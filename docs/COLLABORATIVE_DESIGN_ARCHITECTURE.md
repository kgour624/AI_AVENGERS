# Collaborative Design Architecture

**Status:** Design — not yet implemented
**Created:** 2026-09-19
**Revised:** 2026-09-19 — removed every hardcoded expert/section list; workflow
chat separated from the product chat; tool loop added.
**Scope:** How the workflow produces one authoritative design artifact that an
external coding agent (Claude Code) builds from, and how a client converses with
the experts who wrote it.

---

## 0. Read this first

Two documents already in `docs/` are relevant and one of them lies:

- `docs/CURRENT_ARCHITECTURE.md` — written from the code. Trust it.
- `docs/AIDER_INTEGRATION_ARCHITECTURE.md` — carries a banner: its Python
  samples were written against an Aider API that does not exist. Read it for
  intent only.

This document is written against the code at commit `31becac`. Every claim about
what exists names the file. Where something does not exist, it says so.

---

## 1. What changes, in one table

| | Today | After this change |
|---|---|---|
| Who writes code | Our own Aider loop (`aider_runner.go`) | An external agent (Claude Code), outside this system |
| What we ship | Design artifacts scattered across blackboard events | One **harness folder**: design spine + section docs + agent rules + acceptance criteria |
| How experts collaborate | Each expert posts its own artifacts; nobody reads the others' | Each expert reads the shared spine, then writes its own section and updates the spine |
| Design source of truth | Ambiguous — DB events, plus a `design/` copy per expert workspace | Git repository of the harness folder. Blackboard records *decisions about* it |
| Client feedback | Approve / Request changes at one gate | A chat on any deliverable, with the experts who wrote it, that can amend the design |
| Chat capability | Answers questions | Answers **and acts** — reads the design, searches it, proposes amendments (§7) |
| Aider's job | Generate the application | Author the design documents. Same tool, a much smaller task |

**Aider is not removed.** Its role changes. See §9.

---

## 2. Two principles

### 2.1 Control plane and content plane are separate

- **Control plane — `blackboard_events`** (exists, migration 006). Append-only.
  Who decided what, who asked whom, which gate passed, which contradiction was
  resolved and how. Never the document body.
- **Content plane — a git repository** (exists, `/workspaces/{workflow_id}`). The
  design documents themselves. Git already gives history, per-commit attribution,
  diffs, rollback and a merge path.

Link: every commit that changes the design is recorded as a blackboard event
carrying the commit SHA. From an event you reach the exact document state; from a
document line you reach the decision that produced it.

**Why not store the document in Postgres.** Rendering a multi-section document
from event rows means re-implementing merge, conflict resolution and diff — git
does all three and is already mounted. A DB-rendered document would also strand
`workspace_merger.go` and `files_sse.go`, both of which already operate on that
workspace.

**Why not let N experts free-edit one file.** Concurrent writers on one markdown
file give lost updates and no attribution. Every write is serialised through one
Aider session per expert turn, each ending in exactly one commit.

### 2.2 Static is protocol. Dynamic is domain.

This is the rule that decides every question of the form "is this a fixed list?"

**May be fixed in code** — things that are properties of the *system* and do not
change when the expert roster changes:

- The event type names (§10)
- The names of the four protocol files: `final.md`, `CLAUDE.md`,
  `ACCEPTANCE.md`, `DECISIONS.md`
- The shape of an acceptance criterion (id, owner, statement, verify command)
- The shape of the spine's module-map table

**Must be dynamic — never a constant, never a `switch` on a domain name:**

- Which experts participate, how many, and in what order
- How many waves there are
- Section file names and their numbering
- Which sections exist at all
- Which tools an expert may call
- Every threshold and ceiling (§11)

A concrete consequence: there is **no** `switch expert.Domain { case "react": ... }`
anywhere in this design. Adding a security expert tomorrow requires a database
row, not a code change. Removing one requires nothing.

---

## 3. The artifact: a harness folder

### 3.1 Layout

Four protocol files, plus one section file per participating expert:

```
harness/
  final.md          <- the spine. Read by everyone, every time. Short on purpose.
  CLAUDE.md         <- rules for the coding agent
  ACCEPTANCE.md     <- how "done" is decided
  DECISIONS.md      <- decision log, newest first
  design/
    <one file per participating expert, named and numbered at run time — §3.2>
```

`final.md` is a **spine**, not the whole design. The original plan was a single
file. That breaks at the size this reaches: N experts writing full sections passes
a hundred thousand tokens, Claude Code starts complaining around fifty thousand in
one loaded file, and long before that the model stops attending to the middle.
Expert number N would also have to read N−1 other sections to write its own.

### 3.2 Section files are assigned, not hardcoded

**Rule.** The first time an expert authors in a workflow, it is assigned a section
number and path. That assignment never changes afterwards.

```sql
CREATE TABLE workflow_design_sections (
    workflow_id  UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    expert_id    UUID NOT NULL REFERENCES experts(id),
    section_no   INTEGER NOT NULL,
    section_path TEXT    NOT NULL,
    assigned_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workflow_id, expert_id),
    UNIQUE (workflow_id, section_no),
    UNIQUE (workflow_id, section_path)
);
```

- `section_no` = `(count of existing rows for this workflow + 1) * 10`
- `section_path` = `design/{section_no}-{experts.slug}.md`
- `experts.slug` is already `UNIQUE NOT NULL` (migration 001), so paths cannot
  collide and no new naming scheme is invented.

Worked examples — the same code, three different rosters:

| Roster | Produced paths |
|---|---|
| 3 experts: product-manager, unlimited-go, simplified-react | `design/10-product-manager.md`, `design/20-unlimited-go.md`, `design/30-simplified-react.md` |
| 8 experts (same three + system-design, lld, database, css, testing) | `design/10-…` … `design/80-…`, in first-author order |
| A security expert added to a **running** workflow after five have authored | `design/60-security-expert.md` — the five existing files are not renamed |

**Why assign-once instead of deriving the number from the wave index.** If the
number came from the wave, adding an expert later would renumber and rename files
already committed. That is git churn, broken links in the spine, and broken
references in `DECISIONS.md` — all for cosmetic ordering. The gap in numbering
after a removal is not a problem; a renamed file is.

Sort order for humans falls out of the numeric prefix. Semantic order (who reads
whom) lives in the spine's module map and in the DAG, not in the filename.

### 3.3 What goes in the spine

The spine holds only what **every** expert and the coding agent must agree on. It
stays short on purpose — hundreds of lines, not thousands.

Belongs there:

- One paragraph: what is being built
- **Module map** — generated from `workflow_design_sections` joined to `experts`,
  never typed by hand:

  ```markdown
  | Module | Owner | Section | Status |
  |--------|-------|---------|--------|
  | <from expert capability/domain> | <experts.name> | design/20-unlimited-go.md | authored |
  ```

- Cross-cutting contracts: where the API contract lives, where the data model
  lives, naming conventions, folder layout, error handling, auth approach
- Hard constraints: language and framework versions, explicit out-of-scope
- Statement table with owner and status (§5)
- Pointers to `ACCEPTANCE.md` and `DECISIONS.md`

Does **not** belong there: any expert's reasoning, code samples beyond a few
lines, anything only one expert needs. That goes in its section file.

This is the index-not-dump shape. The coding agent always reads the spine and
pulls one section when it works on that module.

### 3.4 `ACCEPTANCE.md` is what gives us control

Stated plainly: we do not control Claude Code and never will. Control comes from
the harness, not the model — and the part of the harness that creates it is a
checklist the agent can check itself against.

A design says *what to build*. It gives the agent no way to know it has finished,
so the agent decides for itself.

