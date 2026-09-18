-- Phase 5: Production Hardening - Error Recovery
-- Create aider_checkpoints table for resuming after pod restarts

CREATE TABLE IF NOT EXISTS aider_checkpoints (
    workflow_id UUID NOT NULL,
    expert_id UUID NOT NULL,
    task_id UUID NOT NULL,
    current_iteration INT NOT NULL,
    commit_shas JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_observation TEXT,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workflow_id, expert_id, task_id)
);

-- Index for cleanup queries (find old checkpoints)
CREATE INDEX IF NOT EXISTS idx_aider_checkpoints_created_at
    ON aider_checkpoints(created_at);

-- Index for workflow queries (find all checkpoints for a workflow)
CREATE INDEX IF NOT EXISTS idx_aider_checkpoints_workflow_id
    ON aider_checkpoints(workflow_id);

-- Comments for documentation
COMMENT ON TABLE aider_checkpoints IS 'Stores Aider iteration state for recovery after pod restarts';
COMMENT ON COLUMN aider_checkpoints.workflow_id IS 'Workflow this checkpoint belongs to';
COMMENT ON COLUMN aider_checkpoints.expert_id IS 'Expert executing the task';
COMMENT ON COLUMN aider_checkpoints.task_id IS 'Task being executed';
COMMENT ON COLUMN aider_checkpoints.current_iteration IS 'Last completed iteration number';
COMMENT ON COLUMN aider_checkpoints.commit_shas IS 'Array of git commit SHAs produced so far';
COMMENT ON COLUMN aider_checkpoints.last_observation IS 'Observations from last iteration (for debugging)';
COMMENT ON COLUMN aider_checkpoints.completed IS 'True if task completed (should be deleted)';
COMMENT ON COLUMN aider_checkpoints.created_at IS 'When checkpoint was first created';
COMMENT ON COLUMN aider_checkpoints.updated_at IS 'When checkpoint was last updated';
