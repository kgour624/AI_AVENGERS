package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Attempt statuses stored in workflow_task_attempts.status.
const (
	AttemptRunning   = "running"
	AttemptSucceeded = "succeeded"
	AttemptFailed    = "failed"
)

// maxTaskAttempts bounds how many times one (workflow, phase, expert) may run.
//
// WHY a cap exists: without it, a task that fails deterministically — a model
// that cannot answer for this domain, a document check that never passes —
// re-runs on every resume and on every restart, burning credits with no visible
// end. Three covers a transient provider failure and keeps the cost of a broken
// task bounded; after that the task fails with its reason, which the workflow
// shows instead of hiding.
const maxTaskAttempts = 3

// TaskAttempt is one unit of work in a workflow: what one expert was asked to do
// in one phase.
//
// Identity is (WorkflowID, Phase, ExpertID) — the table's UNIQUE key — which is
// what makes re-running safe: a retry updates this row rather than creating a
// second one, so there is no way for a resume to produce the same section twice.
type TaskAttempt struct {
	ID              uuid.UUID
	WorkflowID      uuid.UUID
	Phase           string
	ExpertID        uuid.UUID
	Attempt         int
	Status          string
	ArtifactEventID *uuid.UUID
	Error           string
	StartedAt       time.Time
	FinishedAt      *time.Time
}

// AttemptDecision is what the runner must do with a unit of work.
type AttemptDecision string

const (
	// DecisionRun: nothing has been attempted for this unit yet.
	DecisionRun AttemptDecision = "run"
	// DecisionRetry: a previous attempt failed, or a 'running' row was left
	// behind by a process that died mid-attempt. Re-running is safe because the
	// row — not the work — is what gets replaced.
	DecisionRetry AttemptDecision = "retry"
	// DecisionSkipDone: THIS exact (workflow, phase, expert) already succeeded.
	// This is the only reason to skip work.
	DecisionSkipDone AttemptDecision = "skip_done"
	// DecisionAbandoned: the attempt budget is spent. The runner must record a
	// failure for this expert rather than skip it, because a silent skip is
	// indistinguishable from work that never happened.
	DecisionAbandoned AttemptDecision = "abandoned"
)

// decideAttempt is the entire skip/retry policy as a pure function, so the rule
// that decides whether an expert runs can be unit-tested instead of inferred
// from the runner's control flow.
func decideAttempt(prev *TaskAttempt) AttemptDecision {
	if prev == nil {
		return DecisionRun
	}
	switch prev.Status {
	case AttemptSucceeded:
		// Succeeded for this phase, whatever the checkpoint says.
		return DecisionSkipDone
	case AttemptFailed, AttemptRunning:
		if prev.Attempt >= maxTaskAttempts {
			return DecisionAbandoned
		}
		return DecisionRetry
	default:
		// An unknown status is not evidence of anything. Run rather than skip:
		// skipping on a value nobody wrote would be the old bug in a new place.
		return DecisionRun
	}
}

// TaskAttemptStore owns the attempt ledger.
type TaskAttemptStore struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewTaskAttemptStore builds the store. A nil pool is allowed and makes every
// operation a no-op that reports "no ledger", so the runner keeps working in
// tests and in any wiring that has not provided a database.
func NewTaskAttemptStore(db *pgxpool.Pool, logger *zap.Logger) *TaskAttemptStore {
	return &TaskAttemptStore{db: db, logger: logger}
}

// ErrAttemptsExhausted is returned when a unit has used its whole attempt budget.
var ErrAttemptsExhausted = errors.New("task attempts exhausted")

// Load reads the current row for a unit, or (nil, nil) when it has never run.
func (s *TaskAttemptStore) Load(ctx context.Context, workflowID uuid.UUID, phase string, expertID uuid.UUID) (*TaskAttempt, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	row := s.db.QueryRow(ctx,
		`SELECT id, workflow_id, phase, expert_id, attempt, status,
		        artifact_event_id, error, started_at, finished_at
		   FROM workflow_task_attempts
		  WHERE workflow_id = $1 AND phase = $2 AND expert_id = $3`,
		workflowID, phase, expertID,
	)
	var a TaskAttempt
	if err := row.Scan(&a.ID, &a.WorkflowID, &a.Phase, &a.ExpertID, &a.Attempt,
		&a.Status, &a.ArtifactEventID, &a.Error, &a.StartedAt, &a.FinishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("load task attempt: %w", err)
	}
	return &a, nil
}

