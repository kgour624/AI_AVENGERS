package receptionist

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type Store struct {
    db *pgxpool.Pool
    redis *redis.Client
    logger *zap.Logger // FIX: store_events.go s.logger.Error/Info use karta hai
}

func NewStore(db *pgxpool.Pool, rdb *redis.Client) *Store {
    return &Store{db: db, redis: rdb, logger: zap.NewNop()}
}

// SetLogger - app wiring se real logger inject karo (nil-safe)
func (s *Store) SetLogger(l *zap.Logger) *Store {
    if l != nil { s.logger = l }
    return s
}

func (s *Store) CreateSession(ctx context.Context, sess *ReceptionistSession) error {
    cursorJSON, _ := json.Marshal(sess.Cursor)
    _, err := s.db.Exec(ctx, `INSERT INTO receptionist_sessions (id, tenant_id, admin_id, language, persona, agenda, state, cursor, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
        sess.ID, sess.TenantID, sess.AdminID, sess.Language, sess.Persona, sess.Agenda, sess.State, cursorJSON, sess.CreatedAt, sess.UpdatedAt)
    return err
}

func (s *Store) GetSession(ctx context.Context, id uuid.UUID) (*ReceptionistSession, error) {
    var sess ReceptionistSession
    var cursorJSON []byte
    err := s.db.QueryRow(ctx, `SELECT id, tenant_id, admin_id, language, persona, agenda, state, cursor, created_at, updated_at FROM receptionist_sessions WHERE id=$1`, id).
        Scan(&sess.ID, &sess.TenantID, &sess.AdminID, &sess.Language, &sess.Persona, &sess.Agenda, &sess.State, &cursorJSON, &sess.CreatedAt, &sess.UpdatedAt)
    if err != nil { return nil, err }
    if len(cursorJSON) > 0 { json.Unmarshal(cursorJSON, &sess.Cursor) }
    return &sess, nil
}

func (s *Store) UpdateSessionState(ctx context.Context, id uuid.UUID, state SessionState, cursor SessionCursor) error {
    cj, _ := json.Marshal(cursor)
    _, err := s.db.Exec(ctx, `UPDATE receptionist_sessions SET state=$1, cursor=$2, updated_at=NOW() WHERE id=$3`, state, cj, id)
    return err
}

func (s *Store) UpdateAgenda(ctx context.Context, id uuid.UUID, agenda string) error {
    _, err := s.db.Exec(ctx, `UPDATE receptionist_sessions SET agenda=$1, updated_at=NOW() WHERE id=$2`, agenda, id)
    return err
}

func (s *Store) CreateCheckpoints(ctx context.Context, items []ChecklistItem) error {
    for _, it := range items {
        _, err := s.db.Exec(ctx, `INSERT INTO receptionist_checkpoints (id, session_id, text, order_idx, status, rating_history) VALUES ($1,$2,$3,$4,$5,$6)`,
            it.ID, it.SessionID, it.Text, it.OrderIdx, it.Status, it.RatingHistory)
        if err != nil { return err }
    }
    return nil
}

func (s *Store) GetCheckpoints(ctx context.Context, sessionID uuid.UUID) ([]ChecklistItem, error) {
    rows, err := s.db.Query(ctx, `SELECT id, session_id, text, order_idx, status, committed_summary, rating_history, created_at FROM receptionist_checkpoints WHERE session_id=$1 ORDER BY order_idx ASC`, sessionID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []ChecklistItem
    for rows.Next() {
        var it ChecklistItem
        var summary sql.NullString
        rows.Scan(&it.ID, &it.SessionID, &it.Text, &it.OrderIdx, &it.Status, &summary, &it.RatingHistory, &it.CreatedAt)
        if summary.Valid { it.CommittedSummary = &summary.String }
        out = append(out, it)
    }
    return out, rows.Err()
}

func (s *Store) UpdateCheckpointStatus(ctx context.Context, id uuid.UUID, status ItemStatus, summary *string, ratingHistory []int) error {
    _, err := s.db.Exec(ctx, `UPDATE receptionist_checkpoints SET status=$1, committed_summary=$2, rating_history=$3 WHERE id=$4`, status, summary, ratingHistory, id)
    return err
}

func (s *Store) GetNextPendingCheckpoint(ctx context.Context, sessionID uuid.UUID) (*ChecklistItem, error) {
    var it ChecklistItem
    var summary sql.NullString
    err := s.db.QueryRow(ctx, `SELECT id, session_id, text, order_idx, status, committed_summary, rating_history, created_at FROM receptionist_checkpoints WHERE session_id=$1 AND status IN ('PENDING','REOPENED') ORDER BY order_idx ASC LIMIT 1`, sessionID).
        Scan(&it.ID, &it.SessionID, &it.Text, &it.OrderIdx, &it.Status, &summary, &it.RatingHistory, &it.CreatedAt)
    if err != nil { return nil, err }
    if summary.Valid { it.CommittedSummary = &summary.String }
    return &it, nil
}

func (s *Store) AppendNote(ctx context.Context, n NoteEntry) error {
    _, err := s.db.Exec(ctx, `INSERT INTO receptionist_notes (id, session_id, checkpoint_id, type, summary, ts) VALUES ($1,$2,$3,$4,$5,$6)`, n.ID, n.SessionID, n.CheckpointID, n.Type, n.SummaryText, n.Timestamp)
    return err
}

func (s *Store) GetNotes(ctx context.Context, sessionID uuid.UUID) ([]NoteEntry, error) {
    rows, err := s.db.Query(ctx, `SELECT id, session_id, checkpoint_id, type, summary, ts FROM receptionist_notes WHERE session_id=$1 ORDER BY ts ASC`, sessionID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []NoteEntry
    for rows.Next() {
        var n NoteEntry
        rows.Scan(&n.ID, &n.SessionID, &n.CheckpointID, &n.Type, &n.SummaryText, &n.Timestamp)
        out = append(out, n)
    }
    return out, rows.Err()
}

func (s *Store) AddExpertCall(ctx context.Context, id, sessionID uuid.UUID, checkpointID *uuid.UUID, expertID uuid.UUID, question, relevant string) error {
    _, err := s.db.Exec(ctx, `INSERT INTO receptionist_expert_calls (id, session_id, checkpoint_id, expert_id, question, relevant_answer) VALUES ($1,$2,$3,$4,$5,$6)`, id, sessionID, checkpointID, expertID, question, relevant)
    return err
}

func (s *Store) RateExpertCall(ctx context.Context, id uuid.UUID, rating int) error {
    _, err := s.db.Exec(ctx, `UPDATE receptionist_expert_calls SET rating=$1 WHERE id=$2`, rating, id)
    return err
}

func (s *Store) DeleteSession(ctx context.Context, id uuid.UUID) error {
    _, err := s.db.Exec(ctx, `DELETE FROM receptionist_sessions WHERE id=$1`, id)
    return err
}

func (s *Store) GetExpertCalls(ctx context.Context, sessionID uuid.UUID) ([]map[string]interface{}, error) {
    rows, err := s.db.Query(ctx, `SELECT id, checkpoint_id, expert_id, question, relevant_answer, rating FROM receptionist_expert_calls WHERE session_id=$1 ORDER BY created_at ASC`, sessionID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []map[string]interface{}
    for rows.Next() {
        var id, cid, eid uuid.UUID
        var q, ans string
        var rating sql.NullInt32
        var cidNull uuid.NullUUID
        rows.Scan(&id, &cidNull, &eid, &q, &ans, &rating)
        m := map[string]interface{}{"id": id, "expert_id": eid, "question": q, "answer": ans}
        if cidNull.Valid { m["checkpoint_id"] = cidNull.UUID }
        if rating.Valid { m["rating"] = int(rating.Int32) }
        _ = cid
        out = append(out, m)
    }
    return out, rows.Err()
}

func (s *Store) VerifyTenant(ctx context.Context, sessionID, tenantID uuid.UUID) (bool, error) {
    var exists bool
    err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM receptionist_sessions WHERE id=$1 AND tenant_id=$2)`, sessionID, tenantID).Scan(&exists)
    return exists, err
}

