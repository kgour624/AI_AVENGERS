# Arpit Bhiyani (Last Video) — Memory, Multi-Agent Systems & Incident Auto-Remediation

> **Source:** `Transcripts/Arpit-Bhiyani- Master class Last video Transcript.md` (~1,103 lines).
> **Focus:** memory taxonomy & context management, memory write/decay, multi-agent patterns, orchestration determinism, incident auto-remediation system design, evals.
> **How produced:** read end-to-end; only reusable principles, patterns, anti-patterns, concrete techniques and quotable lines kept.

---

## 1. Memory Taxonomy (the layers)

| Type | What it is | Storage | Lookup style |
|---|---|---|---|
| **In-context (working)** | Everything in the context window: system prompt, tool defs, user prompt, tool outputs, reasoning traces, files | The context window | None — already loaded |
| **External key-value** | Named facts, configs, preferences, decisions, rate limits | Redis, DynamoDB, Postgres, ES, disk | Deterministic key lookup |
| **Vector** | Verbose text/incidents for semantic recall | Any vector DB (PGVector, ES, Chroma) | Cosine / ANN (HNSW) |
| **Episodic** | Key facts/events/decisions extracted from a conversation | Redis/Dynamo/Postgres/graph | Mostly structured query |
| **Semantic** | Facts distilled by aggregating *multiple* episodes (repeated signal → preference) | same | Structured/aggregate |
| **Procedural** | What actions the agent took and what worked — self-improvement | same | Structured |

- **Rule:** Do not treat every answer as a vector problem. *Named fact* → KV/graph; *verbose text* → vector; *repeated theme* → semantic memory.
- **Rule:** Use an LLM to *extract* episodic facts, then store/query them as structured data when possible.
- **Anti-pattern:** Blind "dump everything into a vector DB and retrieve" — loses precision and determinism.

## 2. Context-Window Management

| Strategy | Mechanism | Pros | Cons / failure mode |
|---|---|---|---|
| **Eviction** | Remove content at budget (e.g. sliding window at 90%) | Cheap, deterministic | Can drop needed content; naive window can evict the system prompt |
| **Summarization / compaction** | LLM compresses older messages | General-purpose, safe default; frontier-lab default | Lossy, slower; can skip critical facts |
| **External storage** | Push facts to DB/disk, re-retrieve | Keeps working context small | Needs deterministic keys or reliable retrieval |

