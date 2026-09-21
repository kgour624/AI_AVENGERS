package workflow

// §18 of docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md — git-push export.
//
// The workflow's output is the harness in main/. ZIP download was the first
// delivery route because it carries no credential. This is the second: push
// main/ straight onto the client's own remote, so they receive the commit trail
// and not just a folder of files.
//
// WHAT THIS DELIBERATELY DOES NOT DO
//
//   - No force push. A push that would overwrite existing history is refused
//     by git and reported as such. §18.2 scopes this to "an existing empty
//     repo, or a new repo to create"; a --force flag here would turn a
//     convenience feature into a way to destroy a client's repository from a
//     single API call.
//   - No remote is added to the workspace. `git push <url> <refspec>` takes the
//     URL as an argument and persists nothing, which is what makes §18.4's
//     "no token anywhere on disk" achievable rather than aspirational. The
//     assertion still runs.
//   - The token is never logged. Every git output string passes through
//     scrubSecrets before it reaches the logger or the caller.
//
// The credential-safety rules and the reason the remote host is validated
// before the token is read live in client_repo.go's file comment.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
)

// GitExporter pushes a workflow's harness to a client git remote.
type GitExporter struct {
	db            *pgxpool.Pool
	store         *blackboard.Store
	tokens        RepoTokenSource
	httpClient    *http.Client
	workspaceRoot string
	logger        *zap.Logger
}

// NewGitExporter wires the exporter.
func NewGitExporter(
	db *pgxpool.Pool,
	store *blackboard.Store,
	tokens RepoTokenSource,
	workspaceRoot string,
	logger *zap.Logger,
) *GitExporter {
	return &GitExporter{
		db:            db,
		store:         store,
		tokens:        tokens,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		workspaceRoot: workspaceRoot,
		logger:        logger,
	}
}

// GitExportRequest is one export.
type GitExportRequest struct {
	WorkflowID uuid.UUID
	ClientID   uuid.UUID
	Provider   string // "github" | "gitlab"
	RepoURL    string // target repo, client-supplied — validated, never trusted
	Branch     string // target branch; defaults to "main"
	CreateRepo bool   // create the repo first if it does not exist
	// Private applies only when CreateRepo is true. nil means private.
	//
	// A pointer, not a bool, so the zero value is not "public". A design
	// document is a client's commercial material; defaulting an absent field to
	// a public repository would publish it, and no client would have asked for
	// that by omitting a field.
	Private *bool
}

// GitExportResult is what the client is told.
type GitExportResult struct {
	RepoURL     string    `json:"repo_url"`
	Branch      string    `json:"branch"`
	CommitSHA   string    `json:"commit_sha"`
	RepoCreated bool      `json:"repo_created"`
	EventID     uuid.UUID `json:"event_id"`
}

// Export pushes the harness.
//
// Order of operations matters and is not cosmetic:
//
//	1 ownership   — before anything is read from disk
//	2 remote host — before the token is read from the database (client_repo.go
//	                rule 1: an unvalidated host must never see a credential)
//	3 harness     — before the token is read, so "nothing to export" costs no
//	                decrypt and no provider call
//	4 token       — last of the inputs, first thing scrubbed from every output
func (e *GitExporter) Export(ctx context.Context, req GitExportRequest) (*GitExportResult, error) {
	projectID, err := workflowProject(ctx, e.db, req.WorkflowID, req.ClientID)
	if err != nil {
		return nil, err
	}

	target, err := parseRemoteTarget(req.Provider, req.RepoURL)
	if err != nil {
		return nil, err
	}

	targetBranch := strings.TrimSpace(req.Branch)
	if targetBranch == "" {
		targetBranch = "main"
	}
	if err := validateBranchName(targetBranch); err != nil {
		return nil, err
	}

	mainPath := mainWorkspacePath(e.workspaceRoot, req.WorkflowID)
	if _, statErr := os.Stat(filepath.Join(mainPath, finalMD)); statErr != nil {
		return nil, fmt.Errorf("%w (no %s in the merged workspace)", ErrNoHarness, finalMD)
	}

	// The workspace may hold files the merger rsynced in but never committed —
	// see ensureGitRepo's comment for why main/ could be a non-repository
	// entirely. Both are made good here, so an export never pushes a HEAD that
	// is missing the newest section files.
	if err := ensureGitRepo(ctx, mainPath); err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}
	if err := gitAddCommit(ctx, mainPath, "chore: harness state at export"); err != nil {
		return nil, fmt.Errorf("export: commit pending harness changes: %w", err)
	}
	if !hasCommits(ctx, mainPath) {
		return nil, fmt.Errorf("%w (the workspace has no commits)", ErrNoHarness)
	}

	localBranch, err := currentBranch(ctx, mainPath)
	if err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}

	token, err := clientToken(ctx, e.tokens, projectID, target.Provider)
	if err != nil {
		return nil, err
	}

	created := false
	if req.CreateRepo {
		created, err = e.createRemoteRepo(ctx, target, token, req.Private)
		if err != nil {
			return nil, err
		}
	}

	refspec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", localBranch, targetBranch)
	out, pushErr := gitRun(ctx, mainPath, "push", target.authedURL(token), refspec)
	clean := strings.TrimSpace(scrubSecrets(out, token))

	// The assertion runs whether the push succeeded or failed. A failed push
	// can still have written something, and "we only check on success" is how a
	// leaked credential survives.
	if diskErr := assertNoSecretOnDisk(mainPath, token); diskErr != nil {
		e.logger.Error("export: credential found on disk after push", zap.Error(diskErr))
		return nil, diskErr
	}

	if pushErr != nil {
		return nil, e.pushFailure(target, targetBranch, clean, pushErr)
	}

	sha, err := headSHA(ctx, mainPath)
	if err != nil {
		return nil, fmt.Errorf("export: %w", err)
	}

	e.logger.Info("harness exported",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("repo", target.CleanURL()),
		zap.String("branch", targetBranch),
		zap.String("commit", sha),
		zap.Bool("repo_created", created),
	)

	// §18.2 step 4. PostedByClient because the client triggered the export and
	// no expert authored it — blackboard_events_poster_check (migration 006)
	// requires exactly one of posted_by_expert_id / posted_by_client, so
	// "the system did it" has to be recorded as one of the two, and the client
	// is the honest one.
	ev, err := e.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     req.WorkflowID,
		EventType:      "harness_exported",
		PostedByClient: true,
		Content: map[string]any{
			"provider":     target.Provider,
			"repo_url":     target.CleanURL(),
			"branch":       targetBranch,
			"commit_sha":   sha,
			"repo_created": created,
		},
	})
	if err != nil {
		// The push already happened. Failing the request now would tell the
		// client the export failed when their repo has the code, and a retry
		// would then hit a non-fast-forward. Log and report success.
		e.logger.Error("export: harness pushed but the event could not be recorded",
			zap.String("workflow_id", req.WorkflowID.String()), zap.Error(err))
		return &GitExportResult{
			RepoURL: target.CleanURL(), Branch: targetBranch,
			CommitSHA: sha, RepoCreated: created,
		}, nil
	}

	return &GitExportResult{
		RepoURL:     target.CleanURL(),
		Branch:      targetBranch,
		CommitSHA:   sha,
		RepoCreated: created,
		EventID:     ev.ID,
	}, nil
}

