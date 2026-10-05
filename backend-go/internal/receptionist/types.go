package receptionist

import (
"time"

"github.com/google/uuid"
)

type CheckpointStatus string

const (
StatusPending   CheckpointStatus = "pending"
StatusActive    CheckpointStatus = "active"
StatusCommitted CheckpointStatus = "committed"
StatusReopened  CheckpointStatus = "reopened" // FIX: pehli baar committed ko wapas khola
StatusAmended   CheckpointStatus = "amended"  // FIX: reopened par hi dusra change
)

func (s CheckpointStatus) IsTerminal() bool { return s == StatusCommitted }
func (s CheckpointStatus) IsReopened() bool { return s == StatusReopened || s == StatusAmended }

// Checkpoint - receptionist_checkpoints row ka public shape.
// FIX: OrderIdx + Text add kiye, kyunki DB columns order_idx/text hi hain
// (pehle code IDx/Title use kar raha tha jo kisi bhi table me nahi hai).
type Checkpoint struct {
ID                  uuid.UUID        `json:"id"`
SessionID           uuid.UUID        `json:"session_id"`
SectionKey          string           `json:"section_key"` // auth, db, scale
OrderIdx            int              `json:"order_idx"`
Text                string           `json:"text"`
Status              CheckpointStatus `json:"status"`
Version             int              `json:"version"`
SummaryWarm         []byte           `json:"summary_warm"` // L2 WARM JSON
SummaryHotKey       string           `json:"summary_hot_key"`
ReopenReason        *string          `json:"reopen_reason"`
ReopenedAt          *time.Time       `json:"reopened_at"`
PreviousSummaryWarm []byte           `json:"previous_summary_warm"`
CreatedAt           time.Time        `json:"created_at"`
UpdatedAt           time.Time        `json:"updated_at"`
}

// Session - legacy narrow view (kept for compatibility with older call sites).
type Session struct {
ID            uuid.UUID `json:"id"`
UserID        uuid.UUID `json:"user_id"`
CurrentCursor string    `json:"current_cursor"` // active section_key
Status        string    `json:"status"`
CreatedAt     time.Time `json:"created_at"`
}