# AI Avengers — Implementation Handoff File

> **Purpose:** Track every component's implementation status. Updated after every commit.
> **Rule:** Never say "done" until Definition of Done checklist is complete.

---

## Repository

- **URL:** https://gitlab.com/zepto-group3/ai_avengers
- **Branch:** main (direct push, no MRs)
- **Architecture Doc:** `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`

---

## Overall Status

| Phase | Status | Started | Completed |
|---|---|---|---|
| Phase 1: Foundation | ⏳ IN PROGRESS | 2026-09-04 | — |
| Phase 2: Expert Training | ⏳ NOT STARTED | — | — |
| Phase 3: Core Intelligence | ⏳ NOT STARTED | — | — |
| Phase 4: Project & Chat | ⏳ NOT STARTED | — | — |
| Phase 5: Advanced Features | ⏳ NOT STARTED | — | — |
| Phase 6: Polish & Deploy | ⏳ NOT STARTED | — | — |

---

## Phase 1: Foundation

### Components

| Component | File | Status | What It Does |
|---|---|---|---|
| Architecture Doc | `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` | ✅ COMPLETE | Full system design |
| Handoff File | `HANDOFF.md` | ✅ COMPLETE | This file |
| Go Project Setup | `backend-go/` | ⏳ IN PROGRESS | Go module, directory structure |
| Database Migrations | `backend-go/migrations/` | ⏳ PENDING | All 12 tables |
| Auth Service | `backend-go/internal/auth/` | ⏳ PENDING | JWT + bcrypt + TOTP |
| ML Sidecar | `ml-sidecar/` | ⏳ PENDING | FastAPI + embeddings + reranker |
| Model Gateway | `backend-go/internal/gateway/` | ⏳ PENDING | OpenRouter integration |
| Docker Setup | `docker-compose.yml` | ⏳ PENDING | All services |

### Checkpoint for Phase 1
- [ ] Admin can login with TOTP
- [ ] JWT issued and validated
- [ ] Embeddings generate from ML sidecar
- [ ] Database migrations run clean
- [ ] All services start with docker-compose up

---

## Phase 2: Expert Training

### Components

| Component | File | Status | What It Does |
|---|---|---|---|
| Text Chunker | `backend-go/internal/training/chunker.go` | ⏳ PENDING | 500-800 token chunks |
| Topic Extractor | `backend-go/internal/training/topic_extractor.go` | ⏳ PENDING | LLM-based topic tagging |
| Charter Extractor | `backend-go/internal/training/charter_extractor.go` | ⏳ PENDING | WHY-embedded charter extraction |
| Capability Builder | `backend-go/internal/training/capability_builder.go` | ⏳ PENDING | Depth levels 1-5 per topic |
| Ingestion Pipeline | `backend-go/internal/training/ingestion.go` | ⏳ PENDING | Full transcript → DB pipeline |
| Expert Admin APIs | `backend-go/internal/admin/experts.go` | ⏳ PENDING | CRUD + ingest trigger |

### Checkpoint for Phase 2
- [ ] Upload transcript → expert created
- [ ] Chunks searchable via semantic search
- [ ] Charter extracted with WHY reasoning
- [ ] Capability table populated

---

## Phase 3: Core Intelligence

### Components

| Component | File | Status | What It Does |
|---|---|---|---|
| Memory Manager | `backend-go/internal/memory/manager.go` | ⏳ PENDING | L1/L2/L3 operations |
| Context Budget Manager | `backend-go/internal/context/budget_manager.go` | ⏳ PENDING | Smart context assembly |
| China Wall Enforcer | `backend-go/internal/chinawall/enforcer.go` | ⏳ PENDING | 4-layer citation enforcement |
| Decision Engine | `backend-go/internal/decision/engine.go` | ⏳ PENDING | 5-gate system |
| Expert Orchestrator | `backend-go/internal/orchestrator/orchestrator.go` | ⏳ PENDING | Parallel expert coordination |
| Synthesizer | `backend-go/internal/orchestrator/synthesizer.go` | ⏳ PENDING | Multi-expert result synthesis |

