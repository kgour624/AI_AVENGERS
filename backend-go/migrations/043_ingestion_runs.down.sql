-- Migration 043 DOWN: remove the ingestion run ledger.
--
-- NOTE (honest limitation): this drops the audit trail of what each run actually
-- stored. The corpus itself (course_chunks) is untouched, so nothing becomes
-- unretrievable — but the parsed-vs-stored breakdown cannot be reconstructed
-- after the fact, because it depended on the INSERT command tag at write time.
-- Verify before downgrading if a mismatch is still open.

DROP INDEX IF EXISTS idx_ingestion_runs_expert_created;

DROP INDEX IF EXISTS ingestion_runs_job_file_key;

DROP TABLE IF EXISTS ingestion_runs;