// pushFailure turns git's output into an error the client can act on.
//
// Three causes account for nearly every failed push here and each needs a
// different action from the client, so they are named instead of all surfacing
// as "git push failed".
func (e *GitExporter) pushFailure(target remoteTarget, branch, cleanOutput string, cause error) error {
	lower := strings.ToLower(cleanOutput)
	switch {
	case strings.Contains(lower, "non-fast-forward") || strings.Contains(lower, "fetch first") ||
		strings.Contains(lower, "rejected"):
		return fmt.Errorf("%s already has commits on %s, so the harness was not pushed. "+
			"Export to an empty repository or to a new branch — this never overwrites existing history. "+
			"(git: %s)", target.CleanURL(), branch, cleanOutput)
	case strings.Contains(lower, "authentication failed") || strings.Contains(lower, "403") ||
		strings.Contains(lower, "could not read username"):
		return fmt.Errorf("the connected %s account cannot write to %s — reconnect the provider with "+
			"repository write access. (git: %s)", target.Provider, target.CleanURL(), cleanOutput)
	case strings.Contains(lower, "not found") || strings.Contains(lower, "404"):
		return fmt.Errorf("%s does not exist or is not visible to the connected %s account. "+
			"Create it first, or set create_repo. (git: %s)", target.CleanURL(), target.Provider, cleanOutput)
	default:
		return fmt.Errorf("push to %s failed: %w (git: %s)", target.CleanURL(), cause, cleanOutput)
	}
}

// ============================================================
// Create-repo — §18.3's optional path
// ============================================================

// createRemoteRepo creates the target repository if it does not already exist.
//
// Returns created=false, nil error when the repository is already there: the
// client asked for "make sure this exists and push to it", and a repo that
// already exists satisfies the first half. Failing instead would mean the only
// way to re-export is to remember whether the last attempt got this far.
func (e *GitExporter) createRemoteRepo(ctx context.Context, target remoteTarget, token string, private *bool) (bool, error) {
	isPrivate := true
	if private != nil {
		isPrivate = *private
	}

	segs := strings.Split(target.Path, "/")
	owner := segs[0]
	name := segs[len(segs)-1]

	switch target.Provider {
	case "github":
		return e.createGitHubRepo(ctx, owner, name, isPrivate, token)
	case "gitlab":
		return e.createGitLabProject(ctx, target.Path, name, isPrivate, token)
	default:
		// Unreachable: parseRemoteTarget only produces allowlisted providers.
		return false, fmt.Errorf("%w: cannot create a repo on %s", ErrUnsupportedRemote, target.Provider)
	}
}

