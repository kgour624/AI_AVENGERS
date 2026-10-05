package receptionist

import (
 "context"
 "fmt"
 "strings"
)

// --- Structs to satisfy the mental model ---
type GatewayCheap interface {
	Cheap(ctx context.Context, prompt string, temp float64) (string, error)
}

type SearchAdapter interface {
	Search(ctx context.Context, query string) ([]string, error)
}

type Deps struct {
	DB      *ExtendedStore
	Redis   *ExtendedRedis
	Gateway GatewayCheap
}

type Service struct {
	deps          *Deps
	searchAdapter SearchAdapter
	templateStore *TemplateStore
}

type ExtendedStore struct { *Store }
type ExtendedRedis struct { *Publisher }

func NewAppError(code int, codeStr, msg string) error { return fmt.Errorf("%s: %s", codeStr, msg) }
func NewInternal(msg string, err error) error { return fmt.Errorf("%s: %w", msg, err) }

type FakeSession struct { Phase string; Agenda string; ChecklistPoints []string }
type FakeTx struct {}
type FakeCP struct { ID string; PointIdx int; Question string }

func (s *ExtendedStore) GetSessionForUpdate(ctx context.Context, sid string) (*FakeSession, error) { return &FakeSession{Phase: "PHASE_1_SETUP", ChecklistPoints: []string{"test"}}, nil }
func (s *ExtendedStore) Begin(ctx context.Context) (*FakeTx, error) { return &FakeTx{}, nil }
func (s *ExtendedStore) InsertEvent(ctx context.Context, sid, role, msg string) {}
func (s *ExtendedStore) GetNextUnansweredCheckpoint(ctx context.Context, sid string) *FakeCP { return nil }
func (s *ExtendedStore) MarkCheckpointAnswered(ctx context.Context, cid, msg string) {}
func (s *ExtendedStore) CountUnansweredInPoint(ctx context.Context, sid string, pi int) int { return 0 }
func (s *ExtendedStore) CommitPoint(ctx context.Context, sid string, pi int) {}

func (t *FakeTx) Rollback(ctx context.Context) {}
func (t *FakeTx) Commit(ctx context.Context) {}
func (t *FakeTx) InsertCheckpoint(ctx context.Context, sid string, pi, qi int, q string) {}
func (t *FakeTx) UpdatePhase(ctx context.Context, sid string, p string) {}

func (r *ExtendedRedis) Publish(ctx context.Context, ch, msg string) {}

// --- User's exact code below ---

func (s *Service) Done(ctx context.Context, sessionID string) error {
 sess, err := s.deps.DB.GetSessionForUpdate(ctx, sessionID)
 if err != nil { return NewInternal("get session fail", err) }
 if sess.Phase != "PHASE_1_SETUP" { return NewAppError(400, "ALREADY_DONE", "already committed") }
 if len(sess.ChecklistPoints)==0 { return NewAppError(400, "CHECKLIST_NOT_READY", "checklist empty") }
 query := sess.Agenda + " " + strings.Join(sess.ChecklistPoints, " ")
 hits, _ := s.searchAdapter.Search(ctx, query)
 allQs := s.templateStore.BuildCheckpoints(ctx, sess.Agenda, sess.ChecklistPoints, hits)
 tx, _ := s.deps.DB.Begin(ctx)
 defer tx.Rollback(ctx)
 for pi, qs := range allQs {
  for qi, q := range qs {
   tx.InsertCheckpoint(ctx, sessionID, pi, qi, q)
  }
 }
 tx.UpdatePhase(ctx, sessionID, "PHASE_2_CONVERSATION")
 tx.Commit(ctx)
 firstQ := allQs[0][0]
 s.deps.DB.InsertEvent(ctx, sessionID, "assistant", firstQ)
 s.deps.Redis.Publish(ctx, "session:"+sessionID+":event", firstQ)
 return nil
}

func (s *Service) HandleChat(ctx context.Context, sessionID, userMsg string) (string, error) {
 sess, err := s.deps.DB.GetSessionForUpdate(ctx, sessionID)
 if err != nil { return "", NewInternal("get session fail", err) }
 if sess.Phase != "PHASE_2_CONVERSATION" {
  return "", NewAppError(400, "CHECKLIST_NOT_READY", "Pehle Done daba ke checklist lock karo")
 }
 s.deps.DB.InsertEvent(ctx, sessionID, "user", userMsg)
 curr := s.deps.DB.GetNextUnansweredCheckpoint(ctx, sessionID)
 if curr == nil {
  return "Bhai saare points ho gaye 2/2 ✅, ab Ask-Experts bolo?", nil
 }
 s.deps.DB.MarkCheckpointAnswered(ctx, curr.ID, userMsg)
 remaining := s.deps.DB.CountUnansweredInPoint(ctx, sessionID, curr.PointIdx)
 var resp string
 if remaining == 0 {
  s.deps.DB.CommitPoint(ctx, sessionID, curr.PointIdx)
  next := s.deps.DB.GetNextUnansweredCheckpoint(ctx, sessionID)
  summ, _ := s.deps.Gateway.Cheap(ctx, fmt.Sprintf("Maa jaisi Hinglish me Point %d ka short summary de, fear low kar.", curr.PointIdx+1), 0.6)
  if next != nil {
   resp = summ + "\n\nBadhiya bhai, Point " + fmt.Sprint(curr.PointIdx+1) + " clear 1/2 ✅. Ab next point - " + next.Question
  } else {
   resp = summ + "\n\nSaare points 2/2 done 🎉, bolo Ask-Experts kar du?"
  }
 } else {
  nxt := s.deps.DB.GetNextUnansweredCheckpoint(ctx, sessionID)
  prompt := SecretaryPrompt + fmt.Sprintf(` Current Q:"%s" User:"%s" Ek hi sawaal poocho.`, nxt.Question, userMsg)
  r, _ := s.deps.Gateway.Cheap(ctx, prompt, 0.6)
  resp = r
 }
 s.deps.DB.InsertEvent(ctx, sessionID, "assistant", resp)
 s.deps.Redis.Publish(ctx, "session:"+sessionID+":event", resp)
 return resp, nil
}
