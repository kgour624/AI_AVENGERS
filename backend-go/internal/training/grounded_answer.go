package training

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// groundedAnswerer asks a question against supplied context and judges the answer.
//
// ONE implementation, used by both the capability evaluation (per generated
// question) and the ingest smoke test (per topic probe). They need the same thing —
// an answer that is allowed to say "I do not know", a citation check that cannot be
// satisfied by accident, and a second opinion on whether the answer is actually
// supported — and two copies of that prompt would drift. Drift here is not cosmetic:
// the smoke test is a training gate, so a weaker copy would let a weaker expert
// through.
//
// WHY the citation rule is enforced by the prompt rather than by asking the model to
// report its own faithfulness: a self-report is the same class of evidence as the
// declared capability this area exists to replace.
type groundedAnswerer struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

func newGroundedAnswerer(gw *gateway.ModelGateway, logger *zap.Logger) *groundedAnswerer {
	return &groundedAnswerer{gateway: gw, logger: logger}
}

// Answer asks the question with the context as the only permitted source.
//
// The context must already carry [n] markers, and the caller must number them in the
// same order the prompt shows, so a citation the answer writes can be checked
// against what was actually supplied.
func (a *groundedAnswerer) Answer(ctx context.Context, question, contextBlock string) (string, error) {
	if a == nil || a.gateway == nil {
		return "", fmt.Errorf("grounded answerer: no model gateway")
	}
	prompt := fmt.Sprintf(`Answer the question using ONLY the context below.

Question: %s

Context:
%s

Rules:
- Cite the context you used with [n] markers matching the numbers above.
- If the context does not contain the answer, reply with exactly %s and nothing else.`,
		question, contextBlock, insufficientContextMarker)

	resp, err := a.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelStrong,
		UserPrompt:  prompt,
		MaxTokens:   2048,
		Temperature: 0.2,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Judge asks a second call whether the answer is actually supported by the context.
//
// A separate call, on a different model tier, for the same reason storage
// verification re-reads the database instead of trusting the store's own count: the
// component that produced the answer is not trusted to grade it.
func (a *groundedAnswerer) Judge(ctx context.Context, question, contextBlock, answer string) string {
	if a == nil || a.gateway == nil {
		return "unverifiable"
	}
	prompt := fmt.Sprintf(`Question: %s

Context:
%s

Answer:
%s

Is the answer fully supported by the context above? Judge only whether the context
supports it, not whether the style is good.

Return ONLY JSON: {"verdict":"supported"|"refuted"|"unverifiable","reason":"..."}`,
		question, contextBlock, answer)

	resp, err := a.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   1024,
		Temperature: 0,
	})
	if err != nil {
		a.logger.Warn("grounded answerer: judge call failed (recorded as unverifiable)",
			zap.Error(err))
		return "unverifiable"
	}

	verdict := parseJudgeVerdict(resp.Content)
	if verdict == "unverifiable" {
		// An unparseable judge and a genuinely unsupported answer look identical in
		// the results, so the raw response is logged (clipped) to tell them apart.
		a.logger.Warn("grounded answerer: judge verdict unparseable (recorded as unverifiable)",
			zap.String("raw_response", clipPromptText(resp.Content, 200)))
	}
	return verdict
}

// buildNumberedContext renders chunks as [n] blocks in the order given.
//
// Shared so the numbering the model cites and the numbering a checker validates can
// never disagree.
func buildNumberedContext(texts []string, limitPerText int) string {
	var sb strings.Builder
	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, clipPromptText(text, limitPerText)))
	}
	return sb.String()
}

// refusalOrCitationFailure names the first reason an answer is unusable, or "" when
// it is at least well-formed: non-empty, not a refusal, and carrying a citation.
//
// Pure. Kept separate from the judge so the deterministic part of the verdict is
// testable without a model.
func refusalOrCitationFailure(answer string) string {
	if strings.Contains(answer, insufficientContextMarker) {
		return FailureRefused
	}
	if strings.TrimSpace(answer) == "" {
		return FailureEmpty
	}
	if countCitationMarkers(answer) == 0 {
		return FailureNotCited
	}
	return ""
}
