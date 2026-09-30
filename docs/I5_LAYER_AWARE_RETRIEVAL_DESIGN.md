# I5 — Layer-Aware Retrieval (Direct Implementation Design)

> Status: **DESIGN READY — NOT YET IMPLEMENTED** (jaanbujh kar chhoda gaya)
> Purpose: Future me bina soche direct copy-paste karke implement ho jaye. RRF + Graph naapne ke baad isko alag se A/B karna hai.

---

## 1. Context — Kyun Chhoda Gaya (Purane Developer Ka Note)

> ⚠️ I5 ka ek hissa maine jaanbujh kar chhoda
> Layer-aware retrieval (matlab "kya tootega" poochne pe layer-3 chunks pehle laana) — maine nahi kiya.
> Kyun: wo retrieval ka teesra change hota. Pichle do (RRF, graph) abhi naape hi nahi gaye tumhare environment me. Teesra bhi jodta to teen changes mix ho jaate aur phir kabhi pata na chalta kisne fayda diya. Aur — is waqt ke liye — depth ka sawaal coverage report se hi pura jawab mil jaata hai, jo bina koi risk liye sach bolta hai.

**Matlab:**
- I4 tak done hai: RRF (`retrieval_rank.go` rrfK=60) + Keyword OR semantics + Fusion
- Graph expansion (`GetCourseChunksExpanded`) done par naapa nahi gaya
- I5 (Layer-aware) = 3rd change hota → isliye pending rakha
- Temporary truth: `DepthLayerReport` / `layerFindings()` coverage report se hi bata deta hai ki kaunse topic me layer-3 nahi hai — bina retrieval risk ke.

---

## 2. Current Architecture Map (As-Is)

```
User Question
  → assembler.Assemble()  [assembler.go: fan-out 6 goroutines]
    → getCourseChunks(ctx, expertID, question, topK, expand=false, pref={})
        1. salientTerms() + keyword query (OR, max 8 terms, stopwords removed)
        2. vector search (pgvector) → Top K
        3. keyword search (tsvector @@ tsquery) → Top K
        4. reciprocalRankFusion(lists, rrfK=60) → fused order
        5. sidecar.Rerank(fused) → rescored (fallback: fused order as-is if sidecar down)
        6. applyPreferenceBoost(scores, matches, 1.25) → reorder only (poolFactor=3)
        7. trimChunksToBudget(chunksBudget 35%) → return
  → Budget enforce → LLM prompt
```

**Key files:**
- `backend-go/internal/context/assembler.go` — `Assemble()`, `getCourseChunks()`, `GetCourseChunksForWorkflow()`, `GetCourseChunksExpanded()`, `GetCourseChunksWithPreference()`
- `backend-go/internal/context/retrieval_preference.go` — `RetrievalPreference{Section, Layer, Boost}`, `DefaultPreferenceBoost=1.25`, `retrievalPoolFactor=3`, `applyPreferenceBoost()` (pure, reorder-only)
- `backend-go/internal/context/retrieval_rank.go` — `reciprocalRankFusion()`, `salientTerms()`, `rrfK=60`
- `backend-go/internal/training/depth_layers.go` — `Layer 1=WHAT/WHY, 2=HOW/TRADE-OFFS, 3=FAILURE/EDGE`, `Report()`, `layerFindings()`
- `backend-go/internal/training/capability_eval.go` — `CapabilityRetriever` interface with 3 isolated methods (measured separately)

**Critical invariant (never break):**
`RetrievalPreference` is **reorder-only, never filter**. Every candidate stays, only order changes. This is what makes it safe to leave on. A wrong Layer-3 detection can never hide the correct answer.

---

## 3. Future Integration Design (Copy-Paste Ready)

### 3A. File 1: `backend-go/internal/context/retrieval_preference.go` — Add intent detector

Add after `Matches()` method, before `applyPreferenceBoost()`:

```go
// layerAwareTriggers are substrings that signal the user wants failure/edge content.
// Matched case-insensitively as substring, same style as Section matching.
// WHY substring not token: "what will break under load?" and "breakage" both match "break".
var layerAwareTriggers = []string{
	// Hindi (user's language)
	"kya tootega", "tootega", "toot", "kya fail", "fail hoga", "kharab", "limit",
	// English failure/edge
	"what breaks", "what will break", "break", "failure", "fail", "edge case", "edge",
	"limitation", "limit", "incident", "outage", "crash", "debug", "troubleshoot",
	"trade-off", "tradeoff", "bottleneck", "degrade", "fallback", "war story",
}

// InferLayerPreference returns Layer=3 preference when question intent is about
// what breaks / limitations / failure modes. Otherwise returns empty (inactive).
// Pure — unit-testable, no DB/model call. Caller decides whether to apply it.
func InferLayerPreference(question string) RetrievalPreference {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return RetrievalPreference{}
	}
	for _, trig := range layerAwareTriggers {
		if strings.Contains(q, trig) {
			return RetrievalPreference{Layer: LayerFailure, Boost: DefaultPreferenceBoost}
		}
	}
	return RetrievalPreference{}
}

// LayerFailure re-export for context package callers without importing training.
// Value must match training.LayerFailure (=3).
const LayerFailure = 3
```

