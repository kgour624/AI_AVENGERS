DROP INDEX IF EXISTS idx_file_jobs_pick_composite;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS chunk_count;
ALTER TABLE file_jobs DROP COLUMN IF EXISTS verified;
DROP FUNCTION IF EXISTS check_file_chunks_contiguous();
