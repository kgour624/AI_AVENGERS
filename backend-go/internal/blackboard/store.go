package blackboard

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// redisChannelPrefix is the prefix for all workflow event channels.
// Full channel name: workflow:{workflow_id}:events
// WHY constant: changing this would break all subscribers silently.
const redisChannelPrefix = "workflow"

// Event represents a single blackboard event.
// Mirrors the blackboard_events table row exactly.
type Event struct {
	ID                uuid.UUID       `json:"id"`
	WorkflowID        uuid.UUID       `json:"workflow_id"`
	SequenceNumber    int64           `json:"sequence_number"`
	EventType         string          `json:"event_type"`
	PostedByExpertID  *uuid.UUID      `json:"posted_by_expert_id,omitempty"`
	PostedByClient    bool            `json:"posted_by_client"`
	ToExpertID        *uuid.UUID      `json:"to_expert_id,omitempty"`
	Content           json.RawMessage `json:"content"`
	ReferencesEventIDs []uuid.UUID    `json:"references_event_ids"`
	Revision          int             `json:"revision"`
	DedupKey          string          `json:"dedup_key"`
	PostedAt          time.Time       `json:"posted_at"`
}

// PostRequest is the input to Store.Post.
type PostRequest struct {
	WorkflowID        uuid.UUID
	EventType         string
	PostedByExpertID  *uuid.UUID // nil if posted by client
	PostedByClient    bool
	ToExpertID        *uuid.UUID // nil unless question_to_expert
	Content           interface{} // will be JSON-marshalled
	ReferencesEventIDs []uuid.UUID
	Revision          int // default 1
}

// NotificationPayload is what gets published to Redis.
// Subscribers use this to decide whether to fetch the full event.
type NotificationPayload struct {
	EventID        uuid.UUID `json:"event_id"`
	WorkflowID     uuid.UUID `json:"workflow_id"`
	EventType      string    `json:"event_type"`
	SequenceNumber int64     `json:"sequence_number"`
	PostedAt       time.Time `json:"posted_at"`
}

// Store is the single write path for all blackboard events.
//
// WHY single write path:
//   All blackboard writes go through here. This ensures:
//   1. Dedup key is always computed consistently.
//   2. Redis notification always follows Postgres write.
//   3. No expert can bypass the blackboard by writing directly.
type Store struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	logger *zap.Logger
}

// NewStore creates a new blackboard store.
func NewStore(db *pgxpool.Pool, redisClient *redis.Client, logger *zap.Logger) *Store {
	return &Store{
		db:     db,
		redis:  redisClient,
		logger: logger,
	}
}

