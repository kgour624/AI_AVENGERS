-- Rollback 081: revert parent-child additive objects (safe, no data loss on original chunks)
DROP INDEX IF EXISTS idx_chunks_expert_parent;
DROP INDEX IF EXISTS idx_chunks_parent_id;
ALTER TABLE course_chunks DROP COLUMN IF EXISTS parent_index;
ALTER TABLE course_chunks DROP COLUMN IF EXISTS is_child;
ALTER TABLE course_chunks DROP COLUMN IF EXISTS parent_id;
DROP INDEX IF EXISTS idx_expert_pages_embedding;
DROP INDEX IF EXISTS idx_expert_pages_source;
DROP INDEX IF EXISTS idx_expert_pages_expert;
DROP TABLE IF EXISTS expert_pages;
DELETE FROM system_settings WHERE key IN ('retrieval_config','chunker_config');
