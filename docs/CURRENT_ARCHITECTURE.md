# AI Avengers — Current Architecture (Verified from Code)

> Ye document poore repo ke actual code padh ke likha gaya hai (grep + read_file se
> verify karke), root ke `AI_AVENGERS_SYSTEM_ARCHITECTURE.md` (dated 2026-09-04,
> status "PENDING IMPLEMENTATION") jaisa pre-implementation design doc nahi hai.
> Jaha bhi purane doc se mismatch mila, actual code ko sahi maana gaya hai.
>
> Generated: 2026-09-19 | Source: gitlab:zepto-group3/ai_avengers @ main

---

## 1. Do Alag Systems — Confuse Mat Hona

Ye codebase mein **do bilkul alag features** hain jo dikhte similar hain par unka code separate hai:

| | **Chat Q&A** (`internal/orchestrator`) | **Build Workflow** (`internal/workflow`) |
|---|---|---|
| Kaam | Client expert se sawaal-jawab karta hai | Multiple experts milke real code likhte hain |
| Entry | `POST /chats/:id/messages` | `POST /workflows`, `/start`, `/run` |
| Experts kaise chalte hain | Parallel, **independent** (koi ek doosre ka output nahi dekhta) | Wave-by-wave, **collaborative** (Blackboard se ek doosre ka kaam padhte hain) |
| Output | Text answer + citations | Git commits + code files |
| UI | ChatPage | KanbanPage |

Code mein confirm: `orchestrator.go` ka doc-comment khud likhta hai — *"Multi-agent COLLABORATION... happens in the WORKFLOW flow, not here."* Dono packages ek doosre ko import nahi karte.

---

## 2. High-Level Diagram

```
                          ┌─────────────────────────┐
                          │   React Frontend (SPA)   │
                          │  Vite dev / nginx prod   │
                          └────────────┬─────────────┘
                                       │ HTTPS + SSE
                          ┌────────────▼─────────────┐
                          │   Go Backend (Gin)        │
                          │   :8080                   │
                          └──┬───────┬───────┬────────┘
                              │       │       │
                 ┌────────────┘       │       └─────────────┐
                 │                    │                     │
       ┌─────────▼─────────┐ ┌───────▼────────┐  ┌──────────▼─────────┐
       │  CHAT PATH         │ │  WORKFLOW PATH │  │  ADMIN PATH        │
       │  orchestrator      │ │  workflow      │  │  admin, training   │
       │  decision (5 gates)│ │  engine, DAG,  │  │  ingestion         │
       │  chinawall (4      │ │  Blackboard,   │  │                    │
       │  layers)           │ │  Aider         │  │                    │
       └─────────┬──────────┘ └───────┬────────┘  └──────────┬─────────┘
                 │                    │                       │
                 └───────────┬────────┴───────────┬───────────┘
                             │                    │
                   ┌─────────▼────────┐  ┌────────▼─────────┐
                   │   PostgreSQL      │  │   Redis           │
                   │   + pgvector      │  │   L1 cache,       │
                   │   (source of      │  │   pub/sub for     │
                   │   truth)          │  │   Blackboard SSE  │
                   └───────────────────┘  └───────────────────┘
                             │
                   ┌─────────▼─────────┐   ┌──────────────────┐
                   │  ml-sidecar        │   │  aider-service    │
                   │  (Python,          │   │  (Python,         │
                   │  embeddings +      │   │  FastAPI wrapper  │
                   │  reranking)        │   │  over Aider CLI)  │
                   └────────────────────┘   └───────────────────┘
```

---

## 3. Tech Stack (Actual, Verified)

| Layer | Tech | Notes |
|---|---|---|
| Backend | Go, Gin router | `backend-go/`, entrypoint `cmd/server/main.go` |
| DB | PostgreSQL + pgvector | 16 migrations (`backend-go/migrations/001`→`016`) |
| Cache/PubSub | Redis 7 | L1 memory, Blackboard SSE fan-out, refresh tokens |
| ML | Python FastAPI (`ml-sidecar/`) | embeddings + reranker (bge models), OR CodeCraftAPI as dynamic alternative |
| Code-gen agent | Python FastAPI wrapper over Aider (`aider-service/`) | routes LLM calls back through Go's `ModelGateway` proxy for centralized cost tracking |
| LLM providers | Multiple, pluggable | Anthropic, DeepSeek, Gemini, OpenRouter, CodeCraftAPI, **Cavoti** (OpenAI-compatible) |
| Frontend | React + Vite, TypeScript | `frontend/`, Zustand (client state) + TanStack Query (server state) |
| Orchestration | Docker Compose | `docker-compose.yml` (dev), `docker-compose.prod.yml` |

