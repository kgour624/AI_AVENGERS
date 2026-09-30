package mcpv2

// Elicitation — confirmation form (MCP 3&4 spec).
// 4 states: accept+true / accept+false / decline / cancel
// Uses server. elicitation/request pattern.

type ElicitationRequest struct {
	Message         string         `json:"message"`
	RequestedSchema map[string]any `json:"requestedSchema"`
}

type ElicitationResult struct {
	Action string         `json:"action"` // accept | decline | cancel
	Content map[string]any `json:"content,omitempty"`
}

// NeedsConfirmation returns true for destructive tools (delete_expert, revoke rental).
func NeedsConfirmation(toolName string) bool {
	switch toolName {
	case "delete_expert", "revoke_rental", "revoke_repo", "apply_patch", "fix_error":
		return true
	default:
		return false
	}
}

// ParseElicitation handles 4-state result.
func ParseElicitation(r ElicitationResult) (confirmed bool, shouldProceed bool) {
	switch r.Action {
	case "accept":
		if v, ok := r.Content["confirmed"].(bool); ok {
			return v, v
		}
		return false, false
	case "decline", "cancel":
		return false, false
	default:
		return false, false
	}
}