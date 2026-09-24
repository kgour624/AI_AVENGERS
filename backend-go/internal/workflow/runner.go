package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/monitoring"
	"ai_avengers/backend/internal/observability"
)

// maxDesignAttempts bounds how many times the design phases may be produced
// again at the client's request. Each attempt blocks on a human response at
// the approval gate, so this cannot spin on its own — it is a runaway-cost
// guard, not a policy limit.
const maxDesignAttempts = 10

// decisionChangesRequested is the approval_requests.status value written when
// the client presses "Request Changes" (see handler.decisionToStatus).
const decisionChangesRequested = "changes_requested"

// runnerState is the checkpoint saved after each task completes.
// Written to workflows.runner_state for pod-restart recovery.
type runnerState struct {
	Phase              string   `json:"phase"`
	CompletedExpertIDs []string `json:"completed_expert_ids"`
	FailedExpertIDs    []string `json:"failed_expert_ids"`
}

// WorkflowRunner drives a workflow from start to completion.
//
// FLOW:
//  1. Load workflow + experts from DB
//  2. Load requirement from blackboard
//  3. Planner.Plan() -> []TaskSpec
//  4. DAG.BuildDAG() -> []ExecutionWave (topological sort)
//  5. Post task_plan_ready event (Projector creates workflow_tasks)
//  6. AskClient for plan approval -> workflow pauses
//  7. waitForResume() polls DB
//  8. executeWaves(): wave by wave, parallel within each wave
//  9. AskClient for final approval
//  10. engine.Complete()
//
// RECOVERY (Fix 5):
//
//	On pod restart: main.go calls ResumeOrphanWorkflows()
//	-> finds workflows WHERE status='running'
//	-> re-launches Run() for each
//	-> Run() reads runner_state + current_task_cursor
//	-> skips already-completed tasks
//
// SINGLE WRITE PATH (Fix 1):
//
//	Runner posts events on blackboard.
//	Projector reads events and updates workflow_tasks.
//	Runner never writes directly to workflow_tasks.
type WorkflowRunner struct {
	db              *pgxpool.Pool
	engine          *Engine
	planner         *Planner
	agentLoop       *AgentLoop
	aiderRunner     *AiderRunner            // kept for cross-verifier and any future code-gen path
	authoringRunner *AuthoringRunner        // §9: implementation phase authors design sections
	qaRunner        *QARunner               // §19: QA phase — testing experts propose test cases
	workspaceMerger *WorkspaceMerger        // Merges per-expert workspaces after each Aider wave
	crossVerifier   *CrossVerifier          // Cross-verification protocol (§8)
	costMonitor     *monitoring.CostMonitor // Per-workflow budget cap enforcement
	tools           *Tools
	store           *blackboard.Store
	gateway         *gateway.ModelGateway
	// crSvc is optional: nil disables the change-request watcher.
	// Set via WithChangeRequestService after construction.
	crSvc           *ChangeRequestService
	logger          *zap.Logger
}

func NewWorkflowRunner(
	db *pgxpool.Pool,
	engine *Engine,
	planner *Planner,
	agentLoop *AgentLoop,
	aiderRunner *AiderRunner,
	authoringRunner *AuthoringRunner,
	qaRunner *QARunner,
	workspaceMerger *WorkspaceMerger,
	crossVerifier *CrossVerifier,
	costMonitor *monitoring.CostMonitor,
	tools *Tools,
	store *blackboard.Store,
	gw *gateway.ModelGateway,
	logger *zap.Logger,
) *WorkflowRunner {
	return &WorkflowRunner{
		db:              db,
		engine:          engine,
		planner:         planner,
		agentLoop:       agentLoop,
		aiderRunner:     aiderRunner,
		authoringRunner: authoringRunner,
		qaRunner:        qaRunner,
		workspaceMerger: workspaceMerger,
		crossVerifier:   crossVerifier,
		costMonitor:     costMonitor,
		tools:           tools,
		store:           store,
		gateway:         gw,
		logger:          logger,
	}
}

// WithChangeRequestService attaches the ChangeRequestService to the runner,
// enabling the change-request watcher goroutine.
// Called from main.go after both the runner and the service are constructed.
func (r *WorkflowRunner) WithChangeRequestService(crSvc *ChangeRequestService) {
	r.crSvc = crSvc
}

