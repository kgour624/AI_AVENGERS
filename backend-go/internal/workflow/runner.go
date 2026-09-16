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

// WorkflowRunner drives a workflow from start to completion.
//
// FLOW:
//   1. Load workflow + experts from DB
//   2. Load requirement text from blackboard (requirement_captured event)
//   3. Planner.Plan() → []TaskSpec (LLM decides who does what)
//   4. INSERT workflow_tasks rows
//   5. Post task_plan_ready on blackboard
//   6. AskClient for plan approval → workflow pauses
//   7. [Client approves via POST /approvals/:id/respond]
//   8. Execute tasks in dependency order (parallel where possible)
//   9. AskClient for final approval
//  10. [Client approves → workflow.Complete()]
//
// WHY WorkflowRunner is a separate goroutine:
//   HTTP handler returns immediately after creating the workflow.
//   Runner runs in background, posts SSE events via blackboard.
//   Client polls GET /kanban or subscribes to GET /kanban/stream.
//
// RESUME:
//   If the process restarts mid-workflow, the runner can be re-launched
//   with the same workflowID. It reads the latest checkpoint + blackboard
//   to determine where to resume.
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

// NewWorkflowRunner creates a new WorkflowRunner.
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
		db:        db,
		engine:    engine,
		planner:   planner,
		agentLoop: agentLoop,
		tools:     tools,
		store:     store,
		gateway:   gw,
		logger:    logger,
	}
}

