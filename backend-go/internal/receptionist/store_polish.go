package receptionist

import (
    "context"
    "strings"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

// NOTES
func (s *Store) AddNote(ctx context.Context, sessionID, tenantID uuid.UUID, text string) (uuid.UUID, error) {
    var id uuid.UUID
    err := s.db.QueryRow(ctx, `INSERT INTO receptionist_notes(session_id, tenant_id, text) VALUES($1,$2,$3) RETURNING id`, sessionID, tenantID, text).Scan(&id)
    if err != nil { s.logger.Error("add note failed", zap.Error(err)); return uuid.Nil, err }
    return id, nil
}
func (s *Store) ListNotes(ctx context.Context, sessionID uuid.UUID) ([]Note, error) {
    rows, err := s.db.Query(ctx, `SELECT id, text, created_at FROM receptionist_notes WHERE session_id=$1 ORDER BY created_at ASC`, sessionID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []Note
    for rows.Next() {
        var n Note
        rows.Scan(&n.ID, &n.Text, &n.CreatedAt)
        out = append(out, n)
    }
    return out, rows.Err()
}
func (s *Store) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
    _, err := s.db.Exec(ctx, `DELETE FROM receptionist_notes WHERE id=$1`, noteID)
    return err
}
// Note - receptionist_notes row.
//
// FIX: CreatedAt pehle int64 tha, par DB column `created_at TIMESTAMPTZ` hai.
// pgx timestamptz ko int64 me scan nahi kar sakta, isliye ListNotes har baar
//   "cannot scan timestamptz into *int64"
// deta tha aur notes API 200 ke saath khaali list bhejta tha. Ab time.Time.
type Note struct {
    ID        uuid.UUID `json:"id"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
}

// CHECKLIST PROGRESS - Contiguous prefix (Learnings/codebase-patterns.md)
//
// FIX (do bugs ek saath):
//  1. Column `idx` receptionist_checkpoints me exist hi nahi karta -> sach hai
//     `order_idx` (079 ne add kiya).
//  2. Status DB me lowercase hota hai ('committed'), aur 078 ka default 'PENDING'
//     hai — isliye case-insensitive compare (EqualFold) karna hi correct hai,
//     warna committed checkpoint hamesha "not committed" gina jata tha.
//  3. order_idx NULL-safe (COALESCE) taaki 079 se pehle ke rows scan na todein.
func (s *Store) GetProgress(ctx context.Context, sessionID uuid.UUID) (total, committed, contiguous int, nextIdx int) {
    rows, err := s.db.Query(ctx, `
        SELECT COALESCE(order_idx, 0), COALESCE(status, 'pending')
        FROM receptionist_checkpoints
        WHERE session_id=$1
        ORDER BY order_idx ASC`, sessionID)
    if err != nil {
        s.logger.Error("get progress failed", zap.Error(err))
        return 0, 0, 0, 0
    }
    defer rows.Close()
    var cps []struct{ Idx int; Status string }
    for rows.Next() {
        var idx int
        var st string
        if err := rows.Scan(&idx, &st); err != nil {
            s.logger.Error("scan progress row failed", zap.Error(err))
            return 0, 0, 0, 0
        }
        cps = append(cps, struct{ Idx int; Status string }{idx, st})
    }
    if err := rows.Err(); err != nil {
        s.logger.Error("iterate progress rows failed", zap.Error(err))
        return 0, 0, 0, 0
    }
    total = len(cps)
    for _, cp := range cps {
        if isCommittedStatus(cp.Status) {
            committed++
        }
    }
    // contiguous: 0..k sab committed/approved tabhi k+1, gap pe break
    contiguous = 0
    for i, cp := range cps {
        if isCommittedStatus(cp.Status) {
            contiguous = i + 1
        } else {
            break
        }
    }
    nextIdx = contiguous // next discuss idx
    return total, committed, contiguous, nextIdx
}

// isCommittedStatus - postgres me status lowercase rehta hai ('committed'),
// par purane/default rows 'COMMITTED'/'APPROVED' bhi ho sakte hain. Dono chalein.
func isCommittedStatus(status string) bool {
    return strings.EqualFold(status, "committed") || strings.EqualFold(status, "approved")
}

// FINAL SYNTHESIS
func (s *Store) UpsertFinal(ctx context.Context, sessionID, tenantID uuid.UUID, finalMD, hinglish string, cost float64, latency int) error {
    _, err := s.db.Exec(ctx, `
        INSERT INTO receptionist_final_synthesis(session_id, tenant_id, final_md, hinglish_summary, total_cost_usd, total_latency_ms)
        VALUES($1,$2,$3,$4,$5,$6)
        ON CONFLICT (session_id) DO UPDATE SET final_md=$3, hinglish_summary=$4, total_cost_usd=$5, total_latency_ms=$6, created_at=now()`,
        sessionID, tenantID, finalMD, hinglish, cost, latency)
    return err
}
func (s *Store) GetFinal(ctx context.Context, sessionID uuid.UUID) (string, string, error) {
    var md, hing string
    err := s.db.QueryRow(ctx, `SELECT final_md, hinglish_summary FROM receptionist_final_synthesis WHERE session_id=$1`, sessionID).Scan(&md, &hing)
    return md, hing, err
}
