-- 074_mcp_v2_action.down.sql
DROP INDEX IF EXISTS idx_mcp_v2_tools_action_type;
ALTER TABLE mcp_v2_tools DROP COLUMN IF EXISTS action;
