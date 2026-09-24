# AI Avengers — Target Microservices & Distributed Architecture

| Field | Value |
|-------|-------|
| **Document ID** | `TARGET_MSA_v1` |
| **Status** | **DESIGN LOCKED — implementation not started** |
| **Single source of truth for** | Future service split, distributed topology, contracts, and staged extraction |
| **Depends on (done)** | Stage 0 Scalable Seams — `docs/STAGE0_SCALABLE_SEAMS.md` |
| **Next implementable slice** | Stage 0.1 (still modular monolith; call-site migration onto ports) |
| **Audience** | Any engineer or AI agent with **no prior deep codebase knowledge** |
| **Last updated** | 2026-09-24 |
| **Related (do not replace)** | `docs/CURRENT_ARCHITECTURE.md` (as-built), `docs/STAGE0_SCALABLE_SEAMS.md`, `docs/INTERFACE_FIRST_CONTRACT.md`, `MIGRATIONS.md` |

---

## 0. How to use this document

### For a human engineer

1. Read §1–§3 (why + current → target + non-negotiables).
2. Read §4 (bounded contexts) and §5 (service catalog) for the piece you own.
3. Read §6 (contracts), §7 (data & events), §8 (runtime topology).
4. Implement only the **Stage** you are assigned (§9). Do not jump stages.
5. Use §10 (playbooks) and §11 (checklists) as the acceptance gate.
6. Do **not** invent new services, shared databases, or “platform libs” without updating this doc first.

### For an AI coding agent

```
SYSTEM CONSTRAINTS (must obey):
1. This file is the architecture source of truth for service boundaries.
2. docs/CURRENT_ARCHITECTURE.md is the as-built modular monolith (truth of TODAY).
3. docs/STAGE0_SCALABLE_SEAMS.md is already implemented — do not re-do Stage 0.
4. NEVER split into network services before Stage 2 entry criteria are met.
5. NEVER introduce a shared “platform/common” Go module that multiple services import for business logic (Farley: coupling over DRY).
6. NEVER share a writable database schema across two services.
7. Prefer transactional outbox events over dual-writes and distributed transactions.
8. Keep chat Q&A and build-workflow as SEPARATE domains forever (they already do not import each other).
9. Public HTTP/SSE contracts change only via OpenAPI first (docs/INTERFACE_FIRST_CONTRACT.md).
10. If a task is ambiguous, STOP and ask — do not invent a new service.
```

**When implementing a Stage N item:** open this doc §9.N + the relevant service card in §5 + the contract in §6, then change code only at the seams already defined in `internal/ports`.

---

## 1. Purpose and product invariants

### 1.1 What AI Avengers is

A **constrained multi-expert platform**:

- **Chat Q&A** — clients ask domain experts; answers are citation-bound (China Wall).
- **Build workflows** — multiple experts collaborate (Blackboard + Aider) to produce real code.
- **Expert knowledge** — transcripts → chunks → embeddings → capability charters.
- **Admin-controlled access** — accounts and expert grants provisioned by admins (self-registration may be disabled).

The product moat is **constraint + citation**, not “generic chatbot”.

### 1.2 What this document is NOT

| This doc | Not this doc |
|----------|--------------|
| Target **distributed / microservice** architecture and the path to get there | Replacing `CURRENT_ARCHITECTURE.md` (as-built detail) |
| How to extract services without breaking chat/workflow | A license to rewrite orchestrator/China Wall/decision engine |
| Contracts, events, data ownership, stages | A Kafka/K8s shopping list for day one |

### 1.3 Non-negotiable product invariants (survive every stage)

1. **Chat path and workflow path stay separate** — different UX, different scaling axes, different failure modes. No merging into one “AgentService”.
2. **China Wall semantics** (rerank → coverage → generate-with-citations → strip) stay intact for chat.
3. **Expert grants / entitlement** remain the only gate for domain_expert access; admin/client rules stay explicit.
4. **LLM cost attribution** stays centralized (today: ModelGateway; tomorrow: LLM Gateway service) — Aider never talks to vendors directly.
5. **Soft-delete / audit** of users and L3 event log semantics are preserved.
6. **Zero silent contract drift** — OpenAPI (or event schemas) change before code.

---

## 2. Current state (baseline) → target state (vision)

### 2.1 Baseline (today) — modular monolith + Stage 0 seams

```
┌──────────────────────────────────────────────────────────────┐
│  React SPA                                                   │
└────────────────────────┬─────────────────────────────────────┘
                         │ HTTPS + SSE
┌────────────────────────▼─────────────────────────────────────┐
│  ONE Go process: cmd/server                                  │
│  packages: auth, admin, chat, message, orchestrator,         │
│  decision, chinawall, workflow, blackboard, training,        │
│  gateway, entitlement, knowledge, outbox, ports, …           │
│  SAME Postgres (+ pgvector)  ·  SAME Redis                   │
└────────────┬───────────────────┬──────────────────┬──────────┘
             │                   │                  │
      ml-sidecar (Py)     aider-service (Py)   postgres/redis
```

**Stage 0 already shipped** (`docs/STAGE0_SCALABLE_SEAMS.md`):

| Seam | Location | Role |
|------|----------|------|
| Cross-domain ports | `internal/ports` | `EntitlementCheck`, `KnowledgeReader`, `UsageRecorder`, `EventPublisher` |
| Entitlement adapter | `internal/entitlement` | grants over `user_expert_grants` |
| Knowledge read adapter | `internal/knowledge` | expert catalog reads |
| Transactional outbox | `internal/outbox` + migration `023` | domain events + SKIP LOCKED dispatcher |
| Metrics | `GET /metrics` | counters (LLM, workflow, outbox, entitlement) |

**Explicitly NOT done in Stage 0:** microservice split, Kafka, schema-per-DB, orchestrator rewrite onto `KnowledgeReader`.

