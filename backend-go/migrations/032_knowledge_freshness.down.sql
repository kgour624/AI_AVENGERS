-- Rollback 032: remove C6 knowledge freshness (additive migration).
DROP TABLE IF EXISTS knowledge_refresh_tasks;
DROP INDEX IF EXISTS idx_chunks_expert_created;
DROP INDEX IF EXISTS idx_chunks_expert_model;
ALTER TABLE course_chunks DROP COLUMN IF EXISTS embedding_model;
ALTER TABLE course_chunks DROP COLUMN IF EXISTS embedding_provider;
