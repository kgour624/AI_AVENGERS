package orchestrator

import (
	"context"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

func (o *AgenticOrchestrator) retrieveParentChild(ctx context.Context, expertID uuid.UUID, embedding []float32) ([]ScoredChunk, error) {
	if o.db == nil || len(embedding) == 0 {
		return nil, nil
	}
	vec := pgvector.NewVector(embedding)
	rows, err := o.db.Query(ctx, `SELECT id, chunk_text, COALESCE(topic,''), COALESCE(section_path,''), COALESCE(source_file,''), chunk_index FROM course_chunks WHERE expert_id=$1 AND parent_id IS NOT NULL ORDER BY embedding <=> $2 LIMIT 50`, expertID, vec)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScoredChunk
	rank := 0
	for rows.Next() {
		var id uuid.UUID
		var txt, topic, sp, sf string
		var cidx int
		if err := rows.Scan(&id, &txt, &topic, &sp, &sf, &cidx); err != nil {
			continue
		}
		out = append(out, ScoredChunk{ID: id, Text: txt, Topic: topic, Score: float64(50 - rank), Rank: rank, Source: "parent-child"})
		rank++
	}
	return out, rows.Err()
}

func (o *AgenticOrchestrator) retrieveHybrid(ctx context.Context, expertID uuid.UUID, embedding []float32) ([]ScoredChunk, error) {
	if o.db == nil || len(embedding) == 0 {
		return nil, nil
	}
	vec := pgvector.NewVector(embedding)
	rows, err := o.db.Query(ctx, `SELECT id, chunk_text, COALESCE(topic,''), COALESCE(section_path,''), COALESCE(source_file,''), chunk_index FROM course_chunks WHERE expert_id=$1 ORDER BY embedding <=> $2 LIMIT 15`, expertID, vec)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ScoredChunk
	rank := 0
	for rows.Next() {
		var id uuid.UUID
		var txt, topic, sp, sf string
		var cidx int
		if err := rows.Scan(&id, &txt, &topic, &sp, &sf, &cidx); err != nil {
			continue
		}
		out = append(out, ScoredChunk{ID: id, Text: txt, Topic: topic, Score: float64(15 - rank), Rank: rank, Source: "hybrid"})
		rank++
	}
	return out, rows.Err()
}

func (o *AgenticOrchestrator) retrieveGraph(ctx context.Context, expertID uuid.UUID, embedding []float32) ([]ScoredChunk, error) {
	if o.db == nil {
		return nil, nil
	}
	rows, err := o.db.Query(ctx, `SELECT c.id, c.chunk_text, COALESCE(c.topic,''), COALESCE(c.section_path,''), COALESCE(c.source_file,''), c.chunk_index FROM course_chunks c JOIN concept_chunk_links l ON l.chunk_id = c.id JOIN concepts co ON co.id = l.concept_id WHERE c.expert_id=$1 ORDER BY c.embedding <=> $2 LIMIT 10`, expertID, pgvector.NewVector(embedding))
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	var out []ScoredChunk
	rank := 0
	for rows.Next() {
		var id uuid.UUID
		var txt, topic, sp, sf string
		var cidx int
		if err := rows.Scan(&id, &txt, &topic, &sp, &sf, &cidx); err != nil {
			continue
		}
		out = append(out, ScoredChunk{ID: id, Text: txt, Topic: topic, Score: float64(10 - rank), Rank: rank, Source: "graph"})
		rank++
	}
	return out, rows.Err()
}
