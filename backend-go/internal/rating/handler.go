package rating

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/memory"
	"ai_avengers/backend/internal/response"
)

// RateRequest holds input for rating a message.
type RateRequest struct {
	Score        int    `json:"score" binding:"required,min=1,max=5"`
	Feedback     string `json:"feedback"`
	FeedbackType string `json:"feedback_type"` // accepted/rejected/modified/ignored
	CodeExecuted bool   `json:"code_executed"`
	ExecSuccess  *bool  `json:"execution_success,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Service handles rating business logic.
type Service struct {
	db         *pgxpool.Pool
	memManager *memory.Manager
	logger     *zap.Logger
}

// NewService creates a new rating service.
func NewService(db *pgxpool.Pool, memManager *memory.Manager, logger *zap.Logger) *Service {
	return &Service{db: db, memManager: memManager, logger: logger}
}

// Rate records a rating for a message.
// Also updates:
// 1. Expert avg_rating in experts table
// 2. Chunk boost_factor for cited chunks
// 3. L3 master event log
func (s *Service) Rate(ctx context.Context, messageID, clientID uuid.UUID, req RateRequest) error {
	// Get message details (expert_id, project_id, chat_id)
	var expertID uuid.UUID
	var projectID uuid.UUID
	var chatID uuid.UUID
	var decisionMode string

	err := s.db.QueryRow(ctx,
		`SELECT m.expert_id, c.project_id, m.chat_id, COALESCE(m.decision_mode,'')
		 FROM messages m
		 JOIN chats c ON c.id = m.chat_id
		 WHERE m.id=$1 AND c.client_id=$2`,
		messageID, clientID,
	).Scan(&expertID, &projectID, &chatID, &decisionMode)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	// Save rating
	_, err = s.db.Exec(ctx,
		`INSERT INTO ratings
			(message_id, client_id, expert_id, project_id, score, feedback,
			 feedback_type, code_executed, execution_success, error_message)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		messageID, clientID, expertID, projectID,
		req.Score, req.Feedback, req.FeedbackType,
		req.CodeExecuted, req.ExecSuccess, req.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("save rating failed: %w", err)
	}

	// Update expert avg_rating async
	go s.updateExpertRating(context.Background(), expertID)

	// Update chunk boost factors async
	go s.updateChunkBoosts(context.Background(), messageID, req.Score)

	// Log to L3 async
	go s.logToL3(context.Background(), projectID, expertID, clientID, chatID, messageID, req)

	return nil
}

// updateExpertRating recalculates expert's average rating.
func (s *Service) updateExpertRating(ctx context.Context, expertID uuid.UUID) {
	_, err := s.db.Exec(ctx,
		`UPDATE experts SET
			avg_rating = (
				SELECT ROUND(AVG(score)::numeric, 2)
				FROM ratings WHERE expert_id=$1
			),
			total_ratings = (
				SELECT COUNT(*) FROM ratings WHERE expert_id=$1
			),
			updated_at = NOW()
		 WHERE id=$1`,
		expertID,
	)
	if err != nil {
		s.logger.Warn("update expert rating failed", zap.Error(err))
	}
}

// updateChunkBoosts adjusts boost_factor for cited chunks based on rating.
// WHY chunk boost (Byte by Byte AI + Apna Kiro advanced features):
// Chunks that lead to good responses should be retrieved more often.
// Chunks that lead to bad responses should be penalized.
// This is the self-learning loop.
func (s *Service) updateChunkBoosts(ctx context.Context, messageID uuid.UUID, score int) {
	// Get cited chunk IDs from message
	var citationsJSON []byte
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(citations, '[]'::jsonb) FROM messages WHERE id=$1`,
		messageID,
	).Scan(&citationsJSON)
	if err != nil || len(citationsJSON) == 0 {
		return
	}

	// Determine boost delta
	var boostDelta float64
	switch {
	case score >= 4:
		boostDelta = 0.05 // Boost good chunks
	case score <= 2:
		boostDelta = -0.05 // Penalize bad chunks
	default:
		return // Neutral rating, no change
	}

	// Update boost_factor for all cited chunks
	// Clamp between 0.5 and 1.5
	_, err = s.db.Exec(ctx,
		`UPDATE course_chunks SET
			boost_factor = GREATEST(0.5, LEAST(1.5, boost_factor + $1)),
			times_cited = times_cited + 1
		 WHERE id IN (
			 SELECT (elem->>'chunk_id')::uuid
			 FROM jsonb_array_elements($2::jsonb) AS elem
		 )`,
		boostDelta, string(citationsJSON),
	)
	if err != nil {
		s.logger.Warn("update chunk boosts failed", zap.Error(err))
	}
}

// logToL3 records the rating event in the master event log.
// WHY RecordRating not RecordViolation:
// A rating is client feedback, not a China Wall violation.
// Using RecordViolation was polluting the admin violations list.
func (s *Service) logToL3(
	ctx context.Context,
	projectID, expertID, clientID, chatID, messageID uuid.UUID,
	req RateRequest,
) {
	s.memManager.RecordRating(
		ctx, projectID, expertID, clientID, chatID, messageID,
		req.Score, req.FeedbackType,
	)
}

// Handler handles HTTP requests for ratings.
type Handler struct {
	svc    *Service
	logger *zap.Logger
}

// NewHandler creates a new rating handler.
func NewHandler(svc *Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Rate POST /messages/:id/rate
func (h *Handler) Rate(c *gin.Context) {
	clientID := c.MustGet("user_id").(uuid.UUID)
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "invalid message ID")
		return
	}
	var req RateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_INPUT", err.Error())
		return
	}
	if err := h.svc.Rate(c.Request.Context(), messageID, clientID, req); err != nil {
		h.logger.Error("rate message failed", zap.Error(err))
		response.InternalError(c)
		return
	}
	response.OK(c, map[string]string{"status": "rated"})
}
