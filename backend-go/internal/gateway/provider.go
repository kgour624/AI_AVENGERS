package gateway

import "context"

// LLMProvider is the interface every LLM provider must implement.
//
// WHY interface (composition over inheritance):
//   Each provider (OpenRouter, DeepSeek, Anthropic, Gemini) has:
//   - Different base URLs
//   - Different auth headers
//   - Different model name formats
//   - Different model tiers (cheap/strong/fast map to different models)
//
//   Without this interface, ModelGateway must know all provider details.
//   With this interface, ModelGateway only knows the contract.
//   Adding a new provider = implement this interface, register in factory.
//   Switching provider = swap implementation, zero other changes.
//
// WHEN to add a new provider:
//   1. Create a new file in gateway/providers/
//   2. Implement LLMProvider interface
//   3. Register in ProviderFactory() below
//   That's it. ModelGateway, config, env vars — nothing else changes.
type LLMProvider interface {
	// Name returns the provider identifier ("openrouter", "deepseek", etc.)
	Name() string

	// ModelName returns the provider-specific model name for a given tier.
	// WHY this method exists:
	//   OpenRouter uses "deepseek/deepseek-chat" (namespace/model format)
	//   DeepSeek direct uses "deepseek-chat" (model only)
	//   Anthropic direct uses "claude-3-5-sonnet-20241022"
	//   Each provider knows its own naming — caller just says "cheap" or "strong".
	ModelName(tier ModelType) string

	// CostPer1K returns input and output cost per 1K tokens for a tier.
	// Used for cost tracking. Approximate — actual cost may vary.
	CostPer1K(tier ModelType) (inputCost, outputCost float64)

	// MaxTokens returns the maximum output tokens for a tier.
	MaxTokens(tier ModelType) int

	// Call makes the actual HTTP call to the provider's API.
	// Returns the raw response in a normalized format.
	Call(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
}

// ProviderRequest is the normalized input to any provider.
// Provider implementations translate this to their wire format.
type ProviderRequest struct {
	ModelTier    ModelType
	Messages     []ProviderMessage
	MaxTokens    int
	Temperature  float64
	EnableCache  bool // prompt caching (Anthropic/OpenRouter)
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
