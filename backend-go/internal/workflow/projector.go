package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/blackboard"
)

// Projector subscribes to blackboard events and projects them into
// the workflow_tasks table (Kanban state).
//
// DESIGN: Single Write Path
//   Runner → Post event on blackboard
//   Projector → reads event → updates workflow_tasks
//   Kanban SSE → reads from same Redis channel (no DB polling)
//
// WHY Projector not direct INSERT:
//   If Runner writes directly to workflow_tasks AND blackboard,
//   they can diverge on failure (e.g. blackboard write succeeds,
//   DB write fails). Projector ensures workflow_tasks is always
//   a faithful projection of blackboard_events.
//
// Event → Projection mapping:
//   task_plan_ready      → INSERT workflow_tasks rows
//   task_status_changed  → UPDATE workflow_tasks.status
//   task_failed          → UPDATE workflow_tasks.status = 'failed'
//   artifact posted      → UPDATE workflow_tasks.produced_artifact_event_id
type Projector struct {
	db     *pgxpool.Pool
	store  *blackboard.Store
	sub    *blackboard.Subscriber
	logger *zap.Logger
}

// NewProjector creates a new Projector.
func NewProjector(db *pgxpool.Pool, store *blackboard.Store, sub *blackboard.Subscriber, logger *zap.Logger) *Projector {
	return &Projector{db: db, store: store, sub: sub, logger: logger}
}

// Run starts the projection loop for a workflow.
// Designed to run in a goroutine alongside WorkflowRunner.
// Stops when ctx is cancelled (workflow complete or failed).
//
// Mental execution:
//   workflowID = abc
//   Subscribe from seq=0
//
//   Event: task_plan_ready {tasks: [{expert_id, title, description}, ...]}
//   → INSERT workflow_tasks for each task
//
//   Event: task_status_changed {expert_id, status: "in_progress"}
//   → UPDATE workflow_tasks SET status='in_progress' WHERE assigned_expert_id=expert_id
//
//   Event: architecture_decision (artifact)
//   → UPDATE workflow_tasks SET produced_artifact_event_id=event.ID
//      WHERE assigned_expert_id=event.PostedByExpertID
func (p *Projector) Run(ctx context.Context, workflowID uuid.UUID) {
	p.logger.Info("projector started", zap.String("workflow_id", workflowID.String()))

	eventCh, errCh := p.sub.Subscribe(ctx, workflowID, uuid.Nil, 0)

	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				p.logger.Info("projector: event channel closed",
					zap.String("workflow_id", workflowID.String()),
				)
				return
			}
			p.project(ctx, workflowID, event)

		case err := <-errCh:
			if err != nil {
				p.logger.Warn("projector: subscriber error",
					zap.String("workflow_id", workflowID.String()),
					zap.Error(err),
				)
			}
			return

		case <-ctx.Done():
			return
		}
	}
}

