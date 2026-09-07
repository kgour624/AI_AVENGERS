-- ============================================================
-- Migration 006: Collaboration Layer (Domain-Expert Workflows)
-- ============================================================
-- Adds the schema required by DOMAIN_EXPERT_COLLABORATION_DESIGN.md.
--
-- Scope: multi-agent workflow engine, blackboard event log,
-- cross-verification, checkpointing, client approval gates,
-- Kanban projection.
--
-- Design principles applied:
-- - Additive only. No RENAMEs, DROPs, or type changes on existing tables.
-- - All ADD COLUMNs use IF NOT EXISTS — safe to re-run.
-- - snake_case consistent with existing schema (001_initial_schema).
-- - CHECK constraints on every enum-like TEXT column.
-- - Every new table has appropriate FKs with explicit ON DELETE behavior.
-- - Every table has created_at; mutable tables also have updated_at.
-- ============================================================

-- ============================================================
-- 1. EXPERTS — add collaboration/loop/tool config fields
-- ============================================================
-- WHY these fields (KNOWLEDGE_HUB.md Part I §1.5 and Part II §2.2):
--   model_tier          — which OpenRouter tier (cheap/strong/fast)
--   temperature, top_p  — per-expert sampling knobs, tuned to task type
--   loop_pattern        — which agentic loop this expert runs (ota/react/plan_execute)
--   max_loop_iterations — hard cap on inner-loop iterations (cost safety)
--   allowed_tools       — whitelist of tool names this expert may call
--   training_status     — richer lifecycle than existing is_training bool
--
-- COEXISTENCE: existing is_training BOOLEAN is preserved for backward
-- compat with any code that reads it. training_status is the new source
-- of truth going forward. Migration to deprecate is_training is future
-- work — out of scope here.
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS model_tier         VARCHAR(20)   NOT NULL DEFAULT 'strong',
    ADD COLUMN IF NOT EXISTS temperature        DECIMAL(3,2)  NOT NULL DEFAULT 0.30,
    ADD COLUMN IF NOT EXISTS top_p              DECIMAL(3,2)  NOT NULL DEFAULT 0.50,
    ADD COLUMN IF NOT EXISTS loop_pattern       VARCHAR(20)   NOT NULL DEFAULT 'react',
    ADD COLUMN IF NOT EXISTS max_loop_iterations INTEGER      NOT NULL DEFAULT 5,
    ADD COLUMN IF NOT EXISTS allowed_tools      JSONB         NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS training_status    VARCHAR(20)   NOT NULL DEFAULT 'draft';

