# AI Avengers — Implementation Handoff (Domain-Expert Collaboration Layer)

> **Purpose:** Live progress tracker for the domain-expert collaboration + training layer defined in `DOMAIN_EXPERT_COLLABORATION_DESIGN.md`.
>
> **This is NOT the same as `HANDOFF.md`.** That file tracks the pre-existing backend/frontend phases (Phase 1–6). This file tracks the additive collaboration layer (Phases A–F) that sits on top of the existing system.
>
> **Rule for this file:** never mark a component ✅ without evidence. Evidence = commit hash + file path + a checkpoint condition that was verified.

---

## Metadata

| Field | Value |
|---|---|
| Version | 0.1.0 |
| Started | 2026-09-07 |
| Depends on | `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`, `DOMAIN_EXPERT_COLLABORATION_DESIGN.md`, `KNOWLEDGE_HUB.md` |
| Repo | https://gitlab.com/zepto-group3/ai_avengers |
| Branch | main (direct push, no MRs) |
| Owner | Kiran Nogia (client + admin) |
| Implementer | System Design Architect (via Duo Chat) |

---

## Ground-truth assessment of pre-existing work (as of 2026-09-07)

Based on file-inspection audit of existing repo:

| Layer | Claim in `HANDOFF.md` | Actual state | Evidence |
|---|---|---|---|
| Backend Phase 1 (Foundation) | ✅ COMPLETE | Verified — files exist and are non-trivial | `cmd/server/main.go`, `internal/{config,db,auth,middleware,gateway,ml}/*.go` |
| Backend Phase 2 (Training) | Top table says ✅, per-component table says ⏳ PENDING | **Partial — code exists but per-component table contradicts top status** | `internal/training/*.go` files exist (chunker, charter_extractor, capability_builder) but per-component table lists them as PENDING |
| Backend Phase 3 (Core Intelligence) | Same contradiction | **Partial** | `internal/{orchestrator,decision,chinawall,memory,context}/*.go` exist and inspected as real implementations, but per-component table lists as PENDING |
| Backend Phase 4 (Project & Chat) | Same contradiction | **Partial** | `internal/{project,chat,message}/*.go` exist |
| Backend Phase 5 (Advanced) | Same contradiction | **Partial** | `internal/{repo,rating,monitoring}/*.go` exist |
| Backend Phase 6 (Polish & Deploy) | ✅ in top table, ⏳ in per-component | **Unclear** | Some files exist (`monitoring/cost_monitor.go`) |
| Frontend Phase 1 | ✅ with honest gaps documented | Trust the file-authored gap list | See `HANDOFF.md` Frontend Phase 1 section |

**Verdict:** the existing per-component tables in `HANDOFF.md` (marked PENDING) are more trustworthy than the top-level "all complete" table. The code files DO exist and were inspected as real, non-stub Go code. But comprehensive smoke testing has not been performed by this implementer yet.

**Action item:** Phase A of this handoff includes rewriting `HANDOFF.md`'s top-level status table with verified truth.

---

## Phase A — Truth Reconciliation + Expert Schema (in progress)

**Goal:** verify existing state, apply collaboration-layer schema migration, expose new expert config fields.

### Checkpoint condition for Phase A

- Admin can create an expert via UI/API with all new fields (`model_tier`, `temperature`, `top_p`, `loop_pattern`, `max_loop_iterations`, `allowed_tools`, `training_status`).
- Admin uploads a transcript → expert transitions to `training_status='trained'` only after smoke test passes (5 hand-crafted domain questions get non-REFUSE responses with citations).
- `HANDOFF.md` top-level status table has been rewritten with verified truth.

### Components

