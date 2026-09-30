package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ToolDefinition struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
	HandlerKey  string          `json:"handler_key"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ToolStore struct{ pool *pgxpool.Pool }
func NewToolStore(pool *pgxpool.Pool) *ToolStore { return &ToolStore{pool: pool} }

func (s *ToolStore) ListActive(ctx context.Context) ([]ToolDefinition, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, display_name, description, input_schema, handler_key, is_active, created_at, updated_at FROM mcp_tool_definitions WHERE is_active = true ORDER BY name`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []ToolDefinition
	for rows.Next() {
		var t ToolDefinition
		var schema []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.DisplayName, &t.Description, &schema, &t.HandlerKey, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
		t.InputSchema = json.RawMessage(schema)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *ToolStore) ListAll(ctx context.Context) ([]ToolDefinition, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, display_name, description, input_schema, handler_key, is_active, created_at, updated_at FROM mcp_tool_definitions ORDER BY created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []ToolDefinition
	for rows.Next() {
		var t ToolDefinition
		var schema []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.DisplayName, &t.Description, &schema, &t.HandlerKey, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
		t.InputSchema = json.RawMessage(schema)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *ToolStore) Create(ctx context.Context, t ToolDefinition) (ToolDefinition, error) {
	t.ID = uuid.New()
	schema := string(t.InputSchema)
	if schema == "" { schema = `{"type":"object","properties":{}}` }
	err := s.pool.QueryRow(ctx, `INSERT INTO mcp_tool_definitions (id, name, display_name, description, input_schema, handler_key) VALUES ($1,$2,$3,$4,$5::jsonb,$6) RETURNING created_at, updated_at`, t.ID, t.Name, t.DisplayName, t.Description, schema, t.HandlerKey).Scan(&t.CreatedAt, &t.UpdatedAt)
	t.IsActive = true
	return t, err
}

func (s *ToolStore) Update(ctx context.Context, id uuid.UUID, t ToolDefinition) error {
	_, err := s.pool.Exec(ctx, `UPDATE mcp_tool_definitions SET display_name=$1, description=$2, input_schema=$3::jsonb, handler_key=$4, is_active=$5, updated_at=now() WHERE id=$6`, t.DisplayName, t.Description, string(t.InputSchema), t.HandlerKey, t.IsActive, id)
	return err
}
func (s *ToolStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM mcp_tool_definitions WHERE id=$1`, id)
	return err
}

func (s *ToolStore) ValidateTools(ctx context.Context, names []string) error {
	if len(names) == 0 { return nil }
	rows, err := s.pool.Query(ctx, `SELECT name FROM mcp_tool_definitions WHERE name = ANY($1) AND is_active=true`, names)
	if err != nil { return err }
	defer rows.Close()
	found := make(map[string]bool)
	for rows.Next() {
		var n string
		rows.Scan(&n)
		found[n] = true
	}
	for _, t := range names {
		if !found[t] {
			return &ToolValidationError{Name: t}
		}
	}
	return nil
}

type ToolValidationError struct{ Name string }
func (e *ToolValidationError) Error() string { return "invalid or inactive tool: " + e.Name }
