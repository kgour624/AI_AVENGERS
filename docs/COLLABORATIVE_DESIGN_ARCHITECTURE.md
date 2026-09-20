# Collaborative Design Architecture

**Status:** Design — not yet implemented
**Created:** 2026-09-19
**Scope:** How the AI Avengers workflow produces one authoritative design artifact
that an external coding agent (Claude Code) can build from, and how a client
converses with the experts who wrote it.

---

## 0. Read this first

Two documents already in `docs/` are relevant and one of them lies:

- `docs/CURRENT_ARCHITECTURE.md` — written from the code. Trust it.
- `docs/AIDER_INTEGRATION_ARCHITECTURE.md` — carries a banner: its Python
  samples were written against an Aider API that does not exist. Read it for
  intent only.

This document is written against the code as of commit `31becac`. Every claim
about what exists names the file. Where something does not exist, it says so.

---

## 1. What changes, in one table

| | Today | After this change |
|---|---|---|
| Who writes code | Our own Aider loop (`aider_runner.go`) | An external agent (Claude Code), outside this system |
| What we ship to the client | Design artifacts scattered across blackboard events | One **harness folder**: design spine + section docs + agent rules + acceptance criteria |
| How experts collaborate | Each expert posts its own artifacts; nobody reads the others' | Each expert reads the shared spine, then writes its own section and updates the spine |
| Design source of truth | Ambiguous — DB events, and a `design/` copy in each expert workspace | Git repository of the harness folder. Blackboard records *decisions about* it |
| Client feedback loop | Approve / Request changes at one gate | A chat on any deliverable, with the experts who wrote it, that can amend the design |
| Aider's job | Generate the application | Edit the design documents. Same tool, different and much smaller task |

**Aider is not removed.** Its role changes. See §9 — the reasons are concrete,
not sentimental.

---

## 2. The one idea this rests on

Separate the **control plane** from the **content plane**.

- **Control plane — `blackboard_events` (exists, migration 006).**
  Append-only. Who decided what, who asked whom, which gate passed, which
  contradiction was resolved and how. Never the document body.

- **Content plane — a git repository (exists, `/workspaces/{workflow_id}`).**
  The design documents themselves. Git already gives history, attribution per
  commit, diffs, rollback, and a merge story. We do not need to reinvent any of
  that inside Postgres.

The link between them: every commit that changes the design is recorded as a
blackboard event carrying the commit SHA. From an event you can reach the exact
document state; from a document line you can reach the decision that produced it.

### Why not put the document in the database

We nearly did. Rejected because:

- Rendering a 7-section document from event rows means re-implementing merge,
  conflict resolution, and diff — git does all three already and is right there.
- `workspace_merger.go` and `files_sse.go` already operate on that git workspace.
  Reusing them costs nothing; a DB-rendered document would strand both.

### Why not let seven experts free-edit one file

Also rejected. Concurrent writers on one markdown file gives lost updates and no
attribution. Instead every write is **serialised through one Aider session per
expert turn**, and each session ends in exactly one commit. Serial writes, git
attribution, no locking protocol to invent.

---

## 3. The artifact: a harness folder, not a single file

The original plan was one `final.md`. That breaks at the size this will reach:
seven experts writing full sections lands well past a hundred thousand tokens.
Claude Code starts complaining around fifty thousand in a single loaded file, and
long before that the model stops attending to the middle of it. Expert number
seven would also have to read six other sections to write its own.

So `final.md` becomes a **spine**, and the detail moves out beside it.

```
harness/
  final.md                 <- the spine. Read by everyone, every time.
  CLAUDE.md                <- rules for the coding agent
  ACCEPTANCE.md            <- how "done" is decided
  DECISIONS.md             <- decision log, newest first
  design/
    00-requirements.md     <- Product Manager
    10-architecture.md     <- System Design
    20-data-model.md       <- Database
    30-api-contract.yaml   <- System Design + Backend, jointly
    40-lld.md              <- LLD
    50-backend.md          <- Go / Java
    60-frontend.md         <- React
    65-styling.md          <- CSS
    70-testing.md          <- Testing
```

