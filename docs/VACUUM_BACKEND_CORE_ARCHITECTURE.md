# VACUUM BACKEND — CORE ARCHITECTURE & FUNCTIONALITY SPEC

> **Audience:** System Architect, Backend reviewers, SRE / on-call engineers.
> **Scope:** **Backend only** — the Go `vacuum` subsystem. No React, no hooks, no UI code.
> **Source of truth:** Verified line-by-line against the live code in
> `backend-go/internal/vacuum/**`, `backend-go/cmd/server/main.go` and
> `backend-go/migrations/084..093_*`.
> **Document version:** 1.0 — written after the orphan-reaper + cancel/retry hardening round.

---

## 0. HOW TO READ THIS DOCUMENT

This document is written in five layers, each one deeper than the last:

| Layer | Sections | Answers the question |
|---|---|---|
| Orientation | 1–4 | What is Vacuum, where does it live, what does it talk to? |
| State model | 5–6 | What states exist, who writes them, what do the phase numbers mean? |
| Component deep dive | 7 | What does **every file and every function** do, and how? |
| Correctness | 8–11 | Why is it safe? What breaks it? What happens when it breaks? |
| Operations | 12–16 | Config, metrics, complexity, limitations, indexes, glossary |

Reading order for a new architect:
`§1 → §3 → §5 → §6 → §7.11 (reaper) → §7.12 (pipeline) → §10 (failure taxonomy)`.

Every claim in this document is traceable to a symbol. If a behaviour is
described, the function that implements it is named in backticks, e.g.
`ExecuteOne`, `classifyOpenError`, `sweep`.

---

## 1. PURPOSE — WHAT PROBLEM DOES VACUUM SOLVE?

### 1.1 The product problem

Raw course/lecture transcripts (ASR output, Hinglish, classroom recordings)
contain **"kachra"** — noise that has no informational value:

- filler interjections (`uh`, `umm`, `hmm`, `so guys`, `you know`, `i mean`)
- classroom logistics (`am i audible?`, `can you see my screen?`)
- engagement spam (`like share and subscribe`, `hit the bell icon`)
- repetition, ASR errors, semantic noise, smalltalk
- formatting garbage (`......`, `------`, speaker labels)

The Vacuum subsystem ingests such a file and produces a **cleaned, chunked,
classified, headed, hash-verified** version of it, stored back on disk, with
per-chunk audit rows in Postgres.

### 1.2 The engineering problem (and the hard constraint)

LLMs are good at *suggesting* what is noise and **bad** at *removing* it safely.
An LLM that rewrites text can silently delete content, paraphrase, or hallucinate.

Vacuum's core invariant, repeated in the code comments and enforced in code:

> **The LLM never mutates content. The deterministic algorithm (DSA) is the sole
> mutator. The LLM only produces *suggestions* (`KachraSpan`) that DSA maps back
> to byte offsets, validates, and only then removes.**

This is what the codebase calls **"No-Trust" / "Harness > Prompt"**:
the prompt asks the model for JSON spans; the harness (hash guard, contiguous
range checks, word-boundary checks, byte-exact mapping) decides what actually
happens.

### 1.3 The six phases (product vocabulary)

| Phase | Name | What it does | Mutates content? |
|---|---|---|---|
| 0 | `scheduled` | row enqueued, no work started | – |
| 1 | `picking` | DTS picker claims job with `FOR UPDATE SKIP LOCKED` | – |
| 2 | `cleaning` | P1 regex + P2 trie + P3 LLM suggest/verify + splice | **YES** (DSA only) |
| 3 | `classifying` | bounded concurrent Gemini classifier per chunk | no |
| 4 | `headings` | Claude heading generator (`##`/`###` only) | no |
| 5 | `verifying` | SHA guard + contiguous checks + **atomic commit** | no (writes DB/store) |
| 6 | terminal | `done` / `failed` / `quarantined` / `canceled` | – |

The number `6` that the UI renders as `phase 6` is the terminal band; it is not
"step 6 of 6 work units".

---

## 2. WHERE VACUUM LIVES — PACKAGE TREE

