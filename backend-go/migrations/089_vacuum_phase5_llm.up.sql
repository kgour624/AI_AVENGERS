-- 089 Vacuum Phase 5: LLM Intelligence Layer (Harness > Prompt)
-- I->P->O: Input raw transcript -> Process Tiny Classifier (Gemini Flash temp 0.1) + Heading Gen (Claude Sonnet ## ### only) + Preservation Guard (SHA) -> Output verified headings + preserved cleaned S3
-- No Trust Policy: LLM only returns headings, content jo original S3 se joda — hash mismatch => FAIL
ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS llm_classifier_label TEXT;
ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS llm_heading_count INT NOT NULL DEFAULT 0;
ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS preservation_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS download_count INT NOT NULL DEFAULT 0;
ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS last_downloaded_at TIMESTAMPTZ;
ALTER TABLE file_chunks ADD COLUMN IF NOT EXISTS llm_label TEXT;
ALTER TABLE file_chunks ADD COLUMN IF NOT EXISTS llm_confidence DOUBLE PRECISION;
CREATE INDEX IF NOT EXISTS idx_file_jobs_preservation ON file_jobs(preservation_verified) WHERE status='done';
CREATE INDEX IF NOT EXISTS idx_file_jobs_download ON file_jobs(download_count) WHERE status='done';
DROP VIEW IF EXISTS vacuum_job_stats;
CREATE OR REPLACE VIEW vacuum_job_stats AS
SELECT status,
       COUNT(*)::bigint AS cnt,
       COALESCE(SUM(chunk_count),0)::bigint AS total_chunks,
       COALESCE(AVG(duration_ms),0)::bigint AS avg_duration_ms,
       COALESCE(SUM(download_count),0)::bigint AS total_downloads,
       COALESCE(SUM(CASE WHEN preservation_verified THEN 1 ELSE 0 END),0)::bigint AS preserved_cnt
FROM file_jobs GROUP BY status;
