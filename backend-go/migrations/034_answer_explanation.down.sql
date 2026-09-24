-- Migration 034 down: drop the C9 explanation columns (views are derived,
-- only these three added facts are removed).
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_coverage_check;
ALTER TABLE messages DROP COLUMN IF EXISTS refusal_reason;
ALTER TABLE messages DROP COLUMN IF EXISTS coverage;
ALTER TABLE messages DROP COLUMN IF EXISTS quality_score;
