-- 074_mcp_v2_action.up.sql
-- Dynamic Tool Engine — Interpreter Pattern (Phase 1: SQL_READ + Magic Generate)
-- Adds action JSONB to mcp_v2_tools: {type, config{query, allow_write, timeout_ms}, output_mapping}
ALTER TABLE mcp_v2_tools
  ADD COLUMN IF NOT EXISTS action JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Index for filtering by action type (e.g., WHERE action->>'type' = 'SQL_READ')
CREATE INDEX IF NOT EXISTS idx_mcp_v2_tools_action_type ON mcp_v2_tools ((action->>'type'));

COMMENT ON COLUMN mcp_v2_tools.action IS 'Dynamic Engine Action: {type: SQL_READ|API_CALL|LLM_PROMPT|COMPOSITE, config: {...}, output_mapping, timeout_ms}';