- **Lost-in-the-middle:** LLMs attend most to start and end; middle suffocates. Large windows ≠ equal attention.
- **Quotable:** *"Copy-paste your entire prompt twice and you get a better result"* (forces attention; doubles input tokens).
- **Eviction heuristics:** remove reasoning steps once the decision is taken; remove verbose tool outputs once the subtask is done. **Rule:** *If unsure at step 8 whether step 3's output is needed later, don't evict it.*
- **Deterministic-workflow advantage:** if the shape is known, evict safely (e.g. after step 5, only step 5's output matters). Implement context as a **list of tagged dicts**.
- **Summarization prompt must be prescriptive** (like a doctor's prescription): name exactly what to preserve (commands, numbers, table names, constraint names, breaking changes, action items). Generic "summarize this" loses detail.
- **Rolling summarization** only works for long-running/coding agents, not stateless per-request chat APIs.
- **Tokenizer cost trick:** approximate with word count or a local model (±100–300 tokens is fine).
- **System prompt must live separately** from conversation messages, concatenated at call time — so eviction can't delete it.

## 3. Memory Write Strategy

- **First decide *what* to persist.** *"The cost of noisy memory is very high."* Persisting everything degrades **SNR** → retrieval becomes irrelevant/contradictory → output "goes berserk."
- **Rule of thumb:** persist only decisions that **change the behavior of the agent or future execution**.
- **Operations:** `add` / `upsert` / `delete`.
  - "I like pizza" then "I like burger" → **append**.
  - "I used to like pizza, now I like burger" → **replace/upsert** (edge invalidated).
- **Evolving-edge pattern:** LLM outputs structured memory actions; a deterministic applier mutates the store. Prescriptive prompt defines exactly what constitutes `add` vs `upsert` vs `delete`.
- **Canonical keys must be descriptive:** `user_<id>_preference_food` not `pref`.
- **Tag every memory** at write time with task/agent/chat ID → observability/traceability.
- **Constrain relations:** define an enum (Pydantic) of allowed relations; don't let the LLM invent them.
- **Output format:** RDF triple (subject–verb–object) is easier to emit/parse than free JSON.
- **Cost anti-pattern:** an LLM call on *every* message. Correct: batch 5–10 messages, cheaper model, or run extraction **asynchronously** (CDC/Kafka).
- **Use classic NLP when input is predictable** (NER, pronoun disambiguation); fall back to LLMs for messy input.

## 4. Weighted Retention & Decay

- **Rule:** facts stay as-is; *preferences* get weights/decay. "New Delhi is capital" never decays; "favorite language = Java" must.
- **Exponential decay:** weight = e^(−x) on time-since-registered/access; below a floor → forgotten.
- **Critical events exempt:** marriage/baby must not decay — flag "critical, never forget."
- **Frequency/PageRank analogy:** a repeatedly referenced memory gains weight; simplest proxy = LFU-style access counting.
- **Explicit override:** "remember this forever" → pin it.

## 5. Summarization Types & Hierarchical Consolidation

- **Use-case-shaped summaries:** choose the output format the consumer needs.
- **Hierarchical consolidation:** accumulate structured items (decisions, open items) across the session, consolidate at the **end** — not per-event.
  - Example: create Jira tickets *after human review*, not one tool call per ticket.
  - Example: meeting notes → sectioned decisions / who said what / next steps / open items; each 10-min window merged into a growing doc.
- **Growing-doc pattern:** process comments in batches of 10, rewrite the summary doc each batch.

## 6. Memory Recall (graph query)

- Give the LLM the **schema**: node types, node-ID format, and a **limited relation list**. LLM writes a Cypher/graph query.
- **Rule:** restrict relation count — don't let relations "go berserk." A limited relation set → easy deterministic query crafting.

## 7. Multi-Agent Systems

Core: one agent can't do everything → split into specialized agents.
- **Analogy:** multi-agent ≈ microservices; single agent ≈ monolith. Each agent can own its model, context, and database.

### 7.1 When to go multi-agent
- Subtasks independent → parallelize for speed.
- Need specialized models (planning=Opus, execution=Sonnet; Gemini for English, another for multilingual).
- Differing context lengths per task.
- Single-responsibility / RBAC / tool isolation.

### 7.2 Orchestrator–Specialist
- Orchestrator plans, splits into subtasks, hands off, waits, accumulates only each subagent's **final output**.
- Handoff = create a sub-process/thread/goroutine running that agent's loop; **prevents context bleed** (subagent doesn't see the orchestrator's conversation).
- Subtask splitting is **structured output** (Pydantic: title, description, agent_type).

### 7.3 Critic–Refiner
- Critic re-validates the *final output* (factual consistency, format, SQL performance) and may run its own loop; refiner fixes.
- **Rule:** apply when **correctness > speed** (costs tokens/time).

### 7.4 Mixture of Agents
- Same task to multiple agents, then a **synthesizer** picks/combines the best report. Can stack a critic–refiner. *"Only done by people with a lot of money."*

### 7.5 Deadlocking Agents (anti-pattern)
- **Scenario:** refund agent says "if damaged → escalate to shipping"; shipping says "if refund requested → confirm with refund agent" → infinite ping-pong.
- **Rule:** be ultra-careful with agent ownership crossing team boundaries; prompts are English/non-deterministic → cycles are easy.
- **Fixes:** explicit max-hop counter; default action after N; a **supervisor agent** ("is there a deadlock risk?"); deterministic cycle detection in the delegation graph.

## 8. Orchestration: Determinism vs Intelligence

- **Quotable:** *"Use LLM for intelligence, not orchestration."*
- If you *know* the workflow, don't burn an LLM call deciding to wait — write the loop. Letting the agent decide wastes tokens and kills determinism.
- **Rule:** deterministic workflow → **Temporal/Airflow/DAGs**; LLM only where intelligence is required.
- Teams preferred shared state (DB) + pull-from-queue over A2A handoff for observability/state/checkpoint-resume.
- **Evolve, don't over-engineer:** start with skills/LLM orchestration (fast GTM), convert to deterministic workflows when cost/scale demands.
- **Reliability anti-pattern:** a dynamic workflow where an agent can spawn 10–100 agents → no SLO possible.
- **Agent SDK vs own loop:** SDK fine for MVP; at scale own the loop / use response API + Temporal.
- **Skills vs agents:** use agents when you need **RBAC between agents**; skills within one agent otherwise.

## 9. RAG / Retrieval in Agent Systems

- Runbook lookup is an **open-ended retrieval problem** → vector DB over indexed runbooks.
- **Deterministic vs semantic:** named facts (rate limit) → deterministic tool call with a fixed key format in the tool description; conceptual/verbose → vector.
- **Prompt must tell the retriever what kinds of things exist** so it forms the right query.
- Standard RAG subsystem: ingestion + vector DB kept updated + query tool.

## 10. Incident Auto-Remediation System (System Design)

**Rule — augment, don't replace:** integrate into the existing on-call workflow; you cannot force teams to change overnight. *"Think of it as augmentation, not replacement."*

**Capacity math:**
- 5,000 alerts/day ≈ **0.006 alerts/sec** avg; peak ~10×; cascading outages ~1–2/sec.
- **Dedup/grouping essential** — else 900 raw events in a 90-sec P1 window → ~3 unique incidents; worst case ~30.
- P99 time-to-first-action: **90s P1 / 5min P2 / 15min P3** → avg ≤30s → **~5 sequential LLM calls** per agent (10 with slack); 3 parallel × 5 = 15 calls/incident; ~150 LLM calls/sec peak = top-level rate budget.

**Reliability rules (harness):**
- Alert if your own queue grows / SLA at risk.
- **Model fallback across ≥2 providers** — *"You can't say 'I can't fix my issue because Gemini is down'."* Separate key + Bedrock/Azure.
- Loop must **not be model-specific**; keep a **prompt registry** (per-model templates). Claude tolerates vagueness; Gemini needs explicitness.
- Watch model **deprecation** — eval + notification-driven changes.

**Architecture:** webhook → ticket → dedup/grouping → priority-split Kafka topics → executors (RAG over runbooks, metrics, logs, code, feature flags, similar past incidents) → Postgres (memory + incident state) → output to PagerDuty/Slack/Jira.
- **Include feature-flag audit** — a today-flipped flag can be the cause even if deploy is old.
- **Similar-incident memory guides but does not force** the path (weight, don't dictate).
- **Triage report must be crisp**, hyperlinked, non-verbose — but citations double as feedback loop and provenance.

**Human-in-the-loop:**
- **Reversibility flag on every tool:** reversible vs irreversible (pod restart = irreversible). Staggered rollout — reversible OK; never irreversible without human.
- "Run" button → command executor MCP server; human decides *what*, tool executes.
- **Stop conditions:** diminishing returns, breached SLA ×2, stuck in loop.

## 11. Evaluation (evals as unit tests)

- Evals are the safety net for every lossy choice: summarization, eviction, context-size vs lost-in-the-middle, model swap, prompt change.
- **Write evals for the critical info your system must never lose** (e.g. "output must contain this string").
- Design evals that *depend on the lossy behavior*: after compression, recall the exact rollback command — eviction loses it, summarization loses it, external storage retains it.
- **If evals pass, you can be fairly certain, despite lost-in-the-middle.**
- Next step for evolving memory: verify the LLM emitted the right `add`/`upsert`/`delete` action.

## 12. Career / Building (short)

- Engineers who win **find problems worth solving**.
- Build your own tools; understand the underlying system — *"you need to know, otherwise the LLM goes in a very random direction."*
- Geniuses are made, not born — experiment mindset.

---

## Relevance to AI Avengers

| Item | Transcript mapping |
|---|---|
| **A5 — design-conflict escalation** | Deadlocking agents: cross-team ownership + conflicting escalation prompts create infinite loops → max-hop cap, default action, or supervisor agent; deterministic cycle detection in the delegation graph. |
| **A7 — cross-verification / fail-closed review** | Critic–refiner (correctness > speed); verification as "the ultimate thing"; reversible-vs-irreversible tool flags; human approval of tool args before execution (fail closed on irreversible actions). |
| **B1 — LLM synthesis of multiple expert answers** | Mixture of Agents: same task to multiple agents + a **synthesizer** producing the best combined report; can stack critic–refiner. Orchestrator–specialist accumulates subagent *final outputs*. |
| **B2 — contradiction detection** | Memory add/upsert/delete ("I used to like X, now I like Y" → replace vs append); noisy memory → contradictory retrieval → "output goes berserk"; supervisor/deadlock detection for conflicting agent intents. |
| **B3 — workflow memory write-back** | Persist only decisions that change future behavior; tag writes with task/agent/chat ID; upsert vs append; extract via batched, **asynchronous** CDC pipeline; procedural memory = what the agent did. |
| **B6 — answer quality scoring** | Evals as unit tests over summarization/context strategy; expected-keyword checks on recall; critic–refiner loops; SNR as a memory-quality metric. |
| **B8 — verification-first (claim → evidence)** | Citations in triage reports double as provenance + feedback signal; give the LLM the schema and have it produce verifiable queries; "agent is as good as the information you provided." |
| **C3 — eval harness** | Purpose-written evals per lossy component; eval-driven prompt iteration; model-swap/deprecation regression detection. |
| **C7 — adversarial debate** | Critic agent critiquing another agent's full output (factual consistency, format, SQL perf); mixture-of-agents where independent reports conflict and a synthesizer adjudicates; deadlock is an emergent adversarial loop to detect. |
| **C10 — reliability-as-product** | ≥2 LLM providers + fallback; prompt registry per model; rate-limit budget; dedup/grouping; reversible-action gating; deterministic orchestration (Temporal/Airflow); *"Use LLM for intelligence, not orchestration."* |
