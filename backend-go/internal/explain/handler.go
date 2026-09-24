package explain

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/tenant"
)

// Handler exposes the explainability endpoint.
type Handler struct {
	svc    *Service
	logger *zap.Logger
}

// NewHandler builds the explanation HTTP handler.
func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Get handles GET /messages/:id/explanation.
// Returns the unified "why this answer" view for a message the caller owns.
func (h *Handler) Get(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid message ID")
		return
	}
	uid, ok := c.MustGet("user_id").(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "invalid session")
		return
	}
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	// C4: resolve tenant scope (nil-safe → global). Failure = deny.
	scope, scopeErr := h.resolveScope(c, uid, roleStr)
	if scopeErr != nil {
		response.Forbidden(c, "Tenant scope could not be resolved")
		return
	}

	exp, err := h.svc.Get(c.Request.Context(), messageID, uid, scope)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "message")
			return
		}
		h.logger.Error("build explanation failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, exp)
}

// resolveScope delegates to the C4 tenant service (nil-safe).
func (h *Handler) resolveScope(c *gin.Context, uid uuid.UUID, role string) (tenant.Scope, error) {
	if h.svc.tenants == nil {
		return tenant.Scope{Global: true}, nil
	}
	return h.svc.tenants.Resolve(c.Request.Context(), uid, role)
}
