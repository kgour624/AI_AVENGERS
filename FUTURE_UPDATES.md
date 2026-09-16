# FUTURE_UPDATES.md — AI Avengers Implementation Guide

> **Single Source of Truth** for all deferred Phase 2 improvements.
> Implementor ko koi decision nahi lena — sab kuch yahan likha hai.
> Sirf padho aur implement karo.

---

## How to Use This File

1. Ek item uthao
2. "Already Exists" section padho — kya reuse karna hai
3. "Exact Implementation" section padho — kya banana hai
4. Implement karo — koi assumption mat karo
5. Done hone ke baad is file mein status update karo

---

## Item 1: Observability — Prometheus Metrics Endpoint

**Status:** `TODO`
**Effort:** 1 day
**Priority:** Implement when: first production deployment ya jab "kya fail hua" ka jawab nahi milta

### Already Exists (Reuse Karo)

```
backend-go/internal/observability/phase_timer.go
  - PhaseTimer struct — Start(phase), Stop(phase), Results() []PhaseResult
  - PhaseResult{Phase string, Duration time.Duration}
  - Already used in: orchestrator/orchestrator.go, training/ingestion_pipeline.go

backend-go/cmd/server/main.go
  - GET /health endpoint already exists (line ~310)
  - gin router already wired

backend-go/internal/gateway/model_gateway.go
  - totalCost atomic.Value — already tracks cumulative LLM cost
  - callCount atomic.Int64 — already tracks total LLM calls
```

### Exact Implementation

**File to create:** `backend-go/internal/observability/metrics.go`

```go
package observability

import (
    "sync/atomic"
    "time"
)

// Metrics is a simple in-memory counter store.
// WHY not Prometheus client library: adds 10MB binary size for 1 user.
// WHY this approach: /metrics endpoint returns JSON — Prometheus can
// scrape JSON with json_exporter, or we switch to prometheus/client_go
// later by just changing the Render() method. Zero caller changes.
//
// SOLID: OCP — adding new metrics doesn't change existing code.
// SOLID: SRP — Metrics only counts, doesn't log or alert.
type Metrics struct {
    llmCallsTotal    atomic.Int64
    llmErrorsTotal   atomic.Int64
    llmCostUSDTotal  atomic.Value // float64
    workflowsStarted atomic.Int64
    workflowsFailed  atomic.Int64
    agentLoopIter    atomic.Int64
}

// Global singleton — same pattern as existing gateway.totalCost
var Global = &Metrics{}

func (m *Metrics) IncLLMCall()           { m.llmCallsTotal.Add(1) }
func (m *Metrics) IncLLMError()          { m.llmErrorsTotal.Add(1) }
func (m *Metrics) IncWorkflowStarted()   { m.workflowsStarted.Add(1) }
func (m *Metrics) IncWorkflowFailed()    { m.workflowsFailed.Add(1) }
func (m *Metrics) IncAgentLoopIter()     { m.agentLoopIter.Add(1) }
func (m *Metrics) AddLLMCost(usd float64) {
    // atomic float64 add — same pattern as gateway.totalCost
    for {
        old := m.llmCostUSDTotal.Load()
        var oldF float64
        if old != nil { oldF = old.(float64) }
        if m.llmCostUSDTotal.CompareAndSwap(old, oldF+usd) { return }
    }
}

// Snapshot returns current metric values as a map.
// Used by the /metrics HTTP handler.
func (m *Metrics) Snapshot() map[string]interface{} {
    cost := float64(0)
    if v := m.llmCostUSDTotal.Load(); v != nil { cost = v.(float64) }
    return map[string]interface{}{
        "llm_calls_total":    m.llmCallsTotal.Load(),
        "llm_errors_total":   m.llmErrorsTotal.Load(),
        "llm_cost_usd_total": cost,
        "workflows_started":  m.workflowsStarted.Load(),
        "workflows_failed":   m.workflowsFailed.Load(),
        "agent_loop_iter":    m.agentLoopIter.Load(),
        "uptime_seconds":     int64(time.Since(startTime).Seconds()),
    }
}

var startTime = time.Now()
```

