-- Migration 028: expert versioning + capability drift (C2)
--
-- WHY: an expert's charter/corpus could be replaced by a re-ingest with no
-- record of what changed, so behaviour drift was invisible and
-- irreversible. P8 (§3.1): version the corpus/charter (and the model that
-- produced the embeddings) so a change is observable, and a corpus/model/
-- charter change can be flagged for re-eval (C3) before it is promoted.
--
-- expert_versions: immutable snapshot per ingest/regen. Stores hashes +
-- the declared capability snapshot + the charter text, so an admin can
-- pin a version and roll the charter back without re-ingesting.
-- is_active: the canonical version (one per expert).
--
-- expert_drift_events: append-only record of a detected change between
-- two versions (corpus / capability / charter / model), with a score.
--
-- Additive only. Rollback = drop both tables (down migration).

CREATE TABLE IF NOT EXISTS expert_versions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id             UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    version_number        INTEGER NOT NULL,
    source                VARCHAR(30) NOT NULL DEFAULT 'ingest',
    corpus_hash           CHAR(64) NOT NULL DEFAULT '',
    charter_hash          CHAR(64) NOT NULL DEFAULT '',
    chunk_count           INTEGER NOT NULL DEFAULT 0,
    topic_count           INTEGER NOT NULL DEFAULT 0,
    -- topics as JSONB (not TEXT[]) to avoid array-codec coupling and to
    -- match this codebase's existing JSONB convention for slice data.
    topics                JSONB NOT NULL DEFAULT '[]'::jsonb,
    reasoning_charter     TEXT NOT NULL DEFAULT '',
    clarification_charter JSONB NOT NULL DEFAULT '{}'::jsonb,
    capability_summary    JSONB NOT NULL DEFAULT '{}'::jsonb,
    model                 VARCHAR(100),
    is_active             BOOLEAN NOT NULL DEFAULT FALSE,
    notes                 TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT expert_versions_source_check CHECK (
        source IN ('ingest', 'manual', 'charter_regen')
    ),
    CONSTRAINT expert_versions_unique UNIQUE (expert_id, version_number)
);

-- One active (canonical) version per expert.
CREATE UNIQUE INDEX IF NOT EXISTS idx_expert_versions_active
    ON expert_versions(expert_id) WHERE is_active;

CREATE INDEX IF NOT EXISTS idx_expert_versions_expert
    ON expert_versions(expert_id, version_number DESC);

CREATE TABLE IF NOT EXISTS expert_drift_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id       UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    from_version_id UUID REFERENCES expert_versions(id) ON DELETE SET NULL,
    to_version_id   UUID REFERENCES expert_versions(id) ON DELETE SET NULL,
    drift_type      VARCHAR(30) NOT NULL,
    drift_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    details         JSONB NOT NULL DEFAULT '{}'::jsonb,
    acknowledged    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT expert_drift_type_check CHECK (
        drift_type IN ('corpus', 'capability', 'charter', 'model')
    )
);

CREATE INDEX IF NOT EXISTS idx_expert_drift_expert
    ON expert_drift_events(expert_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_expert_drift_unacked
    ON expert_drift_events(expert_id) WHERE acknowledged = FALSE;

COMMENT ON TABLE expert_versions IS
    'C2: immutable expert corpus/charter/capability snapshots. is_active = pinned canonical version.';
COMMENT ON TABLE expert_drift_events IS
    'C2: append-only drift detections between expert versions (corpus/capability/charter/model).';