```
backend-go/
├── cmd/server/main.go                  ← wiring: service, LLM adapters, reaper, routes
├── internal/vacuum/                    ← THE FEATURE PACKAGE (this document)
│   ├── service.go                      ← 3-firewall wiring root (Brain + Engine + DB + LLMs)
│   ├── storage.go                      ← Storage interface + FSStorage (local FS, inline: keys)
│   ├── progress.go                     ← SSE event model, in-process hub, run registry
│   ├── reaper.go                       ← orphan reaper (boot heal + 1-min sweep, 45m grace)
│   ├── pipeline.go                     ← PickAndExecute + ExecuteOne (the 6-phase DAG)
│   ├── handler.go                      ← pattern CRUD + candidate list (admin routes)
│   ├── handler2.go                     ← jobs CRUD, pick, execute, cancel, delete, retry, SSE, stats
│   ├── handler_phase3.go               ← preview/diff, eval stats, drift, auto-promote
│   ├── evolve.go                       ← AutoPromote (self-learning) + DriftCheck
│   ├── open_error_test.go              ← pins INPUT_MISSING vs STORAGE_OPEN split
│   ├── scan_percent_test.go            ← pins phase-2 progress band maths
│   ├── brain/                          ← Aho-Corasick pattern brain
│   │   ├── trie.go                     ← Trie: Add / Build (fail links) / Search
│   │   └── store.go                    ← versioned copy-on-write snapshot + 30s hot reload
│   ├── engine/                         ← deterministic engine (DSA)
│   │   ├── engine.go                   ← Clean, CleanHybrid, CleanReader, regexes, IsFiller
│   │   └── progress_test.go
│   ├── chunker/
│   │   └── chunker.go                  ← pronoun-aware sentence chunking with byte offsets
│   ├── filecontext/
│   │   └── context.go                  ← per-file visited-set + interval tree
│   ├── llm/                            ← model-facing ports & adapters
│   │   ├── llm.go                      ← interfaces: Classifier, HeadingGenerator, Guard, Kachra*
│   │   ├── gateway_adapter.go          ← breaks the gateway import cycle
│   │   ├── classifier.go               ← Gemini flash classifier + heuristic fallback
│   │   ├── heading.go                  ← Claude heading generator + strict validation
│   │   ├── guard.go                    ← SHA256 preservation guard
│   │   ├── kachra_detector.go          ← detector (suggests spans, temp 0.0)
│   │   ├── kachra_verifier.go          ← verifier (drops <0.70, normalises, caps 20)
│   │   ├── kachra_sink.go              ← persists suggestions to candidate_kachra
│   │   └── cache.go                    ← 10m TTL read-through detector cache
│   └── runner/
│       └── dag.go                      ← generic sequential stage DAG + RunParallel
├── internal/observability/metrics.go   ← vacuum counters surfaced on /vacuum/metrics
└── migrations/084..093_vacuum_*.sql    ← schema, indexes, triggers, views, checks
```

**Layering rule actually followed in the tree:**

- `brain/`, `chunker/`, `filecontext/`, `llm/`, `engine/`, `runner/` are **pure
  libraries**: no `gin`, no per-request `pgxpool` (except the sink), no HTTP.
- `vacuum/*.go` (root of the package) is the **application/feature layer**: SQL,
  HTTP handlers, orchestration.
- `cmd/server/main.go` is the **composition root**: it instantiates and injects.

Dependency direction (arrows = "imports"):

```
cmd/server ──► vacuum(root) ──► engine ──► brain, chunker, filecontext, llm
                    │
                    └────────► llm (directly, for sinks/interfaces)

engine ✗──► vacuum      IMPOSSIBLE (import cycle) — see §7.5.7 WithProgress
```

This single fact explains several design decisions in the codebase: the
`WithProgress` context-value trick, the `ProgressFn` indirection, and why the
`llm` package defines its own `LLMCaller`/`LLMRequest` instead of importing
`internal/gateway`.

### 2.1 The "3-firewall" mental model (used in the code comments)

| Firewall | Layer | Contains | Rule |
|---|---|---|---|
| F1 | Foundation | `pgxpool`, `zap`, `gateway.ModelGateway` | no feature logic |
| F2 | Feature/Business | `vacuum.Service`, `brain.Store`, `engine.Engine` | no HTTP, no request objects |
| F3 | App/Interface | `vacuum.Handler`, `gin` routes | no SQL beyond query composition |

`Service` is the object that crosses F2→F3: handlers hold `*Service` and ask it
for `Engine()`, `Brain()`, `Classifier()`, `Guard()`, `DB()`.

