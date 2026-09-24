package byoexpert

import (
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/docextract"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/tenant"
)

// Handler exposes the tenant-facing BYO endpoints. Registered under
// /byo-experts on the JWT-protected group (all routes use STATIC path
// segments only, so there is no gin wildcard/static route conflict and no
// need to touch the existing expert handler).
type Handler struct {
	svc    *Service
	logger *zap.Logger
}

// NewHandler builds the BYO HTTP handler.
func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// RegisterRoutes mounts the BYO routes on the protected group.
//
//	POST /byo-experts           register a tenant-owned expert
//	GET  /byo-experts           list my tenant's byo experts (incl. drafts)
//	GET  /byo-experts/entitlement  my resolved entitlement
//	POST /byo-experts/ingest    self-service corpus ingest (multipart)
func (h *Handler) RegisterRoutes(g *gin.RouterGroup) {
	byo := g.Group("/byo-experts")
	byo.POST("", h.Register)
	byo.GET("", h.ListMine)
	byo.GET("/entitlement", h.GetEntitlement)
	byo.POST("/ingest", h.StartIngest)
}

// scope resolves the caller's tenant scope + user id. Returns false (and
// writes the error response) on failure — callers must return.
func (h *Handler) scope(c *gin.Context) (tenant.Scope, uuid.UUID, bool) {
	uid, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "invalid session")
		return tenant.Scope{}, uuid.Nil, false
	}
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	sc, err := h.svc.Resolve(c.Request.Context(), uid, roleStr)
	if err != nil {
		response.Forbidden(c, "Tenant scope could not be resolved")
		return tenant.Scope{}, uuid.Nil, false
	}
	return sc, uid, true
}

// Register POST /byo-experts  Body: {name,slug,domain,description}
func (h *Handler) Register(c *gin.Context) {
	sc, uid, ok := h.scope(c)
	if !ok {
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Domain      string `json:"domain" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	e, err := h.svc.Register(c.Request.Context(), sc, uid, RegisterInput{
		Name: req.Name, Slug: req.Slug, Domain: req.Domain, Description: req.Description,
	})
	if err != nil {
		h.respondErr(c, err)
		return
	}
	response.Created(c, e)
}

// ListMine GET /byo-experts
func (h *Handler) ListMine(c *gin.Context) {
	sc, _, ok := h.scope(c)
	if !ok {
		return
	}
	list, err := h.svc.ListMine(c.Request.Context(), sc)
	if err != nil {
		h.logger.Error("list byo experts failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, list)
}

// GetEntitlement GET /byo-experts/entitlement
func (h *Handler) GetEntitlement(c *gin.Context) {
	sc, _, ok := h.scope(c)
	if !ok {
		return
	}
	ent, err := h.svc.Entitlement(c.Request.Context(), sc)
	if err != nil {
		h.logger.Error("resolve byo entitlement failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, ent)
}

// StartIngest POST /byo-experts/ingest  multipart: expert_id + transcript file
func (h *Handler) StartIngest(c *gin.Context) {
	sc, uid, ok := h.scope(c)
	if !ok {
		return
	}
	expertID, err := uuid.Parse(c.PostForm("expert_id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "expert_id must be a valid UUID")
		return
	}
	file, header, err := c.Request.FormFile("transcript")
	if err != nil {
		response.BadRequest(c, "FILE_REQUIRED", "transcript file is required")
		return
	}
	defer file.Close()
	if header.Size > 50*1024*1024 {
		response.BadRequest(c, "FILE_TOO_LARGE", "file must be under 50MB")
		return
	}
	// D3: same format allowlist as the admin path — reject before creating a job
	// so an unusable upload cannot leave an orphan job row.
	if !docextract.IsSupported(header.Filename) {
		response.BadRequest(c, "UNSUPPORTED_FORMAT",
			fmt.Sprintf(".%s is not a supported format. Supported: %s",
				docextract.Extension(header.Filename), docextract.SupportedList()))
		return
	}
	content := make([]byte, header.Size)
	if _, err := io.ReadFull(file, content); err != nil {
		response.InternalError(c)
		return
	}

	jobID, err := h.svc.StartIngest(c.Request.Context(), sc, uid, expertID, header.Filename, content)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	response.Created(c, gin.H{
		"job_id":    jobID,
		"expert_id": expertID,
		"status":    "pending",
		"message":   "ingestion started in background",
	})
}

// respondErr maps the package's sentinel errors to HTTP status codes.
// Fail closed: anything unrecognised is a 500 (never a silent 200).
func (h *Handler) respondErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrDisabled):
		response.ServiceUnavailable(c, "byo experts are disabled")
	case errors.Is(err, ErrNotEntitled), errors.Is(err, ErrScopeRequired),
		errors.Is(err, ErrQuotaExceeded), errors.Is(err, ErrNotOwned):
		response.Forbidden(c, err.Error())
	case errors.Is(err, ErrSlugTaken):
		response.Conflict(c, err.Error())
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "expert")
	case errors.Is(err, ErrInvalidInput):
		response.BadRequest(c, "INVALID_INPUT", err.Error())
	default:
		h.logger.Error("byo request failed", zap.Error(err))
		response.InternalError(c)
	}
}