Each criterion has a fixed shape (protocol — §2.2) and dynamic content:

```markdown
## AC-<SECTION_NO>-<n>   owner: <experts.name>   section: <section_path>
Statement: <what must be true>
Verify:    <a command that exits zero or non-zero>
Done when: <the observable condition>
```

Three properties matter: **checkable without a human**, **owned by exactly one
expert**, **traceable to a section**.

This is the lesson from our own Aider phase, and it is worth repeating because we
paid for it: the loop ran with `taskComplete = (buildErr == nil && testErr == nil)`
inside a container that had no Go toolchain. That condition could never be true,
so no task ever completed, and nobody noticed for days — because no definition of
done had been written down where a human could check it. Fixed in
`internal/workflow/verify.go`. The same mistake at design level is worse: the
client discovers it.

Anyone can write a requirements document. Shipping a design *with* a
machine-checkable definition of done is the part that is hard to copy.

---

## 4. How the experts build it

The existing machinery does most of this and **already is dynamic**:

- `planner.go` — decomposes the requirement into one task per selected expert,
  whatever the roster. It also produces `depends_on_expert_names` per task.
- `dag.go` — topological sort. **Wave count is an output, not a setting.**
- `runner.go` — executes wave by wave, parallel within a wave.
- `agent_loop.go` — runs one design-phase expert with Gates 1/2/3
  (`gate_system.go`) and the client's generic-knowledge dial (migration 017).

None of that changes. What changes is what an expert's turn *produces*.

### 4.1 An expert's turn

```
1. READ    final.md (spine) + the section files of this task's declared
           dependencies. Only declared ones — not everything, so context
           stays bounded as the roster grows.
2. THINK   agent_loop as today: own training via RAG, peers via Gate 2,
           generic knowledge only up to the client's ceiling.
3. WRITE   one Aider session:
             editable  : this expert's own section file, ACCEPTANCE.md,
                         and append-only rows of final.md
             read-only : the dependency section files
           Aider commits once. Commit message names expert and task.
4. RECORD  post design_section_written with commit SHA, section path and the
           acceptance criterion IDs added.
```

Step 3 uses Aider rather than us writing files directly because Aider edits **in
place with diffs** (an expert amending its own section later does not rewrite and
lose the rest), because a surgical edit to a file owned by everyone is exactly
what search/replace is for, and because one commit per turn gives attribution,
diff and revert for free.

### 4.2 Order is decided at run time

There is **no hardcoded chain** in this design. Order comes from
`planner.depends_on_expert_names` → `dag.BuildDAG` → waves. Three experts may
produce two waves; eight may produce five. The planner decides; the DAG validates.

Two guards, both of which already exist and must be kept:

- Cycle detection (`dag.go` — this is the bug fixed in `ce625e0`; the in-degree
  calculation was reversed and every normal two-task plan reported a false cycle).
- Self-dependency rejection (`planner.go`, fixed in `fff440e` — the LLM returned a
  UUID with different casing and a string compare missed it).

If the team wants to *influence* order without touching code, the hint belongs in
the database, fed into the planner prompt:

```sql
ALTER TABLE expert_categories ADD COLUMN authoring_rank INTEGER;
```

`NULL` means no opinion, which is the default and changes nothing. It is a
**hint** in the prompt, not an override of the DAG — otherwise we would have two
competing sources of order, which is the contradiction problem in §5 applied to
ourselves.

### 4.3 The spine write rule

Two experts editing `final.md` in the same wave is the one real conflict.

> An expert may edit its **own** section file freely. It may only **append** to
> the spine's module map and statement table, and may not edit another expert's
> row.

After a wave completes, one **integration turn** runs. It is not a new kind of
agent: it is whichever participating expert the workflow designates as integrator
(selectable, default = the expert with the fewest declared dependencies, i.e. the
most upstream one — computed, not named in code). It reads every section written
in that wave and rewrites the spine's contract areas coherently. That turn's
commit is the only one allowed to restructure the spine.

Mechanically this reuses `workspace_merger.MergeWave(ctx, workflowWorkspace,
expertIDs)`: experts in a wave work in their own workspaces, the merge brings them
together, and the integration turn runs on the merged `main/`.

---

## 5. Contradictions

Today two disagreeing experts both post their view and nothing resolves it. In one
shared document that is worse than untidy: a document containing two opposing
statements makes the coding agent pick one at random, and differently on
different runs.

`orchestrator.go` has a `Contradiction` type and a `synthesize()` that fills it —
but read `synthesize()` before relying on it. It pairs any WARN response with any
ADVISE response and labels the topic `"approach"` with the literal position
`"Proceed with implementation"`. It detects nothing. **Treat contradiction
detection as not built.**

### The protocol

Every spine statement carries an owner and a status:

```markdown
| ID | Statement | Owner | Status | Decided in |
|----|-----------|-------|--------|------------|
| C-014 | Board state authoritative in Postgres; Redis cache only | <expert> | accepted | DEC-007 |
| C-015 | ~~Redis authoritative, Postgres async~~ | <expert> | superseded by C-014 | DEC-007 |
```

- A disagreeing expert does **not** edit the other's statement. It posts
  `design_conflict_raised` naming the statement ID and its own position.
- A raised conflict **blocks** the section it touches from completing.
- Resolution is a client gate, reusing `tools.AskClient` and the approval
  machinery — same mechanism as the design gate, different `gate_name` (§10).
- On resolution the integration turn edits the spine: winner stays `accepted`,
  loser becomes `superseded by <id>` and is struck through, and `DECISIONS.md`
  gains an entry with both positions and the reason.

The loser is **struck through and marked superseded, not deleted.** Deleting loses
the reason and someone re-proposes it in six weeks. But it must be unmistakably
dead on the page, because an agent reading the spine must not act on it.

---

## 6. Workflow chat — deliberately separate from the product chat

### 6.1 The separation is a hard boundary

There are now **two chat systems** and they do not share code paths:

| | Product chat (exists) | Workflow chat (new) |
|---|---|---|
| Purpose | A client talking to experts about a project | A client talking to experts about a **deliverable** inside a workflow |
| Tables | `chats`, `messages` | `workflow_chats`, `workflow_chat_messages` |
| Packages | `internal/chat`, `internal/message`, `internal/orchestrator` | `internal/workflow/chat` (new) |
| Gate implementation | `internal/decision/engine.go` (gates 1–4, structure permission) | `internal/workflow/gate_system.go` — the same gates the design phase uses |
| Can call tools | No | **Yes** (§7) |
| Can change the design | No | Yes, through approval (§7.5) |

**Why separate rather than adding `workflow_id` to `chats`.** This was the earlier
proposal and it is withdrawn. The product chat is shipped and paid for. Threading
workflow concerns through `message/handler.go` and `orchestrator.go` puts every
future workflow-chat change one bug away from breaking it. Duplicating the
pipeline shape costs us some repeated code; coupling costs us the product. The
duplication is the cheaper mistake.

Copy the *shape* of `message.Handler.Send` — assemble, gate, call, save, index,
stream. Reuse only leaves that carry no policy:

- `context.Assembler` — data access for history and training chunks
- `gateway.ModelGateway` — the LLM call
- `chinawall` — formatting and charter text
- `workflow.GateSystem`, `workflow.Tools` — already workflow-scoped

Do **not** reuse `decision.Engine`, `orchestrator.Orchestrator`, `chat.Service`,
or `message.Handler`. Those belong to the product chat.

Using `GateSystem` rather than `decision.Engine` is deliberate: the answers a
client gets in a workflow chat must obey the same knowledge rules as the design
that workflow produced. Two gate implementations giving different answers about
the same deliverable is a defect the client would find first.