// Run drives the workflow to completion.
// Designed to run in a goroutine: go runner.Run(ctx, workflowID)
//
// Mental execution:
//
//	workflowID = abc
//	experts = [PM, SD, DSA, LLD]
//	requirement = "Build URL shortener"
//
//	Plan: [{PM,deps:[]}, {SD,deps:[PM]}, {DSA,deps:[PM]}, {LLD,deps:[SD,DSA]}]
//	DAG:  Wave0=[PM], Wave1=[SD,DSA], Wave2=[LLD]
//
//	Post task_plan_ready -> Projector inserts 4 workflow_tasks
//	AskClient -> pause
//	[Client approves]
//	Wave0: PM runs alone
//	Wave1: SD + DSA run in parallel (errgroup)
//	Wave2: LLD runs after both done
//	AskClient -> pause
//	[Client approves]
//	Complete
func (r *WorkflowRunner) Run(ctx context.Context, workflowID uuid.UUID) {
	log := r.logger.With(zap.String("workflow_id", workflowID.String()))
	log.Info("workflow runner started")
	observability.Global.IncWorkflowStarted()

	// Step 1: Load workflow.
	wf, err := r.engine.GetByID(ctx, workflowID)
	if err != nil {
		log.Error("runner: load workflow failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "load workflow failed")
		return
	}

	// Step 2: Load experts.
	experts, err := r.loadWorkflowExperts(ctx, wf.SelectedExpertIDs)
	if err != nil {
		log.Error("runner: load experts failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "load experts failed")
		return
	}

	// Start the change request watcher goroutine.
	// It runs for the lifetime of this workflow and picks up any change
	// requests the client submits from the workflow chat.
	// crSvc is nil when the runner is constructed without change-request
	// support — the nil check makes the feature opt-in without breaking
	// existing callers that have not called WithChangeRequestService.
	if r.crSvc != nil {
		go r.watchForChangeRequest(ctx, workflowID, experts, nil, r.crSvc)
	}

	// Step 3: Load requirement.
	requirementText := r.loadRequirement(ctx, workflowID, wf.Title)

	// Step 4: Plan.
	log.Info("runner: planning", zap.Int("experts", len(experts)))
	tasks, err := r.planner.Plan(ctx, workflowID, requirementText, experts)
	if err != nil {
		log.Error("runner: planning failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "planning failed: "+err.Error())
		return
	}

	// Step 5: Build DAG (topological sort + cycle check).
	waves, err := BuildDAG(tasks)
	if err != nil {
		log.Error("runner: DAG build failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "DAG build failed: "+err.Error())
		return
	}
	log.Info("runner: DAG built",
		zap.Int("tasks", len(tasks)),
		zap.Int("waves", len(waves)),
	)

	// Step 6: Post task_plan_ready event.
	// Projector will INSERT workflow_tasks rows from this event.
	// Runner does NOT write to workflow_tasks directly.
	planContent := buildPlanContent(tasks, experts)
	_, err = r.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "task_plan_ready",
		PostedByClient: true,
		Content:        planContent,
	})
	if err != nil {
		log.Error("runner: post task_plan_ready failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "post plan failed")
		return
	}

	// Step 7: AskClient for plan approval.
	_, err = r.tools.AskClient(ctx, AskClientRequest{
		WorkflowID:      workflowID,
		FromExpertID:    uuid.Nil,
		GateName:        "intake",
		Summary:         fmt.Sprintf("%d tasks planned for %d experts. Review and approve.", len(tasks), len(experts)),
		ArtifactContent: planContent,
	})
	if err != nil {
		log.Error("runner: AskClient failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "AskClient failed")
		return
	}

	log.Info("runner: waiting for plan approval")
	if err := r.waitForResume(ctx, workflowID); err != nil {
		log.Warn("runner: wait for resume failed", zap.Error(err))
		return
	}

	// Step 8: Execute all phases in sequence.
	// Each design phase runs the full wave set.
	// After design phases, AskClient gates implementation start.
	// Implementation and QA use AiderRunner (file system + git).

	// --- Design phases, repeatable on client request ---
	//
	// The two design phases plus their approval gate run inside a loop. If the
	// client responds "request_changes" at the gate, the design is produced
	// again — picking up whatever generic allowance the client set with that
	// response (executeWaves re-reads workflows.generic_allowance_pct on every
	// attempt). Approving breaks the loop and implementation starts.
	//
	// WHY a loop and not "experts revise in place": the client's dial changes
	// what the experts are ALLOWED to use as source material, which is an
	// input to the gate system, not something an expert can patch afterwards.
	// The design has to be produced again under the new rule.
	//
	// The loop cannot spin on its own — every iteration blocks on a human
	// response at the gate. maxDesignAttempts is only a runaway-cost guard.
	var state *runnerState
	for attempt := 1; attempt <= maxDesignAttempts; attempt++ {
		if attempt == 1 {
			_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseHighLevelDesign, nil, 0)
		} else {
			// Backwards move, so TransitionPhase's forward-only validation
			// does not apply. See Engine.RestartPhase.
			if rErr := r.engine.RestartPhase(ctx, workflowID, PhaseHighLevelDesign); rErr != nil {
				log.Error("runner: restart HLD failed", zap.Error(rErr))
				return
			}
			log.Info("runner: re-running design phases on client request",
				zap.Int("attempt", attempt),
			)
		}

		state = &runnerState{Phase: PhaseHighLevelDesign}
		if err := r.executeWaves(ctx, workflowID, waves, experts, state); err != nil {
			log.Error("runner: HLD waves failed", zap.Error(err))
			if len(state.CompletedExpertIDs) == 0 {
				_ = r.engine.Fail(ctx, workflowID, "HLD phase: all tasks failed")
				return
			}
		}

		if attempt == 1 {
			_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseDetailedDesign, nil, 0)
		} else if rErr := r.engine.RestartPhase(ctx, workflowID, PhaseDetailedDesign); rErr != nil {
			log.Error("runner: restart detailed design failed", zap.Error(rErr))
			return
		}

		state = &runnerState{Phase: PhaseDetailedDesign}
		if err := r.executeWaves(ctx, workflowID, waves, experts, state); err != nil {
			log.Error("runner: DetailedDesign waves failed", zap.Error(err))
			if len(state.CompletedExpertIDs) == 0 {
				_ = r.engine.Fail(ctx, workflowID, "DetailedDesign phase: all tasks failed")
				return
			}
		}

		// Gate: client reads the deliverables and decides.
		decision, gateErr := r.askDesignGate(ctx, workflowID, attempt)
		if gateErr != nil {
			log.Error("runner: design gate failed", zap.Error(gateErr))
			return
		}
		if decision != decisionChangesRequested {
			break
		}
		if attempt == maxDesignAttempts {
			log.Warn("runner: design re-run limit reached, proceeding",
				zap.Int("max_attempts", maxDesignAttempts),
			)
		}
	}

	// The design gate lives inside the loop above (askDesignGate), so by the
	// time we get here the client has approved the design.

	// --- Phase: Implementation (Aider) ---
	_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseImplementation, nil, 0)
	state = &runnerState{Phase: PhaseImplementation}
	if err := r.executeWaves(ctx, workflowID, waves, experts, state); err != nil {
		log.Error("runner: Implementation waves failed", zap.Error(err))
		if len(state.CompletedExpertIDs) == 0 {
			_ = r.engine.Fail(ctx, workflowID, "Implementation phase: all tasks failed")
			return
		}
	}

	// --- Phase: QA (Aider - test generation) ---
	_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseQA, nil, 0)
	state = &runnerState{Phase: PhaseQA}
	if err := r.executeWaves(ctx, workflowID, waves, experts, state); err != nil {
		log.Error("runner: QA waves failed", zap.Error(err))
		// QA failure is non-fatal — code is already written
		log.Warn("runner: QA phase had failures, continuing to handoff",
			zap.Int("completed", len(state.CompletedExpertIDs)),
			zap.Int("failed", len(state.FailedExpertIDs)),
		)
	}

	// Step 9: Final approval.
	_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseHandoff, nil, 0)
	allArtifacts, _ := r.store.GetByType(ctx, workflowID,
		[]string{"architecture_decision", "data_model_proposed", "api_contract_proposed",
			"module_design_proposed", "code_artifact_produced", "requirement_captured"}, 0)
	artifactIDs := make([]uuid.UUID, 0, len(allArtifacts))
	for _, ev := range allArtifacts {
		artifactIDs = append(artifactIDs, ev.ID)
	}
	_, err = r.tools.AskClient(ctx, AskClientRequest{
		WorkflowID:    workflowID,
		FromExpertID:  uuid.Nil,
		GateName:      "handoff",
		Summary:       fmt.Sprintf("%d tasks done (%d failed). %d artifacts produced.", len(state.CompletedExpertIDs), len(state.FailedExpertIDs), len(allArtifacts)),
		CitedEventIDs: artifactIDs,
	})
	if err != nil {
		log.Error("runner: final AskClient failed", zap.Error(err))
		return
	}

	log.Info("runner: waiting for final approval")
	if err := r.waitForResume(ctx, workflowID); err != nil {
		return
	}
	_ = r.engine.Complete(ctx, workflowID)
	log.Info("runner: workflow completed")
}