---

## 3. END-TO-END DATA FLOW (THE CANONICAL HAPPY PATH)

### 3.1 ASCII swimlane

```
[Admin browser]
      │  POST /vacuum/upload (multipart file)
      ▼
[vacuum.Handler.UploadAndEnqueue]                       handler2.go
      │  FSStorage.Put("vacuum/<uuid>/<name>")          storage.go
      │  INSERT INTO file_jobs(s3_key,file_name,file_size) → status='scheduled', phase=0
      ▼
[Postgres: file_jobs]
      │
      │  ── trigger: trg_file_jobs_updated → set_updated_at() ── (every UPDATE)
      ▼
[vacuum.Pipeline.PickAndExecute]                        pipeline.go
      │  1. heal stuck 'verifying' > 2m → scheduled (legacy debounce)
      │  2. SELECT ... WHERE status='scheduled' ORDER BY created_at LIMIT n FOR UPDATE SKIP LOCKED
      │  3. UPDATE status='picking', picked_at=NOW(), picked_by=$, attempts=attempts+1, phase=1
      │  4. commit tx  (locks released; work happens outside the tx)
      │  5. per job: context.WithCancel + registerRun(id, cancel) → goroutine (sem=5)
      ▼
[vacuum.Pipeline.ExecuteOne]  ── 6 stages, each publishing ProgressEvent ──┐
      │                                                                     │
      │ stage 1  open:   FSStorage.Open(s3_key)                             │
      │ stage 2  clean:  engine.CleanHybrid(...)  [DSA + LLM suggest/verify]│
      │ stage 3  classify: classifyChunks()       [bounded 8, per chunk]    │
      │ stage 4  headings: HeadingGenerator.Generate()                      │
      │ stage 5  verify:  PreservationGuard.Verify() + contiguous checks    │
      │                   saveStageCache()  ← checkpoint BEFORE commit      │
      │ stage 6  commit:  tx2: DELETE chunks, INSERT chunks, UPDATE done    │
      │                   FSStorage.Put(outputKey, cleaned)                 │
      │                   re-open output → SHA256 → compare                 │
      ▼                                                                     │
[Postgres: file_jobs status='done'] + [FS: vacuum/<uuid>/<name>.cleaned.txt]│
      │                                                                     │
      │  GET /vacuum/jobs/:id/download                                      │
      ▼                                                                     │
[vacuum.Handler.DownloadCleaned] → streams cleaned text, download_count++   │
                                                                            │
[vacuum.progressHub] ◄──── publishProgress(...) ───────────────────────────┘
      │  SSE fan-out (per job, buffered 64, non-blocking)
      ▼
[Admin browser: GET /vacuum/jobs/:id/stream]
```

### 3.2 Who calls what — call graph (abridged)

```
Handler.UploadAndEnqueue ──► FSStorage.Put
                          └► INSERT file_jobs

Handler.PickJob ──► Pipeline.PickAndExecute ──► Pipeline.ExecuteOne ──► Engine.CleanHybrid
                                                                     ├► classifyChunks → Classifier.Classify
                                                                     ├► HeadingGenerator.Generate
                                                                     ├► PreservationGuard.Verify
                                                                     ├► saveStageCache → FSStorage.Put(.stage.json)
                                                                     └► tx2 commit → file_chunks + file_jobs

Handler.ExecuteJob ──► registerRun + go Pipeline.ExecuteOne   (async, 202)

Handler.CancelJob ──► cancelRun(id)          (soft stop the run)
                   ├► cancelSnapshot()       (real phase/percent, not 0)
                   ├► publishProgress(canceled…)
                   ├► cleanupJob(id)         (delete chunks, row, blobs)
                   └► progressBus.Forget(id)

Handler.DeleteJob ──► cancelRun + cleanupJob + Forget

Handler.RetryJob / RetryLastStageJob ──► jobInputKey + FSStorage.Missing  (honest precheck)
                                      └► UPDATE status='scheduled' (phase 1 or resume_from_stage=5)

vacuum.StartOrphanReaper ──► HealOrphansOnBoot (grace=0)
                         └► every 1m: sweep(grace=45m) ──► isRunning() skip ──► UPDATE failed/ORPHANED
                                                        └► progressBus.Forget + publishProgress(failed)
```

