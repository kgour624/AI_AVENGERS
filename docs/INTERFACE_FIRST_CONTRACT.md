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
