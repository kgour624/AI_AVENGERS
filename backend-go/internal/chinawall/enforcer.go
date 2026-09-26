package chinawall

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/category"
	"ai_avengers/backend/internal/config"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/ml"
)

// CourseChunk is a retrieved chunk with its rerank score.
// WHY SourceFile and ChunkIndex added (Feature #23, 2026-09-23):
//
//	Citation modal showed chunk text but NOT which transcript it came from.
//	Users couldn't tell if a citation was from "React Hooks" or "State Management".
//	ChunkIndex helps locate the exact position in the source transcript.
type CourseChunk struct {
	ID          uuid.UUID
	Text        string
	Topic       string
	RerankScore float32
	// SourceFile: original transcript filename (e.g. "react_hooks_part1.txt")
	// Empty string if chunk has no source_file (legacy data, repo chunks, etc.)
	SourceFile string
	// ChunkIndex: 0-based position in the original transcript
	// Helps users locate "this is chunk #42 out of 150 in that transcript"
	ChunkIndex int
}

// EnforceResult is the output of China Wall enforcement.
type EnforceResult struct {
	Status      string // success | retry | refused | partial
	Answer      string
	Citations   []Citation
	Coverage    string // YES | PARTIAL | NO
	Confidence  float64
	LayerFailed int // 0 = all passed
	Reason      string
	// TemplateSections is set ONLY when the expert's category has a
	// non-empty template_schema (CT-B, CATEGORY_TEMPLATE_HANDOFF.md §4).
	// nil/empty for every flat-text expert (CT-L2) — Answer
	// is the only populated field in that case, exactly as before this
	// feature existed.
	TemplateSections []TemplateSectionResult
	// QualityScore (B6): overall LLM-as-judge score in [0,1] for the
	// flat-path answer. 0 when the judge is disabled, failed open, or
	// the path is structured (structured scoring deferred). Observability
	// only — never used to refuse a cited answer after retries are spent.
	QualityScore float64
	// Claims (B8): atomic claim→evidence reports for the flat-path answer.
	// nil when claim verify is disabled, failed open, or structured path
	// (structured claim verify deferred). Unverifiable/refuted claims are
	// also labelled inline in Answer; never silently asserted (P9).
	Claims []ClaimReport
}

// Citation links a claim to a source chunk.
// WHY SourceName and ChunkIndex added (Feature #23, 2026-09-23):
//
//	Frontend CitationChip modal showed chunk text + score, but NOT which
//	transcript the citation came from. Users couldn't tell if a citation
//	was from "React Hooks" or "State Management" transcript.
//	ChunkIndex helps users locate the exact position: "Chunk #42 of 150".
type Citation struct {
	ChunkID uuid.UUID `json:"chunk_id"`
	Text    string    `json:"text"`
	Score   float32   `json:"score"`
	// SourceName: human-readable transcript filename (e.g. "react_hooks_part1.txt")
	// Empty string if chunk has no source_file (legacy data, repo chunks, etc.)
	// Frontend displays "Unknown Source" when empty.
	SourceName string `json:"source_name,omitempty"`
	// ChunkIndex: 0-based position in the original transcript
	// Frontend displays as 1-based: "Chunk #43" (index 42 + 1)
	ChunkIndex int `json:"chunk_index,omitempty"`
}

// Enforcer implements the 4-layer China Wall system.
//
// Layer 1: Reranker threshold (0.35) — fast reject
// Layer 2: Coverage check (YES/PARTIAL/NO) — LLM check
// Layer 3: Generate with mandatory citations — strong LLM
// Layer 4: Strip uncited claims — regex
//
// DOMAIN BEHAVIOR:
//
//	Each layer's behavior is controlled by DomainProfile, not hardcoded booleans.
//	registry.Get(expertDomain) → *DomainProfile → controls all 4 layers.
//	Unknown domain → BaseProfile (strict defaults, safe fallback).
//
// CONFLICT RESOLUTION (base wall safety net):
//
//	Domain rules apply first.
//	IF Layer 4 output is empty AND profile.StripMode == FULL_STRIP:
//	  → retry with BaseProfile (domain rules over-relaxed the wall)
//	IF profile.StripMode == CODE_EXEMPT AND output has code block:
//	  → valid output, domain rules win
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