**Alternative (if you want to keep single source of truth):** import `training.LayerFailure` instead of re-export. Current codebase avoids `context → training` import (see `ConceptNeighbourFinder` interface comment). So re-export is preferred to keep dependency directed.

**Tests to add in `retrieval_preference_test.go`:**
```go
func TestInferLayerPreference(t *testing.T) {
  tests := []struct{ q string; wantLayer int }{
    {"kya tootega is system me?", 3},
    {"What will break under load?", 3},
    {"explain failure modes of redis", 3},
    {"what is consistent hashing?", 0}, // no trigger → inactive
    {"how does sharding work?", 0},
    {"", 0},
  }
  for _, tc := range tests {
    got := InferLayerPreference(tc.q)
    if got.Layer != tc.wantLayer { t.Errorf(...) }
  }
}
```

### 3B. File 2: `backend-go/internal/context/assembler.go` — Add wrapper + wire in Assemble

**Add new exported wrapper (same pattern as existing 3):**

```go
// GetCourseChunksWithLayerAware is retrieval with automatic Layer-3 nudge for
// "what breaks" intent. Separate method so a pass can measure it against the
// baseline (same reason as GetCourseChunksExpanded / GetCourseChunksWithPreference).
func (a *Assembler) GetCourseChunksWithLayerAware(ctx context.Context, expertID uuid.UUID, question string, topK int) ([]chinawall.CourseChunk, error) {
	pref := InferLayerPreference(question)
	if !pref.Active() {
		return a.getCourseChunks(ctx, expertID, question, topK, false, RetrievalPreference{})
	}
	// Active → fetch 3x pool so boosted chunks outside topK can be promoted
	return a.getCourseChunks(ctx, expertID, question, topK, false, pref)
}
```

**Wire in chat path (opt-in, default OFF for measurement):**

In `Assemble()` goroutine g5, change from:
```go
c, err := a.getCourseChunks(ctx, expertID, question, a.chunksTopK, false, RetrievalPreference{})
```
To measurement-safe branching (feature flag via system_settings or env):

```go
// Layer-aware is OFF by default. Enable via system_settings key `layer_aware_enabled=true`
// after RRF+graph are measured. Keeps measurement clean.
useLayerAware := false // read from system_settings / config when ready
var c []chinawall.CourseChunk
var err error
if useLayerAware {
  c, err = a.GetCourseChunksWithLayerAware(ctx, expertID, question, a.chunksTopK)
} else {
  c, err = a.getCourseChunks(ctx, expertID, question, a.chunksTopK, false, RetrievalPreference{})
}
```

**Why flag OFF by default:** Same principle as `GetCourseChunksExpanded` being opt-in. Lets you run A/B: baseline vs layer-aware in separate eval passes and attribute the delta to exactly one change.

### 3C. File 3: `backend-go/internal/training/capability_eval.go` — Add 4th isolated method

```go
type CapabilityRetriever interface {
	GetCourseChunksForWorkflow(ctx context.Context, expertID uuid.UUID, taskDescription string, topK int) ([]chinawall.CourseChunk, error)
	GetCourseChunksExpanded(ctx context.Context, expertID uuid.UUID, taskDescription string, topK int) ([]chinawall.CourseChunk, error)
	GetCourseChunksWithPreference(ctx context.Context, expertID uuid.UUID, question string, topK int, pref appcontext.RetrievalPreference) ([]chinawall.CourseChunk, error)
	// GetCourseChunksWithLayerAware — I5: auto Layer-3 nudge for "what breaks" intent.
	// Separate method so eval can measure baseline vs graph vs preference vs layer-aware
	// in isolation and the difference is attributable to one change.
	GetCourseChunksWithLayerAware(ctx context.Context, expertID uuid.UUID, question string, topK int) ([]chinawall.CourseChunk, error)
}
```

Then in eval switch:
```go
switch {
case layerAware:
  chunks, err = e.retriever.GetCourseChunksWithLayerAware(ctx, expertID, c.Question, topK)
case layerPreference:
  chunks, err = e.retriever.GetCourseChunksWithPreference(ctx, expertID, c.Question, topK, preferenceForLevel(c.Level))
case graphExpansion:
  chunks, err = e.retriever.GetCourseChunksExpanded(ctx, expertID, c.Question, topK)
default:
  chunks, err = e.retriever.GetCourseChunksForWorkflow(ctx, expertID, c.Question, topK)
}
```

No new migration needed — `course_chunks.layer` column already exists (migration 051+ `chunk_section_path` / depth layers).

