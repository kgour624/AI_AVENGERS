package providers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ai_avengers/backend/internal/gateway"
)

// openAICompatibleResponse is the shared response format used by
// OpenRouter, DeepSeek, and Gemini (all use OpenAI-compatible API).
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

// doOpenAICompatibleCall executes an HTTP request and parses the
// OpenAI-compatible response format. Shared by OpenRouter, DeepSeek, Gemini.
func doOpenAICompatibleCall(client *http.Client, req *http.Request, modelName string) (*gateway.ProviderResponse, error) {
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

	return &gateway.ProviderResponse{
		Content:      result.Choices[0].Message.Content,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		ModelUsed:    usedModel,
	}, nil
}
