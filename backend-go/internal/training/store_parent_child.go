package training

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// storeChunksWithParents is Phase 1.5 dual-write: same batching as storeChunks
// but honors parent_id/parent_index/is_child. Uses ON CONFLICT DO NOTHING +
// post-insert UPDATE to backfill parent link for deduped rows.
func (p *IngestionPipeline) storeChunksWithParents(
	ctx context.Context,
	jobID uuid.UUID,
	expertID uuid.UUID,
	chunks []TextChunk,
	topics []TopicResult,
	embeddings [][]float32,
	sourceFile string,
	parentIDByIndex map[int]uuid.UUID,
) ([]uuid.UUID, StoreStats, error) {
	chunkIDs := make([]uuid.UUID, len(chunks))
	stats := StoreStats{Parsed: len(chunks)}
	embedProvider, embedModel := p.currentEmbedding(ctx)
	var embedModelArg interface{}
	if embedModel != "" {
		embedModelArg = embedModel
	}
	for i := range chunks {
		if chunks[i].ChunkHash == "" {
			chunks[i].ChunkHash = HashChunkText(chunks[i].Text)
		}
	}
	stats.UniqueHashes, stats.Duplicates = countDistinctHashes(chunks)
	if err := p.db.QueryRow(ctx, `SELECT COUNT(*) FROM course_chunks WHERE expert_id=$1 AND source_file=$2`, expertID, sourceFile).Scan(&stats.PreexistingForFile); err != nil {
		return nil, stats, fmt.Errorf("count existing chunks for file: %w", err)
	}
	covered := make(map[string]struct{}, len(chunks))
	const storeBatchSize = 200
	for batchStart := 0; batchStart < len(chunks); batchStart += storeBatchSize {
		batchEnd := min(batchStart+storeBatchSize, len(chunks))
		var sb strings.Builder
		sb.WriteString(`INSERT INTO course_chunks
			(expert_id, chunk_text, chunk_index, topic, subtopic, source_file, embedding, chunk_hash,
			 embedding_provider, embedding_model, section_path, parent_id, parent_index, is_child)
		 VALUES `)
		args := make([]interface{}, 0, (batchEnd-batchStart)*14)
		for i := batchStart; i < batchEnd; i++ {
			if i > batchStart {
				sb.WriteString(",")
			}
			base := len(args)
			sb.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12, base+13, base+14))
			topic := topics[i].Topic
			subtopic := ""
			if i < len(topics) {
				subtopic = topics[i].Subtopic
			}
			var parentID interface{}
			parentIdx := 0
			isChild := false
			if chunks[i].ParentIndex >= 0 {
				isChild = true
				parentIdx = chunks[i].ParentIndex
				if chunks[i].ParentID != "" {
					if pid, err := uuid.Parse(chunks[i].ParentID); err == nil && pid != uuid.Nil {
						parentID = pid
					} else if pid2, ok := parentIDByIndex[chunks[i].ParentIndex]; ok {
						parentID = pid2
					}
				} else if pid2, ok := parentIDByIndex[chunks[i].ParentIndex]; ok {
					parentID = pid2
				}
			}
			args = append(args,
				expertID, chunks[i].Text, chunks[i].Index, topic, subtopic,
				sourceFile, pgvector.NewVector(embeddings[i]), chunks[i].ChunkHash,
				embedProvider, embedModelArg, chunks[i].SectionPath,
				parentID, parentIdx, isChild,
			)
		}
		sb.WriteString(` ON CONFLICT (expert_id, chunk_hash) DO NOTHING`)
		tag, err := p.db.Exec(ctx, sb.String(), args...)
		if err != nil {
			return nil, stats, fmt.Errorf("failed to insert chunks %d..%d: %w", batchStart, batchEnd-1, err)
		}
		batchInserted := int(tag.RowsAffected())
		stats.Inserted += batchInserted
		idByHash, err := p.lookupChunkIDs(ctx, expertID, chunks[batchStart:batchEnd])
		if err != nil {
			return nil, stats, err
		}
		for i := batchStart; i < batchEnd; i++ {
			id, ok := idByHash[chunks[i].ChunkHash]
			if !ok {
				return nil, stats, fmt.Errorf("chunk %d missing after insert (hash %s)", i, chunks[i].ChunkHash)
			}
			chunkIDs[i] = id
			covered[chunks[i].ChunkHash] = struct{}{}
		}
		// Patch parent link for deduped rows where stored parent_id IS NULL
		var ids []uuid.UUID
		var pids []uuid.UUID
		var pidx []int
		for i := batchStart; i < batchEnd; i++ {
			if chunks[i].ParentIndex < 0 {
				continue
			}
			pidVal, _ := parentIDByIndex[chunks[i].ParentIndex]
			if pidVal == uuid.Nil && chunks[i].ParentID != "" {
				if parsed, perr := uuid.Parse(chunks[i].ParentID); perr == nil {
					pidVal = parsed
				}
			}
			if pidVal == uuid.Nil {
				continue
			}
			ids = append(ids, chunkIDs[i])
			pids = append(pids, pidVal)
			pidx = append(pidx, chunks[i].ParentIndex)
		}
		if len(ids) > 0 {
			_, _ = p.db.Exec(ctx, `UPDATE course_chunks AS c SET parent_id=COALESCE(c.parent_id, v.pid), parent_index=CASE WHEN c.parent_id IS NULL THEN v.pidx ELSE c.parent_index END, is_child=CASE WHEN c.parent_id IS NULL THEN TRUE ELSE c.is_child END FROM (SELECT unnest($1::uuid[]) as id, unnest($2::uuid[]) as pid, unnest($3::int[]) as pidx) v WHERE c.id=v.id AND c.parent_id IS NULL`, ids, pids, pidx)
		}
		p.updateJobProgress(ctx, jobID, batchEnd, len(chunks))
		p.emit(ctx, jobID, expertID, StageStoring, "chunk_stored", map[string]interface{}{
			"chunks_done":    len(covered),
			"chunks_total":   stats.UniqueHashes,
			"batch_size":     batchEnd - batchStart,
			"batch_inserted": batchInserted,
		})
	}
	stats.Reused = stats.UniqueHashes - stats.Inserted
	if stats.Reused < 0 {
		stats.Reused = 0
	}
	p.linkChunks(ctx, chunkIDs)
	return chunkIDs, stats, nil
}


