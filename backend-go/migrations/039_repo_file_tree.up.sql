-- Migration 039: repository file tree + content-addressed blob store (Phase 3A)
--
-- WHY this migration exists:
--
-- Until now a synced repository was only stored as RAG text chunks
-- (repo_chunks). There was no file tree, no per-file content, and no way to
-- answer "show me this repo's structure" or "open this file" — the two things
-- an expert needs before it can reason about a client's codebase. This adds
-- that foundation. It also underpins the later phases: the dependency graph
-- (3B), incremental re-index (3C), the human-approved working set (3D) and
-- read-only enforcement (3E) all key off file paths.
--
-- TWO TABLES, deliberately:
--
--   repo_blobs  — content-addressed by git blob SHA-1. Two paths with the same
--                 bytes, or the same path across two commits, share one row.
--                 This is what stops a re-sync (or a later incremental sync)
--                 from re-storing identical file bodies.
--
--   repo_files  — the tree: one row per (connection, commit, path). blob_sha is
--                 NULL for files whose content was intentionally not stored
--                 (binary, unsupported extension, or over the 100KB cap), so
--                 the tree can still be displayed in full without pretending
--                 the content is available. That is why content availability is
--                 read as `blob_sha IS NOT NULL` and there is no separate
--                 boolean column: it would be derivable state.
--
-- Trade-off recorded honestly: repo_blobs is never garbage-collected here. A
-- sync replaces repo_files for the connection, leaving blobs from older
-- commits behind. Content is bounded (100KB/file) and addressed by hash, so
-- the growth is proportional to how much the repository actually changes;
-- collection is deferred to the incremental phase (3C) rather than guessed at
-- now.
--
-- Reversible: down drops both tables and the pin column. No existing column or
-- row is modified, so downgrading loses only the new index.

-- ---------------------------------------------------------------
-- Content-addressed blob store. sha is the git blob SHA-1, computed
-- locally for both providers because GitLab's tree API does not
-- report it while GitHub's does — one key shape for both.
-- ---------------------------------------------------------------
CREATE TABLE repo_blobs (
    sha        VARCHAR(64) PRIMARY KEY,
    size_bytes INTEGER NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------
-- The file tree at a pinned commit.
-- ---------------------------------------------------------------
CREATE TABLE repo_files (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id         UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    repo_connection_id UUID NOT NULL REFERENCES repo_connections(id) ON DELETE CASCADE,
    commit_sha         VARCHAR(64) NOT NULL,
    path               VARCHAR(1000) NOT NULL,

    -- NULL = the tree knows this file exists but its content was not stored.
    blob_sha           VARCHAR(64) REFERENCES repo_blobs(sha) ON DELETE SET NULL,
    language           VARCHAR(50),
    size_bytes         INTEGER,

    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One row per file per commit. Also the ON CONFLICT target used by the sync
-- writer, so it must be a unique constraint/index rather than a plain index.
CREATE UNIQUE INDEX repo_files_connection_commit_path_key
    ON repo_files(repo_connection_id, commit_sha, path);

CREATE INDEX idx_repo_files_connection ON repo_files(repo_connection_id);
CREATE INDEX idx_repo_files_project_path ON repo_files(project_id, path);

-- The commit the stored tree was taken from, so a reader can say which
-- revision it is looking at and a later phase can detect drift.
ALTER TABLE repo_connections
    ADD COLUMN IF NOT EXISTS last_commit_sha VARCHAR(64);
