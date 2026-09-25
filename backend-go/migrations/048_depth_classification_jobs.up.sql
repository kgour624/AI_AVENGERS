-- Migration 048: durable background jobs for depth-layer classification.
--
-- WHY: the classification POST previously ran all model calls inside the HTTP
-- request. A real pass can take longer than the browser/proxy's 30-second request
-- limit; when that request context was cancelled, the LLM calls were cancelled too.
-- The job row now outlives the request, records progress after every batch, and is
-- polled by the UI. A failed worker leaves its reason here instead of disappearing.
--
-- One active job per expert. A partial unique index makes the guarantee hold even if
-- two POST requests arrive at the same time or two API processes receive them.

CREATE TABLE expert_depth_classification_jobs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id      UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    status         VARCHAR(20) NOT NULL DEFAULT 'queued',

    -- The job's snapshot/progress. Updated after each model batch, not once at the end.
    total          INTEGER NOT NULL DEFAULT 0,
    classified     INTEGER NOT NULL DEFAULT 0,
    remaining      INTEGER NOT NULL DEFAULT 0,
    calls          INTEGER NOT NULL DEFAULT 0,
    rejected       INTEGER NOT NULL DEFAULT 0,

    error_message  TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at   TIMESTAMPTZ,

    CONSTRAINT expert_depth_classification_jobs_status_check
        CHECK (status IN ('queued', 'running', 'complete', 'failed')),
    CONSTRAINT expert_depth_classification_jobs_counts_check
        CHECK (total >= 0 AND classified >= 0 AND remaining >= 0 AND calls >= 0 AND rejected >= 0)
);

CREATE UNIQUE INDEX expert_depth_classification_jobs_one_active
    ON expert_depth_classification_jobs(expert_id)
    WHERE status IN ('queued', 'running');

CREATE INDEX idx_depth_classification_jobs_latest
    ON expert_depth_classification_jobs(expert_id, created_at DESC);

COMMENT ON TABLE expert_depth_classification_jobs IS
    'Durable, pollable status for the on-demand depth-layer classifier. The model work '
    'runs outside the HTTP request so browser/proxy timeouts do not cancel it.';
