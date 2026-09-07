package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

// OpenRouterProvider implements gtypes.LLMProvider for OpenRouter.
type OpenRouterProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewOpenRouterProvider(apiKey, baseURL string, client *http.Client) *OpenRouterProvider {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &OpenRouterProvider{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/"), httpClient: client}
}

func (p *OpenRouterProvider) Name() string { return "openrouter" }

func (p *OpenRouterProvider) ModelName(tier gtypes.ModelType) string {
	switch tier {
	case gtypes.ModelStrong:
		return "anthropic/claude-3-5-sonnet"
	case gtypes.ModelFast:
		return "google/gemini-flash-1.5"
	default:
		return "deepseek/deepseek-chat"
	}
}

func (p *OpenRouterProvider) CostPer1K(tier gtypes.ModelType) (float64, float64) {
	switch tier {
	case gtypes.ModelStrong:
		return 0.003, 0.015
	case gtypes.ModelFast:
		return 0.00025, 0.00075
	default:
		return 0.0005, 0.0015
	}
}

func (p *OpenRouterProvider) MaxTokens(tier gtypes.ModelType) int {
	if tier == gtypes.ModelStrong { return 8192 }
	return 4096
}

func (p *OpenRouterProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
	type cacheCtrl struct{ Type string `json:"type"` }
	type message struct {
		Role         string      `json:"role"`
		Content      string      `json:"content"`
		CacheControl interface{} `json:"cache_control,omitempty"`
	}
	var msgs []message
	for i, m := range req.Messages {
		msg := message{Role: m.Role, Content: m.Content}
		if req.EnableCache && i == 0 && m.Role == "system" {
			msg.CacheControl = cacheCtrl{Type: "ephemeral"}
		}
		msgs = append(msgs, msg)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.ModelName(req.ModelTier), "messages": msgs,
		"max_tokens": req.MaxTokens, "temperature": req.Temperature,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openrouter: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://ai-avengers.app")
	httpReq.Header.Set("X-Title", "AI Avengers")
	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
