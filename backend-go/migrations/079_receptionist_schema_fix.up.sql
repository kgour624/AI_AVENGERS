-- Migration 079: Fix receptionist schema to match Go store.go queries
-- Problem: Migration 078 ran with OLD schema (user_id NOT NULL, current_cursor NOT NULL, missing columns)
-- Fix: Add all columns that store.go expects

-- Step 1: Make user_id and current_cursor nullable (they exist in old schema)
ALTER TABLE receptionist_sessions
    ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE receptionist_sessions
    ALTER COLUMN current_cursor DROP NOT NULL;

-- Step 2: Add missing columns that store.go INSERT queries use
ALTER TABLE receptionist_sessions
    ADD COLUMN IF NOT EXISTS tenant_id UUID,
    ADD COLUMN IF NOT EXISTS admin_id UUID,
    ADD COLUMN IF NOT EXISTS language TEXT NOT NULL DEFAULT 'EN',
    ADD COLUMN IF NOT EXISTS persona TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS agenda TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'INIT',
    ADD COLUMN IF NOT EXISTS cursor JSONB,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Step 3: Fix receptionist_checkpoints - add columns store.go uses
ALTER TABLE receptionist_checkpoints
    ALTER COLUMN section_key DROP NOT NULL;

ALTER TABLE receptionist_checkpoints
    ADD COLUMN IF NOT EXISTS text TEXT,
    ADD COLUMN IF NOT EXISTS order_idx INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS rating_history INT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS committed_summary TEXT;
