-- Migration 041: existing-codebase workflows + the human-approved working set (Phase 3D)
--
-- WHY this migration exists:
--
-- Until now every workflow built something from scratch, and every file an
-- expert read came from the training corpus. The existing-codebase workflow is
-- a different environment: the input is the CLIENT's repository, and the set of
-- files an expert may read must be chosen deliberately — suggested by the
-- system, approved by a human — because reading the whole repository is exactly
-- the expensive operation this feature exists to avoid.
--
-- TWO CHANGES:
--
--   1. workflows.mode distinguishes the two environments. It defaults to
--      'scratch' so every existing row keeps its current meaning and no code
--      path changes behaviour until it explicitly opts in.
--
--   2. workflow_codebase_files is the working set: one row per file, in one of
--      three states (pending / approved / rejected).
--
-- WHY ONE TABLE, NOT "suggestions" PLUS "manifest":
--
-- The obvious design is two tables — proposals, and the frozen approved set.
-- That was rejected deliberately: two tables holding the same file list means
-- two copies of the same truth that must be kept in step, and every bug in that
-- sync shows up as an expert reading a file the client never approved — a
-- safety failure, not a cosmetic one. "The manifest" is simply the rows whose
-- status is 'approved', so it is derived from one source instead of duplicated.
--
-- The approval fields are the audit trail: which expert proposed the file, why,
-- who decided, and when. decided_by_client is explicit rather than inferred,
-- because a rule change to how approvals are recorded must not silently rewrite
-- history.
--
-- Reversible: down drops the table and the mode column. Rolling back leaves
-- existing-codebase workflows unreadable as such, which is the correct failure
-- for a downgrade: better to refuse than to silently treat them as scratch.

-- ---------------------------------------------------------------
-- Environment switch. 'scratch' = build something new;
-- 'existing_codebase' = work inside a connected client repository.
-- ---------------------------------------------------------------
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS mode VARCHAR(30) NOT NULL DEFAULT 'scratch'
        CHECK (mode IN ('scratch', 'existing_codebase'));

-- ---------------------------------------------------------------
-- The working set: every file proposed for, or admitted to, one
-- workflow's code context.
-- ---------------------------------------------------------------
CREATE TABLE workflow_codebase_files (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id            UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    project_id             UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    path                   VARCHAR(1000) NOT NULL,

    status                 VARCHAR(20) NOT NULL,

    -- 'expert' = the system proposed it; 'client' = the client added it
    -- directly. Kept separate so the approval history shows who drove the
    -- decision, not just that a file ended up approved.
    source                 VARCHAR(20) NOT NULL,

    suggested_by_expert_id UUID REFERENCES experts(id) ON DELETE SET NULL,
    reason                 TEXT NOT NULL DEFAULT '',
    score                  DOUBLE PRECISION NOT NULL DEFAULT 0,
    hop_depth              INTEGER NOT NULL DEFAULT 0,

    decided_by_client      BOOLEAN NOT NULL DEFAULT FALSE,
    decided_at             TIMESTAMPTZ,

    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT workflow_codebase_files_status_check
        CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT workflow_codebase_files_source_check
        CHECK (source IN ('expert', 'client'))
);

-- One row per file per workflow: approving twice must not create a second
-- entry, and the ON CONFLICT target is this index.
CREATE UNIQUE INDEX workflow_codebase_files_unique
    ON workflow_codebase_files(workflow_id, path);

-- The manifest read ("everything approved for this workflow") and the review
-- queue read ("everything still pending") both scan by status.
CREATE INDEX idx_wcf_workflow_status
    ON workflow_codebase_files(workflow_id, status);
