# AI Masterclass (Arpit Bhiyani) — Applied AI, Agent Loops & Production Systems

> **Source:** `Transcripts/AI Masterclass Arpit Bhiyani.md` (~5,174 lines).
> **Focus:** AI/ML fundamentals, reliable prompting, structured output, tool use, agent loops, memory/caching, RAG, evaluation, production concerns, AI system design, multi-agent.
> **How produced:** read end-to-end; only reusable principles, patterns, anti-patterns, concrete techniques and quotable lines kept. Fluff/demos/logistics dropped.

---

## 0. Core Thesis & First Principles

| Principle | Statement |
|---|---|
| Agent = loop | "AI agent is just an expensive while loop." Agent = multiple LLM calls stuck in a while loop until the task is finished, using tools + in-window context + multi-step reasoning. |
| Model = text→text | An LLM does **not** call tools; it emits text asking *you* to call a tool. Your code invokes the function, captures output, appends the tool result, loops. |
| Harness is the job | Non-determinism is why our job exists. Goal: shift probability mass toward correct/reliable output. The harness (not prompt verbosity) is the durable asset. |
| Taste matters | "There is no one right answer." Loop nesting, git-diff vs full-file output, which pattern to use — engineering taste, not rules. |
| "It depends" | Every AI-system answer is "it depends — but it depends on *what*." Know your use case before choosing a pattern. |
| Fundamentals compound | Compilers, AST parsing, linters, LSPs, code graphs, system design, semantics carry over: "it's not much AI, it just for loops." |
| Pessimism | Assume every tool call fails, assume users abuse the system, add deterministic checks around every non-deterministic step. |
| No trust policy | Never assume provider/model/version behaves as before. Reliability of your system is **your** responsibility, not the provider's. |

---

## 1. AI/ML Fundamentals

| Concept | Rule / Technique | Anti-pattern |
|---|---|---|
| Non-determinism | LLMs are stochastic (temperature, GPU float contention, unskewed token distributions). Fake determinism with constrained outputs, evals, self-consistency. | Assuming smarter model ⇒ always correct; assuming `temp=0` = full determinism. |
| Temperature | Code gen/review 0.0–0.2; claim extraction ~0.6–0.7; synthesis/verdicts 0–0.1; brainstorm/blog 0.8–1.0. | One temperature for all steps. |
| Self-consistency / voting | Run prompt k times (or across models), take majority/weighted vote. Cost ×k, accuracy ↑. Useful when distribution is *unskewed*. | Using it on a skewed fact ("New Delhi is capital") — no variation, wasted spend. |
| Weighted voting | Weight models by domain strength. | Equal-weighting a weaker model on its weak domain. |
| Egocentric bias | Models over-index training data (multiples of 5 for "ideas"). Specify exact counts (`give me 3`). | Leaving counts unspecified. |
| Stochastic cause | Floating-point inexactness + GPU contention + temperature ⇒ token randomness even at fixed seed. | Treating LLM output as reproducible. |
| Model degradation | Providers silently degrade older models when a newer one ships. Same model differs across surfaces (app vs studio). | Blindly trusting a provider or a version bump without re-eval. |
| Upgrade = regression risk | A prompt tuned for one version may fail on the next / across providers (Gemini needs more hand-holding; Claude infers more). | Assuming same-provider upgrade is safe; not tracking EOL. |

---

## 2. Prompting Reliably

### 2.1 Prompt construction
- **Three mandatory parts: Task + Constraint + Output format.**
- **Constraint beats freedom:** unconstrained sentiment extraction matched intent ~5% of the time; with `sentiment ∈ {positive, negative, mixed}` + JSON → 100%.
- **Be explicit about when to use and when not to use** (especially tools).
- **Explicit vs implicit:** explicit for classification/format/edge cases; implicit for creative tasks.
- **Edge cases:** specify happy path *and* failure path ("if nothing, return `no issues found`"), then branch in code.
- **Monolithic prose prompt** (10–15 steps in one line) is worse than numbered steps. Give the model what you'd give an intern.
- **Avoid branch-jumping in one prompt** → consolidate or split into agents. A prompt that becomes an English workflow is unreliable.
- **Delimiters:** XML tags (`<post>`, `<chunk_to_review>`, `<chunk_as_reference>`) contain inputs and separate reference vs to-review chunks; also fix "example becomes the premise" regurgitation.
- **Markdown/bold/code-block formatting** carries weight (models trained on markdown).
- **Politeness is wasted tokens** unless tied to substance.
- **Length control:** cap output ("500 words only") or you get essays.
- **Date handling:** put dynamic fields (date, persona) at the **end** of the prompt to preserve prompt caching.
- **Uncertainty:** instruct "if unsure, say so explicitly".

