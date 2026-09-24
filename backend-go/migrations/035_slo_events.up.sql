-- Migration 035: audit-grade reliability event log (C10)
--
-- WHY: reliability work was internal — no published targets and no durable
-- record of degradation/outage transitions. C10 turns it into a product:
-- SLO targets + a /status surface, backed by an append-only event log so an
-- outage (dependency down, error budget breached, recovery) is auditable
-- after the fact rather than only visible live on a dashboard.
--
-- Additive only. Rollback = migration 035 down.

CREATE TABLE IF NOT EXISTS slo_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind       VARCHAR(40) NOT NULL,
    severity   VARCHAR(10) NOT NULL DEFAULT 'info',
    component  VARCHAR(60),
    detail     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT slo_events_severity_check CHECK (
        severity IN ('info', 'warning', 'critical')
    )
);

CREATE INDEX IF NOT EXISTS idx_slo_events_created ON slo_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_slo_events_kind ON slo_events(kind, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_slo_events_component ON slo_events(component, created_at DESC);

COMMENT ON TABLE slo_events IS
    'C10: append-only, audit-grade reliability event log (degraded/recovered/budget_breach/...).';
COMMENT ON COLUMN slo_events.kind IS
    'Event kind: degraded | recovered | budget_breach | provider_error | ingest_failed | rate_limited | ...';