### 3.3 The single most important sentence in this document

> A job's lifecycle is a **row in `file_jobs`** plus a **blob under
> `VACUUM_STORAGE_ROOT`** plus an **in-process goroutine** — and these three can
> disagree. Every piece of hardening in this subsystem exists because one of the
> three can outlive or predecease the other two.

That mismatch is:
- why `Missing()` exists (`storage.go`),
- why `classifyOpenError` distinguishes `INPUT_MISSING` from `STORAGE_OPEN`,
- why the orphan reaper exists,
- why `registerRun`/`isRunning` exist,
- why `resume_from_stage` + `.stage.json` checkpoints exist.

---

## 4. DATA MODEL — EVERY TABLE, COLUMN, CONSTRAINT AND INDEX

Schema is built by migrations `084` → `093`. Nothing older touches Vacuum.
All statements are idempotent (`IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`).

### 4.1 `file_jobs` — the job queue AND the audit record

Created in `084`, then extended by `086`, `087`, `088`, `089`, `092`, `093`.

| Column | Type | Default | Added in | Meaning |
|---|---|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` | 084 | job id, also the `<uuid>` in the storage path |
| `s3_key` | TEXT NOT NULL | – | 084 | **storage key**, e.g. `vacuum/<uuid>/<file>.md` |
| `file_name` | TEXT NOT NULL | `''` | 084 | original display name |
| `file_size` | BIGINT NOT NULL | `0` | 084 | bytes at upload (informational) |
| `status` | TEXT NOT NULL | `scheduled` | 084 (+093 widened) | state machine — see §5 |
| `phase` | INT NOT NULL | `0` | 084 | `CHECK (phase>=0 AND phase<=6)` |
| `picked_at` | TIMESTAMPTZ | NULL | 084 | when a picker claimed it (progress clock for reaper) |
| `picked_by` | TEXT | NULL | 084 | picker instance label (default `vacuum-picker`) |
| `attempts` | INT NOT NULL | `0` | 084 | incremented on every successful pick |
| `last_error` | TEXT | NULL | 084 | human-readable failure reason |
| `sha256_input` | TEXT | NULL | 084 | hash of raw input (set at phase 5) |
| `sha256_output` | TEXT | NULL | 084 | hash of cleaned text (set at phase 5) |
| `s3_output_key` | TEXT | NULL | 084 | cleaned output storage key |
| `created_at` | TIMESTAMPTZ NOT NULL | `NOW()` | 084 | queue order key |
| `updated_at` | TIMESTAMPTZ NOT NULL | `NOW()` | 084 | maintained by trigger |
| `verified` | BOOLEAN NOT NULL | `FALSE` | 086 | did commit-time verification pass |
| `chunk_count` | INT NOT NULL | `0` | 086 | number of `file_chunks` rows |
| `retry_after` | TIMESTAMPTZ | NULL | 087 | optional retry gate (indexed) |
| `error_code` | TEXT | NULL | 087 | **machine-readable** code — see §10 |
| `finished_at` | TIMESTAMPTZ | NULL | 088 | terminal timestamp (done path) |
| `duration_ms` | INT | NULL | 088 | wall time of the whole run |
| `llm_classifier_label` | TEXT | NULL | 089 | dominant label from stage 3 |
| `llm_heading_count` | INT NOT NULL | `0` | 089 | headings produced in stage 4 |
| `preservation_verified` | BOOLEAN NOT NULL | `FALSE` | 089 | SHA guard result |
| `download_count` | INT NOT NULL | `0` | 089 | how many times output was downloaded |
| `last_downloaded_at` | TIMESTAMPTZ | NULL | 089 | last download timestamp |
| `llm_kachra_count` | INT NOT NULL | `0` | 092 | mapped LLM spans that were removed |
| `filler_filtered_count` | INT NOT NULL | `0` | 092 | chunks dropped as filler |
| `eval_score` | DOUBLE PRECISION | NULL | 092 | `(llm_kachra + headings) / (chunks + 1)` |
| `resume_from_stage` | INT NOT NULL | `0` | 093 | `0` = full run, `>=5` = final-stage resume |

**Status CHECK (after 093):**

```sql
CHECK (status IN ('scheduled','picking','cleaning','verifying',
                  'done','failed','quarantined','canceled'))
