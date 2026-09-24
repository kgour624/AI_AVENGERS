package chinawall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// QualityVerdict is the output of the LLM-as-judge (B6, §3.1 P1/P7).
// Pass is decided solely by Overall >= floor (top-level filter) — never
// by averaging Accuracy/Coverage/Structure. Those component scores exist
// for observability and for the REVISION FEEDBACK block on regenerate.
type QualityVerdict struct {
	Overall   float64 // 0-1, the only gate for Pass
	Accuracy  float64 // 0-1
	Coverage  float64 // 0-1
	Structure float64 // 0-1
	Feedback  string  // short free-text: what to improve (used on regen)
	Pass      bool    // Overall >= floor
	// Method: "judge" on a successful LLM call, "fail_open" when the
	// judge itself failed and the original answer is kept as-is.
	Method string
}

// qualityJudgeMaxChunks caps how much reference material is sent to the
// judge. Full chunk dump is wasteful and can push the judge into the
// generator's own length regime; 5 strongest chunks is enough grounding.
const qualityJudgeMaxChunks = 5

// judgeAnswer runs the LLM-as-judge rubric against a cleaned flat-path
// answer. Reference grounding is the original question + the top-N
// retrieved chunks (no golden answer exists in this flow). Uses
// ModelFast (divergent from the ModelStrong generator) so the judge is
// not the same model marking its own homework.
//
// Failure policy (fail-open): any LLM/parse error returns a Pass=true
// verdict with Method="fail_open". Quality is an enhancement on top of
// the citation wall; a broken judge must never refuse a cited answer.
func (e *Enforcer) judgeAnswer(
	ctx context.Context,
	question string,
	answer string,
	chunks []CourseChunk,
	floor float64,
) QualityVerdict {
	// Overall=0 + Pass=true: ship the answer (fail-open) but leave
	// QualityScore at 0 so observability does not pretend the judge ran.
	failOpen := QualityVerdict{
		Overall:  0,
		Pass:     true,
		Method:   "fail_open",
		Feedback: "judge unavailable — kept original answer",
	}
	if e == nil || e.gateway == nil {
		return failOpen
	}
	if strings.TrimSpace(answer) == "" {
		// Empty answers never reach here in the flat success path, but
		// if they did the wall already refused. Fail-open would pass an
		// empty answer — so fail-closed only for the empty case.
		return QualityVerdict{
			Overall:  0,
			Pass:     false,
			Method:   "empty",
			Feedback: "answer is empty",
		}
	}

	ref := buildJudgeReference(chunks)
	systemPrompt := `You are a strict answer-quality judge for a domain-expert system.
Score the ANSWER against the QUESTION using ONLY the REFERENCE material (the expert's training chunks). Do not use outside knowledge.

Rubric (each 0.0 to 1.0):
- accuracy: factual claims match the reference; no invented details
- coverage: the answer addresses the core of the question
- structure: clear, usable structure (explanation / steps / code as needed)
- overall: your final verdict on whether this answer is good enough to ship

Rules:
1. overall is an independent top-level judgement — do NOT average the other three.
2. Be strict: a partial or hand-wavy answer should score overall below 0.7.
3. feedback: one short sentence naming the single biggest fix if overall < 1.0; empty string if overall is high.
4. Output ONLY one JSON object, no markdown, no preamble:
{"accuracy":0.0,"coverage":0.0,"structure":0.0,"overall":0.0,"feedback":"..."}`

	userPrompt := fmt.Sprintf(
		"QUESTION:\n%s\n\nREFERENCE (expert training chunks):\n%s\n\nANSWER:\n%s",
		question, ref, answer,
	)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelFast, // divergent from ModelStrong generator
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    200,
		Temperature:  0.0,
	})
	if err != nil {
		e.logger.Warn("quality judge call failed — fail-open", zap.Error(err))
		return failOpen
	}

	v, parseErr := parseQualityVerdict(resp.Content, floor)
	if parseErr != nil {
		e.logger.Warn("quality judge parse failed — fail-open",
			zap.Error(parseErr),
			zap.String("raw", truncateForLog(resp.Content, 200)),
		)
		return failOpen
	}
	v.Method = "judge"
	return v
}

// parseQualityVerdict extracts the JSON object from the judge response
// (tolerates accidental markdown fences) and decides Pass from Overall
// alone. Pure function — unit-tested.
func parseQualityVerdict(raw string, floor float64) (QualityVerdict, error) {
	raw = strings.TrimSpace(raw)
	// Strip optional ```json ... ``` fence.
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```JSON")
		raw = strings.TrimPrefix(raw, "```")
		if i := strings.LastIndex(raw, "```"); i >= 0 {
			raw = raw[:i]
		}
		raw = strings.TrimSpace(raw)
	}
	// Take the first {...} object if there is trailing prose.
	if start := strings.Index(raw, "{"); start >= 0 {
		if end := strings.LastIndex(raw, "}"); end > start {
			raw = raw[start : end+1]
		}
	}

	var parsed struct {
		Accuracy  float64 `json:"accuracy"`
		Coverage  float64 `json:"coverage"`
		Structure float64 `json:"structure"`
		Overall   float64 `json:"overall"`
		Feedback  string  `json:"feedback"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return QualityVerdict{}, fmt.Errorf("unmarshal: %w", err)
	}
	// Clamp components into [0,1] so a misbehaving model cannot pass with
	// overall=5 or poison logs with NaNs.
	v := QualityVerdict{
		Accuracy:  clamp01(parsed.Accuracy),
		Coverage:  clamp01(parsed.Coverage),
		Structure: clamp01(parsed.Structure),
		Overall:   clamp01(parsed.Overall),
		Feedback:  strings.TrimSpace(parsed.Feedback),
	}
	if floor <= 0 || floor > 1 {
		floor = 0.7
	}
	v.Pass = v.Overall >= floor
	return v, nil
}

// buildJudgeReference concatenates the strongest chunks (already sorted
// best-first by the retriever) into a compact block for the judge.
func buildJudgeReference(chunks []CourseChunk) string {
	if len(chunks) == 0 {
		return "(no reference chunks)"
	}
	n := len(chunks)
	if n > qualityJudgeMaxChunks {
		n = qualityJudgeMaxChunks
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := chunks[i]
		sb.WriteString(fmt.Sprintf("[REF %d score=%.2f topic=%s]\n%s\n\n",
			i+1, c.RerankScore, c.Topic, truncateForLog(c.Text, 600)))
	}
	return sb.String()
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func truncateForLog(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
