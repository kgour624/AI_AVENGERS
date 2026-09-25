-- Migration 041 DOWN: remove the existing-codebase environment and its working set.
--
-- NOTE (honest limitation): dropping workflow_codebase_files destroys the
-- record of which files a human approved for which workflow. The file contents
-- themselves live in the repository tables and are untouched, and a repository
-- re-sync rebuilds them — but the APPROVAL decisions are not recoverable, and
-- they are the part that represents human judgement. A downgrade therefore
-- loses audit history; that is stated here rather than discovered later.

DROP INDEX IF EXISTS idx_wcf_workflow_status;
DROP INDEX IF EXISTS workflow_codebase_files_unique;

DROP TABLE IF EXISTS workflow_codebase_files;

ALTER TABLE workflows
    DROP CONSTRAINT IF EXISTS workflows_mode_check;

ALTER TABLE workflows
    DROP COLUMN IF EXISTS mode;
