# Feature #8: Real-time Ingestion Progress

## Overview
Ingestion progress has two cooperating layers: a **durable, append-only event
timeline** (`ingestion_job_events`, Postgres) streamed over SSE, and the
**hot job row** (`ingestion_jobs`) used for the progress bar / ETA / cost.

> **UPDATE (T1–T4, training transparency).** The original implementation below
> described a 1-second DB **poll** inside the SSE handler, a browser-memory-only
> event log, fully sequential batch processing, and per-chunk INSERTs. All four
> were replaced:
>
> | Concern | Before | Now |
> |---|---|---|
> | Event log | Built client-side from snapshots; **lost on refresh**, different per viewer | Appended to `ingestion_job_events`; SSE replays from Postgres, so every viewer sees the identical history |
> | SSE transport | 1s poll of `ingestion_jobs`, re-sent even when unchanged | Redis pub/sub wake-up + Postgres read; the 2s snapshot is only for ETA/cost reconciliation |
> | Double confirmation | Admin had to run SQL to check the claimed counts | `verified` event compares the pipeline's claim against real `course_chunks` aggregates (claim vs reality in the UI) |
> | Topic + embed batches | Sequential (one batch at a time) | Worker pool, `INGESTION_WORKERS` (default 3) with a semaphore |
> | Chunk storage | 1 INSERT + up to 2 link UPDATEs **per chunk** (~45k round trips for 15k chunks) | 200-row batched INSERT + one id-resolution query per batch + batched `unnest` link UPDATE |
>
> Two latent resume bugs were also fixed while in there: topic/embed batches
> were being *skipped* on resume even though their results only ever lived in
> memory until the store step — which silently produced `topic="general"`
> chunks (topic stage) and nil embeddings (embed stage). Resume now reuses
> stored topics only when they are actually in the DB, and always recomputes
> embeddings (local sidecar, so it is cheap).
>
> New endpoint for history: `GET /admin/experts/:id/jobs/events?jobId=<uuid>`
> (`after`/`limit` cursors). `jobId` is a query param because
> `GET /experts/:id/jobs/stream` already owns that route tree position.

## Supported upload formats (D3)

Uploads are no longer assumed to be plain text. Every file is converted to text
**before** the pipeline runs, as **stage 0 (`extracting`)** — visible in the job
row and on the timeline.

| Group | Extensions | How it is converted |
|---|---|---|
| Plain text (decoded in the API process) | `.txt` `.text` `.md` `.markdown` `.json` `.log` `.xml` | In-process decode (BOM, UTF‑16, UTF‑8, latin‑1 fallback). **Works even if the ML sidecar is down.** |
| PDF | `.pdf` | pypdf; repeated page headers/footers removed; `# Page N` markers |
| Word | `.docx` `.docm` | python-docx (paragraphs + tables) |
| Word (legacy) | `.doc` | catdoc (falls back to LibreOffice when the image is built with it) |
| Excel | `.xlsx` `.xlsm` | openpyxl; `# Sheet: name` + `column: value \| column: value` rows |
| Excel (legacy) | `.xls` | xlrd, then xls2csv / LibreOffice |
| Delimited | `.csv` `.tsv` | header-aware row rendering |
| PowerPoint | `.pptx` | python-pptx; `# Slide N` + bullets + speaker notes |
| PowerPoint (legacy) | `.ppt` | catppt (falls back to LibreOffice) |
| Web / ebook | `.html` `.htm` `.xhtml` `.epub` | BeautifulSoup (+ `# Chapter N`), scripts/styles/nav stripped |
| Rich text | `.rtf` | striprtf |
| OpenDocument | `.odt` `.ods` `.odp` | odfpy |
| Subtitles | `.srt` `.vtt` | cue numbers and timecodes stripped, spoken text kept |

**Limits & failure modes** (each is reported with a stable reason code and an
admin-readable message on the timeline):
`unsupported_format`, `extension_mismatch` (a `.txt` that is really a PDF),
`pdf_encrypted`, `pdf_no_text` (scanned PDF — OCR is **not** enabled),
`too_large` (above `DOC_MAX_EXTRACTED_CHARS`, default 5,000,000),
`too_many_pages` (`DOC_MAX_PAGES`, default 2000), `timeout`, `corrupt`,
`extractor_unavailable` (ML sidecar down — nothing was ingested, retry).

