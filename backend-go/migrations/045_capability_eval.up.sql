-- Migration 045: measured expert capability (I2)
--
-- WHY this migration exists:
--
-- Until now an expert's capability was DECLARED, never measured. expert_capabilities
-- held a depth_level derived from how many chunks carry a topic, plus can_handle /
-- cannot_handle lists an LLM produced from about 600 characters per topic while each
-- topic averaged fourteen chunks. So the row that answers "what does this expert
-- know?" was a guess dressed as a claim, and Gate 2 read it.
--
-- This migration adds the storage for the opposite approach: ask the expert real
-- questions drawn from its own corpus, and record what actually happened. Three
-- tables, because the three things have different lifetimes:
--
--   expert_capability_cases   the durable question set (ground truth: which chunk
--                             answers each question). Regenerated only when the
--                             corpus changes, so runs stay comparable.
--   expert_capability_eval_runs  one row per evaluation pass: status, budget, and
--                             the aggregate counts the admin screen lists. A run
--                             row is what makes the pass async (a UI can poll it)
--                             and what makes a baseline possible (runs ordered in
--                             time are the regression harness).
--   expert_capability_results one row per case per run: the retrieved chunk ids,
--                             their rank, whether the answer cited and was judged
--                             supported, and why it failed otherwise.
--
-- WHY retrieved ids are stored: "no retrieval logging" is the anti-pattern that
-- makes RAG undebuggable. Without the ids, a failure can only be reported as "it
-- got it wrong" — with them, the admin can see whether retrieval never surfaced
-- the chunk or the answer was ungrounded despite good context, which are different
-- problems with different fixes.
--
-- measured_level is a DIFFERENT SCALE from depth_level and the two must never be
-- compared numerically. depth_level is 1-5 and means "how many chunks mention this
-- topic". measured_level is 1-3 and means "how hard a question about this topic the
-- expert actually answered": 1 = definitions, 2 = mechanics and trade-offs,
-- 3 = failure modes and edge cases. The UI shows them separately, labelled.
--
-- Reversible: down drops the three tables and the four columns. Nothing else reads
-- them yet, so no data migration is needed.

CREATE TABLE expert_capability_cases (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id            UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    topic                VARCHAR(255) NOT NULL,

    -- 1 = definition, 2 = mechanics/trade-offs, 3 = failure modes.
    level                INTEGER NOT NULL,

    question             TEXT NOT NULL,
    -- Ground truth: the chunk whose content answers this question. Retrieval is
    -- scored against this id, so it must be a real corpus row.
    expected_chunk_id    UUID NOT NULL REFERENCES course_chunks(id) ON DELETE CASCADE,
    expected_source_file VARCHAR(500) NOT NULL DEFAULT '',

    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT expert_capability_cases_level_check CHECK (level BETWEEN 1 AND 3)
);

-- Regeneration must not pile up near-duplicate questions for the same expert;
-- the unique key makes regeneration idempotent.
CREATE UNIQUE INDEX expert_capability_cases_question_key
    ON expert_capability_cases(expert_id, question);

CREATE INDEX idx_capability_cases_expert_topic
    ON expert_capability_cases(expert_id, topic);

CREATE TABLE expert_capability_eval_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id     UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,

    status        VARCHAR(20) NOT NULL DEFAULT 'running',

    -- The budget the pass was given, recorded so a score is always read next to
    -- the sample it came from.
    topics_total  INTEGER NOT NULL DEFAULT 0,
    cases_total   INTEGER NOT NULL DEFAULT 0,
    top_k         INTEGER NOT NULL DEFAULT 0,

    -- Aggregate counts. Stored rather than derived because the admin list reads
    -- them directly, and because a baseline is only useful if it cannot change
    -- after the fact.
    cases_passed  INTEGER NOT NULL DEFAULT 0,
    retrieval_hits INTEGER NOT NULL DEFAULT 0,
    grounded      INTEGER NOT NULL DEFAULT 0,
    refused       INTEGER NOT NULL DEFAULT 0,

    error_message TEXT NOT NULL DEFAULT '',
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,

    CONSTRAINT expert_capability_eval_runs_status_check
        CHECK (status IN ('running', 'complete', 'failed'))
);

CREATE INDEX idx_capability_eval_runs_expert
    ON expert_capability_eval_runs(expert_id, started_at DESC);

CREATE TABLE expert_capability_results (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id         UUID NOT NULL REFERENCES expert_capability_eval_runs(id) ON DELETE CASCADE,
    case_id        UUID NOT NULL REFERENCES expert_capability_cases(id) ON DELETE CASCADE,
    expert_id      UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,

    -- Copied from the case so a topic report never needs a join.
    topic          VARCHAR(255) NOT NULL,
    level          INTEGER NOT NULL,

    -- What retrieval actually returned, in rank order. The evidence behind
    -- retrieval_hit and the reason failure_reason can be specific.
    retrieved_ids  UUID[] NOT NULL DEFAULT '{}',
    -- 1-based rank of the expected chunk, 0 = never retrieved in top_k.
    hit_rank       INTEGER NOT NULL DEFAULT 0,

    cited          BOOLEAN NOT NULL DEFAULT FALSE,
    -- LLM judge verdict on whether the answer is supported by the context it was
    -- given: supported / refuted / unverifiable. Empty = not judged.
    grounded       VARCHAR(20) NOT NULL DEFAULT '',
    refused        BOOLEAN NOT NULL DEFAULT FALSE,
    passed         BOOLEAN NOT NULL DEFAULT FALSE,

    answer         TEXT NOT NULL DEFAULT '',
    -- Which check failed first: retrieval_missed | not_cited | ungrounded |
    -- refused | empty. Empty when the case passed, so the report can say why.
    failure_reason VARCHAR(40) NOT NULL DEFAULT '',

    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_capability_results_run
    ON expert_capability_results(run_id, topic, level);

-- The measured verdict is written back onto the capability row it replaces, so
-- every existing reader of expert_capabilities (the topics endpoint, Gate 2) sees
-- a measurement instead of a guess without changing its query.
ALTER TABLE expert_capabilities
    ADD COLUMN IF NOT EXISTS measured_level    INTEGER,
    ADD COLUMN IF NOT EXISTS eval_cases        INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS eval_passed       INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_evaluated_at TIMESTAMPTZ;

ALTER TABLE expert_capabilities
    ADD CONSTRAINT expert_capabilities_measured_level_check
        CHECK (measured_level IS NULL OR measured_level BETWEEN 1 AND 3);

COMMENT ON COLUMN expert_capabilities.measured_level IS
    'Measured answer depth: 1 definitions, 2 mechanics/trade-offs, 3 failure modes. '
    'NULL = never evaluated. NOT comparable to depth_level (1-5, chunk coverage).';
COMMENT ON COLUMN expert_capabilities.depth_level IS
    'Coverage band derived from chunk count (<5 -> 1, 5-14 -> 2, 15-29 -> 3, '
    '30-49 -> 4, >=50 -> 5). Reflects how the corpus was split, not difficulty.';