| # | Component | File | Status | Notes |
|---|---|---|---|---|
| A1 | Design doc: collaboration layer | `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` | ✅ COMMITTED | commit 2026-09-07 |
| A2 | Knowledge Hub | `KNOWLEDGE_HUB.md` | ✅ COMMITTED | this commit |
| A3 | This handoff file | `IMPLEMENTATION_HANDOFF.md` | ✅ COMMITTED | this commit |
| A4 | Verify existing schema (`001` through `005` migrations) | `backend-go/migrations/*.sql` | ✅ COMPLETE (2026-09-07) | All 5 existing migrations read; verified experts/course_chunks/messages/system_settings columns; found that migration numbers 004 and 005 are already taken, and that `llm_calls` table does NOT exist (cost tracked on messages) |
| A5 | Migration: `006_collaboration_layer.up.sql` (new tables + ALTERs) | `backend-go/migrations/006_collaboration_layer.up.sql` | ✅ COMPLETE (2026-09-07) | Renumbered from 004→006. Additive-only. IF NOT EXISTS everywhere, safe to re-run. Adds: experts config columns, course_chunks.chunk_hash, workflows, blackboard_events, workflow_tasks, workflow_checkpoints, approval_requests, messages.workflow_id, two system_settings rows |
| A6 | Migration: `006_collaboration_layer.down.sql` (reversible) | `backend-go/migrations/006_collaboration_layer.down.sql` | ✅ COMPLETE (2026-09-07) | Reverses everything in strict child-before-parent order. Preserves existing `is_training` column (not owned by this migration) |
| A7 | Update `internal/expert/handler.go` to accept new fields | `backend-go/internal/expert/handler.go` | ⏳ NOT STARTED | Preserve backward compatibility |
| A8 | Update `internal/admin/admin_handler.go` expert CRUD | `backend-go/internal/admin/admin_handler.go` | ⏳ NOT STARTED | Admin UI must surface these fields |
| A9 | Chunk deduplication in `internal/training/chunker.go` + `ingestion_pipeline.go` | `backend-go/internal/training/chunker.go`, `ingestion_pipeline.go` | ✅ COMPLETE (2026-09-07) | SHA-256 hash on every chunk. ON CONFLICT DO NOTHING dedup. replaceExisting bool param added. |
| A9b | Caller update: `admin_handler.go` → pass `replaceExisting=false` | `backend-go/internal/admin/admin_handler.go` | ✅ COMPLETE (2026-09-07) | **Broken build fixed.** Was 6-arg call, now 7-arg with false (append mode). |
| A10 | Smoke test step in ingestion pipeline | `backend-go/internal/training/ingestion_pipeline.go` | ⏳ NOT STARTED | 5 hand-crafted questions must return non-REFUSE with citations |
| A11 | Rewrite `HANDOFF.md` top-level status table with verified truth | `HANDOFF.md` | ⏳ NOT STARTED | Requires smoke-testing existing Phase 2–6 code path first |
| A12 | Frontend: expert form fields for new config | `frontend/src/pages/admin/**`, `frontend/src/components/admin/**` | ⏳ NOT STARTED | Depends on A7/A8 API surface |

### Decisions made in Phase A (log)

- **2026-09-07:** locked decision — `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §16 as the authoritative list. Any conflict with existing code must be resolved by either updating design doc (with rationale) or existing code (with migration).
- **2026-09-07:** `KNOWLEDGE_HUB.md` supersedes course transcripts as the primary knowledge source for future implementers. Transcripts are archival only.
- **2026-09-07:** using additive-only migration for schema changes (`ADD COLUMN IF NOT EXISTS`, new tables). No destructive changes to existing schema in Phase A.

### Anti-patterns caught in Phase A

- **Trust-but-verify failure in existing `HANDOFF.md`:** top-level status table said "all complete" but per-component tables said "pending" for the same phases. Fix: rewrite top-level table (item A11) with per-component evidence.
- **No repo access in first design pass:** initial design plan (before repo access was granted) assumed Next.js + Prisma. Reality is Go + Python + React + Vite. Fix: `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` now matches actual stack.
- **Fabricated table reference in own design doc:** `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §15.2 originally said `ALTER TABLE llm_calls ADD COLUMN workflow_id`, but that table does not exist in the schema. Cost tracking is on `messages.cost_usd`, not a separate `llm_calls` table. If I had blindly written the migration without reading `001_initial_schema.up.sql` first, it would have failed on apply. Fix: (a) migration 006 correctly adds `workflow_id` to the real `messages` table; (b) design doc §15.2 updated with the correction and a comment noting the fix, so future readers see how this class of error is caught. Lesson: never trust a design doc's SQL without cross-checking against actual migration files first.
- **Migration number collision avoided:** initial plan was to create `004_collaboration_layer`. Ground-truth check revealed `004_messages_warning_fields` and `005_seed_admin` already occupied 004 and 005. Renumbered to 006. Lesson: `list_repository_tree` on the migrations directory before choosing a migration number is mandatory.
- **Signature change without caller update (2026-09-07):** `IngestTranscript` signature extended with `replaceExisting bool` (7th param) in one commit. `admin_handler.go` caller was NOT updated in the same commit. Build broke on `main`. Fix: always grep for all callers BEFORE committing a signature change, and update them in the SAME commit. Lesson logged in `KNOWLEDGE_HUB.md` §9.8.