// maxGenericAllowancePct caps the client-supplied generic ceiling, mirroring the
// workflow's 0-30% rule. Above this the "generic" share would dominate the
// answer and the China Wall would no longer mean anything.
const maxGenericAllowancePct = 30

// genericKnowledgeRule is the instruction appended to the domain profile when a
// generic allowance is set. Kept beside the cap so the rule and the number can
// never drift apart.
func genericKnowledgeRule(pct float64) string {
	return fmt.Sprintf(`

GENERIC KNOWLEDGE (allowed up to %.0f%%):
- The trained chunks above are still the primary source; use them first.
- If they do not cover part of the question, you MAY answer that part from general knowledge.
- Tag every such part with [GENERIC] so it is visibly not from the corpus.
- Keep generic content under %.0f%% of your answer.
- Never attach a citation to a generic part, and never invent one.`, pct, pct)
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
	// genericAllowancePct (0-30): the client's generic ceiling, the same
	// concept workflows.generic_allowance_pct already uses. 0 (the default)
	// preserves the strict China Wall exactly: no chunk coverage -> refuse.
	// >0 lets the expert answer the uncovered part from general knowledge,
	// tagged [GENERIC] and capped, instead of refusing the whole question.
	genericAllowancePct float64,
	// tokenCh: non-nil enables streaming for Gate 5 generation.
	// nil = blocking Call() (backward compatible, used by smoke test etc.).
	tokenCh chan<- string,
) (*EnforceResult, error) {
	// Load domain profile from registry.
	// O(1) lookup. Falls back to BaseProfile for unknown domains.
	// WHY registry not hardcoded: domain behavior is DB-driven, no redeploy needed.
	profile := e.registry.Get(expertDomain)

	if genericAllowancePct < 0 {
		genericAllowancePct = 0
	}
	if genericAllowancePct > maxGenericAllowancePct {
		genericAllowancePct = maxGenericAllowancePct
	}
	allowGeneric := genericAllowancePct > 0

	// LAYER 1: Reranker threshold
	// Priority: domain profile override > config default > relaxed (last retry).
	// WHY domain override: principle-transfer domains (DSA) need lower threshold.
	// WHY config default: factual domains (medical/legal) use strict 0.35.
	// WHY relaxed on last retry: last-resort fallback to avoid total refusal.
	threshold := e.cfg.RerankerThreshold
	if profile.RerankerThreshold > 0 {
		// Domain profile overrides config default.
		// 0 means "use config default" (backward compatible).
		threshold = profile.RerankerThreshold
	}
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
		// A low reranker score is NOT, by itself, proof the expert cannot answer.
		// For APPLY_PRINCIPLES domains (system design, DSA, coding) questions are
		// answered by applying the principles in the corpus, and a design
		// question can legitimately score low against the wording that was
		// ingested. Hard-refusing there threw away answers a trained expert
		// produced correctly — the production regression where a system-design
		// expert refused its own domain even with generic allowance at 0%.
		// Only LITERAL_MATCH domains (medical/legal/finance), where the answer
		// must be found verbatim, keep the strict score gate.
		if !allowGeneric && profile.CoverageMode != CoverageModeApplyPrinciples {
			if attempt >= e.cfg.MaxRetries {
				return e.buildRefusal("insufficient_relevance",
					fmt.Sprintf("Best relevance score %.2f below threshold %.2f", bestScore, threshold)), nil
			}
			return &EnforceResult{Status: "retry", LayerFailed: 1}, nil
		}
		if profile.CoverageMode == CoverageModeApplyPrinciples {
			e.logger.Info("Layer 1: low reranker score on a principle-transfer domain — answering from the best available chunks",
				zap.String("expert", expertName),
				zap.String("domain", expertDomain),
				zap.Float32("best_score", bestScore),
				zap.Float64("threshold", threshold))
		}
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
		if !allowGeneric {
			// APPLY_PRINCIPLES exists precisely for domains whose questions are
			// answered by APPLYING principles rather than by finding the answer
			// verbatim in the corpus (system design, DSA, coding). A literal
			// "NO" there is a mis-signal, so it must not hard-refuse — this is
			// what made a trained system-design expert refuse its own domain's
			// question. LITERAL_MATCH domains (medical/legal/finance) keep the
			// strict refusal.
			if profile.CoverageMode == CoverageModeApplyPrinciples {
				e.logger.Info("Layer 2: coverage NO on a principle-transfer domain — continuing as PARTIAL",
					zap.String("expert", expertName),
					zap.String("domain", expertDomain),
					zap.String("coverage_mode", string(profile.CoverageMode)))
				coverage = "PARTIAL"
			} else {
				// Tell the client exactly what would unblock this — the refusal
				// is actionable instead of a dead end.
				return e.buildRefusal("not_covered",
					fmt.Sprintf("Content does not cover this question. Generic knowledge is disabled (allowance 0%%); raise the answer basis to allow up to %d%%.", maxGenericAllowancePct)), nil
			}
		} else {
			e.logger.Info("Layer 2 allowed the question via generic allowance",
				zap.String("expert", expertName),
				zap.Float64("generic_allowance_pct", genericAllowancePct))
		}
	}

	if coverage == "PARTIAL" && attempt >= 3 {
		// The decision engine turns "partial" into a refusal ("Cannot provide a
		// complete answer"). For an APPLY_PRINCIPLES domain a partial match is
		// normal and the expert must still answer from what it has — otherwise a
		// trained system-design expert is refused on a question it can solve.
		if profile.CoverageMode != CoverageModeApplyPrinciples {
			return &EnforceResult{
				Status:   "partial",
				Coverage: "PARTIAL",
				Reason:   "Partial coverage after multiple attempts",
			}, nil
		}
		e.logger.Info("Layer 2: PARTIAL coverage on a principle-transfer domain — answering instead of refusing",
			zap.String("expert", expertName),
			zap.String("domain", expertDomain))
	}

	// LAYER 3: Generate with mandatory citations
	// Behavior controlled by profile.CitationMode and profile.SystemPromptExt.
	// With a generic allowance, extend a COPY of the profile — the registry's
	// profile is shared across every concurrent question and must never be
	// mutated (a mutation here would leak one client's allowance into everyone
	// else's prompts).
	if allowGeneric {
		withGeneric := *profile
		withGeneric.SystemPromptExt += genericKnowledgeRule(genericAllowancePct)
		profile = &withGeneric
	}
	generated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, templateSections, defaultLanguage, tokenCh)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// STRUCTURED PATH (CT-B1/B2): category had a non-empty template_schema,
	// so generateWithCitations took the structured branch and returned
	// per-section results instead of one flat Answer. Handled entirely
	// separately from the flat path below so the flat path's existing
	// logic (including its BaseProfile safety net) stays byte-for-byte
	// unchanged for every non-categorized expert (CT-L2).
	//
	// A14: run this BEFORE the flat zero-citation check. Structured
	// answers can legitimately have empty top-level Citations when
	// sections are code/test-only (citation-exempt). enforceStructured
	// owns citation policy for those sections; the flat check below
	// must not refuse them first.
	if len(generated.TemplateSections) > 0 {
		return e.enforceStructured(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, generated, coverage, bestScore, templateSections, defaultLanguage)
	}

	// FLAT PATH only: zero citations => retry/refuse. Structured path
	// already returned above (A14).
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

	// SECURITY FIX: Remove [CHUNK_xxx] tokens from answer before sending to frontend.
	// WHY: These are internal citation markers, not user-facing text.
	// Citations are already extracted into the Citations array.
	// Exposing chunk UUIDs leaks internal database structure.
	cleanAnswer = e.cleanChunkIDs(cleanAnswer)

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
			// SECURITY FIX: Clean chunk IDs from BaseProfile retry path too.
			cleanAnswer = e.cleanChunkIDs(cleanAnswer)
			if strings.TrimSpace(cleanAnswer) == "" {
				return e.buildRefusal("empty_after_strip", "Could not generate a properly cited answer"), nil
			}
			// BaseProfile produced the cleaned answer — keep BaseProfile for
			// quality regen so domain strip rules can't empty it again.
			ans, cites, qScore := e.applyQualityGate(
				ctx, question, chunks, expertName, reasoningCharter, replyContext, BaseProfile,
				cleanAnswer, baseGenerated.Citations,
			)
			claimRes := e.verifyClaims(ctx, ans, chunks, cites)
			return &EnforceResult{
				Status:       "success",
				Answer:       claimRes.Answer,
				Citations:    cites,
				Coverage:     coverage,
				Confidence:   float64(bestScore),
				QualityScore: qScore,
				Claims:       claimRes.Claims,
			}, nil
		} else {
			return e.buildRefusal("empty_after_strip", "Could not generate a properly cited answer"), nil
		}
	}

	ans, cites, qScore := e.applyQualityGate(
		ctx, question, chunks, expertName, reasoningCharter, replyContext, profile,
		cleanAnswer, generated.Citations,
	)
	claimRes := e.verifyClaims(ctx, ans, chunks, cites)
	return &EnforceResult{
		Status:       "success",
		Answer:       claimRes.Answer,
		Citations:    cites,
		Coverage:     coverage,
		Confidence:   float64(bestScore),
		QualityScore: qScore,
		Claims:       claimRes.Claims,
	}, nil
}

