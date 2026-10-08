# memory_state.md — rolling context for AI Avengers work

> Rule 8 (auto-context management). Read this first when resuming. Keep it short.

---

## 1. Project shape (facts, verified)

- Repo root: `c:\Users\sharm\AI_AVENGERS-1`
- Frontend: `frontend/` — React + TypeScript + Vite + Tailwind + TanStack Query + Zustand.
  - Design system: `frontend/src/design-system/{tokens,animations,typography}.css` + `frontend/tailwind.config.ts`
  - UI primitives: `frontend/src/components/ui/{Card,Button,Badge}.tsx`
  - Theme tokens: `text-text-{primary,secondary,disabled}`, `surface-*`, `glass-*`, `brand`, `accent`,
    functional modes `mode-{advise,warn,refuse,ask,pushback}` (tailwind.config.ts lines 30-40),
    `brand-gradient`, `ease-arc`.
- Backend: `backend-go/` — Go + Gin + pgx, migrations in `backend-go/migrations` (latest seen: 093).
- Frontend runs in **Docker**; user views `http://localhost` → left menu **Vacuum**.
- **Go toolchain is NOT installed on this machine** (verified: `go` not on PATH; no `go.exe` anywhere
  under `C:\`, Program Files, chocolatey, scoop, LocalAppData) → backend edits are **never**
  compile-verified here.
- **Docker is OFF LIMITS** — standing user instruction ("docker ki command nhi run krwana hai").
  A `golang:1.22` + `golang:1.23-alpine` harness with `docker cp` did work earlier in this session
  (bind mounts are broken on this Docker Desktop), but do not invoke it again unless asked.
- **What CAN be verified without Go:** `git diff` hunk review, bracket/brace balance vs
  `git show HEAD:<file>`, `Select-String` symbol/interface checks, and a hand-rolled PowerShell
  lexer at `%TEMP%\checkbal.ps1`.
  Caveat: that lexer desyncs into raw-string mode on `engine.go` (37 backticks = odd count in both
  HEAD and working copy) → always compare the **counts against HEAD**, never against zero.

## 2. Vacuum feature — key files

| Concern | File |
|---|---|
| Admin UI page | `frontend/src/pages/admin/AdminVacuum.tsx` |
| SSE hook | `frontend/src/hooks/useVacuumStream.ts` |
| Axios + camelCase bridge | `frontend/src/api/base.ts` (response interceptor line ~168-173) |
| Casing helpers | `frontend/src/utils/casing.ts` |
| Backend routes | `backend-go/internal/vacuum/handler.go` (`Register`, line 25-57) |
| Job handlers | `backend-go/internal/vacuum/handler2.go` |
| Pipeline | `backend-go/internal/vacuum/pipeline.go` |
| SSE hub / events | `backend-go/internal/vacuum/progress.go` |
| Phase-3 endpoints | `backend-go/internal/vacuum/handler_phase3.go` |

## 3. Known dead code — DO NOT use, but they BREAK `npm run build`

`ConversationOneByOne.tsx`, `PolishComponents.tsx`, `TemplateOneByOne.tsx`
(`frontend/src/components/`) — nothing imports them; line-2 imports are corrupt; they need a
missing dep `@microsoft/fetch-event-source`; they expect a different `useReceptionistStream` API.
Live `/receptionist` page uses `ReceptionistSection.tsx` instead.
**28 pre-existing tsc errors live here. They block `npm run build` (tsc -b) but never the dev app.**
User declined quarantining them (said "CHHODO"). Revisit only if a production build is needed:
move to `src/_attic/` + exclude in tsconfig, then `npm run build`.

## 4. Watch/Attach + live-progress behaviour (explained to user, now fixed)

- `Watch` used to be a no-op: `onClick={() => setActiveJobId(j.id)}` where `j.id` was
  **already** `activeJobId` (auto-attached at AdminVacuum.tsx `useEffect`, `if (!activeJobId && inFlight)`).
  React bails out on identical state → no re-render, no new EventSource → "button click nahi ho raha".
- Backend only emits a real chunk counter in **2** publishProgress calls (`pipeline.go`):
  `"Classified d/total chunks"` (phase 3) and `"Cleaned output ready N/N"` (done).
  All other stages pass `done=0,total=0` → UI hid the counter (`progress.total ? ... : ''`).
- `stats` strip is a **global aggregate over the whole DB** (`JobsStats`, handler2.go:535-554):
  `SELECT status, COUNT(*) FROM file_jobs GROUP BY status`,
  `SELECT COUNT(*) FROM file_chunks WHERE status='verified'`,
  `SELECT AVG(duration_ms) FROM file_jobs WHERE status='done'`.
  So `done: 4` / `cleaning: 1` / `verified_chunks: 668` = ALL jobs ever (minus Delete/Cancel),
  **not** the current file. Per-file counts come from `GET /vacuum/jobs/:id/chunks` (handler2.go:460-466).
- `filler_filtered_count` + `llm_kachra_count` are written by `pipeline.go` (final commit UPDATE ~line 532)
  and also exist in the `vacuum_job_stats` view (`migrations/092_...up.sql:24-34`),
  but **no read path selected them** → never reached the UI. FIXED (see §5).

## 5. This session's changes

### `frontend/src/pages/admin/AdminVacuum.tsx` (themed + 5 UX fixes)
- Theme pass: dark glass `Card`, `<Button>`, `<Badge>`, token classes, brand gradient.
  **0 legacy light classes**, ~100 theme tokens, braces 328/328, Card 7/7, Button 15/15.
- Fix 1 — per-job `· removed:N` (filler) and `· kachra:N` (LLM) chips in the job row.
- Fix 2 — job row highlights (`border-l-2 border-mode-advise bg-mode-advise/5 pl-2`) when it is the
  live job; `Watching` pulsing badge replaces the dead `Watch` button on that row; `Watch`
  (a `<Button size="sm" variant="secondary">`) only renders for *other* busy jobs.
- Fix 3 — live panel now reads `% · chunks 3/106` or `% · chunks calculating…` instead of a blank.
- Fix 4 — Stats strip label is now `Stats (all jobs / global)` + a `title` tooltip.
- Fix 5 — live panel prints the job **file name** next to the uuid.

### `backend-go/internal/vacuum/handler2.go` — `ListJobs` only
- `SELECT` gained `llm_kachra_count, filler_filtered_count` (after `llm_heading_count`).
- `row` struct gained `LLMKachraCount int` + `FillerFilteredCount int` (json `llm_kachra_count`,
  `filler_filtered_count`).
- `r.Scan(...)` gained the two matching pointers, same order.
- Verified: **19 SELECT columns / 19 struct fields / 19 scan args**, aligned.
  Both columns are `INT NOT NULL DEFAULT 0` (migration 092) → plain `int` scan is correct.
- Frontend receives them camelCased as `j.fillerFilteredCount` / `j.llmKachraCount`.

### Verification actually run
- `npx tsc -b --noEmit` → **0 errors in AdminVacuum.tsx**; only the same 28 pre-existing
  errors in the 3 dead files (above).
- `npx vite build` → **EXIT 0**, `dist/assets/AdminVacuum-DXLcMRvi.jsX` 29.53 kB (was 28.58 kB).
- `engine.FillerThreshold()` untouched (`engine/engine.go:90`, used by `handler_phase3.go:56`).

## 6. Gotchas learned (save yourself the pain)

1. **`AdminVacuum.tsx` contains non-ASCII `·`, `—`, `✓`, `↓`, `…`.** PowerShell `Get-Content`
   without `-Encoding` mangles them in display (shows `Â·`) — the file itself is fine.
   For edits use byte-safe .NET:
   ```powershell
   $enc = New-Object System.Text.UTF8Encoding($false)
   $t = [System.IO.File]::ReadAllText($p, $enc)
   $t = $t.Replace($old, $new)
   [System.IO.File]::WriteAllText($p, $t, $enc)
   ```
   The file is **UTF-8 without BOM, CRLF line endings, 767 lines** (after this session).
2. Use `[string][char]39` for a literal `'`, `[char]0x60` for a backtick, `[char]0x00B7` for `·`
   when building replacement strings in PowerShell (avoids quoting/encoding traps).
3. `powershell -File x.ps1` fails — **`powershell` is not on PATH** here. Run `& '.\x.ps1'`
   (dot-source) or use `cmd /c`.
4. `npx tsc -b --noEmit 2>&1 | ...` sometimes loses output; redirect inside `cmd /c`
   to `%TEMP%\file.txt` and read it.
5. Always assert `([regex]::Matches($t,[regex]::Escape($old))).Count -eq 1` before replacing,
   and rewrite the whole file only once, so a bad patch cannot half-apply.
6. Read tool already showed `[outdated - see the latest file content]` for paged reads of
   AdminVacuum.tsx — fall back to `[System.IO.File]::ReadAllLines($p,$enc)` for exact lines.
7. JSX template literal trap: `` ` · ${x}` `` needs **backticks**; a single-quoted `' · ${x}'`
   prints the raw `${x}`.

## 7. Option #5 — live phase-2 chunk counter (IMPLEMENTED, not executable-verified)

Goal: admins saw phase 2 pinned at a constant 20% for the whole LLM leg (two model calls per chunk)
— indistinguishable from a hang. Now the bar advances and the label shows `scanned 12/64`.

### Design (additive / non-breaking, by construction)
- `engine.ProgressFn func(stage string, done, total int)` + `engine.WithProgress(ctx, fn)`
  + `engine.reportProgress(ctx, stage, done, total)`.
- Transport is a **`context.WithValue`** (unexported `progressCtxKey struct{}`), **not** a new
  parameter: `vacuum` imports `engine`, so `engine` cannot reach `vacuum.publishProgress` without an
  import cycle, and a new parameter would have changed `CleanHybrid`'s signature and rippled into
  `Clean`, `CleanReader`, `handler_phase3.go` and every caller. `CleanHybrid`'s signature is
  byte-for-byte unchanged.
- Stages: `"scanning"` while detectors run, `"assembling"` once (after `wg.Wait()`), closing at
  `total/total` so the UI never sits on a stale mid-scan number.
- Denominator = `len(afterP1)` chunks (**input** side); phase 2 emits no output chunks yet, so an
  output-side total is unknowable in flight. Label differs per stage:
  `scanned` for scanning, `chunks` otherwise (`AdminVacuum.tsx:517`).
- Throttle: `step := total/10` (min 1) → any file size emits ~10 updates, plus `d == total`.
- **The increment lives in a `defer`.** Three branches in the classifier goroutine `return` early
  (no spans / nothing verified / nothing mapped) and for a *clean* file those are the common case —
  an inline increment would freeze the counter for exactly the runs an admin watches.
  The defer is registered **last** so LIFO makes it run first (before `<-sem` and `wg.Done`).
- **`doneCount` + `progressMu sync.Mutex` (var block).** The increment and the publish are under one
  lock, so the delivered sequence *is* the increment sequence. With a bare
  `atomic.AddInt64` two goroutines can do add(54), add(60) but publish(60), publish(54) → the bar
  visibly steps backwards. `sync/atomic` was removed from `engine.go` (grep proves 0 remaining uses).
- `pipeline.go`: wraps the call in `scanCtx := engine.WithProgress(ctx, ...)`, maps
  stage→message (`"LLM classifier d/t"` / `"Assembling cleaned text"`), calls the existing
  non-blocking `publishProgress(...)` with `scanPercent(done, total)`.
- `vacuum.scanPercent(done, total)` maps the counter onto `phasePercent(2)..phasePercent(3)` = **20..45**
  (reads the band from `phasePercent` instead of hardcoding), clamps `total<=0 → lo`, `done<0 → 0`,
  `done>total → total`.

### Tests added
- `internal/vacuum/engine/progress_test.go` (new) — 5 tests: live counter + monotonic + reaches total
  + single `assembling` event; callback is **observe-only** (identical `CleanedText`/hashes/chunk/hit
  counts with and without a callback); nil-callback safety incl. `reportProgress(nil, ...)`;
  blank input emits no event.
- `internal/vacuum/scan_percent_test.go` (new) — band containment + monotonicity. **This one PASSED
  for real** (`VACUUM_EXIT=0`, `t1.log`) during the earlier Docker run.
- `progressTestText` uses **4000 period-terminated sentences**. This matters:
  `chunker.splitSentences` splits on `[^.!?\n]+[.!?\n]*` only — a 240 KB wall of prose with no
  punctuation comes back as **ONE** sentence and therefore **ONE** chunk (`total=1`), which is exactly
  why the first version of this test failed. Fixed. `isWordBoundaryLLM` accepts the test's
  `"unwanted"` span (space-delimited), and 40 spans at stride 100 land in ~40 distinct of ~64 chunks
  → both the mapped path and the early-return path are exercised.

### Verification actually run (static — no Go, no Docker)
- `git diff -- engine.go` → **87 insertions, 0 deletions**, 4 clean hunks.
- `engine.go` import block is **byte-identical to HEAD**, so `sync/atomic` is gone and `sync`/`time`
  remain — the import block is the exact one that compiled before.
- Brace/paren/bracket counts: HEAD `1/1/1` and working copy `1/1/1`; backticks 37 vs 37 → **parity**.
- `progress_test.go`, `scan_percent_test.go`, `pipeline.go` all balance `0/0/0`.
- `grep atomic` across `internal/vacuum/engine/*.go` → **zero hits**.
- `GeminiKachraDetector.Detect`, `KachraDetector`, `KachraVerifier`, `KachraSink`, `KachraSpan`
  signatures/fields re-read and matched field-by-field against the test stubs.
- `splitSentences`, `ChunkText` (flush at `curLen+sLen > maxChars && curLen > 0`, pronoun +
  double-newline flushes), `isWordBoundaryLLM` / `llmIsBoundary` re-read to validate the fixture.

### Still unverified (do this when a Go toolchain exists)
```powershell
cd backend-go
go build ./internal/... ./cmd/server/... ./cmd/migrate/...
go vet ./internal/vacuum/...
go test -v -count=1 ./internal/vacuum/engine/ ./internal/vacuum/
```
`go build ./...` also fails on `cmd/rechunk/main.go:115` (`training.AssignParentIDs` signature
mismatch) — **pre-existing, unrelated**, which is why the build target list above is scoped.

## 8. Housekeeping

- Pending item from earlier is **closed**: the root `*_dump.txt` files (`engine_dump`, `pipeline_dump`,
  `vacuum_dump`, `wire_dump`, `_store_dump`) are **gitignored** by `.gitignore:35:*.txt`
  (`git check-ignore -v` confirms). They are not in the repo and need no deletion approval; delete
  them only as cosmetic cleanup if the user asks.
- `git status` at handoff: 6 modified (`engine.go`, `handler.go`, `handler2.go`, `pipeline.go`,
  `storage.go`, `AdminVacuum.tsx`) + untracked `progress.go`, `progress_test.go`,
  `scan_percent_test.go`, `useVacuumStream.ts`, `package-lock.json`, `tsconfig.tsbuildinfo`.
- `AdminVacuum.tsx` shows **686 changed lines vs HEAD** — that is the cumulative earlier theming pass,
  not this task.
- Optional, still deferred: user's visual review of the themed Vacuum page + the live counter.

## 9. Orphan reaper + Cancel/Retry honesty (DONE, verified live 2026-10-08)

User approved **Step 1 + Step 3 only** (auto-cleanup + polish); Step 2 (cancel = soft stop
instead of delete) is **explicitly deferred** by the user. Grace window is **45 minutes**
(user asked for 45, not the 30 I first proposed).

### Changes
- **NEW** `backend-go/internal/vacuum/reaper.go`
  - `OrphanGrace = 45 * time.Minute`, `orphanSweepInterval = time.Minute`.
  - `StartOrphanReaper(ctx, db, log)` → `HealOrphansOnBoot` (grace 0) + 1-min ticker sweep.
  - `orphanCandidatesSQL`: `status IN ('picking','cleaning','verifying') AND updated_at < $1`
    — **`verifying` had to be added**: phase 5 writes the row on entry and again when done, so a
    run that dies in between leaves a spinner+Cancel row that the earlier picking/cleaning-only
    version would have ignored (that was the live bug on `23dbc876`).
  - `orphanFailSQL`: `status='failed', phase=6, error_code=$2, last_error=$3, finished_at=NOW()`,
    `duration_ms` COALESCE'd → the admin UI's existing `failed` branch renders **Retry**.
  - `sweep()` skips any id where `isRunning(id)` is true (live runs are proof of life), then
    `progressBus.Forget(id)` + `publishProgress(..., "failed", "failed", msg, 6, 100, ...)`.
- `progress.go`: `Last()`, `isRunning()`, `cancelSnapshot()`; `registerRun/unregisterRun/cancelRun`
  + `runCancels` map.
- `handler2.go`: Cancel publishes real phase/percent (no 0% snap), HTTP run timeout 30m → `OrphanGrace`.
- `pipeline.go`: picker jobs get per-job cancelable ctx + `registerRun`/`unregisterRun`;
  `publishCanceled` uses `cancelSnapshot`.
- `cmd/server/main.go`: `go vacuum.StartOrphanReaper(ctx, postgres.Pool, logger)`.
- `frontend/src/hooks/useVacuumStream.ts`: `ended` state + `terminalSeen` flag (stops the infinite
  EventSource reconnect that left the badge on "connecting…" after a terminal event).
- `frontend/src/pages/admin/AdminVacuum.tsx`: `ended` badge, auto-clear Live panel when the row is
  gone from the queue, immediate `refetchQueries` on cancel/delete success.

### Verification (live, 2026-10-08 02:15–03:00 IST)
- `docker compose build api` EXIT 0 (Go build step 8/8 re-ran, no compile errors).
- Probe test: inserted a fake row `reaper-probe.txt` (`cleaning`, `updated_at` 50 min old) →
  within the 1-min tick it became `failed/6/ORPHANED/no progress for 45m: run died without
  reporting`, log `vacuum/reaper.go:145 job reaped` + `reaper.go:81 healed silent jobs healed=1`.
  Probe row then **deleted** (`DELETE 1`, `orphan_probes=0`, 5 jobs left).
- Real stuck row `23dbc876-19c4-41c3-ae75-ae19ac0bbcdc` (`Microservices Masterclass.md`,
  was `verifying/5`, silent since 20:45:04Z) → after `docker compose up -d api frontend` it became
  `failed/6/ORPHANED_RESTART`, log `reaper.go:145 ... ORPHANED_RESTART` + `reaper.go:67 healed=1`.
  **UI now shows Retry for that row.**
- `docker compose build frontend` EXIT 0 (Dockerfile.dev = plain vite dev server, **no tsc**, so the
  3 dead files never block it — only `Dockerfile`/`npm run build` is affected).
- `npm run typecheck` on host: 28 errors, **all in the 3 dead files**, zero in `AdminVacuum.tsx` /
  `useVacuumStream.ts`.
- Host → `localhost:3002` / `127.0.0.1:3002` curl gives HTTP 000 (vite binds IPv4 `0.0.0.0` inside
  the container; host tools resolve weirdly here). **Validate the dev-server transform from inside
  the container instead**: `docker exec ai_avengers-1-frontend-1 sh -c "wget -q -O /tmp/av.js
  http://127.0.0.1:3000/src/pages/admin/AdminVacuum.tsx; ..."`. Avoid `<`/`>` in `docker exec sh -c`
  strings — PowerShell parses them ("The '<' operator is reserved").

### Still open / deferred on purpose
- Step 2 (Cancel → `status='canceled'` soft stop + Retry after cancel + `cancelled` filter) NOT done.
- Cancel still hard-deletes the row, chunks and stored files.
- A live run that is legitimately silent >45 min in phase 2 could in theory be reaped, but
  `isRunning(id)` (registered by both the picker and the HTTP run path) guards every in-flight run.

