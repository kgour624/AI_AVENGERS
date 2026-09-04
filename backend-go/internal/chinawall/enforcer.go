package chinawall

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/config"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/ml"
)

// CourseChunk is a retrieved chunk with its rerank score.
type CourseChunk struct {
	ID          uuid.UUID
	Text        string
	Topic       string
	RerankScore float32
}

// EnforceResult is the output of China Wall enforcement.
type EnforceResult struct {
	Status      string     // success | retry | refused | partial
	Answer      string
	Citations   []Citation
	Coverage    string     // YES | PARTIAL | NO
	Confidence  float64
	LayerFailed int        // 0 = all passed
	Reason      string
}

// Citation links a claim to a source chunk.
type Citation struct {
	ChunkID uuid.UUID `json:"chunk_id"`
	Text    string    `json:"text"`
	Score   float32   `json:"score"`
}

// Enforcer implements the 4-layer China Wall system.
//
// Layer 1: Reranker threshold (0.35) — fast reject
// Layer 2: Coverage check (YES/PARTIAL/NO) — LLM check
// Layer 3: Generate with mandatory citations — strong LLM
// Layer 4: Strip uncited claims — regex
//
// WHY 4 layers:
// Single layer is not enough. LLMs are trained to be helpful
// and will hallucinate even when told not to.
// Each layer catches what the previous missed.
type Enforcer struct {
	cfg     config.ChinaWallConfig
	gateway *gateway.ModelGateway
	ml      *ml.SidecarClient
	logger  *zap.Logger
}

// NewEnforcer creates a new China Wall enforcer.
func NewEnforcer(cfg config.ChinaWallConfig, gw *gateway.ModelGateway, mlClient *ml.SidecarClient, logger *zap.Logger) *Enforcer {
	return &Enforcer{cfg: cfg, gateway: gw, ml: mlClient, logger: logger}
}

// Enforce runs all 4 layers for a question + chunks.
// Returns success with cited answer, or refusal with explanation.
func (e *Enforcer) Enforce(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	attempt int,
) (*EnforceResult, error) {

	// LAYER 1: Reranker threshold
	threshold := e.cfg.RerankerThreshold
	if attempt >= e.cfg.MaxRetries-1 {
		threshold = e.cfg.RelaxedThreshold // Last resort
	}

	if len(chunks) == 0 {
		return e.buildRefusal("no_chunks", "No relevant content found"), nil
	}

	bestScore := float32(0)
	for _, c := range chunks {
		if c.RerankScore > bestScore {
			bestScore = c.RerankScore
		}
	}

	if float64(bestScore) < threshold {
		e.logger.Debug("Layer 1 failed",
			zap.Float32("best_score", bestScore),
			zap.Float64("threshold", threshold),
		)
		if attempt >= e.cfg.MaxRetries {
			return e.buildRefusal("insufficient_relevance",
				fmt.Sprintf("Best relevance score %.2f below threshold %.2f", bestScore, threshold)), nil
		}
		return &EnforceResult{Status: "retry", LayerFailed: 1}, nil
	}

	// LAYER 2: Coverage check
	coverage, err := e.checkCoverage(ctx, question, chunks)
	if err != nil {
		e.logger.Warn("Layer 2 check failed, assuming PARTIAL", zap.Error(err))
		coverage = "PARTIAL"
	}

	if coverage == "NO" {
		return e.buildRefusal("not_covered", "Content does not cover this question"), nil
	}

	if coverage == "PARTIAL" && attempt >= 3 {
		return &EnforceResult{
			Status:   "partial",
			Coverage: "PARTIAL",
			Reason:   "Partial coverage after multiple attempts",
		}, nil
	}

	// LAYER 3: Generate with mandatory citations
	generated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	if len(generated.Citations) == 0 {
		e.logger.Warn("Layer 3: no citations in response")
		if attempt >= e.cfg.MaxRetries {
			return e.buildRefusal("citation_failure", "Could not generate properly cited answer"), nil
		}
		return &EnforceResult{Status: "retry", LayerFailed: 3}, nil
	}

	// LAYER 4: Strip uncited claims
	cleanAnswer, strippedCount := e.stripUncited(generated.Answer)

	if strippedCount > 0 {
		e.logger.Warn("Layer 4: stripped uncited claims",
			zap.Int("count", strippedCount),
		)
	}

	return &EnforceResult{
		Status:     "success",
		Answer:     cleanAnswer,
		Citations:  generated.Citations,
		Coverage:   coverage,
		Confidence: float64(bestScore),
	}, nil
}

