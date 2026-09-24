-- Migration 032: knowledge freshness / staleness signals (C6)
--
-- WHY: knowledge grows but nothing told you it was outdated or that an
-- embedding-provider change left vectors incompatible. The admin note on
-- the embedding-settings screen already warned "Changing embedding provider
-- requires re-ingesting ALL transcripts" — but there was no record of which
-- model produced each vector, so the warning could not be acted on.
--
-- course_chunks gains embedding_provider/embedding_model (nullable; existing
-- rows stay NULL = unknown and are never flagged as a mismatch). New ingests
-- stamp the model that produced them, so a later provider/model change is
-- detectable. knowledge_refresh_tasks is an append-only-ish work list of
-- refresh actions (dedup + reopen-on-recur per expert+type+key).
--
-- Additive. Rollback = migration 032 down.

ALTER TABLE course_chunks
    ADD COLUMN IF NOT EXISTS embedding_provider VARCHAR(50);
ALTER TABLE course_chunks
    ADD COLUMN IF NOT EXISTS embedding_model VARCHAR(150);

COMMENT ON COLUMN course_chunks.embedding_provider IS
    'C6: embedding provider that produced this vector (sidecar | codecraftapi). NULL = legacy/unknown.';
COMMENT ON COLUMN course_chunks.embedding_model IS
    'C6: embedding model name at ingest time. NULL for sidecar or legacy rows.';

CREATE INDEX IF NOT EXISTS idx_chunks_expert_model
    ON course_chunks(expert_id, embedding_model);
CREATE INDEX IF NOT EXISTS idx_chunks_expert_created
    ON course_chunks(expert_id, created_at DESC);

CREATE TABLE IF NOT EXISTS knowledge_refresh_tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id   UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    task_type   VARCHAR(50) NOT NULL,
    severity    VARCHAR(20) NOT NULL DEFAULT 'medium',
    title       VARCHAR(300) NOT NULL,
    details     JSONB NOT NULL DEFAULT '{}'::jsonb,
    status      VARCHAR(20) NOT NULL DEFAULT 'open',
    dedupe_key  VARCHAR(200) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    CONSTRAINT krt_task_type_check CHECK (task_type IN
        ('embedding_mismatch', 'stale_corpus', 'orphan_reference', 'empty_corpus')),
    CONSTRAINT krt_severity_check CHECK (severity IN ('low', 'medium', 'high')),
    CONSTRAINT krt_status_check   CHECK (status IN ('open', 'acknowledged', 'resolved'))
);

COMMENT ON TABLE knowledge_refresh_tasks IS
    'C6: refresh actions for an expert (stale corpus, embedding mismatch, orphaned citations, empty corpus).';

CREATE UNIQUE INDEX IF NOT EXISTS uq_krt_dedupe
    ON knowledge_refresh_tasks(expert_id, task_type, dedupe_key);
CREATE INDEX IF NOT EXISTS idx_krt_status ON knowledge_refresh_tasks(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_krt_expert ON knowledge_refresh_tasks(expert_id, status);
