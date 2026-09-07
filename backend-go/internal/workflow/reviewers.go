package workflow

// ReviewerRule defines who must and who may review an artifact type.
type ReviewerRule struct {
	// MandatoryReviewers are domains that MUST post an approved review
	// before the artifact can be marked final.
	// If any mandatory reviewer says blocked → escalate to client.
	// If any says changes_requested → revision round (max 3).
	MandatoryReviewers []string

	// OptionalReviewers may review but their verdict is not blocking.
	OptionalReviewers []string
}

// ReviewerMatrix maps artifact event_type → ReviewerRule.
// Source: DOMAIN_EXPERT_COLLABORATION_DESIGN.md §8.1
//
// Domain strings must match the `domain` column in the experts table.
var ReviewerMatrix = map[string]ReviewerRule{
	"requirement_captured": {
		// Client approves via approval gate — no expert mandatory reviewers.
		// The workflow engine handles this via AskClient, not this matrix.
		MandatoryReviewers: []string{},
		OptionalReviewers:  []string{},
	},
	"architecture_decision": {
		MandatoryReviewers: []string{"security", "lld"},
		OptionalReviewers:  []string{"database", "devops"},
	},
	"data_model_proposed": {
		MandatoryReviewers: []string{"system_design", "backend"},
		OptionalReviewers:  []string{"security"},
	},
	"api_contract_proposed": {
		MandatoryReviewers: []string{"frontend", "security"},
		OptionalReviewers:  []string{"qa"},
	},
	"module_design_proposed": {
		MandatoryReviewers: []string{"code_review"},
		OptionalReviewers:  []string{"backend", "frontend"},
	},
	"code_artifact_produced": {
		MandatoryReviewers: []string{"code_review", "qa"},
		OptionalReviewers:  []string{"security"},
	},
	"test_case_proposed": {
		// Owner of the code under test must approve.
		// Determined at runtime by the workflow engine based on which
		// expert produced the artifact being tested.
		// Stored here as a placeholder; engine resolves the actual domain.
		MandatoryReviewers: []string{"_artifact_owner"},
		OptionalReviewers:  []string{},
	},
}

// GetReviewerRule returns the ReviewerRule for an artifact event_type.
// Returns an empty rule (no reviewers) for unknown event types.
func GetReviewerRule(eventType string) ReviewerRule {
	rule, ok := ReviewerMatrix[eventType]
	if !ok {
		return ReviewerRule{}
	}
	return rule
}

// NeedsReview returns true if the event_type has at least one mandatory reviewer.
func NeedsReview(eventType string) bool {
	rule := GetReviewerRule(eventType)
	return len(rule.MandatoryReviewers) > 0
}

// MaxRevisionRounds is the maximum number of revision rounds before
// escalating to the client. Source: DOMAIN_EXPERT_COLLABORATION_DESIGN.md §8.3
const MaxRevisionRounds = 3
