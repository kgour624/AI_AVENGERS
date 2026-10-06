package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ClaudeHeadingGenerator — Sonnet heading generator: ## and ### only.
// No-Trust: returns ONLY headings, never rewrites content.
type ClaudeHeadingGenerator struct{ caller LLMCaller }

func NewClaudeHeadingGenerator(caller LLMCaller) *ClaudeHeadingGenerator {
	return &ClaudeHeadingGenerator{caller: caller}
}

const headingSystem = `You are a heading generator for cleaned transcripts.
Return ONLY headings for the text. Rules:
- Use ONLY "## " (level 2) or "### " (level 3) prefixes.
- Do NOT use "# " (h1) or "####" or deeper.
- Do NOT rewrite or summarize the source text.
- Return JSON array of strings, each string is the full heading line e.g. "## Introduction" or "### Background".
- Max 12 headings. If no clear sections, return [].
Respond ONLY with JSON array. No other text.`

func (c *ClaudeHeadingGenerator) Generate(ctx context.Context, text string) ([]Heading, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	if c.caller == nil {
		return nil, nil
	}
	prompt := truncate(text, 8000)
	resp, err := c.caller.Call(ctx, LLMRequest{
		Model:        "strong",
		SystemPrompt: headingSystem,
		UserPrompt:   prompt,
		MaxTokens:    1024,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(resp.Content)
	raw = extractJSONArr(raw)
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		// Lenient: try line-split fallback
		lines := strings.Split(resp.Content, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "### ") {
				arr = append(arr, l)
			}
		}
		if len(arr) == 0 {
			return nil, fmt.Errorf("heading parse failed: %q err=%w", raw, err)
		}
	}
	headings := make([]Heading, 0, len(arr))
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "### ") {
			headings = append(headings, Heading{Level: 3, Text: strings.TrimSpace(strings.TrimPrefix(s, "### ")), Raw: s})
		} else if strings.HasPrefix(s, "## ") {
			headings = append(headings, Heading{Level: 2, Text: strings.TrimSpace(strings.TrimPrefix(s, "## ")), Raw: s})
		} else {
			// Reject invalid heading — skip instead of fail
			continue
		}
	}
	// Validate harness: if LLM returned bad headings (h1 or #### etc), they were skipped.
	// If none remain, return nil (no headings) not error — pipeline continues.
	if err := c.Validate(headings); err != nil {
		return nil, err
	}
	return headings, nil
}

func (c *ClaudeHeadingGenerator) Validate(headings []Heading) error {
	for _, h := range headings {
		if h.Level != 2 && h.Level != 3 {
			return fmt.Errorf("invalid heading level %d: %q", h.Level, h.Raw)
		}
		if strings.HasPrefix(h.Raw, "# ") || strings.HasPrefix(h.Raw, "####") {
			return fmt.Errorf("heading must be ## or ### only: %q", h.Raw)
		}
		if strings.TrimSpace(h.Text) == "" {
			return fmt.Errorf("empty heading text: %q", h.Raw)
		}
	}
	return nil
}

func extractJSONArr(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		start := strings.Index(s, "[")
		end := strings.LastIndex(s, "]")
		if start >= 0 && end > start {
			return s[start : end+1]
		}
	}
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
