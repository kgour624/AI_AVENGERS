-- The requirement-understanding gate needs its own approval gate name so the
-- client can approve what each expert understood BEFORE any design work starts.
-- Same drop-then-add pattern migration 018 used (Postgres has no
-- ADD CONSTRAINT IF NOT EXISTS).
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
            'understanding',
            'high_level_design',
            'detailed_design',
            'handoff',
            'budget_exceeded',
            'ad_hoc',
            'design_conflict',
            'design_amendment'
        ));
END $$;
