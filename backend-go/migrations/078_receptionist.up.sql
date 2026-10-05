CREATE TABLE IF NOT EXISTS receptionist_sessions (
    id UUID PRIMARY KEY,
    user_id UUID,
    tenant_id UUID,
    admin_id UUID,
    language TEXT NOT NULL DEFAULT 'EN',
    persona TEXT DEFAULT '',
    agenda TEXT DEFAULT '',
    state TEXT NOT NULL DEFAULT 'INIT',
    status TEXT NOT NULL DEFAULT 'INIT',
    cursor JSONB,
    current_cursor TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS receptionist_checkpoints (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    section_key TEXT,
    text TEXT,
    order_idx INT DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'PENDING',
    rating_history INT[] DEFAULT '{}',
    committed_summary TEXT,
    version INT NOT NULL DEFAULT 1,
    summary_warm JSONB,
    summary_hot_key TEXT,
    reopen_reason TEXT,
    reopened_at TIMESTAMPTZ,
    previous_summary_warm JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_receptionist_checkpoints_session_status ON receptionist_checkpoints(session_id, status);
CREATE INDEX IF NOT EXISTS idx_receptionist_checkpoints_section ON receptionist_checkpoints(session_id, section_key);

CREATE TABLE IF NOT EXISTS receptionist_notes (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    checkpoint_id UUID REFERENCES receptionist_checkpoints(id) ON DELETE SET NULL,
    type TEXT NOT NULL,
    summary TEXT NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS receptionist_expert_calls (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    checkpoint_id UUID REFERENCES receptionist_checkpoints(id) ON DELETE SET NULL,
    expert_id UUID NOT NULL,
    question TEXT NOT NULL,
    relevant_answer TEXT NOT NULL,
    rating INT CHECK (rating >=1 AND rating <=5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

