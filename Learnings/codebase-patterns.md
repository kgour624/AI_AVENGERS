# Codebase Patterns — AI_AVENGERS

## Structure & layers
- Backend is a Go module `ai_avengers/backend`; HTTP handlers are in `internal/<domain>`, route/dependency composition is in `cmd/server/main.go`, and DB migrations are paired `.up.sql`/`.down.sql` files. (seen in: `backend-go/go.mod`, `backend-go/internal/admin/admin_handler.go`, `backend-go/cmd/server/main.go`, `backend-go/migrations/`)
- Training pipeline and durable checkpoint/progress logic live in `internal/training`; shared adapters/services are injected through constructors. (seen in: `backend-go/internal/training/ingestion_pipeline.go`, `backend-go/internal/training/checkpoint.go`)
- Python ML sidecar is FastAPI, with `/embed` and `/rerank` endpoints, services split by responsibility, and CPU-bound work passed to bounded `ThreadPoolExecutor`s. (seen in: `ml-sidecar/main.py`, `ml-sidecar/embeddings.py`, `ml-sidecar/reranker.py`)
- Frontend is React + TypeScript + Vite; API calls are in `frontend/src/api`, reusable stream behavior in `frontend/src/hooks`, and admin UI in `frontend/src/components/admin`. (seen in: `frontend/package.json`, `frontend/src/api/admin.ts`, `frontend/src/hooks/useIngestionStream.ts`, `frontend/src/components/admin/`)

## Naming conventions
- Go uses domain package names, exported PascalCase types/functions, unexported camelCase helpers, and stage constants for constrained DB values. (seen in: `backend-go/internal/training/checkpoint.go`, `backend-go/internal/admin/admin_handler.go`)
- TypeScript React components use PascalCase filenames/component names; hooks use `useX`, API functions use camelCase verbs. (seen in: `frontend/src/hooks/useIngestionStream.ts`, `frontend/src/components/admin/IngestionPipelineModal.tsx`, `frontend/src/api/admin.ts`)
- Python uses snake_case functions/variables and PascalCase service/model classes. (seen in: `ml-sidecar/main.py`, `ml-sidecar/embeddings.py`)

## Error handling & logging
- Go HTTP handlers return errors through shared `response` helpers and stable error codes; service code wraps errors with context and uses structured `zap` fields. (seen in: `backend-go/internal/admin/admin_handler.go`, `backend-go/internal/byoexpert/handler.go`, `backend-go/internal/training/ingestion_pipeline.go`)
- Background ingestion records terminal status/error on `ingestion_jobs`; checkpoint writes/logged failures are handled explicitly, and `ErrJobPaused` is a sentinel for an intentional non-crash stop. (seen in: `backend-go/internal/training/checkpoint.go`, `backend-go/internal/training/ingestion_pipeline.go`)
- Timeline/event writes are best-effort observability, while writes required for resume source-of-truth must fail closed. (seen in: `backend-go/internal/jobevents/store.go`, `backend-go/internal/training/prepare.go`)
- Python endpoints raise FastAPI `HTTPException` with appropriate status; services log internal operational details and return structured API models. (seen in: `ml-sidecar/main.py`)

## Design patterns / OOP / SOLID in use
- Constructor dependency injection is the default for backend services/handlers; small interfaces are used at boundaries for testability and to keep package dependencies directed. (seen in: `backend-go/internal/admin/admin_handler.go`, `backend-go/internal/byoexpert/byoexpert.go`, `backend-go/internal/training/ingestion_pipeline.go`)
- PostgreSQL is authoritative for durable job/timeline state; Redis Pub/Sub is used as notification/wakeup, not as the source of truth. (seen in: `backend-go/internal/jobevents/store.go`, `backend-go/internal/jobevents/subscriber.go`, `backend-go/internal/blackboard/`)
- Durable append-only event tables coexist with mutable hot snapshots; migration files document why and rollback behavior. (seen in: `backend-go/migrations/035_slo_events.up.sql`, `backend-go/migrations/036_ingestion_job_events.up.sql`)

## Concurrency model
- Go uses goroutines, `sync.WaitGroup`, mutexes and bounded semaphore channels for independent batch work; pipeline stages remain ordered where later stages depend on earlier results. (seen in: `backend-go/internal/training/ingestion_pipeline.go`, `backend-go/internal/training/checkpoint.go`, `backend-go/internal/orchestrator/orchestrator.go`)
- When concurrent batches can finish out of order, checkpoints retain a contiguous completed prefix so resume cannot skip a gap. (seen in: `backend-go/internal/training/ingestion_pipeline.go`)
- Python sidecar keeps one Uvicorn worker because ML models are memory-heavy; async handles requests, while separate bounded thread pools isolate embedding/reranking and document extraction work. (seen in: `ml-sidecar/Dockerfile`, `ml-sidecar/main.py`)

## Libraries & versions in use
- Go: Go 1.22; Gin 1.10; pgx/v5 5.6; go-redis/v9 9.6.1; zap 1.27; google/uuid 1.6; pgvector-go 0.2.1. (seen in: `backend-go/go.mod`)
- Frontend: React 18.3, TypeScript 5.6, Vite 5.4, TanStack Query 5.56, Zustand 4.5, Axios 1.7, react-dropzone 14.2. (seen in: `frontend/package.json`)
- ML sidecar: Python 3.11, FastAPI 0.111, Uvicorn 0.30, sentence-transformers 3.0.1, Torch 2.3.1, NumPy 1.26.4; document parsers are separately pinned. (seen in: `ml-sidecar/Dockerfile`, `ml-sidecar/requirements.txt`)

## Test conventions
- Go tests use standard `*_test.go` files and `testing`; unit tests commonly test package behavior without external service infrastructure, while backend compile verification uses the Docker build when local Go is unavailable. (seen in: `backend-go/internal/`, `backend-go/Dockerfile`)
- Frontend build command is `npm run build` (`tsc -b && vite build`); lint script exists but requires an ESLint configuration. (seen in: `frontend/package.json`)
- Python source can be syntax-checked with `python -m py_compile`; sidecar dependencies are pinned in requirements. (seen in: `ml-sidecar/requirements.txt`)
