package mcpv2

import (
	"context"
	"sync"
)

// DynamicRegistry handles enable/disable + notifications/list_changed ( §3.4 )
// 2 levels: TYPE available vs INSTANCE list (template list callback) — notify on every create/delete.

type ListChangedNotifier interface {
	NotifyListChanged(ctx context.Context, listType string) error // listType: tools|prompts|resources
}

// DynamicState tracks enabled state per list type.
type DynamicState struct {
	mu       sync.RWMutex
	enabled  map[string]bool
	notifier ListChangedNotifier
}

func NewDynamicState(notifier ListChangedNotifier) *DynamicState {
	return &DynamicState{
		enabled:  map[string]bool{"tools": true, "prompts": true, "resources": true},
		notifier: notifier,
	}
}

// Disable marks type unavailable and notifies clients (auto notifications/tools/list_changed).
func (d *DynamicState) Disable(ctx context.Context, listType string) error {
	d.mu.Lock()
	d.enabled[listType] = false
	d.mu.Unlock()
	if d.notifier != nil {
		return d.notifier.NotifyListChanged(ctx, listType)
	}
	return nil
}

// Enable marks type available and notifies.
func (d *DynamicState) Enable(ctx context.Context, listType string) error {
	d.mu.Lock()
	d.enabled[listType] = true
	d.mu.Unlock()
	if d.notifier != nil {
		return d.notifier.NotifyListChanged(ctx, listType)
	}
	return nil
}

// IsEnabled checks if type is currently enabled.
func (d *DynamicState) IsEnabled(listType string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.enabled[listType]
}