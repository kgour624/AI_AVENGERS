-- T2 stores cross-source contradictions as actionable freshness tasks.
ALTER TABLE knowledge_refresh_tasks
    DROP CONSTRAINT IF EXISTS krt_task_type_check;

ALTER TABLE knowledge_refresh_tasks
    ADD CONSTRAINT krt_task_type_check CHECK (task_type IN
        ('embedding_mismatch', 'stale_corpus', 'orphan_reference', 'empty_corpus', 'corpus_contradiction'));

-- Capability measurement freshness is based on the corpus version it measured.
ALTER TABLE expert_capability_eval_runs
    ADD COLUMN IF NOT EXISTS corpus_updated_at TIMESTAMPTZ;

UPDATE expert_capability_eval_runs r
   SET corpus_updated_at = COALESCE(
       (SELECT MAX(cc.created_at) FROM course_chunks cc WHERE cc.expert_id = r.expert_id),
       r.started_at
   )
 WHERE corpus_updated_at IS NULL AND r.status = 'complete';
