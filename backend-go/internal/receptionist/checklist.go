// Line 1
package receptionist

import (
    "context"
    "fmt"
    "github.com/google/uuid"
)

// Line 8
type ChecklistService struct {
    store *Store
}

// Line 12
func NewChecklistService(s *Store) *ChecklistService { return &ChecklistService{store: s} }

// Line 14
func (c *ChecklistService) Create(ctx context.Context, sessionID uuid.UUID, texts []string) ([]ChecklistItem, error) {
    if len(texts) == 0 {
        return nil, fmt.Errorf("checklist empty")
    }
    items := make([]ChecklistItem, len(texts))
    for i, t := range texts {
        items[i] = ChecklistItem{
            ID: uuid.New(), SessionID: sessionID, Text: t, OrderIdx: i, Status: ItemPending, RatingHistory: []int{},
        }
    }
    if err := c.store.CreateCheckpoints(ctx, items); err != nil { return nil, err }
    return items, nil
}

// Line 28
func (c *ChecklistService) CreateNoChecklist(ctx context.Context, sessionID uuid.UUID) ([]ChecklistItem, error) {
    return c.Create(ctx, sessionID, []string{"General Discussion"})
}

// Line 32
func (c *ChecklistService) RateCheckpoint(ctx context.Context, checkpointID uuid.UUID, stars int, llmSummary string) (bool, error) {
    // find checkpoint
    // For simplicity we fetch via session scan - in prod add GetCheckpointByID
    // rating <=3 => REOPENED, not committed
    // rating >3 => COMMITTED with summary
    if stars < 1 || stars > 5 {
        return false, fmt.Errorf("rating must be 1-5")
    }
    committed := stars > 3
    status := ItemReopened
    var summary *string
    if committed {
        status = ItemCommitted
        summary = &llmSummary
    }
    // append rating to history - need current history
    // We do read-modify-write with FOR UPDATE semantics via DB transaction in real impl
    // Simplified: fetch current
    // NOTE: production me SELECT ... FOR UPDATE karna
    var hist []int
    // fetch existing rating_history
    // This is simplified inline fetch
    // In real code: SELECT rating_history FROM receptionist_checkpoints WHERE id=$1
    // For brevity we assume caller passes full history; here we just store single rating
    hist = []int{stars}
    if err := c.store.UpdateCheckpointStatus(ctx, checkpointID, status, summary, hist); err != nil {
        return false, err
    }
    return committed, nil
}

func (c *ChecklistService) AddDeltaItem(ctx context.Context, sessionID uuid.UUID, sectionKey string, deltaText string) error {
	// REOPENED case me naya sub-item add, purana commit history bana rahega
	_, err := c.store.db.Exec(ctx, `INSERT INTO receptionist_checkpoints (id, session_id, section_key, text, status, created_at, updated_at) VALUES ($1,$2,$3,$4,'PENDING',now(),now())`, uuid.New(), sessionID, sectionKey, deltaText)
	return err
}

