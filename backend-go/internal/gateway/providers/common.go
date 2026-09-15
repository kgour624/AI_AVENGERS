package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

// openAICompatibleResponse is the shared response format used by
// OpenRouter, DeepSeek, Gemini, and CodeCraftAPI (all OpenAI-compatible).
type openAICompatibleResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
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
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// doOpenAICompatibleCall executes an HTTP request and parses the
// OpenAI-compatible response. Shared by OpenRouter, DeepSeek, Gemini, CodeCraftAPI.
func doOpenAICompatibleCall(client *http.Client, req *http.Request, modelName string) (*gtypes.ProviderResponse, error) {
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

	usedModel := result.Model
	if usedModel == "" {
		usedModel = modelName
	}

	return &gtypes.ProviderResponse{
		Content:      result.Choices[0].Message.Content,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		ModelUsed:    usedModel,
	}, nil
}

// doOpenAICompatibleStream executes a streaming HTTP request.
// The HTTP request body must already include stream:true.
// Returns tokenCh (individual tokens) and respCh (final metadata).
// Both channels are closed when the stream ends.
func doOpenAICompatibleStream(
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
				token := chunk.Choices[0].Delta.Content
				if token != "" {
					fullContent.WriteString(token)
					tokens <- token
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
