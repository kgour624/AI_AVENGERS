package selflearning

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/gateway"
)

// minTokensForProcessing is the minimum question length (in whitespace-split
// tokens) before self-learning processing is attempted.
// WHY 20: very short questions ("What is BFS?", "Explain recursion") are
// already domain-aligned — processing adds latency with no accuracy gain.
// The RAG vector search handles these correctly without extraction.
const minTokensForProcessing = 20

// ProcessedQuestion is the output of QuestionProcessor.Process.
// Original is NEVER modified — it is always the raw user question.
// Extracted is the domain-specific signal derived from Original.
// VerificationPassed=true means Extracted is safe to use for RAG.
// VerificationPassed=false means Original should be used instead.
type ProcessedQuestion struct {
	Original           string // raw user question — never modified
	Extracted          string // domain-specific signal for RAG
	UnderstandingLog   string // Step 1 output — used by Step 3 verify
	VerificationPassed bool   // true = use Extracted; false = use Original
	SkippedReason      string // non-empty when processing was skipped
}

// understandingOutput holds the parsed result of Step 1.
type understandingOutput struct {
	CoreProblem    string
	DomainConcepts string
	UserIntent     string
}

// QuestionProcessor implements the Self-Learning Mode:
// Understand → Extract → Verify.
//
// WHY this exists:
//   Raw user questions often contain story noise, platform-specific
//   language, or domain-agnostic phrasing that causes RAG vector search
//   to retrieve wrong chunks (or no chunks), triggering Gate 2 refusal.
//   This processor converts raw questions into domain-specific signal
//   BEFORE the question reaches Gate 1, improving retrieval accuracy
//   without changing any gate logic.
//
// Failure policy:
//   Any step failure → SkippedReason set → caller uses Original.
//   Pipeline is NEVER blocked by self-learning failures.
//
// Cost:
//   3 × ModelCheap calls per question per expert.
//   ~$0.0003 per question (DeepSeek cheap tier).
//   Acceptable: accuracy improvement justifies cost.
type QuestionProcessor struct {
	gateway *gateway.ModelGateway
	logger  *zap.Logger
}

// NewQuestionProcessor creates a new QuestionProcessor.
// gateway must be non-nil. logger must be non-nil.
func NewQuestionProcessor(gw *gateway.ModelGateway, logger *zap.Logger) *QuestionProcessor {
	return &QuestionProcessor{gateway: gw, logger: logger}
}

// Process runs the Understand → Extract → Verify pipeline on the raw question.
//
// expertName, expertDomain, reasoningCharter come from the domain expert.
// chunks are the already-retrieved course chunks (used for domain context).
//
// Mental execution:
//   Input: "Alice is on a chessboard. Knight moves. Find minimum moves."
//   Expert: DSA, domain="dsa", charter="Apply BFS, DP, Greedy patterns"
//
//   Step 1 (Understand):
//     CORE_PROBLEM: Find minimum moves for a knight on a chessboard
//     DOMAIN_CONCEPTS: BFS, shortest path, grid traversal, parity
//     USER_INTENT: Implement minimum knight moves algorithm
//
//   Step 2 (Extract):
//     "BFS shortest path on grid. Knight movement pattern (±1,±2).
//      Find minimum distance between two cells. Grid parity analysis."
//
//   Step 3 (Verify):
//     VERIFIED: YES
//     → VerificationPassed=true, Extracted used for RAG
//
//   RAG now retrieves BFS/grid chunks instead of chess story chunks.
func (p *QuestionProcessor) Process(
	ctx context.Context,
	question string,
	expertName string,
	expertDomain string,
	reasoningCharter string,
	chunks []chinawall.CourseChunk,
) *ProcessedQuestion {
	result := &ProcessedQuestion{
		Original: question,
	}

	// Skip heuristic: very short questions are already domain-aligned.
	// Counting whitespace-split tokens is O(n) but n is tiny (question length).
	if len(strings.Fields(question)) < minTokensForProcessing {
		result.SkippedReason = fmt.Sprintf(
			"question too short (%d tokens < %d threshold)",
			len(strings.Fields(question)), minTokensForProcessing,
		)
		p.logger.Debug("self-learning skipped: short question",
			zap.String("expert", expertName),
			zap.String("reason", result.SkippedReason),
		)
		return result
	}

	// ============================================================
	// STEP 1: UNDERSTAND
	// Expert reads problem through domain lens.
	// Output: structured CORE_PROBLEM / DOMAIN_CONCEPTS / USER_INTENT.
	// ============================================================
	understanding, err := p.stepUnderstand(ctx, question, expertName, expertDomain, reasoningCharter)
	if err != nil {
		result.SkippedReason = "step1_understand_failed: " + err.Error()
		p.logger.Warn("self-learning step 1 failed — using original question",
			zap.String("expert", expertName),
			zap.Error(err),
		)
		return result
	}
	result.UnderstandingLog = formatUnderstandingLog(understanding)

	// ============================================================
	// STEP 2: EXTRACT
	// Rewrite question using only domain-relevant signal.
	// Story noise removed, domain technical terms injected.
	// Constraints and edge cases always preserved.
	// ============================================================
	extracted, err := p.stepExtract(ctx, question, expertName, expertDomain, understanding)
	if err != nil {
		result.SkippedReason = "step2_extract_failed: " + err.Error()
		p.logger.Warn("self-learning step 2 failed — using original question",
			zap.String("expert", expertName),
			zap.Error(err),
		)
		return result
	}
	if strings.TrimSpace(extracted) == "" {
		result.SkippedReason = "step2_extract_empty"
		p.logger.Warn("self-learning step 2 returned empty extraction — using original question",
			zap.String("expert", expertName),
		)
		return result
	}
	result.Extracted = extracted

	// ============================================================
	// STEP 3: VERIFY
	// Check extracted version against original.
	// Ensures no important constraints or edge cases were dropped.
	// On NO: VerificationPassed=false → caller uses Original.
	// ============================================================
	verified, missing, err := p.stepVerify(ctx, question, extracted)
	if err != nil {
		// Verify failed — safe fallback: use original.
		result.SkippedReason = "step3_verify_failed: " + err.Error()
		p.logger.Warn("self-learning step 3 failed — using original question",
			zap.String("expert", expertName),
			zap.Error(err),
		)
		return result
	}

	if !verified {
		// Verification said NO — extraction dropped something important.
		// Use original question. Log missing parts for charter improvement.
		result.SkippedReason = "step3_verify_no: missing=" + missing
		p.logger.Warn("self-learning verification failed — using original question",
			zap.String("expert", expertName),
			zap.String("missing", missing),
		)
		return result
	}

	// All 3 steps passed. Extracted question is safe for RAG.
	result.VerificationPassed = true
	p.logger.Info("self-learning processing complete",
		zap.String("expert", expertName),
		zap.String("domain", expertDomain),
		zap.Int("original_tokens", len(strings.Fields(question))),
		zap.Int("extracted_tokens", len(strings.Fields(extracted))),
	)
	return result
}