### What belongs in `final.md` and what does not

`final.md` holds only what **every** expert and the coding agent must agree on.
It stays short on purpose — a few hundred lines, not a few thousand.

Belongs in the spine:

- One-paragraph statement of what is being built
- The module map: which module exists, which expert owns it, which section file
  describes it
- Cross-cutting contracts: the API contract path, the data model path, naming
  conventions, folder layout, error-handling convention, auth approach
- Hard constraints: language and framework versions, what is explicitly out of scope
- A pointer to `ACCEPTANCE.md` and `DECISIONS.md`

Does **not** belong in the spine: any expert's reasoning, any code sample longer
than a few lines, anything only one expert needs. That goes in its section file.

This is the index-not-dump shape. The coding agent reads the spine always and
pulls one section file when it works on that module.

### `ACCEPTANCE.md` is the part that gives us control

Stated plainly: we do not control Claude Code. We never will. Control does not
come from the model, it comes from the harness around it — and the piece of the
harness that creates control is a checklist the agent can check itself against.

A design document alone tells an agent *what to build*. It gives the agent no way
to know when it is finished, so the agent decides that for itself.

Every section contributes acceptance criteria in a fixed, checkable shape:

```markdown
## AC-BACKEND-03  (owner: Go expert, section: design/50-backend.md)
Statement: POST /boards/{id}/tasks creates a task and broadcasts task.created
           to every subscriber of that board.
Verify:    go test ./internal/board -run TestCreateTaskBroadcast
Done when: the test exists, passes, and asserts at least two subscribers receive it.
```

Three properties matter:

- **Checkable without a human.** A command that exits zero or non-zero.
- **Owned.** If it fails, exactly one expert is accountable.
- **Traceable.** It names the section it came from.

This is the hard lesson from our own Aider phase, and it is worth stating because
we paid for it: the loop ran with `taskComplete = (buildErr == nil && testErr ==
nil)` in a container that had no Go toolchain. That condition could never be
true, so no task ever completed — and nobody noticed for days, because there was
no definition of done that a human had written down and could check. Fixed in
`internal/workflow/verify.go`. The same mistake at the design level would be
worse, because the client is the one who discovers it.

Anyone can write a requirements document. Shipping a design *with* a machine-
checkable definition of done is the thing that is hard to copy.

---

## 4. How the experts build it

The existing machinery does most of this. `planner.go` decomposes the
requirement into one task per expert. `dag.go` orders them by dependency.
`runner.go` executes wave by wave. `agent_loop.go` runs a design-phase expert
with Gates 1/2/3 (`gate_system.go`) and the generic-knowledge dial
(migration 017). None of that changes.

What changes is what an expert's turn *produces*.

### An expert's turn

```
1. READ    spine (final.md) + the section files this expert depends on,
           declared in the DAG. Not every section — only the declared ones.
2. THINK   agent_loop as today: training chunks via RAG, peers via Gate 2,
           generic knowledge only up to the client's dial.
3. WRITE   one Aider session, given:
             - editable:  this expert's own section file, ACCEPTANCE.md, final.md
             - read-only: the dependency sections
           Aider commits once. Commit message names the expert and the task.
4. RECORD  post design_section_written to the blackboard with the commit SHA,
           the section path, and the acceptance criteria IDs added.
```

Step 3 is the one worth defending. Why Aider rather than having the expert
return markdown that we write to disk ourselves?

- Aider edits **in place with diffs**. An expert amending its own section in a
  later round does not rewrite the file and lose the rest.
- Editing the spine is a surgical change to a file owned by everyone. A
  search/replace edit against the current content is exactly the operation Aider
  is built for; regenerating the whole spine is not.
- We get one commit per turn for free: attribution, diff, revert.
- Writing markdown is far inside Aider's competence. Generating a distributed
  application was not. Same tool, a task it can actually do.

### The dependency order

The DAG already exists; the design of the *chain* is a product decision:

```
Product Manager ─┐
                 ├─> System Design ─> LLD ─┬─> Backend ─┐
Requirements ────┘                         ├─> Frontend ─┼─> Testing
                                           └─> Database ─┘
```

