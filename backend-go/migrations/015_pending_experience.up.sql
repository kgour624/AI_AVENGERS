-- ============================================================
-- Migration 015: Pending Experience Bank
-- ============================================================
-- Generic knowledge used in workflows gets saved here for admin review.
-- Admin approves -> moves to course_chunks (expert's training).
-- Admin rejects -> deleted.
--
-- WHY pending (not direct save):
--   Direct save to course_chunks risks polluting training data with
--   hallucinated or low-quality generic knowledge.
--   Admin review ensures only valid knowledge becomes training.
--
-- WHY this table (not course_chunks directly):
--   course_chunks has chunk_hash UNIQUE constraint per expert.
--   pending_experience is a staging area with no such constraint.
--   Admin can review, edit, then approve to course_chunks.
-- ============================================================

CREATE TABLE IF NOT EXISTS pending_experience (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id       UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    task_title      VARCHAR(500) NOT NULL,
    task_domain     VARCHAR(100) NOT NULL DEFAULT '',
    content         TEXT NOT NULL,
    source          VARCHAR(50) NOT NULL DEFAULT 'generic',
    -- source: 'generic' = came from Gate 3 generic LLM
    --         'peer'    = came from Gate 2 peer knowledge
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- status: 'pending' | 'approved' | 'rejected'
    admin_notes     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ,
    CONSTRAINT pending_experience_status_check
        CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT pending_experience_source_check
        CHECK (source IN ('generic', 'peer'))
);

CREATE INDEX IF NOT EXISTS idx_pending_exp_expert
    ON pending_experience(expert_id, status);
CREATE INDEX IF NOT EXISTS idx_pending_exp_workflow
    ON pending_experience(workflow_id);
CREATE INDEX IF NOT EXISTS idx_pending_exp_pending
    ON pending_experience(created_at DESC)
    WHERE status = 'pending';

COMMENT ON TABLE pending_experience IS
    'Staging area for generic knowledge used in workflows. Admin reviews before promoting to course_chunks.';
COMMENT ON COLUMN pending_experience.source IS
    'generic = Gate 3 generic LLM. peer = Gate 2 peer knowledge aggregation.';
