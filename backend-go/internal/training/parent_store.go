package training

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// StoreParents inserts expert_pages rows (additive, ON CONFLICT DO NOTHING).
// Returns map[page_index]page_id for child FK linking. Parent embeddings are NULL initially
// (Phase 2 backfills via embed job); IVFFLAT still works with NULL embeddings skipped.
func StoreParents(ctx context.Context, db *pgxpool.Pool, expertID uuid.UUID, sourceFile string, parents []ParentChunk) (map[int]uuid.UUID, error) {
	if len(parents) == 0 {
		return map[int]uuid.UUID{}, nil
	}
	ids := make(map[int]uuid.UUID, len(parents))
	for _, p := range parents {
		var id uuid.UUID
		err := db.QueryRow(ctx, `
			INSERT INTO expert_pages (expert_id, source_file, page_index, page_text, section_path, token_count)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (expert_id, page_index, source_file) DO UPDATE SET page_text=EXCLUDED.page_text, section_path=EXCLUDED.section_path, token_count=EXCLUDED.token_count
			RETURNING id`, expertID, sourceFile, p.PageIndex, p.Text, p.SectionPath, p.TokenCount).Scan(&id)
		if err != nil {
			return nil, err
		}
		ids[p.PageIndex] = id
	}
	return ids, nil
}

// StoreParentsWithEmbeddings inserts parents with embeddings (when embedder available).
func StoreParentsWithEmbeddings(ctx context.Context, tx pgx.Tx, expertID uuid.UUID, sourceFile string, parents []ParentChunk, embeddings [][]float32) (map[int]uuid.UUID, error) {
	if len(parents) == 0 {
		return map[int]uuid.UUID{}, nil
	}
	ids := make(map[int]uuid.UUID, len(parents))
	for i, p := range parents {
		var vec pgvector.Vector
		if i < len(embeddings) && len(embeddings[i]) > 0 {
			vec = pgvector.NewVector(embeddings[i])
		}
		var id uuid.UUID
		var err error
		if len(vec.Slice()) > 0 {
			err = tx.QueryRow(ctx, `
				INSERT INTO expert_pages (expert_id, source_file, page_index, page_text, section_path, token_count, embedding)
				VALUES ($1,$2,$3,$4,$5,$6,$7)
				ON CONFLICT (expert_id, page_index, source_file) DO UPDATE SET page_text=EXCLUDED.page_text, section_path=EXCLUDED.section_path, token_count=EXCLUDED.token_count, embedding=EXCLUDED.embedding
				RETURNING id`, expertID, sourceFile, p.PageIndex, p.Text, p.SectionPath, p.TokenCount, vec).Scan(&id)
		} else {
			err = tx.QueryRow(ctx, `
				INSERT INTO expert_pages (expert_id, source_file, page_index, page_text, section_path, token_count)
				VALUES ($1,$2,$3,$4,$5,$6)
				ON CONFLICT (expert_id, page_index, source_file) DO UPDATE SET page_text=EXCLUDED.page_text, section_path=EXCLUDED.section_path, token_count=EXCLUDED.token_count
				RETURNING id`, expertID, sourceFile, p.PageIndex, p.Text, p.SectionPath, p.TokenCount).Scan(&id)
		}
		if err != nil {
			return nil, err
		}
		ids[p.PageIndex] = id
	}
	return ids, nil
}

