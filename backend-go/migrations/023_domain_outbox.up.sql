-- Migration 023: transactional outbox for cross-domain events
--
-- WHY outbox (not direct Redis/Kafka publish inside handlers):
--   Same DB transaction as the business write guarantees at-least-once
--   delivery without dual-write loss. Dispatcher (SKIP LOCKED) drains rows.
--   Pattern: message broker on RDBMS — no new infra in Stage 0.
--
-- WHY status + available_at:
--   pending → published | failed. available_at enables delayed retry.
--   FOR UPDATE SKIP LOCKED lets N worker replicas scale without a leader.

CREATE TABLE IF NOT EXISTS domain_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  VARCHAR(64)  NOT NULL,
    aggregate_id    UUID         NOT NULL,
    event_type      VARCHAR(128) NOT NULL,
    payload         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    status          VARCHAR(16)  NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'published', 'failed')),
    attempts        INT          NOT NULL DEFAULT 0,
    available_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ,
    last_error      TEXT
);

CREATE INDEX IF NOT EXISTS idx_domain_outbox_pending
    ON domain_outbox (available_at, created_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_domain_outbox_aggregate
    ON domain_outbox (aggregate_type, aggregate_id);
