// Package workflow — Fully separate from backend-go/internal/workflow (R2). Own tables, own ports.
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Workflow is Business Model (core, no JSON tags on business logic).
type Workflow struct {
	ID          string
	Name        string
	Description string
	Steps       []Step
	Status      string
}

type Step struct {
	Step          int    `json:"step"`
	ExpertID      string `json:"expertId"`
	Tool          string `json:"tool"` // ask_expert|review_code|write_tests|generate_docs
	InputTemplate string `json:"inputTemplate"`
	DependsOn     *int   `json:"dependsOn"`
}

type Run struct {
	ID           string
	WorkflowID   string
	Platform     string
	Status       string // queued|running|completed|failed|canceled
	Input        json.RawMessage
	StepsResults []StepResult
	StartedAt    time.Time
	CompletedAt  *time.Time
}

type StepResult struct {
	Step      int             `json:"step"`
	ExpertID  string          `json:"expertId"`
	Tool      string          `json:"tool"`
	Output    json.RawMessage `json:"output"`
	Citations []string        `json:"citations"`
	DurationMs int            `json:"duration_ms"`
	Error     string          `json:"error,omitempty"`
}

// Service orchestrates run_workflow with FOR UPDATE SKIP LOCKED (RULE 8-F:47) + progress + cancel.
type Service struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewService(pool *pgxpool.Pool, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{pool: pool, logger: logger}
}

// ListWorkflows returns active workflows.
func (s *Service) ListWorkflows(ctx context.Context) ([]Workflow, error) {
	rows, err := s.pool.Query(ctx, `SELECT id::text, name, description, steps, status FROM mcp_workflows WHERE status='active' ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()
	var out []Workflow
	for rows.Next() {
		var w Workflow
		var stepsJSON []byte
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &stepsJSON, &w.Status); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(stepsJSON, &w.Steps)
		out = append(out, w)
	}
	return out, rows.Err()
}

// QueueRun creates a queued run — worker picks with FOR UPDATE SKIP LOCKED.
func (s *Service) QueueRun(ctx context.Context, workflowID string, platform string, input json.RawMessage) (string, error) {
	var runID string
	err := s.pool.QueryRow(ctx, `INSERT INTO mcp_workflow_runs (workflow_id, platform, status, input) VALUES ($1,$2,'queued',$3) RETURNING id::text`, workflowID, platform, input).Scan(&runID)
	if err != nil {
		return "", fmt.Errorf("queue run: %w", err)
	}
	return runID, nil
}

// ClaimNextRun is for worker — mandatory SKIP LOCKED for high throughput (RULE 8-F:47).
func (s *Service) ClaimNextRun(ctx context.Context) (*Run, error) {
	var r Run
	var input, results []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, workflow_id::text, platform, status, input, steps_results, started_at
		FROM mcp_workflow_runs
		WHERE status='queued'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED`).Scan(&r.ID, &r.WorkflowID, &r.Platform, &r.Status, &input, &results, &r.StartedAt)
	if err != nil {
		return nil, err
	}
	r.Input = input
	_ = json.Unmarshal(results, &r.StepsResults)
	return &r, nil
}