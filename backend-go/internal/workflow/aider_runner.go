package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

// AiderCheckpoint represents saved state for resuming after pod restart.
//
// PHASE 5: Production Hardening - Error Recovery
//
// MENTAL MODEL:
//   Pod crashes during iteration 3 of 5
//   On restart:
//     1. Load checkpoint from DB
//     2. Resume from iteration 4
//     3. Continue OTA loop
//
// CROSS-QUESTIONS:
//   Q: What state needs to be saved?
//   A: Iteration number, commits, observations, completion status
//
//   Q: Where to save?
//   A: Database table: aider_checkpoints
//      Key: (workflow_id, expert_id, task_id)
//
//   Q: When to save?
//   A: After each successful iteration (before next iteration starts)
//
//   Q: How to resume?
//   A: Load checkpoint, skip completed iterations, continue from last+1
//
//   Q: What if no checkpoint?
//   A: Start from iteration 1 (normal flow, first run)
//
//   Q: When to delete checkpoint?
//   A: After task completes successfully (cleanup)
//
// EXAMPLE:
//   Iteration 1: Create auth.go → Save checkpoint (iteration=1, commits=[abc123])
//   Iteration 2: Fix import → Save checkpoint (iteration=2, commits=[abc123, def456])
//   [POD CRASH]
//   On restart: Load checkpoint → Resume from iteration 3
//   Iteration 3: Add tests → Save checkpoint (iteration=3, commits=[abc123, def456, ghi789])
//   Task complete → Delete checkpoint
type AiderCheckpoint struct {
	WorkflowID       uuid.UUID `json:"workflow_id"`
	ExpertID         uuid.UUID `json:"expert_id"`
	TaskID           uuid.UUID `json:"task_id"`
	CurrentIteration int       `json:"current_iteration"` // Last completed iteration
	CommitSHAs       []string  `json:"commit_shas"`       // All commits produced so far
	LastObservation  string    `json:"last_observation"`  // Observations from last iteration
	Completed        bool      `json:"completed"`         // True if task completed
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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

	// Step 5: Post coverage metrics (QA phase only) - Phase 4
	if req.WorkflowPhase == "qa" && result.Completed {
		if err := a.postCoverageMetrics(ctx, req, workspacePath); err != nil {
			// Non-fatal: log error but don't fail the task
			a.logger.Error("failed to post coverage metrics",
				zap.String("workflow_id", req.WorkflowID.String()),
				zap.String("expert", req.Expert.Name),
				zap.Error(err),
			)
		}
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
// PHASE 4 ADDITION:
//   QA Phase Support:
//     - Different prompt (focus on test generation)
//     - Load implementation artifacts from previous phase
//     - Track test coverage (>80% target)
//     - Validate test quality (naming, edge cases, error paths)
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
//
//   Q: How does QA phase differ?
//   A: Implementation creates source code, QA creates test files
//      QA validates coverage and test quality
func (a *AiderRunner) runAiderLoop(
	ctx context.Context,
	req AiderRunRequest,
	workspacePath string,
) (*AiderRunResult, error) {
	const maxIterations = 5
	var commitSHAs []string

	// PHASE 5: Load checkpoint for recovery
	// Try to resume from previous run (pod restart)
	checkpoint, err := a.loadCheckpoint(ctx, req.WorkflowID, req.Expert.ID, req.TaskID)
	if err != nil {
		// Non-fatal: log warning and start from beginning
		a.logger.Warn("failed to load checkpoint, starting from beginning",
			zap.Error(err),
			zap.String("workflow_id", req.WorkflowID.String()),
		)
		checkpoint = nil
	}

	// Resume from checkpoint if exists
	startIteration := 1
	if checkpoint != nil {
		// Resume from next iteration
		startIteration = checkpoint.CurrentIteration + 1
		commitSHAs = checkpoint.CommitSHAs

		a.logger.Info("resuming from checkpoint",
			zap.Int("start_iteration", startIteration),
			zap.Int("commits_so_far", len(commitSHAs)),
		)
	}

	for iteration := startIteration; iteration <= maxIterations; iteration++ {
		a.logger.Info("aider iteration starting",
			zap.Int("iteration", iteration),
			zap.Int("max", maxIterations),
		)

		// Step 1: Observe workspace state (phase-aware)
		observations, err := a.observeWorkspace(ctx, req.WorkflowPhase, workspacePath)
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

		// PHASE 5: Save checkpoint after successful iteration
		// This allows resuming from this point if pod crashes
		checkpointToSave := &AiderCheckpoint{
			WorkflowID:       req.WorkflowID,
			ExpertID:         req.Expert.ID,
			TaskID:           req.TaskID,
			CurrentIteration: iteration,
			CommitSHAs:       commitSHAs,
			LastObservation:  observations,
			Completed:        taskComplete,
		}
		if err := a.saveCheckpoint(ctx, checkpointToSave); err != nil {
			// Non-fatal: log error but continue
			// Worst case: restart from beginning on pod crash
			a.logger.Error("failed to save checkpoint",
				zap.Error(err),
				zap.Int("iteration", iteration),
			)
		}

		// Check if task complete
		if taskComplete {
			a.logger.Info("task completed",
				zap.Int("iterations", iteration),
			)

			// PHASE 5: Delete checkpoint on successful completion
			if err := a.deleteCheckpoint(ctx, req.WorkflowID, req.Expert.ID, req.TaskID); err != nil {
				// Non-fatal: log error but don't fail task
				// TTL job will cleanup old checkpoints
				a.logger.Error("failed to delete checkpoint",
					zap.Error(err),
					zap.String("workflow_id", req.WorkflowID.String()),
				)
			}

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
//     - Coverage (QA phase only)
//
// PHASE 4 ADDITION:
//   QA Phase: Include test coverage in observations
//   Example: "Coverage: 45.2% (target: 80.0%)"
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
//
//   Q: Why coverage in QA only?
//   A: Implementation doesn't need coverage (no tests yet)
//      QA needs to know current coverage to improve it
func (a *AiderRunner) observeWorkspace(
	ctx context.Context,
	phase string,
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

	// Test failures and coverage (phase-specific)
	if phase == "qa" {
		// QA phase: run tests with coverage
		testOutput, coverage, testErr := a.runTestsWithCoverage(ctx, workspacePath)
		if testErr != nil {
			observations.WriteString("Test Failures:\n")
			observations.WriteString(testOutput)
			observations.WriteString("\n")
		}
		// Always show coverage (even if 0%)
		observations.WriteString(fmt.Sprintf("Current Coverage: %.1f%% (target: 80.0%%)\n\n", coverage))
	} else {
		// Implementation phase: run tests without coverage
		testOutput, testErr := a.runTests(ctx, workspacePath)
		if testErr != nil {
			observations.WriteString("Test Failures:\n")
			observations.WriteString(testOutput)
			observations.WriteString("\n")
		}
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

// runTestsWithCoverage executes go test with coverage and captures output.
//
// PHASE 4: QA Phase Support
//
// MENTAL MODEL:
//   go test -cover ./... outputs:
//     ok      package1    0.123s  coverage: 85.7% of statements
//     ok      package2    0.456s  coverage: 92.3% of statements
//
//   Parse output to extract coverage percentage
//
// CROSS-QUESTIONS:
//   Q: Why separate method?
//   A: runTests() doesn't capture coverage, need -cover flag
//
//   Q: What if no tests?
//   A: Coverage = 0%, not an error (first iteration)
//
//   Q: What if multiple packages?
//   A: Average coverage across all packages
//
// RETURNS:
//   output: Full test output (for debugging)
//   coverage: Average coverage percentage (0-100)
//   error: Non-nil if tests failed to run (not if tests failed)
func (a *AiderRunner) runTestsWithCoverage(
	ctx context.Context,
	workspacePath string,
) (output string, coverage float64, err error) {
	cmd := exec.CommandContext(ctx, "go", "test", "-cover", "./...")
	cmd.Dir = workspacePath
	outputBytes, cmdErr := cmd.CombinedOutput()
	output = string(outputBytes)

	// Parse coverage from output
	coverage = a.parseCoverage(output)

	// Return command error (if any)
	return output, coverage, cmdErr
}

// parseCoverage extracts average coverage percentage from go test output.
//
// PHASE 4: QA Phase Support
//
// MENTAL MODEL:
//   Input:
//     ok      package1    0.123s  coverage: 85.7% of statements
//     ok      package2    0.456s  coverage: 92.3% of statements
//   Output: 89.0 (average of 85.7 and 92.3)
//
// CROSS-QUESTIONS:
//   Q: Why average?
//   A: Single metric for overall coverage
//      Alternative: weighted by package size (future)
//
//   Q: What if no coverage lines?
//   A: Return 0.0 (no tests yet)
//
//   Q: What if parse fails?
//   A: Return 0.0 (graceful degradation)
//
// REGEX:
//   coverage: (\d+\.\d+)% of statements
//   Captures: 85.7, 92.3, etc.
func (a *AiderRunner) parseCoverage(output string) float64 {
	// Regex to match: coverage: 85.7% of statements
	// Note: Using simple string parsing instead of regex for clarity
	lines := strings.Split(output, "\n")
	var totalCoverage float64
	var count int

	for _, line := range lines {
		// Look for "coverage: XX.X% of statements"
		if strings.Contains(line, "coverage:") && strings.Contains(line, "% of statements") {
			// Extract percentage
			// Example: "ok  	package	0.123s	coverage: 85.7% of statements"
			parts := strings.Split(line, "coverage:")
			if len(parts) < 2 {
				continue
			}
			// parts[1] = " 85.7% of statements"
			percentPart := strings.TrimSpace(parts[1])
			// percentPart = "85.7% of statements"
			percentStr := strings.Split(percentPart, "%")[0]
			// percentStr = "85.7"

			// Parse float
			var percent float64
			if _, err := fmt.Sscanf(percentStr, "%f", &percent); err == nil {
				totalCoverage += percent
				count++
			}
		}
	}

	// Return average
	if count == 0 {
		return 0.0
	}
	return totalCoverage / float64(count)
}

// runAiderIteration executes one Aider iteration.
//
// MENTAL MODEL:
//   Aider CLI call with task + observations
//   Aider generates patches, applies them, commits
//   Returns: (commitSHA, taskComplete, error)
//
// PHASE 4 ADDITION:
//   QA Phase:
//     - Different prompt (focus on test generation)
//     - Check coverage in completion criteria
//     - Validate test quality
//
// CROSS-QUESTIONS:
//   Q: Why --yes flag?
//   A: Auto-accept all changes (no interactive mode)
//      We trust Aider's decisions (can rollback via git)
//
//   Q: How detect TASK_COMPLETE?
//   A: Implementation: build + tests pass
//      QA: tests pass + coverage >80%
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
	// Build Aider message (phase-specific)
	var message string
	if req.WorkflowPhase == "qa" {
		message = a.buildQAPrompt(req.TaskDescription, observations)
	} else {
		// Implementation phase
		message = req.TaskDescription
		if observations != "" {
			message += "\n\nObservations from previous iteration:\n" + observations
		}
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

	// Check if task complete (phase-specific criteria)
	if req.WorkflowPhase == "qa" {
		// QA phase: tests pass + coverage >80%
		testOutput, coverage, testErr := a.runTestsWithCoverage(ctx, workspacePath)
		taskComplete = (testErr == nil && coverage >= 80.0)

		a.logger.Info("aider iteration completed (QA phase)",
			zap.Int("iteration", iteration),
			zap.String("commit", commitSHA),
			zap.Float64("coverage", coverage),
			zap.Bool("tests_pass", testErr == nil),
			zap.Bool("task_complete", taskComplete),
		)

		// Log warning if coverage is close but not enough
		if testErr == nil && coverage >= 75.0 && coverage < 80.0 {
			a.logger.Warn("coverage close to target but not sufficient",
				zap.Float64("coverage", coverage),
				zap.Float64("target", 80.0),
				zap.String("test_output", testOutput),
			)
		}
	} else {
		// Implementation phase: build + tests pass
		_, buildErr := a.runBuild(ctx, workspacePath)
		_, testErr := a.runTests(ctx, workspacePath)
		taskComplete = (buildErr == nil && testErr == nil)

		a.logger.Info("aider iteration completed (implementation phase)",
			zap.Int("iteration", iteration),
			zap.String("commit", commitSHA),
			zap.Bool("build_pass", buildErr == nil),
			zap.Bool("tests_pass", testErr == nil),
			zap.Bool("task_complete", taskComplete),
		)
	}

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

// buildQAPrompt constructs a QA-specific prompt for test generation.
//
// PHASE 4: QA Phase Support
//
// MENTAL MODEL:
//   QA prompt structure:
//     1. Task description (from planner)
//     2. Test requirements (coverage, quality)
//     3. Observations (test failures, coverage)
//     4. Best practices (naming, table-driven tests)
//
// CROSS-QUESTIONS:
//   Q: What makes a good test?
//   A: - Tests all public functions
//      - Covers edge cases (empty input, nil, invalid)
//      - Tests error paths
//      - Clear naming: TestFunctionName_Scenario_ExpectedResult
//
//   Q: Why >80% coverage?
//   A: Industry standard, balances thoroughness vs cost
//
//   Q: Should we enforce table-driven tests?
//   A: Recommend in prompt, not enforce (some tests don't fit)
//
// EXAMPLE OUTPUT:
//   "Generate comprehensive tests for auth.go
//
//    Requirements:
//    - Test all public functions (Authenticate, ValidateToken)
//    - Cover edge cases: empty username, invalid password, expired token
//    - Test error paths: database errors, network failures
//    - Achieve >80% code coverage
//    - Use table-driven tests where appropriate
//
//    Naming convention:
//    - TestFunctionName_Scenario_ExpectedResult
//    - Example: TestAuthenticate_EmptyUsername_ReturnsError
//
//    Observations from previous iteration:
//    Test Failures:
//    TestAuthenticate_ValidCredentials failed: expected nil error, got 'db connection failed'
//    Coverage: 45.2%"
func (a *AiderRunner) buildQAPrompt(taskDescription string, observations string) string {
	var prompt strings.Builder

	// Task description
	prompt.WriteString(taskDescription)
	prompt.WriteString("\n\n")

	// Test requirements
	prompt.WriteString("Test Requirements:\n")
	prompt.WriteString("- Test all public functions and methods\n")
	prompt.WriteString("- Cover edge cases: empty input, nil values, invalid data, boundary conditions\n")
	prompt.WriteString("- Test error paths: database errors, network failures, validation errors\n")
	prompt.WriteString("- Achieve >80% code coverage\n")
	prompt.WriteString("- Use table-driven tests where appropriate\n")
	prompt.WriteString("\n")

	// Naming convention
	prompt.WriteString("Naming Convention:\n")
	prompt.WriteString("- TestFunctionName_Scenario_ExpectedResult\n")
	prompt.WriteString("- Example: TestAuthenticate_EmptyUsername_ReturnsError\n")
	prompt.WriteString("- Example: TestValidateToken_ExpiredToken_ReturnsFalse\n")
	prompt.WriteString("\n")

	// Best practices
	prompt.WriteString("Best Practices:\n")
	prompt.WriteString("- Use t.Run() for subtests\n")
	prompt.WriteString("- Use testify/assert for assertions (if available)\n")
	prompt.WriteString("- Mock external dependencies (database, HTTP clients)\n")
	prompt.WriteString("- Clean up resources in defer statements\n")
	prompt.WriteString("\n")

	// Observations (if any)
	if observations != "" {
		prompt.WriteString("Observations from previous iteration:\n")
		prompt.WriteString(observations)
		prompt.WriteString("\n")
	}

	return prompt.String()
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

	// Log validation success with duration
	validationDuration := time.Since(validationStart)
	a.logger.Info("validation passed, proceeding to publish",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("expert", req.Expert.Name),
		zap.Duration("validation_duration", validationDuration),
	)

	// Warn if validation took too long (> 10 minutes)
	if validationDuration > 10*time.Minute {
		a.logger.Warn("validation took longer than expected",
			zap.String("workflow_id", req.WorkflowID.String()),
			zap.String("expert", req.Expert.Name),
			zap.Duration("duration", validationDuration),
			zap.String("recommendation", "consider optimizing build/test performance"),
		)
	}

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

// postCoverageMetrics posts test coverage metrics to blackboard.
//
// PHASE 4: QA Phase Support
//
// MENTAL MODEL:
//   After QA phase completes, post coverage metrics for:
//     - Frontend display (show coverage badge)
//     - Audit trail (compliance requirements)
//     - Other experts (know test quality)
//
// CROSS-QUESTIONS:
//   Q: When to call this?
//   A: After QA phase completes successfully (result.Completed == true)
//
//   Q: What if coverage < 80%?
//   A: This shouldn't happen (task wouldn't complete)
//      But if called, still post (passed: false)
//
//   Q: How to count tests?
//   A: Parse test output for "PASS" or "ok" lines
//
//   Q: What about package-level coverage?
//   A: Future enhancement. For now, post average.
//
// EVENT STRUCTURE:
//   {
//     "event_type": "test_coverage_achieved",
//     "posted_by_expert_id": "expert-uuid",
//     "content": {
//       "coverage_percent": 85.7,
//       "target_percent": 80.0,
//       "passed": true,
//       "test_count": 42,
//       "package_count": 3
//     }
//   }
func (a *AiderRunner) postCoverageMetrics(
	ctx context.Context,
	req AiderRunRequest,
	workspacePath string,
) error {
	// Run tests with coverage
	testOutput, coverage, testErr := a.runTestsWithCoverage(ctx, workspacePath)

	// Count tests and packages
	testCount := a.countTests(testOutput)
	packageCount := a.countPackages(testOutput)

	// Determine if coverage target met
	const targetCoverage = 80.0
	passed := (testErr == nil && coverage >= targetCoverage)

	// Post event to blackboard
	_, err := a.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:       req.WorkflowID,
		EventType:        "test_coverage_achieved",
		PostedByExpertID: &req.Expert.ID,
		Content: map[string]interface{}{
			"coverage_percent": coverage,
			"target_percent":   targetCoverage,
			"passed":           passed,
			"test_count":       testCount,
			"package_count":    packageCount,
		},
	})
	if err != nil {
		return fmt.Errorf("post coverage event: %w", err)
	}

	a.logger.Info("coverage metrics posted",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.Float64("coverage", coverage),
		zap.Int("test_count", testCount),
		zap.Int("package_count", packageCount),
		zap.Bool("passed", passed),
	)

	return nil
}

// countTests counts the number of tests from test output.
//
// PHASE 4: QA Phase Support
//
// MENTAL MODEL:
//   Test output contains lines like:
//     ok      package1    0.123s
//     ok      package2    0.456s
//   Each "ok" line represents one or more tests in that package.
//
//   More accurate: count "PASS: TestName" lines
//   But "ok" lines are simpler and good enough.
//
// CROSS-QUESTIONS:
//   Q: Why count "ok" lines?
//   A: Simple heuristic, one line per package with tests
//
//   Q: What about individual test count?
//   A: Would need to parse "=== RUN TestName" lines
//      Future enhancement if needed
//
//   Q: What if no tests?
//   A: Return 0 (valid case for first iteration)
func (a *AiderRunner) countTests(output string) int {
	lines := strings.Split(output, "\n")
	count := 0
	for _, line := range lines {
		// Count lines starting with "ok" (package test summary)
		if strings.HasPrefix(strings.TrimSpace(line), "ok") {
			count++
		}
	}
	return count
}

// countPackages counts the number of packages tested.
//
// PHASE 4: QA Phase Support
//
// MENTAL MODEL:
//   Same as countTests() - each "ok" line is one package
//   This is a duplicate for clarity (may diverge in future)
//
// CROSS-QUESTIONS:
//   Q: Why separate method?
//   A: Semantic clarity, may count differently in future
//
//   Q: What about failed packages?
//   A: Count "FAIL" lines too
func (a *AiderRunner) countPackages(output string) int {
	lines := strings.Split(output, "\n")
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Count "ok" and "FAIL" lines (package summaries)
		if strings.HasPrefix(trimmed, "ok") || strings.HasPrefix(trimmed, "FAIL") {
			count++
		}
	}
	return count
}

// saveCheckpoint saves current iteration state to DB for recovery.
//
// PHASE 5: Production Hardening - Error Recovery
//
// MENTAL MODEL:
//   After each successful iteration, save state:
//     INSERT INTO aider_checkpoints (...)
//     ON CONFLICT (workflow_id, expert_id, task_id)
//     DO UPDATE SET current_iteration = ..., updated_at = NOW()
//
// CROSS-QUESTIONS:
//   Q: What if save fails?
//   A: Log error but don't fail iteration (non-fatal)
//      Worst case: restart from beginning (acceptable)
//
//   Q: Should we use transaction?
//   A: No, checkpoint is independent of main workflow
//      Failure to save checkpoint shouldn't rollback iteration
//
//   Q: What about concurrent saves?
//   A: ON CONFLICT handles race conditions
//      Last write wins (acceptable for checkpoints)
//
// SQL:
//   CREATE TABLE aider_checkpoints (
//     workflow_id UUID NOT NULL,
//     expert_id UUID NOT NULL,
//     task_id UUID NOT NULL,
//     current_iteration INT NOT NULL,
//     commit_shas JSONB NOT NULL,
//     last_observation TEXT,
//     completed BOOLEAN DEFAULT FALSE,
//     created_at TIMESTAMPTZ DEFAULT NOW(),
//     updated_at TIMESTAMPTZ DEFAULT NOW(),
//     PRIMARY KEY (workflow_id, expert_id, task_id)
//   );
func (a *AiderRunner) saveCheckpoint(
	ctx context.Context,
	checkpoint *AiderCheckpoint,
) error {
	// Convert commit SHAs to JSONB
	commitSHAsJSON, err := json.Marshal(checkpoint.CommitSHAs)
	if err != nil {
		return fmt.Errorf("marshal commit_shas: %w", err)
	}

	// Upsert checkpoint
	query := `
		INSERT INTO aider_checkpoints (
			workflow_id,
			expert_id,
			task_id,
			current_iteration,
			commit_shas,
			last_observation,
			completed,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (workflow_id, expert_id, task_id)
		DO UPDATE SET
			current_iteration = EXCLUDED.current_iteration,
			commit_shas = EXCLUDED.commit_shas,
			last_observation = EXCLUDED.last_observation,
			completed = EXCLUDED.completed,
			updated_at = NOW()
	`

	_, err = a.db.Exec(ctx, query,
		checkpoint.WorkflowID,
		checkpoint.ExpertID,
		checkpoint.TaskID,
		checkpoint.CurrentIteration,
		commitSHAsJSON,
		checkpoint.LastObservation,
		checkpoint.Completed,
	)
	if err != nil {
		return fmt.Errorf("upsert checkpoint: %w", err)
	}

	a.logger.Debug("checkpoint saved",
		zap.String("workflow_id", checkpoint.WorkflowID.String()),
		zap.String("expert_id", checkpoint.ExpertID.String()),
		zap.Int("iteration", checkpoint.CurrentIteration),
	)

	return nil
}

// loadCheckpoint loads saved state from DB for recovery.
//
// PHASE 5: Production Hardening - Error Recovery
//
// MENTAL MODEL:
//   On pod restart, check if checkpoint exists:
//     SELECT * FROM aider_checkpoints
//     WHERE workflow_id = ? AND expert_id = ? AND task_id = ?
//
//   If exists: Resume from checkpoint.CurrentIteration + 1
//   If not exists: Start from iteration 1 (normal flow)
//
// CROSS-QUESTIONS:
//   Q: What if checkpoint doesn't exist?
//   A: Return nil, nil (not an error, just no checkpoint)
//
//   Q: What if query fails?
//   A: Return error (caller decides: fail or start fresh)
//
//   Q: Should we validate checkpoint?
//   A: Yes, check if completed=true (shouldn't happen)
//      If completed, delete checkpoint and start fresh
func (a *AiderRunner) loadCheckpoint(
	ctx context.Context,
	workflowID, expertID, taskID uuid.UUID,
) (*AiderCheckpoint, error) {
	query := `
		SELECT
			workflow_id,
			expert_id,
			task_id,
			current_iteration,
			commit_shas,
			last_observation,
			completed,
			created_at,
			updated_at
		FROM aider_checkpoints
		WHERE workflow_id = $1
		  AND expert_id = $2
		  AND task_id = $3
	`

	var checkpoint AiderCheckpoint
	var commitSHAsJSON []byte

	err := a.db.QueryRow(ctx, query, workflowID, expertID, taskID).Scan(
		&checkpoint.WorkflowID,
		&checkpoint.ExpertID,
		&checkpoint.TaskID,
		&checkpoint.CurrentIteration,
		&commitSHAsJSON,
		&checkpoint.LastObservation,
		&checkpoint.Completed,
		&checkpoint.CreatedAt,
		&checkpoint.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			// No checkpoint exists (normal case for first run)
			return nil, nil
		}
		return nil, fmt.Errorf("query checkpoint: %w", err)
	}

	// Unmarshal commit SHAs
	if err := json.Unmarshal(commitSHAsJSON, &checkpoint.CommitSHAs); err != nil {
		return nil, fmt.Errorf("unmarshal commit_shas: %w", err)
	}

	// Validate checkpoint
	if checkpoint.Completed {
		// Checkpoint marked as completed but still exists
		// This shouldn't happen, but if it does, delete it
		a.logger.Warn("found completed checkpoint, deleting",
			zap.String("workflow_id", workflowID.String()),
			zap.String("expert_id", expertID.String()),
		)
		if err := a.deleteCheckpoint(ctx, workflowID, expertID, taskID); err != nil {
			a.logger.Error("failed to delete completed checkpoint", zap.Error(err))
		}
		return nil, nil
	}

	a.logger.Info("checkpoint loaded",
		zap.String("workflow_id", workflowID.String()),
		zap.String("expert_id", expertID.String()),
		zap.Int("iteration", checkpoint.CurrentIteration),
		zap.Int("commits", len(checkpoint.CommitSHAs)),
	)

	return &checkpoint, nil
}

// deleteCheckpoint removes checkpoint from DB after task completes.
//
// PHASE 5: Production Hardening - Error Recovery
//
// MENTAL MODEL:
//   After task completes successfully, cleanup:
//     DELETE FROM aider_checkpoints
//     WHERE workflow_id = ? AND expert_id = ? AND task_id = ?
//
// CROSS-QUESTIONS:
//   Q: When to delete?
//   A: After task completes (Completed=true)
//      Also after loading completed checkpoint (cleanup)
//
//   Q: What if delete fails?
//   A: Log error but don't fail task (non-fatal)
//      Checkpoint will be cleaned up by TTL job
//
//   Q: Should we delete on failure?
//   A: No, keep checkpoint for debugging
//      TTL job will clean up old checkpoints (>7 days)
func (a *AiderRunner) deleteCheckpoint(
	ctx context.Context,
	workflowID, expertID, taskID uuid.UUID,
) error {
	query := `
		DELETE FROM aider_checkpoints
		WHERE workflow_id = $1
		  AND expert_id = $2
		  AND task_id = $3
	`

	_, err := a.db.Exec(ctx, query, workflowID, expertID, taskID)
	if err != nil {
		return fmt.Errorf("delete checkpoint: %w", err)
	}

	a.logger.Debug("checkpoint deleted",
		zap.String("workflow_id", workflowID.String()),
		zap.String("expert_id", expertID.String()),
	)

	return nil
}