### Verification steps still owed for Phase A

- [ ] Run migration 006 up against a live Postgres. Confirm no errors.
- [ ] Run migration 006 down. Confirm all new tables/columns gone, existing tables untouched (`is_training` still present on experts).
- [ ] Run 006 up again. Confirm idempotent (all IF NOT EXISTS guards work).
- [ ] `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'experts';` — confirm new columns present with expected types.
- [ ] `SELECT * FROM system_settings WHERE key IN ('workflow_engine', 'blackboard');` — confirm defaults inserted.

Owner: Kiran (has DB access). I can prepare the exact commands as a runbook when needed.

---

## Phase B — First 3 Experts (not started)

**Goal:** create PM, System Design, and Backend experts with real trained corpora. Prove end-to-end training + retrieval + response flow with 3 experts before scaling to 10.

### Checkpoint condition for Phase B

- Each of the 3 experts answers 5 hand-crafted domain questions with correct citations, no hallucination.
- Charter and clarification questions are surfaced correctly in `ASK` mode when input is vague.

### Components (planned; will be broken down when Phase A completes)

| # | Component | Depends on |
|---|---|---|
| B1 | Curate + upload transcripts for PM expert | A10 |
| B2 | Curate + upload transcripts for System Design expert (use Byte Byte AI + Master Class 1/2/3 transcripts) | A10 |
| B3 | Curate + upload transcripts for Backend expert (use Ultimate Go transcript) | A10 |
| B4 | Smoke test each of the 3 experts | B1, B2, B3 |
| B5 | Tune charter / clarification if smoke tests reveal gaps | B4 |

---

## Phase C — Blackboard + Workflow Engine Skeleton (not started)

**Goal:** implement the collaboration layer that enables multiple experts to work concurrently on one workflow.

### Checkpoint condition for Phase C

- A workflow with 3 experts (from Phase B) runs through `INTAKE → HIGH_LEVEL_DESIGN → simulated approval → DETAILED_DESIGN`, with all events visible in `blackboard_events` table.

### Components (planned)

| # | Component | New file |
|---|---|---|
| C1 | Blackboard store (writes events, publishes to Redis pub/sub) | `backend-go/internal/blackboard/store.go` |
| C2 | Blackboard subscriber (Redis subscribe, cursor management) | `backend-go/internal/blackboard/subscriber.go` |
| C3 | Workflow engine state machine | `backend-go/internal/workflow/engine.go` |
| C4 | Workflow tasks projection (Kanban derivation) | `backend-go/internal/workflow/tasks.go` |
| C5 | Blackboard tools (`PostArtifact`, `AskExpert`, `ReadBlackboard`) hooked into orchestrator | `backend-go/internal/orchestrator/tools.go` |
| C6 | `AskClient` tool + `approval_requests` handler | `backend-go/internal/workflow/approval.go` |
| C7 | API endpoints: `POST /workflows`, `GET /workflows/{id}`, `POST /workflows/{id}/approvals/{approval_id}` | `backend-go/internal/api/workflow_handler.go` |

---

## Phase D — Cross-Verification + Validation Pipeline (not started)

### Checkpoint condition

- Backend expert produces a Go file with a deliberate syntax error → validation rejects → OTA loop fixes → passes → Code Reviewer approves → artifact goes `final`.

### Components (planned)