**Where parsing happens, and why:** in the ML sidecar (`POST /extract`), never in
the API process. Parsing untrusted documents is a hostile-input job and the
sidecar is non-root with no DB credentials. The trade-offs:

- Legacy `.doc`/`.ppt`/`.xls` use the **catdoc** tools (~2 MB, installed by
  default). Exotic formats (`.wpd`, `.pages`, …) need LibreOffice, which is an
  **opt-in** build arg (`--build-arg WITH_LIBREOFFICE=true`, ~500 MB).
- The 50 MB upload limit is unchanged; extraction adds a 5M-character cap
  because DOCX/XLSX/PPTX/EPUB are ZIP containers and a small upload can expand
  enormously (zip bomb).
- Extraction writes **no checkpoint**: on failure the admin uploads again.
  Resume/retry are unaffected because they read the previously extracted text
  from `ingestion_jobs.transcript_content`.

## Status: ✅ IMPLEMENTED

Includes:
- Durable event timeline + true-push SSE
- Step-by-step stage visualization
- Error details and recovery options
- Resume from checkpoint functionality
- Live event log backed by the database

## Architecture

### Backend
- **File**: `backend-go/internal/admin/admin_handler.go` (stream + history)
- **Timeline**: `backend-go/internal/jobevents/` (store + subscriber)
- **Endpoint**: `GET /api/v1/admin/experts/:id/jobs/stream`
- **History**: `GET /api/v1/admin/experts/:id/jobs/events?jobId=<uuid>`
- **Protocol**: Server-Sent Events (SSE)
- **Event push**: immediate (Redis pub/sub wake-up → Postgres read)
- **Snapshot reconciliation**: every 2 seconds (ETA / cost / status)
- **Heartbeat**: every 15 seconds

### Frontend
- **Hook**: `frontend/src/hooks/useIngestionStream.ts`
- **Component**: `frontend/src/components/admin/IngestionPipelineModal.tsx`
- **API**: `frontend/src/api/admin.ts`

## Features

### 1. Real-time Progress Updates
- **Event push, not polling**: timeline events arrive as soon as the pipeline writes them
- **Live connection indicator**: Shows "Live" or "Reconnecting..."
- **Progress bar**: Updates in real-time with percentage
- **Chunk counter**: Shows processed/total chunks

### 1a. Timeline event kinds (`ingestion_job_events.kind`)
| kind | Meaning |
|---|---|
| `run_started` | A run (or resume) began — includes worker count |
| `stage_started` / `stage_done` | Stage boundary; `stage_done` carries `duration_ms` |
| `batch_done` | One topic/embed batch finished (batches done/total, chunks, workers) |
| `chunk_stored` | A block of chunks was committed to `course_chunks` |
| `verified` | **Double confirmation** — claim vs real DB aggregates |
| `paused` / `failed` / `complete` | Terminal states, with reason/ledger |

### 1b. Double confirmation (T2)
After storing, the pipeline re-reads the database and compares it with what it
claims it produced, then records a `verified` event:

```json
{ "ok": true,
  "claim":   { "chunks": 873, "topics": 37, "general_chunks": 0 },
  "reality": { "chunks": 873, "topics": 37, "general_chunks": 0, "null_embeddings": 0 },
  "corpus_total_chunks": 1741 }
```

`ok=false` means the numbers disagree — surfaced in the UI in amber instead of
being silently wrong.

### 1c. Concurrency (T3)
Topic extraction and embedding run batches in a bounded worker pool
(`INGESTION_WORKERS`, default 3 — mirrors `orchestrator.expertMaxConcurrency`).
Checkpoints persist the **contiguous completed prefix**, so resume stays correct
when batches finish out of order. Chunk storage is batched (T4) but its
`prev/next` link pass is a separate sequential step because linking is
order-dependent; the waves-sequential / tasks-parallel shape mirrors
`workflow/runner.go`.

### 2. Stage Pipeline Visualization

