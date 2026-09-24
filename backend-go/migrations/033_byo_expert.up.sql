-- Migration 033: bring-your-own-expert (C8)
--
-- WHY: experts were admin-provisioned only. C8 lets a tenant register its own
-- expert (corpus + charter) and ingest source material self-service, without
-- an admin. The isolation boundary already exists (experts.tenant_id, mig 030),
-- so this migration only adds provenance/audit columns and an audit table —
-- no new isolation mechanism (P3: reuse the C4 deterministic pre-filter).
--
-- SAFE + ADDITIVE: origin defaults to 'admin' so every existing expert is
-- unchanged. byo_expert_events is a new, append-only audit trail (C10-friendly).
-- Rollback = migration 033 down.

-- Which path created the expert: 'admin' (platform/admin API) or 'byo'
-- (tenant self-service). Only 'byo' experts are self-managed by their owner.
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS origin VARCHAR(20) NOT NULL DEFAULT 'admin';
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE experts
    DROP CONSTRAINT IF EXISTS experts_origin_check;
ALTER TABLE experts
    ADD CONSTRAINT experts_origin_check CHECK (origin IN ('admin', 'byo'));

CREATE INDEX IF NOT EXISTS idx_experts_origin
    ON experts(tenant_id, origin) WHERE deleted_at IS NULL;

COMMENT ON COLUMN experts.origin IS
    'C8: admin = provisioned via admin API; byo = tenant self-registered, owner-managed.';
COMMENT ON COLUMN experts.created_by_user_id IS
    'C8: the tenant user who registered a byo expert (NULL for admin/legacy rows).';

-- Append-only audit of BYO activity (register / ingest started / completed /
-- failed / entitlement denials). Never updated or deleted — an audit trail.
CREATE TABLE IF NOT EXISTS byo_expert_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID REFERENCES tenants(id) ON DELETE SET NULL,
    user_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    expert_id  UUID REFERENCES experts(id) ON DELETE SET NULL,
    action     VARCHAR(40) NOT NULL,
    detail     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_byo_events_tenant
    ON byo_expert_events(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_byo_events_expert
    ON byo_expert_events(expert_id, created_at DESC);

COMMENT ON TABLE byo_expert_events IS
    'C8: append-only audit of tenant BYO-expert activity (register, ingest, denials).';
