-- Migration 040 DOWN: remove the repository dependency graph.
--
-- NOTE (honest limitation): edges are derived data. They can be rebuilt exactly
-- by re-running a repository sync, so dropping them loses no information that
-- the provider still holds — but until the next sync completes, related-file
-- and impact queries return nothing rather than stale rows.

DROP INDEX IF EXISTS repo_file_edges_unique;
DROP INDEX IF EXISTS idx_repo_file_edges_dst;
DROP INDEX IF EXISTS idx_repo_file_edges_src;

DROP TABLE IF EXISTS repo_file_edges;
