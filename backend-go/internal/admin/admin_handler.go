package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/ml"
	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/training"
)

// AdminHandler handles all admin panel HTTP requests.
// All routes require admin role (enforced by AdminMiddleware).
type AdminHandler struct {
	db         *pgxpool.Pool
	gateway    *gateway.ModelGateway
	mlClient   *ml.SidecarClient
	ingestion  *training.IngestionPipeline
	logger     *zap.Logger
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(
	db *pgxpool.Pool,
	gw *gateway.ModelGateway,
	mlClient *ml.SidecarClient,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		db:        db,
		gateway:   gw,
		mlClient:  mlClient,
		ingestion: training.NewIngestionPipeline(db, mlClient, gw, logger),
		logger:    logger,
	}
}

// ============================================================
// EXPERT MANAGEMENT
// ============================================================

// ListExperts GET /admin/experts
func (h *AdminHandler) ListExperts(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, name, slug, domain, COALESCE(description,''),
		       total_chunks, total_topics, COALESCE(avg_depth_level,0),
		       COALESCE(avg_rating,0), total_ratings,
		       is_active, is_training, created_at, updated_at
		FROM experts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		h.logger.Error("list experts failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type expertRow struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Slug         string    `json:"slug"`
		Domain       string    `json:"domain"`
		Description  string    `json:"description"`
		TotalChunks  int       `json:"total_chunks"`
		TotalTopics  int       `json:"total_topics"`
		AvgDepth     float64   `json:"avg_depth_level"`
		AvgRating    float64   `json:"avg_rating"`
		TotalRatings int       `json:"total_ratings"`
		IsActive     bool      `json:"is_active"`
		IsTraining   bool      `json:"is_training"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	var experts []expertRow
	for rows.Next() {
		var e expertRow
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
			&e.TotalChunks, &e.TotalTopics, &e.AvgDepth,
			&e.AvgRating, &e.TotalRatings,
			&e.IsActive, &e.IsTraining, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			continue
		}
		experts = append(experts, e)
	}
	if experts == nil {
		experts = []expertRow{}
	}
	response.OK(c, experts)
}

// CreateExpert POST /admin/experts
func (h *AdminHandler) CreateExpert(c *gin.Context) {
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

	var id uuid.UUID
	err := h.db.QueryRow(c.Request.Context(),
		`INSERT INTO experts (name, slug, domain, description)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		req.Name, req.Slug, req.Domain, req.Description,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			response.Conflict(c, "slug already exists")
			return
		}
		h.logger.Error("create expert failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, map[string]interface{}{"id": id, "slug": req.Slug})
}

// UpdateExpert PATCH /admin/experts/:id
func (h *AdminHandler) UpdateExpert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	var req struct {
		Name             *string `json:"name"`
		Description      *string `json:"description"`
		IsActive         *bool   `json:"is_active"`
		ReasoningCharter *string `json:"reasoning_charter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.Name != nil {
		_, _ = h.db.Exec(c.Request.Context(),
			`UPDATE experts SET name=$1, updated_at=NOW() WHERE id=$2`, *req.Name, id)
	}
	if req.IsActive != nil {
		_, _ = h.db.Exec(c.Request.Context(),
			`UPDATE experts SET is_active=$1, updated_at=NOW() WHERE id=$2`, *req.IsActive, id)
	}
	if req.ReasoningCharter != nil {
		_, _ = h.db.Exec(c.Request.Context(),
			`UPDATE experts SET reasoning_charter=$1, updated_at=NOW() WHERE id=$2`, *req.ReasoningCharter, id)
	}
	response.OK(c, map[string]string{"status": "updated"})
}

// IngestTranscript POST /admin/experts/:id/ingest
// Accepts multipart form with "transcript" file.
// Starts background ingestion job.
func (h *AdminHandler) IngestTranscript(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	// Get expert name
	var expertName string
	h.db.QueryRow(c.Request.Context(),
		`SELECT name FROM experts WHERE id=$1 AND deleted_at IS NULL`, expertID,
	).Scan(&expertName)
	if expertName == "" {
		response.NotFound(c, "expert")
		return
	}

	// Parse file
	file, header, err := c.Request.FormFile("transcript")
	if err != nil {
		response.BadRequest(c, "FILE_REQUIRED", "transcript file is required")
		return
	}
	defer file.Close()

	// Validate
	if header.Size > 50*1024*1024 {
		response.BadRequest(c, "FILE_TOO_LARGE", "file must be under 50MB")
		return
	}

	content := make([]byte, header.Size)
	if _, err := file.Read(content); err != nil {
		response.InternalError(c)
		return
	}

	// Create ingestion job
	var jobID uuid.UUID
	err = h.db.QueryRow(c.Request.Context(),
		`INSERT INTO ingestion_jobs (expert_id, job_type, status, source_path)
		 VALUES ($1, 'transcript', 'pending', $2)
		 RETURNING id`,
		expertID, header.Filename,
	).Scan(&jobID)
	if err != nil {
		h.logger.Error("create ingestion job failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Mark expert as training
	_, _ = h.db.Exec(c.Request.Context(),
		`UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`, expertID)

	// Start background ingestion
	// WHY goroutine: Ingestion takes minutes. Client gets job ID immediately.
	// WHY replaceExisting=false: append mode per DOMAIN_EXPERT_COLLABORATION_DESIGN.md §5.4.
	// Admin uploading a new transcript adds to the corpus; full-retrain (true) is a
	// separate explicit operation reserved for Phase B.
	go func() {
		_, err := h.ingestion.IngestTranscript(
			context.Background(),
			jobID, expertID, expertName,
			string(content), header.Filename,
			false, // replaceExisting=false → append mode
		)
		if err != nil {
			h.logger.Error("ingestion failed",
				zap.String("job_id", jobID.String()),
				zap.Error(err),
			)
		}
	}()

	response.Created(c, map[string]interface{}{
		"job_id":    jobID,
		"status":    "pending",
		"message":   "ingestion started in background",
		"expert_id": expertID,
	})
}

// GetIngestionJobs GET /admin/experts/:id/jobs
func (h *AdminHandler) GetIngestionJobs(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, status, COALESCE(source_path,''),
		       total_chunks, processed_chunks,
		       COALESCE(error_message,''), started_at, completed_at, created_at
		FROM ingestion_jobs
		WHERE expert_id=$1
		ORDER BY created_at DESC LIMIT 20`, expertID)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type jobRow struct {
		ID              uuid.UUID  `json:"id"`
		Status          string     `json:"status"`
		SourcePath      string     `json:"source_path"`
		TotalChunks     int        `json:"total_chunks"`
		ProcessedChunks int        `json:"processed_chunks"`
		ErrorMessage    string     `json:"error_message,omitempty"`
		StartedAt       *time.Time `json:"started_at"`
		CompletedAt     *time.Time `json:"completed_at"`
		CreatedAt       time.Time  `json:"created_at"`
	}
	var jobs []jobRow
	for rows.Next() {
		var j jobRow
		if err := rows.Scan(
			&j.ID, &j.Status, &j.SourcePath,
			&j.TotalChunks, &j.ProcessedChunks,
			&j.ErrorMessage, &j.StartedAt, &j.CompletedAt, &j.CreatedAt,
		); err != nil {
			continue
		}
		jobs = append(jobs, j)
	}
	if jobs == nil {
		jobs = []jobRow{}
	}
	response.OK(c, jobs)
}

