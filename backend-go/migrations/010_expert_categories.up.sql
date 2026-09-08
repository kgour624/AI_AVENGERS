-- ============================================================
-- Migration 010: Expert Categories + Template System + Reply Threading
-- ============================================================
-- Adds the schema required by CATEGORY_TEMPLATE_HANDOFF.md.
--
-- Scope: admin-managed expert category layer (owns structured JSON
-- answer templates), nullable category assignment on experts,
-- reply-to-message threading on messages, retrofit of existing
-- coding-domain experts into a seeded category.
--
-- Design principles applied (same as migration 006):
-- - Additive only. No RENAMEs, DROPs, or type changes on existing tables.
-- - All ADD COLUMNs use IF NOT EXISTS — safe to re-run.
-- - snake_case consistent with existing schema.
-- - CHECK constraints added via DO blocks so re-runs don't fail.
-- - Every new table has created_at; mutable tables also have updated_at.
--
-- Locked decisions (CATEGORY_TEMPLATE_HANDOFF.md §2):
--   CT-L1: existing DomainRegistry/DomainProfile (chinawall package) —
--          NOT touched by this migration or any file in this change.
--   CT-L2: experts.category_id is NULLABLE — zero regression for any
--          expert not yet retrofitted; flat-text behavior stays the
--          fallback, not an error state.
--   CT-L4: test-case bucket labels (BASE/EDGE/CORNER/STRESS) are a
--          hardcoded Go constant, never stored here as configurable data.
--   CT-L11: both existing and future experts get retrofit-assigned a
--          category where a confident domain match exists.
-- ============================================================