// applyQualityGate (B6) scores a cleaned flat-path answer and, when
// below floor, regenerates up to MaxQualityRetries times with the
// judge's feedback injected into the generation prompt. tokenCh is
// deliberately nil on regen so a second stream does not interleave
// with the already-forwarded first stream (SSE would append a second
// answer after the client already painted the first).
//
// Kill switch: QualityJudgeEnabled=false → return inputs unchanged,
// score 0. Fail-open on judge error → return current best answer.
func (e *Enforcer) applyQualityGate(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	replyContext string,
	profile *DomainProfile,
	answer string,
	citations []Citation,
) (string, []Citation, float64) {
	if e == nil || !e.cfg.QualityJudgeEnabled {
		return answer, citations, 0
	}
	floor := e.cfg.QualityFloor
	if floor <= 0 || floor > 1 {
		floor = 0.7
	}
	maxRetries := e.cfg.MaxQualityRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	bestAnswer := answer
	bestCites := citations
	bestScore := 0.0

	verdict := e.judgeAnswer(ctx, question, answer, chunks, floor)
	bestScore = verdict.Overall
	e.logger.Info("quality judge",
		zap.Float64("overall", verdict.Overall),
		zap.Float64("floor", floor),
		zap.Bool("pass", verdict.Pass),
		zap.String("method", verdict.Method),
		zap.String("feedback", truncateForLog(verdict.Feedback, 120)),
	)
	// Pass, score-only mode, or fail-open (judge broke) → keep answer, no regen.
	if verdict.Pass || maxRetries == 0 || verdict.Method == "fail_open" {
		return bestAnswer, bestCites, bestScore
	}

	feedback := verdict.Feedback
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if strings.TrimSpace(feedback) == "" {
			feedback = "Improve accuracy, coverage and structure against the training material."
		}
		e.logger.Info("quality regenerate",
			zap.Int("attempt", attempt),
			zap.Int("max", maxRetries),
			zap.String("feedback", truncateForLog(feedback, 120)),
		)
		// tokenCh=nil: no streaming on regen (see function doc).
		// revisionFeedback is appended to the user prompt of generateFlatText.
		regenProfile := profile
		if regenProfile == nil {
			regenProfile = BaseProfile
		}
		regen, err := e.generateWithCitations(
			ctx, question, chunks, expertName, reasoningCharter, replyContext,
			regenProfile, nil, "", nil, feedback,
		)
		if err != nil {
			e.logger.Warn("quality regenerate failed", zap.Int("attempt", attempt), zap.Error(err))
			break
		}
		clean, _ := e.stripUncited(regen.Answer, regenProfile.StripMode)
		clean = e.cleanChunkIDs(clean)
		if strings.TrimSpace(clean) == "" {
			e.logger.Warn("quality regenerate produced empty after strip", zap.Int("attempt", attempt))
			continue
		}
		v2 := e.judgeAnswer(ctx, question, clean, chunks, floor)
		e.logger.Info("quality judge after regen",
			zap.Int("attempt", attempt),
			zap.Float64("overall", v2.Overall),
			zap.Bool("pass", v2.Pass),
			zap.String("method", v2.Method),
		)
		if v2.Overall >= bestScore {
			bestAnswer = clean
			bestCites = regen.Citations
			bestScore = v2.Overall
		}
		if v2.Pass {
			return bestAnswer, bestCites, bestScore
		}
		feedback = v2.Feedback
	}
	// Retries spent: ship the best attempt. Citation wall already passed;
	// quality is soft.
	return bestAnswer, bestCites, bestScore
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
			// SECURITY FIX: Remove [CHUNK_xxx] tokens from prose sections.
			// Same rationale as flat-text path above.
			s.Content = e.cleanChunkIDs(s.Content)
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
	baseGenerated, err := e.generateWithCitations(ctx, question, chunks, expertName, reasoningCharter, replyContext, BaseProfile, templateSections, defaultLanguage, nil)
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
//  1. Problem-solving mode (DSA/coding): checks if expert can APPLY principles
//     to solve the problem. Correct for DSA — principles must transfer to new problems.
//  2. Factual mode (default): checks if chunks CONTAIN the answer.
//     Correct for medical/legal/domain-fact experts.
//
// WHY two modes:
//
//	A DSA expert trained on "two pointers, sliding window, prefix sum" SHOULD
//	solve "Longest Common Prefix" by applying string traversal principles.
//	Asking "does the transcript mention Longest Common Prefix?" is wrong —
//	it would refuse every new problem, defeating the purpose of DSA education.
//
// checkCoverage asks cheap LLM if chunks can answer the question.
// Uses Chain-of-Thought prompting.
//
// CoverageModeApplyPrinciples: checks if expert can APPLY principles to solve.
//
//	Correct for DSA, coding — principles transfer to new problems.
//
// CoverageModeLiteralMatch: checks if chunks CONTAIN the answer.
//
//	Correct for medical, legal, finance — facts must be in transcript.
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
	// revisionFeedback (B6): optional judge feedback appended to the user
	// prompt on quality-gated regenerations. Empty on the first generate.
	// Variadic so every existing call site (Enforce / BaseProfile safety
	// net / structured path) stays source-compatible without edits.
	revisionFeedback ...string,
) (*generatedAnswer, error) {
	feedback := ""
	if len(revisionFeedback) > 0 {
		feedback = strings.TrimSpace(revisionFeedback[0])
	}

	// Build context with chunk IDs
	var contextSB strings.Builder
	for _, c := range chunks {
		contextSB.WriteString(fmt.Sprintf("[CHUNK_%s]\n%s\n\n", c.ID, c.Text))
	}

	if len(templateSections) > 0 {
		// Structured path ignores revisionFeedback for now (B6 scope =
		// flat only; per-section judge deferred).
		return e.generateStructured(ctx, question, chunks, expertName, reasoningCharter, replyContext, contextSB.String(), templateSections, defaultLanguage, profile, tokenCh)
	}

	return e.generateFlatText(ctx, question, chunks, expertName, reasoningCharter, replyContext, profile, contextSB.String(), tokenCh, feedback)
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
	revisionFeedback string, // B6: empty on first generate; set on quality regen
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

	// User prompt: the question, plus optional REVISION FEEDBACK from the
	// quality judge (B6). Empty feedback on the first generate keeps the
	// prompt byte-identical to pre-B6 for the hot path.
	userPrompt := question
	if revisionFeedback != "" {
		userPrompt = question + "\n\nREVISION FEEDBACK (previous draft was below quality floor — fix these points, keep all citations):\n" + revisionFeedback
	}

	// Gate 5 generation: streaming when tokenCh non-nil, blocking otherwise.
	//
	// WHY streaming timeout (45s):
	//   Without a timeout, if the provider sends HTTP 200 headers but then
	//   stalls the body (no tokens, no [DONE], no RST/FIN), rawTokenCh
	//   never closes, the for-range loop below hangs forever, tokenDone
	//   never signals, and message/handler.go's <-tokenDone blocks
	//   indefinitely — the user never gets a response.
	//   45s is chosen as: generous enough for slow providers on long answers,
	//   tight enough to fail fast and fall back to blocking Call() which
	//   has its own 120s transport timeout.
	if tokenCh != nil {
		streamCtx, streamCancel := context.WithTimeout(ctx, 45*time.Second)
		rawTokenCh, rawRespCh, streamErr := e.gateway.StreamCall(streamCtx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
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
					streamCancel()
					return nil, ctx.Err()
				}
			}
			<-rawRespCh // drain metadata channel
			streamCancel()
			content := sb.String()
			if content != "" {
				return &generatedAnswer{Answer: content, Citations: e.extractCitations(content, chunks)}, nil
			}
		} else {
			streamCancel()
		}
		// StreamCall failed, timed out, or returned empty — fall through to blocking call.
		// WHY heartbeat goroutine:
		//   Blocking Call() can take 30+ seconds. During this time, no SSE
		//   events are sent to the frontend. Browser SSE connection drops
		//   after ~30s of silence (OS/browser TCP keepalive timeout).
		//   When the blocking call finally returns, the frontend connection
		//   is already dead — UI stays stuck on "Thinking..." forever.
		//   Fix: send a "thinking" SSE heartbeat every 5s while blocking
		//   call runs. This keeps the SSE connection alive without sending
		//   fake content. heartbeatDone stops the goroutine immediately
		//   when the blocking call returns (no goroutine leak).
		e.logger.Warn("streaming failed or empty, falling back to blocking call",
			zap.NamedError("stream_err", streamErr),
		)
		if tokenCh != nil {
			heartbeatDone := make(chan struct{})
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-heartbeatDone:
						return
					case <-ctx.Done():
						return
					case <-ticker.C:
						// Send empty string as heartbeat — handler.go forwards
						// non-empty tokens only, so this keeps SSE alive
						// without injecting fake content into the answer.
						select {
						case tokenCh <- "":
						default: // channel full — skip this heartbeat
						}
					}
				}
			}()
			defer close(heartbeatDone)
		}
	}

	resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelStrong,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    resolveMaxTokens(profile.MaxTokensFlat, DefaultMaxTokensFlat),
		Temperature:  0.4,
	})
	if err != nil {
		return nil, err
	}
	citations := e.extractCitations(resp.Content, chunks)
	return &generatedAnswer{Answer: resp.Content, Citations: citations}, nil
}

