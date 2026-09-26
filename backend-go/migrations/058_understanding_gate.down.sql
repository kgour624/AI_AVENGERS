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
            'intake', 'high_level_design', 'detailed_design', 'handoff',
            'budget_exceeded', 'ad_hoc', 'design_conflict', 'design_amendment'
        ));
END $$;

DELETE FROM approval_requests WHERE gate_name = 'understanding';
