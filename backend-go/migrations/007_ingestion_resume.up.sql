-- ============================================================
-- Migration 007: Resume-capable ingestion job tracking
-- ============================================================
-- Adds fields to ingestion_jobs for:
-- 1. Granular stage tracking (which of 6 stages is running)
-- 2. Live cost accumulation (LLM calls during ingestion)
-- 3. ETA estimation (chunks/sec speed)
-- 4. Crash-recovery checkpoint (durable JSONB state)
-- ============================================================

ALTER TABLE ingestion_jobs
    ADD COLUMN IF NOT EXISTS current_stage              VARCHAR(30)    DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS stage_detail               TEXT           DEFAULT '',
    ADD COLUMN IF NOT EXISTS cost_usd                   DECIMAL(10,6)  DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS estimated_seconds_remaining INTEGER,
    ADD COLUMN IF NOT EXISTS checkpoint_data            JSONB          DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS resumed_from_checkpoint    BOOLEAN        DEFAULT FALSE;

-- CHECK constraint on stage values
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ingestion_jobs_stage_check'
    ) THEN
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
                'failed'
            ));
    END IF;
END $$;

COMMENT ON COLUMN ingestion_jobs.current_stage IS
    'Active pipeline stage. Used for UI progress display and crash recovery.';
COMMENT ON COLUMN ingestion_jobs.checkpoint_data IS
    'JSONB snapshot for crash recovery. Schema: {stage, chunks_done, charter_extracted, last_batch_index, cost_usd_so_far}. Written every 50 chunks.';
COMMENT ON COLUMN ingestion_jobs.cost_usd IS
    'Accumulated LLM cost for this ingestion job in USD.';
