-- One row per (workflow, phase, expert): this IS the idempotency key for a unit
-- of work, and the only thing allowed to decide that a unit is finished.
--
-- WHY it exists: the runner used to decide "already done" from a list inside
-- workflows.runner_state — a mutable JSON blob that is a hint, not evidence. If
-- that list was stale (a crash between an expert finishing and the checkpoint
-- being written, or a workflow resumed from an older phase), the experts were
-- skipped, every phase reported success, and the workflow reached 'completed'
-- having produced nothing and spent nothing. Observed exactly that: a workflow
-- went from intake to done with an empty board and zero LLM calls.
--
-- Now a unit is skipped only when this table proves that the SAME
-- (workflow, phase, expert) already succeeded. A second attempt updates the
-- existing row and bumps attempt — it can never create a second success, so a
-- resume cannot double-produce a design section the way it could before.
--
-- attempt is bounded by the code (maxTaskAttempts): a task that fails
-- deterministically stops retrying instead of burning credits on every resume,
-- and fails with its reason visible on the workflow rather than looping.
CREATE TABLE IF NOT EXISTS workflow_task_attempts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id       UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    phase             TEXT NOT NULL,
    expert_id         UUID NOT NULL REFERENCES experts(id),
    attempt           INTEGER NOT NULL DEFAULT 1 CHECK (attempt >= 1),
    status            TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
    artifact_event_id UUID REFERENCES blackboard_events(id),
    error             TEXT NOT NULL DEFAULT '',
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at       TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workflow_task_attempts_key UNIQUE (workflow_id, phase, expert_id)
);

-- The runner's read is always "what is the state of this workflow's units".
CREATE INDEX IF NOT EXISTS idx_wta_workflow_status ON workflow_task_attempts(workflow_id, status);