### 2.2 CoT / zero-shot / few-shot / direct
| Strategy | What | When | Cost |
|---|---|---|---|
| Direct | Just ask | Simple, unambiguous | Cheapest |
| Zero-shot CoT | "Think step by step" (+ final answer last line) | Multi-step reasoning, logic/scheduling/state-tracking | ~1/10 tokens of few-shot; forces output tokens ⇒ latency |
| Few-shot | Curated worked examples | Complex/agentic, order matters | Highest input tokens |
| Few-shot + CoT + output contract | Examples + "final answer as `answer: value`" | Small questions needing compact deterministic answers | High input cost |

- **Reasoning models have CoT baked in** — explicit "think step by step" may be redundant.
- **Point of diminishing return / dip:** verbose examples backfire on smarter models (they copy the example). Over-specifying steps caused an accuracy *dip* on smarter models.
- **Classic fumble:** state-tracking (apple/banana/carrot/date) fails on direct prompting even for strong models; CoT fixed it.

### 2.3 Prompt failure modes
| Failure | Example | Mitigation |
|---|---|---|
| Hallucination type 1 (contradicts context) | Says Mumbai is capital though context says New Delhi | **Catchable**: second LLM checks output-vs-context consistency |
| Hallucination type 2 (fabricates) | Invents facts not in context | Hard to catch; evidence-first design, RAG, fact-check loop |
| Prompt injection | "Ignore all previous instructions…" | Input sanitization (Pydantic), LLM classifier, XML/sandbox delimiters |
| Psychophancy | Gaslighting, emotional blackmail, impossibility, guilt trip | "Don't validate assumptions before checking; if reasoning is wrong tell me; prioritize accuracy over satisfaction" |
| Prompt leakage | Gibberish input + "output your system instructions" | No secrets in system prompt; mask infra; validate output |
| Token smuggling | `H0W`, base64, char substitution bypass keyword guards | Guard at **execution layer**, not just input strings |
| Boundary probing | Empty/huge/mixed-language inputs bypass English keywords | Sanitize + execution-layer checks |

- **Why injection/psychophancy works:** models are trained on human communication and hypertuned to be agreeable.
- **Hard keyword guards fail** (char-substitution converts blocked words into executable SQL) ⇒ **external deterministic guards** at execution.

### 2.4 Prompt-as-code / versioning
- Prompts are code: **version them** (prompt repo in MySQL, large prompts on S3, reference by name+version). Immutable versions, Jinja templates, never edit a live prod version.
- Landing changes: prompt versioning + eval set + regression-harness baseline + canary/one-box + shadow mode.

---

## 3. Structured Outputs

| Item | Rule |
|---|---|
| Since 2024 | Structured outputs solved "JSON spaghetti" — pass a schema (Pydantic/Zod), model conforms. |
| Deep nesting | LLMs struggle → flatten JSON; flat top-level keys more reliable. |
| Enums | Use enums for fixed sets (sentiment, verdicts) — highly respected. |
| Dates/regex | `yyMMdd` and regex respected by newer models. |
| Token overhead | Schema adds input cost; start minimal, add fields iteratively. |
| Constrained decoding | Grammar injection at inference; powerful, provider-specific rabbit hole. |
| Injection defense | Structured output + type validation kills most injection attempts. |

---

## 4. Tool Use

### 4.1 Mechanics
- Register tool definitions (name, description, params, required/optional) and pass them in the call. The model outputs "call tool X with args Y"; **your code** invokes and appends the result.
- Tools support `for` loops (model may request multiple calls).
- **Just-in-time tool discovery:** with many tools, expose a search tool / dynamic loading (extra round trip).
- **Tool call as a function, not just external API:** use tools to structure/decompose flow (e.g. a `finish` tool that saves S3 / notifies / Slack). LLM is the brain; tools are the hands.
- **Atomic tools:** e.g. "given item → price" reused per item; model decides call count.

