-- Rollback 030: remove tenant isolation (additive migration).
DROP INDEX IF EXISTS idx_projects_tenant;
DROP INDEX IF EXISTS idx_experts_tenant;
DROP INDEX IF EXISTS idx_users_tenant;
ALTER TABLE projects DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE experts  DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users    DROP COLUMN IF EXISTS tenant_id;
DROP TABLE IF EXISTS tenants;
