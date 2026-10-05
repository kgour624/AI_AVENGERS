// Line 1
package receptionist

import (
    "time"
    "github.com/google/uuid"
)

// Line 7
type Language string
const (
    LangEN      Language = "EN"
    LangHI      Language = "HI"
    LangINEN    Language = "IN_EN"
    LangHinglish Language = "HINGLISH"
)

// Line 15
type SessionState string
const (
    StateINIT              SessionState = "INIT"
    StatePhase1Setup       SessionState = "PHASE_1_SETUP"
    StatePhase2Conversation SessionState = "PHASE_2_CONVERSATION"
    StatePhase3Wrapup      SessionState = "PHASE_3_WRAPUP"
    StatePhase4SummaryLoop SessionState = "PHASE_4_SUMMARY_LOOP"
    StatePhase5Final       SessionState = "PHASE_5_FINAL"
    StateCompleted         SessionState = "COMPLETED"
    StateDeleted           SessionState = "DELETED"
)

// Line 27
type CursorStatus string
const (
    CursorAwaitingReview CursorStatus = "AWAITING_HUMAN_REVIEW"
    CursorInFlight       CursorStatus = "IN_FLIGHT"
    CursorCommitted      CursorStatus = "COMMITTED"
)

// Line 35
type SessionCursor struct {
    RunID              uuid.UUID  `json:"run_id"`
    ActiveCheckpointID *uuid.UUID `json:"active_checkpoint_id,omitempty"`
    Status             CursorStatus `json:"status"`
    LastAckEventID     int64      `json:"last_ack_event_id"`
}

// Line 42
type ReceptionistSession struct {
    ID        uuid.UUID    `json:"id"`
    TenantID  uuid.UUID    `json:"tenant_id"`
    AdminID   uuid.UUID    `json:"admin_id"`
    Language  Language     `json:"language"`
    Persona   string       `json:"persona"` // immutable after Phase1, e.g. "software engineer"
    Agenda    string       `json:"agenda"`
    State     SessionState `json:"state"`
    Cursor    SessionCursor `json:"cursor"`
    CreatedAt time.Time    `json:"created_at"`
    UpdatedAt time.Time    `json:"updated_at"`
}

// Line 55
type ItemStatus string
const (
    ItemPending    ItemStatus = "PENDING"
    ItemInProgress ItemStatus = "IN_PROGRESS"
    ItemCommitted  ItemStatus = "COMMITTED"
    ItemReopened   ItemStatus = "REOPENED"
)

// Line 63
type ChecklistItem struct {
    ID               uuid.UUID `json:"id"`
    SessionID        uuid.UUID `json:"session_id"`
    Text             string    `json:"text"`
    OrderIdx         int       `json:"order_idx"`
    Status           ItemStatus `json:"status"`
    CommittedSummary *string   `json:"committed_summary,omitempty"`
    RatingHistory    []int     `json:"rating_history"`
    CreatedAt        time.Time `json:"created_at"`
}

// Line 75
type NoteType string
const (
    NoteCheckpointCommit NoteType = "CHECKPOINT_COMMIT"
    NoteExpertConsult    NoteType = "EXPERT_CONSULT"
    NoteChangeRequest    NoteType = "CHANGE_REQUEST"
    NoteRatingReject     NoteType = "RATING_REJECT"
    NoteFileUpload       NoteType = "FILE_UPLOAD"
)

// Line 85
type NoteEntry struct {
    ID           uuid.UUID  `json:"id"`
    SessionID    uuid.UUID  `json:"session_id"`
    CheckpointID *uuid.UUID `json:"checkpoint_id,omitempty"`
    Type         NoteType   `json:"type"`
    SummaryText  string     `json:"summary_text"`
    Timestamp    time.Time  `json:"timestamp"`
}
