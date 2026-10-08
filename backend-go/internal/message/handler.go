package message

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/auth"
	"ai_avengers/backend/internal/chat"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/collab"
	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/orchestrator"
	"ai_avengers/backend/internal/provenance"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/tenant"
	"ai_avengers/backend/internal/usage"
)

// Attachment limits for the chat. WHY 2MB: a chat attachment is context, not a
// corpus upload — 2MB covers real PDFs/DOCX/XLSX specs while keeping the prompt
// bounded. Files above the cap are reported back in the message, never silently
// dropped.
const (
	maxAttachmentBytes        = 2 * 1024 * 1024
	maxAttachmentsPerMessage  = 5
	maxAttachmentCharsInTotal = 60000
)

// extractAttachments turns every uploaded file into text using the SAME
// extractor the training ingestion uses (docextract -> ml-sidecar /extract).
//
// WHY this replaced the old io.ReadAll: the previous code treated the upload as
// plain text, so a PDF/DOCX/XLSX arrived as binary bytes and the expert answered
// from garbage. A nil extractor (sidecar not configured) still handles the
// plain-text formats.
func (h *Handler) extractAttachments(c *gin.Context) string {
	if h.extractor == nil {
		return ""
	}
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return ""
	}
	files := form.File["file"]
	if len(files) == 0 {
		return ""
	}
	if len(files) > maxAttachmentsPerMessage {
		files = files[:maxAttachmentsPerMessage]
	}
	var sb strings.Builder
	for _, header := range files {
		if header.Size > maxAttachmentBytes {
			fmt.Fprintf(&sb, "\n\n[Attached file skipped: %s — larger than %d MB]",
				header.Filename, maxAttachmentBytes/(1024*1024))
			continue
		}
		opened, openErr := header.Open()
		if openErr != nil {
			continue
		}
		data, readErr := io.ReadAll(opened)
		opened.Close()
		if readErr != nil {
			continue
		}
		result, extractErr := h.extractor.Extract(c.Request.Context(), header.Filename, data)
		if extractErr != nil {
			// Surface the reason (unsupported_format, pdf_encrypted, ...) so the
			// user knows the file was not read instead of getting a silent answer.
			fmt.Fprintf(&sb, "\n\n[Attached file could not be read: %s — %s]",
				header.Filename, extractErr.Error())
			continue
		}
		fmt.Fprintf(&sb, "\n\n[Attached file: %s]\n%s", header.Filename, result.Text)
		if sb.Len() >= maxAttachmentCharsInTotal {
			sb.WriteString("\n\n[Further attachments omitted: context limit reached]")
			break
		}
	}
	return sb.String()
}

func parseUUIDOrNil(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// GetActiveRelay returns the latest non-terminal relay run for a chat plus its
// transparency transcript and, if a section is awaiting review, that section's
// gate payload. Used by the frontend on reconnect to hydrate the live view.
func (h *Handler) GetActiveRelay(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	store := collab.NewRelayStore(h.db)
	run, err := store.GetActiveRun(c.Request.Context(), chatID)
	if err != nil {
		// no active run -> 204-style empty; frontend treats as "nothing to hydrate"
		c.JSON(http.StatusNotFound, gin.H{"relay": nil})
		return
	}
	events, _ := store.EventsAfter(c.Request.Context(), run.ID, 0)

	payload := gin.H{
		"run_id":     run.ID,
		"chat_id":    run.ChatID,
		"status":     run.Status,
		"events":     events,
		"plan":       run.Plan,
		"current_index": run.CurrentIndex,
	}

	if run.Status == collab.RunAwaitingReview {
		sec, err := store.GetSectionByIndex(c.Request.Context(), run.ID, run.CurrentIndex)
		if err == nil {
			payload["pending_review"] = gin.H{
				"section_id":           sec.ID,
				"section_index":        sec.SectionIndex,
				"section_title":        sec.SectionTitle,
				"expert_name":          sec.ExpertName,
				"content":              sec.RawContent,
				"conflict":             sec.ConflictDetected,
				"conflict_explanation": sec.ConflictExplanation,
			}
		}
	}

	if run.Status == collab.RunRelayFailed {
		payload["failure"] = gin.H{
			"failed_at_index": run.FailedAtIndex,
			"reason":          run.FailureReason,
			"retry_until":     run.RetryDeadline,
		}
	}

	c.JSON(http.StatusOK, payload)
}

// ApproveRelaySection is the human-in-the-loop action (ans7 edit + approve,
// ans3 resolve). After writing, if the relay goroutine is no longer alive, it
// re-launches the relay from the persisted cursor so the pipeline continues.
func (h *Handler) ApproveRelaySection(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("runId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid run ID")
		return
	}
	sectionID, err := uuid.Parse(c.Param("sectionId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid section ID")
		return
	}
	var body struct {
		EditedContent    string `json:"edited_content"`
		ResolutionSource string `json:"resolution_source"` // 'expert_a' | 'expert_b' | 'human_merged'
	}
	_ = c.ShouldBindJSON(&body)

	store := collab.NewRelayStore(h.db)
	sec, _ := store.GetSection(c.Request.Context(), sectionID)
	if sec.ID == uuid.Nil {
		response.NotFound(c, "section")
		return
	}

	finalContent := body.EditedContent
	edited := body.EditedContent != ""
	if !edited {
		finalContent = sec.RawContent
	}

	approved, err := store.ApproveSection(c.Request.Context(), runID, sectionID, finalContent, edited, body.ResolutionSource)
	if err != nil {
		response.InternalError(c)
		return
	}
	_ = store.UpdateRunStatus(c.Request.Context(), runID, collab.RunRunning)

	// If the relay goroutine died (client disconnected mid-gate), resume it.
	h.resumeRelayIfInactive(c.Request.Context(), runID)

	c.JSON(http.StatusOK, gin.H{
		"approved":         true,
		"section_index":    approved.SectionIndex,
		"final_content":    approved.FinalContent,
		"edited":           approved.Edited,
		"resolution_source": approved.ResolutionSource,
	})
}

