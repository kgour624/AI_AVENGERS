-- 070_mcp_v2_tools.down.sql
-- Rollback for mcp-v2 tools

DROP TRIGGER IF EXISTS trg_mcp_v2_tools_updated_at ON mcp_v2_tools;
DROP FUNCTION IF EXISTS update_mcp_v2_tools_updated_at();
DROP TABLE IF EXISTS mcp_v2_tools;
