-- Collaborative Design Architecture — schema foundation.
--
-- See docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md. This migration adds only the
-- data layer; no behaviour changes until the Go/TS code that uses these tables
-- ships. Every statement is written to be safe to re-run: golang-migrate tracks
-- versions, but the rest of this migration set is idempotent and this one keeps
-- that property (IF NOT EXISTS on tables/columns, guarded DO blocks on
-- constraints, ON CONFLICT on seed rows).
--
-- Grouped into one migration (018) on purpose — it is one feature, and the doc
-- (§10) commits to a single migration so a reviewer reads one file.

-- ============================================================
-- 1. Section assignment (§3.2)
-- ============================================================
-- One row per (workflow, expert) the first time that expert authors. The
-- assignment never changes afterwards, so section files committed to git are
-- never renumbered or renamed when the roster changes later.
--
--   section_no   = (existing rows for this workflow + 1) * 10   [assigned in Go]
--   section_path = design/{section_no}-{experts.slug}.md        [assigned in Go]
--
-- experts.slug is already UNIQUE NOT NULL (migration 001), so section_path
-- cannot collide. The two UNIQUE constraints below make a collision a database
-- error rather than a silent overwrite — the exact failure mode
-- (event-type -> fixed filename) that turned 11 artifacts into 4 files in the
-- Aider phase (commit 1faf3e5).
CREATE TABLE IF NOT EXISTS workflow_design_sections (
    workflow_id  UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    expert_id    UUID NOT NULL REFERENCES experts(id),
    section_no   INTEGER NOT NULL,
    section_path TEXT    NOT NULL,
    assigned_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workflow_id, expert_id),
    UNIQUE (workflow_id, section_no),
    UNIQUE (workflow_id, section_path)
);

COMMENT ON TABLE workflow_design_sections IS
    'One row per expert per workflow, assigned on first authoring. section_no and section_path never change afterwards. See COLLABORATIVE_DESIGN_ARCHITECTURE.md §3.2.';

-- ============================================================
-- 2. Workflow chat (§6) — fully separate from chats/messages
-- ============================================================
-- These do NOT touch the product chat tables (chats, messages). The separation
-- is deliberate (§6.1): the product chat is shipped, and threading workflow
-- concerns through it would put every workflow-chat change one bug away from
-- breaking a paid feature. Zero regression surface on the product chat.
CREATE TABLE IF NOT EXISTS workflow_chats (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id           UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    client_id             UUID NOT NULL REFERENCES users(id),
    -- The deliverable this chat is about. NULL = a chat about the workflow as a
    -- whole rather than one artifact.
    pinned_event_id       UUID REFERENCES blackboard_events(id),
    title                 VARCHAR(500) NOT NULL,
    -- Knowledge mode (§6.4). NULL = inherit workflows.generic_allowance_pct, so
    -- a chat opened inside a strict workflow is strict by default.
    generic_allowance_pct DECIMAL(5,2),
    message_count         INTEGER NOT NULL DEFAULT 0,
    is_archived           BOOLEAN NOT NULL DEFAULT FALSE,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Same 0..100 bound as workflows after this migration widens it (§4 below).
    -- NULL is allowed (means inherit); the bound only applies when set.
    CONSTRAINT workflow_chats_generic_pct_check
        CHECK (generic_allowance_pct IS NULL OR
               (generic_allowance_pct >= 0 AND generic_allowance_pct <= 100))
);

CREATE INDEX IF NOT EXISTS idx_workflow_chats_workflow ON workflow_chats(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_chats_client   ON workflow_chats(client_id);
CREATE INDEX IF NOT EXISTS idx_workflow_chats_pinned   ON workflow_chats(pinned_event_id) WHERE pinned_event_id IS NOT NULL;

COMMENT ON TABLE workflow_chats IS
    'Deliverable-scoped chat inside a workflow. Separate from chats/messages by design. See COLLABORATIVE_DESIGN_ARCHITECTURE.md §6.';

CREATE TABLE IF NOT EXISTS workflow_chat_participants (
    chat_id   UUID NOT NULL REFERENCES workflow_chats(id) ON DELETE CASCADE,
    expert_id UUID NOT NULL REFERENCES experts(id),
    added_by  UUID NOT NULL REFERENCES users(id),
    added_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, expert_id)
);

COMMENT ON TABLE workflow_chat_participants IS
    'Experts participating in a workflow chat beyond the pinned deliverable''s author. See §6.3.';

