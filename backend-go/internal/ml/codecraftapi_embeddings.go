package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// CodeCraftAPIEmbedder generates embeddings via CodeCraftAPI's /v1/embeddings endpoint.
//
// CodeCraftAPI uses the standard OpenAI embeddings format:
//   POST /v1/embeddings
//   {"model": "...", "input": ["text1", "text2"]}
//   Response: {"data": [{"embedding": [...], "index": 0}, ...]}
//
// WHY model name is read from DB at call time (not stored in constructor):
//   Same pattern as ModelGateway.getActiveProvider() and DynamicEmbedder.
//   Admin changes embedding model in the UI — takes effect on the next
//   Embed() call with no server restart needed.
//   Constructor takes db *pgxpool.Pool to read embedding_model from
//   system_settings at call time.
//
// WHY *http.Client is required in constructor:
//   CodeCraftAPIEmbedder makes HTTP calls to CodeCraftAPI.
//   Caller (main.go) provides a client with 300s timeout — same as
//   SidecarClient — because batch embedding can be slow.
type CodeCraftAPIEmbedder struct {
	apiKey     string
	baseURL    string // trimmed of trailing slash
	db         *pgxpool.Pool // reads embedding_model from system_settings at call time
	httpClient *http.Client
	logger     *zap.Logger
}

// NewCodeCraftAPIEmbedder creates a new CodeCraftAPIEmbedder.
//
// apiKey: CodeCraftAPI API key (cc_...)
// baseURL: CodeCraftAPI base URL (default: https://codecraftapi.com/v1)
// db: used to read embedding_model from system_settings at call time
// httpClient: HTTP client with appropriate timeout (300s recommended)
// logger: for debug/warning logs
func NewCodeCraftAPIEmbedder(
	apiKey string,
	baseURL string,
	db *pgxpool.Pool,
	httpClient *http.Client,
	logger *zap.Logger,
) *CodeCraftAPIEmbedder {
	if baseURL == "" {
		baseURL = "https://codecraftapi.com/v1"
	}
	return &CodeCraftAPIEmbedder{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		db:         db,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Embed generates embeddings for a batch of texts via CodeCraftAPI.
//
// Reads embedding_model from system_settings at call time.
// Returns one embedding vector per input text, sorted by index.
//
// Mental execution:
//   Input: ["How to shard?", "Use consistent hashing"]
//   1. Read embedding_model from DB → "cc-embed-model-X"
//   2. POST /v1/embeddings {"model":"cc-embed-model-X","input":[...]}
//   3. Parse data[].embedding, sort by index
//   4. Return [[...768 floats], [...768 floats]]
func (e *CodeCraftAPIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Read model name from DB at call time
	model, err := e.readEmbeddingModel(ctx)
	if err != nil {
		return nil, err
	}
	if model == "" {
		return nil, fmt.Errorf("codecraftapi: embedding_model not configured in system_settings")
	}

	// Build request body (OpenAI embeddings format)
	body, err := json.Marshal(map[string]interface{}{
		"model": model,
		"input": texts,
	})
	if err != nil {
		return nil, fmt.Errorf("codecraftapi embeddings: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		e.baseURL+"/embeddings",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("codecraftapi embeddings: build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("codecraftapi embeddings: http call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("codecraftapi embeddings: status %d", resp.StatusCode)
	}

	// Parse OpenAI-compatible embeddings response
	var result embeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("codecraftapi embeddings: decode response: %w", err)
	}

	if len(result.Data) != len(texts) {
		return nil, fmt.Errorf(
			"codecraftapi embeddings: expected %d, got %d",
			len(texts), len(result.Data),
		)
	}

	// Sort by index (defensive — OpenAI spec guarantees order but sort is correct)
	sort.Slice(result.Data, func(i, j int) bool {
		return result.Data[i].Index < result.Data[j].Index
	})

	// Extract embedding vectors in sorted order
	embeddings := make([][]float32, len(result.Data))
	for i, item := range result.Data {
		embeddings[i] = item.Embedding
	}

	e.logger.Debug("codecraftapi embeddings generated",
		zap.Int("count", len(texts)),
		zap.String("model", model),
	)

	return embeddings, nil
}

// EmbedSingle generates an embedding for a single text.
// Convenience wrapper around Embed.
func (e *CodeCraftAPIEmbedder) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("codecraftapi: no embedding returned")
	}
	return embeddings[0], nil
}

// readEmbeddingModel reads the embedding_model value from system_settings.
// Returns empty string if not configured (caller handles this case).
func (e *CodeCraftAPIEmbedder) readEmbeddingModel(ctx context.Context) (string, error) {
	var valueJSON []byte
	err := e.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = 'embedding_model'`,
	).Scan(&valueJSON)
	if err != nil {
		// Row missing = not configured yet. Return empty string, caller returns actionable error.
		return "", nil
	}

	var model string
	if err := json.Unmarshal(valueJSON, &model); err != nil {
		return "", fmt.Errorf("codecraftapi: failed to parse embedding_model from system_settings: %w", err)
	}

	return model, nil
}

// embeddingsResponse is the OpenAI-compatible embeddings response format.
// Used by CodeCraftAPI's POST /v1/embeddings endpoint.
type embeddingsResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}
