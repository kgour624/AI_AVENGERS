-- 075_add_collab_sections.up.sql
-- Collaborative Relay mode: when 2+ experts answer in "collaborative" mode,
-- their sequential sections are merged into ONE assistant message instead of
-- one message per expert. collab_sections stores the ordered section list
-- (expert_id, expert_name, section_title, content) so a reload/history fetch
-- renders the same merged view the live SSE stream showed.
-- NULL for every independent-mode message (the default) and every message
-- saved before this feature - additive, zero-regression (same convention as
-- migration 011's template_sections).
ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS collab_sections JSONB NULL;

COMMENT ON COLUMN messages.collab_sections IS 'Collaborative Relay (Phase 1): ordered [{expert_id, expert_name, section_title, content}] when answer_mode=collaborative and 2+ experts responded. NULL for independent-mode messages.';
