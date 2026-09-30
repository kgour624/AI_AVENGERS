-- 068: MCP analytics views (Complete Source of Truth)
CREATE OR REPLACE VIEW v_mcp_daily_cost AS
SELECT date_trunc('day', created_at)::date AS day,
       expert_id, platform, provider, model, tier,
       COUNT(*)::int AS calls,
       SUM(total_tokens)::bigint AS total_tokens,
       SUM(CASE WHEN generic_used THEN 1 ELSE 0 END)::int AS generic_calls,
       SUM(cost_usd)::numeric(12,6) AS cost_usd,
       AVG(duration_ms)::int AS avg_duration_ms
FROM mcp_usage_log GROUP BY 1,2,3,4,5,6;

CREATE OR REPLACE VIEW v_mcp_rental_revenue AS
SELECT r.expert_id, r.plan_id, COUNT(*)::int AS rentals, SUM(r.usage_cost)::numeric(12,6) AS revenue
FROM mcp_rentals r WHERE r.status='active' GROUP BY 1,2;

CREATE OR REPLACE VIEW v_mcp_generic_audit AS
SELECT id, expert_id, platform, generic_percent, ''::text AS reason, created_at FROM mcp_usage_log WHERE generic_used = TRUE;