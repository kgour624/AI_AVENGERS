-- 062: MCP usage + per-expert per-platform limits
-- WHY: Token/Cost engine needs per-call row + hot limits snapshot. PG is Source of Truth (RULE 8-F:46), atomic counters via Redis but durable here.

CREATE TABLE IF NOT EXISTS mcp_usage_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id       UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    platform        TEXT NOT NULL CHECK (platform IN ('claude','cursor','inspector','generic')),
    provider        TEXT NOT NULL DEFAULT 'openrouter' CHECK (provider IN ('openrouter','codecraftapi','deepseek','openai','gemini','anthropic')),
    model           TEXT NOT NULL DEFAULT '',
    tier            TEXT NOT NULL CHECK (tier IN ('cheap','strong','fast')),
    input_tokens    INTEGER NOT NULL DEFAULT 0,
    output_tokens   INTEGER NOT NULL DEFAULT 0,
    total_tokens    INTEGER NOT NULL DEFAULT 0,
    cost_usd        NUMERIC(10,6) NOT NULL DEFAULT 0,
    generic_used    BOOLEAN NOT NULL DEFAULT FALSE,
    generic_percent INTEGER NOT NULL DEFAULT 0 CHECK (generic_percent IN (0,5,10,20)),
    duration_ms     INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mcp_usage_log_expert_platform_created ON mcp_usage_log(expert_id, platform, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mcp_usage_log_created ON mcp_usage_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mcp_usage_log_generic ON mcp_usage_log(generic_used) WHERE generic_used = TRUE;

CREATE TABLE IF NOT EXISTS mcp_expert_limits (
    expert_id           UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    platform            TEXT NOT NULL CHECK (platform IN ('claude','cursor','inspector','generic')),
    daily_token_limit   INTEGER NOT NULL DEFAULT 100000,
    monthly_token_limit INTEGER NOT NULL DEFAULT 2000000,
    current_daily       INTEGER NOT NULL DEFAULT 0,
    current_monthly     INTEGER NOT NULL DEFAULT 0,
    blocked_until       TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (expert_id, platform)
);