package providers

import (
	"context"
	"net/http"
	"strings"

	gtypes "ai_avengers/backend/internal/gateway/types"
)

// CavotiProvider implements gtypes.LLMProvider for Cavoti.
//
// Cavoti is an OpenAI-compatible LLM provider.
// Wire format is identical to OpenRouter/CodeCraftAPI/DeepSeek.
// doOpenAICompatibleCall() from common.go is reused directly.
//
// WHY model names come from constructor (not hardcoded):
//   Admin selects which Cavoti model to use per tier from the admin panel.
//   Same pattern as CodeCraftAPI.
//
// WHY CostPer1K returns 0, 0:
//   Cavoti pricing is checked from the Cavoti dashboard.
//   Returning fabricated numbers would corrupt cost tracking.
type CavotiProvider struct {
	apiKey      string
	baseURL     string
	modelCheap  string
	modelStrong string
	modelFast   string
	httpClient  *http.Client
}

// NewCavotiProvider creates a new CavotiProvider.
func NewCavotiProvider(
	apiKey string,
	baseURL string,
	modelCheap string,
	modelStrong string,
	modelFast string,
	client *http.Client,
) *CavotiProvider {
	if baseURL == "" {
		baseURL = "https://cavoti.com/v1"
	}
	return &CavotiProvider{
		apiKey:      apiKey,
		baseURL:     strings.TrimRight(baseURL, "/"),
		modelCheap:  modelCheap,
		modelStrong: modelStrong,
		modelFast:   modelFast,
		httpClient:  client,
	}
}

func (p *CavotiProvider) Name() string { return "cavoti" }

func (p *CavotiProvider) ModelName(tier gtypes.ModelType) string {
	switch tier {
	case gtypes.ModelStrong:
		return p.modelStrong
	case gtypes.ModelFast:
		return p.modelFast
	default:
		return p.modelCheap
	}
}

// CostPer1K returns (inputCostPer1K, outputCostPer1K) in USD.
// Returns 0, 0 — check Cavoti dashboard for actual pricing.
func (p *CavotiProvider) CostPer1K(_ gtypes.ModelType) (float64, float64) {
	return 0, 0
}

func (p *CavotiProvider) MaxTokens(tier gtypes.ModelType) int {
	if tier == gtypes.ModelStrong {
		return 8192
	}
	return 4096
}

func (p *CavotiProvider) ExtractContent(raw interface{ MarshalJSON() ([]byte, error) }, reasoningContent string) string {
	return gtypes.StandardExtractContent(raw.(interface{ MarshalJSON() ([]byte, error) }), reasoningContent)
}

func (p *CavotiProvider) ExtractStreamToken(content, reasoningContent string) string {
	return gtypes.StandardExtractStreamToken(content, reasoningContent)
}

func (p *CavotiProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
	return doOpenAICompatibleCall(ctx, p.httpClient, p.baseURL, p.apiKey, p.ModelName(req.ModelTier), req)
}

func (p *CavotiProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	return doOpenAICompatibleStream(ctx, p.httpClient, p.baseURL, p.apiKey, p.ModelName(req.ModelTier), req)
}
