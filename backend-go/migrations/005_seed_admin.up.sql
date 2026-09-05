-- Migration 005: Seed default admin user
-- WHY: Fresh deployment has no users. Admin must exist to:
--   1. Upload transcripts (train experts)
--   2. Create/enable experts
--   3. Access admin panel
--
-- IMPORTANT: Change password after first login in production.
-- Default password: Admin@123456
-- Bcrypt hash (cost 12) of Admin@123456:

INSERT INTO users (
    email,
    hashed_password,
    full_name,
    role,
    is_active,
    totp_enabled
)
VALUES (
    'admin@aiavengers.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj4J/HS.iK8i',
    'Admin',
    'admin',
    true,
    false
)
ON CONFLICT (email) DO NOTHING;
-- ON CONFLICT: safe to re-run. If admin already exists, skip.
