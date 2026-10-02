package business

import "context"

// Storer is the Business Port — defines what Business needs from Storage.
// Storage adapter (Postgres) implements this interface (Ultimate Go §7).
// All methods use Business core models (types.go) — Storage translates via parse().
type Storer interface {
	ListExperts(ctx context.Context) ([]Expert, error)
	GetExpert(ctx context.Context, expertID string) (Expert, error)
	SearchChunks(ctx context.Context, expertID string, embedding []float32, limit int, cursor string) ([]Chunk, string, int, error)
	SearchRepoChunks(ctx context.Context, expertID string, embedding []float32, limit int) ([]RepoChunk, error)
	InsertUsageLog(ctx context.Context, row UsageLog) error
	GetExpertLimits(ctx context.Context, expertID, platform string) (ExpertLimits, error)
	IncLimits(ctx context.Context, expertID, platform string, tokens int) error

	// ListTools returns all active MCP tool definitions from mcp_v2_tools.
	// Used by Service.ListTools -> Handler.HandleListTools -> GET /api/v1/mcp-v2/tools
	ListTools(ctx context.Context) ([]ToolDefinition, error)

	// CreateTool inserts a new tool definition into mcp_v2_tools.
	CreateTool(ctx context.Context, t ToolDefinition) error

	// GetTool fetches single tool by name — for GenericExecutor dispatch.
	GetTool(ctx context.Context, name string) (ToolDefinition, error)

	// ExecSQLRead executes read-only SELECT with $1,$2 placeholders — for SQL_READ action.
	ExecSQLRead(ctx context.Context, query string, args []any) ([]map[string]any, error)
}