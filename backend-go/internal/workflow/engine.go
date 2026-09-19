package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Phase constants — must match CHECK constraint in migration 006.
const (
	PhaseIntake          = "intake"
	PhaseHighLevelDesign = "high_level_design"
	PhaseDetailedDesign  = "detailed_design"
	PhaseImplementation  = "implementation"
	PhaseQA              = "qa"
	PhaseHandoff         = "handoff"
	PhaseCompleted       = "completed"
)

// Status constants — must match CHECK constraint in migration 006.
const (
	StatusDraft               = "draft"
	StatusRunning             = "running"
	StatusPausedForApproval   = "paused_for_approval"
	StatusPausedForInput      = "paused_for_client_input"
	StatusCompleted           = "completed"
	StatusCancelled           = "cancelled"
	StatusFailed              = "failed"
)

// phaseOrder defines the valid forward progression.
// Used by TransitionPhase to validate the requested next phase.
var phaseOrder = []string{
	PhaseIntake,
	PhaseHighLevelDesign,
	PhaseDetailedDesign,
	PhaseImplementation,
	PhaseQA,
	PhaseHandoff,
	PhaseCompleted,
}

// Workflow is the in-memory representation of a workflows table row.
type Workflow struct {
	ID                 uuid.UUID       `json:"id"`
	ClientID           uuid.UUID       `json:"client_id"`
	ProjectID          uuid.UUID       `json:"project_id"`
	Title              string          `json:"title"`
	Status             string          `json:"status"`
	CurrentPhase       string          `json:"current_phase"`
	PhaseStartedAt     *time.Time      `json:"phase_started_at"`
	PhaseCompletedAt   *time.Time      `json:"phase_completed_at"`
	SelectedExpertIDs  []uuid.UUID     `json:"selected_expert_ids"`
	CostBudgetUSD      float64         `json:"cost_budget_usd"`
	CostSpentUSD       float64         `json:"cost_spent_usd"`
	CostSoftLimitPct   float64         `json:"cost_soft_limit_pct"`
	CostHardLimitPct   float64         `json:"cost_hard_limit_pct"`
	// GenericAllowancePct: client-set ceiling (0-30) on how much of an
	// expert's answer may be generic knowledge. 0 (default) means trained
	// knowledge + peer experts only. See migration 017.
	GenericAllowancePct float64        `json:"generic_allowance_pct"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// MaxGenericAllowancePct is the product ceiling on generic knowledge.
// Trained knowledge must always remain the majority contributor; the same
// bound is enforced in the DB by workflows_generic_allowance_pct_check.
const MaxGenericAllowancePct = 30.0

// CreateRequest is the input to Engine.Create.
type CreateRequest struct {
	ClientID          uuid.UUID
	ProjectID         uuid.UUID
	Title             string
	SelectedExpertIDs []uuid.UUID
	CostBudgetUSD     float64 // 0 = use default (10.0)
}

// CostLimitResult is returned by UpdateCostSpent.
type CostLimitResult struct {
	SoftLimitHit bool    // crossed soft_limit_pct threshold
	HardLimitHit bool    // crossed hard_limit_pct threshold — workflow must pause
	CostSpentUSD float64 // new total
}

// Engine owns all writes to the workflows table.
// Implements the 6-phase state machine.
//
// WHY engine owns all status writes (Locked Decision L11):
//   Experts post events; engine reads events and decides transitions.
//   No expert can directly mutate workflow status.
type Engine struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewEngine creates a new workflow engine.
func NewEngine(db *pgxpool.Pool, logger *zap.Logger) *Engine {
	return &Engine{db: db, logger: logger}
}

// Create creates a new workflow in 'draft' status.
// Does NOT start it — call Start() when ready.
func (e *Engine) Create(ctx context.Context, req CreateRequest) (*Workflow, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("workflow create: title is required")
	}
	if len(req.SelectedExpertIDs) == 0 {
		return nil, fmt.Errorf("workflow create: at least one expert must be selected")
	}

	budget := req.CostBudgetUSD
	if budget <= 0 {
		budget = 10.0 // default per migration 006
	}

	expertIDsJSON, err := json.Marshal(req.SelectedExpertIDs)
	if err != nil {
		return nil, fmt.Errorf("workflow create: marshal expert ids: %w", err)
	}

	var w Workflow
	var expertIDsRaw []byte
	err = e.db.QueryRow(ctx,
		`INSERT INTO workflows
			(client_id, project_id, title, status, current_phase,
			 selected_expert_ids, cost_budget_usd)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, client_id, project_id, title, status, current_phase,
		           phase_started_at, phase_completed_at,
		           selected_expert_ids, cost_budget_usd, cost_spent_usd,
		           cost_soft_limit_pct, cost_hard_limit_pct,
		           generic_allowance_pct,
		           created_at, updated_at`,
		req.ClientID, req.ProjectID, req.Title,
		StatusDraft, PhaseIntake,
		string(expertIDsJSON), budget,
	).Scan(
		&w.ID, &w.ClientID, &w.ProjectID, &w.Title, &w.Status, &w.CurrentPhase,
		&w.PhaseStartedAt, &w.PhaseCompletedAt,
		&expertIDsRaw, &w.CostBudgetUSD, &w.CostSpentUSD,
		&w.CostSoftLimitPct, &w.CostHardLimitPct,
		&w.GenericAllowancePct,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow create: insert: %w", err)
	}
	if err := json.Unmarshal(expertIDsRaw, &w.SelectedExpertIDs); err != nil {
		w.SelectedExpertIDs = req.SelectedExpertIDs // fallback
	}

	e.logger.Info("workflow created",
		zap.String("workflow_id", w.ID.String()),
		zap.String("title", w.Title),
	)
	return &w, nil
}

// Start transitions a draft workflow to running and enters the intake phase.
// Writes the first phase checkpoint.
func (e *Engine) Start(ctx context.Context, workflowID uuid.UUID) (*Workflow, error) {
	now := time.Now()
	var w Workflow
	var expertIDsRaw []byte
	err := e.db.QueryRow(ctx,
		`UPDATE workflows SET
			status = $1,
			current_phase = $2,
			phase_started_at = $3,
			updated_at = NOW()
		 WHERE id = $4 AND status = $5
		 RETURNING id, client_id, project_id, title, status, current_phase,
		           phase_started_at, phase_completed_at,
		           selected_expert_ids, cost_budget_usd, cost_spent_usd,
		           cost_soft_limit_pct, cost_hard_limit_pct,
		           generic_allowance_pct,
		           created_at, updated_at`,
		StatusRunning, PhaseIntake, now,
		workflowID, StatusDraft,
	).Scan(
		&w.ID, &w.ClientID, &w.ProjectID, &w.Title, &w.Status, &w.CurrentPhase,
		&w.PhaseStartedAt, &w.PhaseCompletedAt,
		&expertIDsRaw, &w.CostBudgetUSD, &w.CostSpentUSD,
		&w.CostSoftLimitPct, &w.CostHardLimitPct,
		&w.GenericAllowancePct,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow start: %w (must be in draft status)", err)
	}
	if err := json.Unmarshal(expertIDsRaw, &w.SelectedExpertIDs); err != nil {
		w.SelectedExpertIDs = []uuid.UUID{}
	}

	e.logger.Info("workflow started",
		zap.String("workflow_id", w.ID.String()),
		zap.String("phase", w.CurrentPhase),
	)
	return &w, nil
}

// TransitionPhase moves the workflow to the next phase.
// Validates that nextPhase is a valid forward transition.
// Writes a phase checkpoint before transitioning.
//
// Mental execution:
//   Current: high_level_design, next: detailed_design
//   1. Validate nextPhase is after current in phaseOrder
//   2. Write checkpoint for completed phase
//   3. UPDATE workflows SET current_phase=detailed_design, phase_started_at=NOW()
//   4. Return updated workflow
func (e *Engine) TransitionPhase(
	ctx context.Context,
	workflowID uuid.UUID,
	nextPhase string,
	checkpointSnapshot interface{},
	lastBlackboardSeq int64,
) (*Workflow, error) {
	// Load current workflow
	w, err := e.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow transition: load: %w", err)
	}

	// Validate forward transition
	if !isValidTransition(w.CurrentPhase, nextPhase) {
		return nil, fmt.Errorf("workflow transition: invalid: %s -> %s", w.CurrentPhase, nextPhase)
	}

	// Write phase checkpoint BEFORE transitioning
	// WHY before: if the UPDATE fails, we still have the checkpoint.
	// On retry, we can detect the checkpoint exists and skip re-writing it.
	if err := e.writeCheckpoint(ctx, workflowID, w.CurrentPhase, checkpointSnapshot, lastBlackboardSeq); err != nil {
		e.logger.Warn("workflow transition: checkpoint write failed (non-fatal)",
			zap.String("workflow_id", workflowID.String()),
			zap.Error(err),
		)
		// Non-fatal: checkpoint failure doesn't block the transition.
		// The blackboard is still the recovery source.
	}

	now := time.Now()
	var updated Workflow
	var expertIDsRaw []byte
	err = e.db.QueryRow(ctx,
		`UPDATE workflows SET
			current_phase = $1,
			phase_started_at = $2,
			phase_completed_at = NULL,
			status = $3,
			updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, client_id, project_id, title, status, current_phase,
		           phase_started_at, phase_completed_at,
		           selected_expert_ids, cost_budget_usd, cost_spent_usd,
		           cost_soft_limit_pct, cost_hard_limit_pct,
		           generic_allowance_pct,
		           created_at, updated_at`,
		nextPhase, now, StatusRunning, workflowID,
	).Scan(
		&updated.ID, &updated.ClientID, &updated.ProjectID, &updated.Title,
		&updated.Status, &updated.CurrentPhase,
		&updated.PhaseStartedAt, &updated.PhaseCompletedAt,
		&expertIDsRaw, &updated.CostBudgetUSD, &updated.CostSpentUSD,
		&updated.CostSoftLimitPct, &updated.CostHardLimitPct,
		&updated.GenericAllowancePct,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow transition: update: %w", err)
	}
	if err := json.Unmarshal(expertIDsRaw, &updated.SelectedExpertIDs); err != nil {
		updated.SelectedExpertIDs = []uuid.UUID{}
	}

	e.logger.Info("workflow phase transition",
		zap.String("workflow_id", workflowID.String()),
		zap.String("from", w.CurrentPhase),
		zap.String("to", nextPhase),
	)
	return &updated, nil
}

// RestartPhase sets current_phase WITHOUT the forward-only validation that
// TransitionPhase applies, and puts the workflow back to running.
//
// WHY a separate method instead of relaxing TransitionPhase: a workflow must
// never drift backwards on its own — that validation is what keeps the state
// machine honest. But a client who reads the deliverables, is not satisfied,
// and asks for the design to be produced again (optionally with a generic
// allowance) IS a legitimate backwards move, and it is driven by an explicit
// human action. Giving it its own named method keeps the two cases
// distinguishable in the code and in the logs.
func (e *Engine) RestartPhase(ctx context.Context, workflowID uuid.UUID, phase string) error {
	_, err := e.db.Exec(ctx,
		`UPDATE workflows SET
			current_phase = $1,
			phase_started_at = NOW(),
			phase_completed_at = NULL,
			status = $2,
			updated_at = NOW()
		 WHERE id = $3`,
		phase, StatusRunning, workflowID,
	)
	if err != nil {
		return fmt.Errorf("workflow restart phase: %w", err)
	}
	e.logger.Info("workflow phase restarted by client request",
		zap.String("workflow_id", workflowID.String()),
		zap.String("phase", phase),
	)
	return nil
}

// PauseForApproval pauses the workflow at a client approval gate.
// Called when the workflow engine creates an approval_request row.
func (e *Engine) PauseForApproval(ctx context.Context, workflowID uuid.UUID) error {
	_, err := e.db.Exec(ctx,
		`UPDATE workflows SET status = $1, updated_at = NOW() WHERE id = $2`,
		StatusPausedForApproval, workflowID,
	)
	if err != nil {
		return fmt.Errorf("workflow pause for approval: %w", err)
	}
	e.logger.Info("workflow paused for approval", zap.String("workflow_id", workflowID.String()))
	return nil
}

// Resume transitions a paused workflow back to running.
// Called after client responds to an approval request.
func (e *Engine) Resume(ctx context.Context, workflowID uuid.UUID) error {
	_, err := e.db.Exec(ctx,
		`UPDATE workflows SET status = $1, updated_at = NOW()
		 WHERE id = $2 AND status IN ($3, $4)`,
		StatusRunning, workflowID,
		StatusPausedForApproval, StatusPausedForInput,
	)
	if err != nil {
		return fmt.Errorf("workflow resume: %w", err)
	}
	e.logger.Info("workflow resumed", zap.String("workflow_id", workflowID.String()))
	return nil
}

// Complete marks the workflow as completed.
func (e *Engine) Complete(ctx context.Context, workflowID uuid.UUID) error {
	now := time.Now()
	_, err := e.db.Exec(ctx,
		`UPDATE workflows SET
			status = $1,
			current_phase = $2,
			phase_completed_at = $3,
			updated_at = NOW()
		 WHERE id = $4`,
		StatusCompleted, PhaseCompleted, now, workflowID,
	)
	if err != nil {
		return fmt.Errorf("workflow complete: %w", err)
	}
	e.logger.Info("workflow completed", zap.String("workflow_id", workflowID.String()))
	return nil
}

// Fail marks the workflow as failed with a reason.
func (e *Engine) Fail(ctx context.Context, workflowID uuid.UUID, reason string) error {
	_, err := e.db.Exec(ctx,
		`UPDATE workflows SET status = $1, updated_at = NOW() WHERE id = $2`,
		StatusFailed, workflowID,
	)
	if err != nil {
		return fmt.Errorf("workflow fail: %w", err)
	}
	e.logger.Error("workflow failed",
		zap.String("workflow_id", workflowID.String()),
		zap.String("reason", reason),
	)
	return nil
}

// GetByID loads a workflow by ID.
func (e *Engine) GetByID(ctx context.Context, workflowID uuid.UUID) (*Workflow, error) {
	var w Workflow
	var expertIDsRaw []byte
	err := e.db.QueryRow(ctx,
		`SELECT id, client_id, project_id, title, status, current_phase,
		        phase_started_at, phase_completed_at,
		        selected_expert_ids, cost_budget_usd, cost_spent_usd,
		        cost_soft_limit_pct, cost_hard_limit_pct,
		        generic_allowance_pct,
		        created_at, updated_at
		 FROM workflows WHERE id = $1`,
		workflowID,
	).Scan(
		&w.ID, &w.ClientID, &w.ProjectID, &w.Title, &w.Status, &w.CurrentPhase,
		&w.PhaseStartedAt, &w.PhaseCompletedAt,
		&expertIDsRaw, &w.CostBudgetUSD, &w.CostSpentUSD,
		&w.CostSoftLimitPct, &w.CostHardLimitPct,
		&w.GenericAllowancePct,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workflow get by id: %w", err)
	}
	if err := json.Unmarshal(expertIDsRaw, &w.SelectedExpertIDs); err != nil {
		w.SelectedExpertIDs = []uuid.UUID{}
	}
	return &w, nil
}

// UpdateCostSpent increments cost_spent_usd and checks soft/hard limits.
// Returns CostLimitResult indicating whether limits were crossed.
//
// Mental execution:
//   budget=10.0, soft=75%, hard=100%, current_spent=7.0, additional=1.0
//   new_spent = 8.0
//   soft_threshold = 10.0 * 0.75 = 7.5 → 8.0 > 7.5 → SoftLimitHit=true
//   hard_threshold = 10.0 * 1.00 = 10.0 → 8.0 < 10.0 → HardLimitHit=false
func (e *Engine) UpdateCostSpent(ctx context.Context, workflowID uuid.UUID, additionalCostUSD float64) (*CostLimitResult, error) {
	var w Workflow
	err := e.db.QueryRow(ctx,
		`UPDATE workflows SET
			cost_spent_usd = cost_spent_usd + $1,
			updated_at = NOW()
		 WHERE id = $2
		 RETURNING cost_spent_usd, cost_budget_usd, cost_soft_limit_pct, cost_hard_limit_pct`,
		additionalCostUSD, workflowID,
	).Scan(&w.CostSpentUSD, &w.CostBudgetUSD, &w.CostSoftLimitPct, &w.CostHardLimitPct)
	if err != nil {
		return nil, fmt.Errorf("workflow update cost: %w", err)
	}

	softThreshold := w.CostBudgetUSD * (w.CostSoftLimitPct / 100.0)
	hardThreshold := w.CostBudgetUSD * (w.CostHardLimitPct / 100.0)

	result := &CostLimitResult{
		CostSpentUSD: w.CostSpentUSD,
		SoftLimitHit: w.CostSpentUSD >= softThreshold,
		HardLimitHit: w.CostSpentUSD >= hardThreshold,
	}

	if result.HardLimitHit {
		e.logger.Warn("workflow hard cost limit hit",
			zap.String("workflow_id", workflowID.String()),
			zap.Float64("spent", w.CostSpentUSD),
			zap.Float64("budget", w.CostBudgetUSD),
		)
	} else if result.SoftLimitHit {
		e.logger.Warn("workflow soft cost limit hit",
			zap.String("workflow_id", workflowID.String()),
			zap.Float64("spent", w.CostSpentUSD),
			zap.Float64("threshold", softThreshold),
		)
	}
	return result, nil
}

// GetLatestCheckpoint loads the most recent phase checkpoint for a workflow.
// Returns nil if no checkpoint exists (new workflow).
func (e *Engine) GetLatestCheckpoint(ctx context.Context, workflowID uuid.UUID) (*Checkpoint, error) {
	var cp Checkpoint
	var snapshotRaw []byte
	err := e.db.QueryRow(ctx,
		`SELECT id, workflow_id, phase, state_snapshot,
		        blackboard_sequence_number, created_at
		 FROM workflow_checkpoints
		 WHERE workflow_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		workflowID,
	).Scan(
		&cp.ID, &cp.WorkflowID, &cp.Phase, &snapshotRaw,
		&cp.BlackboardSequenceNumber, &cp.CreatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // no checkpoint yet
		}
		return nil, fmt.Errorf("workflow get checkpoint: %w", err)
	}
	cp.StateSnapshot = snapshotRaw
	return &cp, nil
}

