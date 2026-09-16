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
