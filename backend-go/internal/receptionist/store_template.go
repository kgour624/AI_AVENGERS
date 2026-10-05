package receptionist

import (
    "context"
    "encoding/json"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

func (s *Store) GetTemplate(ctx context.Context, checkpointID uuid.UUID) ([]Point, string, error) {
    var ptsJSON []byte
    var query string
    err := s.db.QueryRow(ctx, `SELECT points, search_query FROM receptionist_templates WHERE checkpoint_id=$1`, checkpointID).Scan(&ptsJSON, &query)
    if err != nil {
        return nil, "", err // pgx.ErrNoRows = empty
    }
    var pts []Point
    json.Unmarshal(ptsJSON, &pts)
    return pts, query, nil
}

func (s *Store) UpsertTemplate(ctx context.Context, checkpointID, sessionID, tenantID uuid.UUID, idx int, query string, pts []Point) error {
    b, _ := json.Marshal(pts)
    _, err := s.db.Exec(ctx, `
        INSERT INTO receptionist_templates(checkpoint_id, session_id, idx, points, search_query, tenant_id)
        VALUES($1,$2,$3,$4,$5,$6)
        ON CONFLICT (checkpoint_id) DO UPDATE SET points=$4, search_query=$5, cached_at=now()`,
        checkpointID, sessionID, idx, b, query, tenantID)
    if err != nil {
        s.logger.Error("upsert template failed", zap.Error(err))
    }
    return err
}

// Clear old checkpoint - next checkpoint pe purana hatao (tune bola: purana clear + naya search)
func (s *Store) DeleteTemplate(ctx context.Context, checkpointID uuid.UUID) error {
    _, err := s.db.Exec(ctx, `DELETE FROM receptionist_templates WHERE checkpoint_id=$1`, checkpointID)
    return err
}

// For contiguous: list order_idx order.
//
// FIX: pehle yahan `SELECT id, idx, title, status ... ORDER BY idx` tha, par
// receptionist_checkpoints me `idx`/`title` naam ka column hi nahi hai (078 + 079
// ke baad sach sirf `order_idx` aur `text` hai). Isliye query ke saath-saath struct
// fields bhi Checkpoint (types.go) ke OrderIdx/Text par shift kiye gaye.
// COALESCE isliye ki 079 se pehle ke rows me order_idx/text NULL ho sakte hain —
// warna rows.Scan(NULL) par pgx error dega.
func (s *Store) ListCheckpoints(ctx context.Context, sessionID uuid.UUID) ([]Checkpoint, error) {
    rows, err := s.db.Query(ctx, `
        SELECT id, COALESCE(order_idx, 0), COALESCE(text, ''), COALESCE(status, 'pending')
        FROM receptionist_checkpoints
        WHERE session_id=$1
        ORDER BY order_idx ASC`, sessionID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []Checkpoint
    for rows.Next() {
        var cp Checkpoint
        if err := rows.Scan(&cp.ID, &cp.OrderIdx, &cp.Text, &cp.Status); err != nil {
            return nil, err
        }
        out = append(out, cp)
    }
    return out, rows.Err()
}
