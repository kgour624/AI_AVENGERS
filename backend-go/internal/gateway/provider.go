package gateway

import gtypes "ai_avengers/backend/internal/gateway/types"

// Re-export types from gateway/types so existing callers
// (model_gateway.go, all internal packages) don't need to change imports.
// providers/*.go import gateway/types directly — no cycle.

type ModelType = gtypes.ModelType

const (
	ModelCheap  = gtypes.ModelCheap
	ModelStrong = gtypes.ModelStrong
	ModelFast   = gtypes.ModelFast
)

type LLMProvider = gtypes.LLMProvider
type ProviderRequest = gtypes.ProviderRequest
type ProviderMessage = gtypes.ProviderMessage
type ProviderResponse = gtypes.ProviderResponse
