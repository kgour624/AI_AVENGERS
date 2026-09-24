-- Revert to the hard unique constraint from migration 001.
DROP INDEX IF EXISTS idx_users_email_unique_active;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email
    ON users (email)
    WHERE deleted_at IS NULL;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
