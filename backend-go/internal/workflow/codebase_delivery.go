package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// This file implements 3F: delivering an existing-codebase workflow's changes as
// a patch against the exact revision the work was based on.
//
// THE SAFETY INVARIANT: the client's repository is never written to. Changes
// leave as a patch the client applies themselves (and, separately, the existing
// §18 GitExporter can push a mirror repo the client controls). Nothing in this
// file pushes anywhere.
//
// WHY a patch and not "here are the changed files": a patch is reviewable before
// it is applied, applies against a named revision, and can be rejected as a
// whole. Handing over loose files would move the job of working out what
// changed onto the person least able to do it safely.

// ErrDeliveryUnavailable reports that this deployment has no workspace volume,
// so there is nothing to diff.
var ErrDeliveryUnavailable = errors.New("delivery is unavailable: no workspace root is configured")

// ErrNoBaseline reports that the workflow has no captured baseline yet, so a
// diff would not mean anything.
var ErrNoBaseline = errors.New("no baseline captured yet: start the workflow so the approved files are seeded")

// CodebaseChangedFile is one entry of a patch's file list.
type CodebaseChangedFile struct {
	// Status is git's name-status code: A added, M modified, D deleted,
	// R renamed (with a similarity score, e.g. R100).
	Status string `json:"status"`
	Path   string `json:"path"`
}

// CodebaseDelivery is the stored patch and what it is relative to.
type CodebaseDelivery struct {
	WorkflowID    uuid.UUID             `json:"workflow_id"`
	BaseCommitSHA string                `json:"base_commit_sha"`
	BaselineRef   string                `json:"baseline_ref"`
	ChangedFiles  []CodebaseChangedFile `json:"changed_files"`
	PatchBytes    int                   `json:"patch_bytes"`
	GeneratedAt   *time.Time            `json:"generated_at"`
}