| # | Component | Notes |
|---|---|---|
| D1 | Reviewer matrix config | `backend-go/internal/workflow/reviewers.go` (per §8.1 of design doc) |
| D2 | Validation pipeline orchestrator | `backend-go/internal/validation/pipeline.go` |
| D3 | Go validators (syntax via `go/parser`, lint via `staticcheck` subprocess) | `backend-go/internal/validation/go_validator.go` |
| D4 | TS validators (`@babel/parser` and `tsc` subprocess) | `backend-go/internal/validation/ts_validator.go` |
| D5 | Sandbox wrapper (CPU/mem/time limits via `exec.CommandContext`) | `backend-go/internal/validation/sandbox.go` |
| D6 | Revision loop (max 3 rounds) | Integrated in D2 |

---

## Phase E — Checkpointing + Cost Governance + Kanban (not started)

### Checkpoint condition

- Kill the server mid-workflow → restart → workflow resumes from last phase checkpoint → completes successfully.
- Kanban shows real-time card movement.

### Components (planned)

| # | Component | Notes |
|---|---|---|
| E1 | Phase checkpoints (write on every phase transition + approval response) | `backend-go/internal/workflow/checkpoint.go` |
| E2 | Cursor-based event replay for resume | Extends `internal/blackboard/subscriber.go` |
| E3 | Extend `cost_monitor.go` with per-workflow rollup | `backend-go/internal/monitoring/cost_monitor.go` |
| E4 | Soft/hard cost limit enforcement | Same file |
| E5 | Kanban API endpoint | `backend-go/internal/api/kanban_handler.go` |
| E6 | Kanban UI (read-only OK for MVP) | `frontend/src/pages/workflows/kanban.tsx` (new) |

---

## Phase F — Remaining 7 Experts + Handoff Packaging (not started)

### Checkpoint condition

- Client runs a full workflow deliverable on their local machine with produced setup instructions, and the app boots without modification.

### Components (planned)

| # | Component | Notes |
|---|---|---|
| F1–F7 | Curate + upload transcripts for LLD, DB, Frontend, DevOps, Security, QA, Code Reviewer | One per week is realistic |
| F8 | Handoff packaging (DevOps expert produces README/SETUP/env.example) | `backend-go/internal/workflow/handoff.go` |
| F9 | End-to-end test: real requirement → all 6 phases → runnable deliverable | Manual acceptance test |

---

## Global known-follow-ups (not yet fixed, tracked here)

Carried from `HANDOFF.md` Frontend Phase 1 gap list, still relevant:

1. **`GET /me` endpoint missing on backend.** Frontend has a JWT-decode stopgap. Must add real endpoint before Phase B is called complete for the admin UI. Owner: whoever picks up A8.
2. **Ingest multipart field name unverified.** `frontend/src/api/admin.ts` `ingestTranscript` assumes `file` field. Must cross-check `backend-go/internal/admin/admin_handler.go`. Owner: A8.
3. **No ESLint config.** `npm run lint` fails. Add config before Phase E frontend work.
4. **No frontend build has actually been run in a Node environment.** First person with a Node env should `cd frontend && npm install && npm run typecheck && npm run build` and fix whatever surfaces.
5. **Backend Phase 2–6 smoke tests never actually run by this implementer.** A11 requires this.

---

## Cost + time projections

Based on `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §17 estimates for a solo developer:

| Phase | Estimate | Cost driver |
|---|---|---|
| A | 1 week | Migration + schema + admin CRUD |
| B | 2 weeks | Transcript curation is manual |
| C | 2 weeks | Blackboard + workflow engine is the largest new subsystem |
| D | 1.5 weeks | Validation sandboxing needs care |
| E | 1.5 weeks | Frontend Kanban is the visible piece |
| F | 2 weeks | 7 experts, each with curriculum + smoke test |
| **Total** | **~10 weeks** | (Parallelizable if more than one implementer) |

LLM costs for training + smoke testing: rough order of magnitude $50–$200 for 10 experts, dominated by charter extraction (uses `ModelStrong`).

---

## Change log for this file

- 2026-09-07 v0.1.0: initial creation. Phase A started. Design doc + Knowledge Hub + this file committed together.