// Post appends an event to the blackboard.
//
// Idempotent: if an event with the same dedup_key already exists for
// this workflow, the existing event is returned without error.
// This makes retries safe.
//
// Mental execution:
//   Input: PostRequest{WorkflowID: abc, EventType: "architecture_decision", ...}
//   1. Marshal content to JSON
//   2. Compute dedup_key
//   3. INSERT ... ON CONFLICT DO NOTHING RETURNING id, sequence_number, posted_at
//   4. If no rows (duplicate): SELECT existing event
//   5. Publish notification to Redis (best-effort, non-fatal if Redis down)
//   6. Return event
func (s *Store) Post(ctx context.Context, req PostRequest) (*Event, error) {
	// Step 1: Marshal content
	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		return nil, fmt.Errorf("blackboard post: marshal content: %w", err)
	}

	// Step 2: Compute dedup_key
	// Format: SHA-256(event_type + "|" + poster_id + "|" + content_json)
	// WHY include content: same poster posting different content for the
	// same event_type is a legitimate new event (e.g. revised architecture).
	// The revision field handles that, but dedup_key must be unique per
	// logical post attempt.
	posterID := ""
	if req.PostedByExpertID != nil {
		posterID = req.PostedByExpertID.String()
	} else if req.PostedByClient {
		posterID = "client"
	}
	dedupRaw := fmt.Sprintf("%s|%s|%s", req.EventType, posterID, string(contentJSON))
	dedupHash := sha256.Sum256([]byte(dedupRaw))
	dedupKey := fmt.Sprintf("%x", dedupHash)

	// Step 3: Revision default
	revision := req.Revision
	if revision < 1 {
		revision = 1
	}

	// Step 4: Build references array (Postgres UUID[])
	//
	// This used to be json.Marshal, and that was a live bug: it produces
	// ["a1b2-…","c3d4-…"], and the parameter is cast with $7::uuid[]. Postgres
	// reads a leading '[' in an array literal as the start of an explicit
	// dimension specifier (the [1:2]={…} form), so every post with a non-empty
	// references list failed with
	//
	//	malformed array literal … "[" must introduce explicitly-specified array dimensions
	//
	// The empty case marshalled to "[]" only when the slice was non-nil-but-empty
	// and to the correct "{}" when nil, which is why this went unnoticed: the
	// common path posts no references at all. What it broke, silently:
	// cross_verifier.go:274 and :540 (every verification finding), and
	// Tools.AskClient's cited events — i.e. the handoff gate, which passes every
	// artifact id.
	refsLiteral := UUIDArrayLiteral(req.ReferencesEventIDs)

	// Step 5: INSERT with ON CONFLICT DO NOTHING
	// RETURNING gives us the new row's id + sequence_number + posted_at.
	// If conflict (duplicate), RETURNING returns no rows.
	var event Event
	err = s.db.QueryRow(ctx,
		`INSERT INTO blackboard_events
			(workflow_id, event_type,
			 posted_by_expert_id, posted_by_client, to_expert_id,
			 content, references_event_ids, revision, dedup_key)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::uuid[], $8, $9)
		 ON CONFLICT (workflow_id, dedup_key) DO NOTHING
		 RETURNING id, sequence_number, posted_at`,
		req.WorkflowID,
		req.EventType,
		req.PostedByExpertID,
		req.PostedByClient,
		req.ToExpertID,
		string(contentJSON),
		refsLiteral,
		revision,
		dedupKey,
	).Scan(&event.ID, &event.SequenceNumber, &event.PostedAt)

	if err != nil {
		// ON CONFLICT DO NOTHING + RETURNING = no rows on duplicate.
		// pgx surfaces this as pgx.ErrNoRows.
		if err.Error() == "no rows in result set" {
			// Fetch the existing event
			existing, fetchErr := s.GetByDedupKey(ctx, req.WorkflowID, dedupKey)
			if fetchErr != nil {
				return nil, fmt.Errorf("blackboard post: fetch duplicate: %w", fetchErr)
			}
			s.logger.Debug("blackboard post: duplicate suppressed",
				zap.String("workflow_id", req.WorkflowID.String()),
				zap.String("event_type", req.EventType),
				zap.String("dedup_key", dedupKey[:8]),
			)
			return existing, nil
		}
		return nil, fmt.Errorf("blackboard post: insert: %w", err)
	}

	// Populate remaining fields on the returned event
	event.WorkflowID = req.WorkflowID
	event.EventType = req.EventType
	event.PostedByExpertID = req.PostedByExpertID
	event.PostedByClient = req.PostedByClient
	event.ToExpertID = req.ToExpertID
	event.Content = contentJSON
	event.ReferencesEventIDs = req.ReferencesEventIDs
	event.Revision = revision
	event.DedupKey = dedupKey

	s.logger.Debug("blackboard event posted",
		zap.String("workflow_id", req.WorkflowID.String()),
		zap.String("event_type", req.EventType),
		zap.Int64("sequence_number", event.SequenceNumber),
	)

	// Step 6: Publish to Redis (best-effort).
	// WHY best-effort: Redis is notification-only. If Redis is down,
	// the event is still in Postgres. Subscribers will replay on reconnect.
	// A Redis failure must NOT fail the blackboard write.
	s.publishNotification(ctx, event)

	return &event, nil
}

// GetByDedupKey fetches an existing event by its dedup key.
// Used internally when a duplicate post is detected.
func (s *Store) GetByDedupKey(ctx context.Context, workflowID uuid.UUID, dedupKey string) (*Event, error) {
	var e Event
	var contentJSON []byte
	err := s.db.QueryRow(ctx,
		`SELECT id, workflow_id, sequence_number, event_type,
		        posted_by_expert_id, posted_by_client, to_expert_id,
		        content, revision, dedup_key, posted_at
		 FROM blackboard_events
		 WHERE workflow_id = $1 AND dedup_key = $2
		 LIMIT 1`,
		workflowID, dedupKey,
	).Scan(
		&e.ID, &e.WorkflowID, &e.SequenceNumber, &e.EventType,
		&e.PostedByExpertID, &e.PostedByClient, &e.ToExpertID,
		&contentJSON, &e.Revision, &e.DedupKey, &e.PostedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("blackboard get by dedup key: %w", err)
	}
	e.Content = contentJSON
	return &e, nil
}

