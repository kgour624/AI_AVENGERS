-- Migration 033 down: revert bring-your-own-expert (C8).
-- Additive-only up → safe to drop. Experts keep their tenant_id (mig 030);
-- only the BYO provenance/audit surface is removed.
DROP TABLE IF EXISTS byo_expert_events;

ALTER TABLE experts DROP CONSTRAINT IF EXISTS experts_origin_check;
DROP INDEX IF EXISTS idx_experts_origin;
ALTER TABLE experts DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE experts DROP COLUMN IF EXISTS origin;
