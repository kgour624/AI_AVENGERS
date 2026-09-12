# CodeCraftAPI Integration — Implementation Handoff

> **Purpose:** Live progress tracker for CodeCraftAPI integration.
> **Rule:** Never mark a component ✅ without evidence (commit hash + file path + checkpoint verified).
> **Design doc:** `CODECRAFTAPI_INTEGRATION_DESIGN.md`
> **Started:** 2026-09-12
> **Branch:** main (direct push, no MRs)
> **Owner:** Kiran Nogia

---

## Overall Status

| Phase | Status | Evidence |
|---|---|---|
| CC-1: Backend LLM provider | ⏳ NOT STARTED | — |
| CC-2: Embedder interface | ⏳ NOT STARTED | — |
| CC-3: CodeCraftAPI embedder | ⏳ NOT STARTED | — |
| CC-4: Config + gateway wiring | ⏳ NOT STARTED | — |
| CC-5: Admin handler changes | ⏳ NOT STARTED | — |
| CC-6: main.go wiring | ⏳ NOT STARTED | — |
| CC-7: Caller signature updates | ⏳ NOT STARTED | — |
| CC-8: Frontend types + API | ⏳ NOT STARTED | — |
| CC-9: Admin UI changes | ⏳ NOT STARTED | — |

---

## Locked Decisions (Cannot Change Without Discussion)

| # | Decision | Reason |
|---|---|---|
| 1 | Reranking ALWAYS stays on Python sidecar | CodeCraftAPI has no `/v1/rerank`. China Wall Layer 1 depends on bge-reranker-base. |
| 2 | `Embedder` interface (SOLID-D) | Callers depend on abstraction, not `*SidecarClient` concrete type. |
| 3 | `DynamicEmbedder` reads DB at call time | No restart needed to switch embedding provider. Same pattern as `ModelGateway.getActiveProvider()`. |
| 4 | `CostPer1K()` returns `0, 0` | CodeCraftAPI pricing unknown. Honest zero > fabricated number. |
| 5 | Model names in `system_settings` | No migration needed. Consistent with existing LLM settings pattern. |
| 6 | Admin UI proxies model list via backend | API key never leaves server. Frontend never calls CodeCraftAPI directly. |
| 7 | Embedding provider change warning in UI | Changing provider makes existing vectors incompatible. Admin must re-ingest all transcripts. |
| 8 | `SidecarClient` struct unchanged | Zero regression. Wrap it, never modify it. |
| 9 | All caller signature changes in ONE commit | Intermediate state does not compile. Compiler catches all missed call sites. |

---

## Files to CREATE

| File | Phase | Status | What it does |
|---|---|---|---|
| `backend-go/internal/gateway/providers/codecraftapi.go` | CC-1 | ⏳ | `CodeCraftAPIProvider`: implements `LLMProvider` interface |
| `backend-go/internal/ml/embedder.go` | CC-2 | ⏳ | `Embedder` interface + `DynamicEmbedder` + `NewDynamicEmbedder` factory |
| `backend-go/internal/ml/codecraftapi_embeddings.go` | CC-3 | ⏳ | `CodeCraftAPIEmbedder`: calls `/v1/embeddings` |
| `frontend/src/types/codecraftapi.ts` | CC-8 | ⏳ | `CodeCraftModel` type |

---

## Files to MODIFY

