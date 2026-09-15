package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

type DeepSeekProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewDeepSeekProvider(apiKey string, client *http.Client) *DeepSeekProvider {
	return &DeepSeekProvider{apiKey: apiKey, httpClient: client}
}

func (p *DeepSeekProvider) Name() string                                    { return "deepseek" }
func (p *DeepSeekProvider) ModelName(_ gtypes.ModelType) string             { return "deepseek-chat" }
func (p *DeepSeekProvider) CostPer1K(_ gtypes.ModelType) (float64, float64) { return 0.00014, 0.00028 }
func (p *DeepSeekProvider) MaxTokens(_ gtypes.ModelType) int                { return 8192 }

func (p *DeepSeekProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.ModelName(req.ModelTier), "messages": msgs,
		"max_tokens": req.MaxTokens, "temperature": req.Temperature,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepseek: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

func (p *DeepSeekProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.ModelName(req.ModelTier), "messages": msgs,
		"max_tokens": req.MaxTokens, "temperature": req.Temperature,
		"stream": true,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("deepseek stream: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleStream(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}}

func (p *DeepSeekProvider) Name() string { return "deepseek" }

func (p *DeepSeekProvider) ModelName(_ gtypes.ModelType) string { return "deepseek-chat" }

func (p *DeepSeekProvider) CostPer1K(_ gtypes.ModelType) (float64, float64) {
	return 0.00014, 0.00028
}

func (p *DeepSeekProvider) MaxTokens(_ gtypes.ModelType) int { return 8192 }

func (p *DeepSeekProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":       p.ModelName(req.ModelTier),
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"stream":      true,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("deepseek stream: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleStream(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.ModelName(req.ModelTier), "messages": msgs,
		"max_tokens": req.MaxTokens, "temperature": req.Temperature,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepseek: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
