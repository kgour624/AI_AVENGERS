package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GatewayAdapter wires any cheap LLM func to GatewayCheap without import cycle.
// Wire in main.go: NewGatewayAdapter(func(ctx,prompt string)(string,error){ return gw.CompleteCheap(ctx,prompt) })
type GatewayAdapter struct {
	complete func(ctx context.Context, prompt string) (string, error)
}

func NewGatewayAdapter(fn func(ctx context.Context, prompt string) (string, error)) *GatewayAdapter {
	return &GatewayAdapter{complete: fn}
}

func (a *GatewayAdapter) CompleteCheap(ctx context.Context, prompt string) (string, error) {
	if a == nil || a.complete == nil {
		return "", fmt.Errorf("gateway nil")
	}
	return a.complete(ctx, prompt)
}

func (a *GatewayAdapter) EvaluateCoverage(ctx context.Context, question string, chunks []ScoredChunk, draft string) (CoverageResult, error) {
	if a == nil || a.complete == nil {
		return CoverageResult{Answer: draft, Coverage: 1.0}, nil
	}
	ctxText := buildEvalContext(chunks)
	prompt := fmt.Sprintf("Evaluate if draft covers question using ONLY context.\nQuestion: %q\nContext:\n%s\nDraft: %q\n\nReturn JSON {\"coverage\": 0.0-1.0, \"reason\": \"...\", \"answer\": \"final answer\"}. Coverage <0.70 means gap.", question, ctxText, draft)
	raw, err := a.CompleteCheap(ctx, prompt)
	if err != nil {
		return CoverageResult{Answer: draft, Coverage: 1.0, Reason: "eval failed fail-open"}, nil
	}
	var parsed struct {
		Coverage float64 `json:"coverage"`
		Reason   string  `json:"reason"`
		Answer   string  `json:"answer"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return CoverageResult{Answer: draft, Coverage: 0.85, Reason: "non-json eval fail-open"}, nil
	}
	if parsed.Answer == "" {
		parsed.Answer = draft
	}
	if parsed.Coverage < 0 {
		parsed.Coverage = 0
	}
	if parsed.Coverage > 1 {
		parsed.Coverage = 1
	}
	return CoverageResult{Answer: parsed.Answer, Coverage: float32(parsed.Coverage), Reason: parsed.Reason}, nil
}

func (a *GatewayAdapter) RewriteQuery(ctx context.Context, question string, reason string) (string, error) {
	if a == nil || a.complete == nil {
		return "", fmt.Errorf("gateway nil")
	}
	prompt := buildReflectionPrompt(question, reason, 0)
	rewritten, err := a.CompleteCheap(ctx, prompt)
	if err != nil {
		return "", err
	}
	rewritten = strings.TrimSpace(strings.Trim(rewritten, "\"'"))
	if rewritten == "" {
		return question, nil
	}
	return rewritten, nil
}

func buildEvalContext(chunks []ScoredChunk) string {
	var b strings.Builder
	for i, c := range chunks {
		if i >= 3 {
			break
		}
		b.WriteString(fmt.Sprintf("[%d] %s\n", i+1, c.Text))
	}
	return b.String()
}