**6 Stages**:
1. **Splitting Transcript** (chunking)
2. **Extracting Topics** (LLM)
3. **Extracting Charter** (LLM)
4. **Generating Embeddings**
5. **Saving to Database**
6. **Running Smoke Test**

**Visual States**:
- ⏳ **Waiting**: Gray, 40% opacity
- 🔵 **Active**: Blue glow, pulsing indicator
- ✅ **Done**: Green background
- ❌ **Failed**: Red background

### 3. Progress Details

**Top Section**:
- Processed chunks: `1,234 / 5,678 chunks`
- Progress percentage: `45%`
- Animated progress bar with glow effect

**Bottom Section**:
- ETA: `~2m 30s remaining`
- Cost: `$0.0234 / ₹1.97`

### 4. Error Handling

**Failed State**:
```
┌─────────────────────────────────────┐
│ ❌ Failed                           │
│ Error: LLM API rate limit exceeded  │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ↺ Resume Available                  │
│ Checkpoint: embedding               │
│ No re-upload needed                 │
│                                     │
│         [Resume from checkpoint]    │
└─────────────────────────────────────┘
```

**Paused State** (Charter LLM failure):
```
┌─────────────────────────────────────┐
│ ⏸ Paused                            │
│ Charter LLM failed. Waiting for     │
│ admin action.                       │
│ Checkpoint: charter_extraction      │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ↺ Resume Available                  │
│ Fix the issue (e.g. top up API      │
│ credits), then resume.              │
│                                     │
│         [Resume from checkpoint]    │
└─────────────────────────────────────┘
```

### 5. Resume from Checkpoint

**Endpoint**: `POST /api/v1/admin/experts/:id/jobs/:jobId/resume`

**Features**:
- No re-upload required
- Resumes from last successful stage
- Preserves all completed work
- Shows checkpoint stage in UI

**Checkpoint Stages**:
- `pending` → Need transcript
- `chunking` → Need transcript
- `topic_extraction` → Chunks in DB, no transcript needed
- `charter_extraction` → Chunks + topics in DB
- `embedding` → Everything except embeddings
- `storing` → Everything except final storage
- `smoke_test` → Everything except smoke test

### 6. Live Event Log

**Features**:
- Newest events at top
- Timestamp for each event (HH:MM:SS.mmm)
- Color-coded by type:
  - 🔵 **Connected**: Cyan
  - ⚪ **Update**: Gray
  - ✅ **Complete**: Green
  - ❌ **Failed**: Red
  - ⏸ **Paused**: Amber

**Example Log**:
```
14:23:45.123  ✅ Complete — 5,678 chunks stored
14:23:44.890  storing — 99%
14:23:43.567  embedding — 98% — $0.0234
14:23:42.234  embedding — 95%
14:23:40.901  charter_extraction — Extracting charter...
14:23:35.678  topic_extraction — Extracting topics...
14:23:30.345  chunking — Splitting transcript...
14:23:25.012  🔵 SSE connection established
```

## SSE Event Types

### 0. `hello` (initial frame)
```json
{ "type": "hello", "job": { "id": "...", "status": "running", "currentStage": "embedding" }, "ts": "..." }
```
- Sent once on connect with the current job snapshot, immediately followed by a
  replay of the whole timeline as `event` frames.

### 0b. `event` (durable timeline row)
```json
{
  "type": "event",
  "event": {
    "id": "...", "jobId": "...", "expertId": "...",
    "sequenceNumber": 42, "stage": "embedding", "kind": "batch_done",
    "detail": { "batch_index": 7, "batches_done": 8, "batches_total": 40,
                "chunks_done": 200, "chunks_total": 873, "workers": 3 },
    "createdAt": "2026-09-25T14:23:43.567Z"
  },
  "ts": "..."
}
```
- The authoritative timeline. Delivered live and replayed on connect/reconnect
  from Postgres, so it survives a refresh and is identical for every viewer.
- Old clients that only understand `update`/`complete`/`failed` simply ignore
  these frames — the protocol stayed backward compatible.

### 1. `heartbeat`
```json
{
  "type": "heartbeat",
  "ts": "2026-09-22T14:23:45.123Z"
}
```
- Sent every 5 seconds
- Keeps connection alive through proxies
- No state change

