package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
)

// ToolName constants — must match allowed_tools JSONB values in experts table.
const (
	ToolPostArtifact   = "PostArtifact"
	ToolAskExpert      = "AskExpert"
	ToolReadBlackboard = "ReadBlackboard"
	ToolAskClient      = "AskClient"
)

// askExpertTimeout is how long AskExpert waits for a response.
// Configurable via system_settings workflow_engine.default_review_timeout_seconds.
// Default 60s matches the design doc §8.4.
const askExpertTimeout = 60 * time.Second

// Tools provides the 4 blackboard tool implementations.
// Injected into the expert execution loop so experts can call them.
//
// WHY a struct not package-level functions:
//   Tools need access to the blackboard store and workflow engine.
//   Struct injection makes dependencies explicit and testable.
type Tools struct {
	store  *blackboard.Store
	engine *Engine
	sub    *blackboard.Subscriber
	logger *zap.Logger
}

// NewTools creates a new Tools instance.
func NewTools(
	store *blackboard.Store,
	engine *Engine,
	sub *blackboard.Subscriber,
	logger *zap.Logger,
) *Tools {
	return &Tools{
		store:  store,
		engine: engine,
		sub:    sub,
		logger: logger,
	}
}

// PostArtifactRequest is the input to PostArtifact.
type PostArtifactRequest struct {
	WorkflowID         uuid.UUID
	PostedByExpertID   uuid.UUID
	EventType          string      // e.g. "architecture_decision", "code_artifact_produced"
	Content            interface{} // will be JSON-marshalled
	ReferencesEventIDs []uuid.UUID // events this artifact responds to or builds on
	Revision           int         // 1 for new, >1 for revisions
}

// PostArtifact publishes an artifact to the blackboard.
// This is the primary output mechanism for all experts.
//
// Mental execution:
//   Expert: System Design, EventType: architecture_decision
//   Content: {"decision": "Use PostgreSQL", "rationale": "..."}
//   → blackboard.Store.Post() → INSERT blackboard_events
//   → Redis PUBLISH workflow:{id}:events
//   → Return event with sequence_number
func (t *Tools) PostArtifact(ctx context.Context, req PostArtifactRequest) (*blackboard.Event, error) {
	if req.EventType == "" {
		return nil, fmt.Errorf("PostArtifact: event_type is required")
	}

	expertID := req.PostedByExpertID
	event, err := t.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:         req.WorkflowID,
		EventType:          req.EventType,
		PostedByExpertID:   &expertID,
		PostedByClient:     false,
		Content:            req.Content,
		ReferencesEventIDs: req.ReferencesEventIDs,
		Revision:           req.Revision,
	})
	if err != nil {
		return nil, fmt.Errorf("PostArtifact: %w", err)
	}

	t.logger.Info("artifact posted",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("event_type", req.EventType),
		zap.String("expert_id", req.PostedByExpertID.String()),
		zap.Int64("sequence_number", event.SequenceNumber),
	)
	return event, nil
}

// AskExpertRequest is the input to AskExpert.
type AskExpertRequest struct {
	WorkflowID       uuid.UUID
	FromExpertID     uuid.UUID
	ToExpertID       uuid.UUID
	Question         string
	ReferencesEvents []uuid.UUID
}

// AskExpertResult is the response from AskExpert.
type AskExpertResult struct {
	AnswerEvent *blackboard.Event // the review_comment or answer event
	TimedOut    bool              // true if no answer within askExpertTimeout
}

// AskExpert posts a question_to_expert event and waits for a response.
// Returns when the target expert posts a review_comment event that
// references the question event, or when the timeout is reached.
//
// WHY blocking with timeout:
//   The asking expert cannot proceed without the answer.
//   Timeout prevents infinite blocking (§8.4).
//   On timeout: caller decides whether to proceed or escalate.
//
// Mental execution:
//   Frontend expert asks Backend: "What is the auth endpoint signature?"
//   1. Post question_to_expert event (to_expert_id = Backend)
//   2. Subscribe to blackboard events for this workflow
//   3. Wait for a review_comment event that references our question event
//   4. Return the answer event
//   5. If 60s pass with no answer: return TimedOut=true
func (t *Tools) AskExpert(ctx context.Context, req AskExpertRequest) (*AskExpertResult, error) {
	if req.Question == "" {
		return nil, fmt.Errorf("AskExpert: question is required")
	}

	// Step 1: Post the question
	toExpertID := req.ToExpertID
	fromExpertID := req.FromExpertID
	questionEvent, err := t.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:         req.WorkflowID,
		EventType:          "question_to_expert",
		PostedByExpertID:   &fromExpertID,
		ToExpertID:         &toExpertID,
		Content:            map[string]string{"question": req.Question},
		ReferencesEventIDs: req.ReferencesEvents,
	})
	if err != nil {
		return nil, fmt.Errorf("AskExpert: post question: %w", err)
	}

	t.logger.Info("expert question posted",
		zap.String("from", req.FromExpertID.String()),
		zap.String("to", req.ToExpertID.String()),
		zap.Int64("question_seq", questionEvent.SequenceNumber),
	)

	// Step 2: Subscribe and wait for answer
	// We subscribe from the question's sequence_number so we only see
	// events posted AFTER the question.
	timeoutCtx, cancel := context.WithTimeout(ctx, askExpertTimeout)
	defer cancel()

	eventCh, errCh := t.sub.Subscribe(
		timeoutCtx,
		req.WorkflowID,
		req.FromExpertID, // cursor tracked under the asking expert
		questionEvent.SequenceNumber,
	)

	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Channel closed — timeout or context cancelled
				return &AskExpertResult{TimedOut: true}, nil
			}
			// Check if this is a review_comment that references our question
			if event.EventType == "review_comment" && referencesEvent(event, questionEvent.ID) {
				return &AskExpertResult{AnswerEvent: &event}, nil
			}

		case err := <-errCh:
			if err != nil {
				return nil, fmt.Errorf("AskExpert: subscriber error: %w", err)
			}

		case <-timeoutCtx.Done():
			t.logger.Warn("AskExpert timed out",
				zap.String("from", req.FromExpertID.String()),
				zap.String("to", req.ToExpertID.String()),
			)
			return &AskExpertResult{TimedOut: true}, nil
		}
	}
}

