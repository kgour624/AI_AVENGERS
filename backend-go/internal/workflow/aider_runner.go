package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// AiderRunner executes implementation/qa tasks using Aider.
//
// PATTERN: RALF Loop (Ralph Loop from Arpit Bhiyani AI Masterclass)
//   File system is context (not JSON blobs)
//   Git history is memory (meaningful commits)
//   Fresh read every iteration (no stale context)
//   Patch-based changes (unified diff format)
//
// FLOW:
//   1. Initialize workspace: /workspaces/{workflow_id}/
//   2. Git init + initial commit
//   3. Load design artifacts from blackboard → seed files
//   4. Run Aider loop:
//      - Observe: git status, test results, build errors
//      - Think: Aider LLM call with expert's training (Gate 1 only)
//      - Act: Apply patches, commit changes
//      - Repeat until TASK_COMPLETE or max iterations
//   5. Post final code artifacts to blackboard
//   6. Cleanup workspace (optional, keep for debugging)
//
// MENTAL MODEL:
//   Expert: Backend Engineer, Task: "Implement user authentication"
//   
//   Iteration 1:
//     Observe: Empty workspace, design artifacts in blackboard
//     Think:   Aider reads expert's training (Gate 1: auth patterns)
//     Act:     Generate auth.go, commit "feat: add user authentication"
//   
//   Iteration 2:
//     Observe: go build fails (missing import)
//     Think:   Aider analyzes error, checks training
//     Act:     Fix import, commit "fix: add missing crypto import"
//   
//   Iteration 3:
//     Observe: go test fails (password hash test)
//     Think:   Aider reads test output, checks training
//     Act:     Fix hash logic, commit "fix: use bcrypt cost 12"
//   
//   Iteration 4:
//     Observe: All tests pass, build succeeds
//     Think:   Task complete
//     Act:     Post code_artifact_produced to blackboard, exit
type AiderRunner struct {
	db           *pgxpool.Pool
	store        *blackboard.Store
	gateway      *gateway.ModelGateway
	workspaceDir string // Base directory: /workspaces/
	logger       *zap.Logger
}

// NewAiderRunner creates a new AiderRunner instance.
//
// workspaceDir: Base directory for all workspaces (e.g., "/workspaces/")
// Each workflow gets isolated subdirectory: {workspaceDir}/{workflow_id}/
func NewAiderRunner(
	db *pgxpool.Pool,
	store *blackboard.Store,
	gw *gateway.ModelGateway,
	workspaceDir string,
	logger *zap.Logger,
) *AiderRunner {
	return &AiderRunner{
		db:           db,
		store:        store,
		gateway:      gw,
		workspaceDir: workspaceDir,
		logger:       logger,
	}
}

// AiderRunRequest contains parameters for running Aider on a task.
type AiderRunRequest struct {
	WorkflowID      uuid.UUID
	Expert          workflowExpert
	TaskID          uuid.UUID
	TaskTitle       string
	TaskDescription string
	WorkflowPhase   string // "implementation" or "qa"
}

// AiderRunResult contains the outcome of an Aider run.
type AiderRunResult struct {
	CommitSHAs []string // Git commit SHAs produced
	Iterations int      // Number of iterations executed
	Completed  bool     // True if task completed successfully
}

// Run executes the Aider loop for one expert's task.
//
// CRITICAL: This is Phase 1 skeleton - no actual Aider calls yet.
// Phase 2 will add runAiderIteration() with real Aider integration.
//
// Current implementation:
//   1. Initialize workspace
//   2. Seed with design artifacts
//   3. TODO: Run Aider loop (Phase 2)
//   4. TODO: Post code artifacts (Phase 3)
func (a *AiderRunner) Run(ctx context.Context, req AiderRunRequest) (*AiderRunResult, error) {
	a.logger.Info("aider runner started",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
		zap.String("task", req.TaskTitle),
		zap.String("phase", req.WorkflowPhase),
	)

	// Step 1: Initialize workspace
	workspacePath := filepath.Join(a.workspaceDir, req.WorkflowID.String())
	if err := a.initWorkspace(ctx, workspacePath); err != nil {
		return nil, fmt.Errorf("init workspace: %w", err)
	}

	// Step 2: Load design artifacts from blackboard → seed files
	if err := a.seedWorkspace(ctx, req.WorkflowID, workspacePath); err != nil {
		return nil, fmt.Errorf("seed workspace: %w", err)
	}

	// Step 3: Run Aider loop (TODO: Phase 2)
	// For now, just return success
	result := &AiderRunResult{
		CommitSHAs: []string{},
		Iterations: 0,
		Completed:  true,
	}

	a.logger.Info("aider runner completed (skeleton)",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
	)

	return result, nil
}