**Note on OpenAPI:** `backend-go/api/openapi.yaml` aur `internal/api/generated/` exist karte hain but **kabhi use nahi hue** — `generated/` mein sirf `.gitkeep` hai, koi Go file generate nahi hui, koi handler ise import nahi karta. Ye dead scaffolding hai, actual API contract nahi.

---

## 4. Go Backend — Package Map (`backend-go/internal/*`)

| Package | Kya karta hai |
|---|---|
| `orchestrator` | Chat Q&A engine — parallel expert responses, synthesis |
| `decision` | 5-gate decision engine (Gate 0-5, neeche detail mein) |
| `chinawall` | 4-layer citation/hallucination enforcement, domain-profile-driven |
| `workflow` | Multi-agent build-workflow engine (DAG, Blackboard, Aider, Kanban SSE) |
| `blackboard` | Shared event store — workflow experts ek doosre ka kaam yahi se padhte hain |
| `memory` | L1 (Redis)/L2 (Postgres)/L3 (append-only log) — chat context memory |
| `context` | `Assembler` — LLM call se pehle sab context sources parallel fetch karta hai |
| `category` | Expert ke structured-answer templates (sections: prose/code/test_cases) |
| `training` | Transcript ingestion: clean → chunk → topic-extract → charter-extract → embed → store |
| `repo` | GitHub/GitLab OAuth + repo code ko `repo_chunks` mein ingest karna |
| `chat`, `message` | Chat/message DB CRUD + SSE send handler |
| `expert`, `project` | Read-only expert listing, project CRUD |
| `admin` | Admin CRUD — experts, clients, categories, domain-profiles, LLM/embedding settings |
| `auth`, `middleware` | JWT auth, admin-role middleware |
| `gateway` | `ModelGateway` — sab LLM providers ka single entry point, cost tracking |
| `ml` | `Embedder` interface — ml-sidecar client ya CodeCraftAPI dynamic embedder |
| `ratelimit` | Token-bucket rate limiter (per-expert LLM call throttling) |
| `rating` | User rating → chunk boost_factor + expert avg_rating feedback loop |
| `selflearning` | User ke sawaal ko RAG-friendly tareeke se rewrite karta hai |
| `validation` | Go/TS code validation — workflow ke Aider-generated code ke liye |
| `monitoring` | LLM cost budget tracking/alerts |
| `observability` | Phase-wise latency timing utility |
| `config`, `db`, `response` | Config loading, Postgres pool, standard JSON responses |
| `api/generated` | **Unused stub** — sirf `.gitkeep` |

---

## 5. Feature A — Chat Q&A Flow (`internal/orchestrator`)

**Trigger:** `POST /api/v1/chats/:id/messages` → `message.Handler.Send` (SSE endpoint)

**Poora flow:**
```
Client message
   │
   ▼
message.Handler.Send  — SSE stream start
   │
   ▼
chat.Service.SaveMessage  — user message DB mein save
   │
   ▼
orchestrator.Orchestrator.Process
   │
   ├── loadExperts (DB se selected experts load)
   │
   └── HAR expert ke liye ek goroutine (parallel, independent):
       orchestrator.processWithExpert
          │
          ├── ratelimit.RateLimiter    — per-expert rate limit
          ├── context.Assembler.Assemble  — parallel fan-out:
          │     • rolling summary (10% token budget)
          │     • L2 project memory (20%)
          │     • recent messages (20%)
          │     • semantic chat history (15%)
          │     • course chunks — reranked (35%, sabse aakhir mein
          │       rakha jaata hai — "lost in the middle" se bachne ke liye)
          │     • repo chunks (agar repo connected hai)
          │     • reply-thread (agar user ne kisi purane message pe reply kiya)
          │
          ├── category.Registry lookup — agar expert ka structured template hai
          ├── selflearning.QuestionProcessor — sawaal ko RAG-friendly banata hai
          │
          └── decision.Engine.Process  — 5 GATES (neeche detail)
                 │
                 └── Gate 5 pe: chinawall.Enforcer.Enforce — 4 LAYERS (neeche)
   │
   ▼
Sab experts ke results collect (120s timeout)
   │
   ▼
2+ experts → orchestrator.synthesize()  — agreements/contradictions
   │
   ▼
SSE: thinking → chunk/complete (per expert) → synthesis → done
   │
   └── async: memory.Manager.RecordTurn (L1/L2 update, L3 always)
       async: chat.Service.IndexTurn (chat_index semantic search)
       async: rolling summary har 10 turns pe
```

