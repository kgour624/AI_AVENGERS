-- Database-level tenant isolation (A11).
--
-- WHY this exists on top of the application predicates (tenant.Service.Resolve /
-- AssertProject / AssertExperts): those predicates protect the queries that
-- remember to call them. One forgotten predicate in a new handler is a silent
-- cross-client leak, and nothing in the system would notice. A policy on the
-- table cannot be forgotten.
--
-- HONEST STATUS: these policies are inert for the current deployment, because the
-- app connects as the table OWNER and an owner bypasses row security unless the
-- table is put in FORCE mode. That is deliberate — enabling enforcement without
-- first moving request traffic to a non-owner role would make rows disappear from
-- every background path that does not set a tenant (ingestion, the workflow
-- runner, the chat stream), which is the worst possible failure: silent and
-- total. The switch and its exact steps are documented at the bottom of this
-- file, and file 053 does NOT flip it.
--
-- SEMANTICS, taken from the application so the two can never disagree:
--   * tenant_id IS NULL means GLOBAL (platform experts, admin users) — visible to
--     everybody, exactly as tenant.Scope.Allows treats it.
--   * app.tenant_id unset or empty → NOTHING is visible. Fail closed: a request
--     path that forgets the setting must see nothing, not everything.
--   * app.system = 'on' → everything is visible. This is the escape hatch for
--     background work, and it is explicit so that "privileged" is a decision
--     someone made, not a default someone inherited.

-- app_current_tenant() returns the tenant the current transaction is scoped to,
-- or NULL when none is set. NULL is the fail-closed case in every policy below.
CREATE OR REPLACE FUNCTION app_current_tenant() RETURNS uuid
LANGUAGE sql STABLE AS $$
    SELECT NULLIF(current_setting('app.tenant_id', true), '')::uuid
$$;

-- app_is_system() is true only when a transaction has explicitly declared itself
-- system-scoped. Both functions read transaction-local settings, so a pooled
-- connection cannot carry a tenant into the next request.
CREATE OR REPLACE FUNCTION app_is_system() RETURNS boolean
LANGUAGE sql STABLE AS $$
    SELECT COALESCE(current_setting('app.system', true), '') = 'on'
$$;

-- The predicate every policy shares. Kept in one function so the three tables
-- cannot drift apart.
CREATE OR REPLACE FUNCTION app_tenant_visible(row_tenant uuid) RETURNS boolean
LANGUAGE sql STABLE AS $$
    SELECT app_is_system()
        OR (app_current_tenant() IS NOT NULL
            AND (row_tenant IS NULL OR row_tenant = app_current_tenant()))
$$;

ALTER TABLE users    ENABLE ROW LEVEL SECURITY;
ALTER TABLE experts  ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON users;
CREATE POLICY tenant_isolation ON users
    USING (app_tenant_visible(tenant_id))
    WITH CHECK (app_tenant_visible(tenant_id));

DROP POLICY IF EXISTS tenant_isolation ON experts;
CREATE POLICY tenant_isolation ON experts
    USING (app_tenant_visible(tenant_id))
    WITH CHECK (app_tenant_visible(tenant_id));

DROP POLICY IF EXISTS tenant_isolation ON projects;
CREATE POLICY tenant_isolation ON projects
    USING (app_tenant_visible(tenant_id))
    WITH CHECK (app_tenant_visible(tenant_id));

-- ============================================================
-- HOW TO TURN ENFORCEMENT ON (an operator decision, not a migration)
-- ============================================================
-- 1. Create a NON-OWNER role for request traffic:
--      CREATE ROLE avengers_app LOGIN PASSWORD '<strong>';
--      GRANT USAGE ON SCHEMA public TO avengers_app;
--      GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO avengers_app;
--      GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO avengers_app;
--      ALTER DEFAULT PRIVILEGES IN SCHEMA public
--        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO avengers_app;
--    The role must NOT be the owner of these tables and must NOT have BYPASSRLS.
--
-- 2. Point request traffic at it, and set the tenant per transaction:
--      SET LOCAL app.tenant_id = '<uuid>';     -- request-scoped queries
--      SET LOCAL app.system = 'on';           -- background jobs only
--    Both are transaction-local on purpose: a pooled connection must not carry one
--    request's tenant into the next.
--
-- 3. Verify before switching FORCE on, by connecting AS avengers_app:
--      SET app.tenant_id = '<tenant A>';
--      SELECT count(*) FROM projects;   -- only A's rows (plus NULL-tenant rows)
--      RESET app.tenant_id;
--      SELECT count(*) FROM projects;   -- must be 0: fail closed
--      SET app.system = 'on';
--      SELECT count(*) FROM projects;   -- all rows
--
-- 4. Only then, if you want the isolation to survive a connection that is the
--    owner (e.g. a future tool connecting as avengers):
--      ALTER TABLE users    FORCE ROW LEVEL SECURITY;
--      ALTER TABLE experts  FORCE ROW LEVEL SECURITY;
--      ALTER TABLE projects FORCE ROW LEVEL SECURITY;
--    Do this ONLY after every background path sets app.system = 'on', because
--    FORCE applies to the owner too.
