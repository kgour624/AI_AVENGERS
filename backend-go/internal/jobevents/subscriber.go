package jobevents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// pollInterval is the DB fallback poll cadence used when Redis is unavailable.
// WHY 1s: matches the old snapshot cadence, so a Redis outage degrades the
// stream to the previous behaviour instead of stalling it.
const pollInterval = time.Second

// catchUpLimit is the page size for replaying missed events.
const catchUpLimit = 200

// Subscriber streams a job's events from Postgres, using Redis only as a
// wake-up hint (deliberately identical to blackboard.Subscriber).
//
// Usage:
//
//	sub := jobevents.NewSubscriber(store, redisClient, logger)
//	eventCh, errCh := sub.Subscribe(ctx, jobID, 0)
//	for {
//	  select {
//	  case ev := <-eventCh: // forward to SSE
//	  case <-errCh:         // log; stream continues
//	  case <-ctx.Done():    // client left
//	  }
//	}
type Subscriber struct {
	store  *Store
	redis  *redis.Client
	logger *zap.Logger
}

// NewSubscriber creates a subscriber. redisClient may be nil — the subscriber
// then polls Postgres instead of waiting on pub/sub.
func NewSubscriber(store *Store, redisClient *redis.Client, logger *zap.Logger) *Subscriber {
	return &Subscriber{store: store, redis: redisClient, logger: logger}
}

// Subscribe streams events with sequence_number > fromSeq.
//
// fromSeq semantics:
//   - 0  → full history (new viewer opening the modal)
//   - N  → only what happened after the viewer already saw (reconnect)
//
// Both channels are closed when ctx is cancelled. The caller must keep
// draining until then.
func (s *Subscriber) Subscribe(ctx context.Context, jobID uuid.UUID, fromSeq int64) (<-chan Event, <-chan error) {
	eventCh := make(chan Event, 128) // buffered: a slow SSE writer must not stall catch-up
	errCh := make(chan error, 1)
	go s.run(ctx, jobID, fromSeq, eventCh, errCh)
	return eventCh, errCh
}

// run is the internal goroutine driving the subscription.
//
// Algorithm:
//  1. Catch up from Postgres (covers events written before we subscribed).
//  2. If Redis is available: subscribe and catch up on each notification.
//     Otherwise: poll Postgres on a ticker.
//  3. On Redis channel loss: log, wait, re-subscribe, catch up.
func (s *Subscriber) run(ctx context.Context, jobID uuid.UUID, fromSeq int64, eventCh chan<- Event, errCh chan<- error) {
	defer close(eventCh)
	defer close(errCh)

	cursor := fromSeq

	// Step 1: catch up on anything already written. A failure here is
	// recoverable (the poll/ticker below retries), so it is reported on errCh
	// and the subscription continues rather than dying.
	if err := s.catchUp(ctx, jobID, &cursor, eventCh); err != nil && ctx.Err() == nil {
		select {
		case errCh <- fmt.Errorf("jobevents subscriber: initial catch-up failed: %w", err):
		default: // errCh full — the first failure is the interesting one
		}
	}

	// Step 2a: no Redis → polling fallback (older behaviour, still correct).
	if s.redis == nil {
		s.logger.Info("jobevents subscriber: Redis unavailable, polling Postgres",
			zap.String("job_id", jobID.String()),
		)
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.catchUp(ctx, jobID, &cursor, eventCh); err != nil && ctx.Err() == nil {
					s.logger.Warn("jobevents subscriber: poll catch-up failed",
						zap.String("job_id", jobID.String()), zap.Error(err))
				}
			}
		}
	}

	// Step 2b: Redis available → wait for wake-ups, then read Postgres.
	channel := ChannelName(jobID)
	pubsub := s.redis.Subscribe(ctx, channel)
	defer pubsub.Close()

	s.logger.Info("jobevents subscriber started",
		zap.String("job_id", jobID.String()),
		zap.Int64("from_seq", fromSeq),
	)

	// Safety net: even with Redis, poll occasionally. A dropped pub/sub
	// message would otherwise stall the timeline silently, and this is the
	// surface an admin trusts during a live run.
	safety := time.NewTicker(10 * time.Second)
	defer safety.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-safety.C:
			if err := s.catchUp(ctx, jobID, &cursor, eventCh); err != nil && ctx.Err() == nil {
				s.logger.Warn("jobevents subscriber: safety-net catch-up failed",
					zap.String("job_id", jobID.String()), zap.Error(err))
			}

		case msg, ok := <-pubsub.Channel():
			if !ok {
				// Channel closed → Redis connection lost. Re-subscribe.
				s.logger.Warn("jobevents subscriber: Redis channel closed, reconnecting")
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				pubsub.Close()
				pubsub = s.redis.Subscribe(ctx, channel)
				if err := s.catchUp(ctx, jobID, &cursor, eventCh); err != nil && ctx.Err() == nil {
					s.logger.Warn("jobevents subscriber: catch-up after reconnect failed",
						zap.String("job_id", jobID.String()), zap.Error(err))
				}
				continue
			}

			var notif NotificationPayload
			if err := json.Unmarshal([]byte(msg.Payload), &notif); err != nil {
				s.logger.Warn("jobevents subscriber: malformed notification", zap.Error(err))
				continue
			}
			// Skip notifications we have already consumed (at-least-once
			// delivery from Redis, plus the safety-net ticker).
			if notif.SequenceNumber <= cursor {
				continue
			}
			if err := s.catchUp(ctx, jobID, &cursor, eventCh); err != nil && ctx.Err() == nil {
				s.logger.Warn("jobevents subscriber: catch-up on notification failed",
					zap.String("job_id", jobID.String()), zap.Error(err))
			}
		}
	}
}

// catchUp pages through all events after *cursor and delivers them, advancing
// the cursor. Loops until a short page is returned, so a burst larger than
// catchUpLimit is never truncated.
//
// Returns the first error encountered; the cursor is left at the last delivered
// event so a retry resumes exactly where it stopped (no duplicate, no gap).
func (s *Subscriber) catchUp(ctx context.Context, jobID uuid.UUID, cursor *int64, eventCh chan<- Event) error {
	for {
		events, err := s.store.GetSince(ctx, jobID, *cursor, catchUpLimit)
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
		if len(events) < catchUpLimit {
			return nil
		}
	}
}
