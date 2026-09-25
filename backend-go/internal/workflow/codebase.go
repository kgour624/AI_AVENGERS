package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/repo"
	"ai_avengers/backend/internal/response"
)

// This file implements 3D: the human-approved working set for an
// existing-codebase workflow.
//
// THE RULE THE WHOLE FILE SERVES: a domain expert may only read files the
// client has explicitly approved. The system proposes (ranked, with reasons),
// the client decides, and nothing an expert reads is ever the result of the
// system deciding on its own.
//
// It is deliberately ONE table with three states rather than "suggestions" plus
// a separate "manifest". Two tables holding the same file list would be two
// copies of one truth, and a sync bug between them would surface as an expert
// reading an unapproved file — a safety failure, not a display glitch.

// Working-set states. The CHECK constraint in migration 041 is the authority.
const (
	CodebaseStatusPending  = "pending"
	CodebaseStatusApproved = "approved"
	CodebaseStatusRejected = "rejected"
)

// Who put a file forward.
const (
	CodebaseSourceExpert = "expert"
	CodebaseSourceClient = "client"
)

// Blackboard event types emitted by the approval loop.
const (
	EventFileSuggested  = "file_suggested"
	EventFileApproved   = "file_approved"
	EventFileRejected   = "file_rejected"
	EventFileAddedByCli = "client_file_added"
)

// RepoCodeSource is the slice of the repository index this package needs.
//
// Declared here (by the consumer) rather than taking *repo.Service: the
// dependency points inwards, the surface is two methods instead of the whole
// repository API, and a test can supply a stub without a database.
type RepoCodeSource interface {
	SuggestRepoFiles(ctx context.Context, projectID uuid.UUID, requirement string, limit int) ([]repo.RepoFileSuggestion, error)
	RepoFileExists(ctx context.Context, projectID uuid.UUID, path string) (bool, error)
}

