package mcpv2

import (
	"context"
	"sync"
)

// Progress handles long-task progressToken + AbortSignal (MCP 3&4 spec).
// review_codebase / sync_repo / run_workflow — 30s+ tasks

type ProgressToken string

type ProgressNotifier interface {
	NotifyProgress(ctx context.Context, token ProgressToken, progress, total float64) error
}

// CancellationManager — in-memory sync.Map for notifications/cancelled -> context.WithCancel
// Key = string(req.ID) or _meta.progressToken alias. Single-pod prod-safe, no Redis yet.
type CancellationManager struct {
	m sync.Map // key(string) -> context.CancelFunc
}

func (cm *CancellationManager) Store(id string, cancel context.CancelFunc) {
	if id == "" || cancel == nil {
		return
	}
	cm.m.Store(id, cancel)
}

func (cm *CancellationManager) CancelAndDelete(id string) bool {
	if id == "" {
		return false
	}
	v, ok := cm.m.LoadAndDelete(id)
	if !ok {
		return false
	}
	if cancel, ok := v.(context.CancelFunc); ok && cancel != nil {
		cancel()
		return true
	}
	return false
}

func (cm *CancellationManager) Delete(id string) {
	if id == "" {
		return
	}
	cm.m.Delete(id)
}

// ReportProgress sends notifications/progress — caller gets token from _meta.progressToken
func ReportProgress(ctx context.Context, notifier ProgressNotifier, token ProgressToken, progress, total float64) error {
	if notifier == nil || token == "" {
		return nil
	}
	return notifier.NotifyProgress(ctx, token, progress, total)
}

// IsCanceled checks AbortSignal — signal.addEventListener("abort", ()=>kill) pattern
func IsCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}