-- Migration 012: Add updated_at to ingestion_jobs
-- ingestion_jobs was created without updated_at (migration 001).
-- ResumeIngestionJob and RetryIngestionJob in admin_handler.go
-- reference updated_at = NOW() in their UPDATE queries, causing
-- SQLSTATE 42703 (column does not exist) and 500 errors on resume/retry.
ALTER TABLE ingestion_jobs
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Backfill: set updated_at = completed_at for finished jobs,
-- started_at for running jobs, created_at for everything else.
UPDATE ingestion_jobs SET updated_at =
    CASE
        WHEN completed_at IS NOT NULL THEN completed_at
        WHEN started_at IS NOT NULL   THEN started_at
        ELSE created_at
    END
WHERE updated_at IS NULL;

COMMENT ON COLUMN ingestion_jobs.updated_at IS
    'Last modification timestamp. Added in migration 012 to fix
     ResumeIngestionJob/RetryIngestionJob 500 errors (SQLSTATE 42703).';