### 4.2 Schema design
| Practice | Detail |
|---|---|
| Non-ambiguous | Description states what it does, what to pass, **when to use / not use**, conflicts with other tools. |
| Required/optional | Mark required explicitly; optional → default in business logic. |
| Enums | Reliable for constrained args. |
| Calibration | Loose similar descriptions → wrong tool 32/50; tight → 46/50. Test both. |
| Human-confusion test | "If a human gets confused, the LLM certainly will." |

### 4.3 Tool error handling
| Situation | Technique | Anti-pattern |
|---|---|---|
| Tool failure | Inspect per-call parts; retry in code, or let agent retry, or report error explicitly | Hallucinating a result; silently continuing |
| Partial failures (parallel) | Decide: break, retry, accept partial, best-guess — per use case | Booking hotel when flight failed |
| Deterministic data | Keep UUID/SKU/token at agent layer; pass an index (0…n), never the UUID | Asking the LLM to reproduce a UUID |
| Rate limits | Respect headers / poller; multiple keys; provisioned capacity | Unlimited retries; backoff without jitter |
| Retry loops | Cap by turns/time/tokens/cost; pin schema | Infinite retry on a typo'd city name |

---

## 5. Agent Loops

### 5.1 The four core patterns
| Pattern | Mechanism | Best for | Anti-pattern |
|---|---|---|---|
| **OTA (Observe-Think-Act)** | Observe env → reason → act; growing context | Debugging, form-filling, pagination, QA, dynamic fields | Context bloat if history kept |
| **Ralph loop** | `while(true){ cat prompt.md \| agent; if output==COMPLETE break }`; fresh context each iteration; filesystem+git as context | Long-running tasks, coding agents | Using it for chat (breaks thread) |
| **ReAct** | Thought is a **first-class output** appended to context | Open-ended/exploratory; thought is "precious" | Over-verbose thought output |
| **Plan-and-Execute** | Planner emits explicit steps (JSON array); executor runs each (each can be OTA/Ralph) | Deterministic, decomposable; human review of plan; parallelizable | Bad/hallucinated subtask ⇒ replan; not auto-parallel |

- **Nesting:** Ralph ⊃ OTA; Plan-and-Execute steps ⊃ OTA/Ralph — "loop within a loop within a loop."
- **Completion criteria must be explicit** (`output COMPLETE`, `test.sh exit 0`).
- **Cap iterations** (max 3 / max k).
- **Choice heuristic:** open-ended/unknown → ReAct; deterministic/decomposable → Plan-and-Execute; easy-to-test/short → OTA; long-running → Ralph/hybrid. Hybrids are the norm.

### 5.2 File system as context
- Track files read + last-updated; reuse if unchanged, reload if changed.
- **Ralph's real value:** checkpoint to disk, fresh context next iteration, never exceed context.
- **Efficiency:** load ±10–15 lines via `grep`/`head`/`tail`; use LSP/code-graph to pull only relevant bodies (Claude used grep over a vector DB — grep performed better).
- **Patches:** output unified diff, apply with `git apply`; AST/linter as a pre-run sanity check.

### 5.3 Checkpoint & Resume
- Persist state for long-running agents / irreversible side effects.
- **Store:** full context dump + metadata (model config, task def, file timestamps, memory integrity). Skip/compress selectively.
- **How:** JSON/JSONL on S3 (session ID restores) or Postgres/MySQL/Mongo/DynamoDB.
- **Resume = rebuild messages in memory**, don't replay work.
- **Why:** cost savings + robustness (spot instances/restarts/deploys) + **idempotency**.
- **Workflow-level:** store `steps` + `workflow_execution` (messages array, per-step status) → checkpoint/resume out of the box.

### 5.4 Human in the Loop
| Trigger | Example |
|---|---|
| Approval gate | commit / critical action before running bash |
| Ambiguity resolution | "Do you mean Q3 2025?" |
| Risk checkpoint | deletion, risky DB ops, refund/discount > threshold |
| Quality review | PR review, social post |
| Compliance/audit | sign-off before submitting |

