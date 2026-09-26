package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
	"ai_avengers/backend/internal/gateway/providers"
	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/usage"
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
	// WorkflowID attributes this call's cost to a workflow.
	//
	// nil (the default) = not part of a workflow — e.g. the chat flow,
	// which already records its cost on the messages row. Behaviour for
	// those callers is unchanged.
	//
	// When set, Call() adds the call's cost to workflows.cost_spent_usd.
	// WHY here and not at every call site: this is the single place where
	// cost is computed, so attribution cannot be forgotten by a new caller.
	// WHY it does not affect caching: cacheKey() is built from
	// Model + SystemPrompt + UserPrompt only.
	WorkflowID *uuid.UUID

	// Messages carries a full multi-turn conversation.
	//
	// nil (the default) = use SystemPrompt + UserPrompt. Every existing
	// caller is a single-shot prompt, so their behaviour is unchanged.
	//
	// When set, it replaces SystemPrompt/UserPrompt and is passed to the
	// provider verbatim, assistant turns included.
	// WHY this exists: the /llm/proxy endpoint serves Aider, which sends a
	// real conversation (system, few-shot user/assistant pairs, chat
	// history, reminder). Squeezing that into two strings dropped every
	// assistant turn and every message except the last of each role — the
	// model was being asked to edit code it had never been shown writing.
	// ProviderRequest already took []ProviderMessage; only this struct
	// could not express it.
	Messages []ProviderMessage
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
//  1. Create internal/gateway/providers/myprovider.go
//  2. Implement gateway/types.LLMProvider interface
//  3. Add a case in buildProvider() below
//     Nothing else changes.
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
	// usageRec (C5): optional per-call usage recorder. nil = no persistence.
	// C5 single choke point — every real (non-cached) call is recorded here,
	// so a new caller cannot forget to attribute its cost.
	usageRec usage.Recorder
	// limitsMu guards limitsSnap, the cached read of llm_model_limits. The
	// limits are enforced in Call/StreamCall — the same choke point as cost, so
	// a new caller cannot spend tokens a limit was meant to cap.
	limitsMu   sync.RWMutex
	limitsSnap *modelLimitsSnapshot
	// breaker remembers which providers are currently failing, so a provider
	// that is down is skipped instead of costing every request three failed
	// attempts first (P8).
	breaker *Breaker
	// latencyMu guards latency, the per-provider rolling call durations. The
	// numbers exist because "is it up" is only half the question: a provider that
	// answers in 40s is unusable in a way a health flag cannot express.
	latencyMu sync.Mutex
	latency   map[string]*latencyWindow
}

// NewModelGateway creates a new model gateway.
func NewModelGateway(cfg config.LLMConfig, logger *zap.Logger) *ModelGateway {
	g := &ModelGateway{
		cfg: cfg,
		// Transport-level timeouts prevent TCP hangs when the LLM provider
		// drops the connection silently (no RST, no FIN).
		// TLSHandshakeTimeout: fail fast on TLS negotiation hang.
		// ResponseHeaderTimeout: fail if server never sends the first byte.
		// The outer http.Client.Timeout (120s) is the wall-clock cap for
		// the entire round-trip including body read — kept for streaming
		// compatibility (streaming body reads can legitimately take >60s).
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		logger:  logger,
		breaker: NewBreaker(DefaultBreakerConfig(), logger),
		latency: make(map[string]*latencyWindow),
	}
	g.totalCost.Store(float64(0))
	return g
}

// SetBreakerConfig replaces the breaker's configuration. Called from main.go
// with the environment's settings; the default (enabled, 3 failures, 60s) is
// installed at construction so a deployment that configures nothing still gets
// the protection.
func (g *ModelGateway) SetBreakerConfig(cfg BreakerConfig) {
	if g.breaker == nil {
		g.breaker = NewBreaker(cfg, g.logger)
		return
	}
	g.breaker.cfg = cfg
}

// BreakerSnapshot reports each provider's breaker state for the admin surface.
func (g *ModelGateway) BreakerSnapshot() []BreakerStatus {
	if g == nil || g.breaker == nil {
		return nil
	}
	return g.breaker.Snapshot()
}

// SetDB wires the database pool for runtime settings override.
// Called from main.go after DB connects.
func (g *ModelGateway) SetDB(db *pgxpool.Pool) {
	g.db = db
}

// SetUsageRecorder wires C5 usage persistence. Called from main.go after the
// usage service is built. nil = disabled (no change to existing behaviour).
func (g *ModelGateway) SetUsageRecorder(r usage.Recorder) {
	g.usageRec = r
}