```

> Architect note: `classifying` and `headings` are **phase numbers**, not status
> values. Only `cleaning` and `verifying` are used as the "busy" statuses
> (plus `picking`). The reaper therefore only ever needs to heal three statuses.

**Indexes on `file_jobs`:**

| Index | Definition | Why it exists |
|---|---|---|
| `idx_file_jobs_status` | `(status)` | count-by-status dashboards |
| `idx_file_jobs_status_picked` | `(status, picked_at)` | legacy picker scan |
| `idx_file_jobs_created` | `(created_at)` | queue ordering / pagination |
| `idx_file_jobs_pick` | `(created_at) WHERE status='scheduled'` | **partial** — tiny index for the DTS scan |
| `idx_file_jobs_pick_composite` | `(status, created_at) WHERE status IN ('scheduled','picking','cleaning')` | in-flight lookups |
| `idx_file_jobs_pick_sched_retry` | `(status, retry_after, created_at)` | retry-aware picking |
| `idx_file_jobs_failed_retry` | `(retry_after) WHERE status='failed'` | retry sweep |
| `idx_file_jobs_done_at` | `(updated_at) WHERE status='done'` | throughput |
| `idx_file_jobs_done_finished` | `(finished_at DESC) WHERE status='done'` | "recently finished" list |
| `idx_file_jobs_preservation` | `(preservation_verified) WHERE status='done'` | integrity report |
| `idx_file_jobs_download` | `(download_count) WHERE status='done'` | popularity report |
| `idx_file_jobs_canceled` | `(status) WHERE status='canceled'` | canceled filter (093) |
| `idx_file_jobs_s3_unique` | UNIQUE `(s3_key)` | **one job per blob** — 085 |

Note the design pattern: almost every observational index is **partial**. The
table is expected to be dominated by terminal rows, so `WHERE status='...'`
partial indexes stay small and are updated only when a row is in that state.

### 4.2 `file_chunks` — per-chunk audit trail

| Column | Type | Default | Meaning |
|---|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` | row id |
| `file_job_id` | UUID NOT NULL FK→`file_jobs(id) ON DELETE CASCADE` | – | parent |
| `chunk_index` | INT NOT NULL | – | 0-based, **UNIQUE with job** |
| `char_start` | BIGINT NOT NULL | `0` | byte offset in cleaned text |
| `char_end` | BIGINT NOT NULL | `0` | byte offset (exclusive) |
| `hash_input` | TEXT NOT NULL | `''` | hash of the source slice |
| `hash_output` | TEXT NOT NULL | `''` | hash of the cleaned chunk |
| `status` | TEXT NOT NULL | `pending` | `pending｜cleaned｜verified｜failed` |
| `headings` | JSONB NOT NULL | `'[]'` | headings attached to this chunk |
| `created_at` | TIMESTAMPTZ NOT NULL | `NOW()` | – |
| `llm_label` | TEXT | NULL | 089 — classifier label |
| `llm_confidence` | DOUBLE PRECISION | NULL | 089 — classifier confidence |

```sql
UNIQUE (file_job_id, chunk_index)   -- the anti-duplicate guarantee
CREATE INDEX idx_file_chunks_job ON file_chunks(file_job_id);
```

`CASCADE` matters: `cleanupJob` deleting the `file_jobs` row automatically
removes every chunk row (used by Cancel and Delete).

### 4.3 `kachra_patterns` — the "brain" (deterministic pattern DB)

```sql
id            UUID PK
pattern       TEXT NOT NULL
pattern_type  TEXT NOT NULL DEFAULT 'PHRASE'
              CHECK (pattern_type IN ('WORD','PHRASE','REGEX','SEMANTIC'))
category      TEXT NOT NULL DEFAULT 'filler'
              CHECK (category IN ('filler','repetition','asr_error','classroom_meta',
                                  'hinglish','logistics','semantic_noise','smalltalk'))
is_active     BOOLEAN NOT NULL DEFAULT TRUE
hit_count     BIGINT  NOT NULL DEFAULT 0
version       BIGINT  NOT NULL DEFAULT 1
created_at / updated_at  TIMESTAMPTZ
CONSTRAINT kachra_pattern_not_empty CHECK (length(trim(pattern)) > 0)
```

Indexes:

```sql
CREATE UNIQUE INDEX idx_kachra_patterns_unique
  ON kachra_patterns(lower(trim(pattern)), pattern_type, category);
CREATE INDEX idx_kachra_patterns_active  ON kachra_patterns(is_active) WHERE is_active=TRUE;
CREATE INDEX idx_kachra_patterns_version ON kachra_patterns(version);
```

**Why `lower(trim(pattern))` in the unique index:** the application upserts with
`ON CONFLICT (lower(trim(pattern)), pattern_type, category)`, so the index and
the `ON CONFLICT` target must be *expression-identical*. This is a common SQL
foot-gun and is deliberately handled.

**Why the `version` column exists:** it is a monotonically increasing counter
per row; combined with `idx_kachra_patterns_version` it lets the store detect
"anything new since my last build?" cheaply without comparing full row sets.

`085` seeds the curated starter set (16 rows):

```
uh, umm, hmm, ah, er                      → WORD,   filler
so guys, you know, i mean, kind of        → PHRASE, filler
can you see my screen,
is my screen visible,
am i audible                              → PHRASE, classroom_meta
please like share and subscribe,
hit the bell icon                         → PHRASE, classroom_meta
thank you so much                         → PHRASE, logistics
```

`091` widens both `kachra_patterns` and `candidate_kachra` category CHECKs to
include `smalltalk` — because the LLM span type set has **8** values while the
084 CHECK had only **7**, so a `smalltalk` suggestion would have failed the
sink's UPSERT. The migration comment states exactly that. This is a textbook
example of a *schema/contract drift* bug that was found and fixed additively
(`ALTER`, not recreate).

### 4.4 `candidate_kachra` — the self-learning inbox

```sql
id              UUID PK
pattern         TEXT NOT NULL
pattern_type    TEXT NOT NULL DEFAULT 'PHRASE' CHECK (pattern_type IN (...4 values...))
category        TEXT NOT NULL DEFAULT 'filler' CHECK (category IN (...8 values, added 091...))
context_snippet TEXT NOT NULL DEFAULT ''
file_id         UUID            -- nullable: no FK on purpose (see below)
confidence      DOUBLE PRECISION NOT NULL DEFAULT 0.85 CHECK (>=0 AND <=1)
status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected'))
hit_count       BIGINT NOT NULL DEFAULT 1
first_seen_at / last_seen_at / updated_at TIMESTAMPTZ
```

```sql
CREATE UNIQUE INDEX idx_candidate_kachra_unique_pattern
  ON candidate_kachra(lower(trim(pattern)), pattern_type, category);
CREATE INDEX idx_candidate_kachra_status ON candidate_kachra(status);
CREATE INDEX idx_candidate_kachra_hit    ON candidate_kachra(hit_count DESC);
CREATE INDEX idx_candidate_auto_promote  ON candidate_kachra(status, hit_count, confidence)
  WHERE status='pending';           -- 092
```

**Why `file_id` has no FK:** the sink is best-effort telemetry. If the parent job
is deleted (cancel/delete) *while* a sink write is in flight, an FK would turn a
harmless audit write into a transaction error. Losing the back-reference is
acceptable; failing the clean is not. `Preview` (single-text, no job) also writes
candidates with `file_id = NULL`.

### 4.5 `vacuum_eval` — per-run quality telemetry (092)

```sql
id                 UUID PK
file_job_id        UUID REFERENCES file_jobs(id) ON DELETE CASCADE
dsa_hit_count      INT
llm_suggested      INT
llm_verified       INT
llm_mapped         INT
combined_hit_count INT
filler_filtered    INT
duration_ms        INT
created_at         TIMESTAMPTZ DEFAULT NOW()
```

Written **best-effort, after commit** in `ExecuteOne`. Note honestly: in the
current implementation `llm_suggested`, `llm_verified` and `llm_mapped` are all
bound to the same `llmKachra` value — the per-stage counters were collapsed to
one number by the time the insert was written. If you need the real funnel, the
metrics counters (§12) are the accurate source, not this table.

### 4.6 `brain_version` — single-row version singleton (084)

```sql
CREATE TABLE brain_version (
  id INT PRIMARY KEY DEFAULT 1 CHECK (id=1),
  version BIGINT NOT NULL DEFAULT 1,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO brain_version(id, version) VALUES (1,1) ON CONFLICT (id) DO NOTHING;
```