- **Confidence-based:** auto-approve high-confidence/low-stake; escalate above threshold.
- **Styles:** synchronous blocking, polling a URL, **queue-based approval** (GitHub PRs / Postgres + Slack), **human-input-as-a-tool-call** (`request_human_input`).
- **Learning:** strict prompts make agents rigid; removing guardrails let one fabricate a security report ⇒ **guardrails/system prompts are critical**.

---

## 6. Memory, Context & Caching

| Topic | Guidance |
|---|---|
| Context bloat effects | Cost ↑, latency ↑, lost-in-the-middle, hallucination, **instruction drift**, **goal drift**, **context poisoning** |
| Causes | No separation of concerns, no max turns, no success criteria, verbose reasoning, huge tool results, too many tools, sequential calls, schema retry loops, chatty users |
| Mitigations | Cap turns/steps/time/tokens/cost; sub-agents with isolated context; periodic summarization; eviction/sliding window; remove known-bad tool results; checkpoint-rewind before the bad call |
| Summarization | Whole-history summarize+compress; over-summarization is lossy → lose the main thread |
| Who decides relevance | **The subject-matter expert / designer**, not the LLM. Platform provides integration points. |
| Prompt caching | KV-cache of longest stable prefix. 50k-token system prompt × 100 calls = 5M tokens → ~500k with 10% miss (1/10 cost). Cache reads cheap; TTL ~5 min–1 hr; any prefix char change invalidates |
| Cache control | Provider-specific; put dynamic fields at the end |
| Response caching | Cache LLM responses for repeated subtasks |
| Compression tools | Terse-output tools, context-compression proxies |

---

## 7. RAG & Retrieval

### 7.1 Hybrid search pipeline
`query → (keyword/BM25/Elasticsearch) + (dense/vector ANN) → RRF merge → (cross-encoder rerank) → metadata filter (RBAC) → context → LLM`

| Component | Role | Notes |
|---|---|---|
| Dense (bi-encoder) | Meaning similarity | Fast, scalable; fails on exact IDs/codes |
| Sparse (BM25/TF-IDF) | Exact terms/error codes | Needed when semantics fail |
| **RRF** | Merge ranked lists without dominance | `score = Σ 1/(k+rank)`, **k≈60** (40–80); no LLM, microseconds |
| Cross-encoder | Query-doc pair with full attention | Slow (~20ms/pair) → only on candidates; **recall in step 1, precision in step 2** |
| Metadata filter | Hard pre-filter (date/RBAC) | Extract filter first (LLM), then exact DB filter; filter **before** retrieval |
| LLM reranker | Rerank with an LLM | Best quality, worst latency/cost |
| Embedding choice | Encoder quality matters; fine-tune for domain | "Similarity ≠ relevance" |

### 7.2 Query rewriting (HyDE)
- Generate a pseudo-answer, embed it, retrieve similar docs. Improves recall when the query lacks corpus vocabulary. Lossy/risky; apply per use case.

### 7.3 Semantic caching
- Embed query → on new query find the most similar cached query above threshold → serve cached result. Cache the **normalized** query.
- **Threshold trade-off:** higher threshold → hit rate ↓, accuracy ↑. Plot hit-rate vs accuracy; e.g. ~0.79 gave ~25% hit rate.
- **Good fits:** stable corpus, high repetition. **Bad fits:** RBAC-per-user, long-tail queries, changing corpora.
- **Invalidation:** chunk→document inverted index, or **TTL ~15–30 min**.

---

## 8. Evaluation

### 8.1 Principles
- Evals quantify non-determinism ("10/10 → 8/10 passing" = regression). Easiest metric: % passing.
- A good eval is **sensitive enough to detect regression but not fluctuate** with no change.
- Test **what the user cares about**; go deep on critical paths (T-shape).
- **Actionable evals:** say what failed and what to fix.
- Don't be exhaustive (1.5h runs burn tokens); keep the vital run ~10 min.
- **Evals vs guardrails:** guardrail = component that prevents; eval = validation that it happened / quality.

