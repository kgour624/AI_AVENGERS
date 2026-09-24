// Package explain owns C9: the "why this answer" explainability surface.
//
// WHY a dedicated package:
//   An explanation is a VIEW over stored facts — never a new LLM narration
//   (§3.1, C9 design). The facts already live in two places: the `messages`
//   row (decision mode, gate stopped, coverage, judge score, refusal reason,
//   warning/questions, citations) and the C1 provenance record (signed chain,
//   B8 claim→evidence reports, model, signature). This package joins them into
//   one deterministic DTO so trust/debugging does not depend on a fragile or
//   hallucination-prone summary call.
//
// AUTHORIZATION: a caller may only explain their own chat's message. Ownership
// is a client_id match AND the C4 tenant assertion on the chat's project
// (fail closed — a foreign/unknown message is a flat NotFound so existence
// never leaks).
package explain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/provenance"
	"ai_avengers/backend/internal/tenant"
)

// ErrNotFound is returned when the message does not exist or is not owned by
// the caller (the two cases are deliberately indistinguishable).
var ErrNotFound = errors.New("message not found")

// Explanation is the assembled, read-only explanation for one answer.
type Explanation struct {
	Message   MessageFacts    `json:"message"`
	Decision  DecisionFacts   `json:"decision"`
	Gates     []GateStep      `json:"gates"`
	Refusal   *RefusalFacts   `json:"refusal,omitempty"`
	Sources   []SourceRef     `json:"sources"`
	Claims    json.RawMessage `json:"claims,omitempty"`
	Quality   QualityFacts    `json:"quality"`
	Integrity *IntegrityFacts `json:"integrity,omitempty"`
	Expert    *ExpertRef      `json:"expert,omitempty"`
}

// MessageFacts are the raw stored fields of the answered message.
type MessageFacts struct {
	ID          uuid.UUID  `json:"id"`
	ChatID      uuid.UUID  `json:"chat_id"`
	Role        string     `json:"role"`
	ExpertID    *uuid.UUID `json:"expert_id,omitempty"`
	Mode        string     `json:"mode"`
	GateStopped int        `json:"gate_stopped"`
	Confidence  *float64   `json:"confidence,omitempty"`
	Model       string     `json:"model,omitempty"`
	CreatedAt   string     `json:"created_at"`
}

// DecisionFacts is the human-readable decision summary (deterministic).
type DecisionFacts struct {
	Mode        string `json:"mode"`
	Label       string `json:"label"`
	Explanation string `json:"explanation"`
}

// GateStep is one row of the deterministic gate timeline.
type GateStep struct {
	N      int    `json:"n"`
	Name   string `json:"name"`
	Status string `json:"status"` // passed | stopped | not_reached | asked
}

// RefusalFacts explains a refusal / warning / clarification.
type RefusalFacts struct {
	Reason              string   `json:"reason,omitempty"`
	Warning             string   `json:"warning,omitempty"`
	ClarifyingQuestions []string `json:"clarifying_questions,omitempty"`
}

// SourceRef is one cited chunk (mirrors chinawall.Citation's JSON shape).
type SourceRef struct {
	ChunkID    uuid.UUID `json:"chunk_id"`
	Text       string    `json:"text"`
	Score      float32   `json:"score"`
	SourceName string    `json:"source_name,omitempty"`
	ChunkIndex int       `json:"chunk_index,omitempty"`
}

// QualityFacts is the B6 judge score + China Wall coverage verdict.
type QualityFacts struct {
	Score    *float64 `json:"score,omitempty"`
	Coverage string   `json:"coverage,omitempty"`
	JudgeRan bool     `json:"judge_ran"`
}

// IntegrityFacts is the C1 signed-chain verification result.
type IntegrityFacts struct {
	ContentHash    string `json:"content_hash"`
	SignatureValid bool   `json:"signature_valid"`
	SigningKey     string `json:"signing_key_id,omitempty"`
	RecordedAt     string `json:"recorded_at,omitempty"`
}

// ExpertRef identifies the expert that produced the answer.
type ExpertRef struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Domain string    `json:"domain"`
}

// Service assembles explanations from stored facts.
type Service struct {
	db      *pgxpool.Pool
	prov    *provenance.Service
	tenants *tenant.Service
	logger  *zap.Logger
}

