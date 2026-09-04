package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// L3Event is a single immutable event in the master event log.
// NEVER updated, NEVER deleted. Append-only.
//
// WHY append-only:
// Complete audit trail. "What did Expert A decide at turn 47 and WHY?"
// Rating system needs history to learn from.
// Debugging: trace exactly what happened and when.
type L3Event struct {
	ProjectID              uuid.UUID              `json:"project_id"`
	ExpertID               *uuid.UUID             `json:"expert_id,omitempty"`
	ClientID               uuid.UUID              `json:"client_id"`
	ChatID                 *uuid.UUID             `json:"chat_id,omitempty"`
	MessageID              *uuid.UUID             `json:"message_id,omitempty"`
	EventType              string                 `json:"event_type"`
	EventData              map[string]interface{} `json:"event_data"`
	Reasoning              string                 `json:"reasoning"`
	DecisionMade           string                 `json:"decision_made"`
	AlternativesConsidered []string               `json:"alternatives_considered"`
}

// L3EventRecord is a stored event with DB-assigned ID and timestamp.
type L3EventRecord struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	L3Event
}

// Common event types
const (
	EventMessageSent       = "message_sent"
	EventResponseGenerated = "response_generated"
	EventDecisionMade      = "decision_made"
	EventRatingRecorded    = "rating_recorded"
	EventExpertSwitched    = "expert_switched"
	EventChinaWallViolation = "china_wall_violation"
	EventGateStopped       = "gate_stopped"
	EventIngestionComplete = "ingestion_complete"
)

// L3Store handles the append-only master event log.
type L3Store struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewL3Store creates a new L3 store.
func NewL3Store(db *pgxpool.Pool, logger *zap.Logger) *L3Store {
	return &L3Store{db: db, logger: logger}
}

// Append adds a new event to the master log.
// This is the ONLY write operation on L3 — no updates, no deletes.
func (s *L3Store) Append(ctx context.Context, event L3Event) error {
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	// Serialize event data
	eventDataJSON := "{}"
	if event.EventData != nil {
		data, err := marshalJSON(event.EventData)
		if err == nil {
			eventDataJSON = string(data)
		}
	}

	alternativesJSON := "[]"
	if len(event.AlternativesConsidered) > 0 {
		data, err := marshalJSON(event.AlternativesConsidered)
		if err == nil {
			alternativesJSON = string(data)
		}
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO master_event_log
			(project_id, expert_id, client_id, chat_id, message_id,
			 event_type, event_data, reasoning, decision_made, alternatives_considered)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		event.ProjectID, event.ExpertID, event.ClientID,
		event.ChatID, event.MessageID,
		event.EventType, eventDataJSON,
		event.Reasoning, event.DecisionMade, alternativesJSON,
	)
	if err != nil {
		s.logger.Error("L3 append failed",
			zap.String("event_type", event.EventType),
			zap.Error(err),
		)
		return fmt.Errorf("L3 append failed: %w", err)
	}
	return nil
}

// GetProjectTimeline returns paginated events for a project.
// Used by the "Project Timeline" UI feature.
func (s *L3Store) GetProjectTimeline(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]L3EventRecord, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, client_id, chat_id, message_id,
		        event_type, COALESCE(reasoning,''), COALESCE(decision_made,''), created_at
		 FROM master_event_log
		 WHERE project_id=$1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		projectID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("timeline query failed: %w", err)
	}
	defer rows.Close()

	var events []L3EventRecord
	for rows.Next() {
		var e L3EventRecord
		if err := rows.Scan(
			&e.ID, &e.ProjectID, &e.ExpertID, &e.ClientID,
			&e.ChatID, &e.MessageID,
			&e.EventType, &e.Reasoning, &e.DecisionMade, &e.CreatedAt,
		); err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetViolations returns China Wall violation events.
// Used by admin panel to monitor hallucination attempts.
func (s *L3Store) GetViolations(ctx context.Context, limit int) ([]L3EventRecord, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, expert_id, client_id, chat_id, message_id,
		        event_type, COALESCE(reasoning,''), COALESCE(decision_made,''), created_at
		 FROM master_event_log
		 WHERE event_type=$1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		EventChinaWallViolation, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []L3EventRecord
	for rows.Next() {
		var e L3EventRecord
		if err := rows.Scan(
			&e.ID, &e.ProjectID, &e.ExpertID, &e.ClientID,
			&e.ChatID, &e.MessageID,
			&e.EventType, &e.Reasoning, &e.DecisionMade, &e.CreatedAt,
		); err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
