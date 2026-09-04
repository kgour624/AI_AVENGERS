package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
)

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

// modelConfig holds per-model settings.
type modelConfig struct {
	Name              string
	CostPer1KInput    float64
	CostPer1KOutput   float64
	MaxTokens         int
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

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

// ModelGateway is the single interface for all LLM calls.
// Handles: model routing, cost tracking, caching, retry with backoff.
//
// WHY single gateway:
// All LLM calls go through one place — easy to add logging, cost limits,
// model switching, and caching without touching business logic.
type ModelGateway struct {
	cfg        config.LLMConfig
	httpClient *http.Client
	models     map[ModelType]modelConfig
	cache      sync.Map // key: hash(model+prompt) → *LLMResponse
	totalCost  atomic.Value // float64
	callCount  atomic.Int64
	logger     *zap.Logger
}

// NewModelGateway creates a new model gateway.
func NewModelGateway(cfg config.LLMConfig, logger *zap.Logger) *ModelGateway {
	g := &ModelGateway{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // LLM calls can be slow
		},
		logger: logger,
	}

	// Initialize total cost
	g.totalCost.Store(float64(0))

	// Model configurations
	g.models = map[ModelType]modelConfig{
		ModelCheap: {
			Name:            cfg.ModelCheap,
			CostPer1KInput:  0.0005,
			CostPer1KOutput: 0.0015,
			MaxTokens:       4096,
		},
		ModelStrong: {
			Name:            cfg.ModelStrong,
			CostPer1KInput:  0.003,
			CostPer1KOutput: 0.015,
			MaxTokens:       8192,
		},
		ModelFast: {
			Name:            cfg.ModelFast,
			CostPer1KInput:  0.00025,
			CostPer1KOutput: 0.00075,
			MaxTokens:       8192,
		},
	}

	return g
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
func (g *ModelGateway) Call(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	// Check cache
	if req.UseCache {
		cacheKey := g.cacheKey(req)
		if cached, ok := g.cache.Load(cacheKey); ok {
			result := cached.(*LLMResponse)
			result.Cached = true
			return result, nil
		}
	}

	// Get model config
	modelCfg, ok := g.models[req.Model]
	if !ok {
		return nil, fmt.Errorf("unknown model type: %s", req.Model)
	}

	// Set max tokens
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2000 // Sensible default
	}
	if maxTokens > modelCfg.MaxTokens {
		maxTokens = modelCfg.MaxTokens
	}

	// Build messages
	messages := []openRouterMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: req.UserPrompt,
	})

	// Retry with exponential backoff
	// WHY 3 retries: LLM APIs have transient failures (rate limits, timeouts)
	// Exponential backoff: 1s, 2s, 4s — avoids hammering a struggling API
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
		openRouterResp, err := g.callOpenRouter(ctx, modelCfg.Name, messages, maxTokens, req.Temperature)
		if err != nil {
			lastErr = err
			g.logger.Warn("LLM call attempt failed",
				zap.Int("attempt", attempt+1),
				zap.String("model", string(req.Model)),
				zap.Error(err),
			)
			continue
		}

		duration := time.Since(start)

		// Calculate cost
		cost := g.calculateCost(modelCfg,
			openRouterResp.Usage.PromptTokens,
			openRouterResp.Usage.CompletionTokens,
		)

		// Update stats
		currentCost := g.totalCost.Load().(float64)
		g.totalCost.Store(currentCost + cost)
		g.callCount.Add(1)

		result := &LLMResponse{
			Content:      openRouterResp.Choices[0].Message.Content,
			InputTokens:  openRouterResp.Usage.PromptTokens,
			OutputTokens: openRouterResp.Usage.CompletionTokens,
			CostUSD:      cost,
			ModelUsed:    modelCfg.Name,
			DurationMs:   float64(duration.Milliseconds()),
		}

		g.logger.Info("LLM call complete",
			zap.String("model", string(req.Model)),
			zap.Int("input_tokens", result.InputTokens),
			zap.Int("output_tokens", result.OutputTokens),
			zap.Float64("cost_usd", cost),
			zap.Float64("duration_ms", result.DurationMs),
		)

		// Cache result
		if req.UseCache {
			cacheKey := g.cacheKey(req)
			g.cache.Store(cacheKey, result)
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

// callOpenRouter makes the actual HTTP call to OpenRouter.
func (g *ModelGateway) callOpenRouter(
	ctx context.Context,
	modelName string,
	messages []openRouterMessage,
	maxTokens int,
	temperature float64,
) (*openRouterResponse, error) {
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
