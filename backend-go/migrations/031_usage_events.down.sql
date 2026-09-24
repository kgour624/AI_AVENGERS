-- Rollback 031: remove C5 usage analytics (additive migration).
DROP TABLE IF EXISTS usage_budgets;
DROP TABLE IF EXISTS usage_events;
