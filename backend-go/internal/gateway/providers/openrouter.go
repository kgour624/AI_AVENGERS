package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai_avengers/backend/internal/gateway"
)

// OpenRouterProvider implements gateway.LLMProvider for OpenRouter.
// OpenRouter is a multi-model gateway — one API key, many models.
// Model names use namespace/model format: "deepseek/deepseek-chat".
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

// ModelName maps tier → OpenRouter model name.
// WHY hardcoded here: these are OpenRouter-specific names.
// Changing provider doesn't change these — they're internal to this struct.
// Admin can override via system_settings if needed (future).
func (p *OpenRouterProvider) ModelName(tier gateway.ModelType) string {
	switch tier {
	case gateway.ModelStrong:
		return "anthropic/claude-3-5-sonnet" // Best quality for charter extraction
	case gateway.ModelFast:
		return "google/gemini-flash-1.5"
	default: // ModelCheap
		return "deepseek/deepseek-chat"
	}
}

func (p *OpenRouterProvider) CostPer1K(tier gateway.ModelType) (float64, float64) {
	switch tier {
	case gateway.ModelStrong:
		return 0.003, 0.015
	case gateway.ModelFast:
		return 0.00025, 0.00075
	default:
		return 0.0005, 0.0015
	}
}

func (p *OpenRouterProvider) MaxTokens(tier gateway.ModelType) int {
	switch tier {
	case gateway.ModelStrong:
		return 8192
	default:
		return 4096
	}
}

func (p *OpenRouterProvider) Call(ctx context.Context, req gateway.ProviderRequest) (*gateway.ProviderResponse, error) {
	type message struct {
		Role         string      `json:"role"`
		Content      string      `json:"content"`
		CacheControl interface{} `json:"cache_control,omitempty"`
	}
	type cacheCtrl struct {
		Type string `json:"type"`
	}

	var msgs []message
	for i, m := range req.Messages {
		msg := message{Role: m.Role, Content: m.Content}
		// Enable prompt caching on system message (40-60% cost reduction)
		if req.EnableCache && i == 0 && m.Role == "system" {
			msg.CacheControl = cacheCtrl{Type: "ephemeral"}
		}
		msgs = append(msgs, msg)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model":       p.ModelName(req.ModelTier),
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openrouter: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://ai-avengers.app")
	httpReq.Header.Set("X-Title", "AI Avengers")

	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