// generateStructured generates a structured answer section-by-section.
//
// DESIGN DECISION (2026-09-16):
//
//	Previous approach: one LLM call requesting ALL sections as a single
//	JSON object. This failed reliably because:
//	1. DeepSeek and reasoning models emit <think> blocks before JSON.
//	   Even after stripping, the JSON was often malformed or truncated.
//	2. 5 heavy sections (prose + full code + test cases) in one JSON
//	   object routinely exceeded 32K tokens, truncating mid-JSON.
//	3. JSON format forced the model to escape code (\n, \") — LLMs
//	   are unreliable at this, producing invalid JSON on complex code.
//
//	New approach: one LLM call PER SECTION, sequentially.
//	WHY sequential not parallel:
//	- User reads Pattern → Idea → Code → Walkthrough → TestCases in order.
//	  Sequential streaming lets them read each section as it arrives.
//	- Each section is independently China-Wall enforced with citations.
//	- No JSON format — plain text per section, no escaping issues.
//	- Any model works: DeepSeek, Claude, Gemini — all return plain text.
//
//	SOLID: Single Responsibility — each section call has one job.
//	OCP: New section types extend by adding a case, not modifying core.
//	LSP: TemplateSectionResult contract unchanged — callers unaffected.
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
	tokenCh chan<- string,
) (*generatedAnswer, error) {
	if defaultLanguage == "" {
		defaultLanguage = "java"
	}

	allCitations := []Citation{}
	seen := make(map[string]bool)
	resultSections := make([]TemplateSectionResult, 0, len(sections))

	for _, section := range sections {
		content, secCitations, err := e.generateOneSection(
			ctx, question, chunks, expertName, reasoningCharter,
			replyContext, contextText, section, defaultLanguage, profile, tokenCh,
		)
		if err != nil {
			// One section failing should not kill the whole answer.
			// Log and continue — user gets partial answer, not blank screen.
			e.logger.Warn("section generation failed — using empty content",
				zap.String("section", section.Key),
				zap.Error(err),
			)
			content = ""
			secCitations = []Citation{}
		}

		for _, c := range secCitations {
			key := c.ChunkID.String()
			if !seen[key] {
				seen[key] = true
				allCitations = append(allCitations, c)
			}
		}
		resultSections = append(resultSections, TemplateSectionResult{
			Key:       section.Key,
			Label:     section.Label,
			Type:      section.Type,
			Content:   content,
			Citations: secCitations,
		})
	}

	return &generatedAnswer{
		Citations:        allCitations,
		TemplateSections: resultSections,
	}, nil
}

