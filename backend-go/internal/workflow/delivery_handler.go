package workflow

// HTTP surface for the two delivery features:
//
//	POST /api/v1/workflows/:id/export/git    §18 — push the harness to a client remote
//	POST /api/v1/workflows/:id/code-feedback §17 — ingest the repo the client built
//
// A third handler struct in this package, alongside Handler and ChatHandler, for
// the same reason ChatHandler exists separately: these two endpoints need the
// exporter and the feedback service and nothing else, and adding two fields to
// Handler would make its eleven existing routes depend on both being
// constructed.

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// DeliveryHandler serves the export and code-feedback endpoints.
type DeliveryHandler struct {
	exporter *GitExporter
	feedback *CodeFeedbackService
	logger   *zap.Logger
}

// NewDeliveryHandler creates the handler.
func NewDeliveryHandler(exporter *GitExporter, feedback *CodeFeedbackService, logger *zap.Logger) *DeliveryHandler {
	return &DeliveryHandler{exporter: exporter, feedback: feedback, logger: logger}
}

// respondErr maps service errors to HTTP status codes in one place.
//
// Note what is NOT here: internal/repo's ErrNoRepoConnection and
// ErrProviderMismatch. Both arrive wrapped as ErrNoProviderToken
// (client_repo.go), which keeps the HTTP layer from importing internal/repo to
// name error values — and gives all three credential problems the one status
// they all deserve, since the client's fix is the same for each.
func (h *DeliveryHandler) respondErr(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, ErrWorkflowNotFound):
		response.NotFound(c, "workflow")
	case errors.Is(err, ErrNoHarness):
		response.BadRequest(c, "NO_HARNESS", err.Error())
	case errors.Is(err, ErrUnsupportedRemote):
		response.BadRequest(c, "UNSUPPORTED_REMOTE", err.Error())
	case errors.Is(err, ErrNoProviderToken):
		response.BadRequest(c, "NO_PROVIDER_TOKEN", err.Error())
	default:
		// Everything left is a git or provider failure. The message is already
		// written for the client (see GitExporter.pushFailure and
		// CodeFeedbackService.cloneClientRepo) and has been through
		// scrubSecrets, so it is returned rather than swallowed into a generic
		// 500 that would tell the client nothing about a fixable problem.
		h.logger.Error("workflow delivery: "+action, zap.Error(err))
		response.BadRequest(c, "DELIVERY_FAILED", err.Error())
	}
}

type gitExportBody struct {
	Provider   string `json:"provider" binding:"required"`
	RepoURL    string `json:"repo_url" binding:"required"`
	Branch     string `json:"branch"`
	CreateRepo bool   `json:"create_repo"`
	// Private is a pointer so an absent field means private. See
	// GitExportRequest.Private.
	Private *bool `json:"private"`
}

// ExportGit POST /api/v1/workflows/:id/export/git
//
// Synchronous: a push of a design harness is seconds of work, and the client
// wants the commit SHA and the repo URL back in the response rather than having
// to watch the blackboard for it.
func (h *DeliveryHandler) ExportGit(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	var body gitExportBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	result, err := h.exporter.Export(c.Request.Context(), GitExportRequest{
		WorkflowID: workflowID,
		ClientID:   clientID,
		Provider:   body.Provider,
		RepoURL:    body.RepoURL,
		Branch:     body.Branch,
		CreateRepo: body.CreateRepo,
		Private:    body.Private,
	})
	if err != nil {
		h.respondErr(c, err, "export harness")
		return
	}
	response.OK(c, result)
}

type codeFeedbackBody struct {
	Provider string `json:"provider" binding:"required"`
	RepoURL  string `json:"repo_url" binding:"required"`
	Branch   string `json:"branch"`
}

// IngestCodeFeedback POST /api/v1/workflows/:id/code-feedback
//
// Asynchronous, and the response says so. Cloning a repository and running its
// acceptance commands is minutes of work; the findings arrive as blackboard
// events, which is where every other workflow result already arrives.
//
// Everything the CLIENT can fix is still validated before this returns — not
// their workflow, unsupported remote, no harness, no criteria, no provider
// connected — so a bad request fails here with a reason instead of failing
// silently in a goroutine. See CodeFeedbackService.Start for that split.
func (h *DeliveryHandler) IngestCodeFeedback(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid workflow ID")
		return
	}

	var body codeFeedbackBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.feedback.Start(c.Request.Context(), CodeFeedbackRequest{
		WorkflowID: workflowID,
		ClientID:   clientID,
		Provider:   body.Provider,
		RepoURL:    body.RepoURL,
		Branch:     body.Branch,
	}); err != nil {
		h.respondErr(c, err, "ingest code feedback")
		return
	}

	response.OK(c, gin.H{
		"ingest_started": true,
		"note": "The comparison runs in the background. Results arrive on the blackboard as " +
			"code_feedback_ingested, with one design_amendment_proposed per finding.",
	})
}

// RegisterRoutes mounts both endpoints.
//
// The caller MUST pass the JWT-protected group. Both handlers read
// c.MustGet("user_id") for the ownership check, which panics on an
// unauthenticated group — and an export endpoint without an ownership check
// would push one client's design into another client's repository. The
// parameter is named `protected` so a wrong argument is visible at the call
// site, the same convention ChatHandler.RegisterRoutes uses.
//
// The group path and param name (`:id`) match the existing /workflows/:id
// routes deliberately: gin cannot have two different wildcard names at the same
// path position, so registering these under, say, `:wid` would break every
// other workflow route.
func (h *DeliveryHandler) RegisterRoutes(protected *gin.RouterGroup) {
	wf := protected.Group("/workflows/:id")
	{
		wf.POST("/export/git", h.ExportGit)
		wf.POST("/code-feedback", h.IngestCodeFeedback)
	}
}