**File to modify:** `backend-go/cmd/server/main.go`

Existing `/health` route ke baad add karo:
```go
// GET /metrics — Prometheus-compatible JSON metrics
// Add after existing /health route (around line 310)
router.GET("/metrics", func(c *gin.Context) {
    response.OK(c, observability.Global.Snapshot())
})
```

**Files to modify (add IncLLMCall/IncLLMError calls):**
```
backend-go/internal/gateway/model_gateway.go
  - In Call() method, after successful response: observability.Global.IncLLMCall()
  - In Call() method, on error return: observability.Global.IncLLMError()
  - In Call() method, after cost calculation: observability.Global.AddLLMCost(resp.CostUSD)

backend-go/internal/workflow/runner.go
  - In Run(), after engine.Start(): observability.Global.IncWorkflowStarted()
  - In Run(), after engine.Fail(): observability.Global.IncWorkflowFailed()

backend-go/internal/workflow/agent_loop.go
  - In Run() loop, each iteration: observability.Global.IncAgentLoopIter()
```

**Design Patterns Used:**
- Singleton (Global instance)
- Null Object (zero value is valid, no nil checks needed)
- OCP (new metrics = new field + new method, no existing code changes)

---

## Item 2: Circuit Breaker — ML Sidecar Retry Wrapper

**Status:** `TODO`
**Effort:** 2 hours
**Priority:** Implement when: sidecar starts failing in production (check logs for `sidecar` errors)

### Already Exists (Reuse Karo)

```
backend-go/internal/ml/sidecar_client.go
  - SidecarClient struct{baseURL, httpClient, logger}
  - Embed(ctx, []string) ([][]float32, error)
  - Rerank(ctx, query, docs, topK) ([]RerankResult, error)
  - EmbedSingle(ctx, string) ([]float32, error)
  - Already has HTTP timeout via httpClient.Timeout

backend-go/internal/gateway/model_gateway.go
  - tryProvider() already retries 3 times — SAME PATTERN to follow
  - Lines 217-260: retry loop with attempt counter
```

### Exact Implementation

**File to create:** `backend-go/internal/ml/retry.go`

```go
package ml

import (
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
)

// sidecarMaxRetries: 3 attempts total (1 original + 2 retries).
// WHY 3: matches gateway.tryProvider() pattern already in codebase.
// WHY not more: sidecar failures are usually sustained (OOM, crash).
// More retries = more latency, not more success.
const sidecarMaxRetries = 3

// sidecarRetryDelay: wait between retries.
// WHY exponential: 500ms, 1s, 2s — gives sidecar time to recover.
func sidecarRetryDelay(attempt int) time.Duration {
    return time.Duration(500<<uint(attempt)) * time.Millisecond
}

// withRetry runs fn up to sidecarMaxRetries times.
// Returns last error if all attempts fail.
// Generic function — works for Embed, Rerank, EmbedSingle.
//
// Mental execution:
//   attempt 1: fn() -> network error -> wait 500ms
//   attempt 2: fn() -> timeout -> wait 1s
//   attempt 3: fn() -> success -> return result
//   attempt 3: fn() -> error -> return error (all attempts exhausted)
func withRetry[T any](
    ctx context.Context,
    logger *zap.Logger,
    opName string,
    fn func() (T, error),
) (T, error) {
    var lastErr error
    var zero T
    for attempt := 1; attempt <= sidecarMaxRetries; attempt++ {
        result, err := fn()
        if err == nil {
            return result, nil
        }
        lastErr = err
        if attempt < sidecarMaxRetries {
            logger.Warn("sidecar call failed, retrying",
                zap.String("op", opName),
                zap.Int("attempt", attempt),
                zap.Duration("wait", sidecarRetryDelay(attempt)),
                zap.Error(err),
            )
            select {
            case <-ctx.Done():
                return zero, ctx.Err()
            case <-time.After(sidecarRetryDelay(attempt)):
            }
        }
    }
    return zero, fmt.Errorf("sidecar %s failed after %d attempts: %w",
        opName, sidecarMaxRetries, lastErr)
}
```

