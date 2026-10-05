package receptionist

import (
    "context"
    "fmt"
    "strings"
    "time"
    "github.com/google/uuid"
    "go.uber.org/zap"
    "ai_avengers/backend/internal/gateway"
)

type FinalSynthesizer struct {
    store     *Store
    publisher *Publisher
    gw        *gateway.ModelGateway
    logger    *zap.Logger
}

func NewFinalSynthesizer(store *Store, pub *Publisher, gw *gateway.ModelGateway, logger *zap.Logger) *FinalSynthesizer {
    return &FinalSynthesizer{store: store, publisher: pub, gw: gw, logger: logger}
}

// SynthesizeFinal - All checkpoints ke approved expert_responses ko merge (1+n*k+1 pattern)
func (f *FinalSynthesizer) SynthesizeFinal(ctx context.Context, sessionID, tenantID uuid.UUID) (string, string, error) {
    sess, _ := f.store.GetSession(ctx, sessionID)

    // 0. Notes check - baad me wale discuss huye ya nahi
    notes, _ := f.store.ListNotes(ctx, sessionID)
    if len(notes) > 0 {
        // warning only, end allow but SSE pe reminder
        f.store.AppendEvent(ctx, sessionID, tenantID, "live", map[string]any{"msg": fmt.Sprintf("⚠️ %d notes abhi baki hai, discuss kar lo", len(notes))})
    }

    // 1. Collect all expert responses per checkpoint
    cps, _ := f.store.ListCheckpoints(ctx, sessionID)
    var allMDs []string
    totalCost := 0.0
    totalLatency := 0
    for _, cp := range cps {
        resps, _ := f.store.ListExpertResponses(ctx, cp.ID)
        for _, r := range resps {
            if r.Status=="APPROVED" || r.Status=="DONE" {
                allMDs = append(allMDs, fmt.Sprintf("## Checkpoint %d (%s)\n%s", cp.OrderIdx, cp.Text, r.AnswerMD))
                totalCost += r.CostUsd
                totalLatency += r.LatencyMs
            }
        }
    }
    if len(allMDs)==0 {
        return "", "", fmt.Errorf("no expert responses to synthesize")
    }
    joined := strings.Join(allMDs, "\n\n---\n\n")

    // 2. Strong synthesis - Task+Constraint+Output Format (Byte-by-Byte + Arpit)
    f.store.AppendEvent(ctx, sessionID, tenantID, "typing", map[string]any{"msg": "Final synthesis bana rahi hu, 2 sec..."})
    f.publisher.NotifyNewEvent(ctx, sessionID.String(), 0)
    t0 := time.Now()

    systemPrompt := fmt.Sprintf(`You are Brilliant Secretary finalizer. Task: Merge %d expert markdowns for agenda "%s" into ONE complete English markdown.
Constraint: Keep all approved points, no loss, verifiable, headings per checkpoint, code blocks preserved, no technical kachra.
Output Format: Full English markdown only, with Hinglish 4-line summary at end after "--- Hinglish ---".`, len(allMDs), sess.Agenda)
    userPrompt := fmt.Sprintf(`Expert Markdowns to merge:\n%s\n\nNotes to consider: %v`, joined[:min(12000, len(joined))], notes)

    req := gateway.LLMRequest{
        Model: gateway.ModelStrong, SystemPrompt: systemPrompt, UserPrompt: userPrompt,
        Temperature: 0.2, MaxTokens: 4000, // Arpit: synthesis 0-0.2
    }
    resp, err := f.gw.Call(ctx, req)
    if err != nil || resp.Content=="" {
        f.logger.Error("final synthesis failed", zap.Error(err))
        return "", "", fmt.Errorf("synthesis failed: %w", err)
    }
    finalMD := resp.Content
    // split Hinglish part if present
    hinglish := ""
    if idx := strings.Index(finalMD, "--- Hinglish ---"); idx != -1 {
        hinglish = strings.TrimSpace(finalMD[idx+16:])
        finalMD = strings.TrimSpace(finalMD[:idx])
    } else {
        // cheap fallback 2 lines
        cheap, _ := f.gw.Call(ctx, gateway.LLMRequest{Model: gateway.ModelCheap, SystemPrompt: "Summarize Hinglish 4 lines mother-like pyara", UserPrompt: finalMD[:min(1500,len(finalMD))], Temperature: 0.6, MaxTokens: 300})
        if cheap != nil { hinglish = cheap.Content }
    }
    if hinglish=="" { hinglish = "Aapka document taiyaar hai, sab points cover ho gaye!" }

    latency := int(time.Since(t0).Milliseconds()) + totalLatency
    // 3. Upsert Postgres truth
    f.store.UpsertFinal(ctx, sessionID, tenantID, finalMD, hinglish, totalCost, latency)

    ev, _ := f.store.AppendEvent(ctx, sessionID, tenantID, "live", map[string]any{"msg": "Final ready!", "final_md": finalMD, "hinglish": hinglish})
    f.publisher.NotifyNewEvent(ctx, sessionID.String(), ev.ID)
    f.logger.Info("final synthesis done", zap.String("session_id", sessionID.String()), zap.Int("latency_ms", latency), zap.Float64("cost", totalCost))

    return finalMD, hinglish, nil
}

// NOTE: no local min() helper - Go 1.22 builtin min() is used.
// (duplicate package-level min() declarations did not compile)

// EndConversation - phase change + bye greeting
func (f *FinalSynthesizer) EndConversation(ctx context.Context, sessionID, tenantID uuid.UUID) (string, string, error) {
    // Phase check contiguous
    total, _, contiguous, _ := f.store.GetProgress(ctx, sessionID)
    if contiguous < total {
        // allow but warn - steering: fail-closed nahi, graceful
        f.logger.Info("ending with non-contiguous", zap.Int("contiguous", contiguous), zap.Int("total", total))
    }
    md, hing, err := f.SynthesizeFinal(ctx, sessionID, tenantID)
    if err != nil { return "", "", err }
    // update phase
    f.store.db.Exec(ctx, `UPDATE receptionist_sessions SET state='COMPLETED', updated_at=now() WHERE id=$1`, sessionID)
    // pyara bye greeting - Temp 0.8 warm (Arpit table: blog/brainstorm 0.8)
    byePrompt := gateway.LLMRequest{
        Model: gateway.ModelCheap, SystemPrompt: "You are warm mother-like secretary, give pyara bye in Hinglish, 1 line, add Dhanyawaad.", 
        UserPrompt: fmt.Sprintf("Session %s ended, agenda %s, give bye greeting", sessionID, "done"), Temperature: 0.8, MaxTokens: 100,
    }
    byeResp, _ := f.gw.Call(ctx, byePrompt)
    bye := "Dhanyawaad, aapka PRD taiyaar hai! Phir milenge, khayal rakhna 🙏"
    if byeResp != nil && byeResp.Content != "" { bye = byeResp.Content }
    ev, _ := f.store.AppendEvent(ctx, sessionID, tenantID, "phase_change", map[string]any{"phase": "PHASE_3_ENDED", "bye": bye, "final_md": md})
    f.publisher.NotifyNewEvent(ctx, sessionID.String(), ev.ID)
    return md + "\n\n---\n\n" + bye, hing + " | " + bye, nil
}
