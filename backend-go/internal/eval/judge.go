package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Judge is the T3 LLM-judge layer for the golden-set harness. It is OPT-IN:
// when no judge is wired the runner uses only the deterministic Score, so CI
// stays reproducible and offline. Wiring a judge ADDS a component check ("judge")
// to a case's verdict; it never removes the deterministic checks and never
// changes how the vital-failure gate is counted.
//
// WHY an interface: the HTTP/CI harness and any future offline judge share the
// same runner. The real implementation (ModelJudge) calls the model gateway; a
// test injects a fake so no network or key is needed in unit tests.
type Judge interface {
	Judge(ctx context.Context, c Case, o Observed) (JudgeResult, error)
}

// JudgeResult is one judged outcome for a case.
type JudgeResult struct {
	// Score is 0-1: how well the observed answer satisfies the case intent.
	Score float64
	// Passed is Score >= the configured floor.
	Passed bool
	// Feedback is a short sentence naming the biggest issue (empty when good).
	Feedback string
}

// ModelJudge is the production Judge: it asks the cheap model to grade an answer
// against the case intent, using the same fail-open contract as the answer
// quality judge (B6) — a broken judge must not fail a case, it is reported as an
// error and the deterministic result stands.
type ModelJudge struct {
	gw     *gateway.ModelGateway
	floor  float64
	logger *zap.Logger
}

// NewModelJudge builds a judge. A floor <= 0 or > 1 falls back to 0.6.
func NewModelJudge(gw *gateway.ModelGateway, floor float64, logger *zap.Logger) *ModelJudge {
	if floor <= 0 || floor > 1 {
		floor = 0.6
	}
	return &ModelJudge{gw: gw, floor: floor, logger: logger}
}

// Judge scores one case. It returns an error (rather than a passing verdict)
// when the model call or parse fails so the caller can record it as a judge
// error without silently inflating or deflating the case result.
func (m *ModelJudge) Judge(ctx context.Context, c Case, o Observed) (JudgeResult, error) {
	if m == nil || m.gw == nil {
		return JudgeResult{}, fmt.Errorf("eval judge: model gateway is not configured")
	}
	if strings.TrimSpace(o.Content) == "" {
		return JudgeResult{Score: 0, Passed: false, Feedback: "answer is empty"}, nil
	}

	systemPrompt := `You are a strict evaluation judge for a domain-expert answer.
Grade the ANSWER against the QUESTION and the EXPECTED INTENT. Do not reward verbosity.
Output ONLY one JSON object, no markdown:
{"score":0.0,"feedback":"..."}
score is 0.0-1.0 where 1.0 means the answer fully satisfies the intent and 0.0 means it does not.
feedback: one short sentence naming the single biggest issue; empty string when the score is high.`

	userPrompt := fmt.Sprintf(
		"QUESTION:\n%s\n\nEXPECTED INTENT:\n%s\n\nANSWER:\n%s",
		c.Input, describeIntent(c), o.Content,
	)

	resp, err := m.gw.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelFast,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    200,
		Temperature:  0.0,
	})
	if err != nil {
		return JudgeResult{}, fmt.Errorf("eval judge: model call: %w", err)
	}
	result, parseErr := parseJudgeResult(resp.Content, m.floor)
	if parseErr != nil {
		return JudgeResult{}, fmt.Errorf("eval judge: parse: %w", parseErr)
	}
	return result, nil
}

// describeIntent renders a case's machine-checkable expectation as prose for the
// judge, so the rubric and the deterministic checks grade the same intent.
func describeIntent(c Case) string {
	var parts []string
	if len(c.Expect.Keywords) > 0 {
		parts = append(parts, "must cover: "+strings.Join(c.Expect.Keywords, ", "))
	}
	if c.Expect.MustCite {
		parts = append(parts, "must cite at least one source")
	}
	if c.Expect.Refuse {
		parts = append(parts, "must refuse (out of scope)")
	}
	if c.Expect.MinChars > 0 {
		parts = append(parts, fmt.Sprintf("at least %d characters", c.Expect.MinChars))
	}
	if len(parts) == 0 {
		return "(no explicit expectation: a correct, on-topic answer)"
	}
	return strings.Join(parts, "; ")
}

type judgeWire struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

// parseJudgeResult reads the judge's JSON, clamps the score to [0,1] and applies
// the floor. Pure — unit-tested.
func parseJudgeResult(raw string, floor float64) (JudgeResult, error) {
	body := strings.TrimSpace(raw)
	// Tolerate a fenced code block even though the prompt forbids it: a model
	// that wraps its JSON must not fail the case for a cosmetic reason.
	if strings.HasPrefix(body, "```") {
		body = strings.TrimPrefix(body, "```json")
		body = strings.TrimPrefix(body, "```")
		body = strings.TrimSuffix(body, "```")
		body = strings.TrimSpace(body)
	}
	if start := strings.IndexByte(body, '{'); start >= 0 {
		if end := strings.LastIndexByte(body, '}'); end > start {
			body = body[start : end+1]
		}
	}
	var w judgeWire
	if err := json.Unmarshal([]byte(body), &w); err != nil {
		return JudgeResult{}, fmt.Errorf("invalid judge JSON: %w", err)
	}
	score := w.Score
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return JudgeResult{
		Score:    score,
		Passed:   score >= floor,
		Feedback: strings.TrimSpace(w.Feedback),
	}, nil
}
