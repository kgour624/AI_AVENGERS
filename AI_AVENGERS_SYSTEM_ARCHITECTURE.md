# AI Avengers — System Architecture

> **Production-Grade Multi-Agent Platform**  
> Backend: Go | Frontend: React | ML Sidecar: Python | DB: PostgreSQL + pgvector + Redis

---

## Document Status

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | DESIGN COMPLETE — PENDING IMPLEMENTATION |
| Author | System Design Architect |
| Last Updated | 2026-09-04 |
| Target Repo | gitlab.com/zepto-group3/ai_avengers |
| Branch | main |

---

## Table of Contents

1. [Product Vision](#1-product-vision)
2. [Architecture Overview](#2-architecture-overview)
3. [Technology Stack — Decisions with Reasoning](#3-technology-stack)
4. [System Components](#4-system-components)
5. [Database Schema](#5-database-schema)
6. [Memory System — L1/L2/L3](#6-memory-system)
7. [Agent Orchestration Engine](#7-agent-orchestration-engine)
8. [China Wall Enforcement](#8-china-wall-enforcement)
9. [Context Budget Manager](#9-context-budget-manager)
10. [Decision Engine — 5 Gates](#10-decision-engine)
11. [Expert Training Pipeline](#11-expert-training-pipeline)
12. [GitHub/GitLab Integration](#12-githubgitlab-integration)
13. [Rating and Self-Learning System](#13-rating-and-self-learning-system)
14. [Admin Panel](#14-admin-panel)
15. [API Design](#15-api-design)
16. [Security Architecture](#16-security-architecture)
17. [Deployment Architecture](#17-deployment-architecture)
18. [Locked Decisions](#18-locked-decisions)
19. [Implementation Phases](#19-implementation-phases)

---

## 1. Product Vision

### What Is AI Avengers?

AI Avengers is NOT a general-purpose AI assistant. It is a **constrained domain-expert team simulator** where:

- Every expert is trained **exclusively** on specific course transcripts
- Every claim is **cited** to a source chunk — no hallucination possible
- Multiple experts **collaborate** on a single project simultaneously
- The system **remembers** everything across the project lifecycle
- Experts **know why** they are doing what they are doing (WHY principle)

### The Core Insight

> Generic AI tools (Claude, GPT, Gemini) know everything — which means they confidently say wrong things.
> AI Avengers experts know only what they were taught — which means every answer is accurate and cited.

This constraint IS the product. This is the moat.

### Who Uses This?

| Role | What They Do |
|---|---|
| Admin (Kiran Nogia) | Manages experts, uploads transcripts, monitors system |
| Client | Creates projects, selects experts, gets domain-accurate answers |
| Expert (AI Agent) | Answers questions strictly from transcript knowledge |

### Why 2050-Proof?

1. **Domain constraint compounds** — more transcripts = more value, not less
2. **Memory system improves** — L3 master table grows richer over time
3. **Rating system self-corrects** — bad patterns get flagged, good patterns reinforced
4. **Go architecture scales** — goroutines handle millions of concurrent requests
5. **WHY principle** — experts reason like humans, not like autocomplete

---

## 2. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CLIENT BROWSER                                │
│                    React Frontend (SPA)                              │
└─────────────────────────┬───────────────────────────────────────────┘
                          │ HTTPS / WebSocket
┌─────────────────────────▼───────────────────────────────────────────┐
│                     API GATEWAY (Go)                                 │
│              Rate Limiting | Auth | Request Routing                  │
└──────┬──────────────┬──────────────┬──────────────┬─────────────────┘
       │              │              │              │
┌──────▼──────┐ ┌─────▼──────┐ ┌───▼────────┐ ┌──▼──────────────┐
│   Auth      │ │  Project   │ │   Expert   │ │    Admin        │
│  Service    │ │  Service   │ │Orchestrator│ │   Service       │
│   (Go)      │ │   (Go)     │ │   (Go)     │ │    (Go)         │
└──────┬──────┘ └─────┬──────┘ └───┬────────┘ └──┬──────────────┘
       │              │            │              │
┌──────▼──────────────▼────────────▼──────────────▼─────────────────┐
│                    CORE SERVICES LAYER (Go)                         │
│                                                                      │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────────┐   │
│  │   Memory    │  │  China Wall  │  │   Context Budget        │   │
│  │  Manager   │  │  Enforcer    │  │   Manager               │   │
│  │ L1/L2/L3   │  │  4 Layers    │  │   Eviction+Summary      │   │
│  └─────────────┘  └──────────────┘  └─────────────────────────┘   │
│                                                                      │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────────┐   │
│  │  Decision   │  │   Context    │  │   Rating &              │   │
│  │  Engine     │  │  Assembler   │  │   Learning Engine       │   │
│  │  5 Gates    │  │  Smart RAG   │  │                         │   │
│  └─────────────┘  └──────────────┘  └─────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────────┐
│                    DATA LAYER                                        │
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │  PostgreSQL  │  │    Redis     │  │   Python ML Sidecar      │  │
│  │  + pgvector  │  │  L1 Cache    │  │   Embeddings + Reranker  │  │
│  │  L2/L3 Store │  │  Sessions    │  │   bge-base-en-v1.5       │  │
│  └──────────────┘  └──────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### Data Flow — Client Sends a Message

```
Client types requirement
        │
        ▼
API Gateway validates JWT
        │
        ▼
Expert Orchestrator receives request
        │
        ▼
Context Budget Manager checks token budget
        │
        ▼
Context Assembler builds context:
  ├── L2 Group Memory (project history)
  ├── Rolling Summary (last 10 turns)
  ├── Recent Messages (last 3)
  └── Relevant Course Chunks (semantic search)
        │
        ▼
Decision Engine — 5 Gates (per selected expert)
  Gate 1: Enough info? → ASK
  Gate 2: In domain? → REFUSE
  Gate 3: Charter violation? → WARN
  Gate 4: Necessary? → PUSH_BACK
  Gate 5: Generate → ADVISE
        │
        ▼
China Wall Enforcer — 4 Layers
  Layer 1: Reranker threshold (0.35)
  Layer 2: Coverage check
  Layer 3: Mandatory citations
  Layer 4: Strip uncited claims
        │
        ▼
Memory Manager updates:
  ├── L1: This expert's memory for this project
  ├── L2: Group memory update
  └── L3: Master event log append
        │
        ▼
Rating Engine records interaction
        │
        ▼
Response to Client (SSE streaming)
```

---

## 3. Technology Stack

### Decisions with Reasoning

#### Backend: Go 1.22+

**WHEN** handling concurrent expert requests  
**DO** use Go goroutines + channels  
**BECAUSE** Go's M:N threading model handles 10,000+ concurrent goroutines with ~2KB stack each. Python's GIL prevents true parallelism. Node.js single-threaded event loop is not suitable for CPU-bound tasks. When 5 experts process simultaneously, Go runs them truly parallel — Python would serialize them.  
**EXCEPT** ML inference (embeddings, reranking) — Python has mature libraries (sentence-transformers) with no stable Go equivalent.

**Libraries:**
```
github.com/gin-gonic/gin          — HTTP router (fast, production-proven)
github.com/jackc/pgx/v5           — PostgreSQL driver (pgvector support)
github.com/pgvector/pgvector-go   — Vector operations
github.com/redis/go-redis/v9      — Redis client
github.com/golang-jwt/jwt/v5      — JWT authentication
github.com/pquerna/otp            — TOTP for admin (Google Authenticator)
github.com/google/uuid            — UUID generation
golang.org/x/crypto               — bcrypt password hashing
github.com/spf13/viper            — Configuration management
go.uber.org/zap                   — Structured logging
github.com/golang-migrate/migrate — Database migrations
```

#### ML Sidecar: Python 3.11 + FastAPI

**WHEN** embedding generation or reranking is needed  
**DO** call Python ML sidecar via internal HTTP  
**BECAUSE** sentence-transformers, bge-base-en-v1.5, bge-reranker-base are Python-native. ONNX Go bindings exist but are unstable in production. ML sidecar is stateless — horizontally scalable.  
**EXCEPT** never call ML sidecar from hot path synchronously — always async with timeout.

**Models:**
```
BAAI/bge-base-en-v1.5    — 768D embeddings, local, zero cost
BAAI/bge-reranker-base   — Cross-encoder reranking, local, zero cost
```

#### Database: PostgreSQL 15+ with pgvector

**WHEN** storing structured data, vector embeddings, event logs  
**DO** use single PostgreSQL instance with pgvector extension  
**BECAUSE** pgvector supports IVFFlat and HNSW indexes for fast ANN search. Single DB reduces operational complexity. Row-level security for multi-tenancy. JSONB for flexible metadata.  
**EXCEPT** L1 memory (hot path, per-expert, per-project) goes to Redis for O(1) access.

#### Cache: Redis 7+

**WHEN** storing L1 expert memory, sessions, rate limit counters  
**DO** use Redis with TTL-based expiry  
**BECAUSE** O(1) lookup, sub-millisecond latency, built-in TTL. L1 memory is hot data — accessed on every turn.  
**EXCEPT** never store critical data only in Redis — always persist to PostgreSQL as source of truth.

#### LLM Gateway: OpenRouter

**WHEN** making LLM calls  
**DO** route through OpenRouter  
**BECAUSE** single API key, multiple model access, cost tracking, automatic failover.  
**Models:**
```
cheap  — deepseek/deepseek-chat        (tagging, coverage check, metadata)
strong — anthropic/claude-3-5-sonnet   (answer generation, charter extraction)
fast   — google/gemini-flash-1.5       (quick checks, simple tasks)
```

---

## 4. System Components

### 4.1 Expert Orchestrator

**Responsibility:** Coordinates multiple domain experts for a single client request. Routes to correct experts, runs them in parallel, synthesizes results.

**Inputs:** Client message, selected expert IDs, project ID, session token  
**Outputs:** Synthesized response with citations, mode badges, expert attributions  
**Latency target:** < 8 seconds for 3 parallel experts  
**Failure behavior:** If one expert fails, others continue. Partial response returned with failure noted.

```go
// internal/orchestrator/orchestrator.go

type ExpertOrchestrator struct {
    db             *pgx.Pool
    redis          *redis.Client
    memoryManager  *MemoryManager
    contextAssembler *ContextAssembler
    decisionEngine *DecisionEngine
    chinaWall      *ChinaWallEnforcer
    modelGateway   *ModelGateway
    mlSidecar      *MLSidecarClient
}

type OrchestratorRequest struct {
    ProjectID   uuid.UUID
    ClientID    uuid.UUID
    Message     string
    ExpertIDs   []uuid.UUID
    TurnNumber  int
    FileUploads []FileUpload // optional
}

type OrchestratorResponse struct {
    ExpertResponses []ExpertResponse
    Synthesis       *SynthesisResult // nil if single expert
    TurnNumber      int
    TokensUsed      int
    CostUSD         float64
}

type ExpertResponse struct {
    ExpertID    uuid.UUID
    ExpertName  string
    Domain      string
    Mode        ResponseMode // ASK, WARN, PUSH_BACK, REFUSE, ADVISE
    Content     string
    Citations   []Citation
    Confidence  float64
    GateStopped int // which gate stopped (0 = reached Gate 5)
    Warning     string
    Questions   []string // for ASK mode
}

func (o *ExpertOrchestrator) Process(ctx context.Context, req OrchestratorRequest) (*OrchestratorResponse, error) {
    // 1. Validate request
    // 2. Check context budget
    // 3. Assemble context (L2 memory + rolling summary + recent messages)
    // 4. Run experts in parallel via goroutines
    // 5. Collect results
    // 6. Synthesize if multiple experts
    // 7. Update memory (L1, L2, L3)
    // 8. Return response
    
    results := make(chan ExpertResult, len(req.ExpertIDs))
    
    for _, expertID := range req.ExpertIDs {
        go func(eid uuid.UUID) {
            result := o.processWithExpert(ctx, req, eid)
            results <- result
        }(expertID)
    }
    
    // Collect with timeout
    expertResults := o.collectResults(ctx, results, len(req.ExpertIDs), 10*time.Second)
    
    // Synthesize
    response := o.synthesize(expertResults, req)
    
    // Update memory async (non-blocking)
    go o.updateMemory(context.Background(), req, response)
    
    return response, nil
}
```

### 4.2 Model Gateway

**Responsibility:** Single interface for all LLM calls. Cost tracking, retry logic, caching.

```go
// internal/gateway/model_gateway.go

type ModelType string

const (
    ModelCheap  ModelType = "cheap"   // deepseek — tagging, metadata
    ModelStrong ModelType = "strong"  // claude   — answer generation
    ModelFast   ModelType = "fast"    // gemini   — quick checks
)

type ModelGateway struct {
    apiKey      string
    baseURL     string
    httpClient  *http.Client
    cache       *sync.Map // prompt hash → response
    totalCost   atomic.Float64
    callCount   atomic.Int64
}

type LLMRequest struct {
    Model       ModelType
    SystemPrompt string
    UserPrompt  string
    MaxTokens   int
    Temperature float64
    UseCache    bool
}

type LLMResponse struct {
    Content      string
    InputTokens  int
    OutputTokens int
    CostUSD      float64
    ModelUsed    string
    Cached       bool
}

func (g *ModelGateway) Call(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
    // Check cache
    if req.UseCache {
        if cached := g.checkCache(req); cached != nil {
            return cached, nil
        }
    }
    
    // Retry with exponential backoff
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        resp, err := g.callOpenRouter(ctx, req)
        if err == nil {
            g.updateCache(req, resp)
            g.totalCost.Add(resp.CostUSD)
            g.callCount.Add(1)
            return resp, nil
        }
        lastErr = err
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
    }
    
    return nil, fmt.Errorf("all LLM attempts failed: %w", lastErr)
}
```

### 4.3 ML Sidecar Client

**Responsibility:** Go client for Python ML service. Embeddings + reranking.

```go
// internal/ml/sidecar_client.go

type MLSidecarClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

type EmbedRequest struct {
    Texts []string `json:"texts"`
}

type EmbedResponse struct {
    Embeddings [][]float32 `json:"embeddings"` // 768D each
}

type RerankRequest struct {
    Query     string   `json:"query"`
    Documents []string `json:"documents"`
    TopK      int      `json:"top_k"`
}

type RerankResult struct {
    Index int     `json:"index"`
    Score float32 `json:"score"`
    Text  string  `json:"text"`
}

func (c *MLSidecarClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
    // POST /embed with timeout
    // Returns 768D float32 vectors
}

func (c *MLSidecarClient) Rerank(ctx context.Context, query string, docs []string, topK int) ([]RerankResult, error) {
    // POST /rerank with timeout
    // Returns sorted results with scores
}
```

---

## 5. Database Schema

### Design Principles

1. **UUID primary keys** — distributed-safe, no sequential ID leakage
2. **Soft deletes** — `deleted_at` timestamp, never hard delete
3. **Audit columns** — `created_at`, `updated_at` on every table
4. **Row-level security** — clients see only their data
5. **Event sourcing for L3** — append-only, immutable

### Tables

```sql
-- ============================================================
-- USERS & AUTH
-- ============================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    full_name       VARCHAR(255) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'client', -- 'admin' | 'client'
    is_active       BOOLEAN DEFAULT TRUE,
    totp_secret     VARCHAR(255),          -- for admin TOTP
    totp_enabled    BOOLEAN DEFAULT FALSE,
    preferences     JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- ============================================================
-- DOMAIN EXPERTS
-- ============================================================

CREATE TABLE experts (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                  VARCHAR(255) NOT NULL,   -- "Arpit Bhiyani — System Design"
    slug                  VARCHAR(255) UNIQUE NOT NULL, -- "arpit-system-design"
    domain                VARCHAR(100) NOT NULL,   -- "system_design"
    description           TEXT,
    avatar_url            VARCHAR(500),
    
    -- Charter (extracted from transcripts)
    reasoning_charter     TEXT NOT NULL,           -- "Never recommend microservices for team < 10"
    clarification_charter JSONB NOT NULL DEFAULT '{}', -- topic → [questions]
    
    -- Capability
    capability_summary    JSONB DEFAULT '{}',       -- topic → depth_level
    
    -- Stats
    total_chunks          INTEGER DEFAULT 0,
    total_topics          INTEGER DEFAULT 0,
    avg_depth_level       DECIMAL(3,2),
    avg_rating            DECIMAL(3,2) DEFAULT 0,
    total_ratings         INTEGER DEFAULT 0,
    
    -- Status
    is_active             BOOLEAN DEFAULT TRUE,
    is_training           BOOLEAN DEFAULT FALSE,   -- locked during ingestion
    
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_experts_slug ON experts(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_experts_domain ON experts(domain) WHERE is_active = TRUE;

-- ============================================================
-- COURSE CHUNKS (Expert Knowledge Base)
-- ============================================================

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE course_chunks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id        UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    
    -- Content
    chunk_text       TEXT NOT NULL,
    chunk_index      INTEGER NOT NULL,
    
    -- Metadata
    topic            VARCHAR(255),
    subtopic         VARCHAR(255),
    source_file      VARCHAR(500),   -- transcript filename
    
    -- Vector (768D from bge-base-en-v1.5)
    embedding        vector(768) NOT NULL,
    
    -- Navigation
    prev_chunk_id    UUID REFERENCES course_chunks(id),
    next_chunk_id    UUID REFERENCES course_chunks(id),
    
    -- Performance tracking
    times_retrieved  INTEGER DEFAULT 0,
    times_cited      INTEGER DEFAULT 0,
    boost_factor     DECIMAL(5,4) DEFAULT 1.0, -- learned from ratings
    
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chunks_expert ON course_chunks(expert_id);
CREATE INDEX idx_chunks_topic ON course_chunks(expert_id, topic);
CREATE INDEX idx_chunks_embedding ON course_chunks 
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- ============================================================
-- EXPERT CAPABILITIES
-- ============================================================

CREATE TABLE expert_capabilities (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id         UUID NOT NULL REFERENCES experts(id) ON DELETE CASCADE,
    topic             VARCHAR(255) NOT NULL,
    depth_level       INTEGER CHECK (depth_level BETWEEN 1 AND 5),
    chunk_count       INTEGER NOT NULL,
    complexity_ceiling VARCHAR(50), -- basic/intermediate/advanced/expert/master
    can_handle        TEXT[],
    cannot_handle     TEXT[],
    example_questions TEXT[],
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(expert_id, topic)
);

-- ============================================================
-- PROJECTS
-- ============================================================

CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    status          VARCHAR(50) DEFAULT 'active', -- active/completed/archived
    
    -- GitHub/GitLab integration
    repo_url        VARCHAR(500),
    repo_provider   VARCHAR(20),  -- 'github' | 'gitlab' | null
    repo_branch     VARCHAR(255) DEFAULT 'main',
    repo_connected  BOOLEAN DEFAULT FALSE,
    repo_last_sync  TIMESTAMPTZ,
    
    -- Context
    tech_stack      JSONB DEFAULT '{}',
    architecture_type VARCHAR(100), -- monolith/microservices/serverless
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_projects_client ON projects(client_id) WHERE deleted_at IS NULL;

-- ============================================================
-- PROJECT EXPERTS (which experts are in this project)
-- ============================================================

CREATE TABLE project_experts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    expert_id   UUID NOT NULL REFERENCES experts(id),
    added_at    TIMESTAMPTZ DEFAULT NOW(),
    is_active   BOOLEAN DEFAULT TRUE,
    UNIQUE(project_id, expert_id)
);

CREATE INDEX idx_project_experts_project ON project_experts(project_id);

-- ============================================================
-- CHAT WINDOWS (one per project, or multiple per project)
-- ============================================================

CREATE TABLE chats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    client_id       UUID NOT NULL REFERENCES users(id),
    title           VARCHAR(500) NOT NULL,
    message_count   INTEGER DEFAULT 0,
    is_archived     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chats_project ON chats(project_id);

-- ============================================================
-- MESSAGES
-- ============================================================

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL, -- 'user' | 'assistant'
    content         TEXT NOT NULL,
    turn_number     INTEGER NOT NULL,
    
    -- For assistant messages
    expert_id       UUID REFERENCES experts(id),
    decision_mode   VARCHAR(20), -- ASK/WARN/PUSH_BACK/REFUSE/ADVISE
    citations       JSONB,       -- [{chunk_id, text, score}]
    gate_stopped    INTEGER,     -- 0 = reached Gate 5
    confidence      DECIMAL(5,4),
    
    -- Cost tracking
    tokens_used     INTEGER,
    cost_usd        DECIMAL(10,6),
    model_used      VARCHAR(100),
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_chat ON messages(chat_id, turn_number);

-- ============================================================
-- CHAT INDEX (smart indexing for context retrieval)
-- ============================================================

CREATE TABLE chat_index (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    message_id      UUID REFERENCES messages(id),
    turn_number     INTEGER NOT NULL,
    
    -- Metadata (generated by cheap LLM)
    tags            TEXT[],      -- requirement/decision/code/error/fix
    topic           VARCHAR(255),
    one_line_summary TEXT NOT NULL,
    importance      INTEGER CHECK (importance BETWEEN 1 AND 5),
    key_decisions   JSONB DEFAULT '[]',
    
    -- Semantic search
    embedding       vector(768) NOT NULL,
    
    -- Superseding (when decision changes)
    superseded_by   UUID REFERENCES chat_index(id),
    is_superseded   BOOLEAN DEFAULT FALSE,
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chat_index_chat ON chat_index(chat_id);
CREATE INDEX idx_chat_index_importance ON chat_index(chat_id, importance DESC);
CREATE INDEX idx_chat_index_embedding ON chat_index 
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 50);

-- ============================================================
-- CHAT SUMMARIES (rolling summaries every 10 turns)
-- ============================================================

CREATE TABLE chat_summaries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id         UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    summary_text    TEXT NOT NULL,
    turn_range_start INTEGER NOT NULL,
    turn_range_end   INTEGER NOT NULL,
    key_decisions   JSONB DEFAULT '[]',
    open_questions  TEXT[],
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_summaries_chat ON chat_summaries(chat_id, turn_range_end DESC);

-- ============================================================
-- L2 GROUP MEMORY (per project, all experts combined)
-- ============================================================

CREATE TABLE project_memory_l2 (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    expert_id       UUID NOT NULL REFERENCES experts(id),
    
    -- What this expert did in this project
    memory_type     VARCHAR(50) NOT NULL, -- decision/code/error/fix/recommendation
    content         TEXT NOT NULL,
    context         TEXT,                 -- why this decision was made
    turn_reference  INTEGER,              -- which turn this came from
    
    -- Semantic search
    embedding       vector(768),
    
    importance      INTEGER CHECK (importance BETWEEN 1 AND 5),
    is_superseded   BOOLEAN DEFAULT FALSE,
    superseded_at   TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_l2_project ON project_memory_l2(project_id);
CREATE INDEX idx_l2_expert ON project_memory_l2(project_id, expert_id);
CREATE INDEX idx_l2_embedding ON project_memory_l2 
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 30);

-- ============================================================
-- L3 MASTER EVENT LOG (append-only, never delete)
-- ============================================================

CREATE TABLE master_event_log (
    id              BIGSERIAL PRIMARY KEY,  -- sequential for ordering
    project_id      UUID NOT NULL REFERENCES projects(id),
    expert_id       UUID REFERENCES experts(id),
    client_id       UUID NOT NULL REFERENCES users(id),
    chat_id         UUID REFERENCES chats(id),
    message_id      UUID REFERENCES messages(id),
    
    -- Event
    event_type      VARCHAR(100) NOT NULL,
    event_data      JSONB NOT NULL,
    
    -- WHY tracking (Arpit's principle)
    reasoning       TEXT,        -- why this action was taken
    decision_made   TEXT,        -- what was decided
    alternatives_considered JSONB DEFAULT '[]',
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
    -- NO updated_at, NO deleted_at — immutable
);

CREATE INDEX idx_l3_project ON master_event_log(project_id, created_at DESC);
CREATE INDEX idx_l3_expert ON master_event_log(expert_id, created_at DESC);
CREATE INDEX idx_l3_event_type ON master_event_log(event_type);

-- ============================================================
-- RATINGS
-- ============================================================

CREATE TABLE ratings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID NOT NULL REFERENCES messages(id),
    client_id       UUID NOT NULL REFERENCES users(id),
    expert_id       UUID NOT NULL REFERENCES experts(id),
    project_id      UUID NOT NULL REFERENCES projects(id),
    
    score           INTEGER CHECK (score BETWEEN 1 AND 5),
    feedback        TEXT,
    feedback_type   VARCHAR(50), -- accepted/rejected/modified/ignored
    
    -- Code execution result (if applicable)
    code_executed   BOOLEAN DEFAULT FALSE,
    execution_success BOOLEAN,
    error_message   TEXT,
    
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ratings_expert ON ratings(expert_id);
CREATE INDEX idx_ratings_message ON ratings(message_id);

-- ============================================================
-- GITHUB/GITLAB CONNECTIONS
-- ============================================================

CREATE TABLE repo_connections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    client_id       UUID NOT NULL REFERENCES users(id),
    
    provider        VARCHAR(20) NOT NULL, -- github/gitlab
    repo_url        VARCHAR(500) NOT NULL,
    repo_name       VARCHAR(255),
    default_branch  VARCHAR(255) DEFAULT 'main',
    
    -- OAuth tokens (encrypted at rest)
    access_token    TEXT,        -- AES-256 encrypted
    refresh_token   TEXT,        -- AES-256 encrypted
    token_expires_at TIMESTAMPTZ,
    
    -- Sync status
    last_sync_at    TIMESTAMPTZ,
    sync_status     VARCHAR(50) DEFAULT 'pending', -- pending/syncing/complete/failed
    total_chunks    INTEGER DEFAULT 0,
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- INGESTION JOBS
-- ============================================================

CREATE TABLE ingestion_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expert_id       UUID REFERENCES experts(id),
    project_id      UUID REFERENCES projects(id), -- for repo ingestion
    
    job_type        VARCHAR(50) NOT NULL, -- transcript/repo
    status          VARCHAR(50) DEFAULT 'pending', -- pending/running/complete/failed
    
    source_path     VARCHAR(500),
    total_chunks    INTEGER DEFAULT 0,
    processed_chunks INTEGER DEFAULT 0,
    
    error_message   TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- SYSTEM SETTINGS
-- ============================================================

CREATE TABLE system_settings (
    key             VARCHAR(255) PRIMARY KEY,
    value           JSONB NOT NULL,
    description     TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_by      UUID REFERENCES users(id)
);

-- Default settings
INSERT INTO system_settings (key, value, description) VALUES
('china_wall', '{"reranker_threshold": 0.35, "max_retries": 5, "strip_uncited": true}', 'China Wall enforcement config'),
('context', '{"max_tokens": 10000, "recent_messages": 3, "semantic_top_k": 5, "course_chunks": 5}', 'Context assembly config'),
('models', '{"cheap": "deepseek/deepseek-chat", "strong": "anthropic/claude-3-5-sonnet", "fast": "google/gemini-flash-1.5"}', 'LLM model routing'),
('cost_budget', '{"monthly_limit_usd": 1000, "alert_threshold": 0.8}', 'Cost monitoring');
```

---

## 6. Memory System — L1/L2/L3

### Architecture

```
L1 — Individual Expert Memory (Redis)
     Scope: Per expert, per project
     Speed: O(1), sub-millisecond
     TTL: 24 hours (refreshed on access)
     Key pattern: "l1:{project_id}:{expert_id}"
     Content: Last 5 key decisions, current context summary
     WHY: Hot path — accessed on every turn. Redis prevents DB hit.

L2 — Group Memory (PostgreSQL + pgvector)
     Scope: All experts in a project
     Speed: O(log n) with index, ~5ms
     TTL: Project lifetime
     Table: project_memory_l2
     Content: All decisions, code, errors, fixes — with WHY reasoning
     WHY: When expert switches, new expert loads L2 to understand project state.
          Semantic search finds relevant past decisions.

L3 — Master Event Log (PostgreSQL, append-only)
     Scope: Complete project history
     Speed: Write O(1), Read O(log n)
     TTL: Forever — never delete
     Table: master_event_log
     Content: Every event, every decision, every rating, every WHY
     WHY: Audit trail, self-learning, pattern detection, debugging.
          "What did Expert A decide at turn 47 and why?"
```

### Memory Manager — Go Implementation

```go
// internal/memory/manager.go

type MemoryManager struct {
    db    *pgx.Pool
    redis *redis.Client
    ml    *MLSidecarClient
}

// L1 Operations
func (m *MemoryManager) GetL1(ctx context.Context, projectID, expertID uuid.UUID) (*L1Memory, error) {
    key := fmt.Sprintf("l1:%s:%s", projectID, expertID)
    data, err := m.redis.Get(ctx, key).Bytes()
    if err == redis.Nil {
        // L1 miss — rebuild from L2
        return m.rebuildL1FromL2(ctx, projectID, expertID)
    }
    var mem L1Memory
    json.Unmarshal(data, &mem)
    // Refresh TTL on access
    m.redis.Expire(ctx, key, 24*time.Hour)
    return &mem, nil
}

func (m *MemoryManager) UpdateL1(ctx context.Context, projectID, expertID uuid.UUID, update L1Update) error {
    key := fmt.Sprintf("l1:%s:%s", projectID, expertID)
    mem, _ := m.GetL1(ctx, projectID, expertID)
    mem.Apply(update)
    // Keep only last 5 key decisions in L1
    mem.Trim(5)
    data, _ := json.Marshal(mem)
    return m.redis.Set(ctx, key, data, 24*time.Hour).Err()
}

// L2 Operations
func (m *MemoryManager) GetProjectContext(ctx context.Context, projectID uuid.UUID, query string) ([]L2Memory, error) {
    // Semantic search in L2 for relevant project context
    embedding, _ := m.ml.Embed(ctx, []string{query})
    
    rows, _ := m.db.Query(ctx, `
        SELECT id, expert_id, memory_type, content, context, importance,
               1 - (embedding <=> $1) AS similarity
        FROM project_memory_l2
        WHERE project_id = $2
          AND is_superseded = FALSE
        ORDER BY embedding <=> $1
        LIMIT 10
    `, pgvector.NewVector(embedding[0]), projectID)
    
    return scanL2Rows(rows)
}

func (m *MemoryManager) AppendL2(ctx context.Context, entry L2Entry) error {
    embedding, _ := m.ml.Embed(ctx, []string{entry.Content})
    _, err := m.db.Exec(ctx, `
        INSERT INTO project_memory_l2 
        (project_id, expert_id, memory_type, content, context, turn_reference, embedding, importance)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, entry.ProjectID, entry.ExpertID, entry.MemoryType, entry.Content,
       entry.Context, entry.TurnReference, pgvector.NewVector(embedding[0]), entry.Importance)
    return err
}

// L3 Operations
func (m *MemoryManager) AppendL3(ctx context.Context, event L3Event) error {
    // Append-only — never update, never delete
    _, err := m.db.Exec(ctx, `
        INSERT INTO master_event_log 
        (project_id, expert_id, client_id, chat_id, message_id,
         event_type, event_data, reasoning, decision_made, alternatives_considered)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, event.ProjectID, event.ExpertID, event.ClientID, event.ChatID,
       event.MessageID, event.EventType, event.EventData,
       event.Reasoning, event.DecisionMade, event.AlternativesConsidered)
    return err
}

// Rebuild L1 from L2 (cache miss recovery)
func (m *MemoryManager) rebuildL1FromL2(ctx context.Context, projectID, expertID uuid.UUID) (*L1Memory, error) {
    rows, _ := m.db.Query(ctx, `
        SELECT content, context, memory_type, importance, created_at
        FROM project_memory_l2
        WHERE project_id = $1 AND expert_id = $2 AND is_superseded = FALSE
        ORDER BY importance DESC, created_at DESC
        LIMIT 5
    `, projectID, expertID)
    
    mem := &L1Memory{}
    // ... scan and build
    
    // Store in Redis
    key := fmt.Sprintf("l1:%s:%s", projectID, expertID)
    data, _ := json.Marshal(mem)
    m.redis.Set(ctx, key, data, 24*time.Hour)
    
    return mem, nil
}
```

---

## 7. Agent Orchestration Engine

### Expert Selection Algorithm

```go
// internal/orchestrator/expert_selector.go

// When client selects multiple experts:
// 1. Check if experts are active and trained
// 2. Check if question is in their domain (capability check)
// 3. Route to parallel goroutines
// 4. Synthesize results

func (o *ExpertOrchestrator) selectRelevantExperts(
    ctx context.Context,
    projectID uuid.UUID,
    selectedExpertIDs []uuid.UUID,
    question string,
) ([]Expert, error) {
    // Load experts
    experts := o.loadExperts(ctx, selectedExpertIDs)
    
    // Quick capability pre-check (avoid unnecessary LLM calls)
    relevant := []Expert{}
    for _, expert := range experts {
        if o.hasCapability(expert, question) {
            relevant = append(relevant, expert)
        }
    }
    
    return relevant, nil
}

func (o *ExpertOrchestrator) hasCapability(expert Expert, question string) bool {
    // Check capability_summary JSONB
    // If question keywords match any topic with depth >= 2, proceed
    // This is a fast pre-filter before the full 5-gate system
    questionLower := strings.ToLower(question)
    for topic := range expert.CapabilitySummary {
        if strings.Contains(questionLower, strings.ToLower(topic)) {
            return true
        }
    }
    return true // Default: let the 5-gate system decide
}
```

### Synthesis Algorithm

```go
// internal/orchestrator/synthesizer.go

// When multiple experts respond:
// 1. Find agreements
// 2. Find contradictions
// 3. Present both with reasoning
// 4. Let client decide

type SynthesisResult struct {
    Agreements     []string
    Contradictions []Contradiction
    Recommendation string
    AllCitations   []Citation
}

type Contradiction struct {
    Topic      string
    ExpertA    string
    PositionA  string
    ExpertB    string
    PositionB  string
    Resolution string // "Client should decide based on: ..."
}

func (s *Synthesizer) Synthesize(ctx context.Context, responses []ExpertResponse) (*SynthesisResult, error) {
    if len(responses) == 1 {
        return nil, nil // No synthesis needed
    }
    
    // Build synthesis prompt
    prompt := s.buildSynthesisPrompt(responses)
    
    // Use cheap model for synthesis
    resp, _ := s.gateway.Call(ctx, LLMRequest{
        Model:       ModelCheap,
        UserPrompt:  prompt,
        MaxTokens:   1000,
        Temperature: 0.3,
    })
    
    return s.parseSynthesisResponse(resp.Content, responses)
}
```

---

## 8. China Wall Enforcement

### 4-Layer System in Go

```go
// internal/chinawall/enforcer.go

const (
    RerankerThreshold = 0.35
    RelaxedThreshold  = 0.30
    MaxRetries        = 5
)

type ChinaWallEnforcer struct {
    gateway *ModelGateway
    ml      *MLSidecarClient
    db      *pgx.Pool
}

type EnforceResult struct {
    Status      string // success/retry/refused/partial
    Answer      string
    Citations   []Citation
    Coverage    string // YES/PARTIAL/NO
    Confidence  float64
    LayerFailed int    // 0 = all passed
    Reason      string
}

func (e *ChinaWallEnforcer) Enforce(
    ctx context.Context,
    question string,
    chunks []CourseChunk,
    expert Expert,
    attempt int,
) (*EnforceResult, error) {
    
    // LAYER 1: Reranker threshold
    threshold := RerankerThreshold
    if attempt >= 4 {
        threshold = RelaxedThreshold // Last resort
    }
    
    reranked, _ := e.ml.Rerank(ctx, question, chunksToTexts(chunks), 5)
    
    if len(reranked) == 0 || reranked[0].Score < float32(threshold) {
        if attempt >= MaxRetries {
            return e.buildRefusal(question, chunks, "insufficient_relevance"), nil
        }
        return &EnforceResult{Status: "retry", LayerFailed: 1}, nil
    }
    
    topChunks := rerankedToChunks(reranked, chunks)
    
    // LAYER 2: Coverage check
    coverage, _ := e.checkCoverage(ctx, question, topChunks)
    
    if coverage.Status == "NO" {
        return e.buildRefusal(question, chunks, "not_covered"), nil
    }
    
    if coverage.Status == "PARTIAL" && attempt >= 3 {
        return e.buildPartialResponse(question, topChunks, coverage), nil
    }
    
    // LAYER 3: Generate with mandatory citations
    generated, _ := e.generateWithCitations(ctx, question, topChunks, expert)
    
    if len(generated.Citations) == 0 {
        if attempt >= MaxRetries {
            return e.buildRefusal(question, chunks, "citation_failure"), nil
        }
        return &EnforceResult{Status: "retry", LayerFailed: 3}, nil
    }
    
    // LAYER 4: Strip uncited claims
    cleaned := e.stripUncited(generated.Answer, generated.Citations)
    
    if cleaned.StrippedCount > 0 {
        // Log violation
        e.logViolation(ctx, question, cleaned.StrippedCount)
    }
    
    return &EnforceResult{
        Status:     "success",
        Answer:     cleaned.CleanAnswer,
        Citations:  generated.Citations,
        Coverage:   coverage.Status,
        Confidence: coverage.Confidence,
    }, nil
}

// Layer 4: Strip uncited claims using regex
func (e *ChinaWallEnforcer) stripUncited(answer string, citations []Citation) StripResult {
    // Split into sentences
    sentences := splitSentences(answer)
    
    citationPattern := regexp.MustCompile(`\[CHUNK_[a-f0-9-]+\]`)
    
    clean := []string{}
    stripped := 0
    
    for _, sentence := range sentences {
        hasCitation := citationPattern.MatchString(sentence)
        isShort := len(strings.TrimSpace(sentence)) < 20
        isStructural := isHeadingOrTransition(sentence)
        
        if hasCitation || isShort || isStructural {
            clean = append(clean, sentence)
        } else {
            stripped++
        }
    }
    
    return StripResult{
        CleanAnswer:  strings.Join(clean, " "),
        StrippedCount: stripped,
    }
}
```

---

## 9. Context Budget Manager

### Algorithm

```go
// internal/context/budget_manager.go

const (
    MaxContextTokens    = 10000
    RollingSummaryEvery = 10  // turns
    RecentMessagesCount = 3
    SemanticTopK        = 5
    CourseChunksTopK    = 5
)

type ContextBudgetManager struct {
    db      *pgx.Pool
    gateway *ModelGateway
    ml      *MLSidecarClient
}

type AssembledContext struct {
    RollingSummary  string
    RecentMessages  []Message
    RelevantHistory []ChatIndexEntry
    CourseChunks    []CourseChunk
    L2Memory        []L2Memory
    TotalTokens     int
}

func (m *ContextBudgetManager) Assemble(
    ctx context.Context,
    chatID uuid.UUID,
    projectID uuid.UUID,
    expertID uuid.UUID,
    question string,
    turnNumber int,
) (*AssembledContext, error) {
    
    assembled := &AssembledContext{}
    tokensUsed := 0
    
    // 1. Rolling summary (most compressed, highest priority)
    summary, _ := m.getRollingSummary(ctx, chatID)
    if summary != "" {
        assembled.RollingSummary = summary
        tokensUsed += estimateTokens(summary)
    }
    
    // 2. L2 project memory (what other experts did)
    if tokensUsed < MaxContextTokens*40/100 { // 40% budget
        l2, _ := m.getL2Memory(ctx, projectID, question)
        assembled.L2Memory = l2
        for _, m := range l2 {
            tokensUsed += estimateTokens(m.Content)
        }
    }
    
    // 3. Recent messages (last 3)
    if tokensUsed < MaxContextTokens*60/100 { // 60% budget
        recent, _ := m.getRecentMessages(ctx, chatID, RecentMessagesCount)
        assembled.RecentMessages = recent
        for _, msg := range recent {
            tokensUsed += estimateTokens(msg.Content)
        }
    }
    
    // 4. Semantic search in chat history
    if tokensUsed < MaxContextTokens*70/100 { // 70% budget
        history, _ := m.searchChatHistory(ctx, chatID, question, SemanticTopK)
        assembled.RelevantHistory = history
        for _, h := range history {
            tokensUsed += estimateTokens(h.OneLineSummary)
        }
    }
    
    // 5. Course chunks (expert knowledge)
    if tokensUsed < MaxContextTokens*90/100 { // 90% budget
        chunks, _ := m.getCourseChunks(ctx, expertID, question, CourseChunksTopK)
        assembled.CourseChunks = chunks
        for _, c := range chunks {
            tokensUsed += estimateTokens(c.ChunkText)
        }
    }
    
    assembled.TotalTokens = tokensUsed
    
    // Trigger rolling summary if needed
    if turnNumber > 0 && turnNumber%RollingSummaryEvery == 0 {
        go m.generateRollingSummary(context.Background(), chatID, turnNumber)
    }
    
    return assembled, nil
}

func estimateTokens(text string) int {
    return len(text) / 4 // ~4 chars per token
}
```

---

## 10. Decision Engine — 5 Gates

```go
// internal/decision/engine.go

type ResponseMode string

const (
    ModeASK       ResponseMode = "ASK"
    ModeWARN      ResponseMode = "WARN"
    ModePUSHBACK  ResponseMode = "PUSH_BACK"
    ModeREFUSE    ResponseMode = "REFUSE"
    ModeADVISE    ResponseMode = "ADVISE"
)

type DecisionEngine struct {
    gateway   *ModelGateway
    chinaWall *ChinaWallEnforcer
    ml        *MLSidecarClient
}

type DecisionResult struct {
    Mode        ResponseMode
    Content     string
    Citations   []Citation
    Confidence  float64
    GateStopped int
    Warning     string
    Questions   []string
    Alternative string
}

func (e *DecisionEngine) Process(
    ctx context.Context,
    question string,
    context *AssembledContext,
    expert Expert,
    attempt int,
) (*DecisionResult, error) {
    
    // GATE 1: Information Sufficiency
    if result := e.gate1(ctx, question, context, expert); result != nil {
        return result, nil
    }
    
    // GATE 2: Knowledge Coverage
    if result := e.gate2(ctx, question, context.CourseChunks, expert); result != nil {
        return result, nil
    }
    
    // GATE 3: Charter Compliance
    warning := ""
    if result := e.gate3(ctx, question, context, expert); result != nil {
        if result.Mode == ModeWARN {
            warning = result.Warning // Continue but include warning
        } else {
            return result, nil
        }
    }
    
    // GATE 4: Necessity Check
    if result := e.gate4(ctx, question, context, expert); result != nil {
        return result, nil
    }
    
    // GATE 5: Generate Answer (China Wall)
    result, err := e.gate5(ctx, question, context, expert, warning, attempt)
    if err != nil {
        return nil, err
    }
    
    if result.Status == "retry" && attempt < MaxRetries {
        return e.Process(ctx, question, context, expert, attempt+1)
    }
    
    return result, nil
}

// Gate 1: Do we have enough information?
func (e *DecisionEngine) gate1(
    ctx context.Context,
    question string,
    context *AssembledContext,
    expert Expert,
) *DecisionResult {
    
    // Check clarification charter
    clarificationCharter := expert.ClarificationCharter
    
    // Extract topic from question (fast, no LLM)
    topic := extractTopic(question)
    
    if questions, ok := clarificationCharter[topic]; ok {
        if shouldAskClarification(question, context) {
            return &DecisionResult{
                Mode:        ModeASK,
                Questions:   questions[:min(3, len(questions))],
                GateStopped: 1,
            }
        }
    }
    
    return nil // Proceed
}

// Gate 2: Is this in our knowledge domain?
func (e *DecisionEngine) gate2(
    ctx context.Context,
    question string,
    chunks []CourseChunk,
    expert Expert,
) *DecisionResult {
    
    if len(chunks) == 0 {
        return &DecisionResult{
            Mode:        ModeREFUSE,
            Content:     fmt.Sprintf("This topic is not in my training material. I cover: %s", expert.CoveredTopics()),
            GateStopped: 2,
        }
    }
    
    bestScore := float32(0)
    for _, c := range chunks {
        if c.RerankScore > bestScore {
            bestScore = c.RerankScore
        }
    }
    
    if bestScore < RerankerThreshold {
        return &DecisionResult{
            Mode:        ModeREFUSE,
            Content:     fmt.Sprintf("Question doesn't match my training content (score: %.2f < %.2f)", bestScore, RerankerThreshold),
            GateStopped: 2,
        }
    }
    
    return nil // Proceed
}
```

---

## 11. Expert Training Pipeline

### Ingestion Flow

```
Admin uploads transcript (txt/md/pdf)
        │
        ▼
Ingestion Job created (status: pending)
        │
        ▼
Background Worker (Go goroutine) picks up job
        │
        ▼
Text Chunker: 500-800 tokens, 100 overlap, sentence boundaries
        │
        ▼
Topic Extractor: LLM call (cheap model) → topic + subtopic per chunk
        │
        ▼
Embedding Generator: Python ML sidecar → 768D vectors
        │
        ▼
Charter Extractor: LLM call (strong model) → reasoning + clarification charters
        │
        ▼
Capability Builder: Analyze chunks per topic → depth levels 1-5
        │
        ▼
Store in PostgreSQL (course_chunks + expert_capabilities)
        │
        ▼
Update expert stats (total_chunks, total_topics, avg_depth_level)
        │
        ▼
Job status: complete
```

### Charter Extraction — WHY Principle

```go
// internal/training/charter_extractor.go

// This is where the WHY principle is embedded into each expert.
// Every expert must know:
// 1. WHAT they know
// 2. WHY they recommend what they recommend
// 3. WHEN to apply which principle
// 4. WHAT NOT to do (anti-patterns)

const charterExtractionPrompt = `
Analyze this course transcript and extract the instructor's core principles.

For REASONING CHARTER, extract:
1. Strong opinions ("I believe...", "In my experience...")
2. Never/Always statements ("Never use X", "Always consider Y")
3. Decision rules with WHY ("When X, do Y BECAUSE Z")
4. Anti-patterns with reasoning ("Avoid X BECAUSE it causes Y")
5. Trade-off principles ("Choose X over Y when Z")

IMPORTANT: Every rule must include the WHY.
Not just "Never use microservices for small teams"
But "Never use microservices for small teams BECAUSE operational overhead
exceeds development velocity when team < 10 people"

For CLARIFICATION CHARTER, extract:
For each major topic, what questions does the instructor ask BEFORE answering?
These are the questions that determine WHICH answer is correct.

Format reasoning charter as plain text.
Format clarification charter as JSON: {"topic": ["question1", "question2"]}
`
```

---

## 12. GitHub/GitLab Integration

### Flow

```go
// internal/repo/integration.go

// Step 1: OAuth2 connection
// Step 2: Clone/fetch repo
// Step 3: Parse file tree
// Step 4: Identify relevant files (by extension, size)
// Step 5: Chunk files
// Step 6: Generate embeddings
// Step 7: Store in project-specific vector space
// Step 8: Experts can now search codebase

type RepoIntegration struct {
    db      *pgx.Pool
    ml      *MLSidecarClient
    gateway *ModelGateway
}

func (r *RepoIntegration) SyncRepo(ctx context.Context, connectionID uuid.UUID) error {
    conn, _ := r.getConnection(ctx, connectionID)
    
    // Update status
    r.updateSyncStatus(ctx, connectionID, "syncing")
    
    // Clone/fetch repo
    repoPath, _ := r.fetchRepo(ctx, conn)
    defer os.RemoveAll(repoPath) // Cleanup after processing
    
    // Walk file tree
    files, _ := r.walkRepo(repoPath, conn.DefaultBranch)
    
    // Process files in batches
    batchSize := 10
    for i := 0; i < len(files); i += batchSize {
        batch := files[i:min(i+batchSize, len(files))]
        r.processBatch(ctx, conn, batch)
    }
    
    r.updateSyncStatus(ctx, connectionID, "complete")
    return nil
}

func (r *RepoIntegration) walkRepo(repoPath, branch string) ([]RepoFile, error) {
    // Skip: node_modules, .git, vendor, build artifacts
    // Include: .go, .py, .ts, .tsx, .js, .java, .md, .yaml, .json
    // Max file size: 500KB
    
    var files []RepoFile
    filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
        if shouldSkip(path) {
            return filepath.SkipDir
        }
        if isRelevantFile(path) && info.Size() < 500*1024 {
            files = append(files, RepoFile{Path: path, Size: info.Size()})
        }
        return nil
    })
    return files, nil
}
```

---

## 13. Rating and Self-Learning System

```go
// internal/learning/rating_engine.go

// After every expert response:
// Client can rate 1-5 stars + optional feedback
// Rating updates:
// 1. Expert's avg_rating in experts table
// 2. Chunk boost_factor in course_chunks (cited chunks get boosted)
// 3. L3 master event log (for pattern analysis)
// 4. Admin panel shows low-rated responses for review

func (e *RatingEngine) RecordRating(ctx context.Context, rating Rating) error {
    // Store rating
    _, err := e.db.Exec(ctx, `
        INSERT INTO ratings (message_id, client_id, expert_id, project_id, score, feedback, feedback_type)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, rating.MessageID, rating.ClientID, rating.ExpertID, rating.ProjectID,
       rating.Score, rating.Feedback, rating.FeedbackType)
    
    // Update expert avg_rating
    e.db.Exec(ctx, `
        UPDATE experts SET
            avg_rating = (
                SELECT AVG(score) FROM ratings WHERE expert_id = $1
            ),
            total_ratings = total_ratings + 1
        WHERE id = $1
    `, rating.ExpertID)
    
    // Update chunk boost factors (async)
    go e.updateChunkBoosts(context.Background(), rating)
    
    // Append to L3
    go e.memoryManager.AppendL3(context.Background(), L3Event{
        EventType: "rating_recorded",
        EventData: map[string]interface{}{
            "score":    rating.Score,
            "feedback": rating.Feedback,
        },
        Reasoning: "Client feedback on expert response",
    })
    
    return err
}

func (e *RatingEngine) updateChunkBoosts(ctx context.Context, rating Rating) {
    // Get cited chunks from the rated message
    msg, _ := e.getMessage(ctx, rating.MessageID)
    
    for _, citation := range msg.Citations {
        var boostDelta float64
        if rating.Score >= 4 {
            boostDelta = 0.05  // Boost good chunks
        } else if rating.Score <= 2 {
            boostDelta = -0.05 // Penalize bad chunks
        }
        
        if boostDelta != 0 {
            e.db.Exec(ctx, `
                UPDATE course_chunks SET
                    boost_factor = GREATEST(0.5, LEAST(1.5, boost_factor + $1)),
                    times_cited = times_cited + 1
                WHERE id = $2
            `, boostDelta, citation.ChunkID)
        }
    }
}
```

---

## 14. Admin Panel

### Security: Google Authenticator (TOTP)

```go
// internal/auth/totp.go

// Admin login flow:
// 1. Email + password (bcrypt)
// 2. TOTP code (Google Authenticator)
// 3. JWT issued with role=admin
// 4. All admin routes require role=admin in JWT

func (a *AuthService) AdminLogin(ctx context.Context, req AdminLoginRequest) (*TokenPair, error) {
    user, _ := a.getUserByEmail(ctx, req.Email)
    
    if user.Role != "admin" {
        return nil, ErrUnauthorized
    }
    
    // Verify password
    if !bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)) {
        return nil, ErrInvalidCredentials
    }
    
    // Verify TOTP
    if user.TOTPEnabled {
        valid := totp.Validate(req.TOTPCode, user.TOTPSecret)
        if !valid {
            return nil, ErrInvalidTOTP
        }
    }
    
    return a.issueTokens(user)
}
```

### Admin Capabilities

```
1. Expert Management
   - View all experts with stats (avg_rating, total_chunks, total_topics)
   - Enable/disable experts
   - Upload new transcripts → trigger ingestion job
   - View ingestion job status
   - Retrain expert (re-ingest transcripts)
   - Edit reasoning charter manually

2. Client Management
   - View all clients
   - Enable/disable clients
   - View client projects and usage

3. System Monitoring
   - Total cost (daily/monthly)
   - China Wall violation rate
   - Average response time per expert
   - Token usage breakdown

4. Rating Analytics
   - Low-rated responses (score <= 2) — flagged for review
   - Expert performance over time
   - Topic-wise accuracy

5. System Settings
   - China Wall thresholds
   - Context budget limits
   - Model routing
   - Cost budget alerts
```

---

## 15. API Design

### Base URL: `/api/v1`

#### Authentication
```
POST /auth/register          — Client registration
POST /auth/login             — Client login → JWT
POST /auth/admin/login       — Admin login → JWT + TOTP
POST /auth/refresh           — Refresh JWT
POST /auth/logout            — Invalidate token
```

#### Experts
```
GET  /experts                — List active experts
GET  /experts/:id            — Expert details + capabilities
GET  /experts/:id/topics     — Covered topics with depth levels
```

#### Projects
```
POST /projects               — Create project
GET  /projects               — List client's projects
GET  /projects/:id           — Project details
PATCH /projects/:id          — Update project
DELETE /projects/:id         — Soft delete

POST /projects/:id/experts   — Add expert to project
DELETE /projects/:id/experts/:expertId — Remove expert

POST /projects/:id/repo      — Connect GitHub/GitLab repo
POST /projects/:id/repo/sync — Trigger repo sync
GET  /projects/:id/repo/status — Sync status
```

#### Chats
```
POST /projects/:id/chats     — Create chat window
GET  /projects/:id/chats     — List chats
GET  /chats/:id              — Chat details
PATCH /chats/:id             — Update title
DELETE /chats/:id            — Archive chat
```

#### Messages
```
POST /chats/:id/messages     — Send message (SSE streaming response)
GET  /chats/:id/messages     — List messages
GET  /chats/:id/messages/:msgId — Single message

POST /messages/:id/rate      — Rate a response (1-5 stars)
```

#### Memory
```
GET  /projects/:id/memory    — Project L2 memory summary
GET  /projects/:id/timeline  — L3 event log (paginated)
```

#### Admin
```
GET  /admin/experts          — All experts with stats
POST /admin/experts          — Create expert
PATCH /admin/experts/:id     — Update expert
POST /admin/experts/:id/ingest — Upload transcript + trigger ingestion
GET  /admin/experts/:id/jobs — Ingestion job status

GET  /admin/clients          — All clients
PATCH /admin/clients/:id     — Enable/disable client

GET  /admin/stats            — System stats
GET  /admin/violations       — China Wall violations
GET  /admin/ratings          — Rating analytics

GET  /admin/settings         — System settings
PATCH /admin/settings/:key   — Update setting
```

### Response Format

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-09-04T10:00:00Z"
  }
}
```

### SSE Streaming (Message Response)

```
POST /chats/:id/messages
Content-Type: text/event-stream

data: {"type": "thinking", "expert": "Arpit", "gate": 1}
data: {"type": "thinking", "expert": "Arpit", "gate": 5}
data: {"type": "chunk", "expert": "Arpit", "content": "Sharding distributes..."}
data: {"type": "chunk", "expert": "Arpit", "content": "[CHUNK_abc123]"}
data: {"type": "complete", "expert": "Arpit", "mode": "ADVISE", "confidence": 0.92}
data: {"type": "synthesis", "agreements": [...], "contradictions": [...]}
data: {"type": "done"}
```

---

## 16. Security Architecture

```
1. Authentication
   - JWT (access: 15min, refresh: 7 days)
   - bcrypt password hashing (cost factor 12)
   - Admin: TOTP (Google Authenticator) mandatory
   - Rate limiting: 100 req/min per IP, 20 req/min per user

2. Authorization
   - Role-based: admin | client
   - Row-level security in PostgreSQL
   - Clients see only their projects/chats

3. Data Security
   - OAuth tokens encrypted at rest (AES-256)
   - HTTPS only (TLS 1.3)
   - No sensitive data in logs

4. Input Validation
   - All inputs validated before processing
   - SQL injection: parameterized queries only (pgx)
   - XSS: sanitize all user content
   - File uploads: type check + size limit (50MB)

5. Prompt Injection Protection
   - User input never directly in system prompt
   - China Wall prevents external knowledge injection
   - All LLM outputs validated before returning to client
```

---

## 17. Deployment Architecture

```yaml
# docker-compose.yml (Development)
services:
  api:
    build: ./backend-go
    ports: ["8080:8080"]
    depends_on: [postgres, redis, ml-sidecar]
    environment:
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://...
      - ML_SIDECAR_URL=http://ml-sidecar:8001
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}

  ml-sidecar:
    build: ./ml-sidecar
    ports: ["8001:8001"]
    volumes:
      - ml-models:/models  # Cached model weights

  postgres:
    image: pgvector/pgvector:pg15
    volumes: [postgres-data:/var/lib/postgresql/data]

  redis:
    image: redis:7-alpine
    volumes: [redis-data:/data]

  frontend:
    build: ./frontend
    ports: ["3000:80"]
```

### Directory Structure

```
ai_avengers/
├── backend-go/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/
│   │   ├── orchestrator/
│   │   ├── memory/
│   │   ├── chinawall/
│   │   ├── context/
│   │   ├── decision/
│   │   ├── gateway/
│   │   ├── ml/
│   │   ├── training/
│   │   ├── repo/
│   │   ├── learning/
│   │   ├── admin/
│   │   └── db/
│   ├── migrations/
│   ├── go.mod
│   └── go.sum
├── ml-sidecar/
│   ├── main.py
│   ├── embeddings.py
│   ├── reranker.py
│   └── requirements.txt
├── frontend/          # React (to be provided)
├── docs/
│   └── AI_AVENGERS_SYSTEM_ARCHITECTURE.md
└── docker-compose.yml
```

---

## 18. Locked Decisions

These decisions are FINAL. Do not change without explicit discussion.

| # | Decision | Reason | Exception |
|---|---|---|---|
| 1 | Go for backend | Goroutines for parallel experts, no GIL | None |
| 2 | Python ML sidecar | sentence-transformers stability | Replace if Go ONNX matures |
| 3 | PostgreSQL + pgvector | Single DB, vector + relational | None |
| 4 | Redis for L1 memory | O(1) hot path access | None |
| 5 | UUID primary keys | Distributed-safe | None |
| 6 | Append-only L3 | Audit trail, never lose data | None |
| 7 | China Wall 4 layers | Hallucination prevention | Threshold configurable |
| 8 | 5-gate decision engine | Expert behavior, not chatbot | Gates configurable |
| 9 | SSE for streaming | Simpler than WebSocket for one-way | WebSocket if bidirectional needed |
| 10 | Modular monolith v1 | Avoid premature microservices | Extract services when bottleneck found |
| 11 | TOTP for admin | Security — no exceptions | None |
| 12 | WHY in every charter rule | Arpit's principle — experts must reason | None |

---

## 19. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Go project setup (gin, pgx, redis, zap)
- [ ] Database migrations (all tables)
- [ ] Auth service (JWT + bcrypt + TOTP)
- [ ] ML sidecar (FastAPI + embeddings + reranker)
- [ ] Model gateway (OpenRouter integration)
- [ ] Basic API structure

### Phase 2: Expert Training (Week 2-3)
- [ ] Ingestion pipeline (chunking + topic extraction)
- [ ] Charter extractor (WHY principle embedded)
- [ ] Capability builder
- [ ] Expert CRUD + admin endpoints

### Phase 3: Core Intelligence (Week 3-4)
- [ ] Memory manager (L1/L2/L3)
- [ ] Context budget manager
- [ ] China Wall enforcer (4 layers)
- [ ] Decision engine (5 gates)
- [ ] Expert orchestrator (parallel goroutines)

### Phase 4: Project & Chat System (Week 4-5)
- [ ] Project management
- [ ] Chat windows
- [ ] Message processing with SSE streaming
- [ ] Chat index + rolling summaries

### Phase 5: Advanced Features (Week 5-6)
- [ ] GitHub/GitLab OAuth + repo sync
- [ ] Rating system + chunk boost learning
- [ ] Admin panel APIs
- [ ] Multi-expert synthesis

### Phase 6: Polish & Deploy (Week 6-7)
- [ ] Docker setup
- [ ] Rate limiting
- [ ] Monitoring (cost tracking, violation logs)
- [ ] Load testing
- [ ] Documentation

---

## Checkpoint Conditions

How do we know each phase is working?

| Phase | Checkpoint |
|---|---|
| Phase 1 | Admin can login with TOTP. JWT works. Embeddings generate. |
| Phase 2 | Upload transcript → expert created → chunks searchable |
| Phase 3 | Ask question → 5 gates run → China Wall enforces → cited answer |
| Phase 4 | Create project → add experts → send message → get response |
| Phase 5 | Connect GitHub repo → expert can reference code in answers |
| Phase 6 | 100 concurrent users → < 8 second response → no errors |

---

*Architecture designed from transcript knowledge: Agentic AI Engineering Course, 10 Hours of LLM, Arpit Bhiyani Masterclass, Designing Modern Web-Scale Distributed Services, Grokking Advanced System Design, Coding Interview Patterns.*
