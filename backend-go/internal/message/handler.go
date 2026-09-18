package message

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/chat"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/orchestrator"
	"ai_avengers/backend/internal/response"
)

// SendMessageRequest is the input for sending a message.
type SendMessageRequest struct {
	Message   string      `json:"message" binding:"required"`
	ExpertIDs []string    `json:"expert_ids" binding:"required,min=1"`
	// ReplyToMessageID (CT-C1): optional. When set, this message is a
	// reply to a specific prior message (CT-L6: pins exactly that one
	// message by default). nil/absent = fresh question, the existing
	// behavior for every message sent before this feature (CT-L2-style
	// fallback, applied here to messages instead of experts).
	ReplyToMessageID string `json:"reply_to_message_id,omitempty"`
	// IncludeFullThread (CT-L6): explicit opt-in only. false/absent means
	// only the single pinned message is used as reply context — never
	// automatic full-chain inclusion. Ignored if ReplyToMessageID is empty.
	IncludeFullThread bool `json:"include_full_thread,omitempty"`
}

// SSEEvent types for streaming
const (
	SSEThinking   = "thinking"
	SSEChunk      = "chunk"
	SSEComplete   = "complete"
	SSESynthesis  = "synthesis"
	SSEDone       = "done"
	SSEError      = "error"
)

// Handler handles message sending with SSE streaming.
//
// WHY SSE over WebSocket:
// SSE is one-directional (server -> client) which is all we need.
// Simpler than WebSocket, works over HTTP/1.1, auto-reconnect built-in.
// Architecture doc locked decision: SSE for streaming.
type Handler struct {
	chatSvc      *chat.Service
	orchestrator *orchestrator.Orchestrator
	gateway      *gateway.ModelGateway
	embedder     ml.Embedder // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	memManager   *memory.Manager
	logger       *zap.Logger
}

// NewHandler creates a new message handler.
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
func NewHandler(
	chatSvc *chat.Service,
	orch *orchestrator.Orchestrator,
	gw *gateway.ModelGateway,
	embedder ml.Embedder,
	memManager *memory.Manager,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		chatSvc:      chatSvc,
		orchestrator: orch,
		gateway:      gw,
		embedder:     embedder,
		memManager:   memManager,
		logger:       logger,
	}
}

