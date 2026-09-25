package training

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// The ingest gate (I3): the conditions an expert must meet before it is marked
// "trained", taken from DOMAIN_EXPERT_COLLABORATION_DESIGN.md §5.3 and made explicit
// in ONE place.
//
// WHY this exists: the five conditions were documented, but only two were enforced —
// a chunk count nobody read, and a smoke test that checked retrieval alone. Charter
// rules, clarification coverage and capability depth were never consulted, so an
// expert could be published without them while the doc said otherwise. And the one
// capability condition was specified against a depth derived from chunk counts; it
// now reads the MEASURED level, because a coverage band cannot tell "the corpus
// mentions this topic" from "the expert can answer a question about it".
//
// Everything here is deliberately a plain comparison over measured inputs, so the
// judgement can be tested without a database or a model, and so the screen can state
// exactly which condition failed instead of a single pass/fail.

const (
	// §5.3(1) — "at least 100 chunks embedded and stored", to avoid trivially thin
	// experts.
	gateMinChunks = 100

	// §5.3(2) — "charter successfully extracted with at least 5 WHY-embedded rules".
	// Rules are counted as non-empty bullet or numbered lines in the reasoning
	// charter. This is a PROXY and is treated as one: the charter is prose, so any
	// count is a heuristic and says so in the detail string rather than pretending to
	// be a measurement.
	gateMinCharterRules = 5

	// §5.3(3) — "clarification charter has entries for at least 3 topics".
	gateMinClarificationTopics = 3

	// §5.3(4) — the doc asks for "3 topics with depth level >= 3" on the 1-5 coverage
	// scale. On the MEASURED 1-3 scale the equivalent claim is "the expert answers
	// mechanics and trade-offs", i.e. level 2. Requiring level 3 (failure modes)
	// would gate on edge-case questions that a good expert may legitimately miss.
	// Documented as an adaptation, not as the doc's literal number.
	gateMinMeasuredTopics     = 3
	gateRequiredMeasuredLevel = CapabilityLevelMechanics

	// How many topics the in-pipeline measurement pass covers. Bounded because the
	// pass costs one generation call per topic plus two per question, and it runs on
	// every ingest.
	gateMeasureTopics = 10
)

// GateCondition is one §5.3 condition and whether it holds.
type GateCondition struct {
	// Name is stable and machine-readable so the UI can group or translate it.
	Name string `json:"name"`
	// Detail is the sentence shown to the admin: the numbers behind the verdict, so
	// "failed" is never the whole answer.
	Detail string `json:"detail"`
	Met    bool   `json:"met"`
}

// GateInputs is the measured state the gate reads. No judgement lives here: every
// field is a count or a flag produced by something else.
type GateInputs struct {
	Chunks int

	CharterRules        int
	ClarificationTopics int

	// MeasuredTopics counts topics whose MEASURED level reaches
	// gateRequiredMeasuredLevel. CapabilityMeasured distinguishes "measured and
	// shallow" from "never measured": they need different remedies, and the second
	// one must not read as a pass.
	MeasuredTopics     int
	CapabilityMeasured bool

	SmokeProbes   int
	SmokePassed   bool
	SmokeUncited  int
	SmokeNoAnswer int
}

// Gate condition names.
const (
	GateChunks        = "corpus_size"
	GateCharterRules  = "charter_rules"
	GateClarification = "clarification_coverage"
	GateCapability    = "measured_capability"
	GateSmoke         = "smoke_test"
)

// EvaluateIngestGate turns measured inputs into the five §5.3 verdicts, in the order
// the document lists them. Pure — unit-tested.
func EvaluateIngestGate(in GateInputs) []GateCondition {
	conditions := make([]GateCondition, 0, 5)

	conditions = append(conditions, GateCondition{
		Name: GateChunks,
		Met:  in.Chunks >= gateMinChunks,
		Detail: fmt.Sprintf("%d chunks stored (minimum %d)",
			in.Chunks, gateMinChunks),
	})

	conditions = append(conditions, GateCondition{
		Name: GateCharterRules,
		Met:  in.CharterRules >= gateMinCharterRules,
		Detail: fmt.Sprintf("%d reasoning rules in the charter (minimum %d; counted as bullet lines, a proxy for the doc's 'WHY-embedded rules')",
			in.CharterRules, gateMinCharterRules),
	})

	conditions = append(conditions, GateCondition{
		Name: GateClarification,
		Met:  in.ClarificationTopics >= gateMinClarificationTopics,
		Detail: fmt.Sprintf("%d topics have clarifying questions (minimum %d)",
			in.ClarificationTopics, gateMinClarificationTopics),
	})

	// The capability condition fails with a DIFFERENT detail when the expert was
	// never measured. "Measured and shallow" is fixed by a better corpus; "never
	// measured" is fixed by running the pass, and conflating them would send the
	// admin to the wrong remedy.
	switch {
	case !in.CapabilityMeasured:
		conditions = append(conditions, GateCondition{
			Name:   GateCapability,
			Met:    false,
			Detail: "capability has never been measured — the declared coverage band is not evidence (run a capability measurement)",
		})
	default:
		conditions = append(conditions, GateCondition{
			Name: GateCapability,
			Met:  in.MeasuredTopics >= gateMinMeasuredTopics,
			Detail: fmt.Sprintf("%d topics answer mechanics-level questions (minimum %d); measured, not declared",
				in.MeasuredTopics, gateMinMeasuredTopics),
		})
	}

	smokeDetail := fmt.Sprintf("%d of %d probes passed", in.SmokeProbes, in.SmokeProbes)
	if in.SmokeProbes > 0 {
		smokeDetail = fmt.Sprintf("%d of %d probes retrieved a matching chunk with citations",
			in.SmokeProbes-in.SmokeUncited-in.SmokeNoAnswer, in.SmokeProbes)
	}
	if !in.SmokePassed {
		smokeDetail = fmt.Sprintf("%s (needs %d passing probes)", smokeDetail, smokeTestPassThreshold)
	}
	conditions = append(conditions, GateCondition{
		Name:   GateSmoke,
		Met:    in.SmokePassed,
		Detail: smokeDetail,
	})

	return conditions
}