**File to modify:** `backend-go/internal/ml/sidecar_client.go`

Existing `Embed()` function ko wrap karo:
```go
// BEFORE (existing):
func (c *SidecarClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
    // ... existing HTTP call logic ...
}

// AFTER: rename existing to embedOnce, wrap with retry
func (c *SidecarClient) embedOnce(ctx context.Context, texts []string) ([][]float32, error) {
    // ... existing HTTP call logic unchanged ...
}

func (c *SidecarClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
    return withRetry(ctx, c.logger, "Embed", func() ([][]float32, error) {
        return c.embedOnce(ctx, texts)
    })
}
```

Same pattern for `Rerank()` and `EmbedSingle()`.

**Design Patterns Used:**
- Template Method (withRetry is the template, fn is the step)
- Generic function (Go 1.18+) — type-safe, no interface{} casting
- Same retry pattern as existing gateway.tryProvider()

---

## Item 3: Rate Limiting — Per-User Global Workflow Limit

**Status:** `TODO`
**Effort:** 30 minutes
**Priority:** Implement when: multiple users added to system

### Already Exists (Reuse Karo)

```
backend-go/internal/ratelimit/limiter.go
  - RateLimiter interface: Allow(ctx, key) bool
  - TokenBucketLimiter: NewTokenBucketLimiter(capacity, refillRate)
  - NoopLimiter: always returns true (Null Object pattern)
  - ExpertKey(expertID) string — key builder helper

backend-go/internal/middleware/middleware.go
  - RateLimitMiddleware(redisClient, perIP, logger) — already applied globally
  - Pattern to follow for per-user limit

backend-go/cmd/server/main.go
  - Workflow routes group (line ~432)
  - protected middleware group already has user_id in context
```

### Exact Implementation

**File to modify:** `backend-go/internal/ratelimit/limiter.go`

Add one helper function:
```go
// UserWorkflowKey returns the rate limit key for a user's workflow creation.
// Separate from ExpertKey — different resource, different limit.
func UserWorkflowKey(userID string) string {
    return fmt.Sprintf("user:%s:workflow", userID)
}
```

**File to modify:** `backend-go/cmd/server/main.go`

Workflow POST route pe middleware add karo:
```go
// In buildRouter(), find the workflows group (around line 432):
workflows := protected.Group("/workflows")
{
    // ADD THIS: per-user workflow creation rate limit
    // 5 workflows per minute per user — prevents accidental spam
    // Uses existing TokenBucketLimiter (capacity=5, refill=5/60s)
    workflowLimiter := ratelimit.NewTokenBucketLimiter(5, 5.0/60.0)
    workflowCreateLimit := func(c *gin.Context) {
        userID := c.MustGet("user_id").(uuid.UUID)
        if !workflowLimiter.Allow(c.Request.Context(), ratelimit.UserWorkflowKey(userID.String())) {
            response.TooManyRequests(c, "RATE_LIMIT", "Too many workflow creations. Wait 1 minute.")
            c.Abort()
            return
        }
        c.Next()
    }
    workflows.GET("", wfHandler.ListWorkflows)
    workflows.POST("", workflowCreateLimit, wfHandler.CreateWorkflow)  // limit only on POST
    // ... rest unchanged
}
```

**File to modify:** `backend-go/internal/response/response.go`

Add `TooManyRequests` helper (check if already exists first):
```go
func TooManyRequests(c *gin.Context, code, message string) {
    c.JSON(http.StatusTooManyRequests, gin.H{
        "error": gin.H{"code": code, "message": message},
    })
}
```

