-- Migration 019: Seed authoring_rank for existing expert categories.
--
-- authoring_rank was added in migration 018 (ALTER TABLE expert_categories
-- ADD COLUMN IF NOT EXISTS authoring_rank INTEGER) but left NULL for all rows.
-- NULL means "no opinion" — the planner decides order and same-domain detail
-- conflicts fall back to the client instead of auto-resolving.
--
-- This migration seeds sensible defaults based on dependency order:
-- upstream experts (who define contracts others depend on) get lower numbers
-- (higher priority). Testing is always downstream.
--
-- WHY these numbers:
--   System Design / Architecture = 1  (defines the whole system shape)
--   Product / PM                 = 1  (defines requirements everything else follows)
--   LLD / Database / Backend     = 2  (depend on system design)
--   Frontend / CSS / UI          = 3  (depend on backend contracts)
--   Testing / QA                 = 4  (depend on everything else)
--
-- These are HINTS, not overrides. The DAG still decides wave order.
-- A NULL rank is still valid — it means "no auto-resolve, go to client".
--
-- Idempotent: only updates rows where authoring_rank IS NULL, so a re-run
-- or an admin's manual change is never overwritten.

UPDATE expert_categories
SET authoring_rank = 1
WHERE authoring_rank IS NULL
  AND LOWER(TRIM(slug)) IN (
    'system-design', 'system_design', 'systemdesign',
    'architecture',
    'product', 'product-manager', 'product_manager', 'pm'
  );

UPDATE expert_categories
SET authoring_rank = 2
WHERE authoring_rank IS NULL
  AND LOWER(TRIM(slug)) IN (
    'lld', 'low-level-design', 'low_level_design',
    'database', 'db',
    'backend', 'go', 'java', 'python', 'nodejs', 'node',
    'api', 'microservices'
  );

UPDATE expert_categories
SET authoring_rank = 3
WHERE authoring_rank IS NULL
  AND LOWER(TRIM(slug)) IN (
    'frontend', 'react', 'vue', 'angular',
    'css', 'ui', 'ux', 'design-system'
  );

UPDATE expert_categories
SET authoring_rank = 4
WHERE authoring_rank IS NULL
  AND LOWER(TRIM(slug)) IN (
    'testing', 'qa', 'test',
    'java-tester', 'react-tester', 'js-tester',
    'jest', 'junit', 'cypress'
  );

-- Categories that did not match any slug above keep authoring_rank = NULL.
-- That is the safe default: no auto-resolve, conflict goes to the client.
