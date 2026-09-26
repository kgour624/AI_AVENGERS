DROP POLICY IF EXISTS tenant_isolation ON projects;
DROP POLICY IF EXISTS tenant_isolation ON experts;
DROP POLICY IF EXISTS tenant_isolation ON users;

ALTER TABLE projects DISABLE ROW LEVEL SECURITY;
ALTER TABLE experts  DISABLE ROW LEVEL SECURITY;
ALTER TABLE users    DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS app_tenant_visible(uuid);
DROP FUNCTION IF EXISTS app_is_system();
DROP FUNCTION IF EXISTS app_current_tenant();