**Design Patterns Used:**
- Strategy (RateLimiter interface — swap NoopLimiter for testing)
- Middleware chain (gin middleware pattern already in codebase)

---

## Item 4: SSE Reconnect — useKanbanStream Hook

**Status:** `IN PROGRESS` (useKanbanStream implementation next)
**Effort:** 2 hours
**Priority:** Implement NOW — even for 1 user, 5s polling delay is bad UX

### Already Exists (Reuse Karo)

```
frontend/src/hooks/useIngestionStream.ts
  - EXACT SAME PATTERN to follow
  - EventSource with ?token= auth
  - 3s auto-reconnect on error
  - eventLog array for debugging
  - camelizeKeys() for response normalization

frontend/src/api/workflows.ts
  - KanbanTask interface already defined
  - Workflow interface already defined

backend-go/internal/workflow/kanban_sse.go
  - GET /api/v1/workflows/:id/kanban/stream — SSE endpoint exists
  - Events: kanban_task, kanban_artifact, kanban_plan, kanban_approval, kanban_done, kanban_error

frontend/src/pages/workflows/KanbanPage.tsx
  - Currently uses useQuery with refetchInterval: 5000
  - Replace with useKanbanStream hook
```

### Exact Implementation

**File to create:** `frontend/src/hooks/useKanbanStream.ts`