// ============================================================
// CLIENT MANAGEMENT
// ============================================================

// ListClients GET /admin/clients
//
// Feature #7 fix (docs bug list): previously returned only
// id/email/full_name/is_active/last_login/created_at - no way to see
// a client's activity at all. Added project_count and message_count
// as correlated subqueries rather than fabricating them on the
// frontend (there was no data to fabricate FROM). message_count only
// counts role='user' rows (messages the client actually sent), not
// assistant responses, so it reads as "how many times has this client
// asked something" rather than double-counting both sides of a turn.
func (h *AdminHandler) ListClients(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT u.id, u.email, u.full_name, u.is_active, u.last_login, u.created_at,
		       COALESCE((
		           SELECT COUNT(*) FROM projects p
		           WHERE p.client_id = u.id AND p.deleted_at IS NULL
		       ), 0) AS project_count,
		       COALESCE((
		           SELECT COUNT(*) FROM messages m
		           JOIN chats c2 ON c2.id = m.chat_id
		           JOIN projects p2 ON p2.id = c2.project_id
		           WHERE p2.client_id = u.id AND m.role = 'user'
		       ), 0) AS message_count
		FROM users u
		WHERE u.role='client' AND u.deleted_at IS NULL
		ORDER BY u.created_at DESC`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type clientRow struct {
		ID           uuid.UUID  `json:"id"`
		Email        string     `json:"email"`
		FullName     string     `json:"full_name"`
		IsActive     bool       `json:"is_active"`
		LastLogin    *time.Time `json:"last_login"`
		CreatedAt    time.Time  `json:"created_at"`
		ProjectCount int        `json:"project_count"`
		MessageCount int        `json:"message_count"`
	}
	var clients []clientRow
	for rows.Next() {
		var cl clientRow
		if err := rows.Scan(
			&cl.ID, &cl.Email, &cl.FullName, &cl.IsActive, &cl.LastLogin, &cl.CreatedAt,
			&cl.ProjectCount, &cl.MessageCount,
		); err != nil {
			continue
		}
		clients = append(clients, cl)
	}
	if clients == nil {
		clients = []clientRow{}
	}
	response.OK(c, clients)
}

// UpdateClient PATCH /admin/clients/:id
func (h *AdminHandler) UpdateClient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid client ID")
		return
	}
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.IsActive != nil {
		_, _ = h.db.Exec(c.Request.Context(),
			`UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2 AND role='client'`,
			*req.IsActive, id)
	}
	response.OK(c, map[string]string{"status": "updated"})
}

// ============================================================
// SYSTEM STATS
// ============================================================

// GetStats GET /admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	var totalExperts, activeExperts int
	h.db.QueryRow(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active=TRUE) FROM experts WHERE deleted_at IS NULL`).Scan(&totalExperts, &activeExperts)

	var totalClients int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role='client' AND deleted_at IS NULL`).Scan(&totalClients)

	var totalProjects int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE deleted_at IS NULL`).Scan(&totalProjects)

	var totalMessages int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM messages`).Scan(&totalMessages)

	var totalChunks int
	h.db.QueryRow(ctx, `SELECT COALESCE(SUM(total_chunks),0) FROM experts`).Scan(&totalChunks)

	var violationCount int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM master_event_log WHERE event_type='china_wall_violation'`).Scan(&violationCount)

	var avgRating float64
	h.db.QueryRow(ctx, `SELECT COALESCE(AVG(score),0) FROM ratings`).Scan(&avgRating)

	gwStats := h.gateway.GetStats()

	response.OK(c, map[string]interface{}{
		"experts": map[string]int{
			"total":  totalExperts,
			"active": activeExperts,
		},
		"clients":          totalClients,
		"projects":         totalProjects,
		"messages":         totalMessages,
		"total_chunks":     totalChunks,
		"violations":       violationCount,
		"avg_rating":       fmt.Sprintf("%.2f", avgRating),
		"llm_total_calls":  gwStats["total_calls"],
		"llm_total_cost":   gwStats["total_cost"],
	})
}

