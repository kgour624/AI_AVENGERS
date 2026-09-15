-- Migration 012 DOWN: Remove updated_at from ingestion_jobs
ALTER TABLE ingestion_jobs DROP COLUMN IF EXISTS updated_at;