### 6.2 Tables

```sql
CREATE TABLE workflow_chats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    client_id       UUID NOT NULL REFERENCES users(id),
    -- The deliverable this chat is about. NULL = a chat about the workflow as a
    -- whole rather than one artifact.
    pinned_event_id UUID REFERENCES blackboard_events(id),
    title           VARCHAR(500) NOT NULL,
    -- Knowledge mode, §6.4. NULL = inherit workflows.generic_allowance_pct.
    generic_allowance_pct NUMERIC(5,2),
    message_count   INTEGER NOT NULL DEFAULT 0,
    is_archived     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workflow_chat_participants (
    chat_id   UUID NOT NULL REFERENCES workflow_chats(id) ON DELETE CASCADE,
    expert_id UUID NOT NULL REFERENCES experts(id),
    added_by  UUID NOT NULL REFERENCES users(id),
    added_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chat_id, expert_id)
);

CREATE TABLE workflow_chat_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES workflow_chats(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL,   -- user | assistant | tool
    expert_id       UUID REFERENCES experts(id),
    content         TEXT NOT NULL,
    turn_number     INTEGER NOT NULL,
    -- Tool loop, §7. NULL for plain messages.
    tool_name       VARCHAR(100),
    tool_input      JSONB,
    tool_result     JSONB,
    step_number     INTEGER,
    citations       JSONB,
    gate_result     JSONB,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    cost_usd        NUMERIC(12,8) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

The product chat's tables are not touched. Zero regression surface.

### 6.3 Scoping and participants

`pinned_event_id` is the deliverable. From it we reach `posted_by_expert_id` — the
author — so no extra table is needed to know who to bring in. The client may add
more participants; each answers from its own training.

**Cost control.** Three participants means three retrievals and three completions
per message. Default: **only the deliverable's author answers.** Others answer
when addressed explicitly, or when a cheap router (`ModelFast`) judges the
question to touch their domain. Never fan out to everyone by default.

This is not hypothetical. The Aider phase re-ran an identical embedding and
pgvector query on all five iterations of every task until it was hoisted out of
the loop in `31becac`. The same waste appears here multiplied by participants.

### 6.4 Knowledge mode switch

Every workflow chat surface carries a visible switch, and it applies everywhere a
workflow chat exists:

| Mode | Meaning | Effect |
|---|---|---|
| **Trained only** | `generic_allowance_pct = 0` | Answers come from the expert's training and the approved design. If neither covers it, the expert says what is missing rather than filling the gap. |
| **Allow generic up to N%** | `generic_allowance_pct = N` | Generic knowledge permitted for gaps only, up to the ceiling. Every generic claim is labelled as such in the answer. |

- Stored per chat, nullable, **inheriting `workflows.generic_allowance_pct`** when
  unset — so a chat opened inside a strict workflow is strict by default.
- Enforced by the same `GateSystem.RunGates` the design phase uses, which already
  takes `genericAllowancePct` and returns it on `GateResult`. No second
  implementation of the rule.
- The ceiling is **not** 30 in code. See §11 — it moves to `system_settings`.
- The switch is also an audit record: `gate_result` on each message stores which
  mode produced that answer, so a later reader knows whether a statement came from
  training or from generic fill-in.

---

## 7. Making the workflow chat capable, not just conversational

### 7.1 What is missing today

The product chat is genuinely intelligent at **answering**: retrieval over
`course_chunks`, the expert's `reasoning_charter`, gates, citations, confidence,
rolling summary, semantic history, reply threads, per-message cost. In one respect
it beats Claude and Kiro — those have no notion of "this expert may only speak
from its training".

What it cannot do is **act**. Claude and Kiro read files, edit files, run
commands. `message/handler.go` has no tool execution of any kind: its functions
are `Send`, `saveAssistantMessage`, `indexTurn`, `generateRollingSummary`,
`sendSSE`. Nothing else.

The capability exists elsewhere in this codebase. `workflow.Tools` has
`PostArtifact`, `AskExpert`, `ReadBlackboard`, `AskClient` — wired to
`agent_loop.go` only. And `workflowExpert.AllowedTools []string` already exists on
the planner's expert record, so per-expert tool permission is already a data
concept rather than a code one.

So the work is a tool loop in the new workflow-chat package, reusing both.

### 7.2 Tool registry — dynamic, never a hardcoded list

```go
// Tool is one capability the model may invoke. Everything about a tool is data:
// the registry is populated at startup and filtered per expert from
// AllowedTools, so adding or withdrawing a capability is a database change.
type Tool struct {
    Name        string          // "read_design"
    Description string          // shown to the model
    InputSchema json.RawMessage // JSON Schema; the model is told this
    Mutating    bool            // true => result requires client approval (§7.5)
    Handler     func(ctx context.Context, in json.RawMessage) (any, error)
}
```

Which tools a given expert may call = `registry ∩ expert.AllowedTools`. An expert
row with an empty `allowed_tools` gets read-only tools and nothing else — a safe
default, and the default for every expert that exists today.

### 7.3 The loop

This is the `gather → act → verify` loop, the same one Claude Code runs:

```
turn starts
  ├─ assemble context
  │    pinned deliverable + cited events   (first — it is the subject)
  │    expert training chunks via GateSystem.RunGates
  │    peer contributions (Gate 2)
  │    rolling summary + recent turns + semantic history
  │    the design spine + section list (so the model knows what it can open)
  │
  ├─ step 1..maxSteps:
  │    call model with (context, conversation, tool list)
  │    response has no tool call  -> emit as the answer, loop ends
  │    response has a tool call   -> validate input against InputSchema
  │                                  Mutating?  -> create a proposal, do NOT apply
  │                                  read-only? -> execute, append result, continue
  │
  └─ persist every step as a workflow_chat_messages row (role=tool, step_number)
```

Properties that matter:

- **Every step is persisted.** The client can see what the expert read before it
  answered. Claude Code's transparency, and our audit trail.
- **`maxSteps` comes from `system_settings`**, not a constant (§11). Without a cap
  a confused model loops until the budget is gone.
- **Schema validation before execution.** A model that invents an argument gets a
  validation error back as the tool result and can correct itself. It does not get
  to call a handler with garbage.
- **Cost is attributed.** `LLMRequest.WorkflowID` already exists for exactly this
  (added in `ebc9cc8`); a chat step sets it so chat spend lands on the workflow's
  budget and `CostMonitor` sees it.

### 7.4 Tool catalogue

Read-only — available to every participant:

| Tool | Purpose |
|---|---|
| `list_sections` | The module map: which sections exist, who owns them |
| `read_design(path, range?)` | Read the spine or a section file |
| `search_design(query)` | Find where something is stated across the harness |
| `read_blackboard(types?, since?)` | `Tools.ReadBlackboard` — the decision history |
| `get_acceptance(section?)` | Criteria and their verify commands |
| `read_decisions(since?)` | `DECISIONS.md`, newest first |

Mutating — each produces a **proposal**, never a direct write:

| Tool | Purpose |
|---|---|
| `propose_amendment(target, old, new, reason)` | Change a design statement (§7.5) |
| `propose_acceptance(section, statement, verify, done_when)` | Add a criterion. The **id and owner are assigned, not asked for** — `AC-{section_no}-{next}` from `workflow_design_sections`, and the owner token is the expert slug already in the section path. A model filling in its own id is how two `AC-20-01` entries end up in one file; and the owner must be a single token or `acceptanceLineRe` truncates it. |
| `raise_conflict(statement_id, position, reason)` | `design_conflict_raised` (§5) |
| `ask_expert(expert_id, question)` | `Tools.AskExpert` — consult a peer mid-answer |
| `ask_client(summary)` | `Tools.AskClient` — escalate to the human |

Four of these are `workflow.Tools` methods already written and already debugged —
including the foreign-key bug in `AskClient` where a zero UUID was passed as
`posted_by_expert_id` and every gate failed on the first call (`2f40b17`).

### 7.5 Write-back: propose, approve, commit (SHIPPED)

A chat that cannot change the design is a support channel. The value is the loop
closing. But nothing is written because a model said it would be a good idea.

**Implemented.** `internal/workflow/amendment.go` (flow) and `amendment_text.go`
(the deterministic apply). Endpoints:

```
GET  /api/v1/workflows/{id}/amendments               ?status=pending|approved|rejected|all
POST /api/v1/workflows/{id}/amendments/{aid}/respond {decision, edited_text, notes}
```

```
1. PROPOSE   the expert calls propose_amendment: target file, old text, new text,
             reason. One design_amendment_proposed event AND one pending
             approval_requests row (gate_name='design_amendment', which migration
             018 already allows). Nothing on disk.