// GetViolations GET /admin/violations
func (h *AdminHandler) GetViolations(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT id, project_id, expert_id, client_id,
		       event_type, COALESCE(reasoning,''), created_at
		FROM master_event_log
		WHERE event_type='china_wall_violation'
		ORDER BY created_at DESC
		LIMIT 100`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type violationRow struct {
		ID        int64      `json:"id"`
		ProjectID uuid.UUID  `json:"project_id"`
		ExpertID  *uuid.UUID `json:"expert_id"`
		ClientID  uuid.UUID  `json:"client_id"`
		EventType string     `json:"event_type"`
		Reasoning string     `json:"reasoning"`
		CreatedAt time.Time  `json:"created_at"`
	}
	var violations []violationRow
	for rows.Next() {
		var v violationRow
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.ExpertID, &v.ClientID,
			&v.EventType, &v.Reasoning, &v.CreatedAt); err != nil {
			continue
		}
		violations = append(violations, v)
	}
	if violations == nil {
		violations = []violationRow{}
	}
	response.OK(c, violations)
}

// GetRatings GET /admin/ratings
func (h *AdminHandler) GetRatings(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT e.name, e.domain,
		       COUNT(r.id) as total_ratings,
		       ROUND(AVG(r.score)::numeric, 2) as avg_score,
		       COUNT(*) FILTER (WHERE r.score >= 4) as good_ratings,
		       COUNT(*) FILTER (WHERE r.score <= 2) as bad_ratings
		FROM experts e
		LEFT JOIN ratings r ON r.expert_id = e.id
		WHERE e.deleted_at IS NULL
		GROUP BY e.id, e.name, e.domain
		ORDER BY avg_score DESC NULLS LAST`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type ratingRow struct {
		ExpertName   string  `json:"expert_name"`
		Domain       string  `json:"domain"`
		TotalRatings int     `json:"total_ratings"`
		AvgScore     float64 `json:"avg_score"`
		GoodRatings  int     `json:"good_ratings"`
		BadRatings   int     `json:"bad_ratings"`
	}
	var ratings []ratingRow
	for rows.Next() {
		var r ratingRow
		if err := rows.Scan(&r.ExpertName, &r.Domain, &r.TotalRatings,
			&r.AvgScore, &r.GoodRatings, &r.BadRatings); err != nil {
			continue
		}
		ratings = append(ratings, r)
	}
	if ratings == nil {
		ratings = []ratingRow{}
	}
	response.OK(c, ratings)
}

