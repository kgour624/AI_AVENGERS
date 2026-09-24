package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// ErrArtifactBlocked is returned when any reviewer marks an artifact as BLOCKED.
// The runner treats this as a hard stop (client must decide).
var ErrArtifactBlocked = errors.New("artifact blocked")

// CrossVerifier implements the cross-verification protocol from §8.
//
// MENTAL MODEL:
//   After each wave, scan blackboard for new artifacts.
//   For each artifact, find mandatory reviewers (from ReviewerMatrix in reviewers.go).
//   Each reviewer reads the artifact and posts review_comment.
//   If all approve -> artifact is final.
//   If any requests changes -> producer re-runs via AiderRunner with feedback (max MaxRevisionRounds).
//   If any blocks -> escalate to client.
//
// REVIEWER MATRIX: defined in reviewers.go (single source of truth).
//   code_artifact_produced  -> Code Reviewer, QA
//   architecture_decision   -> Security, LLD
//   data_model_proposed     -> System Design, Backend
//   api_contract_proposed   -> Frontend, Security
//   module_design_proposed  -> Code Reviewer
//   test_case_proposed      -> artifact owner
type CrossVerifier struct {
	store       *blackboard.Store
	gateway     *gateway.ModelGateway
	agentLoop   *AgentLoop
	aiderRunner *AiderRunner // used to re-run producer on changes_requested
	tools       *Tools
	logger      *zap.Logger
}

// NewCrossVerifier creates a new CrossVerifier.
// aiderRunner is required for the revision loop (Step 2).
// Pass nil only in tests where revision is not needed.
func NewCrossVerifier(
	store *blackboard.Store,
	gw *gateway.ModelGateway,
	agentLoop *AgentLoop,
	aiderRunner *AiderRunner,
	tools *Tools,
	logger *zap.Logger,
) *CrossVerifier {
	return &CrossVerifier{
		store:       store,
		gateway:     gw,
		agentLoop:   agentLoop,
		aiderRunner: aiderRunner,
		tools:       tools,
		logger:      logger,
	}
}

// VerifyWaveArtifacts runs cross-verification on all artifacts posted
// during a wave. Called after wg.Wait() in executeWaves().
//
// MENTAL MODEL:
//   Input: workflowID, lastSeqBefore (sequence number before wave started)
//   1. Fetch all new events since lastSeqBefore
//   2. Filter to artifact event types (those in ReviewerMatrix)
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

	// Build ID -> expert map for producer lookup in runProducerRevision
	idToExpert := make(map[uuid.UUID]workflowExpert)
	for _, e := range experts {
		idToExpert[e.ID] = e
	}

	// Process each artifact that needs review.
	// Use ReviewerMatrix from reviewers.go — single source of truth.
	// (Removed duplicate lowercase reviewerMatrix that was here before.)
	for _, artifact := range newEvents {
		rule := GetReviewerRule(artifact.EventType)
		if len(rule.MandatoryReviewers) == 0 {
			continue
		}

		// Find which mandatory reviewers are in this workflow
		var reviewers []workflowExpert
		for _, domain := range rule.MandatoryReviewers {
			// Skip placeholder domains (e.g. "_artifact_owner" resolved at runtime)
			if strings.HasPrefix(domain, "_") {
				continue
			}
			if reviewer, ok := domainToExpert[domain]; ok {
				reviewers = append(reviewers, reviewer)
			}
		}

		if len(reviewers) == 0 {
			// No reviewers in this workflow — auto-approve
			cv.logger.Warn("cross-verify: no reviewers in workflow, auto-approving",
				zap.String("event_type", artifact.EventType),
				zap.Strings("needed_domains", rule.MandatoryReviewers),
			)
			continue
		}

		// Run review loop for this artifact
		if err := cv.reviewArtifact(ctx, workflowID, artifact, reviewers, experts, idToExpert); err != nil {
			return fmt.Errorf("cross-verify: artifact %s: %w", artifact.ID, err)
		}
	}

	return nil
}

