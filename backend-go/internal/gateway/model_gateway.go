package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
	"ai_avengers/backend/internal/gateway/providers"
)

// openRouterRequest/Response kept temporarily for getActiveProvider/getAPIKey
// which still read from DB. Will be cleaned up in next pass.
type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature"`
}
type openRouterMessage struct {
	Role         string        `json:"role"`
	Content      string        `json:"content"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}
type cacheControl struct { Type string `json:"type"` }
type openRouterResponse struct {
	Choices []struct { Message struct { Content string `json:"content"` } `json:"message"` } `json:"choices"`
	Usage   struct { PromptTokens int `json:"prompt_tokens"`; CompletionTokens int `json:"completion_tokens"` } `json:"usage"`
	Model   string `json:"model"`
}

// ModelType identifies which LLM to use.
// WHY three tiers:
// cheap  — fast, low cost, for metadata/tagging/coverage checks
// strong — best quality, for answer generation and charter extraction
// fast   — balanced, for quick checks
type ModelType string

const (
	ModelCheap  ModelType = "cheap"
	ModelStrong ModelType = "strong"
	ModelFast   ModelType = "fast"
)

// providerURLs kept for backward compat with getActiveProvider/getAPIKey.
// New code uses LLMProvider interface instead.
var providerURLs = map[config.LLMProvider]string{
	config.ProviderOpenRouter: "https://openrouter.ai/api/v1",
	config.ProviderDeepSeek:   "https://api.deepseek.com/v1",
	config.ProviderAnthropic:  "https://api.anthropic.com/v1",
	config.ProviderGemini:     "https://generativelanguage.googleapis.com/v1beta/openai",
}

// LLMRequest is the input to the model gateway.
type LLMRequest struct {
	Model        ModelType
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Temperature  float64
	UseCache     bool // Cache identical prompts
}

// LLMResponse is the output from the model gateway.
type LLMResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	CostUSD      float64
	ModelUsed    string
	Cached       bool
	DurationMs   float64
}

// openRouterRequest is the OpenRouter API request format.
type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature"`
}

