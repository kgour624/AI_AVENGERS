package message

import (
	"net/http"

	"github.com/google/uuid"
	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/decision"
	"ai_avengers/backend/internal/gateway"
	"ai_avengers/backend/internal/orchestrator"
)

// SSE event types for failure UX (additive, no core generation change)
const (
	SSEStepFailed  = "relay_step"   // StepExpertFailed with retryable flag
	SSERelayFailed = "relay_failed" // Run-level failure with ResumeRunID
)

// Failure SSE payloads

type StepFailedPayload struct {
	Step       string `json:"step"` // "step_expert_failed"
	ExpertID   string `json:"expert_id"`
	ExpertName string `json:"expert_name"`
	Index      int    `json:"index"`
	Reason     string `json:"reason"`
	Retryable  bool   `json:"retryable"`
}

type RelayFailedPayload struct {
	RunID         string `json:"run_id"`
	FailedAtIndex int    `json:"failed_at_index"`
	Reason        string `json:"reason"`
	RetryUntil    string `json:"retry_until"` // RFC3339 deadline for retry
	Retryable     bool   `json:"retryable"`
}

// ensureIndependentRedCards synthesizes a REFUSE red card for any requested
// expert that did not return a response (silent drop -> visible card).
// Additive-only: preserves all successful responses, only fills gaps.
func ensureIndependentRedCards(requested []uuid.UUID, resp *orchestrator.OrchestratorResponse, expertNames map[uuid.UUID]string, domains map[uuid.UUID]string) {
	if resp == nil {
		return
	}
	seen := make(map[uuid.UUID]bool, len(resp.ExpertResponses))
	for _, r := range resp.ExpertResponses {
		seen[r.ExpertID] = true
	}
	for _, want := range requested {
		if seen[want] {
			continue
		}
		// Synthesize infra REFUSE card - visible red error, not silent drop
		name := expertNames[want]
		if name == "" {
			name = want.String()
		}
		domain := domains[want]
		synth := orchestrator.ExpertResponse{
			ExpertID:   want,
			ExpertName: name,
			Domain:     domain,
			Mode:       decision.ModeREFUSE,
			Content:    "Ye Expert abhi network problem ki wajah se jawab nahi de paaya. Kripya dobara try karein.",
			Reason:     "provider unavailable",
			Citations:  []chinawall.Citation{},
			Confidence: 0,
			Error:      "provider unavailable: synthesized missing expert",
		}
		resp.ExpertResponses = append(resp.ExpertResponses, synth)
	}
	// Also normalize any infra error that came back as Error but not REFUSE
	for i, r := range resp.ExpertResponses {
		if r.Error != "" && gateway.IsInfraErrorString(r.Error) {
			if r.Mode != decision.ModeREFUSE {
				resp.ExpertResponses[i].Mode = decision.ModeREFUSE
			}
			if resp.ExpertResponses[i].Reason == "" {
				resp.ExpertResponses[i].Reason = "provider unavailable"
			}
			if resp.ExpertResponses[i].Content == "" {
				resp.ExpertResponses[i].Content = "Ye Expert abhi network problem ki wajah se jawab nahi de paaya. Kripya dobara try karein."
			}
			if resp.ExpertResponses[i].Citations == nil {
				resp.ExpertResponses[i].Citations = []chinawall.Citation{}
			}
		}
	}
}

// emitIndependentFailures ensures every expert (including failed) sends SSEComplete
// so frontend renders a card per row. Must be called inside the SSE forwarder
// after orchestrator returns. Also guarantees citations != nil to avoid frontend crash.
func emitIndependentFailures(w http.ResponseWriter, resp *orchestrator.OrchestratorResponse) {
	if resp == nil {
		return
	}
	for _, er := range resp.ExpertResponses {
		citations := er.Citations
		if citations == nil {
			citations = []chinawall.Citation{}
		}
		// Always emit complete per expert - red card included
		sendSSE(w, SSEComplete, map[string]interface{}{
			"expert_id":         er.ExpertID,
			"expert_name":       er.ExpertName,
			"domain":            er.Domain,
			"mode":              er.Mode,
			"content":           er.Content,
			"citations":         citations,
			"confidence":        er.Confidence,
			"gate_stopped":      er.GateStopped,
			"warning":           er.Warning,
			"questions":         er.Questions,
			"template_sections": er.TemplateSections,
			"claims":            er.Claims,
			"reason":            er.Reason,
			"error":             er.Error,
		})
	}
}

// emitCollaborativeFailure emits StepFailed + RelayFailed for collaborative hard-stop.
// Additive: does not change core expert generation, only SSE emission + retry affordance.
// Caller must have already done store.MarkRunFailed + SetSectionFailure.
func emitCollaborativeFailure(w http.ResponseWriter, runID uuid.UUID, failedAtIndex int, reason string, retryUntil string, expertID uuid.UUID, expertName string) {
	// Step-level failure (per-expert in relay)
	sendSSE(w, SSEStepFailed, StepFailedPayload{
		Step:       "step_expert_failed",
		ExpertID:   expertID.String(),
		ExpertName: expertName,
		Index:      failedAtIndex,
		Reason:     reason,
		Retryable:  true,
	})
	// Run-level failure with ResumeRunID
	sendSSE(w, SSERelayFailed, RelayFailedPayload{
		RunID:         runID.String(),
		FailedAtIndex: failedAtIndex,
		Reason:        reason,
		RetryUntil:    retryUntil,
		Retryable:     true,
	})
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