// generateOneSection generates content for a single template section.
//
// Each section type gets a purpose-built prompt:
//   - prose:      explain + cite training material principles
//   - code:       complete working code, no citations inside code block
//   - test_cases: concrete test cases covering BASE/EDGE/CORNER/STRESS
//
// Streaming: if tokenCh is non-nil, tokens stream to the caller as they
// arrive. A section label header ("## Pattern\n\n") is sent first so the
// user knows which section is streaming.
func (e *Enforcer) generateOneSection(
	ctx context.Context,
	question string,
	chunks []CourseChunk,
	expertName string,
	reasoningCharter string,
	replyContext string,
	contextText string,
	section category.TemplateSection,
	defaultLanguage string,
	profile *DomainProfile,
	tokenCh chan<- string,
) (string, []Citation, error) {
	systemPrompt := e.buildSectionPrompt(
		expertName, reasoningCharter, contextText,
		section, defaultLanguage, profile,
	)
	if replyContext != "" {
		systemPrompt += "\n\n" + replyContext
	}

	userPrompt := fmt.Sprintf(
		"Question: %s\n\nGenerate ONLY the %s section. Plain text only — no JSON, no markdown wrapper.",
		question, section.Label,
	)

	// Send section header to SSE stream so user sees which section is coming.
	if tokenCh != nil {
		header := fmt.Sprintf("\n\n## %s\n\n", section.Label)
		select {
		case tokenCh <- header:
		case <-ctx.Done():
			return "", nil, ctx.Err()
		}
	}

	var content string

	if tokenCh != nil {
		// Streaming path: stream tokens to caller as they arrive.
		streamCtx, streamCancel := context.WithTimeout(ctx, 60*time.Second)
		rawTokenCh, rawRespCh, streamErr := e.gateway.StreamCall(streamCtx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
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
					streamCancel()
					return "", nil, ctx.Err()
				}
			}
			<-rawRespCh
			streamCancel()
			content = sb.String()
		} else {
			streamCancel()
		}
	}

	// Blocking fallback: streaming failed or tokenCh is nil.
	if content == "" {
		resp, err := e.gateway.Call(ctx, gateway.LLMRequest{
			Model:        gateway.ModelStrong,
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			MaxTokens:    resolveMaxTokens(profile.MaxTokensFlat, DefaultMaxTokensFlat),
			Temperature:  0.4,
		})
		if err != nil {
			return "", nil, fmt.Errorf("section %s generation failed: %w", section.Key, err)
		}
		content = resp.Content
	}

	// Extract citations from the generated content.
	// Code sections are citation-exempt by design (code IS the application
	// of cited principles — no per-line citations needed in code blocks).
	var citations []Citation
	if section.Type != category.SectionTypeCode {
		citations = e.extractCitations(content, chunks)
	} else {
		citations = []Citation{}
	}

	return content, citations, nil
}

