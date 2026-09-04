-- Migration 002: Add full-text search support for hybrid retrieval
-- WHY (Byte by Byte AI course):
-- Course taught: keyword-based indexing catches exact matches that
-- semantic search misses (e.g., "PostgreSQL" as exact term).
-- Hybrid search = vector similarity + keyword match = better recall.

-- Add tsvector column to course_chunks for full-text search
ALTER TABLE course_chunks
    ADD COLUMN IF NOT EXISTS chunk_text_tsv tsvector
        GENERATED ALWAYS AS (to_tsvector('english', chunk_text)) STORED;

-- GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS idx_chunks_fts
    ON course_chunks USING GIN (chunk_text_tsv);

-- Add tsvector to project_memory_l2 for keyword search in memory
ALTER TABLE project_memory_l2
    ADD COLUMN IF NOT EXISTS content_tsv tsvector
        GENERATED ALWAYS AS (to_tsvector('english', content)) STORED;

CREATE INDEX IF NOT EXISTS idx_l2_fts
    ON project_memory_l2 USING GIN (content_tsv);
