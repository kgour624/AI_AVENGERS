DROP INDEX IF EXISTS idx_capability_eval_runs_mode;
ALTER TABLE expert_capability_eval_runs DROP COLUMN IF EXISTS layer_preference;
