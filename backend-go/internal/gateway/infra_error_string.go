package gateway

import "strings"

// IsInfraErrorString is a convenience for callers that only have an error string
// (e.g., orchestrator ExpertResponse.Error). Avoids recreating error objects.
func IsInfraErrorString(s string) bool {
	if s == "" {
		return false
	}
	m := strings.ToLower(s)
	return strings.Contains(m, "timeout") ||
		strings.Contains(m, "deadline") ||
		strings.Contains(m, "connection reset") ||
		strings.Contains(m, "connection refused") ||
		strings.Contains(m, "no such host") ||
		strings.Contains(m, "status 429") ||
		strings.Contains(m, "status 500") ||
		strings.Contains(m, "status 502") ||
		strings.Contains(m, "status 503") ||
		strings.Contains(m, "status 504") ||
		strings.Contains(m, " 429 ") ||
		strings.Contains(m, " 500 ") ||
		strings.Contains(m, "rate limit") ||
		strings.Contains(m, "overloaded") ||
		strings.Contains(m, "temporarily unavailable") ||
		strings.Contains(m, "provider unavailable")
}
