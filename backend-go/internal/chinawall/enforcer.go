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
// DOMAIN BEHAVIOR:
//   Each layer's behavior is controlled by DomainProfile, not hardcoded booleans.
//   registry.Get(expertDomain) → *DomainProfile → controls all 4 layers.
//   Unknown domain → BaseProfile (strict defaults, safe fallback).
//
// CONFLICT RESOLUTION (base wall safety net):
//   Domain rules apply first.
//   IF Layer 4 output is empty AND profile.StripMode == FULL_STRIP:
//     → retry with BaseProfile (domain rules over-relaxed the wall)
//   IF profile.StripMode == CODE_EXEMPT AND output has code block:
//     → valid output, domain rules win
//
// WHY 4 layers:
// Single layer is not enough. LLMs are trained to be helpful
// and will hallucinate even when told not to.
// Each layer catches what the previous missed.
type Enforcer struct {
	cfg      config.ChinaWallConfig
	gateway  *gateway.ModelGateway
	ml       *ml.SidecarClient
	logger   *zap.Logger
	registry *DomainRegistry
}

// NewEnforcer creates a new China Wall enforcer.
// registry must be initialized (Init called) before passing here.
func NewEnforcer(cfg config.ChinaWallConfig, gw *gateway.ModelGateway, mlClient *ml.SidecarClient, logger *zap.Logger, registry *DomainRegistry) *Enforcer {
	return &Enforcer{cfg: cfg, gateway: gw, ml: mlClient, logger: logger, registry: registry}
}

// Enforce runs all 4 layers for a question + chunks.
// Returns success with cited answer, or refusal with explanation.
func (e *Enforcer) Enforce(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	expertDomain string,
	reasoningCharter string,
	attempt int,
) (*EnforceResult, error) {
	// Load domain profile from registry.
	// O(1) lookup. Falls back to BaseProfile for unknown domains.
	// WHY registry not hardcoded: domain behavior is DB-driven, no redeploy needed.
	profile := e.registry.Get(expertDomain)

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
	// Behavior controlled by profile.CoverageMode:
	//   APPLY_PRINCIPLES: can expert apply principles to solve? (DSA, coding)
	//   LITERAL_MATCH:    do chunks contain the answer? (medical, legal, finance)
	coverage, err := e.checkCoverage(ctx, question, chunks, profile.CoverageMode)
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
	// Behavior controlled by profile.CitationMode and profile.SystemPromptExt.
	generated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, profile)
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
	// Behavior controlled by profile.StripMode:
	//   CODE_EXEMPT: fenced code blocks kept unconditionally (DSA, coding)
	//   FULL_STRIP:  every sentence without citation is stripped (medical, legal)
	cleanAnswer, strippedCount := e.stripUncited(generated.Answer, profile.StripMode)

	if strippedCount > 0 {
		e.logger.Warn("Layer 4: stripped uncited claims",
			zap.Int("count", strippedCount),
		)
	}

	// CONFLICT RESOLUTION: Base wall safety net.
	// IF domain rules produced empty output AND domain is strict (FULL_STRIP):
	//   → domain rules over-relaxed something upstream, retry with BaseProfile.
	// IF domain is CODE_EXEMPT AND output has a code block:
	//   → valid DSA answer, domain rules win even with no prose.
	// WHY structural not LLM: no extra cost, no circular judgment.
	if strings.TrimSpace(cleanAnswer) == "" {
		if profile.StripMode == StripModeCodeExempt && strings.Contains(generated.Answer, "```") {
			// Code block present but strip removed prose — restore full answer.
			// This means citations were in code (wrong) — keep answer as-is.
			cleanAnswer = generated.Answer
			e.logger.Warn("Layer 4: output empty but code block present, restoring",
				zap.String("domain", profile.Domain),
			)
		} else if profile.Domain != BaseProfile.Domain {
			// Domain rules produced empty output — safety net: retry with BaseProfile.
			e.logger.Warn("Layer 4: domain rules produced empty output, retrying with BaseProfile",
				zap.String("domain", profile.Domain),
			)
			baseGenerated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, BaseProfile)
			if err != nil {
				return nil, fmt.Errorf("base profile generation failed: %w", err)
			}
			cleanAnswer, _ = e.stripUncited(baseGenerated.Answer, BaseProfile.StripMode)
			if strings.TrimSpace(cleanAnswer) == "" {
				return e.buildRefusal("empty_after_strip", "Could not generate a properly cited answer"), nil
			}
			return &EnforceResult{
				Status:     "success",
				Answer:     cleanAnswer,
				Citations:  baseGenerated.Citations,
				Coverage:   coverage,
				Confidence: float64(bestScore),
			}, nil
		} else {
			return e.buildRefusal("empty_after_strip", "Could not generate a properly cited answer"), nil
		}
	}

	return &EnforceResult{
		Status:     "success",
		Answer:     cleanAnswer,
		Citations:  generated.Citations,
		Coverage:   coverage,
		Confidence: float64(bestScore),
	}, nil
}

