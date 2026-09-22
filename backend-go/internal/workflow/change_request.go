package workflow

// ChangeRequestService handles client-initiated change requests raised from the
// workflow chat.
//
// WHAT THIS DOES
//
// When a client types a change suggestion in the workflow chat (e.g. "add bulk
// move endpoint"), ProposeChange:
//   1. Stores the request in change_requests table (audit trail).
//   2. Posts a "change_request" event on the blackboard so the runner can
//      detect it.
//   3. Returns immediately — the runner picks it up asynchronously.
//
// The runner's watchForChangeRequest goroutine polls for pending change_requests
// while the workflow is running. When it finds one it:
//   1. Marks it running.
//   2. Determines which experts are relevant (LLM call — cheap model).
//   3. Re-runs the design phases for those experts (same loop as the initial
//      design, reusing RestartPhase + executeWaves).
//   4. Presents the same approval gate the initial design uses.
//
// WHY NOT REUSE propose_amendment
//
// propose_amendment is a single-file, single-expert, single-change operation.
// A change request is a goal that may touch multiple experts and multiple
// sections — it is a mini redesign, not a text patch.
//
// WHY THE RUNNER POLLS INSTEAD OF BEING CALLED DIRECTLY
//
// The runner is a long-running goroutine that owns the workflow state machine.
// Calling it directly from the HTTP handler would require a channel or a mutex
// shared between the handler and the runner — coupling two layers that are
// deliberately separate. Polling the DB is the same pattern waitForResume uses
// and keeps the runner as the single writer of workflow status.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
	"ai_avengers/backend/internal/gateway"
)

// ErrChangeRequestNotFound is returned when a change_request row does not exist
// or does not belong to the requesting client.
var ErrChangeRequestNotFound = errors.New("change request not found")

// ChangeRequest is one row from the change_requests table.
type ChangeRequest struct {
	ID                uuid.UUID   `json:"id"`
	WorkflowID        uuid.UUID   `json:"workflow_id"`
	ClientID          uuid.UUID   `json:"client_id"`
	SourceChatID      *uuid.UUID  `json:"source_chat_id,omitempty"`
	SourceMessageID   *uuid.UUID  `json:"source_message_id,omitempty"`
	ChangeGoal        string      `json:"change_goal"`
	RelevantExpertIDs []uuid.UUID `json:"relevant_expert_ids"`
	Status            string      `json:"status"`
	BlackboardEventID *uuid.UUID  `json:"blackboard_event_id,omitempty"`
	RequestedAt       time.Time   `json:"requested_at"`
	StartedAt         *time.Time  `json:"started_at,omitempty"`
	CompletedAt       *time.Time  `json:"completed_at,omitempty"`
}

// ChangeRequestService creates and reads change requests.
// The runner (WorkflowRunner) is the only writer of status transitions.
type ChangeRequestService struct {
	db     *pgxpool.Pool
	store  *blackboard.Store
	logger *zap.Logger
}

// NewChangeRequestService creates the service.
func NewChangeRequestService(
	db *pgxpool.Pool,
	store *blackboard.Store,
	logger *zap.Logger,
) *ChangeRequestService {
	return &ChangeRequestService{
		db:     db,
		store:  store,
		logger: logger,
	}
}

