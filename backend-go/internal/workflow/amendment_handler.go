package workflow

// HTTP surface for §7.5's approval step:
//
//	GET  /api/v1/workflows/:id/amendments              list, pending by default
//	POST /api/v1/workflows/:id/amendments/:aid/respond approve / approve_with_edit / reject
//
// WHY THESE ARE NOT THE EXISTING APPROVAL ENDPOINT
//
// There is already POST /workflows/:id/approvals/:aid/respond (handler.go), and
// an amendment IS an approval_requests row, so reusing it looks obvious. It does
// not work, for three concrete reasons:
//
//  1. That handler's body carries only {decision, notes, generic_allowance_pct}.
//     §7.5 requires edit-then-approve to be first-class, and there is nowhere in
//     that shape to put the edited text.
//  2. On approve it calls Engine.Resume, which moves the WORKFLOW from
//     paused_for_approval to running. An amendment must not touch the workflow's
//     run state — §6.1's separation, and the same reason toolAskClient does not
//     call Tools.AskClient. Approving an amendment on a completed workflow would
//     otherwise restart it.
//  3. There is no list endpoint at all. Approval identity reaches the frontend
//     only through the kanban SSE stream, which fires on question_to_client and
//     is rendered only while the workflow is paused_for_approval. An amendment
//     raises no such event and pauses nothing, so without a list there is no way
//     for a client to discover one exists.
//
// So: a separate pair of endpoints over the same table, with the gate_name
// discriminator doing exactly the job it exists for.

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// AmendmentHandler serves the amendment list and decision endpoints.
type AmendmentHandler struct {
	svc    *AmendmentService
	logger *zap.Logger
}

// NewAmendmentHandler creates the handler.
func NewAmendmentHandler(svc *AmendmentService, logger *zap.Logger) *AmendmentHandler {
	return &AmendmentHandler{svc: svc, logger: logger}
}

// respondErr maps service errors to status codes in one place.
func (h *AmendmentHandler) respondErr(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
	case errors.Is(err, ErrAmendmentNotFound):
		response.NotFound(c, "amendment")
	case errors.Is(err, ErrAmendmentResolved):
		// 409, not 400: a second click on Approve is not a malformed request,
		// and the client's UI should treat it as "already done".
		response.Conflict(c, err.Error())
	case errors.Is(err, ErrDesignBusy):
		response.Conflict(c, err.Error())
	case errors.Is(err, ErrAmendmentStale):
		response.Conflict(c, err.Error())
	case errors.Is(err, ErrBadAmendmentTarget):
		response.BadRequest(c, "BAD_TARGET", err.Error())
	case errors.Is(err, ErrNoHarness):
		response.BadRequest(c, "NO_HARNESS", err.Error())
	default:
		h.logger.Error("amendment: "+action, zap.Error(err))
		response.BadRequest(c, "AMENDMENT_FAILED", err.Error())
	}
}

// List GET /api/v1/workflows/:id/amendments
func (h *AmendmentHandler) List(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	status := c.Query("status")
	switch status {
	case "", "pending", "approved", "rejected", "all":
	default:
		response.BadRequest(c, "INVALID_STATUS",
			"status must be pending, approved, rejected or all")
		return
	}

	items, err := h.svc.List(c.Request.Context(), workflowID, clientID, status)
	if err != nil {
		h.respondErr(c, err, "list")
		return
	}
	response.OK(c, gin.H{"amendments": items})
}

type amendmentRespondBody struct {
	Decision   string `json:"decision" binding:"required"`
	EditedText string `json:"edited_text"`
	Notes      string `json:"notes"`
}

// Respond POST /api/v1/workflows/:id/amendments/:aid/respond
//
// Synchronous, because applying an amendment is a string replacement, a file
// write and one commit — milliseconds. The response carries the commit SHA and
// the decision-log id, so the client sees the design actually changed rather
// than being told to go and watch a stream.
func (h *AmendmentHandler) Respond(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}
	approvalID, err := uuid.Parse(c.Param("aid"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid amendment ID")
		return
	}

	var body amendmentRespondBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	outcome, err := h.svc.Respond(c.Request.Context(), workflowID, approvalID, clientID, AmendmentDecision{
		Decision:   body.Decision,
		EditedText: body.EditedText,
		Notes:      body.Notes,
	})
	if err != nil {
		h.respondErr(c, err, "respond")
		return
	}
	response.OK(c, outcome)
}

// RegisterRoutes mounts both endpoints.
//
// Must be the JWT-protected group: both handlers read c.MustGet("user_id") and
// check workflow ownership against it. Without that check any authenticated user
// could approve a write into another client's design. `:id` matches the existing
// /workflows/:id routes so gin's tree merges rather than reporting a wildcard
// conflict.
func (h *AmendmentHandler) RegisterRoutes(protected *gin.RouterGroup) {
	wf := protected.Group("/workflows/:id")
	{
		wf.GET("/amendments", h.List)
		wf.POST("/amendments/:aid/respond", h.Respond)
	}
}
