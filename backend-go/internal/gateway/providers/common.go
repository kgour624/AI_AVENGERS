package providers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

// openAICompatibleResponse is the shared response struct for all
// OpenAI-compatible providers (OpenRouter, DeepSeek, Gemini, CodeCraftAPI).
//
// content is json.RawMessage because CodeCraftAPI does NOT normalize
// the inner payload — it passes through each model's native wire format:
//   - Standard models:  "content": "answer text"          (JSON string)
//   - Claude thinking:  "content": [{"type":"text",...}]  (JSON array)
//   - DeepSeek R1:      "content": "", "reasoning_content": "answer"
//
// Each provider implements LLMProvider.ExtractContent() to handle its
// own format. common.go never inspects the raw bytes directly.
type openAICompatibleResponse struct {
	Choices []struct {
		Message struct {
			Content          json.RawMessage `json:"content"`
			ReasoningContent string          `json:"reasoning_content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// openAIStreamChunk is one SSE data line from an OpenAI-compatible stream.
type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// doOpenAICompatibleCall executes a blocking HTTP request and returns
// a normalized ProviderResponse.
//
// provider.ExtractContent() is called to extract the answer — each
// provider owns its own format logic. common.go is format-agnostic.
func doOpenAICompatibleCall(
	provider gtypes.LLMProvider,
	client *http.Client,
	req *http.Request,
	modelName string,
) (*gtypes.ProviderResponse, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var result openAICompatibleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty choices in response")
	}

	// Delegate format-specific extraction to the provider.
	// No model-specific logic here — common.go stays format-agnostic.
	content := provider.ExtractContent(
		result.Choices[0].Message.Content,
		result.Choices[0].Message.ReasoningContent,
	)
	if content == "" {
		return nil, fmt.Errorf("empty content in response")
	}

	usedModel := result.Model
	if usedModel == "" {
		usedModel = modelName
	}

	return &gtypes.ProviderResponse{
		Content:      content,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		ModelUsed:    usedModel,
	}, nil
}

// doOpenAICompatibleStream executes a streaming HTTP request.
// Returns tokenCh (individual tokens) and respCh (final metadata).
// Both channels are closed when the stream ends OR ctx is cancelled.
//
// provider.ExtractStreamToken() is called per delta — each provider
// owns its own streaming token extraction logic.
func doOpenAICompatibleStream(
	ctx context.Context,
	provider gtypes.LLMProvider,
	client *http.Client,
	req *http.Request,
	modelName string,
) (tokenCh <-chan string, respCh <-chan *gtypes.ProviderResponse, err error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http stream call failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	tokens := make(chan string, 64)
	result := make(chan *gtypes.ProviderResponse, 1)

	go func() {
		defer resp.Body.Close()
		defer close(tokens)
		defer close(result)

		var fullContent strings.Builder
		var inputTokens, outputTokens int
		usedModel := modelName

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if chunk.Model != "" {
				usedModel = chunk.Model
			}
			if chunk.Usage != nil {
				inputTokens = chunk.Usage.PromptTokens
				outputTokens = chunk.Usage.CompletionTokens
			}
			if len(chunk.Choices) > 0 {
				// Delegate token extraction to the provider.
				token := provider.ExtractStreamToken(
					chunk.Choices[0].Delta.Content,
					chunk.Choices[0].Delta.ReasoningContent,
				)
				if token != "" {
					fullContent.WriteString(token)
					select {
					case tokens <- token:
					case <-ctx.Done():
						return
					}
				}
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
