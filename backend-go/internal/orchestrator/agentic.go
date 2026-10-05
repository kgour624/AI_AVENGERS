package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"ai_avengers/backend/internal/ml"
)

type AgenticOrchestrator struct {
	db       *pgxpool.Pool
	embedder ml.Embedder
	gateway  GatewayCheap
	cache    *SemanticCache
	maxHops  int
	rrfK     int
	logger   *zap.Logger
}

func NewAgenticOrchestrator(db *pgxpool.Pool, embedder ml.Embedder, gateway GatewayCheap, cache *SemanticCache, logger *zap.Logger) *AgenticOrchestrator {
	if cache == nil {
		cache = NewSemanticCache()
	}
	return &AgenticOrchestrator{db: db, embedder: embedder, gateway: gateway, cache: cache, maxHops: defaultMaxHops, rrfK: 60, logger: logger}
}

func (o *AgenticOrchestrator) Answer(ctx context.Context, expertID uuid.UUID, question string) (string, []ScoredChunk, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return "", nil, fmt.Errorf("empty question")
	}
	if o.embedder == nil {
		return "", nil, fmt.Errorf("embedder not wired")
	}
	embedding, err := o.embedder.EmbedSingle(ctx, question)
	if err != nil {
		return "", nil, fmt.Errorf("embed failed: %w", err)
	}
	if o.cache != nil {
		if chunks, resp, ok := o.cache.Get(ctx, embedding); ok {
			if o.logger != nil {
				o.logger.Info("agentic cache hit", zap.Int("chunks", len(chunks)))
			}
			return resp, chunks, nil
		}
	}
	var checkpoints []hopCheckpoint
	currentQuery := question
	currentEmbedding := embedding
	var finalAnswer string
	var finalChunks []ScoredChunk
	var lastCoverage float32
	for hop := 0; hop < o.maxHops; hop++ {
		cp := hopCheckpoint{Hop: hop + 1, Query: currentQuery}
		lists, err := o.scatterGather(ctx, expertID, currentQuery, currentEmbedding)
		if err != nil && len(lists) == 0 {
			if o.logger != nil {
				o.logger.Warn("scatter-gather failed", zap.Int("hop", hop+1), zap.Error(err))
			}
			if hop == 0 {
				return "", nil, err
			}
			break
		}
		merged := mergeRRF(lists, o.rrfK)
		top3 := pruneTopK(merged, 3)
		cp.Chunks = len(top3)
		if len(top3) == 0 {
			checkpoints = append(checkpoints, cp)
			break
		}
		coverage, draft, evalErr := o.evaluateWithCheap(ctx, currentQuery, top3)
		if evalErr != nil {
			if o.logger != nil {
				o.logger.Warn("cheap gateway failed", zap.Error(evalErr))
			}
			finalAnswer = buildFallbackAnswer(currentQuery, top3)
			finalChunks = top3
			checkpoints = append(checkpoints, cp)
			break
		}
		cp.Coverage = coverage
		checkpoints = append(checkpoints, cp)
		lastCoverage = coverage
		finalAnswer = draft
		finalChunks = top3
		if !needsReflection(coverage, hop+1, o.maxHops) {
			break
		}
		if o.gateway == nil {
			break
		}
		rewritten, err := o.gateway.RewriteQuery(ctx, currentQuery, fmt.Sprintf("coverage %.2f hop %d", coverage, hop+1))
		if err != nil || strings.TrimSpace(rewritten) == "" {
			break
		}
		rewritten = strings.TrimSpace(rewritten)
		checkpoints[len(checkpoints)-1].Rewritten = rewritten
		currentQuery = rewritten
		if emb, err := o.embedder.EmbedSingle(ctx, currentQuery); err == nil && len(emb) > 0 {
			currentEmbedding = emb
		}
		if o.logger != nil {
			o.logger.Info("reflection hop", zap.Int("hop", hop+1), zap.Float32("coverage", coverage), zap.String("rewritten", rewritten))
		}
	}
	if o.logger != nil && len(checkpoints) > 0 {
		o.logger.Info("agentic checkpoints", zap.Int("hops", len(checkpoints)), zap.Float32("coverage", lastCoverage))
	}
	if o.cache != nil && finalAnswer != "" {
		o.cache.Set(ctx, embedding, finalChunks, finalAnswer)
	}
	if finalAnswer == "" && len(finalChunks) > 0 {
		finalAnswer = buildFallbackAnswer(question, finalChunks)
	}
	return finalAnswer, finalChunks, nil
}

func (o *AgenticOrchestrator) scatterGather(ctx context.Context, expertID uuid.UUID, _ string, emb []float32) ([][]ScoredChunk, error) {
	g, gCtx := errgroup.WithContext(ctx)
	lists := make([][]ScoredChunk, 3)
	g.Go(func() error {
		qCtx, cancel := context.WithTimeout(gCtx, 800*time.Millisecond)
		defer cancel()
		c, err := o.retrieveParentChild(qCtx, expertID, emb)
		if err != nil {
			return err
		}
		lists[0] = c
		return nil
	})
	g.Go(func() error {
		qCtx, cancel := context.WithTimeout(gCtx, 800*time.Millisecond)
		defer cancel()
		c, err := o.retrieveHybrid(qCtx, expertID, emb)
		if err != nil {
			return err
		}
		lists[1] = c
		return nil
	})
	g.Go(func() error {
		qCtx, cancel := context.WithTimeout(gCtx, 800*time.Millisecond)
		defer cancel()
		c, _ := o.retrieveGraph(qCtx, expertID, emb)
		lists[2] = c
		return nil
	})
	if err := g.Wait(); err != nil {
		return filterEmpty(lists), err
	}
	return filterEmpty(lists), nil
}