// reviewArtifact runs the review loop for one artifact.
//
// MENTAL MODEL:
//   Max MaxRevisionRounds revision rounds (reviewers.go §8.3).
//   Each round:
//     1. Each reviewer reads artifact + posts review_comment
//     2. If all approved -> done
//     3. If any changes_requested -> find producer, re-run AiderRunner with feedback
//     4. If any blocked -> escalate to client
//
// STEP 2 — PRODUCER RE-RUN:
//   When changes_requested:
//     - Find producer expert from artifact.PostedByExpertID
//     - Build revised TaskDescription = artifact summary + reviewer feedback
//     - Call aiderRunner.Run() with revised description
//     - Aider re-writes code with feedback in prompt
//     - Next round: reviewers check the new code
//
// CROSS-QUESTIONS:
//   Q: How does reviewer "read" the artifact?
//   A: We build a review prompt with artifact content.
//      Reviewer's ModelGateway call runs with this prompt.
//      Reviewer posts review_comment event.
//
//   Q: How does producer "revise"?
//   A: aiderRunner.Run() called with feedback injected into TaskDescription.
//      Aider sees: original task + "REVISION REQUIRED: [reviewer comments]"
//      Aider re-writes code addressing the feedback.
//
//   Q: What is "blocked"?
//   A: Reviewer posts review_comment with status=blocked.
//      This means a hard stop — client must decide.
//
//   Q: What if aiderRunner is nil?
//   A: Log warning, skip re-run, continue to next round.
//      Graceful degradation: review loop still runs, just no code fix.
func (cv *CrossVerifier) reviewArtifact(
	ctx context.Context,
	workflowID uuid.UUID,
	artifact blackboard.Event,
	reviewers []workflowExpert,
	allExperts []workflowExpert,
	idToExpert map[uuid.UUID]workflowExpert,
) error {
	for round := 1; round <= MaxRevisionRounds; round++ {
		cv.logger.Info("cross-verify: review round",
			zap.String("artifact_id", artifact.ID.String()),
			zap.String("event_type", artifact.EventType),
			zap.Int("round", round),
			zap.Int("reviewers", len(reviewers)),
		)

		allApproved := true
		anyBlocked := false
		reviewedCount := 0
		var failedReviewers []string
		var changeRequests []string

		for _, reviewer := range reviewers {
			status, comment, reviewErr := cv.runReview(ctx, workflowID, artifact, reviewer, allExperts)
			if reviewErr != nil {
				cv.logger.Warn("cross-verify: reviewer failed (skipping)",
					zap.String("reviewer", reviewer.Name),
					zap.Error(reviewErr),
				)
				failedReviewers = append(failedReviewers, reviewer.Name)
				continue // handled after loop so we never falsely approve
			}
			reviewedCount++

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
				PostedByClient: true,
				Content: map[string]interface{}{
					"artifact_id":   artifact.ID.String(),
					"artifact_type": artifact.EventType,
					"reason":        "reviewer blocked artifact",
				},
			})
			return fmt.Errorf("%w: %s", ErrArtifactBlocked, artifact.EventType)
		}

		// Fail closed: if any reviewer failed (or all failed), the verification result is
		// incomplete, so we must not mark the artifact as approved.
		if reviewedCount == 0 || len(failedReviewers) > 0 {
			_, _ = cv.store.Post(ctx, blackboard.PostRequest{
				WorkflowID:         workflowID,
				EventType:          "review_inconclusive",
				PostedByClient:     true,
				ReferencesEventIDs: []uuid.UUID{artifact.ID},
				Content: map[string]interface{}{
					"artifact_id":      artifact.ID.String(),
					"artifact_type":    artifact.EventType,
					"round":            round,
					"reviewed_count":   reviewedCount,
					"failed_reviewers": failedReviewers,
					"reason":           "one or more reviewers failed; verification incomplete",
				},
			})
			// Non-blocking by default at the runner-level; caller decides whether to stop.
			return fmt.Errorf("cross-verify incomplete for %s: reviewed=%d failed=%d", artifact.EventType, reviewedCount, len(failedReviewers))
		}

		if allApproved {
			// All reviewers approved — post final approval event
			_, _ = cv.store.Post(ctx, blackboard.PostRequest{
				WorkflowID: workflowID,
				EventType:  "artifact_approved",
				PostedByClient: true,
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

		// Changes requested — post revision_requested event to blackboard.
		_, _ = cv.store.Post(ctx, blackboard.PostRequest{
			WorkflowID:         workflowID,
			EventType:          "revision_requested",
			PostedByClient:     true,
			ReferencesEventIDs: []uuid.UUID{artifact.ID},
			Content: map[string]interface{}{
				"artifact_id":     artifact.ID.String(),
				"artifact_type":   artifact.EventType,
				"round":           round,
				"change_requests": changeRequests,
			},
		})

		// STEP 2: Re-run producer with reviewer feedback.
		//
		// MENTAL MODEL:
		//   artifact.PostedByExpertID = Backend expert UUID
		//   changeRequests = ["Code Reviewer: error handling missing", "QA: no nil test"]
		//   → build revised task description with feedback appended
		//   → aiderRunner.Run() → Aider sees feedback in prompt → fixes code
		//   → new code_artifact_produced event on blackboard
		//   → next round: reviewers check the new code
		//
		// WHY aiderRunner not agentLoop:
		//   code_artifact_produced artifacts come from AiderRunner (implementation phase).
		//   AgentLoop handles design artifacts (architecture_decision, etc.).
		//   Revision must use the same executor as the original production.
		//   For non-code artifacts (architecture_decision), agentLoop would be used.
		//   Current implementation: only code artifacts trigger aiderRunner revision.
		//   Design artifact revision via agentLoop is future work.
		if round < MaxRevisionRounds {
			cv.runProducerRevision(ctx, workflowID, artifact, changeRequests, idToExpert)
		}
	}

	// Max revisions reached — escalate to client
	cv.logger.Warn("cross-verify: max revisions reached, escalating to client",
		zap.String("artifact_type", artifact.EventType),
	)
	_, _ = cv.store.Post(ctx, blackboard.PostRequest{
		WorkflowID: workflowID,
		EventType:  "review_escalated_to_client",
		PostedByClient: true,
		Content: map[string]interface{}{
			"artifact_id":   artifact.ID.String(),
			"artifact_type": artifact.EventType,
			"reason":        "max revisions reached without consensus",
		},
	})
	// Non-fatal: workflow continues. Client sees escalation on Kanban.
	return nil
}

// runProducerRevision finds the producer expert and re-runs AiderRunner
// with reviewer feedback injected into the task description.
//
// MENTAL MODEL:
//   artifact.PostedByExpertID = *uuid.UUID pointing to Backend expert
//   idToExpert[*artifact.PostedByExpertID] = workflowExpert{Backend}
//   revisedDescription = artifact content summary +
//                        "\n\nREVISION REQUIRED:\n- Code Reviewer: ..."
//   aiderRunner.Run(ctx, AiderRunRequest{
//     WorkflowID:      workflowID,
//     Expert:          backendExpert,
//     TaskID:          uuid.New(),   // fresh ID for this revision run
//     TaskTitle:       "Revision: " + artifact.EventType,
//     TaskDescription: revisedDescription,
//     WorkflowPhase:   PhaseImplementation,
//   })
//
// CROSS-QUESTIONS:
//   Q: Why uuid.New() for TaskID?
//   A: Each revision is a distinct Aider run. Fresh ID prevents
//      checkpoint collision with the original run.
//
//   Q: What if artifact.PostedByExpertID is nil?
//   A: System-posted artifact (no expert). Skip revision. Log warning.
//
//   Q: What if producer not in idToExpert?
//   A: Expert was removed from workflow. Skip revision. Log warning.
//
//   Q: What if aiderRunner is nil?
//   A: Log warning, skip. Graceful degradation.
//
//   Q: What if aiderRunner.Run() fails?
//   A: Log error, non-fatal. Review loop continues to next round.
//      Reviewers will see the same old code and likely request changes again.
//      After MaxRevisionRounds, escalates to client.
//
//   Q: Why only code artifacts (PhaseImplementation)?
//   A: code_artifact_produced comes from AiderRunner.
//      Design artifacts (architecture_decision etc.) come from AgentLoop.
//      Revision executor must match original executor.
//      Design artifact revision via AgentLoop is future work.
func (cv *CrossVerifier) runProducerRevision(
	ctx context.Context,
	workflowID uuid.UUID,
	artifact blackboard.Event,
	changeRequests []string,
	idToExpert map[uuid.UUID]workflowExpert,
) {
	// Guard: aiderRunner must be available
	if cv.aiderRunner == nil {
		cv.logger.Warn("cross-verify: aiderRunner nil, skipping producer revision",
			zap.String("artifact_type", artifact.EventType),
		)
		return
	}

	// Guard: only re-run for code artifacts (AiderRunner produced them).
	// Design artifacts (architecture_decision, data_model_proposed, etc.)
	// are produced by AgentLoop — revision via AgentLoop is future work.
	if artifact.EventType != "code_artifact_produced" {
		cv.logger.Info("cross-verify: non-code artifact, skipping aider revision (future work)",
			zap.String("artifact_type", artifact.EventType),
		)
		return
	}

	// Guard: artifact must have a producer
	if artifact.PostedByExpertID == nil {
		cv.logger.Warn("cross-verify: artifact has no producer (system-posted), skipping revision",
			zap.String("artifact_id", artifact.ID.String()),
		)
		return
	}

	// Find producer expert in workflow
	producer, ok := idToExpert[*artifact.PostedByExpertID]
	if !ok {
		cv.logger.Warn("cross-verify: producer expert not in workflow, skipping revision",
			zap.String("producer_id", artifact.PostedByExpertID.String()),
		)
		return
	}

	// Build revised task description.
	// Structure:
	//   [Original artifact content summary]
	//
	//   REVISION REQUIRED — Reviewer Feedback:
	//   - Code Reviewer: error handling missing in auth flow
	//   - QA: no test for nil token case
	//
	//   Fix ALL issues listed above. Re-implement the affected code.
	artifactSummary := string(artifact.Content)
	if len(artifactSummary) > 1000 {
		artifactSummary = artifactSummary[:1000] + "\n... [truncated]"
	}

	var sb strings.Builder
	sb.WriteString("You previously produced this artifact:\n")
	sb.WriteString(artifactSummary)
	sb.WriteString("\n\nREVISION REQUIRED — Reviewer Feedback:\n")
	for _, req := range changeRequests {
		sb.WriteString("- ")
		sb.WriteString(req)
		sb.WriteString("\n")
	}
	sb.WriteString("\nFix ALL issues listed above. Re-implement the affected code.")
	sb.WriteString("\nPost the revised code_artifact_produced when done.")
	revisedDescription := sb.String()

	cv.logger.Info("cross-verify: re-running producer with reviewer feedback",
		zap.String("producer", producer.Name),
		zap.String("artifact_type", artifact.EventType),
		zap.Strings("change_requests", changeRequests),
	)

	// Re-run AiderRunner for the producer expert.
	// TaskID: fresh UUID so checkpoint doesn't collide with original run.
	// WorkflowPhase: PhaseImplementation (code artifacts always from impl phase).
	_, runErr := cv.aiderRunner.Run(ctx, AiderRunRequest{
		WorkflowID:      workflowID,
		Expert:          producer,
		TaskID:          uuid.New(),
		TaskTitle:       fmt.Sprintf("Revision: %s", artifact.EventType),
		TaskDescription: revisedDescription,
		WorkflowPhase:   PhaseImplementation,
	})
	if runErr != nil {
		// Non-fatal: log error, review loop continues.
		// Reviewers will see old code in next round and request changes again.
		// After MaxRevisionRounds, escalates to client.
		cv.logger.Error("cross-verify: producer revision run failed (non-fatal)",
			zap.String("producer", producer.Name),
			zap.Error(runErr),
		)
		return
	}

	cv.logger.Info("cross-verify: producer revision completed",
		zap.String("producer", producer.Name),
		zap.String("artifact_type", artifact.EventType),
	)
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

	resp, callErr := cv.gateway.Call(ctx, gateway.LLMRequest{
		Model: gateway.ModelCheap, // Reviews use cheap model (fast, low cost)
		// artifact carries its own workflow id — no extra parameter needed.
		WorkflowID:   &artifact.WorkflowID,
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
		PostedByClient:     false,
		ReferencesEventIDs: []uuid.UUID{artifact.ID},
		Content: map[string]interface{}{
			"artifact_id":     artifact.ID.String(),
			"artifact_type":   artifact.EventType,
			"status":          parsedStatus,
			"comment":         parsedComment,
			"reviewer":        reviewer.Name,
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

	// Unexpected format — fail closed.
	// Treat as changes_requested so we never falsely approve an unreviewed artifact.
	// This will either trigger a revision loop (code artifacts) or escalate after
	// MaxRevisionRounds (non-code artifacts).
	trimmed := strings.TrimSpace(response)
	if len(trimmed) > 300 {
		trimmed = trimmed[:300] + "..."
	}
	return "changes_requested", "unparseable review response: " + trimmed
}
