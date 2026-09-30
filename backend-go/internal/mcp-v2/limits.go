package mcpv2

import (
	"context"
	"fmt"
)

// CheckLimits enforces daily/monthly before LLM call — fail-closed (RULE 8-E:43).
// Returns ErrLimitExceeded if current + estimated > limit.
func (s *Service) CheckLimits(ctx context.Context, expertID, platform string, estimatedTokens int) error {
	limits, err := s.deps.DB.GetExpertLimits(ctx, expertID, platform)
	if err != nil {
		return nil // no limits row = unlimited
	}
	if limits.CurrentDaily+estimatedTokens > limits.DailyLimit {
		return fmt.Errorf("%w: daily %d/%d", ErrLimitExceeded, limits.CurrentDaily, limits.DailyLimit)
	}
	if limits.CurrentMonthly+estimatedTokens > limits.MonthlyLimit {
		return fmt.Errorf("%w: monthly %d/%d", ErrLimitExceeded, limits.CurrentMonthly, limits.MonthlyLimit)
	}
	return nil
}