DROP VIEW IF EXISTS vacuum_job_stats;
DROP FUNCTION IF EXISTS check_file_chunks_contiguous();
DROP INDEX IF EXISTS idx_file_jobs_done_at;
DROP INDEX IF EXISTS idx_file_jobs_retry_after;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS error_code;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS retry_after;