// Checkpoint is a phase-level snapshot for resume.
type Checkpoint struct {
	ID                      uuid.UUID       `json:"id"`
	WorkflowID              uuid.UUID       `json:"workflow_id"`
	Phase                   string          `json:"phase"`
	StateSnapshot           json.RawMessage `json:"state_snapshot"`
	BlackboardSequenceNumber int64          `json:"blackboard_sequence_number"`
	CreatedAt               time.Time       `json:"created_at"`
}

// writeCheckpoint writes a phase checkpoint to workflow_checkpoints.
func (e *Engine) writeCheckpoint(
	ctx context.Context,
	workflowID uuid.UUID,
	phase string,
	snapshot interface{},
	blackboardSeq int64,
) error {
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("checkpoint marshal: %w", err)
	}
	_, err = e.db.Exec(ctx,
		`INSERT INTO workflow_checkpoints
			(workflow_id, phase, state_snapshot, blackboard_sequence_number)
		 VALUES ($1, $2, $3, $4)`,
		workflowID, phase, string(snapshotJSON), blackboardSeq,
	)
	if err != nil {
		return fmt.Errorf("checkpoint insert: %w", err)
	}
	e.logger.Debug("workflow checkpoint written",
		zap.String("workflow_id", workflowID.String()),
		zap.String("phase", phase),
		zap.Int64("blackboard_seq", blackboardSeq),
	)
	return nil
}

// isValidTransition checks if moving from currentPhase to nextPhase is a
// valid forward progression in phaseOrder.
// Also allows any phase to transition to 'completed' (early termination).
func isValidTransition(current, next string) bool {
	if next == PhaseCompleted {
		return true
	}
	currentIdx := -1
	nextIdx := -1
	for i, p := range phaseOrder {
		if p == current {
			currentIdx = i
		}
		if p == next {
			nextIdx = i
		}
	}
	if currentIdx == -1 || nextIdx == -1 {
		return false
	}
	// next must be strictly after current
	return nextIdx > currentIdx
}