// recordUsage persists one call's usage when a recorder is wired. Attribution
// is read from the context (usage.WithAttribution) and the request's
// WorkflowID; anything still unknown is resolved at write time by the usage
// service (project←workflow, tenant←project). Best-effort and detached — the
// spend already happened even if ctx is cancelled.
func (g *ModelGateway) recordUsage(ctx context.Context, req LLMRequest, provider, model, tier string, in, out int, cost float64) {
	if g.usageRec == nil || cost <= 0 {
		return
	}
	a, _ := usage.AttributionFrom(ctx)
	if req.WorkflowID != nil {
		a.WorkflowID = req.WorkflowID
	}
	_ = g.usageRec.Record(context.WithoutCancel(ctx), usage.Event{
		Attribution:  a,
		Provider:     provider,
		Tier:         tier,
		Model:        model,
		InputTokens:  in,
		OutputTokens: out,
		CostUSD:      cost,
		OccurredAt:   time.Now().UTC(),
	})
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
	case config.ProviderCavoti:
		modelCheap := g.getModelName(ctx, "cavoti_model_cheap", "")
		modelStrong := g.getModelName(ctx, "cavoti_model_strong", "")
		modelFast := g.getModelName(ctx, "cavoti_model_fast", "")
		return providers.NewCavotiProvider(
			apiKey, g.cfg.CavotiBaseURL,
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
	case config.ProviderCavoti:
		modelCheap := g.getModelName(ctx, "cavoti_model_cheap", "")
		modelStrong := g.getModelName(ctx, "cavoti_model_strong", "")
		modelFast := g.getModelName(ctx, "cavoti_model_fast", "")
		return providers.NewCavotiProvider(
			apiKey, g.cfg.CavotiBaseURL,
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
//
//	If all 3 attempts on the primary provider fail (any error), and
//	'llm_fallback_provider' is configured in system_settings AND is
//	different from the primary, Call() retries 3 more times on the
//	fallback provider. This is per-Call() — global provider is never
//	mutated. If fallback also fails, the original error is returned.
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
		if pricingProvider, ok := p.(interface{ RefreshPricing(context.Context) error }); ok {
			if err := pricingProvider.RefreshPricing(ctx); err != nil {
				g.logger.Warn("provider pricing refresh failed; retaining last known rates",
					zap.String("provider", p.Name()),
					zap.Error(err),
				)
			}
		}

		// The ceiling: the admin's configured maximum when one is set, else the
		// provider's own. See resolveCeiling for why the configured value wins
		// even when it is larger.
		limit, configured := g.resolveModelLimit(ctx, p.Name(), req.Model)
		ceiling := resolveCeiling(p.MaxTokens(req.Model), limit.MaxOutputTokens)
		maxTokens := effectiveOutputBudget(req.MaxTokens, ceiling)

		var messages []ProviderMessage
		if len(req.Messages) > 0 {
			messages = req.Messages
		} else {
			if req.SystemPrompt != "" {
				messages = append(messages, ProviderMessage{Role: "system", Content: req.SystemPrompt})
			}
			messages = append(messages, ProviderMessage{Role: "user", Content: req.UserPrompt})
		}

		// Input guard (P3 fail-closed). Only when an input limit is configured:
		// "not configured" means no opinion, and guessing a model's context
		// window here would refuse requests the provider accepts today.
		if configured && limit.MaxInputTokens > 0 {
			if est := estimateMessagesTokens(messages); est > limit.MaxInputTokens {
				return nil, fmt.Errorf(
					"%w: model=%s estimated_input_tokens=%d max_input_tokens=%d — shorten the request, or raise the limit in Admin → LLM Settings",
					ErrInputTooLong, p.Name(), est, limit.MaxInputTokens,
				)
			}
		}

		provReq := ProviderRequest{
			ModelTier:   req.Model,
			Messages:    messages,
			MaxTokens:   maxTokens,
			Temperature: req.Temperature,
			EnableCache: req.UseCache,
		}

		// escalated: a reasoning model can spend the whole completion budget on
		// thinking and return no visible text. Re-sending the SAME budget cannot
		// help (which is why this used to give up immediately), but a larger one
		// can — so escalate once to the ceiling before giving up.
		escalated := false
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
				// Deterministic failure: the provider returned 200 but no usable
				// content (typically a reasoning model whose completion budget
				// was spent on reasoning). An identical re-send cannot succeed,
				// so stop retrying — but first spend the budget the caller never
				// asked for, which is the difference between a plan and a
				// workflow that dies at intake.
				if errors.Is(err, providers.ErrEmptyContent) {
					if !escalated && ceiling > provReq.MaxTokens {
						escalated = true
						lastErr = err
						g.logger.Warn("LLM returned no visible text — retrying once with the full output budget",
							zap.String("provider", p.Name()),
							zap.Int("from_max_tokens", provReq.MaxTokens),
							zap.Int("to_max_tokens", ceiling),
						)
						provReq.MaxTokens = ceiling
						continue
					}
					g.logger.Warn("LLM call returned empty content — skipping retries (deterministic)",
						zap.String("provider", p.Name()),
						zap.Int("max_tokens", provReq.MaxTokens),
					)
					break
				}
				continue
			}

			duration := time.Since(start)
			inCost, outCost := p.CostPer1K(req.Model)
			cost := float64(provResp.InputTokens)/1000*inCost +
				float64(provResp.OutputTokens)/1000*outCost

			currentCost := g.totalCost.Load().(float64)
			g.totalCost.Store(currentCost + cost)
			g.callCount.Add(1)
			observability.Global.IncLLMCall()
			observability.Global.AddLLMCost(cost)

			// Attribute the spend to the workflow, if this call belongs to one.
			// Only reached on a real provider call — the cache hit above
			// returns early, so a cached answer is never charged twice.
			g.addWorkflowCost(ctx, req.WorkflowID, cost)
			// C5: persist per-call usage for the analytics/budget surface.
			g.recordUsage(ctx, req, p.Name(), provResp.ModelUsed, string(req.Model),
				provResp.InputTokens, provResp.OutputTokens, cost)

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
		observability.Global.IncLLMError()
		return nil, lastErr
	}

	// Try the primary provider, unless the breaker has taken it out of rotation.
	var err error
	if allowed, reason := g.breaker.Allow(primary.Name()); !allowed {
		// WHY skip entirely rather than try: the failure is already known. Three
		// more attempts plus 1s and 2s of backoff would only re-learn it, while
		// adding load to a provider that is already struggling.
		g.logger.Warn("LLM provider skipped by the breaker",
			zap.String("provider", primary.Name()),
			zap.String("reason", reason),
		)
		err = fmt.Errorf("provider %s skipped: %s", primary.Name(), reason)
	} else {
		callStart := time.Now()
		result, callErr := tryProvider(primary)
		g.recordLatency(primary.Name(), float64(time.Since(callStart).Milliseconds()))
		if callErr == nil {
			g.breaker.RecordSuccess(primary.Name())
			return result, nil
		}
		g.breaker.RecordFailure(primary.Name(), callErr)
		err = callErr
	}

	// Primary unavailable. Try fallback if configured.
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

	// The fallback has its own breaker entry: two providers can be unwell
	// independently, and a fallback that is down must not become a second source
	// of three-second waits.
	if allowed, fbReason := g.breaker.Allow(fallback.Name()); !allowed {
		g.logger.Error("fallback LLM provider is also skipped by its breaker",
			zap.String("fallback", fallback.Name()),
			zap.String("reason", fbReason),
		)
		return nil, fmt.Errorf("all LLM attempts failed (primary: %s, fallback: %s skipped: %s): %w",
			primary.Name(), fallback.Name(), fbReason, err)
	}

	g.logger.Warn("primary LLM provider failed — trying fallback",
		zap.String("primary", primary.Name()),
		zap.String("fallback", fallback.Name()),
		zap.Error(err),
	)

	fallbackStart := time.Now()
	result, fallbackErr := tryProvider(fallback)
	g.recordLatency(fallback.Name(), float64(time.Since(fallbackStart).Milliseconds()))
	if fallbackErr == nil {
		g.breaker.RecordSuccess(fallback.Name())
		return result, nil
	}
	g.breaker.RecordFailure(fallback.Name(), fallbackErr)

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

// StreamCall makes a streaming LLM call via the active provider.
// Returns tokenCh (individual tokens as they arrive) and respCh (final metadata).
// Used by chinawall/enforcer.go generateFlatText for Gate 5 generation.
//
// WHY streaming only for Gate 5:
//
//	Gates 1-4 are fast (keyword check, vector search, charter check, necessity).
//	Only Gate 5 (LLM generation) takes 30-67s. Streaming Gate 5 means
//	the user sees the first token in 2-3s instead of waiting 67s.
//
// Fallback: if StreamCall fails, caller should fall back to Call().
func (g *ModelGateway) StreamCall(ctx context.Context, req LLMRequest) (<-chan string, <-chan *LLMResponse, error) {
	provider := g.getProvider(ctx)

	// Breaker gate (P8). Skipping a provider that is known to be failing matters
	// most on this path: streaming is what the chat screen uses, and before this
	// it had no fallback at all — a dead primary simply ended the conversation.
	if allowed, reason := g.breaker.Allow(provider.Name()); !allowed {
		fallback := g.getFallbackProvider(ctx)
		if fallback == nil || fallback.Name() == provider.Name() {
			return nil, nil, fmt.Errorf("provider %s skipped: %s", provider.Name(), reason)
		}
		if fbAllowed, fbReason := g.breaker.Allow(fallback.Name()); !fbAllowed {
			return nil, nil, fmt.Errorf(
				"provider %s skipped (%s) and the fallback %s is skipped too (%s)",
				provider.Name(), reason, fallback.Name(), fbReason,
			)
		}
		g.logger.Warn("stream: primary skipped by the breaker — using the fallback",
			zap.String("primary", provider.Name()),
			zap.String("fallback", fallback.Name()),
			zap.String("reason", reason),
		)
		provider = fallback
	}

	// Same ceiling rule as Call() — see resolveCeiling. No empty-content
	// escalation here: a stream that already started cannot be re-sent as a
	// different request without the caller losing whatever partial answer it
	// already received.
	limit, limitSet := g.resolveModelLimit(ctx, provider.Name(), req.Model)
	if limitSet && limit.MaxInputTokens > 0 {
		est := estimateTokens(req.SystemPrompt) + estimateTokens(req.UserPrompt)
		if est > limit.MaxInputTokens {
			return nil, nil, fmt.Errorf(
				"%w: model=%s estimated_input_tokens=%d max_input_tokens=%d — shorten the request, or raise the limit in Admin → LLM Settings",
				ErrInputTooLong, provider.Name(), est, limit.MaxInputTokens,
			)
		}
	}
	ceiling := resolveCeiling(provider.MaxTokens(req.Model), limit.MaxOutputTokens)
	maxTokens := effectiveOutputBudget(req.MaxTokens, ceiling)

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

	if pricingProvider, ok := provider.(interface{ RefreshPricing(context.Context) error }); ok {
		if err := pricingProvider.RefreshPricing(ctx); err != nil {
			g.logger.Warn("provider pricing refresh failed; retaining last known rates",
				zap.String("provider", provider.Name()),
				zap.Error(err),
			)
		}
	}

	streamStart := time.Now()
	rawTokenCh, rawRespCh, err := provider.StreamCall(ctx, provReq)
	if err != nil {
		g.breaker.RecordFailure(provider.Name(), err)

		// A stream that never started produced no partial answer, so the same
		// request can be served by the other provider without the caller losing
		// anything. A stream that fails MID-FLIGHT is deliberately NOT retried:
		// by then the caller has shown partial text, and silently starting again
		// would duplicate it on screen.
		if fallback := g.getFallbackProvider(ctx); fallback != nil && fallback.Name() != provider.Name() {
			if fbAllowed, _ := g.breaker.Allow(fallback.Name()); fbAllowed {
				g.logger.Warn("stream: primary stream failed to start — trying the fallback",
					zap.String("primary", provider.Name()),
					zap.String("fallback", fallback.Name()),
					zap.Error(err),
				)
				if pricingProvider, ok := fallback.(interface{ RefreshPricing(context.Context) error }); ok {
					if pricingErr := pricingProvider.RefreshPricing(ctx); pricingErr != nil {
						g.logger.Warn("fallback provider pricing refresh failed; retaining last known rates",
							zap.String("provider", fallback.Name()),
							zap.Error(pricingErr),
						)
					}
				}
				tokens, resp, fbErr := fallback.StreamCall(ctx, provReq)
				if fbErr == nil {
					g.breaker.RecordSuccess(fallback.Name())
					provider, rawTokenCh, rawRespCh, err = fallback, tokens, resp, nil
				} else {
					g.breaker.RecordFailure(fallback.Name(), fbErr)
				}
			}
		}
		if err != nil {
			return nil, nil, fmt.Errorf("stream call failed: %w", err)
		}
	}

	// Wrap rawRespCh to accumulate cost + update stats, same as Call().
	tokenCh := rawTokenCh // pass through directly — no wrapping needed
	respCh := make(chan *LLMResponse, 1)

	go func() {
		defer close(respCh)
		provResp, ok := <-rawRespCh
		if !ok || provResp == nil {
			// The stream ended without a final response. That is a provider
			// failure, not an empty answer, and it counts against the breaker —
			// otherwise a provider that accepts streams and then drops them would
			// look perfectly healthy.
			g.breaker.RecordFailure(provider.Name(), fmt.Errorf("stream ended without a final response"))
			return
		}
		g.breaker.RecordSuccess(provider.Name())
		// Full stream duration, measured the same way as a non-streamed call, so
		// the percentiles on the status screen mean one thing.
		g.recordLatency(provider.Name(), float64(time.Since(streamStart).Milliseconds()))
		inCost, outCost := provider.CostPer1K(req.Model)
		cost := float64(provResp.InputTokens)/1000*inCost +
			float64(provResp.OutputTokens)/1000*outCost
		currentCost := g.totalCost.Load().(float64)
		g.totalCost.Store(currentCost + cost)
		g.callCount.Add(1)
		observability.Global.IncLLMCall()
		observability.Global.AddLLMCost(cost)
		// C5: persist per-call usage for the streaming path too.
		g.recordUsage(ctx, req, provider.Name(), provResp.ModelUsed, string(req.Model),
			provResp.InputTokens, provResp.OutputTokens, cost)
		g.logger.Info("LLM stream complete",
			zap.String("provider", provider.Name()),
			zap.String("tier", string(req.Model)),
			zap.Float64("cost_usd", cost),
		)
		respCh <- &LLMResponse{
			Content:      provResp.Content,
			InputTokens:  provResp.InputTokens,
			OutputTokens: provResp.OutputTokens,
			CostUSD:      cost,
			ModelUsed:    provResp.ModelUsed,
		}
	}()

	return tokenCh, respCh, nil
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
	case config.ProviderCavoti:
		return g.cfg.CavotiAPIKey
	default:
		return g.cfg.OpenRouterAPIKey
	}
}

