package knowledge

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

// Reranker is sidecar Cross-Encoder. Rerank is 512 tok max — Phase 2 reranks
// 120-180 tok children (safe) not 1500 tok parents (truncation + O(N^2) trap).
// 50*180^2 vs 5*1500^2 = ~7x cheaper.
type Reranker interface {
	Rerank(ctx context.Context, query string, docs []string) ([]float64, error)
}

// ParentChildRetriever implements Phase 2 critical path:
// ANN Top 50 children -> Cross-Encoder rerank children -> loop till 3 uniques -> fetch parents.
// Config-driven TopK=50, TargetParents=3, scan up to 50, HashSet O(K), ctx->pgx 800ms.
type ParentChildRetriever struct {
	db       *pgxpool.Pool
	reranker Reranker
	logger   *zap.Logger
}

func NewParentChildRetriever(db *pgxpool.Pool, reranker Reranker, logger *zap.Logger) *ParentChildRetriever {
	return &ParentChildRetriever{db: db, reranker: reranker, logger: logger}
}

type ChildHit struct {
	ID          uuid.UUID
	ParentID    *uuid.UUID
	ParentIndex int
	Text        string
	SectionPath string
	Distance    float64
	RerankScore float64
}

type ParentDoc struct {
	ID          uuid.UUID
	PageText    string
	SectionPath string
	TokenCount  int
	SourceFile  string
	Score       float64
	ChildIDs    []uuid.UUID
}

// SearchWithParent is the Phase 2 entry: search children, rerank, dedupe, fetch parents.
// Additive + feature-flagged: when enable_parent_child=false, returns (nil,nil) so caller falls back to legacy reader.
func (r *ParentChildRetriever) SearchWithParent(ctx context.Context, expertID uuid.UUID, queryEmbedding []float32, queryText string) ([]ParentDoc, error) {
	cfg := LoadRetrievalConfig(ctx, r.db)
	if !cfg.EnableParentChild {
		return nil, nil
	}
	topK := cfg.TopKChildren
	if topK == 0 {
		topK = 50
	}
	targetParents := cfg.TopKParents
	if targetParents == 0 {
		targetParents = 3
	}
	rerankTopN := cfg.RerankTopN
	if rerankTopN == 0 {
		rerankTopN = 10
	}

	// ctx propagation with 800ms timeout -> pgx aborts DB query and avoids pool leak (Tweak 2)
	qctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	vec := pgvector.NewVector(queryEmbedding)
	rows, err := r.db.Query(qctx, `
		SELECT id, parent_id, parent_index, text, section_path, embedding <=> $1 AS distance
		FROM course_chunks
		WHERE expert_id=$2 AND parent_id IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $3`, vec, expertID, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []ChildHit
	for rows.Next() {
		var ch ChildHit
		var pid *uuid.UUID
		var pIdx *int
		var txt, sp string
		var dist float64
		var id uuid.UUID
		if err := rows.Scan(&id, &pid, &pIdx, &txt, &sp, &dist); err != nil {
			continue
		}
		ch.ID = id
		ch.ParentID = pid
		if pIdx != nil {
			ch.ParentIndex = *pIdx
		} else {
			ch.ParentIndex = -1
		}
		ch.Text = txt
		ch.SectionPath = sp
		ch.Distance = dist
		ch.RerankScore = -dist
		children = append(children, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return nil, nil
	}

	// Rerank children only (512 safe, O(N^2) trap avoided)
	if r.reranker != nil && queryText != "" {
		docs := make([]string, len(children))
		for i, c := range children {
			docs[i] = c.Text
		}
		// reranker ctx also inherits timeout; abort on cancel
		if scores, err := r.reranker.Rerank(qctx, queryText, docs); err == nil && len(scores) == len(children) {
			for i := range children {
				children[i].RerankScore = scores[i]
			}
		} else if r.logger != nil && err != nil {
			r.logger.Warn("rerank failed, fallback to ANN distance", zap.Error(err))
		}
	}

	sort.Slice(children, func(i, j int) bool { return children[i].RerankScore > children[j].RerankScore })

	// Tweak 1+4: loop till N unique parents, scan up to 50 (HashSet O(K), K=50)
	seen := make(map[uuid.UUID]bool, targetParents)
	var deduped []ChildHit
	var parentIDs []uuid.UUID
	for _, ch := range children {
		if ch.ParentID == nil {
			continue
		}
		pid := *ch.ParentID
		if seen[pid] {
			continue
		}
		seen[pid] = true
		deduped = append(deduped, ch)
		parentIDs = append(parentIDs, pid)
		if len(parentIDs) >= targetParents {
			break
		}
		if len(parentIDs) >= len(children) {
			break
		}
	}
	if len(parentIDs) == 0 {
		return nil, nil
	}
	// fetch parents IN (heap dedupe already done, IN size=3 typical)
	prows, err := r.db.Query(qctx, `SELECT id, page_text, section_path, token_count, source_file FROM expert_pages WHERE id = ANY($1)`, parentIDs)
	if err != nil {
		return nil, err
	}
	defer prows.Close()
	parentMap := make(map[uuid.UUID]*ParentDoc, len(parentIDs))
	for prows.Next() {
		var id uuid.UUID
		var txt, sp, sf string
		var tc int
		var sfPtr *string
		if err := prows.Scan(&id, &txt, &sp, &tc, &sfPtr); err != nil {
			continue
		}
		if sfPtr != nil {
			sf = *sfPtr
		}
		parentMap[id] = &ParentDoc{ID: id, PageText: txt, SectionPath: sp, TokenCount: tc, SourceFile: sf}
	}
	if err := prows.Err(); err != nil {
		return nil, err
	}
	// order by best child score, attach child ids + score + heading_path citation
	var result []ParentDoc
	for _, ch := range deduped {
		if ch.ParentID == nil {
			continue
		}
		pd, ok := parentMap[*ch.ParentID]
		if !ok {
			continue
		}
		pd.Score = ch.RerankScore
		pd.ChildIDs = append(pd.ChildIDs, ch.ID)
		result = append(result, *pd)
	}
	return result, nil
}
