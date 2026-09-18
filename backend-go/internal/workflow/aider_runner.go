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

	// Step 1: Initialize workspace (per-expert isolation)
	// WHY per-expert: Prevent git race conditions in parallel execution
	// Path: /workspaces/{workflow_id}/{expert_id}/
	workspacePath := filepath.Join(a.workspaceDir, req.WorkflowID.String(), req.Expert.ID.String())
	if err := a.initWorkspace(ctx, workspacePath); err != nil {
		return nil, fmt.Errorf("init workspace: %w", err)
	}

	// Step 2: Load design artifacts from blackboard → seed files
	if err := a.seedWorkspace(ctx, req.WorkflowID, workspacePath); err != nil {
		return nil, fmt.Errorf("seed workspace: %w", err)
	}

	// Step 3: Run Aider loop (OTA: Observe → Think → Act)
	result, err := a.runAiderLoop(ctx, req, workspacePath)
	if err != nil {
		return nil, fmt.Errorf("aider loop: %w", err)
	}

	// Step 4: Publish code artifacts to blackboard (Phase 3)
	// WHY: Other experts need to see generated code, frontend needs to display it
	if err := a.publishCodeArtifacts(ctx, req, workspacePath, result.CommitSHAs); err != nil {
		// Non-fatal: log error but don't fail the task
		// Code is already generated and committed to git
		a.logger.Error("failed to publish code artifacts",
			zap.String("workflow_id", req.WorkflowID.String()),
			zap.String("expert", req.Expert.Name),
			zap.Error(err),
		)
	}

	a.logger.Info("aider runner completed",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
		zap.Int("iterations", result.Iterations),
		zap.Bool("completed", result.Completed),
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

// runAiderLoop executes the OTA loop: Observe → Think → Act.
//
// MENTAL MODEL:
//   Max 5 iterations (balance between thoroughness and cost)
//   Each iteration:
//     1. Observe: git status, build errors, test failures
//     2. Think: Aider CLI call with task + observations
//     3. Act: Aider applies patches, commits changes
//   Exit when: TASK_COMPLETE or max iterations reached
//
// CROSS-QUESTIONS:
//   Q: Why max 5 iterations?
//   A: Balance between fixing issues and preventing infinite loops
//      Most tasks complete in 2-3 iterations (initial + 1-2 fixes)
//
//   Q: Why observe before think?
//   A: Provide context (build errors, test failures) to Aider
//      Without observations, Aider can't fix errors
//
//   Q: What if max iterations reached?
//   A: Return partial completion (Completed=false)
//      Caller can decide: retry, escalate, or accept partial
func (a *AiderRunner) runAiderLoop(
	ctx context.Context,
	req AiderRunRequest,
	workspacePath string,
) (*AiderRunResult, error) {
	const maxIterations = 5
	var commitSHAs []string

	for iteration := 1; iteration <= maxIterations; iteration++ {
		a.logger.Info("aider iteration starting",
			zap.Int("iteration", iteration),
			zap.Int("max", maxIterations),
		)

		// Step 1: Observe workspace state
		observations, err := a.observeWorkspace(ctx, workspacePath)
		if err != nil {
			return nil, fmt.Errorf("observe workspace: %w", err)
		}

		// Step 2: Think + Act (Aider iteration)
		commitSHA, taskComplete, err := a.runAiderIteration(
			ctx,
			req,
			workspacePath,
			observations,
			iteration,
		)
		if err != nil {
			return nil, fmt.Errorf("aider iteration %d: %w", iteration, err)
		}

		if commitSHA != "" {
			commitSHAs = append(commitSHAs, commitSHA)
		}

		// Check if task complete
		if taskComplete {
			a.logger.Info("task completed",
				zap.Int("iterations", iteration),
			)
			return &AiderRunResult{
				CommitSHAs: commitSHAs,
				Iterations: iteration,
				Completed:  true,
			}, nil
		}
	}

	// Max iterations reached
	a.logger.Warn("max iterations reached",
		zap.Int("iterations", maxIterations),
	)

	return &AiderRunResult{
		CommitSHAs: commitSHAs,
		Iterations: maxIterations,
		Completed:  false, // Partial completion
	}, nil
}

// observeWorkspace gathers current workspace state for Aider.
//
// MENTAL MODEL:
//   Observations = context for Aider's next iteration
//   Includes:
//     - Git status (what files changed)
//     - Build errors (compile failures)
//     - Test failures (test output)
//
// CROSS-QUESTIONS:
//   Q: Why git status?
//   A: Shows what files Aider created/modified
//      Helps Aider understand current state
//
//   Q: Why build errors?
//   A: Aider needs to see compile errors to fix them
//      Without errors, Aider doesn't know what's wrong
//
//   Q: Why test failures?
//   A: Aider needs test output to fix failing tests
//      Test output shows expected vs actual behavior
func (a *AiderRunner) observeWorkspace(
	ctx context.Context,
	workspacePath string,
) (string, error) {
	var observations strings.Builder

	// Git status
	cmd := exec.CommandContext(ctx, "git", "status", "--short")
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Non-fatal: continue without git status
		a.logger.Warn("git status failed", zap.Error(err))
	} else if len(output) > 0 {
		observations.WriteString("Git Status:\n")
		observations.Write(output)
		observations.WriteString("\n")
	}

	// Build errors
	buildOutput, buildErr := a.runBuild(ctx, workspacePath)
	if buildErr != nil {
		observations.WriteString("Build Errors:\n")
		observations.WriteString(buildOutput)
		observations.WriteString("\n")
	}

	// Test failures
	testOutput, testErr := a.runTests(ctx, workspacePath)
	if testErr != nil {
		observations.WriteString("Test Failures:\n")
		observations.WriteString(testOutput)
		observations.WriteString("\n")
	}

	// If no observations, return empty (first iteration)
	if observations.Len() == 0 {
		return "", nil
	}

	return observations.String(), nil
}

