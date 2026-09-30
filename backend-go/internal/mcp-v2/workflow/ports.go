// Package workflow ports — small interfaces, never import old workflow (R2).
package workflow

import "context"

// WorkflowPort is boundary for MCP-V2 workflow (Discover, don't design — RULE 8-B:32).
type WorkflowPort interface {
	ListWorkflows(ctx context.Context) ([]Workflow, error)
	QueueRun(ctx context.Context, workflowID, platform string, input []byte) (string, error)
	ClaimNextRun(ctx context.Context) (*Run, error)
}

// StorePort for postgres adapter.
type StorePort interface {
	InsertRun(ctx context.Context, workflowID string) (string, error)
}