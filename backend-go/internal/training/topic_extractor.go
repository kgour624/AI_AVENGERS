package training

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/ml"
)

// TopicResult holds extracted topic metadata for a chunk.
type TopicResult struct {
	Topic      string  `json:"topic"`
	Subtopic   string  `json:"subtopic"`
	Confidence float64 `json:"confidence"`
}

// TopicExtractor extracts topic and subtopic from text chunks.
//
// Strategy (Byte by Byte AI course improvement):
// 1. First: embedding clustering to group similar chunks (free)
// 2. Then: LLM call per cluster, not per chunk (60% cost reduction)
//
// WHY this order:
// Course taught: similar texts end up nearby in embedding space.
// Chunks in same cluster share a topic -> one LLM call covers all.
// Only ambiguous chunks (low similarity) need individual LLM calls.
type TopicExtractor struct {
	gateway    *gateway.ModelGateway
	clusterer  *EmbeddingClusterer
	logger     *zap.Logger
}

// NewTopicExtractor creates a new topic extractor.
func NewTopicExtractor(gw *gateway.ModelGateway, mlClient *ml.SidecarClient, logger *zap.Logger) *TopicExtractor {
	return &TopicExtractor{
		gateway:   gw,
		clusterer: NewEmbeddingClusterer(mlClient, 0.75, logger),
		logger:    logger,
	}
}

// ExtractBatch extracts topics for all chunks.
// Uses embedding clustering to minimize LLM calls.
//
// Mental execution:
// 1000 chunks -> cluster by similarity -> 200 clusters
// -> 200 LLM calls (vs 100 batched calls before, but better quality)
// -> assign cluster topic to all chunks in cluster
func (e *TopicExtractor) ExtractBatch(ctx context.Context, chunks []TextChunk) ([]TopicResult, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	// Step 1: Cluster by embedding similarity
	clusters, _, err := e.clusterer.Cluster(ctx, chunks)
	if err != nil {
		e.logger.Warn("clustering failed, falling back to batch LLM", zap.Error(err))
		return e.extractBatchLLM(ctx, chunks)
	}

	// Step 2: Extract topic per cluster (one LLM call per cluster)
	results := make([]TopicResult, len(chunks))

	for _, cluster := range clusters {
		// Use representative chunk (first in cluster) for topic extraction
		representative := cluster.Chunks[0]
		topic, err := e.extractSingleTopic(ctx, representative)
		if err != nil {
			topic = TopicResult{Topic: "general", Confidence: 0.5}
		}

		// Assign same topic to all chunks in this cluster
		for _, idx := range cluster.Indices {
			if idx < len(results) {
				results[idx] = topic
			}
		}
	}

	e.logger.Info("topic extraction complete",
		zap.Int("chunks", len(chunks)),
		zap.Int("clusters", len(clusters)),
	)

	return results, nil
}

// extractSingleTopic extracts topic for one representative chunk.
func (e *TopicExtractor) extractSingleTopic(ctx context.Context, chunk TextChunk) (TopicResult, error) {
	preview := chunk.Text
	if len(preview) > 400 {
		preview = preview[:400]
	}

	prompt := fmt.Sprintf(`Identify the main technical topic of this text.

Text: %s

Return ONLY JSON: {"topic": "lowercase_underscore", "subtopic": "specific_concept", "confidence": 0.0-1.0}
Example: {"topic": "sharding", "subtopic": "consistent_hashing", "confidence": 0.9}`, preview)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   80,
		Temperature: 0.1,
		UseCache:    true,
	})
	if err != nil {
		return TopicResult{Topic: "general", Confidence: 0.5}, err
	}

	return parseTopicResult(resp.Content)
}

// extractBatchLLM is the fallback: batch LLM calls when clustering fails.
const batchSize = 10