// buildSectionPrompt builds a focused system prompt for one section.
// Each section type gets purpose-built instructions — no JSON format,
// no multi-section confusion, just one clear job per call.
func (e *Enforcer) buildSectionPrompt(
	expertName string,
	reasoningCharter string,
	contextText string,
	section category.TemplateSection,
	defaultLanguage string,
	profile *DomainProfile,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You are %s, a domain expert.\n\n", expertName))
	sb.WriteString(fmt.Sprintf("REASONING CHARTER:\n%s\n\n", reasoningCharter))
	sb.WriteString("YOUR TRAINING MATERIAL:\n")
	sb.WriteString(contextText)
	sb.WriteString("\n\n")

	switch section.Type {
	case category.SectionTypeCode:
		sb.WriteString(fmt.Sprintf(
			"Write COMPLETE, WORKING %s code that solves the problem.\n"+
				"Rules:\n"+
				"1. Output ONLY the code in a fenced code block (```%s ... ```)\n"+
				"2. Code must be correct and runnable\n"+
				"3. No explanation outside the code block\n"+
				"4. No [CHUNK_xxx] tokens inside the code\n",
			defaultLanguage, defaultLanguage,
		))

	case category.SectionTypeTestCases:
		sb.WriteString(
			"Generate concrete test cases covering: BASE (normal), EDGE (boundary), CORNER (tricky), STRESS (large input).\n" +
				"Format each as: Input → Expected Output (brief explanation).\n" +
				"No code, no JSON — plain readable text.\n",
		)

	default: // prose sections: pattern, idea, walkthrough, complexity, etc.
		guidance := sectionGuidance(section)
		sb.WriteString(fmt.Sprintf("Write the %s section.\n", section.Label))
		sb.WriteString(fmt.Sprintf("What to write: %s\n\n", guidance))

		if profile.CitationMode == CitationModeLoose {
			sb.WriteString(
				"Citation rules:\n" +
					"- Cite the training-material principles you apply using [CHUNK_uuid] format\n" +
					"- Explain your full reasoning FIRST — citations support the explanation\n" +
					"- NEVER answer with citations alone\n",
			)
		} else {
			sb.WriteString(
				"Citation rules:\n" +
					"- Every factual claim MUST cite a source using [CHUNK_uuid] format\n" +
					"- If information is not in your training material, say so explicitly\n",
			)
		}
	}

	if profile.SystemPromptExt != "" {
		sb.WriteString(fmt.Sprintf("\nAdditional rules: %s\n", profile.SystemPromptExt))
	}

	return sb.String()
}

