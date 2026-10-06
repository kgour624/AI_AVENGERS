-- Rollback 082: drop compat view + extra index only (TABLE expert_pages from 081 stays until 081 rollback)
DROP VIEW IF EXISTS expert_pages_view;
DROP INDEX IF EXISTS idx_expert_pages_expert_page;
