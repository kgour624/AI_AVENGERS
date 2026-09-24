// Package outbox implements the transactional outbox (RDBMS message broker).
//
// Publish appends events; Dispatcher claims batches with FOR UPDATE SKIP LOCKED
// so multiple API/worker replicas can drain safely. Stage 0 handler is a no-op
// logger — subscribers (metering, cache invalidation) plug in later without
// changing publishers.
package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/ports"
)

// Store persists domain events and implements ports.EventPublisher.
type Store struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

var _ ports.EventPublisher = (*Store)(nil)

// NewStore creates an outbox store.
func NewStore(db *pgxpool.Pool, logger *zap.Logger) *Store {
	return &Store{db: db, logger: logger}
}

// Publish inserts events as pending rows. Safe to call outside a larger tx;
// for atomicity with business writes, use PublishTx.
func (s *Store) Publish(ctx context.Context, events ...ports.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.PublishTx(ctx, tx, events...); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// PublishTx inserts events on an existing transaction (preferred).
func (s *Store) PublishTx(ctx context.Context, tx pgx.Tx, events ...ports.DomainEvent) error {
	for _, ev := range events {
		payload, err := json.Marshal(ev.Payload)
		if err != nil {
			return fmt.Errorf("marshal outbox payload: %w", err)
		}
		if payload == nil {
			payload = []byte("{}")
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO domain_outbox (aggregate_type, aggregate_id, event_type, payload)
			VALUES ($1, $2, $3, $4::jsonb)`,
			ev.AggregateType, ev.AggregateID, ev.EventType, payload,
		)
		if err != nil {
			return fmt.Errorf("insert outbox event %s: %w", ev.EventType, err)
		}
	}
	return nil
}

// row is one claimed outbox record.
type row struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       []byte
	Attempts      int
}

// ClaimPending locks up to limit pending rows (SKIP LOCKED) and returns them.
func (s *Store) ClaimPending(ctx context.Context, limit int) ([]row, error) {
	if limit <= 0 {
		limit = 50
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, attempts
		  FROM domain_outbox
		 WHERE status = 'pending'
		   AND available_at <= NOW()
		 ORDER BY created_at
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claimed []row
	var ids []uuid.UUID
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.AggregateType, &r.AggregateID, &r.EventType, &r.Payload, &r.Attempts); err != nil {
			continue
		}
		claimed = append(claimed, r)
		ids = append(ids, r.ID)
	}
	rows.Close()

	if len(ids) == 0 {
		return nil, tx.Commit(ctx)
	}

	// Bump attempts while held so a crash mid-handle still tracks retries.
	// IN ($n…) avoids uuid[] cast issues (json-style arrays break ::uuid[]).
	idArgs := make([]interface{}, len(ids))
	ph := make([]string, len(ids))
	for i, id := range ids {
		idArgs[i] = id
		ph[i] = fmt.Sprintf("$%d", i+1)
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE domain_outbox
		   SET attempts = attempts + 1
		 WHERE id IN (%s)`, strings.Join(ph, ",")),
		idArgs...,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return claimed, nil
}

// MarkPublished marks an event successfully handled.
func (s *Store) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE domain_outbox
		   SET status = 'published', published_at = NOW(), last_error = NULL
		 WHERE id = $1`,
		id,
	)
	return err
}

// MarkFailed schedules a retry (exponential-ish backoff) or permanent fail.
func (s *Store) MarkFailed(ctx context.Context, id uuid.UUID, attempts int, cause error) error {
	msg := ""
	if cause != nil {
		msg = cause.Error()
		if len(msg) > 500 {
			msg = msg[:500]
		}
	}
	const maxAttempts = 10
	if attempts >= maxAttempts {
		_, err := s.db.Exec(ctx, `
			UPDATE domain_outbox
			   SET status = 'failed', last_error = $2
			 WHERE id = $1`,
			id, msg,
		)
		return err
	}
	// 2^min(attempts,6) seconds, capped ~64s
	shift := attempts
	if shift > 6 {
		shift = 6
	}
	delay := time.Duration(1<<uint(shift)) * time.Second
	_, err := s.db.Exec(ctx, `
		UPDATE domain_outbox
		   SET available_at = NOW() + $2::interval,
		       last_error = $3
		 WHERE id = $1 AND status = 'pending'`,
		id, fmt.Sprintf("%d seconds", int(delay.Seconds())), msg,
	)
	return err
}
