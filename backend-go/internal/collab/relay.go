package collab

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// RunExpert is the function relay.go calls once per section, sequentially.
// orchestrator.go supplies this as a closure over its own
// processWithExpert, so collab has zero direct dependency on the
// orchestrator package (avoids an import cycle: orchestrator imports
// collab, never the other way around).
//
// peerContext is "" for the first section (no one has answered yet) and,
// for every section after that, the accumulated content of the sections
// already run in this relay, pre-formatted by RunRelay — the same shape
// orchestrator.processWithExpert already merges into its reply-thread
// context, so collaborative mode needs no new plumbing in
// decision/chinawall.
//
// content is the expert's finished answer text to put in that section.
// err non-nil means this expert's turn failed outright (timeout, rate
// limit, decision engine error, ...) — RunRelay still includes the
// section with a placeholder so the caller can assemble a partial answer
// instead of discarding the whole collaborative response, matching
// processWithExpert's existing per-expert fail-soft posture.
// StepName is the transparency-log taxonomy. The frontend renders these live,
// strictly in arrival order (ans6).
type StepName string

const (
	StepPlan              StepName = "plan"
	StepExpertStarted     StepName = "expert_started"
	StepRetrievingContext StepName = "retrieving_context"
	StepChinaWallCheck    StepName = "china_wall_check"
	StepGeneratingAnswer  StepName = "generating_answer"
	StepExpertCompleted   StepName = "expert_completed"
	StepExpertFailed      StepName = "expert_failed"
	StepCheckingConsistency StepName = "checking_consistency"
	StepConsistencyChecked  StepName = "consistency_checked"
	StepAwaitingReview      StepName = "awaiting_human_review"
	StepHumanApproved       StepName = "human_approved"
	StepRelayFailed         StepName = "relay_failed"
	StepRelayCompleted      StepName = "relay_completed"
)

// StepEvent is one transparency event, emitted live AND persisted to
// collab_relay_events (so reconnect can replay the exact same transcript).
type StepEvent struct {
	Step       StepName       `json:"step"`
	ExpertID   *uuid.UUID     `json:"expert_id,omitempty"`
	ExpertName string         `json:"expert_name,omitempty"`
	Message    string         `json:"message"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// StepFunc is called for every transparency event, in order, as the relay runs.
type StepFunc func(ev StepEvent)

// AppendPeerContext appends a finished section to the accumulated peer-context
// builder, truncated the same way RunRelay already truncates (1500 chars), so
// the next expert is oriented, not given the whole answer verbatim.
func AppendPeerContext(sb *strings.Builder, section Section) {
	peerContent := section.Content
	if len(peerContent) > 1500 {
		peerContent = peerContent[:1500] + "\n... [truncated]"
	}
	fmt.Fprintf(sb, "--- %s (%s) already answered ---\n%s\n\n",
		section.ExpertName, section.SectionTitle, peerContent)
}

// FormatPeerContext is the exported wrapper of the internal formatPeerContext,
// so orchestrator.ProcessCollaborative can reuse the exact same wrapping text
// RunRelay already uses (keeps peer-context identical across both loops).
func FormatPeerContext(peer string) string { return formatPeerContext(peer) }

type RunExpert func(ctx context.Context, expertID uuid.UUID, peerContext string) (content string, err error)

// ProgressFunc is called once a section's content is ready, in relay
// order, so the caller can stream it over SSE as it completes rather than
// waiting for the whole relay to finish. nil is accepted (no progress
// reporting, e.g. in tests).
type ProgressFunc func(section Section, index, total int)

// RelayResult is the sequential relay's output: every section with its
// Content filled in (in final answer order), plus any experts whose turn
// failed outright (so the caller can still assemble a partial answer
// instead of discarding the whole collaborative response, matching the
// orchestrator's existing per-expert fail-soft posture in processWithExpert).
type RelayResult struct {
	Sections []Section
	// Failed maps expert_id -> error message for any section whose
	// RunExpert call returned an error. That section is still included in
	// Sections with a short "could not answer" placeholder Content, so
	// assemble.go always has every planned section to render — never a
	// silently missing expert.
	Failed map[string]string
}

// RunRelay executes the plan sequentially: expert 1 answers first with no
// peer context (same as it would get in independent mode), expert 2 answers
// next with expert 1's finished section injected as peer context, and so
// on. This is the core of Collaborative Relay — later experts can build on
// (agree with, extend, or flag a conflict with) earlier sections instead of
// answering in total isolation.
//
// WHY sequential (not parallel + merge, i.e. not just orchestrator.synthesize):
// synthesize already exists and runs AFTER independent answers are final —
// it summarizes agreements/contradictions but cannot change what either
// expert said, because by the time it runs both answers are already
// generated. The whole point of relay mode is that expert 2's ANSWER ITSELF
// is shaped by having read expert 1's section first — that requires running
// expert 2 strictly after expert 1 completes, not in a goroutine alongside it.
// NOTE (ans2 — hard stop): the `continue`-on-error fail-soft behavior below is
// superseded for the production chat path by orchestrator.ProcessCollaborative,
// which stops the whole relay at the first failed expert. RunRelay is retained
// only for its existing unit tests; production never calls it anymore. If you
// re-wire this, prefer hard-stop semantics.
func RunRelay(
	ctx context.Context,
	plan []Section,
	run RunExpert,
	onProgress ProgressFunc,
) RelayResult {
	result := RelayResult{
		Sections: make([]Section, 0, len(plan)),
		Failed:   make(map[string]string),
	}

	var peerContext strings.Builder
	for i, section := range plan {
		content, err := run(ctx, section.ExpertID, formatPeerContext(peerContext.String()))
		if err != nil {
			section.Content = fmt.Sprintf("This expert could not answer: %s", err.Error())
			result.Failed[section.ExpertID.String()] = err.Error()
			result.Sections = append(result.Sections, section)
			if onProgress != nil {
				onProgress(section, i, len(plan))
			}
			return result
		}

		section.Content = content
		result.Sections = append(result.Sections, section)

		// Append this section to the peer context BEFORE the next
		// iteration, so expert i+1 sees it. Truncated the same way
		// orchestrator.synthesize truncates each expert's content (1500
		// chars) — peer context is meant to orient the next expert, not
		// reproduce the entire answer verbatim in every subsequent prompt.
		peerContent := content
		if len(peerContent) > 1500 {
			peerContent = peerContent[:1500] + "\n... [truncated]"
		}
		fmt.Fprintf(&peerContext, "--- %s (%s) already answered ---\n%s\n\n", section.ExpertName, section.SectionTitle, peerContent)

		if onProgress != nil {
			onProgress(section, i, len(plan))
		}
	}

	return result
}

// formatPeerContext wraps the accumulated peer-section text (built up by
// RunRelay as sections complete) with instructions for the next expert.
// Returns "" when no section has completed yet (first expert in the
// relay), so RunExpert's peerContext argument is "" for that first call,
// exactly like the independent path's unchanged behavior.
func formatPeerContext(peer string) string {
	peer = strings.TrimSpace(peer)
	if peer == "" {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## COLLABORATIVE RELAY (other experts already answered this SAME question, in this order):\n")
	sb.WriteString(peer)
	sb.WriteString("Build on what they said where relevant — agree, add missing detail from your own domain, or clearly flag if you see a conflict. Do not simply repeat their content.\n")
	return sb.String()
}

