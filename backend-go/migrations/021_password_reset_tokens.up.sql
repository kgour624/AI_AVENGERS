-- Migration 021: password_reset_tokens
-- Supports the forgot-password / reset-password flow.
--
-- WHY a separate table (not a column on users):
--   A user can have at most one active reset token at a time, but we
--   want to be able to invalidate old tokens without touching the users
--   row. A separate table also makes it trivial to add an index on
--   token_hash for O(1) lookup and to purge expired rows independently.
--
-- WHY store token_hash not the raw token:
--   The raw token is a secret equivalent to a temporary password.
--   Storing its SHA-256 hash means a DB breach does not expose usable
--   tokens — same reasoning as storing hashed_password on users.
--
-- WHY expires_at 1 hour:
--   Short enough to limit the window for a stolen link, long enough
--   for a user to check their email and act.

CREATE TABLE password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hex digest
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 hour'),
    used_at     TIMESTAMPTZ,                  -- NULL = not yet used
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prt_token_hash ON password_reset_tokens(token_hash)
    WHERE used_at IS NULL;
CREATE INDEX idx_prt_user_id    ON password_reset_tokens(user_id);
