-- Reverse migration 025: drop per-domain Gate 1 threshold config.
DROP TABLE IF EXISTS gate_thresholds CASCADE;