// ProposeChange stores a change request and posts it on the blackboard.
//
// sourceChatID and sourceMessageID are optional — they are the chat and message
// that originated the request, kept for traceability. They may be nil when the
// request is raised outside a chat context.
//
// Mental execution:
//   workflowID = abc, clientID = xyz, goal = "add bulk move endpoint"
//   1. Verify workflow exists and belongs to client
//   2. INSERT change_requests row (status=pending)
//   3. store.Post(change_request event) -> blackboard_event_id
//   4. UPDATE change_requests SET blackboard_event_id = ...
//   5. Return ChangeRequest
func (s *ChangeRequestService) ProposeChange(
	ctx context.Context,
	workflowID, clientID uuid.UUID,
	changeGoal string,
	sourceChatID, sourceMessageID *uuid.UUID,
) (*ChangeRequest, error) {
	if changeGoal == "" {
		return nil, fmt.Errorf("propose change: change_goal is required")
	}

	// Verify workflow exists and belongs to this client.
	// Same ownership check pattern as WorkflowChatService.CreateChat.
	var ownerID uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT client_id FROM workflows WHERE id = $1`, workflowID,
	).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("propose change: %w", ErrChangeRequestNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("propose change: load workflow: %w", err)
	}
	if ownerID != clientID {
		// Same deliberate ambiguity as ErrChatNotFound: don't reveal that the
		// workflow exists but belongs to someone else.
		return nil, fmt.Errorf("propose change: %w", ErrChangeRequestNotFound)
	}

	// Insert the change request row.
	var cr ChangeRequest
	var relevantRaw []byte
	err = s.db.QueryRow(ctx,
		`INSERT INTO change_requests
		     (workflow_id, client_id, source_chat_id, source_message_id, change_goal)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, workflow_id, client_id, source_chat_id, source_message_id,
		           change_goal, relevant_expert_ids, status,
		           blackboard_event_id, requested_at, started_at, completed_at`,
		workflowID, clientID, sourceChatID, sourceMessageID, changeGoal,
	).Scan(
		&cr.ID, &cr.WorkflowID, &cr.ClientID,
		&cr.SourceChatID, &cr.SourceMessageID,
		&cr.ChangeGoal, &relevantRaw, &cr.Status,
		&cr.BlackboardEventID, &cr.RequestedAt, &cr.StartedAt, &cr.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("propose change: insert: %w", err)
	}
	if len(relevantRaw) > 0 {
		_ = json.Unmarshal(relevantRaw, &cr.RelevantExpertIDs)
	}

	// Post the blackboard event so the runner can detect it.
	// The event content carries the change_request_id so the runner can load
	// the full row without a separate query.
	event, err := s.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:     workflowID,
		EventType:      "change_request",
		PostedByClient: true,
		Content: map[string]interface{}{
			"change_request_id": cr.ID.String(),
			"change_goal":       changeGoal,
		},
	})
	if err != nil {
		// Non-fatal: the row exists, the runner can also poll the table.
		// Log and continue rather than rolling back the insert.
		s.logger.Warn("propose change: blackboard post failed (non-fatal)",
			zap.String("change_request_id", cr.ID.String()),
			zap.Error(err),
		)
	} else {
		// Update the row with the blackboard event id for traceability.
		if _, err := s.db.Exec(ctx,
			`UPDATE change_requests SET blackboard_event_id = $1 WHERE id = $2`,
			event.ID, cr.ID,
		); err != nil {
			// Non-fatal: traceability field only.
			s.logger.Warn("propose change: update blackboard_event_id failed (non-fatal)",
				zap.String("change_request_id", cr.ID.String()),
				zap.Error(err),
			)
		} else {
			cr.BlackboardEventID = &event.ID
		}
	}

	s.logger.Info("change request proposed",
		zap.String("change_request_id", cr.ID.String()),
		zap.String("workflow_id", workflowID.String()),
		zap.String("goal", changeGoal),
	)
	return &cr, nil
}

// GetPendingChangeRequest returns the oldest pending change request for a
// workflow, or nil if none exists.
//
// Called by the runner's watchForChangeRequest loop. Oldest-first so requests
// are processed in the order they were submitted.
func (s *ChangeRequestService) GetPendingChangeRequest(
	ctx context.Context,
	workflowID uuid.UUID,
) (*ChangeRequest, error) {
	var cr ChangeRequest
	var relevantRaw []byte
	err := s.db.QueryRow(ctx,
		`SELECT id, workflow_id, client_id, source_chat_id, source_message_id,
		        change_goal, relevant_expert_ids, status,
		        blackboard_event_id, requested_at, started_at, completed_at
		 FROM change_requests
		 WHERE workflow_id = $1 AND status = 'pending'
		 ORDER BY requested_at ASC
		 LIMIT 1`,
		workflowID,
	).Scan(
		&cr.ID, &cr.WorkflowID, &cr.ClientID,
		&cr.SourceChatID, &cr.SourceMessageID,
		&cr.ChangeGoal, &relevantRaw, &cr.Status,
		&cr.BlackboardEventID, &cr.RequestedAt, &cr.StartedAt, &cr.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // no pending request
	}
	if err != nil {
		return nil, fmt.Errorf("get pending change request: %w", err)
	}
	if len(relevantRaw) > 0 {
		_ = json.Unmarshal(relevantRaw, &cr.RelevantExpertIDs)
	}
	return &cr, nil
}

// MarkRunning transitions a change request from pending to running.
// Sets started_at and stores the relevant expert ids determined by the runner.
func (s *ChangeRequestService) MarkRunning(
	ctx context.Context,
	crID uuid.UUID,
	relevantExpertIDs []uuid.UUID,
) error {
	relevantJSON, err := json.Marshal(relevantExpertIDs)
	if err != nil {
		return fmt.Errorf("mark running: marshal expert ids: %w", err)
	}
	now := time.Now()
	tag, err := s.db.Exec(ctx,
		`UPDATE change_requests
		 SET status = 'running',
		     relevant_expert_ids = $1,
		     started_at = $2
		 WHERE id = $3 AND status = 'pending'`,
		string(relevantJSON), now, crID,
	)
	if err != nil {
		return fmt.Errorf("mark running: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("mark running: change request %s is not pending", crID)
	}
	return nil
}

// MarkCompleted transitions a change request from running to completed.
func (s *ChangeRequestService) MarkCompleted(ctx context.Context, crID uuid.UUID) error {
	now := time.Now()
	_, err := s.db.Exec(ctx,
		`UPDATE change_requests
		 SET status = 'completed', completed_at = $1
		 WHERE id = $2 AND status = 'running'`,
		now, crID,
	)
	if err != nil {
		return fmt.Errorf("mark completed: %w", err)
	}
	return nil
}

// MarkCancelled cancels a pending change request.
func (s *ChangeRequestService) MarkCancelled(ctx context.Context, crID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE change_requests
		 SET status = 'cancelled'
		 WHERE id = $1 AND status = 'pending'`,
		crID,
	)
	if err != nil {
		return fmt.Errorf("mark cancelled: %w", err)
	}
	return nil
}