// Run drives the workflow to completion.
// Designed to run in a goroutine: go runner.Run(ctx, workflowID)
//
// Mental execution:
//   workflowID = abc
//   1. Load workflow: {selected_expert_ids: [uuid1, uuid2, uuid3]}
//   2. Load experts: [{SystemDesign, charter}, {DSA, charter}, {LLD, charter}]
//   3. Load requirement: "Build a URL shortener with analytics"
//   4. Plan: [{uuid1, "Design architecture", deps:[]},
//             {uuid2, "Design hash function", deps:[uuid1]},
//             {uuid3, "Design class diagram", deps:[uuid1, uuid2]}]
//   5. INSERT 3 workflow_tasks rows (status=todo)
//   6. Post task_plan_ready event
//   7. AskClient → workflow pauses
//   8. [Client approves]
//   9. Execute:
//      - uuid1 (no deps) → goroutine starts immediately
//      - uuid2 (deps: uuid1) → waits for uuid1 done, then starts
//      - uuid3 (deps: uuid1, uuid2) → waits for both, then starts
//  10. All done → AskClient for final approval
func (r *WorkflowRunner) Run(ctx context.Context, workflowID uuid.UUID) {
	log := r.logger.With(zap.String("workflow_id", workflowID.String()))
	log.Info("workflow runner started")

	// Step 1: Load workflow.
	wf, err := r.engine.GetByID(ctx, workflowID)
	if err != nil {
		log.Error("runner: load workflow failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "load workflow failed: "+err.Error())
		return
	}

	// Step 2: Load workflowExperts.
	experts, err := r.loadWorkflowExperts(ctx, wf.SelectedExpertIDs)
	if err != nil || len(experts) == 0 {
		log.Error("runner: load experts failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "load experts failed")
		return
	}

	// Step 3: Load requirement text.
	// First try: find requirement_captured event on blackboard.
	// Fallback: use workflow title.
	requirementText := r.loadRequirement(ctx, workflowID, wf.Title)

	// Step 4: Plan.
	log.Info("runner: planning tasks", zap.Int("experts", len(experts)))
	tasks, err := r.planner.Plan(ctx, requirementText, experts)
	if err != nil {
		log.Error("runner: planning failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "planning failed: "+err.Error())
		return
	}

	// Step 5: INSERT workflow_tasks rows.
	taskIDs, err := r.insertTasks(ctx, workflowID, tasks)
	if err != nil {
		log.Error("runner: insert tasks failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "insert tasks failed: "+err.Error())
		return
	}

	// Step 6: Post task_plan_ready event.
	planContent := buildPlanContent(tasks, experts)
	_, _ = r.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "task_plan_ready",
		PostedByClient: true, // system event
		Content:        planContent,
	})

	// Step 7: AskClient for plan approval.
	log.Info("runner: requesting plan approval")
	_, err = r.tools.AskClient(ctx, AskClientRequest{
		WorkflowID:      workflowID,
		FromExpertID:    uuid.Nil, // system-generated
		GateName:        "intake",
		Summary:         fmt.Sprintf("Task plan ready: %d tasks for %d experts. Please review and approve.", len(tasks), len(experts)),
		ArtifactContent: planContent,
	})
	if err != nil {
		log.Error("runner: AskClient failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "AskClient failed: "+err.Error())
		return
	}

	// Workflow is now paused. Runner waits for client approval.
	// When client calls POST /approvals/:id/respond with "approve",
	// handler calls engine.Resume() which sets status=running.
	// Runner polls for resume.
	log.Info("runner: waiting for plan approval")
	if err := r.waitForResume(ctx, workflowID); err != nil {
		log.Warn("runner: wait for resume failed or cancelled", zap.Error(err))
		return
	}
	log.Info("runner: plan approved, executing tasks")

	// Step 8: Transition to high_level_design phase.
	_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseHighLevelDesign, nil, 0)

	// Step 9: Execute tasks in dependency order.
	if err := r.executeTasks(ctx, workflowID, tasks, taskIDs, experts); err != nil {
		log.Error("runner: task execution failed", zap.Error(err))
		_ = r.engine.Fail(ctx, workflowID, "task execution failed: "+err.Error())
		return
	}

	// Step 10: Transition to handoff phase + final approval.
	_, _ = r.engine.TransitionPhase(ctx, workflowID, PhaseHandoff, nil, 0)

	// Collect all artifact event IDs for the final approval summary.
	allArtifacts, _ := r.store.GetByType(ctx, workflowID,
		[]string{"architecture_decision", "data_model_proposed", "api_contract_proposed",
			"module_design_proposed", "code_artifact_produced", "requirement_captured"},
		0,
	)
	artifactIDs := make([]uuid.UUID, 0, len(allArtifacts))
	for _, ev := range allArtifacts {
		artifactIDs = append(artifactIDs, ev.ID)
	}

	_, err = r.tools.AskClient(ctx, AskClientRequest{
		WorkflowID:     workflowID,
		FromExpertID:   uuid.Nil,
		GateName:       "handoff",
		Summary:        fmt.Sprintf("All %d tasks completed. %d artifacts produced. Please review and approve final handoff.", len(tasks), len(allArtifacts)),
		CitedEventIDs:  artifactIDs,
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

// executeTasks runs all tasks in dependency order.
// Tasks with no unmet dependencies run in parallel.
// Tasks with dependencies wait for their deps to complete.
//
// Mental execution:
//   tasks = [{uuid1, deps:[]}, {uuid2, deps:[uuid1]}, {uuid3, deps:[uuid1,uuid2]}]
//
//   Round 1: uuid1 has no deps → start goroutine
//   uuid1 completes → signal done
//   Round 2: uuid2 deps=[uuid1] all done → start goroutine
//   uuid2 completes → signal done
//   Round 3: uuid3 deps=[uuid1,uuid2] all done → start goroutine
//   uuid3 completes → all done
func (r *WorkflowRunner) executeTasks(
	ctx context.Context,
	workflowID uuid.UUID,
	tasks []TaskSpec,
	taskIDs map[string]uuid.UUID, // expert_id → task_id
	experts []workflowExpert,
) error {
	// Build expert lookup: expert_id → workflowExpert
	expertMap := make(map[string]workflowExpert, len(experts))
	for _, e := range experts {
		expertMap[e.ID.String()] = e
	}

	// doneCh: expert_id → channel that closes when that expert's task is done.
	// WHY channel not WaitGroup: allows per-expert waiting (not all-or-nothing).
	doneCh := make(map[string]chan struct{}, len(tasks))
	for _, t := range tasks {
		ch := make(chan struct{})
		doneCh[t.ExpertID.String()] = ch
	}

	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for _, task := range tasks {
		wg.Add(1)
		go func(t TaskSpec) {
			defer wg.Done()
			defer close(doneCh[t.ExpertID.String()])

			// Wait for all dependencies to complete.
			for _, depID := range t.DependsOnExpertIDs {
				ch, ok := doneCh[depID.String()]
				if !ok {
					continue // unknown dep — skip
				}
				select {
				case <-ch:
					// dep done
				case <-ctx.Done():
					return
				}
			}

			expert, ok := expertMap[t.ExpertID.String()]
			if !ok {
				return
			}
			taskID, ok := taskIDs[t.ExpertID.String()]
			if !ok {
				return
			}

			_, err := r.agentLoop.Run(ctx, AgentLoopRequest{
				WorkflowID:      workflowID,
				Expert:          expert,
				TaskID:          taskID,
				TaskTitle:       t.Title,
				TaskDescription: t.Description,
			})
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				r.logger.Error("runner: agent loop failed",
					zap.String("expert", expert.Name),
					zap.Error(err),
				)
			}
		}(task)
	}

	wg.Wait()

	errMu.Lock()
	defer errMu.Unlock()
	return firstErr
}

// loadWorkflowExperts loads expert records with workflow-specific fields.
// Reads: id, name, domain, reasoning_charter, loop_pattern,
//        max_loop_iterations, allowed_tools.
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
			 WHERE id = $1
			   AND is_active = TRUE
			   AND deleted_at IS NULL`,
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

// loadRequirement reads the requirement text from the blackboard.
// Falls back to workflow title if no requirement_captured event exists.
func (r *WorkflowRunner) loadRequirement(ctx context.Context, workflowID uuid.UUID, fallback string) string {
	events, err := r.store.GetByType(ctx, workflowID, []string{"requirement_captured"}, 0)
	if err != nil || len(events) == 0 {
		return fallback
	}
	// Use the most recent requirement_captured event.
	latest := events[len(events)-1]
	// Try to extract "text" field from content JSON.
	var content struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(latest.Content, &content); err == nil && content.Text != "" {
		return content.Text
	}
	// Fallback: use raw content as string.
	return string(latest.Content)
}

// insertTasks creates workflow_tasks rows for each TaskSpec.
// Returns map: expert_id string → task_id UUID.
func (r *WorkflowRunner) insertTasks(ctx context.Context, workflowID uuid.UUID, tasks []TaskSpec) (map[string]uuid.UUID, error) {
	taskIDs := make(map[string]uuid.UUID, len(tasks))
	for _, t := range tasks {
		var taskID uuid.UUID
		err := r.db.QueryRow(ctx,
			`INSERT INTO workflow_tasks
				(workflow_id, assigned_expert_id, title, description, status)
			 VALUES ($1, $2, $3, $4, 'todo')
			 RETURNING id`,
			workflowID, t.ExpertID, t.Title, t.Description,
		).Scan(&taskID)
		if err != nil {
			return nil, fmt.Errorf("insert task for expert %s: %w", t.ExpertID, err)
		}
		taskIDs[t.ExpertID.String()] = taskID
	}
	return taskIDs, nil
}

// waitForResume polls the workflow status until it transitions back to "running".
// Returns nil when running, error when context is cancelled or workflow fails.
//
// WHY polling not pub/sub:
//   Approval response comes via HTTP (POST /approvals/:id/respond).
//   That handler calls engine.Resume() which sets status=running in DB.
//   Polling is simple and correct. Interval: 2s. Max wait: context timeout.
func (r *WorkflowRunner) waitForResume(ctx context.Context, workflowID uuid.UUID) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wf, err := r.engine.GetByID(ctx, workflowID)
		if err != nil {
			return fmt.Errorf("waitForResume: load failed: %w", err)
		}
		switch wf.Status {
		case StatusRunning:
			return nil // resumed
		case StatusFailed, StatusCancelled:
			return fmt.Errorf("waitForResume: workflow %s", wf.Status)
		}
		// Still paused — wait 2s and check again.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sleepCh(2):
		}
	}
}

// buildPlanContent builds the task_plan_ready event content.
func buildPlanContent(tasks []TaskSpec, experts []workflowExpert) map[string]interface{} {
	expertNames := make(map[string]string, len(experts))
	for _, e := range experts {
		expertNames[e.ID.String()] = e.Name
	}

	type taskSummary struct {
		ExpertID   string   `json:"expert_id"`
		ExpertName string   `json:"expert_name"`
		Title      string   `json:"title"`
		DependsOn  []string `json:"depends_on_expert_names"`
	}
	var summaries []taskSummary
	for _, t := range tasks {
		var depNames []string
		for _, depID := range t.DependsOnExpertIDs {
			if name, ok := expertNames[depID.String()]; ok {
				depNames = append(depNames, name)
			}
		}
		summaries = append(summaries, taskSummary{
			ExpertID:   t.ExpertID.String(),
			ExpertName: expertNames[t.ExpertID.String()],
			Title:      t.Title,
			DependsOn:  depNames,
		})
	}
	return map[string]interface{}{
		"task_count": len(tasks),
		"tasks":      summaries,
	}
}

// sleepCh returns a channel that closes after n seconds.
// Used by waitForResume to avoid blocking the goroutine with time.Sleep.
func sleepCh(seconds int) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(seconds) * time.Second)
		close(ch)
	}()
	return ch
}