---

## 4. Files & Changes Checklist

| File | Change | Lines | Risk |
|------|--------|-------|------|
| `backend-go/internal/context/retrieval_preference.go` | Add `layerAwareTriggers`, `InferLayerPreference()`, `LayerFailure` const | ~30 | Low — pure, testable |
| `backend-go/internal/context/assembler.go` | Add `GetCourseChunksWithLayerAware()` + flagged branch in `Assemble()` | ~20 | Low — opt-in, default off |
| `backend-go/internal/training/capability_eval.go` | Add 4th interface method + eval switch case | ~10 | Low |
| `backend-go/internal/context/retrieval_preference_test.go` | Add `TestInferLayerPreference` | ~25 | — |
| `backend-go/internal/context/assembler_budget_test.go` | Add layer-aware budget test (chunks still within 35%) | ~15 | — |

**Total: 3 prod files, 2 test files, 0 migrations.**

---

## 5. Rollout Protocol (Strict — Measure Before Merge)

1. **Measure RRF + Graph first** in your environment (hit_rate, MRR, NDCG on golden set). Do not start I5 until those numbers are logged.
2. **Branch for I5:** `feat/i5-layer-aware-retrieval` off main.
3. **Eval pass 1 (offline):** Run `CapabilityRetriever` golden set with `layerAware=true` vs baseline. Expect: "kya tootega" queries promote Layer-3, neutral queries unchanged.
4. **Eval pass 2 (online shadow):** Enable flag for 10% traffic, log `layer_aware_applied` vs not, compare reranker scores before/after boost.
5. **Merge only if:** Layer-aware improves `failure`-tagged queries without regressing neutral queries. If it regresses, tighten trigger list — never widen blindly.
6. **Coverage report stays:** `DepthLayerReport` / `layerFindings()` remains the honest fallback. If a topic has 0 Layer-3 chunks, report says so — retrieval cannot invent depth (Option A principle from `depth_layers.go`).

---

## 6. Observability

**Logs (zap):**
- `layer-aware preference applied` fields: `question_preview`, `layer=3`, `boost=1.25`
- `layer-aware not applied` when `InferLayerPreference` returns inactive (debug level)

**Metrics (if you add prometheus):**
- `layer_aware_queries_total` (counter, label `applied=true/false`)
- `layer_aware_promoted_total` (counter, when boosted chunk entered topK that was outside topK before boost)

**Debug:** `retrieval_preference_test.go` pins trigger list so future edits cannot silently widen it.

---

## 7. Risks & Fallbacks

| Risk | Mitigation |
|------|------------|
| False positive trigger promotes Layer-3 wrongly | Preference is reorder-only; correct chunk still in pool, just one rank lower. No filter = safe. Tighten trigger list if seen. |
| Trigger misses Hindi variation | Add variant to `layerAwareTriggers` — substring match catches inflections (`toot`, `tootega`). Add after seeing real miss in logs. |
| `layer` NULL chunks (unclassified) | `Matches()` checks `layer == 3`; NULL never matches → no boost. Run `DepthClassifier.Classify()` to reduce unclassified. |
| Mixing 3 changes hides attribution | This doc keeps I5 as isolated 4th method — eval attributes delta to one change only. |

**Fallback is already built:** If Layer-aware is off or misses, the answer still works — it just may cite Layer-1/2. And `DepthLayerReport.Findings` honestly tells buyer: *"X topics have NO failure-mode content — expert will explain but cannot answer 'what breaks under load'."* That sentence is the risk-free truth the old developer left you with.

---

## 8. Direct Implementation Steps (When You Say Go)

```bash
# 1. Create branch
git checkout -b feat/i5-layer-aware-retrieval

# 2. Edit 3 files as per §3A/3B/3C (copy-paste blocks above)

# 3. Tests
go test ./internal/context -run TestInferLayerPreference -v
go test ./internal/context -v
go test ./internal/training -run TestCapability -v

# 4. Build
go vet ./...
npm run build  # frontend unchanged, sanity check

# 5. Eval (requires golden set)
go run ./cmd/eval --retriever layer-aware --golden ./eval/golden.json

# 6. Enable flag in system_settings after RRF+graph measured:
# INSERT INTO system_settings(key, value) VALUES ('layer_aware_enabled', 'true')
#    ON CONFLICT (key) DO UPDATE SET value='true';
```

> After this doc, no design thinking is needed. Just copy §3A→§3C blocks into the 3 files, run tests, and flip the flag after measurement.

---

*Source: `backend-go/internal/context/retrieval_preference.go` (preference only reorders), `retrieval_rank.go` (RRF), `depth_layers.go` (Layer 1/2/3 + Report), `assembler.go` (fan-out + isolated retrieval methods), `Learnings/codebase-patterns.md` (Postgres source of truth, small interfaces, three model layers).*