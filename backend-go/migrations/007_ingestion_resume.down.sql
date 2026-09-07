-- Reverse migration 007
ALTER TABLE ingestion_jobs
    DROP COLUMN IF EXISTS current_stage,
    DROP COLUMN IF EXISTS stage_detail,
    DROP COLUMN IF EXISTS cost_usd,
    DROP COLUMN IF EXISTS estimated_seconds_remaining,
    DROP COLUMN IF EXISTS checkpoint_data,
    DROP COLUMN IF EXISTS resumed_from_checkpoint;
