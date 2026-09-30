-- 064: MCP generic relief config per expert (Vault-style time bound)
-- WHY: DecisionEngine gate2 wrapper needs allow_generic + percent + time_bound + reason + L3 audit. No edit in internal/decision/engine.go.

CREATE TABLE IF NOT EXISTS mcp_expert_generic_config (
    expert_id         UUID PRIMARY KEY REFERENCES experts(id) ON DELETE CASCADE,
    allow_generic     BOOLEAN NOT NULL DEFAULT FALSE,
    generic_percent   INTEGER NOT NULL DEFAULT 0 CHECK (generic_percent IN (0,5,10,20)),
    time_bound_until  TIMESTAMPTZ,
    reason            TEXT NOT NULL DEFAULT '',
    set_by            UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);