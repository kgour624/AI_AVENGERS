package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ai_avengers/backend/internal/gateway"
)

// DeepSeekProvider implements gateway.LLMProvider for DeepSeek direct API.
// Model names use simple format: "deepseek-chat", "deepseek-reasoner".
// WHY different from OpenRouter: DeepSeek direct API uses its own naming.
type DeepSeekProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewDeepSeekProvider(apiKey string, client *http.Client) *DeepSeekProvider {
	return &DeepSeekProvider{apiKey: apiKey, httpClient: client}
}

func (p *DeepSeekProvider) Name() string { return "deepseek" }

func (p *DeepSeekProvider) ModelName(tier gateway.ModelType) string {
	switch tier {
	case gateway.ModelStrong:
		return "deepseek-chat" // DeepSeek-V3, best available on direct API
	case gateway.ModelFast:
		return "deepseek-chat"
	default: // ModelCheap
		return "deepseek-chat"
	}
}

func (p *DeepSeekProvider) CostPer1K(tier gateway.ModelType) (float64, float64) {
	// DeepSeek pricing (approximate, as of 2026)
	return 0.00014, 0.00028
}

func (p *DeepSeekProvider) MaxTokens(_ gateway.ModelType) int { return 8192 }

func (p *DeepSeekProvider) Call(ctx context.Context, req gateway.ProviderRequest) (*gateway.ProviderResponse, error) {
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepseek: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
