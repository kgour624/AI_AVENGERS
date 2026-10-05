-- Migration 081: Parent-Child Retrieval (Phase 1 - Additive, Feature-Flagged)
-- WHY: Problem C (Context Orphan) + Problem A (Fixed-Size Chunking)
--   Single 450-tok chunk is both too small for full context and too large for precision.
--   Children 120-180 tok are embedded/searched/reranked (precise), parents 1200-1500 tok
--   are fetched for LLM (full context). Fixes Cross-Encoder 512 truncation trap:
--   50*180^2 vs 5*1500^2 = ~7x cheaper + no truncation.
-- SAFETY: Additive only. No DROP/RENAME. Old rows parent_id=NULL stay searchable.
--   enable_parent_child=false keeps old behaviour. Rollback = flag false + DROP new objects.

CREATE TABLE IF NOT EXISTS expert_pages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id    UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    page_text    TEXT NOT NULL,
    page_index   INTEGER NOT NULL,
    section_path TEXT NOT NULL DEFAULT '',
    token_count  INTEGER NOT NULL DEFAULT 0,
    source_file  VARCHAR(500),
    embedding    vector(768),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (expert_id, page_index, source_file)
);

CREATE INDEX IF NOT EXISTS idx_expert_pages_expert ON expert_pages(expert_id);
CREATE INDEX IF NOT EXISTS idx_expert_pages_source ON expert_pages(expert_id, source_file);
CREATE INDEX IF NOT EXISTS idx_expert_pages_embedding ON expert_pages
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

COMMENT ON TABLE expert_pages IS 'Parent chunks 1200-1500 tok. Children in course_chunks point here via parent_id. Phase 1.';

ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES expert_pages(id) ON DELETE SET NULL;
ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS is_child BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS parent_index INTEGER;

CREATE INDEX IF NOT EXISTS idx_chunks_parent_id ON course_chunks(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_chunks_expert_parent ON course_chunks(expert_id, parent_id);

INSERT INTO system_settings (key, value, description) VALUES
(
    'retrieval_config',
    '{"enable_parent_child": false, "child_soft_limit": 150, "child_hard_limit": 220, "parent_soft_limit": 1200, "parent_hard_limit": 1500, "overlap_tokens": 20, "top_k_children": 50, "top_k_parents": 3, "rerank_top_n": 10, "max_hops": 3, "rrf_k": 60, "atomic_code_fence": true}',
    'Phase 1-3 RAG config. enable_parent_child=false keeps old behaviour. Flip true after re-chunk backfill.'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO system_settings (key, value, description) VALUES
(
    'chunker_config',
    '{"soft_limit": 450, "hard_limit": 600, "overlap_tokens": 75, "min_size": 375, "atomic_code_fence": true, "word_count_estimate": true}',
    'Chunker A: Soft/HardLimit + atomic fence + word-count estimate.'
) ON CONFLICT (key) DO NOTHING;
