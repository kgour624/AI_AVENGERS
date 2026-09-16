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
	engine *Engine
	store  *blackboard.Store
	redis  *redis.Client
	logger *zap.Logger
}

// NewHandler creates a new workflow handler.
func NewHandler(engine *Engine, store *blackboard.Store, redisClient *redis.Client, logger *zap.Logger) *Handler {
	return &Handler{
		engine: engine,
		store:  store,
		redis:  redisClient,
		logger: logger,
	}
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
// Starts the WorkflowRunner in a background goroutine.
// Returns immediately — client polls GET /kanban or streams GET /kanban/stream.
//
// WHY background goroutine:
//   WorkflowRunner can take minutes (multiple LLM calls per expert).
//   HTTP request must return immediately.
//   Client tracks progress via Kanban SSE stream.
//
// Mental execution:
//   POST /workflows/abc/run
//   1. Verify workflow exists and is in 'running' status
//      (client must call POST /start before /run)
//   2. Launch goroutine: go runner.Run(ctx, workflowID)
//   3. Return 202 Accepted immediately
func (h *Handler) RunWorkflow(runner *WorkflowRunner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
			return
		}

		// Verify workflow exists and is in running status.
		wf, err := h.engine.GetByID(c.Request.Context(), id)
		if err != nil {
			response.NotFound(c, "workflow")
			return
		}
		if wf.Status != StatusRunning {
			response.BadRequest(c, "INVALID_STATUS",
				fmt.Sprintf("workflow must be in 'running' status to run (current: %s). Call POST /start first.", wf.Status))
			return
		}

		// Launch WorkflowRunner in background.
		// WHY context.Background() not c.Request.Context():
		//   Request context is cancelled when HTTP response is sent.
		//   Runner must outlive the HTTP request.
		go runner.Run(context.Background(), id)

		h.logger.Info("workflow runner launched",
			zap.String("workflow_id", id.String()),
		)
		c.JSON(202, map[string]interface{}{
			"status":      "accepted",
			"workflow_id": id.String(),
			"message":     "WorkflowRunner started. Track progress via GET /kanban/stream",
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
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	validDecisions := map[string]bool{
		"approve":                  true,
		"approve_with_notes":       true,
		"request_changes":          true,
		"reject_and_restart_phase": true,
		"cancel_workflow":          true,
	}
	if !validDecisions[req.Decision] {
		response.BadRequest(c, "INVALID_DECISION",
			"decision must be: approve, approve_with_notes, request_changes, reject_and_restart_phase, cancel_workflow")
		return
	}

	ctx := c.Request.Context()

	// Map decision to approval_requests.status
	approvalStatus := req.Decision
	if req.Decision == "reject_and_restart_phase" || req.Decision == "cancel_workflow" {
		approvalStatus = "rejected"
	}

	// Update approval_request row
	clientResponseJSON, _ := json.Marshal(map[string]string{
		"decision": req.Decision,
		"notes":    req.Notes,
	})
	_, err = h.engine.db.Exec(ctx,
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

	// Act on decision
	switch req.Decision {
	case "approve", "approve_with_notes":
		if err := h.engine.Resume(ctx, workflowID); err != nil {
			h.logger.Error("resume workflow failed", zap.Error(err))
			response.InternalError(c)
			return
		}
	case "cancel_workflow":
		if err := h.engine.Fail(ctx, workflowID, "cancelled by client"); err != nil {
			h.logger.Error("cancel workflow failed", zap.Error(err))
		}
		// request_changes and reject_and_restart_phase:
		// workflow stays paused; experts read the client_response event
		// from the blackboard and revise their artifacts.
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