### 2. `waiting`
```json
{
  "type": "waiting",
  "ts": "2026-09-22T14:23:45.123Z"
}
```
- No job found yet
- Waiting for ingestion to start

### 3. `update`
```json
{
  "type": "update",
  "job": {
    "id": "...",
    "status": "running",
    "currentStage": "embedding",
    "stageDetail": "Processing chunk 1234/5678",
    "totalChunks": 5678,
    "processedChunks": 1234,
    "costUsd": 0.0234,
    "estimatedSecondsRemaining": 150
  },
  "ts": "2026-09-22T14:23:45.123Z"
}
```

### 4. `complete`
```json
{
  "type": "complete",
  "job": {
    "id": "...",
    "status": "complete",
    "totalChunks": 5678,
    "processedChunks": 5678,
    "costUsd": 0.0234
  },
  "ts": "2026-09-22T14:23:45.123Z"
}
```
- Connection closes automatically

### 5. `failed`
```json
{
  "type": "failed",
  "job": {
    "id": "...",
    "status": "failed",
    "currentStage": "embedding",
    "errorMessage": "LLM API rate limit exceeded",
    "processedChunks": 1234,
    "totalChunks": 5678
  },
  "ts": "2026-09-22T14:23:45.123Z"
}
```
- Connection closes automatically
- Resume button appears

### 6. `llm_failure_decision_required` (Paused)
```json
{
  "type": "llm_failure_decision_required",
  "job": {
    "id": "...",
    "status": "paused",
    "currentStage": "charter_extraction",
    "errorMessage": "Charter LLM failed: 402 Payment Required",
    "processedChunks": 5678,
    "totalChunks": 5678
  },
  "ts": "2026-09-22T14:23:45.123Z"
}
```
- Connection closes automatically
- Retry button appears
- Admin must fix issue (e.g., top up credits)

## Connection Management

### Auto-reconnect
- Reconnects after 3 seconds on connection loss
- Shows "Reconnecting..." indicator
- Stops reconnecting when job is done/failed/paused

### Token Authentication
- SSE cannot send Authorization headers
- Token passed as query parameter: `?token=...`
- Backend reads from query param for SSE endpoints

### Connection Lifecycle
```
1. Open connection
2. Receive "connected" event
3. Receive updates every 1s
4. Receive heartbeats every 5s
5. Job completes/fails/pauses
6. Connection closes
7. No more reconnects
```

## Progress Calculation

### Formula
```typescript
function calcProgress(
  stage: string,
  processed: number,
  total: number,
  status: string
): number {
  if (status === 'complete') return 100
  
  if (status === 'failed') {
    const stageIdx = STAGE_ORDER[stage] ?? 0
    return Math.max((stageIdx / 7) * 100, 5)
  }
  
  const stageIdx = STAGE_ORDER[stage] ?? 0
  const stageBase = ((stageIdx - 1) / 7) * 100
  const stageWidth = (1 / 7) * 100
  
  if (total > 0 && processed > 0) {
    return Math.min(
      stageBase + (processed / total) * stageWidth,
      99
    )
  }
  
  return Math.max(stageBase, 0)
}
```

### Stage Order
```typescript
const STAGE_ORDER = {
  pending: 0,
  chunking: 1,
  topic_extraction: 2,
  charter_extraction: 3,
  embedding: 4,
  storing: 5,
  smoke_test: 6,
  complete: 7,
  failed: -1,
  paused: -1,
}
```

### Progress Bar Colors
- **Running**: Blue with glow
- **Complete**: Green with glow
- **Failed**: Red
- **Paused**: Amber with glow

## Cost Tracking

### Display Format
```
$0.0234 / ₹1.97
```

### Conversion Rate
- USD to INR: 1 USD = 84 INR
- Hardcoded in component

### Cost Sources
- LLM API calls (topic extraction, charter extraction)
- Embedding API calls
- Tracked per job in `cost_usd` column

## ETA Calculation

### Backend Logic
- Tracks time per chunk
- Calculates average processing time
- Estimates remaining time based on remaining chunks
- Updates every second