// createGitHubRepo creates a repository under a user or an organisation.
//
// Which endpoint applies is decided by asking who the token belongs to, not by
// guessing from the path. POST /user/repos silently creates the repo under the
// TOKEN's account, ignoring the owner the client typed — so guessing wrong does
// not fail, it creates a repository at the wrong address and then the push
// 404s against the address the client asked for.
func (e *GitExporter) createGitHubRepo(ctx context.Context, owner, name string, private bool, token string) (bool, error) {
	var me struct {
		Login string `json:"login"`
	}
	if err := e.providerJSON(ctx, http.MethodGet, "https://api.github.com/user", token, nil, &me); err != nil {
		return false, fmt.Errorf("github: identify the connected account: %w", err)
	}

	endpoint := fmt.Sprintf("https://api.github.com/orgs/%s/repos", owner)
	if strings.EqualFold(me.Login, owner) {
		endpoint = "https://api.github.com/user/repos"
	}

	body := map[string]any{"name": name, "private": private, "auto_init": false}
	var out struct {
		FullName string `json:"full_name"`
	}
	err := e.providerJSON(ctx, http.MethodPost, endpoint, token, body, &out)
	if err != nil {
		if isAlreadyExists(err) {
			return false, nil
		}
		return false, fmt.Errorf("github: create %s/%s: %w", owner, name, err)
	}
	return true, nil
}

// createGitLabProject creates a project, resolving a group namespace when the
// path is not directly under the token's own user.
func (e *GitExporter) createGitLabProject(ctx context.Context, fullPath, name string, private bool, token string) (bool, error) {
	segs := strings.Split(fullPath, "/")
	namespacePath := strings.Join(segs[:len(segs)-1], "/")

	var me struct {
		Username string `json:"username"`
	}
	if err := e.providerJSON(ctx, http.MethodGet, "https://gitlab.com/api/v4/user", token, nil, &me); err != nil {
		return false, fmt.Errorf("gitlab: identify the connected account: %w", err)
	}

	visibility := "private"
	if !private {
		visibility = "public"
	}
	body := map[string]any{"name": name, "path": segs[len(segs)-1], "visibility": visibility}

	if !strings.EqualFold(me.Username, namespacePath) {
		// A group (or subgroup) namespace. GitLab's create-project endpoint
		// takes a numeric namespace_id, never a path, so it has to be looked up.
		var namespaces []struct {
			ID       int    `json:"id"`
			FullPath string `json:"full_path"`
		}
		nsURL := "https://gitlab.com/api/v4/namespaces?search=" + url.QueryEscape(segs[len(segs)-2])
		if err := e.providerJSON(ctx, http.MethodGet, nsURL, token, nil, &namespaces); err != nil {
			return false, fmt.Errorf("gitlab: resolve namespace %q: %w", namespacePath, err)
		}
		found := 0
		for _, ns := range namespaces {
			if strings.EqualFold(ns.FullPath, namespacePath) {
				found = ns.ID
				break
			}
		}
		if found == 0 {
			return false, fmt.Errorf("gitlab: the connected account cannot see a namespace %q", namespacePath)
		}
		body["namespace_id"] = found
	}

	var out struct {
		PathWithNamespace string `json:"path_with_namespace"`
	}
	err := e.providerJSON(ctx, http.MethodPost, "https://gitlab.com/api/v4/projects", token, body, &out)
	if err != nil {
		if isAlreadyExists(err) {
			return false, nil
		}
		return false, fmt.Errorf("gitlab: create %s: %w", fullPath, err)
	}
	return true, nil
}

// providerAPIError carries a provider's own words plus the status code, so
// isAlreadyExists can classify it without every caller re-parsing strings.
type providerAPIError struct {
	Status int
	Body   string
}

func (p *providerAPIError) Error() string {
	return fmt.Sprintf("provider returned %d: %s", p.Status, p.Body)
}

// isAlreadyExists reports whether a create call failed because the repository
// is already there.
//
// Both providers answer with a 4xx and a message, not with a distinct status
// code, so the message is what has to be read. Anything unrecognised stays a
// real error — the failure mode to avoid is swallowing "insufficient scope" as
// "already exists" and then pushing to a repository that does not exist.
func isAlreadyExists(err error) bool {
	var apiErr *providerAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Status != http.StatusUnprocessableEntity && apiErr.Status != http.StatusBadRequest &&
		apiErr.Status != http.StatusConflict {
		return false
	}
	lower := strings.ToLower(apiErr.Body)
	return strings.Contains(lower, "already exists") ||
		strings.Contains(lower, "already been taken") ||
		strings.Contains(lower, "has already been taken")
}

// providerJSON performs one authenticated provider API call.
//
// Bearer for both providers, matching internal/repo (fetchGitHubTree,
// fetchGitLabTree) — not two different auth schemes for the same two providers
// in one codebase.
func (e *GitExporter) providerJSON(ctx context.Context, method, endpoint, token string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = strings.NewReader(string(encoded))
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		snippet := strings.TrimSpace(scrubSecrets(string(raw), token))
		if len(snippet) > 500 {
			snippet = snippet[:500] + "...(truncated)"
		}
		return &providerAPIError{Status: resp.StatusCode, Body: snippet}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response from %s: %w", endpoint, err)
	}
	return nil
}
