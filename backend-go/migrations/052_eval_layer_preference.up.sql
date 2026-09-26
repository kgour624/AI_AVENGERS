-- The capability eval gains a third retrieval mode: the section/layer nudge.
--
-- WHY a column and not a new table: a pass is identified by the retrieval it used,
-- and that is already how graph_expansion works. A pass may carry both flags, and
-- the same-mode baseline lookup matches BOTH — so a pass is only ever compared with
-- one that differs in nothing, which is the whole point of the measurement.
--
-- Without this column the third mode would be indistinguishable from a plain pass
-- in the data, and the comparison would silently attribute the nudge's effect to
-- nothing at all.
ALTER TABLE expert_capability_eval_runs
    ADD COLUMN IF NOT EXISTS layer_preference BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_capability_eval_runs_mode
    ON expert_capability_eval_runs(expert_id, graph_expansion, layer_preference, started_at DESC);
