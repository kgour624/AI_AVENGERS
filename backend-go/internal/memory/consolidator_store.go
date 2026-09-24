package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store-side helpers for B7 consolidation / decay. Kept separate from
// consolidator.go so L2Store owns all SQL (same pattern as Supersede).

// listActiveForConsolidate returns up to limit active non-summary rows,
// highest importance first then oldest (so dense old decisions get covered).
func (s *L2Store) listActiveForConsolidate(ctx context.Context, projectID uuid.UUID, limit int) ([]L2Entry, error) {
	if limit <= 0 {
		limit = consolidateMaxEntries
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, memory_type, content,
		        COALESCE(context,''), COALESCE(turn_reference,0), importance, created_at,
		        COALESCE(weight, 1.0)
		 FROM project_memory_l2
		 WHERE project_id=$1
		   AND is_superseded=FALSE
		   AND memory_type <> 'summary'
		 ORDER BY importance DESC, created_at ASC
		 LIMIT $2`,
		projectID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanL2RowsWeighted(rows)
}

// hasRecentSummary is true when an active summary was written inside cooldown.
func (s *L2Store) hasRecentSummary(ctx context.Context, projectID uuid.UUID, cooldown time.Duration) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM project_memory_l2
			 WHERE project_id=$1
			   AND memory_type='summary'
			   AND is_superseded=FALSE
			   AND created_at > NOW() - ($2 * INTERVAL '1 second')
		)`,
		projectID, int(cooldown.Seconds()),
	).Scan(&exists)
	return exists, err
}

// listConsolidateCandidates: active projects with enough raw L2 rows and
// no fresh summary. Cheap aggregate — no per-row pull.
func (s *L2Store) listConsolidateCandidates(ctx context.Context, minEntries int, cooldown time.Duration) ([]uuid.UUID, error) {
	rows, err := s.db.Query(ctx,
		`SELECT l2.project_id
		 FROM project_memory_l2 l2
		 JOIN projects p ON p.id = l2.project_id AND p.deleted_at IS NULL
		 WHERE l2.is_superseded=FALSE
		   AND l2.memory_type <> 'summary'
		 GROUP BY l2.project_id
		 HAVING COUNT(*) >= $1
		    AND NOT EXISTS (
		          SELECT 1 FROM project_memory_l2 s
		           WHERE s.project_id = l2.project_id
		             AND s.memory_type = 'summary'
		             AND s.is_superseded = FALSE
		             AND s.created_at > NOW() - ($2 * INTERVAL '1 second')
		        )
		 LIMIT 50`,
		minEntries, int(cooldown.Seconds()),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// supersedeOlderSummaries marks every active summary for the project as
// superseded EXCEPT the newest one (growing-doc keeps a single live summary).
func (s *L2Store) supersedeOlderSummaries(ctx context.Context, projectID uuid.UUID) (int, error) {
	tag, err := s.db.Exec(ctx,
		`UPDATE project_memory_l2
		 SET is_superseded=TRUE, superseded_at=NOW()
		 WHERE project_id=$1
		   AND memory_type='summary'
		   AND is_superseded=FALSE
		   AND id <> (
		         SELECT id FROM project_memory_l2
		          WHERE project_id=$1 AND memory_type='summary' AND is_superseded=FALSE
		          ORDER BY created_at DESC
		          LIMIT 1
		       )`,
		projectID,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// supersedeMany marks the given ids superseded. Empty slice is a no-op.
func (s *L2Store) supersedeMany(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tag, err := s.db.Exec(ctx,
		`UPDATE project_memory_l2
		 SET is_superseded=TRUE, superseded_at=NOW()
		 WHERE id = ANY($1) AND is_superseded=FALSE`,
		ids,
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// decayPreferences: one SQL pass applies half-life math and supersedes
// prefs that fall below floor. importance >= criticalImportance is exempt.
// Facts are never selected (memory_type = 'preference' only).
func (s *L2Store) decayPreferences(
	ctx context.Context,
	halfLifeDays float64,
	floor float64,
	criticalImportance int,
) error {
	// weight' = weight * 0.5^(idle_days / half_life)
	// PostgreSQL: power(0.5, EXTRACT(EPOCH FROM (NOW()-last_accessed_at))/86400 / half_life)
	_, err := s.db.Exec(ctx,
		`UPDATE project_memory_l2
		 SET weight = GREATEST(
		       0,
		       LEAST(
		         1,
		         weight * power(
		           0.5,
		           (EXTRACT(EPOCH FROM (NOW() - last_accessed_at)) / 86400.0) / $1
		         )
		       )
		     )
		 WHERE is_superseded=FALSE
		   AND memory_type='preference'
		   AND importance < $2
		   AND last_accessed_at < NOW() - INTERVAL '1 day'`,
		halfLifeDays, criticalImportance,
	)
	if err != nil {
		return fmt.Errorf("decay update: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE project_memory_l2
		 SET is_superseded=TRUE, superseded_at=NOW()
		 WHERE is_superseded=FALSE
		   AND memory_type='preference'
		   AND importance < $1
		   AND weight < $2`,
		criticalImportance, floor,
	)
	if err != nil {
		return fmt.Errorf("decay supersede: %w", err)
	}
	return nil
}

// touchAccess bumps last_accessed_at for retrieved rows (hot prefs resist decay).
func (s *L2Store) touchAccess(ctx context.Context, ids []uuid.UUID) {
	if len(ids) == 0 {
		return
	}
	_, _ = s.db.Exec(ctx,
		`UPDATE project_memory_l2 SET last_accessed_at=NOW() WHERE id = ANY($1)`,
		ids,
	)
}

// projectClientID looks up projects.client_id for L3 append (NOT NULL there).
func (s *L2Store) projectClientID(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT client_id FROM projects WHERE id=$1 AND deleted_at IS NULL`,
		projectID,
	).Scan(&id)
	return id, err
}

// scanL2RowsWeighted includes weight (post-026). Older scanL2Rows stays
// for pre-migration-compatible call sites that don't need weight.
func scanL2RowsWeighted(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}) ([]L2Entry, error) {
	var entries []L2Entry
	for rows.Next() {
		var e L2Entry
		if err := rows.Scan(
			&e.ID, &e.ProjectID, &e.ExpertID, &e.MemoryType,
			&e.Content, &e.Context, &e.TurnReference, &e.Importance, &e.CreatedAt,
			&e.Weight,
		); err != nil {
			continue
		}
		if e.Weight <= 0 {
			e.Weight = 1.0
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
