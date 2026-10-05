// Line 1
package receptionist

import (
    "context"
    "time"
    "github.com/google/uuid"
)

// Line 8
type NotesService struct {
    store *Store
}

// Line 12
func NewNotesService(s *Store) *NotesService { return &NotesService{store: s} }

// Line 14
func (n *NotesService) Append(ctx context.Context, sessionID uuid.UUID, checkpointID *uuid.UUID, typ NoteType, summary string) error {
    entry := NoteEntry{
        ID: uuid.New(), SessionID: sessionID, CheckpointID: checkpointID, Type: typ, SummaryText: summary, Timestamp: time.Now().UTC(),
    }
    return n.store.AppendNote(ctx, entry)
}

// Line 22
func (n *NotesService) GetAll(ctx context.Context, sessionID uuid.UUID) ([]NoteEntry, error) {
    return n.store.GetNotes(ctx, sessionID)
}
