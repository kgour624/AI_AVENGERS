-- Rollback: remove seeded admin
-- WARNING: Only run this if you want to remove the default admin.
-- Will fail if admin has created projects/experts (FK constraints).
DELETE FROM users WHERE email = 'admin@aiavengers.com' AND role = 'admin';
