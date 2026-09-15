package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

type AnthropicProvider struct {
	apiKey     string
	httpClient *http.Client
}

func NewAnthropicProvider(apiKey string, client *http.Client) *AnthropicProvider {
	return &AnthropicProvider{apiKey: apiKey, httpClient: client}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }
func (p *AnthropicProvider) ModelName(tier gtypes.ModelType) string {
	if tier == gtypes.ModelStrong { return "claude-3-5-sonnet-20241022" }
	return "claude-3-haiku-20240307"
}
func (p *AnthropicProvider) CostPer1K(tier gtypes.ModelType) (float64, float64) {
	if tier == gtypes.ModelStrong { return 0.003, 0.015 }
	return 0.00025, 0.00125
}
func (p *AnthropicProvider) MaxTokens(tier gtypes.ModelType) int {
	if tier == gtypes.ModelStrong { return 8192 }
	return 4096
}

func (p *AnthropicProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type areq struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		System    string `json:"system,omitempty"`
		Messages  []msg  `json:"messages"`
		Stream    bool   `json:"stream"`
	}
	var sys string
	var msgs []msg
	for _, m := range req.Messages {
		if m.Role == "system" {
			sys = m.Content
		} else {
			msgs = append(msgs, msg{Role: m.Role, Content: m.Content})
		}
	}
	body, _ := json.Marshal(areq{Model: p.ModelName(req.ModelTier), MaxTokens: req.MaxTokens, System: sys, Messages: msgs, Stream: true})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic stream: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic stream: http: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("anthropic stream: status %d", resp.StatusCode)
	}
	tokens := make(chan string, 64)
	result := make(chan *gtypes.ProviderResponse, 1)
	go func() {
		defer resp.Body.Close()
		defer close(tokens)
		defer close(result)
		var sb strings.Builder
		var in, out int
		model := p.ModelName(req.ModelTier)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			var ev struct {
				Type    string `json:"type"`
				Delta   struct{ Type, Text string } `json:"delta"`
				Usage   struct{ InputTokens, OutputTokens int } `json:"usage"`
				Message struct {
					Model string
					Usage struct{ InputTokens int }
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &ev); err != nil {
				continue
			}
			switch ev.Type {
			case "message_start":
				if ev.Message.Model != "" { model = ev.Message.Model }
				in = ev.Message.Usage.InputTokens
			case "content_block_delta":
				if ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
					sb.WriteString(ev.Delta.Text)
					tokens <- ev.Delta.Text
				}
			case "message_delta":
				out = ev.Usage.OutputTokens
			}
		}
		result <- &gtypes.ProviderResponse{Content: sb.String(), InputTokens: in, OutputTokens: out, ModelUsed: model}
	}()
	return tokens, result, nil
}

func (p *AnthropicProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type areq struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		System    string `json:"system,omitempty"`
		Messages  []msg  `json:"messages"`
	}
	var sys string
	var msgs []msg
	for _, m := range req.Messages {
		if m.Role == "system" {
			sys = m.Content
		} else {
			msgs = append(msgs, msg{Role: m.Role, Content: m.Content})
		}
	}
	body, _ := json.Marshal(areq{Model: p.ModelName(req.ModelTier), MaxTokens: req.MaxTokens, System: sys, Messages: msgs})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: status %d", resp.StatusCode)
	}
	var r struct {
		Content []struct{ Type, Text string } `json:"content"`
		Usage   struct{ InputTokens, OutputTokens int } `json:"usage"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("anthropic: decode: %w", err)
	}
	if len(r.Content) == 0 {
		return nil, fmt.Errorf("anthropic: empty response")
	}
	return &gtypes.ProviderResponse{Content: r.Content[0].Text, InputTokens: r.Usage.InputTokens, OutputTokens: r.Usage.OutputTokens, ModelUsed: r.Model}, nil
}