// executeWaves runs execution waves sequentially.
// Within each wave, tasks run in parallel using goroutines + WaitGroup.
//
// Mental execution:
//
//	waves = [[PM], [SD, DSA], [LLD]]
//
//	Wave 0: launch PM goroutine, wait
//	Wave 1: launch SD goroutine + DSA goroutine simultaneously, wait for both
//	Wave 2: launch LLD goroutine, wait
//
// WHY WaitGroup not errgroup:
//
//	errgroup cancels all goroutines on first error.
//	We want ALL tasks in a wave to complete (even if some fail).
//	Failed tasks are marked via blackboard events, not by cancelling others.
//
// WORKSPACE MERGE (Aider phases only):
//
//	After each wave completes, WorkspaceMerger merges per-expert workspaces
//	into a shared main/ workspace so the next wave's experts see all prior work.
//	Merge errors are non-fatal: logged and execution continues.
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

	var firstErr error
	var errMu sync.Mutex

	// Check if current phase requires Aider (implementation/qa).
	// Design phases (high_level_design, detailed_design, handoff) use AgentLoop.
	// Implementation/QA phases use AiderRunner (file system + git).
	useAider := state.Phase == PhaseImplementation || state.Phase == PhaseQA

	// Client's generic ceiling, read fresh for this phase attempt so a re-run
	// picks up a dial the client just changed. Only the design phases use it
	// (AiderRunner builds its own prompt and never allows generic).
	genericAllowancePct := r.loadGenericAllowancePct(ctx, workflowID)

	// workflowWorkspace is the base dir for all expert workspaces in this workflow.
	// Structure: /workspaces/{workflow_id}/{expert_id}/
	workflowWorkspace := fmt.Sprintf("%s/%s", r.aiderRunner.workspaceDir, workflowID.String())

	for waveIdx, wave := range waves {
		r.logger.Info("runner: executing wave",
			zap.Int("wave", waveIdx),
			zap.Int("tasks", len(wave)),
			zap.String("phase", state.Phase),
			zap.Bool("use_aider", useAider),
		)

		// Per-workflow budget cap check before each wave.
		// WHY before wave not after: prevent spending on a wave that would breach the cap.
		// Soft limit: warn + continue (admin visibility, non-blocking).
		// Hard limit: post cost_limit_exceeded event, fail workflow immediately.
		if r.costMonitor != nil {
			softHit, hardHit, costErr := r.costMonitor.CheckWorkflowLimits(ctx, workflowID.String())
			if costErr != nil {
				// Non-fatal: log and continue. Don't block workflow on monitoring failure.
				r.logger.Warn("runner: cost limit check failed (non-fatal)",
					zap.Int("wave", waveIdx),
					zap.Error(costErr),
				)
			} else if hardHit {
				// Hard limit hit: stop workflow immediately.
				r.logger.Error("runner: workflow hard cost limit hit, stopping",
					zap.Int("wave", waveIdx),
					zap.String("workflow_id", workflowID.String()),
				)
				_, _ = r.store.Post(ctx, blackboard.PostRequest{
					WorkflowID:     workflowID,
					EventType:      "cost_limit_exceeded",
					PostedByClient: false,
					Content: map[string]interface{}{
						"limit_type": "hard",
						"wave":       waveIdx,
						"phase":      state.Phase,
					},
				})
				_ = r.engine.Fail(ctx, workflowID, "hard cost limit exceeded")
				return fmt.Errorf("hard cost limit exceeded at wave %d", waveIdx)
			} else if softHit {
				// Soft limit hit: warn, continue.
				r.logger.Warn("runner: workflow soft cost limit hit, continuing",
					zap.Int("wave", waveIdx),
					zap.String("workflow_id", workflowID.String()),
				)
			}
		}

		// Track last blackboard sequence before wave starts (for cross-verification scoping).
		lastSeqBefore := r.getLastBlackboardSeq(ctx, workflowID)

		// Track which expert IDs succeeded in this wave (for post-wave merge).
		var waveMu sync.Mutex
		var waveExpertIDs []string

		var wg sync.WaitGroup
		for _, task := range wave {
			wg.Add(1)
			go func(t TaskSpec) {
				defer wg.Done()

				expert, ok := expertMap[t.ExpertID.String()]
				if !ok {
					return
				}

				// Route to appropriate executor based on phase.
				if useAider {
					// Implementation phase now means AUTHORING (§9 of
					// COLLABORATIVE_DESIGN_ARCHITECTURE.md): each expert edits
					// its own design section and appends to the spine, instead
					// of writing application code. This is the doc's stated
					// change, not an optional mode — §9's table states
					// "Editable files / Verification / Completion" as definite
					// replacements, not a toggle. AiderRunner.Run (application
					// code) is retired from this path.
					//
					// QA phase is NOT covered here — left calling AiderRunner
					// unchanged below. §9 does not define what "QA" means once
					// application code is written by an external agent rather
					// than this system; redefining it was not part of this
					// step's scope and inventing a meaning here would be
					// guessing at a decision the doc never made. Tracked as a
					// gap, not silently resolved.
					if state.Phase == PhaseImplementation {
						reqSpec := AiderRunRequest{
							WorkflowID:      workflowID,
							Expert:          expert,
							TaskID:          uuid.Nil,
							TaskTitle:       t.Title,
							TaskDescription: t.Description,
							WorkflowPhase:   state.Phase,
						}
						result, err := r.authoringRunner.Run(ctx, reqSpec)
						if err != nil {
							r.logger.Error("runner: authoring task failed",
								zap.String("expert", expert.Name),
								zap.Error(err),
							)
							errMu.Lock()
							if firstErr == nil {
								firstErr = err
							}
							state.FailedExpertIDs = append(state.FailedExpertIDs, expert.ID.String())
							errMu.Unlock()
						} else {
							r.logger.Info("runner: authoring task completed",
								zap.String("expert", expert.Name),
								zap.String("section", result.SectionPath),
								zap.Bool("completed", result.Completed),
							)
							errMu.Lock()
							state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
							errMu.Unlock()
							waveMu.Lock()
							waveExpertIDs = append(waveExpertIDs, expert.ID.String())
							waveMu.Unlock()
						}
						r.saveRunnerState(ctx, workflowID, state)
						return
					}

					// QA phase: testing experts (java-tester, react-tester, etc.)
					// read the completed design and propose test cases as
					// acceptance criteria. No Aider, no go build, no git workspace.
					// See qa_runner.go and §19 of COLLABORATIVE_DESIGN_ARCHITECTURE.md.
					qaResult, err := r.qaRunner.Run(ctx, AiderRunRequest{
						WorkflowID:      workflowID,
						Expert:          expert,
						TaskID:          uuid.Nil,
						TaskTitle:       t.Title,
						TaskDescription: t.Description,
						WorkflowPhase:   state.Phase,
					})
					if err != nil {
						r.logger.Error("runner: qa task failed",
							zap.String("expert", expert.Name),
							zap.Error(err),
						)
						errMu.Lock()
						if firstErr == nil {
							firstErr = err
						}
						state.FailedExpertIDs = append(state.FailedExpertIDs, expert.ID.String())
						errMu.Unlock()
					} else {
						r.logger.Info("runner: qa task completed",
							zap.String("expert", expert.Name),
							zap.Int("proposed", qaResult.ProposedCount),
							zap.Int("gaps", qaResult.GapCount),
							zap.Bool("completed", qaResult.Completed),
						)
						errMu.Lock()
						state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
						errMu.Unlock()
						// QA does not produce per-expert workspaces, so there
						// is nothing to merge. waveExpertIDs is intentionally
						// NOT appended here — the post-wave merge below is
						// gated on len(waveExpertIDs) > 0 and is a no-op when
						// the slice is empty.
					}
				} else {
					// Design phases: Use AgentLoop (blackboard-based, existing code).
					_, err := r.agentLoop.Run(ctx, AgentLoopRequest{
						WorkflowID:      workflowID,
						Expert:          expert,
						TaskID:          uuid.Nil,
						TaskTitle:       t.Title,
						TaskDescription: t.Description,
						// WorkflowPhase: passed so AgentLoop knows design vs implementation.
						// Design phases: Gates 1+2+3 active.
						// Implementation phase: Gate 1 only, generic BLOCKED.
						WorkflowPhase: state.Phase,
						// AllExperts: all workflow experts for Gate 2 peer poll.
						AllExperts: experts,
						// GenericAllowancePct: 0 means trained + peer only.
						GenericAllowancePct: genericAllowancePct,
					})
					if err != nil {
						r.logger.Error("runner: task failed",
							zap.String("expert", expert.Name),
							zap.Error(err),
						)
						errMu.Lock()
						if firstErr == nil {
							firstErr = err
						}
						state.FailedExpertIDs = append(state.FailedExpertIDs, expert.ID.String())
						errMu.Unlock()
					} else {
						errMu.Lock()
						state.CompletedExpertIDs = append(state.CompletedExpertIDs, expert.ID.String())
						errMu.Unlock()
					}
				}

				// Save checkpoint after each task.
				r.saveRunnerState(ctx, workflowID, state)
			}(task)
		}
		wg.Wait()

		// Post-wave workspace merge (Aider phases only).
		if useAider && len(waveExpertIDs) > 0 {
			r.logger.Info("runner: merging wave workspaces",
				zap.Int("wave", waveIdx),
				zap.Strings("expert_ids", waveExpertIDs),
			)
			if mergeErr := r.workspaceMerger.MergeWave(ctx, workflowWorkspace, waveExpertIDs); mergeErr != nil {
				r.logger.Error("runner: workspace merge failed (non-fatal)",
					zap.Int("wave", waveIdx),
					zap.Error(mergeErr),
				)
			} else {
				r.logger.Info("runner: wave workspaces merged",
					zap.Int("wave", waveIdx),
					zap.Int("experts_merged", len(waveExpertIDs)),
				)

				// Post wave_completed so the frontend (and any other listener)
				// knows main/ was just updated with this wave's merged code.
				// Non-fatal: logged and ignored on failure, matches the
				// pattern used for every other blackboard.Post call in this
				// function (e.g. task_plan_ready, cost_limit_exceeded above).
				mainPath := fmt.Sprintf("%s/main", workflowWorkspace)
				mergedFiles := listMergedFiles(mainPath)
				_, _ = r.store.Post(ctx, blackboard.PostRequest{
					WorkflowID:     workflowID,
					EventType:      "wave_completed",
					PostedByClient: false,
					Content: map[string]interface{}{
						"wave_index":   waveIdx,
						"phase":        state.Phase,
						"expert_ids":   waveExpertIDs,
						"merged_files": mergedFiles,
					},
				})
			}
		}

		// Post-wave cross-verification (all phases).
		// Run after workspace merge so artifacts are on blackboard.
		// Non-fatal: log error but continue to next wave.
		if r.crossVerifier != nil {
			if cvErr := r.crossVerifier.VerifyWaveArtifacts(
				ctx, workflowID, experts, lastSeqBefore,
			); cvErr != nil {
				r.logger.Error("runner: cross-verification failed (non-fatal)",
					zap.Int("wave", waveIdx),
					zap.Error(cvErr),
				)
			}
		}
	}

	errMu.Lock()
	defer errMu.Unlock()
	return firstErr
}