| File | Phase | Status | What changes |
|---|---|---|---|
| `backend-go/internal/config/config.go` | CC-4 | ⏳ | Add `ProviderCodeCraftAPI` constant + `CodeCraftAPIKey` + `CodeCraftAPIBaseURL` + default |
| `backend-go/internal/gateway/model_gateway.go` | CC-4 | ⏳ | Add `codecraftapi` case in `buildProvider()` + `getAPIKey()` + new `getModelName()` method |
| `backend-go/internal/admin/admin_handler.go` | CC-5 | ⏳ | Add `embedder ml.Embedder` field + `codecraftapi` to valid providers + 3 new handlers |
| `backend-go/cmd/server/main.go` | CC-6 | ⏳ | Build `ccEmbedder` + `DynamicEmbedder` + pass to constructors + 3 new admin routes |
| `backend-go/internal/training/ingestion_pipeline.go` | CC-7 | ⏳ | `*ml.SidecarClient` → `ml.Embedder` (embedding calls only) |
| `backend-go/internal/context/assembler.go` | CC-7 | ⏳ | `*ml.SidecarClient` → `ml.Embedder` (embedding calls only) |
| `backend-go/internal/memory/l2_store.go` | CC-7 | ⏳ | `*ml.SidecarClient` → `ml.Embedder` (embedding calls only) |
| `backend-go/internal/memory/manager.go` | CC-7 | ⏳ | Accept `ml.Embedder` instead of `*ml.SidecarClient` |
| `backend-go/internal/message/handler.go` | CC-7 | ⏳ | `h.mlClient.EmbedSingle()` → `h.embedder.EmbedSingle()` |
| `frontend/src/api/admin.ts` | CC-8 | ⏳ | Add `getCodeCraftModels` + `getEmbeddingSettings` + `updateEmbeddingSettings` + extend `updateLLMSettings` |
| `frontend/src/pages/admin/AdminLLMSettings.tsx` | CC-9 | ⏳ | Add CodeCraftAPI provider card + model picker + embedding settings section |

---

## Files NOT Touched (Zero Regression — Verified)

| File | Why safe |
|---|---|
| `backend-go/internal/chinawall/enforcer.go` | Uses `SidecarClient.Rerank()` only — not in `Embedder` interface |
| `backend-go/internal/decision/engine.go` | No ML calls |
| `backend-go/internal/orchestrator/orchestrator.go` | No ML calls |
| `backend-go/internal/ml/sidecar_client.go` | Struct unchanged. Already satisfies `Embedder` via Go structural typing. |
| `backend-go/internal/auth/` | No ML calls |
| `backend-go/internal/chat/` | No ML calls |
| `backend-go/internal/project/` | No ML calls |
| `backend-go/internal/rating/` | No ML calls |
| `backend-go/internal/repo/` | No ML calls |
| `backend-go/internal/blackboard/` | No ML calls |
| `backend-go/internal/workflow/` | No ML calls |
| `backend-go/internal/validation/` | No ML calls |
| `frontend/src/pages/chat/ChatPage.tsx` | No LLM/embedding config |
| `frontend/src/hooks/useSSEStream.ts` | No LLM/embedding config |

---

## system_settings Keys (No Migration Needed)

New rows in existing `system_settings` table. No schema change.

| Key | Value type | Default | Purpose |
|---|---|---|---|
| `llm_provider` | string JSON | `"openrouter"` | Active LLM provider (existing — add `"codecraftapi"` as valid value) |
| `llm_api_keys` | map JSON | `{}` | API keys per provider (existing — add `"codecraftapi"` key) |
| `codecraftapi_model_cheap` | string JSON | `""` | Model name for cheap tier |
| `codecraftapi_model_strong` | string JSON | `""` | Model name for strong tier |
| `codecraftapi_model_fast` | string JSON | `""` | Model name for fast tier |
| `embedding_provider` | string JSON | `"sidecar"` | `"sidecar"` or `"codecraftapi"` |
| `embedding_model` | string JSON | `""` | Model name for embeddings (only when `embedding_provider = "codecraftapi"`) |

---

## Caller Signature Change Map

ALL must be updated in the SAME commit as `ml/embedder.go` creation.
Compiler will catch any missed call site — let it go red, follow the red.

| Constructor | Param that changes | From | To |
|---|---|---|---|
| `appcontext.NewAssembler` | 2nd param | `*ml.SidecarClient` | `ml.Embedder` |
| `training.NewIngestionPipeline` | 2nd param | `*ml.SidecarClient` | `ml.Embedder` |
| `message.NewHandler` | 4th param | `*ml.SidecarClient` | `ml.Embedder` |
| `admin.NewAdminHandler` | new 4th param | (did not exist) | `ml.Embedder` |
| `memory.NewManager` | 3rd param | `*ml.SidecarClient` | `ml.Embedder` |

---

## Phase Details

### CC-1: Backend LLM Provider

**Goal:** Create `CodeCraftAPIProvider` that implements `LLMProvider` interface.