// NewService builds the explainability service. prov/tenants may be nil
// (provenance section omitted; tenant assertion becomes a no-op global scope).
func NewService(db *pgxpool.Pool, prov *provenance.Service, tenants *tenant.Service, logger *zap.Logger) *Service {
	return &Service{db: db, prov: prov, tenants: tenants, logger: logger}
}

// row is the raw message+chat join used internally.
type row struct {
	id          uuid.UUID
	chatID      uuid.UUID
	role        string
	expertID    *uuid.UUID
	mode        string
	gateStopped *int
	confidence  *float64
	model       string
	warning     *string
	questions   []byte
	citations   []byte
	quality     *float64
	coverage    *string
	refusal     *string
	createdAt   string
	clientID    uuid.UUID
	projectID   uuid.UUID
}

// Get assembles the explanation for a message the caller owns.
func (s *Service) Get(ctx context.Context, messageID, userID uuid.UUID, scope tenant.Scope) (*Explanation, error) {
	if s == nil || s.db == nil {
		return nil, ErrNotFound
	}
	var r row
	err := s.db.QueryRow(ctx, `
		SELECT m.id, m.chat_id, m.role, m.expert_id, COALESCE(m.decision_mode,''),
		       m.gate_stopped, m.confidence, COALESCE(m.model_used,''),
		       m.warning_text, COALESCE(m.clarifying_questions,'[]'::jsonb),
		       COALESCE(m.citations,'[]'::jsonb), m.quality_score, m.coverage,
		       m.refusal_reason, m.created_at::text,
		       c.client_id, COALESCE(c.project_id, '00000000-0000-0000-0000-000000000000'::uuid)
		FROM messages m
		JOIN chats c ON c.id = m.chat_id
		WHERE m.id = $1`, messageID,
	).Scan(&r.id, &r.chatID, &r.role, &r.expertID, &r.mode,
		&r.gateStopped, &r.confidence, &r.model,
		&r.warning, &r.questions,
		&r.citations, &r.quality, &r.coverage,
		&r.refusal, &r.createdAt,
		&r.clientID, &r.projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load message: %w", err)
	}

	// Ownership: client_id match is required, then the C4 tenant assertion on
	// the chat's project. Both failures collapse to ErrNotFound (no probing).
	if r.clientID != userID {
		return nil, ErrNotFound
	}
	projectID := r.projectID
	if projectID != uuid.Nil {
		if err := s.tenants.AssertProject(ctx, scope, projectID); err != nil {
			return nil, ErrNotFound
		}
	}

	gateStopped := 0
	if r.gateStopped != nil {
		gateStopped = *r.gateStopped
	}

	exp := &Explanation{
		Message: MessageFacts{
			ID: r.id, ChatID: r.chatID, Role: r.role, ExpertID: r.expertID,
			Mode: r.mode, GateStopped: gateStopped, Confidence: r.confidence,
			Model: r.model, CreatedAt: r.createdAt,
		},
		Decision: DecisionFacts{
			Mode:        r.mode,
			Label:       ModeLabel(r.mode),
			Explanation: ModeExplanation(r.mode, gateStopped),
		},
		Gates:   GateTimeline(gateStopped),
		Sources: parseSources(r.citations),
		Quality: QualityFacts{Score: r.quality, Coverage: derefStr(r.coverage), JudgeRan: r.quality != nil},
	}

	// Refusal/warning/clarification facts (only when present).
	var questions []string
	if len(r.questions) > 0 {
		_ = json.Unmarshal(r.questions, &questions)
	}
	if derefStr(r.refusal) != "" || derefStr(r.warning) != "" || len(questions) > 0 {
		exp.Refusal = &RefusalFacts{
			Reason:              derefStr(r.refusal),
			Warning:             derefStr(r.warning),
			ClarifyingQuestions: questions,
		}
	}

	// Expert reference (best-effort).
	if r.expertID != nil {
		var e ExpertRef
		if qErr := s.db.QueryRow(ctx,
			`SELECT id, name, domain FROM experts WHERE id=$1`, *r.expertID,
		).Scan(&e.ID, &e.Name, &e.Domain); qErr == nil {
			exp.Expert = &e
		}
	}

	// Provenance: signed chain + B8 claims + integrity (best-effort; C1 may
	// be off or not yet recorded for very old messages).
	if s.prov != nil {
		rec, pErr := s.prov.Get(ctx, provenance.OutputChatMessage, messageID)
		if pErr != nil {
			s.logger.Warn("explain: provenance fetch failed", zap.Error(pErr))
		} else if rec != nil {
			exp.Claims = rec.Claims
			exp.Integrity = &IntegrityFacts{
				ContentHash:    rec.ContentHash,
				SignatureValid: s.prov.Verify(rec),
				SigningKey:     rec.SigningKey,
				RecordedAt:     rec.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			}
		}
	}

	return exp, nil
}