### 2.2 Target vision — independently deployable services

```
                         ┌──────────── Edge / API Gateway ────────────┐
                         │  TLS, rate limit, JWT validate (or pass)   │
                         └───────┬─────────────┬─────────────┬────────┘
           ┌─────────────────────┤             │             └──────────────────┐
           ▼                     ▼             ▼                                ▼
   ┌───────────────┐   ┌───────────────┐ ┌───────────────┐            ┌────────────────┐
   │ Identity &    │   │ Entitlement   │ │ Conversation  │            │ Workflow       │
   │ Access (IdP)  │   │ Service       │ │ (Chat Q&A)    │            │ Orchestrator   │
   └───────┬───────┘   └───────┬───────┘ └───────┬───────┘            └────────┬───────┘
           │                   │                 │                               │
           │            ┌──────▼──────┐   ┌──────▼──────┐                 ┌─────▼──────┐
           │            │ Project     │   │ Expert      │                 │ Blackboard │
           │            │ Service     │   │ Knowledge   │                 │ + Runner   │
           │            └─────────────┘   │ + Ingestion │                 └─────┬──────┘
           │                              └──────┬──────┘                       │
           │                                     │                        ┌─────▼──────┐
           │                              ┌──────▼──────┐                 │ Aider      │
           │                              │ LLM Gateway │◄────────────────│ Workers    │
           │                              └──────┬──────┘                 └────────────┘
           │                                     │
           │                              ┌──────▼──────┐
           │                              │ ML Sidecar  │
           │                              │ (embed/rr)  │
           ▼                              └─────────────┘
   Async bus (starts as outbox→dispatcher; later Kafka/NATS if justified)
```

**Definition of “done” for target:** each box above can be **deployed, scaled, and rolled back independently** without a lockstep release of every other box — Beth Scurry / Farley criterion. Until that is true, we still have a (possibly distributed) modular system, not microservices.

### 2.3 The only safe evolution path (locked)

```
Stage 0     ✅ Seams in-process (ports, entitlement, knowledge, outbox, metrics)
Stage 0.1   → Call sites consume ports; drop raw SQL for access/knowledge across domains
Stage 1     → Process split: cmd/api + cmd/worker (+ optional cmd/ingest) SAME binaries/image family, SAME DB
Stage 2     → Extract 1–2 services with OWN DB + HTTP/gRPC + events (start with Entitlement OR LLM Gateway)
Stage 3     → Extract Conversation, Workflow, Knowledge; multi-DB; real broker if outbox fan-out saturates
Stage 4     → Platform hardening: contract tests, golden paths, multi-region options (only if product needs)
```

**Farley rule baked in:** start modular monolith → evolve when **message/ports stop changing often** and **org/scale pressure** appears. We do **not** copy Netflix day-one topology.

---

## 3. Design principles (locked decisions)

| # | Principle | Application here |
|---|-----------|------------------|
| P1 | **Coupling > DRY across services** | Prefer duplicate small DTOs over a shared business library |
| P2 | **Independently deployable unit = service = repo-or-pipeline scope** | One deploy unit may still live in monorepo initially |
| P3 | **Ports & adapters at every boundary** | `internal/ports` today; HTTP/gRPC adapters tomorrow |
| P4 | **Translate at boundaries** | Services keep local models; never leak storage rows across the wire |
| P5 | **No shared writable DB** | Read-replicas or events for cross-context data; not `JOIN` across service tables |
| P6 | **Outbox over dual-write** | Business tx + outbox row; dispatcher publishes; consumers idempotent |
| P7 | **Sync for request/response; async for facts** | Chat completion sync/SSE; `entitlement.granted` async |
| P8 | **Scale axes separated** | HTTP ingress ≠ long workers ≠ GPU embed ≠ Aider CPU |
| P9 | **SKIP LOCKED for competing workers** | Outbox dispatcher, future job pickers (same pattern as DTS class) |
| P10 | **Measure before splitting** | Split only with metrics (`/metrics`, latency phases, queue lag) proving a distinct axis |
| P11 | **Vitamin vs painkiller** | No multi-region, service mesh, or CQRS-everywhere until a painkiller need |
| P12 | **China Wall & decision gates stay inside Conversation domain** | Not a shared “AI core” library other services import |

### 3.1 Explicit anti-patterns (reject in review)

1. **Distributed monolith** — many repos/processes that must be tested and released together.
2. **Shared `platform/` business package** forcing lockstep upgrades.
3. **One mega “Agent service”** owning chat + workflow + training.
4. **Chatty sync chains** (API → A → B → C → D) for one user click without caching/aggregation.
5. **Distributed transactions / 2PC** across service DBs.
6. **Sharing `users` table** writes from Workflow service.
7. **Kafka “because microservices”** while outbox + single consumer still fine.
8. **Rewriting China Wall** during extraction (extract, don’t redesign).
9. **Breaking Aider → LLM proxy cost attribution**.
10. **Putting domain logic in the API gateway**.

---

## 4. Bounded contexts (the firebreaks)

Each context owns **language, rules, and data**. Align services to these — not to technical layers (“utils”, “managers”).

