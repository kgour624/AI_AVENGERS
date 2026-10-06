-- Migration 082: Parent-Child compat view + index assurance (additive, flag-gated)
-- WHY: 081 created TABLE expert_pages (Option B). Some design docs proposed VIEW expert_pages
--      AS SELECT DISTINCT parent_id FROM course_chunks WHERE is_child=true (Option A).
--      Keep TABLE as source-of-truth (richer: page_text, embedding, token_count), but also
--      provide a lightweight VIEW for code that only needs distinct parent ids, and ensure
--      indexes exist for both paths. Additive only — no DROP/RENAME, zero regression when
--      enable_parent_child=false.
-- SAFETY: CREATE VIEW IF NOT EXISTS + CREATE INDEX IF NOT EXISTS. Old rows parent_id=NULL stay searchable.

-- Ensure hot-path index exists (081 already created, but idempotent for fresh DBs)
CREATE INDEX IF NOT EXISTS idx_chunks_parent_id ON course_chunks(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_chunks_expert_parent ON course_chunks(expert_id, parent_id);

-- Compat view: distinct parents referenced by children (Option A shape, without dropping TABLE)
-- Named expert_pages_view to avoid colliding with TABLE expert_pages. App code can SELECT
-- FROM expert_pages (TABLE, full context) or FROM expert_pages_view (light distinct ids).
CREATE OR REPLACE VIEW expert_pages_view AS
SELECT DISTINCT
    cc.parent_id AS id,
    cc.expert_id,
    cc.parent_index,
    ep.page_text,
    ep.section_path,
    ep.token_count,
    ep.source_file,
    ep.embedding,
    ep.created_at
FROM course_chunks cc
LEFT JOIN expert_pages ep ON ep.id = cc.parent_id
WHERE cc.parent_id IS NOT NULL
  AND cc.is_child = TRUE;

COMMENT ON VIEW expert_pages_view IS 'Compat VIEW: distinct parent_ids from course_chunks (Option A shape) joined to expert_pages TABLE (Option B source). Additive, read-only.';

-- Extra index to speed VIEW join + retrieval fetch (expert_pages lookup by id is PK already; this helps filtering)
CREATE INDEX IF NOT EXISTS idx_expert_pages_expert_page ON expert_pages(expert_id, page_index);