**File:** `backend-go/internal/gateway/providers/codecraftapi.go`

**Checkpoint:** `go build ./...` passes. Provider compiles.

**Anti-patterns to watch:**
- DO NOT hardcode model names — they come from constructor parameters
- DO NOT hardcode base URL — comes from constructor (default: `https://codecraftapi.com/v1`)
- DO NOT write new response parsing — reuse `doOpenAICompatibleCall()` from `common.go`
- DO NOT fabricate `CostPer1K()` values — return `0, 0`

**Interface compliance (all 5 methods required):**
```
Name() string                                                    → "codecraftapi"
ModelName(tier ModelType) string                                 → per-tier model name from constructor
CostPer1K(tier ModelType) (float64, float64)                     → 0, 0
MaxTokens(tier ModelType) int                                    → 8192
Call(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
```

**Mental execution before writing:**
```
Happy path:
  Input: ModelStrong, system="You are...", user="How to shard?"
  1. Build messages slice
  2. Marshal JSON body with model name from constructor
  3. POST https://codecraftapi.com/v1/chat/completions
     Authorization: Bearer cc_xxx
  4. doOpenAICompatibleCall() parses response
  5. Return ProviderResponse{Content, InputTokens, OutputTokens, ModelUsed}

Edge case — empty model name:
  ModelName() returns "" → CodeCraftAPI returns 400
  → doOpenAICompatibleCall returns error "provider returned status 400"
  → ModelGateway retries 3× → all fail → returns error to caller
  → Admin sees error, fixes model name in LLM Settings

Edge case — wrong API key:
  POST → 401 Unauthorized
  → error "provider returned status 401"
  → retries 3× → all fail → error returned
```

| Component | File | Status | What it does |
|---|---|---|---|
| CodeCraftAPIProvider | `providers/codecraftapi.go` | ⏳ | LLMProvider impl for CodeCraftAPI |

---

### CC-2: Embedder Interface

**Goal:** Create `Embedder` interface. `SidecarClient` already satisfies it via Go structural typing.
Create `DynamicEmbedder` that reads `embedding_provider` from DB at call time.

**File:** `backend-go/internal/ml/embedder.go`

**Checkpoint:** `go build ./...` passes. Interface compiles.

**Anti-patterns to watch:**
- DO NOT include `Rerank()` in the interface — reranking is sidecar-only always
- DO NOT include `HealthCheck()` in the interface — not needed by embedding callers
- DynamicEmbedder DB read failure MUST fall back to sidecar (safe default)

**Interface definition:**
```go
type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    EmbedSingle(ctx context.Context, text string) ([]float32, error)
}
```

**Mental execution — DynamicEmbedder.Embed():**
```
Happy path (sidecar):
  1. SELECT value FROM system_settings WHERE key='embedding_provider'
  2. Returns "sidecar" (or row not found)
  3. Delegate to sidecarClient.Embed(ctx, texts)
  4. Return result

Happy path (codecraftapi):
  1. SELECT value FROM system_settings WHERE key='embedding_provider'
  2. Returns "codecraftapi"
  3. Delegate to ccEmbedder.Embed(ctx, texts)
  4. Return result

Edge case — DB read fails:
  1. SELECT fails (DB down)
  2. Log warning: "embedding_provider read failed, falling back to sidecar"
  3. Delegate to sidecarClient.Embed(ctx, texts)
  4. Return result (safe fallback)
```

| Component | File | Status | What it does |
|---|---|---|---|
| Embedder interface | `ml/embedder.go` | ⏳ | Abstract embedding source |
| DynamicEmbedder | `ml/embedder.go` | ⏳ | Reads DB at call time, delegates |
| NewDynamicEmbedder | `ml/embedder.go` | ⏳ | Factory function |

---

### CC-3: CodeCraftAPI Embedder

**Goal:** Create `CodeCraftAPIEmbedder` that calls `/v1/embeddings`.

**File:** `backend-go/internal/ml/codecraftapi_embeddings.go`

**Checkpoint:** `go build ./...` passes. Embedder compiles.