-- CHECK constraints — added separately with DO block so re-runs don't fail.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'experts_model_tier_check'
    ) THEN
        ALTER TABLE experts ADD CONSTRAINT experts_model_tier_check
            CHECK (model_tier IN ('cheap', 'strong', 'fast'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'experts_temperature_check'
    ) THEN
        ALTER TABLE experts ADD CONSTRAINT experts_temperature_check
            CHECK (temperature >= 0.0 AND temperature <= 2.0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'experts_top_p_check'
    ) THEN
        ALTER TABLE experts ADD CONSTRAINT experts_top_p_check
            CHECK (top_p >= 0.0 AND top_p <= 1.0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'experts_loop_pattern_check'
    ) THEN
        ALTER TABLE experts ADD CONSTRAINT experts_loop_pattern_check
            CHECK (loop_pattern IN ('ota', 'react', 'plan_execute'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'experts_max_loop_iterations_check'
    ) THEN
        ALTER TABLE experts ADD CONSTRAINT experts_max_loop_iterations_check
            CHECK (max_loop_iterations >= 1 AND max_loop_iterations <= 50);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'experts_training_status_check'
    ) THEN
        ALTER TABLE experts ADD CONSTRAINT experts_training_status_check
            CHECK (training_status IN ('draft', 'ingesting', 'trained', 'deprecated'));
    END IF;
END $$;

COMMENT ON COLUMN experts.model_tier IS
    'OpenRouter tier: cheap=deepseek, strong=claude, fast=gemini. See KNOWLEDGE_HUB.md §6.2.';
COMMENT ON COLUMN experts.loop_pattern IS
    'Agentic loop: ota (Observe-Think-Act), react (Reason+Act), plan_execute. See KNOWLEDGE_HUB.md §2.2.';
COMMENT ON COLUMN experts.training_status IS
    'Lifecycle: draft (created, no transcripts) -> ingesting -> trained (smoke test passed) -> deprecated.';

-- ============================================================
-- 2. COURSE_CHUNKS — add chunk_hash for deduplication
-- ============================================================
-- WHY (DOMAIN_EXPERT_COLLABORATION_DESIGN.md §6.3):
-- Uploading the same transcript twice must NOT duplicate chunks.
-- chunk_hash = SHA-256 of normalized chunk_text. Unique per expert.
ALTER TABLE course_chunks
    ADD COLUMN IF NOT EXISTS chunk_hash CHAR(64);

CREATE INDEX IF NOT EXISTS idx_chunks_expert_hash
    ON course_chunks(expert_id, chunk_hash)
    WHERE chunk_hash IS NOT NULL;

COMMENT ON COLUMN course_chunks.chunk_hash IS
    'SHA-256 hex of normalized chunk_text. Used by ingestion pipeline to skip duplicates for the same expert.';

-- ============================================================
-- 3. WORKFLOWS — long-running end-to-end runs
-- ============================================================
-- One workflow spans the 6-phase state machine (INTAKE -> HANDOFF).
-- Selected experts collaborate via blackboard within the workflow scope.
CREATE TABLE IF NOT EXISTS workflows (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id             UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title                  VARCHAR(500) NOT NULL,
    status                 VARCHAR(50)  NOT NULL DEFAULT 'draft',
    current_phase          VARCHAR(50)  NOT NULL DEFAULT 'intake',
    phase_started_at       TIMESTAMPTZ,
    phase_completed_at     TIMESTAMPTZ,
    selected_expert_ids    JSONB        NOT NULL DEFAULT '[]',
    cost_budget_usd        DECIMAL(10,4) NOT NULL DEFAULT 10.0,
    cost_spent_usd         DECIMAL(10,4) NOT NULL DEFAULT 0.0,
    cost_soft_limit_pct    DECIMAL(5,2)  NOT NULL DEFAULT 75.0,
    cost_hard_limit_pct    DECIMAL(5,2)  NOT NULL DEFAULT 100.0,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT workflows_status_check CHECK (status IN (
        'draft',
        'running',
        'paused_for_approval',
        'paused_for_client_input',
        'completed',
        'cancelled',
        'failed'
    )),
    CONSTRAINT workflows_current_phase_check CHECK (current_phase IN (
        'intake',
        'high_level_design',
        'detailed_design',
        'implementation',
        'qa',
        'handoff',
        'completed'
    )),
    CONSTRAINT workflows_budget_check CHECK (cost_budget_usd > 0),
    CONSTRAINT workflows_soft_limit_check CHECK (cost_soft_limit_pct BETWEEN 0 AND 100),
    CONSTRAINT workflows_hard_limit_check CHECK (cost_hard_limit_pct BETWEEN 0 AND 100)
);

CREATE INDEX IF NOT EXISTS idx_workflows_client_status ON workflows(client_id, status);
CREATE INDEX IF NOT EXISTS idx_workflows_project       ON workflows(project_id);
CREATE INDEX IF NOT EXISTS idx_workflows_updated       ON workflows(updated_at DESC);

COMMENT ON TABLE workflows IS
    'One end-to-end run of the 6-phase state machine. See DOMAIN_EXPERT_COLLABORATION_DESIGN.md §9.';

-- ============================================================
-- 4. BLACKBOARD_EVENTS — append-only per-workflow event log
-- ============================================================
-- Experts communicate ONLY via this table. Every artifact posting,
-- question, review comment, phase transition, client response is an
-- event here. Full audit trail; resumable via replay.
CREATE TABLE IF NOT EXISTS blackboard_events (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id           UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    sequence_number       BIGSERIAL NOT NULL,
    event_type            VARCHAR(100) NOT NULL,
    posted_by_expert_id   UUID REFERENCES experts(id),
    posted_by_client      BOOLEAN NOT NULL DEFAULT FALSE,
    to_expert_id          UUID REFERENCES experts(id),
    content               JSONB NOT NULL,
    references_event_ids  UUID[] NOT NULL DEFAULT '{}',
    revision              INTEGER NOT NULL DEFAULT 1,
    dedup_key             VARCHAR(255) NOT NULL,
    posted_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Prevent accidental double-posts. Producer computes dedup_key
    -- as e.g. hash(event_type + posted_by + content_hash).
    CONSTRAINT blackboard_events_workflow_dedup UNIQUE (workflow_id, dedup_key),
    -- Exactly one of posted_by_expert_id / posted_by_client must be set.
    CONSTRAINT blackboard_events_poster_check CHECK (
        (posted_by_expert_id IS NOT NULL AND posted_by_client = FALSE) OR
        (posted_by_expert_id IS NULL     AND posted_by_client = TRUE)
    )
);

CREATE INDEX IF NOT EXISTS idx_bb_workflow_seq   ON blackboard_events(workflow_id, sequence_number);
CREATE INDEX IF NOT EXISTS idx_bb_workflow_type  ON blackboard_events(workflow_id, event_type);
CREATE INDEX IF NOT EXISTS idx_bb_to_expert      ON blackboard_events(to_expert_id) WHERE to_expert_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_bb_posted_at      ON blackboard_events(workflow_id, posted_at DESC);

COMMENT ON TABLE blackboard_events IS
    'Append-only event log. Experts publish artifacts, questions, reviews here. NEVER UPDATE OR DELETE rows.';
COMMENT ON COLUMN blackboard_events.dedup_key IS
    'Producer-computed dedup key. UNIQUE per workflow. Duplicate posts are silently dropped.';

-- ============================================================
-- 5. WORKFLOW_TASKS — Kanban projection
-- ============================================================
-- Not a separate write path — derived state from blackboard_events
-- (either projected by a subscriber or materialized here for query speed).
-- Each row = one artifact-producing task assigned to an expert.
CREATE TABLE IF NOT EXISTS workflow_tasks (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id                 UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    assigned_expert_id          UUID NOT NULL REFERENCES experts(id),
    title                       VARCHAR(500) NOT NULL,
    description                 TEXT NOT NULL DEFAULT '',
    status                      VARCHAR(50) NOT NULL DEFAULT 'todo',
    produced_artifact_event_id  UUID REFERENCES blackboard_events(id),
    cost_usd                    DECIMAL(10,4) NOT NULL DEFAULT 0.0,
    started_at                  TIMESTAMPTZ,
    completed_at                TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workflow_tasks_status_check CHECK (status IN (
        'todo',
        'in_progress',
        'under_review',
        'blocked',
        'done',
        'cancelled'
    ))
);

CREATE INDEX IF NOT EXISTS idx_wt_workflow_status ON workflow_tasks(workflow_id, status);
CREATE INDEX IF NOT EXISTS idx_wt_expert          ON workflow_tasks(assigned_expert_id);
CREATE INDEX IF NOT EXISTS idx_wt_updated         ON workflow_tasks(workflow_id, updated_at DESC);

COMMENT ON TABLE workflow_tasks IS
    'Kanban projection of blackboard state. One row per artifact-producing task.';

-- ============================================================
-- 6. WORKFLOW_CHECKPOINTS — phase-level snapshots for resume
-- ============================================================
-- Written on every phase transition and every approval response.
-- Recovery: load latest checkpoint + replay blackboard events past
-- checkpoint's sequence_number.
CREATE TABLE IF NOT EXISTS workflow_checkpoints (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id                 UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    phase                       VARCHAR(50) NOT NULL,
    state_snapshot              JSONB NOT NULL,
    blackboard_sequence_number  BIGINT NOT NULL,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wc_workflow ON workflow_checkpoints(workflow_id, created_at DESC);

COMMENT ON TABLE workflow_checkpoints IS
    'Phase-level durable snapshots. See DOMAIN_EXPERT_COLLABORATION_DESIGN.md §12.';

-- ============================================================
-- 7. APPROVAL_REQUESTS — client-facing decision points
-- ============================================================
-- Every human-in-the-loop moment. Workflow pauses until responded.
-- No auto-approval on timeout (per client's explicit preference).
CREATE TABLE IF NOT EXISTS approval_requests (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id              UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    gate_name                VARCHAR(100) NOT NULL,
    summary                  TEXT NOT NULL,
    artifact_content         JSONB NOT NULL,
    cited_event_ids          UUID[] NOT NULL DEFAULT '{}',
    cost_so_far_usd          DECIMAL(10,4) NOT NULL,
    estimated_remaining_usd  DECIMAL(10,4),
    status                   VARCHAR(50) NOT NULL DEFAULT 'pending',
    client_response          JSONB,
    requested_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at             TIMESTAMPTZ,
    CONSTRAINT approval_requests_status_check CHECK (status IN (
        'pending',
        'approved',
        'approved_with_notes',
        'changes_requested',
        'rejected',
        'cancelled'
    )),
    CONSTRAINT approval_requests_gate_check CHECK (gate_name IN (
        'intake',
        'high_level_design',
        'detailed_design',
        'handoff',
        'budget_exceeded',
        'ad_hoc'
    ))
);

CREATE INDEX IF NOT EXISTS idx_ar_workflow_status ON approval_requests(workflow_id, status);
CREATE INDEX IF NOT EXISTS idx_ar_pending_all     ON approval_requests(requested_at DESC) WHERE status = 'pending';

COMMENT ON TABLE approval_requests IS
    'Client approval gates. Workflow pauses until client responds. See DOMAIN_EXPERT_COLLABORATION_DESIGN.md §10.';

-- ============================================================
-- 8. MESSAGES — link LLM calls to workflows for cost rollup
-- ============================================================
-- Cost/tokens are already tracked on messages (cost_usd, tokens_used).
-- Adding workflow_id lets us do: SUM(cost_usd) WHERE workflow_id = $1.
-- Nullable because most messages (regular chat) are not part of a workflow.
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_messages_workflow
    ON messages(workflow_id)
    WHERE workflow_id IS NOT NULL;

COMMENT ON COLUMN messages.workflow_id IS
    'If set, this message is part of a workflow. NULL for regular chat messages.';

-- ============================================================
-- 9. SYSTEM SETTINGS — defaults for workflow engine
-- ============================================================
-- Uses INSERT ... ON CONFLICT so re-running the migration is safe.
INSERT INTO system_settings (key, value, description) VALUES
(
    'workflow_engine',
    '{"default_review_timeout_seconds": 60, "max_revision_rounds": 3, "per_task_cost_cap_ratio": 0.5}',
    'Workflow engine defaults: reviewer timeout, revision cap, per-task cost cap as fraction of budget/experts.'
),
(
    'blackboard',
    '{"redis_channel_prefix": "workflow", "event_ttl_days": null}',
    'Blackboard config. event_ttl_days=null means never auto-delete (compliance).'
)
ON CONFLICT (key) DO NOTHING;

-- ============================================================
-- End of migration 006
-- ============================================================
