-- ⚠️ YEH FILE GOLANG-MIGRATE KABHI CHALAYEGA HI NAHI (`.up.sql` suffix missing) ⚠️
-- Live version: migrations/080_receptionist_runtime_tables.up.sql
-- Notes already tha agar nahi to create, warna skip
CREATE TABLE IF NOT EXISTS receptionist_notes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  text TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notes_session ON receptionist_notes(session_id);

-- Final synthesis cache per session (1 row, upsert)
CREATE TABLE IF NOT EXISTS receptionist_final_synthesis (
  session_id UUID PRIMARY KEY REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  final_md TEXT NOT NULL, -- English merged markdown
  hinglish_summary TEXT NOT NULL,
  total_cost_usd NUMERIC(8,6),
  total_latency_ms INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- pressure_score already receptionist_sessions me hai
COMMENT ON TABLE receptionist_final_synthesis IS 'Phase 6 end pe Strong synthesis, 1 per session.';
