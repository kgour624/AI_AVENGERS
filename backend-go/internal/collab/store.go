package collab

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RetryWindow is the human-in-the-loop retry deadline (ans4): a failed relay can
// be retried from its failed section for this long; after it elapses the run is
// swept to terminal "failed".
const RetryWindow = 24 * time.Hour

// MaxEventMessageLen caps every transparency event's human-readable message
// (ans5: character limit 3500).
const MaxEventMessageLen = 3500

type RunStatus string

const (
	RunRunning        RunStatus = "running"
	RunAwaitingReview RunStatus = "awaiting_human_review"
	RunRelayFailed    RunStatus = "relay_failed"
	RunRetrying       RunStatus = "retrying"
	RunCompleted      RunStatus = "completed"
	RunFailed         RunStatus = "failed"
)

type SectionStatus string

const (
	SectionPending        SectionStatus = "pending"
	SectionGenerating     SectionStatus = "generating"
	SectionAwaitingReview SectionStatus = "awaiting_review"
	SectionApproved       SectionStatus = "approved"
	SectionFailed         SectionStatus = "failed"
)

type RelayRun struct {
	ID            uuid.UUID
	ChatID        uuid.UUID
	UserMessageID uuid.UUID
	Status        RunStatus
	Plan          []Section
	CurrentIndex  int
	FailedAtIndex *int
	FailureReason string
	RetryDeadline *time.Time
}

type RelaySection struct {
	ID                  uuid.UUID
	RelayRunID          uuid.UUID
	SectionIndex        int
	ExpertID            uuid.UUID
	ExpertName          string
	SectionTitle        string
	RawContent          string
	FinalContent        string
	Edited              bool
	Status              SectionStatus
	ConflictDetected    bool
	ConflictExplanation string
	ResolutionSource    string
	ErrorReason         string
}

type RelayEvent struct {
	SequenceNum int
	Step        string
	ExpertID    *uuid.UUID
	Message     string
	Metadata    map[string]any
	CreatedAt   time.Time
}

type RelayStore struct{ db *pgxpool.Pool }

func NewRelayStore(db *pgxpool.Pool) *RelayStore { return &RelayStore{db: db} }

func (s *RelayStore) CreateRun(ctx context.Context, run RelayRun) (uuid.UUID, error) {
	planJSON, err := json.Marshal(run.Plan)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshal plan: %w", err)
	}
	id := uuid.New()
	_, err = s.db.Exec(ctx,
		`INSERT INTO collab_relay_runs (id, chat_id, user_message_id, status, plan)
		 VALUES ($1,$2,$3,$4,$5)`,
		id, run.ChatID, run.UserMessageID, string(RunRunning), planJSON)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create run: %w", err)
	}
	for i, sec := range run.Plan {
		if err := s.createSection(ctx, id, i, sec); err != nil {
			return uuid.Nil, err
		}
	}
	return id, nil
}

func (s *RelayStore) createSection(ctx context.Context, runID uuid.UUID, index int, sec Section) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO collab_relay_sections
		 (relay_run_id, section_index, expert_id, expert_name, section_title)
		 VALUES ($1,$2,$3,$4,$5)`,
		runID, index, sec.ExpertID, sec.ExpertName, sec.SectionTitle)
	if err != nil {
		return fmt.Errorf("create section %d: %w", index, err)
	}
	return nil
}

// AppendEvent assigns sequence_num atomically and returns the new event.
func (s *RelayStore) AppendEvent(ctx context.Context, runID uuid.UUID, step string, expertID *uuid.UUID, message string, metadata map[string]any) (RelayEvent, error) {
	if len(message) > MaxEventMessageLen {
		message = message[:MaxEventMessageLen] + "...[truncated]"
	}
	metaJSON, _ := json.Marshal(metadata)
	var seq int
	err := s.db.QueryRow(ctx,
		`INSERT INTO collab_relay_events (relay_run_id, sequence_num, step, expert_id, message, metadata)
		 SELECT $1, COALESCE(MAX(sequence_num),0)+1, $2, $3, $4, $5
		   FROM collab_relay_events WHERE relay_run_id = $1
		 RETURNING sequence_num`,
		runID, step, expertID, message, metaJSON).Scan(&seq)
	if err != nil {
		return RelayEvent{}, fmt.Errorf("append event: %w", err)
	}
	return RelayEvent{SequenceNum: seq, Step: step, ExpertID: expertID, Message: message, Metadata: metadata}, nil
}

func (s *RelayStore) UpdateRunStatus(ctx context.Context, runID uuid.UUID, status RunStatus) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_runs SET status=$2, updated_at=now() WHERE id=$1`,
		runID, string(status))
	return err
}

