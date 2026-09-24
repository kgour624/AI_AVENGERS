// Package knowledge is the Expert Knowledge bounded context (read side).
//
// Stage 0: Postgres adapter for ports.KnowledgeReader only.
// Catalog writes stay in admin/training; this package is the seam conversation
// and workflow will migrate onto so experts can scale/extract independently.
package knowledge

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai_avengers/backend/internal/ports"
)

// Reader implements ports.KnowledgeReader against the experts table.
type Reader struct {
	db *pgxpool.Pool
}

var _ ports.KnowledgeReader = (*Reader)(nil)

// NewReader creates a Postgres-backed knowledge reader.
func NewReader(db *pgxpool.Pool) *Reader {
	return &Reader{db: db}
}

// GetExperts returns summaries for the given IDs (any status). Missing IDs are omitted.
func (r *Reader) GetExperts(ctx context.Context, ids []uuid.UUID) ([]ports.ExpertSummary, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Build IN ($1,$2,…) — avoids uuid[] literal casting bugs (see blackboard.UUIDArrayLiteral).
	args := make([]interface{}, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	q := fmt.Sprintf(`
		SELECT id, name, slug, domain,
		       (is_active = TRUE AND is_training = FALSE AND training_status = 'trained' AND deleted_at IS NULL)
		  FROM experts
		 WHERE id IN (%s)`, strings.Join(placeholders, ","))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("get experts: %w", err)
	}
	defer rows.Close()

	var out []ports.ExpertSummary
	for rows.Next() {
		var s ports.ExpertSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Domain, &s.Active); err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// ListTrainedActive returns the public client catalog (trained + active).
func (r *Reader) ListTrainedActive(ctx context.Context) ([]ports.ExpertSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug, domain, TRUE
		  FROM experts
		 WHERE is_active = TRUE
		   AND is_training = FALSE
		   AND training_status = 'trained'
		   AND deleted_at IS NULL
		 ORDER BY avg_rating DESC NULLS LAST, total_chunks DESC`)
	if err != nil {
		return nil, fmt.Errorf("list trained experts: %w", err)
	}
	defer rows.Close()

	var out []ports.ExpertSummary
	for rows.Next() {
		var s ports.ExpertSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Domain, &s.Active); err != nil {
			continue
		}
		out = append(out, s)
	}
	if out == nil {
		out = []ports.ExpertSummary{}
	}
	return out, nil
}