**Anti-patterns to watch:**
- DO NOT hardcode model name — comes from `DynamicEmbedder` reading `embedding_model` from DB
- Sort response by `index` field before returning (defensive)
- Timeout: 300s (same as sidecar — batch embedding can be slow)
- DO NOT use string interpolation for vectors — always `pgvector.NewVector()` as `$N` param

**Mental execution — Embed():**
```
Happy path:
  Input: ["How to shard?", "Use consistent hashing"]
  1. Marshal: {"model": "cc-embed-X", "input": ["How to shard?", "Use consistent hashing"]}
  2. POST https://codecraftapi.com/v1/embeddings
     Authorization: Bearer cc_xxx
  3. Parse: data[0]={embedding:[...], index:0}, data[1]={embedding:[...], index:1}
  4. Sort by index (defensive)
  5. Return [[...768 floats...], [...768 floats...]]

Edge case — empty input:
  Input: []
  1. Return [][]float32{}, nil (no API call needed)

Edge case — wrong model name:
  POST → 400 Bad Request
  Return nil, fmt.Errorf("codecraftapi embeddings: status 400")
  Caller (ingestion_pipeline) fails job → resume-capable checkpoint

Edge case — response has fewer items than input:
  len(data) < len(texts)
  Return nil, fmt.Errorf("codecraftapi embeddings: expected %d, got %d", len(texts), len(data))
```

| Component | File | Status | What it does |
|---|---|---|---|
| CodeCraftAPIEmbedder | `ml/codecraftapi_embeddings.go` | ⏳ | Calls /v1/embeddings |

---

### CC-4: Config + Gateway Wiring

**Goal:** Add `ProviderCodeCraftAPI` constant + config fields + `buildProvider()` case.

**Files:** `config/config.go`, `gateway/model_gateway.go`

**Checkpoint:** `go build ./...` passes.

**config.go changes:**
```go
// Add constant:
const ProviderCodeCraftAPI LLMProvider = "codecraftapi"

// Add to LLMConfig struct:
CodeCraftAPIKey     string  // env: CODECRAFTAPI_KEY
CodeCraftAPIBaseURL string  // env: CODECRAFTAPI_BASE_URL

// Add to Load():
CodeCraftAPIKey:     v.GetString("CODECRAFTAPI_KEY"),
CodeCraftAPIBaseURL: v.GetString("CODECRAFTAPI_BASE_URL"),

// Add to applyDefaults():
if c.LLM.CodeCraftAPIBaseURL == "" {
    c.LLM.CodeCraftAPIBaseURL = "https://codecraftapi.com/v1"
}
```

**model_gateway.go changes:**
```go
// In buildProvider() switch:
case config.ProviderCodeCraftAPI:
    cheapModel  := g.getModelName(ctx, "codecraftapi_model_cheap",  "")
    strongModel := g.getModelName(ctx, "codecraftapi_model_strong", "")
    fastModel   := g.getModelName(ctx, "codecraftapi_model_fast",   "")
    return providers.NewCodeCraftAPIProvider(
        apiKey, g.cfg.CodeCraftAPIBaseURL,
        cheapModel, strongModel, fastModel,
        g.httpClient,
    )

// In getAPIKey() switch:
case config.ProviderCodeCraftAPI:
    return g.cfg.CodeCraftAPIKey

// New method:
func (g *ModelGateway) getModelName(ctx context.Context, settingKey, fallback string) string {
    // Read from system_settings, fall back to provided default
    // Same pattern as getActiveProvider()
}
```

| Component | File | Status | What it does |
|---|---|---|---|
| ProviderCodeCraftAPI constant | `config/config.go` | ⏳ | New provider identifier |
| CodeCraftAPIKey + BaseURL fields | `config/config.go` | ⏳ | Env var config |
| buildProvider() case | `gateway/model_gateway.go` | ⏳ | Builds CodeCraftAPIProvider |
| getAPIKey() case | `gateway/model_gateway.go` | ⏳ | Returns CodeCraftAPI key |
| getModelName() method | `gateway/model_gateway.go` | ⏳ | Reads per-tier model name from DB |

---

### CC-5: Admin Handler Changes

**Goal:** Add `codecraftapi` to valid providers. Add 3 new handlers.

**File:** `backend-go/internal/admin/admin_handler.go`

**Checkpoint:** `go build ./...` passes. New endpoints return expected responses.