### Decision Engine — 5 Gates (`internal/decision/engine.go`)

| Gate | Naam | Kya check karta hai | Fail hone pe mode |
|---|---|---|---|
| 0 | Structure Permission | (naya, purane doc mein nahi tha) Agar expert ki category `ask_structure_permission=true` hai, user se boilerplate preference poochta hai pehle | — (question ban jata hai) |
| 1 | Information Sufficiency | Keyword check + LLM double-check — kya sawaal ke liye enough info hai (problem-solving domains ke liye skip) | `ASK` |
| 2 | Knowledge Coverage | Koi chunk nahi mila, ya best rerank score < 0.20 | `REFUSE` |
| 3 | Charter Compliance | Expert ke `ReasoningCharter` mein "never X" rules regex-scan — violate hua to warning attach hoti hai, **pipeline ruk ta nahi** | `WARN` (continue) |
| 4 | Necessity | Over-engineering keywords (kubernetes, microservices, sharding etc.) + cheap-LLM check ki kya abhi zaroori hai | `PUSH_BACK` |
| 5 | Generate | China Wall se answer generate — 5 attempts tak retry | `ADVISE` / `REFUSE` |

### China Wall — 4 Layers (`internal/chinawall/enforcer.go`)

Purana doc mein 4 layers hardcoded booleans thi; **ab har layer `DomainProfile` (DB-backed) se driven hai** — har domain (medical/legal/DSA/system-design etc.) ka apna profile hai:

| Layer | Kaam | Domain-specific control |
|---|---|---|
| 1 | Reranker threshold reject | `DomainProfile.RerankerThreshold` (last retry pe relax hota hai) |
| 2 | Coverage check | `CoverageModeApplyPrinciples` (DSA — naye problem pe principle apply karo) vs `CoverageModeLiteralMatch` (medical/legal — chunks mein literally answer hona chahiye) |
| 3 | Generate with citations | Flat text (loose/strict citation mode) OR structured per-section (agar expert ki category ho) |
| 4 | Strip uncited claims | `CODE_EXEMPT` (fenced code kabhi strip nahi hota) vs `FULL_STRIP` (har uncited line strip) — agar sab strip ho jaye to `BaseProfile` se ek retry, phir refuse |

---

## 6. Feature B — Build Workflow Engine (`internal/workflow`)

**Trigger:** `POST /api/v1/workflows` (create) → `/start` → `/run`

**Files:** `engine.go` (state machine), `planner.go` (LLM se task-decomposition), `dag.go` (topological sort — **abhi fix hua bug yahi tha**), `runner.go` (wave-by-wave execution), `agent_loop.go` (design-phase experts), `aider_runner.go` (implementation/QA-phase experts, real git workspace), `workspace_merger.go` (per-expert workspaces ko `main/` mein merge), `cross_verifier.go` (reviewer-feedback revision loop), `kanban_sse.go` + `files_sse.go` (live UI streams), `checkpoint.go` (pod-restart resume).

**Phase state machine:** `intake → high_level_design → detailed_design → implementation → qa → handoff → completed`

