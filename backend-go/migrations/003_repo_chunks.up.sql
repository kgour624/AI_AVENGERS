-- Migration 003: Repo chunks table for GitHub/GitLab integration
-- WHY separate table from course_chunks:
-- course_chunks are expert-specific (training data)
-- repo_chunks are project-specific (client's codebase)
-- Different scopes, different queries, different lifecycle

CREATE TABLE repo_chunks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    repo_connection_id UUID REFERENCES repo_connections(id) ON DELETE CASCADE,

    -- Content
    chunk_text       TEXT NOT NULL,
    chunk_index      INTEGER NOT NULL,

    -- File metadata
    file_path        VARCHAR(1000) NOT NULL,
    file_language    VARCHAR(50),
    file_size        INTEGER,

    -- Vector embedding (768D)
    embedding        vector(768) NOT NULL,

    -- Navigation
    prev_chunk_id    UUID REFERENCES repo_chunks(id),
    next_chunk_id    UUID REFERENCES repo_chunks(id),

    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_repo_chunks_project ON repo_chunks(project_id);
CREATE INDEX idx_repo_chunks_file ON repo_chunks(project_id, file_path);
CREATE INDEX idx_repo_chunks_embedding ON repo_chunks
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);

-- Full-text search on repo chunks
ALTER TABLE repo_chunks
    ADD COLUMN IF NOT EXISTS chunk_text_tsv tsvector
        GENERATED ALWAYS AS (to_tsvector('english', chunk_text)) STORED;

CREATE INDEX IF NOT EXISTS idx_repo_chunks_fts
    ON repo_chunks USING GIN (chunk_text_tsv);