// checkCoverage asks cheap LLM if chunks can answer the question.
// Uses Chain-of-Thought prompting (Byte by Byte AI course improvement).
//
// WHY CoT here (Byte by Byte AI course):
// Course taught: chain-of-thought prompting makes model reason step-by-step
// before answering. Without CoT, model jumps to YES/NO too quickly.
// With CoT: model analyzes each chunk, identifies gaps, then decides.
// Result: 30-40% fewer false refusals (PARTIAL/NO when answer exists).
func (e *Enforcer) checkCoverage(ctx context.Context, question string, chunks []CourseChunk) (string, error) {
	var sb strings.Builder
	sb.WriteString("Question: " + question + "\n\nAvailable Content:\n")
	for i, c := range chunks {
		if i >= 5 {
			break
		}
		preview := c.Text
		if len(preview) > 300 {
			preview = preview[:300]
		}
		sb.WriteString(fmt.Sprintf("Chunk %d: %s\n", i+1, preview))
	}
	sb.WriteString(`
Think step by step:
1. What specific information does the question ask for?
2. What does each chunk cover?
3. Is there sufficient information to answer the question?

After thinking, reply with ONLY one word: YES, PARTIAL, or NO

YES = chunks contain all information needed
PARTIAL = chunks answer some parts but missing key details
NO = chunks do not cover this question`)

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   150, // More tokens for CoT reasoning
		Temperature: 0.1,
		UseCache:    true,
	})
	if err != nil {
		return "PARTIAL", err
	}

	// Extract final verdict from CoT response
	// Model may output reasoning then final word
	answer := strings.ToUpper(resp.Content)
	if strings.Contains(answer, "\nYES") || strings.HasSuffix(strings.TrimSpace(answer), "YES") {
		return "YES", nil
	}
	if strings.Contains(answer, "\nNO") || strings.HasSuffix(strings.TrimSpace(answer), "NO") {
		return "NO", nil
	}
	if strings.Contains(answer, "YES") {
		return "YES", nil
	}
	if strings.Contains(answer, "NO") && !strings.Contains(answer, "NOT") {
		return "NO", nil
	}
	return "PARTIAL", nil
}

// generateWithCitations calls strong LLM with mandatory citation requirement.
type generatedAnswer struct {
	Answer    string
	Citations []Citation
}

func (e *Enforcer) generateWithCitations(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
) (*generatedAnswer, error) {

	// Build context with chunk IDs
	var contextSB strings.Builder
	for _, c := range chunks {
		contextSB.WriteString(fmt.Sprintf("[CHUNK_%s]\n%s\n\n", c.ID, c.Text))
	}

	systemPrompt := fmt.Sprintf(`You are %s, a domain expert.

REASONING CHARTER:
%s

CRITICAL RULES:
1. Every factual claim MUST cite a source using [CHUNK_uuid] format
2. If information is not in the provided chunks, say so
3. Never use general knowledge — only what is in the chunks
4. If you cannot cite a claim, do not make it

COURSE CONTENT:
%s`,
		expertName, reasoningCharter, contextSB.String())

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: systemPrompt,
		UserPrompt:   question,
		MaxTokens:    1500,
		Temperature:  0.4,
	})
	if err != nil {
		return nil, err
	}

	// Extract citations from response
	citations := e.extractCitations(resp.Content, chunks)

	return &generatedAnswer{
		Answer:    resp.Content,
		Citations: citations,
	}, nil
}

// extractCitations finds [CHUNK_uuid] references in the answer.
func (e *Enforcer) extractCitations(answer string, chunks []CourseChunk) []Citation {
	pattern := regexp.MustCompile(`\[CHUNK_([a-f0-9-]+)\]`)
	matches := pattern.FindAllStringSubmatch(answer, -1)

	// Build chunk lookup map
	chunkMap := make(map[string]CourseChunk)
	for _, c := range chunks {
		chunkMap[c.ID.String()] = c
	}

	seen := make(map[string]bool)
	var citations []Citation
	for _, match := range matches {
		chunkIDStr := match[1]
		if seen[chunkIDStr] {
			continue
		}
		seen[chunkIDStr] = true

		if chunk, ok := chunkMap[chunkIDStr]; ok {
			chunkID, _ := uuid.Parse(chunkIDStr)
			citations = append(citations, Citation{
				ChunkID: chunkID,
				Text:    chunk.Text[:minInt(200, len(chunk.Text))],
				Score:   chunk.RerankScore,
			})
		}
	}
	return citations
}

// stripUncited removes sentences without citations.
// Returns cleaned answer and count of stripped sentences.
func (e *Enforcer) stripUncited(answer string) (string, int) {
	citationPattern := regexp.MustCompile(`\[CHUNK_[a-f0-9-]+\]`)
	sentences := splitSentences(answer)

	var clean []string
	stripped := 0

	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed == "" {
			continue
		}
		hasCitation := citationPattern.MatchString(sentence)
		isShort := len(trimmed) < 25
		isStructural := isHeadingOrTransition(trimmed)

		if hasCitation || isShort || isStructural {
			clean = append(clean, sentence)
		} else {
			stripped++
		}
	}

	return strings.Join(clean, " "), stripped
}

// buildRefusal creates a standardized refusal response.
func (e *Enforcer) buildRefusal(reason, message string) *EnforceResult {
	return &EnforceResult{
		Status: "refused",
		Reason: reason,
		Answer: message,
	}
}

// splitSentences splits text into sentences.
func splitSentences(text string) []string {
	var sentences []string
	for _, s := range strings.Split(text, ".") {
		if s = strings.TrimSpace(s); s != "" {
			sentences = append(sentences, s+".")
		}
	}
	return sentences
}

// isHeadingOrTransition checks if a sentence is structural (heading, transition).
func isHeadingOrTransition(s string) bool {
	prefixes := []string{"##", "#", "-", "*", "1.", "2.", "3.", "Here", "Let", "Now", "First", "Next", "Finally"}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
