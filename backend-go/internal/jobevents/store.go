// Package jobevents is the durable, append-only event log for transcript
// ingestion jobs — the "training transparency" timeline.
//
// WHY a dedicated package (deliberately mirrors internal/blackboard):
//   The writer lives in internal/training and the readers live in
//   internal/admin (SSE + history endpoint). Training must not import admin
//   (that would be an import cycle: admin already imports training), so the
//   shared type goes in a small package both can depend on.
//
// DESIGN (same shape as blackboard, on purpose — consistency over novelty):
//   * Postgres is the source of truth. Every event is a row.
//   * Redis is notification-only. If Redis is down, events are still durable
//     and a subscriber catches up by replaying from the DB.
//   * sequence_number is BIGSERIAL: monotonic, so a subscriber cursor is just
//     "give me everything with sequence_number > N".
//
// FAILURE POLICY: writing an event must never break ingestion. Append returns
// an error, but every production caller treats it as non-fatal (log + drop) —
// the pipeline is the product, the timeline is observability.
package jobevents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// redisChannelPrefix is the prefix for all ingestion event channels.
// Full channel name: ingestion:{job_id}:events
// WHY constant: changing it would silently break every live subscriber.
const redisChannelPrefix = "ingestion"

// Event kinds. Kept as constants (not free strings) so a typo cannot silently
// create an event the frontend does not know how to render.
const (
	// KindRunStarted: the pipeline began (or resumed) a run for this job.
	KindRunStarted = "run_started"
	// KindStageStarted: a pipeline stage began.
	KindStageStarted = "stage_started"
	// KindStageDone: a pipeline stage finished, with its duration.
	KindStageDone = "stage_done"
	// KindBatchDone: one unit of parallel work finished (topic/embed batch).
	KindBatchDone = "batch_done"
	// KindChunkStored: a block of chunks was committed to the database.
	KindChunkStored = "chunk_stored"
	// KindVerified: the DB was queried to confirm (or contradict) what the
	// pipeline claims it produced. This is the "double confirmation" record.
	KindVerified = "verified"
	// KindPaused: job stopped waiting for admin action.
	KindPaused = "paused"
	// KindFailed: job stopped on an unrecoverable error.
	KindFailed = "failed"
	// KindComplete: job finished successfully.
	KindComplete = "complete"
)

// Event is one row of ingestion_job_events.
type Event struct {
	ID             uuid.UUID       `json:"id"`
	JobID          uuid.UUID       `json:"job_id"`
	ExpertID       uuid.UUID       `json:"expert_id"`
	SequenceNumber int64           `json:"sequence_number"`
	Stage          string          `json:"stage"`
	Kind           string          `json:"kind"`
	Detail         json.RawMessage `json:"detail"`
	CreatedAt      time.Time       `json:"created_at"`
}

// AppendRequest is the input to Store.Append.
type AppendRequest struct {
	JobID    uuid.UUID
	ExpertID uuid.UUID
	Stage    string
	Kind     string
	// Detail is JSON-marshalled. Pass a struct or map — never a pre-encoded
	// string (that would double-encode).
	Detail interface{}
}

// NotificationPayload is what gets published to Redis. Subscribers use it only
// as a wake-up hint; the authoritative payload is always fetched from Postgres.
type NotificationPayload struct {
	EventID        uuid.UUID `json:"event_id"`
	JobID          uuid.UUID `json:"job_id"`
	Kind           string    `json:"kind"`
	SequenceNumber int64     `json:"sequence_number"`
	CreatedAt      time.Time `json:"created_at"`
}

// Store is the single write path for ingestion job events.
//
// WHY single write path: guarantees the Redis notification always follows the
// Postgres write, so a subscriber can never be woken for a row that is not yet
// readable (which would make it silently skip that event).
type Store struct {
	db     *pgxpool.Pool
	redis  *redis.Client // nil-safe: publish is skipped when Redis is absent
	logger *zap.Logger
}

// NewStore creates a new event store. redisClient may be nil (development /
// Redis outage): events are still written to Postgres, subscribers fall back
// to polling.
func NewStore(db *pgxpool.Pool, redisClient *redis.Client, logger *zap.Logger) *Store {
	return &Store{db: db, redis: redisClient, logger: logger}
}