| Context | Ubiquitous language (examples) | Owns data (today’s tables → future) | Does NOT own |
|---------|--------------------------------|--------------------------------------|--------------|
| **Identity & Access** | account, role, TOTP, session, bootstrap | `users`, refresh/password-reset tokens, TOTP secrets | grants as product rule (delegates or emits events) |
| **Entitlement** | grant, capability, plan (future), deny reason | `user_expert_grants` (+ future plans/subscriptions) | expert definitions |
| **Expert Knowledge** | expert, transcript, chunk, charter, domain profile, category | `experts`, `course_chunks`, categories, domain profiles, ingestion jobs | chat messages, workflows |
| **Project** | project, membership, repo link | `projects`, project-experts, repo bindings | chat content |
| **Conversation** | chat, turn, citation, gate, refuse/advise | chats, messages, L1/L2/L3 memory bindings, ratings tied to turns | workflow blackboard |
| **Workflow** | workflow, wave, blackboard event, aider task, kanban | workflows, blackboard events, checkpoints, workspaces metadata | course_chunks source of truth |
| **LLM Platform** | model tier, provider, cost, proxy token | provider settings, cost ledgers, budgets | prompt domain rules |
| **ML Inference** | embedding, rerank score | model weights (sidecar); no business tables | knowledge storage |
| **Usage & Billing** (future) | meter, quota, invoice | usage aggregates | entitlement rules source (reads events) |
| **Admin Experience** | UI aggregation only | none long-term (BFF or fe-for-be) | authoritative writes belong in domain services |

### 4.1 Critical domain split (never collapse)

```
Conversation  ──parallel independent experts, China Wall, decision gates, SSE tokens
Workflow      ──waves, Blackboard collaboration, Aider git workspaces, kanban SSE
```

Code today already documents: *“Multi-agent COLLABORATION happens in the WORKFLOW flow, not here.”* Preserve this in every future topology.

---

## 5. Target service catalog

Each card is enough to staff an engineer. **Do not create services outside this catalog** without a doc revision.

### 5.0 Legend

| Field | Meaning |
|-------|---------|
| **Stage introduced** | Earliest stage this becomes a *network* service (before that it is a package) |
| **Data store** | Sole writable store |
| **Sync API** | Request/response or SSE the UI/other services call |
| **Published events** | Facts others may consume (outbox names are stable) |
| **Consumed events** | Facts this service reacts to |
| **SLOs (initial targets)** | Starting points; refine with measurement |

---

### 5.1 Identity & Access Service (`svc-identity`)

| | |
|--|--|
| **Responsibility** | Register/login/refresh (policy-gated), me, password reset, TOTP, admin-managed accounts CRUD, JWT issuance |
| **Stage introduced** | Stage 2 (optional early extract) or Stage 3 |
| **Data store** | `identity_db`: users, credentials, totp, reset tokens, sessions/refresh |
| **Sync API** | `POST /auth/login\|refresh\|logout`, `GET /auth/me`, admin account routes (or internal only) |
| **Published events** | `identity.account.created`, `identity.account.updated`, `identity.account.deactivated`, `identity.role.changed` |
| **Consumed events** | none required at start |
| **Depends on** | Entitlement only if admin UI needs grant counts (prefer Entitlement owns grants fully) |
| **SLOs** | login p99 < 300ms excl. TOTP; token refresh p99 < 150ms |
| **Notes** | Self-registration flag (`SELF_REGISTRATION_ENABLED`) stays here. Soft-delete + email partial unique (migration 024 pattern) stay here. |

**JWT contract (stable):** access token claims include at least `sub` (account id), `role`, `exp`. Other services **authorize** via Entitlement + role claims; they do **not** re-query identity for every chat token if role is in JWT — except for “is_active” checks on sensitive ops (optional introspection endpoint).

---

### 5.2 Entitlement Service (`svc-entitlement`)

| | |
|--|--|
| **Responsibility** | Answer `Can(account, expert, capability)`; manage grants; future plans/API keys |
| **Stage introduced** | **Stage 2 first extract candidate** (small, stable port already exists) |
| **Data store** | `entitlement_db`: grants, optional plan tables |
| **Sync API** | gRPC/HTTP: `Check`, `FilterAllowed`, `MustAllow`, `GrantedSet`; Admin: set grants |
| **Published events** | `entitlement.granted`, `entitlement.revoked` (already emitted in Stage 0 shape) |
| **Consumed events** | `identity.account.deactivated` → revoke or mark grants inactive |
| **Implements today** | `ports.EntitlementCheck` |
| **SLOs** | check p99 < 20ms (cache hot path); admin update p99 < 200ms |
| **Caching** | per-account grant set in Redis with invalidate on grant events |

**Why extract early:** every chat/workflow/project path needs it; interface already stable; low logic churn; high reuse.

---

### 5.3 Expert Knowledge Service (`svc-knowledge`)

| | |
|--|--|
| **Responsibility** | Expert catalog, domain profiles, categories, transcript ingestion pipeline, chunk storage + vector search APIs, charters |
| **Stage introduced** | Stage 3 |
| **Data store** | `knowledge_db`: experts, course_chunks (pgvector), categories, domain_profiles, ingestion_jobs |
| **Sync API** | `GetExperts`, `ListTrainedActive`, `SearchChunks(expert, queryEmbedding|text, topK)`, admin ingest controls |
| **Published events** | `expert.published`, `expert.updated`, `knowledge.ingestion.completed`, `knowledge.chunkset.updated` |
| **Consumed events** | none critical |
| **Implements today** | `ports.KnowledgeReader` (+ expand with search port before extract) |
| **Workers** | `cmd/ingest` or internal worker: clean → chunk → topic/charter → embed → store (resume-safe) |
| **SLOs** | catalog read p99 < 50ms; chunk search p99 < 200ms excl. embed; ingest throughput measured per transcript GB |

**ML calls:** Knowledge **owns** when to embed; it calls ML Sidecar (or LLM Gateway only if embeddings go through a unified inference plane). Do not put chunk SQL in Conversation after Stage 3.

---

### 5.4 Project Service (`svc-project`)

