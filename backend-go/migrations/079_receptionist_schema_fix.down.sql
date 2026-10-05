ALTER TABLE receptionist_checkpoints
    DROP COLUMN IF EXISTS text,
    DROP COLUMN IF EXISTS order_idx,
    DROP COLUMN IF EXISTS rating_history,
    DROP COLUMN IF EXISTS committed_summary;

ALTER TABLE receptionist_sessions
    DROP COLUMN IF EXISTS tenant_id,
    DROP COLUMN IF EXISTS admin_id,
    DROP COLUMN IF EXISTS language,
    DROP COLUMN IF EXISTS persona,
    DROP COLUMN IF EXISTS agenda,
    DROP COLUMN IF EXISTS state,
    DROP COLUMN IF EXISTS cursor,
    DROP COLUMN IF EXISTS updated_at;
