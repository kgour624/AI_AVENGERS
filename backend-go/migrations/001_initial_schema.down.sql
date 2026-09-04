-- ============================================================
-- Migration 001: Rollback
-- Drops all tables in reverse dependency order
-- ============================================================

DROP TABLE IF EXISTS system_settings CASCADE;
DROP TABLE IF EXISTS ingestion_jobs CASCADE;
DROP TABLE IF EXISTS repo_connections CASCADE;
DROP TABLE IF EXISTS ratings CASCADE;
DROP TABLE IF EXISTS master_event_log CASCADE;
DROP TABLE IF EXISTS project_memory_l2 CASCADE;
DROP TABLE IF EXISTS chat_summaries CASCADE;
DROP TABLE IF EXISTS chat_index CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS project_experts CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS expert_capabilities CASCADE;
DROP TABLE IF EXISTS course_chunks CASCADE;
DROP TABLE IF EXISTS experts CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP EXTENSION IF EXISTS vector;
DROP EXTENSION IF EXISTS "uuid-ossp";
