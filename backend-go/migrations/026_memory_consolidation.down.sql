-- Rollback 026: drop B7 consolidation columns/indexes; restore original
-- memory_type enum. Rows with memory_type IN ('summary','preference') must
-- be removed first or the restored CHECK fails.

DELETE FROM project_memory_l2
 WHERE memory_type IN ('summary', 'preference');

DROP INDEX IF EXISTS idx_l2_project_active_created;
DROP INDEX IF EXISTS idx_l2_decay;

ALTER TABLE project_memory_l2
    DROP CONSTRAINT IF EXISTS l2_weight_range;

ALTER TABLE project_memory_l2
    DROP COLUMN IF EXISTS last_accessed_at;

ALTER TABLE project_memory_l2
    DROP COLUMN IF EXISTS weight;

ALTER TABLE project_memory_l2
    DROP CONSTRAINT IF EXISTS l2_memory_type_check;

ALTER TABLE project_memory_l2
    ADD CONSTRAINT l2_memory_type_check CHECK (
        memory_type IN (
            'decision', 'code', 'error', 'fix', 'recommendation', 'requirement'
        )
    );