| | |
|--|--|
| **Responsibility** | Projects CRUD, project↔expert selection, repo OAuth bind + sync metadata |
| **Stage introduced** | Stage 3 |
| **Data store** | `project_db`: projects, project_experts, repo links, sync status |
| **Sync API** | project CRUD; attach experts (calls Entitlement before attach); repo status |
| **Published events** | `project.created`, `project.expert.attached`, `project.repo.synced` |
| **Consumed events** | `entitlement.revoked` → may detach expert from projects for that account |
| **Repo code chunks** | Either stay in Project (`repo_chunks`) **or** move to Knowledge as `source=repo` — **pick one before Stage 3 extract**. Default recommendation: **keep `repo_chunks` in Project** (project-scoped), Knowledge stays course transcripts only. |

---

### 5.5 Conversation Service (`svc-conversation`) — Chat Q&A

| | |
|--|--|
| **Responsibility** | Chats/messages, orchestrator parallel experts, decision gates, China Wall, context assembly, memory L1/L2/L3 for chat, ratings, SSE streams |
| **Stage introduced** | Stage 3 |
| **Data store** | `conversation_db`: chats, messages, memory tables, ratings, chat_index |
| **Sync API** | chat CRUD; `POST .../messages` (SSE); search |
| **Published events** | `conversation.turn.completed`, `conversation.rating.recorded` |
| **Consumed events** | `entitlement.revoked`, `expert.updated` (invalidate caches), `knowledge.chunkset.updated` (optional cache bust) |
| **Calls** | Entitlement (sync), Knowledge (experts + chunk search), LLM Gateway, Project (metadata), ML via Knowledge or direct rerank as designed |
| **SLOs** | time-to-first-SSE byte p99 < 1s; full multi-expert answer depends on model SLA (track separately) |
| **Hard rule** | Does **not** import Workflow packages or tables |

**Internal modules that move together (do not split further without cause):** `orchestrator`, `decision`, `chinawall`, `context`, `memory`, `selflearning`, `rating` (chat-facing).

---

### 5.6 Workflow Service (`svc-workflow`)

| | |
|--|--|
| **Responsibility** | Workflow state machine, planner, DAG/waves, Blackboard, agent_loop, aider_runner orchestration, kanban/files SSE, checkpoints |
| **Stage introduced** | Stage 3 |
| **Data store** | `workflow_db`: workflows, blackboard events, task plans, checkpoints |
| **Object/volume store** | shared workspaces volume or object storage for git worktrees (Aider) |
| **Sync API** | workflow CRUD/start/run/cancel; kanban SSE; files SSE |
| **Published events** | `workflow.started`, `workflow.wave.completed`, `workflow.completed`, `workflow.failed` |
| **Consumed events** | entitlement changes; expert catalog updates |
| **Calls** | Entitlement, Knowledge (training load once per task), LLM Gateway (design agents + proxy for Aider), Aider workers |
| **SLOs** | control APIs p99 < 300ms; run duration is business metric not API SLO |
| **Hard rule** | Does **not** own China Wall chat path |

---

### 5.7 LLM Gateway Service (`svc-llm-gateway`)

| | |
|--|--|
| **Responsibility** | Single door to all model providers; tiers (cheap/strong/fast); cost accounting; budget; **Aider proxy** (`/llm/proxy`) with service token |
| **Stage introduced** | **Stage 2 strong candidate** (clear boundary, cost critical, already centralized) |
| **Data store** | `llm_db` or tables: provider config, cost ledger, budgets |
| **Sync API** | internal `Complete(chat)`; public-to-workers `POST /llm/proxy[/chat/completions]`; admin settings |
| **Published events** | `llm.call.recorded` (for usage), `llm.budget.exceeded` |
| **Auth for Aider** | `AIDER_PROXY_TOKEN` (process identity, not user JWT) — preserve semantics from CURRENT_ARCHITECTURE §10a |
| **SLOs** | overhead p99 < 50ms above provider; 100% cost attribution for proxy calls with `X-Workflow-ID` |

---

### 5.8 ML Inference Sidecar (`svc-ml`)

| | |
|--|--|
| **Responsibility** | Embeddings + rerank only |
| **Stage introduced** | Already separate process today — keep |
| **Data store** | none (stateless) |
| **Sync API** | embed batch; rerank |
| **Contract** | Freeze OpenAPI from FastAPI; generate Go client (INTERFACE_FIRST B3) |
| **Scaling** | CPU/GPU pool independent of API replicas |

---

### 5.9 Aider Worker (`svc-aider-worker`)

| | |
|--|--|
| **Responsibility** | Execute Aider loops on workspaces; no direct vendor LLM |
| **Stage introduced** | Already separate process — keep; may scale as worker pool in Stage 1+ |
| **Data store** | workspace FS only |
| **Calls** | LLM Gateway proxy only |
| **Queue** | Stage 1+: pull jobs via DB SKIP LOCKED or SQS; do not embed job pull inside API request thread longer than enqueue |

---

### 5.10 Usage / Metering Service (`svc-usage`) — future

| | |
|--|--|
| **Responsibility** | Quotas, bills, meters |
| **Stage introduced** | Stage 3–4 when selling experts commercially requires it |
| **Implements** | `ports.UsageRecorder` |
| **Consumed events** | `llm.call.recorded`, `conversation.turn.completed`, `workflow.completed` |

Until then: `NoopUsageRecorder` remains valid.

---

### 5.11 Admin BFF (optional `svc-admin-bff`)

| | |
|--|--|
| **Responsibility** | Aggregate admin UI reads; **no authoritative business writes** that bypass domain services |
| **Stage introduced** | Stage 3 only if admin fan-in becomes painful |
| **Default** | Admin UI talks to domain services via gateway routes until pain appears |

---

## 6. Contracts (the only legal coupling)

### 6.1 Contract types

