-- ⚠️ YEH FILE GOLANG-MIGRATE KABHI CHALAYEGA HI NAHI ⚠️
-- File-name me `.up.sql` suffix nahi hai, aur cmd/migrate/main.go `file://migrations`
-- source use karta hai — file-driver ka regex `<version>_<name>.<up|down>.sql` match
-- na hone par file chup-chaap SKIP ho jati hai (isliye ye table kabhi bana hi nahi tha).
-- Isi ka replacement ab migrations/080_receptionist_runtime_tables.up.sql me hai;
-- is file ko chhedna nahi hai (rename karne par ye 015 numbering ki wajah se 078 se
-- pehle chalegi aur fresh DB par FK error de degi).
-- Postgres Source of Truth - Redis kabhi truth nahi
CREATE TABLE IF NOT EXISTS conversation_events (
  id BIGSERIAL PRIMARY KEY,
  session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('live','typing','expert_wait','pressure','note','template_ready','phase_change')),
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_conversation_events_session_id_id ON conversation_events(session_id, id);
CREATE INDEX IF NOT EXISTS idx_conversation_events_tenant ON conversation_events(tenant_id);

-- Cold archive ke liye - 30 din baad S3 dumper uthayega (tiered_storage pattern)
-- Hot me sirf last 1 week, warm me 30 din
COMMENT ON TABLE conversation_events IS 'Append-only durable log for SSE replay. Redis only notifies, never stores.';
