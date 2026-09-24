-- Migration 036: durable ingestion job event log (training transparency, T1)
--
-- WHY: ingestion progress was a single mutable row on ingestion_jobs
-- (current_stage / stage_detail / processed_chunks) and the "live event log"
-- in the admin modal was built client-side from 1-second SSE snapshots. So:
--   * refreshing the browser erased the whole timeline,
--   * two admins watching the same job saw different logs,
--   * and there was no way to tell what actually happened after the fact —
--     the only durable record was the server's stdout log.
--
-- This table is the append-only source of truth for "what happened, when":
-- one row per meaningful pipeline action (stage started/done, batch done,
-- verification, pause, failure, completion). ingestion_jobs stays as the hot
-- snapshot the UI needs for a single render; this table is the timeline.
--
-- Shape deliberately mirrors blackboard_events (migration 006):
--   * Postgres is the source of truth; Redis is notification-only.
--   * sequence_number is BIGSERIAL so a subscriber cursor is just
--     "give me everything with sequence_number > N".
--   * NEVER UPDATE OR DELETE rows.
--
-- Additive only. Rollback = migration 036 down.

CREATE TABLE IF NOT EXISTS ingestion_job_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
    expert_id       UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    sequence_number BIGSERIAL NOT NULL,
    stage           VARCHAR(40) NOT NULL,
    kind            VARCHAR(40) NOT NULL,
    detail          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Read path: "all events for this job after sequence N" (SSE catch-up +
-- history endpoint). Ordered by the global sequence_number, which is stable
-- and monotonic per job.
CREATE INDEX IF NOT EXISTS idx_ingestion_events_job_seq
    ON ingestion_job_events(job_id, sequence_number);

-- Secondary read path: "latest activity for this expert" (expert list UI).
CREATE INDEX IF NOT EXISTS idx_ingestion_events_expert
    ON ingestion_job_events(expert_id, created_at DESC);

-- Bounded retention: the event log is per-job and a job is deleted with its
-- expert, so no extra pruning job is required at this scale.

COMMENT ON TABLE ingestion_job_events IS
    'T1: append-only ingestion timeline. NEVER UPDATE OR DELETE rows.';
COMMENT ON COLUMN ingestion_job_events.sequence_number IS
    'Global monotonic cursor. Subscribers ask for sequence_number > last_seen.';
COMMENT ON COLUMN ingestion_job_events.kind IS
    'stage_started | stage_done | batch_done | chunk_stored | verified | paused | failed | complete | run_started';
COMMENT ON COLUMN ingestion_job_events.detail IS
    'Kind-specific payload (batch index, counts, durations, verification ledger, ...).';