**Changes:**
1. Add `embedder ml.Embedder` field to `AdminHandler` struct
2. Update `NewAdminHandler()` signature to accept `ml.Embedder`
3. Update `IngestionPipeline` creation inside `NewAdminHandler` to use `embedder`
4. Add `"codecraftapi"` to `validProviders` map in `UpdateLLMSettings`
5. Add `"codecraftapi"` to `available_providers` list in `GetLLMSettings`
6. Add `codecraftapi_model_cheap/strong/fast` to `GetLLMSettings` response
7. Add `codecraftapi_model_cheap/strong/fast` handling to `UpdateLLMSettings`
8. Add `embedding_provider` and `embedding_model` to `GetLLMSettings` response
9. New handler: `GetCodeCraftModels` — proxies to CodeCraftAPI `/v1/models`
10. New handler: `GetEmbeddingSettings` — reads embedding config from system_settings
11. New handler: `UpdateEmbeddingSettings` — writes embedding config to system_settings

**Mental execution — GetCodeCraftModels:**
```
Happy path:
  1. Read llm_api_keys["codecraftapi"] from system_settings
  2. If empty → 400 "CodeCraftAPI key not configured"
  3. GET https://codecraftapi.com/v1/models
     Authorization: Bearer cc_xxx
  4. Parse model list
  5. Return to admin UI

Edge case — key not configured:
  → 400 {code: "KEY_NOT_CONFIGURED", message: "Add CodeCraftAPI key in LLM Settings first"}

Edge case — CodeCraftAPI /v1/models down:
  → HTTP error from CodeCraftAPI
  → 502 {code: "UPSTREAM_ERROR", message: "Could not fetch models from CodeCraftAPI"}
```

**Mental execution — UpdateEmbeddingSettings:**
```
Happy path (switch to codecraftapi):
  Input: {embedding_provider: "codecraftapi", embedding_model: "cc-embed-X"}
  1. Validate: provider must be "sidecar" or "codecraftapi"
  2. If provider=="codecraftapi" and model=="" → 400 "embedding_model required"
  3. UPSERT system_settings (key='embedding_provider', value='"codecraftapi"')
  4. UPSERT system_settings (key='embedding_model', value='"cc-embed-X"')
  5. Return {status: "updated"}

Happy path (switch back to sidecar):
  Input: {embedding_provider: "sidecar"}
  1. Validate: "sidecar" is valid
  2. UPSERT system_settings (key='embedding_provider', value='"sidecar"')
  3. Return {status: "updated"}
```

| Component | File | Status | What it does |
|---|---|---|---|
| AdminHandler.embedder field | `admin/admin_handler.go` | ⏳ | Holds Embedder for IngestionPipeline |
| NewAdminHandler signature | `admin/admin_handler.go` | ⏳ | Accepts ml.Embedder |
| validProviders + available_providers | `admin/admin_handler.go` | ⏳ | Add "codecraftapi" |
| GetCodeCraftModels handler | `admin/admin_handler.go` | ⏳ | Proxy to /v1/models |
| GetEmbeddingSettings handler | `admin/admin_handler.go` | ⏳ | Read embedding config |
| UpdateEmbeddingSettings handler | `admin/admin_handler.go` | ⏳ | Write embedding config |

---

### CC-6: main.go Wiring

**Goal:** Build `CodeCraftAPIEmbedder` + `DynamicEmbedder`. Pass to all constructors. Register 3 new routes.

**File:** `backend-go/cmd/server/main.go`

**Checkpoint:** `go build ./...` passes. Server starts.

**Changes:**
```go
// After mlClient initialization:
ccEmbedder := ml.NewCodeCraftAPIEmbedder(
    cfg.LLM.CodeCraftAPIKey,
    cfg.LLM.CodeCraftAPIBaseURL,
    logger,
)
embedder := ml.NewDynamicEmbedder(postgres.Pool, mlClient, ccEmbedder, logger)

// Update constructors:
contextAssembler := appcontext.NewAssembler(postgres.Pool, embedder, memManager, ...)
messageHandler := message.NewHandler(chatSvc, orch, modelGateway, embedder, memManager, logger)
adminHandler := adminpkg.NewAdminHandler(postgres.Pool, modelGateway, mlClient, embedder, categoryRegistry, domainRegistry, logger)

// New admin routes:
adminGroup.GET("/codecraftapi/models",  adminHandler.GetCodeCraftModels)
adminGroup.GET("/embedding-settings",  adminHandler.GetEmbeddingSettings)
adminGroup.POST("/embedding-settings", adminHandler.UpdateEmbeddingSettings)
```

