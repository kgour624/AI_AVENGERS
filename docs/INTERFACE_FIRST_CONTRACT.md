# Interface-First Contract - AI Avengers

Status: DESIGN - AWAITING CONFIRMATION BEFORE ENFORCEMENT WORK BEGINS
Author: System Design Architect
Last Updated: 2026-09-06

## 1. What Interface-First Means Here

Blueprint analogy: backend-go/api/openapi.yaml is the blueprint. Go backend and React frontend must both GENERATE their types from it, not hand-write independent guesses. If the spec says fullName is a required string, neither side can drift from that - a generator can't guess, it either matches the spec or the build fails.

## 2. Current State Audit

Finding: Interface-First was designed but never enforced.

- backend-go/api/openapi.yaml exists (1044 lines, camelCase schemas) - written, never used.
- backend-go/api/oapi-codegen.yaml exists - never run.
- Makefile has `make generate` and `make verify-contract` - never run.
- backend-go/internal/api/generated/ contains only a .gitkeep - no Go types were ever generated.
- frontend/src/types/generated.ts does not exist - no TS types were ever generated.
- Both sides instead use hand-written types (frontend types/*.ts, Go handler structs) that were guessed by reading each other's source.

This is the root cause of nearly every bug already logged in HANDOFF.md's blocker list and end-to-end audit: casing mismatch bridge, wrong multipart field name, fabricated AdminStats shape, wrong addProjectExpert return type, Expert fields that don't exist on the public endpoint, missing GET /me, missing OAuth routes, missing DB columns. None of these are possible when both sides generate from one spec instead of hand-guessing.

Conclusion: the blueprint exists, nobody has built from it yet.

## 3. Contract Boundaries Map

Three boundaries exist, each needs its own mechanism:

| Boundary | Source of truth | Generator | Enforcement |
|---|---|---|---|
| B1: Frontend <-> Backend (/api/v1/*) | backend-go/api/openapi.yaml | oapi-codegen (Go) + openapi-typescript (TS) | make verify-contract |
| B2: Backend <-> PostgreSQL | backend-go/migrations/*.up.sql | none today | new: schema-diff check, see section 5 |
| B3: Backend <-> ML Sidecar | not yet defined | none today | new: internal spec, see section 6 |

## 4. B1 - Frontend <-> Backend

WHEN a new endpoint or field is needed DO add it to openapi.yaml first, run make generate, then write code against the generated type BECAUSE that is the only point forcing both sides to agree EXCEPT no exceptions - every past bug came from skipping this.

WHEN openapi.yaml changes DO run make generate and commit both generated files in the same commit BECAUSE verify-contract exists to catch drift but only if it runs EXCEPT none.

WHEN deciding casing DO keep openapi.yaml camelCase and make Go's json struct tags camelCase directly BECAUSE the existing utils/casing.ts bridge only exists because this was never confirmed (HANDOFF.md Gap #1) EXCEPT DB columns stay snake_case, the Go struct tag is the translation point, not a frontend bridge.

Audit needed before enforcement: openapi.yaml must be diffed against every real route in the fixed cmd/server/main.go. Not done yet in this session - first concrete task, not assumed complete.

## 5. B2 - Backend <-> PostgreSQL

No generator wired up today. Two options, neither picked yet:

- Option A: adopt sqlc - generates type-safe Go structs from .sql queries + schema. Strong guarantee, large blast radius (rewrites every hand-written query).
- Option B: keep hand-written pgx queries, add a lightweight CI check that diffs migrations/*.up.sql columns against .Scan() call sites. Weaker guarantee, small blast radius, no new dependency.

## 6. B3 - Backend <-> ML Sidecar

FastAPI already generates an OpenAPI spec for free from ml-sidecar/main.py's Pydantic models, currently unused.

WHEN ml/sidecar_client.go's request/response structs are defined DO generate them from the sidecar's own openapi spec instead of hand-writing a second guess BECAUSE same drift risk as B1, cheaper to fix now while it is only 2 endpoints EXCEPT if the spec is runtime-only, snapshot it into ml-sidecar/openapi.json at build time so oapi-codegen has a stable file.

## 7. Migration Plan - Section by Section

| Phase | Scope | Exit condition |
|---|---|---|
| 1 | Diff openapi.yaml against real routes in main.go. Run make generate for the first time on both sides. | Both generated files exist and compile (go build, tsc --noEmit) |
| 2 | Re-point one module (api/auth.ts + types/auth.ts, smallest/most audited) to import generated.ts. Delete hand-written duplicate only after generated one compiles clean. | Login/Register still work against the real backend |
| 3 | Repeat per module: experts -> projects -> chats/messages -> admin, each with its own VERIFIED/GAP report. | All hand-written types/*.ts deleted except casing bridge if kept |
| 4 | Wire make verify-contract into an actual CI job (none exists today). | Spec change without regenerating fails a pipeline, not just an unused local target |
| 5 | Decide and implement B2 and B3. | Both boundaries have some automated drift check |

Not in scope: rewriting business logic. If a re-point reveals the generated type doesn't match real backend behavior, that is a new bug to report, not something to patch by hand-editing the generated file - fix the spec, regenerate.

## 8. Open Questions - Need Answers Before Phase 1

1. B2: sqlc (Option A) or lightweight column-diff check (Option B)? Or defer B2 entirely for now?
2. Casing: confirm Go should own camelCase JSON tags directly, killing utils/casing.ts?
3. Should Phase 1's diff happen only after someone actually runs go build to confirm the duplicate-buildRouter fix really compiles (HANDOFF.md marks it fixed but never built)?
4. CI platform: repo pushes directly to main, no MRs - does Phase 4 mean a .gitlab-ci.yml running on every push to main (report-only), or a pre-push local hook?

No implementation begins until these four are answered.

## 9. Phase 1 Progress Log

**Prerequisite for Q3 confirmed:** admin ran `docker-compose up` locally, backend and frontend both started successfully, login worked end-to-end. This confirms the duplicate-buildRouter fix (HANDOFF.md) is real - `main.go` has exactly one `buildRouter` definition, verified by reading the full 687-line file. Phase 1's route diff below is against a backend confirmed to actually compile and run, not an assumption.

**Finding 1 - FIXED: response casing violated the spec.** `main.go`'s `buildAuthResponse`, `handleGetMe`, and `handleRefresh` emitted snake_case (`full_name`, `totp_enabled`, `token_pair`, `access_token`, `expires_in_seconds`) while `openapi.yaml`'s `User`/`TokenPair`/`AuthResponse` schemas are camelCase - a real, live contract violation, not a hypothetical. Per §4's locked decision (Go owns camelCase directly), fixed on the Go side.

Cross-checked before touching anything, per the project's own rule (frontend + backend together, not backend alone):
- `frontend/src/utils/casing.ts`'s `camelizeKeys()` fast-paths any key with no underscore - confirmed switching backend to camelCase is a safe no-op for every response that goes through `baseAPI`'s interceptor (`buildAuthResponse`, `handleGetMe`).
- `frontend/src/api/base.ts`'s `refreshSession()` uses a **raw** `axios.post` call that bypasses `camelizeKeys()` entirely (deliberate, to avoid interceptor re-entrancy if refresh itself 401s) and read `res.data.data.access_token` directly. Changing `handleRefresh`'s response casing without also fixing this call in the same commit would have broken token refresh silently. Both were fixed together in one commit.
- Request-side snake_case bindings (`RegisterRequest.FullName` still binds `json:"full_name"`, refresh's body-fallback field still binds `json:"refresh_token"`) were **deliberately left unchanged** - `frontend/src/api/base.ts`'s request interceptor still calls `snakeifyKeys()` on outgoing bodies today, so the current snake_case request bindings are still correct and working. Changing these now, without also removing the frontend's outgoing snakeify step in the same change, would break register. This is Phase 2/4 scope (removing the casing bridge entirely), not this fix.

**Finding 2 - GAP, not yet fixed: `openapi.yaml` is missing ~18 real, registered routes.** Diffed the spec's `paths:` section (ends at `/messages/{id}/rate`) against every route actually registered in `main.go`'s `buildRouter`. Missing entirely from the spec:

```
GET  /repo/oauth/:provider
GET  /repo/callback/:provider
POST /projects/:id/repo
POST /projects/:id/repo/sync
GET  /projects/:id/repo/status
GET  /projects/:id/memory
GET  /projects/:id/timeline
GET  /admin/experts
POST /admin/experts
PATCH /admin/experts/:id
POST /admin/experts/:id/ingest
GET  /admin/experts/:id/jobs
GET  /admin/clients
PATCH /admin/clients/:id
GET  /admin/stats
GET  /admin/violations
GET  /admin/ratings
GET  /admin/settings
PATCH /admin/settings/:key
```

Not fabricating schemas for these yet - each needs its real response shape read from the actual handler/service first (`repo/service.go`, `admin/admin_handler.go`, `memory/manager.go`), same discipline as the casing fix above. This is the next concrete Phase 1 task, pending confirmation to proceed.