var _ = fmt.Sprintf
var _ = time.Now

// GetCheckpoint FOR UPDATE ke sath - race fix
func (s *Store) GetCheckpointForUpdate(ctx context.Context, tx pgxTx, sessionID uuid.UUID, sectionKey string) (*Checkpoint, error) {
	row := tx.QueryRow(ctx, `SELECT id, session_id, section_key, status, version, summary_warm, previous_summary_warm, reopen_reason FROM receptionist_checkpoints WHERE session_id=$1 AND section_key=$2 FOR UPDATE`, sessionID, sectionKey)
	var cp Checkpoint
	var reason *string
	var prev []byte
	var warm []byte
	if err := row.Scan(&cp.ID, &cp.SessionID, &cp.SectionKey, &cp.Status, &cp.Version, &warm, &prev, &reason); err != nil {
		return nil, err
	}
	cp.SummaryWarm = warm
	cp.PreviousSummaryWarm = prev
	cp.ReopenReason = reason
	return &cp, nil
}

// ReopenCheckpoint - CORE FIX
// L2 WARM -> L1 HOT promotion + status REOPENED/AMENDED + version bump
func (s *Store) ReopenCheckpoint(ctx context.Context, sessionID uuid.UUID, sectionKey string, reason string) (*Checkpoint, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)

	cp, err := s.GetCheckpointForUpdate(ctx, tx, sessionID, sectionKey)
	if err != nil { return nil, fmt.Errorf("checkpoint not found: %w", err) }

	if cp.Status != StatusCommitted && cp.Status != StatusReopened && cp.Status != StatusAmended {
		return nil, fmt.Errorf("only committed/reopened can be reopened, current=%s", cp.Status)
	}

	newStatus := StatusReopened
	if cp.Status == StatusReopened || cp.Status == StatusAmended {
		newStatus = StatusAmended // dusri baar change to AMENDED
	}

	// 1. WARM -> HOT Promotion (Redis)
	hotKey := fmt.Sprintf("receptionist:hot:%s:%s", sessionID.String(), sectionKey)
	var warmData map[string]interface{}
	_ = json.Unmarshal(cp.SummaryWarm, &warmData)
	warmJSON, _ := json.Marshal(warmData)
	// HOT me 30 min TTL - Secretary ab isi se baat karega
	if err := s.redis.Set(ctx, hotKey, string(warmJSON), 30*time.Minute).Err(); err != nil {
		return nil, fmt.Errorf("redis HOT promote failed: %w", err)
	}

	// 2. Version bump + audit
	now := time.Now()
	_, err = tx.Exec(ctx, `UPDATE receptionist_checkpoints SET status=$1, version=version+1, reopen_reason=$2, reopened_at=$3, previous_summary_warm=summary_warm, updated_at=$3 WHERE id=$4`,
		string(newStatus), reason, now, cp.ID)
	if err != nil { return nil, err }

	// 3. Cursor ko wapas us point par lao - loop tootne se bacha
	_, err = tx.Exec(ctx, `UPDATE receptionist_sessions SET cursor=$1, updated_at=$2 WHERE id=$3`, sectionKey, now, sessionID)
	if err != nil { return nil, err }

	if err := tx.Commit(ctx); err != nil { return nil, err }

	cp.Status = newStatus
	cp.Version++
	cp.ReopenReason = &reason
	return cp, nil
}