### Display Format
- `~30s` (under 1 minute)
- `~2m 30s` (under 1 hour)
- `~1h 15m` (over 1 hour)

## Error Recovery

### Automatic Checkpoint Saving
- Checkpoint saved after each stage
- Stored in `checkpoint_data` JSONB column
- Includes:
  - Current stage
  - Processed chunks
  - Extracted topics
  - Charter (if extracted)
  - Embeddings (if generated)

### Resume Logic
1. Load checkpoint from DB
2. Determine last successful stage
3. Skip completed stages
4. Resume from next stage
5. No re-upload required (unless chunking failed)

### Transcript Storage
- Stored in `transcript_content` TEXT column
- Required for resume if chunking failed
- Optional if chunks already in DB

## Performance Considerations

### Backend
- SSE connection per expert
- 1 DB query per second
- Minimal CPU usage
- No polling overhead

### Frontend
- Single EventSource connection
- Efficient state updates
- Auto-scroll to newest log entry
- Max 200 log entries kept in memory

### Network
- ~100 bytes per update
- ~6 KB/minute
- ~360 KB/hour
- Negligible bandwidth

## Browser Compatibility

### EventSource Support
- ✅ Chrome 6+
- ✅ Firefox 6+
- ✅ Safari 5+
- ✅ Edge 79+
- ❌ IE (not supported)

### Fallback
- No fallback implemented
- Modern browsers only
- IE users see error message

## Testing Checklist

### Happy Path
- [ ] Start ingestion
- [ ] Modal opens automatically
- [ ] See "Live" indicator
- [ ] Progress bar animates
- [ ] Stages update in sequence
- [ ] ETA decreases
- [ ] Cost increases
- [ ] Log shows events
- [ ] Job completes
- [ ] "Ingestion Complete" shows

### Error Handling
- [ ] Disconnect network mid-ingestion
- [ ] See "Reconnecting..." indicator
- [ ] Reconnect network
- [ ] Updates resume
- [ ] No data loss

### Resume Flow
- [ ] Start ingestion
- [ ] Kill backend mid-ingestion
- [ ] Job fails
- [ ] See error message
- [ ] See "Resume" button
- [ ] Click Resume
- [ ] Job resumes from checkpoint
- [ ] No re-upload required

### Paused Flow
- [ ] Start ingestion
- [ ] Charter LLM fails (402)
- [ ] Job pauses
- [ ] See "Paused" message
- [ ] See "Resume" button
- [ ] Top up API credits
- [ ] Click Resume
- [ ] Job resumes from charter stage

## Known Issues

- `INGESTION_WORKERS` higher than the provider's rate limit (topic extraction)
  or the sidecar's CPU (embedding) does not add throughput — it just moves the
  queue to the provider. 3 is the tuned default.
- Batch-level concurrency means `processed_chunks` in the job row advances in
  *chunk* batches, not strictly one batch at a time; the timeline is the
  accurate per-batch record.

## Rollback

Migration `036_ingestion_job_events.down.sql` drops the timeline. Every emit
site is nil-safe and the SSE handler falls back to snapshot-only updates, so
dropping the table degrades the feature instead of breaking ingestion.

## Future Enhancements

### 1. Pause/Cancel Button
- Allow admin to pause ingestion manually
- Save checkpoint and stop processing
- Resume later

### 2. Parallel Ingestion
- Process multiple experts simultaneously
- Show all jobs in a list
- Click to view details

### 3. Ingestion History
- Show past ingestion jobs
- Filter by status/date
- Export logs

### 4. Cost Alerts
- Set cost threshold
- Alert when exceeded
- Auto-pause on limit

### 5. Progress Notifications
- Browser notifications
- Email on completion
- Slack integration

## Conclusion

The ingestion progress feature is **fully implemented** and includes:
- ✅ Real-time SSE updates (no polling)
- ✅ Step-by-step stage visualization
- ✅ Error details and recovery
- ✅ Resume from checkpoint
- ✅ Live event log
- ✅ Cost tracking
- ✅ ETA calculation
- ✅ Auto-reconnect

**No additional work required for Feature #8.**