// cleanChunkIDs removes [CHUNK_xxx] tokens from the answer text.
// WHY: These are internal citation markers for the LLM, not user-facing text.
// SECURITY: Exposing chunk UUIDs leaks internal database structure.
// UX: Users should see clean text, not technical IDs.
//
// MENTAL MODEL:
//
//	Input:  "Use bcrypt [CHUNK_abc-123] for passwords"
//	Output: "Use bcrypt for passwords"
//
// CROSS-QUESTION:
//
//	Q: Why not remove during generation?
//	A: LLM needs them for citation tracking, remove after extraction
//
//	Q: What if chunk ID is in code block?
//	A: Regex removes all [CHUNK_xxx] tokens, including in code
//	   (code blocks should never have chunk IDs anyway)
func (e *Enforcer) cleanChunkIDs(answer string) string {
	pattern := regexp.MustCompile(`\[CHUNK_[a-f0-9-]+\]`)
	return pattern.ReplaceAllString(answer, "")
}

// extractCitations finds [CHUNK_uuid] references in the answer.
// WHY SourceName and ChunkIndex added (Feature #23, 2026-09-23):
//
//	Frontend CitationChip modal needs to show which transcript a citation
//	came from, not just the chunk text. "Source: react_hooks_part1.txt, Chunk #42"
//	is much more useful than just showing 200 chars of text with no context.
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
				// Feature #23: Include source transcript name and chunk position
				// so frontend can display "Source: react_hooks.txt, Chunk #42"
				// instead of just showing chunk text with no context.
				SourceName: chunk.SourceFile,
				ChunkIndex: chunk.ChunkIndex,
			})
		}
	}
	return citations
}

// stripUncited removes sentences without citations, while preserving
// markdown/code structure.
//
// StripModeCodeExempt: fenced code blocks kept unconditionally.
//
//	Code is the APPLICATION of cited principles — no per-line citations needed.
//	Explanation before/after the code block cites which principles are applied.
//	Explanation lines also kept — citations appear at start, not per-sentence.
//
// StripModeFull: every sentence without citation is stripped.
//
//	Correct for medical/legal/finance — strict grounding required.
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
