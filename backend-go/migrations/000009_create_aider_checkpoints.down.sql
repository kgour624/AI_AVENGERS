-- Phase 5: Production Hardening - Error Recovery
-- Drop aider_checkpoints table

DROP INDEX IF EXISTS idx_aider_checkpoints_workflow_id;
DROP INDEX IF EXISTS idx_aider_checkpoints_created_at;
DROP TABLE IF EXISTS aider_checkpoints;
