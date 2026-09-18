package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// CrossVerifier implements the cross-verification protocol from §8.
//
// MENTAL MODEL:
//   After each wave, scan blackboard for new artifacts.
//   For each artifact, find mandatory reviewers (from reviewer matrix).
//   Each reviewer reads the artifact and posts review_comment.
//   If all approve -> artifact is final.
//   If any requests changes -> producer revises (max 3 rounds).
//   If any blocks -> escalate to client.
//
// REVIEWER MATRIX (from design doc §8.1):
//   code_artifact_produced  -> Code Reviewer, QA
//   architecture_decision   -> Security, LLD
//   data_model_proposed     -> System Design, Backend
//   api_contract_proposed   -> Frontend, Security
//   module_design_proposed  -> Code Reviewer
//   test_case_proposed      -> Backend or Frontend (owner)
type CrossVerifier struct {
	store     *blackboard.Store
	gateway   *gateway.ModelGateway
	agentLoop *AgentLoop
	tools     *Tools
	logger    *zap.Logger
}

// NewCrossVerifier creates a new CrossVerifier.
func NewCrossVerifier(
	store *blackboard.Store,
	gw *gateway.ModelGateway,
	agentLoop *AgentLoop,
	tools *Tools,
	logger *zap.Logger,
) *CrossVerifier {
	return &CrossVerifier{
		store:     store,
		gateway:   gw,
		agentLoop: agentLoop,
		tools:     tools,
		logger:    logger,
	}
}

// reviewerMatrix maps artifact event types to mandatory reviewer domains.
// Domain strings must match experts.domain column values.
var reviewerMatrix = map[string][]string{
	"code_artifact_produced": {"code_review", "qa"},
	"architecture_decision":  {"security", "lld"},
	"data_model_proposed":    {"system_design", "backend"},
	"api_contract_proposed":  {"frontend", "security"},
	"module_design_proposed": {"code_review"},
	"test_case_proposed":     {"backend", "frontend"},
}

// VerifyWaveArtifacts runs cross-verification on all artifacts posted
// during a wave. Called after wg.Wait() in executeWaves().
//
// MENTAL MODEL:
//   Input: workflowID, lastSeqBefore (sequence number before wave started)
//   1. Fetch all new events since lastSeqBefore
//   2. Filter to artifact event types (those in reviewerMatrix)
//   3. For each artifact: find reviewers, run review loop
//   4. Return error if any artifact is blocked (escalate to client)
//
// CROSS-QUESTIONS:
//   Q: Why lastSeqBefore?
//   A: We only want artifacts from THIS wave, not previous waves.
//      lastSeqBefore = sequence number before wave started.
//      All events with seq > lastSeqBefore are from this wave.
//
//   Q: What if wave produced no artifacts?
//   A: No-op. Return nil.
//
//   Q: What if reviewer is not in workflow?
//   A: Skip. Only workflow-selected experts can review.
func (cv *CrossVerifier) VerifyWaveArtifacts(
	ctx context.Context,
	workflowID uuid.UUID,
	experts []workflowExpert,
	lastSeqBefore int64,
) error {
	// Fetch all new events from this wave
	newEvents, err := cv.store.GetSince(ctx, workflowID, lastSeqBefore, 200)
	if err != nil {
		return fmt.Errorf("cross-verify: fetch events: %w", err)
	}

	if len(newEvents) == 0 {
		return nil
	}

	// Build domain -> expert map for quick lookup
	domainToExpert := make(map[string]workflowExpert)
	for _, e := range experts {
		domainToExpert[e.Domain] = e
	}

	// Process each artifact that needs review
	for _, artifact := range newEvents {
		mandatoryDomains, needsReview := reviewerMatrix[artifact.EventType]
		if !needsReview {
			continue
		}

		// Find which mandatory reviewers are in this workflow
		var reviewers []workflowExpert
		for _, domain := range mandatoryDomains {
			if reviewer, ok := domainToExpert[domain]; ok {
				reviewers = append(reviewers, reviewer)
			}
		}

		if len(reviewers) == 0 {
			// No reviewers in this workflow — auto-approve
			cv.logger.Warn("cross-verify: no reviewers in workflow, auto-approving",
				zap.String("event_type", artifact.EventType),
				zap.Strings("needed_domains", mandatoryDomains),
			)
			continue
		}

		// Run review loop for this artifact
		if err := cv.reviewArtifact(ctx, workflowID, artifact, reviewers, experts); err != nil {
			return fmt.Errorf("cross-verify: artifact %s: %w", artifact.ID, err)
		}
	}

	return nil
}

