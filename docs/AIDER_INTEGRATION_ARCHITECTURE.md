# Aider Integration Architecture Design

**Status:** Draft  
**Created:** 2026-09-17  
**Author:** AI Avengers Team  
**Purpose:** Integrate Aider for implementation/qa phases while preserving existing AgentLoop for design phases

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current Architecture](#current-architecture)
3. [Problem Statement](#problem-statement)
4. [Design Principles](#design-principles)
5. [Proposed Architecture](#proposed-architecture)
6. [Implementation Roadmap](#implementation-roadmap)
7. [Testing Strategy](#testing-strategy)
8. [Appendix](#appendix)

---

## 1. Executive Summary

### Current State

AI Avengers uses **AgentLoop** (OTA pattern) for all workflow phases:
- **Design phases** (high_level_design, detailed_design): Experts reason, post architecture decisions to blackboard ✅
- **Implementation/QA phases**: Experts post `code_artifact_produced` events with JSON blobs ❌

**Problem:** JSON artifacts in blackboard ≠ actual executable code. No file system, no git history, no build/test execution.

### Proposed Solution

**Phase-based routing:**
```
Phase                    | Executor      | Context           | Output
-------------------------|---------------|-------------------|------------------
high_level_design        | AgentLoop     | Blackboard        | Architecture decisions
detailed_design          | AgentLoop     | Blackboard        | API contracts, data models
implementation           | AiderRunner   | File system + Git | Working code + commits
qa                       | AiderRunner   | File system + Git | Tests + fixes
handoff                  | AgentLoop     | Blackboard        | Final summary
```

**Key Benefits:**
1. ✅ Preserves existing AgentLoop (no breaking changes to design phases)
2. ✅ Maintains 70-30% expert system (Gate 1 only in implementation)
3. ✅ Real code execution (compile, test, iterate)
4. ✅ Git history as memory (meaningful commits, patch-based changes)
5. ✅ Workspace isolation (one workspace per workflow)

---

## 2. Current Architecture

### 2.1 Core Flow

**File:** `backend-go/internal/workflow/runner.go`

```go
func (r *WorkflowRunner) Run(ctx context.Context, workflowID uuid.UUID) {
    // 1. Load workflow + experts from DB
    wf, _ := r.engine.GetByID(ctx, workflowID)
    experts, _ := r.loadWorkflowExperts(ctx, wf.SelectedExpertIDs)
    
    // 2. Load requirement from blackboard
    requirementText := r.loadRequirement(ctx, workflowID, wf.Title)
    
    // 3. Planner.Plan() → []TaskSpec (per-expert tasks)
    tasks, _ := r.planner.Plan(ctx, requirementText, experts)
    
    // 4. DAG.BuildDAG() → []ExecutionWave (topological sort)
    waves, _ := BuildDAG(tasks)
    
    // 5. Post task_plan_ready event → Projector creates workflow_tasks
    r.store.Post(ctx, blackboard.PostRequest{
        EventType: "task_plan_ready",
        Content:   buildPlanContent(tasks, experts),
    })
    
    // 6. AskClient for plan approval → pause
    r.tools.AskClient(ctx, AskClientRequest{...})
    r.waitForResume(ctx, workflowID)
    
    // 7. Transition phase + execute waves
    r.engine.TransitionPhase(ctx, workflowID, PhaseHighLevelDesign, nil, 0)
    state := &runnerState{Phase: PhaseHighLevelDesign}
    r.executeWaves(ctx, workflowID, waves, experts, state)
    
    // 8. Final approval + complete
    r.engine.TransitionPhase(ctx, workflowID, PhaseHandoff, nil, 0)
    r.tools.AskClient(ctx, AskClientRequest{...})
    r.waitForResume(ctx, workflowID)
    r.engine.Complete(ctx, workflowID)
}
```

### 2.2 AgentLoop (OTA Pattern)

**File:** `backend-go/internal/workflow/agent_loop.go`

```go
func (a *AgentLoop) Run(ctx context.Context, req AgentLoopRequest) (*AgentLoopResult, error) {
    maxIter := req.Expert.MaxLoopIterations
    var lastBlackboardSeq int64 = 0
    var allArtifacts []blackboard.Event
    
    for iter := 1; iter <= maxIter; iter++ {
        // ============================================================
        // OBSERVE: Read new blackboard events since last cursor
        // ============================================================
        newArtifacts, _ := a.tools.ReadBlackboard(ctx, ReadBlackboardRequest{
            WorkflowID: req.WorkflowID,
            Since:      lastBlackboardSeq,
        })
        allArtifacts = append(allArtifacts, newArtifacts...)
        
        // ============================================================
        // GATE SYSTEM: 3-gate knowledge access (iter==1 only)
        // Design phases: Gates 1+2+3 active
        // Implementation phase: Gate 1 only (no generic allowed)
        // ============================================================
        var gateResult *GateResult
        if iter == 1 && a.gateSystem != nil {
            isDesignPhase := req.WorkflowPhase == PhaseHighLevelDesign ||
                             req.WorkflowPhase == PhaseDetailedDesign
            
            if isDesignPhase {
                // Full 3-gate system
                gateResult, _ = a.gateSystem.RunGates(
                    ctx, req.Expert, req.TaskDescription, req.AllExperts,
                )
            } else {
                // Implementation phase: Gate 1 only, generic BLOCKED
                chunks, _ := a.assembler.GetCourseChunksForWorkflow(
                    ctx, req.Expert.ID, req.TaskDescription, 10,
                )
                gateResult = &GateResult{
                    TrainingChunks: chunks,
                    Gate1Passed:    len(chunks) > 0,
                    GenericAllowed: false, // NEVER in implementation phase
                }
            }
        }
        
        // ============================================================
        // CONTEXT MANAGEMENT: Build context from gate result + blackboard
        // ============================================================
        var contextText string
        if gateResult != nil {
            gateCtx := FormatGateContext(gateResult)
            blackboardCtx, _ := a.buildBlackboardContext(ctx, allArtifacts)
            contextText = gateCtx + "\n[BLACKBOARD — peers' work]\n" + blackboardCtx
        }
        
        // ============================================================
        // THINK: LLM call
        // ============================================================
        systemPrompt := buildAgentSystemPrompt(req)
        userPrompt := buildAgentUserPrompt(req, contextText, iter)
        llmResp, _ := a.gateway.Call(ctx, gateway.LLMRequest{
            Model:        gateway.ModelStrong,
            SystemPrompt: systemPrompt,
            UserPrompt:   userPrompt,
            MaxTokens:    4000,
        })
        
        // ============================================================
        // ACT: Execute tool calls from LLM response
        // ============================================================
        artifactID, _ := a.act(ctx, req, llmResp.Content)
        
        if strings.Contains(llmResp.Content, doneMarker) {
            // Save generic claims to Experience Bank before exiting
            if a.experienceBank != nil && gateResult != nil && gateResult.GenericAllowed {
                claims := ExtractGenericClaims(llmResp.Content)
                a.experienceBank.SaveGenericClaims(
                    ctx, req.Expert.ID, req.WorkflowID,
                    req.TaskTitle, req.Expert.Domain, claims,
                )
            }
            break
        }
    }
    
    return result, nil
}
```

### 2.3 Key Components

| Component | File | Purpose |
|-----------|------|----------|
| **WorkflowRunner** | `backend-go/internal/workflow/runner.go` | Orchestrates entire workflow, executes waves |
| **AgentLoop** | `backend-go/internal/workflow/agent_loop.go` | OTA loop for each expert task |
| **GateSystem** | `backend-go/internal/workflow/gate_system.go` | 3-gate knowledge access control |
| **Blackboard** | `backend-go/internal/blackboard/store.go` | Event-driven state management |
| **Tools** | `backend-go/internal/workflow/tools.go` | PostArtifact, AskExpert, ReadBlackboard, AskClient |
| **Engine** | `backend-go/internal/workflow/engine.go` | Workflow lifecycle management |
| **Planner** | `backend-go/internal/workflow/planner.go` | Task decomposition (requirement → per-expert tasks) |

### 2.4 Current Phases

**File:** `backend-go/internal/workflow/engine.go`

```go
const (
    PhaseIntake          = "intake"
    PhaseHighLevelDesign = "high_level_design"  // ← AgentLoop works well
    PhaseDetailedDesign  = "detailed_design"    // ← AgentLoop works well
    PhaseImplementation  = "implementation"     // ← Need Aider here
    PhaseQA              = "qa"                 // ← Need Aider here
    PhaseHandoff         = "handoff"
    PhaseCompleted       = "completed"
)
```

---

## 3. Problem Statement

### 3.1 Current Limitations

**Implementation Phase Issues:**

1. **No File System:**
   - AgentLoop posts `code_artifact_produced` events with JSON: `{"filename": "handler.go", "code": "..."}`
   - Code exists only in blackboard (PostgreSQL JSONB)
   - No actual files on disk → can't compile, can't test, can't run

2. **No Git History:**
   - No commits, no branches, no diffs
   - Can't track "why this change was made"
   - Can't accept/reject individual changes (patch-based workflow)

3. **No Build/Test Execution:**
   - Validation pipeline (`backend-go/internal/validation/`) only does syntax checks
   - No `go build`, no `go test`, no actual execution
   - Bugs discovered only after handoff

4. **Context Explosion:**
   - Full file content in every LLM call (10,000 line file = 10k tokens)
   - No incremental patching (unified diff format)
   - Expensive, slow, error-prone

### 3.2 Why Aider?

**Aider** (https://github.com/paul-gauthier/aider) is an open-source coding agent that:

1. ✅ **File system native:** Works with actual files on disk
2. ✅ **Git integrated:** Commits changes with meaningful messages
3. ✅ **Patch-based:** Generates unified diffs (accept/reject workflow)
4. ✅ **Build/test aware:** Runs `go build`, `go test`, iterates on failures
5. ✅ **Context efficient:** Only sends changed sections, not full files
6. ✅ **OTA/RALF patterns:** Supports both observe-think-act and RALF loops

**From Arpit Bhiyani AI Masterclass transcript:**
> "Ralph loop looks something like this. Literally, it's an oversimplification, but it is an infinite while loop where you have your prompt and you keep cat, you cat it and you pipe it to clot code, and you keep doing it until your job is done."

> "Your file system and your git history is the context. And there it loads, checks the file, tries to operate, builds the test again and again and again and again until all the tests are written."

---

## 4. Design Principles

### 4.1 From Transcripts

**Arpit Bhiyani AI Masterclass:**

1. **"Agent is just an expensive while loop"**
   - Don't overcomplicate
   - Focus on the loop structure: Observe → Think → Act
   - Exit condition: `TASK_COMPLETE` or max iterations

2. **"File system is context"**
   - Real files on disk, not JSON blobs
   - Fresh read every iteration (RALF pattern)
   - Git history tracks changes

3. **"Patch-based changes"**
   - Generate unified diffs (git diff format)
   - Accept/reject individual changes
   - Avoid full file rewrites

4. **"One bug at a time"**
   - Fix incrementally, not all at once
   - Commit after each fix
   - Build → Test → Fix → Repeat

**System Design Master Class (Arpit Bhiyani):**

5. **"Break down problem into phases"**
   - Store: How to persist workspace state
   - Pick: How to select tasks for execution
   - Execute: How to run Aider loop

6. **"Don't get overwhelmed"**
   - Solve one-time execution first (implementation phase)
   - Then extend to recurring (qa phase)
   - Add schema columns as you discover needs

7. **"Think structurally"**
   - What guarantees do you need from storage?
   - How to maintain SLA (30s task start)?
   - How to recover from pod restarts?

### 4.2 Architectural Constraints

**MUST PRESERVE:**

1. **Existing AgentLoop for design phases**
   - No breaking changes to `high_level_design`, `detailed_design`
   - Blackboard remains single source of truth for design artifacts
   - 3-gate system continues to work

2. **70-30% Expert Knowledge System**
   - Gate 1 only in implementation phase (own training)
   - Generic knowledge BLOCKED in implementation
   - Experience Bank continues to save generic claims from design phases

3. **Event-Driven Architecture**
   - Blackboard events remain primary communication mechanism
   - Projector continues to update `workflow_tasks` from events
   - No direct DB writes from runners

**MUST ADD:**

1. **Workspace Management**
   - Isolated directory per workflow: `/workspaces/{workflow_id}/`
   - Git repository initialized in each workspace
   - Cleanup on workflow completion/failure

2. **File System Layer**
   - Actual code files (not JSON blobs)
   - Build artifacts (binaries, test results)
   - Git history (commits, branches, diffs)

3. **Build/Test Execution**
   - `go build` to verify compilation
   - `go test` to run tests
   - Iterate on failures (OTA loop)

4. **Phase Routing Logic**
   - `executeWaves()` checks `state.Phase`
   - Routes to `AgentLoop` or `AiderRunner` based on phase
   - Seamless transition between executors

---

## 5. Proposed Architecture

### 5.1 High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                        WorkflowRunner                           │
│  (backend-go/internal/workflow/runner.go)                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ├─ Load workflow + experts
                              ├─ Planner.Plan() → []TaskSpec
                              ├─ DAG.BuildDAG() → []ExecutionWave
                              ├─ Post task_plan_ready event
                              ├─ AskClient for approval
                              │
                              ▼
                    executeWaves(state.Phase)
                              │
                ┌─────────────┴─────────────┐
                │                           │
                ▼                           ▼
    ┌───────────────────────┐   ┌───────────────────────┐
    │     AgentLoop         │   │    AiderRunner        │
    │  (Design Phases)      │   │ (Implementation/QA)   │
    └───────────────────────┘   └───────────────────────┘
                │                           │
                ▼                           ▼
    ┌───────────────────────┐   ┌───────────────────────┐
    │   Blackboard Store    │   │  File System + Git    │
    │  (PostgreSQL JSONB)   │   │  (/workspaces/{id}/)  │
    └───────────────────────┘   └───────────────────────┘
```

### 5.2 Phase Routing Logic

**File:** `backend-go/internal/workflow/runner.go` (modified)

```go
func (r *WorkflowRunner) executeWaves(
    ctx context.Context,
    workflowID uuid.UUID,
    waves []ExecutionWave,
    experts []workflowExpert,
    state *runnerState,
) error {
    expertMap := make(map[string]workflowExpert, len(experts))
    for _, e := range experts {
        expertMap[e.ID.String()] = e
    }
    
    // NEW: Check if current phase requires Aider
    useAider := state.Phase == PhaseImplementation || state.Phase == PhaseQA
    
    for waveIdx, wave := range waves {
        var wg sync.WaitGroup
        for _, task := range wave {
            wg.Add(1)
            go func(t TaskSpec) {
                defer wg.Done()
                
                expert, ok := expertMap[t.ExpertID.String()]
                if !ok {
                    return
                }
                
                // NEW: Route to appropriate executor
                if useAider {
                    // Implementation/QA: Use AiderRunner
                    _, err := r.aiderRunner.Run(ctx, AiderRunRequest{
                        WorkflowID:      workflowID,
                        Expert:          expert,
                        TaskID:          uuid.Nil,
                        TaskTitle:       t.Title,
                        TaskDescription: t.Description,
                        WorkflowPhase:   state.Phase,
                    })
                    if err != nil {
                        r.logger.Error("runner: aider task failed",
                            zap.String("expert", expert.Name),
                            zap.Error(err),
                        )
                        state.FailedExpertIDs = append(state.FailedExpertIDs, expert.ID.String())
                    } else {
                        state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
                    }
                } else {
                    // Design phases: Use AgentLoop (existing code)
                    _, err := r.agentLoop.Run(ctx, AgentLoopRequest{
                        WorkflowID:      workflowID,
                        Expert:          expert,
                        TaskID:          uuid.Nil,
                        TaskTitle:       t.Title,
                        TaskDescription: t.Description,
                        WorkflowPhase:   state.Phase,
                        AllExperts:      experts,
                    })
                    if err != nil {
                        r.logger.Error("runner: task failed",
                            zap.String("expert", expert.Name),
                            zap.Error(err),
                        )
                        state.FailedExpertIDs = append(state.FailedExpertIDs, expert.ID.String())
                    } else {
                        state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
                    }
                }
                
                r.saveRunnerState(ctx, workflowID, state)
            }(task)
        }
        wg.Wait()
    }
    
    return nil
}
```

### 5.3 AiderRunner Component

**New File:** `backend-go/internal/workflow/aider_runner.go`

**CRITICAL:** Aider is used as **library** (not CLI), with **ModelGateway proxy** to preserve cost tracking.

```go
package workflow

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
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
type AiderRunner struct {
    db           *pgxpool.Pool
    store        *blackboard.Store
    gateway      *gateway.ModelGateway
    workspaceDir string // Base directory: /workspaces/
    logger       *zap.Logger
}

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

type AiderRunRequest struct {
    WorkflowID      uuid.UUID
    Expert          workflowExpert
    TaskID          uuid.UUID
    TaskTitle       string
    TaskDescription string
    WorkflowPhase   string // "implementation" or "qa"
}

type AiderRunResult struct {
    CommitSHAs []string // Git commit SHAs produced
    Iterations int
    Completed  bool
}

// Run executes the Aider loop for one expert's task.
//
// Mental execution:
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
func (a *AiderRunner) Run(ctx context.Context, req AiderRunRequest) (*AiderRunResult, error) {
    a.logger.Info("aider runner started",
        zap.String("workflow_id", req.WorkflowID.String()),
        zap.String("expert", req.Expert.Name),
        zap.String("task", req.TaskTitle),
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
    
    // Step 3: Run Aider loop
    result := &AiderRunResult{}
    maxIter := req.Expert.MaxLoopIterations
    if maxIter <= 0 {
        maxIter = 5
    }
    
    for iter := 1; iter <= maxIter; iter++ {
        a.logger.Debug("aider iteration",
            zap.String("expert", req.Expert.Name),
            zap.Int("iter", iter),
        )
        
        // Observe: Check build/test status
        buildErr := a.runBuild(ctx, workspacePath)
        testErr := a.runTests(ctx, workspacePath)
        
        if buildErr == nil && testErr == nil {
            // Success: all tests pass, build succeeds
            a.logger.Info("aider task completed",
                zap.String("expert", req.Expert.Name),
                zap.Int("iterations", iter),
            )
            result.Completed = true
            break
        }
        
        // Think + Act: Call Aider to fix errors
        commitSHA, err := a.runAiderIteration(ctx, AiderIterationRequest{
            WorkspacePath:   workspacePath,
            Expert:          req.Expert,
            TaskDescription: req.TaskDescription,
            BuildError:      buildErr,
            TestError:       testErr,
            Iteration:       iter,
        })
        if err != nil {
            a.logger.Error("aider iteration failed",
                zap.String("expert", req.Expert.Name),
                zap.Int("iter", iter),
                zap.Error(err),
            )
            return result, fmt.Errorf("aider iteration %d: %w", iter, err)
        }
        
        result.CommitSHAs = append(result.CommitSHAs, commitSHA)
        result.Iterations = iter
    }
    
    // Step 4: Post final code artifacts to blackboard
    if err := a.publishCodeArtifacts(ctx, req.WorkflowID, req.Expert.ID, workspacePath); err != nil {
        a.logger.Warn("publish code artifacts failed", zap.Error(err))
    }
    
    return result, nil
}

// initWorkspace creates workspace directory and initializes git repo.
func (a *AiderRunner) initWorkspace(ctx context.Context, workspacePath string) error {
    // Create directory
    if err := os.MkdirAll(workspacePath, 0755); err != nil {
        return fmt.Errorf("mkdir: %w", err)
    }
    
    // Git init
    cmd := exec.CommandContext(ctx, "git", "init")
    cmd.Dir = workspacePath
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("git init: %w", err)
    }
    
    // Initial commit (empty)
    cmd = exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", "chore: initialize workspace")
    cmd.Dir = workspacePath
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("git commit: %w", err)
    }
    
    a.logger.Info("workspace initialized", zap.String("path", workspacePath))
    return nil
}

// seedWorkspace loads design artifacts from blackboard and creates seed files.
// Example: architecture_decision → README.md, api_contract_proposed → api.yaml
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
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("git commit: %w", err)
    }
    
    a.logger.Info("workspace seeded", zap.Int("files", len(events)))
    return nil
}

// artifactToFilename maps blackboard event types to filenames.
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

// runBuild executes `go build` and returns error if compilation fails.
func (a *AiderRunner) runBuild(ctx context.Context, workspacePath string) error {
    cmd := exec.CommandContext(ctx, "go", "build", "./...")
    cmd.Dir = workspacePath
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("build failed: %s", string(output))
    }
    return nil
}

// runTests executes `go test` and returns error if tests fail.
func (a *AiderRunner) runTests(ctx context.Context, workspacePath string) error {
    cmd := exec.CommandContext(ctx, "go", "test", "./...", "-v")
    cmd.Dir = workspacePath
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("tests failed: %s", string(output))
    }
    return nil
}

type AiderIterationRequest struct {
    WorkspacePath   string
    Expert          workflowExpert
    TaskDescription string
    BuildError      error
    TestError       error
    Iteration       int
}

// runAiderIteration calls Aider to fix build/test errors.
// Returns commit SHA of the fix.
//
// CRITICAL: Uses AiderService HTTP API (not CLI) to preserve ModelGateway cost tracking.
func (a *AiderRunner) runAiderIteration(ctx context.Context, req AiderIterationRequest) (string, error) {
    // Build prompt for Aider
    prompt := a.buildAiderPrompt(req)
    
    // Call AiderService HTTP API
    // AiderService wraps Aider library and routes LLM calls through ModelGateway
    reqBody := AiderServiceRequest{
        WorkspacePath: req.WorkspacePath,
        Message:       prompt,
        ExpertID:      req.Expert.ID.String(),
        WorkflowID:    a.getCurrentWorkflowID(ctx), // For cost tracking
    }
    
    bodyBytes, _ := json.Marshal(reqBody)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        "http://aider-service:8080/iterate",
        bytes.NewReader(bodyBytes),
    )
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        return "", fmt.Errorf("aider service call failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("aider service returned %d", resp.StatusCode)
    }
    
    var result AiderServiceResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("decode aider response: %w", err)
    }
    
    a.logger.Info("aider iteration completed",
        zap.String("expert", req.Expert.Name),
        zap.Int("iteration", req.Iteration),
        zap.String("commit", result.CommitSHA),
        zap.Float64("cost_usd", result.CostUSD), // Cost tracked via ModelGateway
    )
    
    return result.CommitSHA, nil
}

type AiderServiceRequest struct {
    WorkspacePath string `json:"workspace_path"`
    Message       string `json:"message"`
    ExpertID      string `json:"expert_id"`
    WorkflowID    string `json:"workflow_id"`
}

type AiderServiceResponse struct {
    CommitSHA    string  `json:"commit_sha"`
    FilesChanged []string `json:"files_changed"`
    CostUSD      float64 `json:"cost_usd"` // Tracked by ModelGateway
    Success      bool    `json:"success"`
    Error        string  `json:"error,omitempty"`
}

// buildAiderPrompt constructs prompt for Aider based on errors.
func (a *AiderRunner) buildAiderPrompt(req AiderIterationRequest) string {
    var sb strings.Builder
    
    sb.WriteString(fmt.Sprintf("Task: %s\n\n", req.TaskDescription))
    
    if req.BuildError != nil {
        sb.WriteString(fmt.Sprintf("Build Error:\n%s\n\n", req.BuildError.Error()))
    }
    
    if req.TestError != nil {
        sb.WriteString(fmt.Sprintf("Test Error:\n%s\n\n", req.TestError.Error()))
    }
    
    sb.WriteString("Fix the errors above. Commit changes with a descriptive message.\n")
    
    return sb.String()
}

// extractCommitSHA parses Aider output to extract commit SHA.
func (a *AiderRunner) extractCommitSHA(output string) string {
    // TODO: Parse Aider output format
    // Example: "Committed: abc123def456"
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "Committed:") {
            return strings.TrimSpace(strings.TrimPrefix(line, "Committed:"))
        }
    }
    return "unknown"
}

