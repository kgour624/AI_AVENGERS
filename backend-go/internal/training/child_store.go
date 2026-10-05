package training

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// StoreChildrenDual inserts children with parent_id FK when parents exist.
// Additive: if parentIDs empty, falls back to old columns only (backward compat).
// Uses ON CONFLICT (expert_id, chunk_hash) DO NOTHING to keep dedup.
func StoreChildrenDual(ctx context.Context, db *pgxpool.Pool, expertID uuid.UUID, chunks []TextChunk, embeddings [][]float32, topicResults []TopicResult, sourceFile string, parentIDs map[int]uuid.UUID) error {
	if len(chunks) == 0 {
		return nil
	}
	// Batch insert 200 at a time like existing storeBatchSize
	const batch = 200
	for i := 0; i < len(chunks); i += batch {
		end := i + batch
		if end > len(chunks) {
			end = len(chunks)
		}
		if err := storeChildBatch(ctx, db, expertID, chunks[i:end], embeddings[i:end], topicResults[i:end], sourceFile, parentIDs); err != nil {
			return fmt.Errorf("child batch %d: %w", i/batch, err)
		}
	}
	return nil
}

func storeChildBatch(ctx context.Context, db *pgxpool.Pool, expertID uuid.UUID, batch []TextChunk, embs [][]float32, topics []TopicResult, sourceFile string, parentIDs map[int]uuid.UUID) error {
	if len(batch) == 0 {
		return nil
	}
	// Build bulk insert with parent_id when available
	// course_chunks columns: id, expert_id, chunk_index, text, chunk_hash, section_path, topic, confidence, embedding, source_file, parent_id, parent_index, is_child
	for idx, ch := range batch {
		var vec pgvector.Vector
		if idx < len(embs) && len(embs[idx]) > 0 {
			vec = pgvector.NewVector(embs[idx])
		}
		topic := "general"
		conf := 0.5
		if idx < len(topics) && topics[idx].Topic != "" {
			topic = topics[idx].Topic
			conf = topics[idx].Confidence
		}
		parentID := uuid.Nil
		parentIdx := -1
		isChild := false
		if ch.ParentIndex >= 0 {
			parentIdx = ch.ParentIndex
			isChild = true
			if ch.ParentID != "" {
				if pid, err := uuid.Parse(ch.ParentID); err == nil {
					parentID = pid
				}
			} else if pid, ok := parentIDs[ch.ParentIndex]; ok {
				parentID = pid
			}
		}
		if parentID != uuid.Nil {
			_, err := db.Exec(ctx, `
				INSERT INTO course_chunks (id, expert_id, chunk_index, text, chunk_hash, section_path, topic, confidence, embedding, source_file, parent_id, parent_index, is_child)
				VALUES (gen_random_uuid(), $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
				ON CONFLICT (expert_id, chunk_hash) DO NOTHING`,
				expertID, ch.Index, ch.Text, ch.ChunkHash, ch.SectionPath, topic, conf, vec, sourceFile, parentID, parentIdx, isChild)
			if err != nil {
				return err
			}
		} else {
			_, err := db.Exec(ctx, `
				INSERT INTO course_chunks (id, expert_id, chunk_index, text, chunk_hash, section_path, topic, confidence, embedding, source_file, is_child)
				VALUES (gen_random_uuid(), $1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
				ON CONFLICT (expert_id, chunk_hash) DO NOTHING`,
				expertID, ch.Index, ch.Text, ch.ChunkHash, ch.SectionPath, topic, conf, vec, sourceFile, isChild)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
