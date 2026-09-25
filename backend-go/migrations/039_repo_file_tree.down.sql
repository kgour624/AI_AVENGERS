-- Migration 039 DOWN: remove the repo file tree index.
--
-- NOTE (honest limitation): this drops repo_blobs content wholesale. The bytes
-- can be re-fetched from the provider by running a sync again, so nothing
-- irreplaceable is lost — but no backfill happens inside this migration, so a
-- rollback followed by a re-upgrade leaves repo_files empty until the next sync.

ALTER TABLE repo_connections
    DROP COLUMN IF EXISTS last_commit_sha;

DROP INDEX IF EXISTS idx_repo_files_project_path;
DROP INDEX IF EXISTS idx_repo_files_connection;
DROP INDEX IF EXISTS repo_files_connection_commit_path_key;

DROP TABLE IF EXISTS repo_files;
DROP TABLE IF EXISTS repo_blobs;
