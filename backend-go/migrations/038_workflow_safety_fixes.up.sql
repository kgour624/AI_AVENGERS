-- Migration 038: workflow/repo safety fixes (A1, A8, A12)
--
-- WHY this migration exists:
--
-- A1 — repo_connections had NO unique constraint on project_id, but
--      repo/service.go ConnectRepo upserts with `ON CONFLICT (project_id)`.
--      Postgres rejects that upsert with SQLSTATE 42P10 ("there is no unique
--      or exclusion constraint matching the ON CONFLICT specification"), so
--      re-connecting a repo could fail. Trade-off: the code was written for
--      one connection per project, so the schema is what must change.
--      Existing duplicate rows must be collapsed first — and because
--      repo_chunks.repo_connection_id is ON DELETE CASCADE, the surviving
--      row must adopt the doomed rows' chunks BEFORE the delete, or a
--      re-sync would silently lose ingested code.
--
-- A12 — workflows.failure_reason: Engine.Fail() received a reason and only
--       logged it. The admin could never see WHY a workflow failed.
--       workflows.current_task_cursor: added by migration 014 but never read
--       or written by any code (runner.go keeps its own cursor in
--       runner_state). It is misleading dead schema, so it is dropped.
--
-- A8 is additive code (engine posts workflow_completed/workflow_failed to
-- the blackboard) and needs no schema change.
--
-- Reversible: down restores the old index and columns. The duplicate-row
-- collapse is one-way by nature (the discarded tokens are stale anyway).

-- ---------------------------------------------------------------
-- A1, step 1: re-point chunks of duplicate connections at the row
-- that will survive (most recent per project).
-- ---------------------------------------------------------------
WITH ranked AS (
    SELECT id,
           project_id,
           ROW_NUMBER() OVER (
               PARTITION BY project_id
               ORDER BY created_at DESC, id DESC
           ) AS rn,
           FIRST_VALUE(id) OVER (
               PARTITION BY project_id
               ORDER BY created_at DESC, id DESC
           ) AS keep_id
    FROM repo_connections
)
UPDATE repo_chunks rc
SET repo_connection_id = r.keep_id
FROM ranked r
WHERE rc.repo_connection_id = r.id
  AND r.rn > 1;

-- ---------------------------------------------------------------
-- A1, step 2: delete the duplicate connection rows.
-- Chunks already point at the survivor, so the CASCADE cannot
-- destroy data. project_id denormalized fields on the survivor are
-- already the latest (ORDER BY created_at DESC picks the newest).
-- ---------------------------------------------------------------
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY project_id
               ORDER BY created_at DESC, id DESC
           ) AS rn
    FROM repo_connections
)
DELETE FROM repo_connections rc
USING ranked r
WHERE rc.id = r.id
  AND r.rn > 1;

-- ---------------------------------------------------------------
-- A1, step 3: enforce one connection per project. The old plain
-- index becomes redundant once the unique constraint's implicit
-- index exists, so it is dropped.
-- ---------------------------------------------------------------
ALTER TABLE repo_connections
    DROP CONSTRAINT IF EXISTS repo_connections_project_id_key;
ALTER TABLE repo_connections
    ADD CONSTRAINT repo_connections_project_id_key UNIQUE (project_id);
DROP INDEX IF EXISTS idx_repo_connections_project;

-- ---------------------------------------------------------------
-- A12: let a failed workflow record its reason, drop the dead cursor.
-- ---------------------------------------------------------------
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS failure_reason TEXT;
ALTER TABLE workflows
    DROP COLUMN IF EXISTS current_task_cursor;