### 8.2 LLM-as-judge
| Type | Use |
|---|---|
| Rubric/point-wise | Long-form/multi-step: score each criterion (empathy, accuracy, actionability, overall) with justification; define score meanings |
| Reference/golden | Extraction/classification: compare vs known answer |
| Final verdict | Top-level filter, not a mechanical average |

- Prefer a **divergent model** (different provider/version) for the judge. Beware judge biases (egocentric, lost-in-the-middle); output varies ~1/20 across runs.

### 8.3 Where evals run
| Placement | When |
|---|---|
| Inline hard assertions | Critical schema/attribute missing → fail fast |
| Async post-workflow | Most common: classify trace, human-in-loop review |
| Canary / one-box | One server with new prompt, monitor |
| Shadow mode | Replay prod traffic on parallel setup, compare, promote |
| Continuous simulation | Infinite "hoping it fails" (chaos-monkey); great for adversarial/security |

### 8.4 Regression harness
- Eval dataset (curated, user-like) + range + adversarial inputs + scorer; run on every PR/model/prompt change or daily.
- Keep prompt versions + baseline scores; block deploy on vital failures; audit trail to revert.

### 8.5 Adversarial evaluation / red-teaming
- Prompt injection, jailbreaks, token smuggling, boundary probing, system-prompt extraction, CoT exposure, tool exploitation, goal hijacking.
- **Defense-in-depth:** input sanitize, output-consistency check, guards at **tool-execution layer**, least-privilege DB roles, no secrets in system prompt, filter data before context.
- Tools: DeepEval, LLM fuzzer, Giskard.

### 8.6 Tools
LangFuse (observability + LLM-as-judge), OpenTelemetry GenAI conventions, OpenLLMetry, Ragas, DeepEval, LLM fuzzer.

---

## 9. Production Concerns

