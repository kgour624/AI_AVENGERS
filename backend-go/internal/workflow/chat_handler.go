package workflow

// HTTP surface for the workflow chat (§6).
//
// A separate handler struct from workflow.Handler on purpose: this one needs the
// chat service and nothing else, and workflow.Handler is already wired into
// eleven routes. Adding a field there would make every workflow route depend on
// the chat service being constructed.

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// ChatHandler serves the workflow chat endpoints.
type ChatHandler struct {
	svc    *WorkflowChatService
	logger *zap.Logger
}

// NewChatHandler creates the handler.
func NewChatHandler(svc *WorkflowChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{svc: svc, logger: logger}
}

// respondErr maps service errors onto HTTP status codes in one place, so every
// endpoint answers the same way for the same cause.
//
// ErrChatNotFound covers "not found" and "belongs to another client" together —
// the service merges them deliberately (telling a caller a chat exists but is
// someone else's leaks other clients' workflows), and the handler must not
// undo that by answering differently.
func (h *ChatHandler) respondErr(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, ErrChatNotFound):
		response.NotFound(c, "workflow chat")
	case errors.Is(err, ErrExpertNotFound):
		response.BadRequest(c, "EXPERT_NOT_FOUND", err.Error())
	case errors.Is(err, ErrNotParticipant):
		response.BadRequest(c, "NOT_PARTICIPANT",
			"that expert is not part of this conversation — add it first")
	case errors.Is(err, ErrNoResponder):
		response.BadRequest(c, "NO_RESPONDER",
			"this chat has no expert to answer — add a participant")
	default:
		h.logger.Error("workflow chat: "+action, zap.Error(err))
		response.InternalError(c)
	}
}

// ============================================================
// Chats
// ============================================================

type createChatRequest struct {
	// PinnedEventID is the deliverable this chat is about. Optional: omit for a
	// conversation about the workflow as a whole.
	PinnedEventID string `json:"pinned_event_id"`
	Title         string `json:"title"`
}

// CreateChat POST /api/v1/workflows/:id/chats
func (h *ChatHandler) CreateChat(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	var req createChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	// Empty string is valid and means "not pinned" — only a non-empty value is
	// parsed, so a malformed id is rejected rather than silently ignored.
	var pinned *uuid.UUID
	if req.PinnedEventID != "" {
		id, err := uuid.Parse(req.PinnedEventID)
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid pinned_event_id")
			return
		}
		pinned = &id
	}

	ch, err := h.svc.CreateChat(c.Request.Context(), workflowID, clientID, pinned, req.Title)
	if err != nil {
		h.respondErr(c, err, "create chat")
		return
	}
	response.Created(c, ch)
}

// ListChats GET /api/v1/workflows/:id/chats
func (h *ChatHandler) ListChats(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	chats, err := h.svc.ListChats(c.Request.Context(), workflowID, clientID)
	if err != nil {
		h.respondErr(c, err, "list chats")
		return
	}
	response.OK(c, gin.H{"chats": chats})
}

// GetChat GET /api/v1/workflow-chats/:cid
//
// Participants are returned alongside the chat: the UI needs both to render the
// conversation header, and two round trips for one screen is waste.
func (h *ChatHandler) GetChat(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}

	ch, err := h.svc.GetChat(c.Request.Context(), chatID, clientID)
	if err != nil {
		h.respondErr(c, err, "get chat")
		return
	}
	participants, err := h.svc.ListParticipants(c.Request.Context(), chatID)
	if err != nil {
		h.respondErr(c, err, "list participants")
		return
	}
	response.OK(c, gin.H{"chat": ch, "participants": participants})
}

// ============================================================
// Knowledge mode (§6.4)
// ============================================================

type knowledgeModeRequest struct {
	// GenericAllowancePct nil clears the chat override so the workflow's value
	// applies again. 0 means trained-only. A value above the configured ceiling
	// is clamped by the service, not rejected.
	GenericAllowancePct *float64 `json:"generic_allowance_pct"`
}

// SetKnowledgeMode PATCH /api/v1/workflow-chats/:cid/knowledge-mode
func (h *ChatHandler) SetKnowledgeMode(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}

	var req knowledgeModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	ch, err := h.svc.SetKnowledgeMode(c.Request.Context(), chatID, clientID, req.GenericAllowancePct)
	if err != nil {
		h.respondErr(c, err, "set knowledge mode")
		return
	}
	response.OK(c, ch)
}

// ============================================================
// Participants (§6.3)
// ============================================================

type participantRequest struct {
	ExpertID string `json:"expert_id" binding:"required"`
}

// AddParticipant POST /api/v1/workflow-chats/:cid/participants
func (h *ChatHandler) AddParticipant(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}

	var req participantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	expertID, err := uuid.Parse(req.ExpertID)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert_id")
		return
	}

	if err := h.svc.AddParticipant(c.Request.Context(), chatID, clientID, expertID); err != nil {
		h.respondErr(c, err, "add participant")
		return
	}
	participants, err := h.svc.ListParticipants(c.Request.Context(), chatID)
	if err != nil {
		h.respondErr(c, err, "list participants")
		return
	}
	response.OK(c, gin.H{"participants": participants})
}

