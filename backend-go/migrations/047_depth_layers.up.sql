-- Migration 047: what KIND of content each chunk is (I5, option A)
--
-- WHY this migration exists:
--
-- The expert's "depth" was a lie with a number attached. depth_level came from how many
-- chunks mention a topic — so a topic with fourteen chunks of introductory definitions
-- scored the same as fourteen chunks of failure modes and trade-offs. The admin's real
-- question ("is this expert deep, or does it only know what things are?") had no answer
-- in the data at all.
--
-- This column answers it, and answers it truthfully, because it is DERIVED FROM THE
-- COURSE rather than written by a model. Option A: nothing is invented. Every chunk that
-- already exists gets classified by what kind of content it is:
--
--   1 = WHAT / WHY        a definition, a purpose, when to use the thing
--   2 = HOW / TRADE-OFFS  mechanics, comparisons, costs and benefits
--   3 = FAILURE / EDGE    what breaks, limitations, debugging, war stories
--
-- A topic with 20 chunks at layer 1 and none at layer 3 will answer "what is X" and
-- cannot answer "what breaks when X is under load". That is a real, checkable statement
-- about the expert's depth, and it is the kind of thing a buyer of a $699 course should
-- be able to see.
--
-- NULL means not classified yet. Deliberately nullable: classification is an on-demand,
-- resumable pass (it costs model calls), so a corpus can be partially classified, and
-- "not looked at" must be distinguishable from "looked at and found shallow".
--
-- Reversible: down drops the column and its index. Nothing else reads them, and the
-- chunks themselves are untouched.

ALTER TABLE course_chunks
    ADD COLUMN IF NOT EXISTS layer SMALLINT;

ALTER TABLE course_chunks
    ADD CONSTRAINT course_chunks_layer_check
        CHECK (layer IS NULL OR layer BETWEEN 1 AND 3);

-- Coverage reporting scans by expert and groups by layer; classification scans for the
-- rows still missing one.
CREATE INDEX IF NOT EXISTS idx_chunks_expert_layer
    ON course_chunks(expert_id, layer);

COMMENT ON COLUMN course_chunks.layer IS
    'Content kind: 1 what/why (definition), 2 how/trade-offs (mechanics), 3 failure/edge. '
    'NULL = not classified yet. Derived from the course by classification, never generated.';
