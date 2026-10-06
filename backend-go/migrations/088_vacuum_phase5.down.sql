DROP VIEW IF EXISTS vacuum_job_stats;
CREATE OR REPLACE VIEW vacuum_job_stats AS SELECT status, COUNT(*)::bigint AS cnt, COALESCE(SUM(chunk_count),0)::bigint AS total_chunks FROM file_jobs GROUP BY status;
DROP INDEX IF EXISTS idx_file_jobs_done_finished;
DROP INDEX IF EXISTS idx_file_jobs_failed_retry;
DROP INDEX IF EXISTS idx_file_jobs_pick_sched_retry;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS duration_ms;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS finished_at;
