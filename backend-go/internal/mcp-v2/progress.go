package mcpv2

import "context"

// Progress handles long-task progressToken + AbortSignal (MCP 3&4 spec).
// review_codebase / sync_repo / run_workflow — 30s+ tasks

type ProgressToken string

type ProgressNotifier interface {
	NotifyProgress(ctx context.Context, token ProgressToken, progress, total float64) error
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