// project handles one blackboard event and updates workflow_tasks accordingly.
func (p *Projector) project(ctx context.Context, workflowID uuid.UUID, event blackboard.Event) {
	switch event.EventType {

	case "task_plan_ready":
		// INSERT workflow_tasks rows from the plan.
		// Content: {tasks: [{expert_id, title, description}, ...]}
		var plan struct {
			Tasks []struct {
				ExpertID    string `json:"expert_id"`
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"tasks"`
		}
		if err := json.Unmarshal(event.Content, &plan); err != nil {
			p.logger.Warn("projector: parse task_plan_ready failed", zap.Error(err))
			return
		}
		for _, t := range plan.Tasks {
			expertID, err := uuid.Parse(t.ExpertID)
			if err != nil {
				continue
			}
			_, err = p.db.Exec(ctx,
				`INSERT INTO workflow_tasks
					(workflow_id, assigned_expert_id, title, description, status)
				 VALUES ($1, $2, $3, $4, 'todo')
				 ON CONFLICT (workflow_id, assigned_expert_id) DO NOTHING`,
				workflowID, expertID, t.Title, t.Description,
			)
			if err != nil {
				p.logger.Warn("projector: insert task failed",
					zap.String("expert_id", t.ExpertID),
					zap.Error(err),
				)
			}
		}

	case "task_status_changed":
		// UPDATE workflow_tasks.status
		// Content: {expert_id, status, started_at?, completed_at?}
		var payload struct {
			ExpertID string `json:"expert_id"`
			Status   string `json:"status"`
		}
		if err := json.Unmarshal(event.Content, &payload); err != nil {
			return
		}
		expertID, err := uuid.Parse(payload.ExpertID)
		if err != nil {
			return
		}
		if _, err := p.db.Exec(ctx,
			`UPDATE workflow_tasks SET
				status = $1,
				started_at   = CASE WHEN $1 = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
				completed_at = CASE WHEN $1 IN ('done','failed') THEN NOW() ELSE completed_at END,
				updated_at   = NOW()
			 WHERE workflow_id = $2 AND assigned_expert_id = $3`,
			payload.Status, workflowID, expertID,
		); err != nil {
			p.logger.Warn("projector: task_status_changed update failed",
				zap.String("workflow_id", workflowID.String()),
				zap.String("expert_id", expertID.String()),
				zap.String("status", payload.Status),
				zap.Error(err),
			)
		}

	case "task_failed":
		// UPDATE workflow_tasks.status = 'failed'
		// Content: {expert_id, reason}
		var payload struct {
			ExpertID string `json:"expert_id"`
		}
		if err := json.Unmarshal(event.Content, &payload); err != nil {
			return
		}
		expertID, err := uuid.Parse(payload.ExpertID)
		if err != nil {
			return
		}
		if _, err := p.db.Exec(ctx,
			`UPDATE workflow_tasks SET status='failed', completed_at=NOW(), updated_at=NOW()
			 WHERE workflow_id=$1 AND assigned_expert_id=$2`,
			workflowID, expertID,
		); err != nil {
			p.logger.Warn("projector: task_failed update failed",
				zap.String("workflow_id", workflowID.String()),
				zap.String("expert_id", expertID.String()),
				zap.Error(err),
			)
		}

	case "code_artifact_produced":
		// INSERT new workflow_tasks row for each code file.
		// Each file is a separate artifact that should be visible in Kanban.
		//
		// MENTAL MODEL:
		//   Input: event with {filename, file_path, content, language, lines_of_code}
		//   Action: INSERT workflow_tasks row
		//   Output: New row with status='done', artifact_type='code_file'
		//
		// CROSS-QUESTION:
		//   Q: Why INSERT instead of UPDATE?
		//   A: Multiple files should create multiple rows, not overwrite one row
		//
		//   Q: Why status='done'?
		//   A: File already created by Aider, not a future task
		//
		//   Q: What if event.PostedByExpertID is nil?
		//   A: Skip event (invalid, should always have expert_id)
		//
		// EXAMPLE:
		//   Event: {filename: "auth.go", language: "go", lines_of_code: 150}
		//   Result: INSERT workflow_tasks (
		//     title = "Generated: auth.go",
		//     status = 'done',
		//     artifact_type = 'code_file',
		//     artifact_data = {filename, file_path, language, lines_of_code}
		//   )
		p.logger.Debug("code_artifact_produced: received event",
			zap.String("workflow_id", workflowID.String()),
			zap.String("event_id", event.ID.String()),
		)

		if event.PostedByExpertID == nil {
			p.logger.Warn("code_artifact_produced: missing expert_id",
				zap.String("event_id", event.ID.String()),
			)
			return
		}

		// Parse event data
		var artifactData struct {
			Filename     string `json:"filename"`
			FilePath     string `json:"file_path"`
			Language     string `json:"language"`
			LinesOfCode  int    `json:"lines_of_code"`
			CommitSHA    string `json:"commit_sha"`
		}
		if err := json.Unmarshal(event.Content, &artifactData); err != nil {
			p.logger.Warn("code_artifact_produced: parse failed",
				zap.String("event_id", event.ID.String()),
				zap.Error(err),
			)
			return
		}

		// Validate required fields
		if artifactData.Filename == "" {
			p.logger.Warn("code_artifact_produced: missing filename",
				zap.String("event_id", event.ID.String()),
			)
			return
		}

		// Use fallback for optional fields
		if artifactData.Language == "" {
			artifactData.Language = "unknown"
		}
		if artifactData.FilePath == "" {
			artifactData.FilePath = artifactData.Filename
		}

		// INSERT workflow_tasks row
		title := fmt.Sprintf("Generated: %s", artifactData.Filename)
		description := fmt.Sprintf("Code file: %s (%s, %d lines)",
			artifactData.FilePath,
			artifactData.Language,
			artifactData.LinesOfCode,
		)

		_, err := p.db.Exec(ctx,
			`INSERT INTO workflow_tasks (
				workflow_id,
				assigned_expert_id,
				title,
				description,
				status,
				artifact_type,
				artifact_data,
				produced_artifact_event_id,
				started_at,
				completed_at
			) VALUES ($1, $2, $3, $4, 'done', 'code_file', $5, $6, NOW(), NOW())`,
			workflowID,
			*event.PostedByExpertID,
			title,
			description,
			event.Content, // Store full artifact data as JSON
			event.ID,
		)
		if err != nil {
			p.logger.Warn("code_artifact_produced: insert task failed",
				zap.String("filename", artifactData.Filename),
				zap.Error(err),
			)
			return
		}

		p.logger.Info("code_artifact_produced: task created",
			zap.String("workflow_id", workflowID.String()),
			zap.String("filename", artifactData.Filename),
			zap.String("language", artifactData.Language),
			zap.Int("lines", artifactData.LinesOfCode),
		)

	default:
		// Artifact events: update produced_artifact_event_id.
		// Note: code_artifact_produced has its own case handler above.
		artifactTypes := map[string]bool{
			"architecture_decision":  true,
			"data_model_proposed":    true,
			"api_contract_proposed":  true,
			"module_design_proposed": true,
			"test_case_proposed":     true,
			"requirement_captured":   true,
		}
		if !artifactTypes[event.EventType] {
			return
		}
		if event.PostedByExpertID == nil {
			return
		}
		if _, err := p.db.Exec(ctx,
			`UPDATE workflow_tasks
			 SET produced_artifact_event_id = $1, updated_at = NOW()
			 WHERE workflow_id = $2 AND assigned_expert_id = $3`,
			event.ID, workflowID, *event.PostedByExpertID,
		); err != nil {
			p.logger.Warn("projector: artifact link update failed",
				zap.String("workflow_id", workflowID.String()),
				zap.String("event_type", event.EventType),
				zap.String("event_id", event.ID.String()),
				zap.String("expert_id", event.PostedByExpertID.String()),
				zap.Error(err),
			)
		}
	}
}

// PostTaskStatus posts a task_status_changed event on the blackboard.
// Called by AgentLoop instead of direct DB write.
// Projector will pick this up and update workflow_tasks.
func PostTaskStatus(ctx context.Context, store *blackboard.Store, workflowID, expertID uuid.UUID, status string) error {
	_, err := store.Post(ctx, blackboard.PostRequest{
		WorkflowID:       workflowID,
		EventType:        "task_status_changed",
		PostedByExpertID: &expertID,
		Content: map[string]string{
			"expert_id": expertID.String(),
			"status":    status,
		},
	})
	if err != nil {
		return fmt.Errorf("PostTaskStatus: %w", err)
	}
	return nil
}

// PostTaskFailed posts a task_failed event on the blackboard.
// Called by AgentLoop when LLM call fails fatally.
func PostTaskFailed(ctx context.Context, store *blackboard.Store, workflowID, expertID uuid.UUID, reason string) error {
	_, err := store.Post(ctx, blackboard.PostRequest{
		WorkflowID:       workflowID,
		EventType:        "task_failed",
		PostedByExpertID: &expertID,
		Content: map[string]interface{}{
			"expert_id": expertID.String(),
			"reason":    reason,
		},
	})
	if err != nil {
		return fmt.Errorf("PostTaskFailed: %w", err)
	}
	return nil
}
