package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai_avengers/backend/internal/observability"
)

// GeminiKachraDetector — LLM suggests {text,reason,type,confidence}, temp 0.0 strict JSON.
// No-Trust: never mutates content; DSA is final remover.
type GeminiKachraDetector struct{ caller LLMCaller }

func NewGeminiKachraDetector(caller LLMCaller) *GeminiKachraDetector {
	return &GeminiKachraDetector{caller: caller}
}

const kachraDetectSystem = `You are a kachra (noise) detector for course transcripts.
Identify ONLY noise text spans that should be removed.
Return STRICT JSON array only — no markdown, no explanation.

Rules:
- Each item: {"text":"verbatim substring","reason":"why kachra 10-20 words","type":"<one of 8>","confidence":0.0-1.0}
- "text" MUST be exact verbatim substring of input (no paraphrase).
- "reason" short justification e.g. "filler interjection, no information".
- "type" exactly one of: filler | repetition | asr_error | classroom_meta | hinglish | logistics | semantic_noise | smalltalk
- "confidence" 0.0-1.0 (0.70+ high certainty)
- Do NOT rewrite. Do NOT return cleaned text. Only spans.
- Max 20 spans. If no kachra, return [].
- Respond ONLY with JSON array.`

func (g *GeminiKachraDetector) Detect(ctx context.Context, chunkText string) ([]KachraSpan, error) {
	if strings.TrimSpace(chunkText) == "" {
		return nil, nil
	}
	if g.caller == nil {
		return nil, nil
	}
	prompt := truncate(chunkText, 6000)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		spans, err := g.doDetect(ctx, prompt)
		if err == nil {
			observability.Global.IncKachraSuggested(int64(len(spans)))
			return spans, nil
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
	return nil, fmt.Errorf("kachra detect retries exhausted: %w", lastErr)
}

func (g *GeminiKachraDetector) doDetect(ctx context.Context, prompt string) ([]KachraSpan, error) {
	resp, err := g.caller.Call(ctx, LLMRequest{
		Model:        "fast",
		SystemPrompt: kachraDetectSystem,
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
		return nil, fmt.Errorf("kachra detect parse failed: %q err=%w", raw, err)
	}
	return normalizeKachraSpans(spans), nil
}
