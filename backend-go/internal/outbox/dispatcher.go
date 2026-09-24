package outbox

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Handler processes one claimed event. Return nil to mark published.
type Handler func(ctx context.Context, aggregateType, eventType string, aggregateID string, payload []byte) error

// Dispatcher periodically drains pending outbox rows.
type Dispatcher struct {
	store    *Store
	handler  Handler
	logger   *zap.Logger
	interval time.Duration
	batch    int
}

// NewDispatcher builds a dispatcher. handler may be nil (log-only Stage 0).
func NewDispatcher(store *Store, handler Handler, logger *zap.Logger) *Dispatcher {
	if handler == nil {
		handler = func(ctx context.Context, aggregateType, eventType, aggregateID string, payload []byte) error {
			logger.Debug("outbox event (no subscriber)",
				zap.String("aggregate_type", aggregateType),
				zap.String("event_type", eventType),
				zap.String("aggregate_id", aggregateID),
				zap.Int("payload_bytes", len(payload)),
			)
			return nil
		}
	}
	return &Dispatcher{
		store:    store,
		handler:  handler,
		logger:   logger,
		interval: 2 * time.Second,
		batch:    50,
	}
}

// Run blocks until ctx is cancelled, draining the outbox on an interval.
func (d *Dispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	d.logger.Info("outbox dispatcher started",
		zap.Duration("interval", d.interval),
		zap.Int("batch", d.batch),
	)
	for {
		select {
		case <-ctx.Done():
			d.logger.Info("outbox dispatcher stopped")
			return
		case <-ticker.C:
			d.drain(ctx)
		}
	}
}

func (d *Dispatcher) drain(ctx context.Context) {
	claimed, err := d.store.ClaimPending(ctx, d.batch)
	if err != nil {
		d.logger.Warn("outbox claim failed", zap.Error(err))
		return
	}
	for _, r := range claimed {
		hErr := d.handler(ctx, r.AggregateType, r.EventType, r.AggregateID.String(), r.Payload)
		if hErr != nil {
			d.logger.Warn("outbox handle failed",
				zap.String("event_type", r.EventType),
				zap.String("id", r.ID.String()),
				zap.Error(hErr),
			)
			_ = d.store.MarkFailed(ctx, r.ID, r.Attempts+1, hErr)
			continue
		}
		if err := d.store.MarkPublished(ctx, r.ID); err != nil {
			d.logger.Warn("outbox mark published failed", zap.Error(err))
			continue
		}
	}
}