2. APPROVE   the client lists pending amendments and answers approve,
             approve_with_edit or reject. Idempotent — a second click is a 409.
3. APPLY     an exact text replacement, then one commit. NOT an Aider session,
             and NOT on a branch — see below.
4. RECORD    design_amended with commit SHA, approval id, decision id, approver.
5. MERGE     not needed — nothing was branched.
6. LOG       DECISIONS.md gains an entry, newest first, with a DEC-nnn id.
```

Step 2 is not optional. The design is the source of truth for everything
downstream; an unapproved write is a silent change to the contract the code was
built against.

#### Two departures from the plan above, both deliberate

**No Aider session.** An amendment is a structured `{target, old_text, new_text}`,
so applying it is an exact string replacement. Handing that to a model buys
nothing and costs three things already paid for on the Aider path:
non-determinism (it rewrites text nobody asked about), money per amendment, and
the failure mode where the model proposes no edits and the run silently produces
nothing — fixed twice already (`1faf3e5`, `d9fa1c3`). `applyAmendmentText` has a
defined answer for every input:

| Input | Result |
|---|---|
| `old_text` empty | append (this is how a new criterion or statement row is added) |
| `old_text` appears exactly once | replace |
| `old_text` appears 0 times | **refuse** — the design moved; propose again |
| `old_text` appears more than once | **refuse**, with the count — the proposal does not say which one |

The last two are the point. A model asked to "replace this line" when the line
occurs twice will pick one; picking one is a guess at the contract the code is
built against. This is the same rule Aider's own SEARCH/REPLACE blocks follow,
which is why no model is needed to follow it.

**No branch.** The branch had two stated jobs. "Gives us the diff to render at
step 2" is already satisfied — `old_text`/`new_text` *is* the diff, stored
structurally in `artifact_content`, and it is what the approval screen renders.
"Keeps the proposal out of `main/` until approved" is satisfied by writing nothing
until approval, which is the actual behaviour. Against that, a branch adds a real
hazard: `main/` is the shared workspace a running wave rsyncs into, and a
`git checkout` there would swap files under a wave in progress. Step 5 disappears
with it — there is nothing to merge back.

#### Why not the existing approvals endpoint

`POST /workflows/{id}/approvals/{aid}/respond` already exists and an amendment is
an `approval_requests` row, so reusing it looks obvious. Three reasons it cannot
be:

1. Its body is `{decision, notes, generic_allowance_pct}` — there is nowhere to
   put edited text, so edit-then-approve could not be first-class.
2. On approve it calls `Engine.Resume`, moving the workflow to `running`.
   Approving an amendment on a finished workflow would restart it.
3. There is no list endpoint. Approval identity reaches the frontend only over
   the kanban SSE stream, which fires on `question_to_client` and renders only
   while the workflow is `paused_for_approval`. An amendment raises no such event
   and pauses nothing, so without a list a client cannot discover one exists.

#### The one limitation, named

An approved amendment is **refused while the workflow is running** (409, "the
design is being written right now"). A wave merge rsyncs each expert workspace
over `main/`, and those workspaces were seeded from `main/` *before* the
amendment — so a merge landing after an apply would silently overwrite it.
Refusing is honest; making `WorkspaceMerger` amendment-aware is the real fix and
a larger change than this.

Concurrency within one process is handled by a mutex around read-modify-write of
the harness files. That assumes one api process per workspace directory, which
holds because the workspace is a local volume. A shared-storage deployment would
need a Postgres advisory lock instead.

### 7.6 What "as intelligent as Claude/Kiro" means here, concretely

"As intelligent as" is not testable, so it is replaced by behaviours that are:

| Behaviour | How |
|---|---|
| Reads the design before answering, instead of guessing | `read_design` / `search_design` in the loop, and the spine in the assembled context |
| Shows its work | Every tool step persisted and streamed |
| Asks instead of inventing when the design is silent | Gate rules in Trained-only mode + `ask_client` |
| Consults a colleague when the question is not its domain | `ask_expert`, and the participant router (§6.3) |
| Can change things, not only describe them | Mutating tools through approval (§7.5) |
| Remembers the conversation | Rolling summary, recent turns, semantic history — already built |
| Cites sources | `citations` per message — already built |
| Stays inside its competence | `GateSystem` — Claude and Kiro have no equivalent |

The last row is the one where we are ahead, and it is the product's reason to
exist. It should not be traded away chasing parity on the others.

---

## 8. Handoff to the coding agent

The workflow's output is a git repository containing the harness folder.

`CLAUDE.md`, generated by us from the spine — every list in it comes from the
database and the harness, none of it typed by hand:

```markdown
# Build rules

