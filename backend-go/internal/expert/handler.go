package expert

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
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
	logger *zap.Logger
}

// NewHandler creates a new expert handler.
func NewHandler(db *pgxpool.Pool, logger *zap.Logger) *Handler {
	return &Handler{db: db, logger: logger}
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

	var e ExpertPublic
	// WHY training_status='trained': same as ListActive — only fully trained
	// experts are visible to clients. Draft/ingesting experts are admin-only.
	err = h.db.QueryRow(c.Request.Context(), `
		SELECT id, name, slug, domain, COALESCE(description,''),
		       total_chunks, total_topics,
		       COALESCE(avg_depth_level,0), COALESCE(avg_rating,0), created_at
		FROM experts
		WHERE id=$1
		  AND is_active=TRUE
		  AND is_training=FALSE
		  AND training_status='trained'
		  AND deleted_at IS NULL`, id,
	).Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
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
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT topic, depth_level, chunk_count, complexity_ceiling
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
		var depth, chunkCount int
		if err := rows.Scan(&topic, &depth, &chunkCount, &complexity); err != nil {
			continue
		}
		topics = append(topics, map[string]interface{}{
			"topic": topic, "depth_level": depth,
			"chunk_count": chunkCount, "complexity_ceiling": complexity,
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