### Checkpoint for Phase 3
- [ ] Ask question → 5 gates run in sequence
- [ ] China Wall blocks uncited claims
- [ ] Multiple experts run in parallel
- [ ] L1/L2/L3 memory updates after each turn

---

## Phase 4: Project & Chat

### Components

| Component | File | Status | What It Does |
|---|---|---|---|
| Project Service | `backend-go/internal/project/service.go` | ⏳ PENDING | Project CRUD |
| Chat Service | `backend-go/internal/chat/service.go` | ⏳ PENDING | Chat window management |
| Message Handler | `backend-go/internal/message/handler.go` | ⏳ PENDING | SSE streaming responses |
| Chat Index Manager | `backend-go/internal/chat/index_manager.go` | ⏳ PENDING | Smart turn indexing |
| Rolling Summary | `backend-go/internal/chat/summary_generator.go` | ⏳ PENDING | Every 10 turns |

### Checkpoint for Phase 4
- [ ] Create project → add experts → send message → get SSE response
- [ ] Chat index updated after each turn
- [ ] Rolling summary generated at turn 10, 20, 30...

---

## Phase 5: Advanced Features

### Components

| Component | File | Status | What It Does |
|---|---|---|---|
| GitHub/GitLab OAuth | `backend-go/internal/repo/oauth.go` | ⏳ PENDING | OAuth2 flow |
| Repo Sync Worker | `backend-go/internal/repo/sync_worker.go` | ⏳ PENDING | Clone + chunk + embed repo |
| Rating Engine | `backend-go/internal/learning/rating_engine.go` | ⏳ PENDING | 1-5 stars + chunk boost |
| Admin Panel APIs | `backend-go/internal/admin/` | ⏳ PENDING | Full admin endpoints |

### Checkpoint for Phase 5
- [ ] Connect GitHub repo → expert can reference code
- [ ] Rate response → chunk boost_factor updates
- [ ] Admin can view violation logs and ratings

---

## Phase 6: Polish & Deploy

### Components

| Component | File | Status | What It Does |
|---|---|---|---|
| Rate Limiting | `backend-go/internal/middleware/rate_limiter.go` | ⏳ PENDING | 100 req/min per IP |
| Cost Monitoring | `backend-go/internal/monitoring/cost_tracker.go` | ⏳ PENDING | Daily/monthly cost alerts |
| Docker Production | `docker-compose.prod.yml` | ⏳ PENDING | Production config |
| Load Testing | `tests/load/` | ⏳ PENDING | 100 concurrent users |

---

## Decisions Made

| # | Decision | Reason | Date |
|---|---|---|---|
| 1 | Go backend | Goroutines for parallel experts | 2026-09-04 |
| 2 | Python ML sidecar | sentence-transformers stability | 2026-09-04 |
| 3 | PostgreSQL + pgvector | Single DB, vector + relational | 2026-09-04 |
| 4 | Redis for L1 memory | O(1) hot path | 2026-09-04 |
| 5 | UUID primary keys | Distributed-safe | 2026-09-04 |
| 6 | Append-only L3 | Audit trail | 2026-09-04 |
| 7 | Modular monolith v1 | No premature microservices | 2026-09-04 |
| 8 | TOTP for admin | Security | 2026-09-04 |
| 9 | WHY in every charter rule | Arpit's principle | 2026-09-04 |
| 10 | SSE for streaming | Simpler than WebSocket | 2026-09-04 |

---

## Anti-Patterns Found and Fixed

*Will be updated as implementation progresses.*

---

## Current Session Notes

- Architecture doc complete and pushed
- Handoff file created
- Phase 1 implementation starting now
- Go project structure being created

---

*Last updated: 2026-09-04*
