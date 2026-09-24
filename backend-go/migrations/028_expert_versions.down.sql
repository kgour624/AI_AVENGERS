-- Rollback 028: drop C2 expert versioning tables (additive migration).
DROP TABLE IF EXISTS expert_drift_events;
DROP TABLE IF EXISTS expert_versions;
