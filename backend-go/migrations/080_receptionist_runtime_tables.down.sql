-- Migration 080 DOWN: 080 me jo banaya gaya tha, wahi wapas hatao.
-- (078/079 se aaye asli receptionist tables ko chhua nahi jata.)

-- 5. receptionist_notes - Go store ke liye add kiye gaye columns hatao.
DROP INDEX IF EXISTS idx_notes_session_created;

ALTER TABLE receptionist_notes DROP COLUMN IF EXISTS created_at;
ALTER TABLE receptionist_notes DROP COLUMN IF EXISTS text;
ALTER TABLE receptionist_notes DROP COLUMN IF EXISTS tenant_id;

-- 4. final synthesis cache
DROP TABLE IF EXISTS receptionist_final_synthesis;

-- 3. expert responses
DROP TABLE IF EXISTS expert_responses;

-- 2. per-checkpoint templates
DROP TABLE IF EXISTS receptionist_templates;

-- 1. durable event log
DROP TABLE IF EXISTS conversation_events;
