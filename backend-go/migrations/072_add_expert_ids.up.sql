-- 072_add_expert_ids: granular expert-level MCP token scope
-- WHY: broad domain restriction exposes all experts in a domain.
-- expert_ids allows a token to reach only specific experts (e.g. "High Level Design SLR").

ALTER TABLE mcp_tokens ADD COLUMN IF NOT EXISTS expert_ids TEXT[] NOT NULL DEFAULT '{}';

-- GIN index for efficiently checking token scope on expert_id
CREATE INDEX IF NOT EXISTS idx_mcp_tokens_expert_ids ON mcp_tokens USING GIN (expert_ids);