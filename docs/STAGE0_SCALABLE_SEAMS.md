# Stage 0 — Scalable Seams (Modular Monolith Hardening)

**Status:** implemented  
**Goal:** expert sell + independent future scale of experts / chat / workflow  
**Constraint:** zero break of core chat, workflow, experts, admin, auth features

## What landed

| Seam | Package | Role |
|------|---------|------|
| Cross-domain ports | `internal/ports` | `EntitlementCheck`, `KnowledgeReader`, `UsageRecorder`, `EventPublisher` |
| Entitlement | `internal/entitlement` | Implements `EntitlementCheck` over `user_expert_grants` (same rules as before) |
| Knowledge read | `internal/knowledge` | Implements `KnowledgeReader` (catalog read port; call sites migrate incrementally) |
| Outbox | `internal/outbox` + migration `023` | Transactional domain events + SKIP LOCKED dispatcher |
| Metrics | `internal/observability/metrics.go` | `GET /metrics` JSON counters |

## Behaviour preserved

- Chat / project / workflow expert access still goes through `auth.MustHaveExpertAccess` — now a thin delegate to `entitlement.Checker` (identical SQL + role rules).
- Public APIs, JWT, admin, training, orchestrator paths untouched except optional metrics hooks.
- Grant admin writes still succeed if outbox is down (event publish is non-fatal after commit).

## New behaviour (additive)

1. `domain_outbox` table + background dispatcher (2s tick).
2. `entitlement.granted` events after managed-account create / set-grants.
3. `GET /metrics` — llm calls/errors/cost, workflows started/failed, agent loop iters, outbox published, entitlement denied.
4. LLM gateway + workflow runner + agent loop increment metrics (no logic change).

## Run migration

```bash
# same path as existing migrations
migrate -path backend-go/migrations -database "$DATABASE_URL" up
# or your compose migrate service
```

## What we deliberately did NOT do

- No microservice split, no Kafka, no schema-per-DB split.
- Did not rewrite orchestrator/admin SQL onto `KnowledgeReader` yet (port exists; migration is Stage 0.1 to avoid risk).
- Did not change OpenAPI public contracts.

## Next (Stage 0.1 / 1) when you approve

1. Inject `ports.EntitlementCheck` into message/project/workflow handlers (drop `*pgxpool` from access checks).
2. Point orchestrator expert load at `KnowledgeReader`.
3. `cmd/api` vs `cmd/worker` entrypoint split (same image).
