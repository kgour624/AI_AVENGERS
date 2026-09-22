package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/response"
)

// Handler handles all workflow + approval HTTP endpoints.
type Handler struct {
	engine    *Engine
	store     *blackboard.Store
	redis     *redis.Client
	projector *Projector
	logger    *zap.Logger
}

// NewHandler creates a new workflow handler.
func NewHandler(engine *Engine, store *blackboard.Store, redisClient *redis.Client, projector *Projector, logger *zap.Logger) *Handler {
	return &Handler{
		engine:    engine,
		store:     store,
		redis:     redisClient,
		projector: projector,
		logger:    logger,
	}
}

// ListWorkflows GET /api/v1/workflows
// Returns all workflows for the authenticated client, newest first.
func (h *Handler) ListWorkflows(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)

	rows, err := h.engine.db.Query(c.Request.Context(),
		`SELECT id, client_id, project_id, title, status, current_phase,
		        phase_started_at, phase_completed_at,
		        selected_expert_ids, cost_budget_usd, cost_spent_usd,
		        cost_soft_limit_pct, cost_hard_limit_pct,
		        generic_allowance_pct,
		        created_at, updated_at
		 FROM workflows
		 WHERE client_id = $1
		 ORDER BY updated_at DESC
		 LIMIT 50`,
		clientID,
	)
	if err != nil {
		h.logger.Error("list workflows failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var w Workflow
		var expertIDsRaw []byte
		if err := rows.Scan(
			&w.ID, &w.ClientID, &w.ProjectID, &w.Title, &w.Status, &w.CurrentPhase,
			&w.PhaseStartedAt, &w.PhaseCompletedAt,
			&expertIDsRaw, &w.CostBudgetUSD, &w.CostSpentUSD,
			&w.CostSoftLimitPct, &w.CostHardLimitPct,
		&w.GenericAllowancePct,
			&w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			continue
		}
		if len(expertIDsRaw) > 0 {
			_ = json.Unmarshal(expertIDsRaw, &w.SelectedExpertIDs)
		}
		workflows = append(workflows, w)
	}
	if workflows == nil {
		workflows = []Workflow{}
	}
	response.OK(c, workflows)
}

