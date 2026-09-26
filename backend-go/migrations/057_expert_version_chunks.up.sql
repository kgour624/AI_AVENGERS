-- T4: full corpus rollback for an expert version.
--
-- WHY: migration 028 snapshots only hashes + charter, so a version could be
-- pinned to roll the CHARTER back, but the corpus itself could not be restored
-- (expertversion.go documented this as deferred: "full corpus rollback needs the
-- original transcripts"). Re-ingesting changed answers forever, with no way back.
--
-- expert_version_chunks stores the actual chunk rows (including the embedding,
-- topic and source file) at snapshot time, so a version can be restored exactly
-- without the original transcripts. It is additive and fully reversible: the
-- down migration drops the table and leaves expert_versions untouched.

CREATE TABLE IF NOT EXISTS expert_version_chunks (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id         UUID NOT NULL REFERENCES expert_versions(id) ON DELETE CASCADE,
    expert_id          UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    chunk_index        INTEGER NOT NULL,
    chunk_text         TEXT NOT NULL,
    topic              VARCHAR(255),
    subtopic           VARCHAR(255),
    source_file        VARCHAR(500),
    embedding          vector(768),
    embedding_provider VARCHAR(50),
    embedding_model    VARCHAR(150),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expert_version_chunks_version
    ON expert_version_chunks(version_id, chunk_index);

COMMENT ON TABLE expert_version_chunks IS
    'T4: chunk rows captured with an expert version, so the corpus can be restored on rollback.';