Backend, Frontend and Database are one wave: they read the same LLD and touch
different section files, so they can run in parallel with no write conflict.
They all touch the spine, so **spine edits are serialised** — see §5.

### The spine write rule

Two experts editing `final.md` in the same wave is the one real conflict. Rule:

> An expert may edit its **own** section file freely. It may only **append** to
> the spine's module map, and it may not edit another expert's spine row.

After a wave completes, one **integration turn** runs — an existing expert with
the System Design role, not a new kind of agent — which reads every section file
written in that wave and rewrites the spine's contract areas coherently. That
turn's commit is the only one allowed to restructure the spine.

Mechanically this reuses `workspace_merger.go`: experts in a wave work in their
own workspaces, the merge brings them together, and the integration turn runs on
the merged `main/`.

---

## 5. Contradictions

Today, two disagreeing experts both post their view and nothing resolves it. In a
single shared document that is worse than untidy: a document containing two
opposing statements makes the coding agent pick one at random, and it will pick
differently on different runs.

`orchestrator.go` has a `Contradiction` type and a `synthesize()` that fills it —
but read `synthesize()` before relying on it. It currently pairs any WARN
response with any ADVISE response and labels the topic `"approach"` with the
literal position `"Proceed with implementation"`. It detects nothing; it is a
placeholder with a real type around it. Treat contradiction detection as **not
built**.

### The protocol

Every spine statement carries an owner and a status:

```markdown
| ID | Statement | Owner | Status | Decided in |
|----|-----------|-------|--------|------------|
| C-014 | Board state is authoritative in Postgres; Redis is cache only | System Design | accepted | DEC-007 |
| C-015 | ~~Redis is authoritative, Postgres is async~~ | Backend | superseded by C-014 | DEC-007 |
```

- An expert that disagrees does **not** edit the other expert's statement. It
  posts `design_conflict_raised` naming the statement ID and its own position.
- A raised conflict **blocks** the section it touches from being marked complete.
- Resolution is a client gate, reusing `AskClient` and the approval machinery in
  `tools.go` — the same mechanism as the design approval gate, with a different
  `gate_name`.
- On resolution an integration turn edits the spine: the winning statement stays
  `accepted`, the losing one becomes `superseded by <id>` and is struck through,
  and `DECISIONS.md` gets an entry with both positions and the reason.

The losing statement is **struck through and marked superseded, not deleted**.
Deleting it loses the reason, and six weeks later somebody re-proposes it. But it
must be unmistakably dead on the page, because an agent reading the spine must
not be able to act on it.

---

## 6. Deliverable-scoped chat

The client selects a deliverable in the Deliverables panel
(`frontend/src/pages/workflows/KanbanPage.tsx` → `ArtifactsPanel`) and talks to
the expert or experts who produced it.

### Most of this already exists

This is the important finding of this design pass, so it is stated before any new
work is proposed.

| Capability | Where it lives | State |
|---|---|---|
| Chats and messages, expert-attributed | `chats`, `messages` (migration 001); `internal/chat/service.go` | Exists |
| Streaming send pipeline | `internal/message/handler.go` `Send` | Exists |
| Context assembly: rolling summary, recent turns, semantic history, training chunks, project memory, reply threads | `internal/context/assembler.go` `Assemble` | Exists |
| Multi-expert fan-out in parallel, per-expert semaphore | `internal/orchestrator/orchestrator.go` `Process` | Exists |
| Gates 1–4, structure-permission, charter WARN | `internal/decision/engine.go`, `internal/chinawall` | Exists |
| Citations, confidence, cost per message | `messages` columns | Exists |
| Adding experts to a conversation | `project_experts` (migration 001) | Exists |
| Contradiction detection across expert answers | `orchestrator.synthesize` | **Placeholder — see §5** |
| Chat scoped to a workflow deliverable | — | **New** |
| Deliverable content as a context source | — | **New** |
| Writing an agreed change back into the design | — | **New** |

So this is not "build a chat system". It is three additions to a chat system that
already works.

### Addition 1 — scope a chat to a deliverable

Migration `018`:

