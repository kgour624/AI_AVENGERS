-- Reverse migration 008
ALTER TABLE course_chunks
    DROP CONSTRAINT IF EXISTS course_chunks_expert_hash_unique;

-- Restore the original non-unique index from migration 006
CREATE INDEX IF NOT EXISTS idx_chunks_expert_hash
    ON course_chunks(expert_id, chunk_hash)
    WHERE chunk_hash IS NOT NULL;
