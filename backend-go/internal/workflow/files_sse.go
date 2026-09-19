package workflow

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/response"
)

// SSE event type constants for the file/code stream.
// Naming mirrors kanban_sse.go's SSEKanban* constants exactly.
const (
	SSEFileArtifact = "file_artifact" // expert produced/modified a code file (code_artifact_produced)
	SSEFileWave     = "file_wave"     // a wave's code was merged into main/ (wave_completed)
	SSEFileDone     = "file_done"     // workflow completed or failed
	SSEFileError    = "file_error"    // stream error
)

// StreamFiles handles GET /api/v1/workflows/:id/files/stream
// Streams live file/code updates via SSE — the file-browser counterpart
// to StreamKanban (kanban_sse.go). Same subscriber, same connection
// lifecycle, different event filter.
//
// DESIGN:
//   Subscribes to the same workflow blackboard Redis channel as Kanban.
//   Only translates the two event types relevant to a file browser:
//     code_artifact_produced -> SSE file_artifact (includes "operation": create|modify)
//     wave_completed          -> SSE file_wave (main/ was just updated)
//   Everything else is ignored — a file browser doesn't care about task
//   status or approval gates, that's StreamKanban's job.
//
// Mental execution:
//   Client connects to /workflows/abc/files/stream
//   1. Verify workflow exists
//   2. Subscribe to blackboard Redis channel (from seq 0 — full history)
//   3. For each new blackboard event:
//      - code_artifact_produced -> SSE file_artifact
//      - wave_completed          -> SSE file_wave
//      - workflow_completed/failed -> SSE file_done + close
//      - everything else -> ignored
//   4. Client disconnects -> ctx.Done() -> subscriber stops
func (h *Handler) StreamFiles(c *gin.Context) {
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
	// WHY from 0: client may have just connected and needs full file history
	// to render the file browser (not just files produced after connecting).
	eventCh, errCh := sub.Subscribe(c.Request.Context(), workflowID, uuid.Nil, 0)

	// ✅ Same fix as StreamKanban: write initial ping BEFORE entering stream
	// loop. Go/Gin doesn't send 200 OK headers until first write, so without
	// this the browser stays in CONNECTING state if no matching events exist
	// yet (e.g. workflow still in design phase, no code_artifact_produced
	// events posted).
	c.Writer.Write([]byte(": connected\n\n"))
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed — subscriber stopped.
				sendKanbanSSE(w, SSEFileDone, map[string]string{"reason": "stream_ended"})
				return false
			}
			h.handleFileEvent(w, event)
			if event.EventType == "workflow_completed" || event.EventType == "workflow_failed" {
				sendKanbanSSE(w, SSEFileDone, map[string]interface{}{
					"event_type":  event.EventType,
					"workflow_id": workflowID.String(),
				})
				return false
			}
			return true

		case err := <-errCh:
			if err != nil {
				h.logger.Warn("file stream: subscriber error",
					zap.String("workflow_id", workflowID.String()),
					zap.Error(err),
				)
				sendKanbanSSE(w, SSEFileError, map[string]string{"message": err.Error()})
			}
			return false

		case <-c.Request.Context().Done():
			// Client disconnected.
			return false
		}
	})
}

// handleFileEvent translates a blackboard event into a file-stream SSE
// event. Only code_artifact_produced and wave_completed are forwarded;
// everything else is dropped (StreamKanban already covers task/approval
// updates — this stream is scoped to code/file content only).
func (h *Handler) handleFileEvent(w io.Writer, event blackboard.Event) {
	payload := map[string]interface{}{
		"sequence_number":     event.SequenceNumber,
		"event_type":          event.EventType,
		"posted_by_expert_id": event.PostedByExpertID,
		"posted_at":           event.PostedAt,
		"content":             json.RawMessage(event.Content),
	}

	switch event.EventType {
	case "code_artifact_produced":
		sendKanbanSSE(w, SSEFileArtifact, payload)
	case "wave_completed":
		sendKanbanSSE(w, SSEFileWave, payload)
	default:
		// Not a file-relevant event — ignored.
	}
}
