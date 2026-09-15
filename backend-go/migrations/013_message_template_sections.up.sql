-- ============================================================
-- Migration 011: messages.template_sections (persist structured answers)
-- ============================================================
-- FIXES A REAL, REPORTED BUG (2026-09-08 RCA, round 7):
-- message/handler.go's saveAssistantMessage has documented this exact
-- gap since CT-B4 as a "KNOWN GAP" comment: a categorized expert's
-- structured answer (chinawall.EnforceResult.TemplateSections) is
-- streamed correctly over SSE on the FIRST render, but was never
-- persisted to the messages table - only resp.Content (which is EMPTY
-- for a structured response; the structured path only populates
-- TemplateSections, never Answer/Content) was saved. Result: a page
-- reload, or any fetch via GET /chats/:id/messages, shows a
-- categorized expert's past answer as completely blank - while a
-- fresh SSE stream renders it correctly. This is the exact
-- "sometimes it shows something, sometimes nothing at all" behavior
-- reported by the admin.
--
-- Design: nullable JSONB column, additive only (no RENAME/DROP/type
-- change on existing columns), matching migration 010's own stated
-- design principles. NULL = flat-text message (CT-L2) or any message
-- saved before this migration - zero regression, existing rows are
-- simply never re-read through this column.
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS template_sections JSONB;

COMMENT ON COLUMN messages.template_sections IS
    'JSONB array mirroring chinawall.TemplateSectionResult ([{key,label,type,content,citations}, ...]). NULL for every flat-text message (CT-L2) and every message saved before migration 011. Populated only for a categorized expert''s structured answer, so it survives page reloads / message history fetches instead of appearing blank. See HANDOFF.md 2026-09-08 round 7.';
