-- 071_mcp_v2_schema.down.sql — rollback dynamic schema column
ALTER TABLE mcp_v2_tools DROP COLUMN IF EXISTS input_schema;