`CHECK (id=1)` makes it structurally impossible to have two rows. The row is the
**hot-reload signal**.

### 4.7 Triggers — the invisible logic you must know about

| Trigger | Table | Event | Effect |
|---|---|---|---|
| `trg_kachra_bump` | `kachra_patterns` | AFTER INSERT/UPDATE/DELETE **FOR EACH ROW** | `bump_brain_version()` → `UPDATE brain_version SET version=version+1` |
| `trg_kachra_updated` | `kachra_patterns` | BEFORE UPDATE | `set_updated_at()` |
| `trg_candidate_updated` | `candidate_kachra` | BEFORE UPDATE | `set_updated_at()` |
| `trg_file_jobs_updated` | `file_jobs` | BEFORE UPDATE | `set_updated_at()` |

**Consequences an architect must know:**

1. `updated_at` on `file_jobs` is **never** written by application code — it is
   always the DB. The reaper deliberately does **not** rely on it, because a row
   that is stuck mid-run keeps bumping `updated_at`; the reaper uses
   `COALESCE(picked_at, created_at)` instead — see §7.11.1.
2. `bump_brain_version` fires **per row**, so
   `UPDATE kachra_patterns SET hit_count=hit_count+1` for N patterns bumps the
   version N times. Harmless (version is a counter, not a content hash) and
   intentional: the store only treats it as a change *signal*.
3. `AutoPromote` inserts into `kachra_patterns` inside a transaction; the bump
   therefore commits atomically with the promote. No lost-wakeup window.

### 4.8 `vacuum_job_stats` — the aggregate view (rewritten in 087, 088, 089, 092)

Final shape (092):

```sql
SELECT status,
       COUNT(*)::bigint                                                        AS cnt,
       COALESCE(SUM(chunk_count),0)::bigint                                    AS total_chunks,
       COALESCE(AVG(duration_ms),0)::bigint                                    AS avg_duration_ms,
       COALESCE(SUM(download_count),0)::bigint                                 AS total_downloads,
       COALESCE(SUM(CASE WHEN preservation_verified THEN 1 ELSE 0 END),0)::bigint AS preserved_cnt,
       COALESCE(SUM(llm_kachra_count),0)::bigint                               AS total_llm_kachra,
       COALESCE(SUM(filler_filtered_count),0)::bigint                          AS total_filler_filtered,
       COALESCE(AVG(eval_score),0)::double precision                           AS avg_eval_score
FROM file_jobs
GROUP BY status;
```

**Why it is rewritten and not extended column-by-column:** a view cannot gain a
column with `ALTER`; each migration that adds a metric must
`DROP VIEW ... ; CREATE OR REPLACE VIEW ...`. Every `.down.sql` reverses this by
recreating the previous shape — which is why the down files for 088/089/092 each
contain a `CREATE OR REPLACE VIEW` of the *older* definition.

**Why `COALESCE(...,0)` everywhere:** `SUM` over an empty group returns NULL, and
a dashboard that receives `null` renders "NaN%". The view guarantees `0`.

**Why `cnt`/`total_chunks` are `::bigint`:** `COUNT(*)` is `bigint` and `SUM(int)`
is `bigint` in Postgres; averaging a `bigint` yields `numeric`, which some Go
scanners decode awkwardly. Casting keeps the driver happy and types stable.

### 4.9 Schema invariants summary (the contract)

| # | Invariant | Enforced by |
|---|---|---|
| S1 | One job per storage key | `idx_file_jobs_s3_unique` |
| S2 | `phase` ∈ [0,6] | `file_jobs_phase_check` |
| S3 | `status` ∈ the 8-value set | `file_jobs_status_check` (093) |
| S4 | `chunk_index` unique per job and 0-based contiguous | `UNIQUE(file_job_id, chunk_index)` + app check |
| S5 | Deleted job ⇒ chunks deleted | FK `ON DELETE CASCADE` |
| S6 | Any pattern change bumps brain version | `trg_kachra_bump` |
| S7 | `updated_at` always fresh on update | `set_updated_at` triggers |
| S8 | Pattern uniqueness case/whitespace-insensitive | unique expression index |
| S9 | Category vocabulary is exactly 8 values | CHECK constraints (091) |
| S10 | `brain_version` has exactly one row | `CHECK (id=1)` |

<--CURSOR-->





