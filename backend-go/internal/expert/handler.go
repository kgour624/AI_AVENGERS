package expert

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/tenant"
)

// ExpertPublic is the client-facing view of an expert.
type ExpertPublic struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Domain        string    `json:"domain"`
	Description   string    `json:"description"`
	TotalChunks   int       `json:"total_chunks"`
	TotalTopics   int       `json:"total_topics"`
	AvgDepthLevel float64   `json:"avg_depth_level"`
	AvgRating     float64   `json:"avg_rating"`
	CreatedAt     time.Time `json:"created_at"`
}

// Handler handles public expert endpoints.
type Handler struct {
	db     *pgxpool.Pool
	tenant *tenant.Service // C4: nil-safe — unwired → global scope, no filter
	logger *zap.Logger
}

// NewHandler creates a new expert handler.
func NewHandler(db *pgxpool.Pool, tenantSvc *tenant.Service, logger *zap.Logger) *Handler {
	return &Handler{db: db, tenant: tenantSvc, logger: logger}
}

// ListActive GET /experts
// Returns only experts that are active AND fully trained.
// WHY training_status='trained' filter:
//   Migration 006 adds training_status (default 'draft').
//   An expert is only ready for client use after ingestion pipeline
//   completes smoke test and sets training_status='trained' (A10).
//   Without this filter, draft experts with no chunks appear publicly
//   and return REFUSE on every question — bad client experience.
// WHY keep is_training=FALSE:
//   Belt-and-suspenders. Existing code sets is_training=TRUE during
//   ingestion. Both conditions must be true for public visibility.
func (h *Handler) ListActive(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	userID, _ := c.Get("user_id")
	uid, _ := userID.(uuid.UUID)

	// domain_expert: only granted experts; admin/client: full trained catalog.
	query := `
		SELECT id, name, slug, domain, COALESCE(description,''),
		       total_chunks, total_topics,
		       COALESCE(avg_depth_level,0), COALESCE(avg_rating,0), created_at
		FROM experts
		WHERE is_active=TRUE
		  AND is_training=FALSE
		  AND training_status='trained'
		  AND deleted_at IS NULL`
	args := []interface{}{}
	if roleStr == "domain_expert" {
		query += `
		  AND id IN (SELECT expert_id FROM user_expert_grants WHERE user_id = $1)`
		args = append(args, uid)
	}
	// C4: only experts visible to the caller's tenant. Platform experts
	// (tenant_id NULL) stay visible to everyone. Fail closed on an
	// unresolvable scope rather than leaking the whole catalog.
	scope, scopeErr := h.tenant.Resolve(c.Request.Context(), uid, roleStr)
	if scopeErr != nil {
		response.Forbidden(c, "Tenant scope could not be resolved")
		return
	}
	if !scope.Global {
		if scope.TenantID == nil {
			response.Forbidden(c, "Tenant scope could not be resolved")
			return
		}
		query += fmt.Sprintf(" AND (tenant_id IS NULL OR tenant_id = $%d)", len(args)+1)
		args = append(args, *scope.TenantID)
	}
	query += `
		ORDER BY avg_rating DESC, total_chunks DESC`

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		h.logger.Error("list experts failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	var experts []ExpertPublic
	for rows.Next() {
		var e ExpertPublic
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
			&e.TotalChunks, &e.TotalTopics, &e.AvgDepthLevel, &e.AvgRating, &e.CreatedAt); err != nil {
			continue
		}
		experts = append(experts, e)
	}
	if experts == nil {
		experts = []ExpertPublic{}
	}
	response.OK(c, experts)
}

// GetByID GET /experts/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr == "domain_expert" {
		userID := c.MustGet("user_id").(uuid.UUID)
		var granted bool
		_ = h.db.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM user_expert_grants WHERE user_id=$1 AND expert_id=$2)`,
			userID, id,
		).Scan(&granted)
		if !granted {
			response.NotFound(c, "expert")
			return
		}
	}

	// C4: hide experts owned by another tenant. Fail closed / NotFound so
	// the caller cannot probe for the existence of other tenants' experts.
	scope, scopeErr := h.tenant.Resolve(c.Request.Context(), c.MustGet("user_id").(uuid.UUID), roleStr)
	if scopeErr != nil {
		response.NotFound(c, "expert")
		return
	}
	q := `
		SELECT id, name, slug, domain, COALESCE(description,''),
		       total_chunks, total_topics,
		       COALESCE(avg_depth_level,0), COALESCE(avg_rating,0), created_at
		FROM experts
		WHERE id=$1
		  AND is_active=TRUE
		  AND is_training=FALSE
		  AND training_status='trained'
		  AND deleted_at IS NULL`
	qargs := []interface{}{id}
	if !scope.Global {
		if scope.TenantID == nil {
			response.NotFound(c, "expert")
			return
		}
		q += ` AND (tenant_id IS NULL OR tenant_id = $2)`
		qargs = append(qargs, *scope.TenantID)
	}

	var e ExpertPublic
	// WHY training_status='trained': same as ListActive — only fully trained
	// experts are visible to clients. Draft/ingesting experts are admin-only.
	err = h.db.QueryRow(c.Request.Context(), q, qargs...).Scan(
		&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
		&e.TotalChunks, &e.TotalTopics, &e.AvgDepthLevel, &e.AvgRating, &e.CreatedAt)
	if err != nil {
		response.NotFound(c, "expert")
		return
	}
	response.OK(c, e)
}

// GetTopics GET /experts/:id/topics
func (h *Handler) GetTopics(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	// I2: this returns BOTH the declared coverage band and the measured depth, so
	// the caller can show them side by side. They are different scales and must
	// never be compared numerically - see migration 045's column comments.
	//
	// can_handle / cannot_handle used to be omitted here ("backend gap"): they were
	// generated from a few hundred characters per topic, so returning them would
	// have published a guess. They are now overwritten by the measured result of an
	// actual evaluation pass, which is why they are returned with eval_cases so the
	// caller can tell a measured list from the old declared one.
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT topic, depth_level, chunk_count, complexity_ceiling,
		       COALESCE(measured_level,0), COALESCE(eval_cases,0), COALESCE(eval_passed,0),
		       COALESCE(can_handle,'{}'), COALESCE(cannot_handle,'{}'),
		       last_evaluated_at
		FROM expert_capabilities
		WHERE expert_id=$1
		ORDER BY depth_level DESC, chunk_count DESC`, id)
	if err != nil {
		h.logger.Error("get topics failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	var topics []map[string]interface{}
	for rows.Next() {
		var topic, complexity string
		var depth, chunkCount, measuredLevel, evalCases, evalPassed int
		var canHandle, cannotHandle []string
		var lastEvaluatedAt *time.Time
		if err := rows.Scan(&topic, &depth, &chunkCount, &complexity,
			&measuredLevel, &evalCases, &evalPassed, &canHandle, &cannotHandle,
			&lastEvaluatedAt); err != nil {
			continue
		}
		topics = append(topics, map[string]interface{}{
			"topic": topic, "depth_level": depth,
			"chunk_count": chunkCount, "complexity_ceiling": complexity,
			"measured_level":    measuredLevel,
			"eval_cases":        evalCases,
			"eval_passed":       evalPassed,
			"can_handle":        canHandle,
			"cannot_handle":     cannotHandle,
			"last_evaluated_at": lastEvaluatedAt,
		})
	}
	if topics == nil {
		topics = []map[string]interface{}{}
	}
	response.OK(c, map[string]interface{}{
		"expert_id": id,
		"topics":    topics,
		"total":     fmt.Sprintf("%d", len(topics)),
	})
}
