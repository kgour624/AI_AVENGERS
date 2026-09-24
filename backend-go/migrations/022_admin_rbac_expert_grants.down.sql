-- Reverse migration 022: admin RBAC + expert grants

DROP INDEX IF EXISTS idx_admin_bootstrap_token_hash;
DROP TABLE IF EXISTS admin_bootstrap_tokens;

DROP INDEX IF EXISTS idx_user_expert_grants_user;
DROP INDEX IF EXISTS idx_user_expert_grants_expert;
DROP TABLE IF EXISTS user_expert_grants;

-- Restore original role constraint. domain_expert rows must not exist
-- for this to succeed — down is for clean rollback in dev only.
UPDATE users SET role = 'client' WHERE role = 'domain_expert';
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'client'));