// runBuild executes go build and captures output.
//
// MENTAL MODEL:
//   go build ./... compiles all packages
//   Returns: (output, error)
//     - error != nil: build failed (compile errors)
//     - error == nil: build succeeded
//
// CROSS-QUESTIONS:
//   Q: Why go build ./...?
//   A: Builds all packages in workspace (not just main)
//      Catches compile errors in all files
//
//   Q: Why capture output?
//   A: Output contains error messages for Aider
//      Aider needs to see errors to fix them
func (a *AiderRunner) runBuild(
	ctx context.Context,
	workspacePath string,
) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runTests executes go test and captures output.
//
// MENTAL MODEL:
//   go test ./... runs all tests
//   Returns: (output, error)
//     - error != nil: tests failed
//     - error == nil: tests passed
//
// CROSS-QUESTIONS:
//   Q: Why go test ./...?
//   A: Runs all tests in workspace
//      Catches test failures in all packages
//
//   Q: Why capture output?
//   A: Output contains test failure details
//      Aider needs to see failures to fix them
func (a *AiderRunner) runTests(
	ctx context.Context,
	workspacePath string,
) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runAiderIteration executes one Aider iteration.
//
// MENTAL MODEL:
//   Aider CLI call with task + observations
//   Aider generates patches, applies them, commits
//   Returns: (commitSHA, taskComplete, error)
//
// CROSS-QUESTIONS:
//   Q: Why --yes flag?
//   A: Auto-accept all changes (no interactive mode)
//      We trust Aider's decisions (can rollback via git)
//
//   Q: How detect TASK_COMPLETE?
//   A: Check Aider output for "TASK_COMPLETE" marker
//      Or: build + tests pass (heuristic)
//
//   Q: What if Aider fails?
//   A: Return error, caller retries or escalates
func (a *AiderRunner) runAiderIteration(
	ctx context.Context,
	req AiderRunRequest,
	workspacePath string,
	observations string,
	iteration int,
) (commitSHA string, taskComplete bool, err error) {
	// Build Aider message
	message := req.TaskDescription
	if observations != "" {
		message += "\n\nObservations from previous iteration:\n" + observations
	}

	// Add expert training (Gate 1 only)
	training, err := a.loadExpertTraining(ctx, req.Expert.ID)
	if err != nil {
		a.logger.Warn("failed to load expert training",
			zap.Error(err),
			zap.String("expert_id", req.Expert.ID.String()),
		)
		// Non-fatal: continue without training
	} else if training != "" {
		message += "\n\nExpert Training (Gate 1 - Domain Patterns):\n" + training
	}

	// Run Aider CLI
	// aider --yes --message "<message>" *.go
	cmd := exec.CommandContext(ctx, "aider", "--yes", "--message", message)
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		a.logger.Error("aider failed",
			zap.Error(err),
			zap.String("output", string(output)),
		)
		return "", false, fmt.Errorf("aider: %w (output: %s)", err, string(output))
	}

	a.logger.Debug("aider output",
		zap.String("output", string(output)),
	)

	// Extract commit SHA
	commitSHA, err = a.extractCommitSHA(ctx, workspacePath)
	if err != nil {
		return "", false, fmt.Errorf("extract commit SHA: %w", err)
	}

	// Check if task complete
	// Heuristic: build + tests pass
	_, buildErr := a.runBuild(ctx, workspacePath)
	_, testErr := a.runTests(ctx, workspacePath)
	taskComplete = (buildErr == nil && testErr == nil)

	a.logger.Info("aider iteration completed",
		zap.Int("iteration", iteration),
		zap.String("commit", commitSHA),
		zap.Bool("task_complete", taskComplete),
	)

	return commitSHA, taskComplete, nil
}