// Append writes one event and notifies subscribers (best-effort).
//
// Mental execution:
//   Input: {job=J, stage="embedding", kind="batch_done", detail={index:3}}
//   1. INSERT ... RETURNING id, sequence_number, created_at
//   2. Publish {job_id, kind, sequence_number} to ingestion:J:events
//   3. Return the full Event (callers may ignore it)
//
// On INSERT failure the error is returned; callers log and continue. A missing
// timeline must never abort an ingestion run.
func (s *Store) Append(ctx context.Context, req AppendRequest) (*Event, error) {
	detailJSON, err := json.Marshal(req.Detail)
	if err != nil {
		// A non-marshalable detail is a programming error, not an ingestion
		// error. Store an explicit marker so the row still exists and the
		// mistake is visible in the timeline instead of vanishing.
		s.logger.Warn("jobevents: detail marshal failed",
			zap.String("job_id", req.JobID.String()),
			zap.String("kind", req.Kind),
			zap.Error(err),
		)
		detailJSON = []byte(`{"marshal_error":true}`)
	}

	var event Event
	err = s.db.QueryRow(ctx,
		`INSERT INTO ingestion_job_events (job_id, expert_id, stage, kind, detail)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, sequence_number, created_at`,
		req.JobID, req.ExpertID, req.Stage, req.Kind, string(detailJSON),
	).Scan(&event.ID, &event.SequenceNumber, &event.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("jobevents append: %w", err)
	}

	event.JobID = req.JobID
	event.ExpertID = req.ExpertID
	event.Stage = req.Stage
	event.Kind = req.Kind
	event.Detail = detailJSON

	s.publishNotification(ctx, event)
	return &event, nil
}

// GetSince returns events for a job with sequence_number > since, oldest
// first. Used for the initial replay and for catch-up after a notification.
func (s *Store) GetSince(ctx context.Context, jobID uuid.UUID, since int64, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, job_id, expert_id, sequence_number, stage, kind, detail, created_at
		   FROM ingestion_job_events
		  WHERE job_id = $1 AND sequence_number > $2
		  ORDER BY sequence_number ASC
		  LIMIT $3`,
		jobID, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("jobevents get since: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var detail []byte
		if err := rows.Scan(
			&e.ID, &e.JobID, &e.ExpertID, &e.SequenceNumber,
			&e.Stage, &e.Kind, &detail, &e.CreatedAt,
		); err != nil {
			s.logger.Warn("jobevents get since: scan error", zap.Error(err))
			continue
		}
		e.Detail = detail
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("jobevents get since: rows error: %w", err)
	}
	return events, nil
}

// Subscriber returns a subscriber sharing this store's DB + Redis clients.
// WHY a method: the readers (admin SSE / history) should not have to be handed
// the Redis client separately — they only care about the timeline.
func (s *Store) Subscriber(logger *zap.Logger) *Subscriber {
	return NewSubscriber(s, s.redis, logger)
}

// Enabled reports whether a timeline store is wired. Nil-safe so callers can
// keep working before the migration is applied.
func (s *Store) Enabled() bool { return s != nil }

// ChannelName returns the Redis pub/sub channel for a job's events.
func ChannelName(jobID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:events", redisChannelPrefix, jobID.String())
}

// publishNotification publishes a lightweight wake-up to Redis.
// Non-fatal by design: Postgres already holds the event, and subscribers
// replay from the DB on reconnect or on the next notification.
func (s *Store) publishNotification(ctx context.Context, event Event) {
	if s.redis == nil {
		return
	}
	payload, err := json.Marshal(NotificationPayload{
		EventID:        event.ID,
		JobID:          event.JobID,
		Kind:           event.Kind,
		SequenceNumber: event.SequenceNumber,
		CreatedAt:      event.CreatedAt,
	})
	if err != nil {
		s.logger.Warn("jobevents: failed to marshal notification", zap.Error(err))
		return
	}
	channel := ChannelName(event.JobID)
	if err := s.redis.Publish(ctx, channel, string(payload)).Err(); err != nil {
		s.logger.Warn("jobevents: Redis publish failed (non-fatal)",
			zap.String("channel", channel),
			zap.Error(err),
		)
	}
}