// captureBaseline commits the seeded working set and records it as the point the
// workflow's own changes are measured from.
//
// WHY a git commit rather than storing file hashes: git already computes
// content-addressed diffs, and the workspace is a git repository by the time the
// merge runs. Reusing it keeps "what changed" consistent with the same mechanism
// the harness uses everywhere else.
//
// The FIRST baseline wins. A later re-seed must not move the reference point,
// because that would silently drop changes already made from the delivered
// patch.
func (s *CodebaseService) captureBaseline(ctx context.Context, workflowID, projectID uuid.UUID, mainDir string) error {
	if err := ensureGitRepo(ctx, mainDir); err != nil {
		return fmt.Errorf("prepare workspace repo: %w", err)
	}
	if _, err := gitRun(ctx, mainDir, "add", "-A"); err != nil {
		return fmt.Errorf("stage baseline: %w", err)
	}

	// Seeding an unchanged working set a second time has nothing to commit.
	// That is not a failure: the existing baseline is already correct.
	if out, err := gitRun(ctx, mainDir, "commit", "-m", "baseline: approved working set"); err != nil {
		if !strings.Contains(out, "nothing to commit") && !strings.Contains(out, "nothing added") {
			return fmt.Errorf("commit baseline: %w (%s)", err, strings.TrimSpace(out))
		}
	}

	refOut, err := gitRun(ctx, mainDir, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("resolve baseline ref: %w", err)
	}
	baselineRef := strings.TrimSpace(refOut)

	baseSHA, err := s.repo.BaseCommitSHA(ctx, projectID)
	if err != nil {
		// Non-fatal: the patch is still valid, it just cannot name the upstream
		// revision it applies to. Saying so beats refusing to deliver.
		s.logger.Warn("codebase baseline: upstream base commit unavailable",
			zap.String("workflow_id", workflowID.String()), zap.Error(err))
		baseSHA = ""
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO codebase_deliveries (workflow_id, project_id, base_commit_sha, baseline_ref)
		 VALUES ($1, $2, NULLIF($3, ''), $4)
		 ON CONFLICT (workflow_id) DO UPDATE SET
			base_commit_sha = COALESCE(EXCLUDED.base_commit_sha, codebase_deliveries.base_commit_sha),
			-- keep the ORIGINAL baseline: moving it would drop already-made changes
			baseline_ref = COALESCE(codebase_deliveries.baseline_ref, EXCLUDED.baseline_ref),
			updated_at = NOW()`,
		workflowID, projectID, baseSHA, baselineRef,
	)
	if err != nil {
		return fmt.Errorf("record codebase baseline: %w", err)
	}

	s.logger.Info("codebase baseline captured",
		zap.String("workflow_id", workflowID.String()),
		zap.String("baseline_ref", baselineRef),
		zap.String("base_commit_sha", baseSHA),
	)
	return nil
}

// GeneratePatch diffs the workflow's workspace against its baseline and stores
// the result.
func (s *CodebaseService) GeneratePatch(ctx context.Context, workflowID uuid.UUID) (*CodebaseDelivery, error) {
	if s.workspaceRoot == "" {
		return nil, ErrDeliveryUnavailable
	}
	// Rejects scratch workflows and confirms the workflow exists.
	if _, err := s.workflowProject(ctx, workflowID); err != nil {
		return nil, err
	}

	delivery, err := s.loadDelivery(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if delivery.BaselineRef == "" {
		return nil, ErrNoBaseline
	}

	mainDir := filepath.Join(s.workspaceRoot, workflowID.String(), "main")

	patch, err := gitRun(ctx, mainDir, "diff", delivery.BaselineRef, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("generate patch: %w", err)
	}
	nameStatus, err := gitRun(ctx, mainDir, "diff", "--name-status", delivery.BaselineRef, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("list changed files: %w", err)
	}
	changed := parseNameStatus(nameStatus)

	encoded, err := json.Marshal(changed)
	if err != nil {
		return nil, fmt.Errorf("encode changed files: %w", err)
	}

	now := time.Now()
	err = s.db.QueryRow(ctx,
		`UPDATE codebase_deliveries
		    SET patch=$1, patch_bytes=$2, changed_files=$3, generated_at=$4, updated_at=NOW()
		  WHERE workflow_id=$5
		  RETURNING generated_at`,
		patch, len(patch), encoded, now, workflowID,
	).Scan(&now)
	if err != nil {
		return nil, fmt.Errorf("store patch: %w", err)
	}

	delivery.ChangedFiles = changed
	delivery.PatchBytes = len(patch)
	delivery.GeneratedAt = &now

	s.logger.Info("codebase patch generated",
		zap.String("workflow_id", workflowID.String()),
		zap.Int("changed_files", len(changed)),
		zap.Int("patch_bytes", len(patch)),
	)
	return delivery, nil
}

// GetDelivery returns the stored delivery without regenerating it.
func (s *CodebaseService) GetDelivery(ctx context.Context, workflowID uuid.UUID) (*CodebaseDelivery, error) {
	if _, err := s.workflowProject(ctx, workflowID); err != nil {
		return nil, err
	}
	return s.loadDelivery(ctx, workflowID)
}

// PatchText returns the stored patch body for download.
func (s *CodebaseService) PatchText(ctx context.Context, workflowID uuid.UUID) (string, error) {
	if _, err := s.workflowProject(ctx, workflowID); err != nil {
		return "", err
	}
	var patch string
	err := s.db.QueryRow(ctx,
		`SELECT patch FROM codebase_deliveries WHERE workflow_id=$1`, workflowID,
	).Scan(&patch)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoBaseline
	}
	if err != nil {
		return "", fmt.Errorf("load patch: %w", err)
	}
	if patch == "" {
		return "", ErrNoBaseline
	}
	return patch, nil
}

// loadDelivery reads the delivery row, returning an empty shell when the
// workflow has not been seeded yet rather than an error: "nothing has happened"
// is a normal answer for a workflow that has not started.
func (s *CodebaseService) loadDelivery(ctx context.Context, workflowID uuid.UUID) (*CodebaseDelivery, error) {
	var (
		baseSHA     string
		baselineRef string
		rawFiles    []byte
		patchBytes  int
		generatedAt *time.Time
	)
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(base_commit_sha, ''), COALESCE(baseline_ref, ''),
		        changed_files, patch_bytes, generated_at
		   FROM codebase_deliveries WHERE workflow_id=$1`,
		workflowID,
	).Scan(&baseSHA, &baselineRef, &rawFiles, &patchBytes, &generatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &CodebaseDelivery{
			WorkflowID:   workflowID,
			ChangedFiles: []CodebaseChangedFile{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load delivery: %w", err)
	}

	changed := []CodebaseChangedFile{}
	if len(rawFiles) > 0 {
		if err := json.Unmarshal(rawFiles, &changed); err != nil {
			// A malformed list must not take the endpoint down; report an empty
			// list and let the log carry the real story.
			s.logger.Warn("codebase delivery: changed_files could not be decoded",
				zap.String("workflow_id", workflowID.String()), zap.Error(err))
			changed = []CodebaseChangedFile{}
		}
	}

	return &CodebaseDelivery{
		WorkflowID:    workflowID,
		BaseCommitSHA: baseSHA,
		BaselineRef:   baselineRef,
		ChangedFiles:  changed,
		PatchBytes:    patchBytes,
		GeneratedAt:   generatedAt,
	}, nil
}

// parseNameStatus turns `git diff --name-status` output into entries.
//
// Format is TAB-separated: "M\tpath", "A\tpath", and for renames
// "R100\told\tnew" — where the LAST field is the path that now exists, which is
// the one a reader needs.
func parseNameStatus(output string) []CodebaseChangedFile {
	files := []CodebaseChangedFile{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		files = append(files, CodebaseChangedFile{
			Status: strings.TrimSpace(parts[0]),
			Path:   strings.TrimSpace(parts[len(parts)-1]),
		})
	}
	return files
}

// ---------------------------------------------------------------------------
// HTTP layer
// ---------------------------------------------------------------------------

// GeneratePatch POST /workflows/:id/codebase/patch
func (h *CodebaseHandler) GeneratePatch(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}

	delivery, err := h.svc.GeneratePatch(c.Request.Context(), workflowID)
	switch {
	case errors.Is(err, ErrNoBaseline):
		response.Conflict(c, "no baseline yet: start the workflow so the approved files are seeded first")
		return
	case errors.Is(err, ErrDeliveryUnavailable):
		response.ServiceUnavailable(c, "patch generation is unavailable in this deployment")
		return
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("generate codebase patch failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, delivery)
}

// GetPatch GET /workflows/:id/codebase/patch
func (h *CodebaseHandler) GetPatch(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}

	delivery, err := h.svc.GetDelivery(c.Request.Context(), workflowID)
	switch {
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("load codebase delivery failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, delivery)
}

// DownloadPatch GET /workflows/:id/codebase/patch/download
//
// Served as text/x-patch with a filename so the client saves a real .patch and
// can review it before applying, which is the whole point of delivering a patch.
func (h *CodebaseHandler) DownloadPatch(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}

	patch, err := h.svc.PatchText(c.Request.Context(), workflowID)
	switch {
	case errors.Is(err, ErrNoBaseline):
		response.Conflict(c, "no patch has been generated for this workflow yet")
		return
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("load codebase patch failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="workflow-%s.patch"`, workflowID.String()[:8]))
	c.Data(200, "text/x-patch; charset=utf-8", []byte(patch))
}
