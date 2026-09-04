package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Expert is the admin view of an expert (includes all stats).
type Expert struct {
	ID                   uuid.UUID              `json:"id"`
	Name                 string                 `json:"name"`
	Slug                 string                 `json:"slug"`
	Domain               string                 `json:"domain"`
	Description          string                 `json:"description"`
	ReasoningCharter     string                 `json:"reasoning_charter"`
	ClarificationCharter map[string]interface{} `json:"clarification_charter"`
	CapabilitySummary    map[string]interface{} `json:"capability_summary"`
	TotalChunks          int                    `json:"total_chunks"`
	TotalTopics          int                    `json:"total_topics"`
	AvgDepthLevel        float64                `json:"avg_depth_level"`
	AvgRating            float64                `json:"avg_rating"`
	TotalRatings         int                    `json:"total_ratings"`
	IsActive             bool                   `json:"is_active"`
	IsTraining           bool                   `json:"is_training"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// CreateExpertRequest holds input for creating a new expert.
type CreateExpertRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Domain      string `json:"domain" binding:"required"`
	Description string `json:"description"`
}

// UpdateExpertRequest holds input for updating an expert.
type UpdateExpertRequest struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	IsActive         *bool   `json:"is_active"`
	ReasoningCharter *string `json:"reasoning_charter"`
}

// ExpertService handles expert management for admin.
type ExpertService struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewExpertService creates a new expert service.
func NewExpertService(db *pgxpool.Pool, logger *zap.Logger) *ExpertService {
	return &ExpertService{db: db, logger: logger}
}

// ListAll returns all experts with stats.
func (s *ExpertService) ListAll(ctx context.Context) ([]Expert, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, slug, domain, COALESCE(description,''),
		       COALESCE(reasoning_charter,''), clarification_charter, capability_summary,
		       total_chunks, total_topics, COALESCE(avg_depth_level,0),
		       COALESCE(avg_rating,0), total_ratings,
		       is_active, is_training, created_at, updated_at
		FROM experts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var experts []Expert
	for rows.Next() {
		var e Expert
		var clarJSON, capJSON []byte
		err := rows.Scan(
			&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
			&e.ReasoningCharter, &clarJSON, &capJSON,
			&e.TotalChunks, &e.TotalTopics, &e.AvgDepthLevel,
			&e.AvgRating, &e.TotalRatings,
			&e.IsActive, &e.IsTraining, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		experts = append(experts, e)
	}
	return experts, nil
}

// Create creates a new expert record.
func (s *ExpertService) Create(ctx context.Context, req CreateExpertRequest) (*Expert, error) {
	// Check slug uniqueness
	var exists bool
	_ = s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM experts WHERE slug=$1 AND deleted_at IS NULL)`,
		req.Slug,
	).Scan(&exists)
	if exists {
		return nil, fmt.Errorf("slug already exists: %s", req.Slug)
	}

	var e Expert
	err := s.db.QueryRow(ctx, `
		INSERT INTO experts (name, slug, domain, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, slug, domain, COALESCE(description,''), is_active, is_training, created_at, updated_at`,
		req.Name, req.Slug, req.Domain, req.Description,
	).Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
		&e.IsActive, &e.IsTraining, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	s.logger.Info("expert created", zap.String("id", e.ID.String()), zap.String("slug", e.Slug))
	return &e, nil
}

// Update updates an expert's fields.
func (s *ExpertService) Update(ctx context.Context, id uuid.UUID, req UpdateExpertRequest) error {
	if req.Name != nil {
		_, err := s.db.Exec(ctx, `UPDATE experts SET name=$1, updated_at=NOW() WHERE id=$2`, *req.Name, id)
		if err != nil {
			return err
		}
	}
	if req.IsActive != nil {
		_, err := s.db.Exec(ctx, `UPDATE experts SET is_active=$1, updated_at=NOW() WHERE id=$2`, *req.IsActive, id)
		if err != nil {
			return err
		}
	}
	if req.ReasoningCharter != nil {
		_, err := s.db.Exec(ctx, `UPDATE experts SET reasoning_charter=$1, updated_at=NOW() WHERE id=$2`, *req.ReasoningCharter, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateIngestionJob creates a new ingestion job record.
func (s *ExpertService) CreateIngestionJob(ctx context.Context, expertID uuid.UUID, sourceFile string) (uuid.UUID, error) {
	var jobID uuid.UUID
	err := s.db.QueryRow(ctx,
		`INSERT INTO ingestion_jobs (expert_id, job_type, status, source_path)
		 VALUES ($1, 'transcript', 'pending', $2)
		 RETURNING id`,
		expertID, sourceFile,
	).Scan(&jobID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create job failed: %w", err)
	}
	return jobID, nil
}

// GetIngestionJobs returns all ingestion jobs for an expert.
func (s *ExpertService) GetIngestionJobs(ctx context.Context, expertID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, status, COALESCE(source_path,''), total_chunks, processed_chunks,
		       COALESCE(error_message,''), started_at, completed_at, created_at
		FROM ingestion_jobs
		WHERE expert_id=$1
		ORDER BY created_at DESC
		LIMIT 20`,
		expertID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var status, sourcePath, errMsg string
		var total, processed int
		var startedAt, completedAt *time.Time
		var createdAt time.Time

		if err := rows.Scan(&id, &status, &sourcePath, &total, &processed, &errMsg, &startedAt, &completedAt, &createdAt); err != nil {
			continue
		}
		jobs = append(jobs, map[string]interface{}{
			"id": id, "status": status, "source_path": sourcePath,
			"total_chunks": total, "processed_chunks": processed,
			"error_message": errMsg, "started_at": startedAt,
			"completed_at": completedAt, "created_at": createdAt,
		})
	}
	return jobs, nil
}