// ReadBlackboardRequest is the input to ReadBlackboard.
type ReadBlackboardRequest struct {
	WorkflowID uuid.UUID
	EventTypes []string // filter by event type; empty = all types
	Since      int64    // sequence_number cursor; 0 = from beginning
}

// ReadBlackboard returns events from the blackboard matching the filter.
// Used by experts to consume artifacts posted by other experts.
//
// Mental execution:
//   Frontend expert reads: event_types=["api_contract_proposed"], since=0
//   → Returns all api_contract_proposed events for this workflow
//   → Frontend expert uses these to know what endpoints to implement
func (t *Tools) ReadBlackboard(ctx context.Context, req ReadBlackboardRequest) ([]blackboard.Event, error) {
	if len(req.EventTypes) == 0 {
		// No type filter: return all events since cursor
		return t.store.GetSince(ctx, req.WorkflowID, req.Since, 200)
	}
	return t.store.GetByType(ctx, req.WorkflowID, req.EventTypes, req.Since)
}

// AskClientRequest is the input to AskClient.
type AskClientRequest struct {
	WorkflowID       uuid.UUID
	FromExpertID     uuid.UUID
	GateName         string // must match approval_requests.gate_name CHECK constraint
	Summary          string
	ArtifactContent  interface{}
	CitedEventIDs    []uuid.UUID
	CostSoFarUSD     float64
	EstimatedRemUSD  float64
}

// AskClient posts a question_to_client event, creates an approval_request row,
// and pauses the workflow.
//
// WHY this pauses the workflow (Locked Decision L5):
//   No auto-approval on timeout. Workflow stays paused until client acts.
//   This is the client's explicit preference.
//
// The workflow resumes when the client calls
// POST /api/v1/workflows/{id}/approvals/{approval_id}.
//
// Mental execution:
//   System Design expert: "Architecture is ready for review"
//   1. Post question_to_client event on blackboard
//   2. INSERT approval_requests row
//   3. Engine.PauseForApproval()
//   4. Return approval_id so caller can track it
func (t *Tools) AskClient(ctx context.Context, req AskClientRequest) (uuid.UUID, error) {
	if req.Summary == "" {
		return uuid.Nil, fmt.Errorf("AskClient: summary is required")
	}
	if req.GateName == "" {
		req.GateName = "ad_hoc"
	}

	// Step 1: Post question_to_client event on blackboard
	fromExpertID := req.FromExpertID
	_, err := t.store.Post(ctx, blackboard.PostRequest{
		WorkflowID:       req.WorkflowID,
		EventType:        "question_to_client",
		PostedByExpertID: &fromExpertID,
		Content: map[string]interface{}{
			"gate_name": req.GateName,
			"summary":   req.Summary,
		},
		ReferencesEventIDs: req.CitedEventIDs,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("AskClient: post event: %w", err)
	}

	// Step 2: Create approval_request row
	approvalID, err := t.createApprovalRequest(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("AskClient: create approval: %w", err)
	}

	// Step 3: Pause the workflow
	if err := t.engine.PauseForApproval(ctx, req.WorkflowID); err != nil {
		return uuid.Nil, fmt.Errorf("AskClient: pause workflow: %w", err)
	}

	t.logger.Info("workflow paused — waiting for client",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("gate", req.GateName),
		zap.String("approval_id", approvalID.String()),
	)
	return approvalID, nil
}

// createApprovalRequest inserts a row into approval_requests.
func (t *Tools) createApprovalRequest(ctx context.Context, req AskClientRequest) (uuid.UUID, error) {
	var approvalID uuid.UUID

	// Marshal artifact content
	artifactJSON := []byte("{}")
	if req.ArtifactContent != nil {
		var err error
		artifactJSON, err = marshalJSON(req.ArtifactContent)
		if err != nil {
			return uuid.Nil, fmt.Errorf("marshal artifact: %w", err)
		}
	}

	// Build cited_event_ids as Postgres UUID[]
	citedJSON := []byte("{}")
	if len(req.CitedEventIDs) > 0 {
		var err error
		citedJSON, err = marshalJSON(req.CitedEventIDs)
		if err != nil {
			return uuid.Nil, fmt.Errorf("marshal cited events: %w", err)
		}
	}

	err := t.engine.db.QueryRow(ctx,
		`INSERT INTO approval_requests
			(workflow_id, gate_name, summary, artifact_content,
			 cited_event_ids, cost_so_far_usd, estimated_remaining_usd)
		 VALUES ($1, $2, $3, $4, $5::uuid[], $6, $7)
		 RETURNING id`,
		req.WorkflowID,
		req.GateName,
		req.Summary,
		string(artifactJSON),
		string(citedJSON),
		req.CostSoFarUSD,
		req.EstimatedRemUSD,
	).Scan(&approvalID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert approval_request: %w", err)
	}
	return approvalID, nil
}

// referencesEvent checks if an event's ReferencesEventIDs contains targetID.
func referencesEvent(event blackboard.Event, targetID uuid.UUID) bool {
	for _, ref := range event.ReferencesEventIDs {
		if ref == targetID {
			return true
		}
	}
	return false
}

// marshalJSON is a thin wrapper around json.Marshal used internally.
func marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
