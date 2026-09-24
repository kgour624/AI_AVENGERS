-- Migration 037 DOWN: remove 'extracting' from the allowed current_stage values.
--
-- Any row still sitting at 'extracting' would violate the restored constraint,
-- so those rows are moved back to 'pending' first. That is the honest value:
-- 'pending' means "created, no stage completed", and no checkpoint exists for
-- an extraction that never finished, so resume restarts from the top.

UPDATE ingestion_jobs SET current_stage = 'pending' WHERE current_stage = 'extracting';

ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS ingestion_jobs_stage_check;
ALTER TABLE ingestion_jobs ADD CONSTRAINT ingestion_jobs_stage_check
    CHECK (current_stage IN (
        'pending',
        'chunking',
        'topic_extraction',
        'charter_extraction',
        'embedding',
        'storing',
        'smoke_test',
        'complete',
        'failed',
        'paused'
    ));