**Note on memory.NewManager:** `NewManager` takes `*ml.SidecarClient` for L2Store.
L2Store uses `Embed()` for storing entries. This must change to `ml.Embedder`.
See CC-7 for the L2Store change. `NewManager` signature changes in CC-7.

| Component | File | Status | What it does |
|---|---|---|---|
| ccEmbedder construction | `cmd/server/main.go` | ⏳ | Build CodeCraftAPIEmbedder |
| DynamicEmbedder construction | `cmd/server/main.go` | ⏳ | Build DynamicEmbedder |
| Updated constructor calls | `cmd/server/main.go` | ⏳ | Pass embedder to all callers |
| 3 new admin routes | `cmd/server/main.go` | ⏳ | Register new endpoints |

---

### CC-7: Caller Signature Updates

**Goal:** Update all 5 constructors + their internal embedding calls to use `ml.Embedder`.

**CRITICAL:** All 5 files must be updated in ONE commit. Intermediate state does not compile.

**Files:**
- `backend-go/internal/training/ingestion_pipeline.go`
- `backend-go/internal/context/assembler.go`
- `backend-go/internal/memory/l2_store.go`
- `backend-go/internal/memory/manager.go`
- `backend-go/internal/message/handler.go`

**Checkpoint:** `go build ./...` passes with zero errors.

**Changes per file:**

`ingestion_pipeline.go`:
```go
// Before:
type IngestionPipeline struct { ml *ml.SidecarClient ... }
func NewIngestionPipeline(db, mlClient *ml.SidecarClient, gw, logger) *IngestionPipeline

// After:
type IngestionPipeline struct { embedder ml.Embedder ... }
func NewIngestionPipeline(db *pgxpool.Pool, embedder ml.Embedder, gw *gateway.ModelGateway, logger *zap.Logger) *IngestionPipeline
// All internal calls: p.ml.Embed() → p.embedder.Embed()
```

`assembler.go`:
```go
// Before:
type Assembler struct { ml *ml.SidecarClient ... }
func NewAssembler(db, mlClient *ml.SidecarClient, ...) *Assembler

// After:
type Assembler struct { embedder ml.Embedder ... }
func NewAssembler(db *pgxpool.Pool, embedder ml.Embedder, ...) *Assembler
// All internal calls: a.ml.Embed() → a.embedder.Embed()
// All internal calls: a.ml.EmbedSingle() → a.embedder.EmbedSingle()
```

`l2_store.go`:
```go
// Before:
type L2Store struct { ml *ml.SidecarClient ... }
func NewL2Store(db, mlClient *ml.SidecarClient, logger) *L2Store

// After:
type L2Store struct { embedder ml.Embedder ... }
func NewL2Store(db *pgxpool.Pool, embedder ml.Embedder, logger *zap.Logger) *L2Store
// All internal calls: s.ml.Embed() → s.embedder.Embed()
```

`manager.go`:
```go
// Before:
func NewManager(db, redisClient, mlClient *ml.SidecarClient, logger) *Manager

// After:
func NewManager(db *pgxpool.Pool, redisClient *redis.Client, embedder ml.Embedder, logger *zap.Logger) *Manager
// Pass embedder to NewL2Store()
```

`message/handler.go`:
```go
// Before:
type Handler struct { mlClient *ml.SidecarClient ... }
func NewHandler(chatSvc, orch, gw, mlClient *ml.SidecarClient, memManager, logger) *Handler

// After:
type Handler struct { embedder ml.Embedder ... }
func NewHandler(chatSvc *chat.Service, orch *orchestrator.Orchestrator, gw *gateway.ModelGateway, embedder ml.Embedder, memManager *memory.Manager, logger *zap.Logger) *Handler
// Internal call: h.mlClient.EmbedSingle() → h.embedder.EmbedSingle()
```

