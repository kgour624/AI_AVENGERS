package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// ExtractContent: Cavoti is OpenAI-compatible — standard plain string content.
// Falls back to reasoningContent for DeepSeek-style models routed via Cavoti.
func (p *CavotiProvider) ExtractContent(raw json.RawMessage, reasoningContent string) string {
	if len(raw) == 0 {
		return reasoningContent
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && strings.TrimSpace(s) != "" {
		return s
	}
	return reasoningContent
}

// ExtractStreamToken: standard delta.content + reasoning_content fallback.
func (p *CavotiProvider) ExtractStreamToken(content, reasoningContent string) string {
	return gtypes.StandardExtractStreamToken(content, reasoningContent)
}

// Call makes a blocking LLM call to Cavoti's OpenAI-compatible endpoint.
// Builds *http.Request first, then delegates to doOpenAICompatibleCall in common.go.
// Pattern is identical to CodeCraftAPIProvider.Call().
func (p *CavotiProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, err := json.Marshal(map[string]interface{}{
		"model":       p.ModelName(req.ModelTier),
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("cavoti: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cavoti: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleCall(p, p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

// StreamCall makes a streaming LLM call to Cavoti's OpenAI-compatible endpoint.
// Pattern is identical to CodeCraftAPIProvider.StreamCall().
func (p *CavotiProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
	var msgs []map[string]string
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]string{"role": m.Role, "content": m.Content})
	}
	body, err := json.Marshal(map[string]interface{}{
		"model":       p.ModelName(req.ModelTier),
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"stream":      true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cavoti stream: marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("cavoti stream: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleStream(ctx, p, p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

// Compile-time interface check — fails at build time if CavotiProvider
// does not fully implement gtypes.LLMProvider.
var _ gtypes.LLMProvider = (*CavotiProvider)(nil)
