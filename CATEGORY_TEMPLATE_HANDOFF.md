# AI Avengers — Expert Category & Structured Template System (Handoff)

> **Purpose:** Live progress tracker for the Category layer, structured JSON
> response templates, and reply/thread-context feature.
>
> **This is NOT `HANDOFF.md` or `IMPLEMENTATION_HANDOFF.md`.** Those track the
> pre-existing backend/frontend phases (1-6) and the collaboration layer
> (Phases A-F). This file tracks an additive feature that sits on top of both
> without modifying their locked behavior.
>
> **Rule for this file (same as `IMPLEMENTATION_HANDOFF.md`):** never mark a
> component \u2705 without evidence. Evidence = commit hash + file path + a
> checkpoint condition that was verified. "Code written" is not evidence that
> it runs.

---

## Metadata

| Field | Value |
|---|---|
| Version | 0.1.0 |
| Started | 2026-09-08 |
| Depends on | `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`, `DOMAIN_EXPERT_COLLABORATION_DESIGN.md`, `KNOWLEDGE_HUB.md` |
| Repo | https://gitlab.com/zepto-group3/ai_avengers |
| Branch | main (direct push, no MRs) |
| Owner | Kiran Nogia (client + admin) |
| Requested via | Duo Chat design session, 2026-09-08 |

---

## 1. Problem statement (WHY this exists)

Today every domain expert is created flat (`experts.domain` free text) with no
structure to its answers. For coding-style experts (DSA, LLD, System Design,
OOPs, ...) the client wants a **consistent, predictable answer shape** per
category of expert (e.g. DSA -> Pattern / Idea / Code / Walkthrough / Test
Cases), driven **only** by what that expert was actually trained on \u2014 never
generic LLM knowledge. Admin should design this shape once per category and
have it apply automatically to every expert placed in that category.

