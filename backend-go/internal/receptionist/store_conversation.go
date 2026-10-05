package receptionist

import (
    "context"
    "github.com/google/uuid"
)

// pressure = base 0 + reopen*10 + lowRating*15 + delaySec/10 + dismissedAfterExplain*5
// 0-100 cap, 70+ = High 🔥
func calcPressure(reopenCount, lowRatingCount, dismissedCount int, delaySec int) int {
    s := reopenCount*10 + lowRatingCount*15 + dismissedCount*5 + delaySec/10
    if s > 100 { s = 100 }
    return s
}

func (s *Store) UpdatePressure(ctx context.Context, sessionID uuid.UUID, score int) error {
    _, err := s.db.Exec(ctx, `UPDATE receptionist_sessions SET pressure_score=$1, updated_at=now() WHERE id=$2`, score, sessionID)
    return err
}
