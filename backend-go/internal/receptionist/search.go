package receptionist

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Small Interface at Boundary - SOLID I & D (Ultimate Go rule)
type SearchProvider interface {
	Search(ctx context.Context, query string) ([]Point, error)
}

// Business Model - storage se alag (Three Model Layers)
type Point struct {
	ID               string `json:"id"`
	QText            string `json:"q_text"`
	Status           string `json:"status"` // draft -> discussing -> approved/dismissed/judge_explained
	Source           string `json:"source"` // search
	Citation         string `json:"citation,omitempty"`
	JudgeExplained   bool   `json:"judge_explained,omitempty"`
	DismissedReason  string `json:"dismissed_reason,omitempty"`
}

// App Model - JSON tags yahan
type AppPoint struct {
	QText string `json:"q_text"`
}

// GatewaySearch - Cheap tier Temp 0.6 (Arpit + Byte-by-Byte learnings)
type GatewaySearch struct {
	gw     *gateway.ModelGateway
	rdb    *redis.Client
	logger *zap.Logger
}

func NewGatewaySearch(gw *gateway.ModelGateway, rdb *redis.Client, logger *zap.Logger) *GatewaySearch {
	return &GatewaySearch{gw: gw, rdb: rdb, logger: logger}
}

// normalize for cache key - DSA: simple correct first
func normalizeQuery(q string) string {
	return strings.ToLower(strings.TrimSpace(strings.ReplaceAll(q, " ", "_")))
}

func (s *GatewaySearch) Search(ctx context.Context, query string) ([]Point, error) {
	// 1. Redis cache 1hr check - Postgres hi truth, Redis sirf fast path
	cacheKey := fmt.Sprintf("search:%s", normalizeQuery(query))
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		var pts []Point
		if err := json.Unmarshal([]byte(cached), &pts); err == nil {
			s.logger.Info("search cache hit", zap.String("key", cacheKey), zap.Int("points", len(pts)))
			return pts, nil
		}
	}

	// 2. Cheap LLM call - Task + Constraint + Output Format mandatory (Byte-by-Byte)
	// Temp 0.6 for claim extraction (Arpit table: 0.6-0.7)
	systemPrompt := `You are a checklist search engine. Task: Generate required sections/questions for the given document type. Constraint: Return ONLY JSON array, no markdown, no fluff. Output Format: [{"q_text":"..."}] exactly 8-10 items, 2026 industry standard, verifiable. If not in context, return empty array.`
	userPrompt := fmt.Sprintf(`Agenda: %s. Generate checklist points as JSON.`, query)

	req := gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.6,
		MaxTokens:    1500,
	}

	// gateway.Call is expensive while loop - harness is job (Arpit: Agent = while loop)
	resp, err := s.gw.Call(ctx, req)
	if err != nil {
		s.logger.Error("search llm failed", zap.Error(err), zap.String("query", query))
		return nil, fmt.Errorf("search failed: %w", err)
	}
	if resp.Content == "" {
		return nil, fmt.Errorf("empty search content")
	}

	// 3. Parse integrity - parse() at every transition (Ultimate Go)
	var raw []map[string]string
	if err := json.Unmarshal([]byte(resp.Content), &raw); err != nil {
		// try extract JSON array if model added fluff
		start := strings.Index(resp.Content, "[")
		end := strings.LastIndex(resp.Content, "]")
		if start >= 0 && end > start {
			json.Unmarshal([]byte(resp.Content[start:end+1]), &raw)
		} else {
			return nil, fmt.Errorf("parse search json failed: %w", err)
		}
	}

	pts := make([]Point, 0, len(raw))
	for i, r := range raw {
		q := strings.TrimSpace(r["q_text"])
		if q == "" {
			continue
		}
		pts = append(pts, Point{
			ID:     fmt.Sprintf("p%d_%d", time.Now().UnixMilli(), i),
			QText:  q,
			Status: "draft",
			Source: "search",
		})
	}
	if len(pts) == 0 {
		return nil, fmt.Errorf("no points generated")
	}

	// 4. Cache 1hr - best effort, fail bhi ho to Postgres se chalega
	b, _ := json.Marshal(pts)
	_ = s.rdb.Set(ctx, cacheKey, b, time.Hour).Err()
	s.logger.Info("search done", zap.String("query", query), zap.Int("points", len(pts)), zap.String("model", "cheap"))

	return pts, nil
}
