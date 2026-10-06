package vacuum

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
)

type Handler struct {
	db      *pgxpool.Pool
	service *Service
	logger  *zap.Logger
}

func NewHandler(db *pgxpool.Pool, svc *Service, logger *zap.Logger) *Handler {
	return &Handler{db: db, service: svc, logger: logger}
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/vacuum/patterns", h.ListPatterns)
	r.POST("/vacuum/patterns", h.CreatePattern)
	r.PATCH("/vacuum/patterns/:id", h.UpdatePattern)
	r.DELETE("/vacuum/patterns/:id", h.DeletePattern)
	r.GET("/vacuum/candidates", h.ListCandidates)
	r.POST("/vacuum/candidates/:id/approve", h.ApproveCandidate)
	r.POST("/vacuum/candidates/:id/reject", h.RejectCandidate)
	r.POST("/vacuum/candidates/bulk-approve", h.BulkApproveCandidates)
	r.POST("/vacuum/candidates/bulk-reject", h.BulkRejectCandidates)
	r.GET("/vacuum/jobs", h.ListJobs)
	r.POST("/vacuum/jobs", h.CreateJob)
	r.GET("/vacuum/jobs/:id", h.GetJob)
	r.POST("/vacuum/jobs/:id/retry", h.RetryJob)
	r.POST("/vacuum/jobs/:id/execute", h.ExecuteJob)
	r.GET("/vacuum/jobs/:id/download", h.DownloadCleaned)
	r.POST("/vacuum/jobs/pick", h.PickJob)
	r.POST("/vacuum/jobs/pick-execute", h.PickJob)
	r.POST("/vacuum/upload", h.UploadAndEnqueue)
	r.GET("/vacuum/jobs/:id/chunks", h.ListChunks)
	r.GET("/vacuum/stats", h.JobsStats)
	r.POST("/vacuum/preview", h.Preview)
	r.GET("/vacuum/brain/version", h.BrainVersion)
	r.GET("/vacuum/metrics", h.VacuumMetrics)
}

func (h *Handler) ListPatterns(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `SELECT id::text, pattern, pattern_type, category, is_active, hit_count, version FROM kachra_patterns ORDER BY updated_at DESC LIMIT 200`)
	if err != nil {
		h.logger.Error("list patterns failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	defer rows.Close()
	var out []KachraPattern
	for rows.Next() {
		var p KachraPattern
		if err := rows.Scan(&p.ID, &p.Pattern, &p.PatternType, &p.Category, &p.IsActive, &p.HitCount, &p.Version); err != nil {
			continue
		}
		out = append(out, p)
	}
	if out == nil {
		out = []KachraPattern{}
	}
	response.OK(c, out)
}

func (h *Handler) CreatePattern(c *gin.Context) {
	var req struct {
		Pattern     string `json:"pattern" binding:"required"`
		PatternType string `json:"pattern_type"`
		Category    string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.PatternType == "" {
		req.PatternType = "PHRASE"
	}
	if req.Category == "" {
		req.Category = "filler"
	}
	req.Pattern = strings.TrimSpace(req.Pattern)
	var id string
	err := h.db.QueryRow(c.Request.Context(), `INSERT INTO kachra_patterns(pattern, pattern_type, category) VALUES ($1,$2,$3) RETURNING id::text`, req.Pattern, req.PatternType, req.Category).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(c, "pattern already exists")
			return
		}
		h.logger.Error("create pattern failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	go func() { _ = h.service.Brain().Reload(context.Background()) }()
	response.Created(c, gin.H{"id": id})
}

func (h *Handler) UpdatePattern(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	var req struct {
		IsActive *bool   `json:"is_active"`
		Category *string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if req.IsActive != nil {
		_, _ = h.db.Exec(c.Request.Context(), `UPDATE kachra_patterns SET is_active=$1 WHERE id=$2`, *req.IsActive, id)
	}
	if req.Category != nil {
		_, _ = h.db.Exec(c.Request.Context(), `UPDATE kachra_patterns SET category=$1 WHERE id=$2`, *req.Category, id)
	}
	go func() { _ = h.service.Brain().Reload(context.Background()) }()
	response.OK(c, gin.H{"status": "updated"})
}

func (h *Handler) DeletePattern(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid id")
		return
	}
	_, _ = h.db.Exec(c.Request.Context(), `DELETE FROM kachra_patterns WHERE id=$1`, id)
	go func() { _ = h.service.Brain().Reload(context.Background()) }()
	response.OK(c, gin.H{"status": "deleted"})
}

func (h *Handler) ListCandidates(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		status = "pending"
	}
	rows, err := h.db.Query(c.Request.Context(), `SELECT id::text, pattern, category, status, hit_count, confidence, context_snippet FROM candidate_kachra WHERE status=$1 ORDER BY hit_count DESC LIMIT 200`, status)
	if err != nil {
		response.InternalError(c)
		return
	}
	defer rows.Close()
	var out []CandidateRow
	for rows.Next() {
		var r CandidateRow
		if err := rows.Scan(&r.ID, &r.Pattern, &r.Category, &r.Status, &r.HitCount, &r.Confidence, &r.Context); err != nil {
			continue
		}
		out = append(out, r)
	}
	if out == nil {
		out = []CandidateRow{}
	}
	response.OK(c, out)
}

