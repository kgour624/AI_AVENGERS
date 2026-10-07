package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GeminiClassifier — Flash temp 0.1 classifier for chunks.
// ModelTier "fast" maps to Gemini Flash via gateway. Retries on transient errors.
type GeminiClassifier struct {
	caller LLMCaller
}

func NewGeminiClassifier(caller LLMCaller) *GeminiClassifier {
	return &GeminiClassifier{caller: caller}
}

const classifierSystem = `You are a tiny text classifier for transcript chunks.
Classify the chunk into exactly one label: "filler" | "content" | "music" | "noise" | "code".
Respond ONLY with JSON: {"label":"<one of the 5>","confidence":0.0-1.0}
No other text.`

var validLabels = map[string]bool{
	"filler": true, "content": true, "music": true, "noise": true, "code": true,
}

type classifierJSON struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

func (g *GeminiClassifier) Classify(ctx context.Context, text string) (string, float64, error) {
	if strings.TrimSpace(text) == "" {
		return "content", 1.0, nil
	}
	// Pass-through when no LLM wired (dev / test without creds) — deterministic fallback.
	if g.caller == nil {
		return heuristicLabel(text), 0.55, nil
	}
	// Retry 3x with backoff for transient LLM errors
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		label, conf, err := g.doClassify(ctx, text)
		if err == nil {
			return label, conf, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return "", 0, ctx.Err()
		}
		// Only retry on transient (rate limit / 5xx style messages)
		if !isTransient(err) {
			return "", 0, err
		}
		select {
		case <-ctx.Done():
			return "", 0, ctx.Err()
		case <-time.After(time.Duration(200*(attempt+1)) * time.Millisecond):
		}
	}
	// After retries fall back to heuristic so pipeline does not fail open
	return heuristicLabel(text), 0.5, fmt.Errorf("classifier fallback after retries: %w", lastErr)
}

func (g *GeminiClassifier) doClassify(ctx context.Context, text string) (string, float64, error) {
	// Truncate to first 2k chars to control cost
	prompt := truncate(text, 2000)
	resp, err := g.caller.Call(ctx, LLMRequest{
		Model:        "fast",
		SystemPrompt: classifierSystem,
		UserPrompt:   prompt,
		MaxTokens:    8100,
		Temperature:  0.1,
	})
	if err != nil {
		return "", 0, err
	}
	raw := strings.TrimSpace(resp.Content)
	// Extract JSON object if wrapped in markdown
	raw = extractJSON(raw)
	var j classifierJSON
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		// Fallback: try to parse label directly
		lower := strings.ToLower(strings.TrimSpace(raw))
		if validLabels[lower] {
			return lower, 0.6, nil
		}
		return "", 0, fmt.Errorf("classifier parse failed: %q err=%w", raw, err)
	}
	j.Label = strings.ToLower(strings.TrimSpace(j.Label))
	if !validLabels[j.Label] {
		return "", 0, fmt.Errorf("invalid label %q", j.Label)
	}
	if j.Confidence < 0 {
		j.Confidence = 0
	}
	if j.Confidence > 1 {
		j.Confidence = 1
	}
	return j.Label, j.Confidence, nil
}

func heuristicLabel(text string) string {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "[music]") || strings.Contains(lower, "[applause]") {
		return "music"
	}
	if strings.Contains(lower, "\n```") || strings.Contains(lower, "func ") {
		return "code"
	}
	// Very short repeated filler
	if len(strings.Fields(text)) < 4 {
		return "filler"
	}
	return "content"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// strip ```json ... ``` fences
	if strings.HasPrefix(s, "```") {
		// find first { and last }
		start := strings.Index(s, "{")
		end := strings.LastIndex(s, "}")
		if start >= 0 && end > start {
			return s[start : end+1]
		}
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func isTransient(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") || strings.Contains(msg, "rate") ||
		strings.Contains(msg, "503") || strings.Contains(msg, "502") ||
		strings.Contains(msg, "504") || strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "overload")
}