**Poora flow:**
```
POST /workflows/:id/run
   │
   ▼
Runner.Run
   │
   ├── Step 1-2: DB se workflow + experts load, blackboard se requirement load
   │
   ├── Step 3: Planner.Plan  — LLM: requirement → ek task per expert (with deps)
   │
   ├── Step 4: BuildDAG  — Kahn's algorithm: tasks → parallel "waves"
   │              (Wave 0 = no-deps tasks, Wave 1 = depends-on-Wave-0, ...)
   │
   ├── Step 5: task_plan_ready event post (client ko approval ke liye)
   │
   ├── Step 6-7: AskClient approval → workflow pause karta hai
   │
   └── Step 8: executeWaves — HAR wave:
         │
         ├── Design phase → agent_loop.Run (Blackboard-based artifacts)
         │   Implementation/QA phase → aider_runner.Run:
         │      • per-expert isolated git workspace (/workspaces/{wf}/{expert}/)
         │      • seedWorkspace: blackboard design docs + main/ ka prior-wave code
         │      • Aider loop: observe (git status/build/test) → think (LLM) →
         │        act (patch apply)
         │      • publishCodeArtifacts: code_artifact_produced event
         │        (filename, content, operation: create/modify)
         │
         ├── (wave khatam) WorkspaceMerger.MergeWave → per-expert code ko
         │   main/ mein merge, conflict-check
         │   → wave_completed event post
         │
         └── cross_verifier.VerifyWaveArtifacts — reviewer matrix se code
             review; "changes_requested" pe runProducerRevision (Aider ko
             feedback ke saath phir se run karo, max 3 rounds)
   │
   ▼
Step 9-10: final approval → Engine.Complete
```

**Live UI (Kanban + Files):**
- `StreamKanban` (`kanban_sse.go`) — Blackboard events → task status/approval-gate SSE
- `StreamFiles` (`files_sse.go`, **abhi add kiya gaya**) — sirf `code_artifact_produced` + `wave_completed` events → frontend Files panel

---

## 7. Memory System (Chat ke liye, `internal/memory`)

| Layer | Storage | Scope | Kab likhta hai |
|---|---|---|---|
| L1 | Redis, TTL 24h | Per-expert, per-project, hot data | Har turn (agar importance ≥ 3) |
| L2 | Postgres (`project_memory_l2`) + pgvector | Sab experts ka combined project memory | Har turn (agar importance ≥ 3) |
| L3 | Postgres (`master_event_log`), append-only | Poora audit trail — kabhi delete/update nahi | **Hamesha**, har turn |

`memory.Manager` dono kaam karta hai: `GetProjectContext` (turn se pehle L1+L2 fetch) aur `RecordTurn` (turn ke baad async likhta hai, `orchestrator.updateMemory` se call hota hai).

---

## 8. Knowledge Ingestion (`internal/training`)

Admin transcript upload karta hai → `POST /admin/experts/:id/ingest` →
```
Load text → clean (noise strip ~45-55%) → chunk (dedup hash) →
topic extract (batched cheap-LLM) → charter extract (strong LLM,
reasoning + clarification rules) → embed (ml.Embedder) →
store in course_chunks (ON CONFLICT DO NOTHING dedup) →
capability table build → expert stats update
```
Resume-safe (`checkpoint.go` — interrupted ingestion resume kar sakta hai).

`internal/repo/` bilkul same `TextChunker` reuse karta hai, par GitHub/GitLab repo code ko `repo_chunks` (course_chunks se alag table) mein daalta hai — project-specific code context ke liye.

---

## 9. Frontend (`frontend/src`)

| Area | Detail |
|---|---|
| Pages | `auth/`, `chat/`, `projects/`, `experts/`, `workflows/` (WorkflowsPage, **KanbanPage**), `admin/` (8 sub-pages) |
| API clients | `frontend/src/api/*.ts` — backend ke har route-group ka 1:1 mirror (auth, chats, experts, projects, repo, workflows, admin, messages) |
| State | **Zustand** — client-only state (authStore, replyStore, streamStore, uiStore). Server data ke liye **TanStack Query** (alag concern, koi overlap nahi) |
| Auth | Access token sirf **in-memory** (Zustand), kabhi localStorage nahi (XSS-safe). Refresh token httpOnly cookie mein. Hard-reload pe `/auth/refresh` + `/auth/me` se restore |
| Live updates | Raw `EventSource` (SSE) — no wrapper library. Har stream (`useKanbanStream`, `useFileStream`) same pattern: `?token=` query-param auth, `camelizeKeys` snake↔camel conversion, 3s auto-reconnect |

---

## 10. Deployment (`docker-compose.yml`)

7 services: `migrate` (one-shot DB migration runner) → `postgres` (pgvector) → `redis` → `ml-sidecar` (Python, port 8001) → `aider-service` (Python, port 8082) → `api` (Go, port 8080) → `frontend` (Vite dev, port 3001).