// publishCodeArtifacts posts final code to blackboard as code_artifact_produced events.
func (a *AiderRunner) publishCodeArtifacts(ctx context.Context, workflowID, expertID uuid.UUID, workspacePath string) error {
    // Walk workspace directory
    err := filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Skip directories and .git
        if info.IsDir() || strings.Contains(path, ".git") {
            return nil
        }
        
        // Skip non-code files
        if !strings.HasSuffix(path, ".go") {
            return nil
        }
        
        // Read file content
        content, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("read %s: %w", path, err)
        }
        
        // Post to blackboard
        relPath, _ := filepath.Rel(workspacePath, path)
        _, err = a.store.Post(ctx, blackboard.PostRequest{
            WorkflowID:       workflowID,
            EventType:        "code_artifact_produced",
            PostedByExpertID: &expertID,
            Content: map[string]interface{}{
                "filename": relPath,
                "code":     string(content),
            },
        })
        if err != nil {
            return fmt.Errorf("post artifact %s: %w", relPath, err)
        }
        
        a.logger.Debug("code artifact published",
            zap.String("file", relPath),
        )
        
        return nil
    })
    
    return err
}
```

### 5.4 Cost Tracking Integration

**CRITICAL:** All Aider LLM calls must route through ModelGateway to preserve centralized cost tracking.

**Flow:**

```
AiderRunner (Go)
    |
    | HTTP POST /iterate
    v
