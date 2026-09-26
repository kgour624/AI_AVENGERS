package capability

import "fmt"

// Requirement is what a phase needs from the expert set.
type Requirement struct {
	Kind     Kind
	Language string // only meaningful when Kind is implementation/testing
}

// Phase constants mirror workflow's phases without importing it (the workflow
// package imports this one, and a cycle would force the mapping into the runner
// itself, which is exactly the hardcoding this package exists to avoid).
const (
	PhaseHighLevelDesign = "high_level_design"
	PhaseDetailedDesign  = "detailed_design"
	PhaseImplementation  = "implementation"
	PhaseQA              = "qa"
	PhaseHandoff         = "handoff"
)

// RequirementForPhase returns what a phase needs. Design and handoff need
// design; code needs an implementation expert; QA needs testing. Anything
// unknown returns an empty Kind, and callers must treat that as "no
// requirement" rather than "any expert will do".
func RequirementForPhase(phase, language string) Requirement {
	switch phase {
	case PhaseHighLevelDesign, PhaseDetailedDesign, PhaseHandoff:
		return Requirement{Kind: KindDesign}
	case PhaseImplementation:
		return Requirement{Kind: KindImplementation, Language: LanguageAliases(language)}
	case PhaseQA:
		return Requirement{Kind: KindTesting, Language: LanguageAliases(language)}
	}
	return Requirement{}
}

// Expert is the minimum this package needs to know about an expert.
type Expert struct {
	ID   string
	Name string
	// Declaration is the expert's declared capability (from its domain profile).
	Declaration Declaration
}

// MissingExpertError is returned when a phase cannot run because no expert
// matches its requirement. It is intentionally an error, not a silent skip:
// running a design expert on a code task produced the incident this package
// fixes, and a client must be told which expert to add.
type MissingExpertError struct {
	Phase       string
	Requirement Requirement
	Have        []Kind
}

func (e *MissingExpertError) Error() string {
	need := string(e.Requirement.Kind)
	if e.Requirement.Language != "" {
		need = fmt.Sprintf("%s (%s)", need, e.Requirement.Language)
	}
	if len(e.Have) == 0 {
		return fmt.Sprintf("%s phase needs a %s expert, but the workflow has no experts with a declared capability", e.Phase, need)
	}
	have := make([]string, 0, len(e.Have))
	for _, k := range e.Have {
		if k == KindUnclassified {
			have = append(have, "unclassified")
			continue
		}
		have = append(have, string(k))
	}
	return fmt.Sprintf("%s phase needs a %s expert, but the workflow only has: %s", e.Phase, need, join(have))
}

func join(v []string) string {
	out := ""
	for i, s := range v {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

// SelectForPhase returns the experts that may run a phase.
//
// Matching rules, in order:
//  1. The expert's declared kind must equal the requirement's kind.
//  2. When the requirement names a language, an expert that declares a
//     DIFFERENT language is excluded — a Python programmer must not be handed a
//     Go task. Experts that declare no language are allowed through only when
//     the requirement has no language, so "implementation" never quietly means
//     "any language".
//
// A phase with no requirement (empty Kind) selects nobody, so an unknown phase
// cannot fall back to "all experts".
func SelectForPhase(phase string, experts []Expert, language string) ([]Expert, error) {
	req := RequirementForPhase(phase, language)
	if req.Kind == KindUnclassified {
		return nil, nil
	}

	var matched []Expert
	have := make([]Kind, 0, len(experts))
	for _, e := range experts {
		have = append(have, e.Declaration.Kind)
		if e.Declaration.Kind != req.Kind {
			continue
		}
		if req.Kind == KindImplementation || req.Kind == KindTesting {
			if req.Language != "" && e.Declaration.Language != "" && e.Declaration.Language != req.Language {
				continue
			}
			if req.Language != "" && e.Declaration.Language == "" {
				// An implementation expert with no language cannot be trusted to
				// own a language-specific task.
				continue
			}
		}
		matched = append(matched, e)
	}
	if len(matched) == 0 {
		return nil, &MissingExpertError{Phase: phase, Requirement: req, Have: have}
	}
	return matched, nil
}
