-- MCP access tokens and audit trail.
--
-- WHY tokens live here and not in the admin API: an MCP client (Claude Code)
-- runs on a developer's machine and must reach the experts without an admin
-- session or password. A token is scoped to exactly what it may do (domains,
-- tools), is revocable, and never exposes the account behind it.
--
-- WHY the token itself is not stored: only its SHA-256 hash is kept, so a
-- database leak does not hand somebody a working credential. The plaintext is
-- shown once, at creation.
CREATE TABLE IF NOT EXISTS mcp_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    label        TEXT NOT NULL,
    -- Empty array = every domain / every tool. Keeping the allow-lists as
    -- arrays means adding a scope needs no schema change.
    domains      TEXT[] NOT NULL DEFAULT '{}',
    tools        TEXT[] NOT NULL DEFAULT '{}',
    -- Soft revoke: the row stays so the audit trail keeps a readable history.
    revoked_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    request_count BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mcp_tokens_user ON mcp_tokens(user_id);

-- One row per tool call. Append-only by design: this is the record of what an
-- external agent asked the experts, and it must not be editable.
--
-- input_bytes + input_sha256 instead of the raw payload: whoever asks the
-- question should be able to see THAT it was asked and prove WHICH question it
-- was, without the product retaining the client's code.
CREATE TABLE IF NOT EXISTS mcp_audit (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_id     UUID REFERENCES mcp_tokens(id) ON DELETE SET NULL,
    token_label  TEXT NOT NULL DEFAULT '',
    tool         TEXT NOT NULL,
    domain       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL,
    error_code   TEXT NOT NULL DEFAULT '',
    latency_ms   INTEGER NOT NULL DEFAULT 0,
    input_bytes  INTEGER NOT NULL DEFAULT 0,
    input_sha256 TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mcp_audit_token_created ON mcp_audit(token_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mcp_audit_created ON mcp_audit(created_at DESC);
