-- Migration 024: make users.email uniqueness respect soft-delete
--
-- WHY: users.email had a hard UNIQUE constraint (001). A soft-deleted account
-- (deleted_at set, added by the admin delete flow) kept its email reserved
-- forever, so re-creating an account with the same email failed at the DB even
-- though every application-level check (WHERE deleted_at IS NULL) said it was
-- free. This replaces the hard constraint with a partial unique index so an
-- email can be reused once the old account is deleted.
--
-- Additive/safe: drops the auto-named constraint and adds an equivalent
-- partial unique index; no data change.

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

-- Partial unique: at most one LIVE user per email; deleted rows are ignored.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique_active
    ON users (email)
    WHERE deleted_at IS NULL;

-- Drop the now-redundant non-unique partial index from 001 (the unique one
-- above serves the same lookups).
DROP INDEX IF EXISTS idx_users_email;