// GatePassed reports whether every condition holds.
func GatePassed(conditions []GateCondition) bool {
	for _, c := range conditions {
		if !c.Met {
			return false
		}
	}
	return true
}

// UnmetGateReasons renders the failed conditions as sentences for the job's warning
// text, so the admin reads which condition blocked training rather than only that
// something did.
func UnmetGateReasons(conditions []GateCondition) []string {
	var reasons []string
	for _, c := range conditions {
		if c.Met {
			continue
		}
		reasons = append(reasons, fmt.Sprintf("%s: %s", c.Name, c.Detail))
	}
	return reasons
}

// CountCharterRules counts the reasoning rules in a charter.
//
// Pure. A rule is a non-empty line that the extractor marked as a bullet ("- ",
// "* ") or numbered item ("1. ", "2) "). The charter is prose, so this is a proxy for
// §5.3's "WHY-embedded rules" — labelled as one everywhere it is reported, rather
// than dressed up as a measurement.
func CountCharterRules(reasoningCharter string) int {
	count := 0
	for _, raw := range strings.Split(reasoningCharter, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			count++
			continue
		}
		// Numbered item: digits followed by '.' or ')' and a space.
		i := 0
		for i < len(line) && line[i] >= '0' && line[i] <= '9' {
			i++
		}
		if i > 0 && i < len(line)-1 && (line[i] == '.' || line[i] == ')') && line[i+1] == ' ' {
			count++
		}
	}
	return count
}

// CapabilityMeasurer runs one capability measurement pass and reports only whether it
// succeeded.
//
// WHY a function and not the evaluator itself: measuring needs the context assembler,
// which the pipeline does not have and should not. Declaring the dependency as a
// single injected call keeps the pipeline free of the retrieval wiring while still
// making the gate depend on a measurement actually having happened.
type CapabilityMeasurer func(ctx context.Context, expertID uuid.UUID, topics int) error

// SetCapabilityMeasurer wires the measurement pass the ingest gate requires.
//
// Not wiring it is a supported state: the gate then reports that capability has never
// been measured and the expert stays in draft, which is the correct outcome for an
// environment that cannot measure.
func (p *IngestionPipeline) SetCapabilityMeasurer(m CapabilityMeasurer) {
	p.capabilityMeasurer = m
}

// collectGateInputs reads every number the gate needs.
func (p *IngestionPipeline) collectGateInputs(
	ctx context.Context,
	expertID uuid.UUID,
	charter *Charter,
	smoke SmokeTestResult,
) (GateInputs, error) {
	in := GateInputs{
		SmokeProbes:   smoke.Probes,
		SmokePassed:   smoke.Passed,
		SmokeUncited:  smoke.Uncited,
		SmokeNoAnswer: smoke.NoAnswer,
	}

	if err := p.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM course_chunks WHERE expert_id = $1`, expertID,
	).Scan(&in.Chunks); err != nil {
		return in, fmt.Errorf("gate: count chunks: %w", err)
	}

	if charter != nil {
		in.CharterRules = CountCharterRules(charter.ReasoningCharter)
		in.ClarificationTopics = len(charter.ClarificationCharter)
	}

	// Two questions, deliberately not conflated: has anything been measured at all,
	// and how many topics reach the required level. A single COUNT would make "never
	// measured" indistinguishable from "measured and shallow", and the two need
	// different remedies.
	var capabilityRows int
	if err := p.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(BOOL_OR(measured_level IS NOT NULL), FALSE)
		  FROM expert_capabilities
		 WHERE expert_id = $1`, expertID,
	).Scan(&capabilityRows, &in.CapabilityMeasured); err != nil {
		return in, fmt.Errorf("gate: read capabilities: %w", err)
	}
	if in.CapabilityMeasured {
		if err := p.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM expert_capabilities
			 WHERE expert_id = $1 AND measured_level >= $2`,
			expertID, gateRequiredMeasuredLevel,
		).Scan(&in.MeasuredTopics); err != nil {
			return in, fmt.Errorf("gate: count measured topics: %w", err)
		}
	}

	return in, nil
}
