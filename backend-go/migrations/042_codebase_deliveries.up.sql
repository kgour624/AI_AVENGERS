-- Migration 042: codebase delivery — patch against the pinned base (Phase 3F)
--
-- WHY this migration exists:
--
-- Phases 3A-3E let a workflow read an approved slice of a client's repository
-- and change it. Nothing yet answers the two questions the client actually
-- cares about: "what exactly did you change?" and "how do I get it into my
-- codebase safely?"
--
-- THE SAFETY INVARIANT THIS TABLE EXISTS TO SUPPORT: the client's repository is
-- NEVER written to. Changes leave as a patch (and optionally as a mirror repo
-- the client controls), never as a push to the branch they handed us. The patch
-- is generated against the exact commit the work was based on, so the client can
-- apply it themselves and see precisely what will happen first.
--
--   base_commit_sha — the repository commit the work started from (the pinned
--                     revision from repo_connections.last_commit_sha). Without
--                     it a patch has no meaning: "changed" relative to what?
--   baseline_ref    — the workspace-local git commit that captured the SEEDED
--                     state. The diff is baseline_ref..HEAD, which is exactly
--                     the workflow's own work and nothing else.
--   patch           — the unified diff, stored so the client can download the
--                     same bytes they reviewed rather than a re-computed file.
--   changed_files   — name-status list, stored as JSONB because it is a small
--                     list read whole, never queried per element.
--
-- ONE ROW PER WORKFLOW (the unique index): a workflow has one delivery. Making
-- the patch a mutable single row means regenerating it cannot leave two
-- competing answers to "what did you change?".
--
-- Reversible: down drops the table. Patches can be regenerated from the
-- workspace as long as it still exists, so nothing irreplaceable is lost.

CREATE TABLE codebase_deliveries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    base_commit_sha VARCHAR(64),
    baseline_ref    VARCHAR(64),

    changed_files   JSONB NOT NULL DEFAULT '[]'::jsonb,
    patch           TEXT NOT NULL DEFAULT '',
    patch_bytes     INTEGER NOT NULL DEFAULT 0,

    generated_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX codebase_deliveries_workflow_key
    ON codebase_deliveries(workflow_id);