// RetryRelay re-launches a failed relay from its failed section, only valid
// before retry_deadline (ans4).
func (h *Handler) RetryRelay(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("runId"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid run ID")
		return
	}
	store := collab.NewRelayStore(h.db)
	run, err := store.GetRun(c.Request.Context(), runID)
	if err != nil {
		response.NotFound(c, "relay")
		return
	}
	if run.RetryDeadline != nil && time.Now().After(*run.RetryDeadline) {
		response.BadRequest(c, "RETRY_EXPIRED", "retry window elapsed")
		return
	}
	h.resumeRelayIfInactive(c.Request.Context(), runID)
	c.JSON(http.StatusOK, gin.H{"retrying": true, "run_id": runID})
}

func (h *Handler) resumeRelayIfInactive(ctx context.Context, runID uuid.UUID) {
	// Idempotent claim: only proceed if the run is NOT currently progressing
	// (status 'running' means a live goroutine owns it). A 0-row update means
	// the live relay is still alive — skip. This survives process restarts.
	tag, err := h.db.Exec(ctx,
		`UPDATE collab_relay_runs SET status='retrying', updated_at=now()
		  WHERE id=$1 AND status IN ('awaiting_human_review','relay_failed')`, runID)
	if err != nil || tag.RowsAffected() == 0 {
		return
	}
	go func() { _ = h.runResume(context.Background(), runID) }()
}

func (h *Handler) runResume(ctx context.Context, runID uuid.UUID) error {
	store := collab.NewRelayStore(h.db)
	run, err := store.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	// Rebuild a minimal request from the stored run.
	req := orchestrator.OrchestratorRequest{
		ChatID:      run.ChatID,
		ClientID:    uuid.Nil,
		TurnNumber:  0,
		ExpertIDs:   sectionExpertIDs(run.Plan),
		ResumeRunID: runID,
	}
	_, err = h.orchestrator.ProcessCollaborative(ctx, req, nil, nil)
	return err
}

func sectionExpertIDs(plan []collab.Section) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(plan))
	for _, s := range plan {
		ids = append(ids, s.ExpertID)
	}
	return ids
}