| Boundary | Source of truth | Tooling | Rule |
|----------|-----------------|---------|------|
| Public HTTP/SSE (FE ↔ system) | `backend-go/api/openapi.yaml` | oapi-codegen + openapi-typescript | Spec first; generate; CI `verify-contract` |
| Service ↔ service sync | Per-service OpenAPI or protobuf | buf / oapi | Versioned; additive |
| Service ↔ service async | Event schema registry (JSON Schema or Avro later) | start as versioned Go structs + JSON in outbox | Additive fields; never reuse names |
| Backend ↔ ML | Sidecar OpenAPI snapshot | generate client | No hand-duplicated DTOs long term |
| Backend ↔ Postgres | Migrations in **owning** service | migrate | No cross-service migration |

See also: `docs/INTERFACE_FIRST_CONTRACT.md` (enforcement plan). This target doc **requires** Interface-First before Stage 2 network splits.

### 6.2 Event envelope (stable)

All outbox/bus events use:

```json
{
  "event_id": "uuid",
  "event_type": "entitlement.granted",
  "event_version": 1,
  "occurred_at": "ISO-8601",
  "aggregate_type": "account",
  "aggregate_id": "uuid",
  "correlation_id": "uuid",
  "causation_id": "uuid",
  "payload": { }
}
```

**Rules:**

- `event_type` is namespaced (`domain.action`).
- Consumers are **idempotent** on `event_id` (inbox table).
- Dispatcher uses **SKIP LOCKED** (already in Stage 0).
- At-least-once delivery assumed; exactly-once is **not** promised.
- Payload carries **business facts**, not table rows (Farley / Evans: translate).

### 6.3 Initial event catalog (extend, don’t rename)

| Event type | Producer | Consumers | Payload (min) |
|------------|----------|-----------|---------------|
| `entitlement.granted` | Entitlement / Identity admin path | Conversation cache, Project, Usage | account_id, expert_ids[], actor_id |
| `entitlement.revoked` | Entitlement | Conversation, Project, Workflow | account_id, expert_ids[] |
| `identity.account.created` | Identity | Admin analytics, Usage | account_id, role |
| `identity.account.deactivated` | Identity | Entitlement, all session caches | account_id |
| `expert.published` | Knowledge | Conversation, Workflow caches | expert_id, slug, version |
| `knowledge.ingestion.completed` | Knowledge | Admin UI, metrics | expert_id, job_id, chunk_count |
| `conversation.turn.completed` | Conversation | Usage, rating pipelines | chat_id, account_id, expert_ids, tokens |
| `workflow.started` / `.completed` / `.failed` | Workflow | Usage, Admin | workflow_id, account_id, cost |
| `llm.call.recorded` | LLM Gateway | Usage, monitoring | account_id?, workflow_id?, cost_usd, model, tokens |

Stage 0 already publishes `entitlement.granted` — keep name stable.

### 6.4 Authn/z between services

| Caller | Callee | Mechanism |
|--------|--------|-----------|
| Browser | Edge/API | User JWT (access) + refresh cookie pattern as today |
| Conversation → Entitlement | mTLS or internal JWT + service mesh later; start with **network policy + service token** in private network |
| Aider → LLM Gateway | `AIDER_PROXY_TOKEN` Bearer |
| Workers → DB | private network only |
| Admin human | JWT role `admin` |

**Authorization rule:** Identity authenticates; **Entitlement authorizes expert use**; role in JWT authorizes admin routes.

### 6.5 Error and idempotency conventions

- Public errors: stable `code` + human `message` (no stack traces).
- Writes that may retry: require `Idempotency-Key` on create payment-like ops; for chat messages, client message id optional later.
- Consumer inbox: `INSERT event_id PRIMARY KEY` ignore duplicates.

---

## 7. Data ownership, consistency, and storage layout

### 7.1 Database-per-service (target)

| Service | Database | Notes |
|---------|----------|-------|
| Identity | `identity_db` | users authoritative |
| Entitlement | `entitlement_db` | grants authoritative |
| Knowledge | `knowledge_db` | pgvector here |
| Project | `project_db` | |
| Conversation | `conversation_db` | memory L2/L3 |
| Workflow | `workflow_db` | blackboard |
| LLM Gateway | `llm_db` | costs |
| Shared until Stage 2 | single `ai_avengers` | **allowed** through Stage 1 |

**Stage 1 rule:** still one physical Postgres is OK if **schemas or table prefixes** match ownership and cross-domain writes go through ports/events only.

### 7.2 Consistency model

| Interaction | Model |
|-------------|-------|
| Single-service write | Strong (local ACID) |
| Cross-service fact propagation | **Eventual** via outbox |
| User-facing grant check | Strong read on Entitlement (sync) — do not rely on eventual cache alone for deny |
| Cost dashboards | Eventual OK |

**Forbidden:** 2PC across `conversation_db` and `knowledge_db`.

### 7.3 Caching policy

- Entitlement grant sets: Redis, TTL + event invalidation.
- Expert summaries: Redis/local LRU, bust on `expert.*` events.
- Embeddings: never cache as source of truth; chunks are.

### 7.4 File / workspace storage

- Aider workspaces: volume today; target object storage (S3) + ephemeral worker disks when multi-node.
- Ingestion uploads: object storage keyed by `expert_id` + `job_id`.

### 7.5 Migration ownership

- Each service owns `migrations/` for its DB.
- Until split: keep monorepo `backend-go/migrations` but **comment header owns context** (already style in 023/024).
- Never add FKs across future service boundaries (e.g. workflow table FK to users is OK *until* extract — replace with soft id reference at extract time).

---

## 8. Runtime topology, scaling, and deployment

### 8.1 Process topology by stage

| Stage | Processes |
|-------|-----------|
| 0 / 0.1 | `api` (all HTTP), embedded outbox dispatcher, ml-sidecar, aider-service, postgres, redis, frontend |
| 1 | `api` (HTTP only, no heavy loops), `worker` (outbox, workflow run, ingest), ml, aider pool, postgres, redis |
| 2 | + `svc-entitlement` and/or `svc-llm-gateway` as separate deployables |
| 3 | + conversation, workflow, knowledge, identity, project as needed |