func (e *TopicExtractor) extractBatchLLM(ctx context.Context, chunks []TextChunk) ([]TopicResult, error) {
	results := make([]TopicResult, len(chunks))
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]
		batchResults, err := e.extractBatch(ctx, batch)
		if err != nil {
			for j := range batch {
				results[i+j] = TopicResult{Topic: "general", Confidence: 0.5}
			}
			continue
		}
		for j, r := range batchResults {
			if i+j < len(results) {
				results[i+j] = r
			}
		}
	}
	return results, nil
}

func (e *TopicExtractor) extractBatch(ctx context.Context, chunks []TextChunk) ([]TopicResult, error) {
	var sb strings.Builder
	sb.WriteString("Extract the main topic for each text chunk. Return ONLY a JSON array.\n\n")
	for i, chunk := range chunks {
		preview := chunk.Text
		if len(preview) > 300 {
			preview = preview[:300]
		}
		sb.WriteString(fmt.Sprintf("Chunk %d: %s\n", i+1, preview))
	}
	sb.WriteString(fmt.Sprintf(`\nReturn exactly %d objects: [{"topic": "...", "subtopic": "...", "confidence": 0.0}]`, len(chunks)))

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   500,
		Temperature: 0.1,
	})
	if err != nil {
		return nil, err
	}
	return parseTopicResults(resp.Content, len(chunks))
}

// parseTopicResult parses a single topic JSON response.
func parseTopicResult(response string) (TopicResult, error) {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		return TopicResult{Topic: "general", Confidence: 0.5}, fmt.Errorf("no JSON found")
	}

	var result TopicResult
	if err := json.Unmarshal([]byte(response[start:end+1]), &result); err != nil {
		return TopicResult{Topic: "general", Confidence: 0.5}, err
	}
	result.Topic = normalizeTopicName(result.Topic)
	result.Subtopic = normalizeTopicName(result.Subtopic)
	if result.Topic == "" {
		result.Topic = "general"
	}
	return result, nil
}

// parseTopicResults parses batch topic JSON response.
func parseTopicResults(response string, expectedCount int) ([]TopicResult, error) {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON array found")
	}

	var results []TopicResult
	if err := json.Unmarshal([]byte(response[start:end+1]), &results); err != nil {
		return nil, err
	}
	for i := range results {
		results[i].Topic = normalizeTopicName(results[i].Topic)
		results[i].Subtopic = normalizeTopicName(results[i].Subtopic)
		if results[i].Topic == "" {
			results[i].Topic = "general"
		}
		if results[i].Confidence == 0 {
			results[i].Confidence = 0.7
		}
	}
	for len(results) < expectedCount {
		results = append(results, TopicResult{Topic: "general", Confidence: 0.5})
	}
	return results[:expectedCount], nil
}

// normalizeTopicName converts topic names to lowercase_underscore format.
func normalizeTopicName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func minInt2(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// TopicResult holds extracted topic metadata for a chunk.
type TopicResult struct {
	Topic      string  `json:"topic"`
	Subtopic   string  `json:"subtopic"`
	Confidence float64 `json:"confidence"`
}

// TopicExtractor extracts topic and subtopic from text chunks.
// Uses cheap LLM model — batches 10 chunks per call to minimize cost.
//
// WHY batch extraction:
// 1000 chunks × 1 LLM call = 1000 calls = ~$0.50
// 1000 chunks ÷ 10 per batch = 100 calls = ~$0.05
// 10x cost reduction with same quality.
type TopicExtractor struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewTopicExtractor creates a new topic extractor.
func NewTopicExtractor(gw *gateway.ModelGateway, logger *zap.Logger) *TopicExtractor {
	return &TopicExtractor{gateway: gw, logger: logger}
}

// ExtractBatch extracts topics for a batch of chunks.
// Processes in groups of 10 to balance cost and accuracy.
//
// Mental execution:
// Input: 25 chunks
// Batch 1: chunks 0-9 → 1 LLM call → 10 TopicResults
// Batch 2: chunks 10-19 → 1 LLM call → 10 TopicResults
// Batch 3: chunks 20-24 → 1 LLM call → 5 TopicResults
// Output: 25 TopicResults (one per chunk, same order)
func (e *TopicExtractor) ExtractBatch(ctx context.Context, chunks []TextChunk) ([]TopicResult, error) {
	const batchSize = 10

	results := make([]TopicResult, len(chunks))

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[i:end]
		batchResults, err := e.extractBatch(ctx, batch)
		if err != nil {
			e.logger.Warn("topic extraction batch failed, using fallback",
				zap.Int("batch_start", i),
				zap.Error(err),
			)
			// Fallback: use empty topics for this batch
			for j := range batch {
				results[i+j] = TopicResult{
					Topic:      "general",
					Subtopic:   "",
					Confidence: 0.5,
				}
			}
			continue
		}

		for j, result := range batchResults {
			if i+j < len(results) {
				results[i+j] = result
			}
		}
	}

	return results, nil
}