// Send POST /chats/:id/messages
// Streams response via SSE.
//
// Flow:
// 1. Validate request
// 2. Save user message to DB
// 3. Get turn number
// 4. Stream SSE: "thinking" events while processing
// 5. Run orchestrator (parallel experts)
// 6. Stream SSE: expert responses one by one
// 7. Stream SSE: synthesis if multiple experts
// 8. Save assistant messages to DB
// 9. Update memory async
// 10. Stream SSE: "done"
//
// Mental execution:
// Client sends: "How should I design the user table?"
// SSE stream:
//   data: {"type":"thinking","expert":"DB Expert","gate":1}
//   data: {"type":"thinking","expert":"DB Expert","gate":5}
//   data: {"type":"chunk","expert":"DB Expert","content":"Use UUID..."}
//   data: {"type":"complete","expert":"DB Expert","mode":"ADVISE"}
//   data: {"type":"done"}
func (h *Handler) Send(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}

	// Parse request — support both JSON and multipart (file upload)
	var req SendMessageRequest
	var fileContent string

	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		req.Message = c.PostForm("message")
		expertIDsStr := c.PostForm("expert_ids")
		if err := json.Unmarshal([]byte(expertIDsStr), &req.ExpertIDs); err != nil {
			response.BadRequest(c, "INVALID_INPUT", "expert_ids must be JSON array")
			return
		}
		// CT-D4 fix: the multipart branch never read reply_to_message_id/
		// include_full_thread at all — frontend's useSSEStream.ts sends
		// these as regular form fields on this SAME branch (alongside
		// message/expert_ids above) whenever a file is attached to a
		// reply, but they were silently dropped here while the JSON
		// branch's c.ShouldBindJSON(&req) below correctly picked them up
		// via SendMessageRequest's json tags. Fixed by reading them the
		// same way message/expert_ids are read on this branch.
		req.ReplyToMessageID = c.PostForm("reply_to_message_id")
		req.IncludeFullThread = c.PostForm("include_full_thread") == "true"
		// Handle file upload
		if file, header, err := c.Request.FormFile("file"); err == nil {
			defer file.Close()
			if header.Size < 1*1024*1024 { // Max 1MB for context
				if content, err := io.ReadAll(file); err == nil {
					fileContent = fmt.Sprintf("\n\n[Attached file: %s]\n%s",
						header.Filename, string(content))
				}
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}
	}

	if req.Message == "" {
		response.BadRequest(c, "INVALID_INPUT", "message is required")
		return
	}

	// Parse expert IDs
	var expertIDs []uuid.UUID
	for _, idStr := range req.ExpertIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			response.BadRequest(c, "INVALID_ID", fmt.Sprintf("invalid expert ID: %s", idStr))
			return
		}
		expertIDs = append(expertIDs, id)
	}

	// Parse reply_to_message_id (CT-C1). Empty string is valid (fresh
	// question) — only parse+validate when non-empty, matching the
	// existing expertIDs loop's error-on-malformed-input pattern above.
	var replyToMessageID *uuid.UUID
	if req.ReplyToMessageID != "" {
		parsed, err := uuid.Parse(req.ReplyToMessageID)
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid reply_to_message_id")
			return
		}
		replyToMessageID = &parsed
	}

	// Verify chat ownership
	ch, err := h.chatSvc.GetByID(c.Request.Context(), chatID, clientID)
	if err != nil {
		response.NotFound(c, "chat")
		return
	}

	// Get turn number
	turnNumber, err := h.chatSvc.GetNextTurnNumber(c.Request.Context(), chatID)
	if err != nil {
		response.InternalError(c)
		return
	}

	// Append file content to message if present
	fullMessage := req.Message
	if fileContent != "" {
		fullMessage += fileContent
	}

	// Save user message
	userMsgID, err := h.chatSvc.SaveMessage(c.Request.Context(), chat.Message{
		ChatID:           chatID,
		Role:             "user",
		Content:          fullMessage,
		TurnNumber:       turnNumber,
		ReplyToMessageID: replyToMessageID,
	})
	if err != nil {
		h.logger.Error("save user message failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Increment message count
	_ = h.chatSvc.IncrementMessageCount(c.Request.Context(), chatID)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Stream processing
	c.Stream(func(w io.Writer) bool {
		// Send thinking event
		sendSSE(w, SSEThinking, map[string]interface{}{
			"message": "Processing your request...",
			"experts": len(expertIDs),
		})

		// Streaming: create a token channel when exactly one expert is selected.
		// WHY single expert only: multiple experts run in parallel goroutines.
		// Mixing tokens from N experts into one SSE stream would interleave
		// tokens from different experts, making the response unreadable.
		// Single expert = 95% of real usage (DSA, System Design one at a time).
		// Multi-expert = blocking path (same as before, no regression).
		var tokenCh chan string
		var tokenDone chan struct{}
		if len(expertIDs) == 1 {
			tokenCh = make(chan string, 128)
			tokenDone = make(chan struct{})
			// Goroutine: forward tokens from channel to SSE as they arrive.
			// Runs concurrently with orchestrator.Process().
			// Empty string tokens are heartbeats (sent by blocking-fallback
			// path in enforcer.go to keep SSE connection alive) — skip them
			// so no fake content reaches the frontend.
			go func() {
				defer close(tokenDone)
				for token := range tokenCh {
					if token == "" {
						continue // heartbeat — keep connection alive, no content
					}
					sendSSE(w, SSEChunk, map[string]interface{}{
						"content":   token,
						"expert_id": expertIDs[0].String(),
					})
				}
			}()
		}

		// Run orchestrator
		orchestratorReq := orchestrator.OrchestratorRequest{
			ProjectID:         ch.ProjectID,
			ClientID:          clientID,
			ChatID:            chatID,
			Message:           fullMessage,
			ExpertIDs:         expertIDs,
			TurnNumber:        turnNumber,
			ReplyToMessageID:  replyToMessageID,
			IncludeFullThread: req.IncludeFullThread,
			UserMessageID:     userMsgID,
		}
		if tokenCh != nil {
			orchestratorReq.TokenCh = tokenCh
		}

		orchestratorResp, err := h.orchestrator.Process(c.Request.Context(), orchestratorReq)

		// Wait for token forwarding goroutine to finish before sending SSEComplete.
		// This ensures all streamed tokens arrive before the complete event.
		if tokenDone != nil {
			<-tokenDone
		}

		if err != nil {
			h.logger.Error("orchestrator failed", zap.Error(err))
			sendSSE(w, SSEError, map[string]string{"message": "Processing failed"})
			sendSSE(w, SSEDone, nil)
			return false
		}

		// Stream each expert response
		// savedMessageIDs: collect message_id per expert for SSEDone.
		// WHY sync save (not async goroutine):
		//   Frontend needs message_id to render the Reply button.
		//   If save is async, SSEDone arrives before DB write completes,
		//   message_id is unavailable, Reply button never renders.
		//   Save is fast (single INSERT, <5ms) — sync cost is negligible
		//   compared to the 2-10s LLM generation that just completed.
		savedMessageIDs := make(map[string]string) // expert_id -> message_id
		for _, expertResp := range orchestratorResp.ExpertResponses {
			// Ensure citations is never null in SSE payload.
			// WHY: frontend calls citations.map() — null crashes JS.
			// REFUSE/ASK modes have no citations — send [] not null.
			citations := expertResp.Citations
			if citations == nil {
				citations = []chinawall.Citation{}
			}
			// Send complete expert response
			sendSSE(w, SSEComplete, map[string]interface{}{
				"expert_id":   expertResp.ExpertID,
				"expert_name": expertResp.ExpertName,
				"domain":      expertResp.Domain,
				"mode":        expertResp.Mode,
				"content":     expertResp.Content,
				"citations":   citations,
				"confidence":  expertResp.Confidence,
				"gate_stopped": expertResp.GateStopped,
				"warning":     expertResp.Warning,
				"questions":   expertResp.Questions,
				// CT-B4: nil/omitted for every flat-text expert response (CT-L2).
				// Frontend (CT-D5, not yet built) renders this when present,
				// falls back to "content" above otherwise.
				"template_sections": expertResp.TemplateSections,
			})

			// Save assistant message SYNCHRONOUSLY.
			// WHY sync: message_id needed in SSEDone for Reply button.
			// Save is <5ms — negligible after 2-10s LLM generation.
			savedID := h.saveAssistantMessage(context.Background(), chatID, expertResp, turnNumber)
			if savedID.String() != uuid.Nil.String() {
				savedMessageIDs[expertResp.ExpertID.String()] = savedID.String()
			}
		}

		// Send synthesis if multiple experts
		if orchestratorResp.Synthesis != nil {
			sendSSE(w, SSESynthesis, orchestratorResp.Synthesis)
		}

		// Update chat index async
		go h.indexTurn(context.Background(), chatID, fullMessage, orchestratorResp, turnNumber)

		// Generate rolling summary every 10 turns
		if turnNumber%10 == 0 {
			go h.generateRollingSummary(context.Background(), chatID, turnNumber)
		}

		sendSSE(w, SSEDone, map[string]interface{}{
			"turn_number":  turnNumber,
			"duration_ms":  orchestratorResp.DurationMs,
			// message_ids: expert_id -> saved message_id.
			// Frontend uses this to render the Reply button immediately
			// after SSEDone, without a separate API call to fetch message_id.
			// Empty map when all saves failed (graceful degradation).
			"message_ids": savedMessageIDs,
		})

		// CRITICAL: Flush the writer to ensure SSEDone reaches frontend immediately.
		// WHY: Without flush, Gin may buffer the final event, causing frontend
		// to show "Streaming..." indefinitely even though backend is done.
		// BUG FIX: Frontend "Streaming..." persists after completion.
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		return false // Stop streaming
	})
}

// saveAssistantMessage saves an expert's response to the messages table.
// Returns the saved message ID so the caller can include it in SSEDone.
// WHY return ID: frontend needs message_id to render the Reply button.
// If save fails, returns uuid.Nil — caller sends SSEDone without message_id.
func (h *Handler) saveAssistantMessage(
	ctx context.Context,
	chatID uuid.UUID,
	resp orchestrator.ExpertResponse,
	turnNumber int,
) uuid.UUID {
	if resp.Error != "" {
		return uuid.Nil // Don't save failed responses
	}

	// FIXED 2026-09-08 (RCA round 7, migration 011): this WAS a documented
	// KNOWN GAP - the messages table had no column for resp.TemplateSections,
	// so a categorized expert's structured sections streamed correctly over
	// SSE (see sendSSE(w, SSEComplete, ...) above, which DOES include
	// "template_sections") but were never persisted. Content alone is empty
	// for a structured response (the structured generation path only
	// populates TemplateSections, never Answer/Content), so a page reload
	// rendered the answer as completely blank - while the first, live SSE
	// render looked correct. messages.template_sections (migration 011)
	// now stores exactly what was streamed, so ListMessages/reload renders
	// identically to the live stream.
	expertID := resp.ExpertID
	savedID, err := h.chatSvc.SaveMessage(ctx, chat.Message{
		ChatID:              chatID,
		Role:                "assistant",
		Content:             resp.Content,
		TurnNumber:          turnNumber,
		ExpertID:            &expertID,
		DecisionMode:        string(resp.Mode),
		Confidence:          resp.Confidence,
		// Persist Gate 3 WARN text and Gate 1 ASK questions
		// so they survive page reloads (Bug 3 fix)
		WarningText:         resp.Warning,
		ClarifyingQuestions: resp.Questions,
		// Bug 3.4 fix (docs bug list): resp.Citations was never passed
		// through here, so the messages.citations column was always
		// NULL - this silently broke the rating->chunk-boost feedback
		// loop (rating/handler.go's updateChunkBoosts reads citations
		// back from a saved message to know which chunks to boost).
		Citations:           resp.Citations,
		// CT-C4: when this response IS a structure-permission ASK
		// (orchestrator.go's structurePermissionAskParent sentinel),
		// this sets the ASK message's OWN reply_to_message_id back to
		// the user's question — so a later reply-to-this-ASK can walk
		// one more parent level and recover the original question
		// (decision/engine.go's gateStructurePermission). nil for every
		// other response — identical to before this feature existed.
		ReplyToMessageID:    resp.ReplyToUserMessageID,
		// nil for every flat-text response (CT-L2) - resp.TemplateSections
		// is only non-nil for a categorized expert's structured answer.
		// pgx encodes a nil []chinawall.TemplateSectionResult as SQL NULL
		// for the JSONB column (interface{} field, same pattern already
		// used for Citations above), never an empty-but-present JSON value.
		TemplateSections:    resp.TemplateSections,
	})
	if err != nil {
		h.logger.Warn("save assistant message failed",
			zap.String("expert", resp.ExpertName),
			zap.Error(err),
		)
		_ = h.chatSvc.IncrementMessageCount(ctx, chatID)
		return uuid.Nil
	}
	_ = h.chatSvc.IncrementMessageCount(ctx, chatID)
	return savedID
}

// indexTurn creates a chat_index entry for semantic search.
// Uses cheap LLM to generate one-line summary and extract topic.
func (h *Handler) indexTurn(
	ctx context.Context,
	chatID uuid.UUID,
	userMessage string,
	orchestratorResp *orchestrator.OrchestratorResponse,
	turnNumber int,
) {
	if len(orchestratorResp.ExpertResponses) == 0 {
		return
	}

	// Build combined turn text for summarization
	firstResp := orchestratorResp.ExpertResponses[0]
	turnText := fmt.Sprintf("User: %s\nAssistant: %s",
		userMessage[:minInt(200, len(userMessage))],
		firstResp.Content[:minInt(300, len(firstResp.Content))],
	)

	// Generate summary via cheap LLM
	prompt := fmt.Sprintf(`Summarize this conversation turn in one line (max 100 chars).
Also identify the main topic.

%s

Return JSON: {"summary": "...", "topic": "...", "importance": 1-5}`, turnText)

	resp, err := h.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   100,
		Temperature: 0.1,
		UseCache:    false,
	})

	summary := userMessage[:minInt(100, len(userMessage))]
	topic := "general"
	importance := 3

	if err == nil {
		var result struct {
			Summary    string `json:"summary"`
			Topic      string `json:"topic"`
			Importance int    `json:"importance"`
		}
		clean := strings.TrimSpace(resp.Content)
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		clean = strings.TrimSpace(clean)
		start := strings.Index(clean, "{")
		end := strings.LastIndex(clean, "}")
		if start != -1 && end != -1 {
			if err := json.Unmarshal([]byte(clean[start:end+1]), &result); err == nil {
				if result.Summary != "" {
					summary = result.Summary
				}
				if result.Topic != "" {
					topic = result.Topic
				}
				if result.Importance >= 1 && result.Importance <= 5 {
					importance = result.Importance
				}
			}
		}
	}

	// Generate embedding for semantic search
	var embedding []float32
	if emb, err := h.embedder.EmbedSingle(ctx, summary); err == nil {
		embedding = emb
	}

	// Save to chat_index
	_ = h.chatSvc.IndexTurn(ctx, chatID, uuid.New(), turnNumber, summary, topic, importance, embedding)
}

