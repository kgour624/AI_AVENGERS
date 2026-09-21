package workflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DesignSection is one expert's assigned slot in a workflow's design folder.
// Once assigned it never changes — see COLLABORATIVE_DESIGN_ARCHITECTURE.md §3.2.
type DesignSection struct {
	WorkflowID  uuid.UUID
	ExpertID    uuid.UUID
	SectionNo   int
	SectionPath string
}

// ErrExpertNotFound is returned when a section is requested for an expert id
// that does not exist in the experts table. Distinguished from an assignment
// race so a caller can tell a real data error apart from a transient retry.
var ErrExpertNotFound = errors.New("expert not found")

// designSectionMaxRetries bounds the assign loop. Each retry happens only when a
// concurrent authoring turn grabbed the section number this turn computed; the
// number strictly advances each round, so a handful of retries covers far more
// experts than any real wave contains.
const designSectionMaxRetries = 8

// DesignSectionStore assigns and reads per-expert design section slots.
//
// It follows blackboard.Store's shape: a pgx pool plus a logger, INSERT ...
// ON CONFLICT DO NOTHING RETURNING, and a re-query when RETURNING is empty.
// No advisory lock — nothing in this codebase uses one, and the ON-CONFLICT +
// re-query pattern is already how blackboard.Post handles concurrent writers.
type DesignSectionStore struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewDesignSectionStore creates a section store.
func NewDesignSectionStore(db *pgxpool.Pool, logger *zap.Logger) *DesignSectionStore {
	return &DesignSectionStore{db: db, logger: logger}
}

// AssignSection returns the expert's section in this workflow, assigning one on
// first call and returning the same one on every call after. Idempotent.
//
// The number is MAX(section_no)+10, not (count+1)*10 as the doc's formula reads.
// They agree while rows are only ever added (the normal case — assignment is
// once-per-expert and rows are deleted only when the whole workflow is). They
// differ after a hypothetical deletion: (count+1)*10 could reproduce a number a
// surviving row already holds and hit the UNIQUE(workflow_id, section_no)
// constraint, whereas MAX+10 is always free. MAX+10 satisfies the doc's intent
// (monotonic, unique, ×10 spacing) and is strictly safer, so it is used here.
//
// Concurrency: two experts in one wave author in parallel and can compute the
// same next number. The INSERT then hits UNIQUE(workflow_id, section_no); the
// loser re-reads MAX (now advanced) and retries. A second call for an
// already-assigned expert hits the (workflow_id, expert_id) primary key and the
// re-query returns that existing row — which is the idempotent path.
func (s *DesignSectionStore) AssignSection(ctx context.Context, workflowID, expertID uuid.UUID) (DesignSection, error) {
	for attempt := 0; attempt < designSectionMaxRetries; attempt++ {
		// One statement computes the next number and the path atomically.
		// The path is built here (not in Go) so it is derived from the same
		// MAX read that fixes the number — no second round trip, no window for
		// the number and the path to disagree. next_no is cast to text
		// explicitly rather than relying on || 's implicit cast.
		// CROSS JOIN experts yields no row when the expert id is unknown, so
		// the INSERT inserts nothing and the not-found case is handled below.
		var sec DesignSection
		err := s.db.QueryRow(ctx,
			`INSERT INTO workflow_design_sections (workflow_id, expert_id, section_no, section_path)
			 SELECT $1, $2, c.next_no,
			        'design/' || c.next_no::text || '-' || e.slug || '.md'
			 FROM (
			     SELECT COALESCE(MAX(section_no), 0) + 10 AS next_no
			     FROM workflow_design_sections
			     WHERE workflow_id = $1
			 ) c
			 CROSS JOIN (SELECT slug FROM experts WHERE id = $2) e
			 ON CONFLICT DO NOTHING
			 RETURNING workflow_id, expert_id, section_no, section_path`,
			workflowID, expertID,
		).Scan(&sec.WorkflowID, &sec.ExpertID, &sec.SectionNo, &sec.SectionPath)
		if err == nil {
			s.logger.Info("design section assigned",
				zap.String("workflow_id", workflowID.String()),
				zap.String("expert_id", expertID.String()),
				zap.Int("section_no", sec.SectionNo),
				zap.String("section_path", sec.SectionPath),
			)
			return sec, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return DesignSection{}, fmt.Errorf("assign section: insert: %w", err)
		}

		// RETURNING was empty. Three possible reasons:
		//   a) this expert already has a section  -> re-query returns it (done)
		//   b) another expert took our number      -> re-query empty, retry
		//   c) the expert id does not exist         -> re-query empty, no rows
		//                                              in experts -> ErrExpertNotFound
		existing, found, qErr := s.GetSection(ctx, workflowID, expertID)
		if qErr != nil {
			return DesignSection{}, fmt.Errorf("assign section: re-query: %w", qErr)
		}
		if found {
			return existing, nil
		}

		// Not assigned. Is that because the expert does not exist (c), or a
		// number collision (b)? Check once; a genuine missing expert must not
		// be retried into a confusing "max retries" error.
		var exists bool
		if err := s.db.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM experts WHERE id = $1)`, expertID,
		).Scan(&exists); err != nil {
			return DesignSection{}, fmt.Errorf("assign section: expert check: %w", err)
		}
		if !exists {
			return DesignSection{}, fmt.Errorf("assign section: %w: %s", ErrExpertNotFound, expertID)
		}
		// Collision — loop and try the next number.
		s.logger.Debug("design section number collision, retrying",
			zap.String("workflow_id", workflowID.String()),
			zap.String("expert_id", expertID.String()),
			zap.Int("attempt", attempt+1),
		)
	}
	return DesignSection{}, fmt.Errorf("assign section: gave up after %d attempts (persistent section_no contention)", designSectionMaxRetries)
}

// GetSection returns the expert's section if one has been assigned.
// found is false with a nil error when no section exists yet.
func (s *DesignSectionStore) GetSection(ctx context.Context, workflowID, expertID uuid.UUID) (DesignSection, bool, error) {
	var sec DesignSection
	err := s.db.QueryRow(ctx,
		`SELECT workflow_id, expert_id, section_no, section_path
		 FROM workflow_design_sections
		 WHERE workflow_id = $1 AND expert_id = $2`,
		workflowID, expertID,
	).Scan(&sec.WorkflowID, &sec.ExpertID, &sec.SectionNo, &sec.SectionPath)
	if errors.Is(err, pgx.ErrNoRows) {
		return DesignSection{}, false, nil
	}
	if err != nil {
		return DesignSection{}, false, fmt.Errorf("get section: %w", err)
	}
	return sec, true, nil
}

// ListSections returns every assigned section for a workflow, ordered by number.
//
// This is what the spine's module map and the generated CLAUDE.md read: the
// list of section files is derived from here, never hand-written, so it always
// reflects the real roster (§3.3, §8).
func (s *DesignSectionStore) ListSections(ctx context.Context, workflowID uuid.UUID) ([]DesignSection, error) {
	rows, err := s.db.Query(ctx,
		`SELECT workflow_id, expert_id, section_no, section_path
		 FROM workflow_design_sections
		 WHERE workflow_id = $1
		 ORDER BY section_no`,
		workflowID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sections: %w", err)
	}
	defer rows.Close()

	var out []DesignSection
	for rows.Next() {
		var sec DesignSection
		if err := rows.Scan(&sec.WorkflowID, &sec.ExpertID, &sec.SectionNo, &sec.SectionPath); err != nil {
			return nil, fmt.Errorf("list sections: scan: %w", err)
		}
		out = append(out, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list sections: rows: %w", err)
	}
	return out, nil
}