// getModelName reads a per-tier model name from system_settings.
// Falls back to the provided default if the key is missing or empty.
//
// WHY this method exists:
//
//	CodeCraftAPI exposes a live model catalog. Admin selects which model
//	to use per tier from the admin panel. Model names are stored in
//	system_settings (e.g. key="codecraftapi_model_strong").
//	Same pattern as getActiveProvider() and getAPIKey().
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
//
// Multi-turn requests key on the whole conversation. Without this, two
// different conversations would collide: both leave SystemPrompt and
// UserPrompt empty, so the key would reduce to the model tier alone.
func (g *ModelGateway) cacheKey(req LLMRequest) string {
	if len(req.Messages) > 0 {
		var b strings.Builder
		b.WriteString(string(req.Model))
		for _, m := range req.Messages {
			b.WriteString("\x00")
			b.WriteString(m.Role)
			b.WriteString("\x00")
			b.WriteString(m.Content)
		}
		return b.String()
	}
	return fmt.Sprintf("%s:%s:%s", req.Model, req.SystemPrompt, req.UserPrompt)
}

// addWorkflowCost adds cost to workflows.cost_spent_usd for one workflow.
//
// No-op when workflowID is nil (non-workflow call) or the DB is not wired.
// Non-fatal on error: a monitoring write must never fail an LLM response
// the caller already paid for — matches how cost checks are treated in
// WorkflowRunner.executeWaves (logged, execution continues).
//
// WHY the column has to be maintained here: nothing wrote it before, so it
// stayed 0.0 forever. Two things depended on it —
//  1. the Kanban header, which showed "$0.0000 / $10.00" no matter what
//     was actually spent;
//  2. CostMonitor.CheckWorkflowLimits, which reads this exact column as
//     its "fast path" and therefore never saw a limit as reached, so the
//     per-workflow soft/hard cost caps could not fire at all.
func (g *ModelGateway) addWorkflowCost(ctx context.Context, workflowID *uuid.UUID, cost float64) {
	if workflowID == nil || g.db == nil || cost <= 0 {
		return
	}
	// context.WithoutCancel: the spend already happened. If the caller's
	// context is cancelled right after the provider replied, the money is
	// still gone and must still be recorded.
	if _, err := g.db.Exec(context.WithoutCancel(ctx),
		`UPDATE workflows SET cost_spent_usd = cost_spent_usd + $1, updated_at = NOW()
		 WHERE id = $2`,
		cost, *workflowID,
	); err != nil {
		g.logger.Warn("failed to attribute LLM cost to workflow",
			zap.String("workflow_id", workflowID.String()),
			zap.Float64("cost_usd", cost),
			zap.Error(err),
		)
	}
}
