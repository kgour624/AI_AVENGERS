package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ai_avengers/backend/internal/gateway"
)

// GeminiProvider implements gateway.LLMProvider for Google Gemini.
// Uses OpenAI-compatible endpoint (generativelanguage.googleapis.com/v1beta/openai).
type GeminiProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewGeminiProvider(apiKey string, client *http.Client) *GeminiProvider {
	return &GeminiProvider{apiKey: apiKey, httpClient: client}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) ModelName(tier gateway.ModelType) string {
	switch tier {
	case gateway.ModelStrong:
		return "gemini-1.5-pro"
	case gateway.ModelFast:
		return "gemini-1.5-flash"
	default:
		return "gemini-1.5-flash"
	}
}

func (p *GeminiProvider) CostPer1K(_ gateway.ModelType) (float64, float64) {
	return 0.00035, 0.00105
}

func (p *GeminiProvider) MaxTokens(_ gateway.ModelType) int { return 8192 }

func (p *GeminiProvider) Call(ctx context.Context, req gateway.ProviderRequest) (*gateway.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model":       p.ModelName(req.ModelTier),
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	})

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions?key=%s",
		p.apiKey,
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gemini: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
