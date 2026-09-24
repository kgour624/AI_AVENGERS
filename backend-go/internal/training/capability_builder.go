package training

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// CapabilityResult holds analyzed capability for a single topic.
type CapabilityResult struct {
	Topic             string
	DepthLevel        int
	ChunkCount        int
	ComplexityCeiling string
	CanHandle         []string
	CannotHandle      []string
	ExampleQuestions  []string
}

// CapabilityBuilder analyzes course chunks to determine expert capabilities.
// Generates depth levels (1-5) per topic and what the expert can/cannot handle.
//
// WHY capability table matters:
// Without it, experts attempt to answer everything and fail.
// With it, Gate 2 (Knowledge Coverage) can quickly check if a question
// is even in the expert's domain before doing expensive semantic search.
type CapabilityBuilder struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewCapabilityBuilder creates a new capability builder.
func NewCapabilityBuilder(gw *gateway.ModelGateway, logger *zap.Logger) *CapabilityBuilder {
	return &CapabilityBuilder{gateway: gw, logger: logger}
}

// Build analyzes chunks grouped by topic and builds capability table.
//
// Mental execution:
// Input: 500 chunks with topics [sharding, caching, kafka, database]
// Step 1: Group chunks by topic
// Step 2: For each topic, calculate depth level
// Step 3: Extract can/cannot handle
// Step 4: Generate example questions
// Output: 4 CapabilityResults
func (b *CapabilityBuilder) Build(ctx context.Context, chunks []TextChunk, topics []TopicResult) ([]CapabilityResult, error) {
	if len(chunks) != len(topics) {
		return nil, fmt.Errorf("chunks and topics length mismatch: %d vs %d", len(chunks), len(topics))
	}

	// Group chunks by topic
	topicChunks := make(map[string][]TextChunk)
	for i, chunk := range chunks {
		topic := topics[i].Topic
		if topic == "" {
			topic = "general"
		}
		topicChunks[topic] = append(topicChunks[topic], chunk)
	}

	var results []CapabilityResult

	for topic, topicChunkList := range topicChunks {
		result, err := b.analyzeTopicCapability(ctx, topic, topicChunkList)
		if err != nil {
			b.logger.Warn("capability analysis failed for topic",
				zap.String("topic", topic),
				zap.Error(err),
			)
			// Use basic capability on failure
			result = b.basicCapability(topic, topicChunkList)
		}
		results = append(results, result)
	}

	// Sort by depth level descending (strongest topics first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].DepthLevel > results[j].DepthLevel
	})

	b.logger.Info("capability table built",
		zap.Int("topics", len(results)),
	)

	return results, nil
}

// analyzeTopicCapability analyzes a single topic's depth and capabilities.
func (b *CapabilityBuilder) analyzeTopicCapability(ctx context.Context, topic string, chunks []TextChunk) (CapabilityResult, error) {
	chunkCount := len(chunks)
	depthLevel := b.calculateDepthLevel(chunks)
	complexityCeiling := depthLevelToComplexity(depthLevel)

	// Build text sample for LLM analysis
	// Use first 3 chunks as representative sample
	sampleChunks := chunks
	if len(sampleChunks) > 3 {
		sampleChunks = sampleChunks[:3]
	}

	var sampleTexts []string
	for _, c := range sampleChunks {
		preview := c.Text
		if len(preview) > 200 {
			preview = preview[:200]
		}
		sampleTexts = append(sampleTexts, preview)
	}
	sample := strings.Join(sampleTexts, "\n---\n")

	// Ask LLM to identify specific capabilities
	canHandle, cannotHandle, err := b.extractCapabilities(ctx, topic, sample, depthLevel)
	if err != nil {
		// Non-fatal: use generated defaults
		canHandle = generateCanHandle(topic, depthLevel)
		cannotHandle = []string{}
	}

	exampleQuestions := generateExampleQuestions(topic, depthLevel)

	return CapabilityResult{
		Topic:             topic,
		DepthLevel:        depthLevel,
		ChunkCount:        chunkCount,
		ComplexityCeiling: complexityCeiling,
		CanHandle:         canHandle,
		CannotHandle:      cannotHandle,
		ExampleQuestions:  exampleQuestions,
	}, nil
}

