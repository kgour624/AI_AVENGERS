package receptionist

import (
"context"
"fmt"
"sync"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5/pgxpool"
"go.uber.org/zap"

"ai_avengers/backend/internal/gateway"
)

// ExpertFanOut - Business layer (App layer ko import nahi karta).
// NOTE: ye logic pehle expert_fanout.go me tha. Wahan packet-level duplicate
// `ExpertResponse` struct declare ho raha tha (store_expert.go me bhi same naam
// ka struct hai), isliye duplicate file delete ki gayi aur ab store wala
// ExpertResponse hi use hota hai. Koi duplicate min()/max() helper bhi nahi hai
// (Go 1.22 ka builtin min/max chalta hai).
type ExpertFanOut struct {
db        *pgxpool.Pool
store     *Store
publisher *Publisher
gw        *gateway.ModelGateway // Strong tier for answer, Cheap for Hinglish summary
logger    *zap.Logger
}

func NewExpertFanOut(db *pgxpool.Pool, store *Store, pub *Publisher, gw *gateway.ModelGateway, logger *zap.Logger) *ExpertFanOut {
if logger == nil {
logger = zap.NewNop()
}
return &ExpertFanOut{db: db, store: store, publisher: pub, gw: gw, logger: logger}
}

// CallExperts - client-selected expert_ids pe parallel (semaphore 5), full batch wait.
// Client selection final hoti hai, fixed 3 nahi.
func (e *ExpertFanOut) CallExperts(ctx context.Context, sessionID, checkpointID, tenantID uuid.UUID, expertIDs []uuid.UUID, approvedPoints []Point, agenda string) ([]ExpertResponse, error) {
if len(expertIDs) == 0 {
return nil, fmt.Errorf("no experts selected")
}

// semaphore 5 - bounded concurrency, DB connections bachao
sem := make(chan struct{}, 5)
var wg sync.WaitGroup
var mu sync.Mutex
results := make([]ExpertResponse, 0, len(expertIDs))
var firstErr error

_, _ = e.store.AppendEvent(ctx, sessionID, tenantID, "expert_wait", map[string]any{"done": 0, "total": len(expertIDs), "msg": fmt.Sprintf("Experts se baat (0/%d, 0s)", len(expertIDs))})
startAll := time.Now()

for _, eid := range expertIDs {
wg.Add(1)
go func(expertID uuid.UUID) {
defer wg.Done()
sem <- struct{}{}        // acquire
defer func() { <-sem }() // release

ctxChild, cancel := context.WithTimeout(ctx, 15*time.Second)
defer cancel()
t0 := time.Now()

// 1. Expert meta (tenant isolation)
var name, domain, desc string
err := e.db.QueryRow(ctxChild, `SELECT name, domain, description FROM experts WHERE id=$1 AND tenant_id=$2`, expertID, tenantID).Scan(&name, &domain, &desc)
if err != nil {
e.logger.Error("expert not found", zap.Error(err), zap.String("expert_id", expertID.String()))
mu.Lock()
if firstErr == nil {
firstErr = err
}
mu.Unlock()
return
}

// 2. RAG top3 chunks
rows, _ := e.db.Query(ctxChild, `SELECT chunk_text FROM course_chunks WHERE expert_id=$1 ORDER BY times_cited DESC LIMIT 3`, expertID)
ragContext := ""
if rows != nil {
defer rows.Close()
for rows.Next() {
var chunk string
_ = rows.Scan(&chunk)
if len(chunk) > 800 {
chunk = chunk[:800]
}
ragContext += chunk + "\n---\n"
}
}
if ragContext == "" {
ragContext = "No additional context, use your expertise."
}

// 3. Task + Constraint + Output Format
systemPrompt := fmt.Sprintf("You are %s, domain %s. %s. Task: Answer checklist for agenda \"%s\" using approved points. Constraint: Use RAG context only, verifiable, English markdown, no fluff. Output Format: Full markdown with headings per point, code blocks if needed.", name, domain, desc, agenda)
userPrompt := fmt.Sprintf("Approved Points: %v\nRAG Context:\n%s\nInstruction: Give complete markdown answer per point, synthesis quality.", approvedPoints, ragContext)

resp, cerr := e.gw.Call(ctxChild, gateway.LLMRequest{
Model:        gateway.ModelStrong,
SystemPrompt: systemPrompt,
UserPrompt:   userPrompt,
Temperature:  0.2,
MaxTokens:    3000,
})
if cerr != nil || resp == nil || resp.Content == "" {
e.logger.Error("gateway call failed", zap.Error(cerr), zap.String("expert", name))
mu.Lock()
if firstErr == nil {
firstErr = fmt.Errorf("expert %s failed", name)
}
mu.Unlock()
return
}

// 4. Cheap Hinglish 2-line summary
summaryResp, _ := e.gw.Call(ctxChild, gateway.LLMRequest{
Model:        gateway.ModelCheap,
SystemPrompt: "Summarize in Hinglish 2 lines, mother-like, pyaar se.",
UserPrompt:   resp.Content[:min(1000, len(resp.Content))],
Temperature:  0.6,
MaxTokens:    200,
})
hinglish := ""
if summaryResp != nil {
hinglish = summaryResp.Content
}

latency := int(time.Since(t0).Milliseconds())
cost := resp.CostUSD

// 5. Postgres truth (FK NOT NULL)
er := ExpertResponse{
ID: uuid.New(), CheckpointID: checkpointID, SessionID: sessionID, TenantID: tenantID,
ExpertID: expertID, AnswerMD: resp.Content, HinglishSummary: hinglish,
LatencyMs: latency, CostUsd: cost, Status: "DONE",
}
if err := e.store.InsertExpertResponse(ctxChild, er); err != nil {
mu.Lock()
if firstErr == nil {
firstErr = err
}
mu.Unlock()
return
}

// 6. Progress event + SSE notify
mu.Lock()
results = append(results, er)
done := len(results)
mu.Unlock()

elapsed := int(time.Since(startAll).Seconds())
ev, _ := e.store.AppendEvent(ctx, sessionID, tenantID, "expert_wait", map[string]any{"done": done, "total": len(expertIDs), "elapsed": elapsed, "expert": name})
if ev != nil {
_ = e.publisher.NotifyNewEvent(ctx, sessionID.String(), ev.ID)
}
expertLatency.WithLabelValues(name).Observe(float64(latency))
e.logger.Info("expert done", zap.String("expert", name), zap.Int("latency_ms", latency))
}(eid)
}

// full batch wait - adhura nahi (client ko poora answer chahiye)
wg.Wait()

if len(results) == 0 && firstErr != nil {
return nil, firstErr
}

_, _ = e.store.AppendEvent(ctx, sessionID, tenantID, "live", map[string]any{"msg": fmt.Sprintf("All %d experts answered, cards ready", len(results))})
return results, nil
}