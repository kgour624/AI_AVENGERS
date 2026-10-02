-- 071_mcp_v2_schema.up.sql
-- mcp-v2: 100% Dynamic Tool Management — add JSONB input_schema for tools/list JSON-RPC

ALTER TABLE mcp_v2_tools
  ADD COLUMN IF NOT EXISTS input_schema JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE mcp_v2_tools SET input_schema = '{
  "type": "object",
  "properties": {
    "category": {"type": "string", "description": "Optional category filter"},
    "limit": {"type": "integer", "description": "Max experts to return", "default": 20},
    "cursor": {"type": "string", "description": "Pagination cursor"}
  },
  "required": [],
  "additionalProperties": false
}'::jsonb WHERE name = 'list_experts';

UPDATE mcp_v2_tools SET input_schema = '{
  "type": "object",
  "properties": {
    "expert_id": {"type": "string", "description": "Expert ID or slug"},
    "track": {"type": "string", "description": "Standards track"}
  },
  "required": ["expert_id"],
  "additionalProperties": false
}'::jsonb WHERE name = 'get_standards';

UPDATE mcp_v2_tools SET input_schema = '{
  "type": "object",
  "properties": {
    "expert_id": {"type": "string", "description": "Expert ID or slug to query"},
    "question": {"type": "string", "description": "User question to answer with grounded context"},
    "platform": {"type": "string", "description": "Calling platform", "default": "generic"},
    "limit": {"type": "integer", "description": "Number of chunks to retrieve", "default": 5}
  },
  "required": ["expert_id", "question"],
  "additionalProperties": false
}'::jsonb WHERE name = 'ask_expert';

UPDATE mcp_v2_tools SET input_schema = '{
  "type": "object",
  "properties": {
    "expert_id": {"type": "string", "description": "Expert ID or slug"},
    "diff": {"type": "string", "description": "Unified diff or code change to review"},
    "change_description": {"type": "string", "description": "Human description of the change intent"},
    "file_path": {"type": "string", "description": "Optional file path context"}
  },
  "required": ["expert_id", "diff"],
  "additionalProperties": false
}'::jsonb WHERE name = 'review_change';

INSERT INTO mcp_v2_tools (name, display_name, description, input_schema) VALUES
('get_standards', 'Get Standards', 'Fetch expert coding and quality standards', '{"type":"object","properties":{"expert_id":{"type":"string"},"track":{"type":"string"}},"required":["expert_id"],"additionalProperties":false}'::jsonb),
('ask_expert', 'Ask Expert', 'Ask a domain expert a grounded question (RAG)', '{"type":"object","properties":{"expert_id":{"type":"string"},"question":{"type":"string"},"platform":{"type":"string","default":"generic"},"limit":{"type":"integer","default":5}},"required":["expert_id","question"],"additionalProperties":false}'::jsonb),
('review_change', 'Review Change', 'Review a code diff against expert standards', '{"type":"object","properties":{"expert_id":{"type":"string"},"diff":{"type":"string"},"change_description":{"type":"string"},"file_path":{"type":"string"}},"required":["expert_id","diff"],"additionalProperties":false}'::jsonb)
ON CONFLICT (name) DO NOTHING;

UPDATE mcp_v2_tools SET input_schema = '{"type":"object","properties":{"expert_id":{"type":"string"},"query":{"type":"string"},"limit":{"type":"integer","default":5}},"required":["expert_id","query"],"additionalProperties":false}'::jsonb WHERE name IN ('search_course_content', 'search_repo_content') AND input_schema = '{}'::jsonb;