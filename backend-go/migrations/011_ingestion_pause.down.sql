-- ============================================================
-- Migration 011 DOWN: Remove ingestion pause state
-- ============================================================

-- Remove index first (depends on column)
DROP INDEX IF EXISTS idx_ingestion_jobs_paused;

-- Remove paused_at column
ALTER TABLE ingestion_jobs DROP COLUMN IF EXISTS paused_at;

-- Restore original 4-value status constraint
ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS job_status_check;
ALTER TABLE ingestion_jobs ADD CONSTRAINT job_status_check
    CHECK (status IN ('pending', 'running', 'complete', 'failed'));
