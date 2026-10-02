-- 073_mcp_v2_expert_id_fix.up.sql
-- Production-grade fix: list_experts -> ask_expert expert_id roundtrip contract
-- WHY: Claude's mcp_AI-AVENGERS_ask_expert sends prefixed name + no stable ID → 5 tools failed (INVALID_ARGS/UNKNOWN_EXPERT).
-- list_experts now returns id:uuid, handler normalizes prefix (mcp_AI-AVENGERS_ask_expert -> ask_expert) and requires expert_id.
-- This migration hardens mcp_v2_tools.input_schema to require expert_id as stable UUID from list_experts (slug/name stays as alias via GetExpert fallback).
-- Idempotent: safe to re-run. No structural change to mcp_tokens (done in 072).

-- ask_expert: expert_id required (stable UUID from list_experts, copy exactly)
UPDATE mcp_v2_tools SET
  input_schema = '{
    "type": "object",
    "properties": {
      "expert_id": {"type": "string", "description": "Stable expert UUID from list_experts — preferred, copy exactly. Also accepts slug/name as alias"},
      "question": {"type": "string", "description": "User question to answer with grounded context"},
      "platform": {"type": "string", "description": "Calling platform", "default": "generic"},
      "tier": {"type": "string", "description": "LLM tier", "default": "strong"},
      "limit": {"type": "integer", "description": "Number of chunks to retrieve", "default": 5}
    },
    "required": ["expert_id", "question"],
    "additionalProperties": false
  }'::jsonb,
  description = 'Ask a domain expert a grounded question (RAG) — use id from list_experts'
WHERE name = 'ask_expert';

-- get_standards: expert_id required
UPDATE mcp_v2_tools SET
  input_schema = '{
    "type": "object",
    "properties": {
      "expert_id": {"type": "string", "description": "Stable expert UUID from list_experts — preferred, copy exactly. Also accepts slug/name as alias"},
      "track": {"type": "string", "description": "Standards track", "default": "general"}
    },
    "required": ["expert_id"],
    "additionalProperties": false
  }'::jsonb
WHERE name = 'get_standards';

-- review_change: expert_id + diff required
UPDATE mcp_v2_tools SET
  input_schema = '{
    "type": "object",
    "properties": {
      "expert_id": {"type": "string", "description": "Stable expert UUID from list_experts — preferred, copy exactly. Also accepts slug/name as alias"},
      "diff": {"type": "string", "description": "Unified diff or code change to review"},
      "change_description": {"type": "string", "description": "Human description of the change intent"},
      "file_path": {"type": "string", "description": "Optional file path context"}
    },
    "required": ["expert_id", "diff"],
    "additionalProperties": false
  }'::jsonb
WHERE name = 'review_change';

-- search_course_content / search_repo_content / search_chunks: expert_id + query required (alias for handler)
UPDATE mcp_v2_tools SET
  input_schema = '{
    "type": "object",
    "properties": {
      "expert_id": {"type": "string", "description": "Stable expert UUID from list_experts — preferred, copy exactly. Also accepts slug/name as alias"},
      "query": {"type": "string", "description": "Search query"},
      "limit": {"type": "integer", "description": "Max results", "default": 20},
      "cursor": {"type": "string", "description": "Pagination cursor"}
    },
    "required": ["expert_id", "query"],
    "additionalProperties": false
  }'::jsonb
WHERE name IN ('search_course_content', 'search_repo_content', 'search_chunks');

-- list_experts: documents that it returns id for roundtrip
UPDATE mcp_v2_tools SET
  input_schema = '{
    "type": "object",
    "properties": {
      "category": {"type": "string", "description": "Optional category filter"},
      "limit": {"type": "integer", "description": "Max experts to return", "default": 20},
      "cursor": {"type": "string", "description": "Pagination cursor"}
    },
    "required": [],
    "additionalProperties": false
  }'::jsonb,
  description = 'List all active domain experts — returns id (UUID) to use as expert_id in ask_expert/search/get_standards'
WHERE name = 'list_experts';

-- Ensure missing search_chunks tool exists (handler aliases it)
INSERT INTO mcp_v2_tools (name, display_name, description, input_schema) VALUES
('search_chunks', 'Search Chunks', 'Semantic search over expert chunks — alias for course/repo search, uses expert_id',
  '{"type":"object","properties":{"expert_id":{"type":"string"},"query":{"type":"string"},"limit":{"type":"integer","default":20}},"required":["expert_id","query"],"additionalProperties":false}'::jsonb)
ON CONFLICT (name) DO NOTHING;