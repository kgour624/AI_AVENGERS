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
	gateway   *gateway.ModelGateway
	clusterer *EmbeddingClusterer
	logger    *zap.Logger
}

// NewTopicExtractor creates a new topic extractor.
// NewTopicExtractor creates a new topic extractor.
// embedder satisfies ml.Embedder — passed through to EmbeddingClusterer.
func NewTopicExtractor(gw *gateway.ModelGateway, embedder ml.Embedder, logger *zap.Logger) *TopicExtractor {
	return &TopicExtractor{
		gateway:   gw,
		clusterer: NewEmbeddingClusterer(embedder, 0.75, logger),
		logger:    logger,
	}
}

// ExtractBatch extracts topics for all chunks.
// Uses embedding clustering to minimize LLM calls.
//
// Mental execution:
// 1000 chunks -> cluster by similarity -> 200 clusters
// -> 200 LLM calls (one per cluster representative)
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
