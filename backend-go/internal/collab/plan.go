// Package collab implements Collaborative Relay mode: when a chat message
// selects "collaborative" with 2+ experts, their answers are produced as ONE
// section-organized response (sequential relay) instead of N independent
// per-expert answers.
//
// Design (locked):
//   - plan.go:        decide section order + section title per expert.
//   - relay.go:        run the decision engine once per expert/section,
//     sequentially, feeding each section the prior sections' content as
//     peer context (reuses decision.Engine's existing replyContext slot —
//     no new plumbing in decision/chinawall).
//   - assemble.go:     pure Markdown join of the ordered sections.
//
// NOT YET IMPLEMENTED: a consistency.go cross-section contradiction check
// (an optional post-assembly LLM pass flagging disagreements between
// sections) is a natural Phase 2 addition but is out of scope here —
// assemble.go's plain join is the full extent of post-relay processing
// today.
package collab

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/gateway"
)

// Section is one expert's slot in the relay — both the PLAN (title +
// ordering, before that expert has answered) and, once Content is filled in
// by relay.go, the final per-expert contribution to the assembled answer.
type Section struct {
	ExpertID     uuid.UUID `json:"expert_id"`
	ExpertName   string    `json:"expert_name"`
	SectionTitle string    `json:"section_title"`
	// Content is empty until relay.go runs this expert's turn.
	Content string `json:"content,omitempty"`
	// --- Debug metadata (additive, Phase-2 observability) ---
	// Populated AFTER relay.go returns, by orchestrator.go's
	// ProcessCollaborative, from the same ExpertResponse it already
	// stashes in responsesByExpert for OrchestratorResponse.ExpertResponses.
	// relay.go/RunExpert's signature is untouched (still just
	// (content string, err error)) — these fields are filled in a
	// second pass over result.Sections, never inside RunRelay itself.
	// omitempty on every field so a pre-existing persisted row (saved
	// before this feature) round-trips through json.Unmarshal with these
	// simply absent — zero-regression, same convention as every other
	// additive field in this codebase (see TemplateSections, CollabSections
	// in chat/service.go).
	//
	// Coverage: China Wall verdict YES|PARTIAL|NO for this expert's
	// retrieved chunks against the question it was actually asked.
	Coverage string `json:"coverage,omitempty"`
	// Confidence: decision engine's confidence score for this section.
	Confidence float64 `json:"confidence,omitempty"`
	// Mode: decision engine's ResponseMode (ADVISE|WARN|ASK|PUSH_BACK|REFUSE).
	Mode string `json:"mode,omitempty"`
	// GenericAllowancePct: the request-level generic-knowledge ceiling
	// (0-30) that was in effect for this entire turn — same value for
	// every section in a given relay, repeated here per-section purely so
	// a single collab_sections row is self-contained for debugging
	// (no need to cross-reference the request log to know what allowance
	// a REFUSE/PARTIAL happened under).
	GenericAllowancePct float64 `json:"generic_allowance_pct,omitempty"`
}

// ExpertInfo is the minimal shape plan.go needs about each selected expert.
// Deliberately small (not orchestrator.expertRecord) so this package has no
// dependency on orchestrator — the caller adapts its own expert records into
// this shape.
type ExpertInfo struct {
	ID           uuid.UUID
	Name         string
	Domain       string
	CategoryName string // "" when the expert has no category — Domain is used instead
}

// FallbackPlan builds the deterministic Phase 1 plan: one section per
// expert, in the exact order the caller passed them (== DB/request load
// order, never re-sorted), titled with the expert's own category name when
// it has one, else its domain name. Used whenever PlanSections is skipped
// (gw nil) or fails (LLM error / malformed JSON) — the relay must always
// have a plan to run, never block the whole collaborative answer on a
// planning LLM call (same fail-open posture as orchestrator.synthesize's
// fallback merge).
func FallbackPlan(experts []ExpertInfo) []Section {
	sections := make([]Section, 0, len(experts))
	for _, e := range experts {
		title := e.CategoryName
		if strings.TrimSpace(title) == "" {
			title = e.Domain
		}
		sections = append(sections, Section{
			ExpertID:     e.ID,
			ExpertName:   e.Name,
			SectionTitle: title,
		})
	}
	return sections
}