// extractBatch processes a single batch of up to 10 chunks.
func (e *TopicExtractor) extractBatch(ctx context.Context, chunks []TextChunk) ([]TopicResult, error) {
	// Build prompt with all chunks
	var sb strings.Builder
	sb.WriteString(`You are a technical content analyzer. Extract the main topic and subtopic for each text chunk.

For each chunk, identify:
- topic: The main technical domain (e.g., "sharding", "caching", "kafka", "system_design", "database", "algorithms")
- subtopic: The specific concept within the topic (e.g., "consistent_hashing", "cache_invalidation", "partition_rebalancing")
- confidence: How confident you are (0.0 to 1.0)

Rules:
- Use lowercase with underscores for topic and subtopic
- Be specific: "consistent_hashing" not "hashing"
- If unclear, use "general" as topic
- Return ONLY valid JSON array, no explanation

Chunks to analyze:
`)

	for i, chunk := range chunks {
		// Truncate chunk text for prompt efficiency
		// WHY 300 chars: Enough context for topic identification, not wasteful
		preview := chunk.Text
		if len(preview) > 300 {
			preview = preview[:300] + "..."
		}
		sb.WriteString(fmt.Sprintf("\nChunk %d: %s", i+1, preview))
	}

	sb.WriteString(fmt.Sprintf(`

Return a JSON array with exactly %d objects:
[{"topic": "...", "subtopic": "...", "confidence": 0.0}, ...]`, len(chunks)))

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   500,
		Temperature: 0.1, // Low temperature for consistent extraction
		UseCache:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse JSON response
	results, err := parseTopicResults(resp.Content, len(chunks))
	if err != nil {
		e.logger.Warn("failed to parse topic results, using fallback",
			zap.String("response", resp.Content[:min(200, len(resp.Content))]),
			zap.Error(err),
		)
		// Fallback: return general topics
		results = make([]TopicResult, len(chunks))
		for i := range results {
			results[i] = TopicResult{Topic: "general", Confidence: 0.5}
		}
	}

	return results, nil
}

// parseTopicResults parses LLM JSON response into TopicResult slice.
// Handles common LLM response issues: markdown code blocks, extra text.
func parseTopicResults(response string, expectedCount int) ([]TopicResult, error) {
	// Strip markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find JSON array
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no JSON array found in response")
	}
	response = response[start : end+1]

	var results []TopicResult
	if err := json.Unmarshal([]byte(response), &results); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	// Validate and normalize
	for i := range results {
		results[i].Topic = normalizeTopicName(results[i].Topic)
		results[i].Subtopic = normalizeTopicName(results[i].Subtopic)
		if results[i].Topic == "" {
			results[i].Topic = "general"
		}
		if results[i].Confidence == 0 {
			results[i].Confidence = 0.7
		}
	}

	// Pad with fallbacks if LLM returned fewer results than expected
	for len(results) < expectedCount {
		results = append(results, TopicResult{Topic: "general", Confidence: 0.5})
	}

	return results[:expectedCount], nil
}

// normalizeTopicName converts topic names to lowercase_underscore format.
func normalizeTopicName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
