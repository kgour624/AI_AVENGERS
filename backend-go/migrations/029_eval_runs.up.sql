-- Migration 029: evaluation runs + baseline (C3)
--
-- WHY: there was no automated quality gate — every prompt/model/corpus
-- change was traced manually. P7 (§3.1): run a golden set on every
-- PR/model/prompt change, store baseline scores, and block on VITAL
-- failures (tiered must-have vs good-to-have). This table stores one row
-- per run with per-case results so a quality/cost delta is computable.
--
-- is_baseline: the reference run for a suite. A partial unique index
-- keeps a single baseline per suite; the harness promotes a run to
-- baseline only deliberately (never automatically on every CI run).
--
-- Additive only. Rollback = drop the table (down migration).

CREATE TABLE IF NOT EXISTS eval_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    suite         VARCHAR(50) NOT NULL,
    model         VARCHAR(100),
    total         INTEGER NOT NULL DEFAULT 0,
    passed        INTEGER NOT NULL DEFAULT 0,
    vital_total   INTEGER NOT NULL DEFAULT 0,
    vital_failed  INTEGER NOT NULL DEFAULT 0,
    score         DOUBLE PRECISION NOT NULL DEFAULT 0,
    cost_usd      DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_baseline   BOOLEAN NOT NULL DEFAULT FALSE,
    results       JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_eval_runs_suite
    ON eval_runs(suite, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_eval_runs_baseline
    ON eval_runs(suite) WHERE is_baseline;

COMMENT ON TABLE eval_runs IS
    'C3: golden-set evaluation runs. score = passed/total. is_baseline = reference for delta/regression.';
COMMENT ON COLUMN eval_runs.results IS
    'Per-case results: [{case_id, vital, passed, checks[], error, cost_usd, latency_ms}].';
