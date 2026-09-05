-- Migration 004: Add warning_text and clarifying_questions to messages
-- WHY: Gate 3 WARN text and Gate 1 ASK clarifying questions were
-- streamed to client via SSE but never persisted. On page reload,
-- these were permanently lost. This migration fixes that data loss.
--
-- warning_text: Gate 3 WARN reasoning (why the charter was triggered)
-- clarifying_questions: Gate 1 ASK questions (what info is needed)

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS warning_text TEXT,
    ADD COLUMN IF NOT EXISTS clarifying_questions JSONB DEFAULT '[]';

COMMENT ON COLUMN messages.warning_text IS
    'Gate 3 WARN reasoning text. NULL means no warning was triggered.';

COMMENT ON COLUMN messages.clarifying_questions IS
    'Gate 1 ASK clarifying questions list. Empty array means no clarification needed.';