// PromoteWarmToHot - L1 patch jo kahi se bhi call ho sakta hai
func (s *Store) PromoteWarmToHot(ctx context.Context, sessionID uuid.UUID, sectionKey string) error {
	var warm []byte
	err := s.db.QueryRow(ctx, `SELECT summary_warm FROM receptionist_checkpoints WHERE session_id=$1 AND section_key=$2`, sessionID, sectionKey).Scan(&warm)
	if err != nil { return err }
	hotKey := fmt.Sprintf("receptionist:hot:%s:%s", sessionID.String(), sectionKey)
	return s.redis.Set(ctx, hotKey, string(warm), 30*time.Minute).Err()
}

func (s *Store) CommitCheckpoint(ctx context.Context, sessionID uuid.UUID, sectionKey string, summaryWarm []byte) error {
	hotKey := fmt.Sprintf("receptionist:hot:%s:%s", sessionID.String(), sectionKey)
	// Commit par HOT delete (Tiered Storage pattern)
	_ = s.redis.Del(ctx, hotKey).Err()
	_, err := s.db.Exec(ctx, `UPDATE receptionist_checkpoints SET status='committed', summary_warm=$1, updated_at=now() WHERE session_id=$2 AND section_key=$3`, summaryWarm, sessionID, sectionKey)
	return err
}

// pgxTx interface taaki transaction mock ho sake â€” pgx/v5 exact signature
type pgxTx interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (s *Store) IsChecklistCommitted(ctx context.Context, sessionID uuid.UUID) bool {
    var count int
    err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM receptionist_checkpoints WHERE session_id=$1 AND LOWER(status)='committed'`, sessionID).Scan(&count)
    if err != nil {
        return false
    }
    return count > 0
}
