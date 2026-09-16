ALTER TABLE workflows
    DROP COLUMN IF EXISTS runner_state,
    DROP COLUMN IF EXISTS current_task_cursor;

ALTER TABLE workflow_tasks
    DROP CONSTRAINT IF EXISTS workflow_tasks_workflow_expert_unique;
