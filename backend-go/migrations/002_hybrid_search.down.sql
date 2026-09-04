-- Rollback migration 002
DROP INDEX IF EXISTS idx_chunks_fts;
DROP INDEX IF EXISTS idx_l2_fts;
ALTER TABLE course_chunks DROP COLUMN IF EXISTS chunk_text_tsv;
ALTER TABLE project_memory_l2 DROP COLUMN IF EXISTS content_tsv;
