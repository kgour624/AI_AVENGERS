-- 020_change_requests.up.sql
--
-- Stores change requests raised from the workflow chat.
-- A change request is a client's free-text instruction to re-run the design
-- phases with a new goal. It is separate from approval_requests (which are
-- engine-driven gates) because it is client-initiated at any time.
--
-- status lifecycle:
--   pending   -> runner has not yet picked it up
--   running   -> runner is executing the redesign
--   completed -> redesign done, approval gate presented
--   cancelled -> client cancelled before runner picked it up

CREATE TABLE IF NOT EXISTS change_requests (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id       UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    client_id         UUID NOT NULL REFERENCES users(id),
    -- The chat and message that originated this request (for traceability).
    source_chat_id    UUID REFERENCES workflow_chats(id) ON DELETE SET NULL,
    source_message_id UUID REFERENCES workflow_chat_messages(id) ON DELETE SET NULL,
    -- The free-text goal the client typed.
    change_goal       TEXT NOT NULL,
    -- Which experts the runner decided are relevant (populated when running starts).
    relevant_expert_ids JSONB NOT NULL DEFAULT '[]',
    status            TEXT NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending','running','completed','cancelled')),
    -- The blackboard event id of the change_request event posted for this.
    blackboard_event_id UUID REFERENCES blackboard_events(id) ON DELETE SET NULL,
    requested_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_change_requests_workflow
    ON change_requests(workflow_id, status);
CREATE INDEX IF NOT EXISTS idx_change_requests_pending
    ON change_requests(workflow_id)
    WHERE status = 'pending';
