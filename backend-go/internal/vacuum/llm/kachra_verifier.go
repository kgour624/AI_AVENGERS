package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai_avengers/backend/internal/observability"
)

// GeminiKachraVerifier — second LLM call that drops confidence < 0.70 and non-verbatim spans.
type GeminiKachraVerifier struct{ caller LLMCaller }

func NewGeminiKachraVerifier(caller LLMCaller) *GeminiKachraVerifier {
	return &GeminiKachraVerifier{caller: caller}
}

const kachraVerifySystem = `You are a strict verifier for kachra (noise) spans.
Given the original chunk and candidate spans, KEEP ONLY spans that are truly noise.
Return STRICT JSON array of kept spans with same keys: {"text","reason","type","confidence"}.
Rules:
- If span text is NOT verbatim in the chunk, drop it.
- If confidence < 0.70, drop it.
- If span is meaningful content (code, domain term, explanation), drop it.
- Keep "type" from input if valid; else correct to one of: filler | repetition | asr_error | classroom_meta | hinglish | logistics | semantic_noise | smalltalk
- Max same count as input. If none verified, return [].
- Respond ONLY with JSON array.`

var validKachraType = map[string]bool{
	"filler": true, "repetition": true, "asr_error": true, "classroom_meta": true,
	"hinglish": true, "logistics": true, "semantic_noise": true, "smalltalk": true,
}

func (g *GeminiKachraVerifier) Verify(ctx context.Context, chunkText string, spans []KachraSpan) ([]KachraSpan, error) {
	if len(spans) == 0 || strings.TrimSpace(chunkText) == "" {
		return spans, nil
	}
	if g.caller == nil {
		filtered := make([]KachraSpan, 0, len(spans))
		for _, s := range spans {
			if s.Confidence >= 0.70 && strings.TrimSpace(s.Reason) != "" && validKachraType[s.Type] {
				filtered = append(filtered, s)
			}
		}
		return filtered, nil
	}
	payload := buildVerifyPrompt(chunkText, spans)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		out, err := g.doVerify(ctx, payload)
		if err == nil {
			observability.Global.IncKachraVerified(int64(len(out)))
			return out, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !isTransient(err) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(200*(attempt+1)) * time.Millisecond):
		}
	}
	return nil, fmt.Errorf("kachra verify retries exhausted: %w", lastErr)
}

func (g *GeminiKachraVerifier) doVerify(ctx context.Context, prompt string) ([]KachraSpan, error) {
	resp, err := g.caller.Call(ctx, LLMRequest{
		Model:        "fast",
		SystemPrompt: kachraVerifySystem,
		UserPrompt:   prompt,
		MaxTokens:    8100,
		Temperature:  0.0,
	})
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(resp.Content)
	raw = extractJSONArr(raw)
	var spans []KachraSpan
	if err := json.Unmarshal([]byte(raw), &spans); err != nil {
		return nil, fmt.Errorf("kachra verify parse failed: %q err=%w", raw, err)
	}
	out := normalizeKachraSpans(spans)
	filtered := make([]KachraSpan, 0, len(out))
	for _, s := range out {
		if s.Confidence >= 0.70 {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

func buildVerifyPrompt(chunkText string, spans []KachraSpan) string {
	var b strings.Builder
	b.WriteString("CHUNK:\n")
	b.WriteString(truncate(chunkText, 4000))
	b.WriteString("\n\nCANDIDATE_SPANS_JSON:\n")
	j, _ := json.Marshal(spans)
	b.Write(j)
	return b.String()
}

func normalizeKachraSpans(in []KachraSpan) []KachraSpan {
	if len(in) == 0 {
		return nil
	}
	out := make([]KachraSpan, 0, len(in))
	seen := map[string]bool{}
	for _, s := range in {
		s.Text = strings.TrimSpace(s.Text)
		s.Reason = strings.TrimSpace(s.Reason)
		s.Type = strings.ToLower(strings.TrimSpace(s.Type))
		if s.Text == "" || s.Reason == "" {
			continue
		}
		if !validKachraType[s.Type] {
			s.Type = "filler"
		}
		if s.Confidence < 0 {
			s.Confidence = 0
		}
		if s.Confidence > 1 {
			s.Confidence = 1
		}
		key := strings.ToLower(s.Text) + "|" + s.Type
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
		if len(out) >= 20 {
			break
		}
	}
	return out
}
