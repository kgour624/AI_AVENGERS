-- Migration 042 DOWN: remove the stored delivery records.
--
-- NOTE (honest limitation): the patch bytes live here, so dropping this table
-- discards any delivery the client has not downloaded yet. The workspace it was
-- generated from is untouched, so a regenerated delivery produces the same
-- diff — but only while that workspace still exists, and workspaces are not
-- guaranteed to survive a restart. Download before downgrading.

DROP INDEX IF EXISTS codebase_deliveries_workflow_key;

DROP TABLE IF EXISTS codebase_deliveries;
