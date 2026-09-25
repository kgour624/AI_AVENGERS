-- Migration 044: a job status that tells the truth about a degraded corpus (Phase C)
--
-- WHY this migration exists:
--
-- 'complete' was written unconditionally at the end of every run, whatever the
-- store had actually produced and whatever the smoke test had said. Two facts
-- the pipeline computed were simply discarded: the storage verdict from
-- verifyStoredCorpus, and the smoke test result. So a run whose corpus was
-- missing rows — or whose expert could not answer a single probe — still told
-- the admin "Complete", while the experts row silently said draft. Two
-- contradictory truths, and the reassuring one was on the screen the admin was
-- watching.
--
-- The fix is a status that admits the run finished but did not fully succeed:
-- 'complete_with_warnings'. It means "do not treat this corpus as good". The
-- reason is written to ingestion_jobs.error_message, and the full breakdown stays
-- in ingestion_runs (migration 043).
--
-- PostgreSQL cannot add a value to a CHECK constraint in place, so the
-- constraint is dropped and recreated — the same approach migration 011 used to
-- introduce 'paused'.
--
-- Reversible: the down migration first rewrites any 'complete_with_warnings' row
-- to 'complete' (the run DID finish; only the nuance is lost), then restores the
-- narrower constraint. Without that rewrite the ADD CONSTRAINT would fail on
-- exactly the rows the new status was introduced for.

ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS job_status_check;

ALTER TABLE ingestion_jobs ADD CONSTRAINT job_status_check
    CHECK (status IN ('pending', 'running', 'complete', 'complete_with_warnings', 'failed', 'paused'));

COMMENT ON COLUMN ingestion_jobs.error_message IS
    'Failure reason for status=failed, and the warning summary for '
    'status=complete_with_warnings. NULL/empty means the run finished cleanly.';
