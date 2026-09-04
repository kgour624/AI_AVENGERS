package admin

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/response"
	"ai_avengers/backend/internal/training"
)

// Handler holds dependencies for admin HTTP handlers.
type Handler struct {
	expertSvc  *ExpertService
	ingestion  *training.IngestionPipeline
	logger     *zap.Logger
}

// NewHandler creates a new admin handler.
func NewHandler(expertSvc *ExpertService, ingestion *training.IngestionPipeline, logger *zap.Logger) *Handler {
	return &Handler{expertSvc: expertSvc, ingestion: ingestion, logger: logger}
}

// ListExperts GET /admin/experts
func (h *Handler) ListExperts(c *gin.Context) {
	experts, err := h.expertSvc.ListAll(c.Request.Context())
	if err != nil {
		h.logger.Error("list experts failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, experts)
}

// CreateExpert POST /admin/experts
func (h *Handler) CreateExpert(c *gin.Context) {
	var req CreateExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	expert, err := h.expertSvc.Create(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "slug already exists") {
			response.Conflict(c, err.Error())
			return
		}
		h.logger.Error("create expert failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.Created(c, expert)
}

// UpdateExpert PATCH /admin/experts/:id
func (h *Handler) UpdateExpert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	var req UpdateExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}

	if err := h.expertSvc.Update(c.Request.Context(), id, req); err != nil {
		h.logger.Error("update expert failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "updated"})
}

// IngestTranscript POST /admin/experts/:id/ingest
// Accepts multipart form with "transcript" file field.
// Starts background ingestion job.
//
// Mental execution:
// 1. Parse expert ID from URL
// 2. Read uploaded file (max 50MB)
// 3. Validate file type (.txt or .md only)
// 4. Create ingestion job record
// 5. Start background goroutine for ingestion
// 6. Return job ID immediately (async)
func (h *Handler) IngestTranscript(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	// Parse multipart file
	file, header, err := c.Request.FormFile("transcript")
	if err != nil {
		response.BadRequest(c, "FILE_REQUIRED", "transcript file is required")
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".txt" && ext != ".md" {
		response.BadRequest(c, "INVALID_FILE_TYPE", "only .txt and .md files supported")
		return
	}

	// Validate file size (50MB max)
	if header.Size > 50*1024*1024 {
		response.BadRequest(c, "FILE_TOO_LARGE", "file must be under 50MB")
		return
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("read file failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	if len(strings.TrimSpace(string(content))) == 0 {
		response.BadRequest(c, "EMPTY_FILE", "transcript file is empty")
		return
	}

	// Get expert name for charter extraction
	expertName := c.PostForm("expert_name")
	if expertName == "" {
		expertName = header.Filename
	}

	// Create ingestion job
	jobID, err := h.expertSvc.CreateIngestionJob(c.Request.Context(), expertID, header.Filename)
	if err != nil {
		h.logger.Error("create job failed", zap.Error(err))
		response.InternalError(c)
		return
	}

	// Mark expert as training
	isTraining := true
	_ = h.expertSvc.Update(c.Request.Context(), expertID, UpdateExpertRequest{IsActive: nil})
	_, _ = h.expertSvc.db.Exec(c.Request.Context(),
		`UPDATE experts SET is_training=$1 WHERE id=$2`, isTraining, expertID)

	// Start background ingestion
	// WHY goroutine: Ingestion takes minutes. Client gets job ID immediately.
	// Client polls GET /admin/experts/:id/jobs to check progress.
	go func() {
		ctx := c.Request.Context()
		// Use background context so it survives request completion
		ctx = context.Background()
		_, err := h.ingestion.IngestTranscript(
			ctx, jobID, expertID, expertName,
			string(content), header.Filename,
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
func (h *Handler) GetIngestionJobs(c *gin.Context) {
	expertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid expert ID")
		return
	}

	jobs, err := h.expertSvc.GetIngestionJobs(c.Request.Context(), expertID)
	if err != nil {
		h.logger.Error("get jobs failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, jobs)
}

// context import needed for background goroutine
var context = struct {
	Background func() interface{}
}{}
