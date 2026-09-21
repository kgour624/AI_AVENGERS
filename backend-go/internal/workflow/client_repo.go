package workflow

// Shared plumbing for the two features that act on a CLIENT's git remote:
//
//   export_git.go     §18 — push the finished harness to the client's repo
//   code_feedback.go  §17 — clone the repo the client built, and read it
//
// Both need the same four things, and both get them from here so a fix lands
// once: a credential, a validated remote, a git invocation that cannot hang,
// and a guarantee that the credential never reaches disk or a log line.
//
// The remote validation and the log scrubbing live in client_remote.go — the
// standard-library-only half, split out because it is the security boundary and
// because it is the part that can be exercised directly in an environment with
// no module cache.
//
// WHY AN INTERFACE INSTEAD OF IMPORTING internal/repo
//
// The token lives encrypted in repo_connections and only internal/repo holds
// the key. This package needs the token and nothing else from there. Taking
// *repo.Service as a field would make every future change in internal/repo a
// possible compile break here, for a one-method dependency. RepoTokenSource is
// that one method; *repo.Service already satisfies it.
//
// THE SECURITY RULES, STATED ONCE
//
// A client's git token is the most dangerous value this codebase handles: it
// is write access to their source. Three rules, all enforced in code rather
// than left to the caller to remember:
//
//  1. The remote host is checked against a fixed allowlist BEFORE the token is
//     ever read. Without that check, a request body saying
//     {"repo_url": "https://evil.example/x"} would make us hand the client's
//     GitHub token to evil.example. That is not a hypothetical: the URL is
//     client-supplied input on both endpoints. Enforced by parseRemoteTarget.
//  2. No token is ever written into a git config. `git push <url>` takes the
//     URL as an argument and persists nothing; `git clone <url>` DOES persist
//     it as remote.origin.url, so cloneClientRepo removes that remote
//     immediately. Both paths then call assertNoSecretOnDisk.
//  3. Every byte of git output passes through scrubSecrets before it is logged
//     or returned. git prints the remote URL in several of its error messages,
//     credentials included.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RepoTokenSource yields a client's git provider token. Implemented by
// *repo.Service.
type RepoTokenSource interface {
	AccessTokenForProject(ctx context.Context, projectID uuid.UUID, provider string) (string, error)
}

// Errors shared by both features.
var (
	// ErrWorkflowNotFound covers "no such workflow" and "belongs to another
	// client" together, for the same reason ErrChatNotFound does
	// (chat_service.go): answering differently would tell a caller that
	// someone else's workflow exists.
	ErrWorkflowNotFound = errors.New("workflow not found")

	// ErrNoHarness means the workflow has not produced a harness yet, so
	// there is nothing to push and nothing to compare code against.
	ErrNoHarness = errors.New("this workflow has no harness yet — run the design phase first")

	// ErrNoProviderToken means we could not obtain a usable credential for the
	// client's git provider.
	//
	// Every cause — nothing connected, the wrong provider connected, a token
	// that no longer decrypts — collapses into this one sentinel on purpose.
	// All three are fixed by the same client action (reconnect the provider),
	// all three deserve the same HTTP status, and collapsing them here is what
	// lets the HTTP layer classify the failure WITHOUT importing internal/repo
	// just to name its error values. The specific cause is not lost: it stays
	// in the wrapped message.
	ErrNoProviderToken = errors.New("no usable git provider credential for this project")
)

