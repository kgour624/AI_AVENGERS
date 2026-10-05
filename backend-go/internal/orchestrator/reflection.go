package orchestrator

import (
	"context"
	"fmt"
	"strings"
)

// MaxHops = 3 per spec: 1 initial + up to 2 rewrites. Loop bounded, checkpoint per hop.
const defaultMaxHops = 3

// CoverageResult is Cheap gateway's self-evaluation of answer coverage.
type CoverageResult struct {
	Answer   string  `json:"answer"`
	Coverage float32 `json:"coverage"` // 0..1, cheap model's self-score that chunks cover question
	Reason   string  `json:"reason"`
}

// GatewayCheap is subset of gateway.ModelGateway used by orchestrator (cheap model only).
// Implemented as interface to avoid import cycle; adapter in orchestrator_adapter.go wires real gateway.
type GatewayCheap interface {
	CompleteCheap(ctx context.Context, prompt string) (string, error)
	EvaluateCoverage(ctx context.Context, question string, chunks []ScoredChunk, draft string) (CoverageResult, error)
	RewriteQuery(ctx context.Context, question string, reason string) (string, error)
}

// needsReflection decides whether to loop. Threshold 0.70 coverage; hop < MaxHops.
func needsReflection(coverage float32, hop, maxHops int) bool {
	if maxHops <= 0 {
		maxHops = defaultMaxHops
	}
	return hop < maxHops && coverage < 0.70
}

// buildReflectionPrompt builds cheap rewrite prompt (bounded, no unbounded string growth).
func buildReflectionPrompt(question, reason string, hop int) string {
	question = strings.TrimSpace(question)
	reason = strings.TrimSpace(reason)
	if len(reason) > 300 {
		reason = reason[:300]
	}
	return fmt.Sprintf("Rewrite this question to better retrieve relevant chunks (hop %d/3). Original: %q. Gap: %q. Return ONLY the rewritten question, no explanation.", hop+1, question, reason)
}

// hopCheckpoint records state per hop for bounded loop observability.
type hopCheckpoint struct {
	Hop      int     `json:"hop"`
	Query    string  `json:"query"`
	Coverage float32 `json:"coverage"`
	Chunks   int     `json:"chunks"`
	Rewritten string `json:"rewritten,omitempty"`
}