// ============================================================
// Relevance determination
// ============================================================

// DetermineRelevantExperts asks the LLM (cheap model) which experts need to act
// on this change goal, given the full expert roster.
//
// Returns the subset of expertIDs that are relevant. If the LLM call fails or
// returns an empty set, ALL experts are returned — the safe default is
// "everyone re-runs" rather than "nobody re-runs".
//
// Mental execution:
//   goal = "add bulk move endpoint"
//   experts = [PM, SystemDesign, Backend, Frontend, DB, DevOps, Security, QA]
//   LLM response: ["Backend", "SystemDesign", "QA"]
//   -> return IDs of Backend, SystemDesign, QA
//   If LLM fails -> return all 8 IDs
func DetermineRelevantExperts(
	ctx context.Context,
	gw *gateway.ModelGateway,
	workflowID uuid.UUID,
	changeGoal string,
	allExperts []workflowExpert,
	logger *zap.Logger,
) []uuid.UUID {
	if len(allExperts) == 0 {
		return nil
	}

	// Build expert list for the prompt.
	var expertLines strings.Builder
	for _, e := range allExperts {
		fmt.Fprintf(&expertLines, "- %s (domain: %s)\n", e.Name, e.Domain)
	}

	systemPrompt := `You are a software project coordinator.
Given a change goal and a list of domain experts, decide which experts need to
update their design section to implement this change.

Rules:
- Return ONLY the expert names that are directly affected by this change.
- An expert is affected if the change touches their domain.
- An expert is NOT affected if the change is entirely outside their domain.
- Return a JSON array of expert names, e.g. ["Backend Expert", "DB Expert"].
- If you are unsure, include the expert.
- Return ONLY the JSON array, no explanation.`

	userPrompt := fmt.Sprintf(
		"Change goal: %s\n\nExperts:\n%s\n\nWhich experts need to update their section?",
		changeGoal, expertLines.String(),
	)

	resp, err := gw.Call(ctx, gateway.LLMRequest{
		Model:        gateway.ModelCheap,
		WorkflowID:   &workflowID,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    200,
	})
	if err != nil {
		logger.Warn("determine relevant experts: LLM call failed, using all experts",
			zap.String("workflow_id", workflowID.String()),
			zap.Error(err),
		)
		return allExpertIDs(allExperts)
	}

	// Parse the JSON array of names.
	var names []string
	if err := json.Unmarshal([]byte(extractJSONArray(resp.Content)), &names); err != nil {
		logger.Warn("determine relevant experts: parse failed, using all experts",
			zap.String("workflow_id", workflowID.String()),
			zap.String("raw", resp.Content),
			zap.Error(err),
		)
		return allExpertIDs(allExperts)
	}

	if len(names) == 0 {
		logger.Warn("determine relevant experts: LLM returned empty set, using all experts",
			zap.String("workflow_id", workflowID.String()),
		)
		return allExpertIDs(allExperts)
	}

	// Match names to expert IDs (case-insensitive).
	nameSet := make(map[string]struct{}, len(names))
	for _, n := range names {
		nameSet[strings.ToLower(strings.TrimSpace(n))] = struct{}{}
	}

	var relevant []uuid.UUID
	for _, e := range allExperts {
		if _, ok := nameSet[strings.ToLower(e.Name)]; ok {
			relevant = append(relevant, e.ID)
		}
	}

	if len(relevant) == 0 {
		// Names didn't match — safe fallback.
		logger.Warn("determine relevant experts: no name matched, using all experts",
			zap.String("workflow_id", workflowID.String()),
			zap.Strings("llm_names", names),
		)
		return allExpertIDs(allExperts)
	}

	logger.Info("determine relevant experts",
		zap.String("workflow_id", workflowID.String()),
		zap.String("goal", changeGoal),
		zap.Int("relevant", len(relevant)),
		zap.Int("total", len(allExperts)),
	)
	return relevant
}

// allExpertIDs returns all expert IDs from a slice.
func allExpertIDs(experts []workflowExpert) []uuid.UUID {
	ids := make([]uuid.UUID, len(experts))
	for i, e := range experts {
		ids[i] = e.ID
	}
	return ids
}

// extractJSONArray finds the first '[' ... ']' substring in s.
// The LLM sometimes wraps the array in prose; this strips the prose.
func extractJSONArray(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start == -1 || end == -1 || end <= start {
		return "[]"
	}
	return s[start : end+1]
}