// stepUnderstand runs Step 1: Expert reads problem through domain lens.
//
// Mental execution:
//   Input: chess problem, DSA expert
//   System prompt: "You are DSA Expert. Read and UNDERSTAND. Do NOT solve."
//   Output: CORE_PROBLEM: Find min knight moves / DOMAIN_CONCEPTS: BFS, grid
//   Parse → understandingOutput{CoreProblem: "...", DomainConcepts: "BFS, grid"}
func (p *QuestionProcessor) stepUnderstand(
	ctx context.Context,
	question string,
	expertName string,
	expertDomain string,
	reasoningCharter string,
) (*understandingOutput, error) {
	systemPrompt := fmt.Sprintf(`You are %s, a domain expert in %s.

Your ONLY job right now is to READ and UNDERSTAND the following question.
Do NOT solve it. Do NOT answer it. Do NOT write any code.

Read it carefully and tell me:
1. What is the CORE problem being asked? (1-2 sentences, domain-specific)
2. Which concepts from your domain (%s) are relevant here? (comma-separated technical terms)
3. What is the user actually trying to learn or solve? (1 sentence)

Be specific to YOUR domain. Ignore parts that are not relevant to %s.

REASONING CHARTER:
%s

Output EXACTLY in this format (no extra text):
CORE_PROBLEM: [1-2 sentences]
DOMAIN_CONCEPTS: [comma-separated technical terms]
USER_INTENT: [1 sentence]`,
		expertName, expertDomain, expertDomain, expertDomain, reasoningCharter)

	resp, err := p.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: systemPrompt,
		UserPrompt:   question,
		MaxTokens:    200,
		Temperature:  0.1, // low temperature: deterministic understanding
	})
	if err != nil {
		return nil, fmt.Errorf("understand LLM call failed: %w", err)
	}

	parsed := parseUnderstandingOutput(resp.Content)
	if parsed.CoreProblem == "" {
		return nil, fmt.Errorf("understand output missing CORE_PROBLEM field")
	}
	return parsed, nil
}

