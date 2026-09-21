-- Reverse migration 019: clear the seeded authoring_rank values.
-- Sets all ranks back to NULL (the pre-019 state).
-- Admin-set values are also cleared — this is a full rollback.
UPDATE expert_categories SET authoring_rank = NULL;
