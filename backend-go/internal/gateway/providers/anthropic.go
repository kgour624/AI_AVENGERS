package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ai_avengers/backend/internal/gateway"
)

// AnthropicProvider implements gateway.LLMProvider for Anthropic direct API.
type AnthropicProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewAnthropicProvider(apiKey string, client *http.Client) *AnthropicProvider {
	return &AnthropicProvider{apiKey: apiKey, httpClient: client}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) ModelName(tier gateway.ModelType) string {
	switch tier {
	case gateway.ModelStrong:
		return "claude-3-5-sonnet-20241022"
	case gateway.ModelFast:
		return "claude-3-haiku-20240307"
	default:
		return "claude-3-haiku-20240307"
	}
}

func (p *AnthropicProvider) CostPer1K(tier gateway.ModelType) (float64, float64) {
	switch tier {
	case gateway.ModelStrong:
		return 0.003, 0.015
	default:
		return 0.00025, 0.00125
	}
}

func (p *AnthropicProvider) MaxTokens(tier gateway.ModelType) int {
	switch tier {
	case gateway.ModelStrong:
		return 8192
	default:
		return 4096
	}
}

func (p *AnthropicProvider) Call(ctx context.Context, req gateway.ProviderRequest) (*gateway.ProviderResponse, error) {
	// Anthropic uses a different API format (not OpenAI-compatible)
	type contentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type anthropicMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type anthropicReq struct {
		Model     string         `json:"model"`
		MaxTokens int            `json:"max_tokens"`
		System    string         `json:"system,omitempty"`
		Messages  []anthropicMsg `json:"messages"`
	}

	var systemPrompt string
	var msgs []anthropicMsg
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemPrompt = m.Content
		} else {
			msgs = append(msgs, anthropicMsg{Role: m.Role, Content: m.Content})
		}
	}

	body, _ := json.Marshal(anthropicReq{
		Model:     p.ModelName(req.ModelTier),
		MaxTokens: req.MaxTokens,
		System:    systemPrompt,
		Messages:  msgs,
	})

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Anthropic response format is different
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: http call: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: status %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("anthropic: decode: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("anthropic: empty response")
	}
	return &gateway.ProviderResponse{
		Content:      result.Content[0].Text,
		InputTokens:  result.Usage.InputTokens,
		OutputTokens: result.Usage.OutputTokens,
		ModelUsed:    result.Model,
	}, nil
}
