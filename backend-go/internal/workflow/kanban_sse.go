package workflow

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/response"
)

// SSE event type constants for the Kanban stream.
const (
	SSEKanbanTask     = "kanban_task"     // workflow_task status changed
	SSEKanbanArtifact = "kanban_artifact" // expert posted an artifact
	SSEKanbanPlan     = "kanban_plan"     // task plan ready for approval
	SSEKanbanApproval = "kanban_approval" // approval gate opened
	SSEKanbanDone     = "kanban_done"     // workflow completed or failed
	SSEKanbanError    = "kanban_error"    // stream error
)

// StreamKanban handles GET /api/v1/workflows/:id/kanban/stream
// Streams live Kanban updates via SSE.
//
// DESIGN:
//   Subscribes to the workflow's blackboard Redis channel.
//   Translates blackboard events into SSE events for the frontend.
//   Closes when: workflow completes/fails, client disconnects, or context cancelled.
//
// Mental execution:
//   Client connects to /workflows/abc/kanban/stream
//   1. Verify workflow exists
//   2. Send current Kanban state (snapshot)
//   3. Subscribe to blackboard Redis channel
//   4. For each new blackboard event:
//      - task_plan_ready → SSE kanban_plan
//      - artifact events → SSE kanban_artifact
//      - workflow_completed → SSE kanban_done + close
//   5. Client disconnects → ctx.Done() → subscriber stops
func (h *Handler) StreamKanban(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	// Verify workflow exists.
	_, err = h.engine.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		response.NotFound(c, "workflow")
		return
	}

	// Set SSE headers.
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create subscriber for this workflow.
	// Each client gets its own subscriber — Redis fan-out handles multiple clients.
	sub := blackboard.NewSubscriber(h.store, h.redis, h.logger)

	// Subscribe from sequence 0 to get all events (including historical).
	// WHY from 0: client may have just connected and needs full history.
	// Frontend deduplicates by sequence_number if needed.
	eventCh, errCh := sub.Subscribe(c.Request.Context(), workflowID, uuid.Nil, 0)

	// ✅ FIX: Write initial ping BEFORE entering stream loop.
	// WHY: Go/Gin doesn't send 200 OK headers until first write.
	// Without this, browser stays in CONNECTING state if no events exist.
	// SSE comment format: lines starting with ':' are ignored by EventSource.
	c.Writer.Write([]byte(": connected\n\n"))
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed — subscriber stopped.
				sendKanbanSSE(w, SSEKanbanDone, map[string]string{"reason": "stream_ended"})
				return false
			}
			h.handleKanbanEvent(w, event)
			// Check if workflow is done.
			if event.EventType == "workflow_completed" || event.EventType == "workflow_failed" {
				sendKanbanSSE(w, SSEKanbanDone, map[string]interface{}{
					"event_type": event.EventType,
					"workflow_id": workflowID.String(),
				})
				return false
			}
			return true

		case err := <-errCh:
			if err != nil {
				h.logger.Warn("kanban stream: subscriber error",
					zap.String("workflow_id", workflowID.String()),
					zap.Error(err),
				)
				sendKanbanSSE(w, SSEKanbanError, map[string]string{"message": err.Error()})
			}
			return false

		case <-c.Request.Context().Done():
			// Client disconnected.
			return false
		}
	})
}

// handleKanbanEvent translates a blackboard event into an SSE event.
//
// Event type mapping:
//   task_plan_ready          → kanban_plan
//   architecture_decision    → kanban_artifact
//   data_model_proposed      → kanban_artifact
//   api_contract_proposed    → kanban_artifact
//   module_design_proposed   → kanban_artifact
//   code_artifact_produced   → kanban_artifact
//   test_case_proposed       → kanban_artifact
//   requirement_captured     → kanban_artifact
//   question_to_client       → kanban_approval
//   (all others)             → kanban_task (generic update)
func (h *Handler) handleKanbanEvent(w io.Writer, event blackboard.Event) {
	artifactTypes := map[string]bool{
		"architecture_decision":  true,
		"data_model_proposed":    true,
		"api_contract_proposed":  true,
		"module_design_proposed": true,
		"code_artifact_produced": true,
		"test_case_proposed":     true,
		"requirement_captured":   true,
	}

	payload := map[string]interface{}{
		"sequence_number":    event.SequenceNumber,
		"event_type":         event.EventType,
		"posted_by_expert_id": event.PostedByExpertID,
		"posted_at":          event.PostedAt,
		"content":            json.RawMessage(event.Content),
	}

	switch {
	case event.EventType == "task_plan_ready":
		sendKanbanSSE(w, SSEKanbanPlan, payload)
	case artifactTypes[event.EventType]:
		sendKanbanSSE(w, SSEKanbanArtifact, payload)
	case event.EventType == "question_to_client":
		sendKanbanSSE(w, SSEKanbanApproval, payload)
	default:
		sendKanbanSSE(w, SSEKanbanTask, payload)
	}
}

// sendKanbanSSE writes a single SSE event to the response writer.
func sendKanbanSSE(w io.Writer, eventType string, data interface{}) {
	var payload string
	if data != nil {
		b, err := json.Marshal(map[string]interface{}{
			"type": eventType,
			"data": data,
		})
		if err == nil {
			payload = string(b)
		}
	} else {
		payload = fmt.Sprintf(`{"type":"%s"}`, eventType)
	}
	fmt.Fprintf(w, "data: %s\n\n", payload)
}
