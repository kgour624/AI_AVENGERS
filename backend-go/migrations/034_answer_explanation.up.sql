-- Migration 034: persist explanation facts on answers (C9)
--
-- WHY: the "why this answer" surface is a VIEW over stored facts, never a new
-- LLM narration. Provenance (C1) already stores the signed chain (citations,
-- B8 claim→evidence reports, model, decision mode, gate stopped). Three facts
-- the surface needs were computed but never persisted (B6 judge score, China
-- Wall coverage verdict, and the Gate-5 refusal reason), so they are added here
-- — the minimum needed so the view is assembled, not invented.
--
-- Additive only. Rollback = migration 034 down.

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS quality_score  DECIMAL(5,4),
    ADD COLUMN IF NOT EXISTS coverage       VARCHAR(10),
    ADD COLUMN IF NOT EXISTS refusal_reason TEXT;

ALTER TABLE messages
    DROP CONSTRAINT IF EXISTS messages_coverage_check;
ALTER TABLE messages
    ADD CONSTRAINT messages_coverage_check CHECK (
        coverage IS NULL OR coverage IN ('YES', 'PARTIAL', 'NO')
    );

COMMENT ON COLUMN messages.quality_score IS
    'C9: B6 LLM-as-judge overall score [0,1]. NULL when the judge was off/failed-open.';
COMMENT ON COLUMN messages.coverage IS
    'C9: China Wall coverage verdict YES|PARTIAL|NO. NULL for non-generated modes.';
COMMENT ON COLUMN messages.refusal_reason IS
    'C9: Gate-5 / partial China Wall refusal reason. NULL unless a refusal was explained.';
