-- Migration 040: repository dependency graph (Phase 3B)
--
-- WHY this migration exists:
--
-- Phase 3A stored what files a repository HAS. That answers "show me the tree"
-- but not "which files does this requirement actually touch". In a real
-- codebase a file is meaningless in isolation: it imports, requires and
-- includes others, and the set of files an expert needs is a connected region
-- of that graph, not a flat list. Storing the edges is what turns "read the
-- whole repo" (expensive, and the user explicitly does not want it) into
-- "walk outwards from the relevant files, one hop at a time, with approval".
--
-- ONE DIRECTED EDGE TABLE, deliberately:
--
--   repo_file_edges — src_path -> dst_path. Direction matters: "what does this
--   file depend on" (outgoing) and "what would break if this file changed"
--   (incoming) are different questions, and the incoming one is the whole basis
--   of impact analysis. A single undirected table would erase that distinction.
--
-- Edges are FILE-level, because that is the granularity the working set is kept
-- at. Go imports resolve to a package directory rather than a file, so a Go
-- import becomes edges to every file directly in that package. That
-- over-approximates (it can include files the importer never names) and the
-- limitation is recorded here rather than hidden: symbol-level precision needs
-- a type-checked parse and is deliberately out of scope at this layer.
--
-- LIKE 3A: rows are written per (connection, commit, src, dst). A sync replaces
-- the graph for the connection, so a deleted import disappears from the index.
-- No MUTATION of existing tables, so this is reversible by dropping the table.
--
-- Reversible: down drops the table and its indexes.

CREATE TABLE repo_file_edges (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id         UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    repo_connection_id UUID NOT NULL REFERENCES repo_connections(id) ON DELETE CASCADE,
    commit_sha         VARCHAR(64) NOT NULL,

    src_path           VARCHAR(1000) NOT NULL,
    dst_path           VARCHAR(1000) NOT NULL,

    -- How the edge was discovered. Kept so a future extractor can be added
    -- (or an extractor's edges invalidated) without guessing what wrote a row.
    kind               VARCHAR(20) NOT NULL,

    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Outgoing lookups: "what does this file depend on".
CREATE INDEX idx_repo_file_edges_src
    ON repo_file_edges(repo_connection_id, src_path);

-- Incoming lookups: "what depends on this file" (impact analysis).
CREATE INDEX idx_repo_file_edges_dst
    ON repo_file_edges(repo_connection_id, dst_path);

-- One row per distinct edge. Also the ON CONFLICT target used by the writer,
-- so it must be a unique index rather than a plain one.
CREATE UNIQUE INDEX repo_file_edges_unique
    ON repo_file_edges(repo_connection_id, commit_sha, src_path, dst_path, kind);
