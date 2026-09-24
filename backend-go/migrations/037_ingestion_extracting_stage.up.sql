-- Migration 037: allow 'extracting' as a current_stage (document ingestion, D2)
--
-- WHY: uploaded files are no longer assumed to be plain text. An upload can now
-- be a PDF/DOCX/XLSX/PPTX/... that must be converted to text before chunking
-- can start. That conversion is a real, visible step — it can take seconds and
-- it can fail (encrypted PDF, scanned PDF with no text layer, corrupt file) —
-- so the admin needs to see it in the job row, not just in the event timeline.
--
-- The CHECK constraint from migration 011 enumerates the legal stage values,
-- so 'extracting' has to be added here or every UPDATE that sets it is rejected
-- and the job silently stays at 'pending'.
--
-- Ordering note: 'extracting' runs BEFORE 'chunking'. Nothing else changes —
-- checkpoint_data.stage (the resume source of truth) never stores 'extracting',
-- because no checkpoint is written during extraction.
--
-- Additive and reversible: migration 037 down restores the 011 constraint.

ALTER TABLE ingestion_jobs DROP CONSTRAINT IF EXISTS ingestion_jobs_stage_check;
ALTER TABLE ingestion_jobs ADD CONSTRAINT ingestion_jobs_stage_check
    CHECK (current_stage IN (
        'pending',
        'extracting',
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
