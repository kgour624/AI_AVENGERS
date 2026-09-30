package mcpv2

import (
	"context"
	"sync"
)

// Subscriptions handles SubscribeRequestSchema + notifications/resources/updated ( §3.4 )
// In-memory Set<string> for local; Durable Object storage for Cloudflare worker — same interface.

type ResourceUpdatedNotifier interface {
	NotifyUpdated(ctx context.Context, uri string) error
}

type SubscriptionStore struct {
	mu       sync.RWMutex
	subs     map[string]bool // uri -> subscribed
	notifier ResourceUpdatedNotifier
}

func NewSubscriptionStore(notifier ResourceUpdatedNotifier) *SubscriptionStore {
	return &SubscriptionStore{subs: make(map[string]bool), notifier: notifier}
}

// Subscribe adds uri to Set — handler calls on SubscribeRequestSchema
func (s *SubscriptionStore) Subscribe(uri string) {
	s.mu.Lock()
	s.subs[uri] = true
	s.mu.Unlock()
}

// Unsubscribe removes uri
func (s *SubscriptionStore) Unsubscribe(uri string) {
	s.mu.Lock()
	delete(s.subs, uri)
	s.mu.Unlock()
}

// IsSubscribed checks if uri is watched
func (s *SubscriptionStore) IsSubscribed(uri string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subs[uri]
}

// NotifyUpdated sends notifications/resources/updated if subscribed — caller does db.subscribe -> notification
func (s *SubscriptionStore) NotifyUpdated(ctx context.Context, uri string) error {
	if !s.IsSubscribed(uri) {
		return nil
	}
	if s.notifier != nil {
		return s.notifier.NotifyUpdated(ctx, uri)
	}
	return nil
}