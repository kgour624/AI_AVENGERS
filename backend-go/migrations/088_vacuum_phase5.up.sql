-- 088 Vacuum Phase 5: Scale & Observability — Input scheduled/failed retryable -> Process concurrent DTS + metrics + contiguous hard verify -> Output done/finance observable pagination
-- Mental I->P->O: Input= file_jobs scheduled with retry_after filter, Process= concurrent PickAndExecute (sem 5) + duration_ms/finished_at + atomic metrics, Output= paginated jobs + avg duration + contiguous verified chunks

ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ;
ALTER TABLE file_jobs ADD COLUMN IF NOT EXISTS duration_ms INT;

-- Retry-aware picker index: scheduled with retry_after tracking (no NOW() in predicate — NOW() immutable issue)
CREATE INDEX IF NOT EXISTS idx_file_jobs_pick_sched_retry ON file_jobs(status, retry_after, created_at);
CREATE INDEX IF NOT EXISTS idx_file_jobs_failed_retry ON file_jobs(retry_after) WHERE status='failed';

-- Done observability composite
CREATE INDEX IF NOT EXISTS idx_file_jobs_done_finished ON file_jobs(finished_at DESC) WHERE status='done';

-- Ensure chunk ordering still contiguous via same function (pipeline hard verifies 0..n)
-- No new trigger; pipeline verifies max_idx == count-1 and char ranges contiguous at commit

-- Stats view now includes avg duration
DROP VIEW IF EXISTS vacuum_job_stats;
CREATE OR REPLACE VIEW vacuum_job_stats AS
SELECT status,
       COUNT(*)::bigint AS cnt,
       COALESCE(SUM(chunk_count),0)::bigint AS total_chunks,
       COALESCE(AVG(duration_ms),0)::bigint AS avg_duration_ms
FROM file_jobs GROUP BY status;
