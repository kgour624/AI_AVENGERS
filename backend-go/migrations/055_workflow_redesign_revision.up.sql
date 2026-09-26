-- A client-requested redesign is fresh work even if that expert/phase already
-- succeeded. The initial design uses the all-zero UUID; each approval gate's
-- ID identifies its requested redesign across a server restart.
ALTER TABLE workflow_task_attempts
    ADD COLUMN IF NOT EXISTS revision_id UUID NOT NULL
        DEFAULT '00000000-0000-0000-0000-000000000000';
