package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/chinawall"
)

// tryParentChildRetrieval is Phase 2 critical path (context assembler side).
// Mirrors knowledge.ParentChildRetriever but lives in context package to avoid import cycle.
// Additive + feature-flagged: when retrieval_config.enable_parent_child=false, returns (nil,nil) so caller falls back.
// Returns parent page_text as CourseChunk (Text=parent page, Topic from best child, heading via SectionPath).
func (a *Assembler) tryParentChildRetrieval(ctx context.Context, expertID uuid.UUID, embedding []float32, question string, limit int) ([]chinawall.CourseChunk, error) {
	// Load retrieval_config best-effort (defaults if missing)
	cfg := struct {
		Enable     bool `json:"enable_parent_child"`
		TopKChild  int  `json:"top_k_children"`
		TopKParent int  `json:"top_k_parents"`
	}{}
	var raw []byte
	if err := a.db.QueryRow(ctx, `SELECT value FROM system_settings WHERE key='retrieval_config'`).Scan(&raw); err == nil && len(raw) > 0 {
		_ = jsonUnmarshalRaw(raw, &cfg)
	}
	if !cfg.Enable {
		return nil, nil
	}
	topK := cfg.TopKChild
	if topK == 0 {
		topK = 50
	}
	targetParents := cfg.TopKParent
	if targetParents == 0 {
		targetParents = 3
	}
	if limit > 0 && targetParents > limit {
		targetParents = limit
	}
	qctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()
	vec := pgvector.NewVector(embedding)
	// Use non-reserved alias 'c' for course_chunks to avoid PG reserved keyword 'ch'
	rows, err := a.db.Query(qctx, `
		SELECT c.id, c.parent_id, c.parent_index, c.chunk_text, c.section_path, COALESCE(c.topic,''), c.embedding <=> $1 AS distance
		FROM course_chunks c
		WHERE c.expert_id=$2 AND c.parent_id IS NOT NULL
		ORDER BY c.embedding <=> $1
		LIMIT $3`, vec, expertID, topK)
	if err != nil {
		return nil, fmt.Errorf("parent-child ANN failed: %w", err)
	}
	defer rows.Close()
	type hit struct {
		ID          uuid.UUID
		ParentID    *uuid.UUID
		ParentIndex int
		Text        string
		SectionPath string
		Topic       string
		Dist        float64
		Score       float64
	}
	var children []hit
	for rows.Next() {
		var h hit
		var pid *uuid.UUID
		var pIdx *int
		var txt, sp, topic string
		var dist float64
		var id uuid.UUID
		if err := rows.Scan(&id, &pid, &pIdx, &txt, &sp, &topic, &dist); err != nil {
			continue
		}
		h.ID = id
		h.ParentID = pid
		if pIdx != nil {
			h.ParentIndex = *pIdx
		} else {
			h.ParentIndex = -1
		}
		h.Text = txt
		h.SectionPath = sp
		h.Topic = topic
		h.Dist = dist
		h.Score = -dist
		children = append(children, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, nil
	}
	// Rerank children only (safe 180 tok) via sidecar if available
	// Rerank returns []RerankResult{Index, Score} sorted desc. Map back to children by Index.
	if a.sidecar != nil && strings.TrimSpace(question) != "" {
		docs := make([]string, len(children))
		for i, c := range children {
			docs[i] = c.Text
		}
		if reranked, err := a.sidecar.Rerank(qctx, question, docs, len(children)); err == nil {
			for _, r := range reranked {
				if r.Index >= 0 && r.Index < len(children) {
					children[r.Index].Score = float64(r.Score)
				}
			}
		} else if a.logger != nil {
			a.logger.Warn("parent-child rerank failed, fallback to ANN", zap.Error(err))
		}
	}
	// sort by score desc
	for i := 0; i < len(children)-1; i++ {
		for j := i + 1; j < len(children); j++ {
			if children[j].Score > children[i].Score {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
	// dedupe to 3 unique parents
	seen := make(map[uuid.UUID]bool, targetParents)
	var deduped []hit
	var parentIDs []uuid.UUID
	for _, h := range children {
		if h.ParentID == nil {
			continue
		}
		pid := *h.ParentID
		if seen[pid] {
			continue
		}
		seen[pid] = true
		deduped = append(deduped, h)
		parentIDs = append(parentIDs, pid)
		if len(parentIDs) >= targetParents {
			break
		}
	}
	if len(parentIDs) == 0 {
		return nil, nil
	}
	prows, err := a.db.Query(qctx, `SELECT id, page_text, COALESCE(section_path,''), token_count, COALESCE(source_file,''), COALESCE(chunk_index,0) FROM expert_pages WHERE id = ANY($1)`, parentIDs)
	if err != nil {
		return nil, err
	}
	defer prows.Close()
	type pdoc struct {
		Text        string
		SectionPath string
		TokenCount  int
		SourceFile  string
		ChunkIndex  int
	}
	parentMap := make(map[uuid.UUID]pdoc, len(parentIDs))
	for prows.Next() {
		var id uuid.UUID
		var txt, sp, sf string
		var tc, cidx int
		if err := prows.Scan(&id, &txt, &sp, &tc, &sf, &cidx); err != nil {
			continue
		}
		_ = tc
		parentMap[id] = pdoc{Text: txt, SectionPath: sp, SourceFile: sf, ChunkIndex: cidx}
	}
	var out []chinawall.CourseChunk
	for _, h := range deduped {
		if h.ParentID == nil {
			continue
		}
		pd, ok := parentMap[*h.ParentID]
		if !ok {
			continue
		}
		out = append(out, chinawall.CourseChunk{
			ID:         *h.ParentID,
			Text:       pd.Text,
			Topic:      h.Topic,
			RerankScore: float32(h.Score),
			SourceFile: pd.SourceFile,
			ChunkIndex: pd.ChunkIndex,
		})
		_ = pd.SectionPath
	}
	return out, nil
}

func jsonUnmarshalRaw(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
