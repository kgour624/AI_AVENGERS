package orchestrator

import (
	"context"
	"fmt"
	"strings"
)

func filterEmpty(lists [][]ScoredChunk) [][]ScoredChunk {
	out := make([][]ScoredChunk, 0, len(lists))
	for _, l := range lists {
		if len(l) > 0 {
			out = append(out, l)
		}
	}
	return out
}

func (o *AgenticOrchestrator) evaluateWithCheap(ctx context.Context, question string, chunks []ScoredChunk) (float32, string, error) {
	if o.gateway == nil {
		return 0, buildFallbackAnswer(question, chunks), nil
	}
	prompt := buildDraftPrompt(question, chunks)
	draft, err := o.gateway.CompleteCheap(ctx, prompt)
	if err != nil {
		return 0, "", err
	}
	cov, err := o.gateway.EvaluateCoverage(ctx, question, chunks, draft)
	if err != nil {
		return 1.0, draft, nil
	}
	return cov.Coverage, cov.Answer, nil
}

func buildDraftPrompt(question string, chunks []ScoredChunk) string {
	var b strings.Builder
	b.WriteString("Answer concisely using ONLY context below. If insufficient say so.\n\nContext:\n")
	for i, c := range chunks {
		if i >= 3 {
			break
		}
		b.WriteString(fmt.Sprintf("[%d] %s\n", i+1, c.Text))
	}
	b.WriteString("\nQuestion: ")
	b.WriteString(question)
	b.WriteString("\nAnswer:")
	return b.String()
}

func buildFallbackAnswer(_ string, chunks []ScoredChunk) string {
	if len(chunks) == 0 {
		return "I don't have enough context to answer that."
	}
	var b strings.Builder
	b.WriteString(chunks[0].Text)
	if len(chunks) > 1 && len(b.String()) < 400 {
		b.WriteString("\n\n")
		b.WriteString(chunks[1].Text)
	}
	return b.String()
}
