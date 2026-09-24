-- Migration 027: durable provenance chain for answers + artifacts (C1)
--
-- WHY: citations exist on messages/blackboard rows but there is no
-- signed, queryable chain tying an output to its sources (transcript/
-- chunk), the expert that produced it, the gates that fired, and the
-- model used. C1 (§3.1 P9) reuses the B8 span anchors as the provenance
-- primitive and stores one immutable chain per output.
--
-- Additive only. Rollback = run the down migration (drops the table).
--
-- signature: HMAC-SHA256 over chain (canonical JSON), hex-encoded.
-- chain    : canonical JSON of the whole record; the signed bytes.
-- content_hash: sha256 of the shipped output text (tamper check).
-- UNIQUE(output_type, output_id): one chain per output; re-writes are
-- rejected (append-only provenance, never silently rewritten).

CREATE TABLE IF NOT EXISTS provenance_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    output_type     VARCHAR(30) NOT NULL,
    output_id       UUID NOT NULL,
    project_id      UUID REFERENCES projects(id)  ON DELETE SET NULL,
    expert_id       UUID REFERENCES experts(id)   ON DELETE SET NULL,
    chat_id         UUID REFERENCES chats(id)     ON DELETE SET NULL,
    workflow_id     UUID REFERENCES workflows(id) ON DELETE SET NULL,
    model           VARCHAR(100),
    decision_mode   VARCHAR(20),
    gate_stopped    INTEGER,
    content_hash    CHAR(64) NOT NULL,
    claims          JSONB NOT NULL DEFAULT '[]'::jsonb,
    citations       JSONB NOT NULL DEFAULT '[]'::jsonb,
    -- chain is TEXT (not JSONB) so the exact signed bytes are preserved.
    -- JSONB normalizes key order/whitespace, which would break HMAC verify.
    -- Querying uses the structured columns above; chain is the signed doc.
    chain           TEXT NOT NULL,
    signature       TEXT NOT NULL,
    signing_key_id  VARCHAR(20) NOT NULL DEFAULT 'v1',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT provenance_output_type_check CHECK (
        output_type IN ('chat_message', 'workflow_artifact')
    ),
    CONSTRAINT provenance_unique UNIQUE (output_type, output_id)
);

CREATE INDEX IF NOT EXISTS idx_provenance_project
    ON provenance_records(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_provenance_expert
    ON provenance_records(expert_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_provenance_workflow
    ON provenance_records(workflow_id, created_at DESC)
    WHERE workflow_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_provenance_output
    ON provenance_records(output_type, output_id);

COMMENT ON TABLE provenance_records IS
    'C1: immutable signed provenance chain per answer/artifact. Append-only; UNIQUE(output_type,output_id).';
COMMENT ON COLUMN provenance_records.signature IS
    'HMAC-SHA256(signing_key, chain) hex. Verifies the chain was not tampered with.';
COMMENT ON COLUMN provenance_records.chain IS
    'Canonical JSON text of the signed provenance chain (exact signed bytes; source of truth for verification).';
