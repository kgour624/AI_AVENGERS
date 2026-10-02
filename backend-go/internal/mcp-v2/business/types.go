// Package business — Core domain types (Business model — RULE 8-B:33, Ultimate Go §7).
// Business defines its own types; App imports Business (imports down only — §1).
package business

// Expert is Business model (core). No JSON tags.
type Expert struct {
	ID      string
	Name    string
	Slug    string
	Charter string
}

type Chunk struct {
	ID       string
	ExpertID string
	Text     string
	Score    float32
}

type RepoChunk struct {
	ID       string
	FilePath string
	Text     string
	Score    float32
}

type UsageLog struct {
	ExpertID       string
	Platform       string
	Provider       string
	Model          string
	Tier           string
	InputTokens    int
	OutputTokens   int
	TotalTokens    int
	CostUSD        float64
	GenericUsed    bool
	GenericPercent int
	DurationMs     int
}

type ExpertLimits struct {
	ExpertID         string
	Platform         string
	DailyLimit       int
	MonthlyLimit     int
	CurrentDaily     int
	CurrentMonthly   int
	BlockedUntilUnix *int64
}

type Message struct{ Role, Content string }

type GatewayUsage struct{ InputTokens, OutputTokens int; CostUSD float64 }

type GateResult struct{ GateStopped int; Allowed bool }

// ToolDefinition is the Core Business Model for MCP tools — single source of truth.
// Maps to mcp_v2_tools table (id, name, display_name, description, is_active, input_schema, action).
// Business model — used by Storer -> Service -> Handler -> Frontend.
// InputSchema is the JSON Schema for MCP tools/list (JSONB in DB -> map[string]any in Go).
// Action is the Dynamic Engine interpreter config (Phase 1: SQL_READ). JSONB -> ActionDef.
type ToolDefinition struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	IsActive    bool           `json:"is_active"`
	InputSchema map[string]any `json:"input_schema"`
	Action      ActionDef      `json:"action"`
}

// ActionType defines interpreter strategy — Interpreter Pattern (not code-gen).
type ActionType string

const (
	ActionSQLRead  ActionType = "SQL_READ"
	ActionAPICall  ActionType = "API_CALL"
	ActionLLMPrompt ActionType = "LLM_PROMPT"
	ActionComposite ActionType = "COMPOSITE"
)

// ActionDef is the Dynamic Engine payload stored as JSONB in mcp_v2_tools.action.
// Phase 1: SQL_READ only — {type:"SQL_READ", config:{query, allow_write:false}, timeout_ms}.
// Future types need no ALTER TABLE — just new ActionType + Executor strategy.
type ActionDef struct {
	Type          ActionType     `json:"type"`
	Config        map[string]any `json:"config"`
	OutputMapping string         `json:"output_mapping,omitempty"`
	TimeoutMs     int            `json:"timeout_ms,omitempty"`
}

// IsEmpty reports whether no dynamic action is configured (legacy static tool).
func (a ActionDef) IsEmpty() bool { return a.Type == "" }

// SQLQuery extracts SQL query from config — safe accessor.
func (a ActionDef) SQLQuery() string {
	if a.Config == nil {
		return ""
	}
	q, _ := a.Config["query"].(string)
	return q
}