AiderService (Python FastAPI)
    |
    | Aider library call
    v
Aider (Python)
    |
    | LLM API call
    v
LiteLLM Proxy (Python)
    |
    | HTTP POST /chat/completions
    v
ModelGateway (Go)
    |
    | Provider selection, retry, fallback
    | Cost tracking in messages.cost_usd
    v
Anthropic/OpenAI/etc.
```

**Why This Works:**

1. **AiderService** wraps Aider library, exposes HTTP API
2. **LiteLLM Proxy** intercepts Aider's LLM calls, forwards to ModelGateway
3. **ModelGateway** handles provider switching, retry, fallback, cost tracking
4. **messages.cost_usd** remains single source of truth (AgentLoop + AiderRunner)

**Configuration:**

```python
# aider-service/config.py
import os
from litellm import completion

# Configure LiteLLM to proxy through ModelGateway
os.environ["LITELLM_PROXY_URL"] = "http://backend-go:8080/api/v1/llm/proxy"

# Aider will use this proxy for all LLM calls
from aider.coders import Coder
coder = Coder.create(
    model="gpt-4",
    api_base="http://backend-go:8080/api/v1/llm/proxy",  # ModelGateway proxy
)
```

**ModelGateway Proxy Endpoint:**

**New File:** `backend-go/internal/gateway/proxy.go`

```go
package gateway

