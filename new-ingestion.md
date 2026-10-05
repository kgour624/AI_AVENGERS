# AI Avengers: End-to-End Ingestion, Retrieval & Agentic Inference Architecture

## 1. Executive Summary & Business Impact
This document outlines the core architectural logic and end-to-end functionality of the upgraded AI Avengers Retrieval-Augmented Generation (RAG) pipeline. The system has transitioned from a naive, single-shot vector retrieval model to a highly sophisticated, multi-phase Agentic RAG system. 

For the business, this translates to:
- **Zero "Lost-in-the-Middle" Hallucinations**: By preserving document hierarchy (Parent-Child retrieval), the AI never loses the broader context of a snippet.
- **Cost-Efficiency**: By using app-side semantic caching and leveraging "Cheap" model classifiers (low temperature) to gate expensive generation steps, API costs are drastically reduced.
- **Fail-Safe Determinism**: The orchestration logic lives in deterministic Go code (not in the LLM's prompt), ensuring that timeouts, circuit breaking, and database limits are strictly enforced.

---

## 2. The Core Mental Model: 3 Problems = 1 Unified Pipeline
The architecture applies the fundamental engineering rule: **Break into phases (Store → Pick → Execute)**. We solve one thing well, evaluate component-wise, and only then chain them together. 

While the system is conceptually divided into three phases, the storage mechanism acts as a Single Source of Truth (SSOT): **PostgreSQL + pgvector + API Gateway**.

The three core challenges addressed in this upgrade are:
1. **Problem A (Ingestion)**: Fixed-Size Chunking leading to semantic boundary violations.
2. **Problem C (Retrieval)**: Context Orphans leading to generic, ungrounded answers.
3. **Problem D (Inference)**: Single-Shot retrieval failing on multi-hop, complex queries.

---

## 3. Phase 1: Semantic & Markdown-Aware Chunking (Problem A)

### 3.1. The Flaw in Legacy Chunking
Previously, the system used fixed-size chunking (e.g., 500 tokens). This resulted in blind cuts (`Boundary Violation`). If a large Markdown file contained a `## Deep Architecture` section with a 600-token code fence, a blind 500-token cut would split the code block in half. The LLM would receive fragmented logic, which is an anti-pattern: *wrong chunking ensures that even if retrieval is perfectly relevant, the generation will be garbage.*

### 3.2. The Config-Driven Semantic Solution
Instead of hardcoding limits, the system now uses a configuration-driven approach backed by a Stack-based Interval Merge algorithm.

#### Data Structures & Limits
```go
type ChunkConfig struct {
    SoftLimit     int
    HardLimit     int
    OverlapTokens int
}
```
These values are fetched dynamically from the environment or database. 

#### The Algorithm
The chunker parses the document into a Markdown Abstract Syntax Tree (AST), understanding elements like Headings (`##`, `###`), Code Fences (` ``` `), Lists, and Tables. 
- **Heading Hierarchy**: A stack tracks heading levels. Every `##` starts a new Parent Interval.
- **Interval Merging (Greedy approach)**: The chunker attempts to split at natural boundaries (like `\n\n`). It merges intervals greedily until it hits the `SoftLimit`. It only forces a split if a block exceeds the `HardLimit`.

#### Code Flow Example
```go
func (c *Chunker) ChunkMarkdown(ctx context.Context, raw string, cfg ChunkConfig) ([]Chunk, error) {
    blocks := parseMarkdownBlocks(raw) // Atomic units: Heading, CodeFence
    var chunks []Chunk
    var cur strings.Builder
    var curTokens int

    for _, b := range blocks {
        t := estimateTokens(b.Text) // Word-count/4 approx (sub-millisecond cost)
        
        if curTokens + t > cfg.SoftLimit && curTokens > 0 {
            chunks = append(chunks, flush(cur))
            // Semantic overlap applied here (last 1-2 sentences)
        }
        
        if t > cfg.HardLimit { 
            // Atomic preservation: Do not cut a 600-token code block in half!
            chunks = append(chunks, Chunk{Text: b.Text, Type: b.Type, HeadingPath: b.Path})
        } else { 
            cur.WriteString(b.Text)
            curTokens += t 
        }
    }
    return chunks, nil
}
```

### 3.3. Key Business & Architectural Rules Apply Here:
- **Atomic Derivability**: Do not store derivable splits. A massive code block is kept as a single atomic unit to preserve correctness.
- **Intelligent Overlap**: Overlap is no longer a fixed 50 tokens. It is semantically aware (the last 1-2 sentences of the previous chunk).
- **Cost Mitigation**: Token estimation uses a fast local heuristic (`word-count / 4`) rather than an expensive API call to a remote tokenizer like tiktoken.
- **Evaluation Harness (C3)**: The golden test set validates that code blocks are *never* cut mid-fence.

---

## 4. Phase 2: Parent-Child Retrieval (Problem C)

### 4.1. The "Context Orphan" Problem
When an LLM retrieves a pinpoint 500-token chunk, it suffers from severe "Lost-in-the-Middle" syndrome. It finds the exact vector match, but lacks the surrounding context of the previous page. This violates the memory taxonomy rule: *Vector search is for finding verbose text, Key-Value (KV) lookup is for named facts.*

### 4.2. Two-Level Indexing (B+ Tree Philosophy)
The solution is a two-tiered data structure that separates precision (searching) from recall (context delivery).

#### Database Schema (PostgreSQL + pgvector)
```sql
-- The Parent: Large context window (e.g., 1500 tokens)
CREATE TABLE expert_pages (
    id uuid PRIMARY KEY, 
    course_id uuid, 
    heading_path text, 
    full_text text, 
    token_count int
);

-- The Children: Small, dense pinpoint chunks (e.g., 150 tokens)
CREATE TABLE expert_chunks (
    id uuid PRIMARY KEY, 
    parent_id uuid REFERENCES expert_pages(id), 
    chunk_text text, 
    embedding vector(1536), 
    order_idx int
);

-- Indexing for fast Approximate Nearest Neighbor (ANN)
CREATE INDEX ON expert_chunks USING ivfflat (embedding vector_cosine_ops);
```

#### The Retrieval Flow
1. **Embed the Query**: The user's query is converted to a vector embedding.
2. **Pinpoint Search (Children)**: ANN search on `expert_chunks` using Cosine Similarity (k=20 to 50).
3. **Parent Resolution (O(1) Hash Map)**: Deduplicate the resulting `parent_id`s.
4. **Context Fetch**: Perform a single `SELECT * FROM expert_pages WHERE id IN (...)` to fetch the full 1500-token parents.
5. **Cross-Encoder Rerank**: Rerank the massive parent texts to ensure top relevance.
6. **Delivery**: Provide the Top 3 Parents (3 * 1500 = 4500 tokens of rich context) to the LLM.

### 4.3. Taxonomy & Grounding
- **In-Context Working Memory**: Top 3 Parents.
- **Episodic Memory**: The `heading_path` of the parent.
- **Grounding Requirement**: The system prompt strictly enforces that the LLM must cite the `parent_id` and the `span_start`/`span_end` to prove faithfulness (result ↔ context).

```go
func (r *Reader) SearchWithParent(ctx context.Context, query string) ([]Page, error) {
    qVec := r.deps.Embedder.Embed(ctx, query)
    childIDs := r.deps.DB.ANNChildren(ctx, qVec, 20) // Fast pgvector search
    parentIDs := dedupeByParent(childIDs) // Hashmap deduplication
    parents := r.deps.DB.GetParents(ctx, parentIDs) // Batch fetch
    reranked := r.deps.Gateway.Rerank(ctx, query, parents) // Cross-Encoder validation
    return reranked[:3], nil
}
```

---

## 5. Phase 3: Agentic Reflection Loop (Problem D)

### 5.1. The Single-Shot Fallacy
Single-shot retrieval relies on a "One-Hop Assumption". Complex multi-hop queries (e.g., "How does load balancing interact with database sharding?") will fail because a single ANN query will almost never return both isolated concepts simultaneously. 

### 5.2. Bounded Agentic Loop (Hybrid OTA)
We treat complex queries as a Directed Acyclic Graph (DAG) where nodes are sub-queries. The orchestrator uses Breadth-First Search (BFS) to parallelize sub-query retrieval.

#### Architecture Highlights:
- **LLMs Emit Text, Not Tool Calls**: The "Harness" (Go code) drives the logic. The LLM is merely evaluated for coverage.
- **Priority Heap & RRF**: The system maintains a priority heap of the best parents. After every hop, results are merged using Reciprocal Rank Fusion (RRF). Score formula: `score = Σ 1/(k+rank)` (where k≈60).
- **Monotonic Predicate**: The loop runs until a threshold is met (e.g., coverage_score >= 0.85). This behaves like binary search on an answer.

#### The Harness Code Logic
```go
func (o *Orchestrator) AnswerAgentic(ctx context.Context, q string) (string, error) {
    // 1. Query Planning
    plan := o.deps.Gateway.Complete(ctx, "Decompose query into sub-queries JSON", q) 
    
    // 2. Parallel Initial Retrieval
    retrieved := o.reader.SearchWithParentBatch(ctx, plan.SubQueries) 
    
    // 3. The Bounded Loop (MaxHops prevents infinite spend)
    for hops := 0; hops < MaxHops; hops++ { 
        // Cheap classifier (temp 0.1) evaluates coverage
        verdict := o.deps.Gateway.Cheap(ctx, `Coverage check: query vs context. Return JSON {sufficient:bool, missing:string}`, 0.1)
        
        if verdict.Sufficient { break }
        
        // Generate targeted follow-up for missing context
        followUp := o.deps.Gateway.Cheap(ctx, "Generate follow-up query for missing: "+verdict.Missing, 0.6)
        more := o.reader.SearchWithParent(ctx, followUp)
        
        // Merge using Reciprocal Rank Fusion
        retrieved = mergeRRF(retrieved, more) 
        
        // Checkpoint to Postgres (Idempotency + Observability)
        o.deps.DB.InsertEvent(ctx, sessionID, "hop", retrieved)
    }
    
    // 4. Final Generation heavily grounded in cited parents
    final := o.deps.Gateway.Complete(ctx, SystemPrompt+` Use only provided parents, cite parent_id`, retrieved)
    return final, nil
}
```

### 5.3. Observability & Circuit Breaking
- **Unit Economics**: A "Cheap" check call (200 tokens) prevents a massive 10k-token hallucinated generation.
- **Circuit Breaker**: If the loop detects the identical "missing query" twice in a row, the loop breaks to prevent multi-agent deadlocks.
- **Fail-Closed Pessimism**: If the Gateway throws an error mid-loop, the system falls back to returning the partial context with an honest "I'm not completely sure" disclaimer.

---

## 6. Production Hardening: The 4 Critical Tweaks

To make this architecture production-grade, 4 specific tweaks were implemented. These are not optional—they are the difference between a prototype and an enterprise system.

### Tweak 1: Target N Unique Parents Loop (Fixing Context Starvation)
- **The Bug**: If we pull the top 10 children, and 8 of them belong to the exact same parent, deduplication results in only 1 or 2 unique parents being fetched. This starves the LLM of context.
- **The Fix**: We introduced a dynamic scan loop. `targetUniqueParents` is configured (e.g., to 3). The system scans the ranked children (up to K=50) until it successfully collects 3 *unique* parent IDs.
- **Performance**: Scanning a hashset up to 50 items is an O(K) operation that takes microseconds in Go, but guarantees the LLM receives the full requested context volume.

### Tweak 2: Strict Context Cancellation (`ctx` → `pgx`)
- **The Bug**: Using `errgroup.WithContext` with an 800ms timeout is useless if that context isn't aggressively passed down to the database driver. This leads to connection pool leaks where Postgres continues computing abandoned queries.
- **The Fix**: Every `deps.DB.Query` explicitly receives the timeout `ctx`. When Go's timeout fires (`ctx.Done()`), the pgx driver immediately sends a `CancelRequest` to Postgres.
- **Verification**: `pg_stat_activity` must show zero stuck queries, and the DB connection pool's `acquiredConns` metric will remain stable, preventing connection exhaustion under load.

### Tweak 3: App-Side Semantic Caching (Day-Zero Cost Saver)
- **The Concept**: Why re-run expensive LLM orchestration if the identical semantic question was asked 5 minutes ago?
- **The Avoided Trap**: Setting up RediSearch was deemed infrastructure overkill (high cost, high ops maintenance).
- **The App-Side Fix**: We maintain a lightweight in-memory cache (or basic Redis LIST) containing the last 200 queries per course/domain. 
    - Payload: `{embedding []float32, finalText string, createdAt}`
    - When a new query arrives, we embed it. In Go, we perform a brute-force cosine similarity (dot product) against the 200 cached vectors.
    - Since `O(200 * 1536)` floating-point operations take less than ~0.3ms in Go, this is effectively free compute.
    - If `cosine > 0.98` (Configurable `SemanticThreshold`), we return the cached answer immediately.
- **Business Rule**: "Embeddings are cheap, LLMs are expensive."

### Tweak 4: 50 vs 20 Children Retrieval (Optimizing Cross-Encoders)
- **The Math**: Running a cross-encoder on huge parent chunks (1500 tokens) is computationally crushing (Quadratic scaling: 5 * 1500² ≈ 11.2M ops). 
- **The Optimization**: Instead, we retrieve the Top 50 *children* from the vector DB. We run the cross-encoder *only on the children* (which are small, 150 tokens each). 50 * 180² ≈ 1.6M ops. 
- **Result**: It is 7x cheaper to accurately rank the children first, and then map the top unique children to their respective parents. This increases recall while drastically lowering latency and compute costs.

---

## 7. The Final End-to-End Execution Flow

When a user asks a question, the data flows exactly as follows:

```javascript
// PHASE 1: INGESTION
Ingest (A): Markdown AST Parsing → Construct Parent(1500 tokens) → Split to Children(150 tokens) 
            + Semantic Overlap (2 lines) + Enforce cfg.Soft/Hard limits
            → Store in expert_pages + expert_chunks (pgvector ivfflat indexing)

// PHASE 2: RETRIEVAL
Retrieve (C): Query Embed (qVec) → ANN Top 50 children 
              → Cross-Encoder rerank children (Safe 512 token limit) 
              → Scan loop until 3 Unique Parents identified 
              → Fetch Full Parents (SQL IN clause) 
              → Prune priority heap (Top 3 RRF)

// PHASE 3: INFERENCE & AGENTIC LOOP
Infer (D): Semantic Cache app-side check (Threshold 0.98) 
           → Cache Miss 
           → Errgroup parallel sub-queries (Strict ctx→pgx timeout 800ms) 
           → Merge results using RRF (k≈60) 
           → MaxHops Coverage Classifier via Cheap Gateway (Temp 0.1) 
           → Generate followUp query (if coverage insufficient) 
           → Heap Prune 
           → Final LLM Complete (Strict instruction to cite parent_id spans)
```

## 8. Execution Order & Rollout Strategy
To safely implement this into production without regressions, the architecture will be deployed in this exact sequence:

1. **`knowledge/reader.go`**: Implement Fix 1 + Tweak 1+4 (Fetch 50 children, scan loop till 3 unique parents).
2. **`orchestrator.go`**: Implement Fix 2 + Tweak 2 (Introduce errgroup and strict `ctx` propagation to the database).
3. **`orchestrator.go`**: Implement Fix 4 (Heap pruning) and Fix 3 (App-side semantic caching for 200 items).
4. **Evaluation Harness**: Integrate regression tests asserting `unique_parents@3`, monitoring `pool_active_conns`, and tracking `cache_hit_rate`.

**Conclusion**: This architecture relies on no magic. It systematically addresses retrieval flaws through robust Data Structures, tiered storage logic, and deterministic Go orchestration, completely eliminating the reliance on monolithic, blind vector dumps.
