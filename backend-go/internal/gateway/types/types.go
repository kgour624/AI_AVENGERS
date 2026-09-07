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

import "context"

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
type LLMProvider interface {
	// Name returns the provider identifier ("openrouter", "deepseek", etc.)
	Name() string

	// ModelName returns the provider-specific model name for a tier.
	// Each provider knows its own naming convention internally.
	ModelName(tier ModelType) string

	// CostPer1K returns input and output cost per 1K tokens.
	CostPer1K(tier ModelType) (inputCost, outputCost float64)

	// MaxTokens returns the maximum output tokens for a tier.
	MaxTokens(tier ModelType) int

	// Call makes the actual HTTP call to the provider's API.
	Call(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
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
