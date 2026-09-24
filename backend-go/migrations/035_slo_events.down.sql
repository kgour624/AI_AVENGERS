-- Migration 035 down: drop the C10 reliability event log.
DROP TABLE IF EXISTS slo_events;