### 8.2 Scaling axes (what to scale independently)

| Axis | Scale trigger | Unit |
|------|---------------|------|
| Public HTTP/SSE | concurrent users, SSE connections | `api` replicas |
| Workflow execution | active workflows, wave depth | `worker` / workflow runners |
| Aider | implementation tasks | aider workers + workspace IOPS |
| Embed/rerank | ingest + chat QPS | ml replicas / GPU |
| LLM | provider rate limits | gateway + provider keys; not more chat pods alone |
| Outbox dispatch | event lag | dispatcher consumers (SKIP LOCKED safe) |

**Do not** scale only on queue length vanity metrics — use phase timings (pick/start/complete) like the distributed task scheduler pattern.

### 8.3 Workflow execution as distributed task pattern (target)

Map the “store → pick → execute” pattern onto workflows:

1. **Store:** workflow run requested → row `status=scheduled` + plan.
2. **Pick:** worker `SELECT … FOR UPDATE SKIP LOCKED` eligible runs/waves.
3. **Execute:** agent_loop / aider; write checkpoints; emit events.
4. **Observe:** `picked_at`, `started_at`, `completed_at` for SLA diagnostics.

Chat remains request-driven SSE (not this pattern), except optional async post-processing (memory index).

### 8.4 Deployment units (containers)

Minimum production set (target):

- `frontend`, `gateway` (nginx/envoy), `api`, `worker`, `ml`, `aider`, `postgres`, `redis`, `migrate` jobs per DB

Kubernetes optional at Stage 2+; Docker Compose remains valid through Stage 1.

### 8.5 Observability (required before Stage 2)

| Signal | Standard |
|--------|----------|
| Metrics | Prometheus format preferred at Stage 2; JSON `/metrics` acceptable Stage 0–1 |
| Logs | structured JSON; **UTC** timestamps; no secrets |
| Traces | OpenTelemetry trace_id on HTTP + event envelope `correlation_id` |
| Health | liveness cheap; readiness = DB ping + dependency critical path |
| Business | entitlement denies, workflow fail rate, LLM cost/day, outbox lag |

Platform bootstrap (Farley/Spotify idea): one **service template** with metrics, health, auth middleware, outbox client — copy template, don’t create a fat shared business platform.

### 8.6 Resilience

| Concern | Approach |
|---------|----------|
| Provider outage | LLM Gateway circuit breaker + degraded messages |
| Worker crash | checkpoints (workflow already has); retry from last wave |
| Poison message | outbox/DLQ after N attempts |
| Redis loss | entitlement falls back to DB; SSE may drop (reconnect) |
| Partial deploy | additive contracts; feature flags per service |

---

## 9. Staged roadmap (implement only your stage)

### Stage 0 — Scalable seams ✅ DONE

**Ref:** `docs/STAGE0_SCALABLE_SEAMS.md`

- Ports, entitlement, knowledge reader, outbox 023, metrics.

**Exit criteria:** met.

---

### Stage 0.1 — Port adoption (modular monolith hardening)

**Goal:** every cross-domain call goes through ports; still **one** process, **one** DB.

**Work items (ordered):**

1. Inject `ports.EntitlementCheck` into message / project / workflow handlers; remove duplicate SQL access checks (keep `auth.MustHaveExpertAccess` as thin wrapper or delete once unused).
2. Point orchestrator expert load + any admin expert lists used by chat/workflow at `ports.KnowledgeReader` (extend port if `GetExperts` insufficient — add methods **to the port first**).
3. Add `ports.ChunkSearcher` (or extend KnowledgeReader) **interface only** if needed; implement in `internal/knowledge` — call-site migration can be partial but no new raw SQL from orchestrator into `course_chunks`.
4. Ensure grant writes always publish outbox events (already); add `entitlement.revoked` on revoke paths.
5. Document remaining illegal imports in CI grep (e.g. `orchestrator` must not import `training`).

**Exit criteria:**

- [ ] No package outside `entitlement` reads `user_expert_grants`.
- [ ] No package outside `knowledge`/`training` writes `course_chunks`.
- [ ] Chat and workflow still pass existing smoke tests (login, send message, start workflow).
- [ ] `/metrics` shows entitlement denies and outbox published > 0 under test.

**Non-goals:** new services, new broker, OpenAPI full enforcement (can start in parallel).

---

### Stage 1 — Process split (`api` vs `worker`)

**Goal:** separate **scaling axes** without network service boundaries yet.

**Work items:**

1. `cmd/api` — HTTP/SSE only; enqueue workflow runs; no long aider loops in-request.
2. `cmd/worker` — outbox dispatcher, workflow runner, ingestion jobs (flags to enable subsets).
3. Shared packages OK; same Docker image with different entrypoint/command.
4. Job rows: reuse/extend workflow recovery tables; picker uses SKIP LOCKED.
5. Horizontal: N API + M workers against same DB.

**Exit criteria:**

- [ ] Killing workers leaves API healthy (chat still works).
- [ ] Workflow runs complete with API replica count = 1 and workers = 2 under load test.
- [ ] No dual outbox dispatchors corrupting state (claim via SKIP LOCKED).

---

### Stage 2 — First real service extracts

**Goal:** prove independent deployability on the **smallest stable** boundaries.

**Recommended order (pick one first):**

**Option A — Entitlement service (recommended first)**

1. New deployable wrapping current `internal/entitlement` + grants schema copy/move.
2. Conversation/Workflow/API call it via HTTP/gRPC adapter implementing `ports.EntitlementCheck`.
3. Own DB or own Postgres schema with **no** joins to users except by UUID.
4. Contract tests for Check/Filter/Grants API.
5. Deploy entitlement on its own cadence once; older API still works.