`api` aur `aider-service` dono ek shared `workspaces` Docker volume mount karte hain — yehi wo directory hai jaha per-expert git workspaces (`internal/workflow/aider_runner.go`) bante hain.

---

## 10a. Aider ↔ ModelGateway Proxy Contract

`aider-service` kisi bhi model vendor se seedha baat nahi karta. Uske saare LLM
calls Go backend pe wapas aate hain, OpenAI chat-completions format mein.

**Routes** (dono ek hi handler — `internal/gateway/proxy.go`):

```
POST /api/v1/llm/proxy
POST /api/v1/llm/proxy/chat/completions
```

Doosra path zaroori hai kyunki litellm ke andar ka OpenAI client apne base URL
ke aage khud `/chat/completions` jodta hai. Aider ka `OPENAI_API_BASE` pehla
path hai, request doosre pe land karti hai.

**Auth:** shared secret (`AIDER_PROXY_TOKEN`), user JWT nahi —
`middleware.ServiceTokenMiddleware`. `aider-service` ek process hai, logged-in
user nahi; uske paas access token lene ka koi raasta nahi hai. Token
`Authorization: Bearer <token>` ke roop mein aata hai kyunki OpenAI client
library us path pe sirf `api_key` ko control karne deti hai. Token blank ho toh
route har call ko **503** deta hai — khaali secret ka matlab "auth off" nahi ho
sakta ek aise endpoint pe jo paisa kharch karta hai.

**Request:** poori conversation forward hoti hai (`LLMRequest.Messages`),
sirf last system + last user nahi. `model` field **ignore** hota hai — kaun sa
asli model chalega ye admin ki setting hai, aur gateway ka `strong` tier hamesha
use hota hai. `stream: true` aur `tools` explicit **400** dete hain (SSE aur
tool-calling provider contract mein nahi hain — chup-chaap galat jawab dene se
behtar hai saaf mana karna).

**Response:** poora OpenAI envelope — `id`, `object`, `created`, `model`,
`choices[].finish_reason`, aur `usage`. litellm isko OpenAI SDK ke
`ChatCompletion` model mein parse karta hai, jo in fields ke bina fail hota hai;
`usage` ke bina Aider ki har iteration free dikhti thi.

**Cost attribution:** `X-Workflow-ID` header ko UUID mein parse karke
`LLMRequest.WorkflowID` set hota hai, jisse gateway `workflows.cost_spent_usd`
badhata hai. Pehle ye header sirf log hota tha.

**Aider side** (`aider-service/main.py`) ke teen non-obvious points:

| Cheez | Kyun |
|---|---|
| Model name mein `openai/` prefix | Yehi litellm transport `OPENAI_API_BASE` maanta hai. Prefix hataya toh litellm vendor endpoint hardcode kar deta hai aur proxy bypass ho jaata hai. Prefix ke baad ka naam sirf label hai. |
| `OPENAI_API_BASE`/`OPENAI_API_KEY` env vars, constructor args nahi | litellm credentials sirf environment se padhta hai. Env var se weak model (commit message + history summary) bhi cover ho jaata hai, jo hum haath se banate hi nahi. |
| `GitRepo(io, fnames=[], git_dname=workspace)` explicitly | Coder khud banata hai toh `git_dname=None` deta hai, jo process ki cwd (`/app`) ban jaati hai — workflow ka workspace nahi. `chdir` option nahi hai: FastAPI sync endpoints thread-pool pe chalte hain, cwd global hai, concurrent tasks race karenge. |
| `stream=False` Coder pe | Coder default streaming hai, proxy ek complete JSON body deta hai. |
| `model.extra_params["extra_headers"]` | `send_completion` sirf `extra_params` ko litellm call mein merge karta hai. `model.extra_headers` set karna ek aisi attribute likhna hai jise koi padhta nahi. |
| HEAD before/after compare | Aider LLM/parse failures apne console pe likh ke normally return karta hai. Compare ke bina caller ko "success" milta tha jabki kuch commit hua hi nahi; ab `num_exhausted_context_windows` / `num_malformed_responses` se asli wajah report hoti hai. |

---

## 11. Recently Fixed Bugs (Is Session, Verified Against Code)

