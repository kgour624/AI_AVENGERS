-- Migration 046: concept relationships (I4, the "understanding" layer)
--
-- WHY this migration exists:
--
-- An expert's knowledge was flat. course_chunks carries a topic label per chunk, and
-- expert_capabilities aggregates those labels into a row each — so the system knew
-- WHICH topics a corpus mentions and how many chunks cover them, and nothing about how
-- the topics relate. A topic list is an index; it is not understanding. A human engineer
-- knows that consistent hashing is part_of sharding, that cache stampede is a
-- failure mode arising in caching, and that SQL and NoSQL are contrasted as
-- alternatives — and that structure is what makes an answer explain WHY rather than
-- recite.
--
-- This table stores those relationships: directed, typed, with the reason, and
-- anchored to a chunk that supports it so a relation can be checked rather than
-- believed.
--
-- WHY a separate table and not columns on expert_capabilities: a relationship has two
-- ends and a type. Denormalising it onto the topic rows would mean storing each edge
-- twice and keeping the copies consistent.
--
-- WHY the relation is a closed set: the graph is read by retrieval, which decides
-- whether to pull a neighbour's chunks into the candidate set. An open-ended string
-- would make that decision depend on whatever wording the model produced; a closed set
-- keeps the behaviour predictable and the CHECK constraint keeps the data honest.
--
-- Reversible: down drops the table. Nothing else references it, and the topics it
-- describes are unaffected.

CREATE TABLE expert_concept_edges (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id   UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,

    -- Both ends are topic labels from expert_capabilities / course_chunks. Names, not
    -- foreign keys to a capabilities row: a relation is still true if the topic is
    -- renamed, and a topic can legitimately be dropped from the capability table by a
    -- later repair without invalidating the edge between two other topics.
    from_topic  VARCHAR(255) NOT NULL,
    to_topic    VARCHAR(255) NOT NULL,

    relation    VARCHAR(32) NOT NULL,

    -- One sentence saying why, so a surprising edge can be judged instead of deleted.
    rationale   TEXT NOT NULL DEFAULT '',

    -- Anchored to the chunk that supports the relation. Nullable because a relation
    -- can be inferred from the corpus as a whole rather than one passage — but when it
    -- is set, the claim is checkable.
    evidence_chunk_id UUID REFERENCES course_chunks(id) ON DELETE SET NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT expert_concept_edges_relation_check CHECK (
        relation IN ('prerequisite_of', 'part_of', 'contrasts_with', 'used_with', 'fails_with')
    ),
    CONSTRAINT expert_concept_edges_not_self CHECK (from_topic <> to_topic)
);

-- Re-extraction must be idempotent: the same relation between the same two topics is
-- one row, however many times extraction runs.
CREATE UNIQUE INDEX expert_concept_edges_key
    ON expert_concept_edges(expert_id, from_topic, to_topic, relation);

-- Retrieval reads neighbours of the topics a question matched, so the lookup is by
-- either end of the edge.
CREATE INDEX idx_concept_edges_from ON expert_concept_edges(expert_id, from_topic);
CREATE INDEX idx_concept_edges_to ON expert_concept_edges(expert_id, to_topic);

COMMENT ON TABLE expert_concept_edges IS
    'Typed, directed relationships between an expert''s topics. Used to expand retrieval '
    'with neighbouring concepts, and to report how the corpus is structured.';

-- The retrieval mode a capability evaluation ran in. WHY it is stored on the run:
-- comparing a graph-expanded run against a plain one would attribute the difference to
-- the wrong change. The report compares like with like, so the mode has to be recorded
-- rather than inferred.
ALTER TABLE expert_capability_eval_runs
    ADD COLUMN IF NOT EXISTS graph_expansion BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN expert_capability_eval_runs.graph_expansion IS
    'Whether this pass retrieved with concept-graph expansion. Baseline comparisons are '
    'only made against a pass with the same value.';