// CodebaseFile is one entry in a workflow's working set.
type CodebaseFile struct {
	ID         uuid.UUID  `json:"id"`
	WorkflowID uuid.UUID  `json:"workflow_id"`
	Path       string     `json:"path"`
	Status     string     `json:"status"`
	Source     string     `json:"source"`
	Reason     string     `json:"reason"`
	Score      float64    `json:"score"`
	HopDepth   int        `json:"hop_depth"`
	DecidedAt  *time.Time `json:"decided_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CodebaseService owns the working-set rules.
type CodebaseService struct {
	db     *pgxpool.Pool
	store  *blackboard.Store
	repo   RepoCodeSource
	logger *zap.Logger
}

func NewCodebaseService(db *pgxpool.Pool, store *blackboard.Store, repoSource RepoCodeSource, logger *zap.Logger) *CodebaseService {
	return &CodebaseService{db: db, store: store, repo: repoSource, logger: logger}
}

// workflowProject returns the project a workflow belongs to, and refuses if the
// workflow is not in the existing-codebase environment.
//
// WHY the mode check lives here: a scratch workflow has no repository context,
// so proposing repository files for it would produce a working set that means
// nothing. Refusing with a clear error is better than admitting rows that no
// later phase will ever read.
func (s *CodebaseService) workflowProject(ctx context.Context, workflowID uuid.UUID) (uuid.UUID, error) {
	var projectID uuid.UUID
	var mode, status string
	err := s.db.QueryRow(ctx,
		`SELECT project_id, mode, status FROM workflows WHERE id=$1`, workflowID,
	).Scan(&projectID, &mode, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrWorkflowNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("load workflow: %w", err)
	}
	if mode != ModeExistingCodebase {
		return uuid.Nil, ErrNotCodebaseWorkflow
	}
	return projectID, nil
}

// ErrNotCodebaseWorkflow lets the handler explain that this workflow is in the
// scratch environment, where a repository working set has no meaning.
//
// ErrWorkflowNotFound already exists in this package (client_repo.go) and
// intentionally covers "no such workflow" and "belongs to another client"
// together, so it is reused here rather than declared again.
var ErrNotCodebaseWorkflow = errors.New("workflow is not in the existing-codebase environment")

// ListFiles returns the whole working set, newest proposals first.
func (s *CodebaseService) ListFiles(ctx context.Context, workflowID uuid.UUID) ([]CodebaseFile, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, workflow_id, path, status, source, reason, score, hop_depth,
		        decided_at, created_at
		   FROM workflow_codebase_files
		  WHERE workflow_id=$1
		  ORDER BY status, score DESC, path`,
		workflowID,
	)
	if err != nil {
		return nil, fmt.Errorf("list codebase files: %w", err)
	}
	defer rows.Close()

	files := []CodebaseFile{}
	for rows.Next() {
		var file CodebaseFile
		if err := rows.Scan(&file.ID, &file.WorkflowID, &file.Path, &file.Status,
			&file.Source, &file.Reason, &file.Score, &file.HopDepth,
			&file.DecidedAt, &file.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan codebase file: %w", err)
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

// Manifest returns the approved paths — the only files an expert may read.
//
// This is the method 3E enforces against: the runner asks for the manifest and
// refuses anything outside it.
func (s *CodebaseService) Manifest(ctx context.Context, workflowID uuid.UUID) ([]string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT path FROM workflow_codebase_files
		  WHERE workflow_id=$1 AND status=$2
		  ORDER BY path`,
		workflowID, CodebaseStatusApproved,
	)
	if err != nil {
		return nil, fmt.Errorf("load codebase manifest: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan codebase manifest: %w", err)
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

// requirementText reads the workflow's requirement from the blackboard.
//
// Reads the LATEST requirement_captured event rather than a column, because
// that is already the source of truth the runner uses (runner.loadRequirement);
// introducing a second place to look would let the two disagree.
func (s *CodebaseService) requirementText(ctx context.Context, workflowID uuid.UUID) string {
	var raw []byte
	err := s.db.QueryRow(ctx,
		`SELECT content FROM blackboard_events
		  WHERE workflow_id=$1 AND event_type='requirement_captured'
		  ORDER BY created_at DESC LIMIT 1`,
		workflowID,
	).Scan(&raw)
	if err != nil {
		return ""
	}
	var payload struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Text)
}

// SuggestForWorkflow asks the repository ranker for candidate files and records
// them as pending.
//
// WHY ON CONFLICT DO NOTHING: a client who has already rejected a file must not
// have that decision silently overwritten by the next suggestion run. The
// unique index is (workflow_id, path), so an existing row — of ANY status — is
// left exactly as it is.
func (s *CodebaseService) SuggestForWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]CodebaseFile, error) {
	projectID, err := s.workflowProject(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	requirement := s.requirementText(ctx, workflowID)
	if requirement == "" {
		return nil, ErrNoRequirement
	}

	suggestions, err := s.repo.SuggestRepoFiles(ctx, projectID, requirement, limit)
	if err != nil {
		return nil, fmt.Errorf("suggest repo files: %w", err)
	}

	added := []CodebaseFile{}
	for _, suggestion := range suggestions {
		var file CodebaseFile
		err := s.db.QueryRow(ctx,
			`INSERT INTO workflow_codebase_files
				(workflow_id, project_id, path, status, source, reason, score, hop_depth)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (workflow_id, path) DO NOTHING
			 RETURNING id, workflow_id, path, status, source, reason, score, hop_depth,
			           decided_at, created_at`,
			workflowID, projectID, suggestion.Path,
			CodebaseStatusPending, CodebaseSourceExpert,
			suggestion.Reason, suggestion.Score, 0,
		).Scan(&file.ID, &file.WorkflowID, &file.Path, &file.Status, &file.Source,
			&file.Reason, &file.Score, &file.HopDepth, &file.DecidedAt, &file.CreatedAt)

		if errors.Is(err, pgx.ErrNoRows) {
			continue // already known to this workflow (any status): keep the human's decision
		}
		if err != nil {
			return nil, fmt.Errorf("store codebase suggestion: %w", err)
		}
		added = append(added, file)

		s.postEvent(ctx, workflowID, EventFileSuggested, map[string]interface{}{
			"path": file.Path, "reason": file.Reason, "score": file.Score,
		})
	}

	s.logger.Info("codebase suggestions recorded",
		zap.String("workflow_id", workflowID.String()),
		zap.Int("candidates", len(suggestions)),
		zap.Int("new_rows", len(added)),
	)
	return added, nil
}

// ErrNoRequirement is returned when a workflow has no requirement text yet, so
// there is nothing to rank files against.
var ErrNoRequirement = errors.New("workflow has no requirement text to rank files against")

// AddFile records a file the CLIENT chose directly. The client is the approver,
// so the row is approved on arrival — that is the whole point of letting them
// bypass the suggestion queue.
func (s *CodebaseService) AddFile(ctx context.Context, workflowID uuid.UUID, path string) (*CodebaseFile, error) {
	projectID, err := s.workflowProject(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrEmptyPath
	}

	// Verify against the stored tree: an approved path that does not exist would
	// make the approval record claim something untrue.
	exists, err := s.repo.RepoFileExists(ctx, projectID, path)
	if err != nil {
		return nil, fmt.Errorf("verify repo file: %w", err)
	}
	if !exists {
		return nil, ErrRepoPathNotFound
	}

	var file CodebaseFile
	err = s.db.QueryRow(ctx,
		`INSERT INTO workflow_codebase_files
			(workflow_id, project_id, path, status, source, reason, decided_by_client, decided_at)
		 VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW())
		 ON CONFLICT (workflow_id, path) DO UPDATE SET
			status=EXCLUDED.status,
			source=EXCLUDED.source,
			decided_by_client=TRUE,
			decided_at=NOW()
		 RETURNING id, workflow_id, path, status, source, reason, score, hop_depth,
		           decided_at, created_at`,
		workflowID, projectID, path,
		CodebaseStatusApproved, CodebaseSourceClient,
		"added by the client",
	).Scan(&file.ID, &file.WorkflowID, &file.Path, &file.Status, &file.Source,
		&file.Reason, &file.Score, &file.HopDepth, &file.DecidedAt, &file.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add codebase file: %w", err)
	}

	s.postEvent(ctx, workflowID, EventFileAddedByCli, map[string]interface{}{"path": file.Path})
	return &file, nil
}

// ErrEmptyPath / ErrRepoPathNotFound keep handler status mapping simple.
var (
	ErrEmptyPath        = errors.New("path is required")
	ErrRepoPathNotFound = errors.New("path is not part of the repository's stored tree")
)

// Decide approves or rejects one file.
func (s *CodebaseService) Decide(ctx context.Context, workflowID, fileID uuid.UUID, decision string) (*CodebaseFile, error) {
	status, err := normalizeDecision(decision)
	if err != nil {
		return nil, err
	}
	if _, err := s.workflowProject(ctx, workflowID); err != nil {
		return nil, err
	}

	var file CodebaseFile
	err = s.db.QueryRow(ctx,
		`UPDATE workflow_codebase_files
		    SET status=$1, decided_by_client=TRUE, decided_at=NOW()
		  WHERE id=$2 AND workflow_id=$3
		  RETURNING id, workflow_id, path, status, source, reason, score, hop_depth,
		           decided_at, created_at`,
		status, fileID, workflowID,
	).Scan(&file.ID, &file.WorkflowID, &file.Path, &file.Status, &file.Source,
		&file.Reason, &file.Score, &file.HopDepth, &file.DecidedAt, &file.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCodebaseFileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("decide codebase file: %w", err)
	}

	s.postEvent(ctx, workflowID, decisionEvent(status), map[string]interface{}{"path": file.Path})
	return &file, nil
}

// DecideBulk applies one decision to many files. It reports how many rows
// actually changed, so the UI can tell "3 approved" from "3 already approved".
func (s *CodebaseService) DecideBulk(ctx context.Context, workflowID uuid.UUID, fileIDs []uuid.UUID, decision string) (int, error) {
	status, err := normalizeDecision(decision)
	if err != nil {
		return 0, err
	}
	if _, err := s.workflowProject(ctx, workflowID); err != nil {
		return 0, err
	}
	if len(fileIDs) == 0 {
		return 0, nil
	}

	rows, err := s.db.Query(ctx,
		`UPDATE workflow_codebase_files
		    SET status=$1, decided_by_client=TRUE, decided_at=NOW()
		  WHERE workflow_id=$2 AND id = ANY($3) AND status <> $1
		  RETURNING path`,
		status, workflowID, fileIDs,
	)
	if err != nil {
		return 0, fmt.Errorf("bulk decide codebase files: %w", err)
	}
	defer rows.Close()

	var changed int
	eventType := decisionEvent(status)
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return 0, fmt.Errorf("scan bulk decision: %w", err)
		}
		changed++
		s.postEvent(ctx, workflowID, eventType, map[string]interface{}{"path": p})
	}
	return changed, rows.Err()
}

// ErrCodebaseFileNotFound reports that the file is not in this workflow's set.
var ErrCodebaseFileNotFound = errors.New("codebase file not found in this workflow")

func normalizeDecision(decision string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "approve", "approved":
		return CodebaseStatusApproved, nil
	case "reject", "rejected":
		return CodebaseStatusRejected, nil
	default:
		return "", ErrInvalidDecision
	}
}

func decisionEvent(status string) string {
	if status == CodebaseStatusApproved {
		return EventFileApproved
	}
	return EventFileRejected
}

// ErrInvalidDecision reports a decision word the API does not accept.
var ErrInvalidDecision = errors.New("decision must be approve or reject")

// postEvent is best-effort observability: the row is already committed, and the
// approval itself must not fail because a timeline event could not be written.
func (s *CodebaseService) postEvent(ctx context.Context, workflowID uuid.UUID, eventType string, content map[string]interface{}) {
	if s.store == nil {
		return
	}
	if _, err := s.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      eventType,
		PostedByClient: true, // every event here records a client-side decision
		Content:        content,
	}); err != nil {
		s.logger.Warn("codebase event post failed (non-fatal)",
			zap.String("workflow_id", workflowID.String()),
			zap.String("event_type", eventType),
			zap.Error(err),
		)
	}
}

// ---------------------------------------------------------------------------
// HTTP layer
// ---------------------------------------------------------------------------

// CodebaseHandler exposes the working set. It takes the DB directly for the
// ownership check, so it does not depend on the workflow handler's internals.
type CodebaseHandler struct {
	svc    *CodebaseService
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewCodebaseHandler(svc *CodebaseService, db *pgxpool.Pool, logger *zap.Logger) *CodebaseHandler {
	return &CodebaseHandler{svc: svc, db: db, logger: logger}
}

// assertWorkflowAccess allows the owning client, or any admin.
//
// WHY not the tenant helper used elsewhere: a workflow's owner is a single
// client_id, and the working set is that client's private decision record — a
// same-tenant colleague must not approve files on their behalf.
func (h *CodebaseHandler) assertWorkflowAccess(c *gin.Context, workflowID uuid.UUID) bool {
	clientID, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "invalid session")
		return false
	}
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	var owner uuid.UUID
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT client_id FROM workflows WHERE id=$1`, workflowID,
	).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		response.NotFound(c, "workflow")
		return false
	}
	if err != nil {
		response.InternalError(c)
		return false
	}
	if roleStr != "admin" && owner != clientID {
		response.Forbidden(c, "workflow does not belong to this client")
		return false
	}
	return true
}

