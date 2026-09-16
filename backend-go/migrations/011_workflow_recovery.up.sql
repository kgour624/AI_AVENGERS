-- Migration 011: Workflow Runner State + Recovery
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS runner_state          JSONB,
    ADD COLUMN IF NOT EXISTS current_task_cursor   BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN workflows.runner_state IS
    'WorkflowRunner checkpoint: {phase, completed_expert_ids, failed_expert_ids}.';
COMMENT ON COLUMN workflows.current_task_cursor IS
    'Last blackboard sequence_number processed by WorkflowRunner. Used for pod-restart resume.';

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS template_sections     JSONB,
    ADD COLUMN IF NOT EXISTS reply_to_message_id   UUID REFERENCES messages(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_messages_reply_to
    ON messages(reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;

-- Unique constraint required by Projector's ON CONFLICT (workflow_id, assigned_expert_id) DO NOTHING.
-- Without this, INSERT in projector.go fails with: there is no unique constraint matching ON CONFLICT.
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
