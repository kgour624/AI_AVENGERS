-- 077_collab_relay_pipeline.up.sql
-- Live transparency + human-in-the-loop gate for Collaborative Relay mode
-- (expert chat ONLY). Three tables: runs (one per collaborative turn),
-- sections (per-expert runtime state incl. human edit/conflict), events
-- (append-only transparency log for live step rendering + reconnect replay).
-- No relationship to the workflow system's approval_requests table.

CREATE TABLE IF NOT EXISTS collab_relay_runs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id          UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_message_id  UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    status           TEXT NOT NULL DEFAULT 'running',
    plan             JSONB NOT NULL,          -- ordered [{expert_id, expert_name, section_title}]
    current_index    INT  NOT NULL DEFAULT 0, -- which section is in progress / was last attempted
    failed_at_index  INT,
    failure_reason   TEXT,
    retry_deadline   TIMESTAMPTZ,             -- now() + 24h, set only when status='relay_failed'
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS collab_relay_sections (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_run_id          UUID NOT NULL REFERENCES collab_relay_runs(id) ON DELETE CASCADE,
    section_index         INT  NOT NULL,
    expert_id             UUID NOT NULL REFERENCES experts(id),
    expert_name           TEXT NOT NULL,
    section_title         TEXT NOT NULL,
    raw_content           TEXT,               -- what the expert generated
    final_content         TEXT,               -- what human approved (== raw_content if no edit)
    edited                BOOLEAN NOT NULL DEFAULT false,
    status                TEXT NOT NULL DEFAULT 'pending',
    conflict_detected     BOOLEAN NOT NULL DEFAULT false,
    conflict_explanation  TEXT,
    resolution_source     TEXT,               -- 'expert_a' | 'expert_b' | 'human_merged' | NULL
    error_reason          TEXT,
    approved_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(relay_run_id, section_index)
);

CREATE TABLE IF NOT EXISTS collab_relay_events (
    id            BIGSERIAL PRIMARY KEY,
    relay_run_id  UUID NOT NULL REFERENCES collab_relay_runs(id) ON DELETE CASCADE,
    sequence_num  INT  NOT NULL,             -- strictly increasing per run, for replay
    step          TEXT NOT NULL,             -- see StepName taxonomy
    expert_id     UUID,
    message       TEXT NOT NULL,             -- human-readable, capped at 3500 chars
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(relay_run_id, sequence_num)
);

CREATE INDEX IF NOT EXISTS idx_relay_runs_chat    ON collab_relay_runs(chat_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_relay_events_run   ON collab_relay_events(relay_run_id, sequence_num);
CREATE INDEX IF NOT EXISTS idx_relay_sections_run ON collab_relay_sections(relay_run_id, section_index);

-- fail-closed status CHECKs: unknown states are rejected, never silently accepted.
ALTER TABLE collab_relay_runs DROP CONSTRAINT IF EXISTS collab_relay_runs_status_check;
ALTER TABLE collab_relay_runs ADD CONSTRAINT collab_relay_runs_status_check
    CHECK (status IN ('running','awaiting_human_review','relay_failed','retrying','completed','failed'));

ALTER TABLE collab_relay_sections DROP CONSTRAINT IF EXISTS collab_relay_sections_status_check;
ALTER TABLE collab_relay_sections ADD CONSTRAINT collab_relay_sections_status_check
    CHECK (status IN ('pending','generating','awaiting_review','approved','failed'));
