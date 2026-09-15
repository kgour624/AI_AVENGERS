package chinawall

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/category"
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
	// TemplateSections is set ONLY when the expert's category has a
	// non-empty template_schema (CT-B, CATEGORY_TEMPLATE_HANDOFF.md §4).
	// nil/empty for every flat-text expert (CT-L2 fallback) — Answer
	// is the only populated field in that case, exactly as before this
	// feature existed.
	TemplateSections []TemplateSectionResult
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
	// replyContext is a pre-formatted string produced by
	// orchestrator.formatReplyContext(). Empty string for every fresh
	// (non-reply) question — all downstream functions check for "" and
	// skip injection, so the flat-text and structured paths are
	// completely unaffected for the vast majority of calls (CT-L2).
	replyContext string,
	attempt int,
	// templateSections/defaultLanguage (CT-B) are nil/"" for every expert
	// with no category or an empty template_schema — the entire flat-text
	// code path below this point is untouched and behaves identically to
	// before this feature (CT-L2). Only decision/engine.go passes these
	// through from expert.TemplateSections/expert.DefaultLanguage.
	templateSections []category.TemplateSection,
	defaultLanguage string,
	// tokenCh: non-nil enables streaming for Gate 5 generation.
	// nil = blocking Call() (backward compatible, used by smoke test etc.).
	tokenCh chan<- string,
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
	generated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, templateSections, defaultLanguage, tokenCh)
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

	// STRUCTURED PATH (CT-B1/B2): category had a non-empty template_schema,
	// so generateWithCitations took the structured branch and returned
	// per-section results instead of one flat Answer. Handled entirely
	// separately from the flat path below so the flat path's existing
	// logic (including its BaseProfile safety net) stays byte-for-byte
	// unchanged for every non-categorized expert (CT-L2).
	if len(generated.TemplateSections) > 0 {
		return e.enforceStructured(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, generated, coverage, bestScore, templateSections, defaultLanguage)
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
			baseGenerated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, replyContext, BaseProfile, nil, "", tokenCh)
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

// enforceStructured runs Layer 4 (per-section) and the BaseProfile safety
// net for the structured (categorized-expert) path. Mirrors the flat
// path's conflict-resolution logic in Enforce() above, but scoped to
// sections instead of one Answer string:
//   - prose sections: existing per-domain StripMode logic (unchanged
//     stripUncited call), applied to that section's text only.
//   - code / test_cases sections: structurally exempt (CT-B2) — never
//     passed through stripUncited, same principle as StripModeCodeExempt
//     for whole answers, just correctly scoped per-section instead of
//     accidentally exempting an entire prose answer that merely
//     contains one code block.
//   - if every section ends up empty AND this is not already BaseProfile:
//     retry once with BaseProfile, same safety net rationale as the flat
//     path ("domain rules over-relaxed something upstream").
func (e *Enforcer) enforceStructured(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	replyContext string,
	profile *DomainProfile,
	generated *generatedAnswer,
	coverage string,
	bestScore float32,
	templateSections []category.TemplateSection,
	defaultLanguage string,
) (*EnforceResult, error) {
	strippedCount := 0
	anyContent := false
	for i := range generated.TemplateSections {
		s := &generated.TemplateSections[i]
		if s.Type == category.SectionTypeProse {
			clean, n := e.stripUncited(s.Content, profile.StripMode)
			s.Content = clean
			strippedCount += n
		}
		if strings.TrimSpace(s.Content) != "" {
			anyContent = true
		}
	}
	if strippedCount > 0 {
		e.logger.Warn("Layer 4 (structured): stripped uncited prose sections",
			zap.Int("count", strippedCount),
		)
	}

	if anyContent {
		return &EnforceResult{
			Status:           "success",
			TemplateSections: generated.TemplateSections,
			Citations:        generated.Citations,
			Coverage:         coverage,
			Confidence:       float64(bestScore),
		}, nil
	}

	if profile.Domain == BaseProfile.Domain {
		return e.buildRefusal("empty_after_strip", "Could not generate a properly cited structured answer"), nil
	}

	e.logger.Warn("Layer 4 (structured): all sections empty, retrying with BaseProfile",
		zap.String("domain", profile.Domain),
	)
	baseGenerated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, replyContext, BaseProfile, templateSections, defaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("base profile structured generation failed: %w", err)
	}
	stillEmpty := true
	for i := range baseGenerated.TemplateSections {
		s := &baseGenerated.TemplateSections[i]
		if s.Type == category.SectionTypeProse {
			s.Content, _ = e.stripUncited(s.Content, BaseProfile.StripMode)
		}
		if strings.TrimSpace(s.Content) != "" {
			stillEmpty = false
		}
	}
	if stillEmpty {
		return e.buildRefusal("empty_after_strip", "Could not generate a properly cited structured answer"), nil
	}
	return &EnforceResult{
		Status:           "success",
		TemplateSections: baseGenerated.TemplateSections,
		Citations:        baseGenerated.Citations,
		Coverage:         coverage,
		Confidence:       float64(bestScore),
	}, nil
}