| Component | File | Status | What changes |
|---|---|---|---|
| IngestionPipeline | `training/ingestion_pipeline.go` | ⏳ | *SidecarClient → ml.Embedder |
| Assembler | `context/assembler.go` | ⏳ | *SidecarClient → ml.Embedder |
| L2Store | `memory/l2_store.go` | ⏳ | *SidecarClient → ml.Embedder |
| Manager | `memory/manager.go` | ⏳ | *SidecarClient → ml.Embedder |
| message.Handler | `message/handler.go` | ⏳ | *SidecarClient → ml.Embedder |

---

### CC-8: Frontend Types + API

**Goal:** Add `CodeCraftModel` type. Add 3 new API functions. Extend `updateLLMSettings`.

**Files:** `frontend/src/types/codecraftapi.ts`, `frontend/src/api/admin.ts`

**Checkpoint:** `npm run typecheck` passes.

**`codecraftapi.ts`:**
```typescript
export interface CodeCraftModel {
  id: string
  name?: string
  description?: string
  // Additional fields from /v1/models response — optional, shape unknown until tested
  [key: string]: unknown
}
```

**`admin.ts` additions:**
```typescript
// Import new type
import type { CodeCraftModel } from '@/types/codecraftapi'

// New functions:
export const getCodeCraftModels = () =>
  baseAPI.get<ApiResponse<CodeCraftModel[]>>('/api/v1/admin/codecraftapi/models')
    .then(res => res.data.data!)

export interface EmbeddingSettings {
  embeddingProvider: 'sidecar' | 'codecraftapi'
  embeddingModel: string
  availableProviders: string[]
  note: string
}

export const getEmbeddingSettings = () =>
  baseAPI.get<ApiResponse<EmbeddingSettings>>('/api/v1/admin/embedding-settings')
    .then(res => res.data.data!)

export interface UpdateEmbeddingSettingsRequest {
  embeddingProvider: 'sidecar' | 'codecraftapi'
  embeddingModel?: string
}

export const updateEmbeddingSettings = (req: UpdateEmbeddingSettingsRequest) =>
  baseAPI.post<ApiResponse<{ status: string }>>('/api/v1/admin/embedding-settings', req)
    .then(res => res.data.data!)
```

**`admin.ts` modification — extend `updateLLMSettings` request type:**
```typescript
export const updateLLMSettings = (req: {
  provider: string
  apiKeys: Record<string, string>
  // NEW: per-tier model names for CodeCraftAPI
  codecraftapiModelCheap?: string
  codecraftapiModelStrong?: string
  codecraftapiModelFast?: string
}) => ...
```

| Component | File | Status | What it does |
|---|---|---|---|
| CodeCraftModel type | `types/codecraftapi.ts` | ⏳ | Model catalog item type |
| getCodeCraftModels | `api/admin.ts` | ⏳ | Fetch model list from backend proxy |
| EmbeddingSettings type | `api/admin.ts` | ⏳ | Embedding config response type |
| getEmbeddingSettings | `api/admin.ts` | ⏳ | Read embedding config |
| updateEmbeddingSettings | `api/admin.ts` | ⏳ | Write embedding config |
| updateLLMSettings extended | `api/admin.ts` | ⏳ | Add codecraftapi model name fields |

---

### CC-9: Admin UI Changes

**Goal:** Add CodeCraftAPI provider card + model picker + embedding settings section.

**File:** `frontend/src/pages/admin/AdminLLMSettings.tsx`

**Checkpoint:** Admin can select CodeCraftAPI, fetch models, pick per-tier models, configure embeddings, save.

**Changes:**
1. Add CodeCraftAPI entry to `PROVIDERS` array
2. Add `useQuery` for `getCodeCraftModels` (only when `activeProvider === 'codecraftapi'`)
3. Add model picker section (3 dropdowns: cheap/strong/fast) — shown only when codecraftapi selected
4. Add `useQuery` for `getEmbeddingSettings`
5. Add embedding settings section (toggle + model picker + warning)
6. Update `mutation.mutationFn` to include model names in `updateLLMSettings` call
7. Add separate mutation for `updateEmbeddingSettings`