```sql
ALTER TABLE chats ADD COLUMN workflow_id     UUID REFERENCES workflows(id) ON DELETE CASCADE;
ALTER TABLE chats ADD COLUMN pinned_event_id UUID REFERENCES blackboard_events(id);

CREATE INDEX idx_chats_workflow ON chats(workflow_id) WHERE workflow_id IS NOT NULL;
```

Both nullable. A nullable column is deliberate, not laziness: every existing
project chat has neither, and must keep working untouched. `workflow_id IS NULL`
means "ordinary project chat", exactly as today.

`pinned_event_id` is the deliverable. From it we reach `posted_by_expert_id` —
which is how we know who to bring into the conversation, with no new table.

### Addition 2 — the deliverable as context

`Assemble` gains one more source, alongside the existing ones:

```go
// PinnedDeliverable is the blackboard artifact this chat is about, plus the
// artifacts it cites via references_event_ids. nil for an ordinary project
// chat, which is every chat that exists today.
PinnedDeliverable *DeliverableContext
```

Budgeted like every other source. It goes **first** in `FormatForPrompt`: it is
the subject of the conversation, and the one thing that must not be the part the
model skims.

Answers are therefore grounded in **deliverable + that expert's training**, which
is the requirement.

### Addition 3 — participants

The chat's expert set starts as the deliverable's author. The client can add
more, and every added expert answers from its own training — `orchestrator.Process`
already fans out and already has the semaphore to keep that from stampeding.

**Cost control, because this is where the bill grows.** Three experts on a chat
means three RAG queries and three completions per message. Default: only the
deliverable's **author** answers. Other participants answer when the client
addresses them explicitly, or when a cheap router (`ModelFast`) judges the
question to touch their domain. Never fan out to everyone by default.

Proof that this is a real cost and not a hypothetical: the Aider phase re-ran an
identical embedding and pgvector query on all five iterations of every task
before it was hoisted out of the loop in `31becac`. The same shape of waste
appears here multiplied by the number of participants.

---

## 7. Writing an agreed change back into the design

A chat that cannot change the design is a support channel. The value is in the
loop closing.

### The protocol

```
1. The expert proposes.     A structured amendment, not prose:
                              target file, target statement ID, old text,
                              new text, reason.
2. The client approves.     Rendered as a diff. Nothing is written on the
                            strength of a model saying it would be a good idea.
3. Aider applies.           One session, one commit, on a branch named
                            amend/{chat_id}/{n}.
4. The blackboard records.  design_amended, with the commit SHA, the chat ID,
                            the message ID and the approving user.
5. DECISIONS.md grows.      Newest first, so the next reader sees the current
                            thinking without reading history.
```

Step 2 is not negotiable. The design is the single source of truth for everything
downstream; an unapproved write to it is a silent change to the contract the code
was built against.

### Why a branch

An amendment arrives while a workflow may be mid-phase. A branch keeps the
proposal out of `main/` until it is approved, and gives us a diff to render at
step 2. Merge to `main/` on approval — which is `workspace_merger.go`'s existing
job.

### The next increment

The client's stated intent, in their words: if something is unclear later, ask
the experts, get the answer written into the design, hand the design to the
coding agent, and let it continue from there.

For that to work, the agent must be able to tell **what changed since it last
built**. `DECISIONS.md` carries that: each entry records the commit range, so the
next run reads "since DEC-011 these four statements changed" rather than
re-reading the whole design and guessing.

---

## 8. Handoff to the coding agent

The workflow's output is a git repository containing the harness folder. The
client takes it to Claude Code.

`CLAUDE.md`, generated by us, is what constrains the agent:

```markdown
# Build rules

## Source of truth
final.md is authoritative. Read it in full before anything else.
design/*.md are the detail; open one only when working on that module.
A statement marked `superseded` is dead. Do not implement it.

## Boundaries
Do not invent endpoints, tables, or config keys that design/30-api-contract.yaml
and design/20-data-model.md do not define. If something is missing, stop and say
what is missing. Do not fill the gap.

## Definition of done
ACCEPTANCE.md lists every criterion with the command that checks it.
Run them. A change is not done until its criteria pass.

## Conventions
<folder layout, naming, error handling, logging — from the spine>
```

