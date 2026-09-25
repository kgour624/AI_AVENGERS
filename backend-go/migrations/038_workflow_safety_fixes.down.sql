-- Migration 038 DOWN: restore the pre-038 schema.
--
-- NOTE (honest limitation): the duplicate repo_connections rows that 038
-- collapsed cannot be recreated — the discarded rows' tokens are stale by
-- definition and their chunks were re-pointed at the survivor. Dropping the
-- unique constraint restores the old (buggy) shape so the schema matches the
-- pre-038 code, but the data loss is not reversible.

ALTER TABLE workflows
    DROP COLUMN IF EXISTS failure_reason;

-- current_task_cursor was NOT NULL DEFAULT 0 in migration 014; restore the
-- same shape. No row ever carried a meaningful value (nothing wrote it).
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS current_task_cursor BIGINT NOT NULL DEFAULT 0;

ALTER TABLE repo_connections
    DROP CONSTRAINT IF EXISTS repo_connections_project_id_key;
CREATE INDEX IF NOT EXISTS idx_repo_connections_project
    ON repo_connections(project_id);
