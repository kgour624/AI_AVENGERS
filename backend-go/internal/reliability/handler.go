package reliability

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

// Handler exposes the public status surface + admin reliability API.
type Handler struct {
	svc     *Service
	version string
	logger  *zap.Logger
}

// NewHandler builds the reliability HTTP handler.
func NewHandler(svc *Service, version string, logger *zap.Logger) *Handler {
	if strings.TrimSpace(version) == "" {
		version = "1.0.0"
	}
	return &Handler{svc: svc, version: version, logger: logger}
}

// PublicStatus handles GET /status — no auth. Product-facing reliability view:
// dependency components + the published SLO snapshot. Error strings on
// unhealthy dependencies are withheld publicly (they can name internal hosts);
// the admin view carries them.
func (h *Handler) PublicStatus(c *gin.Context) {
	report := h.svc.Status(c.Request.Context(), h.version)
	public := StatusReport{
		Status:      report.Status,
		Version:     report.Version,
		SLO:         report.SLO,
		GeneratedAt: report.GeneratedAt,
	}
	public.Components = make([]DependencyStatus, 0, len(report.Components))
	for _, comp := range report.Components {
		public.Components = append(public.Components, DependencyStatus{
			Name:   comp.Name,
			Status: comp.Status,
			// Error deliberately omitted on the public surface.
		})
	}
	response.OK(c, public)
}

// AdminStatus handles GET /admin/reliability/status — full detail (incl. probe
// error strings + targets). Admin-only (mounted under adminGroup).
func (h *Handler) AdminStatus(c *gin.Context) {
	report := h.svc.Status(c.Request.Context(), h.version)
	report.SLO = ptrSnapshot(h.svc.Snapshot())
	response.OK(c, report)
}

// ListEvents handles GET /admin/reliability/events?kind=&limit=
func (h *Handler) ListEvents(c *gin.Context) {
	limit := 100
	if l := strings.TrimSpace(c.Query("limit")); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	events, err := h.svc.ListEvents(c.Request.Context(), strings.TrimSpace(c.Query("kind")), limit)
	if err != nil {
		h.logger.Error("list reliability events failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, events)
}

func ptrSnapshot(s SLISnapshot) *SLISnapshot { return &s }
