-- ============================================================
-- Migration 011 rollback
-- ============================================================
ALTER TABLE messages
    DROP COLUMN IF EXISTS template_sections;