The two rules that matter are the boundary rule and the definition of done. The
first stops invention; the second removes the agent's freedom to declare itself
finished. Everything else is convention.

---

## 9. What Aider's role becomes

Aider stays. The reasons are specific:

1. **It is the collaborative editing engine for the design.** Per-expert git
   workspaces, per-turn commits, in-place diffs, a merge path — that is
   `aider_runner.go` plus `workspace_merger.go`, already built and already
   debugged through six root causes (`d9fa1c3`, `007f88c`, `0a4d3a4`, `31becac`).
2. **Editing markdown is inside its competence.** Generating a distributed
   application was not. The failure was the size of the task, not the tool.
3. **Future optionality.** If an external coding agent turns out not to fit, the
   code-generation path is still here. Deleting it to buy a smaller diff would be
   a bad trade against that option.

What must change in how we use it:

| Concern | Change |
|---|---|
| Phase name | The Aider path is no longer "implementation". It is the **authoring** step of the design phase |
| Editable files | The expert's own section, `ACCEPTANCE.md`, and append-only rows of `final.md`. Never another expert's section |
| Verification | `verify.go` build/test checks do not apply to markdown. Authoring turns need a document check instead: spine parses, every referenced section file exists, every acceptance criterion has an owner and a command, no statement is both `accepted` and `superseded` |
| Completion | `taskComplete` for an authoring turn = the document check passes. Not `go build` |

That last row matters. `verify.go` currently reports `unavailable` when it finds
no `go.mod` or `package.json`, and `AllPassed()` deliberately treats unavailable
as not-complete. For a markdown-only workspace that is correct today and wrong
tomorrow: an authoring turn would never complete. The document check is the
missing verifier, and it must be added in the same change that switches the phase
over — not after.

---

## 10. Data model changes

Migration `018`, one file:

```sql
-- Chats can be scoped to a workflow deliverable.
ALTER TABLE chats ADD COLUMN workflow_id     UUID REFERENCES workflows(id) ON DELETE CASCADE;
ALTER TABLE chats ADD COLUMN pinned_event_id UUID REFERENCES blackboard_events(id);
CREATE INDEX idx_chats_workflow ON chats(workflow_id) WHERE workflow_id IS NOT NULL;

-- Which experts participate in a chat, beyond the deliverable's author.
CREATE TABLE chat_participants (
    chat_id    UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    expert_id  UUID NOT NULL REFERENCES experts(id),
    added_by   UUID NOT NULL REFERENCES users(id),
    added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, expert_id)
);
```

`blackboard_events.event_type` is `VARCHAR(100)` with **no CHECK constraint**
(migration 006, verified). New event types need no migration:

| Event type | Posted by | Content |
|---|---|---|
| `design_section_written` | expert | `{section_path, commit_sha, acceptance_ids[]}` |
| `design_conflict_raised` | expert | `{statement_id, my_position, reason}` |
| `design_conflict_resolved` | client | `{statement_id, winner, superseded[], reason}` |
| `design_amendment_proposed` | expert | `{chat_id, message_id, target, old, new, reason}` |
| `design_amended` | client | `{chat_id, commit_sha, statement_ids[]}` |
| `harness_published` | system | `{commit_sha, spine_path, acceptance_count}` |

`approval_requests.gate_name` **does** have a CHECK constraint (migration 006):
`'intake','high_level_design','detailed_design','handoff','budget_exceeded','ad_hoc'`.
Conflict resolution and amendment approval must either use `ad_hoc` or add values
in migration 018. Recommended: add `design_conflict` and `design_amendment`, so
the Kanban can label them meaningfully. We have already been bitten twice by this
exact constraint (`a698d0b`, and the `ingestion_jobs_stage_check` violation) —
check the constraint before inventing a value, every time.

---

## 11. Mental execution — a full trace

Requirement: *"Real-time Kanban board. Go backend, React frontend, WebSocket
updates."* Experts: PM, System Design, LLD, Go, React, Testing. Generic dial 0%.