// problemSolvingDomains: expert domains where principle-application
// mode is correct. Set by admin at expert creation. Stable per expert.
// WHY domain not question: same question to different experts = different behavior.
var problemSolvingDomains = map[string]bool{
	"dsa": true, "algorithms": true, "coding": true,
	"programming": true, "computer_science": true,
	"software_engineering": true, "data_structures": true,
	"competitive_programming": true,
}

// IsProblemSolvingDomain is exported for use by decision/engine.go Gate 1.
func IsProblemSolvingDomain(domain string) bool {
	return problemSolvingDomains[strings.ToLower(domain)]
}

// internal alias for use within this package
func isProblemSolvingDomain(domain string) bool {
	return IsProblemSolvingDomain(domain)
}

// checkCoverage asks cheap LLM if chunks can answer the question.
// Uses Chain-of-Thought prompting (Byte by Byte AI course improvement).
//
// TWO MODES:
// 1. Problem-solving mode (DSA/coding): checks if expert can APPLY principles
//    to solve the problem. Correct for DSA — principles must transfer to new problems.
// 2. Factual mode (default): checks if chunks CONTAIN the answer.
//    Correct for medical/legal/domain-fact experts.
//
// WHY two modes:
//   A DSA expert trained on "two pointers, sliding window, prefix sum" SHOULD
//   solve "Longest Common Prefix" by applying string traversal principles.
//   Asking "does the transcript mention Longest Common Prefix?" is wrong —
//   it would refuse every new problem, defeating the purpose of DSA education.
func (e *Enforcer) checkCoverage(ctx context.Context, question string, chunks []CourseChunk, isProblemSolving bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("Question: " + question + "\n\nAvailable Knowledge:\n")
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

	if isProblemSolving {
		// Problem-solving mode: check if principles are applicable
		sb.WriteString(`
Think step by step:
1. What algorithmic concepts or data structures does this problem require?
2. What relevant concepts, techniques, or principles do the chunks contain?
3. Can the expert APPLY the knowledge in these chunks to solve this problem?
   (The exact problem does NOT need to be in the chunks — principles transfer)

After thinking, reply with ONLY one word: YES, PARTIAL, or NO

YES = chunks contain principles/techniques applicable to solve this problem
PARTIAL = chunks have related concepts but missing some key technique
NO = chunks are completely unrelated to this type of problem`)
	} else {
		// Factual mode: check if chunks contain the answer
		sb.WriteString(`
Think step by step:
1. What specific information does the question ask for?
2. What does each chunk cover?
3. Is there sufficient information to answer the question?

After thinking, reply with ONLY one word: YES, PARTIAL, or NO

YES = chunks contain all information needed
PARTIAL = chunks answer some parts but missing key details
NO = chunks do not cover this question`)
	}

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
	isProblemSolving bool,
) (*generatedAnswer, error) {

	// Build context with chunk IDs
	var contextSB strings.Builder
	for _, c := range chunks {
		contextSB.WriteString(fmt.Sprintf("[CHUNK_%s]\n%s\n\n", c.ID, c.Text))
	}

	var systemPrompt string
	if isProblemSolving {
		// Problem-solving mode: apply principles to solve new problems.
		// WHY different prompt: DSA education = learn principles, apply to new problems.
		// Strict "only cite chunks" would refuse every new LeetCode problem.
		systemPrompt = fmt.Sprintf(`You are %s, a domain expert in algorithms and data structures.

REASONING CHARTER:
%s

YOUR TRAINING MATERIAL (use these principles to solve problems):
%s

CRITICAL RULES:
1. Structure your answer as: Approach explanation → Code → Complexity analysis
2. In the EXPLANATION (before and after code): cite which principles you are applying
   Format: "Using the sliding window technique [CHUNK_uuid] we..."
3. Write COMPLETE, WORKING code in a fenced code block (java/python/etc.)
   The code block must be clean — NO [CHUNK_xxx] tokens inside the code
4. After the code: explain time and space complexity with citations
5. If a technique is NOT in your training material, say so explicitly
6. The code must be correct and runnable — this is the primary deliverable`,
			expertName, reasoningCharter, contextSB.String())
	} else {
		// Factual mode: strict grounding, only cite chunks.
		systemPrompt = fmt.Sprintf(`You are %s, a domain expert.

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
	}

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

// stripUncited removes sentences without citations, while preserving
// markdown/code structure.
//
// isProblemSolving: when true, entire fenced code blocks are kept
// unconditionally. Code is the APPLICATION of cited principles —
// it doesn't need per-line citations. The explanation before/after
// the code block cites which principles are being applied.
func (e *Enforcer) stripUncited(answer string, isProblemSolving bool) (string, int) {
	citationPattern := regexp.MustCompile(`\[CHUNK_[a-f0-9-]+\]`)
	sentences := splitSentences(answer)

	var clean []string
	stripped := 0
	inCodeBlock := false

	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)

		// Track fenced code block boundaries
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			clean = append(clean, sentence)
			continue
		}

		// Inside a code block: always keep (problem-solving mode)
		// or keep if protected (factual mode — protectCode already handled it)
		if inCodeBlock {
			clean = append(clean, sentence)
			continue
		}

		if trimmed == "" {
			// Blank-line marker — always keep for paragraph spacing.
			clean = append(clean, sentence)
			continue
		}

		hasCitation := citationPattern.MatchString(sentence)
		isShort := len(trimmed) < 25
		isStructural := isHeadingOrTransition(trimmed)

		if hasCitation || isShort || isStructural {
			clean = append(clean, sentence)
		} else if isProblemSolving {
			// In problem-solving mode, keep explanation lines even without
			// inline citations — the approach explanation is part of the answer.
			// Citations appear at the start of the explanation, not per-sentence.
			clean = append(clean, sentence)
		} else {
			stripped++
		}
	}

	return strings.Join(clean, ""), stripped
}

// buildRefusal creates a standardized refusal response.
func (e *Enforcer) buildRefusal(reason, message string) *EnforceResult {
	return &EnforceResult{
		Status: "refused",
		Reason: reason,
		Answer: message,
	}
}

// splitSentences splits text into sentence-like units for citation
// enforcement, WITHOUT destroying markdown or code (Bug 3.1 fix,
// docs bug list).
//
// The old implementation did strings.Split(text, ".") on the raw
// text, so every period ANYWHERE - inside code (fmt.Println(),
// db.user.id), version numbers (v1.0), decimals (3.14), and across
// markdown headings/bullets/code fences - was treated as a sentence
// boundary, then stripUncited rejoined survivors with a single
// space, flattening all structure into one scrambled line.
//
// Fix, in order (Go's regexp/RE2 has no lookaround, so this is a
// two-pass extract-then-restore approach rather than one clever regex):
//  1. protectCode() replaces fenced ``` code blocks and inline `code`
//     spans with placeholder tokens FIRST, so nothing inside them is
//     ever split or touched.
//  2. Split on newlines before anything else, so each markdown line
//     (heading, bullet, blank line) stays intact as its own unit.
//  3. Within one non-blank line, split only on ". " (period
//     immediately followed by a space) - fmt.Println(), db.user.id,
//     v1.0, and 3.14 never have a space right after that internal
//     period, so none of them match; a real sentence boundary almost
//     always does.
//  4. Each returned sentence carries its own trailing separator
//     (". " mid-line, "\n" at end-of-line, or bare "\n" for a blank
//     line) so stripUncited can rejoin survivors with plain
//     concatenation and get the original structure back.
func splitSentences(text string) []string {
	protected, codeBlocks := protectCode(text)

	var sentences []string
	for _, line := range strings.Split(protected, "\n") {
		if strings.TrimSpace(line) == "" {
			// Blank line = paragraph break. Keep as its own unit so
			// paragraph spacing survives even if neighboring sentences
			// get stripped.
			sentences = append(sentences, "\n")
			continue
		}
		parts := strings.Split(line, ". ")
		for i, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			if i < len(parts)-1 {
				trimmed += ". "
			} else {
				trimmed += "\n"
			}
			sentences = append(sentences, restoreCode(trimmed, codeBlocks))
		}
	}
	return sentences
}

// protectCode replaces fenced code blocks (```...```) and inline code
// spans (`...`) with placeholder tokens so splitSentences never splits
// or rejoins their content. Returns the substituted text and a map
// from placeholder token -> original code text, to be reversed by
// restoreCode once each final sentence unit is decided.
func protectCode(text string) (string, map[string]string) {
	blocks := make(map[string]string)
	n := 0
	nextToken := func(match string) string {
		token := fmt.Sprintf("\x00CODE%d\x00", n)
		blocks[token] = match
		n++
		return token
	}

	fenced := regexp.MustCompile("(?s)```.*?```")
	out := fenced.ReplaceAllStringFunc(text, nextToken)

	inline := regexp.MustCompile("`[^`\n]+`")
	out = inline.ReplaceAllStringFunc(out, nextToken)

	return out, blocks
}

// restoreCode reverses protectCode's substitution for one sentence.
func restoreCode(sentence string, blocks map[string]string) string {
	for token, original := range blocks {
		sentence = strings.ReplaceAll(sentence, token, original)
	}
	return sentence
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