// calculateDepthLevel determines depth level (1-5) based on content analysis.
//
// Depth levels:
// 1 - Basic: Definitions, high-level concepts (< 5 chunks)
// 2 - Intermediate: Common patterns, basic implementation (5-15 chunks)
// 3 - Advanced: Trade-offs, edge cases, optimization (15-30 chunks)
// 4 - Expert: Deep internals, complex scenarios (30-50 chunks)
// 5 - Master: Research-level, novel solutions (> 50 chunks)
//
// WHY chunk count as primary signal:
// More chunks = more content = deeper coverage.
// Content analysis adjusts for quality.
func (b *CapabilityBuilder) calculateDepthLevel(chunks []TextChunk) int {
	chunkCount := len(chunks)

	// Base level from chunk count
	baseLevel := 1
	switch {
	case chunkCount >= 50:
		baseLevel = 5
	case chunkCount >= 30:
		baseLevel = 4
	case chunkCount >= 15:
		baseLevel = 3
	case chunkCount >= 5:
		baseLevel = 2
	default:
		baseLevel = 1
	}

	// Adjust based on content depth keywords
	allText := strings.Builder{}
	for _, c := range chunks {
		allText.WriteString(strings.ToLower(c.Text))
		allText.WriteString(" ")
	}
	combined := allText.String()

	expertKeywords := []string{
		"internal implementation", "algorithm complexity", "distributed consensus",
		"lock-free", "memory model", "cache coherence", "linearizability",
	}
	advancedKeywords := []string{
		"trade-off", "optimization", "performance tuning", "edge case",
		"bottleneck", "profiling", "benchmark",
	}

	expertCount := 0
	for _, kw := range expertKeywords {
		if strings.Contains(combined, kw) {
			expertCount++
		}
	}

	advancedCount := 0
	for _, kw := range advancedKeywords {
		if strings.Contains(combined, kw) {
			advancedCount++
		}
	}

	// Boost level if expert content found
	if expertCount >= 3 && baseLevel < 5 {
		baseLevel = min(5, baseLevel+1)
	} else if advancedCount >= 3 && baseLevel < 4 {
		baseLevel = min(4, baseLevel+1)
	}

	return baseLevel
}

// extractCapabilities uses LLM to identify specific can/cannot handle items.
func (b *CapabilityBuilder) extractCapabilities(ctx context.Context, topic string, sample string, depthLevel int) ([]string, []string, error) {
	prompt := fmt.Sprintf(`Based on this course content about "%s" (depth level %d/5):

%s

List what this expert CAN and CANNOT handle.

Return JSON:
{"can_handle": ["specific capability 1", "specific capability 2"], "cannot_handle": ["limitation 1"]}

Be specific. Max 8 items each. Return ONLY JSON.`,
		topic, depthLevel, sample)

	resp, err := b.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  prompt,
		MaxTokens:   2048, // P2: was 400 — below the reasoning-model floor
		Temperature: 0.2,
	})
	if err != nil {
		return nil, nil, err
	}

	return parseCapabilityJSON(resp.Content)
}

// basicCapability generates capability without LLM (fallback).
func (b *CapabilityBuilder) basicCapability(topic string, chunks []TextChunk) CapabilityResult {
	depthLevel := b.calculateDepthLevel(chunks)
	return CapabilityResult{
		Topic:             topic,
		DepthLevel:        depthLevel,
		ChunkCount:        len(chunks),
		ComplexityCeiling: depthLevelToComplexity(depthLevel),
		CanHandle:         generateCanHandle(topic, depthLevel),
		CannotHandle:      []string{},
		ExampleQuestions:  generateExampleQuestions(topic, depthLevel),
	}
}

// depthLevelToComplexity maps depth level to complexity ceiling string.
func depthLevelToComplexity(level int) string {
	switch level {
	case 1:
		return "basic"
	case 2:
		return "intermediate"
	case 3:
		return "advanced"
	case 4:
		return "expert"
	case 5:
		return "master"
	default:
		return "intermediate"
	}
}

// generateCanHandle creates generic can-handle list based on depth.
func generateCanHandle(topic string, depth int) []string {
	base := []string{
		fmt.Sprintf("Explain what %s is and when to use it", topic),
		fmt.Sprintf("Describe common %s patterns", topic),
	}
	if depth >= 2 {
		base = append(base, fmt.Sprintf("Implement basic %s solutions", topic))
	}
	if depth >= 3 {
		base = append(base, fmt.Sprintf("Analyze %s trade-offs", topic))
		base = append(base, fmt.Sprintf("Optimize %s for specific use cases", topic))
	}
	if depth >= 4 {
		base = append(base, fmt.Sprintf("Debug complex %s issues", topic))
		base = append(base, fmt.Sprintf("Design %s for large-scale systems", topic))
	}
	return base
}

// generateExampleQuestions creates example questions based on depth.
func generateExampleQuestions(topic string, depth int) []string {
	switch {
	case depth <= 1:
		return []string{
			fmt.Sprintf("What is %s?", topic),
			fmt.Sprintf("When should I use %s?", topic),
		}
	case depth == 2:
		return []string{
			fmt.Sprintf("How do I implement %s?", topic),
			fmt.Sprintf("What are common %s patterns?", topic),
			fmt.Sprintf("Can you show a %s example?", topic),
		}
	default:
		return []string{
			fmt.Sprintf("What are the trade-offs of different %s approaches?", topic),
			fmt.Sprintf("How do I optimize %s for my use case?", topic),
			fmt.Sprintf("What are common %s pitfalls?", topic),
			fmt.Sprintf("How does %s work internally?", topic),
		}
	}
}

// parseCapabilityJSON parses LLM JSON response for capabilities.
func parseCapabilityJSON(response string) (canHandle []string, cannotHandle []string, err error) {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 {
		return nil, nil, fmt.Errorf("no JSON found")
	}
	response = response[start : end+1]

	var result struct {
		CanHandle    []string `json:"can_handle"`
		CannotHandle []string `json:"cannot_handle"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, nil, err
	}

	// Limit sizes
	if len(result.CanHandle) > 8 {
		result.CanHandle = result.CanHandle[:8]
	}
	if len(result.CannotHandle) > 5 {
		result.CannotHandle = result.CannotHandle[:5]
	}

	return result.CanHandle, result.CannotHandle, nil
}