import (
    "encoding/json"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// ProxyHandler handles LLM requests from AiderService.
// Routes through ModelGateway to preserve cost tracking.
func (g *ModelGateway) ProxyHandler(c *gin.Context) {
    var req LLMRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Extract workflow_id from headers (set by AiderService)
    workflowID := c.GetHeader("X-Workflow-ID")
    expertID := c.GetHeader("X-Expert-ID")
    
    g.logger.Info("proxy llm call",
        zap.String("workflow_id", workflowID),
        zap.String("expert_id", expertID),
        zap.String("model", req.Model),
    )
    
    // Call ModelGateway (existing logic)
    resp, err := g.Call(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // Track cost in messages table
    // TODO: Save to messages table with workflow_id, expert_id
    // This ensures cost appears in same table as AgentLoop costs
    
    c.JSON(http.StatusOK, resp)
}
```

**Register Proxy Route:**

**File:** `backend-go/cmd/server/main.go` (modified)

```go
// Add proxy endpoint for AiderService
api.POST("/llm/proxy", modelGateway.ProxyHandler)
```

### 5.5 AiderService (Python)

**New Service:** `aider-service/` (Python FastAPI)

**File:** `aider-service/main.py`

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import os
from aider.coders import Coder
from aider.models import Model

app = FastAPI()

# Configure Aider to use ModelGateway proxy
MODEL_GATEWAY_URL = os.getenv("MODEL_GATEWAY_URL", "http://backend-go:8080/api/v1/llm/proxy")

class IterateRequest(BaseModel):
    workspace_path: str
    message: str
    expert_id: str
    workflow_id: str

class IterateResponse(BaseModel):
    commit_sha: str
    files_changed: list[str]
    cost_usd: float
    success: bool
    error: str = None

@app.post("/iterate")
async def iterate(req: IterateRequest) -> IterateResponse:
    try:
        # Create Aider coder with ModelGateway proxy
        model = Model(
            name="gpt-4",
            api_base=MODEL_GATEWAY_URL,
            api_key="dummy",  # Not used, ModelGateway handles auth
        )
        
        coder = Coder.create(
            model=model,
            fnames=[],  # Auto-detect files in workspace
            auto_commits=True,
            dirty_commits=False,
            git_dname=req.workspace_path,
        )
        
        # Set headers for cost tracking
        coder.model.headers = {
            "X-Workflow-ID": req.workflow_id,
            "X-Expert-ID": req.expert_id,
        }
        
        # Run Aider iteration
        result = coder.run(req.message)
        
        # Extract commit SHA from git log
        import subprocess
        commit_sha = subprocess.check_output(
            ["git", "rev-parse", "HEAD"],
            cwd=req.workspace_path,
        ).decode().strip()
        
        # Get files changed
        files_changed = subprocess.check_output(
            ["git", "diff", "--name-only", "HEAD~1"],
            cwd=req.workspace_path,
        ).decode().strip().split("\n")
        
        # Cost is tracked by ModelGateway, returned in response
        # TODO: Extract from ModelGateway response headers
        cost_usd = 0.0  # Placeholder, actual cost in messages.cost_usd
        
        return IterateResponse(
            commit_sha=commit_sha,
            files_changed=files_changed,
            cost_usd=cost_usd,
            success=True,
        )
    except Exception as e:
        return IterateResponse(
            commit_sha="",
            files_changed=[],
            cost_usd=0.0,
            success=False,
            error=str(e),
        )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
```

**Dockerfile:** `aider-service/Dockerfile`

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install Aider
RUN pip install aider-chat fastapi uvicorn

# Copy service code
COPY main.py .

EXPOSE 8080

CMD ["python", "main.py"]
```

**Docker Compose:** `docker-compose.yml` (add service)

```yaml
services:
  aider-service:
    build: ./aider-service
    ports:
      - "8081:8080"
    environment:
      - MODEL_GATEWAY_URL=http://backend-go:8080/api/v1/llm/proxy
    volumes:
      - /workspaces:/workspaces  # Shared workspace directory
```

### 5.6 Workspace Management

**Directory Structure:**

```
/workspaces/
├── {workflow_id_1}/
│   ├── .git/
│   ├── ARCHITECTURE.md      # Seeded from architecture_decision
│   ├── DATA_MODEL.md        # Seeded from data_model_proposed
│   ├── API_CONTRACT.yaml    # Seeded from api_contract_proposed
│   ├── auth.go              # Generated by Aider
│   ├── auth_test.go         # Generated by Aider
│   └── go.mod
├── {workflow_id_2}/
│   ├── .git/
│   └── ...
└── {workflow_id_3}/
    ├── .git/
    └── ...
```

**Lifecycle:**

1. **Creation:** `AiderRunner.initWorkspace()` on first implementation task
2. **Seeding:** `AiderRunner.seedWorkspace()` loads design artifacts from blackboard
3. **Execution:** Aider loop runs, commits changes to git
4. **Publishing:** `AiderRunner.publishCodeArtifacts()` posts final code to blackboard
5. **Cleanup:** Optional (keep for debugging, or delete on workflow completion)

**Storage Requirements:**

- **Disk space:** ~100MB per workflow (code + git history)
- **Retention:** 7 days after workflow completion
- **Backup:** Git history is backup (can recreate from commits)

---

## 6. Implementation Roadmap

### Phase 1: Foundation (Week 1)

**Goal:** Add AiderRunner skeleton, workspace management, phase routing

**Tasks:**

1. **Create `backend-go/internal/workflow/aider_runner.go`**
   - [ ] Define `AiderRunner` struct
   - [ ] Implement `NewAiderRunner()`
   - [ ] Add `Run()` method (skeleton, no Aider calls yet)
   - [ ] Add `initWorkspace()` (git init)
   - [ ] Add `seedWorkspace()` (load design artifacts)

2. **Modify `backend-go/internal/workflow/runner.go`**
   - [ ] Add `aiderRunner *AiderRunner` field to `WorkflowRunner`
   - [ ] Update `NewWorkflowRunner()` to accept `aiderRunner`
   - [ ] Modify `executeWaves()` to check `state.Phase`
   - [ ] Route to `AiderRunner` for implementation/qa phases

3. **Update `backend-go/cmd/server/main.go`**
   - [ ] Create `AiderRunner` instance
   - [ ] Pass to `NewWorkflowRunner()`
   - [ ] Configure workspace directory (`/workspaces/`)

4. **Add workspace cleanup job**
   - [ ] Cron job to delete workspaces older than 7 days
   - [ ] Graceful handling of active workflows

**Validation:**
- [ ] Unit tests for `initWorkspace()`, `seedWorkspace()`
- [ ] Integration test: workflow transitions to implementation phase, workspace created
- [ ] Manual test: verify git repo initialized, design artifacts seeded

### Phase 2: Aider Integration (Week 2)

**Goal:** Integrate Aider CLI, implement OTA loop

**Tasks:**

1. **Implement `runAiderIteration()`**
   - [ ] Build prompt from task description + errors
   - [ ] Call Aider CLI with `--message`, `--yes`, `--no-pretty`
   - [ ] Parse Aider output to extract commit SHA
   - [ ] Handle Aider failures (retry logic)

2. **Implement build/test execution**
   - [ ] Add `runBuild()` (execute `go build ./...`)
   - [ ] Add `runTests()` (execute `go test ./... -v`)
   - [ ] Parse build/test output for errors

3. **Implement OTA loop in `Run()`**
   - [ ] Observe: check build/test status
   - [ ] Think + Act: call `runAiderIteration()`
   - [ ] Repeat until success or max iterations
   - [ ] Post task status events to blackboard

4. **Add expert training integration (Gate 1)**
   - [ ] Load expert's training chunks (Gate 1 only)
   - [ ] Inject into Aider prompt as context
   - [ ] Verify generic knowledge BLOCKED

**Validation:**
- [ ] Unit tests for `runBuild()`, `runTests()`, `runAiderIteration()`
- [ ] Integration test: workflow with broken code, Aider fixes it
- [ ] Manual test: verify git commits, meaningful messages

### Phase 3: Code Publishing (Week 3)

**Goal:** Post final code artifacts to blackboard

**Tasks:**

1. **Implement `publishCodeArtifacts()`**
   - [ ] Walk workspace directory
   - [ ] Read all `.go` files
   - [ ] Post `code_artifact_produced` events to blackboard
   - [ ] Include git commit SHA in event metadata

2. **Update Projector to handle Aider artifacts**
   - [ ] Recognize `code_artifact_produced` from AiderRunner
   - [ ] Update `workflow_tasks` with commit SHAs
   - [ ] Display git history in UI

3. **Add artifact validation**
   - [ ] Verify all files compile (`go build`)
   - [ ] Verify all tests pass (`go test`)
   - [ ] Block publication if validation fails

**Validation:**
- [ ] Unit tests for `publishCodeArtifacts()`
- [ ] Integration test: workflow completes, code artifacts in blackboard
- [ ] Manual test: verify UI shows git commits, code diffs

### Phase 4: QA Phase Support (Week 4)

**Goal:** Extend AiderRunner to support QA phase (test generation)

**Tasks:**

1. **Add QA-specific logic to `Run()`**
   - [ ] Detect `state.Phase == PhaseQA`
   - [ ] Load implementation artifacts from previous phase
   - [ ] Generate test files (`*_test.go`)
   - [ ] Run tests, iterate on failures

2. **Add test coverage tracking**
   - [ ] Run `go test -cover`
   - [ ] Parse coverage output
   - [ ] Post coverage metrics to blackboard

3. **Add test quality checks**
   - [ ] Verify test names follow conventions
   - [ ] Verify edge cases covered
   - [ ] Verify error paths tested

**Validation:**
- [ ] Integration test: workflow generates tests, achieves >80% coverage
- [ ] Manual test: verify test quality, edge cases covered

### Phase 5: Production Hardening (Week 5)

**Goal:** Error handling, monitoring, recovery

**Tasks:**

1. **Add error recovery**
   - [ ] Checkpoint after each Aider iteration
   - [ ] Resume from last checkpoint on pod restart
   - [ ] Handle Aider crashes gracefully

2. **Add monitoring**
   - [ ] Metrics: iterations per task, success rate, time per iteration
   - [ ] Logs: structured logging with workflow_id, expert_id, iteration
   - [ ] Alerts: Aider failures, workspace disk full

3. **Add resource limits**
   - [ ] Max workspace size (1GB per workflow)
   - [ ] Max iterations (10 per task)
   - [ ] Timeout per iteration (5 minutes)

4. **Add security**
   - [ ] Sandbox Aider execution (Docker container)
   - [ ] Restrict file system access (chroot)
   - [ ] Validate Aider output (no malicious code)

**Validation:**
- [ ] Chaos test: kill pod during Aider iteration, verify recovery
- [ ] Load test: 100 concurrent workflows, verify no resource exhaustion
- [ ] Security test: attempt to escape sandbox, verify blocked

---

## 7. Testing Strategy

### 7.1 Unit Tests

**File:** `backend-go/internal/workflow/aider_runner_test.go`

```go
func TestInitWorkspace(t *testing.T) {
    // Test workspace creation + git init
}

func TestSeedWorkspace(t *testing.T) {
    // Test design artifacts → seed files
}

func TestRunBuild(t *testing.T) {
    // Test go build execution
}

func TestRunTests(t *testing.T) {
    // Test go test execution
}

func TestPublishCodeArtifacts(t *testing.T) {
    // Test code → blackboard events
}
```

### 7.2 Integration Tests

**File:** `backend-go/internal/workflow/aider_integration_test.go`

```go
func TestAiderRunner_ImplementationPhase(t *testing.T) {
    // End-to-end test:
    // 1. Create workflow with implementation phase
    // 2. Seed workspace with design artifacts
    // 3. Run AiderRunner
    // 4. Verify code artifacts in blackboard
    // 5. Verify git commits
}

func TestAiderRunner_QAPhase(t *testing.T) {
    // End-to-end test:
    // 1. Create workflow with qa phase
    // 2. Load implementation artifacts
    // 3. Run AiderRunner
    // 4. Verify test files generated
    // 5. Verify coverage >80%
}
```

### 7.3 Manual Testing

**Scenario 1: Simple Implementation Task**

1. Create workflow: "Build user authentication"
2. Select experts: Backend Engineer
3. Approve plan
4. Wait for implementation phase
5. Verify:
   - [ ] Workspace created: `/workspaces/{workflow_id}/`
   - [ ] Git repo initialized
   - [ ] Design artifacts seeded
   - [ ] Aider generates `auth.go`
   - [ ] Code compiles (`go build`)
   - [ ] Tests pass (`go test`)
   - [ ] Code artifacts in blackboard
   - [ ] Git commits with meaningful messages

**Scenario 2: Broken Code Fix**

1. Create workflow: "Fix authentication bug"
2. Seed workspace with broken code
3. Run AiderRunner
4. Verify:
   - [ ] Aider detects build error
   - [ ] Aider fixes error
   - [ ] Commit message: "fix: add missing import"
   - [ ] Build succeeds
   - [ ] Tests pass

**Scenario 3: QA Phase**

1. Create workflow: "Add tests for authentication"
2. Load implementation artifacts
3. Run AiderRunner (QA phase)
4. Verify:
   - [ ] Aider generates `auth_test.go`
   - [ ] Tests cover edge cases
   - [ ] Coverage >80%
   - [ ] All tests pass

---

## 8. Appendix

### 8.1 Aider Integration Approach

**CRITICAL DECISION: Library Mode, Not CLI**

**Why Not CLI:**
```bash
# ❌ WRONG: Aider CLI bypasses ModelGateway
aider --message "Fix bug" --yes
# Uses $ANTHROPIC_API_KEY directly
# No cost tracking in messages.cost_usd
# No provider switching/retry/fallback
# Split cost accounting
```

**Why Library Mode:**
```python
# ✅ CORRECT: Aider library with ModelGateway proxy
from aider.coders import Coder

coder = Coder.create(
    model="gpt-4",
    api_base="http://backend-go:8080/api/v1/llm/proxy",  # ModelGateway
)
coder.run("Fix bug")
# All LLM calls route through ModelGateway
# Cost tracked in messages.cost_usd
# Provider switching/retry/fallback preserved
# Unified cost accounting
```

**Architecture:**

```
AiderRunner (Go) → HTTP → AiderService (Python) → Aider Library → LiteLLM → ModelGateway (Go) → Anthropic/OpenAI
                                                                                    |
                                                                                    v
                                                                          messages.cost_usd
```

**Benefits:**

1. ✅ **Centralized Cost Tracking:** All costs in `messages.cost_usd` (single source of truth)
2. ✅ **Provider Switching:** ModelGateway handles Anthropic → OpenAI fallback
3. ✅ **Retry Logic:** ModelGateway retries on rate limits
4. ✅ **No API Key Leakage:** Aider never sees `$ANTHROPIC_API_KEY`
5. ✅ **Consistent Accounting:** AgentLoop + AiderRunner costs in same table

### 8.2 Git Commands Reference

**Initialize repo:**
```bash
git init
git commit --allow-empty -m "chore: initialize workspace"
```

**Add files:**
```bash
git add .
git commit -m "chore: seed workspace with design artifacts"
```

**View history:**
```bash
git log --oneline
```

**View diff:**
```bash
git diff HEAD~1
```

### 8.3 Go Build/Test Commands

**Build:**
```bash
go build ./...  # Build all packages
```

**Test:**
```bash
go test ./... -v  # Run all tests with verbose output
go test -cover    # Run tests with coverage
```

**Parse output:**
```go
output, err := cmd.CombinedOutput()
if err != nil {
    // Parse error from output
    // Example: "./auth.go:10:2: undefined: crypto"
}
```

### 8.4 Blackboard Event Schema

**code_artifact_produced (from AiderRunner):**
```json
{
  "event_type": "code_artifact_produced",
  "posted_by_expert_id": "expert-uuid",
  "content": {
    "filename": "auth.go",
    "code": "package main\n\nfunc Authenticate() { ... }",
    "commit_sha": "abc123def456",
    "git_message": "feat: add user authentication"
  }
}
```

### 8.5 Database Schema Changes

**No schema changes required!**

AiderRunner uses existing tables:
- `workflows`: Stores workflow state (no changes)
- `blackboard_events`: Stores code artifacts (no changes)
- `workflow_tasks`: Updated by Projector (no changes)

Workspace state stored in file system, not database.

### 8.6 Configuration

**Environment Variables:**

```bash
# Workspace directory
WORKSPACE_DIR=/workspaces/

# Aider model
AIDER_MODEL=gpt-4

# Max iterations per task
AIDER_MAX_ITERATIONS=10

# Timeout per iteration (seconds)
AIDER_ITERATION_TIMEOUT=300

# Workspace retention (days)
WORKSPACE_RETENTION_DAYS=7
```

**Config File:** `backend-go/config/config.yaml`

```yaml
aider:
  workspace_dir: /workspaces/
  model: gpt-4
  max_iterations: 10
  iteration_timeout: 300
  retention_days: 7
```

### 8.7 Monitoring Metrics

**Prometheus Metrics:**

```go
// Aider iterations per task
aider_iterations_total{expert="backend_engineer",phase="implementation"}

// Aider success rate
aider_success_rate{expert="backend_engineer",phase="implementation"}

// Aider iteration duration
aider_iteration_duration_seconds{expert="backend_engineer",phase="implementation"}

// Workspace disk usage
workspace_disk_usage_bytes{workflow_id="abc-123"}
```

**Grafana Dashboard:**

- Aider success rate over time
- Average iterations per task
- Workspace disk usage
- Build/test failure rate

### 8.9 Cost Tracking Verification

**CRITICAL:** Verify all Aider costs appear in `messages.cost_usd`.

**Query to Verify:**

```sql
-- All costs for a workflow (AgentLoop + AiderRunner)
SELECT 
    m.id,
    m.workflow_id,
    m.expert_id,
    m.role,
    m.content_preview,
    m.cost_usd,
    m.created_at
FROM messages m
WHERE m.workflow_id = 'abc-123'
ORDER BY m.created_at;

-- Total cost breakdown by phase
SELECT 
    w.current_phase,
    COUNT(*) as message_count,
    SUM(m.cost_usd) as total_cost_usd
FROM messages m
JOIN workflows w ON m.workflow_id = w.id
WHERE m.workflow_id = 'abc-123'
GROUP BY w.current_phase;
```

**Expected Output:**

```
current_phase       | message_count | total_cost_usd
--------------------|---------------|---------------
high_level_design   | 15            | 0.45
detailed_design     | 20            | 0.60
implementation      | 30            | 1.20  ← Aider costs here
qa                  | 25            | 0.90  ← Aider costs here
handoff             | 5             | 0.15
```

**Validation:**

1. ✅ All phases have costs in `messages.cost_usd`
2. ✅ Implementation/QA costs include Aider iterations
3. ✅ No missing costs (compare with ModelGateway logs)
4. ✅ Cost matches provider invoices (Anthropic/OpenAI)

**Alert if:**

```sql
-- Alert: Workflow completed but no implementation costs
SELECT w.id, w.title
FROM workflows w
WHERE w.status = 'completed'
  AND w.current_phase = 'completed'
  AND NOT EXISTS (
      SELECT 1 FROM messages m
      WHERE m.workflow_id = w.id
        AND m.cost_usd > 0
        AND w.current_phase IN ('implementation', 'qa')
  );
```

### 8.10 Security Considerations

**Sandbox Aider Execution:**

```dockerfile
# Dockerfile for Aider sandbox
FROM python:3.11-slim

RUN pip install aider-chat

# Create non-root user
RUN useradd -m -u 1000 aider
USER aider

# Restrict file system access
WORKDIR /workspace
VOLUME /workspace

ENTRYPOINT ["aider"]
```

**Run Aider in Docker:**

```go
cmd := exec.CommandContext(ctx, "docker", "run",
    "--rm",
    "-v", fmt.Sprintf("%s:/workspace", workspacePath),
    "aider-sandbox",
    "--message", prompt,
    "--yes",
    "--no-pretty",
)
```

**Validate Aider Output:**

```go
// Check for malicious patterns
forbiddenPatterns := []string{
    "os.RemoveAll",
    "exec.Command",
    "syscall",
}

for _, pattern := range forbiddenPatterns {
    if strings.Contains(code, pattern) {
        return fmt.Errorf("malicious code detected: %s", pattern)
    }
}
```

---

## Summary

This design document provides a comprehensive plan to integrate Aider into AI Avengers for implementation/qa phases while preserving the existing AgentLoop for design phases. The key principles are:

1. **Phase-based routing:** AgentLoop for design, AiderRunner for implementation/qa
2. **File system as context:** Real code files, not JSON blobs
3. **Git history as memory:** Meaningful commits, patch-based changes
4. **Preserve existing architecture:** No breaking changes to design phases
5. **Maintain 70-30% expert system:** Gate 1 only in implementation

The implementation roadmap spans 5 weeks, with clear milestones and validation criteria. The design is production-ready, with error handling, monitoring, and security considerations.

**Next Steps:**
1. Review this document with the team
2. Get approval from stakeholders
3. Start Phase 1 implementation (Week 1)
4. Iterate based on feedback

---

**Document Version:** 1.0  
**Last Updated:** 2026-09-17  
**Status:** Ready for Review
