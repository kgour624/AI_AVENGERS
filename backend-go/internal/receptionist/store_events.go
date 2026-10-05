package receptionist

import (
"context"
"encoding/json"
"time"

"github.com/google/uuid"
"go.uber.org/zap"
)

// ConversationEvent - conversation_events row (Postgres = source of truth)
type ConversationEvent struct {
ID        int64           `json:"id" db:"id"`
SessionID uuid.UUID       `json:"session_id" db:"session_id"`
TenantID  uuid.UUID       `json:"tenant_id" db:"tenant_id"`
Type      string          `json:"type" db:"type"`
Payload   json.RawMessage `json:"payload" db:"payload"`
CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// AppendEvent - Postgres me hi likho, Redis me nahi (durable truth)
func (s *Store) AppendEvent(ctx context.Context, sessionID, tenantID uuid.UUID, typ string, payload any) (*ConversationEvent, error) {
b, _ := json.Marshal(payload)
var ev ConversationEvent
err := s.db.QueryRow(ctx,
`INSERT INTO conversation_events(session_id, tenant_id, type, payload) VALUES($1,$2,$3,$4) RETURNING id, session_id, tenant_id, type, payload, created_at`,
sessionID, tenantID, typ, b,
).Scan(&ev.ID, &ev.SessionID, &ev.TenantID, &ev.Type, &ev.Payload, &ev.CreatedAt)
if err != nil {
s.logger.Error("append event failed", zap.Error(err), zap.String("type", typ))
return nil, err
}
return &ev, nil
}

// ListEvents - SSE replay ke liye sequence based read
func (s *Store) ListEvents(ctx context.Context, sessionID uuid.UUID, afterID int64, limit int) ([]ConversationEvent, error) {
rows, err := s.db.Query(ctx,
`SELECT id, session_id, tenant_id, type, payload, created_at FROM conversation_events
 WHERE session_id=$1 AND id > $2 ORDER BY id ASC LIMIT $3`, sessionID, afterID, limit)
if err != nil {
return nil, err
}
defer rows.Close()
var out []ConversationEvent
for rows.Next() {
var ev ConversationEvent
if err := rows.Scan(&ev.ID, &ev.SessionID, &ev.TenantID, &ev.Type, &ev.Payload, &ev.CreatedAt); err != nil {
return nil, err
}
out = append(out, ev)
}
return out, rows.Err()
}

// PickNextCheckpointsForUpdate - FOR UPDATE SKIP LOCKED picker
// (Learnings/codebase-patterns.md: 5 picker same checkpoint nahi uthayenge).
// FIX: real columns order_idx/text hain (idx/title kisi table me nahi hain) aur
// status lowercase 'active' hai ('DISCUSSING' nahi).
func (s *Store) PickNextCheckpointsForUpdate(ctx context.Context, sessionID uuid.UUID, limit int) ([]Checkpoint, error) {
rows, err := s.db.Query(ctx,
`SELECT id, session_id, section_key, order_idx, text, status FROM receptionist_checkpoints
 WHERE session_id=$1 AND status='active'
 ORDER BY order_idx ASC LIMIT $2 FOR UPDATE SKIP LOCKED`, sessionID, limit)
if err != nil {
return nil, err
}
defer rows.Close()
var out []Checkpoint
for rows.Next() {
var cp Checkpoint
if err := rows.Scan(&cp.ID, &cp.SessionID, &cp.SectionKey, &cp.OrderIdx, &cp.Text, &cp.Status); err != nil {
return nil, err
}
out = append(out, cp)
}
return out, rows.Err()
}