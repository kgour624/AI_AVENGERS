// Package business — Business layer, no net/http import (RULE 8-A:28).
package business

import (
	"context"
	"fmt"
)

// ExpertService enforces China Wall isolation per expert (WHERE expert_id=$1).
// FIXED: Business never imports App (mcpv2). Uses its own Storer + types (Ultimate Go §1,§7).
type ExpertService struct {
	db Storer
}

func NewExpertService(db Storer) *ExpertService { return &ExpertService{db: db} }

// ValidateExpert ensures expert exists and is active.
func (s *ExpertService) ValidateExpert(ctx context.Context, expertID string) (Expert, error) {
	if expertID == "" {
		return Expert{}, fmt.Errorf("expertId required")
	}
	return s.db.GetExpert(ctx, expertID)
}