-- ============================================================
-- 1. EXPERT_CATEGORIES — admin-owned category + template layer
-- ============================================================
-- template_schema shape (documented, not enforced by a DB CHECK because
-- JSONB structural validation is done at the application layer in
-- admin_handler.go, matching the existing pattern for allowed_tools):
-- {
--   "sections": [
--     {"key": "pattern", "label": "Pattern", "type": "prose", "required": true},
--     {"key": "code",    "label": "Code",    "type": "code",  "required": true},
--     ...
--   ]
-- }
-- Valid section "type" values (enforced at application layer, not DB):
-- prose | code | test_cases.
CREATE TABLE IF NOT EXISTS expert_categories (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                     VARCHAR(255) NOT NULL,
    slug                     VARCHAR(255) UNIQUE NOT NULL,
    description              TEXT,
    template_schema          JSONB        NOT NULL DEFAULT '{}',
    default_language         VARCHAR(20)  NOT NULL DEFAULT 'java',
    ask_structure_permission BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by               UUID REFERENCES users(id),
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expert_categories_slug ON expert_categories(slug);

COMMENT ON TABLE expert_categories IS
    'Admin-managed category layer above experts.domain. Owns the structured JSON answer template (template_schema) that applies to every expert placed in this category. See CATEGORY_TEMPLATE_HANDOFF.md.';
COMMENT ON COLUMN expert_categories.template_schema IS
    'JSONB: {"sections": [{"key","label","type" (prose|code|test_cases),"required"}, ...]}. Empty {} means no structured template — expert falls back to flat-text behavior (CT-L2).';
COMMENT ON COLUMN expert_categories.default_language IS
    'Code language used for any "code"-type template section. Defaults to java at the code level (CT-L5), overridable per category by admin.';
COMMENT ON COLUMN expert_categories.ask_structure_permission IS
    'If true, a fresh (non-reply) question to an expert in this category triggers an ASK-mode structure-permission question before generating the full templated answer (CT-L9).';

-- ============================================================
-- 2. EXPERTS — nullable category assignment
-- ============================================================
-- WHY nullable (CT-L2): an expert with category_id = NULL must behave
-- exactly as it does today (flat free-text response). This is the
-- backward-compat fallback, not an error state. Category is a NEW
-- layer ABOVE the existing experts.domain column (CT-L1) — domain is
-- unchanged and still drives chinawall.DomainRegistry lookups.
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES expert_categories(id);

CREATE INDEX IF NOT EXISTS idx_experts_category ON experts(category_id) WHERE category_id IS NOT NULL;

COMMENT ON COLUMN experts.category_id IS
    'Nullable FK to expert_categories. NULL = flat-text behavior (no structured template), the safe default for any expert not yet retrofitted. See CATEGORY_TEMPLATE_HANDOFF.md CT-L2.';

-- ============================================================
-- 3. MESSAGES — reply-to threading (single parent pointer)
-- ============================================================
-- Self-referencing, single immediate-parent pointer only. Full-chain
-- resolution (when the client opts into "include full thread") is a
-- runtime traversal in context/assembler.go, NOT a stored list —
-- keeps writes trivial and avoids denormalized drift (CATEGORY_TEMPLATE_HANDOFF.md §5).
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_messages_reply_to
    ON messages(reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;

COMMENT ON COLUMN messages.reply_to_message_id IS
    'Immediate parent message this one replies to. NULL = fresh question, not a reply. Full-thread resolution walks this pointer upward at query time, capped by system_settings.reply_thread_max_depth. See CATEGORY_TEMPLATE_HANDOFF.md §5.';

-- ============================================================
-- 4. RETROFIT — seed a Coding category and assign existing coding-
--    domain experts into it (CT-L11: existing experts get retrofitted,
--    not left permanently uncategorized).
-- ============================================================
-- WHY these three domain values specifically: verified against the
-- actual DefaultProfiles list in backend-go/internal/chinawall/domain_profile.go
-- (domains: dsa, algorithms, system_design, coding, oops, medical,
-- finance, legal). dsa/algorithms/coding are the three coding-style
-- domains this category's template (Pattern/Idea/Code/Walkthrough/
-- TestCases) is designed for. system_design and oops are intentionally
-- excluded here — they need their own category with a different
-- template shape (admin will create that separately per CT-L2/L8),
-- not force-fitted into this one.
DO $$
DECLARE
    v_category_id UUID;
BEGIN
    -- Idempotent: only seed if a category with this slug doesn't exist yet.
    -- This matches migration 006's "admin overrides preserved" pattern —
    -- if this migration is re-run, or if admin already renamed/edited the
    -- seeded row, we must not clobber it.
    IF NOT EXISTS (SELECT 1 FROM expert_categories WHERE slug = 'coding') THEN
        INSERT INTO expert_categories (name, slug, description, template_schema, default_language)
        VALUES (
            'Coding',
            'coding',
            'DSA / algorithms / general coding experts. Structured answer template: Pattern, Idea, Code, Walkthrough, Test Cases.',
            '{"sections": [
                {"key": "pattern",     "label": "Pattern",     "type": "prose",      "required": true},
                {"key": "idea",        "label": "Idea",        "type": "prose",      "required": true},
                {"key": "code",        "label": "Code",        "type": "code",       "required": true},
                {"key": "walkthrough", "label": "Walkthrough", "type": "prose",      "required": true},
                {"key": "test_cases",  "label": "Test Cases",  "type": "test_cases", "required": true}
            ]}'::jsonb,
            'java'
        )
        RETURNING id INTO v_category_id;

        -- Retrofit: assign existing experts whose domain matches this
        -- category's scope. Case-insensitive match, matching the same
        -- normalization chinawall.DomainRegistry.Get() already applies
        -- (strings.ToLower(strings.TrimSpace(domain))).
        -- Only touches experts that don't already have a category_id,
        -- so re-running this migration never overwrites an admin's
        -- manual (re)assignment.
        UPDATE experts
        SET category_id = v_category_id,
            updated_at = NOW()
        WHERE category_id IS NULL
          AND LOWER(TRIM(domain)) IN ('dsa', 'algorithms', 'coding');
    END IF;
END $$;

-- ============================================================
-- 5. SYSTEM SETTINGS — reply-thread depth cap
-- ============================================================
INSERT INTO system_settings (key, value, description) VALUES
(
    'reply_thread_max_depth',
    '10',
    'Safety cap on how many ancestor messages are walked when a client opts into "include full thread" on a reply. See CATEGORY_TEMPLATE_HANDOFF.md §5.'
)
ON CONFLICT (key) DO NOTHING;

-- ============================================================
-- End of migration 010
-- ============================================================