```typescript
// Input: workflowId string | null
// Output: KanbanStreamState
//
// Internally:
//   1. Open EventSource to /api/v1/workflows/:id/kanban/stream?token=...
//   2. On each SSE event: parse JSON, update tasks map
//   3. On error: close, wait 3s, reconnect (same as useIngestionStream)
//   4. On kanban_done event: close connection, set isDone=true
//
// State shape:
//   tasks: KanbanTask[]          — current kanban board state
//   workflow: Workflow | null    — workflow metadata (from kanban_plan event)
//   isConnected: boolean         — SSE connection alive?
//   isDone: boolean              — workflow completed/failed?
//   eventLog: LogEntry[]         — last 100 events for debugging

import { useEffect, useRef, useState } from 'react'
import type { KanbanTask, Workflow } from '@/api/workflows'
import { useAuthStore } from '@/stores/authStore'
import { camelizeKeys } from '@/utils/casing'

export interface KanbanStreamState {
  tasks: KanbanTask[]
  workflow: Workflow | null
  isConnected: boolean
  isDone: boolean
  eventLog: Array<{ ts: string; type: string; message: string }>
}

export function useKanbanStream(workflowId: string | null): KanbanStreamState {
  const [state, setState] = useState<KanbanStreamState>({
    tasks: [],
    workflow: null,
    isConnected: false,
    isDone: false,
    eventLog: [],
  })

  const esRef = useRef<EventSource | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  // tasksMap: task.id -> KanbanTask — O(1) update by id
  const tasksMapRef = useRef<Map<string, KanbanTask>>(new Map())

  useEffect(() => {
    if (!workflowId) return

    const apiBase = import.meta.env.VITE_API_URL ?? ''
    const url = `${apiBase}/api/v1/workflows/${workflowId}/kanban/stream`

    const addLog = (type: string, message: string) => {
      const ts = new Date().toISOString()
      setState((prev) => ({
        ...prev,
        eventLog: [{ ts, type, message }, ...prev.eventLog].slice(0, 100),
      }))
    }

    const connect = () => {
      if (esRef.current) { esRef.current.close(); esRef.current = null }
      setState((prev) => ({ ...prev, isConnected: false }))

      const token = useAuthStore.getState().accessToken ?? ''
      const sseUrl = token ? `${url}?token=${encodeURIComponent(token)}` : url
      const es = new EventSource(sseUrl, { withCredentials: true })
      esRef.current = es

      es.onopen = () => {
        setState((prev) => ({ ...prev, isConnected: true }))
        addLog('connected', 'SSE connected')
      }

      es.onmessage = (event) => {
        try {
          const raw = JSON.parse(event.data)
          // Backend sends: {type: "kanban_task", data: {...}}
          const { type, data } = camelizeKeys<{ type: string; data: unknown }>(raw)

          if (type === 'kanban_done') {
            setState((prev) => ({ ...prev, isDone: true, isConnected: false }))
            addLog('done', 'Workflow completed')
            es.close(); esRef.current = null
            return
          }

          if (type === 'kanban_error') {
            addLog('error', JSON.stringify(data))
            return
          }

          if (type === 'kanban_task' || type === 'kanban_artifact') {
            // data.content has task status info
            // Parse expert_id + status from content
            const content = data as {
              event_type: string
              posted_by_expert_id?: string
              content?: { expert_id?: string; status?: string }
            }
            // Update task in map if we have expert_id + status
            const expertId = content.content?.expert_id ?? content.posted_by_expert_id
            const status = content.content?.status
            if (expertId && status) {
              const existing = [...tasksMapRef.current.values()]
                .find((t) => t.assignedExpertId === expertId)
              if (existing) {
                const updated = { ...existing, status: status as KanbanTask['status'] }
                tasksMapRef.current.set(existing.id, updated)
                setState((prev) => ({
                  ...prev,
                  tasks: [...tasksMapRef.current.values()],
                }))
              }
            }
            addLog(type, `${content.event_type ?? type}`)
          }

          if (type === 'kanban_plan') {
            // task_plan_ready: build initial task list from plan
            const plan = data as {
              content?: {
                tasks?: Array<{
                  expert_id: string
                  expert_name: string
                  title: string
                  description: string
                }>
              }
            }
            const planTasks = plan.content?.tasks ?? []
            tasksMapRef.current.clear()
            planTasks.forEach((t, i) => {
              const task: KanbanTask = {
                id: `plan-${i}`,
                workflowId: workflowId,
                assignedExpertId: t.expert_id,
                expertName: t.expert_name,
                domain: '',
                title: t.title,
                description: t.description,
                status: 'todo',
                costUsd: 0,
                updatedAt: new Date().toISOString(),
              }
              tasksMapRef.current.set(task.id, task)
            })
            setState((prev) => ({
              ...prev,
              tasks: [...tasksMapRef.current.values()],
            }))
            addLog('plan', `${planTasks.length} tasks planned`)
          }

          if (type === 'kanban_approval') {
            addLog('approval', 'Waiting for approval')
          }
        } catch {
          addLog('parse_error', `Failed to parse: ${event.data.slice(0, 50)}`)
        }
      }

      es.onerror = () => {
        addLog('error', 'SSE disconnected. Reconnecting in 3s...')
        setState((prev) => ({ ...prev, isConnected: false }))
        es.close(); esRef.current = null
        setState((prev) => {
          if (!prev.isDone) {
            retryRef.current = setTimeout(connect, 3000)
          }
          return prev
        })
      }
    }

    connect()
    return () => {
      esRef.current?.close()
      if (retryRef.current) clearTimeout(retryRef.current)
    }
  }, [workflowId])

  return state
}
```

**File to modify:** `frontend/src/pages/workflows/KanbanPage.tsx`

```typescript
// REMOVE these imports:
import { useQuery } from '@tanstack/react-query'
import { getKanban, getWorkflow } from '@/api/workflows'

// ADD this import:
import { useKanbanStream } from '@/hooks/useKanbanStream'

// REMOVE:
const { data: workflow } = useQuery({ queryKey: ['workflow', id], queryFn: () => getWorkflow(id!), ... })
const { data: kanban, isLoading } = useQuery({ queryKey: ['workflow', id, 'kanban'], queryFn: () => getKanban(id!), ... })
const tasks = kanban?.tasks ?? []

// ADD:
const stream = useKanbanStream(id ?? null)
const tasks = stream.tasks
const isLoading = !stream.isConnected && tasks.length === 0
// workflow data comes from stream.workflow (populated from kanban_plan event)
// OR keep a separate useQuery for workflow metadata (title, status, cost)
```

