# CodeCraftAPI Integration — System Design Document

> **Status:** DESIGN COMPLETE — PENDING IMPLEMENTATION  
> **Author:** System Design Architect  
> **Date:** 2026-09-12  
> **Branch:** main  
> **Depends on:** `AI_AVENGERS_SYSTEM_ARCHITECTURE.md`, `HANDOFF.md`, `IMPLEMENTATION_HANDOFF.md`

---

## Table of Contents

1. [What Is CodeCraftAPI](#1-what-is-codecraftapi)
2. [R&D Findings](#2-rd-findings)
3. [Recursive Self Cross-Questioning](#3-recursive-self-cross-questioning)
4. [Architecture — Before vs After](#4-architecture--before-vs-after)
5. [Component Specifications](#5-component-specifications)
6. [Files Touched — Complete List](#6-files-touched--complete-list)
7. [Caller Signature Change Map](#7-caller-signature-change-map)
8. [system_settings Keys](#8-system_settings-keys)
9. [Design Patterns Applied](#9-design-patterns-applied)
10. [SOLID Principles Applied](#10-solid-principles-applied)
11. [Failure Scenarios](#11-failure-scenarios)
12. [Quantitative Analysis](#12-quantitative-analysis)
13. [Locked Decisions](#13-locked-decisions)
14. [Admin UI Design](#14-admin-ui-design)
15. [Out of Scope](#15-out-of-scope)

---

## 1. What Is CodeCraftAPI

CodeCraftAPI is a third-party AI model aggregator — similar to OpenRouter — that provides
access to multiple AI models through a single API key.

```
Base URL:    https://codecraftapi.com/v1
Auth:        Authorization: Bearer cc_your_key   (key prefix: cc_)
Format:      OpenAI-compatible (same wire format as OpenRouter, DeepSeek, Gemini)

Endpoints:
  GET  /v1/models              → list all available models
  GET  /v1/models/{id}         → single model details
  POST /v1/chat/completions    → LLM calls (OpenAI chat format)
  POST /v1/embeddings          → embeddings (OpenAI embeddings format)
```

### Key Differentiator vs Other Providers

| Provider | LLM Calls | Embeddings | Reranking |
|---|---|---|---|
| OpenRouter | ✅ | ❌ | ❌ |
| DeepSeek | ✅ | ❌ | ❌ |
| Anthropic | ✅ | ❌ | ❌ |
| Gemini | ✅ | ❌ | ❌ |
| **CodeCraftAPI** | **✅** | **✅** | **❌** |
| Python Sidecar | ❌ | ✅ | ✅ |

CodeCraftAPI is the ONLY external provider that also offers embeddings.
This means it can optionally replace the Python sidecar for embedding generation.
Reranking (bge-reranker-base) has NO equivalent in CodeCraftAPI — sidecar stays for that.

---

## 2. R&D Findings

### 2.1 API Format Verification

CodeCraftAPI uses OpenAI-compatible format for both endpoints:

**Chat completions (same as OpenRouter/DeepSeek/Gemini):**
```json
POST /v1/chat/completions
Authorization: Bearer cc_your_key
{
  "model": "model-id-from-catalog",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user",   "content": "..."}
  ],
  "max_tokens": 2000,
  "temperature": 0.3
}
```

Response: standard OpenAI `choices[0].message.content` + `usage.prompt_tokens` + `usage.completion_tokens`.
This means `doOpenAICompatibleCall()` in `providers/common.go` can be reused directly.
Zero new response-parsing code needed.

**Embeddings (OpenAI format):**
```json
POST /v1/embeddings
Authorization: Bearer cc_your_key
{
  "model": "embedding-model-id",
  "input": ["text one", "text two"]
}
```

Response:
```json
{
  "data": [
    {"embedding": [0.1, 0.2, ...], "index": 0},
    {"embedding": [0.3, 0.4, ...], "index": 1}
  ]
}
```

### 2.2 Model Catalog

CodeCraftAPI exposes a live model catalog via `GET /v1/models`.
This is different from other providers where model names are hardcoded.
Admin must be able to:
1. Fetch the live model list from the admin panel
2. Assign a model per tier (cheap / strong / fast)
3. Assign a model for embeddings

### 2.3 Impact on Existing Codebase

Embeddings currently come from ONE source: `ml/sidecar_client.go`.
Called from these files:
- `training/ingestion_pipeline.go` — Step 4: embed chunks during training
- `context/assembler.go` — embed query for semantic search
- `memory/l2_store.go` — embed L2 memory entries
- `message/handler.go` — embed turn summary for chat_index

All four must be updated to use an `Embedder` interface instead of `*ml.SidecarClient`.

---

## 3. Recursive Self Cross-Questioning

Every question answered before writing a single line of design.

### Q1: Does CodeCraftAPI's chat format match our existing `doOpenAICompatibleCall()`?
**A:** YES. OpenAI-compatible format = same request/response shape.
`doOpenAICompatibleCall()` in `providers/common.go` is reused directly.
Zero new parsing code needed for LLM calls.

### Q2: Does CodeCraftAPI's embeddings format match OpenAI's?
**A:** YES. Standard OpenAI embeddings format:
`POST /v1/embeddings` with `{"model": "...", "input": ["text1", "text2"]}`
→ `{"data": [{"embedding": [...], "index": 0}]}`.
This is the industry standard. CodeCraftAPI follows it.

### Q3: Can we replace the Python sidecar entirely?
**A:** NO. The sidecar has TWO functions: `Embed()` AND `Rerank()`.
CodeCraftAPI has NO `/v1/rerank` endpoint.
Reranking (bge-reranker-base) is critical for China Wall Layer 1 (threshold 0.35).
Sidecar MUST stay for reranking. Embeddings can optionally come from CodeCraftAPI.

### Q4: Where are model names currently stored?
**A:** Hardcoded in each provider's `ModelName()` method.
For CodeCraftAPI, model names CANNOT be hardcoded — the catalog is dynamic.
Admin must configure per-tier model names. Storage: `system_settings` table
(new rows, no migration needed). Same pattern as existing `llm_provider` key.

### Q5: What is the exact `system_settings` structure for LLM?
**A:** Two existing keys:
- `llm_provider` → `"openrouter"` (string JSON)
- `llm_api_keys` → `{"openrouter": "sk-...", "deepseek": "sk-..."}` (map JSON)

CodeCraftAPI key stored as `{"codecraftapi": "cc_..."}` in the same map.
New keys for model names and embedding config (see Section 8).

### Q6: Does `available_providers` hardcoded list in `GetLLMSettings` need updating?
**A:** YES. Currently: `["openrouter", "deepseek", "anthropic", "gemini"]`.
Must become: `["openrouter", "deepseek", "anthropic", "gemini", "codecraftapi"]`.

### Q7: Does `validProviders` map in `UpdateLLMSettings` need updating?
**A:** YES. Same list, same change. Both must be updated in the same commit.

### Q8: What SOLID principles apply?
**A:**
- **S:** `CodeCraftAPIProvider` handles LLM only. `CodeCraftAPIEmbedder` handles embeddings only.
- **O:** `LLMProvider` interface already open for extension. New implementation, no interface change.
- **L:** `CodeCraftAPIProvider` satisfies all 5 `LLMProvider` methods completely.
- **I:** Extract `Embedder` interface — callers needing only embeddings don't depend on `Rerank()`.
- **D:** `IngestionPipeline`, `ContextAssembler`, `L2Store`, `MessageHandler` depend on `Embedder` interface, not `*SidecarClient` concrete type.

### Q9: What design patterns apply?
**A:**
- **Strategy:** `Embedder` interface with two strategies: `SidecarEmbedder` and `CodeCraftAPIEmbedder`.
- **Factory:** `NewDynamicEmbedder()` returns the right implementation based on DB config.
- **Adapter:** `CodeCraftAPIProvider` adapts CodeCraftAPI's API to our `LLMProvider` interface.
- **Proxy:** `GET /admin/codecraftapi/models` proxies to CodeCraftAPI — admin UI never calls CodeCraftAPI directly.

### Q10: What are the failure scenarios?
**A:**
1. CodeCraftAPI LLM down → existing 3-attempt retry + exponential backoff handles it.
2. CodeCraftAPI embeddings fail during training → ingestion job fails → resume-capable checkpoint handles it.
3. Admin sets wrong model name → CodeCraftAPI returns 400 → logged, surfaced to admin.
4. API key wrong → 401 from CodeCraftAPI → logged, surfaced to admin.
5. `DynamicEmbedder` DB read fails → falls back to sidecar (safe default).

### Q11: What about cost tracking?
**A:** `CostPer1K()` returns `0, 0` for CodeCraftAPI.
CodeCraftAPI pricing is unknown at design time.
Honest zero is better than fabricated numbers.
Admin checks their CodeCraftAPI dashboard for actual costs.

### Q12: Does `message/handler.go` need updating?
**A:** YES — caught during mental execution.
`message/handler.go` calls `h.mlClient.EmbedSingle()` in `indexTurn()`.
This must change to `h.embedder.EmbedSingle()`.
This was NOT in the initial file list — added after mental execution caught it.

### Q13: Does `memory/manager.go` need updating?
**A:** YES. `NewManager()` currently takes `*ml.SidecarClient`.
After this change, it must take `ml.Embedder` and pass it to `L2Store`.

### Q14: Is a DB migration needed?
**A:** NO. All new config stored as new rows in existing `system_settings` table.
The table already has `key VARCHAR(255) PRIMARY KEY, value JSONB` structure.
New rows = no schema change = no migration file.

### Q15: What happens if admin switches embedding provider mid-project?
**A:** Existing vectors in `course_chunks` and `project_memory_l2` were generated
by the OLD provider. New embeddings from the NEW provider will be in a different
vector space. Cosine similarity between old and new vectors is meaningless.
Admin MUST re-ingest all transcripts after switching embedding provider.
UI must show a clear warning before allowing the switch.

### Q16: Does `SidecarClient` struct change?
**A:** NO. `SidecarClient` is unchanged. We wrap it with `SidecarEmbedder`.
All existing callers of `SidecarClient.Rerank()` (chinawall/enforcer.go) are unaffected.

### Q17: Which files are verified safe (zero regression)?
**A:** Verified by reading each file's imports and method calls:
- `chinawall/enforcer.go` — uses `SidecarClient.Rerank()` only, not `Embed()` ✅
- `decision/engine.go` — no ML calls ✅
- `orchestrator/orchestrator.go` — no ML calls ✅
- `auth/`, `chat/`, `project/`, `rating/`, `repo/` — no ML calls ✅
- `ml/sidecar_client.go` — struct unchanged, `SidecarEmbedder` wraps it ✅

---

## 4. Architecture — Before vs After

### Before

```
ModelGateway
  ├── OpenRouterProvider   (LLMProvider interface)
  ├── DeepSeekProvider     (LLMProvider interface)
  ├── AnthropicProvider    (LLMProvider interface)
  └── GeminiProvider       (LLMProvider interface)

SidecarClient (concrete, no interface)
  ├── Embed()              ← ONLY embedding source in entire system
  ├── EmbedSingle()        ← ONLY embedding source in entire system
  └── Rerank()             ← ONLY reranking source (stays this way)

Callers of Embed/EmbedSingle (all take *ml.SidecarClient directly):
  ├── training/ingestion_pipeline.go
  ├── context/assembler.go
  ├── memory/l2_store.go
  └── message/handler.go
```

### After

```
ModelGateway
  ├── OpenRouterProvider       (LLMProvider interface)
  ├── DeepSeekProvider         (LLMProvider interface)
  ├── AnthropicProvider        (LLMProvider interface)
  ├── GeminiProvider           (LLMProvider interface)
  └── CodeCraftAPIProvider     (LLMProvider interface)  ← NEW

Embedder interface (NEW — ml/embedder.go)
  ├── SidecarEmbedder          (Embedder interface)  ← wraps SidecarClient.Embed/EmbedSingle
  └── CodeCraftAPIEmbedder     (Embedder interface)  ← NEW, calls /v1/embeddings

DynamicEmbedder (NEW — ml/embedder.go)
  └── reads embedding_provider from system_settings at call time
      → delegates to SidecarEmbedder OR CodeCraftAPIEmbedder

SidecarClient (UNCHANGED)
  ├── Embed()       → used by SidecarEmbedder
  ├── EmbedSingle() → used by SidecarEmbedder
  └── Rerank()      → still called directly by chinawall/enforcer.go (no interface)

Callers of Embed/EmbedSingle (all now take ml.Embedder interface):
  ├── training/ingestion_pipeline.go
  ├── context/assembler.go
  ├── memory/l2_store.go
  └── message/handler.go
```

### Data Flow — LLM Call with CodeCraftAPI

```
Admin sets provider=codecraftapi in LLM Settings panel
        │
        ▼
Admin sets model names: cheap=cc-model-A, strong=cc-model-B, fast=cc-model-C
        │
        ▼
Client sends message → orchestrator → decision engine → chinawall
        │
        ▼
chinawall.Enforce() calls gateway.Call(ctx, LLMRequest{Model: ModelStrong, ...})
        │
        ▼
ModelGateway.getProvider(ctx)
  → reads system_settings.llm_provider = "codecraftapi"
  → builds CodeCraftAPIProvider(apiKey, baseURL, cheapModel, strongModel, fastModel)
        │
        ▼
CodeCraftAPIProvider.Call(ctx, ProviderRequest{ModelTier: ModelStrong, ...})
  → POST https://codecraftapi.com/v1/chat/completions
    Authorization: Bearer cc_xxx
    {"model": "cc-model-B", "messages": [...], "max_tokens": 2000}
        │
        ▼
doOpenAICompatibleCall() parses response (reused from common.go)
        │
        ▼
ProviderResponse{Content: "...", InputTokens: 150, OutputTokens: 300}
```

### Data Flow — Embedding with CodeCraftAPI

```
Admin enables "Use CodeCraftAPI for embeddings" in LLM Settings
Admin selects embedding model: cc-embed-model-X
        │
        ▼
Admin uploads transcript → IngestTranscript() → Step 4: embed chunks
        │
        ▼
ingestion_pipeline.go calls embedder.Embed(ctx, chunkTexts)
        │
        ▼
DynamicEmbedder.Embed(ctx, texts)
  → reads system_settings.embedding_provider = "codecraftapi"
  → delegates to CodeCraftAPIEmbedder.Embed(ctx, texts)
        │
        ▼
CodeCraftAPIEmbedder.Embed(ctx, texts)
  → POST https://codecraftapi.com/v1/embeddings
    Authorization: Bearer cc_xxx
    {"model": "cc-embed-model-X", "input": ["chunk text 1", "chunk text 2", ...]}
        │
        ▼
Parse response: data[].embedding sorted by index
Return [][]float32 (768D or model-specific dimension)
        │
        ▼
Stored in course_chunks.embedding (pgvector column)
```

---

## 5. Component Specifications

### 5.1 `CodeCraftAPIProvider` — `providers/codecraftapi.go`

**Responsibility:** Implement `LLMProvider` interface for CodeCraftAPI LLM calls.

**Inputs:** API key (string), base URL (string), model names per tier (cheap/strong/fast strings), `*http.Client`

**Outputs:** `*ProviderResponse` (content, input tokens, output tokens, model used)

**Latency target:** 2–8 seconds (same as other providers)

**Failure behavior:** Returns error → `ModelGateway.Call()` retries up to 3 times with exponential backoff

**Locked decisions:**
- Model names come from constructor parameters (admin-configured), NOT hardcoded
- Auth header: `Authorization: Bearer cc_your_key` (same pattern as all other providers)
- Base URL: `https://codecraftapi.com/v1` (configurable via constructor, not hardcoded)
- Reuses `doOpenAICompatibleCall()` from `common.go` — zero new response-parsing code
- `CostPer1K()` returns `0, 0` — pricing unknown, honest default

**Mental execution — `Call()` happy path:**
```
Input: ModelStrong, system="You are a system design expert", user="How to shard?"
1. Build messages: [{role:system, content:"You are..."}, {role:user, content:"How to shard?"}]
2. Marshal body: {"model": "cc-model-strong", "messages": [...], "max_tokens": 2000, "temperature": 0.3}
3. POST https://codecraftapi.com/v1/chat/completions
   Header: Authorization: Bearer cc_xxx
   Header: Content-Type: application/json
4. doOpenAICompatibleCall() reads response
5. Return ProviderResponse{Content: "Use consistent hashing...", InputTokens: 150, OutputTokens: 300, ModelUsed: "cc-model-strong"}
```

**Mental execution — `Call()` failure path:**
```
Input: same as above
1–3. Same as above
4. HTTP 401 Unauthorized (wrong API key)
5. doOpenAICompatibleCall() returns error: "provider returned status 401"
6. ModelGateway.Call() logs warning, attempt 1 failed
7. Backoff 1s, retry attempt 2
8. Same 401 → attempt 2 failed
9. Backoff 2s, retry attempt 3
10. Same 401 → all attempts failed
11. Return nil, fmt.Errorf("all LLM attempts failed: provider returned status 401")
```

**Interface compliance check:**
```go
// All 5 methods must be implemented:
func (p *CodeCraftAPIProvider) Name() string                                          // "codecraftapi"
func (p *CodeCraftAPIProvider) ModelName(tier gtypes.ModelType) string               // per-tier model name
func (p *CodeCraftAPIProvider) CostPer1K(tier gtypes.ModelType) (float64, float64)   // 0, 0
func (p *CodeCraftAPIProvider) MaxTokens(tier gtypes.ModelType) int                  // 8192 default
func (p *CodeCraftAPIProvider) Call(ctx context.Context, req gtypes.ProviderRequest) (*gtypes.ProviderResponse, error)
```

---

### 5.2 `Embedder` Interface — `ml/embedder.go`

**Responsibility:** Abstract embedding source. Allows sidecar and CodeCraftAPI to be swapped without changing callers.

```go
type Embedder interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
    EmbedSingle(ctx context.Context, text string) ([]float32, error)
}
```

**Why NOT include `Rerank()` in this interface:**
Reranking is ALWAYS sidecar-only. CodeCraftAPI has no reranking endpoint.
Including `Rerank()` would force `CodeCraftAPIEmbedder` to implement a method
it cannot implement — violating Interface Segregation (SOLID-I).
Callers that need reranking (`chinawall/enforcer.go`) continue to take `*ml.SidecarClient` directly.

---

### 5.3 `SidecarEmbedder` — `ml/embedder.go`

**Responsibility:** Wrap `SidecarClient` to satisfy `Embedder` interface.

**Why a wrapper instead of making `SidecarClient` implement `Embedder` directly:**
`SidecarClient` already has `Embed()` and `EmbedSingle()` with the correct signatures.
It ALREADY satisfies the `Embedder` interface implicitly in Go (structural typing).
However, `SidecarClient` also has `Rerank()` and `HealthCheck()` — it is a larger type.
The wrapper makes the intent explicit: "this is the embedding-only view of the sidecar".
Alternatively, `SidecarClient` can directly satisfy `Embedder` without a wrapper.
Decision: use `SidecarClient` directly as `Embedder` (Go structural typing handles it).
No wrapper struct needed — simpler, less code, same result.

**Mental execution:**
```
DynamicEmbedder reads embedding_provider = "sidecar"
→ returns sidecarClient (which satisfies Embedder interface)
→ caller calls embedder.Embed(ctx, texts)
→ sidecarClient.Embed(ctx, texts) executes
→ POST http://localhost:8001/embed
→ returns [][]float32
```

---

### 5.4 `CodeCraftAPIEmbedder` — `ml/codecraftapi_embeddings.go`

**Responsibility:** Call CodeCraftAPI `/v1/embeddings` to generate embeddings.

**Inputs:** API key (string), base URL (string), `*pgxpool.Pool` (to read `embedding_model` from DB at call time), `*http.Client`

**Outputs:** `[][]float32` (dimension depends on model)

**Latency target:** 50–500ms per batch (unknown, assume similar to sidecar)

**Failure behavior:** Returns error → caller (ingestion pipeline) fails the job → resume-capable checkpoint handles recovery

**Locked decisions:**
- Timeout: 300s (same as sidecar — batch embedding can be slow)
- Sort response by `index` field before returning (defensive, OpenAI spec guarantees order but sort is correct)
- Model name is read from `system_settings` key `embedding_model` at CALL TIME (not stored in constructor)
  WHY: Same pattern as `ModelGateway.getActiveProvider()` and `DynamicEmbedder.Embed()`. Admin changes model
  name in UI → takes effect on next embed call, no restart needed.
- Constructor takes `db *pgxpool.Pool` NOT `modelName string`
- `*http.Client` is required in constructor (makes HTTP calls to CodeCraftAPI)

**Mental execution — `Embed()` happy path:**
```
Input: ["How to shard?", "Use consistent hashing"]
1. SELECT value FROM system_settings WHERE key='embedding_model' → "cc-embed-model-X"
2. If model empty → return nil, fmt.Errorf("codecraftapi: embedding_model not configured")
3. Marshal: {"model": "cc-embed-model-X", "input": ["How to shard?", "Use consistent hashing"]}
4. POST https://codecraftapi.com/v1/embeddings
   Header: Authorization: Bearer cc_xxx
5. Parse response:
   data[0] = {embedding: [0.1, 0.2, ...768 floats], index: 0}
   data[1] = {embedding: [0.3, 0.4, ...768 floats], index: 1}
6. Sort by index (defensive)
7. Return [[0.1, 0.2, ...], [0.3, 0.4, ...]]
```

**Mental execution — `Embed()` failure path:**
```
Input: ["How to shard?"]
1. Marshal body
2. POST → HTTP 400 Bad Request (wrong model name)
3. Return nil, fmt.Errorf("codecraftapi embeddings: status 400")
4. Caller (ingestion_pipeline.go) receives error
5. IngestTranscript() returns error, job status = "failed"
6. Admin sees failure in ingestion modal
7. Admin fixes model name in embedding settings
8. Admin resumes job from checkpoint
```

**Mental execution — `EmbedSingle()` happy path:**
```
Input: "How to shard?"
1. Call Embed(ctx, []string{"How to shard?"})
2. If len(result) == 0 → return nil, error
3. Return result[0]
```

---

### 5.5 `DynamicEmbedder` — `ml/embedder.go`

**Responsibility:** Read `embedding_provider` from `system_settings` at call time.
Delegate to `SidecarClient` or `CodeCraftAPIEmbedder` based on the setting.

**Why dynamic (reads DB at call time) instead of static (reads at startup):**
Same pattern as `ModelGateway.getActiveProvider()` — admin switches provider
from UI, takes effect on next call, no server restart needed.
DB read overhead: ~1ms (single SELECT, pgx pool). Acceptable at single-user scale.

**Failure behavior:** If DB read fails → falls back to sidecar (safe default).

**Mental execution — `Embed()` with provider switch:**
```
Admin changes embedding_provider from "sidecar" to "codecraftapi" in UI
→ POST /admin/embedding-settings {embedding_provider: "codecraftapi", embedding_model: "cc-embed-X"}
→ system_settings updated

Next call to embedder.Embed(ctx, texts):
1. DynamicEmbedder.Embed() called
2. SELECT value FROM system_settings WHERE key='embedding_provider'
3. Returns "codecraftapi"
4. Delegate to codecraftAPIEmbedder.Embed(ctx, texts)
5. Returns embeddings from CodeCraftAPI
```

---

### 5.6 `GetCodeCraftModels` Handler — `admin_handler.go`

**Route:** `GET /admin/codecraftapi/models`

**Responsibility:** Proxy to CodeCraftAPI `/v1/models`. Admin UI calls this to populate model picker dropdowns.

**Why proxy instead of direct frontend call:**
API key must never leave the server. Frontend never calls CodeCraftAPI directly.

**Mental execution — happy path:**
```
1. Read codecraftapi API key from system_settings (llm_api_keys["codecraftapi"])
2. If key empty → return 400 "CodeCraftAPI key not configured. Add key in LLM Settings first."
3. GET https://codecraftapi.com/v1/models
   Header: Authorization: Bearer cc_xxx
4. Parse response (model list)
5. Return model list to admin UI
```

**Mental execution — key not configured:**
```
1. Read key → empty string
2. Return 400 {code: "KEY_NOT_CONFIGURED", message: "CodeCraftAPI key not configured. Add key in LLM Settings first."}
3. Admin UI shows error: "Save your CodeCraftAPI key first, then fetch models."
```

---

### 5.7 `GetEmbeddingSettings` / `UpdateEmbeddingSettings` — `admin_handler.go`

**Routes:**
- `GET /admin/embedding-settings`
- `POST /admin/embedding-settings`

**Responsibility:** Read/write embedding provider config from `system_settings`.

**`GetEmbeddingSettings` response:**
```json
{
  "embedding_provider": "sidecar",
  "embedding_model": "",
  "available_providers": ["sidecar", "codecraftapi"],
  "note": "Changing embedding provider requires re-ingesting all transcripts."
}
```

**`UpdateEmbeddingSettings` request:**
```json
{
  "embedding_provider": "codecraftapi",
  "embedding_model": "cc-embed-model-X"
}
```

**Validation:**
- `embedding_provider` must be `"sidecar"` or `"codecraftapi"`
- If `embedding_provider = "codecraftapi"` and `embedding_model` is empty → return 400

---

### 5.8 `config.go` Changes

**Add constant:**
```go
const ProviderCodeCraftAPI LLMProvider = "codecraftapi"
```

**Add fields to `LLMConfig`:**
```go
CodeCraftAPIKey     string  // env: CODECRAFTAPI_KEY
CodeCraftAPIBaseURL string  // env: CODECRAFTAPI_BASE_URL (default: https://codecraftapi.com/v1)
```

**Add to `applyDefaults()`:**
```go
if c.LLM.CodeCraftAPIBaseURL == "" {
    c.LLM.CodeCraftAPIBaseURL = "https://codecraftapi.com/v1"
}
```

---

### 5.9 `model_gateway.go` Changes

**`buildProvider()` — add case:**
```go
case config.ProviderCodeCraftAPI:
    cheapModel  := g.getModelName(ctx, "codecraftapi_model_cheap",  "")
    strongModel := g.getModelName(ctx, "codecraftapi_model_strong", "")
    fastModel   := g.getModelName(ctx, "codecraftapi_model_fast",   "")
    return providers.NewCodeCraftAPIProvider(
        apiKey, g.cfg.CodeCraftAPIBaseURL,
        cheapModel, strongModel, fastModel,
        g.httpClient,
    )
```

**`getAPIKey()` — add case:**
```go
case config.ProviderCodeCraftAPI:
    return g.cfg.CodeCraftAPIKey
```

**New method `getModelName(ctx, settingKey, fallback string) string`:**
Reads from `system_settings` by key. Falls back to provided default.
Same pattern as `getActiveProvider()` and `getAPIKey()`.

---

### 5.10 `main.go` Changes

**New wiring for embedder:**
```go
// Build CodeCraftAPI embedder.
// Constructor takes db (to read embedding_model at call time) + httpClient + logger.
// Does NOT take model name — reads it from system_settings on every Embed() call.
// WHY: same pattern as ModelGateway.getActiveProvider() — no restart needed on model change.
ccEmbedder := ml.NewCodeCraftAPIEmbedder(
    cfg.LLM.CodeCraftAPIKey,
    cfg.LLM.CodeCraftAPIBaseURL,
    postgres.Pool,   // reads embedding_model from system_settings at call time
    &http.Client{Timeout: 300 * time.Second}, // 300s — batch embedding can be slow
    logger,
)

// DynamicEmbedder: reads embedding_provider from DB at call time.
// Falls back to mlClient (sidecar) when provider = "sidecar" or DB read fails.
embedder := ml.NewDynamicEmbedder(postgres.Pool, mlClient, ccEmbedder, logger)

// Pass embedder (not mlClient) to all embedding callers:
contextAssembler := appcontext.NewAssembler(postgres.Pool, embedder, memManager, ...)
messageHandler := message.NewHandler(chatSvc, orch, modelGateway, embedder, memManager, logger)
adminHandler := adminpkg.NewAdminHandler(postgres.Pool, modelGateway, mlClient, embedder, categoryRegistry, domainRegistry, logger)
// memory.NewManager also changes — see CC-7
```

**New admin routes:**
```go
adminGroup.GET("/codecraftapi/models",    adminHandler.GetCodeCraftModels)
adminGroup.GET("/embedding-settings",    adminHandler.GetEmbeddingSettings)
adminGroup.POST("/embedding-settings",   adminHandler.UpdateEmbeddingSettings)
```

**`AdminHandler` constructor change:**
```go
// AdminHandler needs embedder to create IngestionPipeline with correct embedding source
func NewAdminHandler(
    db *pgxpool.Pool,
    gw *gateway.ModelGateway,
    mlClient *ml.SidecarClient,
    embedder ml.Embedder,          // NEW parameter
    categoryReg *category.Registry,
    domainReg *chinawall.DomainRegistry,
    logger *zap.Logger,
) *AdminHandler
```

---

### 5.11 Frontend — `AdminLLMSettings.tsx` Changes

**Add to `PROVIDERS` array:**
```typescript
{
  value: 'codecraftapi',
  label: 'CodeCraftAPI (multi-model gateway)',
  desc: 'Single key, access to multiple AI models. Also supports embeddings.'
}
```

**Add model picker section (shown only when `codecraftapi` is selected):**
- Button: "Fetch Available Models" → calls `getCodeCraftModels()`
- Three `<select>` dropdowns populated from model list:
  - Cheap tier model (for tagging, metadata, coverage checks)
  - Strong tier model (for answer generation, charter extraction)
  - Fast tier model (for quick checks, simple tasks)
- Saved via `updateLLMSettings()` with new fields `codecraftapiModelCheap/Strong/Fast`

**Add embedding settings section (always visible, not just for codecraftapi):**
- Toggle: "Use CodeCraftAPI for embeddings" (default: off = sidecar)
- When toggled on:
  - Dropdown: embedding model (populated from same model list)
  - Warning banner: "⚠️ Changing embedding provider requires re-ingesting ALL transcripts. Existing vectors will be incompatible."
- Saved via `updateEmbeddingSettings()`

---

### 5.12 Frontend — `admin.ts` Changes

**New type:**
```typescript
export interface CodeCraftModel {
  id: string
  name?: string
  description?: string
}
```

**New functions:**
```typescript
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

**Modify `updateLLMSettings` request type:**
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

---

## 6. Files Touched — Complete List

### Files to CREATE

| File | What it does |
|---|---|
| `backend-go/internal/gateway/providers/codecraftapi.go` | `CodeCraftAPIProvider`: implements `LLMProvider` interface |
| `backend-go/internal/ml/embedder.go` | `Embedder` interface + `DynamicEmbedder` + `NewDynamicEmbedder` factory |
| `backend-go/internal/ml/codecraftapi_embeddings.go` | `CodeCraftAPIEmbedder`: calls `/v1/embeddings` |
| `frontend/src/types/codecraftapi.ts` | `CodeCraftModel` type |

### Files to MODIFY

| File | What changes |
|---|---|
| `backend-go/internal/config/config.go` | Add `ProviderCodeCraftAPI` constant + `CodeCraftAPIKey` + `CodeCraftAPIBaseURL` fields + default |
| `backend-go/internal/gateway/model_gateway.go` | Add `codecraftapi` case in `buildProvider()` + `getAPIKey()` + new `getModelName()` method |
| `backend-go/internal/admin/admin_handler.go` | Add `embedder ml.Embedder` field + `codecraftapi` to valid providers + 3 new handlers + `IngestionPipeline` uses embedder |
| `backend-go/cmd/server/main.go` | Build `ccEmbedder` + `DynamicEmbedder` + pass to constructors + 3 new admin routes + updated `NewAdminHandler` call |
| `backend-go/internal/training/ingestion_pipeline.go` | `*ml.SidecarClient` → `ml.Embedder` for embedding calls only |
| `backend-go/internal/context/assembler.go` | `*ml.SidecarClient` → `ml.Embedder` for embedding calls only |
| `backend-go/internal/memory/l2_store.go` | `*ml.SidecarClient` → `ml.Embedder` for embedding calls only |
| `backend-go/internal/memory/manager.go` | Accept `ml.Embedder` instead of `*ml.SidecarClient` |
| `backend-go/internal/message/handler.go` | `h.mlClient.EmbedSingle()` → `h.embedder.EmbedSingle()` |
| `frontend/src/api/admin.ts` | Add `getCodeCraftModels` + `getEmbeddingSettings` + `updateEmbeddingSettings` + extend `updateLLMSettings` request type |
| `frontend/src/pages/admin/AdminLLMSettings.tsx` | Add CodeCraftAPI provider card + model picker + embedding settings section |

### Files NOT Touched (Zero Regression — Verified)

| File | Why safe |
|---|---|
| `backend-go/internal/chinawall/enforcer.go` | Uses `SidecarClient.Rerank()` only — not in `Embedder` interface |
| `backend-go/internal/decision/engine.go` | No ML calls |
| `backend-go/internal/orchestrator/orchestrator.go` | No ML calls |
| `backend-go/internal/ml/sidecar_client.go` | Struct unchanged. Already satisfies `Embedder` interface via Go structural typing. |
| `backend-go/internal/memory/helpers.go` | Only JSON marshaling helper (`marshalJSON`). Zero ML calls. Verified by reading source. |
| `backend-go/internal/memory/l1_store.go` | Uses Redis only. `rebuildL1FromL2()` queries l2_store directly — no `Embed()` call. |
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

## 7. Caller Signature Change Map

Every constructor that changes signature. ALL must be updated in the SAME commit
as the interface creation. Compiler will catch any missed call site — let it go red.

| Constructor | Parameter that changes | From | To |
|---|---|---|---|
| `appcontext.NewAssembler` | 2nd param | `*ml.SidecarClient` | `ml.Embedder` |
| `training.NewIngestionPipeline` | 2nd param | `*ml.SidecarClient` | `ml.Embedder` |
| `message.NewHandler` | 4th param | `*ml.SidecarClient` | `ml.Embedder` |
| `admin.NewAdminHandler` | new 4th param | (did not exist) | `ml.Embedder` |
| `memory.NewManager` | 3rd param | `*ml.SidecarClient` | `ml.Embedder` |

**Rule:** Change all 5 constructors + their call sites in `main.go` in ONE commit.
Do NOT split across multiple commits — intermediate state will not compile.

---

## 8. system_settings Keys

No migration needed. New rows in existing `system_settings` table.

| Key | Value type | Default | Purpose |
|---|---|---|---|
| `llm_provider` | string JSON | `"openrouter"` | Active LLM provider (existing — add `"codecraftapi"` as valid value) |
| `llm_api_keys` | map JSON | `{}` | API keys per provider (existing — add `"codecraftapi"` key) |
| `codecraftapi_model_cheap` | string JSON | `""` | Model name for cheap tier (tagging, metadata) |
| `codecraftapi_model_strong` | string JSON | `""` | Model name for strong tier (answer generation) |
| `codecraftapi_model_fast` | string JSON | `""` | Model name for fast tier (quick checks) |
| `embedding_provider` | string JSON | `"sidecar"` | `"sidecar"` or `"codecraftapi"` |
| `embedding_model` | string JSON | `""` | Model name for embeddings (only used when `embedding_provider = "codecraftapi"`) |

---

## 9. Design Patterns Applied

### Strategy Pattern — Embedding Source

```
Embedder (interface)
    ├── SidecarClient     (strategy 1: local Python sidecar)
    └── CodeCraftAPIEmbedder (strategy 2: remote CodeCraftAPI)

DynamicEmbedder selects strategy at runtime based on system_settings.
```

### Factory Pattern — Embedder Creation

```
NewDynamicEmbedder(db, sidecarClient, ccEmbedder, logger) → Embedder
    Reads embedding_provider from DB at call time.
    Returns the correct strategy.
```

### Adapter Pattern — CodeCraftAPI → LLMProvider

```
CodeCraftAPIProvider adapts:
    CodeCraftAPI's OpenAI-compatible API
    → our internal LLMProvider interface

Same pattern as OpenRouterProvider, DeepSeekProvider, etc.
```

### Proxy Pattern — Model Catalog

```
Admin UI → GET /admin/codecraftapi/models (our backend)
         → GET https://codecraftapi.com/v1/models (CodeCraftAPI)

API key never leaves the server.
Admin UI never calls CodeCraftAPI directly.
```

---

## 10. SOLID Principles Applied

| Principle | How applied |
|---|---|
| **S — Single Responsibility** | `CodeCraftAPIProvider` handles LLM calls only. `CodeCraftAPIEmbedder` handles embeddings only. Two separate structs, two separate files. |
| **O — Open/Closed** | `LLMProvider` interface is open for extension. Adding `CodeCraftAPIProvider` requires zero changes to the interface or `ModelGateway` logic. |
| **L — Liskov Substitution** | `CodeCraftAPIProvider` satisfies all 5 `LLMProvider` methods. `CodeCraftAPIEmbedder` satisfies both `Embedder` methods. Any caller can use either without knowing the concrete type. |
| **I — Interface Segregation** | `Embedder` interface has only `Embed()` and `EmbedSingle()`. `Rerank()` is NOT included. Callers that only need embeddings don't depend on reranking. |
| **D — Dependency Inversion** | `IngestionPipeline`, `ContextAssembler`, `L2Store`, `MessageHandler` all depend on `ml.Embedder` interface, not `*ml.SidecarClient` concrete type. |

---

## 11. Failure Scenarios

| Scenario | What happens | Recovery |
|---|---|---|
| CodeCraftAPI LLM down | `Call()` returns error | `ModelGateway` retries 3× with exponential backoff |
| CodeCraftAPI embeddings fail during training | `Embed()` returns error | `IngestTranscript()` fails job. Resume-capable checkpoint. Admin resumes. |
| Wrong model name configured | CodeCraftAPI returns 400 | Logged. Admin sees error in ingestion modal or chat. Admin fixes model name. |
| Wrong API key | CodeCraftAPI returns 401 | Logged. Admin sees error. Admin fixes key in LLM Settings. |
| `DynamicEmbedder` DB read fails | Falls back to sidecar | Safe default. Logged as warning. |
| Admin switches embedding provider mid-project | Vectors become incompatible | UI shows warning before save. Admin must re-ingest all transcripts. |
| CodeCraftAPI `/v1/models` down | `GetCodeCraftModels` returns 502 | Admin sees error. Can still manually type model name. |

---

## 12. Quantitative Analysis

**Who triggers this:** Single admin user (Kiran). Not a high-frequency path.

**LLM call frequency:** On-demand per user message. Same as existing providers.

**Embedding frequency:**
- Training: batch of up to 1000 chunks × ~500 chars each = ~500KB per batch
- Context assembly: 1 query embed per turn = ~100 chars = negligible
- L2 memory: 1 embed per turn = ~200 chars = negligible

**`DynamicEmbedder` DB overhead:** 1 SELECT per `Embed()` call.
- Training: 1 SELECT per batch of 25 chunks = ~40 SELECTs per 1000-chunk transcript
- Context assembly: 1 SELECT per turn = negligible
- Total overhead: ~1ms per SELECT × 40 = 40ms per transcript ingestion = acceptable

**Memory:** No new in-memory state. `DynamicEmbedder` reads DB per call.

**Bottleneck:** None new. CodeCraftAPI latency is the same order as other providers.

---

## 13. Locked Decisions

| # | Decision | Reason | Exception |
|---|---|---|---|
| 1 | Reranking ALWAYS stays on Python sidecar | CodeCraftAPI has no `/v1/rerank`. bge-reranker-base is critical for China Wall Layer 1. | None |
| 2 | `Embedder` interface (SOLID-D) | Callers depend on abstraction, not implementation. | None |
| 3 | `DynamicEmbedder` reads DB at call time | Same pattern as `ModelGateway.getActiveProvider()`. No restart needed. | None |
| 4 | `CostPer1K()` returns `0, 0` | CodeCraftAPI pricing unknown. Honest zero > fabricated number. | Update when pricing is known |
| 5 | Model names stored in `system_settings` | No migration needed. Consistent with existing LLM settings pattern. | None |
| 6 | Admin UI proxies model list via backend | API key never leaves server. | None |
| 7 | Embedding provider change warning in UI | Changing provider makes existing vectors incompatible. Admin must re-ingest. | None |
| 8 | `SidecarClient` struct unchanged | Zero regression. Wrap it, never modify it. | None |
| 9 | All caller signature changes in ONE commit | Intermediate state does not compile. Compiler catches all missed call sites. | None |

---

## 14. Admin UI Design

### LLM Settings Page — Changes

```
┌─────────────────────────────────────────────────────────────┐
│  LLM Settings                                               │
│  Change provider or API keys without restarting the server. │
├─────────────────────────────────────────────────────────────┤
│  Active Provider                                            │
│  ○ OpenRouter (multi-model gateway)                         │
│  ○ DeepSeek (direct)                                        │
│  ○ Anthropic (direct)                                       │
│  ○ Google Gemini (direct)                                   │
│  ● CodeCraftAPI (multi-model gateway)  ← NEW                │
│    Single key, access to multiple AI models.                │
│    Also supports embeddings.                                │
├─────────────────────────────────────────────────────────────┤
│  API Keys                                                   │
│  [OpenRouter API Key    ] [password input]                  │
│  [DeepSeek API Key      ] [password input]                  │
│  [Anthropic API Key     ] [password input]                  │
│  [Google Gemini API Key ] [password input]                  │
│  [CodeCraftAPI Key      ] [password input]  ← NEW           │
├─────────────────────────────────────────────────────────────┤
│  CodeCraftAPI Model Selection  ← NEW (shown when selected)  │
│  [Fetch Available Models]                                   │
│                                                             │
│  Cheap tier (tagging, metadata):  [dropdown: model list]   │
│  Strong tier (answer generation): [dropdown: model list]   │
│  Fast tier (quick checks):        [dropdown: model list]   │
├─────────────────────────────────────────────────────────────┤
│  Embedding Settings  ← NEW (always visible)                 │
│  ○ Use Python sidecar (local, default)                      │
│  ○ Use CodeCraftAPI for embeddings                          │
│                                                             │
│  [when codecraftapi selected:]                              │
│  Embedding model: [dropdown: model list]                    │
│                                                             │
│  ⚠️  Changing embedding provider requires re-ingesting      │
│     ALL transcripts. Existing vectors will be incompatible. │
├─────────────────────────────────────────────────────────────┤
│  [Save Settings]                                            │
└─────────────────────────────────────────────────────────────┘
```

---

## 15. Out of Scope

- Streaming LLM responses (token-by-token) — not in CodeCraftAPI design, not in existing system
- CodeCraftAPI-specific features beyond chat/embeddings (e.g., fine-tuning, image generation)
- Automatic failover between providers (e.g., CodeCraftAPI down → fallback to OpenRouter)
- Cost tracking for CodeCraftAPI (pricing unknown at design time)
- Reranking via CodeCraftAPI (no endpoint exists)
- Multi-provider simultaneous use (one active provider at a time, same as existing design)
