-- Migration 044 DOWN: remove the complete_with_warnings status.
--
-- Rows already carrying it are folded back to 'complete' BEFORE the narrower
-- constraint is restored — they did finish, and leaving them would make the
-- ADD CONSTRAINT fail. The nuance ("this corpus is degraded") is lost on
-- rollback; the ingestion_runs ledger from migration 043 keeps the underlying
-- numbers, so nothing factual disappears.

UPDATE ingestion_jobs
   SET status = 'complete',
       error_message = NULL
 WHERE status = 'complete_with_warnings';

ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS job_status_check;

ALTER TABLE ingestion_jobs ADD CONSTRAINT job_status_check
    CHECK (status IN ('pending', 'running', 'complete', 'failed', 'paused'));

COMMENT ON COLUMN ingestion_jobs.error_message IS
    'Failure reason for status=failed. NULL/empty means the run finished cleanly.';