// PlanSections asks a cheap LLM call to order the experts and title each
// section based on the actual question, e.g. "Database Schema" before "API
// Design" for a backend question, rather than always using category/domain
// names verbatim. On any failure (gw nil, call error, malformed JSON, or a
// returned plan that doesn't name every expert exactly once) it falls back
// to FallbackPlan — planning must never block or crash the relay (§3.1 P3).
// UNWIRED (future phase) — ProcessCollaborative NO LONGER calls PlanSections.
//
// The collaborative order is now always FallbackPlan(experts): the order the
// client selected the experts in the UI. PlanSections (the cheap LLM that
// reorders experts + writes section titles) is deliberately UNWIRED, not
// deleted, so a future phase can re-wire it without rewriting the function.
// See collab_relay_runs.plan — it is populated from FallbackPlan.
//
// To re-wire: change ProcessCollaborative's plan assignment from
//   plan := collab.FallbackPlan(infos)
// back to
//   plan := collab.PlanSections(ctx, o.gw, req.Message, infos, o.logger)
// (PlanSections already falls back to FallbackPlan internally on any LLM error.)
func PlanSections(ctx context.Context, gw *gateway.ModelGateway, question string, experts []ExpertInfo, logger *zap.Logger) []Section {
	fallback := FallbackPlan(experts)
	if gw == nil {
		return fallback
	}

	var sb strings.Builder
	sb.WriteString("Multiple domain experts will answer ONE question together, in sequence, as a single ")
	sb.WriteString("organized answer (not separate independent answers). Decide the best section order and a ")
	sb.WriteString("short, specific section title for each expert, based on what the question actually needs.\n\n")
	sb.WriteString("Question: ")
	sb.WriteString(question)
	sb.WriteString("\n\nExperts (every one MUST appear exactly once in your answer, identified by id):\n")
	for _, e := range experts {
		category := e.CategoryName
		if strings.TrimSpace(category) == "" {
			category = e.Domain
		}
		sb.WriteString("- id=\"" + e.ID.String() + "\" name=\"" + e.Name + "\" domain=\"" + e.Domain + "\" category=\"" + category + "\"\n")
	}
	sb.WriteString("\nReturn JSON only, no prose outside the JSON:\n")
	sb.WriteString(`{"sections": [{"expert_id": "...", "section_title": "..."}, ...]}` + "\n")
	sb.WriteString("The sections array must contain every expert id exactly once, in the order they should answer.")

	resp, err := gw.Call(ctx, gateway.LLMRequest{
		Model:       gateway.ModelCheap,
		UserPrompt:  sb.String(),
		MaxTokens:   400,
		Temperature: 0.1,
	})
	if err != nil {
		logger.Warn("collab: PlanSections LLM call failed, using fallback plan", zap.Error(err))
		return fallback
	}

	clean := strings.TrimSpace(resp.Content)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start == -1 || end == -1 || end < start {
		logger.Warn("collab: PlanSections LLM response had no JSON object, using fallback plan")
		return fallback
	}

	var parsed struct {
		Sections []struct {
			ExpertID     string `json:"expert_id"`
			SectionTitle string `json:"section_title"`
		} `json:"sections"`
	}
	if err := json.Unmarshal([]byte(clean[start:end+1]), &parsed); err != nil {
		logger.Warn("collab: PlanSections LLM response JSON parse failed, using fallback plan", zap.Error(err))
		return fallback
	}

	byID := make(map[uuid.UUID]ExpertInfo, len(experts))
	for _, e := range experts {
		byID[e.ID] = e
	}

	seen := make(map[uuid.UUID]bool, len(experts))
	var sections []Section
	for _, s := range parsed.Sections {
		id, perr := uuid.Parse(s.ExpertID)
		if perr != nil {
			continue
		}
		info, ok := byID[id]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		title := strings.TrimSpace(s.SectionTitle)
		if title == "" {
			title = info.CategoryName
			if title == "" {
				title = info.Domain
			}
		}
		sections = append(sections, Section{
			ExpertID:     info.ID,
			ExpertName:   info.Name,
			SectionTitle: title,
		})
	}

	// The plan must name every expert exactly once — anything else (missing
	// expert, duplicate, hallucinated id) is a malformed plan. Fall back
	// rather than silently dropping an expert's answer from the relay.
	if len(sections) != len(experts) {
		logger.Warn("collab: PlanSections LLM plan did not cover every expert exactly once, using fallback plan",
			zap.Int("expected", len(experts)), zap.Int("got", len(sections)))
		return fallback
	}
	return sections
}