// extractCommitSHA gets the latest commit SHA.
//
// MENTAL MODEL:
//   git rev-parse HEAD returns current commit SHA
//   Used to track which commits Aider created
func (a *AiderRunner) extractCommitSHA(
	ctx context.Context,
	workspacePath string,
) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// loadExpertTraining loads expert's training from DB (Gate 1 only).
//
// MENTAL MODEL:
//   Gate 1: Domain-specific patterns (auth, caching, schema design)
//   Gate 2: Peer review (not used in implementation phase)
//   Gate 3: Approval (not used in implementation phase)
//
// CROSS-QUESTIONS:
//   Q: Why Gate 1 only?
//   A: Implementation phase doesn't need peer review/approval
//      Gate 1 provides domain patterns (sufficient for coding)
//
//   Q: What if training is empty?
//   A: Non-fatal, Aider works without training (less optimal)
//      Training improves quality but isn't required
//
//   Q: How is training structured?
//   A: Free-form text with patterns, examples, anti-patterns
//      Expert admin defines training content
//
// EXAMPLE TRAINING (Backend Expert):
//   Authentication Patterns:
//   - Use bcrypt for password hashing (cost 12)
//   - JWT tokens with 15-minute expiry
//   - Refresh tokens in HTTP-only cookies
//
//   Anti-patterns:
//   - Never store passwords in plain text
//   - Never use MD5/SHA1 for passwords
func (a *AiderRunner) loadExpertTraining(
	ctx context.Context,
	expertID uuid.UUID,
) (string, error) {
	var training string
	err := a.db.QueryRow(ctx,
		`SELECT COALESCE(reasoning_charter, '') FROM experts WHERE id = $1`,
		expertID,
	).Scan(&training)
	if err != nil {
		return "", fmt.Errorf("query expert training: %w", err)
	}
	return training, nil
}

