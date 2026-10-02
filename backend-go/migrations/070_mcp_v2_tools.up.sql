-- 070_mcp_v2_tools.up.sql
-- mcp-v2: Single source of truth for MCP tools (replaces hardcoded []string)
-- Hexagon: Storage (postgres) -> Business (ToolDefinition) -> Service -> Handler

CREATE TABLE IF NOT EXISTS mcp_v2_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE, -- canonical tool name e.g. 'search_course_content' - used in token payload
    display_name TEXT NOT NULL, -- UI label e.g. 'Search Course Content'
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mcp_v2_tools_is_active ON mcp_v2_tools(is_active);
CREATE INDEX IF NOT EXISTS idx_mcp_v2_tools_name ON mcp_v2_tools(name);

-- Auto-update updated_at on modification
DROP TRIGGER IF EXISTS trg_mcp_v2_tools_updated_at ON mcp_v2_tools;
CREATE OR REPLACE FUNCTION update_mcp_v2_tools_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mcp_v2_tools_updated_at
BEFORE UPDATE ON mcp_v2_tools
FOR EACH ROW EXECUTE FUNCTION update_mcp_v2_tools_updated_at();

-- Seed initial tools (from old hardcoded list + expanded for v2)
INSERT INTO mcp_v2_tools (name, display_name, description) VALUES
('search_course_content', 'Search Course Content', 'Semantic search over course chunks via pgvector (experts)'),
('search_repo_content', 'Search Repo Content', 'Semantic search over connected GitHub repo chunks'),
('list_experts', 'List Experts', 'List all active domain experts'),
('get_usage_logs', 'Get Usage Logs', 'Fetch MCP token usage analytics and limits')
ON CONFLICT (name) DO NOTHING;
