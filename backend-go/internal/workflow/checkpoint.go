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

// CheckpointManager writes and reads workflow phase checkpoints.
// Used for crash recovery and deployment resume.
//
// Two-level checkpointing (DOMAIN_EXPERT_COLLABORATION_DESIGN.md §12):
//   Level 1 (this file): phase-level snapshots after every phase transition.
//   Level 2: event-level via blackboard_events sequence_number cursor
//             (handled by blackboard.Subscriber).
type CheckpointManager struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewCheckpointManager creates a new checkpoint manager.
func NewCheckpointManager(db *pgxpool.Pool, logger *zap.Logger) *CheckpointManager {
	return &CheckpointManager{db: db, logger: logger}
}

// PhaseSnapshot is the data stored in a phase checkpoint.
// Contains everything needed to resume the next phase from scratch.
type PhaseSnapshot struct {
	CompletedPhase    string        `json:"completed_phase"`
	NextPhase         string        `json:"next_phase"`
	SelectedExpertIDs []uuid.UUID   `json:"selected_expert_ids"`
	FinalArtifacts    []ArtifactRef `json:"final_artifacts"`
	CostSpentUSD      float64       `json:"cost_spent_usd"`
	CreatedAt         time.Time     `json:"created_at"`
	Extra             []byte        `json:"extra,omitempty"`
}

// ArtifactRef is a reference to a final artifact on the blackboard.
type ArtifactRef struct {
	EventID    uuid.UUID `json:"event_id"`
	EventType  string    `json:"event_type"`
	ExpertID   uuid.UUID `json:"expert_id"`
	SequenceNo int64     `json:"sequence_no"`
}

// Write writes a phase checkpoint to workflow_checkpoints.
func (m *CheckpointManager) Write(
	ctx context.Context,
	workflowID uuid.UUID,
	snapshot PhaseSnapshot,
	blackboardSeq int64,
) error {
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("checkpoint write: marshal snapshot: %w", err)
	}

	_, err = m.db.Exec(ctx,
		`INSERT INTO workflow_checkpoints
			(workflow_id, phase, state_snapshot, blackboard_sequence_number)
		 VALUES ($1, $2, $3, $4)`,
		workflowID,
		snapshot.CompletedPhase,
		string(snapshotJSON),
		blackboardSeq,
	)
	if err != nil {
		return fmt.Errorf("checkpoint write: insert: %w", err)
	}

	m.logger.Info("checkpoint written",
		zap.String("workflow_id", workflowID.String()),
		zap.String("phase", snapshot.CompletedPhase),
		zap.Int64("blackboard_seq", blackboardSeq),
	)
	return nil
}

// LoadLatest loads the most recent checkpoint for a workflow.
// Returns nil if no checkpoint exists (new workflow, start from beginning).
func (m *CheckpointManager) LoadLatest(
	ctx context.Context,
	workflowID uuid.UUID,
) (*Checkpoint, error) {
	var cp Checkpoint
	var snapshotRaw []byte

	err := m.db.QueryRow(ctx,
		`SELECT id, workflow_id, phase, state_snapshot,
		        blackboard_sequence_number, created_at
		 FROM workflow_checkpoints
		 WHERE workflow_id = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		workflowID,
	).Scan(
		&cp.ID, &cp.WorkflowID, &cp.Phase,
		&snapshotRaw,
		&cp.BlackboardSequenceNumber,
		&cp.CreatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("checkpoint load: %w", err)
	}
	cp.StateSnapshot = snapshotRaw
	return &cp, nil
}

// LoadSnapshot unmarshals the state_snapshot JSON into a PhaseSnapshot.
func LoadSnapshot(cp *Checkpoint) (*PhaseSnapshot, error) {
	var snap PhaseSnapshot
	if err := json.Unmarshal(cp.StateSnapshot, &snap); err != nil {
		return nil, fmt.Errorf("load snapshot: %w", err)
	}
	return &snap, nil
}