## Source of truth
final.md is authoritative. Read it in full before anything else.
design/*.md are the detail; open one only when working on that module.
A statement marked `superseded` is dead. Do not implement it.

## Boundaries
Do not invent endpoints, tables or config keys that the contract and data-model
sections do not define. If something is missing, stop and say what is missing.
Do not fill the gap.

## Definition of done
ACCEPTANCE.md lists every criterion with the command that checks it.
Run them. A change is not done until its criteria pass.

## Conventions
<generated from the spine's contract section>
```

The boundary rule stops invention. The definition of done removes the agent's
freedom to declare itself finished. Everything else is convention.

---

## 9. Aider's role

Aider stays. Reasons, specific:

1. **It is the collaborative editing engine for the design.** Per-expert git
   workspaces, per-turn commits, in-place diffs, a merge path — `aider_runner.go`
   plus `workspace_merger.go`, already built and debugged through six root causes
   (`d9fa1c3`, `007f88c`, `0a4d3a4`, `31becac`).
2. **Editing markdown is inside its competence.** Generating a distributed
   application was not. The failure was the size of the task, not the tool.
3. **Optionality.** If the external-agent route does not fit, the code-generation
   path is still here. Deleting it to buy a smaller diff is a bad trade.

What must change:

| Concern | Change |
|---|---|
| Phase name | The Aider path is no longer "implementation". It is the **authoring** step of the design phase |
| Editable files | The expert's own section, `ACCEPTANCE.md`, append-only spine rows. Never another expert's section |
| Verification | `verify.go` build/test checks do not apply to markdown. Authoring needs a **document check**: spine parses; every path in `workflow_design_sections` exists on disk; every acceptance criterion has an owner and a verify command; no statement is both `accepted` and `superseded`; every `references_event_ids` resolves |
| Completion | `taskComplete` for an authoring turn = the document check passes. Not `go build` |

That last row is load-bearing. `verify.go` reports `unavailable` when it finds no
`go.mod` or `package.json`, and `AllPassed()` deliberately treats unavailable as
not-complete. For a markdown-only workspace that is correct today and wrong
tomorrow: an authoring turn would never complete. The document check must ship in
the same change that switches the phase over, not after.

---

## 10. Data model changes

One migration, `018`.

```sql
-- §3.2 section assignment
CREATE TABLE workflow_design_sections (...);            -- see §3.2

-- §6.2 workflow chat, fully separate from chats/messages
CREATE TABLE workflow_chats (...);
CREATE TABLE workflow_chat_participants (...);
CREATE TABLE workflow_chat_messages (...);

-- §4.2 optional ordering hint, NULL = no opinion
ALTER TABLE expert_categories ADD COLUMN authoring_rank INTEGER;

-- §5, §7.5 new gates. approval_requests.gate_name HAS a CHECK constraint
-- (migration 006, verified: intake, high_level_design, detailed_design,
-- handoff, budget_exceeded, ad_hoc). Extend it rather than reusing ad_hoc, so
-- the Kanban can label these meaningfully.
ALTER TABLE approval_requests DROP CONSTRAINT IF EXISTS approval_requests_gate_check;
ALTER TABLE approval_requests ADD CONSTRAINT approval_requests_gate_check
    CHECK (gate_name IN ('intake','high_level_design','detailed_design','handoff',
                         'budget_exceeded','ad_hoc',
                         'design_conflict','design_amendment'));

-- §11 thresholds and ceilings move out of code into settings.
-- system_settings already exists and ModelGateway already reads it.
```

`blackboard_events.event_type` is `VARCHAR(100)` with **no CHECK** (migration 006,
verified), so new event types need no migration:

| Event type | Posted by | Content |
|---|---|---|
| `design_section_written` | expert | `{section_path, commit_sha, acceptance_ids[]}` |
| `design_conflict_raised` | expert | `{statement_id, my_position, reason}` |
| `design_conflict_resolved` | client | `{statement_id, winner, superseded[], reason}` |
| `design_amendment_proposed` | expert | `{chat_id, message_id, target, old, new, reason}` |
| `design_amended` | client | `{chat_id, commit_sha, statement_ids[]}` |
| `harness_published` | system | `{commit_sha, section_count, acceptance_count}` |

We have been bitten by a CHECK constraint twice already (`a698d0b`, and the
`ingestion_jobs_stage_check` violation). Check the constraint before inventing a
value, every time.

---

## 11. Static values that must move to the database

Per §2.2, these are currently constants in Go and must become
`system_settings` rows, admin-editable, with the current value as the default:

| Value | Where it is now | Why it must move |
|---|---|---|
| `gate1StrongThreshold = 0.70` | `gate_system.go:27` | We already changed this once under pressure when measured scores came in at 0.504 and 0.668. A tuning knob does not belong behind a rebuild |
| `gate1UsableThreshold = 0.40` | `gate_system.go:38` | Same |
| `gate1MaxChunks` | `gate_system.go` | Same |
| Generic ceiling `30` | CHECK in migration 017 | See the note below — this one is not a straight move |
| `maxIterations = 5` | `aider_runner.go` | Cost/quality trade-off |
| `maxDesignAttempts = 10` | `runner.go` | Same |
| Coverage target `80.0` | `aider_runner.go` | Per-project policy |
| Tool-loop `maxSteps` | new (§7.3) | Same class — must be settings from day one, not a constant added and moved later |

None of this is cosmetic. Every one of these is a number someone will want to
change during a client engagement, and a rebuild-and-redeploy to change a
threshold is how a tuning problem becomes an outage.

### The generic ceiling is a special case

Migration 017 states its own reasoning in the file: the `30` is in the schema
**on purpose**, so that no code path — present or future, ours or a new
contributor's — can exceed it. That argument is correct and should not be
discarded because §2.2 says "nothing static".

The distinction is between a **safety bound** and a **tuning value**:

- The CHECK stays, as a safety bound. Widen it to `0–100`, which is the real
  limit of the quantity.
- The **business ceiling** (today 30) moves to `system_settings`, and the handler
  enforces it — as it already does, alongside the CHECK.

Result: an admin can lower the ceiling to 10 for a strict client, or raise it to
50, without a migration; and no bug can ever produce a value outside `0–100`.
Both properties, neither sacrificed.

---

## 12. Mental execution

Everything below is **one example run**, not a specification. The roster, the wave
count and the file names are all outputs.

### 12.1 Six experts

Requirement: *"Real-time Kanban board. Go backend, React frontend, WebSocket
updates."*
Selected experts: `product-manager`, `system-design`, `lld`, `unlimited-go`,
`simplified-react`, `testing`. Generic ceiling 0.

```
planner.go   6 tasks, dependencies from the LLM
dag.go       waves: [product-manager] [system-design] [lld]
                    [unlimited-go, simplified-react] [testing]      -> 5 waves

wave 0  product-manager
  assign  section_no 10 -> design/10-product-manager.md
  read    final.md (absent -> created from the protocol template)
  write   its section; spine module-map row; AC-10-01..03
  commit  a1b2c3   event design_section_written

wave 1  system-design
  assign  20 -> design/20-system-design.md
  read    spine + design/10-product-manager.md (read-only)
  write   its section; contract pointers in the spine
          statement C-014 "Board state authoritative in Postgres; Redis cache only"
  commit  d4e5f6

wave 2  lld            assign 30 -> design/30-lld.md        commit g7h8i9

wave 3  unlimited-go and simplified-react IN PARALLEL, separate workspaces
  unlimited-go     assign 40 -> design/40-unlimited-go.md
                   proposes C-015 "Redis authoritative, Postgres async"
                   -> conflicts with C-014
                   event design_conflict_raised; its section stays incomplete
  simplified-react assign 50 -> design/50-simplified-react.md
                   AC-50-01..04                              commit j1k2l3

  MergeWave -> main/
  conflict open -> AskClient(gate_name=design_conflict)
  client picks C-014
  event design_conflict_resolved {winner C-014, superseded [C-015]}

  integration turn (most-upstream participant, computed = system-design)
    spine: C-014 accepted; C-015 struck through "superseded by C-014"
    DECISIONS.md: DEC-007 with both positions and the reason
    design/40-unlimited-go.md rewritten to match C-014
  commit m4n5o6

wave 4  testing        assign 60 -> design/60-testing.md
                       AC-60-01..06                          commit p7q8r9

document check  spine parses; 6 section paths in workflow_design_sections,
                6 present on disk; 17 criteria each with owner + command;
                0 statements both accepted and superseded            -> PASS
event           harness_published {p7q8r9, sections 6, acceptance 17}
```

### 12.2 The same code with three experts

Selected: `product-manager`, `unlimited-go`, `simplified-react`.
Planner: go and react both depend on pm, and not on each other.

```
dag.go   waves: [product-manager] [unlimited-go, simplified-react]   -> 2 waves
paths    design/10-product-manager.md
         design/20-unlimited-go.md
         design/30-simplified-react.md
```

No integration turn is needed after wave 0 (single author), one after wave 1.
Nothing in code changed.

### 12.3 An expert added to a running workflow

Five experts have authored; `security-expert` is added and a design re-run is
requested (the existing `RestartPhase` + design loop from `e2fc2c2`).

```
assign   section_no = (5 existing + 1) * 10 = 60 -> design/60-security-expert.md
paths 10..50 are NOT renumbered and NOT renamed
planner re-runs; the DAG now places security-expert after system-design
the spine gains one module-map row; every existing DECISIONS reference still
  resolves, because no file moved
```

### 12.4 A deliverable chat that does work

```
client selects the contract deliverable (event E-112, author system-design)
POST /workflows/{id}/deliverables/E-112/chat
  -> workflow_chats {workflow_id, pinned_event_id=E-112, generic_allowance_pct=NULL}
  -> participants: system-design
  -> switch shows "Trained only" (inherited from the workflow's 0)

client: "Why is there no bulk-move endpoint? Dragging 10 cards is 10 calls."

tool loop
  step 1  list_sections            -> 6 sections, owners
  step 2  read_design(contract)    -> current endpoints
  step 3  search_design("batch")   -> no match anywhere
  step 4  no tool call -> answer, citing the contract and its training

client adds unlimited-go as a participant
unlimited-go answers from ITS training: batching changes the WebSocket fan-out —
  one event per card or one batch event is a contract decision, not an
  implementation detail

system-design calls propose_amendment
  target design/20-system-design.md contract block + spine statement C-022
  new    POST /boards/{id}/tasks:batch-move, event task.moved.batch
  reason 10x fewer round trips on drag; consistent with C-014
  -> design_amendment_proposed. Nothing written.

client sees the diff, approves
Aider applies on amend/{chat_id}/1 -> commit s1t2u3 -> merged to main/
event design_amended {chat_id, s1t2u3, [C-022]}
DECISIONS.md: DEC-008, newest first

next Claude Code run reads DECISIONS.md: "since DEC-007, C-022 added"
  -> implements only the delta
```

---

## 13. Cross-verification against what was asked

| Asked for | Delivered in | Honest gap |
|---|---|---|
| All domain experts collaborate on one design | §4 | — |
| Nothing fixed — 3, 6 or 8 experts all work | §3.2, §4.2, §12.1–12.3 | Verified against three rosters. The only fixed things are the four protocol filenames and the event-type names (§2.2) |
| One file, single source of truth | §3 | **Not literally one file.** Reason in §3: one file breaks past ~100k tokens and the model stops reading its middle |
| Each expert reads what the others wrote | §4.1 step 1 | Only *declared* dependencies. Deliberate — otherwise context grows with the roster |
| Download and hand to Claude Code | §8 | Export endpoint is new work (§16 Q6) |
| Strong design ⇒ strong result from any builder | §3.4 + §8 | A design cannot make a weak agent strong. Acceptance criteria make weakness **visible**, which is the achievable goal |
| Chat per deliverable with its author | §6.3 | New: scoping and deliverable context |
| Existing chat stays completely separate | §6.1 | Done — separate package, separate tables, no shared handler |
| Add more experts to a conversation | §6.3 | Router to avoid paying for every expert on every message is new |
| Answers from deliverable + training | §6.3, §7.3 | New context source; the retrieval half exists |
| Switch: RAG-only vs generic up to N% | §6.4 | New per-chat column; the rule itself is `GateSystem`, already built |
| As intelligent as Claude/Kiro | §7.6 | Replaced with eight testable behaviours. Two are already ahead of them (gates, citations); four need the tool loop |
| Chat can write into the design | §7.5 | All new. The approval step must not be skipped |
| Incremental: next run continues | §7.5 step 6 + `DECISIONS.md` ranges | New |
| Keep Aider | §9 | The document check must ship with the switch, or authoring never completes |

---

## 14. Failure modes we have already paid for

Every row is a bug from the last week. This design can reproduce all of them.

| What happened | Commit | Guard here |
|---|---|---|
| `taskComplete` could never be true — no Go toolchain in the image | `31becac` | §9: authoring needs a *document* check. Never leave a phase with no reachable definition of done |
| Design artifacts overwrote each other — event type mapped to a fixed filename, 11 artifacts became 4 files | `1faf3e5` | §3.2: one file per expert, assigned once. Never derive a filename from a type |
| Design docs sat in the workspace unopened; the model was asked to implement what it could not see | `1faf3e5` | §4.1: dependencies passed as explicit read-only inputs; §7.4 `read_design` |
| A nil slice became JSON `null`; the receiver rejected every request; the caller never checked the status code so the reason was thrown away | `0a4d3a4` | Every new endpoint checks status before decoding and logs the body on failure |
| Rules said "your training is the ONLY source of truth" while zero chunks were retrieved, so the model correctly wrote nothing | `1faf3e5` | §6.4: the approved design is a named source alongside training |
| `gate_name` CHECK rejected an invented value | `a698d0b` | §10: constraint read, then extended explicitly |
| False cycle in the DAG — in-degree reversed | `ce625e0` | §4.2: keep the guard; the roster is dynamic so this path runs for every plan |
| `AskClient` passed a zero UUID as an FK and every gate failed on the first call | `2f40b17` | §7.4 reuses the fixed `Tools` methods rather than new call sites |
| A local migration fix assumed missing because the clone was stale | this session | Fetch before concluding anything about remote state |

---

## 15. Not in scope

- Generating the application ourselves — Claude Code's job now.
- A visual editor for the design. Chat plus approved diffs is the editing surface.
- Several humans concurrently editing one workflow's design.
- Replacing `agent_loop`, `gate_system`, or the generic dial. All reused.
- Auto-resolving contradictions — a product decision, goes to the client (§5).
- Any change to the product chat (§6.1).

---

## 16. Decisions (resolved with the team, 2026-09-19)

These were open questions in the first draft. They are now decided. Each records
the choice and the reason, so nobody re-opens them without knowing why.

1. **Who writes acceptance criteria — DECIDED: each expert for its own section,
   then a testing-expert review pass.** Ownership sits with the author, who best
   knows what can break; the testing expert makes a second pass to catch gaps and
   inconsistency. This is §3.4 plus a review step at the end of the DAG.

2. **Two experts trained on the same domain disagree — DECIDED: split by scope.**
   An architecture-level disagreement (something in the spine's statement table)
   goes to the client as a conflict (§5). A small, same-domain detail is settled
   automatically by rank — `expert_categories.authoring_rank`, the same column
   §4.2 adds. The higher rank wins, silently, and the decision is logged in
   `DECISIONS.md` so it is visible without stopping the client. Rationale: the
   client must own product-shaping choices, but must not be pulled into every
   small technical difference.

   Mechanically: when `raise_conflict` (§7.4) fires, the handler checks whether
   the two statements are in the spine's **statement table** (architecture-level →
   client gate) or inside a single section's detail (same-domain → rank auto).
   The classifier is the statement's location, not an LLM judgement — location is
   deterministic; a judgement is not.

3. **Does built code report back — DECIDED: not now (Option A). The full plan for
   later (Option B) is written in §17 so it never has to be reverse-engineered
   from transcripts.** The loop ends at handoff for now. §17 is a complete,
   self-contained plan for the feedback loop, to be implemented only when the
   basic flow (design → chat → amend → handoff) is proven in production.

4. **Untrained experts — DECIDED: not a policy question, it was a data accident.**
   The zero-chunk experts in recent runs were a human error (a SQL query run
   against the training data), not experts intended to work untrained. Going
   forward no untrained expert enters a workflow — that is guaranteed by process,
   at expert-selection time, not by a block in the authoring path. One safety net
   remains: if an expert with `total_chunks = 0` ever reaches an authoring turn,
   the turn fails loudly with "expert has no training loaded" rather than
   authoring from nothing. This is a guard against the accident recurring, not a
   feature.

5. **Harness export — DECIDED: ZIP now (Option A). Git push (Option B) full plan
   in §18.** A download button that zips the harness folder ships first. §18 is
   the complete plan for pushing to a client's git remote, reusing the existing
   repo-connect/OAuth feature, to be implemented when a client actually asks for
   it.

6. **Integrator — DECIDED: automatic by default, with a manual override toggle.**
   By default the integrator is the most-upstream participant, computed as in
   §4.3 (fewest declared dependencies — typically System Design or Product
   Manager). But some workflows will not have those roles at all, so a per-workflow
   setting can (a) turn the auto choice off and (b) name a specific integrator.
   Stored on the workflow; auto is the default and needs no configuration.

7. **Tool permissions — DECIDED: per expert (Option A).** `AllowedTools` is
   already a per-expert column (`workflowExpert.AllowedTools`), so mutating-tool
   permission is granted per expert, not per category. Default remains empty =
   read-only, which is safe and is what every expert has today. An admin grants a
   mutating tool to a specific expert deliberately.

---

## 17. Option B — the code-feedback loop (SHIPPED)

**Implemented.** `internal/workflow/code_feedback.go` (orchestration) and
`code_feedback_scan.go` (the comparator). Endpoint:
`POST /api/v1/workflows/{id}/code-feedback` with `{provider, repo_url, branch}`.

The plan below is what was built, with one step deliberately left out — see
§17.7. What changed from the plan as written, and why, is in §17.6.

### 17.1 What it is

Today the workflow ends when the client downloads the harness. Option B closes
the loop: the code Claude Code produces comes back, the experts see what was
actually built, and the next design increment is informed by it — build errors,
deviations from the contract, decisions the coding agent had to make that the
design left open.

### 17.2 The shape

```
client builds with Claude Code in their own repo
        │
        ▼
1. INGEST   client points the workflow at the built repo (the repo-connect
            feature from §18 already knows how to read a repo). We do NOT
            re-run the design; we read the result.
2. DIFF     compare the built code against ACCEPTANCE.md:
              - which acceptance criteria have matching code / passing tests
              - which endpoints/tables exist that the contract never defined
                (the coding agent invented them → a gap in our design)
              - which contract items have no implementation (unbuilt)
3. REPORT   post code_feedback_ingested to the blackboard with the findings.
            This is content the experts read on the next turn — exactly like
            observeWorkspace feeds the Aider loop today (verify.go pattern).
4. AMEND    each finding becomes a proposed design amendment (§7.5) routed to
            the owning expert: "the coding agent added POST /x that the contract
            does not define — accept into the contract, or is this a defect?"
5. GATE     client approves the amendments, exactly as §7.5.
6. INCREMENT the design now reflects reality; DECISIONS.md records the round;
            the next Claude Code run reads the delta since the last decision.
```

### 17.3 What it reuses (nothing new invented)

| Need | Existing piece |
|---|---|
| Read a client repo | Repo-connect / OAuth (§18, and `internal/repo`) |
| Compare code to expectation | `verify.go` per-project build/test, run against the client repo |
| "Findings the experts read" | The observations pattern in `observeWorkspace` |
| Turn a finding into a design change | `propose_amendment` + approval (§7.5) |
| Record the round | `DECISIONS.md` + a new `code_feedback_ingested` event |

### 17.4 New pieces required

- One event type: `code_feedback_ingested` (no migration — `event_type` has no
  CHECK, §10).
- A comparator that maps acceptance criteria to code presence. This is the only
  genuinely new logic, and it is bounded: for each `AC-*` with a `Verify` command,
  run it against the client repo and record pass/fail; for each contract endpoint
  and table, grep the repo for its presence.
- An ingest endpoint: `POST /workflows/{id}/code-feedback` taking a repo
  reference.

### 17.5 The mental model, end to end

```
design v1 shipped (DEC-011 the latest decision)
client builds, comes back with repo R
INGEST R
DIFF against ACCEPTANCE.md:
   AC-20-03 (batch-move endpoint)  -> test passes            ✓
   AC-50-01 (WebSocket reconnect)  -> no matching code       ✗ unbuilt
   POST /boards/{id}/archive       -> in repo, not in contract  ← invented
REPORT code_feedback_ingested { built: [...], unbuilt: [AC-50-01],
                                invented: [POST /boards/{id}/archive] }
AMEND:
   unbuilt AC-50-01   -> routed to the frontend expert: "still required? or drop?"
   invented archive   -> routed to system-design: "fold into contract, or defect?"
client approves: keep AC-50-01, adopt archive into the contract as C-030
Aider commits both amendments -> design v2
DECISIONS.md: DEC-012 "reconciled with build R; C-030 added, AC-50-01 reaffirmed"
next build reads: since DEC-011, C-030 added
```

### 17.6 What changed while building it

Five things the plan did not say, each found by running the code rather than by
reading it.

**1. The comparator's reach is narrower than "compare code to the design", and
the boundary is now written down.** Four buckets, no more: does a criterion's own
`Verify` command pass; does a path the design names appear in the repo's source;
does a path the repo declares appear nowhere in the design; and — not attempted —
whether the implementation is *correct*. Everything found is a **question routed
to the owning expert**, never an automatic write. That is what makes a heuristic
acceptable: a false positive costs one question in the amendment queue.

**2. The harness is inside the client's repo, which almost made the whole report
meaningless.** The §18 export pushes `final.md`, `ACCEPTANCE.md`, `DECISIONS.md`,
`CLAUDE.md` and `design/` into the client's repository. A naive scan therefore
finds every designed endpoint "present in the code" — because the design document
*is* in the code — and the unbuilt list is empty forever. The scan excludes the
harness files at the repo root, excludes `design/`, and excludes **all markdown**:
the question is "was it built", and a document describing an endpoint is not an
endpoint.

**3. A `Verify:` line is model-authored text, and running it is executing that
text.** It runs through an allowlist of executables (`go`, `npm`, `npx`, `node`,
`python`, `pytest`, `make`, `grep`, `test`, `ls`, `cat`, and the package-manager
variants) with no shell: `sh`, `bash`, `curl`, `rm`, `git` are absent by
intention, and an unquoted shell operator (`|`, `&&`, `;`, `>`, backtick) makes
the criterion **unverifiable with that stated as the reason** rather than a
refusal the client cannot interpret. Quoted operators are fine —
`go test -run 'TestA|TestB'` is an ordinary command.

**4. One regexp bug that would have inverted the entire report.** The endpoint
extractor's first version did not allow a pipe between the method and the path.
A design that writes its API as a markdown table with Method and Path in separate
columns — the shape §3.3 itself uses — matched *nothing*, so every designed
endpoint would have been reported as "the code declares an endpoint the design
never defined", with the design sitting right there. Caught by running the
extractor against a table.

**5. Findings are written on their own context.** The analysis has a 20-minute
budget; the report and the amendments are written on a fresh 60-second one.
Sharing the analysis context meant that the slow, failing run — the one whose
report matters most — was the one that silently produced none.

### 17.7 What is NOT built, and why not

Step 6 of §17.2 also asks for a `DECISIONS.md` entry recording the round. That is
a **write to the harness**, and writing to the harness is §7.5 steps 3-6 (an
Aider session on an `amend/` branch, then a merge) — still unbuilt. Writing
`DECISIONS.md` from the ingest would be a second, unapproved way to change the
design, which is precisely what §7.5 step 2 exists to prevent: *"an unapproved
write is a silent change to the contract the code was built against."*

So the round is recorded on the blackboard, every amendment references the report
event that produced it, and the `DECISIONS.md` entry arrives with the approved
amendments once the apply path exists.

### 17.8 Events this adds

No migration: `blackboard_events.event_type` has no CHECK (§10).

| Event | Posted when |
|---|---|
| `code_feedback_ingested` | one per ingest, carrying the whole report |
| `code_feedback_failed` | the ingest could not finish (clone failed, timed out) |
| `code_feedback_question` | one per finding, with `to_expert_id` set to the routed expert and `references_event_ids` pointing at the report |

**Why `code_feedback_question` and not `design_amendment_proposed`.** They were
the same event type at first, and that conflated two things needing different
actions from different people. A `design_amendment_proposed` carries a concrete
`{old_text, new_text}` and a pending approval row: the **client** approves it and
it is written (§7.5). A finding here has no such text — it is a question an
**expert** must answer ("still required, or drop it?"), and the expert's answer is
what then becomes a real amendment via `propose_amendment`. One event type for
both would leave the UI unable to tell "you must approve this" from "an expert
must answer this", and would put un-approvable rows in the amendments list.

Routing: an unsatisfied criterion and a designed-but-absent contract item go to
the expert owning that **section path** (matched against `workflow_design_sections`,
never against the owner name in the markdown — `acceptanceLineRe` captures the
owner as `\S+`, so a two-word expert name arrives truncated). An
undesigned-but-built item goes to the integrator (§16.6).

---

## 18. Option B — git-push export (SHIPPED)

**Implemented.** `internal/workflow/export_git.go`, with the credential rules in
`client_remote.go` and `client_repo.go`. Endpoint:
`POST /api/v1/workflows/{id}/export/git` with
`{provider, repo_url, branch, create_repo, private}`. Synchronous — the response
carries the commit SHA.

Note on §16.5: this was planned as the *second* delivery route, after ZIP. ZIP
was never built, so this is currently the only export. That is a gap to close,
not a decision that was made.

### 18.1 What exists to build on

`internal/repo` already handles GitHub/GitLab connection: OAuth flow, token
storage (encrypted, per `ENCRYPTION_KEY`), and PAT-based connect. The workflow
already produces the harness as a git repository under `/workspaces/{workflow_id}`.
So export is mostly wiring two things that already exist.

### 18.2 The shape

```
1. CONNECT   client connects a git provider (exists: OAuth or PAT, internal/repo)
2. CHOOSE    a target: an existing empty repo, or a new repo to create
3. PUSH      the harness workspace's main/ branch is pushed to the target:
               git remote add client <url-with-token>
               git push client main
             The workspace is already a git repo with real commit history —
             per-expert authorship and every amendment — so the client receives
             not just files but the full decision trail in git log.
4. RECORD    harness_exported event { provider, repo_url, commit_sha }
```

### 18.3 New pieces required

- An export endpoint: `POST /workflows/{id}/export/git` taking a provider +
  target repo reference.
- Create-repo call (optional path): the provider APIs both support it; only
  needed for "new repo" rather than "existing empty repo".
- Token injection into the push URL, then scrubbed from logs. The token is never
  written to disk in the workspace — it goes into the remote URL for one push and
  the remote is removed after.

### 18.4 The one risk, named — and what was actually done about it

Pushing carries a token. §18.4 originally said the remote must be added and then
removed. **No remote is added at all.** `git push <url> <refspec>` takes the URL
as an argument and persists nothing; verified directly — after a push,
`git remote -v` is empty and `.git/config` contains zero `url` lines. `git clone`
is the opposite: it *does* write the URL to `remote.origin.url`, so the ingest
path removes that remote immediately after cloning. Both paths then run
`assertNoSecretOnDisk`, which is §18.4's "mental check before shipping" turned
into an assertion that runs every time — including after a *failed* push, because
a failed push can still have written something.

Two more guards that the plan did not mention and that matter more than the
config question:

**The remote host is validated before the token is ever read.** The repo URL is
client-supplied input. Without an allowlist, a request body saying
`{"repo_url": "https://evil.example/x"}` hands the client's write-scoped token to
`evil.example`. Only `github.com` and `gitlab.com` are accepted, the host is taken
from the allowlist rather than from the input, credentials embedded in the input
URL are discarded, and ssh remotes are refused with the real reason. Self-hosted
GitLab is deliberately absent: `internal/repo` hardcodes `gitlab.com` for its API
calls, so a self-hosted instance is not supported anywhere in the product.

**Every byte of git output is scrubbed before it is logged or returned.** git
echoes the push URL verbatim in several of its error messages.

Also refused: force push. A push that would overwrite existing history is
rejected by git and reported as *"this repo already has commits on that branch"*
with the action to take. A `--force` flag here would turn a convenience feature
into a way to destroy a client's repository from one API call.

### 18.5 A real bug this uncovered

§18.1 claimed the harness workspace "is already a git repository with real commit
history". **It was not.** `WorkspaceMerger` creates `main/` with `os.MkdirAll` and
rsyncs into it with `--exclude .git`; nothing ever ran `git init` there. Two live
consequences, both pre-dating this work:

- `MergeWave`'s multi-expert path calls `commitMerge`, whose first command is
  `git add .` inside `main/`. In a directory that is not a repository — and whose
  parents are not either — that command fails with
  `fatal: not a git repository`, so **every multi-expert wave merge was failing**
  at that step.
- `MergeWave`'s single-expert path never committed at all, so whether the harness
  had any history depended on how many experts happened to be in the wave.

Fixed in `ensureGitRepo`, called from both merge paths, so the merge is fixed too
rather than only the export.

One thing to not overstate: because rsync excludes `.git`, per-expert commit
authorship does **not** survive into `main/`. `main/`'s log is one commit per wave
merge. That is a real decision trail, just coarser than "every expert's own
commits" — §18.2's wording was optimistic.

Related: `git init` picks the branch name from `init.defaultBranch`, which is
unset in this image, so the local branch is whatever the installed git defaults to
(`master` today). The push reads the actual branch and maps it onto the requested
target branch. Assuming `main` would have been a push to a branch that does not
exist.

---

## 19. Still open — genuinely undecided

Only these remain for the team; everything in §16 is settled.

- **`authoring_rank` seeding.** §4.2 and §16.2 both use it. Who sets the initial
  ranks per category, and by what criteria? Until set, all ranks are NULL, which
  means: no ordering hint (planner decides order) and same-domain conflicts fall
  back to the client instead of auto-resolving. Safe default; just less automatic.
- **Testing-review depth (§16.1).** Does the testing expert only *review*
  acceptance criteria, or may it *add* its own? This design allows adding; confirm
  that is wanted.
- **Amendments are refused while a wave is running (§7.5).** The safe answer, not
  the right one. `WorkspaceMerger` rsyncs expert workspaces over `main/`, and
  those were seeded before the amendment, so a merge would overwrite an applied
  amendment. Making the merger amendment-aware is the real fix.
- **No frontend for amendments.** The endpoints exist and are tested; nothing in
  the React app lists or renders them yet. The existing `ApprovalGate` cannot be
  reused — it only appears while the workflow is `paused_for_approval`, and an
  amendment deliberately pauses nothing.
- **ZIP export (§16.5).** Recorded as the first delivery route and never built.
  §18 shipped instead, so the only way to get the harness out today requires the
  client to connect a git provider.
- **QA phase under the new model.** The implementation phase now authors design
  (§9); the QA phase still runs `AiderRunner` unchanged. What "QA" means when the
  code is written by Claude Code outside this system is not defined anywhere in
  this document, and was not guessed at.
- **Verify commands cannot actually run in the api image.** §17 executes each
  criterion's `Verify` command against the cloned repo through `verify.go`'s
  `runProjectChecks`, which correctly reports `unavailable` when the executable is
  missing. The api image is alpine plus git and rsync — no Go toolchain, no node —
  so in practice nearly every criterion comes back *unverifiable*, and the useful
  half of the report is the grep-based contract diff. Making the acceptance checks
  real means running them somewhere that has a toolchain. Named here because
  "unverifiable" is honest but is not the same as working.
