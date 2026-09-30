-- Migration: 069_mcp_tool_definitions.up.sql
-- Purpose: Production Grade - Hardcoded AllowedTools (get_standards, ask_expert) ko DB driven banao
-- Format 100% compatible hai internal/mcp/tools.go -> Tool{name, description, input_schema} se
-- taki dusri functionality break na ho

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Table: mcp_tool_definitions
CREATE TABLE IF NOT EXISTS mcp_tool_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE CHECK (name ~ '^[a-z][a-z0-9_]{2,64}$'),
    display_name TEXT NOT NULL CHECK (char_length(display_name) >= 2 AND char_length(display_name) <= 100),
    description TEXT NOT NULL CHECK (char_length(description) >= 5 AND char_length(description) <= 500),
    input_schema JSONB NOT NULL DEFAULT '{"type":"object","properties":{},"required":[]}'::jsonb,
    handler_key TEXT NOT NULL DEFAULT 'generic' CHECK (handler_key ~ '^[a-z][a-z0-9_]{2,64}$'),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for production performance
CREATE INDEX IF NOT EXISTS idx_mcp_tools_active ON mcp_tool_definitions(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_mcp_tools_name ON mcp_tool_definitions(name);
CREATE INDEX IF NOT EXISTS idx_mcp_tools_created_at ON mcp_tool_definitions(created_at DESC);

-- Auto update updated_at on every UPDATE (production grade)
CREATE OR REPLACE FUNCTION set_mcp_tool_updated_at()
RETURNS TRIGGER AS $$BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_mcp_tool_definitions_updated_at ON mcp_tool_definitions;
CREATE TRIGGER trg_mcp_tool_definitions_updated_at
BEFORE UPDATE ON mcp_tool_definitions
FOR EACH ROW EXECUTE FUNCTION set_mcp_tool_updated_at();

-- Seed existing hardcoded tools taaki purana data aur existing tokens break na ho
-- Format exactly same hai jo internal/mcp/tools.go me hai
INSERT INTO mcp_tool_definitions (name, display_name, description, input_schema, handler_key, is_active) VALUES
('get_standards', 'Get Standards', 'Fetch domain standards and templates from Chinawall/Category', '{"type":"object","properties":{"domain":{"type":"string","description":"Domain name"}},"required":["domain"]}'::jsonb, 'get_standards', true),
('ask_expert', 'Ask Expert', 'Ask expert knowledge base via RAG answerer', '{"type":"object","properties":{"query":{"type":"string","description":"User query"},"expert_id":{"type":"string"}},"required":["query"]}'::jsonb, 'ask_expert', true)
ON CONFLICT (name) DO NOTHING;

-- Verification comment
-- SELECT * FROM mcp_tool_definitions;