// globalRegistry is set by the application at startup via SetGlobalRegistry.
// Used by IsProblemSolvingDomain for backward compatibility with decision/engine.go Gate 1.
// WHY global: Gate 1 is a package-level function, not a method on Enforcer.
var globalRegistry *DomainRegistry

// SetGlobalRegistry must be called once at startup after registry.Init().
// This allows IsProblemSolvingDomain to use the DB-backed registry.
func SetGlobalRegistry(r *DomainRegistry) {
	globalRegistry = r
}

// IsProblemSolvingDomain is exported for use by decision/engine.go Gate 1.
// Returns true if the domain's profile has Gate1Skip=true.
// Falls back to BaseProfile (Gate1Skip=false) if registry not set or domain unknown.
func IsProblemSolvingDomain(domain string) bool {
	if globalRegistry == nil {
		// Registry not initialized yet — safe default: do not skip Gate 1.
		return false
	}
	return globalRegistry.Get(domain).Gate1Skip
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
// checkCoverage asks cheap LLM if chunks can answer the question.
// Uses Chain-of-Thought prompting.
//
// CoverageModeApplyPrinciples: checks if expert can APPLY principles to solve.
//   Correct for DSA, coding — principles transfer to new problems.
// CoverageModeLiteralMatch: checks if chunks CONTAIN the answer.
//   Correct for medical, legal, finance — facts must be in transcript.
func (e *Enforcer) checkCoverage(ctx context.Context, question string, chunks []CourseChunk, mode CoverageMode) (string, error) {
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

	if mode == CoverageModeApplyPrinciples {
		// Principle-application mode: check if principles are applicable.
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
		// Literal match mode: check if chunks contain the answer.
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
	// TemplateSections is set ONLY by the structured branch (len(templateSections) > 0
	// on the call into generateWithCitations below). nil for every flat-text call —
	// callers must check len(TemplateSections) > 0, never assume Answer is unset.
	TemplateSections []TemplateSectionResult
}

// generateWithCitations calls strong LLM with mandatory citation requirement.
// Behavior controlled by profile.CitationMode and profile.SystemPromptExt.
//
// templateSections/defaultLanguage (CT-B1): when non-empty, this function
// takes the structured JSON branch (generateStructured) instead of the
// existing flat-text branch below. Flat-text branch is completely
// unmodified from before this feature — same variables, same prompts,
// same gateway call (CT-L2).
func (e *Enforcer) generateWithCitations(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	replyContext string,
	profile *DomainProfile,
	templateSections []category.TemplateSection,
	defaultLanguage string,
	tokenCh chan<- string,
) (*generatedAnswer, error) {

	// Build context with chunk IDs
	var contextSB strings.Builder
	for _, c := range chunks {
		contextSB.WriteString(fmt.Sprintf("[CHUNK_%s]\n%s\n\n", c.ID, c.Text))
	}

	if len(templateSections) > 0 {
		return e.generateStructured(ctx, question, chunks, expertName, reasoningCharter, replyContext, contextSB.String(), templateSections, defaultLanguage, profile)
	}

	return e.generateFlatText(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, contextSB.String(), tokenCh)
}

// generateFlatText is the ORIGINAL flat-text generation path, extracted
// unchanged from generateWithCitations so generateStructured can reuse
// it as a fallback (RCA 2026-09-08, see below) instead of duplicating
// its prompts. Behavior is byte-for-byte identical to before this
// extraction — same variables, same prompts, same gateway call, same
// 1500 MaxTokens (CT-L2: this path is what every non-categorized
// expert has always used and must keep using unchanged).
func (e *Enforcer) generateFlatText(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	replyContext string,
	profile *DomainProfile,
	contextText string,
	tokenCh chan<- string,
) (*generatedAnswer, error) {
	var systemPrompt string
	if profile.CitationMode == CitationModeLoose {
		// Loose citation mode: cite principles in explanation, code blocks exempt.
		// WHY: DSA/coding education = learn principles, apply to new problems.
		// Strict "only cite chunks" would refuse every new LeetCode problem.
		systemPrompt = fmt.Sprintf(`You are %s, a domain expert.

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
6. The code must be correct and runnable — this is the primary deliverable
%s`,
			expertName, reasoningCharter, contextText, profile.SystemPromptExt)
	} else {
		// Strict citation mode: every factual claim must have inline citation.
		// WHY: medical/legal/finance — no uncited claims allowed.
		systemPrompt = fmt.Sprintf(`You are %s, a domain expert.

REASONING CHARTER:
%s

CRITICAL RULES:
1. Every factual claim MUST cite a source using [CHUNK_uuid] format
2. If information is not in the provided chunks, say so
3. Never use general knowledge — only what is in the chunks
4. If you cannot cite a claim, do not make it
%s

COURSE CONTENT:
%s`,
			expertName, reasoningCharter, profile.SystemPromptExt, contextText)
	}

	// Inject reply context into system prompt when the user is replying
	// to a prior message. Empty string for fresh questions — no-op (CT-L2).
	if replyContext != "" {
		systemPrompt += "\n\n" + replyContext
	}

	// Gate 5 generation: streaming when tokenCh non-nil, blocking otherwise.
	if tokenCh != nil {
		rawTokenCh, rawRespCh, streamErr := e.gateway.StreamCall(ctx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			SystemPrompt: systemPrompt,
			UserPrompt:   question,
			MaxTokens:    resolveMaxTokens(profile.MaxTokensFlat, DefaultMaxTokensFlat),
			Temperature:  0.4,
		})
		if streamErr == nil {
			var sb strings.Builder
			for token := range rawTokenCh {
				sb.WriteString(token)
				select {
				case tokenCh <- token:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			<-rawRespCh // drain metadata channel
			content := sb.String()
			if content != "" {
				return &generatedAnswer{Answer: content, Citations: e.extractCitations(content, chunks)}, nil
			}
		}
		// StreamCall failed or returned empty — fall through to blocking call
		e.logger.Warn("streaming failed or empty, falling back to blocking call")
	}

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: systemPrompt,
		UserPrompt:   question,
		MaxTokens:    resolveMaxTokens(profile.MaxTokensFlat, DefaultMaxTokensFlat),
		Temperature:  0.4,
	})
	if err != nil {
		return nil, err
	}
	citations := e.extractCitations(resp.Content, chunks)
	return &generatedAnswer{Answer: resp.Content, Citations: citations}, nil
}

// generateStructured is the CT-B1 structured JSON generation branch.
// Called from generateWithCitations only when the expert's category has
// a non-empty template_schema. Builds a JSON-only prompt (template.go's
// buildStructuredPrompt), calls the strong model once, parses the result
// into per-section text (template.go's parseStructuredResponse), then
// extracts citations independently per section so Layer 4 can strip each
// prose section on its own (CT-B2) without touching code/test_cases text.
//
// Mental execution (fixed, post-RCA):
// sections = [pattern(prose), idea(prose), code(code), walkthrough(prose), test_cases(test_cases)]
// LLM returns valid JSON with all 5 keys -> parseStructuredResponse succeeds
//   -> 5 TemplateSectionResult entries, citations extracted per prose section
//   -> allCitations = union of pattern/idea/walkthrough citations (code/test_cases
//      sections are not expected to contain [CHUNK_xxx] tokens per the prompt's
//      rule #4, so extractCitations on them normally returns empty, which is
//      correct — not a bug).
//
// Edge case (RCA 2026-09-08, real production symptom — reported as:
// mode badge / confidence % / citations all render correctly, but the
// actual answer body is garbled or near-empty): a DSA-style question
// needs prose + a full working code block + 4 test-case buckets ALL
// inside one JSON object — this routinely exceeds MaxTokens and the
// model's response gets cut off mid-JSON. json.Unmarshal on a
// truncated object always fails. The PREVIOUS fix for this dumped the
// raw, half-formed JSON text straight into Answer as a "fallback" —
// this is not a real fallback, it is broken output disguised as one:
// stray braces/quotes/[CHUNK_xxx] tokens render as garbled or
// near-invisible markdown, while Layer 1-2's mode/confidence and
// Layer 3's extracted citations (independent metadata) still display
// correctly — producing exactly the reported symptom.
//
// REAL fix: when parsing fails, fall back to generateFlatText — the
// SAME reliable flat-prose+code generation path every non-categorized
// expert already uses. This produces a genuine, complete answer
// instead of surfacing a parser failure as content. The caller
// (Enforce) already treats a zero-TemplateSections, non-empty-Answer
// generatedAnswer as the flat path, so this transparently degrades
// the WHOLE response to flat-text for this one turn — not a
// half-structured, half-broken hybrid.
func (e *Enforcer) generateStructured(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	replyContext string,
	contextText string,
	sections []category.TemplateSection,
	defaultLanguage string,
	profile *DomainProfile,
) (*generatedAnswer, error) {
	if defaultLanguage == "" {
		defaultLanguage = "java" // CT-L5: hardcoded default, admin-overridable per category
	}

	systemPrompt := buildStructuredPrompt(expertName, reasoningCharter, contextText, sections, defaultLanguage, profile)

	// Inject reply context when the user is replying to a prior message.
	// Empty string for fresh questions — no-op (CT-L2).
	if replyContext != "" {
		systemPrompt += "\n\n" + replyContext
	}

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: systemPrompt,
		UserPrompt:   question,
		// Admin-configurable per domain (profile.MaxTokensStructured),
		// falls back to DefaultMaxTokensStructured (3500) when unset (0).
		// See DomainProfile's field comment for the 2026-09-08 RCA this
		// setting exists to prevent from recurring on a different domain.
		MaxTokens:   resolveMaxTokens(profile.MaxTokensStructured, DefaultMaxTokensStructured),
		Temperature: 0.4,
	})
	if err != nil {
		return nil, err
	}

	parsed, parseErr := parseStructuredResponse(resp.Content, sections)
	if parseErr != nil {
		e.logger.Warn("structured response parse failed (likely truncated JSON) — falling back to flat-text generation for this turn",
			zap.Error(parseErr),
			zap.Int("raw_response_length", len(resp.Content)),
		)
		return e.generateFlatText(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, contextText)
	}

	// 2026-09-08 RCA: initialized non-nil - a structured answer with
	// zero prose citations (e.g. only code/test_cases sections) must
	// not marshal this as JSON null (crashes frontend citations.map()).
	allCitations := []Citation{}
	seen := make(map[string]bool)
	resultSections := make([]TemplateSectionResult, 0, len(sections))
	for _, s := range sections {
		text := parsed[s.Key]
		secCitations := e.extractCitations(text, chunks)
		for _, c := range secCitations {
			key := c.ChunkID.String()
			if !seen[key] {
				seen[key] = true
				allCitations = append(allCitations, c)
			}
		}
		resultSections = append(resultSections, TemplateSectionResult{
			Key:       s.Key,
			Label:     s.Label,
			Type:      s.Type,
			Content:   text,
			Citations: secCitations,
		})
	}

	return &generatedAnswer{
		Citations:        allCitations,
		TemplateSections: resultSections,
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
	// 2026-09-08 RCA: initialized non-nil (not "var citations []Citation")
	// so zero matches marshals as JSON [] not null - a nil slice here
	// crashed the frontend's citations.map() (parseCitations.ts) whenever
	// a section/answer had zero [CHUNK_xxx] tokens, e.g. a code-type
	// template section, which by design never contains citation tokens.
	citations := []Citation{}
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
// StripModeCodeExempt: fenced code blocks kept unconditionally.
//   Code is the APPLICATION of cited principles — no per-line citations needed.
//   Explanation before/after the code block cites which principles are applied.
//   Explanation lines also kept — citations appear at start, not per-sentence.
//
// StripModeFull: every sentence without citation is stripped.
//   Correct for medical/legal/finance — strict grounding required.
func (e *Enforcer) stripUncited(answer string, mode StripMode) (string, int) {
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

		// Inside a code block: always keep regardless of mode.
		// WHY: code content is never sentence-split or citation-checked.
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
		} else if mode == StripModeCodeExempt {
			// CODE_EXEMPT mode: keep explanation lines even without inline citations.
			// Citations appear at the start of the explanation block, not per-sentence.
			clean = append(clean, sentence)
		} else {
			// FULL_STRIP mode: strip uncited sentences.
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

// resolveMaxTokens returns override if it is a positive value, or
// fallback otherwise. Centralizes the "0 means use default" contract
// documented on DomainProfile.MaxTokensFlat/MaxTokensStructured, so
// every call site (generateFlatText, generateStructured) applies it
// identically instead of duplicating an inline if-check.
func resolveMaxTokens(override, fallback int) int {
	if override > 0 {
		return override
	}
	return fallback
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
