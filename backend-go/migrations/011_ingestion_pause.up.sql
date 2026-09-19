-- ============================================================
-- Migration 011: Ingestion pause state
-- ============================================================
-- Adds 'paused' to ingestion_jobs.status so the pipeline can
-- stop mid-run when charter LLM fails and wait for admin action.
-- Adds paused_at for the 24h auto-fail background checker.
-- ============================================================

-- Step 1: Drop the existing status CHECK constraint.
-- WHY drop+recreate instead of ALTER: PostgreSQL does not support
-- ADD VALUE to a CHECK constraint in-place. Drop+recreate is the
-- only portable approach.
ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS job_status_check;

-- Step 2: Recreate with 'paused' included.
ALTER TABLE ingestion_jobs ADD CONSTRAINT job_status_check
    CHECK (status IN ('pending', 'running', 'complete', 'failed', 'paused'));

ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS ingestion_jobs_stage_check;
ALTER TABLE ingestion_jobs ADD CONSTRAINT ingestion_jobs_stage_check
    CHECK (current_stage IN ('pending', 'chunking', 'topic_extraction', 'charter_extraction', 'embedding', 'storing', 'smoke_test', 'complete', 'failed', 'paused'));

-- Step 3: Add paused_at column.
-- Set when status transitions to 'paused'.
-- NULL for all non-paused jobs.
-- Used by the 24h auto-fail background checker:
--   WHERE status='paused' AND paused_at < NOW() - INTERVAL '24 hours'
ALTER TABLE ingestion_jobs
    ADD COLUMN IF NOT EXISTS paused_at TIMESTAMPTZ;

-- Step 4: Index for the auto-fail checker query.
-- Partial index: only paused rows. Tiny index, fast lookup.
CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_paused
    ON ingestion_jobs(paused_at)
    WHERE status = 'paused';

COMMENT ON COLUMN ingestion_jobs.paused_at IS
    'Set when job transitions to status=paused (LLM failure during charter extraction). '
    'Used by the 24h auto-fail background checker. NULL for all non-paused jobs.';
