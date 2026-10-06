-- 085 Vacuum Phase 2: seed deterministic + semantic filler + hot-reload + DTS hardening
-- Input: empty kachra_patterns -> Process: seed curated patterns -> Output: brain v2 ready + DTS picker index optimal

-- Seed curated deterministic patterns (P1 regex also handles these; DB copy enables trie search + hit_count learning)
INSERT INTO kachra_patterns(pattern, pattern_type, category) VALUES
 ('uh','WORD','filler'),('umm','WORD','filler'),('hmm','WORD','filler'),('ah','WORD','filler'),('er','WORD','filler'),
 ('so guys','PHRASE','filler'),('you know','PHRASE','filler'),('i mean','PHRASE','filler'),('kind of','PHRASE','filler'),
 ('can you see my screen','PHRASE','classroom_meta'),('is my screen visible','PHRASE','classroom_meta'),('am i audible','PHRASE','classroom_meta'),
 ('please like share and subscribe','PHRASE','classroom_meta'),('hit the bell icon','PHRASE','classroom_meta'),
 ('thank you so much','PHRASE','logistics')
ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO NOTHING;

-- Harden DTS picker: partial index for scheduled only (SKIP LOCKED scan)
CREATE INDEX IF NOT EXISTS idx_file_jobs_pick ON file_jobs(created_at) WHERE status='scheduled';
-- S3 dedup (already unique per s3_key pending but add explicit)
CREATE UNIQUE INDEX IF NOT EXISTS idx_file_jobs_s3_unique ON file_jobs(s3_key);