CREATE TABLE IF NOT EXISTS workflow_chat_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES workflow_chats(id) ON DELETE CASCADE,
    -- user | assistant | tool  (tool = one step of the §7 tool loop)
    role            VARCHAR(20) NOT NULL,
    expert_id       UUID REFERENCES experts(id),
    content         TEXT NOT NULL,
    turn_number     INTEGER NOT NULL,
    -- Tool loop (§7). NULL for plain user/assistant messages.
    tool_name       VARCHAR(100),
    tool_input      JSONB,
    tool_result     JSONB,
    step_number     INTEGER,
    citations       JSONB,
    -- Which knowledge mode produced this answer (§6.4) — an audit record so a
    -- later reader knows whether a statement came from training or generic fill.
    gate_result     JSONB,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    cost_usd        DECIMAL(12,8) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workflow_chat_messages_role_check
        CHECK (role IN ('user', 'assistant', 'tool'))
);

CREATE INDEX IF NOT EXISTS idx_wcm_chat_created ON workflow_chat_messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_wcm_chat_turn    ON workflow_chat_messages(chat_id, turn_number, step_number);

COMMENT ON TABLE workflow_chat_messages IS
    'Messages and tool-loop steps for a workflow chat. role=tool rows record one gather/act step. See §6.2, §7.';

-- ============================================================
-- 3. Ordering hint + integrator choice
-- ============================================================
-- authoring_rank (§4.2, §16.2): a hint fed to the planner prompt, NOT an override
-- of the DAG. NULL = no opinion (the default), which changes nothing. Also used
-- by §16.2 to auto-settle a same-domain detail conflict: higher rank wins.
ALTER TABLE expert_categories ADD COLUMN IF NOT EXISTS authoring_rank INTEGER;

COMMENT ON COLUMN expert_categories.authoring_rank IS
    'Optional ordering hint for the planner and tie-breaker for same-domain detail conflicts. NULL = no opinion. See §4.2, §16.2.';

-- Integrator choice (§16.6): auto by default (most-upstream participant, computed
-- in Go), with a manual override for workflows that have no System Design or PM
-- role. integrator_auto=TRUE ignores integrator_expert_id.
ALTER TABLE workflows ADD COLUMN IF NOT EXISTS integrator_auto      BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE workflows ADD COLUMN IF NOT EXISTS integrator_expert_id UUID REFERENCES experts(id);

COMMENT ON COLUMN workflows.integrator_auto IS
    'TRUE = integrator is the most-upstream participant, computed. FALSE = use integrator_expert_id. See §16.6.';

-- ============================================================
-- 4. Widen the generic ceiling from a hard 30 to the real 0..100 bound (§11)
-- ============================================================
-- The 30 was a business ceiling baked into a table constraint. Per §11 it splits
-- into a safety bound (this CHECK, 0..100 — the real range of a percentage) and a
-- business ceiling that moves to system_settings and is enforced by the handler.
-- Result: an admin can lower or raise the business ceiling without a migration,
-- and no bug can ever store a value outside 0..100.
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
        CHECK (generic_allowance_pct >= 0 AND generic_allowance_pct <= 100);
END $$;

COMMENT ON COLUMN workflows.generic_allowance_pct IS
    'Client-set generic-knowledge percentage, 0..100 (safety bound). The business ceiling (default 30) lives in system_settings.generic_allowance_ceiling and is enforced in the handler. See §11.';

-- ============================================================
-- 5. Extend the approval gate CHECK (§5, §7.5)
-- ============================================================
-- approval_requests.gate_name HAS a CHECK (migration 006). Conflict resolution
-- and amendment approval need two new gate names so the Kanban can label them.
-- Postgres has no ADD CONSTRAINT IF NOT EXISTS, so drop-then-add, guarded.
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
            'ad_hoc',
            'design_conflict',
            'design_amendment'
        ));
END $$;

-- ============================================================
-- 6. Tunable values move from Go constants to settings (§11)
-- ============================================================
-- Each row's default is the value currently hardcoded in Go, so seeding these
-- changes no behaviour until the code is switched to read them. ON CONFLICT keeps
-- the migration re-runnable and never overwrites an admin's later change.
INSERT INTO system_settings (key, value, description) VALUES
    ('gate1_strong_threshold', '0.70',
     'Gate 1: own-training score at/above which generic is fully blocked. Was gate_system.go:27.'),
    ('gate1_usable_threshold', '0.40',
     'Gate 1: own-training score at/above which training is usable as principles. Was gate_system.go:38.'),
    ('gate1_max_chunks', '5',
     'Gate 1: max own-training chunks into the prompt. Was gate_system.go:43.'),
    ('aider_max_iterations', '5',
     'Aider loop max iterations per task. Was aider_runner.go:521.'),
    ('max_design_attempts', '10',
     'Max design re-runs on client request. Was runner.go:26.'),
    ('qa_coverage_target', '80.0',
     'QA phase Go coverage target percent. Was aider_runner.go:1175.'),
    ('tool_loop_max_steps', '8',
     'Workflow-chat tool loop step cap (§7.3). New — no prior constant.'),
    ('generic_allowance_ceiling', '30',
     'Business ceiling on generic_allowance_pct, enforced in the handler (§11). Safety bound (0..100) stays in the workflows CHECK.')
ON CONFLICT (key) DO NOTHING;
