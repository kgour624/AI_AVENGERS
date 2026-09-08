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
| CT-A1 | Migration 010 (categories, retrofit, reply_to_message_id, system_settings) | `backend-go/migrations/010_expert_categories.up.sql` + `.down.sql` | \u23f3 NOT STARTED |
| CT-A2 | `CategoryRegistry` (mirrors `chinawall.DomainRegistry` pattern) | `backend-go/internal/category/registry.go` | \u23f3 NOT STARTED |
| CT-A3 | Admin category CRUD handlers + routes | `backend-go/internal/admin/admin_handler.go`, `cmd/server/main.go` | \u23f3 NOT STARTED |
| CT-A4 | `CreateExpert`/`UpdateExpert` require/accept `category_id` | `backend-go/internal/admin/admin_handler.go` | \u23f3 NOT STARTED |

**Checkpoint CT-A:** admin can create a category with a template schema via
API, create/retrofit an expert into it, and `GET /admin/expert-categories`
returns it with the schema intact.

### Phase CT-B \u2014 Structured generation (backend)

| # | Component | File | Status |
|---|---|---|---|
| CT-B1 | Structured JSON prompt + parse path in Layer 3 | `backend-go/internal/chinawall/enforcer.go` | \u23f3 NOT STARTED |
| CT-B2 | Per-section Layer 4 strip logic | `backend-go/internal/chinawall/enforcer.go` | \u23f3 NOT STARTED |
| CT-B3 | Hardcoded test-case bucket constants | `backend-go/internal/chinawall/template.go` (new) | \u23f3 NOT STARTED |
| CT-B4 | `DecisionResult`/`ExpertResponse` carry structured sections through decision engine + orchestrator | `backend-go/internal/decision/engine.go`, `backend-go/internal/orchestrator/orchestrator.go` | \u23f3 NOT STARTED |

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