// SendMessageRequest is the input for sending a message.
type SendMessageRequest struct {
	Message   string   `json:"message" binding:"required"`
	ExpertIDs []string `json:"expert_ids" binding:"required,min=1"`
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
	// TemplateName (T-CAT): optional. Chooses which named answer format of the
	// expert's category to use (e.g. "Code" vs "Approach"). Empty = the
	// category's default variant, which is what every pre-existing client sends.
	TemplateName string `json:"template_name,omitempty"`
	// GenericAllowancePct (0-30): how much general knowledge this message may
	// use when the trained chunks do not cover the question. 0/absent = strict
	// China Wall (refuse), which is what every pre-existing client sends.
	GenericAllowancePct float64 `json:"generic_allowance_pct,omitempty"`
	// AnswerMode (Collaborative Relay): "" / "independent" (default, every
	// pre-existing client) runs each selected expert in parallel exactly as
	// before. "collaborative" requires 2+ expert_ids and instead runs them
	// SEQUENTIALLY as one section-organized answer (internal/collab) — see
	// orchestrator.ProcessCollaborative. Any other value is treated as
	// "independent" (unknown values fail safe to the existing behavior,
	// never to an error).
	AnswerMode string `json:"answer_mode,omitempty"`
	// ResumeRunID (Collaborative Relay retry/resume): when set, continue the
	// given relay run from its persisted cursor instead of starting fresh.
	ResumeRunID string `json:"resume_run_id,omitempty"`
}

// SSEEvent types for streaming
const (
	SSEThinking  = "thinking"
	SSEChunk     = "chunk"
	SSEComplete  = "complete"
	SSESynthesis = "synthesis"
	// SSESection (Collaborative Relay): one finished section, sent as each
	// expert's turn completes in relay order. Independent mode (the
	// existing SSEComplete above) never sends this event.
	SSESection = "collab_section"
	// SSERelayStep (Collaborative Relay, Phase 2): one transparency-log event.
	// Carries step, expert_id, message, metadata. Rendered live, strictly in
	// arrival order; persisted server-side to collab_relay_events for replay.
	SSERelayStep = "relay_step"
	SSEDone    = "done"
	SSEError   = "error"
)

// Handler handles message sending with SSE streaming.
//
// WHY SSE over WebSocket:
// SSE is one-directional (server -> client) which is all we need.
// Simpler than WebSocket, works over HTTP/1.1, auto-reconnect built-in.
// Architecture doc locked decision: SSE for streaming.
type Handler struct {
	db           *pgxpool.Pool
	chatSvc      *chat.Service
	orchestrator *orchestrator.Orchestrator
	gateway      *gateway.ModelGateway
	// extractor (nil-safe) turns an attached PDF/DOCX/XLSX/... into text using
	// the same ml-sidecar path as training ingestion.
	extractor  *docextract.Extractor
	embedder   ml.Embedder // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	memManager *memory.Manager
	// prov (C1): optional signed provenance chain recorder. Nil-safe — when
	// unset, answers are still saved, only the provenance chain is skipped.
	prov *provenance.Service
	// tenant (C4): isolation enforcement. Nil-safe — an unwired service is
	// disabled and resolves a global scope, so every assertion is a no-op.
	tenant *tenant.Service
	logger *zap.Logger
}

// NewHandler creates a new message handler.
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
// SetExtractor wires the document extractor used for chat attachments.
func (h *Handler) SetExtractor(e *docextract.Extractor) { h.extractor = e }

