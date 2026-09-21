-- Reverse of 018. Order matters: drop dependent objects before the things they
-- reference. Chat message/participant tables reference workflow_chats, which
-- references workflows; sections and columns come last.

-- 6. Settings seeded by this migration.
DELETE FROM system_settings WHERE key IN (
    'gate1_strong_threshold',
    'gate1_usable_threshold',
    'gate1_max_chunks',
    'aider_max_iterations',
    'max_design_attempts',
    'qa_coverage_target',
    'tool_loop_max_steps',
    'generic_allowance_ceiling'
);

-- 5. Restore the original 6-value approval gate CHECK.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'approval_requests_gate_check'
    ) THEN
        ALTER TABLE approval_requests DROP CONSTRAINT approval_requests_gate_check;
    END IF;

    ALTER TABLE approval_requests
        ADD CONSTRAINT approval_requests_gate_check
        CHECK (gate_name IN (
            'intake',
            'high_level_design',
            'detailed_design',
            'handoff',
            'budget_exceeded',
            'ad_hoc'
        ));
END $$;

-- 4. Restore the 30 ceiling on workflows.generic_allowance_pct.
--
-- KNOWN LIMITATION: if any workflow row holds a value in 31..100 (only possible
-- after this feature has been used), this ADD CONSTRAINT will fail — Postgres
-- validates existing rows. That is correct: rolling back a feature whose data is
-- out of the old range cannot silently succeed. Bring such rows to <=30 first.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'workflows_generic_allowance_pct_check'
    ) THEN
        ALTER TABLE workflows DROP CONSTRAINT workflows_generic_allowance_pct_check;
    END IF;

    ALTER TABLE workflows
        ADD CONSTRAINT workflows_generic_allowance_pct_check
        CHECK (generic_allowance_pct >= 0 AND generic_allowance_pct <= 30);
END $$;

-- 3. Integrator columns and the ordering hint.
ALTER TABLE workflows DROP COLUMN IF EXISTS integrator_expert_id;
ALTER TABLE workflows DROP COLUMN IF EXISTS integrator_auto;
ALTER TABLE expert_categories DROP COLUMN IF EXISTS authoring_rank;

-- 2. Workflow chat tables — children before parent.
DROP TABLE IF EXISTS workflow_chat_messages;
DROP TABLE IF EXISTS workflow_chat_participants;
DROP TABLE IF EXISTS workflow_chats;

-- 1. Section assignment.
DROP TABLE IF EXISTS workflow_design_sections;