// listMergedFiles returns the relative paths of all files currently in
// mainPath (the shared main/ workspace), excluding .git/. Used to populate
// the wave_completed event so listeners (e.g. a live file browser) know
// what's in main/ without doing their own filesystem walk.
//
// Non-fatal: returns nil (not an error) if mainPath can't be read, e.g. the
// merge itself failed or produced nothing.
func listMergedFiles(mainPath string) []string {
	var files []string
	_ = filepath.Walk(mainPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// Skip .git/ directory — same check as publishCodeArtifacts()
		// in aider_runner.go (kept identical on purpose).
		if strings.Contains(path, ".git/") || strings.Contains(path, ".git\\") {
			return nil
		}
		rel, relErr := filepath.Rel(mainPath, path)
		if relErr != nil {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	return files
}

// getLastBlackboardSeq returns the current max sequence_number on the blackboard.
// Used to scope cross-verification to artifacts from the current wave only.
func (r *WorkflowRunner) getLastBlackboardSeq(ctx context.Context, workflowID uuid.UUID) int64 {
	var seq int64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(MAX(sequence_number), 0) FROM blackboard_events WHERE workflow_id = $1`,
		workflowID,
	).Scan(&seq)
	if err != nil {
		r.logger.Warn("runner: getLastBlackboardSeq failed", zap.Error(err))
		return 0
	}
	return seq
}

// saveRunnerState writes runner_state + current_task_cursor to workflows table.
// Called after each task completes for pod-restart recovery.
func (r *WorkflowRunner) saveRunnerState(ctx context.Context, workflowID uuid.UUID, state *runnerState) {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		r.logger.Warn("runner: marshal state failed", zap.Error(err))
		return
	}
	_, err = r.db.Exec(ctx,
		`UPDATE workflows SET runner_state=$1, updated_at=NOW() WHERE id=$2`,
		string(stateJSON), workflowID,
	)
	if err != nil {
		r.logger.Warn("runner: save state failed", zap.Error(err))
	}
}

// loadWorkflowExperts loads expert records with workflow-specific fields.
func (r *WorkflowRunner) loadWorkflowExperts(ctx context.Context, expertIDs []uuid.UUID) ([]workflowExpert, error) {
	var experts []workflowExpert
	for _, id := range expertIDs {
		var e workflowExpert
		var toolsJSON []byte
		err := r.db.QueryRow(ctx,
			`SELECT id, name, domain,
			        COALESCE(reasoning_charter, ''),
			        COALESCE(loop_pattern, 'ota'),
			        COALESCE(max_loop_iterations, 5),
			        COALESCE(allowed_tools, '[]'::jsonb)
			 FROM experts
			 WHERE id=$1 AND is_active=TRUE AND deleted_at IS NULL`,
			id,
		).Scan(&e.ID, &e.Name, &e.Domain, &e.ReasoningCharter,
			&e.LoopPattern, &e.MaxLoopIterations, &toolsJSON)
		if err != nil {
			r.logger.Warn("runner: expert not found", zap.String("id", id.String()))
			continue
		}
		if len(toolsJSON) > 0 {
			_ = json.Unmarshal(toolsJSON, &e.AllowedTools)
		}
		experts = append(experts, e)
	}
	if len(experts) == 0 {
		return nil, fmt.Errorf("no active experts found")
	}
	return experts, nil
}

// loadRequirement reads requirement text from blackboard.
// Falls back to workflow title if no requirement_captured event.
func (r *WorkflowRunner) loadRequirement(ctx context.Context, workflowID uuid.UUID, fallback string) string {
	events, err := r.store.GetByType(ctx, workflowID, []string{"requirement_captured"}, 0)
	if err != nil || len(events) == 0 {
		return fallback
	}
	latest := events[len(events)-1]
	var content struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(latest.Content, &content); err == nil && content.Text != "" {
		return content.Text
	}
	return string(latest.Content)
}

// askDesignGate posts the design approval gate, waits for the client, and
// returns the decision recorded on the approval row.
//
// GateName must be one of approval_requests_gate_check (migration 006):
//
//	'intake','high_level_design','detailed_design','handoff',
//	'budget_exceeded','ad_hoc'
//
// 'detailed_design' is correct here: this gate sits immediately after the
// Detailed Design phase and gates the start of implementation — "Client
// approval gate 3" in DOMAIN_EXPERT_COLLABORATION_DESIGN.md §9.
func (r *WorkflowRunner) askDesignGate(
	ctx context.Context,
	workflowID uuid.UUID,
	attempt int,
) (string, error) {
	summary := "Design phases complete. Read the deliverables, then Approve to start " +
		"implementation — or Request Changes to have the design produced again " +
		"(optionally allowing some generic knowledge)."
	if attempt > 1 {
		summary = fmt.Sprintf("Design re-run #%d complete. %s", attempt, summary)
	}

	if _, err := r.tools.AskClient(ctx, AskClientRequest{
		WorkflowID:   workflowID,
		FromExpertID: uuid.Nil,
		GateName:     "detailed_design",
		Summary:      summary,
	}); err != nil {
		_ = r.engine.Fail(ctx, workflowID, "design AskClient failed")
		return "", fmt.Errorf("design gate: %w", err)
	}

	if err := r.waitForResume(ctx, workflowID); err != nil {
		return "", fmt.Errorf("design gate wait: %w", err)
	}

	decision := r.lastApprovalStatus(ctx, workflowID)
	r.logger.Info("runner: design gate answered",
		zap.String("workflow_id", workflowID.String()),
		zap.Int("attempt", attempt),
		zap.String("decision", decision),
	)
	return decision, nil
}

// lastApprovalStatus returns the status of the most recent approval row for a
// workflow, e.g. "approved" | "changes_requested" | "rejected".
//
// Returns "" on error, which the caller treats as "not a re-run request" —
// failing to read the decision must not trap the workflow in the design loop.
func (r *WorkflowRunner) lastApprovalStatus(ctx context.Context, workflowID uuid.UUID) string {
	var status string
	err := r.db.QueryRow(ctx,
		`SELECT status FROM approval_requests
		 WHERE workflow_id = $1
		 ORDER BY requested_at DESC
		 LIMIT 1`,
		workflowID,
	).Scan(&status)
	if err != nil {
		r.logger.Warn("runner: could not read approval decision", zap.Error(err))
		return ""
	}
	return status
}

// loadGenericAllowancePct reads the client's generic ceiling for a workflow.
//
// Read fresh on every phase attempt (not cached from the workflow loaded at
// Run start) precisely so that a client raising the dial at the gate and
// asking for a re-run takes effect on the next attempt.
//
// Returns 0 on error — the safe default is "trained knowledge only".
func (r *WorkflowRunner) loadGenericAllowancePct(ctx context.Context, workflowID uuid.UUID) float64 {
	var pct float64
	err := r.db.QueryRow(ctx,
		`SELECT generic_allowance_pct FROM workflows WHERE id = $1`,
		workflowID,
	).Scan(&pct)
	if err != nil {
		r.logger.Warn("runner: could not read generic_allowance_pct, defaulting to 0",
			zap.String("workflow_id", workflowID.String()),
			zap.Error(err),
		)
		return 0
	}
	return pct
}

// waitForResume polls workflow status until it transitions to 'running'.
// Returns nil when running, error when cancelled or workflow fails/cancels.
func (r *WorkflowRunner) waitForResume(ctx context.Context, workflowID uuid.UUID) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		wf, err := r.engine.GetByID(ctx, workflowID)
		if err != nil {
			return fmt.Errorf("waitForResume: %w", err)
		}
		switch wf.Status {
		case StatusRunning:
			return nil
		case StatusFailed, StatusCancelled:
			return fmt.Errorf("workflow %s", wf.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// buildPlanContent builds the task_plan_ready event content.
// This is what Projector reads to INSERT workflow_tasks rows.
func buildPlanContent(tasks []TaskSpec, experts []workflowExpert) map[string]interface{} {
	expertNames := make(map[string]string, len(experts))
	for _, e := range experts {
		expertNames[e.ID.String()] = e.Name
	}
	type taskEntry struct {
		ExpertID    string   `json:"expert_id"`
		ExpertName  string   `json:"expert_name"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		DependsOn   []string `json:"depends_on_expert_names"`
	}
	var entries []taskEntry
	for _, t := range tasks {
		var depNames []string
		for _, depID := range t.DependsOnExpertIDs {
			if name, ok := expertNames[depID.String()]; ok {
				depNames = append(depNames, name)
			}
		}
		entries = append(entries, taskEntry{
			ExpertID:    t.ExpertID.String(),
			ExpertName:  expertNames[t.ExpertID.String()],
			Title:       t.Title,
			Description: t.Description,
			DependsOn:   depNames,
		})
	}
	return map[string]interface{}{
		"task_count": len(tasks),
		"tasks":      entries,
	}
}

// watchForChangeRequest polls for pending change requests while the workflow is
// running. When it finds one it:
//   1. Determines which experts are relevant (LLM call — cheap model).
//   2. Marks the request as running.
//   3. Re-runs the design phases for those experts (same loop as the initial
//      design, reusing RestartPhase + executeWaves).
//   4. Presents the same approval gate the initial design uses.
//   5. Marks the request as completed.
//
// This runs as a goroutine alongside the main Run() goroutine. It exits when
// ctx is cancelled or the workflow reaches a terminal state.
//
// WHY POLLING AND NOT REDIS PUB/SUB
//
// Polling the DB every 3 seconds is cheap (one indexed query on status='pending')
// and consistent with waitForResume's polling pattern. The latency (up to 3s)
// is acceptable for a human-initiated action.
//
// WHY NOT BLOCK Run() ON CHANGE REQUESTS
//
// Run() drives the workflow state machine forward. A change request can arrive
// at any point — including while the workflow is paused at an approval gate.
// Blocking Run() would mean the change request is only processed after the
// current gate is resolved. A separate goroutine lets the change request be
// processed immediately.
//
// waves may be nil on the first call (before planning completes). The watcher
// guards against this and uses the full expert list as a fallback.
func (r *WorkflowRunner) watchForChangeRequest(
	ctx context.Context,
	workflowID uuid.UUID,
	allExperts []workflowExpert,
	waves []ExecutionWave,
	crSvc *ChangeRequestService,
) {
	log := r.logger.With(
		zap.String("workflow_id", workflowID.String()),
		zap.String("goroutine", "watchForChangeRequest"),
	)
	log.Info("change request watcher started")

	for {
		select {
		case <-ctx.Done():
			log.Info("change request watcher stopped (context cancelled)")
			return
		case <-time.After(3 * time.Second):
		}

		// Check workflow is still in a live state.
		wf, err := r.engine.GetByID(ctx, workflowID)
		if err != nil {
			log.Warn("watcher: load workflow failed", zap.Error(err))
			continue
		}
		if wf.Status == StatusCompleted || wf.Status == StatusFailed || wf.Status == StatusCancelled {
			log.Info("change request watcher stopped (workflow terminal)",
				zap.String("status", wf.Status))
			return
		}

		// Look for a pending change request.
		cr, err := crSvc.GetPendingChangeRequest(ctx, workflowID)
		if err != nil {
			log.Warn("watcher: get pending change request failed", zap.Error(err))
			continue
		}
		if cr == nil {
			continue // nothing pending
		}

		log.Info("change request detected, starting redesign",
			zap.String("change_request_id", cr.ID.String()),
			zap.String("goal", cr.ChangeGoal),
		)

		// Determine which experts are relevant.
		relevantIDs := DetermineRelevantExperts(
			ctx, r.gateway, workflowID, cr.ChangeGoal, allExperts, r.logger,
		)

		// Mark running before touching the workflow state.
		if err := crSvc.MarkRunning(ctx, cr.ID, relevantIDs); err != nil {
			log.Error("watcher: mark running failed", zap.Error(err))
			continue
		}

		// Build a filtered expert list.
		relevantIDSet := make(map[string]struct{}, len(relevantIDs))
		for _, id := range relevantIDs {
			relevantIDSet[id.String()] = struct{}{}
		}
		var relevantExperts []workflowExpert
		for _, e := range allExperts {
			if _, ok := relevantIDSet[e.ID.String()]; ok {
				relevantExperts = append(relevantExperts, e)
			}
		}
		if len(relevantExperts) == 0 {
			relevantExperts = allExperts
		}

		// Filter waves to only include relevant experts.
		// If waves is nil (planning not yet done) or all waves become empty,
		// fall back to the full wave set.
		var relevantWaves []ExecutionWave
		if waves != nil {
			relevantWaves = filterWavesForExperts(waves, relevantIDs)
		}
		if len(relevantWaves) == 0 && waves != nil {
			log.Warn("watcher: no relevant waves after filtering, using all waves",
				zap.String("change_request_id", cr.ID.String()),
			)
			relevantWaves = waves
		}
		if len(relevantWaves) == 0 {
			// Planning not yet done — build single-wave from relevant experts.
			var singleWave ExecutionWave
			for _, e := range relevantExperts {
				singleWave = append(singleWave, TaskSpec{
					ExpertID:    e.ID,
					Title:       "Redesign for: " + cr.ChangeGoal,
					Description: cr.ChangeGoal,
				})
			}
			relevantWaves = []ExecutionWave{singleWave}
		}

		// Post a blackboard event so the Kanban board shows the redesign starting.
		_, _ = r.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:     workflowID,
			EventType:      "change_request_started",
			PostedByClient: false,
			Content: map[string]interface{}{
				"change_request_id":   cr.ID.String(),
				"change_goal":         cr.ChangeGoal,
				"relevant_expert_ids": relevantIDs,
			},
		})

		// Re-run design phases for relevant experts.
		// WHY RestartPhase and not TransitionPhase: this is a backwards move.
		// TransitionPhase's forward-only validation would reject it.
		// RestartPhase is the named method for legitimate backwards moves.
		if err := r.engine.RestartPhase(ctx, workflowID, PhaseHighLevelDesign); err != nil {
			log.Error("watcher: restart HLD failed", zap.Error(err))
			_ = crSvc.MarkCompleted(ctx, cr.ID)
			continue
		}

		hldState := &runnerState{Phase: PhaseHighLevelDesign}
		if err := r.executeWaves(ctx, workflowID, relevantWaves, relevantExperts, hldState); err != nil {
			log.Error("watcher: HLD waves failed", zap.Error(err))
			if len(hldState.CompletedExpertIDs) == 0 {
				_ = crSvc.MarkCompleted(ctx, cr.ID)
				continue
			}
		}

		if err := r.engine.RestartPhase(ctx, workflowID, PhaseDetailedDesign); err != nil {
			log.Error("watcher: restart DetailedDesign failed", zap.Error(err))
			_ = crSvc.MarkCompleted(ctx, cr.ID)
			continue
		}
		ddState := &runnerState{Phase: PhaseDetailedDesign}
		if err := r.executeWaves(ctx, workflowID, relevantWaves, relevantExperts, ddState); err != nil {
			log.Error("watcher: DetailedDesign waves failed", zap.Error(err))
			if len(ddState.CompletedExpertIDs) == 0 {
				_ = crSvc.MarkCompleted(ctx, cr.ID)
				continue
			}
		}

		// Approval gate — same as the initial design gate.
		summary := fmt.Sprintf(
			"Change request redesign complete.\n\nChange goal: %s\n\n"+
				"%d expert(s) updated their design sections. "+
				"Review the updated deliverables and approve to continue, "+
				"or request further changes.",
			cr.ChangeGoal, len(relevantExperts),
		)
		if _, err := r.tools.AskClient(ctx, AskClientRequest{
			WorkflowID:   workflowID,
			FromExpertID: uuid.Nil,
			GateName:     "detailed_design",
			Summary:      summary,
		}); err != nil {
			log.Error("watcher: AskClient failed", zap.Error(err))
			_ = crSvc.MarkCompleted(ctx, cr.ID)
			continue
		}

		if err := r.waitForResume(ctx, workflowID); err != nil {
			log.Warn("watcher: waitForResume failed", zap.Error(err))
			_ = crSvc.MarkCompleted(ctx, cr.ID)
			continue
		}

		_ = crSvc.MarkCompleted(ctx, cr.ID)

		_, _ = r.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:     workflowID,
			EventType:      "change_request_completed",
			PostedByClient: false,
			Content: map[string]interface{}{
				"change_request_id": cr.ID.String(),
				"change_goal":       cr.ChangeGoal,
			},
		})

		log.Info("change request redesign completed",
			zap.String("change_request_id", cr.ID.String()),
		)
	}
}

