ALTER TABLE file_jobs DROP COLUMN IF EXISTS llm_classifier_label;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS llm_heading_count;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS preservation_verified;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS download_count;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS last_downloaded_at;
ALTER TABLE file_chunks DROP COLUMN IF EXISTS llm_label;
ALTER TABLE file_chunks DROP COLUMN IF EXISTS llm_confidence;
DROP INDEX IF EXISTS idx_file_jobs_preservation;
DROP INDEX IF EXISTS idx_file_jobs_download;
DROP VIEW IF EXISTS vacuum_job_stats;
CREATE OR REPLACE VIEW vacuum_job_stats AS
SELECT status, COUNT(*)::bigint AS cnt, COALESCE(SUM(chunk_count),0)::bigint AS total_chunks, COALESCE(AVG(duration_ms),0)::bigint AS avg_duration_ms FROM file_jobs GROUP BY status;
