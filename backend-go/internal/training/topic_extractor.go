package training

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

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
