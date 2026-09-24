-- Migration 031: usage events + budgets (C5)
--
-- WHY: cost was invisible. The gateway computed a cost on every LLM call
-- but only kept it in memory (and workflows.cost_spent_usd for the workflow
-- path) — messages.cost_usd was never written, so per-project / per-expert /
-- per-tenant spend could not be reported or capped. C5 makes cost a product
-- surface: every real (non-cached) LLM call is appended here with whatever
-- attribution the caller had (tenant/project/expert/chat/workflow/use-case),
-- and budgets/alert thresholds live in usage_budgets.
--
-- Append-only, additive. Rollback = migration 031 down.

CREATE TABLE IF NOT EXISTS usage_events (
    id            BIGSERIAL PRIMARY KEY,
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id     UUID REFERENCES tenants(id)  ON DELETE SET NULL,
    project_id    UUID REFERENCES projects(id) ON DELETE SET NULL,
    expert_id     UUID REFERENCES experts(id)  ON DELETE SET NULL,
    chat_id       UUID,        -- no FK: chat rows may be archived/purged
    workflow_id   UUID,        -- no FK: workflow rows may be purged
    account_id    UUID REFERENCES users(id)    ON DELETE SET NULL,
    provider      VARCHAR(50),
    tier          VARCHAR(20), -- cheap | strong | fast
    model         VARCHAR(150),
    use_case      VARCHAR(50), -- chat | synthesis | memory_consolidate | summary | index | workflow | ...
    input_tokens  INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd      DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE usage_events IS
    'C5: append-only per-LLM-call usage. Attribution columns nullable; tenant derived from project when absent.';

CREATE INDEX IF NOT EXISTS idx_usage_events_time    ON usage_events(occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant  ON usage_events(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_events_project ON usage_events(project_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_events_expert  ON usage_events(expert_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_events_model   ON usage_events(model);

-- Budgets: one row per tenant, plus one global row (tenant_id NULL) that
-- applies as the platform default. The COALESCE unique index treats NULL as
-- a concrete sentinel so both "one global row" and "one row per tenant" hold.
CREATE TABLE IF NOT EXISTS usage_budgets (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID REFERENCES tenants(id) ON DELETE CASCADE,
    monthly_limit_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    alert_threshold   DOUBLE PRECISION NOT NULL DEFAULT 0.8,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT usage_budgets_limit_check CHECK (monthly_limit_usd >= 0),
    CONSTRAINT usage_budgets_threshold_check CHECK (alert_threshold > 0 AND alert_threshold <= 1)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_usage_budgets_scope
    ON usage_budgets((COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid)));

COMMENT ON TABLE usage_budgets IS
    'C5: monthly cost budget + alert threshold per tenant; tenant_id NULL = platform default.';
