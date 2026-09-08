-- ============================================================
-- Migration 010 DOWN: reverse Expert Categories + Template System + Reply Threading
-- ============================================================
-- Reverses migration 010 in strict child-before-parent order.
-- Preserves all tables/columns not owned by this migration.
-- ============================================================

-- Reverse system_settings seed
DELETE FROM system_settings WHERE key = 'reply_thread_max_depth';

-- Reverse messages.reply_to_message_id (drop index first, then column)
DROP INDEX IF EXISTS idx_messages_reply_to;
ALTER TABLE messages DROP COLUMN IF EXISTS reply_to_message_id;

-- Reverse experts.category_id (drop index first, then column)
-- WHY safe: this also nulls out the FK; no data loss beyond the
-- category assignment itself, which is exactly what "down" should do.
DROP INDEX IF EXISTS idx_experts_category;
ALTER TABLE experts DROP COLUMN IF EXISTS category_id;

-- Reverse expert_categories table (must be last — was referenced by
-- the two columns above, both already dropped)
DROP INDEX IF EXISTS idx_expert_categories_slug;
DROP TABLE IF EXISTS expert_categories;

-- ============================================================
-- End of migration 010 down
-- ============================================================