// openRouterMessage is a single message in the conversation.
type openRouterMessage struct {
	Role         string        `json:"role"`
	Content      string        `json:"content"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}

// cacheControl enables Anthropic-style prompt caching via OpenRouter.
// WHY: System prompts are identical across calls to same expert.
// Caching saves 40-60% on token costs for repeated expert calls.
type cacheControl struct {
	Type string `json:"type"` // "ephemeral"
}

// openRouterResponse is the OpenRouter API response format.
type openRouterResponse struct {
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

// providerURLs maps provider → base URL for the chat completions endpoint.
var providerURLs = map[config.LLMProvider]string{
	config.ProviderOpenRouter: "https://openrouter.ai/api/v1",
	config.ProviderDeepSeek:   "https://api.deepseek.com/v1",
	config.ProviderAnthropic:  "https://api.anthropic.com/v1",
	config.ProviderGemini:     "https://generativelanguage.googleapis.com/v1beta/openai",
}

// ModelGateway is the single interface for all LLM calls.
// Delegates all provider-specific concerns to LLMProvider interface.
// To add a new provider: implement LLMProvider, register in buildProvider().
type ModelGateway struct {
	cfg        config.LLMConfig
	db         *pgxpool.Pool
	httpClient *http.Client
	cache      sync.Map
	totalCost  atomic.Value
	callCount  atomic.Int64
	logger     *zap.Logger
	// provider is built lazily and cached. Rebuilt when active provider changes.
	providerMu   sync.RWMutex
	providerName string      // last built provider name
	provider     LLMProvider // current active implementation
}

// NewModelGateway creates a new model gateway.
func NewModelGateway(cfg config.LLMConfig, logger *zap.Logger) *ModelGateway {
	g := &ModelGateway{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		logger:     logger,
	}
	g.totalCost.Store(float64(0))
	return g
}

// buildProvider creates the correct LLMProvider implementation.
// Called when provider changes or on first use.
// WHY factory here not in constructor:
//   Active provider may change at runtime (admin panel).
//   DB is not available at construction time.
func (g *ModelGateway) buildProvider(ctx context.Context) LLMProvider {
	providerName := string(g.getActiveProvider(ctx))
	apiKey := g.getAPIKey(ctx, config.LLMProvider(providerName))

	switch config.LLMProvider(providerName) {
	case config.ProviderDeepSeek:
		return providers.NewDeepSeekProvider(apiKey, g.httpClient)
	case config.ProviderAnthropic:
		return providers.NewAnthropicProvider(apiKey, g.httpClient)
	case config.ProviderGemini:
		return providers.NewGeminiProvider(apiKey, g.httpClient)
	default: // openrouter
		return providers.NewOpenRouterProvider(apiKey, g.cfg.OpenRouterBaseURL, g.httpClient)
	}
}

// getProvider returns the cached provider, rebuilding if active provider changed.
func (g *ModelGateway) getProvider(ctx context.Context) LLMProvider {
	currentName := string(g.getActiveProvider(ctx))

	g.providerMu.RLock()
	if g.provider != nil && g.providerName == currentName {
		p := g.provider
		g.providerMu.RUnlock()
		return p
	}
	g.providerMu.RUnlock()

	// Rebuild
	newProvider := g.buildProvider(ctx)
	g.providerMu.Lock()
	g.provider = newProvider
	g.providerName = currentName
	g.providerMu.Unlock()

	g.logger.Info("LLM provider switched", zap.String("provider", currentName))
	return newProvider
}

// Call makes an LLM call via OpenRouter.
// Retries up to 3 times with exponential backoff on failure.
// Caches responses if UseCache=true.
//
// Mental execution:
// Input: {Model: cheap, UserPrompt: "tag this turn", MaxTokens: 200}
// 1. Check cache — miss
// 2. Build OpenRouter request
// 3. POST to OpenRouter
// 4. Parse response
// 5. Track cost
// 6. Cache result
// 7. Return content
// Call makes an LLM call via the active provider.
// Provider resolved at call time — switching in admin panel takes effect immediately.
func (g *ModelGateway) Call(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	if req.UseCache {
		if cached, ok := g.cache.Load(g.cacheKey(req)); ok {
			result := cached.(*LLMResponse)
			result.Cached = true
			return result, nil
		}
	}

	provider := g.getProvider(ctx)

	maxTokens := req.MaxTokens
	if maxTokens <= 0 { maxTokens = 2000 }
	if provMax := provider.MaxTokens(req.Model); maxTokens > provMax { maxTokens = provMax }

	var messages []ProviderMessage
	if req.SystemPrompt != "" {
		messages = append(messages, ProviderMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, ProviderMessage{Role: "user", Content: req.UserPrompt})

	provReq := ProviderRequest{
		ModelTier: req.Model, Messages: messages,
		MaxTokens: maxTokens, Temperature: req.Temperature, EnableCache: req.UseCache,
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		start := time.Now()
		provResp, err := provider.Call(ctx, provReq)
		if err != nil {
			lastErr = err
			g.logger.Warn("LLM call attempt failed",
				zap.Int("attempt", attempt+1),
				zap.String("provider", provider.Name()),
				zap.Error(err))
			continue
		}

		duration := time.Since(start)
		inCost, outCost := provider.CostPer1K(req.Model)
		cost := float64(provResp.InputTokens)/1000*inCost + float64(provResp.OutputTokens)/1000*outCost

		currentCost := g.totalCost.Load().(float64)
		g.totalCost.Store(currentCost + cost)
		g.callCount.Add(1)

		result := &LLMResponse{
			Content: provResp.Content, InputTokens: provResp.InputTokens,
			OutputTokens: provResp.OutputTokens, CostUSD: cost,
			ModelUsed: provResp.ModelUsed, DurationMs: float64(duration.Milliseconds()),
		}
		g.logger.Info("LLM call complete",
			zap.String("provider", provider.Name()),
			zap.String("tier", string(req.Model)),
			zap.Float64("cost_usd", cost),
			zap.Float64("duration_ms", result.DurationMs))

		if req.UseCache { g.cache.Store(g.cacheKey(req), result) }
		return result, nil
	}
	return nil, fmt.Errorf("all LLM attempts failed: %w", lastErr)
}

// SetDB wires the database pool for runtime settings override.
// Called from main.go after DB connects.
// WHY separate from constructor: gateway is created before DB connects.
func (g *ModelGateway) SetDB(db *pgxpool.Pool) {
	g.db = db
}

// getActiveProvider reads the active LLM provider from DB system_settings.
// Falls back to cfg.Provider (env var). Falls back to openrouter.
func (g *ModelGateway) getActiveProvider(ctx context.Context) config.LLMProvider {
	if g.db != nil {
		var valueJSON []byte
		err := g.db.QueryRow(ctx,
			`SELECT value FROM system_settings WHERE key = 'llm_provider'`,
		).Scan(&valueJSON)
		if err == nil && len(valueJSON) > 0 {
			// value is JSONB string like "\"openrouter\"" or "\"deepseek\""
			var provider string
			if json.Unmarshal(valueJSON, &provider) == nil && provider != "" {
				return config.LLMProvider(provider)
			}
		}
	}
	if g.cfg.Provider != "" {
		return g.cfg.Provider
	}
	return config.ProviderOpenRouter
}

// getAPIKey returns the API key for the given provider.
// DB value (from admin panel) takes precedence over env var.
func (g *ModelGateway) getAPIKey(ctx context.Context, provider config.LLMProvider) string {
	// Check DB override first
	if g.db != nil {
		var valueJSON []byte
		err := g.db.QueryRow(ctx,
			`SELECT value FROM system_settings WHERE key = 'llm_api_keys'`,
		).Scan(&valueJSON)
		if err == nil && len(valueJSON) > 0 {
			var keys map[string]string
			if json.Unmarshal(valueJSON, &keys) == nil {
				if key, ok := keys[string(provider)]; ok && key != "" {
					return key
				}
			}
		}
	}
	// Fall back to env var
	switch provider {
	case config.ProviderDeepSeek:
		return g.cfg.DeepSeekAPIKey
	case config.ProviderAnthropic:
		return g.cfg.AnthropicAPIKey
	case config.ProviderGemini:
		return g.cfg.GeminiAPIKey
	default: // openrouter
		return g.cfg.OpenRouterAPIKey
	}
}

// callProvider routes an LLM call to the correct provider.
// Replaces callOpenRouter() — same OpenAI-compatible format for all providers.
func (g *ModelGateway) callProvider(
	ctx context.Context,
	modelName string,
	messages []openRouterMessage,
	maxTokens int,
	temperature float64,
) (*openRouterResponse, error) {
	provider := g.getActiveProvider(ctx)
	apiKey := g.getAPIKey(ctx, provider)
	if apiKey == "" {
		return nil, fmt.Errorf("no API key configured for provider %s", provider)
	}

	baseURL, ok := providerURLs[provider]
	if !ok {
		baseURL = providerURLs[config.ProviderOpenRouter]
	}
	// Allow env override of base URL (for OpenRouter custom endpoints)
	if provider == config.ProviderOpenRouter && g.cfg.OpenRouterBaseURL != "" {
		baseURL = strings.TrimRight(g.cfg.OpenRouterBaseURL, "/")
	}

	reqBody := openRouterRequest{
		Model:       modelName,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Anthropic uses x-api-key; everyone else uses Authorization: Bearer
	if provider == config.ProviderAnthropic {
		httpReq.Header.Set("x-api-key", apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if provider == config.ProviderOpenRouter {
		httpReq.Header.Set("HTTP-Referer", "https://ai-avengers.app")
		httpReq.Header.Set("X-Title", "AI Avengers")
	}

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider %s returned status %d", provider, resp.StatusCode)
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response from %s", provider)
	}
	return &result, nil
}

// GetStats returns usage statistics.
func (g *ModelGateway) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_calls": g.callCount.Load(),
		"total_cost":  g.totalCost.Load().(float64),
	}
}

// callOpenRouter makes the actual HTTP call to OpenRouter.
// Supports prompt caching for system prompts (40-60% cost reduction).
//
// WHY prompt caching (Byte by Byte AI course):
// Course taught: LLMs process tokens sequentially.
// System prompt is the same for every call to the same expert.
// Anthropic/OpenRouter supports prefix caching: system prompt
// is cached after first call, subsequent calls only pay for new tokens.
// For 1000 calls with 500-token system prompt: saves 500,000 tokens.
func (g *ModelGateway) callOpenRouter(
	ctx context.Context,
	modelName string,
	messages []openRouterMessage,
	maxTokens int,
	temperature float64,
) (*openRouterResponse, error) {
	// Enable cache_control on system message if present
	// This tells Anthropic/OpenRouter to cache the system prompt prefix
	for i, msg := range messages {
		if msg.Role == "system" {
			messages[i].CacheControl = &cacheControl{Type: "ephemeral"}
			break
		}
	}

	reqBody := openRouterRequest{
		Model:       modelName,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		g.cfg.OpenRouterBaseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.cfg.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://ai-avengers.app")
	req.Header.Set("X-Title", "AI Avengers")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenRouter returned status %d", resp.StatusCode)
	}

	var openRouterResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &openRouterResp, nil
}

// calculateCost computes the USD cost of an LLM call.
func (g *ModelGateway) calculateCost(cfg modelConfig, inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) / 1000 * cfg.CostPer1KInput
	outputCost := float64(outputTokens) / 1000 * cfg.CostPer1KOutput
	return inputCost + outputCost
}

// cacheKey generates a cache key for a request.
// WHY hash: Prevents memory issues with long prompts as map keys.
func (g *ModelGateway) cacheKey(req LLMRequest) string {
	return fmt.Sprintf("%s:%s:%s", req.Model, req.SystemPrompt, req.UserPrompt)
}