**Option B — LLM Gateway service**

1. Move ModelGateway + proxy routes.
2. Aider and Conversation point to gateway URL.
3. Cost ledger ownership moves with it.

**Exit criteria:**

- [ ] Extracted service deployable without rebuilding chat binary (API rebuild only for URL config).
- [ ] Contract test suite green.
- [ ] Chaos: entitlement down → chat returns 503/deny **gracefully**, no monorepo panic.
- [ ] No shared business Go module required at same commit hash.

**Parallel mandatory track:** Interface-First Phase 1–2 from `INTERFACE_FIRST_CONTRACT.md` completed for touched routes.

---

### Stage 3 — Domain service extraction wave

Extract in dependency order:

```
Identity → Entitlement (if not done) → Knowledge → Project → LLM Gateway (if not done) → Conversation → Workflow
```

For each extract:

1. Freeze port + OpenAPI/events.
2. Move tables + migrations to service.
3. Replace in-process adapter with remote adapter behind **same port interface**.
4. Dual-run shadow traffic if risk high (optional).
5. Delete old SQL paths.

**Exit criteria per service:** independent pipeline + own DB + own SLO dashboard + rollback plan.

---

### Stage 4 — Hardening (only with product need)

- Kafka/NATS only if outbox+dispatcher lag or fan-out consumer count justifies.
- Multi-region, read replicas, CQRS read models for admin analytics.
- Usage/Billing service.
- Full contract matrix (PACT/Specmatic-style) across all services.
- Golden technology stack template for new services.

---

## 10. Implementation playbooks (copy-paste for engineers / AI)

### 10.1 Playbook — “Add a field to public API”

1. Edit `openapi.yaml`.
2. Generate FE + BE types.
3. Implement in **owning** service only.
4. Additive only (no renames); version if breaking.
5. Update this doc event/API tables if cross-service.

### 10.2 Playbook — “Check expert access”

```
ALWAYS:
  decision := entitlement.Can(ctx, accountID, role, expertID, capability)
NEVER:
  SELECT … FROM user_expert_grants in chat/workflow packages
```

Capability enum: `chat` | `workflow` | `api` | `project` (extend only in ports + entitlement).

### 10.3 Playbook — “Publish a domain event”

1. In **same DB transaction** as the business write, insert outbox row (Stage 0–1) or outbox in owning service DB (Stage 2+).
2. Dispatcher publishes.
3. Consumer writes inbox + handles idempotently.
4. Add row to §6.3 catalog in the same PR as first producer.

### 10.4 Playbook — “Extract package → service”

1. Confirm Stage entry criteria.
2. Ensure 100% traffic uses port interfaces (no leakage).
3. Define remote API identical in behavior to in-process adapter tests.
4. Ship adapter flip behind config `ENTITLEMENT_MODE=inproc|remote`.
5. Default remote in staging → prod.
6. Remove inproc only after N days + metrics OK.

### 10.5 Playbook — “Change China Wall / decision gates”

1. Work **only** inside Conversation domain packages.
2. Do not “share” helpers into Workflow.
3. Add metrics for refuse/ask/pushback rates.
4. Update `CURRENT_ARCHITECTURE.md` behavior section if user-visible.

### 10.6 Playbook — “Change Aider / implementation phase”

1. Preserve LLM proxy contract (§5.7 / CURRENT_ARCHITECTURE 10a).
2. Workspace path + git isolation rules stay.
3. Cost header `X-Workflow-ID` required.
4. Workers are stateless + checkpointed.

### 10.7 Playbook — “New bounded context?”

Reject by default. If unavoidable: write ADR appendix in this file, name events, name owner team, name data store, update §5 catalog — then implement.

---

## 11. Testing & acceptance strategy

### 11.1 Test layers

| Layer | Where | Purpose |
|-------|-------|---------|
| Unit | domain packages | gates, DAG, entitlement matrix |
| Contract | OpenAPI/event schemas | consumer/provider compatibility |
| Integration | service + its DB | migrations + adapters |
| Smoke E2E | compose stack | login → chat SSE → workflow start |
| Load | staging | prove scale axis before split |

### 11.2 Entitlement matrix (must remain green forever)

| Role | Expert grant | Chat | Workflow |
|------|--------------|------|----------|
| admin | n/a | allow all trained | allow |
| client | n/a (product rule today: all requested) | allow per product rule | allow |
| domain_expert | missing | deny | deny |
| domain_expert | granted | allow | allow |
| any | inactive account | deny | deny |

If product changes client rules, update **this matrix + entitlement service only**.

### 11.3 Extraction acceptance checklist (per service)

- [ ] Own datastore; no cross-DB FK
- [ ] Implements documented sync API + events
- [ ] Remote adapter satisfies existing port tests
- [ ] Independent deploy pipeline
- [ ] Dashboards: latency, error rate, dependency errors
- [ ] Runbook: failure modes + customer impact
- [ ] Rollback: previous version + forward-fix migrations

---

## 12. Security architecture (distributed)

1. **Secrets:** per-service; never bake JWT secret into FE; rotate proxy tokens.
2. **PII:** identity owns email/name; other services store `account_id` only when possible.
3. **Admin bootstrap / TOTP:** remain Identity concerns.
4. **China Wall data:** course chunks may be sensitive IP — Knowledge DB access locked down.
5. **Workspace code:** treat as customer IP; isolate per workflow on disk.
6. **Service auth:** no unauthenticated internal admin routes on public listeners.
7. **Audit:** L3-style append log remains for conversation; admin actions emit events.

---

## 13. Mapping: today’s packages → tomorrow’s services