**Design Patterns Used:**
- Same pattern as useIngestionStream (consistency — no new patterns to learn)
- Ref for mutable map (tasksMapRef) — avoids stale closure issues
- Null Object (workflowId=null = no-op)

---

## Item 5: Scalability — Redis Cluster + K8s Config

**Status:** `TODO`
**Effort:** Config change only — NO code change
**Priority:** Implement when: single Redis instance becomes bottleneck (>10k concurrent users)

### Already Exists (Reuse Karo)

```
backend-go/internal/db/redis.go (or similar)
  - Redis client initialized with single URL
  - go-redis library used — cluster-compatible

frontend/.env or docker-compose.yml
  - REDIS_URL environment variable
```

### Exact Implementation

**No code changes needed.** Only config:

```yaml
# docker-compose.yml or K8s ConfigMap
# BEFORE:
REDIS_URL: redis://localhost:6379

# AFTER (Redis Cluster):
REDIS_URL: redis+cluster://node1:6379,node2:6379,node3:6379
```

go-redis automatically handles cluster mode when URL starts with `redis+cluster://`.

**For K8s HPA (Horizontal Pod Autoscaler):**
```yaml
# k8s/deployment.yaml — add resource limits
resources:
  requests:
    cpu: "250m"
    memory: "256Mi"
  limits:
    cpu: "1000m"
    memory: "512Mi"
```

**No application code changes. Zero risk.**

---

## Item 6: Health Checks — Already Complete

**Status:** `DONE`
**File:** `backend-go/cmd/server/main.go`

Already implemented:
```go
router.GET("/health", func(c *gin.Context) {
    health := map[string]string{"status": "ok", "version": "1.0.0"}
    // Checks: postgres.HealthCheck(), redisClient.HealthCheck(), mlClient.HealthCheck()
    // Returns 200 if all healthy, 200 with "degraded" if any unhealthy
})
```

Graceful shutdown already implemented:
```go
// 30s timeout with server.Shutdown(shutdownCtx)
// Drains in-flight requests before stopping
```

**Nothing to do.**

---

## Principles Followed in This File

### SOLID
- **SRP:** Har function ek kaam karta hai (Metrics sirf count karta hai, retry sirf retry karta hai)
- **OCP:** New metrics = new field, no existing code change. New retry target = new wrapper, no withRetry change.
- **DIP:** Callers depend on interfaces (RateLimiter, Embedder), not concrete types

### Design Patterns (GoF)
- **Singleton:** `observability.Global` — same as existing `gateway.totalCost`
- **Template Method:** `withRetry[T]` — algorithm fixed, step (fn) varies
- **Strategy:** `RateLimiter` interface — swap implementations without caller change
- **Null Object:** `NoopLimiter`, `workflowId=null` — eliminates nil checks

### Go Idioms
- Generic function `withRetry[T any]` — type-safe, no interface{} casting
- `atomic.Int64` / `atomic.Value` — lock-free counters (same as existing codebase)
- `select { case <-ctx.Done() }` — context-aware retry (same as existing codebase)

### Frontend Patterns
- `useRef` for mutable state that shouldn't trigger re-renders (tasksMapRef)
- Same hook structure as `useIngestionStream` — consistency over novelty
- `camelizeKeys()` — already used everywhere, keep using

---

## Implementation Order (When Time Comes)

```
1. Item 4 (useKanbanStream)  — IN PROGRESS, highest UX value
2. Item 2 (Sidecar Retry)    — 2 hours, prevents silent failures
3. Item 3 (Rate Limiting)    — 30 min, prevents cost explosion on multi-user
4. Item 1 (Metrics)          — 1 day, needed for production debugging
5. Item 5 (Redis Cluster)    — Config only, do when scaling
6. Item 6 (Health Checks)    — Already done
```

---

*Last updated: 2026-09-16 | Next: Implement Item 4 (useKanbanStream)*