// gitRun runs one git command and returns its combined output.
//
// GIT_TERMINAL_PROMPT=0 is the load-bearing line. With it unset, a rejected
// token does not fail — git asks for a username on the terminal, there is no
// terminal, and the command blocks until the context deadline. That is a
// request that hangs for minutes instead of returning "authentication failed",
// and it is the single most likely thing to go wrong on these two endpoints.
//
// The caller is responsible for scrubbing the returned text. gitRun cannot do
// it: it does not know which of its arguments is a secret.
func gitRun(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ensureGitRepo makes path a git repository with a commit identity, if it is
// not one already.
//
// Why this is needed at all — a real bug, found while implementing §18:
//
// §18 assumes the harness workspace "is already a git repository with real
// commit history". It is not. WorkspaceMerger creates main/ with os.MkdirAll
// and rsyncs into it with `--exclude .git`; nothing ever runs `git init`
// there. Consequences, both live before this change:
//
//   - MergeWave's multi-expert path calls commitMerge, whose first command is
//     `git add .` inside main/. In a directory that is not a repository (and
//     whose parents are not either — /workspaces/{id}/main) that command
//     fails, commitMerge returns an error, and the whole wave merge fails.
//   - MergeWave's single-expert path never committed at all, so main/ was a
//     plain directory of files.
//
// So the export had no history to push and, worse, multi-expert merges were
// failing for a reason that looked like a git problem. Fixing it here rather
// than only inside the exporter means the merge path is fixed too, and there
// is one definition of "make this directory a repo" instead of two.
//
// Note on what the history actually contains, so §18's promise is not
// overstated: because rsync excludes .git, per-expert commit authorship does
// NOT survive into main/. main/'s log is one commit per wave merge (plus any
// amendment commits). That is a real decision trail, just a coarser one than
// "every expert's own commits".
func ensureGitRepo(ctx context.Context, path string) error {
	if st, err := os.Stat(filepath.Join(path, ".git")); err == nil && st.IsDir() {
		return nil
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", path, err)
	}
	if out, err := gitRun(ctx, path, "init"); err != nil {
		return fmt.Errorf("git init: %w (output: %s)", err, strings.TrimSpace(out))
	}
	// Same identity initWorkspace uses (aider_runner.go). Set per-repo, not
	// globally, because the container has no global git config and a commit
	// without user.email fails outright.
	if out, err := gitRun(ctx, path, "config", "user.name", "AI Avengers"); err != nil {
		return fmt.Errorf("git config user.name: %w (output: %s)", err, strings.TrimSpace(out))
	}
	if out, err := gitRun(ctx, path, "config", "user.email", "ai@avengers.dev"); err != nil {
		return fmt.Errorf("git config user.email: %w (output: %s)", err, strings.TrimSpace(out))
	}
	return nil
}

// currentBranch returns the checked-out branch name of a repository.
//
// Not assumed to be "main": `git init` picks the branch from
// init.defaultBranch, which is unset in this image, so git's built-in default
// applies and has changed between git versions. Reading it is one command;
// guessing it is a push to a branch that does not exist.
func currentBranch(ctx context.Context, repoPath string) (string, error) {
	out, err := gitRun(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("read current branch: %w (output: %s)", err, strings.TrimSpace(out))
	}
	branch := strings.TrimSpace(out)
	if branch == "" || branch == "HEAD" {
		return "", fmt.Errorf("repository at %s has no branch checked out", repoPath)
	}
	return branch, nil
}

// hasCommits reports whether the repository has at least one commit.
// A fresh `git init` has none, and both pushing and diffing need one.
func hasCommits(ctx context.Context, repoPath string) bool {
	_, err := gitRun(ctx, repoPath, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// headSHA returns the full commit SHA of HEAD.
func headSHA(ctx context.Context, repoPath string) (string, error) {
	out, err := gitRun(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("read HEAD: %w (output: %s)", err, strings.TrimSpace(out))
	}
	return strings.TrimSpace(out), nil
}

// assertNoSecretOnDisk checks that a repository's git config does not contain
// any of the given secrets.
//
// This is §18.4's stated pre-ship check, written as an assertion instead of a
// note: "after a push, git remote -v in the workspace must show no client
// remote and no token anywhere on disk". Cheap enough to run every time, and
// the failure it catches — a token left in .git/config where the next
// developer greps the workspace — is not one to find later.
func assertNoSecretOnDisk(repoPath string, secrets ...string) error {
	data, err := os.ReadFile(filepath.Join(repoPath, ".git", "config"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read git config: %w", err)
	}
	text := string(data)
	for _, s := range secrets {
		if s != "" && strings.Contains(text, s) {
			return fmt.Errorf("a credential was left in %s/.git/config — refusing to continue", repoPath)
		}
	}
	return nil
}

// clientToken fetches a client's provider credential, normalising every
// failure into ErrNoProviderToken. See that variable for why.
func clientToken(ctx context.Context, tokens RepoTokenSource, projectID uuid.UUID, provider string) (string, error) {
	token, err := tokens.AccessTokenForProject(ctx, projectID, provider)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNoProviderToken, err)
	}
	if strings.TrimSpace(token) == "" {
		return "", ErrNoProviderToken
	}
	return token, nil
}

// mainWorkspacePath is the merged harness directory for a workflow. The same
// path AuthoringRunner.seedAuthoringWorkspace reads from and WorkspaceMerger
// writes to; expressed once so the three cannot drift.
func mainWorkspacePath(workspaceRoot string, workflowID uuid.UUID) string {
	return filepath.Join(workspaceRoot, workflowID.String(), "main")
}

// workflowProject loads a workflow's project, checking that this client owns
// it. Same shape and same merge-not-found-with-not-yours rule as
// WorkflowChatService.CreateChat.
func workflowProject(ctx context.Context, db *pgxpool.Pool, workflowID, clientID uuid.UUID) (uuid.UUID, error) {
	var ownerID, projectID uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT client_id, project_id FROM workflows WHERE id = $1`, workflowID,
	).Scan(&ownerID, &projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrWorkflowNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("load workflow: %w", err)
	}
	if ownerID != clientID {
		return uuid.Nil, ErrWorkflowNotFound
	}
	return projectID, nil
}
