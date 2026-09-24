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

// CodeCraftAPIProvider implements gtypes.LLMProvider for CodeCraftAPI.
//
// CodeCraftAPI is an AI model aggregator (similar to OpenRouter) that
// provides access to multiple models through a single API key.
//
// API format: OpenAI-compatible (same wire format as OpenRouter/DeepSeek/Gemini).
// This means doOpenAICompatibleCall() from common.go is reused directly.
// Zero new response-parsing code needed.
//
// WHY model names come from constructor (not hardcoded):
//   CodeCraftAPI exposes a live model catalog via GET /v1/models.
//   Admin selects which model to use per tier from the admin panel.
//   Hardcoding would require a code change every time admin wants a different model.
//   Constructor receives admin-configured names; empty string is valid at construction
//   time (admin may not have configured yet) — Call() will get a 400 from CodeCraftAPI
//   which surfaces to the admin as an actionable error.
//
// WHY CostPer1K returns 0, 0:
//   CodeCraftAPI pricing is not publicly documented at integration time.
//   Returning fabricated numbers would corrupt cost tracking.
//   Admin checks their CodeCraftAPI dashboard for actual costs.
type CodeCraftAPIProvider struct {
	apiKey      string
	baseURL     string // trimmed of trailing slash, e.g. "https://codecraftapi.com/v1"
	modelCheap  string // admin-configured model name for cheap tier
	modelStrong string // admin-configured model name for strong tier
	modelFast   string // admin-configured model name for fast tier
	httpClient  *http.Client
}

// NewCodeCraftAPIProvider creates a new CodeCraftAPIProvider.
//
// modelCheap/Strong/Fast are the admin-configured model names per tier.
// Empty string is valid — Call() will return an error from CodeCraftAPI
// (400 Bad Request) which surfaces to the admin as an actionable error.
func NewCodeCraftAPIProvider(
	apiKey string,
	baseURL string,
	modelCheap string,
	modelStrong string,
	modelFast string,
	client *http.Client,
) *CodeCraftAPIProvider {
	if baseURL == "" {
		baseURL = "https://codecraftapi.com/v1"
	}
	return &CodeCraftAPIProvider{
		apiKey:      apiKey,
		baseURL:     strings.TrimRight(baseURL, "/"),
		modelCheap:  modelCheap,
		modelStrong: modelStrong,
		modelFast:   modelFast,
		httpClient:  client,
	}
}

// Name returns the provider identifier used in system_settings and admin UI.
func (p *CodeCraftAPIProvider) Name() string { return "codecraftapi" }

// ModelName returns the admin-configured model name for the given tier.
// Returns empty string if admin has not configured a model for this tier.
// An empty model name will cause CodeCraftAPI to return 400 — this is
// intentional: it surfaces as an actionable error to the admin.
func (p *CodeCraftAPIProvider) ModelName(tier gtypes.ModelType) string {
	switch tier {
	case gtypes.ModelStrong:
		return p.modelStrong
	case gtypes.ModelFast:
		return p.modelFast
	default: // ModelCheap and any future tiers
		return p.modelCheap
	}
}

// CostPer1K returns 0, 0 because CodeCraftAPI pricing is not publicly
// documented at integration time. Admin checks their CodeCraftAPI dashboard.
func (p *CodeCraftAPIProvider) CostPer1K(_ gtypes.ModelType) (float64, float64) {
	return 0, 0
}

// MaxTokens returns the per-call ceiling for CodeCraftAPI.
//
// WHY 21000: CodeCraftAPI's docs (https://codecraftapi.com/docs/reasoning) warn
// that reasoning tokens count toward completion_tokens, so a tight budget can be
// spent entirely on reasoning and leave `content` empty. They recommend 16000+
// for hard problems. The previous ceiling of 8192 silently capped every caller
// below that guidance — the gateway caps maxTokens to this value — which is why
// charter extraction failed with "empty content in response". 21000 gives the
// reasoning + answer room the docs ask for while staying a sane per-call bound.
func (p *CodeCraftAPIProvider) MaxTokens(_ gtypes.ModelType) int { return 21000 }

// ExtractContent: CodeCraftAPI passes through native model formats.
// Claude thinking models return a typed array; DeepSeek uses reasoning_content.
// This provider handles both via the same logic as AnthropicProvider
// (array check first) plus reasoning_content fallback for DeepSeek-style models.
//
// WHY not StandardExtractContent:
//   CodeCraftAPI is a semi-raw proxy — it does NOT normalize inner payload.
//   Claude Opus 5 via CodeCraftAPI returns content as [{type:text,text:...}].
//   DeepSeek-V4-Flash via CodeCraftAPI returns content="", reasoning_content="...".
//   Both cases must be handled here so any model in the catalog works.
func (p *CodeCraftAPIProvider) ExtractContent(raw json.RawMessage, reasoningContent string) string {
	if len(raw) == 0 {
		return reasoningContent
	}
	// Claude-style typed array
	if raw[0] == '[' {
		var blocks []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
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
		return reasoningContent
	}
	// Standard plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && strings.TrimSpace(s) != "" {
		return s
	}
	// DeepSeek-style: content empty, answer in reasoning_content
	return reasoningContent
}

// ExtractStreamToken: standard delta.content + reasoning_content fallback.
// Covers both standard streaming models and DeepSeek reasoning streaming.
func (p *CodeCraftAPIProvider) ExtractStreamToken(content, reasoningContent string) string {
	return gtypes.StandardExtractStreamToken(content, reasoningContent)
}
//
// Uses doOpenAICompatibleCall() from common.go — CodeCraftAPI uses the
// same OpenAI-compatible response format as OpenRouter/DeepSeek/Gemini.
//
// WHY no prompt caching (EnableCache not handled):
//   CodeCraftAPI's support for Anthropic-style cache_control headers is
//   unknown at integration time. Adding cache_control to an API that
//   doesn't support it could break the call. Omitting it is safe —
//   worst case is no caching, not a broken call.
func (p *CodeCraftAPIProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error) {
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
		return nil, fmt.Errorf("codecraftapi: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("codecraftapi: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleCall(p, p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

// StreamCall implements streaming for CodeCraftAPI.
func (p *CodeCraftAPIProvider) StreamCall(ctx context.Context, req gtypes.ProviderRequest) (<-chan string, <-chan *gtypes.ProviderResponse, error) {
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
		return nil, nil, fmt.Errorf("codecraftapi stream: marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("codecraftapi stream: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return doOpenAICompatibleStream(ctx, p, p.httpClient, httpReq, p.ModelName(req.ModelTier))
}

var _ gtypes.LLMProvider = (*CodeCraftAPIProvider)(nil)
