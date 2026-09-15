package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

type GeminiProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewGeminiProvider(apiKey string, client *http.Client) *GeminiProvider {
	return &GeminiProvider{apiKey: apiKey, httpClient: client}
}

func (p *GeminiProvider) Name() string { return "gemini" }
func (p *GeminiProvider) ModelName(tier gtypes.ModelType) string {
	if tier == gtypes.ModelStrong { return "gemini-1.5-pro" }
	return "gemini-1.5-flash"
}
func (p *GeminiProvider) CostPer1K(_ gtypes.ModelType) (float64, float64) { return 0.00035, 0.00105 }
func (p *GeminiProvider) MaxTokens(_ gtypes.ModelType) int                 { return 8192 }

func (p *GeminiProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.ModelName(req.ModelTier), "messages": msgs,
		"max_tokens": req.MaxTokens, "temperature": req.Temperature,
	})
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions?key=%s",
		p.apiKey,
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil { return nil, fmt.Errorf("gemini: %w", err) }
	httpReq.Header.Set("Content-Type", "application/json")
	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

func (p *GeminiProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": p.ModelName(req.ModelTier), "messages": msgs,
		"max_tokens": req.MaxTokens, "temperature": req.Temperature,
		"stream": true,
	})
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions?key=%s",
		p.apiKey,
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("gemini stream: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return doOpenAICompatibleStream(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

func (p *GeminiProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
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
		"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("gemini stream: %w", err)
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
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions?key=%s",
		p.apiKey,
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil { return nil, fmt.Errorf("gemini: %w", err) }
	httpReq.Header.Set("Content-Type", "application/json")
	return doOpenAICompatibleCall(p.httpClient, httpReq, p.ModelName(req.ModelTier))
}