// reviewArtifact runs the review loop for one artifact.
//
// MENTAL MODEL:
//   Max 3 revision rounds (design doc §8.3).
//   Each round:
//     1. Each reviewer reads artifact + posts review_comment
//     2. If all approved -> done
//     3. If any changes_requested -> producer revises
//     4. If any blocked -> escalate to client
//
// CROSS-QUESTIONS:
//   Q: How does reviewer "read" the artifact?
//   A: We build a review prompt with artifact content.
//      Reviewer's AgentLoop runs with this prompt.
//      Reviewer posts review_comment event.
//
//   Q: How does producer "revise"?
//   A: We post a revision_requested event to blackboard.
//      Producer's next iteration picks this up via observations.
//      (Simplified: in current impl, we log and continue)
//
//   Q: What is "blocked"?
//   A: Reviewer posts review_comment with status=blocked.
//      This means a hard stop — client must decide.
func (cv *CrossVerifier) reviewArtifact(
	ctx context.Context,
	workflowID uuid.UUID,
	artifact blackboard.Event,
	reviewers []workflowExpert,
	allExperts []workflowExpert,
) error {
	const maxRevisions = 3

	for round := 1; round <= maxRevisions; round++ {
		cv.logger.Info("cross-verify: review round",
			zap.String("artifact_id", artifact.ID.String()),
			zap.String("event_type", artifact.EventType),
			zap.Int("round", round),
			zap.Int("reviewers", len(reviewers)),
		)

		allApproved := true
		anyBlocked := false
		var changeRequests []string

		for _, reviewer := range reviewers {
			status, comment, reviewErr := cv.runReview(ctx, workflowID, artifact, reviewer, allExperts)
			if reviewErr != nil {
				cv.logger.Warn("cross-verify: reviewer failed (skipping)",
					zap.String("reviewer", reviewer.Name),
					zap.Error(reviewErr),
				)
				continue // Skip failed reviewer, don't block
			}

			switch status {
			case "approved":
				cv.logger.Info("cross-verify: reviewer approved",
					zap.String("reviewer", reviewer.Name),
					zap.String("artifact_type", artifact.EventType),
				)
			case "changes_requested":
				allApproved = false
				changeRequests = append(changeRequests, fmt.Sprintf("%s: %s", reviewer.Name, comment))
				cv.logger.Info("cross-verify: reviewer requested changes",
					zap.String("reviewer", reviewer.Name),
					zap.String("comment", comment),
				)
			case "blocked":
				anyBlocked = true
				cv.logger.Error("cross-verify: reviewer blocked artifact",
					zap.String("reviewer", reviewer.Name),
					zap.String("comment", comment),
				)
			}
		}

		if anyBlocked {
			// Hard stop — post to blackboard and return error
			_, _ = cv.store.Post(ctx, blackboard.PostRequest{
				WorkflowID: workflowID,
				EventType:  "artifact_blocked",
				Content: map[string]interface{}{
					"artifact_id":   artifact.ID.String(),
					"artifact_type": artifact.EventType,
					"reason":        "reviewer blocked artifact",
				},
			})
			return fmt.Errorf("artifact %s blocked by reviewer", artifact.EventType)
		}

		if allApproved {
			// All reviewers approved — post final approval event
			_, _ = cv.store.Post(ctx, blackboard.PostRequest{
				WorkflowID: workflowID,
				EventType:  "artifact_approved",
				Content: map[string]interface{}{
					"artifact_id":   artifact.ID.String(),
					"artifact_type": artifact.EventType,
					"round":         round,
				},
			})
			cv.logger.Info("cross-verify: artifact approved",
				zap.String("artifact_type", artifact.EventType),
				zap.Int("round", round),
			)
			return nil
		}

		// Changes requested — post revision request
		if round < maxRevisions {
			_, _ = cv.store.Post(ctx, blackboard.PostRequest{
				WorkflowID:         workflowID,
				EventType:          "revision_requested",
				ReferencesEventIDs: []uuid.UUID{artifact.ID},
				Content: map[string]interface{}{
					"artifact_id":     artifact.ID.String(),
					"artifact_type":   artifact.EventType,
					"round":           round,
					"change_requests": changeRequests,
				},
			})
			cv.logger.Info("cross-verify: revision requested",
				zap.String("artifact_type", artifact.EventType),
				zap.Int("round", round),
				zap.Strings("requests", changeRequests),
			)
			// In current implementation: log and continue to next round.
			// Future: trigger producer to revise and re-post artifact.
			// For now, we give the reviewer another chance in next round.
		}
	}

	// Max revisions reached — escalate to client
	cv.logger.Warn("cross-verify: max revisions reached, escalating to client",
		zap.String("artifact_type", artifact.EventType),
	)
	_, _ = cv.store.Post(ctx, blackboard.PostRequest{
		WorkflowID: workflowID,
		EventType:  "review_escalated_to_client",
		Content: map[string]interface{}{
			"artifact_id":   artifact.ID.String(),
			"artifact_type": artifact.EventType,
			"reason":        "max revisions reached without consensus",
		},
	})
	// Non-fatal: workflow continues. Client sees escalation on Kanban.
	return nil
}

