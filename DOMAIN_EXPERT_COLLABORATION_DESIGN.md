# AI Avengers — Domain Expert Collaboration & Training Design

> **Purpose:** Single source of truth for how domain experts are trained, how they collaborate on production requirements, and how the client stays in control end-to-end.
>
> **Scope:** This document is the missing piece between `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` (which defines the platform primitives — orchestrator, 5-gate decision engine, China Wall, L1/L2/L3 memory, training pipeline) and actual product usage (client gives requirement → experts collaborate → deliverable ships).
>
> **This document does not redesign anything that already exists.** It defines the collaboration layer, training curation strategy, cross-verification protocol, client approval gates, and the specific set of experts required to build production software. Everything else is deferred to `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`.

---

## Document Metadata

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | DESIGN COMPLETE — READY FOR IMPLEMENTATION |
| Author | System Design Architect (via Duo Chat, based on Byte Byte AI 6-week cohort + Arpit Bhaiya AI Masterclass) |
| Depends on | `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` v1.0.0 |
| Client / Admin | Kiran Nogia |

---

## Table of Contents

1. [Product Reality Check — What Exists Today](#1-product-reality-check)
2. [The 8 Gaps — Definitive Resolution](#2-the-8-gaps--definitive-resolution)
3. [Course Concepts → Product Mapping](#3-course-concepts--product-mapping)
4. [Domain Expert Taxonomy — The 10 Roles](#4-domain-expert-taxonomy)
5. [Per-Expert Training Curriculum](#5-per-expert-training-curriculum)
6. [Training Pipeline — End-to-End Flow](#6-training-pipeline)
7. [Collaboration Blackboard — How Experts Talk](#7-collaboration-blackboard)
8. [Cross-Verification Protocol](#8-cross-verification-protocol)
9. [Workflow State Machine — Requirement to Deliverable](#9-workflow-state-machine)
10. [Client Approval Gates](#10-client-approval-gates)
11. [Output Validation Pipeline](#11-output-validation-pipeline)
12. [Checkpoint & Resume for Multi-Agent Workflows](#12-checkpoint--resume)
13. [Cost Governance per Workflow](#13-cost-governance)
14. [Kanban Board — Client-Facing Transparency](#14-kanban-board)
15. [Schema Extensions Required](#15-schema-extensions-required)
16. [Locked Decisions](#16-locked-decisions)
17. [Implementation Phases](#17-implementation-phases)
18. [Definition of Done — Per Component](#18-definition-of-done)

---

## 1. Product Reality Check

### 1.1 What already exists (verified from `HANDOFF.md` + code inspection)

| Layer | Status | File(s) |
|---|---|---|
| Go backend foundation (config, DB pools, JWT, TOTP, middleware) | IMPLEMENTED | `backend-go/cmd/server/main.go`, `internal/{config,db,auth,middleware}/*` |
| Model gateway (OpenRouter, retry, cache, cost tracking) | IMPLEMENTED | `internal/gateway/model_gateway.go` |
| ML sidecar (bge-base-en-v1.5 embeddings + bge-reranker-base) | IMPLEMENTED | `ml-sidecar/*`, `internal/ml/sidecar_client.go` |
| Training pipeline (recursive chunker, embedding clusterer, topic extractor, charter extractor, capability builder) | IMPLEMENTED | `internal/training/*.go` |
| Memory (L1 Redis, L2 pgvector + hybrid search, L3 append-only log, Manager coordinator) | IMPLEMENTED | `internal/memory/*.go` |
| Context assembler (L2 + rolling summary + recent messages + hybrid search) | IMPLEMENTED | `internal/context/assembler.go` |
| 5-gate decision engine (ASK / WARN / PUSH_BACK / REFUSE / ADVISE) | IMPLEMENTED | `internal/decision/engine.go` |
| China Wall citation enforcer (4 layers) | IMPLEMENTED | `internal/chinawall/enforcer.go` |
| Orchestrator with parallel goroutine expert dispatch + synthesis | IMPLEMENTED (single-turn) | `internal/orchestrator/orchestrator.go` |
| Chat / message / project / rating services | IMPLEMENTED | `internal/{chat,message,project,rating}/*.go` |
| Cost monitor | IMPLEMENTED | `internal/monitoring/cost_monitor.go` |
| Repo integration (GitHub/GitLab OAuth + sync) | IMPLEMENTED | `internal/repo/service.go` |
| React + Vite + TS frontend Phase 1 (auth, routing, stores, API layer, page stubs) | IMPLEMENTED (stubs) | `frontend/src/**` |

### 1.2 What does NOT exist yet (what this document adds)

1. Multi-turn, multi-phase **workflow engine** on top of the orchestrator (orchestrator today is a single-turn synchronous request-response, not a long-running workflow spanning requirement → design → code → handoff).
2. **Blackboard pattern** — a durable, pub/sub-backed event log where experts post artifacts (design decisions, API contracts, code, questions) for other experts to consume asynchronously without blocking each other.
3. **Cross-verification protocol** — deterministic rules for which expert reviews which artifact, how consensus is measured, how disagreements escalate.
4. **Client approval gates** — pause-resume workflow state, notification, response capture.
5. **Output validation pipeline** — deterministic static checks (AST parse, lint, typecheck, formatter) run against every code artifact before it can leave the platform, so that code shipped to the client's local machine never fails to parse.
6. **Workflow-level checkpoint/resume** — the memory system already logs turns, but there is no restart-from-arbitrary-point mechanism for a workflow that spans many turns across many experts.
7. **Per-workflow cost budgeting** — cost monitor tracks totals, but there is no per-workflow budget with soft (warn) and hard (halt) limits.
8. **The 10 concrete domain experts** — expert framework exists (`internal/expert/handler.go`, `internal/training/*`), but none of the actual experts have been created, and no per-expert training curriculum has been defined.
9. **Kanban board** — no client-facing view of what each expert is doing right now.
10. **Client-runnable deliverable packaging** — since there is no CI/CD, code that leaves the platform must be self-contained, syntactically valid, and runnable on the client's local machine with clear setup instructions.

**This document specifies all 10.**

---

## 2. The 8 Gaps — Definitive Resolution

These are the 8 gaps I raised earlier. Each one now has a locked resolution, mapped to concrete components in this document.

| # | Gap | Resolution | Section |
|---|---|---|---|
| 1 | Concurrent multi-agent orchestration | Workflow engine on top of the existing single-turn orchestrator, using Redis Streams for durable event queues and goroutines for concurrent expert execution within a workflow phase | §7, §9 |
| 2 | Shared context / blackboard | Postgres `blackboard_events` table (append-only, per-workflow) + Redis pub/sub channels (`workflow:{id}:events`) for real-time notification | §7, §15 |
| 3 | Domain expert definitions | 10 experts defined in §4, each with charter, training curriculum, tool access, and model tier | §4, §5 |
| 4 | RAG pipeline (per-expert namespacing) | Reuse existing pgvector store; add `expert_id` scoping to all retrieval queries so each expert only retrieves from its own trained chunks + shared project context (L2) | §5, §6 |
| 5 | Cross-verification protocol | Deterministic reviewer matrix (§8) + automatic review-round trigger when an artifact of a given `event_type` is posted | §8 |
| 6 | Checkpoint & resume (workflow level) | `workflow_checkpoints` table with full state snapshot after every phase transition, plus event replay from `blackboard_events` for fine-grained resume | §12, §15 |
| 7 | Output validation & evals | Validation pipeline runs after every code-generating expert produces output: syntax → lint → typecheck → optional unit test generation. Failure re-enters OTA loop | §11 |
| 8 | Cost observability (per-workflow) | Extend existing `cost_monitor.go` with per-workflow rollup, soft/hard budget limits, and client-facing cost card | §13 |

Everything else in this document either builds on or references these resolutions.

---

## 3. Course Concepts → Product Mapping

This is the honest accounting of which course concepts drive which product decisions. Nothing here is name-dropped; every mapping is applied concretely somewhere in this doc.

### 3.1 From Byte Byte AI (6-week cohort)

| Course concept | Product application | Where |
|---|---|---|
| LLM = next-token predictor over vocabulary | Model selection matrix — code-generating experts get lower temperature (deterministic next token), design experts get higher | §4 (per-expert model tier) |
| Sub-word BPE tokenization | Token-count-aware context budgeting (never use character length as proxy) | §6.4, existing `context/assembler.go` |
| Pre-training gives implicit world knowledge; post-training aligns to instructions | We do NOT fine-tune model weights (would require self-hosted OSS model). We simulate SFT via (a) charter as system prompt, (b) in-context few-shot examples from transcripts, (c) RAG grounding | §5 (training pipeline), §6 |
| Sampling: greedy → beam → top-P | top-P + temperature configured per expert type. Code = (T=0.15, top-P=0.1). Design = (T=0.4, top-P=0.5). PM = (T=0.7, top-P=0.9) | §4.3 |
| Decoder-only transformer, context window is finite | Context Budget Manager (already exists) evicts oldest, keeps rolling summary + recent + retrieved. Blackboard events are stored durably but only summarized slices enter LLM context | §6.4, §7 |
| Post-training data quality > quantity | Expert transcripts must be curated, not dumped. Duplicate detection + charter extraction + capability depth grading before an expert is considered "trained" | §5, §6 |

### 3.2 From Arpit Bhaiya AI Masterclass (agentic patterns)

| Course concept | Product application | Where |
|---|---|---|
| Agent = expensive while loop | Every expert is a bounded loop with explicit max iterations and cost cap | §4.4 (per-expert loop config) |
| Observe-Think-Act (OTA) | Used by code-generating experts (Backend, Frontend, DB) — write → validate → fix → repeat, max 5 iterations | §11 |
| Ralph loop (fresh context per iteration, filesystem as state) | Used by long-running expert tasks where accumulating context would blow the window. Applied to "code migration" and "bulk refactor" workflows (future) | §17 (Phase 5) |
| React (thought is precious) | Used by design experts (System Design, LLD, Security) — the reasoning trace itself must live in context because it drives the next tool call | §4.4 |
| Plan-and-Execute | Used at the workflow-engine level (not inside an expert). The workflow engine plans phases, executes each phase (which may fan out to parallel experts), and consults the client at approval gates | §9 |
| File system as context | Existing repo integration (`internal/repo/service.go`) — experts reference the client's connected repo instead of loading it into prompt | Already implemented |
| Checkpoint & resume | Workflow-level checkpoints after every phase (§12). Also fine-grained: any expert task can be resumed from its last completed OTA iteration | §12 |
| Human-in-the-loop as tool call | Client approval gates are exposed to experts as an `AskClient` tool. When an expert emits this tool call, the workflow pauses and waits for client response before resuming | §10 |
| Guardrails / system prompts / no silent failures | Already enforced by (a) charter as system prompt, (b) China Wall citation enforcement, (c) 5-gate decision engine | Existing |
| Cost observability per loop iteration | Every `LLMCall` row already records tokens + cost. Extend with `workflow_id` scoping | §13 |

---

## 4. Domain Expert Taxonomy

### 4.1 The 10 experts

These are the concrete `experts` rows that must exist in the database before any real client workflow can run. Each is a real engineering role, not a marketing label.

| # | Expert | Domain | Primary responsibility |
|---|---|---|---|
| 1 | Product Manager | `product` | Requirement gathering, scope definition, acceptance criteria |
| 2 | System Design Architect | `system_design` | High-level architecture, service boundaries, technology selection, non-functional requirements |
| 3 | Low-Level Design Expert | `lld` | Module structure, class/function design, design patterns, interface contracts |
| 4 | Database Expert | `database` | Schema, indexes, migrations, query design, data modeling |
| 5 | Backend Expert | `backend` | API implementation, business logic, integrations, error handling |
| 6 | Frontend Expert | `frontend` | UI components, state management, routing, accessibility |
| 7 | DevOps Expert | `devops` | Local run instructions, Dockerfiles, env configuration, deployment topology (CI/CD is out of scope by client decision) |
| 8 | Security Expert | `security` | Auth, authorization, secrets, input validation, dependency scanning |
| 9 | QA / Testing Expert | `qa` | Test strategy, unit + integration tests, edge cases, coverage |
| 10 | Code Reviewer | `code_review` | Post-generation review, anti-pattern detection, style consistency, cross-file coherence |

### 4.2 Expert record — schema-level fields

Every expert row in the `experts` table must have:

- `id` (UUID)
- `name` (human-readable)
- `domain` (one of the 10 above; enum)
- `reasoning_charter` (WHY-embedded rules the expert follows; extracted by `charter_extractor.go`)
- `clarification_charter` (JSONB map of topic → questions the expert asks when input is vague; used by Gate 1)
- `model_tier` (`cheap` | `strong` | `fast` — routes through existing `ModelGateway`)
- `temperature`, `top_p` (per §4.3)
- `loop_pattern` (`ota` | `react` | `plan_execute` — see §4.4)
- `max_loop_iterations` (int, hard cap; see §13)
- `allowed_tools` (JSONB array of tool names the expert can call; see §7.3)
- `training_status` (`draft` | `ingesting` | `trained` | `deprecated`)
- `created_by` (admin user id)
- `created_at`, `updated_at`

**Additive to existing schema.** No breaking changes to existing `experts` table (see §15 for exact ALTERs).

### 4.3 Per-expert sampling configuration

Determined by task type. Code-generating experts need determinism (same input → same output → same test result). Design experts need controlled exploration.

| Expert | Model tier | Temperature | Top-P | Reason |
|---|---|---|---|---|
| PM | strong | 0.6 | 0.9 | Needs to ask varied clarifying questions, explore user intent |
| System Design | strong | 0.4 | 0.5 | Reasoning-heavy, some option exploration acceptable |
| LLD | strong | 0.3 | 0.3 | Design patterns are largely deterministic once constraints are known |
| Database | strong | 0.2 | 0.2 | Schema mistakes are expensive; determinism matters |
| Backend | strong | 0.15 | 0.1 | Code generation — determinism critical for validation pipeline |
| Frontend | strong | 0.15 | 0.1 | Same as Backend |
| DevOps | strong | 0.2 | 0.2 | Config files must parse correctly |
| Security | strong | 0.2 | 0.3 | Conservative by default; slight exploration for threat modeling |
| QA | strong | 0.3 | 0.3 | Test cases benefit from some variety in edge-case discovery |
| Code Reviewer | cheap | 0.1 | 0.1 | Deterministic checks; cheap because runs on every artifact |

### 4.4 Per-expert loop pattern

| Expert | Loop pattern | Rationale |
|---|---|---|
| PM | React | Thought is precious — clarifying questions must be reasoned |
| System Design | React | Trade-off exploration requires reasoning in context |
| LLD | React | Same as System Design |
| Database | Plan-and-Execute | Break schema into tables → design each table → link foreign keys |
| Backend | OTA | Write route → run validation pipeline → fix errors → repeat (max 5) |
| Frontend | OTA | Same as Backend |
| DevOps | OTA | Generate config → lint/validate → fix |
| Security | React | Threat modeling is exploratory reasoning |
| QA | Plan-and-Execute | Plan test cases → generate each → validate |
| Code Reviewer | React | Reviews one artifact at a time, thought-driven |

---

## 5. Per-Expert Training Curriculum

This is what the admin (Kiran) must upload for each expert. Training is what makes an expert accurate; without focused transcripts, experts hallucinate.

### 5.1 Training source types

| Source type | Format | Ingestion path |
|---|---|---|
| Course transcript | `.md`, `.txt` | Existing `training/ingestion_pipeline.go` |
| Reference book / documentation excerpt | `.md`, `.pdf` (converted upstream) | Same |
| Client's existing codebase | Git repo URL | Existing `internal/repo/service.go` — chunked and embedded per-expert namespace |
| Past workflow decisions (self-generated) | Auto-fed from L3 event log | New: nightly job appends high-rated decisions to expert's training corpus |

### 5.2 Curriculum per expert (minimum viable)

These are the specific transcripts / sources each expert needs to be considered "trained". Client provides these; admin uploads via `POST /api/v1/admin/experts/{id}/ingest`.

| Expert | Required training material |
|---|---|
| PM | Product management fundamentals, user story writing, requirement elicitation techniques, MVP scoping frameworks |
| System Design | Byte Byte AI 6-week cohort transcripts (already have), Designing Data-Intensive Applications summaries, cloud architecture patterns |
| LLD | Design patterns catalog, SOLID principles, refactoring catalog, domain-driven design tactical patterns |
| Database | Relational modeling, indexing strategies, query optimization, migration patterns, pgvector specifics if project uses it |
| Backend | Chosen framework's official docs summary (e.g., Go stdlib + Gin, or Node + Express), REST API design principles, error handling patterns |
| Frontend | Chosen framework's docs (React + hooks), state management patterns (Zustand / Redux / Context), accessibility WCAG basics, CSS/Tailwind fundamentals |
| DevOps | Docker fundamentals, docker-compose patterns, environment configuration best practices, .env management, secrets handling |
| Security | OWASP Top 10, authentication patterns (JWT, OAuth), authorization models (RBAC, ABAC), secure coding checklists |
| QA | Test pyramid, unit vs integration vs e2e boundaries, mocking strategies, coverage philosophy |
| Code Reviewer | Code review checklists per language, common anti-patterns, style guides (Prettier, gofmt, PEP8) |

### 5.3 What "trained" means (definition)

An expert transitions from `ingesting` → `trained` when ALL of the following are true (enforced by ingestion pipeline):

1. At least 100 chunks embedded and stored (avoids trivially thin experts)
2. Charter successfully extracted with at least 5 WHY-embedded rules
3. Clarification charter has entries for at least 3 topics
4. Capability table has at least 3 topics with depth level ≥ 3
5. Smoke test passes: 5 hand-crafted domain questions get non-REFUSE responses with citations from the expert's own corpus

Smoke test is a new step (see §17 Phase B).

### 5.4 Retraining

Admin can add new transcripts anytime. Ingestion pipeline runs again, chunks are appended (not replaced), capability depths recomputed. Existing L1/L2 memory is preserved (project-scoped) — retraining only affects the source corpus, not project memory.

---

## 6. Training Pipeline

This section describes the end-to-end flow when admin uploads a transcript for an expert.

### 6.1 Flow

```
Admin uploads transcript for Expert X
        │
        ▼
Admin handler receives multipart/form-data
        │
        ▼
training/ingestion_pipeline.go:
  1. training/chunker.go            — split into 500-800 token chunks
  2. ml/sidecar_client.go Embed     — bge-base-en-v1.5, 768D
  3. training/embedding_clusterer   — group semantically similar chunks
  4. training/topic_extractor       — cluster-first topic tagging (60% cost reduction)
  5. training/charter_extractor     — few-shot extraction of WHY-embedded rules
  6. training/capability_builder    — depth 1-5 per topic
        │
        ▼
Persist:
  - course_chunks (chunk text, embedding, topic, expert_id)
  - expert_charter (reasoning + clarification, per expert)
  - expert_capability (topic → depth, can_handle, cannot_handle)
        │
        ▼
Smoke test (§5.3 item 5)
        │
        ▼
Mark expert `trained`
        │
        ▼
Admin UI shows: green badge, capability table, sample questions
```

### 6.2 Per-expert namespacing

Every chunk row has `expert_id`. Retrieval queries filter `WHERE expert_id = $1`. This is the RAG namespace mentioned in Gap #4.

Exception: L2 group memory is per-`project_id`, cross-expert — this is intentional so all experts in a workflow share design context.

### 6.3 Deduplication

Before embedding, `training/chunker.go` should hash each chunk (SHA-256 of normalized text) and skip if a chunk with the same hash already exists for that expert. Prevents accidentally duplicating transcripts.

**Current state:** dedup is not explicit in existing chunker code path per file inspection. Add in §17 Phase A.

### 6.4 Token budgeting

All chunk sizing, retrieval limits, and context assembly use `tiktoken`-equivalent tokenizer counts (not character length). Existing `context/assembler.go` handles this; the training pipeline must match its token accounting so retrieved chunks don't overflow the assembly budget.

---

## 7. Collaboration Blackboard

This is the mechanism that lets experts work concurrently without stepping on each other, and lets them consume each other's outputs asynchronously (Frontend needs Backend's API contract; DB needs System Design's data model; etc.).

### 7.1 Model

- **Durable store:** Postgres table `blackboard_events` (append-only, per workflow).
- **Real-time notification:** Redis pub/sub channel `workflow:{workflow_id}:events`.
- **Consumer model:** Each expert subscribes to event types relevant to its domain (§7.2). When a new event of a subscribed type is posted, the expert is either (a) triggered to act, or (b) triggered to review, per the workflow phase.

### 7.2 Event types

| Event type | Posted by | Typical consumers |
|---|---|---|
| `requirement_captured` | PM | System Design, Security |
| `architecture_decision` | System Design | LLD, DB, Backend, Frontend, DevOps, Security |
| `data_model_proposed` | DB | Backend, System Design |
| `api_contract_proposed` | Backend | Frontend, QA, Security |
| `module_design_proposed` | LLD | Backend, Frontend |
| `code_artifact_produced` | Backend / Frontend / DB / DevOps | Code Reviewer, QA, Security |
| `security_finding` | Security | Backend, Frontend, DevOps, System Design |
| `test_case_proposed` | QA | Backend, Frontend |
| `review_comment` | Code Reviewer | Original producer of the reviewed artifact |
| `question_to_expert` | Any expert | Named expert (via `to_expert_id`) |
| `question_to_client` | Any expert | Client (via approval gate, §10) |
| `client_response` | Client | Expert that raised the question |
| `phase_transition` | Workflow engine | All experts in the workflow |

### 7.3 Blackboard as a tool call

Experts don't write to Postgres directly. They call tools:

- `PostArtifact(event_type, content, references)` — publishes to blackboard.
- `AskExpert(to_expert_id, question)` — publishes `question_to_expert`, waits (with timeout) for a `review_comment` or answer event.
- `AskClient(question, options)` — publishes `question_to_client`, pauses workflow (§10).
- `ReadBlackboard(event_types, since)` — reads relevant events posted by others.

These tools are the ONLY way experts communicate. This constraint is the entire point: it makes the collaboration debuggable, auditable, and resumable.

### 7.4 Ordering and idempotency

- Each event has `id`, `sequence_number` (monotonic per workflow), `posted_at`.
- Each event has `dedup_key` (hash of `event_type + posted_by + content_hash`). Duplicate posts are silently dropped.
- Consumers track their last-processed `sequence_number` in Redis so replays after crash resume from the right spot.

---

## 8. Cross-Verification Protocol

This is Gap #5. The rule: no artifact is considered final until at least one designated reviewer has posted a `review_comment` with status `approved` or `changes_requested`.

### 8.1 Reviewer matrix

| Artifact `event_type` | Mandatory reviewers | Optional reviewers |
|---|---|---|
| `requirement_captured` | Client (approval gate) | — |
| `architecture_decision` | Security, LLD | DB, DevOps |
| `data_model_proposed` | System Design, Backend | Security (if PII involved) |
| `api_contract_proposed` | Frontend, Security | QA |
| `module_design_proposed` | Code Reviewer | Backend or Frontend (whichever will implement) |
| `code_artifact_produced` | Code Reviewer, QA | Security (if touches auth/data) |
| `test_case_proposed` | Backend or Frontend (owner of code under test) | — |

### 8.2 Review outcomes

A `review_comment` event has `status` field, one of:

- `approved` — proceed.
- `changes_requested` — original producer must post a new revision (increment `revision` field).
- `blocked` — hard stop, escalate to client via `question_to_client`.

### 8.3 Consensus

- If all mandatory reviewers say `approved` → artifact is `final`.
- If any mandatory reviewer says `blocked` → workflow pauses at this artifact, client is notified.
- If any mandatory reviewer says `changes_requested` → producer revises, reviewers re-review. Max 3 revision rounds; after that, escalate to client.

### 8.4 Timeout

Reviewers have a soft timeout per artifact (default 60 seconds of wall time; configurable). If a reviewer doesn't respond, workflow engine logs it and proceeds if the reviewer was optional; escalates if mandatory.

---

## 9. Workflow State Machine

A workflow is one end-to-end run from client requirement to final deliverable. It is longer-lived than a chat turn and spans multiple experts across multiple phases.

### 9.1 Phases

```
           ┌───────────────────────────────────────────────────┐
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐                                          │
  │ 1. INTAKE        │  PM elicits requirement                  │
  └────────┬─────────┘  Client approval gate 1                  │
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐  System Design + Security parallel        │
  │ 2. HIGH_LEVEL    │  LLD joins after arch decision            │
  │    DESIGN        │  Client approval gate 2                  │
  └────────┬─────────┘                                          │
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐  DB + Backend + Frontend parallel         │
  │ 3. DETAILED      │  Cross-verify contracts                  │
  │    DESIGN        │  Client approval gate 3                  │
  └────────┬─────────┘                                          │
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐  Backend + Frontend + DB + DevOps parallel│
  │ 4. IMPLEMENTA-   │  Each in OTA loop with validation        │
  │    TION          │  Code Reviewer runs on every artifact    │
  └────────┬─────────┘                                          │
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐  QA generates tests                       │
  │ 5. QA            │  Security final pass                     │
  └────────┬─────────┘                                          │
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐  DevOps produces run instructions        │
  │ 6. HANDOFF       │  Packaging + README + .env.example       │
  │                  │  Client approval gate 4 (final)         │
  └────────┬─────────┘                                          │
           │                                                   │
           ▼                                                   │
  ┌─────────────────┐                                          │
  │  COMPLETED       │  Deliverable delivered                  │
  └─────────────────┘                                          │

  Any phase can loop back to any earlier phase if the client rejects at a gate.
```

### 9.2 State fields

A `workflows` table row has:

- `id`, `client_id`, `project_id`, `title`
- `status`: `draft` | `running` | `paused_for_approval` | `paused_for_client_input` | `completed` | `cancelled` | `failed`
- `current_phase`: enum of the 6 phases above
- `phase_started_at`, `phase_completed_at`
- `selected_expert_ids`: JSONB array (subset of the 10 experts the client chose for this workflow)
- `cost_budget_usd`, `cost_spent_usd` (§13)
- `created_at`, `updated_at`

### 9.3 Transitions

Only the workflow engine writes to `workflows.status` and `workflows.current_phase`. Experts influence transitions by posting `phase_completion_signal` events on the blackboard, which the workflow engine consumes and decides whether to transition.

A phase transitions to the next when:
- All artifacts required by that phase are `final` (§8), AND
- The client approval gate for that phase (if any) is `approved`.

---

## 10. Client Approval Gates

The client (Kiran) approves at every major decision point, as required.

### 10.1 Gate points

1. After `INTAKE` — approve the captured requirement + scope.
2. After `HIGH_LEVEL_DESIGN` — approve architecture + technology choices + security posture.
3. After `DETAILED_DESIGN` — approve data model + API contracts + module design.
4. After `HANDOFF` — final acceptance of deliverable.

Ad-hoc gates: any expert can raise `question_to_client` at any time via the `AskClient` tool (§7.3). This is the fine-grained human-in-the-loop.

### 10.2 Mechanism

When the workflow reaches a gate:

1. Workflow engine sets `workflows.status = paused_for_approval`.
2. Writes a row to `approval_requests` table.
3. Emits `question_to_client` event on blackboard.
4. Frontend polls / SSE-subscribes and shows an approval card.
5. Client responds via `POST /api/v1/workflows/{id}/approvals/{approval_id}`.
6. Workflow engine consumes response, sets status back to `running`, resumes.

### 10.3 Approval payload

Each approval request includes:
- A human-readable summary of what's being approved.
- Full artifact content (design doc, code, whatever).
- Cited sources (China Wall citations).
- Cost so far + estimated cost of continuing.
- Options: `approve`, `approve_with_notes`, `request_changes` (with free-text feedback), `reject_and_restart_phase`, `cancel_workflow`.

### 10.4 Timeouts

By client's preference: **no auto-approval on timeout.** Workflow stays paused indefinitely until client acts. Client dashboard shows all paused workflows prominently.

---

## 11. Output Validation Pipeline

This is Gap #7. Critical because there's no CI/CD — code shipped to client's local must run.

### 11.1 What gets validated

Every `code_artifact_produced` event. The producing expert calls `PostArtifact(event_type='code_artifact_produced', ...)`. The workflow engine intercepts this call and runs the validation pipeline before the event actually lands on the blackboard.

### 11.2 Pipeline stages

Executed in order; fail-fast:

1. **Syntax check** — language-specific AST parse.
   - JS/TS: `@babel/parser` (called via ml-sidecar or a new `internal/validation/parse_sidecar`)
   - Go: `go/parser` (native, in-process)
   - Python: `ast.parse` (via ml-sidecar)
   - SQL: `pg_query_go` (native)
   - JSON / YAML: stdlib unmarshalers

2. **Lint** — language-appropriate linter.
   - JS/TS: ESLint (sidecar)
   - Go: `staticcheck` (sidecar)
   - Python: `ruff` (sidecar)

3. **Format check** — Prettier / gofmt / ruff format. Any diff = fail (auto-fixable, but expert must re-produce to keep provenance clean).

4. **Typecheck** (typed languages only) — `tsc --noEmit`, `go build`.

5. **Test generation (optional)** — QA expert generates unit tests. Tests are run in an isolated subprocess with resource limits. Failure of a generated test means either the test is wrong or the code is wrong — Code Reviewer and QA cross-review to decide.

### 11.3 On failure

The expert enters an OTA loop iteration:
- Observe: validation error output.
- Think: read the error, decide fix.
- Act: emit revised code.
- Repeat, up to `max_loop_iterations` (default 5).

If all iterations exhausted: escalate to Code Reviewer for suggestions, then to client if still stuck.

### 11.4 Sandboxing

All validators run in a subprocess with:
- CPU limit: 5s wall time.
- Memory limit: 512MB.
- Filesystem: read-only tmpfs for artifact, no network.
- Timeout wrapper in Go: `context.WithTimeout` on `exec.CommandContext`.

Generated tests also run in this sandbox — no arbitrary code execution against the platform host.

---

## 12. Checkpoint & Resume

### 12.1 Two levels of checkpointing

**Level 1 — Phase checkpoint (coarse-grained):**

After every phase transition (§9), workflow engine writes a row to `workflow_checkpoints`:
- `workflow_id`
- `phase` (the phase JUST completed)
- `state_snapshot` (JSONB): selected experts, all final artifacts from this phase, cost so far, next phase to enter
- `blackboard_sequence_number` (last consumed event)
- `created_at`

Resume from a phase checkpoint: reload state, re-enter next phase fresh.

**Level 2 — Event checkpoint (fine-grained):**

The blackboard is append-only, so per-expert progress within a phase is recoverable by replaying events from the last consumed `sequence_number`. Each expert tracks its cursor in Redis (`workflow:{id}:expert:{id}:cursor`).

Resume within a phase: reload expert cursors, resume subscription from that sequence number.

### 12.2 When checkpoints are written

- Level 1: after every phase transition (mandatory).
- Level 1: also on any client approval response (opportunistic, cheap).
- Level 2: continuous (every event is durable in Postgres).

### 12.3 Restart scenarios

| Scenario | Recovery |
|---|---|
| Server crash mid-phase | Restart → load latest phase checkpoint → replay blackboard events past that checkpoint → resume experts from their cursors |
| Deployment | Same as crash — new binary loads checkpoints, resumes |
| Client paused workflow for a week | Wake up when client responds → load checkpoint → resume |
| Cost budget hit hard limit | Workflow forcibly paused, checkpoint saved, client notified — client can raise budget and resume |

---

## 13. Cost Governance

### 13.1 Cost boundaries

Existing `LLMCall` records tokens + cost per call. Extend with `workflow_id` column so cost is rollable-up per workflow.

Every workflow has:
- `cost_budget_usd` (set by client at workflow creation).
- `cost_soft_limit_pct` (default 75%): when spent crosses this, admin + client notified.
- `cost_hard_limit_pct` (default 100%): when spent crosses this, workflow auto-pauses (like an approval gate), client must approve budget increase to resume.

### 13.2 Per-expert cost bound

Each expert has `max_loop_iterations`. In addition, each expert task has an implicit cost cap: if a single task exceeds `(cost_budget_usd / len(selected_experts)) * 0.5`, that task is aborted (goroutine cancelled), the expert emits a `budget_exceeded` event, and Code Reviewer/PM triage.

### 13.3 Cost visibility

Client dashboard shows, per workflow:
- Budget vs spent (bar chart).
- Per-expert breakdown (which expert spent how much).
- Per-phase breakdown (which phase was most expensive).
- Projected final cost (based on current phase and historical averages).

---

## 14. Kanban Board

Client-facing view of what every expert is doing right now.

### 14.1 Columns

- **To Do** — task assigned to an expert but not started.
- **In Progress** — expert is actively producing.
- **Under Review** — artifact posted, awaiting reviewers (§8).
- **Blocked** — waiting on client (approval gate or question).
- **Done** — artifact `final`.

### 14.2 Cards

Each card = one artifact task. Fields:
- Title (auto-generated from artifact type)
- Assigned expert
- Reviewers (with status)
- Cost so far
- Time in current column
- Link to full artifact content

### 14.3 Data source

Derived entirely from `blackboard_events` + `workflow_tasks` (see §15). No separate write path; Kanban is a projection.

### 14.4 Realtime updates

Frontend subscribes to `workflow:{id}:events` via SSE (existing SSE infra). Every new event updates the board without polling.

---

## 15. Schema Extensions Required

All additive. No breaking changes. Migration file: `backend-go/migrations/004_collaboration_layer.up.sql`.

### 15.1 New tables

```sql
-- Workflows: long-running end-to-end runs.
CREATE TABLE workflows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id UUID NOT NULL REFERENCES users(id),
  project_id UUID NOT NULL REFERENCES projects(id),
  title TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN
    ('draft','running','paused_for_approval','paused_for_client_input',
     'completed','cancelled','failed')),
  current_phase TEXT NOT NULL CHECK (current_phase IN
    ('intake','high_level_design','detailed_design','implementation','qa','handoff','completed')),
  phase_started_at TIMESTAMPTZ,
  phase_completed_at TIMESTAMPTZ,
  selected_expert_ids JSONB NOT NULL DEFAULT '[]',
  cost_budget_usd NUMERIC(10,4) NOT NULL DEFAULT 10.0,
  cost_spent_usd NUMERIC(10,4) NOT NULL DEFAULT 0.0,
  cost_soft_limit_pct NUMERIC(5,2) NOT NULL DEFAULT 75.0,
  cost_hard_limit_pct NUMERIC(5,2) NOT NULL DEFAULT 100.0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_workflows_client_status ON workflows(client_id, status);
CREATE INDEX idx_workflows_project ON workflows(project_id);

-- Blackboard events: append-only event log per workflow.
CREATE TABLE blackboard_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  sequence_number BIGSERIAL,
  event_type TEXT NOT NULL,
  posted_by_expert_id UUID REFERENCES experts(id),
  posted_by_client BOOLEAN NOT NULL DEFAULT FALSE,
  to_expert_id UUID REFERENCES experts(id),   -- for question_to_expert
  content JSONB NOT NULL,
  references_event_ids UUID[] DEFAULT '{}',   -- what this event responds to
  revision INT NOT NULL DEFAULT 1,
  dedup_key TEXT NOT NULL,
  posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workflow_id, dedup_key)
);
CREATE INDEX idx_bb_workflow_seq ON blackboard_events(workflow_id, sequence_number);
CREATE INDEX idx_bb_workflow_type ON blackboard_events(workflow_id, event_type);
CREATE INDEX idx_bb_to_expert ON blackboard_events(to_expert_id) WHERE to_expert_id IS NOT NULL;

-- Workflow tasks: what each expert is working on (Kanban).
CREATE TABLE workflow_tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  assigned_expert_id UUID NOT NULL REFERENCES experts(id),
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN
    ('todo','in_progress','under_review','blocked','done','cancelled')),
  produced_artifact_event_id UUID REFERENCES blackboard_events(id),
  cost_usd NUMERIC(10,4) NOT NULL DEFAULT 0.0,
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wt_workflow_status ON workflow_tasks(workflow_id, status);
CREATE INDEX idx_wt_expert ON workflow_tasks(assigned_expert_id);

-- Workflow checkpoints: phase-level snapshots for resume.
CREATE TABLE workflow_checkpoints (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  phase TEXT NOT NULL,
  state_snapshot JSONB NOT NULL,
  blackboard_sequence_number BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_wc_workflow ON workflow_checkpoints(workflow_id, created_at DESC);

-- Approval requests: client-facing decision points.
CREATE TABLE approval_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  gate_name TEXT NOT NULL,  -- 'intake','high_level_design','detailed_design','handoff','ad_hoc'
  summary TEXT NOT NULL,
  artifact_content JSONB NOT NULL,
  cited_event_ids UUID[] DEFAULT '{}',
  cost_so_far_usd NUMERIC(10,4) NOT NULL,
  estimated_remaining_usd NUMERIC(10,4),
  status TEXT NOT NULL CHECK (status IN ('pending','approved','changes_requested','rejected','cancelled')),
  client_response JSONB,
  requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  responded_at TIMESTAMPTZ
);
CREATE INDEX idx_ar_workflow_status ON approval_requests(workflow_id, status);
```

### 15.2 ALTERs to existing tables

```sql
-- Expert config fields (from §4.2).
ALTER TABLE experts
  ADD COLUMN IF NOT EXISTS model_tier TEXT NOT NULL DEFAULT 'strong',
  ADD COLUMN IF NOT EXISTS temperature NUMERIC(3,2) NOT NULL DEFAULT 0.3,
  ADD COLUMN IF NOT EXISTS top_p NUMERIC(3,2) NOT NULL DEFAULT 0.5,
  ADD COLUMN IF NOT EXISTS loop_pattern TEXT NOT NULL DEFAULT 'react'
    CHECK (loop_pattern IN ('ota','react','plan_execute')),
  ADD COLUMN IF NOT EXISTS max_loop_iterations INT NOT NULL DEFAULT 5,
  ADD COLUMN IF NOT EXISTS allowed_tools JSONB NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS training_status TEXT NOT NULL DEFAULT 'draft'
    CHECK (training_status IN ('draft','ingesting','trained','deprecated'));

-- Cost tracking per workflow.
-- CORRECTION: an earlier draft of this doc referenced a table named
-- `llm_calls`. That table does not exist. In the actual schema, cost
-- and tokens are tracked directly on the `messages` table
-- (messages.cost_usd, messages.tokens_used, messages.model_used).
-- Per-workflow cost rollup is therefore:
--   SELECT SUM(cost_usd) FROM messages WHERE workflow_id = $1;
ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_messages_workflow ON messages(workflow_id) WHERE workflow_id IS NOT NULL;

-- Per-expert namespacing for chunks. course_chunks already has expert_id
-- (verified against 001_initial_schema.up.sql). Only need to add chunk_hash.
ALTER TABLE course_chunks
  ADD COLUMN IF NOT EXISTS chunk_hash CHAR(64);  -- SHA-256 hex, for dedup (§6.3)
CREATE INDEX IF NOT EXISTS idx_chunks_expert_hash ON course_chunks(expert_id, chunk_hash) WHERE chunk_hash IS NOT NULL;
```

**Actual migration file:** `backend-go/migrations/006_collaboration_layer.up.sql` (up) and `.down.sql` (reverse). Migration number is 006 because 004_messages_warning_fields and 005_seed_admin already exist in main. The up file contains the CHECK constraints, indexes, and system_settings inserts in addition to the DDL shown here.

**Verification step before applying:** admin should `\d experts`, `\d course_chunks`, `\d messages`, and `\d system_settings` in psql to confirm the existing schema before running this migration. The migration uses `ADD COLUMN IF NOT EXISTS` and `CREATE TABLE IF NOT EXISTS` throughout, so re-runs are safe.

---

## 16. Locked Decisions

These cannot be changed without a formal architecture update discussion.

| # | Decision | Reason |
|---|---|---|
| L1 | Blackboard is append-only Postgres + Redis pub/sub | Full audit trail; recovery via replay; no message loss |
| L2 | Experts communicate ONLY through tool calls, never direct function calls | Debuggability, testability, replay |
| L3 | No CI/CD; deliverables are runnable locally by client | Client requirement — explicit |
| L4 | Validation pipeline blocks artifact publication on failure | Zero-tolerance for syntactically invalid code shipping to client |
| L5 | No auto-approval on client timeout | Client's explicit preference |
| L6 | Per-workflow hard cost cap that halts workflow | Cost safety |
| L7 | 10 concrete expert roles, no ad-hoc expert creation without admin curriculum | Prevents untrained experts from being used |
| L8 | Expert charter is generated by training pipeline, admin can edit but cannot bypass | Charter is the guardrail; hand-editing is a footgun |
| L9 | No fine-tuning of model weights; alignment via charter + RAG + in-context examples | We use hosted models via OpenRouter; weight fine-tuning would require self-hosted OSS models |
| L10 | The existing 5-gate decision engine + China Wall are the single source of truth for grounding; no bypasses for any expert | Prevents "trusted expert" backdoors |
| L11 | Workflow phase transitions are engine-controlled, not expert-controlled | Experts request; engine decides |
| L12 | tiktoken-equivalent token counting throughout, never char length | LLM billing and context windows are token-based |

---

## 17. Implementation Phases

Each phase is self-contained and independently valuable. Phases are numbered A–F to avoid collision with the existing `HANDOFF.md` Phase 1–6.

### Phase A — Truth reconciliation + expert schema (1 week)

1. Rewrite `HANDOFF.md` "Overall Status" table by verifying each claimed component against actual code + a runtime smoke test. No component stays marked ✅ without evidence.
2. Apply migration `004_collaboration_layer.up.sql` (§15).
3. Update `internal/expert/handler.go` and admin CRUD to expose new expert fields (model_tier, temperature, top_p, loop_pattern, max_loop_iterations, allowed_tools, training_status).
4. Add chunk deduplication in `training/chunker.go` (§6.3).
5. Add per-expert smoke test in ingestion pipeline (§5.3 item 5).

**Checkpoint:** admin can create an expert with all fields, upload a transcript, and see `training_status=trained` only after smoke test passes.

### Phase B — First 3 experts (2 weeks)

Create PM, System Design, and Backend experts with actual curricula. Do only 3 so the collaboration flow can be exercised end-to-end without the operational overhead of all 10.

1. Curate + upload transcripts per §5.2 for these 3 experts.
2. Manually smoke test each expert (5 questions each).
3. Tune charter and clarification questions if smoke test surfaces gaps.

**Checkpoint:** each of the 3 experts answers 5 domain questions with correct citations, no hallucination.

### Phase C — Blackboard + workflow engine skeleton (2 weeks)

1. Implement `internal/blackboard/store.go` (writes to `blackboard_events`, publishes to Redis).
2. Implement `internal/blackboard/subscriber.go` (Redis subscribe, cursor management).
3. Implement `internal/workflow/engine.go` (state machine per §9, phase transitions).
4. Implement blackboard tools (`PostArtifact`, `AskExpert`, `ReadBlackboard`) — hook into existing expert-processing goroutines in `orchestrator.go`.
5. Wire `AskClient` tool through to `approval_requests` table.

**Checkpoint:** a workflow with 3 experts (from Phase B) can run through INTAKE → HIGH_LEVEL_DESIGN → simulated approval → DETAILED_DESIGN, with all events visible in `blackboard_events`.

### Phase D — Cross-verification + validation pipeline (1.5 weeks)

1. Implement reviewer matrix (§8.1) as config in `internal/workflow/reviewers.go`.
2. Implement `internal/validation/pipeline.go` per §11 stages. Start with Go and TS validators (native + tsc in a subprocess). Python + SQL in a follow-up.
3. Wire validation into `PostArtifact` for `code_artifact_produced` events.
4. Add revision loop up to 3 rounds.

**Checkpoint:** Backend expert produces a Go file with a deliberate syntax error → validation rejects → OTA loop fixes → passes → Code Reviewer approves → artifact goes `final`.

### Phase E — Checkpointing + cost governance + Kanban (1.5 weeks)

1. Implement phase checkpoints (§12) — write on every phase transition and approval response.
2. Implement cursor-based resume (§12).
3. Extend `cost_monitor.go` with per-workflow rollup + soft/hard limit enforcement.
4. Build Kanban API: `GET /api/v1/workflows/{id}/kanban` returns projection.
5. Build Kanban UI on frontend (drag-drop optional; read-only is fine for MVP).

**Checkpoint:** kill the server mid-workflow → restart → workflow resumes from last phase checkpoint → completes successfully. Kanban shows real-time card movement.

### Phase F — Remaining 7 experts + handoff packaging (2 weeks)

1. Curate + upload transcripts for LLD, DB, Frontend, DevOps, Security, QA, Code Reviewer.
2. Smoke test each.
3. Implement handoff packaging: DevOps expert produces `README.md`, `SETUP.md`, `.env.example` for the client deliverable; workflow zips or commits final state.
4. End-to-end test: real requirement ("build me a todo app") → all 6 phases → deliverable runnable on client's local.

**Checkpoint:** client runs deliverable on their local machine with the produced setup instructions, and the app boots without modification.

**Total realistic timeline: 10 weeks for a solo developer, faster with parallelism.**

---

## 18. Definition of Done — Per Component

A component is done only when EVERY box below is checked. This applies uniformly across all phases.

- [ ] Code written and pushed to `main`
- [ ] File exists on `main` (verified by `git ls-tree`)
- [ ] Every function mentally executed for happy path + at least 2 edge cases
- [ ] All anti-pattern checks passed (async/await, silent errors, hardcoded values, DB field verified against schema, circular deps, complete resets, call-site consistency)
- [ ] All callers updated when a signature changed
- [ ] Types consistent end-to-end (Go: `go vet`, `staticcheck`; TS: `tsc --noEmit`)
- [ ] `HANDOFF.md` updated with truthful status (no "complete" without evidence)
- [ ] Locked decisions (§16) not violated
- [ ] Checkpoint condition for the containing phase met and demonstrably passing
- [ ] No gaps left unresolved; if any, they are listed explicitly in `HANDOFF.md` under "Known follow-up"

---

## Appendix A — What this document does NOT cover

Explicitly out of scope, so no one wastes time looking here:

- CI/CD pipeline design — client rejected CI/CD by requirement.
- Multi-tenancy across multiple clients — current design is single-client (Kiran).
- Fine-tuning open-source models — see Locked Decision L9.
- Voice input / voice output — future scope.
- Multi-region deployment — future scope.
- Non-code deliverables (design docs as final output) — implicitly supported (workflow can stop at HIGH_LEVEL_DESIGN if client wants only a design), but not explicitly optimized for.
- Any redesign of the existing 5-gate decision engine, China Wall enforcer, L1/L2/L3 memory, orchestrator single-turn logic, training pipeline, or model gateway. Those are locked in `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` and this document only builds on top of them.

---

## Appendix B — Glossary

| Term | Meaning in this project |
|---|---|
| Expert | An AI agent representing one engineering role; row in `experts` table |
| Charter | WHY-embedded rules extracted from training corpus; the expert's system prompt backbone |
| China Wall | The 4-layer citation enforcement that prevents uncited claims from being returned |
| Blackboard | The `blackboard_events` table + Redis pub/sub — the shared workspace |
| Workflow | One end-to-end run of the 6-phase state machine, tracked in `workflows` table |
| Phase | One of the 6 stages a workflow progresses through |
| Artifact | Any concrete output an expert produces (design doc, code file, test, etc.) |
| Gate | A client approval checkpoint at the end of a phase |
| Validation pipeline | Deterministic checks (syntax → lint → format → typecheck → tests) on code artifacts |
| Checkpoint | Durable snapshot of workflow state for resume |
| Client | The human (Kiran) who requests work and approves deliverables |
| Admin | Same person for now (Kiran) — the one who trains experts |
| Loop pattern | One of OTA / React / Plan-Execute — how the expert internally reasons |

---

**End of document.**