| Today (`internal/…`) | Tomorrow |
|----------------------|----------|
| `auth`, parts of admin accounts | Identity |
| `entitlement`, grant parts of auth admin | Entitlement |
| `expert`, `knowledge`, `training`, `category`, domain profiles | Knowledge |
| `project`, `repo` | Project |
| `chat`, `message`, `orchestrator`, `decision`, `chinawall`, `context`, `memory`, `selflearning`, `rating` | Conversation |
| `workflow`, `blackboard`, validation (workflow), monitoring workflow-ish | Workflow |
| `gateway` | LLM Gateway |
| `ml` client | stays client; server = ml-sidecar |
| `outbox` | library pattern copied per service (duplicate OK) |
| `ports` | **lives in each consumer** long-term OR tiny **api-only** module of interfaces — never business code |
| `admin` handlers | BFF or gateway routes to domain services |
| `observability` | template copied per service |

---

## 14. Configuration surface (target)

| Variable / config | Owner | Notes |
|-------------------|-------|-------|
| `DATABASE_URL` | each service | separate DSN per Stage 2+ |
| `REDIS_URL` | shared infra OK | key prefix per service |
| `JWT_SECRET` / keys | Identity issues; others validate |
| `SELF_REGISTRATION_ENABLED` | Identity | |
| `AIDER_PROXY_TOKEN` | LLM Gateway + Aider | |
| `ENTITLEMENT_URL` | consumers | Stage 2+ |
| `LLM_GATEWAY_URL` | consumers | Stage 2+ |
| `KNOWLEDGE_URL` | consumers | Stage 3+ |
| `OUTBOX_POLL_MS` | each dispatcher | |

---

## 15. Decision log (architecture ADR summary)

| ID | Decision | Status | Rationale |
|----|----------|--------|-----------|
| D1 | Modular monolith through Stage 1 | Locked | Farley/Shoup: don’t pay microservice tax before scale/org need |
| D2 | Ports in Stage 0 before any split | Done | Extraction = adapter swap |
| D3 | Outbox not Kafka first | Locked | Painkiller when lag/fan-out demands |
| D4 | Chat ≠ Workflow forever | Locked | Different collaboration model & scale axes |
| D5 | Entitlement as first extract candidate | Locked | Stable port, hot path, small surface |
| D6 | LLM Gateway centralized forever | Locked | Cost, provider policy, Aider proxy |
| D7 | No shared business platform module | Locked | Avoid lockstep coupling |
| D8 | OpenAPI-first public contract | Locked | INTERFACE_FIRST |
| D9 | Duplicate DTOs across services preferred to shared models | Locked | Coupling cost |
| D10 | SKIP LOCKED workers | Locked | Proven pattern (DTS / outbox) |
| D11 | Client/expert sell path goes through Entitlement capabilities | Locked | Stage 0 capability enum |
| D12 | Repo chunks default own by Project | Locked pending revisit | Project-scoped code context |

---

## 16. What “good” looks like (executive checklist)

- [ ] New engineer implements a grant change **only** in Entitlement + event consumers — no chat SQL.
- [ ] Chat latency incident does not require restarting workflow workers.
- [ ] Ingest GPU saturation does not take down auth.
- [ ] Aider storm does not bypass cost tracking.
- [ ] You can deploy Entitlement on Tuesday and Conversation on Thursday.
- [ ] Killing the broker/dispatcher spikes lag metrics but does not corrupt ACID local writes.
- [ ] This document + OpenAPI + event catalog are sufficient to implement without reading all of `HANDOFF.md`.

---

## 17. Immediate next actions (after this doc is accepted)

1. **Do not start Stage 2** until Stage 0.1 exit criteria are green.
2. Implement Stage 0.1 port adoption PRs (small, reviewable).
3. Parallel: Interface-First Phase 1 (spec ↔ routes diff + generate).
4. Add CI grep guards for illegal cross-domain SQL.
5. Only then schedule Entitlement extract design spike (Option A).

---

## 18. Appendix A — Capability & role reference

**Roles:** `admin` | `domain_expert` | `client`

**Capabilities:** `chat` | `workflow` | `project` | `api`

**Grant table source of truth:** Entitlement context (`user_expert_grants` today).

---

## 19. Appendix B — Relationship to other docs

| Doc | Role vs this file |
|-----|-------------------|
| `docs/STAGE0_SCALABLE_SEAMS.md` | Completed foundation; child of this roadmap |
| `docs/CURRENT_ARCHITECTURE.md` | As-built monolith detail; update when behavior changes |
| `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` | Historical pre-implementation vision; **not** authority if conflict |
| `docs/INTERFACE_FIRST_CONTRACT.md` | Contract enforcement mechanics |
| `docs/FRONTEND_SYSTEM_DESIGN.md` | FE structure; consumes public OpenAPI only |
| `FUTURE_UPDATES.md` | Tactical backlog; must not contradict this file — if conflict, **this file wins** for distribution topology |
| `MIGRATIONS.md` | Migration numbers; service extracts split migration ownership later |

---

## 20. Appendix C — AI agent task template

```markdown
## Task
Stage: 0.1 | 1 | 2 | 3
Service/package: …
Goal: …

## Read first
- docs/TARGET_MICROSERVICES_ARCHITECTURE.md (§ relevant)
- docs/STAGE0_SCALABLE_SEAMS.md
- docs/CURRENT_ARCHITECTURE.md (§ feature area)
- internal/ports/ports.go

## Constraints
- Do not break chat SSE or workflow run
- Do not add new services in Stage < 2
- Do not share DB writes across domains
- OpenAPI first if public route changes
- Prefer extending ports over concrete imports

## Acceptance
- [ ] Exit criteria from §9 for this stage
- [ ] Tests or smoke steps listed
- [ ] Metrics/logs updated if new path
- [ ] Doc touch if new event/API
```

---

**End of TARGET_MSA_v1**

*Maintainers: any change to service boundaries, event names, or stage exit criteria requires a PR that updates this file in the same change set as code — otherwise the change is incomplete.*