| Bug | File | Fix |
|---|---|---|
| Aider `Model(name="gpt-4", api_base=..., api_key=...)` | `aider-service/main.py` | Teen aise keyword args jo `Model.__init__` mein exist hi nahi karte — pehla Aider run turant crash hua. Positional model name + env-var credentials. |
| Aider `Coder.create(git_dname=...)` | `aider-service/main.py` | `Coder.__init__` mein `git_dname` nahi hai aur `**kwargs` bhi nahi — workspace `repo=GitRepo(...)` se jaata hai |
| `/llm/proxy` JWT-protected tha | `cmd/server/main.go` | `aider-service` JWT le hi nahi sakta → har Aider LLM call 401. Ab shared-secret middleware. |
| Proxy conversation collapse | `internal/gateway/proxy.go` | Sirf last system + last user bachte the, saare assistant turns gir jaate the. Ab poori conversation pass hoti hai (`LLMRequest.Messages`). |
| Anthropic multi-system overwrite | `providers/anthropic.go` | `sys = m.Content` assign tha — do system messages mein pehla (asli instructions) chup-chaap gir jaata tha, sirf trailing reminder bachta tha. Ab join hota hai. |
| Validation sandbox nested path | `internal/validation/sandbox.go` | `os.WriteFile` intermediate dirs nahi banata, toh `foundational/logger/logger.go` jaise har artifact ENOENT pe "validation failed" hota tha. `MkdirAll` add kiya. |
| Design artifacts ek doosre ko overwrite kar rahe the | `aider_runner.go` `artifactFilename` | event **type** → fixed filename tha, toh 11 artifacts sirf 4 files bane — 7 approved design decisions Aider tak pahunchne se pehle hi gayab. Ab `design/{seq}_{type}.{md\|yaml}`, ek file per artifact. |
| Design docs Aider ke chat mein khulte hi nahi the | `aider_runner.go` + `aider-service/main.py` | Aider sirf chat ki files padhta hai. Docs disk pe the par band — model se kaha ja raha tha ek design implement karo jo wo dekh hi nahi sakta. Ab `read_only_files` se jaate hain. |
| Implementation phase ke paas prompt hi nahi tha | `aider_runner.go` `buildImplementationPrompt` | `message = req.TaskDescription` — planner ka **design-phase** description tha ("System design for..."), toh model ne design likha, code nahi. QA ke paas `buildQAPrompt` tha, implementation ke paas kuch nahi. |
| Training khaali hone pe rules code likhne se rok dete the | `aider_runner.go` `build7030EnforcementInstructions` | "training is your ONLY source of truth" + zero chunks = model ka obedient jawab "koi code nahi". Ab approved design docs bhi named source hain (workflow ka apna output, generic knowledge nahi). |
| `max:5` jhooth tha | `aider_runner.go` `runAiderLoop` | koi bhi iteration error = poora task fail, iteration 2-5 kabhi nahi chalti. Ab retry signal hai; fail sirf tab jab 5 iterations ke baad ek bhi commit na ho. |
| Proxy `max_tokens` 4000 | `internal/gateway/proxy.go` | Aider `max_tokens` bhejta nahi, toh yehi default har code-gen call pe lagta tha — reply SEARCH/REPLACE block ke beech kat jaati. Ab 8000 (gateway provider cap pe clamp karta hai). |
| DAG false-cycle detection | `dag.go` | In-degree calculation reversed thi — har normal 2-task plan (ek task no-deps, doosra usi pe depend) false-cycle error deta tha |
| Self-dependency case-mismatch | `planner.go` | String-compare se UUID-compare mein badla (LLM case-mismatch UUID return karta tha) |
| Missing `operation` field | `aider_runner.go` | `code_artifact_produced` event mein create/modify tag add kiya |
| Missing `wave_completed` event | `runner.go` | Wave-merge ke baad event post nahi hota tha, ab hota hai |
| Kanban SSE stuck at CONNECTING | `kanban_sse.go` | Initial `": connected\n\n"` ping stream-loop se pehle likhne se fix hua |

---

## 12. Jo Doc Mein NAHI Hai (Explored Nahi Kiya Gaya)

`internal/auth`, `internal/gateway` (providers ke andar ka detail), `internal/expert`, `internal/project`, `internal/blackboard`, `internal/db`, `internal/config`, `internal/middleware`, `internal/ml`, `internal/response` — line-by-line nahi padhe gaye, sirf call-sites/naming se purpose inferred hai. Agar inka deep-dive chahiye, bata dena.
