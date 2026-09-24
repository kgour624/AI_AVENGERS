# AI Avengers — Reliability → Intelligence → Moat Roadmap (A13 → C10)

| Field | Value |
|-------|-------|
| **Document ID** | `RIM_ROADMAP_v2` |
| **Status** | **DESIGN LOCKED — Phase A (A1–A19) DONE on fork branch; B1+ pending** |
| **Single source of truth for** | The locked task order from A13 → C10 and the mandatory per-task working protocol |
| **Depends on (done)** | Phase A1–A12 (merged to `origin/main` via PR #5; roadmap doc via PR #6) |
| **Audience** | Any engineer or AI coding agent with **no prior deep codebase knowledge** |
| **Last updated** | 2026-09-24 |
| **v2 changes** | Added §3.1 design principles distilled from `Learnings/` (Byte-by-Byte AI, Arpit Bhiyani masterclass + last video); enriched A13/B1/B2/B3/B6/B7/B8/C3/C7/C10 cards; added **A19** (deadlock guard) and **B9** (context-budget policy) surfaced by the learnings. |
| **Related (do not replace)** | `docs/CURRENT_ARCHITECTURE.md`, `docs/TARGET_MICROSERVICES_ARCHITECTURE.md`, `docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md`, `KNOWLEDGE_HUB.md`, `Learnings/README.md` |

---

## 0. How to use this document (READ FIRST)

This file is the **only** place the task order lives. If anything disagrees with a chat message, this doc wins.

### 0.1 For a human engineer
1. Read §1 (phases + order), §2 (mandatory per-task protocol — **strict**), §3 (constraints + §3.1 principles).
2. Pick the **next un-done** task in §4 (Phase A), §5 (Phase B), §6 (Phase C). Do **not** jump.
3. Each task card gives: *Kya / Kyo / Change / Knowledge / Files+risk+rollback / Limitation*.
4. Mark the task done in this doc's status table when merged.

### 0.2 For an AI coding agent (system constraints — must obey)
```
SYSTEM CONSTRAINTS:
1. This file is the single source of truth for task order and task scope.
2. Follow §2 protocol EXACTLY for every task. No code before explicit approval.
3. Never do two tasks at once. One task -> 3-line summary -> ask before the next.
4. Never merge the two chat systems: standalone Q&A (orchestrator/decision/
   chinawall/message) vs workflow deliverable chat (workflow/*). Separate forever.
5. Preserve China Wall semantics, cited-answer invariants, and the app's
   single-write-path rules (blackboard events -> projector -> workflow_tasks).
6. Apply §3.1 principles to every design decision in the task.
7. No local Go toolchain is assumed. Every task ends with: "user/CI run
   `go build ./...` + the listed manual test".
8. If a task is ambiguous, STOP and ask. Do not invent scope.
9. Update this doc's status table (and the memory backlog) when a task is done.
```

---

## 1. Phases and locked order

```
Phase A  RELIABILITY   -> make the existing product correct and crash-safe first
Phase B  INTELLIGENCE  -> make the answers/workflows actually smarter
Phase C  MOAT          -> build what nobody can copy (provenance, eval, trust)
```

**Why this order:** a smarter system on top of a crashing/lossy system is worse than a simple system that is correct. Reliability is a prerequisite, not a phase to "come back to".

| Phase | Theme | Tasks |
|-------|-------|-------|
| A | Reliability | A1–A12 **done**; A13–A19 remaining |
| B | Intelligence | B1–B9 |
| C | Moat | C1–C10 |

---

## 2. Mandatory per-task protocol (STRICT)

> This is the same template used for A1–A12. It is **not optional**.
> Every task, without exception, produces the following block and then **waits**.

### 2.1 The task card (must be shown before any code)

```
TASK #<id> (<An/Bn/Cn>) — <one-line title>

1. Kya karenge — simple bhasha me
   <What we build/change, in plain language. No jargon.>

2. Kyo — kaunsa problem solve ho raha hai
   <The concrete problem this fixes, with file:line evidence.>

3. Kya change dikhega — user ko kya farak padega
   <Observable behaviour change: what the client/admin/dev will see.>

4. Knowledge chahiye
   <Table: which skills are needed (System Design / Go / AI / Multi-agent /
    DB) so the agent can guide, or the user can take over.>

5. Files + risk + rollback
   <Exact files touched; risk level + why; how to revert (commit/hunk).>

6. Limitation (tumhari help)
   <What cannot be verified locally; what the user/CI must run.>
```

### 2.2 The loop (do not skip steps)

1. **Present** the task card for the next un-done task.
2. **Wait** for explicit approval from the user: the word **`start`** (or equivalent).
   - No approval → no code. Reading code to verify the card is allowed; editing is not.
3. **Implement** only that task's scope.
4. **Verify** as far as possible (statics/read-back; no local Go assumed).
5. Give a **3-line summary** only:
   ```
   1. <what changed>
   2. <result / effect>
   3. <files + "go build ./..." reminder>
   ```
6. **Ask** before presenting the next task's card. Never bundle tasks.
7. **Record**: mark the task done in §7 status table + update memory backlog.

### 2.3 Hard rules
- Bundling two tasks in one approval = forbidden.
- Silent scope expansion = forbidden. If a fix needs extra files, say so in the card first.
- "Temporary" dead code / swallowed errors / fabricated IDs = forbidden (see `KNOWLEDGE_HUB.md` §6.3).
- Every write path must stay honest: blackboard → projector is the single write path for `workflow_tasks`.

---

## 3. Global constraints (apply to every phase)

| # | Constraint |
|---|------------|
| G1 | Chat Q&A and build-workflow stay separate packages/domains forever. |
| G2 | China Wall (rerank → coverage → generate-with-citations → strip) semantics preserved for chat. |
| G3 | No fabricated UUIDs / no dead code paths / no swallowed errors on write paths. |
| G4 | Fail-closed by default: verification/approval/gates must never pass on missing data. |
| G5 | Cost attribution stays in `ModelGateway`; no direct vendor calls from Aider/tools. |
| G6 | Public contracts change by OpenAPI/event-schema first (`docs/INTERFACE_FIRST_CONTRACT.md`). |
| G7 | Every task is crash-safe: a pod restart must not skip human gates or re-run completed work. |
| G8 | Deliver via fork branch → PR to `origin/main` (direct push to `origin/main` is not permitted for the working account). |

### 3.1 Design principles (distilled from `Learnings/` — apply in every task)

> Source: `Learnings/byte_by_byte_ai_6_week.md`, `Learnings/arpit_bhiyani_ai_masterclass.md`, `Learnings/arpit_bhiyani_last_video.md`. These are the durable rules the courses converged on; use them to judge design choices.

| # | Principle | Rule to apply here | Anti-pattern to avoid |
|---|-----------|--------------------|-----------------------|
| P1 | **Constraint beats freedom** | Any LLM step that must be machine-parsed uses a strict schema + enums (structured output), not free prose. | Free-text output we then regex-parse. |
| P2 | **Use LLM for intelligence, not orchestration** | The workflow engine keeps deterministic control flow (DAG, waves, gates). LLM only where a real judgement is needed. | Letting the model decide when to wait/branch/spawn. |
| P3 | **Fail-closed, pessimistic periphery** | Assume every tool/DB/LLM call fails; assert around every non-deterministic step; on missing data, refuse — never guess. | Silently continuing after an error; defaulting to "approved". |
| P4 | **Equality of signal: memory quality > quantity** | Persist only facts that **change future behaviour**; noisy memory lowers SNR and causes contradictory output. | Dumping everything into memory/vector store. |
| P5 | **Context is a budget** | Bound history; beware lost-in-the-middle; summarization must be **prescriptive** (name what to keep); keep system prompt in a separate list. | Unbounded context; generic "summarize this". |
| P6 | **Idempotency + checkpoint/resume** | Side effects guarded by idempotency keys; state persisted so restart never re-runs irreversibly. | Re-calling external effects after a crash. |
| P7 | **Evals are unit tests** | Every lossy/non-deterministic component gets a golden-set eval; evaluate components **before** chaining. | Judging only end-to-end, not knowing what broke. |
| P8 | **No model/provider trust** | Prompt+model registry (per-model templates); keep a second provider fallback; watch deprecation. | Hard-coding one model/version across the system. |
| P9 | **Citations = verification + provenance + feedback** | Prefer span-anchored evidence; an untraceable claim is "unverified", not silently asserted. | Attaching citations after the fact without claim mapping. |
| P10 | **Deterministic-first, evolve** | Start simple; convert LLM orchestration to deterministic workflows when cost/scale demands. | Over-engineering with agents where a loop would do. |
| P11 | **Keep gates cheap** | Cheap model, low temperature, single pass, no chain-of-thought for mechanical checks (vagueness, coverage). | Heavy prompting/CoT on a cheap gate. |

---

## 4. Phase A — Reliability (remaining: A13 → A19)

> All A-item problems below are **verified** to exist in code (file:line). Order is by severity.

### A13 — `decision/engine.go`: `gate1WithLLM` runs unconditionally (defeats Gate1Skip)
| Part | Detail |
|------|--------|
| **Kya** | Gate 1 returns `nil` in two different cases: (a) problem-solving domain (skip), (b) keyword check found nothing. `Process` cannot tell them apart and always calls `gate1WithLLM`, which has **no** Gate1Skip guard. |
| **Kyo** | `decision/engine.go:132-143` calls `gate1` then `gate1WithLLM` unconditionally; `gate1:258-260` skips for problem-solving domains but the LLM check re-enables the ASK path. Contradicts "Problem-solving experts skip Gate 1 entirely" + adds an LLM call to every clear question. |
| **Change** | Problem-solving experts (DSA/coding) are never asked a clarification ASK; clear/long questions skip the extra LLM vagueness call. |
| **Design (§3.1)** | **P11/P1:** the check stays cheap — cheap model, `Temperature: 0.1`, single pass, **strict schema** `{"clear": bool, "missing": [string]}` with a fixed enum verdict; no few-shot/CoT. **P3:** on LLM error, fail safe → return "not vague" (do not block). **P10/P2:** problem-solving domains are objective/verifiable → skip the LLM check (deterministic rule), not an LLM decision. Future knob: if the gate mislabels, upgrade the gate model tier (not add reasoning). |
| **Knowledge** | Go (control flow, signal-vs-nil refactor), AI (zero-shot structured output), System Design (gate/fail policy). |
| **Files/Risk/Rollback** | `internal/decision/engine.go` only. Risk: Low. Rollback: 1 hunk. |
| **Limitation** | `go build ./...`; manual: ask a vague DSA question + a clear long question to a problem-solving expert. |

### A14 — `chinawall/enforcer.go`: zero-citation check runs before `enforceStructured`
| Part | Detail |
|------|--------|
| **Kya** | The flat-path `len(Citations)==0 → retry/refuse` check runs **before** the `len(TemplateSections)>0 → enforceStructured` branch, so structured answers whose sections are code/test-only (citation-exempt) can never succeed. |
| **Kyo** | `enforcer.go:209-215` precedes `:223`; code/test_cases are deliberately citation-exempt → legitimate empty citations become a refusal. |
| **Change** | Structured category answers with code/test sections return success instead of a citation-failure refusal. |
| **Design (§3.1)** | **P3:** keep fail-closed where citations ARE required (prose claims), but exempt verifiably code/test-only sections. |
| **Knowledge** | Go, China Wall design, template/section model. |
| **Files/Risk/Rollback** | `internal/chinawall/enforcer.go` only. Risk: Medium (touches refusal path). Rollback: 1 hunk. |
| **Limitation** | `go build ./...`; manual: run a structured-category expert whose answer is code-only. |

### A15 — `cross_verifier.go`: revision loop never re-fetches the revised artifact
| Part | Detail |
|------|--------|
| **Kya** | `reviewArtifact` reuses the same `artifact` value every round; it never re-reads the new `code_artifact_produced` posted by `runProducerRevision`, so rounds 2–3 review stale content and can never converge. |
| **Kyo** | `cross_verifier.go:213-214` (same `artifact` each round) + `:333-335` (posts a new artifact, not refreshed). Wastes review LLM calls; guaranteed escalation. |
| **Change** | Reviewers judge the current revision; the loop can reach consensus; fewer wasted calls. |
| **Design (§3.1)** | **P3/P10:** review must operate on latest state; deterministic re-fetch between rounds. |
| **Knowledge** | Go, multi-agent review protocol, blackboard. |
| **Files/Risk/Rollback** | `internal/workflow/cross_verifier.go` only. Risk: Medium. Rollback: 1 hunk. |
| **Limitation** | `go build ./...`; manual: force a `changes_requested` round and observe a fresh artifact reviewed. |

### A16 — `runner.go`: `TransitionPhase` return values ignored
| Part | Detail |
|------|--------|
| **Kya** | Every `TransitionPhase` call discards its error (`_, _ =`). On DB error / invalid transition, the runner advances while `workflows.current_phase` does not → phase drift vs `runner_state`. |
| **Kyo** | `runner.go:346/373/430/449/469`. Resume and UI read the persisted phase; drift breaks both. |
| **Change** | A failed phase transition fails the workflow loudly instead of silently diverging. |
| **Design (§3.1)** | **P3/P6:** state-machine writes must be honest; a failed transition means stop, not proceed. |
| **Knowledge** | Go, state machine, crash recovery. |
| **Files/Risk/Rollback** | `internal/workflow/runner.go` only. Risk: Low–Medium (new failure exit). Rollback: 1 hunk. |
| **Limitation** | `go build ./...`; manual: simulate a DB error mid-transition (hard) — at minimum code-read verification. |

### A17 — `runner.go`: watcher vs main `Run` concurrent `executeWaves` (single-flight)
| Part | Detail |
|------|--------|
| **Kya** | Even while `running`, the change-request watcher can call `executeWaves` concurrently with the main `Run` goroutine on the same workflow (shared `runner_state`, double work). A12 fixed the *paused* case; this is the *running* case. |
| **Kyo** | `runner.go:183` starts the watcher alongside `Run`; both can drive phases. |
| **Change** | One writer per workflow at a time: a change request waits for the current wave/gate to finish before redesign starts. |
| **Design (§3.1)** | **P2/P10:** orchestration is deterministic and single-writer; LLM never decides concurrency. |
| **Knowledge** | Go concurrency, System Design (single-writer). |
| **Files/Risk/Rollback** | `internal/workflow/runner.go` (+ maybe a small per-workflow mutex type). Risk: Medium. Rollback: revert hunks. |
| **Limitation** | `go build ./...`; manual: submit a CR mid-wave and confirm serialization. |

### A18 — `runner.go`: watcher launched with `waves=nil` (always one flat redesign wave)
| Part | Detail |
|------|--------|
| **Kya** | The watcher is started before planning with `waves=nil`, so redesign always falls back to a single flat wave that ignores task dependencies. |
| **Kyo** | `runner.go:183` passes `nil`; fallback at `:1352-1363` builds one wave from all relevant experts. |
| **Change** | Redesign waves reflect the real DAG (ordering/concurrency preserved). |
| **Design (§3.1)** | **P2/P10:** reuse the deterministic planner output; do not invent a second scheduling path. |
| **Knowledge** | Go, DAG scheduling, System Design. |
| **Files/Risk/Rollback** | `internal/workflow/runner.go` only. Risk: Medium. Rollback: hunks. |
| **Limitation** | `go build ./...`; manual: submit a CR and inspect `change_request_started` wave shape. |

### A19 — Cross-expert escalation: deadlock / cycle guard (max-hop + default action)
| Part | Detail |
|------|--------|
| **Kya** | Expert-to-expert conflict escalation (A5) and reviewer↔producer loops (A7/A15) can ping-pong: a reviewer requests changes, the producer "fixes", the reviewer escalates again — or two experts each defer to the other. There is no bounded hop counter or deadlock detector, so a live workflow can loop until the revision cap, burning cost without progress. |
| **Kyo** | `cross_verifier.go` loops to `MaxRevisionRounds` with no cycle detection; `tool_loop.go` escalation has no hop counter; multi-agent deadlock is a known failure (Learnings: "deadlocking agents" — cross-owner prompts create infinite loops). |
| **Change** | A conflicting pair settles after N hops: a deterministic default action fires (escalate to client with the two positions) instead of looping. |
| **Design (§3.1)** | **P2/P3:** cycle detection is deterministic (delegation-graph / hop counter), not an LLM decision; a supervisor/default action resolves the deadlock; fail-closed escalation to the client. |
| **Knowledge** | Go, multi-agent systems, state machine. |
| **Files/Risk/Rollback** | `internal/workflow/cross_verifier.go` + `tool_loop.go`. Risk: Medium. Rollback: revert hunks. |
| **Limitation** | `go build ./...`; manual: force a contested artifact and confirm it settles by hop N (no infinite loop). |

---

## 5. Phase B — Intelligence (B1 → B9)

> **B1 (LLM synthesis) and B8 (verification-first) are the locked endpoint names from the original backlog.**
> B2–B7 + B9 are defined here as the ordered path between/around them; each is refined into its own card at task time under §2.

### B1 — Real LLM synthesis of multi-expert answers
| Part | Detail |
|------|--------|
| **Kya** | Replace the heuristic `synthesize()` (concatenation + "multiple experts agree") with an LLM synthesis that merges agreeing experts and surfaces disagreements as contrast, not as a warning dump. |
| **Kyo** | `orchestrator.go:285-289` + `synthesize` block builds pseudo-agreements/contradictions without reading the actual answer text. |
| **Change** | Multi-expert answers read as one coherent, attributed synthesis with real disagreements called out. |
| **Design (§3.1)** | **P1/P2:** one dedicated synthesis/aggregation call (≈ `1 + n + 1`), structured output (agreements[], disagreements[], synthesis); **P11:** temperature 0–0.1. **"Treat all verdicts as final"** in the synthesis prompt (else majority is ignored). Keep a **non-LLM merge fallback** (RRF-style) if the synthesis call fails. Consider a **divergent model** for the synthesis vs the experts. |
| **Knowledge** | AI (prompting, structured output), Go, System Design (cost/latency budget). |
| **Files/Risk/Rollback** | `internal/orchestrator/orchestrator.go` (+ gateway call). Risk: Medium (new LLM cost + latency). Rollback: 1 function. |
| **Limitation** | `go build ./...`; manual: 2-expert question must show a real merged answer. |

### B2 — Semantic contradiction detection & resolution policy
| Part | Detail |
|------|--------|
| **Kya** | Detect genuine contradictions between experts (not just "expert A vs warning") and attach a resolution policy (block / flag / ask client). |
| **Kyo** | Current `Contradictions` are synthesized from warnings (`orchestrator.go:544-554`) — not semantic. |
| **Change** | Real conflicting advice is flagged precisely; resolvable conflicts are noted, unresolvable ones escalate. |
| **Design (§3.1)** | Distinguish **Type-1** (contradicts provided context — catchable) from **Type-2** (fabrication — hard). Use a second/divergent LLM consistency check; disagreements across multi-pass signal genuine ambiguity. **P3:** unresolvable → escalate (feeds A5/A19). |
| **Knowledge** | AI, Multi-agent, Go. |
| **Files/Risk/Rollback** | `internal/orchestrator/*`, maybe `decision/*`. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: seed two experts that conflict and inspect the result JSON. |
| **DONE (B2 + B2b)** | `Contradiction` gained `Type`/`Resolution` (normalize* fail-closed: unknown type→fabrication, unknown resolution→escalate; fabrication never downgraded to noted). `SynthesisResult` gained `Escalations` (the escalate subset), `EscalationSummary`, `NeedsEscalation` — derived once in `applyEscalations`, used by both `synthesize` and `synthesizeFallback`. Chat has no approval_requests table (workflow-only, G1); the client acts by replying, same as a Gate 1 ASK. |

### B3 — Workflow memory write-back (complete A6)
| Part | Detail |
|------|--------|
| **Kya** | Record workflow expert decisions into project memory (L1/L2/L3) so `[PROJECT MEMORY]` (added in A6) is populated by workflows, not only by standalone chat. |
| **Kyo** | Only `orchestrator.updateMemory` calls `RecordTurn`; workflow path never does → A6 injection is near-empty for workflow-only projects. `RecordTurn` takes `chatID uuid.UUID` (value) — a `*uuid.UUID` change is needed because `master_event_log.chat_id` is nullable. |
| **Change** | Later waves/phases and future workflows see earlier settled decisions. |
| **Design (§3.1)** | **P4:** persist **only decisions that change future behaviour** (not every message); tag each memory with workflow/expert/phase ids for traceability; **add / upsert / replace** semantics (superseded decision → upsert, not append); extraction runs **asynchronously** (off the hot loop), batched, cheap model. **P5:** keep the persisted form compact. |
| **Knowledge** | Go, DB (nullable FK, add/upsert), AI (memory tiers). |
| **Files/Risk/Rollback** | `internal/workflow/agent_loop.go` or `runner.go`; `internal/memory/manager.go` (signature), `internal/orchestrator/orchestrator.go` (caller). Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: run a workflow then query L2 memory for the project. |

### B4 — Gate threshold calibration from feedback
| Part | Detail |
|------|--------|
| **Kya** | Use ratings/feedback (already recorded via `RecordRating`) to tune Gate 1–4 thresholds per domain instead of fixed constants. |
| **Kyo** | Gate decisions use static numbers; rating signal exists but is unused for tuning. |
| **Change** | Domains that are over-refusing loosen; domains that over-answer tighten. |
| **Design (§3.1)** | **P7:** any tuning must be validated by an eval set before/after (do not tune blind); **P8:** thresholds live in config, not code. |
| **Knowledge** | AI, System Design, DB. |
| **Files/Risk/Rollback** | `internal/decision/*`, `internal/rating/*`, config. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: seed ratings and observe threshold shift. |

### B5 — Question pre-processing / self-learning hardening
| Part | Detail |
|------|--------|
| **Kya** | Improve story-wrapped question extraction (DSA/algorithm) robustness and its verification step before RAG. |
| **Kyo** | Self-learning exists (`orchestrator.go:418-441`) but verification/fallback behavior is coarse. |
| **Change** | Fewer wrong-chunk retrievals on story-wrapped problems. |
| **Design (§3.1)** | **P7:** component-wise eval for extraction before chaining into retrieval. |
| **Knowledge** | AI, Go. |
| **Files/Risk/Rollback** | `internal/*` self-learning package. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: DSA story problem → correct chunks. |

### B6 — Answer quality scoring & regeneration
| Part | Detail |
|------|--------|
| **Kya** | Score an answer against the question + citations; regenerate when below a floor (bounded attempts, cost-capped). |
| **Kyo** | China Wall already retries on citation failure but not on quality/coverage soft signals. |
| **Change** | Fewer low-quality answers reach the client. |
| **Design (§3.1)** | **P7/P1:** an **LLM-as-judge rubric** (accuracy / coverage / structure / overall) with a reference/golden answer where possible; regenerate until score ≥ benchmark, capped (P10). Prefer a **divergent judge model**. Final verdict is a top-level filter, not a mechanical average. |
| **Knowledge** | AI, China Wall, Go. |
| **Files/Risk/Rollback** | `internal/chinawall/*`, `internal/decision/*`. Risk: Medium–High (cost). Rollback: revert. |
| **Limitation** | `go build ./...`; manual: force a weak answer and observe one regeneration. |

### B7 — Project memory consolidation (L2 → summaries → L3)
| Part | Detail |
|------|--------|
| **Kya** | Periodically consolidate L2 entries into project-level summaries and prune superseded decisions, so memory quality scales with project age. |
| **Kyo** | Memory grows unbounded; semantic search returns stale/superseded decisions. |
| **Change** | Long-lived projects keep a compact, current decision set. |
| **Design (§3.1)** | **P4/P5:** consolidate at session/phase end, not per-event; **prescriptive summarization** (name exactly what to keep: decisions, numbers, constraints, open items); track SNR; apply **decay** to preferences (not facts) and never decay critical decisions. External-storage + re-retrieve beats blind eviction when the item may be needed later. |
| **Knowledge** | DB (pgvector), AI (summarization), System Design. |
| **Files/Risk/Rollback** | `internal/memory/*` + a job. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: inspect summary row after consolidation. |

### B8 — Verification-first answers (claim → evidence)
| Part | Detail |
|------|--------|
| **Kya** | Require each material claim to map to a citation/evidence unit; refuse or mark unverifiable claims explicitly rather than embedding them silently. |
| **Kyo** | Citations are attached at the end; individual claims are not verified against chunks. |
| **Change** | Every assertion is traceable; unverifiable statements are labelled, not hidden. |
| **Design (§3.1)** | Follow the atomic-claim pattern: extract claims → verify each with a verdict (`supported / refuted / unverifiable`) + confidence + justification + **span anchors**; synthesize. **P9:** a claim with no span is "unverified". **P3:** default to unverifiable on uncertainty. Self-consistency/majority may aggregate evidence. |
| **Knowledge** | AI, China Wall, System Design (trust). |
| **Files/Risk/Rollback** | `internal/chinawall/*`, `internal/decision/*`. Risk: High (core answer path). Rollback: feature flag + revert. |
| **Limitation** | `go build ./...`; manual: answer with a fabricated claim must be marked/refused. |

### B9 — Context-budget policy (bound history; eviction vs summarization vs external)
| Part | Detail |
|------|--------|
| **Kya** | Multi-turn chat and multi-step workflows currently grow context unbounded. Add an explicit context policy: bound turns, and choose eviction / summarization / external-storage per data type. |
| **Kyo** | Session history is prepended without a hard budget; long threads risk context overflow / lost-in-the-middle / instruction drift. (Learnings: context is a budget; system prompt must live separately so eviction can't delete it.) |
| **Change** | Long conversations/workflows stay within budget, keep the system prompt, and retain the facts that matter. |
| **Design (§3.1)** | **P5:** system prompt stored in a separate slot; hard budget per turn; prescriptive summarization for old turns; externalize verbose items and re-retrieve when needed; "if unsure whether an old step is needed later, don't evict". Token counting approximated (no extra network call). |
| **Knowledge** | Go, AI (context engineering), DB. |
| **Files/Risk/Rollback** | `internal/message/*`, `internal/chat/*`, `internal/workflow/*` (context assembly). Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: long-thread test stays in budget and still answers with prior facts. |

---

## 6. Phase C — Moat (C1 → C10)

> **C1 (provenance) and C10 (reliability-as-product) are the locked endpoint names.**
> C2–C9 are defined here as the ordered path; refined per task at task time.

### C1 — Provenance for every answer and artifact
| Part | Detail |
|------|--------|
| **Kya** | Every answer/artifact carries a signed, queryable provenance chain: source transcript/chunk → expert → gates fired → model → timestamp. |
| **Kyo** | Citations exist but are not a verifiable chain stored durably per output. |
| **Change** | Any output can be audited end-to-end ("where did this come from?"). |
| **Design (§3.1)** | **P9:** reuse the span anchors from B8 as the provenance primitive; provenance doubles as feedback (stale refs signal knowledge drift → C6). |
| **Knowledge** | DB, System Design, crypto (signing), Go. |
| **Files/Risk/Rollback** | `internal/chat/*`, `internal/workflow/*`, `internal/memory/*`, migrations. Risk: High (schema + core paths). Rollback: additive migration + flag. |
| **Limitation** | `go build ./...` + migration run; manual: fetch provenance for one answer. |

### C2 — Expert versioning & capability drift detection
| Part | Detail |
|------|--------|
| **Kya** | Version expert charters/corpora; detect when an expert's answers drift from its declared capability. |
| **Kyo** | No versioning; a corpus re-ingest silently changes behavior. |
| **Change** | Admin sees drift and can pin/rollback an expert version. |
| **Design (§3.1)** | **P8:** also version prompts/models; a corpus/model change should trigger a re-eval (C3) before promotion. |
| **Knowledge** | DB, AI (eval), System Design. |
| **Files/Risk/Rollback** | `internal/training/*`, `internal/admin/*`, migrations. Risk: Medium–High. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: re-ingest, observe a drift flag. |

### C3 — Evaluation harness + golden set regression
| Part | Detail |
|------|--------|
| **Kya** | A golden Q/A + workflow set that runs in CI to catch regressions in answer quality, gates, and cost. |
| **Kyo** | No automated quality gate; every change is manually traced. |
| **Change** | Every PR shows a quality/cost delta; regressions block merge. |
| **Design (§3.1)** | **P7:** golden set (user-like + adversarial) + scorer; run on every PR/model/prompt change; store baseline scores; block on **vital** failures (tier must-have vs good-to-have); eval each component before chaining. Placement ladder: inline asserts → async post-workflow → canary/one-box → shadow → continuous simulation. Tools: LangFuse / DeepEval-style. |
| **Knowledge** | AI (eval), Go (test harness), CI. |
| **Files/Risk/Rollback** | new `internal/eval/*` + CI config. Risk: Low (additive). Rollback: disable job. |
| **Limitation** | Needs API keys/quota in CI; user decides budget. |

### C4 — Tenant isolation & enterprise controls
| Part | Detail |
|------|--------|
| **Kya** | Hard tenant boundaries (data, memory, cost, experts) for enterprise customers. |
| **Kyo** | Single-tenant assumptions in queries/memory. |
| **Change** | Customers can be isolated; enterprise deals become possible. |
| **Design (§3.1)** | Per-tenant vector/graph filtering is an open problem — pick a deterministic filter (metadata pre-filter) over "hope retrieval separates tenants". **P3:** fail closed if tenant scope is unknown. |
| **Knowledge** | DB (RLS), System Design, Go. |
| **Files/Risk/Rollback** | broad + migrations. Risk: High. Rollback: staged + flag. |
| **Limitation** | `go build ./...` + migration; manual isolation test. |

### C5 — Cost & usage analytics product
| Part | Detail |
|------|--------|
| **Kya** | Expose per-project/per-expert cost and usage with budgets and alerts as a first-class surface. |
| **Kyo** | CostMonitor caps exist but are not a product surface. |
| **Change** | Admins control and forecast spend; usage is visible. |
| **Design (§3.1)** | Centralized gateway attribution (already the rule, G5); group by agent/model/use-case; tie budgets to evals so cost tradeoffs are explicit (P7). |
| **Knowledge** | DB, Go, System Design. |
| **Files/Risk/Rollback** | `internal/monitoring/*`, `internal/admin/*`, frontend. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: budget alert fires. |

### C6 — Knowledge freshness / staleness detection
| Part | Detail |
|------|--------|
| **Kya** | Detect stale chunks/experts (age, contradiction with newer sources) and surface refresh tasks. |
| **Kyo** | Knowledge grows; nothing tells you it is outdated. |
| **Change** | Experts stay current; stale answers are flagged. |
| **Design (§3.1)** | Invalidation via chunk→document inverted index; model/embedding change ⇒ re-embed; provenance feedback (P9) flags refs that no longer verify. |
| **Knowledge** | AI, DB, System Design. |
| **Files/Risk/Rollback** | `internal/training/*`, `internal/knowledge/*`. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: seed an old chunk, observe flag. |

### C7 — Adversarial review / debate protocol
| Part | Detail |
|------|--------|
| **Kya** | Extend cross-verification into a bounded adversarial debate for high-stakes artifacts (propose → attack → defend → verdict), with fail-closed verdicts (builds on A7). |
| **Kyo** | Cross-verification is single-pass review; no adversarial pressure. |
| **Change** | High-stakes design/code gets stress-tested before handoff. |
| **Design (§3.1)** | **Critic–refiner** where correctness > speed (cost accepted, bounded); include a red-team catalog (prompt injection, token smuggling, boundary probing, system-prompt extraction, tool exploitation); **P2/P3:** bound hops (A19), fail closed. |
| **Knowledge** | Multi-agent, AI, Go. |
| **Files/Risk/Rollback** | `internal/workflow/cross_verifier.go` + new rounds. Risk: Medium–High (cost). Rollback: flag. |
| **Limitation** | `go build ./...`; manual: force a contested artifact. |

### C8 — Bring-your-own-expert / connector knowledge
| Part | Detail |
|------|--------|
| **Kya** | Let a customer register their own expert (corpus + charter) and/or connect external knowledge sources safely (entitlement + isolation aware). |
| **Kyo** | Experts are admin-provisioned only. |
| **Change** | Customers extend the system with their own domain knowledge. |
| **Design (§3.1)** | Connector layer centralizes OAuth/refresh/health (Learnings: `/invoke` per tool, central auth); **P3:** least-privilege + fail closed on unknown scope. |
| **Knowledge** | System Design, Go, DB, connectors. |
| **Files/Risk/Rollback** | `internal/training/*`, `internal/admin/*`, connectors. Risk: High. Rollback: staged + flag. |
| **Limitation** | `go build ./...`; manual: register + query a custom expert. |

### C9 — Explainability surface ("why this answer")
| Part | Detail |
|------|--------|
| **Kya** | UI/API that shows which gates fired, which chunks were used, which experts contributed, and why a refusal happened. |
| **Kyo** | Reasoning traces exist partially but not as a unified explanation. |
| **Change** | Users trust and can debug answers; support burden drops. |
| **Design (§3.1)** | Surface the provenance chain (C1) + gate outcomes + the judge score (B6) — explanation is a view over stored facts, not a new LLM narration. |
| **Knowledge** | Frontend, Go, System Design. |
| **Files/Risk/Rollback** | `frontend/*`, `internal/decision/*`, `internal/workflow/*`. Risk: Medium. Rollback: revert. |
| **Limitation** | `npm run build` + manual; user runs frontend toolchain. |

### C10 — Reliability-as-product (SLOs, status, audit-grade logs)
| Part | Detail |
|------|--------|
| **Kya** | Turn the reliability work into a product: published SLOs, health/status surface, audit-grade event logs, and known-failure transparency. |
| **Kyo** | Reliability is internal; customers cannot see guarantees. |
| **Change** | Customers buy a reliability contract, not just features. |
| **Design (§3.1)** | **P8:** ≥2 provider fallback + prompt/model registry; **P2:** deterministic orchestration with checkpoint/resume; circuit breakers + retries with jitter + rate-limit budget; observability (traces/cost/latency) feeding evals (P7). "Reliability of the system is our responsibility, not the provider's." |
| **Knowledge** | System Design, observability, Go, DB. |
| **Files/Risk/Rollback** | `internal/observability/*`, `internal/monitoring/*`, frontend, docs. Risk: Medium. Rollback: revert. |
| **Limitation** | `go build ./...`; manual: status endpoint + SLO doc review. |

---

## 7. Status table (update on every merge)

| Task | Title (short) | Phase | Status | Commit/PR |
|------|---------------|-------|--------|-----------|
| A1 | single-expert SSE hang | A | ✅ done | PR #5 / `7268332` |
| A2 | real messages.id + MAX+1 turn | A | ✅ done | `7268332` |
| A3 | CR service wiring | A | ✅ done | `7268332` |
| A4 | crash-resume (runner_state) | A | ✅ done | `7268332` |
| A5 | conflict escalation approval row | A | ✅ done | `7268332` |
| A6 | `[PROJECT MEMORY]` injection | A | ✅ done | `7268332` |
| A7 | fail-closed cross-verification | A | ✅ done | `7268332` |
| A8 | no send-on-closed-channel panic | A | ✅ done | `7268332` |
| A9 | runner_state snapshot race | A | ✅ done | `7268332` |
| A10 | projector error logging | A | ✅ done | `7268332` |
| A11 | QA proposals → real approval rows | A | ✅ done | `7268332` |
| A12 | watcher running-only guard | A | ✅ done | `7268332` |
| A13 | gate1WithLLM defeats Gate1Skip | A | ✅ done | (pending PR, this branch) |
| A14 | zero-citation check ordering | A | ✅ done | (pending PR, this branch) |
| A15 | revision loop re-fetch artifact | A | ✅ done | (pending PR, this branch) |
| A16 | TransitionPhase errors ignored | A | ✅ done | (pending PR, this branch) |
| A17 | watcher/main single-flight | A | ✅ done | (pending PR, this branch) |
| A18 | watcher waves=nil | A | ✅ done | (pending PR, this branch) |
| A19 | escalation deadlock/cycle guard | A | ✅ done | (pending PR, this branch) |
| B1 | LLM synthesis | B | ✅ done | (pending PR, this branch) |
| B2 | semantic contradiction detection | B | ✅ done | (pending PR, this branch) |
| B3 | workflow memory write-back | B | ⬜ pending | — |
| B4 | gate threshold calibration | B | ⬜ pending | — |
| B5 | self-learning hardening | B | ⬜ pending | — |
| B6 | answer quality regeneration | B | ⬜ pending | — |
| B7 | memory consolidation | B | ⬜ pending | — |
| B8 | verification-first answers | B | ⬜ pending | — |
| B9 | context-budget policy | B | ⬜ pending | — |
| C1 | provenance chain | C | ⬜ pending | — |
| C2 | expert versioning/drift | C | ⬜ pending | — |
| C3 | eval harness + golden set | C | ⬜ pending | — |
| C4 | tenant isolation | C | ⬜ pending | — |
| C5 | cost/usage product | C | ⬜ pending | — |
| C6 | knowledge freshness | C | ⬜ pending | — |
| C7 | adversarial debate | C | ⬜ pending | — |
| C8 | bring-your-own-expert | C | ⬜ pending | — |
| C9 | explainability surface | C | ⬜ pending | — |
| C10 | reliability-as-product | C | ⬜ pending | — |

---

## 8. Delivery mechanics (current)

- Working copy: `C:\Users\sharm\GensparkCode\AI_AVENGERS` (mirror of the temp repo).
- History merged to `origin/main`: A1–A12 (`7268332` via PR #5), roadmap doc `1daa371` via PR #6.
- Learnings commits (`925d7b9`, `bf20863`) are on fork branch `reliability-a1-a12` **only** — need a PR to reach `origin/main`.
- Direct push to `origin/main` (`kgour624/AI_AVENGERS`) is **denied** for the working account `mehrasneha161-ai` (403).
- Delivery path: push branch to fork `mehrasneha161-ai/AI_AVENGERS` → open PR → `origin/main`.
  - PR URL: `https://github.com/kgour624/AI_AVENGERS/compare/main...mehrasneha161-ai:AI_AVENGERS:reliability-a1-a12?expand=1`
- No local Go/Node/gh toolchain assumed. CI (or the user) runs `go build ./...` and the per-task manual test.

## 9. Cross-references for implementing agents

| Need | File |
|------|------|
| As-built architecture | `docs/CURRENT_ARCHITECTURE.md` |
| Service split / stages | `docs/TARGET_MICROSERVICES_ARCHITECTURE.md` |
| Workflow design model | `docs/COLLABORATIVE_DESIGN_ARCHITECTURE.md` |
| Bug list + anti-patterns | `KNOWLEDGE_HUB.md` §6.3 |
| Contracts-first rule | `docs/INTERFACE_FIRST_CONTRACT.md` |
| Distilled course learnings | `Learnings/README.md` (+ `Learnings/*.md`) |
| Current work memory | `memory: project_ai_avengers_backlog.md` |
