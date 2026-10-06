-- 084: Vacuum Brain Foundation (Phase 1)
CREATE TABLE IF NOT EXISTS kachra_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern TEXT NOT NULL,
    pattern_type TEXT NOT NULL DEFAULT 'PHRASE' CHECK (pattern_type IN ('WORD','PHRASE','REGEX','SEMANTIC')),
    category TEXT NOT NULL DEFAULT 'filler' CHECK (category IN ('filler','repetition','asr_error','classroom_meta','hinglish','logistics','semantic_noise')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    hit_count BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT kachra_pattern_not_empty CHECK (length(trim(pattern)) > 0)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_kachra_patterns_unique ON kachra_patterns(lower(trim(pattern)), pattern_type, category);
CREATE INDEX IF NOT EXISTS idx_kachra_patterns_active ON kachra_patterns(is_active) WHERE is_active=TRUE;
CREATE INDEX IF NOT EXISTS idx_kachra_patterns_version ON kachra_patterns(version);

CREATE TABLE IF NOT EXISTS candidate_kachra (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern TEXT NOT NULL,
    pattern_type TEXT NOT NULL DEFAULT 'PHRASE' CHECK (pattern_type IN ('WORD','PHRASE','REGEX','SEMANTIC')),
    category TEXT NOT NULL DEFAULT 'filler',
    context_snippet TEXT NOT NULL DEFAULT '',
    file_id UUID,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0.85 CHECK (confidence>=0 AND confidence<=1),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
    hit_count BIGINT NOT NULL DEFAULT 1,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_candidate_kachra_unique_pattern ON candidate_kachra(lower(trim(pattern)), pattern_type, category);
CREATE INDEX IF NOT EXISTS idx_candidate_kachra_status ON candidate_kachra(status);
CREATE INDEX IF NOT EXISTS idx_candidate_kachra_hit ON candidate_kachra(hit_count DESC);

CREATE TABLE IF NOT EXISTS file_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    s3_key TEXT NOT NULL,
    file_name TEXT NOT NULL DEFAULT '',
    file_size BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','picking','cleaning','verifying','done','failed','quarantined')),
    phase INT NOT NULL DEFAULT 0 CHECK (phase>=0 AND phase<=6),
    picked_at TIMESTAMPTZ, picked_by TEXT, attempts INT NOT NULL DEFAULT 0,
    last_error TEXT, sha256_input TEXT, sha256_output TEXT, s3_output_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_file_jobs_status ON file_jobs(status);
CREATE INDEX IF NOT EXISTS idx_file_jobs_status_picked ON file_jobs(status, picked_at);
CREATE INDEX IF NOT EXISTS idx_file_jobs_created ON file_jobs(created_at);

CREATE TABLE IF NOT EXISTS file_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_job_id UUID NOT NULL REFERENCES file_jobs(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL, char_start BIGINT NOT NULL DEFAULT 0, char_end BIGINT NOT NULL DEFAULT 0,
    hash_input TEXT NOT NULL DEFAULT '', hash_output TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','cleaned','verified','failed')),
    headings JSONB NOT NULL DEFAULT '[]', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(file_job_id, chunk_index)
);
CREATE INDEX IF NOT EXISTS idx_file_chunks_job ON file_chunks(file_job_id);

CREATE TABLE IF NOT EXISTS brain_version (id INT PRIMARY KEY DEFAULT 1 CHECK (id=1), version BIGINT NOT NULL DEFAULT 1, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
INSERT INTO brain_version(id, version) VALUES (1,1) ON CONFLICT (id) DO NOTHING;

CREATE OR REPLACE FUNCTION bump_brain_version() RETURNS TRIGGER AS $$ BEGIN UPDATE brain_version SET version=version+1, updated_at=NOW() WHERE id=1; RETURN NEW; END; $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_kachra_bump ON kachra_patterns;
CREATE TRIGGER trg_kachra_bump AFTER INSERT OR UPDATE OR DELETE ON kachra_patterns FOR EACH ROW EXECUTE FUNCTION bump_brain_version();
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at=NOW(); RETURN NEW; END; $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_kachra_updated ON kachra_patterns; CREATE TRIGGER trg_kachra_updated BEFORE UPDATE ON kachra_patterns FOR EACH ROW EXECUTE FUNCTION set_updated_at();
DROP TRIGGER IF EXISTS trg_candidate_updated ON candidate_kachra; CREATE TRIGGER trg_candidate_updated BEFORE UPDATE ON candidate_kachra FOR EACH ROW EXECUTE FUNCTION set_updated_at();
DROP TRIGGER IF EXISTS trg_file_jobs_updated ON file_jobs; CREATE TRIGGER trg_file_jobs_updated BEFORE UPDATE ON file_jobs FOR EACH ROW EXECUTE FUNCTION set_updated_at();