func (s *RelayStore) MarkRunFailed(ctx context.Context, runID uuid.UUID, atIndex int, reason string, deadline time.Time) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_runs
		    SET status='relay_failed', failed_at_index=$2, failure_reason=$3,
		        retry_deadline=$4, updated_at=now()
		  WHERE id=$1`,
		runID, atIndex, reason, deadline)
	return err
}

func (s *RelayStore) BeginRetry(ctx context.Context, runID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_runs SET status='retrying', updated_at=now() WHERE id=$1`, runID)
	return err
}

func (s *RelayStore) MarkSectionStatus(ctx context.Context, runID uuid.UUID, index int, status SectionStatus) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_sections SET status=$3, updated_at=now()
		  WHERE relay_run_id=$1 AND section_index=$2`,
		runID, index, string(status))
	return err
}

func (s *RelayStore) SetSectionRawContent(ctx context.Context, runID uuid.UUID, index int, raw string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_sections
		    SET raw_content=$3, status='awaiting_review', updated_at=now()
		  WHERE relay_run_id=$1 AND section_index=$2`,
		runID, index, raw)
	return err
}

func (s *RelayStore) SetSectionFailure(ctx context.Context, runID uuid.UUID, index int, reason string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_sections
		    SET status='failed', error_reason=$3, updated_at=now()
		  WHERE relay_run_id=$1 AND section_index=$2`,
		runID, index, reason)
	return err
}

func (s *RelayStore) SetSectionConflict(ctx context.Context, runID uuid.UUID, index int, conflict bool, explanation string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE collab_relay_sections
		    SET conflict_detected=$3, conflict_explanation=$4, updated_at=now()
		  WHERE relay_run_id=$1 AND section_index=$2`,
		runID, index, conflict, explanation)
	return err
}

// ApproveSection is the human-in-the-loop write: stores the (possibly edited)
// final content and resolution, flips the section to approved, and returns the
// updated row so the blocked relay can continue with the human's content.
func (s *RelayStore) ApproveSection(ctx context.Context, runID, sectionID uuid.UUID, finalContent string, edited bool, resolutionSource string) (RelaySection, error) {
	var rec RelaySection
	err := s.db.QueryRow(ctx,
		`UPDATE collab_relay_sections
		    SET final_content=$3, edited=$4, resolution_source=$5,
		        status='approved', approved_at=now(), updated_at=now()
		  WHERE id=$1 AND relay_run_id=$2
		  RETURNING relay_run_id, section_index, expert_id, expert_name,
		            section_title, COALESCE(final_content,''), edited, status,
		            conflict_detected, COALESCE(conflict_explanation,''),
		            COALESCE(resolution_source,''), COALESCE(error_reason,'')`,
		sectionID, runID, finalContent, edited, nullIfEmpty(resolutionSource),
	).Scan(&rec.RelayRunID, &rec.SectionIndex, &rec.ExpertID, &rec.ExpertName,
		&rec.SectionTitle, &rec.FinalContent, &rec.Edited, &rec.Status,
		&rec.ConflictDetected, &rec.ConflictExplanation, &rec.ResolutionSource,
		&rec.ErrorReason)
	if err != nil {
		return RelaySection{}, fmt.Errorf("approve section: %w", err)
	}
	rec.ID = sectionID
	return rec, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// WaitForSectionApproval blocks until the section is approved or failed, or ctx
// is cancelled (client disconnect / retry deadline). This is the gate primitive.
func (s *RelayStore) WaitForSectionApproval(ctx context.Context, sectionID uuid.UUID) (RelaySection, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return RelaySection{}, ctx.Err()
		case <-ticker.C:
			rec, err := s.GetSection(ctx, sectionID)
			if err != nil {
				return RelaySection{}, err
			}
			switch rec.Status {
			case SectionApproved:
				return rec, nil
			case SectionFailed:
				return rec, fmt.Errorf("section failed: %s", rec.ErrorReason)
			}
		}
	}
}

func (s *RelayStore) GetRun(ctx context.Context, runID uuid.UUID) (RelayRun, error) {
	var r RelayRun
	var planJSON []byte
	var failedIdx *int
	var reason *string
	var deadline *time.Time
	err := s.db.QueryRow(ctx,
		`SELECT id, chat_id, user_message_id, status, plan, current_index,
		        failed_at_index, failure_reason, retry_deadline
		   FROM collab_relay_runs WHERE id=$1`,
		runID,
	).Scan(&r.ID, &r.ChatID, &r.UserMessageID, &r.Status, &planJSON,
		&r.CurrentIndex, &failedIdx, &reason, &deadline)
	if err != nil {
		return r, err
	}
	_ = json.Unmarshal(planJSON, &r.Plan)
	if failedIdx != nil {
		r.FailedAtIndex = failedIdx
	}
	if reason != nil {
		r.FailureReason = *reason
	}
	r.RetryDeadline = deadline
	return r, nil
}

