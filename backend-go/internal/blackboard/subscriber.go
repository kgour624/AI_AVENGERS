package blackboard

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// cursorKeyTTL is how long a cursor key lives in Redis without being refreshed.
// 7 days: covers the longest realistic workflow pause (client on vacation).
// After TTL, cursor is lost and subscriber replays from fromSeq.
const cursorKeyTTL = 7 * 24 * time.Hour

// Subscriber consumes blackboard events for a specific expert in a workflow.
//
// Usage:
//   sub := NewSubscriber(store, redisClient, logger)
//   eventCh, errCh := sub.Subscribe(ctx, workflowID, expertID, fromSeq)
//   for {
//     select {
//     case event := <-eventCh: // process event
//     case err := <-errCh:    // handle error / reconnect
//     case <-ctx.Done():      // workflow phase complete
//     }
//   }
type Subscriber struct {
	store  *Store
	redis  *redis.Client
	logger *zap.Logger
}

// NewSubscriber creates a new blackboard subscriber.
func NewSubscriber(store *Store, redisClient *redis.Client, logger *zap.Logger) *Subscriber {
	return &Subscriber{
		store:  store,
		redis:  redisClient,
		logger: logger,
	}
}

// Subscribe starts consuming events for a workflow.
// Returns a channel of events and a channel of fatal errors.
//
// fromSeq: start delivering events with sequence_number > fromSeq.
//   Pass 0 to start from the beginning of the workflow.
//   Pass the last checkpoint's sequence_number to resume mid-phase.
//
// The caller must drain eventCh and errCh until ctx is cancelled.
// Closing ctx stops the subscription cleanly.
func (s *Subscriber) Subscribe(
	ctx context.Context,
	workflowID uuid.UUID,
	expertID uuid.UUID,
	fromSeq int64,
) (<-chan Event, <-chan error) {
	eventCh := make(chan Event, 64) // buffered: don't block on slow consumers
	errCh := make(chan error, 1)

	go s.run(ctx, workflowID, expertID, fromSeq, eventCh, errCh)
	return eventCh, errCh
}

// LoadCursor loads the last-processed sequence_number for an expert in a workflow.
// Returns 0 if no cursor exists (start from beginning).
func (s *Subscriber) LoadCursor(ctx context.Context, workflowID, expertID uuid.UUID) int64 {
	key := cursorKey(workflowID, expertID)
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		// redis.Nil = key doesn't exist = start from 0
		return 0
	}
	cursor, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return cursor
}

// SaveCursor persists the last-processed sequence_number for an expert.
func (s *Subscriber) SaveCursor(ctx context.Context, workflowID, expertID uuid.UUID, seq int64) {
	key := cursorKey(workflowID, expertID)
	if err := s.redis.Set(ctx, key, strconv.FormatInt(seq, 10), cursorKeyTTL).Err(); err != nil {
		// Non-fatal: cursor loss means replay from checkpoint, not data loss.
		s.logger.Warn("blackboard: failed to save cursor",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

// run is the internal goroutine that drives the subscription loop.
//
// Algorithm:
// 1. Catch up: fetch all events since fromSeq from Postgres.
// 2. Subscribe to Redis channel.
// 3. On each Redis notification: fetch new events since cursor from Postgres.
// 4. Deliver events to eventCh, update cursor.
// 5. On Redis error: log, wait 1s, re-subscribe (Redis is notification-only).
func (s *Subscriber) run(
	ctx context.Context,
	workflowID uuid.UUID,
	expertID uuid.UUID,
	fromSeq int64,
	eventCh chan<- Event,
	errCh chan<- error,
) {
	defer close(eventCh)
	defer close(errCh)

	cursor := fromSeq

	// Step 1: Catch up on any events posted before we subscribed.
	// This handles the race between "subscribe" and "events already posted".
	if err := s.catchUp(ctx, workflowID, &cursor, eventCh); err != nil {
		if ctx.Err() == nil {
			errCh <- fmt.Errorf("blackboard subscriber: initial catch-up failed: %w", err)
		}
		return
	}

	// Step 2: Subscribe to Redis channel.
	channel := ChannelName(workflowID)
	pubsub := s.redis.Subscribe(ctx, channel)
	defer pubsub.Close()

	s.logger.Info("blackboard subscriber started",
		zap.String("workflow_id", workflowID.String()),
		zap.String("expert_id", expertID.String()),
		zap.Int64("from_seq", fromSeq),
	)

	for {
		select {
		case <-ctx.Done():
			// Clean shutdown
			s.logger.Info("blackboard subscriber stopped",
				zap.String("workflow_id", workflowID.String()),
				zap.Int64("cursor", cursor),
			)
			return

		case msg, ok := <-pubsub.Channel():
			if !ok {
				// Channel closed — Redis connection lost.
				// Wait and re-subscribe.
				s.logger.Warn("blackboard subscriber: Redis channel closed, reconnecting")
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				// Re-subscribe
				pubsub.Close()
				pubsub = s.redis.Subscribe(ctx, channel)
				// Catch up on events missed during reconnect
				if err := s.catchUp(ctx, workflowID, &cursor, eventCh); err != nil {
					if ctx.Err() == nil {
						s.logger.Warn("blackboard subscriber: catch-up after reconnect failed", zap.Error(err))
					}
				}
				continue
			}

			// Parse notification to get sequence_number hint.
			// We don't use the payload content directly — we fetch from Postgres.
			var notif NotificationPayload
			if err := json.Unmarshal([]byte(msg.Payload), &notif); err != nil {
				s.logger.Warn("blackboard subscriber: malformed notification", zap.Error(err))
				continue
			}

			// Only fetch if this notification is ahead of our cursor.
			// Handles duplicate notifications (Redis at-least-once delivery).
			if notif.SequenceNumber <= cursor {
				continue
			}

			// Step 3: Fetch new events from Postgres since cursor.
			if err := s.catchUp(ctx, workflowID, &cursor, eventCh); err != nil {
				if ctx.Err() == nil {
					s.logger.Warn("blackboard subscriber: catch-up on notification failed", zap.Error(err))
				}
			}

			// Save cursor to Redis after processing.
			s.SaveCursor(ctx, workflowID, expertID, cursor)
		}
	}
}

// catchUp fetches all events since *cursor from Postgres and delivers them.
// Updates *cursor to the last delivered sequence_number.
func (s *Subscriber) catchUp(
	ctx context.Context,
	workflowID uuid.UUID,
	cursor *int64,
	eventCh chan<- Event,
) error {
	events, err := s.store.GetSince(ctx, workflowID, *cursor, 200)
	if err != nil {
		return err
	}
	for _, event := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case eventCh <- event:
			*cursor = event.SequenceNumber
		}
	}
	return nil
}

// cursorKey returns the Redis key for an expert's cursor in a workflow.
func cursorKey(workflowID, expertID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:expert:%s:cursor", redisChannelPrefix, workflowID.String(), expertID.String())
}