// initWorkspace creates workspace directory and initializes git repo.
//
// MENTAL MODEL:
//   Input: /workspaces/abc-123/
//   Actions:
//     1. mkdir -p /workspaces/abc-123/
//     2. cd /workspaces/abc-123/ && git init
//     3. git commit --allow-empty -m "chore: initialize workspace"
//   Output: Empty git repo with initial commit
//
// WHY initial commit?
//   - Provides base for git diff (HEAD~1)
//   - Allows git log to work immediately
//   - Clean starting point for all changes
func (a *AiderRunner) initWorkspace(ctx context.Context, workspacePath string) error {
	// Create directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Git init
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %w (output: %s)", err, string(output))
	}

	// Configure git user (required for commits)
	cmd = exec.CommandContext(ctx, "git", "config", "user.name", "AI Avengers")
	cmd.Dir = workspacePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config user.name: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "config", "user.email", "ai@avengers.dev")
	cmd.Dir = workspacePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config user.email: %w", err)
	}

	// Initial commit (empty)
	cmd = exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", "chore: initialize workspace")
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w (output: %s)", err, string(output))
	}

	a.logger.Info("workspace initialized",
		zap.String("path", workspacePath),
	)

	return nil
}

// seedWorkspace loads design artifacts from blackboard and creates seed files.
//
// MENTAL MODEL:
//   Blackboard events:
//     - architecture_decision → ARCHITECTURE.md
//     - data_model_proposed → DATA_MODEL.md
//     - api_contract_proposed → API_CONTRACT.yaml
//     - module_design_proposed → MODULE_DESIGN.md
//   
//   Actions:
//     1. Query blackboard for design artifacts
//     2. For each artifact: create file with content
//     3. git add . && git commit -m "chore: seed workspace with design artifacts"
//   
//   Output: Workspace with design files committed
//
// WHY seed from blackboard?
//   - Design artifacts provide context for implementation
//   - Aider can reference architecture decisions
//   - Git history shows what was provided vs generated
func (a *AiderRunner) seedWorkspace(ctx context.Context, workflowID uuid.UUID, workspacePath string) error {
	// Load design artifacts from blackboard
	events, err := a.store.GetByType(ctx, workflowID, []string{
		"architecture_decision",
		"data_model_proposed",
		"api_contract_proposed",
		"module_design_proposed",
	}, 0)
	if err != nil {
		return fmt.Errorf("load design artifacts: %w", err)
	}

	if len(events) == 0 {
		a.logger.Info("no design artifacts to seed",
			zap.String("workflow_id", workflowID.String()),
		)
		return nil
	}

	// Create seed files
	for _, ev := range events {
		filename := a.artifactToFilename(ev.EventType)
		content := string(ev.Content)

		filePath := filepath.Join(workspacePath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}

		a.logger.Debug("seed file created",
			zap.String("file", filename),
			zap.String("event_type", ev.EventType),
		)
	}

	// Git commit seed files
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = workspacePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", "chore: seed workspace with design artifacts")
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if nothing to commit (no design artifacts)
		if !strings.Contains(string(output), "nothing to commit") {
			return fmt.Errorf("git commit: %w (output: %s)", err, string(output))
		}
	}

	a.logger.Info("workspace seeded",
		zap.String("workflow_id", workflowID.String()),
		zap.Int("files", len(events)),
	)

	return nil
}

// artifactToFilename maps blackboard event types to filenames.
//
// MENTAL MODEL:
//   Event Type                → Filename
//   architecture_decision     → ARCHITECTURE.md
//   data_model_proposed       → DATA_MODEL.md
//   api_contract_proposed     → API_CONTRACT.yaml
//   module_design_proposed    → MODULE_DESIGN.md
//   (default)                 → DESIGN.md
func (a *AiderRunner) artifactToFilename(eventType string) string {
	switch eventType {
	case "architecture_decision":
		return "ARCHITECTURE.md"
	case "data_model_proposed":
		return "DATA_MODEL.md"
	case "api_contract_proposed":
		return "API_CONTRACT.yaml"
	case "module_design_proposed":
		return "MODULE_DESIGN.md"
	default:
		return "DESIGN.md"
	}
}