// CreateWorkflow POST /api/v1/workflows
// Body: {project_id, title, selected_expert_ids, cost_budget_usd?}
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var req struct {
		ProjectID         uuid.UUID   `json:"project_id" binding:"required"`
		Title             string      `json:"title" binding:"required"`
		SelectedExpertIDs []uuid.UUID `json:"selected_expert_ids" binding:"required"`
		CostBudgetUSD     float64     `json:"cost_budget_usd"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	clientID := c.MustGet("user_id").(uuid.UUID)

	w, err := h.engine.Create(c.Request.Context(), CreateRequest{
		ClientID:          clientID,
		ProjectID:         req.ProjectID,
		Title:             req.Title,
		SelectedExpertIDs: req.SelectedExpertIDs,
		CostBudgetUSD:     req.CostBudgetUSD,
	})
	if err != nil {
		h.logger.Error("create workflow failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, w)
}

// GetWorkflow GET /api/v1/workflows/:id
func (h *Handler) GetWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	w, err := h.engine.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "workflow")
		return
	}
	response.OK(c, w)
}

// StartWorkflow POST /api/v1/workflows/:id/start
func (h *Handler) StartWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	w, err := h.engine.Start(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("start workflow failed", zap.Error(err))
		response.BadRequest(c, "CANNOT_START", err.Error())
		return
	}
	response.OK(c, w)
}

// GetBlackboard GET /api/v1/workflows/:id/blackboard?since=0&types=architecture_decision,api_contract_proposed
func (h *Handler) GetBlackboard(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	sinceStr := c.DefaultQuery("since", "0")
	since, _ := strconv.ParseInt(sinceStr, 10, 64)

	// Parse optional event type filter
	typesParam := c.Query("types")
	var eventTypes []string
	if typesParam != "" {
		// comma-separated list
		for _, t := range splitComma(typesParam) {
			if t != "" {
				eventTypes = append(eventTypes, t)
			}
		}
	}

	var events []blackboard.Event
	if len(eventTypes) > 0 {
		events, err = h.store.GetByType(c.Request.Context(), id, eventTypes, since)
	} else {
		events, err = h.store.GetSince(c.Request.Context(), id, since, 200)
	}
	if err != nil {
		h.logger.Error("get blackboard failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if events == nil {
		events = []blackboard.Event{}
	}
	response.OK(c, map[string]interface{}{
		"workflow_id": id,
		"events":      events,
		"count":       len(events),
	})
}

// GetKanban GET /api/v1/workflows/:id/kanban
// Returns workflow_tasks projection for the Kanban board.
func (h *Handler) GetKanban(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	rows, err := h.engine.db.Query(c.Request.Context(),
		`SELECT wt.id, wt.workflow_id, wt.assigned_expert_id,
		        e.name AS expert_name, e.domain,
		        wt.title, wt.description, wt.status,
		        wt.cost_usd, wt.started_at, wt.completed_at, wt.updated_at
		 FROM workflow_tasks wt
		 JOIN experts e ON e.id = wt.assigned_expert_id
		 WHERE wt.workflow_id = $1
		 ORDER BY wt.updated_at DESC`,
		id,
	)
	if err != nil {
		h.logger.Error("get kanban failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type taskRow struct {
		ID               uuid.UUID  `json:"id"`
		WorkflowID       uuid.UUID  `json:"workflow_id"`
		AssignedExpertID uuid.UUID  `json:"assigned_expert_id"`
		ExpertName       string     `json:"expert_name"`
		Domain           string     `json:"domain"`
		Title            string     `json:"title"`
		Description      string     `json:"description"`
		Status           string     `json:"status"`
		CostUSD          float64    `json:"cost_usd"`
		StartedAt        *string    `json:"started_at"`
		CompletedAt      *string    `json:"completed_at"`
		UpdatedAt        string     `json:"updated_at"`
	}

	var tasks []taskRow
	for rows.Next() {
		var t taskRow
		if err := rows.Scan(
			&t.ID, &t.WorkflowID, &t.AssignedExpertID,
			&t.ExpertName, &t.Domain,
			&t.Title, &t.Description, &t.Status,
			&t.CostUSD, &t.StartedAt, &t.CompletedAt, &t.UpdatedAt,
		); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []taskRow{}
	}
	response.OK(c, map[string]interface{}{
		"workflow_id": id,
		"tasks":       tasks,
	})
}

// RunWorkflow POST /api/v1/workflows/:id/run
// Starts WorkflowRunner + Projector in background goroutines.
// Returns 202 immediately.
//
// Mental execution:
//   POST /workflows/abc/run
//   1. Verify workflow exists + status=running
//   2. go projector.Run(ctx, workflowID)  <- projects blackboard events to workflow_tasks
//   3. go runner.Run(ctx, workflowID)     <- drives workflow to completion
//   4. Return 202 Accepted
//
// WHY Projector starts first:
//   Runner will post task_plan_ready event almost immediately.
//   Projector must be subscribed before that event arrives.
//   Starting Projector first ensures no events are missed.
func (h *Handler) RunWorkflow(runner *WorkflowRunner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
			return
		}
		wf, err := h.engine.GetByID(c.Request.Context(), id)
		if err != nil {
			response.NotFound(c, "workflow")
			return
		}
		if wf.Status != StatusRunning {
			response.BadRequest(c, "INVALID_STATUS",
				fmt.Sprintf("workflow must be 'running' (current: %s). Call POST /start first.", wf.Status))
			return
		}

		// WHY context.Background(): both goroutines must outlive the HTTP request.
		runCtx := context.Background()

		// Start Projector FIRST: must be subscribed before Runner posts events.
		go h.projector.Run(runCtx, id)

		// Start Runner: drives workflow to completion.
		go runner.Run(runCtx, id)

		h.logger.Info("workflow runner + projector launched",
			zap.String("workflow_id", id.String()),
		)
		c.JSON(202, map[string]interface{}{
			"status":      "accepted",
			"workflow_id": id.String(),
			"message":     "WorkflowRunner started. Track via GET /kanban/stream",
		})
	}
}
// Body: {decision, notes?}
// decision: approve | approve_with_notes | request_changes | reject_and_restart_phase | cancel_workflow
func (h *Handler) RespondToApproval(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	approvalID, err := uuid.Parse(c.Param("aid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid approval ID")
		return
	}

	var req struct {
		Decision string `json:"decision" binding:"required"`
		Notes    string `json:"notes"`
		// GenericAllowancePct: optional. Sent with "request_changes" when the
		// client is not satisfied with a trained-knowledge-only design and
		// wants the design produced again with a bounded generic allowance.
		// Pointer so "not sent" is distinguishable from an explicit 0.
		GenericAllowancePct *float64 `json:"generic_allowance_pct"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	// Validate the dial before anything else. The DB has the same 0-30 CHECK
	// (workflows_generic_allowance_pct_check, migration 017); this gives the
	// client a clear 400 instead of a 500 from a constraint violation.
	if req.GenericAllowancePct != nil {
		pct := *req.GenericAllowancePct
		if pct < 0 || pct > MaxGenericAllowancePct {
			response.BadRequest(c, "INVALID_GENERIC_ALLOWANCE",
				fmt.Sprintf("generic_allowance_pct must be between 0 and %.0f", MaxGenericAllowancePct))
			return
		}
	}

	// decisionToStatus maps the client's decision (API vocabulary) to the
	// approval_requests.status value (DB vocabulary). These two vocabularies
	// are NOT the same and must be translated explicitly.
	//
	// The DB values are fixed by approval_requests_status_check (migration 006):
	//   'pending','approved','approved_with_notes','changes_requested',
	//   'rejected','cancelled'
	//
	// BUG FIX: this used to be a map[string]bool for validation plus a
	// separate `approvalStatus := req.Decision` with an if-statement that
	// only remapped reject_and_restart_phase and cancel_workflow. The other
	// three decisions were written to the DB verbatim:
	//   "approve"            -> invalid (DB wants 'approved')
	//   "approve_with_notes" -> invalid (DB wants 'approved_with_notes')
	//   "request_changes"    -> invalid (DB wants 'changes_requested')
	// so the UPDATE violated the CHECK constraint and the handler returned
	// HTTP 500. Clicking "Approve" in the Kanban approval gate could never
	// succeed — 3 of the 5 decisions were permanently broken.
	//
	// WHY one map instead of two lists: the root cause was two parallel
	// lists of decisions that had to be kept in sync and weren't. A single
	// map makes the valid set and the DB translation the same source of
	// truth, so adding a decision cannot silently produce an invalid status.
	decisionToStatus := map[string]string{
		"approve":                  "approved",
		"approve_with_notes":       "approved_with_notes",
		"request_changes":          "changes_requested",
		"reject_and_restart_phase": "rejected",
		"cancel_workflow":          "cancelled",
	}
	approvalStatus, ok := decisionToStatus[req.Decision]
	if !ok {
		response.BadRequest(c, "INVALID_DECISION",
			"decision must be: approve, approve_with_notes, request_changes, reject_and_restart_phase, cancel_workflow")
		return
	}

	ctx := c.Request.Context()

	// Update approval_request row
	clientResponseJSON, _ := json.Marshal(map[string]string{
		"decision": req.Decision,
		"notes":    req.Notes,
	})
	tag, err := h.engine.db.Exec(ctx,
		`UPDATE approval_requests SET
			status = $1,
			client_response = $2,
			responded_at = NOW()
		 WHERE id = $3 AND workflow_id = $4 AND status = 'pending'`,
		approvalStatus, string(clientResponseJSON), approvalID, workflowID,
	)
	if err != nil {
		h.logger.Error("update approval failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// No row updated means this approval is not pending: either it was
	// already responded to, or the id does not belong to this workflow.
	// Stop here — do NOT post client_response or resume.
	//
	// WHY this guard matters (observed in production): the UPDATE's
	// "AND status = 'pending'" makes a repeat response a no-op, but the
	// old code ignored the result and still returned 200 AND called
	// Resume(). The approval gate stays on screen until the 10s workflow
	// poll refreshes, so a user clicking Approve a few times produced
	// several 200s and several Resume() calls for one approval (6 were
	// seen in one 3-second window). A late duplicate is worse than noise:
	// if the workflow has meanwhile paused at the NEXT gate, that stray
	// Resume() un-pauses it and the client never gets to approve that
	// gate — a human-in-the-loop checkpoint is skipped.
	if tag.RowsAffected() == 0 {
		h.logger.Info("approval already responded or not pending",
			zap.String("workflow_id", workflowID.String()),
			zap.String("approval_id", approvalID.String()),
			zap.String("decision", req.Decision),
		)
		response.Conflict(c, "this approval has already been responded to")
		return
	}

	// Post client_response event on blackboard
	_, _ = h.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "client_response",
		PostedByClient: true,
		Content: map[string]string{
			"decision":    req.Decision,
			"notes":       req.Notes,
			"approval_id": approvalID.String(),
		},
	})

	// Persist the generic dial BEFORE resuming, so the runner's next phase
	// attempt reads the new value (it re-reads the column per attempt).
	if req.GenericAllowancePct != nil {
		if _, err := h.engine.db.Exec(ctx,
			`UPDATE workflows SET generic_allowance_pct = $1, updated_at = NOW() WHERE id = $2`,
			*req.GenericAllowancePct, workflowID,
		); err != nil {
			h.logger.Error("update generic_allowance_pct failed", zap.Error(err))
			response.InternalError(c)
			return
		}
		h.logger.Info("generic allowance set by client",
			zap.String("workflow_id", workflowID.String()),
			zap.Float64("generic_allowance_pct", *req.GenericAllowancePct),
		)
	}

	// Act on decision
	switch req.Decision {
	case "approve", "approve_with_notes":
		if err := h.engine.Resume(ctx, workflowID); err != nil {
			h.logger.Error("resume workflow failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	case "request_changes":
		// Resume so the runner wakes up and re-runs the design phases with the
		// dial the client just set.
		//
		// WHY this changed: previously request_changes left the workflow
		// paused forever with a comment saying "experts read the
		// client_response event and revise" — nothing did that, so the
		// workflow was simply stuck. The runner now owns the re-run
		// (see WorkflowRunner.Run's design loop), and it can only act once
		// the workflow is running again.
		if err := h.engine.Resume(ctx, workflowID); err != nil {
			h.logger.Error("resume for re-run failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	case "cancel_workflow":
		if err := h.engine.Fail(ctx, workflowID, "cancelled by client"); err != nil {
			h.logger.Error("cancel workflow failed", zap.Error(err))
		}
		// reject_and_restart_phase: workflow stays paused. Unlike
		// request_changes there is no implemented restart semantic for it yet,
		// and resuming without one would silently move the workflow forward.
	}

	h.logger.Info("approval responded",
		zap.String("workflow_id", workflowID.String()),
		zap.String("approval_id", approvalID.String()),
		zap.String("decision", req.Decision),
	)
	response.OK(c, map[string]string{
		"status":   "ok",
		"decision": req.Decision,
	})
}

// splitComma splits a comma-separated string into a slice.
func splitComma(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// CancelWorkflow handles POST /workflows/:id/cancel
func (h *Handler) CancelWorkflow(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	wfID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow id")
		return
	}
	if err := h.engine.Cancel(c.Request.Context(), wfID, clientID); err != nil {
		h.logger.Error("cancel workflow failed", zap.Error(err))
		response.BadRequest(c, "CANCEL_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"status": "cancelled"})
}

// RetryTask handles POST /workflows/:id/tasks/:taskId/retry
// Resets a failed task to 'todo' status so the workflow runner can retry it.
// WHY this exists: A single task failure (e.g., LLM timeout, rate limit) can
// block the entire workflow. This endpoint lets the client manually retry
// without restarting the whole workflow.
func (h *Handler) RetryTask(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	wfID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow id")
		return
	}
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid task id")
		return
	}

	// Verify workflow ownership
	var ownerID uuid.UUID
	if err := h.engine.db.QueryRow(c.Request.Context(),
		`SELECT client_id FROM workflows WHERE id = $1`, wfID,
	).Scan(&ownerID); err != nil {
		response.NotFound(c, "workflow")
		return
	}
	if ownerID != clientID {
		response.Forbidden(c, "not your workflow")
		return
	}

	// Reset task status to 'todo' and clear timestamps
	// WHY clear completed_at: the task is no longer complete
	// WHY keep started_at: preserves audit trail of first attempt
	tag, err := h.engine.db.Exec(c.Request.Context(),
		`UPDATE workflow_tasks SET
			status = 'todo',
			completed_at = NULL,
			updated_at = NOW()
		 WHERE id = $1 AND workflow_id = $2 AND status IN ('failed', 'blocked')`,
		taskID, wfID,
	)
	if err != nil {
		h.logger.Error("retry task failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	if tag.RowsAffected() == 0 {
		response.BadRequest(c, "RETRY_FAILED", "task not found or not in failed/blocked state")
		return
	}

	// Post task_retry_requested event on blackboard so the workflow runner
	// knows to pick up this task again
	_, _ = h.store.Post(c.Request.Context(), blackboard.PostRequest{
		WorkflowID:     wfID,
		EventType:      "task_retry_requested",
		PostedByClient: true,
		Content: map[string]string{
			"task_id": taskID.String(),
		},
	})

	h.logger.Info("task retry requested",
		zap.String("workflow_id", wfID.String()),
		zap.String("task_id", taskID.String()),
	)
	response.OK(c, gin.H{"status": "task reset to todo"})
}
