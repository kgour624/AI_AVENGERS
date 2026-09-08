-- Migration 009: Store transcript content in ingestion_jobs for resume support
--
-- WHY: If a job fails during chunking (stage < topic_extraction),
-- chunks are not yet in DB. Resume needs the original transcript.
-- Without this column, early-stage failures require re-upload.
ALTER TABLE ingestion_jobs
    ADD COLUMN IF NOT EXISTS transcript_content TEXT;

COMMENT ON COLUMN ingestion_jobs.transcript_content IS
    'Original transcript text. Stored at upload time for resume support.
     NULL for jobs created before migration 009.
     Used by ResumeIngestionJob when checkpoint stage < topic_extraction.';
