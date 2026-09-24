-- Migration 026: project memory consolidation + weighted decay (B7)
--
-- WHY: L2 grows unbounded; semantic search returns stale/noisy decisions.
-- Roadmap B7 + Arpit last-video P4/P5: hierarchical consolidate at
-- session/phase end (not per-event); decay preferences not facts;
-- critical (importance >= 5) never decay; externalize via supersede
-- (row kept, is_superseded=TRUE) so L3/audit + re-retrieve still work —
-- never hard-delete L2 rows.
--
-- memory_type expands with:
--   'summary'     — project-level consolidated doc (growing-doc pattern)
--   'preference'  — soft preference eligible for exponential weight decay
-- Existing types (decision/code/error/fix/recommendation/requirement)
-- stay facts: weight stays 1.0, never decayed.
--
-- weight: [0, 1]. Search ranks by weight * importance. Below floor
-- (0.05) a preference is superseded (forgotten), not deleted.
-- last_accessed_at: refreshed on semantic/keyword hit so frequently
-- retrieved prefs resist decay (LFU-style proxy).

ALTER TABLE project_memory_l2
    DROP CONSTRAINT IF EXISTS l2_memory_type_check;

ALTER TABLE project_memory_l2
    ADD CONSTRAINT l2_memory_type_check CHECK (
        memory_type IN (
            'decision', 'code', 'error', 'fix', 'recommendation', 'requirement',
            'summary', 'preference'
        )
    );

ALTER TABLE project_memory_l2
    ADD COLUMN IF NOT EXISTS weight REAL NOT NULL DEFAULT 1.0;

ALTER TABLE project_memory_l2
    ADD COLUMN IF NOT EXISTS last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE project_memory_l2
    DROP CONSTRAINT IF EXISTS l2_weight_range;

ALTER TABLE project_memory_l2
    ADD CONSTRAINT l2_weight_range CHECK (weight >= 0 AND weight <= 1);

-- Active prefs needing decay scans; active raw entries needing consolidate.
CREATE INDEX IF NOT EXISTS idx_l2_decay
    ON project_memory_l2 (memory_type, last_accessed_at)
    WHERE is_superseded = FALSE AND memory_type = 'preference';

CREATE INDEX IF NOT EXISTS idx_l2_project_active_created
    ON project_memory_l2 (project_id, created_at)
    WHERE is_superseded = FALSE AND memory_type <> 'summary';

COMMENT ON COLUMN project_memory_l2.weight IS
    'B7 retention weight. Facts stay 1.0; preferences decay exponentially from last_accessed_at. importance>=5 never decays.';
COMMENT ON COLUMN project_memory_l2.last_accessed_at IS
    'B7: bumped on retrieval hit so hot preferences resist decay.';
