-- ⚠️ YEH FILE GOLANG-MIGRATE KABHI CHALAYEGA HI NAHI (`.up.sql` suffix missing) ⚠️
-- Live version: migrations/080_receptionist_runtime_tables.up.sql
-- FK NOT NULL - tera purana bug: hawa me save ho raha tha, ab checkpoint_id mandatory
CREATE TABLE IF NOT EXISTS expert_responses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  checkpoint_id UUID NOT NULL REFERENCES receptionist_checkpoints(id) ON DELETE CASCADE,
  session_id UUID NOT NULL REFERENCES receptionist_sessions(id) ON DELETE CASCADE,
  expert_id UUID NOT NULL REFERENCES experts(id),
  tenant_id UUID NOT NULL,
  answer_md TEXT NOT NULL, -- English markdown full (Strong)
  hinglish_summary TEXT, -- Cheap synthesis 2 lines
  latency_ms INT NOT NULL,
  cost_usd NUMERIC(8,6),
  rating SMALLINT CHECK (rating BETWEEN 1 AND 5),
  status TEXT NOT NULL DEFAULT 'DONE' CHECK (status IN ('DONE','REOPENED','APPROVED')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_expert_responses_checkpoint ON expert_responses(checkpoint_id);
CREATE INDEX IF NOT EXISTS idx_expert_responses_session ON expert_responses(session_id);
CREATE INDEX IF NOT EXISTS idx_expert_responses_tenant ON expert_responses(tenant_id);

-- experts + course_chunks already hai - RAG ke liye
-- course_chunks(expert_id, chunk_text, embedding vector(1536), times_cited)
-- HNSW index already: CREATE INDEX ON course_chunks USING hnsw (embedding vector_cosine_ops);

COMMENT ON TABLE expert_responses IS 'FK NOT NULL fix. One row per expert per checkpoint. Rating per-card.';
