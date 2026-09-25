-- Migration 046 DOWN: remove concept relationships and the retrieval-mode marker.
--
-- NOTE (honest limitation): this drops the extracted relationships. They are generated,
-- not derived, so a re-extraction produces a similar but not identical graph — and the
-- per-run graph_expansion marker disappears with the column, so older evaluation runs
-- can no longer be told apart by retrieval mode. The corpus, the topics and the
-- evaluation evidence are all untouched.

DROP INDEX IF EXISTS idx_concept_edges_to;
DROP INDEX IF EXISTS idx_concept_edges_from;
DROP INDEX IF EXISTS expert_concept_edges_key;
DROP TABLE IF EXISTS expert_concept_edges;

ALTER TABLE expert_capability_eval_runs
    DROP COLUMN IF EXISTS graph_expansion;