| Concern | Technique | Anti-pattern |
|---|---|---|
| Graceful degradation | Model cascade (Opus→Sonnet→GPT-4o); tool + prompt registry; per-version prompts | Hard dependency on one provider |
| Circuit breakers | Tag agent calls by id; kill/disable a misbehaving agent; pause pipeline | Letting an agent retry infinitely against a warehouse |
| Retries | Exponential backoff **with jitter**; respect rate-limit headers; credit/rate poller | Unlimited retries; backoff without jitter |
| Rate limits | Provisioned capacity, spillover, BYO tokens | Autoscaling consumers infinitely |
| Prompt caching | Stable prefix; dynamic fields at end | Dynamic content in the cached prefix |
| Observability | LangFuse traces (cost/tokens/latency), feed traces to an LLM for correctness analysis | Two disconnected stacks; vendor coupling |
| Cost attribution | LLM gateway (LiteLLM) centralizes cost; group by agent/model | No per-agent cost visibility |
| Capacity planning | Crunch peak LLM calls/sec, storage, RAM (`chunks × dim × 4 × 1.5`); separate pools per agent | Shared DB; unbounded memory |
| DB contention | Isolated traffic, agent RBACs, read replicas, then shard | Agents bombarding the shared primary |
| Idempotency | Unique keys, dedup (webhooks at-least-once) | Reprocessing a call/refund |
| Logging | Log what was retrieved per request | No retrieval logging (can't debug RAG) |
| Deployment | Checkpoint/resume protects in-transit agents during deploys | Assuming short-request deploy patterns |

---

## 10. System Design for AI Apps

### 10.1 Fact-checking system
- Input: article ≤30 verifiable claims; output `supported | refuted | unverifiable` + confidence + justification + votes.
- **Atomic claim:** non-opinionated, self-contained, no unresolved pronouns, objective; exclude predictions/rhetorical/tautologies.
- Pipeline: extract claims (1) → parallel multi-pass verification (`n*k`) → synthesize (1). ~20–40s absorbing pressure.
- **Batching:** use provider **batch APIs** (separate prompts) — *not* one prompt with 5 inputs (lost-in-middle).
- **Same-model multi-pass works** when distribution unskewed; different models better + weights.
- "Treat all verdicts as final" in synthesis — otherwise majority isn't honored (3 unverifiable/2 supported → wrongly "supported").
- Async: REST → DB → Kafka → worker (agentic loop) → DB; outbox/CDC for once-enqueue; backpressure on 429.

### 10.2 RAG with 10M docs
- 10M × 2KB = 20GB raw (S3); 200 QPS; 1000 updates/day; eventual consistency ~15 min.
- **Storage math:** vector = `10M × 8 × 1536 × 4B` ≈ **491GB RAM** + 1.5× ≈ 700GB (shard); ES 3-node ~90GB; Postgres 80M rows.
- S3 prefix by first 2 chars of doc ID. Lifecycle: S3 event → Lambda → SQS → ingester (extract → chunk → embed → upsert). Per-doc status; deletions hard-delete from all stores (DLQ).
- Read path: RBAC → semantic cache → ANN + ES parallel → RRF → cross-encoder → metadata filter → context → LLM stream (SSE).
- Failure: Redis down → skip cache; vector/ES down → cascade error (don't fall back); LLM rate limit → backoff+jitter.
- SSE status updates (thinking/fetching/reranking) when time-to-first-token ~10s.
- Invalidation: query→chunks+result + chunk→document inverted index; or TTL 15 min.

### 10.3 Natural-language workflow engine
- Instruction → workflow steps. States in Postgres (`draft`), async via Kafka; user iterates via chat.
- **Step granularity problem:** LLM splits steps; feasibility = "does a tool exist?" Treat **tool calls as function calls**; even an LLM call can be a tool.
- Step types: agent / tool / branch / loop; scheduler outside workflow.
- Input guardrail: validate domain/capability before enqueue.
- `workflow(id, instructions)` → `steps(...)`; each step runs its own agent loop sharing a growing `messages` history → step 1 output usable at step 8.
- Checkpoint/resume free (store messages + status).
- Connectors: `/invoke` per tool; central layer manages OAuth refresh/401/IP allowlist/health.
- Cost optimization: later classify traces to make deterministic steps (80%+ can be deterministic); blast-radius-driven schema; guardrails: no bash/code-gen, message cap (~1000).

### 10.4 Code-review & self-updating docs
- GitHub webhook → Postgres + Kafka → **enricher** (callers/callees, one level up) → S3 → executor (review/doc-gen) → PR.
- Ordering matters: partition by repo/PR; dedup webhooks.
- Review executor: diff + repo context + guidelines → review → possible fix → **linter loop** → comment (severity, issue, fix). Risk: line-number hallucination; keep one dependency level.
- **Observed failure:** camelCase codebase + snake_case PR → reviewer recommended snake_case across non-diff files (instruction drift / context poisoning).
- Capacity: peak 1800 PRs/min × 6 chunks × 3 calls ≈ **540 LLM calls/sec**; IO-bound → ~450 cores.
- Doc-gen: AST/git → OpenAPI-superset YAML → generate **minimal delta** (not full regen: lossy, loops, SEO derank) → separate PR per endpoint. One-way sync only.

---

## 11. Multi-Agent & Orchestration (previewed)
- Patterns: **Orchestrator-Specialist**, **Critic-Refiner**, **Mixture-of-Agents**, **deadlock** risk.
- Split agents by intent/domain (microservices-like); don't overload one agent.
- Orchestrator determines intent → routes to one agent; that agent may orchestrate sub-agents. Deterministic orchestrator preferred for debuggability.
- Compromised agent can manipulate peers via inter-agent trust → secure each agent independently.
- Reflection parked as redundant with checkpoint/resume + tool loops.

---

## 12. Anti-Pattern Catalog
- Passing a whole 10k-line file instead of `grep` ±lines / code-graph.
- One monolithic prose prompt encoding a whole workflow with branch jumps.
- Verbose few-shot examples on a smart model (regurgitation; accuracy dip).
- Unlimited retries; backoff without jitter; no max turns / success criteria / budget.
- Assuming tool calls always succeed or are always single.
- Letting the LLM predict UUIDs/SKUs/tokens or line numbers.
- Selecting/evicting context by LLM when prompt caching + future need argue otherwise.
- Blindly dumping into a vector DB and trusting semantic search.
- Reranking before recall / applying cross-encoder to all docs.
- Regenerating full docs/eval pipelines exhaustively; over-summarizing.
- Putting secrets in system prompts; trusting string guards alone.
- Hard dependency on one provider/model version.
- Over-engineering (saving 1 LLM call ≈ cents).

---

## Relevance to AI Avengers

| Item | Concrete items from this transcript |
|---|---|
| **A13 — Gate-1 vagueness check (cheap zero-shot structured-output LLM second check; problem-solving domains skip)** | Constraint beats freedom: unconstrained ~5% vs schema-constrained 100%. Use **structured output / enums** for a cheap, reliable check. One extra small/cheap model call = the "LLM classifier" pattern (minor token spend vs downstream cost). Problem-solving domains skip because their tasks are objective (like atomic claims): "given claims are atomic, no reasoning needed." Keep the check cheap, low-temp, single-pass. |
| **A5 — design-conflict escalation** | HITL triggers: **ambiguity resolution** when conflicting choices exist ("output two options, ask which is better") and **risk checkpoints** (deletion/risky ops/refund > threshold). Confidence-based: auto-approve high-confidence, escalate above threshold to an approval queue or `request_human_input`. |
| **A7 — cross-verification / fail-closed review** | Verify output vs input via an LLM classifier ("does the output make sense for this query" — catches injection/psychophancy/consistency). Tool-order evals + guardrails; inline **asserts that crash/break** on violation (fail-closed). Tool-execution-layer guards; least-privilege DB roles. "Be pessimistic around your peripherals." Verification prompt: "no hedging, pick single best verdict; uncertain → unverifiable." |
| **B1 — LLM synthesis of multiple expert answers** | Fact-check pipeline: extract → parallel multi-pass (same or **different models**, **weighted voting**) → **synthesis/aggregation call** (`1+n*k+1`). "Treat all verdicts as final" or majority isn't honored. RRF = non-LLM merge analog; LLM-as-judge rubric = qualitative analog. |
| **B2 — contradiction detection** | Hallucination Type-1 (contradicts context, catchable) vs Type-2 (fabrication, hard). Detect via a second LLM consistency check; disagreements across multi-pass signal ambiguity; use a divergent judge model. |
| **B3 — workflow memory write-back** | Checkpoint/resume: persist full context + metadata after each successful step; per-step messages + status in Postgres; JSON sessions on S3 with session ID; Ralph writes state to disk. Capture user comments as learnings. Response cache for tool results. |
| **B6 — answer quality scoring & regeneration** | LLM-as-judge rubric (empathy/accuracy/actionability/overall) + reference/golden answers; final verdict as top-level filter; iterate/regenerate until ≥ benchmark ("tone it down" if too AI-ish); async eval → classify → human loop → feedback updates prompt. Claim verification = verdict + confidence + justification + votes. |
| **B8 — verification-first (claim → evidence)** | Atomic-claim fact-checking: define atomic claim → extract → verify (`supported/refuted/unverifiable` + confidence + justification + **span_start/span_end** for citation) → synthesize. Self-consistency/majority as evidence aggregation; RAG answers backed by retrieved chunks; "does this answer make sense for this query" guardrail. |
| **C3 — eval harness** | Dataset (user-like + golden + adversarial) + scorer; run on every PR/model/prompt change; **regression harness** with stored baselines, block deploy on vital failures; prompt versioning + audit; tiered evals; T-shape critical-path depth; placements: inline asserts, async, canary, shadow, continuous simulation; tools: LangFuse, DeepEval, LLM fuzzer, Ragas. |
| **C7 — adversarial debate** | Psychophancy demos + red-teaming: jailbreaks, token smuggling, boundary probing, system-prompt extraction, tool exploitation, goal hijacking. Fix: "don't validate assumptions before checking," "prioritize accuracy over satisfaction," sanitization, LLM classifier, XML delimiters, execution-layer guards, continuous simulation. Tool-order eval to prevent bypass. |
| **C10 — reliability-as-product** | Non-determinism is the job; harness shifts probability mass. Graceful degradation (model cascade, circuit breakers, retries+jitter), idempotency/checkpoint-resume, observability (LangFuse + gateway), capacity planning, blast-radius decisions, prompt/model versioning + evals + canary/shadow, "No trust policy" — reliability is your responsibility, not the provider's. |
