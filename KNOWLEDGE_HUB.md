# AI Avengers — Knowledge Hub

> **Purpose:** Single, self-contained knowledge transfer document. Any AI model (or human engineer) picking up work on this codebase should be able to read only this file plus `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` and `DOMAIN_EXPERT_COLLABORATION_DESIGN.md`, and be immediately productive without needing to watch any of the source courses.
>
> **Source material distilled here:**
> 1. Byte Byte AI 6-week cohort — LLM foundations (pre-training, tokenization, transformer, sampling, post-training)
> 2. Arpit Bhaiya AI Masterclass — Agentic patterns (OTA, Ralph, React, Plan-Execute, human-in-the-loop, checkpointing)
> 3. Ultimate Go Software Design with Kubernetes 2.0 — Go layering, dependency direction, refactoring philosophy
> 4. System Design Master Class 1 — Distributed task scheduler, message broker on RDBMS, observability & SLA-driven design
> 5. System Design Master Class 2 — Distributed systems approach, load balancer internals, horizontal scaling, health checks
> 6. System Design Master Class 3 — Tiered storage (hot/warm/cold), S3 internals, log-structured storage, merge & compaction
>
> **Total course investment:** ~350k INR. This document is the durable return on that investment.

---

## How to use this document

- Read top-to-bottom once for orientation.
- Then treat it as a reference: every section has explicit **Rules**, **Do this**, **Do NOT do this**, and **Applied in AI Avengers as** callouts.
- Every rule is written in WHEN → DO → BECAUSE → EXCEPT format where applicable (Arpit's charter principle).
- Never fabricate a rule that is not in this document. If you don't know, ask the admin.

---

## Table of Contents

1. [Part I — LLM Foundations (What every AI agent needs to know about itself)](#part-i--llm-foundations)
2. [Part II — Agentic Patterns (How agents actually think and act)](#part-ii--agentic-patterns)
3. [Part III — Go Software Design (How to write and organize Go code in this repo)](#part-iii--go-software-design)
4. [Part IV — Distributed Systems Approach (How to think about scaling and failure)](#part-iv--distributed-systems-approach)
5. [Part V — Storage & Data Layer (When to use what, and why)](#part-v--storage--data-layer)
6. [Part VI — Product-Specific Application (How every principle maps to AI Avengers code)](#part-vi--product-specific-application)
7. [Part VII — Universal Rules (Do / Don't)](#part-vii--universal-rules)
8. [Part VIII — Worked Examples](#part-viii--worked-examples)

---

# Part I — LLM Foundations

## 1.1 What an LLM actually is (no mysticism)

An LLM is a **next-token predictor**. Given a sequence of tokens as input, it outputs a probability distribution over its vocabulary (usually 50K–200K tokens). One token is then sampled from that distribution. That token is appended to the input. The process repeats until a stop condition.

That is literally all it does. Every downstream capability — code generation, reasoning, tool use — emerges from this single mechanism plus training.

**Applied in AI Avengers as:** every prompt we send to `internal/gateway/model_gateway.go` is fundamentally a next-token generation request. Understanding this kills the temptation to treat the LLM as an oracle.

## 1.2 Training happens in two stages

**Pre-training** (very expensive, done once by model provider):
- Train on cleaned internet data (Common Crawl → FineWeb / Dolma / C4).
- Sub-word BPE tokenization (byte-pair encoding). Result: vocabulary of ~50K–200K tokens.
- Model architecture: decoder-only transformer stack (embedding layer → N transformer blocks → linear head producing vocab-sized logits → softmax).
- Objective: minimize cross-entropy loss on next-token prediction.
- Output: a **base model** with implicit world knowledge but no instruction-following.

**Post-training** (cheaper, task-specific):
- **SFT (Supervised Fine-Tuning):** train on curated {prompt, response} demonstration pairs.
- **RL (Reinforcement Learning from Human Feedback):** train the model to prefer responses humans rate higher.
- Output: the model you actually call via API (GPT-4, Claude Sonnet, DeepSeek, etc.).

**Applied in AI Avengers as:** we do NOT fine-tune model weights (we use hosted models via OpenRouter). We simulate post-training via three levers:
1. Charter as system prompt (extracted by `internal/training/charter_extractor.go`).
2. RAG grounding (retrieved chunks injected into context by `internal/context/assembler.go`).
3. In-context few-shot examples in the prompt.

## 1.3 Tokenization details that matter in practice

- **Sub-word BPE is what all modern LLMs use.** Word-level tokenizers make vocab explode; char-level makes sequences too long.
- **Count tokens, never characters.** A 500-character string in English is roughly 100–150 tokens; in code it varies wildly.
- **Same text → different token counts** across models (Claude vs GPT vs Gemini use different tokenizers).
- **Whitespace, punctuation, and case all affect tokenization.** `" love"` and `"love"` are often different tokens.

**Rule (WHEN → DO → BECAUSE):**
- WHEN building context or estimating cost → DO count with `tiktoken` (or the model's actual tokenizer) → BECAUSE character length is a lie and will silently overflow the context window.
- EXCEPT during rough back-of-envelope estimation, where `chars / 4` is a safe rule of thumb for English prose.

**Applied in AI Avengers as:** `internal/context/assembler.go` must count tokens. `internal/training/chunker.go` produces 500–800 token chunks (not char). Cost calculations in `internal/gateway/model_gateway.go` use the model's reported input/output token counts.

## 1.4 Context window is finite and expensive

Every model has a hard context window (e.g., 200K tokens for Claude Sonnet). Every token you put in the prompt:
1. Costs money (input token pricing).
2. Slows the response (bigger prompt = more compute).
3. Increases hallucination risk if the context is noisy or contradictory.

**Rule:**
- WHEN assembling a prompt → DO include only what is needed for this specific turn → BECAUSE more context is not more intelligence, it is more noise.
- EXCEPT for essential grounding (retrieved RAG chunks with citations, active project memory summary, current user message).

**Applied in AI Avengers as:** `internal/context/assembler.go` builds context as: L2 group memory (relevant slice) + rolling summary + recent messages + top-K retrieved course chunks. It does NOT dump the whole codebase or full transcript.

## 1.5 Sampling strategies (top-P, temperature)

Once the model produces a probability distribution over the next token, we must pick one.

- **Greedy** — always pick the highest-probability token. Deterministic but prone to loops ("I'm not sure if I'll ever be able to walk with my dog. I'm not sure if I'll ever be able to walk with my dog...").
- **Beam search** — keep top-K candidate paths, expand each, pick best cumulative. Still deterministic. Rarely used in modern LLM serving.
- **Multinomial sampling** — sample proportional to probability. Adds randomness but can pick very unlikely tokens occasionally.
- **Top-K sampling** — restrict to K highest-probability tokens, then multinomial. Fixed K is limiting.
- **Top-P (nucleus) sampling — the standard.** Restrict to smallest set of tokens whose cumulative probability ≥ P, then multinomial. K adapts per step.
- **Temperature** — pre-sampling scaling of logits. Higher T = flatter distribution = more randomness. Lower T = sharper = more deterministic.

**Rule (per task type):**

| Task type | Temperature | Top-P | Reason |
|---|---|---|---|
| Code generation | 0.1–0.2 | 0.1 | Must compile; determinism helps validation loop converge |
| Design decisions / reasoning | 0.3–0.5 | 0.5 | Some exploration acceptable |
| Requirement gathering / creative | 0.6–0.8 | 0.9 | Diverse clarifying questions |
| Deterministic checks (linting, review) | 0.1 | 0.1 | Consistent verdicts |

**Applied in AI Avengers as:** per-expert `temperature` and `top_p` fields on the `experts` table (per `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §4.3).

## 1.6 What LLMs cannot do (understand this cold)

- **They do not reason. They pattern-match.** Reasoning that appears in output is the pattern-matched shape of reasoning seen in training data.
- **They have no persistent memory between calls.** Everything they "remember" was in the prompt.
- **They do not know what they don't know.** They will confidently produce plausible-sounding fabrication (hallucination).
- **They cannot execute code.** If a response contains code, that code has not been run; assume it may not work.
- **They cannot access the internet, files, or your database** unless you explicitly build tool-use into the workflow.

**Applied in AI Avengers as:** the China Wall (`internal/chinawall/enforcer.go`) exists precisely because we cannot trust unverified LLM output. Every claim must be cited to a source chunk from the expert's own trained corpus.

---

# Part II — Agentic Patterns

## 2.1 What an agent is (Arpit's definition)

> **An agent is an expensive while loop.**

Nothing more. An LLM call inside a loop, with tools, until a completion condition is met. Every "agent framework" (Claude Agent SDK, Anti-Gravity SDK, LangGraph, etc.) is just an implementation of this loop with conveniences.

**Rule:** if you can write the loop yourself in ~50 lines, do that. Frameworks are opaque, slow, and expensive in tokens. Use them only when the abstraction genuinely saves work.

## 2.2 The five patterns (know when to use each)

### 2.2.1 Observe → Think → Act (OTA)

```
loop:
    state = observe_environment()  // run code, read file, check API
    thought = llm.think(state, goal)
    action = execute(thought)
    if done: break
```

**When:** short feedback loops where you can cheaply verify success. Code + run + fix. Web scraping with dynamic pages. Form filling.

**Cost:** grows context each iteration (thought + observation added). Cheap for small loops, expensive for long ones.

**Applied in AI Avengers as:** code-generating experts (Backend, Frontend, DB, DevOps). Loop pattern config: `loop_pattern='ota'`, `max_loop_iterations=5`.

### 2.2.2 Ralph loop

```
while not done:
    fresh_context = read_relevant_files_from_disk()
    llm(prompt + fresh_context) --> makes changes to disk
```

**Key property:** each iteration starts with a **fresh context**. State lives in the filesystem, not in the LLM's window.

**When:** long-running tasks where accumulating context would blow the window. Codebase migrations, bulk refactors, spec-driven implementation.

**Cost:** each iteration re-reads and re-processes; higher per-iteration cost but avoids catastrophic context blow-up.

**Applied in AI Avengers as:** future scope (Phase F+). Not used in initial rollout — our tasks are bounded enough for OTA/React.

### 2.2.3 React (Reason + Act)

```
loop:
    llm outputs: "Thought: I need to X. Action: call_tool(...)"
    execute the tool call
    observation = tool_result
    append observation to context
    llm outputs next Thought + Action
    if final answer produced: break
```

**Key property:** the model's own reasoning trace ("Thought:") is part of the context. Reasoning drives the next tool selection.

**When:** exploratory tasks with no clear plan up-front. Design decisions, debugging, research. "Thought is precious" — when the value of the intermediate reasoning is high, keep it in context.

**When NOT:** trivial tasks that don't need reasoning. Wasteful.

**Applied in AI Avengers as:** PM, System Design, LLD, Security, Code Reviewer experts. Loop pattern: `loop_pattern='react'`.

### 2.2.4 Plan-and-Execute

```
plan = llm.plan(task)  // outputs list of steps
for step in plan:
    result = execute_step(step, prior_context)
    prior_context.append(result)
```

**When:** the task decomposes cleanly into a known sequence. Trip planning day-by-day, itemized cost breakdown, structured report writing, DB schema design table-by-table.

**Danger:** if a step depends on information only discovered mid-execution, and the plan didn't anticipate it, plan becomes stale. Mitigation: allow re-planning after every N steps, or on step failure.

**Applied in AI Avengers as:** DB expert (plan tables → design each), QA expert (plan test cases → generate each). Also used at workflow-engine level (`internal/workflow/engine.go`) across the 6 phases in `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §9.

### 2.2.5 Hybrid / Nested

Real systems combine these. Workflow uses Plan-and-Execute across phases. Within a phase, one expert uses React for design, another uses OTA to write code. Within the OTA loop, a Ralph-style file-system-as-context read may happen.

**Rule:** don't pick one and stick to it dogmatically. Pick per task per expert.

## 2.3 Tool calls are the only way agents talk to the world

Agents interact with everything (filesystem, DB, other agents, the client, external APIs) via tools. Tools are functions the LLM can request. Framework catches the request, executes the function, returns result.

**Rules:**
- Every tool must have a clear, single responsibility.
- Every tool schema must include: purpose, exact input parameters (typed), exact return format, examples of when to use.
- Never expose a raw `bash` or `run_arbitrary_code` tool. Provide narrow tools (`RunPythonFile`, `RunGoTests`) that constrain what the LLM can request.
- Sandbox every tool execution (subprocess with CPU/memory/time limits, no network).

**Applied in AI Avengers as:** experts have `allowed_tools JSONB` field. Blackboard tools defined in `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §7.3 (`PostArtifact`, `AskExpert`, `ReadBlackboard`, `AskClient`).

## 2.4 Human-in-the-loop (HITL)

High-stakes actions must gate on human approval. Implement HITL as a **tool call**: agent calls `AskClient(question, options)`, workflow pauses, client responds, workflow resumes.

**When to gate:**
- Approval gates (spending money, deploying, sending communication).
- Ambiguity resolution (two valid interpretations of requirement).
- Risk checkpoints (destructive operations).
- Exception cases (thresholds breached).

**Rule:** no auto-approval on timeout (client's explicit preference for AI Avengers). Workflow stays paused indefinitely.

## 2.5 Checkpoint & resume

Every long-running agent must be resumable from arbitrary point. Reasons:
- Server crashes mid-run.
- Deployment.
- Client paused, came back a week later.
- Cost budget hit.

**How:**
- After every meaningful step (or every phase transition), snapshot state to durable storage.
- On restart, load latest snapshot and resume.
- For event-driven systems, an append-only event log is itself a checkpoint (replay from last consumed sequence number).

**Applied in AI Avengers as:** two-level checkpointing in `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §12. Phase-level snapshots in `workflow_checkpoints` table + event replay from `blackboard_events`.

## 2.6 Guardrails are not optional

Any production agent must have:
- **System prompt with explicit rules** ("You must call tool X before doing Y", "Never make up values").
- **Input sanitization** — detect prompt injection attempts (embedded instructions in user data).
- **Output validation** — don't ship LLM output without deterministic checks (schema validation for structured output, AST parse for code, citation check for factual claims).
- **Cost caps** — hard limit on loop iterations, tokens per task, dollars per workflow.

See also Part VII (Universal Rules).

---

# Part III — Go Software Design

This section captures what to do when writing Go code in `backend-go/`. Every rule below is either derived from the Ultimate Go with K8s course or already visible in the existing codebase.

## 3.1 Layering: mental model is everything

The project has these horizontal layers, each with a stable meaning. Never violate the direction of imports.

```
 cmd/           (main packages: server, migrate) — wires everything, no logic
   │
   ▼
 internal/api/   (protocol layer: HTTP handlers, SSE, request/response types)
   │          Uses HTTP-specific concerns (gin, http.Request). ONLY layer that imports net/http.
   ▼
 internal/{orchestrator, decision, chinawall, memory, ...}/  (app layer: business use cases)
   │          Protocol-agnostic. Never imports net/http.
   ▼
 internal/{db, ml, gateway, ...}/  (business/infrastructure: DB access, external APIs)
   │          Talks to Postgres, Redis, OpenRouter, ml-sidecar.
   ▼
 Foundation (pkg or vendored libs: web framework, logging, validation)
```

**Rules:**
- WHEN a package needs `net/http` → DO put it in the API layer → BECAUSE HTTP is a protocol concern.
- EXCEPT for genuinely protocol-tied clients (e.g., `auth_client` that calls another HTTP service) which live in app layer with the HTTP import noted.
- WHEN you have a domain concept that spans layers (e.g., "expert") → DO create sub-packages at each layer (`internal/expert/handler.go`, plus the app/business slices as needed) → BECAUSE consistency of mental model outweighs terse code.

**How to verify:** run `grep -r "net/http" internal/` and audit. Only allowed in `internal/api/`, `internal/gateway/`, `internal/ml/` (which is an HTTP client), and `internal/middleware/`.

## 3.2 Data models: parse, don't validate late

Every request that enters the API layer must be parsed and validated at the boundary. Once inside the app layer, types must be trusted.

**Pattern:**
```go
// API-layer type (what wire sends)
type NewExpertRequest struct {
    Name    string `json:"name" validate:"required,min=1,max=100"`
    Domain  string `json:"domain" validate:"required,oneof=product system_design lld ..."`
}

// App-layer type (what business logic uses)
type Expert struct {
    ID     uuid.UUID
    Name   string
    Domain Domain  // typed enum
}

// Conversion function — the ONLY way to cross the layer
func (r NewExpertRequest) ToExpert() (Expert, error) {
    if err := validator.Struct(r); err != nil { return Expert{}, err }
    return Expert{ID: uuid.New(), Name: r.Name, Domain: parseDomain(r.Domain)}, nil
}
```

**Rule:** the app layer never sees raw request structs. Always convert at the boundary.

## 3.3 Errors: handle or bubble, never silently drop

- Every `err` must be either returned, wrapped and returned, logged, or explicitly ignored with a comment explaining why.
- Wrap with `%w`: `fmt.Errorf("load experts failed: %w", err)`.
- Never `_ = someFunc()` without a comment saying why the error is safe to ignore.
- Middleware for HTTP: one central error handler in `internal/middleware/` translates app-layer errors to HTTP status codes.

## 3.4 Concurrency: goroutines + channels, with structure

Existing `internal/orchestrator/orchestrator.go` shows the canonical pattern for parallel experts:

```go
resultCh := make(chan ExpertResponse, len(experts))
var wg sync.WaitGroup
for _, exp := range experts {
    wg.Add(1)
    go func(expert expertRecord) {
        defer wg.Done()
        resultCh <- o.processWithExpert(ctx, req, expert)
    }(exp)
}
go func() { wg.Wait(); close(resultCh) }()

// Collect with context timeout
timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
for {
    select {
    case r, ok := <-resultCh:
        if !ok { goto done }
        results = append(results, r)
    case <-timeoutCtx.Done():
        goto done
    }
}
done:
```

**Rules:**
- Always pass a captured loop variable by value to a goroutine (`func(expert expertRecord)`).
- Always have a timeout via `context.Context`.
- Always `defer wg.Done()`.
- Never spawn unbounded goroutines under user input. Cap the fan-out.
- Use a buffered channel with size equal to expected max senders (avoids goroutine leak on early return).

## 3.5 DB access: pgx pool, prepared statements, explicit columns

- Reuse the pool from `internal/db/postgres.go`. Never create ad-hoc pools.
- Never `SELECT *`. List columns explicitly.
- Use pgvector types via `github.com/pgvector/pgvector-go` for vector columns.
- Wrap multi-statement operations in transactions when they must be atomic (`tx, err := pool.BeginTx(ctx, pgx.TxOptions{})`, `defer tx.Rollback(ctx)`, `tx.Commit(ctx)`).
- For high-concurrency reads with locking (task pickers, message brokers): use `SELECT ... FOR UPDATE SKIP LOCKED` (Postgres 9.5+). See Part IV §4.5.

## 3.6 Config: fail fast, validate at startup

- `internal/config/config.go` loads env vars and validates presence at startup.
- Never silently default a critical value (API keys, DB URLs). Missing = crash with clear error.
- Cache-related config (TTLs, batch sizes) can default, but defaults must be documented at the field.

## 3.7 Testing philosophy

- Unit tests where logic is genuinely testable in isolation (parsing, pure functions, algorithm correctness).
- Integration tests for anything that touches DB, Redis, ML sidecar, or LLM — use `docker-compose` test stack.
- Never mock the LLM in integration tests — use a low-cost model tier (`ModelCheap`) and assert on structural properties of output, not exact text.

---

# Part IV — Distributed Systems Approach

## 4.1 Start with day-zero architecture

One server, one database, one cache. Get it working. Then load-test each component. Scale the actual bottleneck, not the imagined one.

**Rule:** do NOT design for hypothetical scale on day zero. AI Avengers is single-client (Kiran). Design for 1 client + a handful of concurrent workflows. Add sharding, read replicas, multi-region only when metrics prove need.

## 4.2 Anything that can go wrong will go wrong

Every network call: assume it fails. Every DB write: assume it partially succeeds. Every LLM call: assume it times out. Every process: assume it crashes mid-operation.

**How this shapes AI Avengers:**
- All memory writes are async, non-blocking (`memory/manager.go` uses `go func() { ... }()` for L1/L2 updates).
- Every LLM call has retry with exponential backoff (`gateway/model_gateway.go`).
- Every long workflow is checkpointed (§2.5).
- Every reviewer/expert response has a timeout; workflow doesn't hang forever.

## 4.3 True horizontal scalability = every layer

A distributed system is only truly scalable if EVERY layer scales horizontally. One stateful DB with a load-balanced API tier in front is not a distributed system — it's a distributed API layer with a single point of failure.

**For AI Avengers:**
- API layer (Go): stateless, scale horizontally.
- Orchestrator: stateless workflow engine, scale horizontally.
- ML sidecar: stateless, scale horizontally.
- Postgres: single instance for now. Add read replicas when read load justifies. Shard by `client_id` if we ever go multi-tenant.
- Redis: single instance; cluster if hot-path memory grows.
- Blackboard events: partitioned by `workflow_id` for consumer parallelism.

## 4.4 Load balancer internals (what it really does)

A load balancer is just a TCP-level packet forwarder with a strategy for picking backends:

- Read TCP packet from client.
- Pick backend by strategy (round-robin, weighted round-robin, least-connections, consistent-hash).
- Establish TCP connection to backend (or reuse from pool).
- `io.Copy` bytes both directions.
- Track health via periodic health-check to a `/health` endpoint (not just "is the VM up").

**Strategy selection:**

| Strategy | When to use |
|---|---|
| Round-robin | Uniform infra, uniform request latency |
| Weighted round-robin | Heterogeneous infra (mixed instance sizes) |
| Least-connections | High variance in request latency (analytics, long-running LLM calls) |
| Consistent hash | Session affinity, cache warming per backend |

**Applied in AI Avengers as:** we use nginx (see `nginx/nginx.conf`) as reverse proxy. Choice of strategy for the LLM-serving Go tier: **least-connections** because LLM response times vary from 500ms to 30s.

## 4.5 The queue picker pattern (from distributed task scheduler)

Separating **picking** work from **executing** it is one of the most reusable patterns in the industry.

**Anti-pattern:** one machine that both queries the DB for pending work and executes it. Fails because:
- Each machine = 1 DB connection = you can't scale past DB connection limit.
- Long-running tasks block the picker.
- Specialized executors (GPU vs CPU) can't be routed to.

**Right pattern:**
1. **Pickers:** small pool of workers whose only job is `SELECT ... FOR UPDATE SKIP LOCKED` a batch of tasks, mark them picked, push to a queue (SQS/Kafka/Redis Streams).
2. **Executors:** separate pool per task type, consume from queue, execute, update status.

**Key SQL pattern:**
```sql
SELECT id, payload FROM tasks
WHERE scheduled_at <= NOW() + INTERVAL '30 seconds'
  AND picked_at IS NULL
ORDER BY scheduled_at ASC
LIMIT 10
FOR UPDATE SKIP LOCKED;

UPDATE tasks SET picked_at = NOW(), status = 'in_progress'
WHERE id = ANY($1);
```

`SKIP LOCKED` is critical: without it, concurrent pickers block each other. With it, each picker grabs a different set of rows.

**Applied in AI Avengers as:** the workflow engine will use this exact pattern for phase task dispatch. Blackboard event processing by subscriber experts uses Redis Streams consumer groups (Redis-native `SKIP LOCKED` equivalent via `XREADGROUP`).

## 4.6 Observability BEFORE optimization

You cannot fix what you cannot measure. Before optimizing anything, instrument every phase.

**For every request/task, measure:**
- `received_at` — when it entered the system
- `picked_at` — when a worker grabbed it
- `started_at` — when execution began
- `completed_at` or `failed_at` — when it ended

Difference between consecutive timestamps = per-phase latency. This lets you identify the bottleneck.

**Store separately:** `completed_at` and `failed_at`. Reason: failed tasks tend to fail fast, and mixing them into average completion time skews the metric badly.

**Applied in AI Avengers as:** `workflow_tasks` table has `started_at`, `completed_at`. `llm_calls` has `latency_ms`. Extend both with `picked_at` when workflow engine is built.

## 4.7 Reactive updates > polling

When a component needs to know about state changes elsewhere:
- **Poll:** cheap to build, but always stale by up to your poll interval, and wastes queries.
- **Push (pub-sub):** immediate propagation, no waste, but requires infrastructure.

**Rule:** use push (Redis pub-sub, Postgres LISTEN/NOTIFY, or CDC) for anything user-facing or where staleness > 5 seconds is unacceptable.

**Applied in AI Avengers as:** blackboard events publish to `workflow:{id}:events` Redis channel. Frontend SSE subscribes. No polling for chat responses or workflow state.

---

# Part V — Storage & Data Layer

## 5.1 Hot / warm / cold tiered storage

Every piece of data has a lifecycle. Access frequency drops over time. Storage cost varies by tier. Match them.

| Tier | Storage | Access pattern | Example in AI Avengers |
|---|---|---|---|
| Hot | Postgres, Redis | Frequent read+write, low latency | Active workflows, L1 expert memory, active `blackboard_events` |
| Warm | Postgres (separate tables, or read-optimized), pgvector | Frequent read, rare write | Completed workflows (last 30 days), L2 project memory |
| Cold | S3, Glacier | Rare access, compliance/audit | Archived workflow snapshots, historical L3 events > 6 months |

**Rule:** don't bloat hot storage with data no one queries. Move it. Indexes on huge tables slow every query.

**Applied in AI Avengers as:**
- Today: everything in Postgres (single tier). Fine for single-client scale.
- Future (Phase F+): archive `blackboard_events` older than 90 days to S3-compatible storage. Archive `workflow_checkpoints` for completed workflows.

## 5.2 Log-structured storage: why S3 works

S3 (and Kafka, and BitCask, and RocksDB WAL) all use append-only files on cheap magnetic disks. Never random-write to a magnetic disk — it destroys throughput.

**Rules that apply universally:**
- Append-only writes = high throughput on cheap hardware.
- Random reads are OK if you maintain an index (byte offset per key).
- Never update a record in place; write a new version and let compaction reclaim space.
- Compaction: periodic job that reads N old files, drops stale versions, writes a merged file. Reclaims space.

**Applied in AI Avengers as:** `blackboard_events` and `master_event_log` (L3) are effectively log-structured (append-only, sequence-numbered, never updated). We rely on Postgres to store them; the mental model is what matters.

## 5.3 pgvector for RAG

- Use `pgvector` extension for embeddings (already installed per `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`).
- Choose index: `ivfflat` (faster to build, slower queries) or `hnsw` (slower to build, faster queries).
- Distance metric: cosine for text embeddings (bge-base-en-v1.5 outputs are normalized).
- Query pattern:
  ```sql
  SELECT chunk_id, text, embedding <=> $1 AS distance
  FROM course_chunks
  WHERE expert_id = $2
  ORDER BY embedding <=> $1
  LIMIT 20;
  ```
- Follow with reranker (`bge-reranker-base` via ml-sidecar) for top-K refinement.

## 5.4 Hybrid search (semantic + keyword)

Semantic search misses exact-match terms ("function name `parseUserQuery`"). Keyword search misses paraphrases. Combine:
- Semantic: pgvector cosine.
- Keyword: Postgres `tsvector` full-text search or BM25 (external).
- Merge results: rank by weighted score, dedupe.

**Applied in AI Avengers as:** existing `internal/memory/l2_store.go` and `internal/context/assembler.go` already implement hybrid search per `HANDOFF.md` improvement #5 and #8.

## 5.5 CDC (Change Data Capture) — mentioned in courses, future scope here

CDC = tail the database's replication log and stream changes to downstream consumers. Enables ETL to warehouses, cache invalidation, search index sync, without application code doing dual-writes.

**Applied in AI Avengers as:** not used today. Future use case: streaming `llm_calls` to a warehouse for cost analytics.

---

# Part VI — Product-Specific Application

This section maps every principle in Parts I–V to concrete AI Avengers files and decisions. When you write new code, check that it matches.

## 6.1 File → principle map

| File | Principles applied |
|---|---|
| `internal/gateway/model_gateway.go` | §1.5 sampling, §2.6 cost caps, retry with backoff (§4.2) |
| `internal/context/assembler.go` | §1.3 token counting, §1.4 minimum context, §5.4 hybrid search |
| `internal/training/chunker.go` | §1.3 sub-word aware chunking (500–800 tokens) |
| `internal/training/charter_extractor.go` | §1.2 SFT simulation via charter as system prompt |
| `internal/decision/engine.go` | §2.6 guardrails (5 gates), §1.6 anti-hallucination |
| `internal/chinawall/enforcer.go` | §1.6 anti-hallucination (citation required), §2.6 output validation |
| `internal/orchestrator/orchestrator.go` | §3.4 concurrency pattern, §2.1 agent-as-loop |
| `internal/memory/{l1,l2,l3}_store.go` | §5.1 tiered storage (hot/warm/log), §4.7 reactive |
| `internal/memory/manager.go` | §4.2 async non-blocking writes |
| `nginx/nginx.conf` | §4.4 least-connections strategy |

## 6.2 Model tier selection (which model for which task)

We use OpenRouter (`internal/gateway/model_gateway.go`) with three tiers:

- `ModelCheap` — `deepseek/deepseek-chat` — for tagging, coverage checks, metadata, code review (deterministic, high volume).
- `ModelStrong` — `anthropic/claude-3-5-sonnet` — for answer generation, charter extraction, design decisions.
- `ModelFast` — `google/gemini-flash-1.5` — for quick checks, simple classification.

**Rule:** default to `ModelCheap` for anything the customer never sees. Reserve `ModelStrong` for user-facing output and design decisions.

## 6.3 Anti-patterns explicit to this codebase

From `HANDOFF.md` (existing bug list) and course wisdom:

- **Do not** fabricate UUIDs to satisfy FK constraints. If the real ID isn't available yet, use `NULL` and make the column nullable (per Bug 3.2 fix in `memory/manager.go`).
- **Do not** write dead code paths (Gate 1's LLM fallback was never called — fixed per Bug 3.5).
- **Do not** parse SSE frames without cross-read buffering (documented in frontend `useSSEStream.ts`).
- **Do not** allow concurrent 401s to trigger multiple refresh-token calls (single-flight guard in `frontend/src/api/base.ts`).
- **Do not** use `SELECT *` in auth queries (existing fix).
- **Do not** call `gin.Default()` — use `gin.New()` with explicit middleware (existing decision).
- **Do not** load ML models per-request — load once at startup (ml-sidecar decision).
- **Do not** run ML sidecar with multiple gunicorn workers (each worker = another model copy in RAM).

---

# Part VII — Universal Rules

## 7.1 Cross-question BEFORE implementing

Every task, before writing any code, answer:

1. **What exact problem does this solve?** (Not "add a feature" but "admin needs to see per-workflow cost".)
2. **Who triggers it?** (User action / schedule / event / always-on.)
3. **What are the inputs?** (Exact types, sources.)
4. **What are the outputs?** (Exact types, destinations.)
5. **What happens on failure?** (Silent / alert / retry / fallback.)
6. **Can it be disabled?** Is it always-on?
7. **Is it configurable from admin panel?**
8. **Which existing files does it modify?**
9. **Which existing patterns must it follow?**
10. **Does it conflict with any locked decision?**

If any of these are unanswered, ASK before implementing.

## 7.2 Quantitative before qualitative

Numbers decide architecture. Before designing:

- Frequency: how often does this happen?
- Data size: bytes per operation, total in memory?
- Concurrency: how many parallel invocations?
- Latency target: is there one? If yes, what's the budget per stage?

Examples:
- "3 experts in parallel" → 3 goroutines is fine.
- "5000 requests/second" → need connection pooling analysis.
- "1 KB per event, 1M events" → 1 GB total, fits in Postgres, skip S3 archive.

## 7.3 Read → Understand → Verify → Write

Never write code without first reading the surrounding code.

1. **Read:** every file you'll touch, plus every file that imports it or that it imports.
2. **Understand:** what calls this, what does this call, exact types.
3. **Verify:** trace inputs → outputs mentally; check every field name against schema; check every function signature at every call site.
4. **Write:** one file at a time, one component per commit.

## 7.4 Anti-pattern checklist (run on every function)

- [ ] async: does this use `await` / call blocking I/O? If yes, is the caller aware?
- [ ] fire-and-forget: any goroutine without a way to observe its outcome? Add explicit error handling.
- [ ] paths: any hardcoded paths? Use env or `filepath.Join`.
- [ ] magic values: any hardcoded strings/numbers? Extract to constants.
- [ ] DB fields: verified against actual schema, not assumed?
- [ ] external API fields: verified against docs, not assumed?
- [ ] interface completeness: added a field? Updated every constructor, every reset, every serializer?
- [ ] call-site consistency: changed signature? Updated every caller?
- [ ] handler types: registering a callback that can be async? Signature allows `error` return?
- [ ] circular deps: File A imports B and B imports A? Refactor.
- [ ] silent errors: any empty `catch` or `_ = err`? Log or rethrow.
- [ ] reset completeness: added state? Zero-value / reset function updated?
- [ ] assumption check: used a value without reading its source? Read the source.

## 7.5 Definition of Done (never claim "done" without this)

- [ ] Code written and committed
- [ ] File exists on `main` (verified via API or `git ls-tree`)
- [ ] Mentally executed for happy path + at least 2 edge cases
- [ ] Anti-pattern checklist (§7.4) passed
- [ ] All callers updated
- [ ] Types consistent end-to-end
- [ ] `IMPLEMENTATION_HANDOFF.md` updated with truthful status
- [ ] Locked decisions not violated
- [ ] Checkpoint condition for the containing phase demonstrably met

## 7.6 Commit protocol

- Format: `feat(scope): what it does` or `fix(scope): what was broken and how fixed` or `docs(scope): what was documented`.
- One component per commit.
- After commit: verify file exists on branch (list tree or fetch file).
- Never claim success without verification.

---

# Part IX — Deep Dives (New — Added 2026-09-07)

This section captures concepts from the 6 source courses that were not fully represented in Parts I–VIII. Read this alongside the earlier parts.

## 9.1 Post-Training in Full Detail (Byte Byte AI Week 1)

### SFT (Supervised Fine-Tuning)

Goal: convert a base model (next-token predictor) into an instruction-follower.

**How:**
1. Curate demonstration data: `{prompt, response}` pairs in a special-token format.
   ```
   <|prompt|> Give three tips for staying healthy. <|response|> 1. Exercise daily...
   ```
2. Continue training the base model on this data using the same cross-entropy loss.
3. Result: SFT model that answers questions instead of continuing them.

**Applied in AI Avengers as:** we do NOT fine-tune model weights. We simulate SFT via:
- Charter as system prompt (extracted by `internal/training/charter_extractor.go`).
- RAG grounding (retrieved chunks injected by `internal/context/assembler.go`).
- Few-shot examples in the prompt.

### RLHF (Reinforcement Learning from Human Feedback)

Goal: make the model prefer responses humans rate higher.

**How:**
1. Collect human preference data: for the same prompt, show two responses, human picks better one.
2. Train a **reward model** on this preference data.
3. Use RL (PPO algorithm) to update the SFT model to maximize reward model score.
4. Result: model that is more helpful, harmless, honest.

**Applied in AI Avengers as:** the rating system (`internal/rating/handler.go`) collects preference data. `boost_factor` on `course_chunks` is a lightweight analog — chunks cited in high-rated responses get boosted. Full RLHF is future scope.

### Thinking Models (o1/o3 style — Test-Time Compute)

Instead of generating the answer directly, the model generates a long chain-of-thought ("thinking") before the final answer. More compute at inference time = better answers on hard problems.

**Rule:** use thinking models (Claude Sonnet with extended thinking, or o3) for:
- Complex multi-step reasoning (architecture decisions, security analysis).
- Problems where the intermediate reasoning is as valuable as the answer.

**Do NOT use** for: simple lookups, tagging, metadata extraction — wasteful.

**Applied in AI Avengers as:** `ModelStrong` (claude-3-5-sonnet) is used for answer generation and charter extraction. If extended thinking is enabled, it applies here. `ModelCheap` (deepseek) is used for everything else.

## 9.2 Sampling Algorithms — Complete Reference (Byte Byte AI Week 1)

### The full hierarchy

```
Deterministic:
  Greedy search    — always pick highest-probability token. Fast, repetitive.
  Beam search      — keep top-K paths, pick best cumulative. Better than greedy, still repetitive.

Stochastic:
  Multinomial      — sample proportional to probability. Random but can pick very unlikely tokens.
  Top-K sampling   — restrict to K highest tokens, then multinomial. Fixed K is limiting.
  Top-P (nucleus)  — restrict to smallest set with cumulative prob >= P, then multinomial. STANDARD.
```

### Why greedy fails in production

Greedy picks the highest-probability token at every step. Because some sequences are very common on the internet ("I'm not sure if I'll ever be able to..."), greedy loops on them. Never use greedy for user-facing text generation.

### Why Top-P is the standard

Top-P dynamically adjusts K based on the model's confidence:
- When model is very confident (one token has 89% probability): K=1, only that token considered.
- When model is uncertain (many tokens each ~10%): K=many, more exploration.

This is better than fixed Top-K because it adapts to the distribution shape.

### Temperature — what it actually does

Temperature scales the raw logits BEFORE softmax:
- `T < 1.0` → sharper distribution → more deterministic → model picks its top choice more often.
- `T > 1.0` → flatter distribution → more random → model explores more.
- `T = 0` → equivalent to greedy.

**Rule table (from course + applied in AI Avengers):**

| Task | Temperature | Top-P | Why |
|---|---|---|---|
| Code generation | 0.1–0.2 | 0.1 | Must compile; determinism helps OTA loop converge |
| Architecture decisions | 0.3–0.5 | 0.5 | Some exploration acceptable |
| Charter extraction | 0.3 | 0.5 | Need consistent rules, not creative |
| Clarifying questions | 0.6–0.8 | 0.9 | Diverse questions are better |
| Coverage check / linting | 0.1 | 0.1 | Consistent verdicts |

## 9.3 Agentic Patterns — Deep Dive (Arpit Masterclass)

### The core insight: agents are just while loops

> "AI agent is just an expensive while loop." — Arpit Bhiyani

Every agent framework (Claude Agent SDK, Anti-Gravity SDK, LangGraph) is an implementation of this loop with conveniences. If you can write the loop in ~50 lines, do that. Frameworks are opaque, slow, and expensive in tokens.

### OTA vs Ralph — when to use which

**OTA (Observe → Think → Act):**
- Context grows each iteration (observation + thought appended).
- Best when: you can cheaply run the code and observe the result. Short feedback loops.
- Example: fix a Python syntax error. Run → see error → fix → run again.
- Context fills up over many iterations → expensive for long tasks.

**Ralph loop:**
- Each iteration starts with a FRESH context. State lives on disk/filesystem, not in LLM window.
- Best when: long-running tasks where context would blow up. Codebase migrations, bulk refactors.
- Key property: `while not done: fresh_context = read_from_disk(); llm(prompt + fresh_context)`.
- WHY it works: file system IS the context. Changes persist to disk. Next iteration reads fresh.
- Downside: each iteration re-reads and re-processes. Higher per-iteration cost.

**Rule:** OTA for short loops where you can observe. Ralph for long-running tasks where context would overflow. You can nest: Ralph at top level, OTA inside each Ralph iteration.

### React — when thought is precious

React makes the model's reasoning trace a first-class citizen in the context:
```
loop:
    llm outputs: "Thought: I need to search for X. Action: search(X)"
    execute search(X)
    observation = result
    append observation to context
    llm outputs next Thought + Action
    if final answer: break
```

**WHY it works:** the thought is in the context. The next token generation is conditioned on the reasoning. This dramatically improves accuracy on multi-step problems.

**When thought is precious:** exploratory tasks, open-ended problems, multi-step reasoning where intermediate steps depend on each other.

**When thought is NOT precious:** simple lookups, deterministic tasks. Wasteful.

### Plan-and-Execute — when to decompose

```
plan = llm.plan(task)  // outputs list of steps
for step in plan:
    result = execute_step(step, prior_context)
    prior_context.append(result)
```

**When to use:** task decomposes cleanly into a known sequence. Trip planning, itemized cost breakdown, structured report writing, DB schema design table-by-table.

**Key danger:** if a step depends on information only discovered mid-execution, and the plan didn't anticipate it, plan becomes stale. Mitigation: allow re-planning after every N steps, or on step failure.

**Parallel execution:** if steps are independent, spin up sub-agents in parallel. This is where Plan-and-Execute beats React for structured tasks.

**Rule:** always smaller tasks. Larger context = more hallucination. Break into chunks, verify each, move forward.

### Human-in-the-Loop — implementation patterns

Three ways to implement HITL:

1. **Top-level confirmation:** before every high-stakes action, ask. Like Claude asking before running bash commands.
2. **Confidence-based:** if model confidence < threshold, pause and ask. Good for support tickets, medical triage.
3. **Exception-based:** run autonomously, pause only when something unexpected happens (tool call fails, cost threshold hit, ambiguous requirement).

**Implementation:** HITL is a tool call. Agent calls `AskClient(question, options)`. Workflow pauses. Client responds. Workflow resumes.

**Rule for AI Avengers:** no auto-approval on timeout. Workflow stays paused indefinitely. Client's explicit response required.

### Checkpoint & Resume — production requirement

Every long-running agent MUST be resumable. Reasons:
- Server crashes mid-run.
- Deployment (rolling restart).
- Client paused, came back a week later.
- Cost budget hit.
- Idempotency: don't make the same phone call twice.

**Implementation (from Arpit's example):**
```python
# After every successful step:
save_checkpoint({
    "history": full_context,  # all messages so far
    "metadata": {"model_config": ..., "task_definition": ..., "file_timestamps": ...},
    "step_index": current_step
})

# On restart:
if checkpoint_exists():
    context = load_checkpoint()["history"]
    resume_from_step = load_checkpoint()["step_index"]
```

**What to store:** full context (all messages), metadata (model config, task definition, file timestamps), step index.

**Optimization:** you can skip tool call responses that are already consumed by later steps. But be careful — if you miss it, you need another round-trip to re-fetch. Prompt caching makes keeping everything cheaper than selective pruning.

**Applied in AI Avengers as:** two-level checkpointing. Phase-level snapshots in `workflow_checkpoints` table. Event replay from `blackboard_events` for fine-grained resume.

### File System as Context (Ralph loop detail)

When an agent reads a file, it should track metadata:
- File path
- Last modified timestamp

On next iteration: if file is already in context AND timestamp hasn't changed → don't reload. If file was modified → reload.

This prevents bloating context with redundant re-reads of unchanged files.

**Applied in AI Avengers as:** future scope for the code-generating experts (Backend, Frontend, DB). When they read repo files, they should track this metadata.

## 9.4 Distributed Task Scheduler — Design Pattern (System Design Master Class 1)

This is the canonical pattern for any system that needs to execute tasks at a scheduled time with a strict SLA.

### The three phases: Store → Pick → Execute

**Store:**
```sql
CREATE TABLE tasks (
    id          UUID PRIMARY KEY,
    command     JSONB NOT NULL,      -- what to execute
    scheduled_at BIGINT NOT NULL,    -- Unix epoch, minute-level granularity
    status      VARCHAR(20) DEFAULT 'pending',  -- pending/in_progress/completed/failed
    picked_at   TIMESTAMPTZ,
    started_at  TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);
```

**Pick (the critical part):**
```sql
-- Pickers run every 30 seconds, grab tasks due in next 30s
SELECT id, command FROM tasks
WHERE scheduled_at <= EXTRACT(EPOCH FROM NOW()) + 30
  AND status = 'pending'
ORDER BY scheduled_at ASC
LIMIT 100
FOR UPDATE SKIP LOCKED;  -- CRITICAL: concurrent pickers don't block each other

UPDATE tasks SET status='in_progress', picked_at=NOW()
WHERE id = ANY($1);
```

**Execute:** separate worker pool consumes from queue, executes, updates status.

### Key insight: recurring task = one-time task scheduled multiple times

Solve one-time execution really well first. Then recurring is just: after a task completes, schedule the next occurrence. Don't try to solve both at once.

### SLA design principle

You cannot guarantee completion time (task might run for hours). You CAN guarantee start time. Design your SLA around what you can control.

**Applied in AI Avengers as:** the workflow engine uses this exact pattern for phase task dispatch. `workflow_tasks` table has `picked_at`, `started_at`, `completed_at`. Pickers use `SELECT FOR UPDATE SKIP LOCKED`.

## 9.5 Message Broker on RDBMS (System Design Master Class 1)

Building a message broker on a relational DB teaches you what properties you need from ANY storage layer.

**Core table:**
```sql
CREATE TABLE messages (
    id          BIGSERIAL PRIMARY KEY,  -- sequential for ordering
    topic       VARCHAR(255) NOT NULL,
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
    -- NO deleted_at: messages are consumed, not deleted
);

CREATE TABLE consumer_offsets (
    consumer_group  VARCHAR(255),
    topic           VARCHAR(255),
    last_offset     BIGINT DEFAULT 0,
    PRIMARY KEY (consumer_group, topic)
);
```

**Consumer reads:**
```sql
SELECT id, payload FROM messages
WHERE topic = $1 AND id > $2  -- $2 = last_offset for this consumer group
ORDER BY id ASC
LIMIT 100;
```

**Key properties needed from storage:**
- Sequential ordering (BIGSERIAL gives this).
- Efficient range scan by offset (B-tree index on id).
- Concurrent consumers don't interfere (each has its own offset).
- Append-only (never update a message).

**Applied in AI Avengers as:** `blackboard_events` table is effectively a message broker. `sequence_number BIGSERIAL` gives ordering. Experts consume from their last-seen sequence number. Redis pub-sub provides real-time notification; Postgres provides durability.

## 9.6 Go Layering — Mental Model (Ultimate Go Course)

The Ultimate Go course's core teaching: **mental model is everything**. If your mental model of the codebase is wrong, every change you make will be wrong.

### The refactoring philosophy

1. **Compiler-driven refactoring:** make a change, let things go red, follow the red. The compiler tells you every place that needs updating. Never try to find all callers manually.
2. **Move things to where they belong:** auth is an app-layer concern, not a business-layer concern. By the time you're in the business layer, auth should already be done.
3. **Precision in naming:** if you have `authenticateService` and `authenticateLocal`, you've lost the mental model. Rename to `authenticate` (client call) and `bearer`/`basic` (protocol-specific). Now it's obvious.
4. **Function type over interface for callbacks:** don't create an interface for a single-method callback. Use a function type. It's simpler and more composable.

### Layer violations to watch for

- `net/http` imported in business layer → violation. HTTP is protocol, belongs in API layer.
- Auth logic in business layer → violation. Auth is app-layer concern.
- DB test depending on auth package → violation. DB tests should not care about auth.

**Applied in AI Avengers as:** `internal/api/` is the only layer that imports `net/http`. `internal/middleware/` handles auth. Business logic in `internal/{orchestrator,decision,chinawall,memory,context}/` is protocol-agnostic.

## 9.7 Storage Internals — S3 and LSM Trees (System Design Master Class 3)

### Why S3 is fast (log-structured storage)

S3 stores data in append-only files on cheap magnetic disks. The key insight:
- **Magnetic disk sequential write:** ~200 MB/s.
- **Magnetic disk random write:** ~1 MB/s.
- **Conclusion:** never random-write to magnetic disk. Always append.

S3 maintains an index (byte offset per key) for fast reads. Writes are always appends. Compaction periodically merges old files and drops stale versions.

### LSM Trees (Log-Structured Merge Trees)

Used by RocksDB, Cassandra, LevelDB. The write path:
1. Write to in-memory buffer (MemTable).
2. When MemTable is full, flush to disk as an immutable SSTable (Sorted String Table).
3. SSTables are sorted by key, so range scans are fast.
4. Compaction: periodically merge SSTables, drop deleted/stale versions.

**Read path:** check MemTable → check Bloom filter (is key in this SSTable?) → binary search SSTable.

**Bloom filter:** probabilistic data structure. "Is this key definitely NOT in this SSTable?" If yes, skip the SSTable. Reduces disk reads dramatically.

**Applied in AI Avengers as:** we use Postgres (B-tree indexes, not LSM). But the mental model applies to `blackboard_events` and `master_event_log` — both are append-only, never updated. Future: if we move to Cassandra or RocksDB for event storage, LSM tree properties apply directly.

## 9.8 Anti-Patterns Discovered in This Session (2026-09-07)

These are NEW anti-patterns not in the original §6.3 or §7.4. Add them to your checklist.

### Anti-pattern: Signature change without caller update

**What happened:** `IngestTranscript` signature was extended with `replaceExisting bool` (7th param). The commit was made. The caller in `admin_handler.go` was not updated in the same commit. Build broke on `main`.

**Rule:** WHEN changing a function signature → DO grep for all callers BEFORE committing → BECAUSE Go will not compile with mismatched call sites. The compiler is your friend — use it before committing, not after.

**Correct process:**
1. Change the function signature.
2. `grep -r "FunctionName(" internal/` to find all callers.
3. Update every caller in the SAME commit.
4. Only then commit.

### Anti-pattern: Trusting design doc SQL without cross-checking schema

**What happened:** `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §15 referenced `ALTER TABLE llm_calls ADD COLUMN workflow_id`. The `llm_calls` table does not exist in the schema. Cost tracking is on `messages.cost_usd`.

**Rule:** WHEN writing a migration → DO read every existing migration file first → BECAUSE design docs are written before implementation and may reference tables that were never created or were renamed.

### Anti-pattern: Trusting handoff file top-level status over per-component tables

**What happened:** `HANDOFF.md` top-level table said "Phase 2: ✅ COMPLETE" but per-component table said "⏳ PENDING" for the same phase.

**Rule:** WHEN reading handoff files → DO trust the per-component tables over the top-level summary → BECAUSE top-level summaries are often updated optimistically while per-component tables reflect actual state.

---

# Part VIII — Worked Examples

## 8.1 Example: "Add a new tool for experts to call"

1. **Cross-question:** what is the tool for? What inputs/outputs? Which experts can call it? What's the failure mode?
2. **Read:** existing tool definitions (blackboard tools in the workflow engine when built), how tools are registered in the LLM call path.
3. **Design:** tool schema (name, description, parameters, return type). Which experts have it in `allowed_tools`?
4. **Implement:**
   - Go function with the tool logic (in appropriate `internal/` package).
   - Register in tool registry.
   - Update expert seed/config to include tool in `allowed_tools` for eligible experts.
5. **Test:** call the tool via a test expert; verify the LLM correctly invokes it based on description.
6. **Document:** add to `IMPLEMENTATION_HANDOFF.md`.

## 8.2 Example: "Backend expert produced code that doesn't compile"

1. **Observe:** validation pipeline (§11 of design doc) caught syntax error at line 42.
2. **Feed back:** validation output goes into expert's OTA loop as "observation".
3. **Think:** LLM reads the error, decides fix.
4. **Act:** LLM emits corrected code.
5. **Validate again:** if passes → artifact goes to Code Reviewer. If fails → iterate up to `max_loop_iterations` (default 5).
6. **Escalate:** if still failing, escalate to Code Reviewer for suggestions, then to client via `AskClient` if still stuck.

## 8.3 Example: "Client rejected the architecture at gate 2"

1. Workflow status was `paused_for_approval`. Client responded with `request_changes` and free-text feedback.
2. Workflow engine sets status back to `running`, transitions to `HIGH_LEVEL_DESIGN` again (rollback to previous phase).
3. Feedback is posted as `client_response` event on blackboard.
4. System Design + Security experts see the event, revise, re-post.
5. New cross-verification round. New approval gate.

## 8.4 Example: "Cost hit soft limit at 75% of budget"

1. Cost monitor detects `cost_spent_usd / cost_budget_usd > 0.75`.
2. Emits notification to admin + client (no workflow pause).
3. Client can preemptively raise budget from the workflow dashboard.
4. If cost reaches 100% (hard limit): workflow auto-pauses with status `paused_for_approval` and gate `budget_exceeded`. Client must approve budget increase to resume.

---

## Closing Note — For Future Model or Engineer Picking This Up

Read this document. Read `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`. Read `DOMAIN_EXPERT_COLLABORATION_DESIGN.md`. Read `IMPLEMENTATION_HANDOFF.md` for current state.

Do not read the source course transcripts unless you are debugging a specific concept described here and want the original phrasing. The three docs above are the complete distillation.

When you're unsure: ask the admin (Kiran). Do not fabricate. Do not silently guess.

When a bug turns out to be architectural: raise it as a locked-decision-conflict per `DOMAIN_EXPERT_COLLABORATION_DESIGN.md` §16. Do not quietly override.

Good luck.
