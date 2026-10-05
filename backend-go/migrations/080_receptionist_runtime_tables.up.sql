-- Migration 080: receptionist ke woh tables jo ab tak kabhi apply hi nahi huye.
--
-- ASLI BUG (compile ke baad yeh runtime par phatta hai):
--   migrations/000015_receptionist_events.sql
--   migrations/000016_receptionist_templates.sql
--   migrations/000017_expert_responses_fix.sql
--   migrations/000018_polish_end.sql
-- in chaaron files ke naam me golang-migrate ka zaroori `.up.sql` suffix nahi hai.
-- cmd/migrate/main.go `file://migrations` source use karta hai, aur file-driver ka
-- regex hai: <version>_<name>.<up|down>.sql — jo match na kare woh chup-chaap SKIP
-- ho jata hai (error bhi nahi deta). Isliye ye tables kabhi bane hi nahi, aur
-- internal/receptionist ke store methods runtime par
--   ERROR: relation "conversation_events" does not exist
--   ERROR: relation "receptionist_templates" does not exist
--   ERROR: relation "expert_responses" does not exist
--   ERROR: relation "receptionist_final_synthesis" does not exist
-- de rahe the. Yahi migration woh gap bharta hai (sab idempotent, safe re-run).

-- ---------------------------------------------------------------------------
-- 1. conversation_events - AppendEvent / ListEvents ka durable log (SSE replay)
--    src: store_events.go (INSERT ... RETURNING id, session_id, tenant_id, type, payload, created_at)
--
--    NOTE: purane 000015 wala CHECK (type IN ('live','typing',...)) jaan-boojh kar
--    NAHI lagaya — handler "note", "template_ready", "phase_change" jaise types
--    bhi append karta hai, aur us CHECK ke saath har note insert CHECK violation
--    deta tha. Type validation ab application layer par hai.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS conversation_events (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_events_session_id_id ON conversation_events(session_id, id);
CREATE INDEX IF NOT EXISTS idx_conversation_events_tenant ON conversation_events(tenant_id);

COMMENT ON TABLE conversation_events IS 'Append-only durable log for SSE replay. Redis only notifies, never stores.';

-- ---------------------------------------------------------------------------
-- 2. receptionist_templates - per-checkpoint dynamic template (1:1)
--    src: store_template.go (INSERT ... (checkpoint_id, session_id, idx, points, search_query, tenant_id)
--                           ON CONFLICT (checkpoint_id) DO UPDATE SET points, search_query, cached_at)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS receptionist_templates (
    checkpoint_id UUID PRIMARY KEY REFERENCES receptionist_checkpoints(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    idx INT NOT NULL DEFAULT 0,
    points JSONB NOT NULL DEFAULT '[]',
    search_query TEXT NOT NULL DEFAULT '',
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id UUID
);

CREATE INDEX IF NOT EXISTS idx_templates_session_idx ON receptionist_templates(session_id, idx);
CREATE INDEX IF NOT EXISTS idx_templates_tenant ON receptionist_templates(tenant_id);

COMMENT ON TABLE receptionist_templates IS 'Dynamic search result per checkpoint. Next checkpoint pe DELETE old checkpoint_id. Redis only cache, Postgres truth.';

-- ---------------------------------------------------------------------------
-- 3. expert_responses - one row per expert per checkpoint
--    src: store_expert.go (INSERT ... (id, checkpoint_id, session_id, expert_id, tenant_id,
--                                     answer_md, hinglish_summary, latency_ms, cost_usd, status))
--
--    NOTE: 078 ne iske bajaye receptionist_expert_calls banaya tha, par Go store
--    expert_responses hi use karta hai — isliye canonical table yahi rahega.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS expert_responses (
    id UUID PRIMARY KEY,
    checkpoint_id UUID NOT NULL REFERENCES receptionist_checkpoints(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    expert_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    answer_md TEXT NOT NULL,
    hinglish_summary TEXT,
    latency_ms INT NOT NULL DEFAULT 0,
    cost_usd NUMERIC(8,6),
    rating SMALLINT CHECK (rating BETWEEN 1 AND 5),
    status TEXT NOT NULL DEFAULT 'DONE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expert_responses_checkpoint ON expert_responses(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_expert_responses_session ON expert_responses(session_id);
CREATE INDEX IF NOT EXISTS idx_expert_responses_tenant ON expert_responses(tenant_id);

COMMENT ON TABLE expert_responses IS 'FK NOT NULL fix. One row per expert per checkpoint. Rating per-card.';

-- ---------------------------------------------------------------------------
-- 4. receptionist_final_synthesis - Phase 6 ka single cached row per session
--    src: store_polish.go (INSERT ... ON CONFLICT (session_id) DO UPDATE SET
--                          final_md, hinglish_summary, total_cost_usd, total_latency_ms, created_at=now())
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS receptionist_final_synthesis (
    session_id UUID PRIMARY KEY REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    final_md TEXT NOT NULL,
    hinglish_summary TEXT NOT NULL,
    total_cost_usd NUMERIC(8,6),
    total_latency_ms INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE receptionist_final_synthesis IS 'Phase 6 end pe Strong synthesis, 1 per session.';

-- ---------------------------------------------------------------------------
-- 5. receptionist_notes - column mismatch fix.
--
--    078 ne notes banaya tha shape: (id, session_id, checkpoint_id, type NOT NULL,
--    summary NOT NULL, ts) — par Go store (store_polish.go) ka shape hai:
--    (session_id, tenant_id, text, created_at). Isliye AddNote hamesha
--      ERROR: column "tenant_id" of relation "receptionist_notes" does not exist
--    deta tha. Yahan dono shapes reconcile kiye gaye hain: Go store ke columns add
--    kiye gaye aur purane NOT NULL constraints dheele kiye gaye.
-- ---------------------------------------------------------------------------
ALTER TABLE receptionist_notes ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE receptionist_notes ADD COLUMN IF NOT EXISTS text TEXT;
ALTER TABLE receptionist_notes ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE receptionist_notes ALTER COLUMN type DROP NOT NULL;
ALTER TABLE receptionist_notes ALTER COLUMN summary DROP NOT NULL;

-- Purane (078-shape) rows ka data naye column me shift karo. "summary"/"ts" sirf
-- tab exist karte hain jab 078 chala ho, isliye guard lagaya gaya — warna fresh
-- DB par ye statement khud fail ho jati.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'receptionist_notes' AND column_name = 'summary'
    ) THEN
        UPDATE receptionist_notes SET text = COALESCE(text, summary) WHERE text IS NULL;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'receptionist_notes' AND column_name = 'ts'
    ) THEN
        UPDATE receptionist_notes SET created_at = COALESCE(created_at, ts) WHERE created_at IS NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_notes_session_created ON receptionist_notes(session_id, created_at);
