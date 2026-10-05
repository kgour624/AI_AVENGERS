package receptionist

import (
    "context"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

type ExpertResponse struct {
    ID              uuid.UUID `json:"id" db:"id"`
    CheckpointID    uuid.UUID `json:"checkpoint_id" db:"checkpoint_id"`
    SessionID       uuid.UUID `json:"session_id" db:"session_id"`
    ExpertID        uuid.UUID `json:"expert_id" db:"expert_id"`
    TenantID        uuid.UUID `json:"tenant_id" db:"tenant_id"` // FIX: InsertExpertResponse r.TenantID use karta hai
    AnswerMD        string    `json:"answer_md" db:"answer_md"`
    HinglishSummary string    `json:"hinglish_summary" db:"hinglish_summary"`
    LatencyMs       int       `json:"latency_ms" db:"latency_ms"`
    CostUsd         float64   `json:"cost_usd" db:"cost_usd"`
    Rating          *int      `json:"rating" db:"rating"`
    Status          string    `json:"status" db:"status"`
}

// Insert - FK NOT NULL enforce, fail-closed
func (s *Store) InsertExpertResponse(ctx context.Context, r ExpertResponse) error {
    _, err := s.db.Exec(ctx, `
        INSERT INTO expert_responses(id, checkpoint_id, session_id, expert_id, tenant_id, answer_md, hinglish_summary, latency_ms, cost_usd, status)
        VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
        r.ID, r.CheckpointID, r.SessionID, r.ExpertID, r.TenantID, r.AnswerMD, r.HinglishSummary, r.LatencyMs, r.CostUsd, r.Status)
    if err != nil {
        s.logger.Error("insert expert response failed", zap.Error(err), zap.String("checkpoint_id", r.CheckpointID.String()))
    }
    return err
}

func (s *Store) ListExpertResponses(ctx context.Context, checkpointID uuid.UUID) ([]ExpertResponse, error) {
    rows, err := s.db.Query(ctx, `SELECT id, checkpoint_id, session_id, expert_id, answer_md, hinglish_summary, latency_ms, cost_usd, rating, status FROM expert_responses WHERE checkpoint_id=$1 ORDER BY created_at ASC`, checkpointID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []ExpertResponse
    for rows.Next() {
        var r ExpertResponse
        rows.Scan(&r.ID, &r.CheckpointID, &r.SessionID, &r.ExpertID, &r.AnswerMD, &r.HinglishSummary, &r.LatencyMs, &r.CostUsd, &r.Rating, &r.Status)
        out = append(out, r)
    }
    return out, rows.Err()
}

func (s *Store) RateExpertResponse(ctx context.Context, responseID uuid.UUID, rating int) error {
    status := "REOPENED"
    if rating > 3 { status = "APPROVED" }
    _, err := s.db.Exec(ctx, `UPDATE expert_responses SET rating=$1, status=$2 WHERE id=$3`, rating, status, responseID)
    return err
}

// Suggest experts - dynamic suggest, client final tick
func (s *Store) SuggestExperts(ctx context.Context, tenantID uuid.UUID, agenda string, limit int) ([]Expert, error) {
    rows, err := s.db.Query(ctx, `SELECT id, name, domain, description FROM experts WHERE tenant_id=$1 ORDER BY times_cited DESC LIMIT $2`, tenantID, limit)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []Expert
    for rows.Next() {
        var e Expert
        rows.Scan(&e.ID, &e.Name, &e.Domain, &e.Description)
        out = append(out, e)
    }
    return out, nil
}

type Expert struct { ID uuid.UUID; Name string; Domain string; Description string }
