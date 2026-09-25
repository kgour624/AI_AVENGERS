-- Migration 048 DOWN: remove durable depth-classification job state.
--
-- NOTE: an in-flight classifier cannot survive dropping its status table. Stop the
-- backend/job workers before rolling this migration back. Existing chunk.layer values
-- are untouched; only job progress/history is removed.

DROP INDEX IF EXISTS idx_depth_classification_jobs_latest;
DROP INDEX IF EXISTS expert_depth_classification_jobs_one_active;
DROP TABLE IF EXISTS expert_depth_classification_jobs;
