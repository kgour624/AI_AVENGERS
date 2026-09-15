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

// openAICompatibleResponse is the shared response format used by
// OpenRouter, DeepSeek, Gemini, CodeCraftAPI (all OpenAI-compatible).
//
// reasoning_content: populated by reasoning/thinking models
// (DeepSeek-R1, Qwen-QwQ, Claude thinking mode, etc.) when they
// return their chain-of-thought separately from the final answer.
// Some models (e.g. DeepSeek-V4-Flash, Qwen3.x) leave content=""
// and put the actual answer in reasoning_content instead.
// WHY handle both: CodeCraftAPI proxies many model families. A model
// that puts its answer in reasoning_content would otherwise return
// empty content → ChinaWall rejects → unnecessary fallback call.
type openAICompatibleResponse struct {
	Choices []struct {
		Message struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
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
// reasoning_content in delta: same as above — reasoning models stream
// their thinking tokens here. We forward them to the caller the same
// way as regular content tokens so the user sees output immediately
// instead of waiting for the full reasoning phase to complete silently.
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

// resolveContent returns the best available content from a message.
// Priority: content (standard) → reasoning_content (thinking models).
// WHY: reasoning models like DeepSeek-V4-Flash, Qwen3.x, Claude thinking
// mode leave content="" and put the actual answer in reasoning_content.
// Without this fallback, every reasoning model call returns empty content
// → ChinaWall rejects → unnecessary 2nd LLM call → 34+ second latency.
func resolveContent(content, reasoningContent string) string {
	if strings.TrimSpace(content) != "" {
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

	// Use resolveContent: handles both standard models (content field)
	// and reasoning models (reasoning_content field when content is empty).
	content := resolveContent(
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
				// resolveContent per delta: forward whichever field has
				// the token. Standard models use delta.content; reasoning
				// models use delta.reasoning_content. Both are forwarded
				// so the user sees output immediately in either case.
				token := resolveContent(
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