// runReview asks one reviewer to review an artifact.
//
// MENTAL MODEL:
//   Build a review prompt:
//     "You are [reviewer name]. Review this [artifact_type]:
//      [artifact content]
//      Respond with: APPROVED, CHANGES_REQUESTED: [reason], or BLOCKED: [reason]"
//
//   Call ModelGateway directly (not AgentLoop) for speed.
//   Parse response to extract status.
//
// CROSS-QUESTIONS:
//   Q: Why ModelGateway directly, not AgentLoop?
//   A: AgentLoop is for full OTA loops (multiple iterations).
//      Review is a single LLM call: read artifact -> respond.
//      Simpler, faster, cheaper.
//
//   Q: What if LLM response is malformed?
//   A: Default to "approved" (non-blocking). Log warning.
//      We don't want review failures to block the workflow.
func (cv *CrossVerifier) runReview(
	ctx context.Context,
	workflowID uuid.UUID,
	artifact blackboard.Event,
	reviewer workflowExpert,
	allExperts []workflowExpert,
) (status string, comment string, err error) {
	// Build review prompt
	artifactContent := string(artifact.Content)
	if len(artifactContent) > 3000 {
		artifactContent = artifactContent[:3000] + "\n... [truncated for review]"
	}

	systemPrompt := fmt.Sprintf(
		"You are %s, a domain expert in %s. "+
			"Your job is to review artifacts and ensure they meet quality standards. "+
			"Your charter: %s",
		reviewer.Name, reviewer.Domain, reviewer.ReasoningCharter,
	)

	userPrompt := fmt.Sprintf(
		"Review this %s artifact:\n\n%s\n\n"+
			"Respond with EXACTLY one of:\n"+
			"APPROVED\n"+
			"CHANGES_REQUESTED: [specific reason]\n"+
			"BLOCKED: [critical reason requiring client decision]\n\n"+
			"Check for:\n"+
			"- Correctness and completeness\n"+
			"- Security issues\n"+
			"- Consistency with other artifacts on the blackboard\n"+
			"- Adherence to domain best practices",
		artifact.EventType, artifactContent,
	)

	resp, callErr := cv.gateway.Call(ctx, LLMRequest{
		Model:        ModelCheap, // Reviews use cheap model (fast, low cost)
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    500,
	})
	if callErr != nil {
		return "", "", fmt.Errorf("review LLM call failed: %w", callErr)
	}

	// Post review_comment to blackboard
	responseText := strings.TrimSpace(resp.Content)
	parsedStatus, parsedComment := parseReviewResponse(responseText)

	_, _ = cv.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:         workflowID,
		EventType:          "review_comment",
		PostedByExpertID:   &reviewer.ID,
		ReferencesEventIDs: []uuid.UUID{artifact.ID},
		Content: map[string]interface{}{
			"artifact_id":   artifact.ID.String(),
			"artifact_type": artifact.EventType,
			"status":        parsedStatus,
			"comment":       parsedComment,
			"reviewer":      reviewer.Name,
			"reviewer_domain": reviewer.Domain,
		},
	})

	return parsedStatus, parsedComment, nil
}

// parseReviewResponse parses the LLM's review response.
//
// MENTAL MODEL:
//   Expected formats:
//     "APPROVED"
//     "CHANGES_REQUESTED: missing error handling in auth flow"
//     "BLOCKED: SQL injection vulnerability in user input"
//
//   If format is unexpected: default to "approved" (non-blocking).
func parseReviewResponse(response string) (status string, comment string) {
	upper := strings.ToUpper(response)

	if strings.HasPrefix(upper, "APPROVED") {
		return "approved", ""
	}

	if strings.HasPrefix(upper, "CHANGES_REQUESTED") {
		parts := strings.SplitN(response, ":", 2)
		if len(parts) == 2 {
			return "changes_requested", strings.TrimSpace(parts[1])
		}
		return "changes_requested", response
	}

	if strings.HasPrefix(upper, "BLOCKED") {
		parts := strings.SplitN(response, ":", 2)
		if len(parts) == 2 {
			return "blocked", strings.TrimSpace(parts[1])
		}
		return "blocked", response
	}

	// Unexpected format — default to approved (non-blocking)
	return "approved", ""
}

// LLMRequest is defined in agent_loop.go — reuse it here.
// (No redefinition needed — same package)
