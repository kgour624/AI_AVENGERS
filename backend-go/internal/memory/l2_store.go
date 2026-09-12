package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/ml"
)

// L2Entry is a single memory record in the group memory table.
type L2Entry struct {
	ID           uuid.UUID `json:"id"`
	ProjectID    uuid.UUID `json:"project_id"`
	ExpertID     uuid.UUID `json:"expert_id"`
	MemoryType   string    `json:"memory_type"`
	Content      string    `json:"content"`
	Context      string    `json:"context"`
	TurnReference int      `json:"turn_reference"`
	Importance   int       `json:"importance"`
	CreatedAt    time.Time `json:"created_at"`
}

// L2Store handles PostgreSQL-backed group memory operations.
// Scope: All experts in a project.
// WHY PostgreSQL for L2:
// Needs semantic search (pgvector) + relational queries.
// "What did Expert A decide at turn 15?" = structured query.
// "What decisions are related to sharding?" = vector search.
type L2Store struct {
	db       *pgxpool.Pool
	embedder ml.Embedder // ml.Embedder interface: sidecar or CodeCraftAPI, resolved at call time
	logger   *zap.Logger
}

// NewL2Store creates a new L2 store.
// embedder satisfies ml.Embedder — either *ml.SidecarClient (default) or
// *ml.DynamicEmbedder (when CodeCraftAPI embeddings are enabled).
func NewL2Store(db *pgxpool.Pool, embedder ml.Embedder, logger *zap.Logger) *L2Store {
	return &L2Store{db: db, embedder: embedder, logger: logger}
}

// Append adds a new memory entry to L2.
// Generates embedding for semantic search.
func (s *L2Store) Append(ctx context.Context, entry L2Entry) error {
	// Generate embedding for semantic search
	embedding, err := s.embedder.EmbedSingle(ctx, entry.Content)
	if err != nil {
		s.logger.Warn("embedding failed for L2 entry, storing without vector", zap.Error(err))
		// Store without embedding — won't be searchable but won't fail
		_, err = s.db.Exec(ctx,
			`INSERT INTO project_memory_l2
				(project_id, expert_id, memory_type, content, context, turn_reference, importance)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			entry.ProjectID, entry.ExpertID, entry.MemoryType,
			entry.Content, entry.Context, entry.TurnReference, entry.Importance,
		)
		return err
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO project_memory_l2
			(project_id, expert_id, memory_type, content, context, turn_reference, embedding, importance)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		entry.ProjectID, entry.ExpertID, entry.MemoryType,
		entry.Content, entry.Context, entry.TurnReference,
		pgvector.NewVector(embedding), entry.Importance,
	)
	return err
}

// SearchByQuery finds relevant L2 entries using HYBRID search.
// Combines semantic search + keyword fallback.
//
// WHY hybrid for L2 (Byte by Byte AI course):
// Course taught: keyword-based indexing catches exact matches.
// "What did expert say about PostgreSQL?" — keyword match is better.
// "What decisions were made about scaling?" — semantic is better.
// Hybrid covers both cases.
func (s *L2Store) SearchByQuery(ctx context.Context, projectID uuid.UUID, query string, limit int) ([]L2Entry, error) {
	embedding, err := s.ml.EmbedSingle(ctx, query)
	if err != nil {
		// Embedding failed — fall back to keyword search
		return s.searchByKeyword(ctx, projectID, query, limit)
	}

	// Semantic search
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, memory_type, content,
		        COALESCE(context,''), COALESCE(turn_reference,0), importance, created_at
		 FROM project_memory_l2
		 WHERE project_id=$1 AND is_superseded=FALSE AND embedding IS NOT NULL
		 ORDER BY embedding <=> $2
		 LIMIT $3`,
		projectID, pgvector.NewVector(embedding), limit,
	)
	if err != nil {
		return s.searchByKeyword(ctx, projectID, query, limit)
	}
	defer rows.Close()

	entries, err := scanL2Rows(rows)
	if err != nil {
		return nil, err
	}

	// If semantic search returned few results, supplement with keyword search
	if len(entries) < limit/2 {
		kwEntries, kwErr := s.searchByKeyword(ctx, projectID, query, limit-len(entries))
		if kwErr == nil {
			// Deduplicate by ID
			seen := make(map[uuid.UUID]bool)
			for _, e := range entries {
				seen[e.ID] = true
			}
			for _, e := range kwEntries {
				if !seen[e.ID] {
					entries = append(entries, e)
				}
			}
		}
	}

	return entries, nil
}

// searchByKeyword searches L2 using PostgreSQL full-text search.
// Used as fallback when semantic search returns few results.
func (s *L2Store) searchByKeyword(ctx context.Context, projectID uuid.UUID, query string, limit int) ([]L2Entry, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, memory_type, content,
		        COALESCE(context,''), COALESCE(turn_reference,0), importance, created_at
		 FROM project_memory_l2
		 WHERE project_id=$1
		   AND is_superseded=FALSE
		   AND content_tsv @@ plainto_tsquery('english', $2)
		 ORDER BY importance DESC, created_at DESC
		 LIMIT $3`,
		projectID, query, limit,
	)
	if err != nil {
		return s.GetRecent(ctx, projectID, limit)
	}
	defer rows.Close()
	return scanL2Rows(rows)
}

// GetByExpert returns all L2 entries for a specific expert in a project.
func (s *L2Store) GetByExpert(ctx context.Context, projectID, expertID uuid.UUID) ([]L2Entry, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, memory_type, content,
		        COALESCE(context,''), COALESCE(turn_reference,0), importance, created_at
		 FROM project_memory_l2
		 WHERE project_id=$1 AND expert_id=$2 AND is_superseded=FALSE
		 ORDER BY importance DESC, created_at DESC
		 LIMIT 20`,
		projectID, expertID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanL2Rows(rows)
}

// GetRecent returns most recent L2 entries for a project.
func (s *L2Store) GetRecent(ctx context.Context, projectID uuid.UUID, limit int) ([]L2Entry, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, memory_type, content,
		        COALESCE(context,''), COALESCE(turn_reference,0), importance, created_at
		 FROM project_memory_l2
		 WHERE project_id=$1 AND is_superseded=FALSE
		 ORDER BY importance DESC, created_at DESC
		 LIMIT $2`,
		projectID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanL2Rows(rows)
}

// Supersede marks an entry as superseded (decision changed).
func (s *L2Store) Supersede(ctx context.Context, entryID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE project_memory_l2 SET is_superseded=TRUE, superseded_at=NOW() WHERE id=$1`,
		entryID,
	)
	return err
}

// scanL2Rows scans database rows into L2Entry slice.
func scanL2Rows(rows interface{ Next() bool; Scan(...interface{}) error; Err() error }) ([]L2Entry, error) {
	var entries []L2Entry
	for rows.Next() {
		var e L2Entry
		if err := rows.Scan(
			&e.ID, &e.ProjectID, &e.ExpertID, &e.MemoryType,
			&e.Content, &e.Context, &e.TurnReference, &e.Importance, &e.CreatedAt,
		); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