// generateRollingSummary generates a rolling summary every 10 turns.
// WHY rolling summary (Arpit Bhiyani + Byte by Byte AI):
// Context window is finite. Summary compresses history.
// Prevents lost-in-middle problem for long conversations.
func (h *Handler) generateRollingSummary(ctx context.Context, chatID uuid.UUID, turnNumber int) {
	// Get last 10 chat index entries
	rows, err := h.chatSvc.GetDB().Query(ctx,
		`SELECT one_line_summary, topic, turn_number
		 FROM chat_index
		 WHERE chat_id=$1
		 ORDER BY turn_number DESC
		 LIMIT 10`,
		chatID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	var summaries []string
	for rows.Next() {
		var s, topic string
		var turn int
		if err := rows.Scan(&s, &topic, &turn); err != nil {
			continue
		}
		summaries = append(summaries, fmt.Sprintf("Turn %d [%s]: %s", turn, topic, s))
	}

	if len(summaries) == 0 {
		return
	}

	prompt := fmt.Sprintf(`Summarize this conversation history in 3-5 sentences.
Capture: main topics, key decisions, current project state.

%s

Summary:`, strings.Join(summaries, "\n"))

	resp, err := h.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   300,
		Temperature: 0.3,
	})
	if err != nil {
		return
	}

	turnStart := turnNumber - 9
	if turnStart < 1 {
		turnStart = 1
	}
	_ = h.chatSvc.SaveRollingSummary(ctx, chatID, resp.Content, turnStart, turnNumber)
}

// sendSSE writes a single SSE event to the response writer.
func sendSSE(w io.Writer, eventType string, data interface{}) {
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
