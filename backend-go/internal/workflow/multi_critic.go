package workflow

import (
	"fmt"
	"strings"

	"ai_avengers/backend/internal/gateway"
)

// MultiCriticPolicy decides whether an artifact is reviewed by more than one
// critic, and by which model the second one speaks.
//
// WHY this exists (A5 / C7): a review is the gate that decides whether an
// artifact ships, and until now that gate was ONE call to the CHEAPEST model —
// the artifact was written by a strong one and judged by a weaker one. A single
// reviewer also has no way to signal doubt: it either approves or it does not, so
// a reviewer that half-noticed something wrong approves anyway and the doubt is
// lost. Asking twice, on deliberately different models, surfaces exactly that.
type MultiCriticPolicy struct {
	// Enabled: the kill switch. false restores the single-critic review exactly.
	Enabled bool
	// SecondCritic: the model tier the second opinion uses. Empty → ModelFast.
	// Deliberately NOT the same tier as the first critic: two samples from one
	// model share its blind spots, which is the thing this is meant to catch.
	SecondCritic gateway.ModelType
}

// DefaultMultiCriticPolicy is what a deployment gets when it configures nothing.
//
// Enabled by default, and that is a judgement: the cost is one extra review call
// per artifact per round, while the thing it protects is whether wrong work is
// accepted. The second critic uses the middle tier — strong enough to disagree
// with substance, cheap enough to run on every artifact.
func DefaultMultiCriticPolicy() MultiCriticPolicy {
	return MultiCriticPolicy{Enabled: true, SecondCritic: gateway.ModelFast}
}

// CriticOpinion is one critic's verdict, kept individually so a disagreement can
// be shown instead of averaged away.
type CriticOpinion struct {
	// Critic names who spoke, Model names which model said it.
	Critic  string `json:"critic"`
	Model   string `json:"model"`
	Status  string `json:"status"`
	Comment string `json:"comment,omitempty"`
}

// combineCriticVerdicts turns several opinions into the one verdict the workflow
// acts on.
//
// Pure — unit-tested, because this is the rule that decides whether work is
// accepted. The policy, stated plainly:
//
//   - any BLOCKED → blocked. One critic seeing a reason for the client to decide
//     is enough; a second opinion must never overrule that.
//   - any CHANGES_REQUESTED → changes_requested. A split is NOT an approval: one
//     critic saying something is wrong is evidence that a single reader would
//     have missed it, and "one of two approved" is the weakest possible reason to
//     ship.
//   - all approved → approved.
//   - no opinions at all → no verdict. It returns empty strings so the caller
//     treats it as a failure rather than reading silence as consent.
func combineCriticVerdicts(opinions []CriticOpinion) (status string, comment string) {
	if len(opinions) == 0 {
		return "", ""
	}

	var blockers, changers []CriticOpinion
	approved := 0
	for _, o := range opinions {
		switch o.Status {
		case "blocked":
			blockers = append(blockers, o)
		case "changes_requested":
			changers = append(changers, o)
		default:
			// Any other value — including approved — only counts once it has been
			// seen as approved below, so an unknown status can never act as a yes.
			if o.Status == "approved" {
				approved++
			} else {
				changers = append(changers, CriticOpinion{
					Critic:  o.Critic,
					Model:   o.Model,
					Status:  "changes_requested",
					Comment: "review verdict could not be read: " + strings.TrimSpace(o.Comment),
				})
			}
		}
	}

	if len(blockers) > 0 {
		return "blocked", describeDissentingCritics(blockers)
	}
	if len(changers) > 0 {
		return "changes_requested", describeDissentingCritics(changers)
	}
	if approved == len(opinions) {
		return "approved", ""
	}
	// Unreachable in practice (every branch above consumes a status), but a
	// missing verdict must not become approval.
	return "", ""
}

// describeDissentingCritics names who objected and why, so the screen and the
// revision prompt both carry the disagreement instead of a summary that hides it.
func describeDissentingCritics(opinions []CriticOpinion) string {
	parts := make([]string, 0, len(opinions))
	for _, o := range opinions {
		label := o.Critic
		if o.Model != "" {
			label = fmt.Sprintf("%s (%s)", o.Critic, o.Model)
		}
		if strings.TrimSpace(o.Comment) == "" {
			parts = append(parts, label)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", label, strings.TrimSpace(o.Comment)))
	}
	return strings.Join(parts, " | ")
}