// GetSince returns all events for a workflow with sequence_number > since.
// Used by subscribers to catch up after a cursor.
//
// Mental execution:
//   Input: workflowID=abc, since=42
//   Output: events with sequence_number 43, 44, 45, ...
//   Used by: subscriber on reconnect, workflow engine on resume
func (s *Store) GetSince(ctx context.Context, workflowID uuid.UUID, since int64, limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, workflow_id, sequence_number, event_type,
		        posted_by_expert_id, posted_by_client, to_expert_id,
		        content, revision, dedup_key, posted_at
		 FROM blackboard_events
		 WHERE workflow_id = $1 AND sequence_number > $2
		 ORDER BY sequence_number ASC
		 LIMIT $3`,
		workflowID, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("blackboard get since: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var contentJSON []byte
		if err := rows.Scan(
			&e.ID, &e.WorkflowID, &e.SequenceNumber, &e.EventType,
			&e.PostedByExpertID, &e.PostedByClient, &e.ToExpertID,
			&contentJSON, &e.Revision, &e.DedupKey, &e.PostedAt,
		); err != nil {
			s.logger.Warn("blackboard get since: scan error", zap.Error(err))
			continue
		}
		e.Content = contentJSON
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("blackboard get since: rows error: %w", err)
	}
	return events, nil
}

// GetByType returns events of a specific type for a workflow.
// Used by experts reading relevant artifacts from the blackboard.
func (s *Store) GetByType(ctx context.Context, workflowID uuid.UUID, eventTypes []string, since int64) ([]Event, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, workflow_id, sequence_number, event_type,
		        posted_by_expert_id, posted_by_client, to_expert_id,
		        content, revision, dedup_key, posted_at
		 FROM blackboard_events
		 WHERE workflow_id = $1
		   AND event_type = ANY($2)
		   AND sequence_number > $3
		 ORDER BY sequence_number ASC`,
		workflowID, eventTypes, since,
	)
	if err != nil {
		return nil, fmt.Errorf("blackboard get by type: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var contentJSON []byte
		if err := rows.Scan(
			&e.ID, &e.WorkflowID, &e.SequenceNumber, &e.EventType,
			&e.PostedByExpertID, &e.PostedByClient, &e.ToExpertID,
			&contentJSON, &e.Revision, &e.DedupKey, &e.PostedAt,
		); err != nil {
			continue
		}
		e.Content = contentJSON
		events = append(events, e)
	}
	return events, nil
}

// ChannelName returns the Redis pub/sub channel name for a workflow.
func ChannelName(workflowID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:events", redisChannelPrefix, workflowID.String())
}

// publishNotification publishes a lightweight notification to Redis.
// Non-fatal: logs warning on error, does not return error to caller.
func (s *Store) publishNotification(ctx context.Context, event Event) {
	payload := NotificationPayload{
		EventID:        event.ID,
		WorkflowID:     event.WorkflowID,
		EventType:      event.EventType,
		SequenceNumber: event.SequenceNumber,
		PostedAt:       event.PostedAt,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		s.logger.Warn("blackboard: failed to marshal notification", zap.Error(err))
		return
	}
	channel := ChannelName(event.WorkflowID)
	if err := s.redis.Publish(ctx, channel, string(payloadJSON)).Err(); err != nil {
		// Non-fatal: Redis is notification-only. Event is already in Postgres.
		s.logger.Warn("blackboard: Redis publish failed (non-fatal)",
			zap.String("channel", channel),
			zap.Error(err),
		)
	}
}

// UUIDArrayLiteral renders a slice of UUIDs as a Postgres array literal:
// {a1b2…,c3d4…}. An empty or nil slice gives {}.
//
// Exported because two tables in two packages carry a UUID[] column written
// through a ::uuid[] text cast — blackboard_events.references_event_ids here,
// and approval_requests.cited_event_ids in internal/workflow — and both had the
// same json.Marshal bug. One implementation means one place to be right.
//
// No quoting or escaping: a uuid.UUID stringifies to hex and hyphens only, so
// there is no input that could need it. That is a property of the type, not an
// assumption about the caller — which is why this takes []uuid.UUID and not
// []string.
func UUIDArrayLiteral(ids []uuid.UUID) string {
	if len(ids) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, id := range ids {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(id.String())
	}
	b.WriteByte('}')
	return b.String()
}
