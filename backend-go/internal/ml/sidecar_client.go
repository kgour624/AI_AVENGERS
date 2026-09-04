package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
)

// SidecarClient is the Go client for the Python ML sidecar service.
// Handles: embeddings (bge-base-en-v1.5) and reranking (bge-reranker-base).
//
// WHY a separate Python sidecar:
// sentence-transformers is Python-native. Go ONNX bindings exist but
// are unstable in production. ML sidecar is stateless — horizontally scalable.
// Internal HTTP call adds ~2ms overhead — acceptable for ML operations.
type SidecarClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// EmbedRequest is the request body for the /embed endpoint.
type EmbedRequest struct {
	Texts []string `json:"texts"`
}

// EmbedResponse is the response from the /embed endpoint.
// Embeddings are 768-dimensional float32 vectors (bge-base-en-v1.5).
type EmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Model      string      `json:"model"`
	DurationMs float64     `json:"duration_ms"`
}

// RerankRequest is the request body for the /rerank endpoint.
type RerankRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopK      int      `json:"top_k"`
}

// RerankResult is a single reranked document with its score.
type RerankResult struct {
	Index int     `json:"index"`  // Original index in documents slice
	Score float32 `json:"score"`  // Relevance score (0.0 to 1.0)
	Text  string  `json:"text"`   // Document text
}

// RerankResponse is the response from the /rerank endpoint.
type RerankResponse struct {
	Results    []RerankResult `json:"results"`
	Model      string         `json:"model"`
	DurationMs float64        `json:"duration_ms"`
}

// NewSidecarClient creates a new ML sidecar client.
func NewSidecarClient(cfg config.MLConfig, logger *zap.Logger) *SidecarClient {
	return &SidecarClient{
		baseURL: cfg.SidecarURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
		logger: logger,
	}
}

// Embed generates 768D embeddings for a batch of texts.
// Uses bge-base-en-v1.5 model (local, zero cost).
//
// Mental execution:
// Input: ["How does sharding work?", "Consistent hashing explanation"]
// Output: [[0.1, 0.2, ...768 floats], [0.3, 0.1, ...768 floats]]
// Error cases: empty texts, sidecar down, timeout
func (c *SidecarClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	reqBody := EmbedRequest{Texts: texts}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embed request: %w", err)
	}

	resp, err := c.post(ctx, "/embed", body)
	if err != nil {
		return nil, fmt.Errorf("embed request failed: %w", err)
	}
	defer resp.Body.Close()

	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embed response: %w", err)
	}

	if len(embedResp.Embeddings) != len(texts) {
		return nil, fmt.Errorf(
			"embedding count mismatch: expected %d, got %d",
			len(texts), len(embedResp.Embeddings),
		)
	}

	c.logger.Debug("embeddings generated",
		zap.Int("count", len(texts)),
		zap.Float64("duration_ms", embedResp.DurationMs),
	)

	return embedResp.Embeddings, nil
}

// EmbedSingle generates an embedding for a single text.
// Convenience wrapper around Embed.
func (c *SidecarClient) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := c.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// Rerank scores documents by relevance to a query.
// Uses bge-reranker-base (cross-encoder, more accurate than bi-encoder).
// Returns top-K results sorted by score descending.
//
// WHY reranking after semantic search:
// Semantic search (cosine similarity) is fast but approximate.
// Reranker is slower but more accurate — used to refine top-20 to top-5.
// This is the China Wall Layer 1 check.
//
// Mental execution:
// Input: query="sharding", docs=["consistent hashing...", "kafka topics...", ...], topK=5
// Output: [{index:0, score:0.87, text:"consistent hashing..."}, ...]
// Sorted by score DESC, only top-K returned
func (c *SidecarClient) Rerank(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error) {
	if len(documents) == 0 {
		return []RerankResult{}, nil
	}

	if topK <= 0 || topK > len(documents) {
		topK = len(documents)
	}

	reqBody := RerankRequest{
		Query:     query,
		Documents: documents,
		TopK:      topK,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rerank request: %w", err)
	}

	resp, err := c.post(ctx, "/rerank", body)
	if err != nil {
		return nil, fmt.Errorf("rerank request failed: %w", err)
	}
	defer resp.Body.Close()

	var rerankResp RerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&rerankResp); err != nil {
		return nil, fmt.Errorf("failed to decode rerank response: %w", err)
	}

	c.logger.Debug("reranking complete",
		zap.Int("input_docs", len(documents)),
		zap.Int("top_k", topK),
		zap.Float64("duration_ms", rerankResp.DurationMs),
	)

	return rerankResp.Results, nil
}

// HealthCheck verifies the ML sidecar is running.
func (c *SidecarClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ML sidecar health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ML sidecar returned status %d", resp.StatusCode)
	}

	return nil
}

// post makes a POST request to the ML sidecar.
func (c *SidecarClient) post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ML sidecar returned status %d for %s", resp.StatusCode, path)
	}

	return resp, nil
}