```
planner.go        6 tasks, one per expert
dag.go            waves: [PM] [SysDesign] [LLD] [Go, React] [Testing]

wave 0  PM
  read    final.md (does not exist -> created from template)
  write   design/00-requirements.md
          final.md module map: + row "requirements | PM | design/00-requirements.md"
          ACCEPTANCE.md: + AC-REQ-01..03
  commit  a1b2c3  "design(pm): capture requirements"
  event   design_section_written {design/00-requirements.md, a1b2c3, [AC-REQ-01..03]}

wave 1  System Design
  read    final.md + design/00-requirements.md   (read-only)
  write   design/10-architecture.md, design/30-api-contract.yaml
          final.md: contracts section now points at 30-api-contract.yaml
          C-014 "Board state authoritative in Postgres; Redis cache only"
  commit  d4e5f6
  event   design_section_written

wave 2  LLD
  read    final.md + 00, 10, 30
  write   design/40-lld.md   (classes, SOLID boundaries, error taxonomy)
  commit  g7h8i9

wave 3  Go and React in PARALLEL, separate workspaces
  Go      read  final.md + 10, 30, 40
          write design/50-backend.md
                proposes C-015 "Redis authoritative, Postgres async"
                -> CONFLICTS with C-014
          event design_conflict_raised {C-014, ...}
          NOTE  section stays incomplete; the conflict blocks it
  React   read  final.md + 30, 40
          write design/60-frontend.md, ACCEPTANCE.md + AC-FE-01..04
          commit j1k2l3

  workspace_merger.MergeWave -> main/
  conflict open -> AskClient(gate_name=design_conflict)
  client picks C-014
  event   design_conflict_resolved {winner C-014, superseded [C-015]}

  integration turn (System Design on merged main/)
          final.md: C-014 accepted; C-015 struck through, "superseded by C-014"
          DECISIONS.md: DEC-007, both positions, reason
          design/50-backend.md rewritten to match C-014
  commit  m4n5o6

wave 4  Testing
  read    final.md + every section
  write   design/70-testing.md, ACCEPTANCE.md + AC-TEST-01..06
  commit  p7q8r9

document check   spine parses; 9 sections referenced, 9 exist;
                 17 acceptance criteria, each with owner + command;
                 0 statements both accepted and superseded   -> PASS
event            harness_published {p7q8r9, final.md, 17}
client           downloads harness/ -> Claude Code -> code
```

### Now the chat, on the same workflow

```
client selects deliverable "design/30-api-contract.yaml"  (event E-112, author: System Design)
POST /workflows/{id}/deliverables/E-112/chat
  -> chats row {workflow_id, pinned_event_id=E-112}
  -> chat_participants: System Design

client: "Why is there no bulk-move endpoint? Dragging 10 cards will be 10 calls."

Assemble: PinnedDeliverable = E-112 + cited events
          CourseChunks      = System Design training, top-k for this question
          RecentMessages, RollingSummary as today
decision.Engine gates 1-4 -> ADVISE
answer grounded in the contract + training, with citations

client adds the Go expert
Go expert answers from ITS training: "batching changes the WebSocket fan-out;
  one event per card or one batch event — that is a contract decision"

System Design proposes an amendment:
  target  design/30-api-contract.yaml
  new     POST /boards/{id}/tasks:batch-move  +  task.moved.batch event
  reason  10x fewer round trips on drag; matches C-014

client approves the rendered diff
Aider applies on branch amend/{chat_id}/1 -> commit s1t2u3 -> merged to main/
event  design_amended {chat_id, s1t2u3, [C-022]}
DECISIONS.md: DEC-008, newest first

next Claude Code run: reads DECISIONS.md, sees "since DEC-007: C-022 added",
  implements only the delta
```

---

## 12. Cross-verification against the requirement

