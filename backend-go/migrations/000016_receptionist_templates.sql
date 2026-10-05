-- ⚠️ YEH FILE GOLANG-MIGRATE KABHI CHALAYEGA HI NAHI (`.up.sql` suffix missing) ⚠️
-- Live version: migrations/080_receptionist_runtime_tables.up.sql
-- Per-checkpoint dynamic template - ek checkpoint = ek template (1:1)
-- Next checkpoint pe purana clear karna hai (TRUNCATE pattern)
CREATE TABLE IF NOT EXISTS receptionist_templates (
  checkpoint_id UUID PRIMARY KEY REFERENCES receptionist_checkpoints(id) ON DELETE CASCADE,
  session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
  idx INT NOT NULL, -- checkpoint idx copy for fast clear
  points JSONB NOT NULL, -- [{id,q_text,status:draft/discussing/approved/dismissed, source:search, citation}]
  search_query TEXT NOT NULL,
  cached_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  tenant_id UUID NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_templates_session_idx ON receptionist_templates(session_id, idx);
CREATE INDEX IF NOT EXISTS idx_templates_tenant ON receptionist_templates(tenant_id);

COMMENT ON TABLE receptionist_templates IS 'Dynamic search result per checkpoint. Next checkpoint pe DELETE old checkpoint_id. Redis only cache, Postgres truth.';