// publishCodeArtifacts walks the workspace and posts all code files to blackboard.
//
// MENTAL MODEL:
//   Input: workspace path, expert ID, commit SHAs
//   Actions:
//     1. Walk workspace recursively
//     2. Skip: .git/, design artifacts (ARCHITECTURE.md, etc.)
//     3. For each code file:
//        - Read content
//        - Detect language from extension
//        - Count lines of code
//        - Post code_artifact_produced event to blackboard
//   Output: N events posted (one per code file)
//
// CROSS-QUESTIONS:
//   Q: Why walk recursively?
//   A: Aider may create subdirectories (internal/auth/handler.go)
//
//   Q: Why skip design artifacts?
//   A: Already in blackboard from design phase, not code
//
//   Q: Why detect language?
//   A: Frontend syntax highlighting, validation
//
//   Q: What if no code files?
//   A: Not an error, just post no events (task may have failed)
//
// EXAMPLE:
//   Workspace: /workspaces/abc-123/backend-expert-id/
//   Files:
//     - ARCHITECTURE.md (skip)
//     - shortener.go (post)
//     - shortener_test.go (post)
//     - internal/db/schema.sql (post)
//   Result: 3 events posted
func (a *AiderRunner) publishCodeArtifacts(
	ctx context.Context,
	req AiderRunRequest,
	workspacePath string,
	commitSHAs []string,
) error {
	if len(commitSHAs) == 0 {
		// No commits = no code generated
		a.logger.Info("no commits to publish",
			zap.String("workflow_id", req.WorkflowID.String()),
			zap.String("expert", req.Expert.Name),
		)
		return nil
	}

	// STEP 1: Validate code before publishing
	//
	// MENTAL MODEL:
	//   Only publish code that compiles and passes tests
	//   Broken code should not appear in Kanban as "done"
	//
	// CROSS-QUESTIONS:
	//   Q: Why validate here? Aider already checked build/tests.
	//   A: Aider may have exited after max iterations (5) with broken code.
	//      This is the final gate before publishing.
	//
	//   Q: Should validation be optional?
	//   A: Future enhancement. For now, always validate.
	//      Some projects have flaky tests, may want to skip.
	//
	//   Q: What if validation fails?
	//   A: Return error, don't publish. Task marked as failed.
	//      Expert needs to investigate why code is broken.
	//
	// EXAMPLE:
	//   Scenario 1: Valid code
	//     runBuild() → error == nil ✅
	//     runTests() → error == nil ✅
	//     Continue to publish ✅
	//
	//   Scenario 2: Broken code
	//     runBuild() → error != nil ❌
	//     Return error, don't publish ✅
	//     Task marked as failed ✅

	a.logger.Info("validating code before publishing",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
	)

	// Track validation start time for metrics
	validationStart := time.Now()

	// Create timeout context for validation
	// MENTAL MODEL:
	//   Build/tests may hang forever (infinite loop, deadlock)
	//   Timeout prevents blocking workflow indefinitely
	//   15 minutes is generous (most builds < 5 min, tests < 10 min)
	//
	// CROSS-QUESTION:
	//   Q: Why 15 minutes?
	//   A: Covers slow builds (large codebases) + slow tests (integration tests)
	//      If validation takes > 15 min, something is wrong
	//
	//   Q: Should timeout be configurable?
	//   A: Future enhancement. For now, fixed 15 minutes.
	//
	//   Q: What happens on timeout?
	//   A: Return error, don't publish. Task marked as failed.
	validationCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	// Validate build
	buildOutput, buildErr := a.runBuild(validationCtx, workspacePath)
	if buildErr != nil {
		// Check if timeout
		if validationCtx.Err() == context.DeadlineExceeded {
			a.logger.Error("build validation timeout",
				zap.String("workflow_id", req.WorkflowID.String()),
				zap.String("expert", req.Expert.Name),
				zap.Duration("timeout", 15*time.Minute),
			)
			return fmt.Errorf("build validation timeout (15 minutes)")
		}
	if buildErr != nil {
		a.logger.Error("build validation failed",
			zap.String("workflow_id", req.WorkflowID.String()),
			zap.String("expert", req.Expert.Name),
			zap.String("output", buildOutput),
			zap.Error(buildErr),
		)
		return fmt.Errorf("build validation failed: %w\nOutput: %s", buildErr, buildOutput)
	}

	// Validate tests
	testOutput, testErr := a.runTests(validationCtx, workspacePath)
	if testErr != nil {
		// Check if timeout
		if validationCtx.Err() == context.DeadlineExceeded {
			a.logger.Error("test validation timeout",
				zap.String("workflow_id", req.WorkflowID.String()),
				zap.String("expert", req.Expert.Name),
				zap.Duration("timeout", 15*time.Minute),
			)
			return fmt.Errorf("test validation timeout (15 minutes)")
		}
		a.logger.Error("test validation failed",
			zap.String("workflow_id", req.WorkflowID.String()),
			zap.String("expert", req.Expert.Name),
			zap.String("output", testOutput),
			zap.Error(testErr),
		)
		return fmt.Errorf("test validation failed: %w\nOutput: %s", testErr, testOutput)
	}

	a.logger.Info("validation passed, proceeding to publish",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
	)

	// Use latest commit SHA for all artifacts
	latestCommitSHA := commitSHAs[len(commitSHAs)-1]

	publishedCount := 0
	err := filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip .git/ directory
		if strings.Contains(path, ".git/") || strings.Contains(path, ".git\\") {
			return nil
		}

		// Skip design artifacts (already in blackboard from design phase)
		filename := filepath.Base(path)
		if filename == "ARCHITECTURE.md" ||
			filename == "DATA_MODEL.md" ||
			filename == "API_CONTRACTS.md" ||
			filename == ".gitignore" ||
			filename == "README.md" {
			return nil
		}

		// Only publish code files
		language := detectLanguage(path)
		if language == "" {
			// Unknown file type, skip
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			a.logger.Warn("failed to read file",
				zap.String("path", path),
				zap.Error(err),
			)
			return nil // Skip this file, continue walking
		}

		// Get relative path (remove workspace prefix)
		relPath, err := filepath.Rel(workspacePath, path)
		if err != nil {
			relPath = filename // Fallback to just filename
		}

		// Count lines of code
		linesOfCode := countLines(content)

		// Post to blackboard
		event := blackboard.Event{
			Type:       "code_artifact_produced",
			WorkflowID: req.WorkflowID,
			ExpertID:   req.Expert.ID,
			Data: map[string]interface{}{
				"filename":      filename,
				"file_path":     relPath,
				"content":       string(content),
				"language":      language,
				"commit_sha":    latestCommitSHA,
				"lines_of_code": linesOfCode,
				"phase":         req.WorkflowPhase,
			},
		}

		if err := a.bbStore.Post(ctx, event); err != nil {
			a.logger.Error("failed to post code artifact",
				zap.String("file", relPath),
				zap.Error(err),
			)
			return nil // Continue walking even if one post fails
		}

		publishedCount++
		a.logger.Info("published code artifact",
			zap.String("file", relPath),
			zap.String("language", language),
			zap.Int("lines", linesOfCode),
		)

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk workspace: %w", err)
	}

	a.logger.Info("code artifacts published",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
		zap.Int("count", publishedCount),
	)

	return nil
}

// detectLanguage returns the language name based on file extension.
// Returns empty string for unknown/non-code files.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".sql":
		return "sql"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	case ".rs":
		return "rust"
	case ".sh", ".bash":
		return "shell"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss", ".sass":
		return "scss"
	default:
		return "" // Unknown file type
	}
}

// countLines counts the number of lines in a file.
func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := 1 // Start at 1 (file with no newlines = 1 line)
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	return lines
}
