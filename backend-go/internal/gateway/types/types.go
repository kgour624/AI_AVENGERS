// Package types holds the shared contracts for the gateway layer.
// WHY a separate package:
//   gateway/model_gateway.go needs to import providers (to build them).
//   providers/*.go need to implement LLMProvider (defined here).
//   If both lived in gateway, providers would import gateway = cycle.
//   Neutral types package breaks the cycle:
//     gateway  → gateway/types  (no cycle)
//     providers → gateway/types  (no cycle)
//     gateway  → providers       (one-way, for factory only)
package types

import (
	"context"
	"encoding/json"
)

// ModelType identifies which LLM tier to use.
type ModelType string

const (
	ModelCheap  ModelType = "cheap"
	ModelStrong ModelType = "strong"
	ModelFast   ModelType = "fast"
)

// LLMProvider is the interface every LLM provider must implement.
//
// WHEN to add a new provider:
//   1. Create providers/myprovider.go
//   2. Implement this interface
//   3. Register in gateway.buildProvider() switch
//   Nothing else changes — ModelGateway, config, env vars untouched.
//
// CONTENT EXTRACTION (ExtractContent / ExtractStreamToken):
//   CodeCraftAPI proxies many model families but does NOT normalize
//   the inner payload — it passes through each model's native format.
//   Claude thinking models return content as a typed array.
//   DeepSeek reasoning models return an extra reasoning_content field.
//   Future models may use yet another format.
//
//   WHY on the interface (not in common.go):
//     Putting model-specific format logic in common.go creates tight
//     coupling — every new model family requires a code change there.
//     Each provider knows its own wire format; the interface lets each
//     provider own that knowledge. common.go becomes format-agnostic:
//     it calls ExtractContent() and never inspects the raw bytes itself.
//
//   Default implementations (StandardExtractContent /
//   StandardExtractStreamToken) handle the plain-string + reasoning_content
//   case that covers most OpenAI-compatible models. Providers that need
//   different behavior (e.g. Claude array format) override only those two
//   methods — everything else (HTTP, retry, cost tracking) is unchanged.
type LLMProvider interface {
	// Name returns the provider identifier ("openrouter", "deepseek", etc.)
	Name() string

	// ModelName returns the provider-specific model name for a tier.
	ModelName(tier ModelType) string

	// CostPer1K returns input and output cost per 1K tokens.
	CostPer1K(tier ModelType) (inputCost, outputCost float64)

	// MaxTokens returns the maximum output tokens for a tier.
	MaxTokens(tier ModelType) int

	// Call makes the actual HTTP call to the provider's API.
	Call(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)

	// StreamCall makes a streaming HTTP call to the provider's API.
	StreamCall(ctx context.Context, req ProviderRequest) (tokenCh <-chan string, respCh <-chan *ProviderResponse, err error)

	// ExtractContent extracts the final answer text from a message's
	// raw content field and optional reasoningContent field.
	//
	// raw is json.RawMessage — the exact bytes of choices[0].message.content
	// as received from the API. It may be:
	//   - a JSON string:  "answer text"
	//   - a JSON array:   [{"type":"text","text":"..."},...]
	//   - JSON null / empty
	//
	// reasoningContent is choices[0].message.reasoning_content (may be "").
	//
	// The provider returns the best available answer string, or "" if none.
	ExtractContent(raw json.RawMessage, reasoningContent string) string

	// ExtractStreamToken extracts the token text from one streaming delta.
	//
	// content is delta.content (standard field).
	// reasoningContent is delta.reasoning_content (reasoning models).
	//
	// The provider returns whichever field carries the token for this
	// model family, or "" if neither has content.
	ExtractStreamToken(content, reasoningContent string) string
}

// ProviderRequest is the normalized input to any provider.
type ProviderRequest struct {
	ModelTier   ModelType
	Messages    []ProviderMessage
	MaxTokens   int
	Temperature float64
	EnableCache bool
}

// ProviderMessage is a single message in the conversation.
type ProviderMessage struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// ProviderResponse is the normalized output from any provider.
type ProviderResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	ModelUsed    string
}

// StandardExtractContent is the default implementation for providers
// that use the standard OpenAI format (plain string content field).
// Also handles the reasoning_content fallback for DeepSeek-style models.
//
// Providers that need different behavior (e.g. Claude array format)
// should NOT embed this — they implement ExtractContent directly.
func StandardExtractContent(raw json.RawMessage, reasoningContent string) string {
	if len(raw) == 0 {
		return reasoningContent
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return s
	}
	return reasoningContent
}

// StandardExtractStreamToken is the default implementation for providers
// that use delta.content for streaming tokens.
func StandardExtractStreamToken(content, reasoningContent string) string {
	if content != "" {
		return content
	}
	return reasoningContent
}