// parseSources unmarshals the stored citations JSONB. Pure (unit-tested):
// garbage/empty → empty slice, never nil-in-a-way-that-crashes JSON clients.
func parseSources(raw json.RawMessage) []SourceRef {
	out := []SourceRef{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// gateNames is the canonical 5-gate naming (mirrors decision/engine.go).
var gateNames = map[int]string{
	1: "Information Sufficiency",
	2: "Knowledge Coverage",
	3: "Charter Compliance",
	4: "Necessity Check",
	5: "Generate Answer (China Wall)",
}

// GateName returns the human name for a gate number.
func GateName(n int) string {
	if name, ok := gateNames[n]; ok {
		return name
	}
	return "Unknown gate"
}

// GateTimeline builds the deterministic per-gate status list from the stored
// gate_stopped value. Pure (unit-tested).
//
//	gateStopped == 0  → ALL gates passed
//	gateStopped == -1 → structure-permission ASK sentinel (Gate 0)
//	gateStopped 1..5  → gates below passed, that gate stopped, above not reached
//	otherwise         → unknown → every gate "not_reached"
func GateTimeline(gateStopped int) []GateStep {
	if gateStopped == -1 {
		return []GateStep{{N: 0, Name: "Structure Permission", Status: "asked"}}
	}
	if gateStopped < 0 || gateStopped > 5 {
		out := make([]GateStep, 0, 5)
		for n := 1; n <= 5; n++ {
			out = append(out, GateStep{N: n, Name: GateName(n), Status: "not_reached"})
		}
		return out
	}
	out := make([]GateStep, 0, 5)
	for n := 1; n <= 5; n++ {
		status := "passed"
		switch {
		case gateStopped == 0:
			status = "passed"
		case n < gateStopped:
			status = "passed"
		case n == gateStopped:
			status = "stopped"
		default:
			status = "not_reached"
		}
		out = append(out, GateStep{N: n, Name: GateName(n), Status: status})
	}
	return out
}

// ModeLabel maps a decision mode to a short human label. Pure.
func ModeLabel(mode string) string {
	switch mode {
	case "ADVISE":
		return "Answered"
	case "ASK":
		return "Needs more information"
	case "WARN":
		return "Answered with a warning"
	case "PUSH_BACK":
		return "Pushed back"
	case "REFUSE":
		return "Refused"
	default:
		return "Unknown"
	}
}

// ModeExplanation is the deterministic "why" for a decision mode + gate.
// Pure (unit-tested). No LLM: the explanation is derived from stored facts.
func ModeExplanation(mode string, gateStopped int) string {
	switch mode {
	case "ADVISE":
		return "The expert answered using citations retrieved from its corpus; the China Wall required every claim to be backed by a source."
	case "ASK":
		return "The expert asked for more information before answering (Gate 1, information sufficiency): the question lacked enough context for a grounded answer."
	case "WARN":
		return "The expert answered but flagged a charter concern (Gate 3, charter compliance): the question touched a principle the expert must warn about."
	case "PUSH_BACK":
		return "The expert pushed back (Gate 4, necessity check): the request was judged unnecessary or a simpler path was recommended."
	case "REFUSE":
		if gateStopped == 2 {
			return "The expert refused (Gate 2, knowledge coverage): the question is outside the expert's trained domain."
		}
		return "The expert refused (Gate 5, China Wall): the answer could not be fully grounded in cited sources, so it was withheld rather than asserted uncited."
	default:
		return "No decision explanation is available for this message."
	}
}
