-- ============================================================
-- Migration 014: Workflow Runner State + Recovery
-- ============================================================
-- Adds runner_state and current_task_cursor to workflows table.
-- Required for pod-restart recovery: startup scans for running
-- workflows and resumes them from last known cursor.
--
-- Also adds UNIQUE constraint on workflow_tasks(workflow_id, assigned_expert_id)
-- required by Projector's ON CONFLICT DO NOTHING.
--
-- NOTE: template_sections (messages) is in migration 013.
--       reply_to_message_id (messages) is in migration 010.
--       This migration adds ONLY workflow-runner-specific columns.
-- ============================================================

-- ============================================================
-- 1. WORKFLOWS — runner recovery columns
-- ============================================================
-- runner_state: JSON snapshot of WorkflowRunner state at last checkpoint.
--   Stores: {phase, completed_expert_ids, failed_expert_ids}
--   Written after every task completes.
--   On resume: runner reads this + replays blackboard from cursor.
--
-- current_task_cursor: blackboard sequence_number of the last event
--   the runner processed. On resume, runner replays from this cursor.
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS runner_state          JSONB,
    ADD COLUMN IF NOT EXISTS current_task_cursor   BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN workflows.runner_state IS
    'WorkflowRunner checkpoint: {phase, completed_expert_ids, failed_expert_ids}. Written after each task.';
COMMENT ON COLUMN workflows.current_task_cursor IS
    'Last blackboard sequence_number processed by WorkflowRunner. Used for pod-restart resume.';

-- ============================================================
-- 2. WORKFLOW_TASKS — unique constraint for Projector
-- ============================================================
-- Projector uses ON CONFLICT (workflow_id, assigned_expert_id) DO NOTHING.
-- Without this constraint, Postgres throws:
--   ERROR: there is no unique constraint matching given keys for referenced table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'workflow_tasks_workflow_expert_unique'
    ) THEN
        ALTER TABLE workflow_tasks
            ADD CONSTRAINT workflow_tasks_workflow_expert_unique
            UNIQUE (workflow_id, assigned_expert_id);
    END IF;
END $$;

-- ============================================================
-- End of migration 014
-- ============================================================
