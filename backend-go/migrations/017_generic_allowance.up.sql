-- Client-controlled generic-knowledge allowance.
--
-- WHY this column exists:
--   Experts are trained on specific course transcripts. Before this, whether
--   an expert was allowed to fall back on generic LLM knowledge was decided
--   by an LLM "coverage check" at run time, which routinely answered "NO
--   coverage" for a broad task and handed the ENTIRE task to generic — the
--   trained material was fetched and then effectively ignored.
--
--   That decision now belongs to the client, not to a model:
--     0  (default) = trained knowledge + peer experts only. No generic.
--     >0           = at most this percentage of the output may be generic,
--                    and every such claim must be tagged [GENERIC].
--
--   The client raises it from the approval gate only after reading the
--   deliverables, and re-runs the phase with the new value.
--
-- WHY the 30 ceiling is in the schema and not only in the handler:
--   It is a product rule — trained knowledge must stay the majority
--   contributor. A CHECK constraint makes it impossible for any code path,
--   present or future, to exceed it.

ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS generic_allowance_pct DECIMAL(5,2) NOT NULL DEFAULT 0;

-- ADD CONSTRAINT has no IF NOT EXISTS in PostgreSQL, so guard it explicitly
-- to keep this migration safe to re-run (golang-migrate tracks versions, but
-- every other migration here is written to be idempotent).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'workflows_generic_allowance_pct_check'
    ) THEN
        ALTER TABLE workflows
            ADD CONSTRAINT workflows_generic_allowance_pct_check
            CHECK (generic_allowance_pct >= 0 AND generic_allowance_pct <= 30);
    END IF;
END $$;

COMMENT ON COLUMN workflows.generic_allowance_pct IS
    'Client-set ceiling (0-30) on how much of an expert answer may be generic knowledge. 0 = trained + peer knowledge only.';
