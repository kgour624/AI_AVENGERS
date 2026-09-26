-- VIP Pass vs Strict, chosen per expert at creation time in the admin panel.
--
-- 'full_strip' (default, strict): Layer 4 strips every sentence with no citation,
-- so an expert may only say what the training material backs.
-- 'code_exempt' (VIP pass): fenced code blocks survive uncited — for experts whose
-- job is to produce code.
--
-- WHY per-expert rather than per-domain: two experts of the same domain (a design
-- expert and a coding expert) legitimately need different strictness, and the
-- choice belongs to whoever creates that expert.
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS strip_mode VARCHAR(20) NOT NULL DEFAULT 'full_strip';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experts_strip_mode_check') THEN
        ALTER TABLE experts DROP CONSTRAINT experts_strip_mode_check;
    END IF;
    ALTER TABLE experts
        ADD CONSTRAINT experts_strip_mode_check
        CHECK (strip_mode IN ('full_strip', 'code_exempt'));
END $$;

COMMENT ON COLUMN experts.strip_mode IS
    'full_strip = strict (uncited sentences stripped); code_exempt = VIP pass (fenced code kept uncited). Chosen per expert in the admin panel.';