| The client asked for | Where this design delivers it | Honest gap |
|---|---|---|
| All domain experts collaborate on one design | §4. Shared spine, per-expert sections, declared dependency order | The chain in §4 is a first cut. The team must confirm the real SDLC order |
| One file, single source of truth | §3. Spine is authoritative; sections are its detail | Not literally one file. Reason in §3 — one file breaks at this size |
| Each expert reads what the others wrote | §4 step 1. Dependency sections are read-only inputs | Only *declared* dependencies, not everything. Deliberate, to control context |
| Download the design and give it to Claude Code | §8. Git repo with `CLAUDE.md` + `ACCEPTANCE.md` | Export/packaging endpoint is new work |
| Design strong enough that any builder produces a strong result | §3 `ACCEPTANCE.md` + §8 boundary rule | A design cannot make a weak agent strong. Acceptance criteria are what make weakness *visible*, which is the achievable goal |
| Chat per deliverable with its author | §6. `pinned_event_id` -> `posted_by_expert_id` | New: scoping, deliverable context, participants |
| Add more experts to a conversation | §6 Addition 3 + `chat_participants` | Router to avoid paying for every expert on every message is new |
| Answers from deliverable + training | §6 Addition 2 | New context source; the RAG half exists |
| As intelligent as Claude/Kiro chat | §6. Gates, citations, memory, threads, multi-expert — all built | Honest: "as intelligent as" is not a testable target. Suggest replacing it with specific behaviours before building |
| Chat can write the change into the design | §7. Proposal -> approval -> commit -> event | All new, and the approval step must not be skipped |
| Incremental: next run continues from the updated design | §7 + `DECISIONS.md` commit ranges | New |
| Keep Aider | §9. Role changes, path stays | The document check in §9 must ship with the switch, or authoring turns never complete |

---

## 13. Failure modes we have already paid for

Each of these is a bug we hit in the last week. They are listed because this
design can reproduce every one of them.

| What happened | Where | Guard in this design |
|---|---|---|
| `taskComplete` could never be true — no Go toolchain in the image | `31becac` | §9: authoring turns need a *document* check. Never leave a phase with no reachable definition of done |
| Design artifacts overwrote each other — event type mapped to a fixed filename, 11 artifacts became 4 files | `1faf3e5` | §3: one file per section, owned by one expert. Never derive a filename from a type |
| Design docs sat in the workspace unopened; the model was asked to implement what it could not see | `1faf3e5` | §4 step 1: dependency sections are passed as explicit read-only inputs |
| A nil slice became JSON `null` and the receiver rejected every request; the caller never checked the status code, so the reason was discarded | `0a4d3a4` | Every new endpoint checks status before decoding, and logs the body on failure |
| Rules said "your training is the ONLY source of truth" while zero chunks were retrieved, so the model correctly wrote nothing | `1faf3e5` | §6: the deliverable is a named source alongside training. An expert with thin training still has the approved design |
| `gate_name` CHECK constraint rejected a value we invented | `a698d0b` | §10: constraint checked and named before use |
| A local migration fix was assumed missing because the clone was stale | this session | Fetch before concluding anything about remote state |

---

## 14. Not in scope

- Generating the application ourselves. That is Claude Code's job now.
- A visual editor for the design. Chat plus approved diffs is the editing surface.
- Multi-tenant concurrent editing of one workflow's design by several humans.
- Replacing `agent_loop`, the gate system, or the generic-knowledge dial. All
  reused unchanged.
- Auto-resolving contradictions. A contradiction is a product decision and goes
  to the client (§5).

---

## 15. Open questions for the team

1. **Expert order.** §4's chain is a guess at your SDLC. Confirm it, including
   where CSS and Testing sit, and whether Database precedes or follows LLD.
2. **Who writes acceptance criteria?** Every expert for its own section, or the
   Testing expert for all of them? Per-expert gives ownership; Testing gives
   consistency. This design assumes per-expert with a Testing review pass.
3. **Two experts trained on the same domain.** Their training will disagree
   sometimes. Is that a conflict for the client (§5), or does seniority decide?
4. **How does the built code report back?** Right now the loop ends at handoff.
   If Claude Code's output should feed the next design increment, that path needs
   designing — it is not in this document.
5. **`is_training` experts.** Two of the three experts in recent runs retrieved
   zero chunks. An expert with no training contributing to the single source of
   truth is a risk. Should an untrained expert be blocked from authoring?
6. **Harness export.** Zip download, a git remote we push to, or a generated
   repository? Affects §8 and the client's workflow.