func NewHandler(
	db *pgxpool.Pool,
	chatSvc *chat.Service,
	orch *orchestrator.Orchestrator,
	gw *gateway.ModelGateway,
	embedder ml.Embedder,
	memManager *memory.Manager,
	prov *provenance.Service,
	tenantSvc *tenant.Service,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		db:           db,
		chatSvc:      chatSvc,
		orchestrator: orch,
		gateway:      gw,
		embedder:     embedder,
		memManager:   memManager,
		prov:         prov,
		tenant:       tenantSvc,
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
//
//	data: {"type":"thinking","expert":"DB Expert","gate":1}
//	data: {"type":"thinking","expert":"DB Expert","gate":5}
//	data: {"type":"chunk","expert":"DB Expert","content":"Use UUID..."}
//	data: {"type":"complete","expert":"DB Expert","mode":"ADVISE"}
//	data: {"type":"done"}
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
		if v := c.PostForm("template_name"); v != "" {
			req.TemplateName = v
		}
		if v := c.PostForm("generic_allowance_pct"); v != "" {
			// A malformed value is treated as 0 (strict) rather than failing the
			// whole message — the safe direction.
			if pct, convErr := strconv.ParseFloat(v, 64); convErr == nil {
				req.GenericAllowancePct = pct
			}
		}
		req.AnswerMode = c.PostForm("answer_mode")
		fileContent = h.extractAttachments(c)
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

	// Account-level expert grants via entitlement port (domain_expert only; admin/client unrestricted).
	// auth.MustHaveExpertAccess delegates to entitlement.Checker — same rules, sellable-expert seam.
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if err := auth.MustHaveExpertAccess(c.Request.Context(), h.db, clientID, roleStr, expertIDs); err != nil {
		response.Forbidden(c, err.Error())
		return
	}

	// C4: tenant boundary. Resolve the caller's scope, then verify the
	// requested experts are visible to it. Fail closed — an unresolvable
	// scope denies the request (P3). Nil-safe: an unwired service resolves
	// a global scope and both checks are no-ops.
	scope, err := h.tenant.Resolve(c.Request.Context(), clientID, roleStr)
	if err != nil {
		h.logger.Warn("tenant scope unresolved", zap.Error(err))
		response.Forbidden(c, "Tenant scope could not be resolved")
		return
	}
	if err := h.tenant.AssertExperts(c.Request.Context(), scope, expertIDs); err != nil {
		response.Forbidden(c, err.Error())
		return
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

	// C4: the chat's project must live in the caller's tenant too — a
	// client_id match alone is not sufficient once tenants exist.
	if err := h.tenant.AssertProject(c.Request.Context(), scope, ch.ProjectID); err != nil {
		response.Forbidden(c, err.Error())
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
		// stopForwarding: signal the SSE forwarder to exit WITHOUT closing
		// tokenCh. Closing tokenCh while an expert goroutine may still send
		// (orchestrator timeout path returns without waiting for wg) panics
		// with "send on closed channel" and crashes the process — Gin recovery
		// does not catch panics in non-Gin goroutines. Producers stop via
		// <-ctx.Done() once the HTTP handler returns and net/http cancels the
		// request context; the channel is then GC'd.
		var stopForwarding chan struct{}
		if len(expertIDs) == 1 {
			tokenCh = make(chan string, 128)
			tokenDone = make(chan struct{})
			stopForwarding = make(chan struct{})
			// Goroutine: forward tokens from channel to SSE as they arrive.
			// Runs concurrently with orchestrator.Process().
			// Empty string tokens are heartbeats (sent by blocking-fallback
			// path in enforcer.go to keep SSE connection alive) — skip them
			// so no fake content reaches the frontend.
			go func() {
				defer close(tokenDone)
				forward := func(token string) {
					if token == "" {
						return // heartbeat — keep connection alive, no content
					}
					sendSSE(w, SSEChunk, map[string]interface{}{
						"content":   token,
						"expert_id": expertIDs[0].String(),
					})
				}
				for {
					select {
					case token, ok := <-tokenCh:
						if !ok {
							return
						}
						forward(token)
					case <-stopForwarding:
						// Drain whatever is already buffered so happy-path
						// tokens are not dropped, then exit. Do not block
						// waiting for future sends (producers may still be
						// alive on the orchestrator-timeout path).
						for {
							select {
							case token, ok := <-tokenCh:
								if !ok {
									return
								}
								forward(token)
							default:
								return
							}
						}
					case <-c.Request.Context().Done():
						// Ghost fix: browser/app band ho gaya while live typing.
						// Client disconnect => forwarder turant band, producers
						// (orchestrator -> enforcer -> gateway LLM) bhi isi ctx
						// ke Done() se cancel hokar ghost process nahi bante.
						return
					}
				}
			}()
		}

		// Run orchestrator.
		// AnswerMode (Collaborative Relay): "collaborative" with 2+
		// experts runs them SEQUENTIALLY as one section-organized answer
		// (orchestrator.ProcessCollaborative) instead of the parallel
		// independent-mode path below. Any other value (including an
		// unknown string, or "collaborative" with only 1 expert
		// selected) falls back to independent mode unchanged.
		collabMode := req.AnswerMode == "collaborative" && len(expertIDs) >= 2

		orchestratorReq := orchestrator.OrchestratorRequest{
			ProjectID:           ch.ProjectID,
			ClientID:            clientID,
			ChatID:              chatID,
			Message:             fullMessage,
			ExpertIDs:           expertIDs,
			TurnNumber:          turnNumber,
			ReplyToMessageID:    replyToMessageID,
			IncludeFullThread:   req.IncludeFullThread,
			UserMessageID:       userMsgID,
			TemplateName:        req.TemplateName,
			GenericAllowancePct: req.GenericAllowancePct,
		}
		if tokenCh != nil {
			orchestratorReq.TokenCh = tokenCh
		}

		var orchestratorResp *orchestrator.OrchestratorResponse
		var err error
		if collabMode {
			orchestratorReq.ResumeRunID = parseUUIDOrNil(req.ResumeRunID)
			orchestratorResp, err = h.orchestrator.ProcessCollaborative(
				c.Request.Context(), orchestratorReq,
				func(ev collab.StepEvent) {
					sendSSE(w, SSERelayStep, map[string]interface{}{
						"step":        string(ev.Step),
						"expert_id":   ev.ExpertID,
						"expert_name": ev.ExpertName,
						"message":     ev.Message,
						"metadata":    ev.Metadata,
					})
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				},
				func(section collab.Section, index, total int) {
					sendSSE(w, SSESection, map[string]interface{}{
						"expert_id":     section.ExpertID,
						"expert_name":   section.ExpertName,
						"section_title": section.SectionTitle,
						"content":       section.Content,
						"index":         index,
						"total":         total,
					})
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				},
			)
		} else {
			orchestratorResp, err = h.orchestrator.Process(c.Request.Context(), orchestratorReq)
		}

		// Stop the forwarder (do NOT close tokenCh — a live producer may still send).
		if stopForwarding != nil {
			close(stopForwarding)
		}

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

		// savedMessageIDs: collect message_id per expert for SSEDone
		// (collab mode uses a single "collab" key instead — see below).
		savedMessageIDs := make(map[string]string) // expert_id -> message_id
		if collabMode && len(orchestratorResp.CollabSections) > 0 {
			// Collaborative Relay: ONE merged message instead of one per
			// expert. Content is the assembled Markdown; collab_sections
			// (migration 075) preserves the structured per-expert data so a
			// reload renders the same sectioned view the live stream showed.
			savedID := h.saveCollabMessage(context.Background(), chatID, orchestratorResp.CollabSections, turnNumber)
			if savedID != uuid.Nil {
				savedMessageIDs["collab"] = savedID.String()
			}
		} else {
			// Stream each expert response
			// savedMessageIDs: collect message_id per expert for SSEDone.
			// WHY sync save (not async goroutine):
			//   Frontend needs message_id to render the Reply button.
			//   If save is async, SSEDone arrives before DB write completes,
			//   message_id is unavailable, Reply button never renders.
			//   Save is fast (single INSERT, <5ms) — sync cost is negligible
			//   compared to the 2-10s LLM generation that just completed.
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
					"expert_id":    expertResp.ExpertID,
					"expert_name":  expertResp.ExpertName,
					"domain":       expertResp.Domain,
					"mode":         expertResp.Mode,
					"content":      expertResp.Content,
					"citations":    citations,
					"confidence":   expertResp.Confidence,
					"gate_stopped": expertResp.GateStopped,
					"warning":      expertResp.Warning,
					"questions":    expertResp.Questions,
					// CT-B4: nil/omitted for every flat-text expert response (CT-L2).
					// Frontend (CT-D5, not yet built) renders this when present,
					// falls back to "content" above otherwise.
					"template_sections": expertResp.TemplateSections,
					// B8: claim→evidence reports; nil/omitted when verify off.
					// Verification labels are never written into content (removed by product decision).
					"claims": expertResp.Claims,
				})

				// Save assistant message SYNCHRONOUSLY.
				// WHY sync: message_id needed in SSEDone for Reply button.
				// Save is <5ms — negligible after 2-10s LLM generation.
				savedID := h.saveAssistantMessage(context.Background(), chatID, expertResp, turnNumber)
				if savedID.String() != uuid.Nil.String() {
					savedMessageIDs[expertResp.ExpertID.String()] = savedID.String()
				}
			}
		}

		// Send synthesis if multiple experts
		if orchestratorResp.Synthesis != nil {
			sendSSE(w, SSESynthesis, orchestratorResp.Synthesis)
		}

		// Update chat index async — use a real messages.id so the FK holds.
		// Prefer the first successfully saved assistant message; fall back to
		// the user message (always saved before streaming). Fake uuid.New()
		// used to FK-fail silently and leave chat_index empty forever.
		indexMsgID := userMsgID
		if idStr, ok := savedMessageIDs["collab"]; ok {
			if parsed, perr := uuid.Parse(idStr); perr == nil && parsed != uuid.Nil {
				indexMsgID = parsed
			}
		} else {
			for _, expertResp := range orchestratorResp.ExpertResponses {
				if idStr, ok := savedMessageIDs[expertResp.ExpertID.String()]; ok {
					if parsed, perr := uuid.Parse(idStr); perr == nil && parsed != uuid.Nil {
						indexMsgID = parsed
						break
					}
				}
			}
		}
		go h.indexTurn(context.Background(), chatID, indexMsgID, fullMessage, orchestratorResp, turnNumber)

		// Generate rolling summary every 10 turns
		if turnNumber%10 == 0 {
			go h.generateRollingSummary(context.Background(), chatID, turnNumber)
		}

		sendSSE(w, SSEDone, map[string]interface{}{
			"turn_number": turnNumber,
			"duration_ms": orchestratorResp.DurationMs,
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
		ChatID:       chatID,
		Role:         "assistant",
		Content:      resp.Content,
		TurnNumber:   turnNumber,
		ExpertID:     &expertID,
		DecisionMode: string(resp.Mode),
		Confidence:   resp.Confidence,
		// Persist Gate 3 WARN text and Gate 1 ASK questions
		// so they survive page reloads (Bug 3 fix)
		WarningText:         resp.Warning,
		ClarifyingQuestions: resp.Questions,
		// Bug 3.4 fix (docs bug list): resp.Citations was never passed
		// through here, so the messages.citations column was always
		// NULL - this silently broke the rating->chunk-boost feedback
		// loop (rating/handler.go's updateChunkBoosts reads citations
		// back from a saved message to know which chunks to boost).
		Citations: resp.Citations,
		// CT-C4: when this response IS a structure-permission ASK
		// (orchestrator.go's structurePermissionAskParent sentinel),
		// this sets the ASK message's OWN reply_to_message_id back to
		// the user's question — so a later reply-to-this-ASK can walk
		// one more parent level and recover the original question
		// (decision/engine.go's gateStructurePermission). nil for every
		// other response — identical to before this feature existed.
		ReplyToMessageID: resp.ReplyToUserMessageID,
		// nil for every flat-text response (CT-L2) - resp.TemplateSections
		// is only non-nil for a categorized expert's structured answer.
		// pgx encodes a nil []chinawall.TemplateSectionResult as SQL NULL
		// for the JSONB column (interface{} field, same pattern already
		// used for Citations above), never an empty-but-present JSON value.
		TemplateSections: resp.TemplateSections,
		// C9: persist the B6 judge score, China Wall coverage verdict and
		// Gate-5 refusal reason so the "why this answer" view is assembled
		// from stored facts (never a fresh LLM narration).
		QualityScore:  resp.QualityScore,
		Coverage:      resp.Coverage,
		RefusalReason: resp.Reason,
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

	// C1: record a signed provenance chain for this answer (async,
	// best-effort). Reuses the B8 span anchors in resp.Claims as the
	// provenance primitive. model = the generation tier (ModelStrong),
	// matching gateway.proxy.go's convention of recording the tier name.
	if h.prov != nil {
		expertIDCopy := expertID
		gateStopped := resp.GateStopped
		go func() {
			recErr := h.prov.RecordChatAnswer(context.Background(), provenance.ChatAnswer{
				MessageID:    savedID,
				ChatID:       chatID,
				ExpertID:     &expertIDCopy,
				Model:        string(gateway.ModelStrong),
				DecisionMode: string(resp.Mode),
				GateStopped:  &gateStopped,
				Content:      resp.Content,
				Claims:       resp.Claims,
				Citations:    resp.Citations,
			})
			if recErr != nil {
				h.logger.Warn("provenance record failed",
					zap.String("message_id", savedID.String()),
					zap.Error(recErr),
				)
			}
		}()
	}
	return savedID
}

// saveCollabMessage saves a Collaborative Relay turn's merged sections as
// ONE assistant message (no single expert_id — the message represents all
// of them). Content is the assembled Markdown (collab.AssembleMarkdown);
// collab_sections (migration 075) stores the structured per-expert data so
// ListMessages/reload renders the identical sectioned view the live SSE
// stream showed, same convention as saveAssistantMessage's TemplateSections.
// Returns uuid.Nil on save failure, same fail-soft contract as
// saveAssistantMessage, so the caller can still send SSEDone.
func (h *Handler) saveCollabMessage(
	ctx context.Context,
	chatID uuid.UUID,
	sections []collab.Section,
	turnNumber int,
) uuid.UUID {
	savedID, err := h.chatSvc.SaveMessage(ctx, chat.Message{
		ChatID:         chatID,
		Role:           "assistant",
		Content:        collab.AssembleMarkdown(sections),
		TurnNumber:     turnNumber,
		CollabSections: sections,
	})
	if err != nil {
		h.logger.Warn("save collab message failed", zap.Error(err))
		_ = h.chatSvc.IncrementMessageCount(ctx, chatID)
		return uuid.Nil
	}
	_ = h.chatSvc.IncrementMessageCount(ctx, chatID)
	return savedID
}

// indexTurn creates a chat_index entry for semantic search.
// Uses cheap LLM to generate one-line summary and extract topic.
// messageID must be an existing messages.id (FK on chat_index.message_id).
func (h *Handler) indexTurn(
	ctx context.Context,
	chatID uuid.UUID,
	messageID uuid.UUID,
	userMessage string,
	orchestratorResp *orchestrator.OrchestratorResponse,
	turnNumber int,
) {
	// C5: attribute the (cheap) index call to this chat.
	ctx = usage.WithAttribution(ctx, usage.Attribution{
		ChatID: &chatID, UseCase: usage.UseCaseIndex,
	})
	if len(orchestratorResp.ExpertResponses) == 0 {
		return
	}
	if messageID == uuid.Nil {
		h.logger.Warn("indexTurn skipped: nil message_id",
			zap.String("chat_id", chatID.String()),
			zap.Int("turn", turnNumber),
		)
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

	// Save to chat_index with a real messages.id (FK). Never swallow the error —
	// a silent fail left semantic history + rolling summaries dead forever.
	if err := h.chatSvc.IndexTurn(ctx, chatID, messageID, turnNumber, summary, topic, importance, embedding); err != nil {
		h.logger.Error("indexTurn insert failed",
			zap.String("chat_id", chatID.String()),
			zap.String("message_id", messageID.String()),
			zap.Int("turn", turnNumber),
			zap.Error(err),
		)
	}
}

// generateRollingSummary generates a rolling summary every 10 turns.
// WHY rolling summary (Arpit Bhiyani + Byte by Byte AI):
// Context window is finite. Summary compresses history.
// Prevents lost-in-middle problem for long conversations.
func (h *Handler) generateRollingSummary(ctx context.Context, chatID uuid.UUID, turnNumber int) {
	// C5: attribute the (cheap) summary call to this chat.
	ctx = usage.WithAttribution(ctx, usage.Attribution{
		ChatID: &chatID, UseCase: usage.UseCaseSummary,
	})
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

	prompt := buildRollingSummaryPrompt(summaries)

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

// buildRollingSummaryPrompt builds a PRESCRIPTIVE summarization prompt
// (B9, §3.1 P5 / Arpit last-video §5): a generic "summarize this" loses
// the details that matter. Name exactly what to keep. Pure — unit-tested.
func buildRollingSummaryPrompt(summaries []string) string {
	instructions := `Compress this conversation history into a compact, high-signal rolling summary.
Preserve EXACTLY these if present (do not paraphrase away):
- decisions made and their rationale
- concrete numbers, versions, IDs, limits, and rates
- constraints / must-not rules and tech choices locked
- open questions still blocking work
- commands, file paths, or config keys mentioned
Drop chit-chat, greetings, and restated questions.
Format as short labelled bullets. Omit any section that has nothing.

HISTORY:
` + strings.Join(summaries, "\n") + `

ROLLING SUMMARY:`
	return instructions
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

// DeleteMessage handles DELETE /messages/:id
func (h *Handler) DeleteMessage(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	msgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid message id")
		return
	}
	if err := h.chatSvc.DeleteMessage(c.Request.Context(), msgID, clientID); err != nil {
		if err == chat.ErrNotFound {
			response.NotFound(c, "Message")
			return
		}
		h.logger.Error("delete message failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"status": "deleted"})
}

// UpdateMessage handles PATCH /messages/:id
func (h *Handler) UpdateMessage(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	msgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid message id")
		return
	}
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}
	if err := h.chatSvc.UpdateMessageContent(c.Request.Context(), msgID, clientID, req.Content); err != nil {
		if err == chat.ErrNotFound {
			response.NotFound(c, "Message")
			return
		}
		h.logger.Error("update message failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, gin.H{"status": "updated"})
}
