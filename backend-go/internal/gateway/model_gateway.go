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

// LLMRequest is the input to the model gateway.
// Uses ModelType from gateway/types (re-exported via provider.go).
type LLMRequest struct {
	Model        ModelType
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Temperature  float64
	UseCache     bool
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

// ModelGateway is the single interface for all LLM calls.
// Delegates all provider-specific concerns to LLMProvider interface.
//
// To add a new provider:
//   1. Create internal/gateway/providers/myprovider.go
//   2. Implement gateway/types.LLMProvider interface
//   3. Add a case in buildProvider() below
//   Nothing else changes.
type ModelGateway struct {
	cfg        config.LLMConfig
	db         *pgxpool.Pool // nil = env config only, no DB override
	httpClient *http.Client
	cache      sync.Map
	totalCost  atomic.Value
	callCount  atomic.Int64
	logger     *zap.Logger
	// provider is built lazily, cached, rebuilt when active provider changes.
	providerMu   sync.RWMutex
	providerName string
	provider     LLMProvider
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

// SetDB wires the database pool for runtime settings override.
// Called from main.go after DB connects.
func (g *ModelGateway) SetDB(db *pgxpool.Pool) {
	g.db = db
}

// buildProviderByName creates an LLMProvider for the given provider name.
// Used by getFallbackProvider() to build the fallback without touching
// the cached primary provider (g.provider / g.providerName).
func (g *ModelGateway) buildProviderByName(ctx context.Context, providerName string) LLMProvider {
	apiKey := g.getAPIKey(ctx, config.LLMProvider(providerName))
	switch config.LLMProvider(providerName) {
	case config.ProviderDeepSeek:
		return providers.NewDeepSeekProvider(apiKey, g.httpClient)
	case config.ProviderAnthropic:
		return providers.NewAnthropicProvider(apiKey, g.httpClient)
	case config.ProviderGemini:
		return providers.NewGeminiProvider(apiKey, g.httpClient)
	case config.ProviderCodeCraftAPI:
		modelCheap := g.getModelName(ctx, "codecraftapi_model_cheap", "")
		modelStrong := g.getModelName(ctx, "codecraftapi_model_strong", "")
		modelFast := g.getModelName(ctx, "codecraftapi_model_fast", "")
		return providers.NewCodeCraftAPIProvider(
			apiKey, g.cfg.CodeCraftAPIBaseURL,
			modelCheap, modelStrong, modelFast,
			g.httpClient,
		)
	default: // openrouter
		return providers.NewOpenRouterProvider(apiKey, g.cfg.OpenRouterBaseURL, g.httpClient)
	}
}

// getFallbackProvider reads 'llm_fallback_provider' from system_settings
// and returns a ready-to-use LLMProvider, or nil if no fallback is configured.
// Returns nil when DB is not wired or key is missing/empty.
func (g *ModelGateway) getFallbackProvider(ctx context.Context) LLMProvider {
	if g.db == nil {
		return nil
	}
	var valueJSON []byte
	err := g.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'llm_fallback_provider'`,
	).Scan(&valueJSON)
	if err != nil || len(valueJSON) == 0 {
		return nil
	}
	var name string
	if json.Unmarshal(valueJSON, &name) != nil || name == "" {
		return nil
	}
	return g.buildProviderByName(ctx, name)
}

// buildProvider creates the correct LLMProvider implementation for the
// currently active provider. Called lazily on first use and on change.
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
	case config.ProviderCodeCraftAPI:
		// Model names are admin-configured per tier, stored in system_settings.
		// Empty string is valid at construction time — Call() will get a 400
		// from CodeCraftAPI which surfaces as an actionable error to the admin.
		modelCheap := g.getModelName(ctx, "codecraftapi_model_cheap", "")
		modelStrong := g.getModelName(ctx, "codecraftapi_model_strong", "")
		modelFast := g.getModelName(ctx, "codecraftapi_model_fast", "")
		return providers.NewCodeCraftAPIProvider(
			apiKey, g.cfg.CodeCraftAPIBaseURL,
			modelCheap, modelStrong, modelFast,
			g.httpClient,
		)
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

	newProvider := g.buildProvider(ctx)
	g.providerMu.Lock()
	g.provider = newProvider
	g.providerName = currentName
	g.providerMu.Unlock()

	g.logger.Info("LLM provider active", zap.String("provider", currentName))
	return newProvider
}

// Call makes an LLM call via the active provider.
// Provider is resolved at call time — switching provider in admin panel
// takes effect on the next Call() with no restart needed.
//
// Fallback behavior:
//   If all 3 attempts on the primary provider fail (any error), and
//   'llm_fallback_provider' is configured in system_settings AND is
//   different from the primary, Call() retries 3 more times on the
//   fallback provider. This is per-Call() — global provider is never
//   mutated. If fallback also fails, the original error is returned.
func (g *ModelGateway) Call(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	if req.UseCache {
		if cached, ok := g.cache.Load(g.cacheKey(req)); ok {
			result := cached.(*LLMResponse)
			result.Cached = true
			return result, nil
		}
	}

	primary := g.getProvider(ctx)

	// tryProvider runs up to 3 attempts on the given provider.
	// Returns (response, nil) on first success.
	// Returns (nil, lastErr) if all attempts fail.
	tryProvider := func(p LLMProvider) (*LLMResponse, error) {
		maxTokens := req.MaxTokens
		if maxTokens <= 0 {
			maxTokens = 2000
		}
		if provMax := p.MaxTokens(req.Model); maxTokens > provMax {
			maxTokens = provMax
		}

		var messages []ProviderMessage
		if req.SystemPrompt != "" {
			messages = append(messages, ProviderMessage{Role: "system", Content: req.SystemPrompt})
		}
		messages = append(messages, ProviderMessage{Role: "user", Content: req.UserPrompt})

		provReq := ProviderRequest{
			ModelTier:   req.Model,
			Messages:    messages,
			MaxTokens:   maxTokens,
			Temperature: req.Temperature,
			EnableCache: req.UseCache,
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
			provResp, err := p.Call(ctx, provReq)
			if err != nil {
				lastErr = err
				g.logger.Warn("LLM call attempt failed",
					zap.Int("attempt", attempt+1),
					zap.String("provider", p.Name()),
					zap.Error(err),
				)
				continue
			}

			duration := time.Since(start)
			inCost, outCost := p.CostPer1K(req.Model)
			cost := float64(provResp.InputTokens)/1000*inCost +
				float64(provResp.OutputTokens)/1000*outCost

			currentCost := g.totalCost.Load().(float64)
			g.totalCost.Store(currentCost + cost)
			g.callCount.Add(1)

			result := &LLMResponse{
				Content:      provResp.Content,
				InputTokens:  provResp.InputTokens,
				OutputTokens: provResp.OutputTokens,
				CostUSD:      cost,
				ModelUsed:    provResp.ModelUsed,
				DurationMs:   float64(duration.Milliseconds()),
			}

			g.logger.Info("LLM call complete",
				zap.String("provider", p.Name()),
				zap.String("tier", string(req.Model)),
				zap.Float64("cost_usd", cost),
				zap.Float64("duration_ms", result.DurationMs),
			)

			if req.UseCache {
				g.cache.Store(g.cacheKey(req), result)
			}
			return result, nil
		}
		return nil, lastErr
	}

	// Try primary provider first.
	result, err := tryProvider(primary)
	if err == nil {
		return result, nil
	}

	// Primary failed all 3 attempts. Try fallback if configured.
	// getFallbackProvider() returns nil when no fallback is set —
	// in that case we return the primary's error unchanged (backward compatible).
	fallback := g.getFallbackProvider(ctx)
	if fallback == nil {
		return nil, fmt.Errorf("all LLM attempts failed: %w", err)
	}
	if fallback.Name() == primary.Name() {
		// Fallback is the same provider as primary — retrying is pointless.
		g.logger.Warn("LLM fallback provider is same as primary — skipping fallback",
			zap.String("provider", primary.Name()),
		)
		return nil, fmt.Errorf("all LLM attempts failed: %w", err)
	}

	g.logger.Warn("primary LLM provider failed — trying fallback",
		zap.String("primary", primary.Name()),
		zap.String("fallback", fallback.Name()),
		zap.Error(err),
	)

	result, fallbackErr := tryProvider(fallback)
	if fallbackErr == nil {
		return result, nil
	}

	g.logger.Error("both primary and fallback LLM providers failed",
		zap.String("primary", primary.Name()),
		zap.String("fallback", fallback.Name()),
		zap.NamedError("primary_err", err),
		zap.NamedError("fallback_err", fallbackErr),
	)
	return nil, fmt.Errorf("all LLM attempts failed (primary: %s, fallback: %s): %w",
		primary.Name(), fallback.Name(), fallbackErr)
}

// GetStats returns usage statistics.
func (g *ModelGateway) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_calls": g.callCount.Load(),
		"total_cost":  g.totalCost.Load().(float64),
	}
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
	switch provider {
	case config.ProviderDeepSeek:
		return g.cfg.DeepSeekAPIKey
	case config.ProviderAnthropic:
		return g.cfg.AnthropicAPIKey
	case config.ProviderGemini:
		return g.cfg.GeminiAPIKey
	case config.ProviderCodeCraftAPI:
		return g.cfg.CodeCraftAPIKey
	default:
		return g.cfg.OpenRouterAPIKey
	}
}

// getModelName reads a per-tier model name from system_settings.
// Falls back to the provided default if the key is missing or empty.
//
// WHY this method exists:
//   CodeCraftAPI exposes a live model catalog. Admin selects which model
//   to use per tier from the admin panel. Model names are stored in
//   system_settings (e.g. key="codecraftapi_model_strong").
//   Same pattern as getActiveProvider() and getAPIKey().
func (g *ModelGateway) getModelName(ctx context.Context, settingKey, fallback string) string {
	if g.db == nil {
		return fallback
	}
	var valueJSON []byte
	err := g.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = $1`,
		settingKey,
	).Scan(&valueJSON)
	if err != nil || len(valueJSON) == 0 {
		return fallback
	}
	var name string
	if json.Unmarshal(valueJSON, &name) != nil || name == "" {
		return fallback
	}
	return name
}

// cacheKey generates a stable cache key for a request.
func (g *ModelGateway) cacheKey(req LLMRequest) string {
	return fmt.Sprintf("%s:%s:%s", req.Model, req.SystemPrompt, req.UserPrompt)
}
