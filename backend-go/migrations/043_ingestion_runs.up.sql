-- Migration 043: one durable, queryable ledger per ingestion run (Phase B)
--
-- WHY this migration exists:
--
-- The ingestion timeline (ingestion_job_events) records history, but history is
-- not queryable state: answering "which of this expert's files landed short, and
-- by how much?" meant reading events by hand. Worse, the run's own numbers were
-- WRONG. The pipeline compared the chunks it PARSED against the rows in
-- course_chunks and reported the difference as a mismatch — but that difference
-- is precisely what ON CONFLICT (expert_id, chunk_hash) DO NOTHING is designed to
-- produce. A 395-chunk document legitimately becomes 277 rows when 118 chunks
-- repeat text the corpus already holds. Calling that "mismatch" hides the real
-- question ("is anything MISSING?") behind a harmless one, and it trains the
-- admin to ignore the amber box.
--
-- This table stores the honest breakdown per run:
--
--   chunks_parsed               chunks handed to the store
--   chunks_duplicate            same dedup key twice WITHIN this run
--   chunks_inserted             rows actually created (INSERT command tag)
--   chunks_reused               distinct chunks that already existed
--   chunks_preexisting_for_file rows for this expert+file before the store
--   chunks_stored_for_file      rows for this expert+file after the store
--   chunks_general_stored       stored rows whose topic is NULL/''/'general'
--   chunks_fallback             chunks whose topic extraction failed (claim)
--   null_embeddings             stored rows with no vector (invisible to search)
--   corpus_total                the expert's whole corpus after the run
--
-- The rule these numbers enable is an IDENTITY, not a guess:
--
--   chunks_stored_for_file == chunks_preexisting_for_file + chunks_inserted
--
-- Every inserted row carries this source_file, and a row skipped by the conflict
-- clause adds nothing — so the identity can only fail if a row was really lost,
-- or something else wrote to the same expert+file concurrently. That is the
-- signal worth an amber box.
--
-- WHY a table and not just events: the UI needs "how many files are short?" and
-- Phase D needs a row to reconcile against. Both are queries over state, and a
-- timeline is not a state store.
--
-- 'repaired' is accepted from the start so Phase D can mark a reconciled run
-- without a second migration touching this constraint.
--
-- Reversible: down drops the table. Nothing else references it, and the
-- checklist of chunks it describes lives in course_chunks either way.

CREATE TABLE ingestion_runs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id                      UUID NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
    expert_id                   UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    source_file                 VARCHAR(500) NOT NULL,

    chunks_parsed               INTEGER NOT NULL DEFAULT 0,
    chunks_duplicate            INTEGER NOT NULL DEFAULT 0,
    chunks_inserted             INTEGER NOT NULL DEFAULT 0,
    chunks_reused               INTEGER NOT NULL DEFAULT 0,
    chunks_preexisting_for_file INTEGER NOT NULL DEFAULT 0,
    chunks_stored_for_file      INTEGER NOT NULL DEFAULT 0,
    chunks_general_stored       INTEGER NOT NULL DEFAULT 0,
    chunks_fallback             INTEGER NOT NULL DEFAULT 0,
    null_embeddings             INTEGER NOT NULL DEFAULT 0,
    corpus_total                INTEGER NOT NULL DEFAULT 0,

    verification_status         VARCHAR(20) NOT NULL DEFAULT 'not_checked',
    mismatch_reason             TEXT NOT NULL DEFAULT '',

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ingestion_runs_verification_status_check
        CHECK (verification_status IN ('not_checked', 'verified', 'mismatch', 'repaired'))
);

-- One row per file per job. A resumed run re-stores the same file, and upserting
-- one ledger row keeps the latest truth instead of accumulating duplicates.
CREATE UNIQUE INDEX ingestion_runs_job_file_key
    ON ingestion_runs(job_id, source_file);

-- The audit read ("this expert's recent runs, newest first") and the
-- mismatch filter both scan here.
CREATE INDEX idx_ingestion_runs_expert_created
    ON ingestion_runs(expert_id, created_at DESC);
