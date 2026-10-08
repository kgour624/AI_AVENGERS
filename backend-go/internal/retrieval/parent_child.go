package retrieval

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ParentSearchTimeout is the DB + rerank budget for parent-child ANN.
// 3000ms: 800ms timed out on complex queries + heavy DB. Graceful fallback
// on timeout keeps chat working (legacy path) instead of 500.
const ParentSearchTimeout = 3000 // milliseconds, kept as const for single source

// IsTimeoutError matches timeout-flavored errors that must be soft fallback,
// not 500. Mirrors context.isTimeoutError — single source lives here so both
// context and knowledge callers never drift.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline") ||
		strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "i/o timeout")
}

// IsGracefulTimeout reports whether err is a deadline/timeout that should
// trigger graceful fallback (return nil,nil, warn) instead of bubbling 500.
func IsGracefulTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || IsTimeoutError(err)
}

// LogDedupeShortfall warns when deduped parents are fewer than requested.
// Additive observability only — never returns error. Helps frontend decide
// to show "only 1 section matched" hint instead of silent 1/3.
func LogDedupeShortfall(logger *zap.Logger, parentIDs []uuid.UUID, childrenCount, target int, expertID uuid.UUID) {
	if logger == nil {
		return
	}
	if len(parentIDs) < target && childrenCount > 0 {
		logger.Warn("parent-child dedupe shortfall — fewer unique parents than target",
			zap.String("expert_id", expertID.String()),
			zap.Int("unique_parents", len(parentIDs)),
			zap.Int("target_parents", target),
			zap.Int("children_scanned", childrenCount),
		)
	}
}

// LogOrphanParents warns when expert_pages is missing for a parent_id.
// Silent skip hides data corruption; log makes it visible as metric.
func LogOrphanParents(logger *zap.Logger, parentIDs []uuid.UUID, found map[uuid.UUID]bool, expertID uuid.UUID) {
	if logger == nil || len(parentIDs) == 0 {
		return
	}
	for _, pid := range parentIDs {
		if !found[pid] {
			logger.Warn("parent-child orphan parent_id — expert_pages row missing",
				zap.String("expert_id", expertID.String()),
				zap.String("parent_id", pid.String()),
			)
		}
	}
}
