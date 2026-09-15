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

// claudeContentBlock is one element of Claude's content array format.
// Claude thinking/extended models return content as an array of typed
// blocks instead of a plain string:
//   [{"type":"thinking","thinking":"..."}, {"type":"text","text":"..."}]
// We extract only "text" blocks for the final answer.
type claudeContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Thinking string `json:"thinking"`
}

// openAICompatibleResponse is the shared response format used by
// OpenRouter, DeepSeek, Gemini, CodeCraftAPI (all OpenAI-compatible).
//
// content is json.RawMessage because different model families use
// different formats:
//   - Standard models:  "content": "answer text"
//   - Reasoning models: "content": "", "reasoning_content": "answer"
//   - Claude thinking:  "content": [{"type":"text","text":"answer"},...]
// extractMessageContent() handles all three cases.
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
//
// reasoning_content in delta: reasoning models stream thinking tokens here.
// We forward them so the user sees output immediately during thinking phase.
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

// extractMessageContent parses the content field from a message.
// Handles 3 formats:
//   1. Plain string:   "content": "answer"              → standard models
//   2. Empty string:   "content": "", reasoning_content  → DeepSeek-R1, Qwen
//   3. Content array:  "content": [{"type":"text",...}]  → Claude thinking/Opus
//
// Priority: text blocks from array → plain string → reasoning_content.
// WHY: Claude Opus 5 via CodeCraftAPI returns content as a typed array.
// Without this, content parses as empty string → ChinaWall rejects →
// unnecessary retry → 34+ second latency.
func extractMessageContent(raw json.RawMessage, reasoningContent string) string {
	if len(raw) == 0 {
		return reasoningContent
	}

	// Try array format first (Claude thinking models)
	if raw[0] == '[' {
		var blocks []claudeContentBlock
		if err := json.Unmarshal(raw, &blocks); err == nil {
			var sb strings.Builder
			for _, b := range blocks {
				if b.Type == "text" && b.Text != "" {
					sb.WriteString(b.Text)
				}
			}
			if result := strings.TrimSpace(sb.String()); result != "" {
				return result
			}
		}
		// Array parsed but no text blocks — fall through to reasoning_content
		return reasoningContent
	}

	// Try plain string format (standard + reasoning models)
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}

	// content empty → use reasoning_content (DeepSeek-R1, Qwen3.x)
	return reasoningContent
}

// resolveStreamToken returns the best token from a streaming delta.
// Standard models use delta.content; reasoning models use delta.reasoning_content.
func resolveStreamToken(content, reasoningContent string) string {
	if content != "" {
		return content
	}
	return reasoningContent
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

	// extractMessageContent handles all 3 content formats:
	// plain string, reasoning_content fallback, Claude array.
	content := extractMessageContent(
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
// The HTTP request body must already include stream:true.
// Returns tokenCh (individual tokens) and respCh (final metadata).
// Both channels are closed when the stream ends OR ctx is cancelled.
//
// WHY ctx param:
//   Without ctx, the internal goroutine had no way to exit if the provider
//   sent HTTP 200 headers but then stalled the body (no [DONE], no tokens,
//   no RST/FIN). scanner.Scan() blocked forever, tokens channel never
//   closed, callers hung indefinitely.
//   With ctx: select on ctx.Done() in the token-send path ensures the
//   goroutine exits cleanly when the caller's context is cancelled
//   (request disconnect, streaming timeout, orchestrator timeout).
//
// WHY reasoning_content forwarded as tokens:
//   Reasoning models stream thinking tokens in delta.reasoning_content
//   instead of delta.content. Without forwarding these, the user sees
//   nothing for 20-30s while the model thinks, then gets the answer
//   all at once — defeating the purpose of streaming.
func doOpenAICompatibleStream(
	ctx context.Context,
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
			// Check context cancellation on every line.
			// WHY here not just on send: scanner.Scan() itself can block
			// if the provider stalls mid-stream. Checking ctx.Err() after
			// each successful Scan() ensures we exit promptly when the
			// caller's context is cancelled even between token arrivals.
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
				// resolveStreamToken: standard models use delta.content;
				// reasoning models use delta.reasoning_content.
				token := resolveStreamToken(
					chunk.Choices[0].Delta.Content,
					chunk.Choices[0].Delta.ReasoningContent,
				)
				if token != "" {
					fullContent.WriteString(token)
					// Send token or exit if context cancelled.
					// WHY select: without this, tokens <- token blocks
					// if the caller's tokenCh is full AND ctx is done —
					// goroutine would leak instead of exiting.
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