// GetActiveRun returns the latest run for a chat that is not terminal
// (completed/failed). Used by reconnect to hydrate the gate/failure/transcript.
func (s *RelayStore) GetActiveRun(ctx context.Context, chatID uuid.UUID) (*RelayRun, error) {
	var r RelayRun
	var planJSON []byte
	err := s.db.QueryRow(ctx,
		`SELECT id, chat_id, user_message_id, status, plan, current_index,
		        failed_at_index, failure_reason, retry_deadline
		   FROM collab_relay_runs
		  WHERE chat_id=$1 AND status NOT IN ('completed','failed')
		  ORDER BY created_at DESC LIMIT 1`,
		chatID,
	).Scan(&r.ID, &r.ChatID, &r.UserMessageID, &r.Status, &planJSON,
		&r.CurrentIndex, &r.FailedAtIndex, &r.FailureReason, &r.RetryDeadline)
	if err != nil {
		return nil, err // pgx.ErrNoRows propagates; caller treats as "no active run"
	}
	_ = json.Unmarshal(planJSON, &r.Plan)
	return &r, nil
}

func (s *RelayStore) GetSection(ctx context.Context, sectionID uuid.UUID) (RelaySection, error) {
	var rec RelaySection
	err := s.db.QueryRow(ctx,
		`SELECT id, relay_run_id, section_index, expert_id, expert_name, section_title,
		        COALESCE(raw_content,''), COALESCE(final_content,''), edited, status,
		        conflict_detected, COALESCE(conflict_explanation,''),
		        COALESCE(resolution_source,''), COALESCE(error_reason,'')
		   FROM collab_relay_sections WHERE id=$1`,
		sectionID,
	).Scan(&rec.ID, &rec.RelayRunID, &rec.SectionIndex, &rec.ExpertID, &rec.ExpertName,
		&rec.SectionTitle, &rec.RawContent, &rec.FinalContent, &rec.Edited, &rec.Status,
		&rec.ConflictDetected, &rec.ConflictExplanation, &rec.ResolutionSource, &rec.ErrorReason)
	return rec, err
}

func (s *RelayStore) GetSectionByIndex(ctx context.Context, runID uuid.UUID, index int) (RelaySection, error) {
	var rec RelaySection
	err := s.db.QueryRow(ctx,
		`SELECT id, relay_run_id, section_index, expert_id, expert_name, section_title,
		        COALESCE(raw_content,''), COALESCE(final_content,''), edited, status,
		        conflict_detected, COALESCE(conflict_explanation,''),
		        COALESCE(resolution_source,''), COALESCE(error_reason,'')
		   FROM collab_relay_sections
		  WHERE relay_run_id=$1 AND section_index=$2`,
		runID, index,
	).Scan(&rec.ID, &rec.RelayRunID, &rec.SectionIndex, &rec.ExpertID, &rec.ExpertName,
		&rec.SectionTitle, &rec.RawContent, &rec.FinalContent, &rec.Edited, &rec.Status,
		&rec.ConflictDetected, &rec.ConflictExplanation, &rec.ResolutionSource, &rec.ErrorReason)
	return rec, err
}

// EventsAfter returns all transparency events after a sequence number, ordered,
// for reconnect replay (sequence 0 => full history, same as workflow's
// subscribe-from-sequence-0 idea but implemented here for the chat system only).
func (s *RelayStore) EventsAfter(ctx context.Context, runID uuid.UUID, afterSeq int) ([]RelayEvent, error) {
	rows, err := s.db.Query(ctx,
		`SELECT sequence_num, step, expert_id, message, metadata, created_at
		   FROM collab_relay_events
		  WHERE relay_run_id=$1 AND sequence_num > $2
		  ORDER BY sequence_num ASC`,
		runID, afterSeq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RelayEvent
	for rows.Next() {
		var e RelayEvent
		var metaJSON []byte
		if err := rows.Scan(&e.SequenceNum, &e.Step, &e.ExpertID, &e.Message, &metaJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &e.Metadata)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SweepExpiredFailures flips relay_failed runs past their retry_deadline to
// terminal "failed". Called periodically by a background job.
func (s *RelayStore) SweepExpiredFailures(ctx context.Context) (int64, error) {
	tag, err := s.db.Exec(ctx,
		`UPDATE collab_relay_runs
		    SET status='failed', updated_at=now()
		  WHERE status='relay_failed' AND retry_deadline IS NOT NULL
		    AND retry_deadline < now()`)
	return tag.RowsAffected(), err
}
