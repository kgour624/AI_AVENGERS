ALTER TABLE expert_capability_eval_runs
    DROP COLUMN IF EXISTS corpus_updated_at;

DELETE FROM knowledge_refresh_tasks
 WHERE task_type = 'corpus_contradiction';

ALTER TABLE knowledge_refresh_tasks
    DROP CONSTRAINT IF EXISTS krt_task_type_check;

ALTER TABLE knowledge_refresh_tasks
    ADD CONSTRAINT krt_task_type_check CHECK (task_type IN
        ('embedding_mismatch', 'stale_corpus', 'orphan_reference', 'empty_corpus'));
