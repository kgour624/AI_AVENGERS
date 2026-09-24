-- Migration 030: tenant isolation (C4)
--
-- WHY: the system assumed a single tenant — every 'client' user could see
-- the whole expert catalog, and nothing named the isolation boundary an
-- enterprise customer needs. C4 introduces a first-class tenant and the
-- deterministic metadata pre-filter the design calls for (explicit
-- tenant_id columns + predicates) instead of hoping retrieval separates
-- tenants (P3: fail closed when scope is unknown).
--
-- STAGED + SAFE: existing rows are backfilled onto ONE 'default' tenant,
-- and existing experts stay NULL = platform/global (visible to every
-- tenant). So with isolation enabled the current single-tenant behaviour
-- is unchanged until an admin creates a second tenant and re-homes a
-- user/expert. Additive only; rollback = migration 030 down.

CREATE TABLE IF NOT EXISTS tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    slug       VARCHAR(255) UNIQUE NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'active',
    settings   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tenants_status_check CHECK (status IN ('active', 'suspended'))
);

COMMENT ON TABLE tenants IS
    'C4: enterprise isolation boundary. users/projects belong to one tenant; experts NULL = platform-shared.';

-- Seed the default tenant (idempotent).
INSERT INTO tenants (name, slug)
VALUES ('Default', 'default')
ON CONFLICT (slug) DO NOTHING;

-- users.tenant_id: NULL = global (admin) or not yet assigned → fails closed.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id) WHERE deleted_at IS NULL;

-- experts.tenant_id: NULL = platform/global expert (visible to all tenants).
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_experts_tenant ON experts(tenant_id) WHERE deleted_at IS NULL;

-- projects.tenant_id: explicit filter column (denormalised from the owner's
-- tenant) so retrieval/authorization never has to join to decide scope.
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_projects_tenant ON projects(tenant_id) WHERE deleted_at IS NULL;

-- Backfill: every non-admin user joins the default tenant. Admins stay NULL
-- (global scope — they operate across all tenants by design).
UPDATE users
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default')
WHERE tenant_id IS NULL AND role <> 'admin';

-- Backfill: every project joins its owner's tenant (falls back to default).
UPDATE projects p
SET tenant_id = COALESCE(
    (SELECT u.tenant_id FROM users u WHERE u.id = p.client_id),
    (SELECT id FROM tenants WHERE slug = 'default')
)
WHERE p.tenant_id IS NULL;
