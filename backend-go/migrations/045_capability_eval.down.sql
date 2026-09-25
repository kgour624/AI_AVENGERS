-- Migration 045 DOWN: remove measured capability.
--
-- NOTE (honest limitation): this drops the question set AND every evaluation run.
-- The ground truth (which chunk answers which question) is generated, not derived,
-- so it cannot be reconstructed from the corpus — a re-run produces a similar but
-- not identical set, and comparisons to older runs are lost. The corpus itself is
-- untouched, and measured_level simply reverts to NULL (the declared coverage band
-- stays in depth_level, exactly as before this migration).

DROP INDEX IF EXISTS idx_capability_results_run;
DROP TABLE IF EXISTS expert_capability_results;

DROP INDEX IF EXISTS idx_capability_eval_runs_expert;
DROP TABLE IF EXISTS expert_capability_eval_runs;

DROP INDEX IF EXISTS idx_capability_cases_expert_topic;
DROP INDEX IF EXISTS expert_capability_cases_question_key;
DROP TABLE IF EXISTS expert_capability_cases;

ALTER TABLE expert_capabilities
    DROP CONSTRAINT IF EXISTS expert_capabilities_measured_level_check;

ALTER TABLE expert_capabilities
    DROP COLUMN IF EXISTS measured_level,
    DROP COLUMN IF EXISTS eval_cases,
    DROP COLUMN IF EXISTS eval_passed,
    DROP COLUMN IF EXISTS last_evaluated_at;

COMMENT ON COLUMN expert_capabilities.depth_level IS NULL;
