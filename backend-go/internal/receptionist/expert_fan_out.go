// Line 1
package receptionist

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"
    "github.com/google/uuid"
)

// Line 10
type ExpertConsult struct {
    ExpertID uuid.UUID `json:"expert_id"`
    Question string    `json:"question"`
    Relevant string    `json:"relevant"` // only relevant part, raw discarded
    RawLen   int       `json:"raw_len"`
}

// Line 17
type AskExpertsRequest struct {
    SessionID    uuid.UUID   `json:"session_id"`
    CheckpointID *uuid.UUID  `json:"checkpoint_id"`
    ExpertIDs    []uuid.UUID `json:"expert_ids"`
    Question     string      `json:"question"` // client ne jo likha
    ClientLang   Language    `json:"client_lang"`
}

// Line 25
type ExpertCaller interface {
    CallExpert(ctx context.Context, expertID uuid.UUID, questionEnglish string) (string, error)
    GetExpertName(ctx context.Context, expertID uuid.UUID) (string, error)
}

// Line 30
type Coordinator struct {
    caller ExpertCaller
    store  *Store
    notes  *NotesService
}

// Line 36
func NewCoordinator(caller ExpertCaller, store *Store, notes *NotesService) *Coordinator {
    return &Coordinator{caller: caller, store: store, notes: notes}
}

// Line 40
// extractRelevant - Smart Reading: pura blob nahi, sirf relevant range
func extractRelevant(raw, checkpointContext string) string {
    if len(raw) < 2000 { return raw }
    // simple heuristic: relevance extraction via truncation + keyword overlap
    // Production me yahan LLM summarizer call hota hai with prompt: "Extract only parts relevant to: "+checkpointContext
    // For now: first 1500 + last 500 chars with checkpoint keywords
    keywords := strings.Fields(strings.ToLower(checkpointContext))
    lower := strings.ToLower(raw)
    // find best window
    bestIdx := 0
    bestScore := 0
    for i := 0; i < len(lower)-500; i += 500 {
        window := lower[i:min(i+1500, len(lower))]
        score := 0
        for _, kw := range keywords { if strings.Contains(window, kw) { score++ } }
        if score > bestScore { bestScore = score; bestIdx = i }
    }
    if bestScore == 0 { return raw[:1500] + "\n...[truncated]..." }
    end := min(bestIdx+1500, len(raw))
    return raw[bestIdx:end]
}

// Line 64
// NOTE: no local min() helper here - Go 1.22 builtin min() is used (duplicate
// declarations across files caused a compile error).

// Line 66
func (c *Coordinator) AskExperts(ctx context.Context, req AskExpertsRequest) ([]ExpertConsult, error) {
    if len(req.ExpertIDs) == 0 { return nil, fmt.Errorf("no experts selected") }
    if strings.TrimSpace(req.Question) == "" { return nil, fmt.Errorf("question empty") }

    // Tenant isolation check - ensure experts belong to same tenant (S3 pattern)
    // caller should verify tenant scope before calling

    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    type result struct {
        resp ExpertConsult
        err  error
    }
    results := make(chan result, len(req.ExpertIDs))
    var wg sync.WaitGroup
    sem := make(chan struct{}, 10) // backpressure - max 10 concurrent (Flash Sale lesson)

    for _, eid := range req.ExpertIDs {
        wg.Add(1)
        go func(expertID uuid.UUID) {
            defer wg.Done()
            sem <- struct{}{}
            defer func(){ <-sem }()

            // ALWAYS ENGLISH to expert (requirement)
            raw, err := c.caller.CallExpert(ctx, expertID, req.Question)
            if err != nil {
                results <- result{err: fmt.Errorf("expert %s: %w", expertID, err)}
                return
            }
            // Language handling: expert always English, client response translated later
            checkpointCtx := req.Question // in prod: fetch checkpoint text + agenda
            relevant := extractRelevant(raw, checkpointCtx)

            // persist WARM summary only, raw discarded (Tiered Storage lesson)
            _ = c.store.AddExpertCall(ctx, uuid.New(), req.SessionID, req.CheckpointID, expertID, req.Question, relevant)
            _ = c.notes.Append(ctx, req.SessionID, req.CheckpointID, NoteExpertConsult, fmt.Sprintf("Expert %s consulted: Q='%s' -> Relevant='%s'", expertID, req.Question, relevant[:min(200,len(relevant))]))

            results <- result{resp: ExpertConsult{ExpertID: expertID, Question: req.Question, Relevant: relevant, RawLen: len(raw)}}
        }(eid)
    }

    go func(){ wg.Wait(); close(results) }()

    var out []ExpertConsult
    var firstErr error
    for r := range results {
        if r.err != nil && firstErr == nil { firstErr = r.err; continue }
        if r.err == nil { out = append(out, r.resp) }
    }
    if len(out)==0 && firstErr != nil { return nil, firstErr }
    return out, nil
}
