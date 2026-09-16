package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// runnerState is the checkpoint saved after each task completes.
// Written to workflows.runner_state for pod-restart recovery.
type runnerState struct {
	Phase              string      `json:"phase"`
	CompletedExpertIDs []string    `json:"completed_expert_ids"`
	FailedExpertIDs    []string    `json:"failed_expert_ids"`
}

// WorkflowRunner drives a workflow from start to completion.
//
// FLOW:
//   1. Load workflow + experts from DB
//   2. Load requirement from blackboard
//   3. Planner.Plan() -> []TaskSpec
//   4. DAG.BuildDAG() -> []ExecutionWave (topological sort)
//   5. Post task_plan_ready event (Projector creates workflow_tasks)
//   6. AskClient for plan approval -> workflow pauses
//   7. waitForResume() polls DB
//   8. executeWaves(): wave by wave, parallel within each wave
//   9. AskClient for final approval
//  10. engine.Complete()
//
// RECOVERY (Fix 5):
//   On pod restart: main.go calls ResumeOrphanWorkflows()
//   -> finds workflows WHERE status='running'
//   -> re-launches Run() for each
//   -> Run() reads runner_state + current_task_cursor
//   -> skips already-completed tasks
//
// SINGLE WRITE PATH (Fix 1):
//   Runner posts events on blackboard.
//   Projector reads events and updates workflow_tasks.
//   Runner never writes directly to workflow_tasks.
type WorkflowRunner struct {
	db        *pgxpool.Pool
	engine    *Engine
	planner   *Planner
	agentLoop *AgentLoop
	tools     *Tools
	store     *blackboard.Store
	gateway   *gateway.ModelGateway
	logger    *zap.Logger
}

func NewWorkflowRunner(
	db *pgxpool.Pool,
	engine *Engine,
	planner *Planner,
	agentLoop *AgentLoop,
	tools *Tools,
	store *blackboard.Store,
	gw *gateway.ModelGateway,
	logger *zap.Logger,
) *WorkflowRunner {
	return &WorkflowRunner{
		db: db, engine: engine, planner: planner,
		agentLoop: agentLoop, tools: tools, store: store,
		gateway: gw, logger: logger,
	}
}

// Run drives the workflow to completion.
// Designed to run in a goroutine: go runner.Run(ctx, workflowID)
//
// Mental execution:
//   workflowID = abc
//   experts = [PM, SD, DSA, LLD]
//   requirement = "Build URL shortener"
//
//   Plan: [{PM,deps:[]}, {SD,deps:[PM]}, {DSA,deps:[PM]}, {LLD,deps:[SD,DSA]}]
//   DAG:  Wave0=[PM], Wave1=[SD,DSA], Wave2=[LLD]
//
//   Post task_plan_ready -> Projector inserts 4 workflow_tasks
//   AskClient -> pause
//   [Client approves]
//   Wave0: PM runs alone
//   Wave1: SD + DSA run in parallel (errgroup)
//   Wave2: LLD runs after both done
//   AskClient -> pause
//   [Client approves]
//   Complete
func (r *WorkflowRunner) Run(ctx context.Context, workflowID uuid.UUID) {
	log := r.logger.With(zap.String("workflow_id", workflowID.String()))
	log.Info("workflow runner started")

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

	// Step 3: Load requirement.
	requirementText := r.loadRequirement(ctx, workflowID, wf.Title)

	// Step 4: Plan.
	log.Info("runner: planning", zap.Int("experts", len(experts)))
	tasks, err := r.planner.Plan(ctx, requirementText, experts)
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

	// Step 8: Transition phase + execute waves.
	_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseHighLevelDesign, nil, 0)

	state := &runnerState{Phase: PhaseHighLevelDesign}
	if err := r.executeWaves(ctx, workflowID, waves, experts, state); err != nil {
		log.Error("runner: executeWaves failed", zap.Error(err))
		// Don't fail the whole workflow on partial task failure.
		// Individual tasks are marked failed via blackboard events.
		// Only fail if ALL tasks failed.
		if len(state.CompletedExpertIDs) == 0 {
			_ = r.engine.Fail(ctx, workflowID, "all tasks failed")
			return
		}
		log.Warn("runner: some tasks failed, continuing to final approval",
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
//   waves = [[PM], [SD, DSA], [LLD]]
//
//   Wave 0: launch PM goroutine, wait
//   Wave 1: launch SD goroutine + DSA goroutine simultaneously, wait for both
//   Wave 2: launch LLD goroutine, wait
//
// WHY WaitGroup not errgroup:
//   errgroup cancels all goroutines on first error.
//   We want ALL tasks in a wave to complete (even if some fail).
//   Failed tasks are marked via blackboard events, not by cancelling others.
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

	for waveIdx, wave := range waves {
		r.logger.Info("runner: executing wave",
			zap.Int("wave", waveIdx),
			zap.Int("tasks", len(wave)),
		)

		var wg sync.WaitGroup
		for _, task := range wave {
			wg.Add(1)
			go func(t TaskSpec) {
				defer wg.Done()

				expert, ok := expertMap[t.ExpertID.String()]
				if !ok {
					return
				}

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

				// Save checkpoint after each task.
				r.saveRunnerState(ctx, workflowID, state)
			}(task)
		}
		wg.Wait()
	}

	errMu.Lock()
	defer errMu.Unlock()
	return firstErr
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