// ============================================================
// SETTINGS
// ============================================================

// GetSettings GET /admin/settings
func (h *AdminHandler) GetSettings(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(),
		`SELECT key, value, COALESCE(description,''), updated_at FROM system_settings ORDER BY key`)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()

	type settingRow struct {
		Key         string      `json:"key"`
		Value       interface{} `json:"value"`
		Description string      `json:"description"`
		UpdatedAt   time.Time   `json:"updated_at"`
	}
	var settings []settingRow
	for rows.Next() {
		var s settingRow
		var valueJSON []byte
		if err := rows.Scan(&s.Key, &valueJSON, &s.Description, &s.UpdatedAt); err != nil {
			continue
		}
		var v interface{}
		if err := json.Unmarshal(valueJSON, &v); err == nil {
			s.Value = v
		}
		settings = append(settings, s)
	}
	if settings == nil {
		settings = []settingRow{}
	}
	response.OK(c, settings)
}

// UpdateSetting PATCH /admin/settings/:key
func (h *AdminHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value interface{} `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		response.BadRequest(c, "INVALID_VALUE", "value must be valid JSON")
		return
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	_, err = h.db.Exec(c.Request.Context(),
		`INSERT INTO system_settings (key, value, updated_by)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE SET
			value=EXCLUDED.value,
			updated_by=EXCLUDED.updated_by,
			updated_at=NOW()`,
		key, string(valueJSON), adminID,
	)
	if err != nil {
		h.logger.Error("update setting failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "updated", "key": key})
}

// isUniqueViolation checks if error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unique")
}
