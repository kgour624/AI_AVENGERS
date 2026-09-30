-- 063: MCP per-expert per-platform provider map (cheap|strong|fast)
-- WHY: Multi-provider router needs per-expert override + fallback to system_settings. Vault ref encrypted at rest (AES-256) like repo_connections.access_token.

CREATE TABLE IF NOT EXISTS mcp_expert_provider_map (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id     UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    platform      TEXT NOT NULL CHECK (platform IN ('claude','cursor','generic')),
    tier          TEXT NOT NULL CHECK (tier IN ('cheap','strong','fast')),
    provider      TEXT NOT NULL CHECK (provider IN ('openrouter','codecraftapi','deepseek','openai','gemini','anthropic')),
    model         TEXT NOT NULL DEFAULT '',
    api_key_ref   TEXT NOT NULL DEFAULT '',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (expert_id, platform, tier)
);
CREATE INDEX IF NOT EXISTS idx_mcp_provider_map_expert ON mcp_expert_provider_map(expert_id);