ALTER TABLE workflows DROP COLUMN IF EXISTS runner_state, DROP COLUMN IF EXISTS current_task_cursor;
ALTER TABLE messages DROP COLUMN IF EXISTS template_sections, DROP COLUMN IF EXISTS reply_to_message_id;