// RemoveParticipant DELETE /api/v1/workflow-chats/:cid/participants/:eid
func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}
	expertID, err := uuid.Parse(c.Param("eid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	if err := h.svc.RemoveParticipant(c.Request.Context(), chatID, clientID, expertID); err != nil {
		h.respondErr(c, err, "remove participant")
		return
	}
	response.OK(c, gin.H{"removed": expertID})
}

// ============================================================
// Messages
// ============================================================

type sendMessageRequest struct {
	Message string `json:"message" binding:"required"`
	// ExpertID addresses one participant. Empty means the default responder:
	// the pinned deliverable's author (§6.3). Every participant is NOT asked by
	// default — that would be one retrieval and one completion per participant
	// per message.
	ExpertID string `json:"expert_id"`
}

// Send POST /api/v1/workflow-chats/:cid/messages
func (h *ChatHandler) Send(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	var expertID *uuid.UUID
	if req.ExpertID != "" {
		id, err := uuid.Parse(req.ExpertID)
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid expert_id")
			return
		}
		expertID = &id
	}

	msg, err := h.svc.Send(c.Request.Context(), chatID, clientID, expertID, req.Message)
	if err != nil {
		h.respondErr(c, err, "send message")
		return
	}
	response.OK(c, msg)
}

// ListMessages GET /api/v1/workflow-chats/:cid/messages
func (h *ChatHandler) ListMessages(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	chatID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid chat ID")
		return
	}

	// limit is optional; the service clamps anything out of range.
	limit := 0
	if v := c.Query("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			limit = n
		}
	}

	msgs, err := h.svc.ListMessages(c.Request.Context(), chatID, clientID, limit)
	if err != nil {
		h.respondErr(c, err, "list messages")
		return
	}
	response.OK(c, gin.H{"messages": msgs})
}

// ProposeChange POST /api/v1/workflow-chats/:cid/propose-change
//
// Client sends a free-text change goal. The service stores it and posts a
// blackboard event. The runner picks it up and re-runs the relevant design
// sections. The client is then asked to approve the updated design.
//
// Body: {"change_goal": "add bulk move endpoint"}
func (h *ChatHandler) ProposeChange(crSvc *ChangeRequestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.MustGet("user_id").(uuid.UUID)
		chatID, err := uuid.Parse(c.Param("cid"))
		if err != nil {
			response.BadRequest(c, "INVALID_ID", "invalid chat ID")
			return
		}

		var req struct {
			ChangeGoal string `json:"change_goal" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "INVALID_INPUT", err.Error())
			return
		}

		cr, err := h.svc.ProposeChange(c.Request.Context(), chatID, clientID, req.ChangeGoal, crSvc)
		if err != nil {
			h.respondErr(c, err, "propose change")
			return
		}
		response.Created(c, cr)
	}
}

// parsePositiveInt parses a small positive integer without pulling in strconv
// error handling at every call site. Returns an error for anything else, which
// callers treat as "use the default".
func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, errors.New("not a number")
		}
		n = n*10 + int(r-'0')
		if n > 100000 { // guard against an absurd query string
			return 0, errors.New("too large")
		}
	}
	if n == 0 {
		return 0, errors.New("not positive")
	}
	return n, nil
}

// RegisterRoutes mounts the workflow chat endpoints.
//
// Two groups because the ids differ: creating and listing chats is scoped to a
// workflow (:id), while everything afterwards is scoped to a chat (:cid). Using
// :id for both inside one gin group would collide — gin cannot have two
// different wildcard names at the same path position.
//
// The caller MUST pass the JWT-protected group, not the bare /api/v1 group.
// Every handler here reads c.MustGet("user_id"), which AuthMiddleware sets — on
// an unauthenticated group that panics, and worse, the ownership checks these
// endpoints rely on would have no identity to check against. The parameter is
// named `protected` so a wrong argument is visible at the call site.
func (h *ChatHandler) RegisterRoutes(protected *gin.RouterGroup) {
	wf := protected.Group("/workflows/:id")
	{
		wf.POST("/chats", h.CreateChat)
		wf.GET("/chats", h.ListChats)
	}

	ch := protected.Group("/workflow-chats/:cid")
	{
		ch.GET("", h.GetChat)
		ch.PATCH("/knowledge-mode", h.SetKnowledgeMode)
		ch.POST("/participants", h.AddParticipant)
		ch.DELETE("/participants/:eid", h.RemoveParticipant)
		ch.POST("/messages", h.Send)
		ch.GET("/messages", h.ListMessages)
	}
}

// RegisterRoutesWithCR is RegisterRoutes with ChangeRequestService injected.
// Call this from main.go instead of RegisterRoutes when the change-request
// feature is wired in.
func (h *ChatHandler) RegisterRoutesWithCR(protected *gin.RouterGroup, crSvc *ChangeRequestService) {
	wf := protected.Group("/workflows/:id")
	{
		wf.POST("/chats", h.CreateChat)
		wf.GET("/chats", h.ListChats)
	}

	ch := protected.Group("/workflow-chats/:cid")
	{
		ch.GET("", h.GetChat)
		ch.PATCH("/knowledge-mode", h.SetKnowledgeMode)
		ch.POST("/participants", h.AddParticipant)
		ch.DELETE("/participants/:eid", h.RemoveParticipant)
		ch.POST("/messages", h.Send)
		ch.GET("/messages", h.ListMessages)
		// Coordinated redesign: client proposes a change goal from the chat.
		ch.POST("/propose-change", h.ProposeChange(crSvc))
	}
}
