-- Rollback migration 004
ALTER TABLE messages
    DROP COLUMN IF EXISTS warning_text,
    DROP COLUMN IF EXISTS clarifying_questions;
