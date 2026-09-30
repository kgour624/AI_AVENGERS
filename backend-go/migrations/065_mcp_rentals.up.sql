-- 065: Rental marketplace
CREATE TABLE IF NOT EXISTS mcp_rental_plans (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id        UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    name             TEXT NOT NULL CHECK (name IN ('Hourly','Daily','PerCall')),
    price_usd        NUMERIC(10,2) NOT NULL DEFAULT 0,
    token_limit      INTEGER NOT NULL DEFAULT 0,
    generic_allow    BOOLEAN NOT NULL DEFAULT FALSE,
    allowed_platforms TEXT[] NOT NULL DEFAULT '{claude,cursor}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS mcp_rentals (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    renter_client_id TEXT NOT NULL,
    expert_id        UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    plan_id          UUID NOT NULL REFERENCES mcp_rental_plans(id) ON DELETE CASCADE,
    status           TEXT NOT NULL CHECK (status IN ('active','expired','revoked')),
    started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ,
    usage_tokens     INTEGER NOT NULL DEFAULT 0,
    usage_cost       NUMERIC(10,6) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mcp_rentals_expert_status ON mcp_rentals(expert_id, status);
CREATE TABLE IF NOT EXISTS mcp_rental_usage_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rental_id      UUID NOT NULL REFERENCES mcp_rentals(id) ON DELETE CASCADE,
    usage_log_id   UUID NOT NULL REFERENCES mcp_usage_log(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);