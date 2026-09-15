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

// AnthropicProvider implements gtypes.LLMProvider for Anthropic direct API.
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

// StreamCall implements streaming for Anthropic using their SSE format.
// Anthropic's streaming format differs from OpenAI-compatible:
//   data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}
// We parse this directly instead of using doOpenAICompatibleStream.
func (p *AnthropicProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	type anthropicMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type anthropicReq struct {
		Model     string         `json:"model"`
		MaxTokens int            `json:"max_tokens"`
		System    string         `json:"system,omitempty"`
		Messages  []anthropicMsg `json:"messages"`
		Stream    bool           `json:"stream"`
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
		Model: p.ModelName(req.ModelTier), MaxTokens: req.MaxTokens,
		System: systemPrompt, Messages: msgs, Stream: true,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
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

		var fullContent strings.Builder
		var inputTokens, outputTokens int
		usedModel := p.ModelName(req.ModelTier)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
				Message struct {
					Model string `json:"model"`
					Usage struct {
						InputTokens int `json:"input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			switch event.Type {
			case "message_start":
				if event.Message.Model != "" {
					usedModel = event.Message.Model
				}
				inputTokens = event.Message.Usage.InputTokens
			case "content_block_delta":
				if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
					fullContent.WriteString(event.Delta.Text)
					tokens <- event.Delta.Text
				}
			case "message_delta":
				outputTokens = event.Usage.OutputTokens
			case "message_stop":
				// stream complete
			}
		}
		result <- &gtypes.ProviderResponse{
			Content:      fullContent.String(),
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			ModelUsed:    usedModel,
		}
	}()
	return tokens, result, nil
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
		if m.Role == "system" { systemPrompt = m.Content } else {
			msgs = append(msgs, anthropicMsg{Role: m.Role, Content: m.Content})
		}
	}
	body, _ := json.Marshal(anthropicReq{
		Model: p.ModelName(req.ModelTier), MaxTokens: req.MaxTokens,
		System: systemPrompt, Messages: msgs,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("anthropic: http: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: status %d", resp.StatusCode)
	}
	var result struct {
		Content []struct { Type string `json:"type"`; Text string `json:"text"` } `json:"content"`
		Usage   struct { InputTokens int `json:"input_tokens"`; OutputTokens int `json:"output_tokens"` } `json:"usage"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("anthropic: decode: %w", err)
	}
	if len(result.Content) == 0 { return nil, fmt.Errorf("anthropic: empty response") }
	return &gtypes.ProviderResponse{
		Content: result.Content[0].Text,
		InputTokens: result.Usage.InputTokens, OutputTokens: result.Usage.OutputTokens,
		ModelUsed: result.Model,
	}, nil
}