**Anti-patterns to watch:**
- DO NOT call `getCodeCraftModels` on every render — only when provider = codecraftapi AND user clicks "Fetch Models"
- DO NOT allow saving embedding_provider=codecraftapi without embedding_model — validate before submit
- DO NOT hide the warning about re-ingestion — it must be visible before admin can save
- DO NOT use `queryClient.invalidateQueries` for loader-sourced data — use `useRevalidator().revalidate()` if needed

| Component | File | Status | What it does |
|---|---|---|---|
| CodeCraftAPI provider card | `AdminLLMSettings.tsx` | ⏳ | 5th provider option |
| Model picker section | `AdminLLMSettings.tsx` | ⏳ | Per-tier model dropdowns |
| Embedding settings section | `AdminLLMSettings.tsx` | ⏳ | Toggle + model + warning |

---

## Anti-Patterns Checklist (Run on Every Function)

- [ ] **Reranking:** Never add `Rerank()` to `Embedder` interface
- [ ] **Model names:** Never hardcode in `CodeCraftAPIProvider` — always from constructor
- [ ] **Base URL:** Never hardcode in `CodeCraftAPIProvider` — always from constructor
- [ ] **Response parsing:** Never write new parser — reuse `doOpenAICompatibleCall()`
- [ ] **Cost:** Never fabricate `CostPer1K()` — return `0, 0`
- [ ] **API key:** Never expose in frontend — always proxy via backend
- [ ] **SidecarClient:** Never modify struct — wrap it
- [ ] **Vectors:** Never use string interpolation — always `pgvector.NewVector()` as `$N` param
- [ ] **Embedding switch warning:** Always show before admin saves embedding provider change
- [ ] **Caller updates:** All 5 constructors in ONE commit — never split
- [ ] **Async errors:** Never empty catch — log or rethrow
- [ ] **DB fields:** Always verify field names against schema before using
- [ ] **Signature changes:** Always grep all callers before committing

---

## Checkpoint Conditions

| Phase | How to verify |
|---|---|
| CC-1 | `go build ./...` passes. Admin sets provider=codecraftapi, sends a message, gets response. |
| CC-2 | `go build ./...` passes. Interface compiles. |
| CC-3 | `go build ./...` passes. Admin enables codecraftapi embeddings, uploads transcript, chunks get embedded. |
| CC-4 | `go build ./...` passes. |
| CC-5 | `go build ./...` passes. `GET /admin/codecraftapi/models` returns model list. |
| CC-6 | `go build ./...` passes. Server starts. All 3 new routes registered. |
| CC-7 | `go build ./...` passes with zero errors. Compiler confirms all callers updated. |
| CC-8 | `npm run typecheck` passes. |
| CC-9 | Admin can: select codecraftapi → fetch models → pick per-tier models → save → send message → get response. |

---

## Definition of Done (Per Component)

- [ ] Code written
- [ ] Mentally executed: inputs → outputs, happy path + at least 2 edge cases
- [ ] Anti-pattern checklist passed
- [ ] All callers updated (compiler confirms for Go, typecheck for TS)
- [ ] Types consistent end-to-end
- [ ] Committed to main
- [ ] File verified on branch (tree listing done)
- [ ] This handoff file updated with ✅ status + evidence
- [ ] No gaps left unresolved
- [ ] Checkpoint condition met

---

## Known Gaps / Future Work

| Gap | Impact | When to fix |
|---|---|---|
| `CostPer1K()` returns `0, 0` | Cost tracking shows $0 for CodeCraftAPI calls | When CodeCraftAPI pricing is known |
| No automatic failover between providers | If CodeCraftAPI is down, admin must manually switch | Future enhancement |
| Embedding dimension unknown | CodeCraftAPI embedding dimension may differ from sidecar's 768D | Verify after first test with real API key |
| `/v1/models` response shape unknown | `CodeCraftModel` type uses `[key: string]: unknown` | Verify after first API call, tighten type |
| No streaming LLM responses | Responses pop in all at once | Separate feature, not in this scope |

---

## Change Log

| Date | Change | Author |
|---|---|---|
| 2026-09-12 | Initial design doc + handoff file created | System Design Architect |
