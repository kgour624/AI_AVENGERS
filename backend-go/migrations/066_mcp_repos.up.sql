-- 066: MCP repo sync (GitHub/GitLab) + repo chunks pgvector
-- WHY: review_code needs live repo search without manual paste. Encrypted at rest like repo_connections.access_token (AES-256).

CREATE TABLE IF NOT EXISTS mcp_repo_connections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id           UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    provider            TEXT NOT NULL CHECK (provider IN ('github','gitlab')),
    repo_url            TEXT NOT NULL,
    branch              TEXT NOT NULL DEFAULT 'main',
    access_token_encrypted TEXT NOT NULL DEFAULT '',
    webhook_secret_ref  TEXT NOT NULL DEFAULT '',
    last_synced_at      TIMESTAMPTZ,
    status              TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled','error')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (expert_id, repo_url)
);
CREATE INDEX IF NOT EXISTS idx_mcp_repo_conn_expert ON mcp_repo_connections(expert_id);

CREATE TABLE IF NOT EXISTS mcp_repo_chunks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_connection_id  UUID NOT NULL REFERENCES mcp_repo_connections(id) ON DELETE CASCADE,
    file_path           TEXT NOT NULL,
    chunk_text          TEXT NOT NULL,
    embedding           vector(768),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mcp_repo_chunks_conn ON mcp_repo_chunks(repo_connection_id);
-- ivfflat requires rows before index; create conditionally via separate migration if needed
-- CREATE INDEX idx_mcp_repo_chunks_embedding ON mcp_repo_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);