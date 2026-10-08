package gateway

import (
	"context"
	"errors"
	"net"
	"strings"
)

// IsInfraError classifies provider infra vs logical errors.
// Additive-only: no change to StreamCall / generateFlat core.
// Infra = 429 / 5xx / timeout / network / context deadline -> map to
// ExpertResponse{Mode: REFUSE, Reason:"provider unavailable"} for red card.
// Logical = 400 / 403 / 422 / gate REFUSE / ASK -> normal non-retryable.
func IsInfraError(err error) bool {
	if err == nil {
		return false
	}
	// Context timeout / deadline is infra (provider hung)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		// Canceled can be ghost-hang vs user cancel; treat as infra only
		// when wrapped with timeout/network text. Pure user cancel is not infra retry.
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "deadline") || strings.Contains(msg, "timeout") || strings.Contains(msg, "context") {
			// If caller propagated c.Request.Context(), user disconnect will also be Canceled.
			// Caller (orchestrator) checks ctx.Done() first to avoid misclassifying user cancel as infra.
			return true
		}
		return false
	}
	msg := strings.ToLower(err.Error())
	// Network / timeout substrings
	if strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "eof") && strings.Contains(msg, "unexpected") ||
		strings.Contains(msg, "tls handshake") ||
		strings.Contains(msg, "network is unreachable") {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	// HTTP status based
	// StreamCall wraps status as "... status 429 ..." / "... status 5xx ..."
	if strings.Contains(msg, "status 429") ||
		strings.Contains(msg, "status 500") ||
		strings.Contains(msg, "status 502") ||
		strings.Contains(msg, "status 503") ||
		strings.Contains(msg, "status 504") ||
		strings.Contains(msg, " 429 ") ||
		strings.Contains(msg, " 500 ") ||
		strings.Contains(msg, " 502 ") ||
		strings.Contains(msg, " 503 ") ||
		strings.Contains(msg, " 504 ") {
		return true
	}
	if strings.Contains(msg, "rate limit") || strings.Contains(msg, "overloaded") || strings.Contains(msg, "temporarily unavailable") {
		return true
	}
	return false
}

// IsRetryableInfra returns true only for infra that safe to retry via ResumeRunID.
// 429 and 5xx and timeout/network are retryable; logical 4xx are not.
func IsRetryableInfra(err error) bool {
	return IsInfraError(err)
}