// parseWorkflowID reads and validates the :id param, replying on failure.
func parseWorkflowID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return uuid.Nil, false
	}
	return id, true
}

// ListFiles GET /workflows/:id/codebase/files
func (h *CodebaseHandler) ListFiles(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}
	files, err := h.svc.ListFiles(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Error("list codebase files failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	approved := 0
	for _, file := range files {
		if file.Status == CodebaseStatusApproved {
			approved++
		}
	}
	response.OK(c, gin.H{"files": files, "count": len(files), "approved": approved})
}

// Suggest POST /workflows/:id/codebase/suggest
func (h *CodebaseHandler) Suggest(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}
	var body struct {
		Limit int `json:"limit"`
	}
	_ = c.ShouldBindJSON(&body) // an empty body is valid: default limit applies

	added, err := h.svc.SuggestForWorkflow(c.Request.Context(), workflowID, body.Limit)
	switch {
	case errors.Is(err, ErrNoRequirement):
		response.Conflict(c, "this workflow has no requirement text yet, so there is nothing to rank files against")
		return
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("suggest codebase files failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"suggestions": added, "count": len(added)})
}

// AddFile POST /workflows/:id/codebase/files
func (h *CodebaseHandler) AddFile(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	file, err := h.svc.AddFile(c.Request.Context(), workflowID, body.Path)
	switch {
	case errors.Is(err, ErrEmptyPath):
		response.BadRequest(c, "MISSING_PATH", "path is required")
		return
	case errors.Is(err, ErrRepoPathNotFound):
		response.NotFound(c, "repo file")
		return
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("add codebase file failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, file)
}

// Decide POST /workflows/:id/codebase/files/:fileId/decide
func (h *CodebaseHandler) Decide(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}
	fileID, err := uuid.Parse(c.Param("fileId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid codebase file ID")
		return
	}
	var body struct {
		Decision string `json:"decision"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	file, err := h.svc.Decide(c.Request.Context(), workflowID, fileID, body.Decision)
	switch {
	case errors.Is(err, ErrInvalidDecision):
		response.BadRequest(c, "INVALID_DECISION", "decision must be approve or reject")
		return
	case errors.Is(err, ErrCodebaseFileNotFound):
		response.NotFound(c, "codebase file")
		return
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("decide codebase file failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, file)
}

// DecideBulk POST /workflows/:id/codebase/files/decide-bulk
func (h *CodebaseHandler) DecideBulk(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}
	var body struct {
		FileIDs  []uuid.UUID `json:"file_ids"`
		Decision string      `json:"decision"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	changed, err := h.svc.DecideBulk(c.Request.Context(), workflowID, body.FileIDs, body.Decision)
	switch {
	case errors.Is(err, ErrInvalidDecision):
		response.BadRequest(c, "INVALID_DECISION", "decision must be approve or reject")
		return
	case errors.Is(err, ErrNotCodebaseWorkflow):
		response.Conflict(c, "this workflow is not in the existing-codebase environment")
		return
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
		return
	case err != nil:
		h.logger.Error("bulk decide codebase files failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"changed": changed})
}

// Manifest GET /workflows/:id/codebase/manifest
func (h *CodebaseHandler) Manifest(c *gin.Context) {
	workflowID, ok := parseWorkflowID(c)
	if !ok || !h.assertWorkflowAccess(c, workflowID) {
		return
	}
	paths, err := h.svc.Manifest(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Error("load codebase manifest failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"paths": paths, "count": len(paths)})
}
