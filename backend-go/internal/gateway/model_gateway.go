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
	if maxTokens <= 0 {
		maxTokens = 2000
	}
	if provMax := provider.MaxTokens(req.Model); maxTokens > provMax {
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
		provResp, err := provider.Call(ctx, provReq)
		if err != nil {
			lastErr = err
			g.logger.Warn("LLM call attempt failed",
				zap.Int("attempt", attempt+1),
				zap.String("provider", provider.Name()),
				zap.Error(err),
			)
			continue
		}

		duration := time.Since(start)
		inCost, outCost := provider.CostPer1K(req.Model)
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
			zap.String("provider", provider.Name()),
			zap.String("tier", string(req.Model)),
			zap.Float64("cost_usd", cost),
			zap.Float64("duration_ms", result.DurationMs),
		)

		if req.UseCache {
			g.cache.Store(g.cacheKey(req), result)
		}
		return result, nil
	}
	return nil, fmt.Errorf("all LLM attempts failed: %w", lastErr)
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
	default:
		return g.cfg.OpenRouterAPIKey
	}
}

// cacheKey generates a stable cache key for a request.
func (g *ModelGateway) cacheKey(req LLMRequest) string {
	return fmt.Sprintf("%s:%s:%s", req.Model, req.SystemPrompt, req.UserPrompt)
}
