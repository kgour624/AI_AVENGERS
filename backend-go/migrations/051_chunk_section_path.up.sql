-- Section metadata for course chunks.
--
-- WHY: the chunker now splits on the document's own headings before packing to a
-- size target, so every chunk belongs to exactly one section ("RAG > Chunking")
-- instead of possibly straddling two topics. The path is stored because the
-- design's retrieval pipeline applies a metadata filter BEFORE retrieval, and a
-- filter cannot use metadata that was thrown away at ingest time.
--
-- '' (not NULL) when the source had no headings — a plain transcript — or the
-- chunk is preamble. Existing rows keep '' and stay fully searchable, so this
-- column's arrival changes no retrieval behaviour on its own: an expert gains
-- section metadata when its transcript is ingested again.
ALTER TABLE course_chunks ADD COLUMN IF NOT EXISTS section_path TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN course_chunks.section_path IS
    'Heading trail this chunk belongs to ("RAG > Chunking"); empty when the source had no headings.';