// stepExtract runs Step 2: Rewrite question using only domain-relevant signal.
//
// Mental execution:
//   Input: chess problem + understanding (BFS, grid, parity)
//   System prompt: "Rewrite using domain terms. Remove story. Keep constraints."
//   Output: "BFS shortest path on grid. Knight movement (±1,±2). Min distance."
func (p *QuestionProcessor) stepExtract(
	ctx context.Context,
	question string,
	expertName string,
	expertDomain string,
	understanding *understandingOutput,
) (string, error) {
	systemPrompt := fmt.Sprintf(`You are %s, a domain expert in %s.

Based on your understanding of the problem:
CORE_PROBLEM: %s
DOMAIN_CONCEPTS: %s
USER_INTENT: %s

Now rewrite the question using ONLY the domain-relevant parts.

Rules:
1. Use technical terms from %s — NOT story language or character names
2. Keep ALL numerical constraints and limits (they are ALWAYS signal, never noise)
3. Keep ALL edge cases mentioned (large inputs, empty inputs, special cases)
4. Keep ALL performance requirements (time complexity, latency, scale)
5. Remove story characters, platform instructions, irrelevant narrative
6. If a concept maps to a known %s term, use that term
7. If the problem has multiple parts, keep all parts

Output ONLY the rewritten question. No explanation. No preamble. No "Here is..."`,
		expertName, expertDomain,
		understanding.CoreProblem,
		understanding.DomainConcepts,
		understanding.UserIntent,
		expertDomain, expertDomain)

	resp, err := p.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: systemPrompt,
		UserPrompt:   question,
		MaxTokens:    300,
		Temperature:  0.1, // low temperature: deterministic extraction
	})
	if err != nil {
		return "", fmt.Errorf("extract LLM call failed: %w", err)
	}

	return strings.TrimSpace(resp.Content), nil
}

// stepVerify runs Step 3: Check extracted version against original.
//
// Returns (verified bool, missing string, err error).
// verified=true means extraction is safe to use.
// verified=false means something important was dropped — use original.
//
// Mental execution:
//   Input: original chess problem + extracted BFS version
//   System prompt: "Does extracted preserve ALL important info?"
//   Output: "VERIFIED: YES" → verified=true, missing=""
//   OR: "VERIFIED: NO\nMISSING: starting position constraint" → verified=false
func (p *QuestionProcessor) stepVerify(
	ctx context.Context,
	original string,
	extracted string,
) (verified bool, missing string, err error) {
	systemPrompt := `You are a verification agent.

Check if the EXTRACTED VERSION preserves ALL important information from the ORIGINAL QUESTION.

Specifically check:
1. Are all numerical constraints preserved? (counts, limits, thresholds, sizes)
2. Are all edge cases mentioned? (large inputs, empty inputs, special cases, overflow)
3. Is the core problem still the same?
4. Are any domain-critical details missing?
5. Are all performance requirements preserved? (time complexity, latency, scale)

Answer with ONLY one of these two formats:

Format 1 (if everything is preserved):
VERIFIED: YES

Format 2 (if something important is missing):
VERIFIED: NO
MISSING: [specific list of what is missing]`

	userPrompt := fmt.Sprintf("ORIGINAL QUESTION:\n%s\n\nEXTRACTED VERSION:\n%s",
		original, extracted)

	resp, err := p.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    100,
		Temperature:  0.0, // zero temperature: binary yes/no decision
	})
	if err != nil {
		return false, "", fmt.Errorf("verify LLM call failed: %w", err)
	}

	return parseVerifyOutput(resp.Content)
}

// ============================================================
// PARSING HELPERS
// ============================================================

// parseUnderstandingOutput parses the structured output of Step 1.
// Tolerates minor formatting variations (extra spaces, different casing).
//
// Mental execution:
//   Input: "CORE_PROBLEM: Find min knight moves\nDOMAIN_CONCEPTS: BFS, grid\nUSER_INTENT: Implement"
//   Output: understandingOutput{CoreProblem: "Find min knight moves", ...}
func parseUnderstandingOutput(raw string) *understandingOutput {
	out := &understandingOutput{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "CORE_PROBLEM:"); ok {
			out.CoreProblem = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "DOMAIN_CONCEPTS:"); ok {
			out.DomainConcepts = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "USER_INTENT:"); ok {
			out.UserIntent = strings.TrimSpace(after)
		}
	}
	return out
}

// parseVerifyOutput parses the binary YES/NO output of Step 3.
//
// Mental execution:
//   Input: "VERIFIED: YES" → (true, "", nil)
//   Input: "VERIFIED: NO\nMISSING: starting position" → (false, "starting position", nil)
//   Input: "something unexpected" → (false, "", nil) — safe fallback
func parseVerifyOutput(raw string) (verified bool, missing string, err error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VERIFIED:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "VERIFIED:"))
			if strings.EqualFold(val, "YES") {
				return true, "", nil
			}
			// VERIFIED: NO — continue to find MISSING line
		}
		if strings.HasPrefix(line, "MISSING:") {
			missing = strings.TrimSpace(strings.TrimPrefix(line, "MISSING:"))
		}
	}
	// If we reach here: either VERIFIED: NO was found (missing may be set)
	// or output was unexpected (safe fallback: treat as not verified).
	return false, missing, nil
}

// formatUnderstandingLog formats the understanding output as a single string
// for logging and for passing to Step 2's system prompt.
func formatUnderstandingLog(u *understandingOutput) string {
	return fmt.Sprintf("CORE_PROBLEM: %s\nDOMAIN_CONCEPTS: %s\nUSER_INTENT: %s",
		u.CoreProblem, u.DomainConcepts, u.UserIntent)
}
