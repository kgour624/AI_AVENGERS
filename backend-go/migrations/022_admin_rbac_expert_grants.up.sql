-- Migration 022: admin bootstrap RBAC + per-account expert grants
--
-- WHY additive role 'domain_expert':
--   Existing public self-register still creates 'client' (all trained
--   experts visible). Admin-created 'domain_expert' accounts only see
--   and use experts explicitly granted in user_expert_grants.
--
-- WHY user_expert_grants (not project_experts):
--   project_experts is project-scoped selection. This table is
--   account-level authorization — who is allowed to use which AI expert.
--
-- WHY admin_bootstrap_tokens:
--   First admin must not be created on the public login/register pages.
--   Bootstrap uses a one-time hashed token (same pattern as 021 password
--   reset). The raw token is never stored. Pending fields hold the
--   in-progress registration until TOTP is verified.

-- Extend role constraint: admin | client | domain_expert
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'client', 'domain_expert'));

CREATE TABLE IF NOT EXISTS user_expert_grants (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expert_id  UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, expert_id)
);

CREATE INDEX IF NOT EXISTS idx_user_expert_grants_expert
    ON user_expert_grants(expert_id);
CREATE INDEX IF NOT EXISTS idx_user_expert_grants_user
    ON user_expert_grants(user_id);

CREATE TABLE IF NOT EXISTS admin_bootstrap_tokens (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash             VARCHAR(64) NOT NULL UNIQUE,
    expires_at             TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    used_at                TIMESTAMPTZ,
    pending_email          VARCHAR(255),
    pending_full_name      VARCHAR(255),
    pending_password_hash  VARCHAR(255),
    pending_totp_secret    VARCHAR(255),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_bootstrap_token_hash
    ON admin_bootstrap_tokens(token_hash)
    WHERE used_at IS NULL;
