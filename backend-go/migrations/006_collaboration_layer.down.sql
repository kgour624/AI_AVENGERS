-- ============================================================
-- Migration 006 DOWN: Reverse the collaboration layer
-- ============================================================
-- Ordered opposite of the up migration:
--   1. Delete inserted system_settings rows
--   2. Drop new columns from existing tables
--   3. Drop new tables (children before parents)
--   4. Drop new indexes and constraints implicitly with columns/tables
--
-- Note: existing is_training column on experts is NOT touched —
-- it was there before this migration and remains after rollback.
-- ============================================================

-- 9. Remove system_settings entries introduced by this migration
DELETE FROM system_settings WHERE key IN ('workflow_engine', 'blackboard');

-- 8. Drop workflow_id from messages
DROP INDEX IF EXISTS idx_messages_workflow;
ALTER TABLE messages
    DROP COLUMN IF EXISTS workflow_id;

-- 7. approval_requests
DROP TABLE IF EXISTS approval_requests;

-- 6. workflow_checkpoints
DROP TABLE IF EXISTS workflow_checkpoints;

-- 5. workflow_tasks (references blackboard_events, so drop before it)
DROP TABLE IF EXISTS workflow_tasks;

-- 4. blackboard_events (references workflows, so drop before it)
DROP TABLE IF EXISTS blackboard_events;

-- 3. workflows
DROP TABLE IF EXISTS workflows;

-- 2. course_chunks: drop chunk_hash column + its index
DROP INDEX IF EXISTS idx_chunks_expert_hash;
ALTER TABLE course_chunks
    DROP COLUMN IF EXISTS chunk_hash;

-- 1. experts: drop the collaboration/loop/tool config columns.
-- CHECK constraints are dropped automatically with the columns.
ALTER TABLE experts
    DROP COLUMN IF EXISTS training_status,
    DROP COLUMN IF EXISTS allowed_tools,
    DROP COLUMN IF EXISTS max_loop_iterations,
    DROP COLUMN IF EXISTS loop_pattern,
    DROP COLUMN IF EXISTS top_p,
    DROP COLUMN IF EXISTS temperature,
    DROP COLUMN IF EXISTS model_tier;

-- ============================================================
-- End of migration 006 down
-- ============================================================
