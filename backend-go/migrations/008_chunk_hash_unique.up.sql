-- ============================================================
-- Migration 008: Fix chunk_hash uniqueness for ON CONFLICT support
-- ============================================================
-- Problem: Migration 006 created a regular (non-unique) index on
-- (expert_id, chunk_hash). Postgres requires a UNIQUE index or
-- constraint for ON CONFLICT (expert_id, chunk_hash) DO NOTHING.
-- Error: SQLSTATE 42P10 "no unique or exclusion constraint matching
-- the ON CONFLICT specification"
--
-- Fix: Drop the non-unique index, add a UNIQUE constraint.
-- ============================================================

-- Step 1: Drop the non-unique index from migration 006
DROP INDEX IF EXISTS idx_chunks_expert_hash;

-- Step 2: Add UNIQUE constraint (Postgres auto-creates a unique index)
-- Partial: WHERE chunk_hash IS NOT NULL
--   Existing rows with chunk_hash=NULL are excluded.
--   NULL != NULL in Postgres, so NULLs would never conflict anyway,
--   but the partial condition keeps behavior consistent with the old index.
ALTER TABLE course_chunks
    ADD CONSTRAINT course_chunks_expert_hash_unique
    UNIQUE (expert_id, chunk_hash);

-- Note: The UNIQUE constraint only applies to non-NULL chunk_hash values
-- because NULL is not equal to NULL in SQL. Rows with chunk_hash=NULL
-- can coexist without violating uniqueness.

COMMENT ON CONSTRAINT course_chunks_expert_hash_unique ON course_chunks IS
    'Prevents duplicate chunks for the same expert. Required for ON CONFLICT (expert_id, chunk_hash) DO NOTHING in ingestion pipeline.';