Second problem: multi-turn follow-up questions ("the answer you just gave \u2014
explain X part") lose context today. There is no way to pin a specific prior
answer as the subject of a follow-up, and no way to selectively loop in
sibling experts on a follow-up with the original answer as context.

---

## 2. Locked decisions for this feature

| # | Decision | Reason |
|---|---|---|
| CT-L1 | Category is a NEW layer above `experts.domain`, not a replacement | Preserves existing China Wall `DomainRegistry`/`DomainProfile` behavior untouched (§16 L10 of `DOMAIN_EXPERT_COLLABORATION_DESIGN.md`: 5-gate + China Wall stay single source of truth for grounding) |
| CT-L2 | `experts.category_id` is nullable | Zero regression for any expert not yet retrofitted; flat-text behavior is the fallback, not an error state |
| CT-L3 | Template output format is strict JSON, not markdown-heading convention | Reliability \u2014 client chose JSON explicitly over the more flexible but format-fragile markdown-heading approach |
| CT-L4 | Test case category labels are a hardcoded Go constant list: `BASE, EDGE, CORNER, STRESS` | Client explicit choice \u2014 not admin-configurable, unlike the rest of the template |
| CT-L5 | Default code language is Java, set at the code level (not admin-required input), overridable per category | Client explicit choice |
| CT-L6 | Reply pins exactly one message by default; "full thread" is an explicit opt-in toggle, never automatic | Client: "jho pin kra sirf uska hoga lekin kbhi kbhi need hote hai poori chain" \u2014 must stay bounded and predictable |
| CT-L7 | Reply targets only the original expert by default; looping in other experts is an explicit UI opt-in | Client explicit choice |
| CT-L8 | Sub-category is deferred. Category -> Domain Expert is the only hierarchy for now | Client: "sub category ko abhi ke liye skip karte hai" |
| CT-L9 | Structure-permission question reuses the reply mechanism (ASK-mode message + reply_to pointer), not a separate flow | Design choice to avoid duplicate plumbing for a one-off interaction |
| CT-L10 | Existing chat/message list, virtualization, streamStore are NOT modified. Reply state is a new isolated store slice | Client: "jho existing hai usko nahi chhedna" |
| CT-L11 | Both existing (DSA) and all future experts get retrofit-assigned a category; nothing stays permanently uncategorized by design | Client: "purane experts ko bhi unki respective categories mein daal do" |

---

## 3. Schema design (migration 010 \u2014 additive only)

### New table: `expert_categories`

```sql
CREATE TABLE expert_categories (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                     VARCHAR(255) NOT NULL,
    slug                     VARCHAR(255) UNIQUE NOT NULL,
    description              TEXT,
    template_schema          JSONB NOT NULL DEFAULT '{}',
    default_language         VARCHAR(20) NOT NULL DEFAULT 'java',
    ask_structure_permission BOOLEAN NOT NULL DEFAULT FALSE,
    created_by               UUID REFERENCES users(id),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

`template_schema` shape (admin-designed, example for a DSA-style category):
```json
{
  "sections": [
    {"key": "pattern",     "label": "Pattern",     "type": "prose",      "required": true},
    {"key": "idea",        "label": "Idea",        "type": "prose",      "required": true},
    {"key": "code",        "label": "Code",        "type": "code",       "required": true},
    {"key": "walkthrough", "label": "Walkthrough", "type": "prose",      "required": true},
    {"key": "test_cases",  "label": "Test Cases",  "type": "test_cases", "required": true}
  ]
}
```
Valid `type` values: `prose`, `code`, `test_cases`. `code` sections always use
`default_language`. `test_cases` sections always render the hardcoded
BASE/EDGE/CORNER/STRESS buckets (CT-L4) \u2014 never admin-defined labels.

### `experts` table
```sql
ALTER TABLE experts
    ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES expert_categories(id);
```
Nullable per CT-L2.

### `messages` table
```sql
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_messages_reply_to ON messages(reply_to_message_id) WHERE reply_to_message_id IS NOT NULL;
```
Self-referencing, single parent pointer. Full-chain resolution is a runtime
traversal (§5), not a stored list \u2014 keeps writes trivial and avoids denormalized
drift.

### Retrofit (same migration)
Best-effort `UPDATE experts SET category_id = <seeded Coding category id> WHERE domain IN ('dsa','algorithms','coding')`
against a small seed category inserted in the same migration. Anything
unmatched stays `category_id IS NULL` \u2014 admin panel must surface these as
"Uncategorized" so they get fixed manually (CT-L11 requires eventual
assignment, not automatic silent success).

### `system_settings` addition
`reply_thread_max_depth` (default 10) \u2014 safety cap for full-thread traversal (§5).

---

## 4. Structured generation (China Wall integration)

ChinaWall Enforcer (`internal/chinawall/enforcer.go`) Layer 3
(`generateWithCitations`) gains a parallel structured path:

- IF expert has `category_id` AND category has non-empty `template_schema.sections`
  \u2192 prompt LLM for strict JSON matching the schema, one key per section.
- ELSE \u2192 existing flat-text behavior, byte-for-byte unchanged (CT-L2 fallback).

Layer 4 (`stripUncited`) runs **per-section** when structured:
- `prose` sections: existing `StripMode` logic (`CODE_EXEMPT` / `FULL_STRIP`) applied to that section's text only.
- `code` / `test_cases` sections: structurally exempt, same principle as existing `StripModeCodeExempt`, just correctly scoped instead of whole-answer.

No changes to `DomainProfile`, `DomainRegistry`, `CoverageMode`, `CitationMode`
types or values (CT-L1).

---

## 5. Reply / thread-context design

- Client clicks **Reply** on any `ExpertResponse` \u2192 frontend sets `reply_to_message_id` on the next outgoing message.
- Backend default: pin **only** that one message + its own question as context (cheapest, matches "sirf pin kra sirf uska").
- Frontend **"Include full thread"** checkbox (explicit opt-in, CT-L6) \u2192 backend walks `reply_to_message_id` pointers upward, root-ward, capped at `reply_thread_max_depth` (default 10, from `system_settings`).
- This resolves into a new `ReplyThread` context source in `context/assembler.go`, added as its own budgeted slice \u2014 does not touch rolling summary / L2 / recent-messages / course-chunk logic (additive only).
- **Expert scope on reply:** default = only the originally-replied-to expert (natural default since each message row has one `expert_id`). Optional **expert multi-select** in reply UI (CT-L7) lets client loop in additional experts; when selected, those experts also receive the pinned/threaded content as part of their assembled context so they are not blind to the prior answer.

---

## 6. Structure-permission gate (CT-L9)

When category has `ask_structure_permission = true` AND the incoming message
is NOT itself a reply (fresh question) \u2192 engine first emits an `ASK`-mode
message ("Chahiye structure/boilerplate ya sirf logic likh doon?"). The
client's answer is itself sent as a **reply** to that ASK message. Backend
detects the reply target was a structure-permission ASK, walks one level up
to recover the original question, then generates the full structured answer
factoring in the stated preference. Reuses §5's reply mechanism entirely \u2014 no
separate flow.

---

## 7. Frontend scope

| Area | Change |
|---|---|
| Admin: new Categories page | CRUD for `expert_categories` + template builder (add/remove/reorder sections, set label/type/required per section, set `default_language`, toggle `ask_structure_permission`) |
| Admin: `CreateExpertModal.tsx` | Add required category dropdown (replaces free-text-only domain entry) |
| Chat: `ExpertResponse.tsx` (or new sibling component) | Reply button per response; structured-section renderer when `templateSections[]` present (prose / code / fixed test-case buckets) |
| Chat: reply UI | "Replying to: <preview>" chip, "Include full thread" checkbox, optional expert multi-select \u2014 **new isolated Zustand slice**, does not modify `streamStore.ts` (CT-L10) |
| `CodeBlock.tsx` | Add copy-to-clipboard button |
| Response card | Add whole-response copy button (client asked for both scopes) |

---

## 8. Implementation phases

### Phase CT-A \u2014 Schema + Category CRUD (backend)

| # | Component | File | Status |
|---|---|---|---|
| CT-A1 | Migration 010 (categories, retrofit, reply_to_message_id, system_settings) | `backend-go/migrations/010_expert_categories.up.sql` + `.down.sql` | \u2705 COMMITTED (2026-09-08) |
| CT-A2 | `CategoryRegistry` (mirrors `chinawall.DomainRegistry` pattern) | `backend-go/internal/category/registry.go` | \u2705 COMMITTED (2026-09-08) |
| CT-A3 | Admin category CRUD handlers + routes | `backend-go/internal/admin/admin_handler.go`, `cmd/server/main.go` | \u2705 COMMITTED (2026-09-08) |
| CT-A4 | `CreateExpert`/`UpdateExpert` accept optional `category_id` | `backend-go/internal/admin/admin_handler.go` | \u2705 COMMITTED (2026-09-08) |

**Evidence (commits on `main`, 2026-09-08, verified by re-reading each file
from `main` after push, not just "code written"):**

1. `feat(category): add migration 010` \u2014 `backend-go/migrations/010_expert_categories.up.sql`
   (169 lines, verified present on `main`) + matching `.down.sql`. Retrofit
   domain list (dsa, algorithms, coding) cross-checked against the actual
   DefaultProfiles entries in `backend-go/internal/chinawall/domain_profile.go`
   (read in full first) rather than guessed.
2. `feat(category): add CategoryRegistry` \u2014 `backend-go/internal/category/registry.go`
   (271 lines, verified present on `main`). Lifecycle mirrors
   `chinawall/domain_registry.go`, which was read in full first and is NOT
   modified by this change (CT-L1).
3. `feat(category): add admin CRUD handlers ...` \u2014 `backend-go/internal/admin/admin_handler.go`
   (1515 lines, verified present on `main`). Added ListExpertCategories,
   CreateExpertCategory, GetExpertCategory, UpdateExpertCategory,
   validateTemplateSchema (rejects any section type outside prose/code/
   test_cases). CreateExpert/UpdateExpert extended with optional
   category_id, validated against the live CategoryRegistry cache before
   any DB write (400 on unknown id, per CT-L2 \u2014 NOT required).
4. `fix(category): wire CategoryRegistry into main.go ...` \u2014
   `backend-go/cmd/server/main.go`. Necessary because commit 3 changed
   NewAdminHandler's signature without updating its only call site, which
   would have left main between commits 3 and 4 non-compiling. Fixed in the
   very next commit, same session; re-read the full file from `main`
   afterward to confirm import, registry init, buildRouter signature/call
   site, NewAdminHandler call site, and the 4 new route registrations are
   mutually consistent.

**Checkpoint CT-A \u2014 honest status:** admin can create a category with a
template schema via `POST /admin/expert-categories`, list/get/update it, and
create/retrofit an expert into it via `category_id`. **This has NOT been
verified against a live running server or database** \u2014 there is no live DB
or Go toolchain available in this session to run `go build ./...`, apply
migration 010, or make a real HTTP request. This mirrors the same honestly-
stated gap already present in `HANDOFF.md` and `IMPLEMENTATION_HANDOFF.md`
Phase A \u2014 correctness here was verified by re-reading every changed file in
full after each commit and manually tracing signatures/call-sites/field
names against the actual schema and existing patterns, NOT by an actual
build or migration run.

**Owed before this phase can be called done:**
- [ ] `cd backend-go && go build ./...` \u2014 confirm the package compiles.
- [ ] Apply migration 010 up against a live Postgres; confirm no errors;
      confirm `SELECT * FROM expert_categories WHERE slug='coding'` returns
      the seeded row with the 5-section template_schema intact.
- [ ] Confirm retrofit matched the expected rows (0 matches is valid if no
      dsa/algorithms/coding-domain expert exists yet).
- [ ] `POST /admin/expert-categories` with a real admin JWT \u2014 confirm 201
      and that CategoryRegistry.Reload picks it up immediately.
- [ ] `POST /admin/experts` with an invalid category_id \u2014 confirm 400
      INVALID_CATEGORY_ID, not a raw DB FK violation.

Owner: Kiran (has DB + running-server access). Runbook commands available on request.

### Phase CT-B \u2014 Structured generation (backend)

| # | Component | File | Status |
|---|---|---|---|
| CT-B1 | Structured JSON prompt + parse path in Layer 3 | `backend-go/internal/chinawall/enforcer.go`, `backend-go/internal/chinawall/template.go` | DONE (2026-09-08) |
| CT-B2 | Per-section Layer 4 strip logic | `backend-go/internal/chinawall/enforcer.go` (`enforceStructured`) | DONE (2026-09-08) |
| CT-B3 | Hardcoded test-case bucket constants | `backend-go/internal/chinawall/template.go` (new) | DONE (2026-09-08) |
| CT-B4 | `DecisionResult`/`ExpertResponse` carry structured sections through decision engine + orchestrator | `backend-go/internal/decision/engine.go`, `backend-go/internal/orchestrator/orchestrator.go`, `backend-go/internal/message/handler.go`, `backend-go/cmd/server/main.go` | DONE (2026-09-08) |

**Evidence (commits on main, 2026-09-08, verified by re-reading each file
from main after push):**

1. feat(chinawall): add template.go -- new file, 169 lines. TestCaseBuckets
   hardcoded [BASE, EDGE, CORNER, STRESS] order (CT-L4, never model or
   admin controlled). buildStructuredPrompt builds the JSON-only Layer 3
   prompt; parseStructuredResponse/renderTestCaseBuckets parse the
   model's JSON into per-section text, re-ordering test_cases buckets to
   the fixed order regardless of what order the model returned them in.
2. feat(chinawall): CT-B1/B2 -- enforcer.go. Enforce() gained two new
   trailing params (templateSections []category.TemplateSection,
   defaultLanguage string), both nil/empty for every non-categorized
   expert (CT-L2) -- verified the entire pre-existing flat-text branch
   (Layer 4 strip + BaseProfile safety net) is untouched other than
   passing the two new params through unchanged. Added enforceStructured
   (per-section Layer 4 + its own BaseProfile retry, mirroring the flat
   path's conflict-resolution logic) and generateStructured (Layer 3
   structured branch, called from generateWithCitations only when
   len(templateSections) > 0).
3. **3 compile-breaking bugs found and fixed in the SAME session, before
   moving to CT-B4** (caught by re-reading the full file end-to-end and
   tracing every call site, per Step 4 "mentally execute before writing"):
   - generateStructured's defined parameter order did not match its call
     site inside generateWithCitations -- fixed by reordering the
     function signature to match the call site.
   - The flat path's own BaseProfile safety-net retry (inside Enforce(),
     unrelated to the new structured code) still called the OLD 6-arg
     generateWithCitations after the signature grew to 8 args -- fixed
     by adding nil, "" to that call site.
   - decision/engine.go's call to chinaWall.Enforce() still used the
     OLD 7-arg signature after Enforce() grew to 9 args -- fixed together
     with CT-B4's Expert.TemplateSections/DefaultLanguage addition in
     the same commit that changed the call site, per the anti-pattern
     checklist rule "changing a method signature -> update every call
     site in the same commit."
4. feat(decision): CT-B4 part 1 -- engine.go. Added
   Expert.TemplateSections/DefaultLanguage (nil/empty for every
   non-categorized expert) and DecisionResult.TemplateSections, wired
   through Process()'s Gate 5 call into the new Enforce() signature
   and the "success" case's return value.
5. feat(orchestrator): CT-B4 part 2 -- orchestrator.go. loadExperts'
   SQL now selects category_id (verified against migration 010's actual
   column name before writing the query, not assumed) into a new
   nullable expertRecord.CategoryID. processWithExpert resolves it via
   a new categoryRegistry *category.Registry field on Orchestrator --
   nil-safe at every step (nil registry, nil CategoryID, unknown id, or
   empty template_schema.Sections all fall through to the flat-text
   path identically). NewOrchestrator gained a categoryRegistry param;
   its only call site (main.go) was updated in the SAME commit.
6. feat(message): CT-B4 completion -- message/handler.go. SSE
   SSEComplete payload now includes template_sections. **Known,
   explicitly documented gap** (not silently worked around): structured
   sections are NOT YET persisted to the messages table --
   saveAssistantMessage still only saves resp.Content, which is empty
   for a structured response. A page reload will show a categorized
   expert's past answer as blank even though the live SSE stream
   renders it correctly. Fixing this needs either a new
   messages.template_sections JSONB column (its own migration) or a
   text-serialization fallback -- deliberately left for a later phase
   since it is outside CT-B's stated checkpoint ("the raw API response
   contains a template_sections array" -- satisfied for the live SSE
   response, not for reload-from-DB).

**Checkpoint CT-B -- honest status:** a categorized expert's live SSE
response (SSEComplete event) now contains a template_sections array
matching the category's schema keys, each independently citation-processed
(Layer 4 applied per prose section, code/test_cases sections structurally
exempt). **This has NOT been verified against a live running server, a
real LLM call, or a real HTTP/SSE connection** -- there is no Go
toolchain, Postgres, ML sidecar, or LLM provider available in this
session. Every claim above was verified by re-reading each committed
file from main and manually tracing every call site, parameter order,
and struct field name -- including catching and fixing 3 real
signature-mismatch bugs this way before they could reach a build step
-- but this is not a substitute for go build ./... or an actual
generation request. Same honesty bar as Phase CT-A and every entry in
HANDOFF.md.

**Owed before this phase can be called done:**
- [ ] cd backend-go && go build ./... -- confirm the package compiles.
      This is the single most important open item; 3 mismatches were
      already caught by manual tracing but manual tracing cannot
      guarantee it catches everything a real compiler would.
- [ ] Create a category with a real template_schema (CT-A, already
      committed) + an expert with that category_id + real trained
      content, send it a question, confirm the SSE complete event's
      template_sections array has the right keys with real generated
      text and non-empty citations on prose sections.
- [ ] Confirm a malformed-JSON LLM response degrades to the documented
      fallback (flat citation extraction on raw content) instead of
      crashing or silently returning empty sections.
- [ ] Confirm reload-from-DB gap (documented above) matches expectation
      -- decide whether to fix now or defer to a dedicated follow-up.

Owner: Kiran (has DB + running-server access). Runbook commands available on request. \u23f3 NOT STARTED |

**Checkpoint CT-B:** a categorized expert with a template answers a coding
question and the raw API response contains a `template_sections` array
matching the category's schema keys, each properly citation-processed.

### Phase CT-C \u2014 Reply / thread context (backend)

| # | Component | File | Status |
|---|---|---|---|
| CT-C1 | `reply_to_message_id` accepted on send + persisted | `backend-go/internal/message/handler.go`, `backend-go/internal/chat/service.go` | \u23f3 NOT STARTED |
| CT-C2 | `ReplyThread` context source (pinned or full-chain, depth-capped) | `backend-go/internal/context/assembler.go` | \u23f3 NOT STARTED |
| CT-C3 | Multi-expert loop-in on reply passes thread context to all selected experts | `backend-go/internal/orchestrator/orchestrator.go` | \u23f3 NOT STARTED |
| CT-C4 | Structure-permission ASK gate + reply-to-ASK detection | `backend-go/internal/decision/engine.go` | \u23f3 NOT STARTED |

**Checkpoint CT-C:** replying to a specific past message with thread-off
changes only that Q&A's context; toggling thread-on and setting depth=3
demonstrably includes 3 ancestor turns; killing the server mid-chat and
resuming does not break reply resolution (existing message table is durable).

### Phase CT-D \u2014 Frontend

| # | Component | File | Status |
|---|---|---|---|
| CT-D1 | Admin Categories page + template builder | `frontend/src/pages/admin/AdminCategories.tsx` (new) | \u23f3 NOT STARTED |
| CT-D2 | `CreateExpertModal.tsx` category dropdown | `frontend/src/components/admin/CreateExpertModal.tsx` | \u23f3 NOT STARTED |
| CT-D3 | Reply state slice (isolated, new) | `frontend/src/stores/replyStore.ts` (new) | \u23f3 NOT STARTED |
| CT-D4 | Reply UI (chip, full-thread toggle, expert multi-select) | `frontend/src/components/chat/MessageInput.tsx` | \u23f3 NOT STARTED |
| CT-D5 | Structured section renderer | `frontend/src/components/chat/ExpertResponse.tsx` (or new sibling) | \u23f3 NOT STARTED |
| CT-D6 | Copy button in `CodeBlock.tsx` | `frontend/src/components/chat/CodeBlock.tsx` | \u23f3 NOT STARTED |
| CT-D7 | Whole-response copy button | `frontend/src/components/chat/ExpertResponse.tsx` | \u23f3 NOT STARTED |

**Checkpoint CT-D:** admin builds a template visually, an expert in that
category renders Pattern/Idea/Code/Walkthrough/TestCases as distinct UI
sections, reply + full-thread toggle work end-to-end, copy buttons work on
both code blocks and whole responses.

---

## 9. Definition of Done (per component, same bar as `IMPLEMENTATION_HANDOFF.md`)

- [ ] Code written and pushed to `main`
- [ ] File exists on `main` (verified via tree listing)
- [ ] Mentally executed for happy path + at least 2 edge cases (empty template, no category, max thread depth exceeded, non-Java category override, reply-to-a-reply chain)
- [ ] Anti-pattern checklist passed (signature changes -> all callers updated in same commit; DB fields verified against actual schema; no silent errors; nullable handling correct)
- [ ] Existing behavior (flat-text experts, existing chat/message flow, `DomainRegistry`) verified NOT to have regressed
- [ ] This file updated with truthful status \u2014 no \u2705 without evidence
- [ ] Locked decisions (§2, CT-L1 through CT-L11) not violated

---

## 10. Known open questions for later (explicitly deferred, not forgotten)

- Sub-category layer (CT-L8 defers this)
- Whether structure-permission gate should also apply mid-thread (currently scoped to fresh questions only)
- Kanban/workflow integration with categorized experts (out of scope \u2014 collaboration layer Phase B experts don't exist yet per `IMPLEMENTATION_HANDOFF.md`)

---

## Change log

- 2026-09-08 v0.1.0: initial design captured from Duo Chat session. No code written yet.
