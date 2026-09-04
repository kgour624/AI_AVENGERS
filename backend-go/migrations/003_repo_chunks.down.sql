-- Rollback migration 003
DROP INDEX IF EXISTS idx_repo_chunks_fts;
DROP INDEX IF EXISTS idx_repo_chunks_embedding;
DROP INDEX IF EXISTS idx_repo_chunks_file;
DROP INDEX IF EXISTS idx_repo_chunks_project;
DROP TABLE IF EXISTS repo_chunks CASCADE;
