-- Migration 047 DOWN: remove the content-kind column.
--
-- NOTE (honest limitation): this drops every layer classification. It is derived from
-- the corpus, so it CAN be recomputed — but not identically, since classification is a
-- model judgement. The chunks, embeddings, topics and evaluation evidence are untouched,
-- and depth_level (chunk coverage) still exists, so no factual data is lost.

DROP INDEX IF EXISTS idx_chunks_expert_layer;

ALTER TABLE course_chunks
    DROP CONSTRAINT IF EXISTS course_chunks_layer_check;

ALTER TABLE course_chunks
    DROP COLUMN IF EXISTS layer;
