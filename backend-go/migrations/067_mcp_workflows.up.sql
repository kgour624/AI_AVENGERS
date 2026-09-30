-- 067: MCP workflows (fully separate from internal/workflow)
CREATE TABLE IF NOT EXISTS mcp_workflows (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    steps       JSONB NOT NULL DEFAULT '[]',
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS mcp_workflow_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id   UUID NOT NULL REFERENCES mcp_workflows(id) ON DELETE CASCADE,
    platform      TEXT NOT NULL CHECK (platform IN ('claude','cursor','inspector','generic')),
    status        TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running','completed','failed','canceled','queued')),
    input         JSONB NOT NULL DEFAULT '{}',
    steps_results JSONB NOT NULL DEFAULT '[]',
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mcp_workflow_runs_workflow ON mcp_workflow_runs(workflow_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mcp_workflow_runs_status ON mcp_workflow_runs(status) WHERE status IN ('running','queued');