// filterWavesForExperts returns a copy of waves containing only tasks whose
// expert is in the relevant set. Waves that become empty after filtering are
// dropped.
func filterWavesForExperts(waves []ExecutionWave, relevantIDs []uuid.UUID) []ExecutionWave {
	relevantSet := make(map[string]struct{}, len(relevantIDs))
	for _, id := range relevantIDs {
		relevantSet[id.String()] = struct{}{}
	}

	var filtered []ExecutionWave
	for _, wave := range waves {
		var tasks []TaskSpec
		for _, t := range wave {
			if _, ok := relevantSet[t.ExpertID.String()]; ok {
				tasks = append(tasks, t)
			}
		}
		if len(tasks) > 0 {
			filtered = append(filtered, tasks)
		}
	}
	return filtered
}

// ResumeOrphanWorkflows finds workflows stuck in 'running' status
// (e.g. after pod restart) and re-launches their runners.
// Called from main.go on startup.
func (r *WorkflowRunner) ResumeOrphanWorkflows(ctx context.Context) {
	rows, err := r.db.Query(ctx,
		`SELECT id FROM workflows WHERE status = $1`,
		StatusRunning,
	)
	if err != nil {
		r.logger.Error("ResumeOrphanWorkflows: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return
	}
	r.logger.Info("ResumeOrphanWorkflows: resuming", zap.Int("count", len(ids)))
	for _, id := range ids {
		go r.Run(ctx, id)
	}
}
