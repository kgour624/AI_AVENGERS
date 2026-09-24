-- Migration 036 DOWN: drop the ingestion job event log (T1 rollback).
-- Non-destructive to ingestion itself: the pipeline falls back to the
-- ingestion_jobs snapshot if this table is absent (see jobevents.Store).

DROP TABLE IF EXISTS ingestion_job_events;
