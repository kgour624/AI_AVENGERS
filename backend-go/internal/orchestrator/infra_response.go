package orchestrator

import (
	"github.com/google/uuid"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/gateway"
)

// infraRefuseResponse builds the per-expert red card for Independent mode.
// Additive-only: does not change core expert generation; only maps infra errors
// (429/5xx/timeout/network) to a visible REFUSE card so frontend never silently drops.
// Content is Hindi user-facing, Reason is machine tag "provider unavailable".
func infraRefuseResponse(expertID uuid.UUID, expertName, domain string, infraErr error) ExpertResponse {
	msg := "Ye Expert abhi network problem ki wajah se jawab nahi de paaya. Kripya dobara try karein."
	if gateway.IsInfraError(infraErr) {
		// keep Hindi content, machine reason for frontend red card
	}
	return ExpertResponse{
		ExpertID:   expertID,
		ExpertName: expertName,
		Domain:     domain,
		Mode:       decision.ModeREFUSE,
		Content:    msg,
		Reason:     "provider unavailable",
		Citations:  []chinawall.Citation{},
		Confidence: 0,
		Error:      infraErr.Error(),
	}
}

// isInfra maps any error to infra classification via gateway.IsInfraError
func isInfra(err error) bool {
	return gateway.IsInfraError(err)
}