// Claim decides whether a unit may run and, when it may, takes it: the row is
// created or bumped to 'running' with the attempt counter incremented.
//
// Returns the row to complete later, the decision, and an error. When the
// decision is DecisionAbandoned the caller must record a failure for the expert
// — an exhausted unit is not a finished one.
func (s *TaskAttemptStore) Claim(ctx context.Context, workflowID uuid.UUID, phase string, expertID uuid.UUID) (*TaskAttempt, AttemptDecision, error) {
	if s == nil || s.db == nil {
		// No ledger wired: fall back to running the work. Doing the work twice
		// is recoverable; skipping work that never happened is not.
		return nil, DecisionRun, nil
	}

	prev, err := s.Load(ctx, workflowID, phase, expertID)
	if err != nil {
		return nil, DecisionRun, err
	}
	decision := decideAttempt(prev)
	if decision == DecisionSkipDone || decision == DecisionAbandoned {
		return prev, decision, nil
	}

	row := s.db.QueryRow(ctx,
		`INSERT INTO workflow_task_attempts
		     (workflow_id, phase, expert_id, attempt, status, started_at, updated_at)
		 VALUES ($1, $2, $3, 1, $4, NOW(), NOW())
		 ON CONFLICT (workflow_id, phase, expert_id) DO UPDATE SET
		     attempt     = workflow_task_attempts.attempt + 1,
		     status      = $4,
		     error       = '',
		     started_at  = NOW(),
		     finished_at = NULL,
		     updated_at  = NOW()
		 RETURNING id, workflow_id, phase, expert_id, attempt, status, started_at`,
		workflowID, phase, expertID, AttemptRunning,
	)
	var a TaskAttempt
	if err := row.Scan(&a.ID, &a.WorkflowID, &a.Phase, &a.ExpertID, &a.Attempt, &a.Status, &a.StartedAt); err != nil {
		return nil, DecisionRun, fmt.Errorf("claim task attempt: %w", err)
	}
	return &a, decision, nil
}

// Succeed marks a claimed unit as finished, recording the artifact it produced
// when the runner knows its event id.
func (s *TaskAttemptStore) Succeed(ctx context.Context, attempt *TaskAttempt, artifactEventID *uuid.UUID) error {
	if s == nil || s.db == nil || attempt == nil {
		return nil
	}
	if _, err := s.db.Exec(ctx,
		`UPDATE workflow_task_attempts
		    SET status = $2, artifact_event_id = $3, error = '',
		        finished_at = NOW(), updated_at = NOW()
		  WHERE id = $1`,
		attempt.ID, AttemptSucceeded, artifactEventID,
	); err != nil {
		return fmt.Errorf("complete task attempt: %w", err)
	}
	return nil
}

// Fail marks a claimed unit as failed and keeps the reason, so a resume can
// retry it and a human can see what went wrong.
func (s *TaskAttemptStore) Fail(ctx context.Context, attempt *TaskAttempt, reason string) error {
	if s == nil || s.db == nil || attempt == nil {
		return nil
	}
	// Bound the stored text: a provider error can be a wall of JSON and this
	// column exists to explain, not to archive.
	if len(reason) > 1000 {
		reason = reason[:1000]
	}
	if _, err := s.db.Exec(ctx,
		`UPDATE workflow_task_attempts
		    SET status = $2, error = $3, finished_at = NOW(), updated_at = NOW()
		  WHERE id = $1`,
		attempt.ID, AttemptFailed, reason,
	); err != nil {
		return fmt.Errorf("fail task attempt: %w", err)
	}
	return nil
}

// Summarize reports how many units of a workflow succeeded, failed, or are still
// running. Used by the end-to-end check so "the workflow says completed" can be
// compared against "the ledger says something actually ran".
func (s *TaskAttemptStore) Summarize(ctx context.Context, workflowID uuid.UUID) (map[string]int, error) {
	out := map[string]int{}
	if s == nil || s.db == nil {
		return out, nil
	}
	rows, err := s.db.Query(ctx,
		`SELECT status, COUNT(*) FROM workflow_task_attempts
		  WHERE workflow_id = $1 GROUP BY status`,
		workflowID,
	)
	if err != nil {
		return nil, fmt.Errorf("summarize task attempts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, fmt.Errorf("scan task attempt summary: %w", err)
		}
		out[strings.TrimSpace(status)] = n
	}
	return out, rows.Err()
}

// containsID reports whether a list of UUID strings already holds id. The
// checkpoint list is append-only in practice; this keeps a re-adopted expert
// from appearing twice and inflating the handoff summary.
func containsID(ids []string, id string) bool {
	for _, existing := range ids {
		if existing == id {
			return true
		}
	}
	return false
}

// attemptNumber is the attempt count for logging, tolerating a nil row (the
// ledger can be unwired, in which case there is no number to report).
func attemptNumber(a *TaskAttempt) int {
	if a == nil {
		return 0
	}
	return a.Attempt
}
