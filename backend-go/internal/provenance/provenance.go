// Package provenance implements the C1 signed, queryable provenance
// chain for every answer (chat message) and artifact (blackboard event).
//
// §3.1 P9: the B8 span anchors are reused as the provenance primitive —
// a claim maps to a span in the answer and to the chunk(s) that back it.
// The full chain (output → expert → gates fired → model → sources →
// timestamp) is canonicalised, HMAC-signed, and stored append-only, so
// any output can be audited end-to-end and tampering is detectable.
//
// Failure policy: recording is best-effort and MUST NOT fail the answer
// path (fail-open, like B6/B8). If the insert fails, the message/artifact
// was already produced; log and continue.
package provenance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// OutputType discriminates the two provenance sources.
type OutputType string

const (
	OutputChatMessage      OutputType = "chat_message"
	OutputWorkflowArtifact OutputType = "workflow_artifact"
)

// signingKeyID labels the key generation used for a signature so a
// future key rotation can be recognised without re-signing history.
const signingKeyID = "v1"

// Record is a persisted provenance chain.
type Record struct {
	ID           uuid.UUID       `json:"id"`
	OutputType   OutputType      `json:"output_type"`
	OutputID     uuid.UUID       `json:"output_id"`
	ProjectID    *uuid.UUID      `json:"project_id,omitempty"`
	ExpertID     *uuid.UUID      `json:"expert_id,omitempty"`
	ChatID       *uuid.UUID      `json:"chat_id,omitempty"`
	WorkflowID   *uuid.UUID      `json:"workflow_id,omitempty"`
	Model        string          `json:"model,omitempty"`
	DecisionMode string          `json:"decision_mode,omitempty"`
	GateStopped  *int            `json:"gate_stopped,omitempty"`
	ContentHash  string          `json:"content_hash"`
	Claims       json.RawMessage `json:"claims"`
	Citations    json.RawMessage `json:"citations"`
	// Chain is the canonical JSON text that was signed. Exact bytes.
	Chain      string    `json:"chain"`
	Signature  string    `json:"signature"`
	SigningKey string    `json:"signing_key_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChatAnswer is the input for recording a chat answer's provenance.
type ChatAnswer struct {
	MessageID    uuid.UUID
	ChatID       uuid.UUID
	ProjectID    *uuid.UUID
	ExpertID     *uuid.UUID
	Model        string
	DecisionMode string
	GateStopped  *int
	Content      string
	// Claims/Citations are stored as JSONB. interface{} keeps this package
	// independent of chinawall (same convention as chat.Message.Citations).
	Claims    interface{}
	Citations interface{}
}

// Artifact is the input for recording a workflow artifact's provenance.
type Artifact struct {
	EventID     uuid.UUID
	WorkflowID  uuid.UUID
	EventType   string
	ExpertID    *uuid.UUID
	ContentHash string
	References  []uuid.UUID
}

// chainDoc is the canonical, deterministic shape that gets signed.
// Struct field order (not map order) makes json.Marshal byte-stable.
type chainDoc struct {
	OutputType   string          `json:"output_type"`
	OutputID     string          `json:"output_id"`
	ProjectID    string          `json:"project_id,omitempty"`
	ExpertID     string          `json:"expert_id,omitempty"`
	ChatID       string          `json:"chat_id,omitempty"`
	WorkflowID   string          `json:"workflow_id,omitempty"`
	EventType    string          `json:"event_type,omitempty"`
	Model        string          `json:"model,omitempty"`
	DecisionMode string          `json:"decision_mode,omitempty"`
	GateStopped  *int            `json:"gate_stopped,omitempty"`
	ContentHash  string          `json:"content_hash"`
	References   []string        `json:"references,omitempty"`
	Claims       json.RawMessage `json:"claims"`
	Citations    json.RawMessage `json:"citations"`
	CreatedAt    string          `json:"created_at"`
}

// Service records and fetches provenance chains.
type Service struct {
	db      *pgxpool.Pool
	key     []byte
	enabled bool
	logger  *zap.Logger
}

// NewService builds a provenance service. enabled=false (or an empty key)
// makes every Record* call a no-op — the rollback switch.
func NewService(db *pgxpool.Pool, key []byte, enabled bool, logger *zap.Logger) *Service {
	return &Service{db: db, key: key, enabled: enabled, logger: logger}
}

// Enabled reports whether recording is active.
func (s *Service) Enabled() bool {
	return s != nil && s.enabled && s.db != nil && len(s.key) > 0
}

// RecordChatAnswer stores a signed chain for a saved assistant message.
// Best-effort: returns an error for logging, but callers treat it as
// non-fatal (answer already shipped).
func (s *Service) RecordChatAnswer(ctx context.Context, in ChatAnswer) error {
	if !s.Enabled() || in.MessageID == uuid.Nil {
		return nil
	}
	claims := normalizeJSON(in.Claims)
	cites := normalizeJSON(in.Citations)
	hash := contentHash(in.Content)
	now := time.Now().UTC()

	// Resolve project_id from the chat when the caller did not supply it,
	// so the chain is traceable to a project without extra handler plumbing.
	projectID := in.ProjectID
	if projectID == nil {
		projectID = s.resolveProject(ctx, in.ChatID)
	}

	doc := chainDoc{
		OutputType:   string(OutputChatMessage),
		OutputID:     in.MessageID.String(),
		ProjectID:    uuidStr(projectID),
		ExpertID:     uuidStr(in.ExpertID),
		ChatID:       in.ChatID.String(),
		Model:        in.Model,
		DecisionMode: in.DecisionMode,
		GateStopped:  in.GateStopped,
		ContentHash:  hash,
		Claims:       claims,
		Citations:    cites,
		CreatedAt:    now.Format(time.RFC3339Nano),
	}
	return s.insert(ctx, OutputChatMessage, in.MessageID, projectID, in.ExpertID,
		&in.ChatID, nil, in.Model, in.DecisionMode, in.GateStopped, hash, claims, cites, doc, now)
}

// resolveProject maps a chat to its project_id (best-effort). Returns nil
// on any error so a missing lookup never blocks provenance.
func (s *Service) resolveProject(ctx context.Context, chatID uuid.UUID) *uuid.UUID {
	if s == nil || s.db == nil || chatID == uuid.Nil {
		return nil
	}
	var projectID uuid.UUID
	if err := s.db.QueryRow(ctx,
		`SELECT project_id FROM chats WHERE id=$1`, chatID,
	).Scan(&projectID); err != nil || projectID == uuid.Nil {
		return nil
	}
	return &projectID
}

// RecordArtifact stores a signed chain for a blackboard artifact event.
func (s *Service) RecordArtifact(ctx context.Context, in Artifact) error {
	if !s.Enabled() || in.EventID == uuid.Nil {
		return nil
	}
	hash := in.ContentHash
	if hash == "" {
		hash = contentHash(in.EventID.String())
	}
	refs := make([]string, 0, len(in.References))
	for _, r := range in.References {
		refs = append(refs, r.String())
	}
	now := time.Now().UTC()
	doc := chainDoc{
		OutputType:  string(OutputWorkflowArtifact),
		OutputID:    in.EventID.String(),
		WorkflowID:  in.WorkflowID.String(),
		ExpertID:    uuidStr(in.ExpertID),
		EventType:   in.EventType,
		ContentHash: hash,
		References:  refs,
		Claims:      json.RawMessage("[]"),
		Citations:   json.RawMessage("[]"),
		CreatedAt:   now.Format(time.RFC3339Nano),
	}
	wf := in.WorkflowID
	return s.insert(ctx, OutputWorkflowArtifact, in.EventID, nil, in.ExpertID,
		nil, &wf, "", "", nil, hash, json.RawMessage("[]"), json.RawMessage("[]"), doc, now)
}

func (s *Service) insert(
	ctx context.Context,
	outputType OutputType,
	outputID uuid.UUID,
	projectID, expertID *uuid.UUID,
	chatID, workflowID *uuid.UUID,
	model, decisionMode string,
	gateStopped *int,
	hash string,
	claims, cites json.RawMessage,
	doc chainDoc,
	createdAt time.Time,
) error {
	chainBytes, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("provenance: marshal chain: %w", err)
	}
	sig := sign(s.key, chainBytes)

	tag, err := s.db.Exec(ctx,
		`INSERT INTO provenance_records
			(output_type, output_id, project_id, expert_id, chat_id, workflow_id,
			 model, decision_mode, gate_stopped, content_hash, claims, citations,
			 chain, signature, signing_key_id, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		 ON CONFLICT (output_type, output_id) DO NOTHING`,
		string(outputType), outputID, projectID, expertID, chatID, workflowID,
		nullStr(model), nullStr(decisionMode), gateStopped, hash,
		string(claims), string(cites), string(chainBytes), sig, signingKeyID, createdAt,
	)
	if err != nil {
		return fmt.Errorf("provenance: insert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Append-only: an existing chain is never rewritten.
		s.logger.Debug("provenance already recorded, skipped",
			zap.String("output_type", string(outputType)),
			zap.String("output_id", outputID.String()),
		)
	}
	return nil
}

// Get returns the stored chain for one output, or (nil, nil) when absent.
func (s *Service) Get(ctx context.Context, outputType OutputType, outputID uuid.UUID) (*Record, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var r Record
	var claims, cites []byte
	err := s.db.QueryRow(ctx,
		`SELECT id, output_type, output_id, project_id, expert_id, chat_id, workflow_id,
		        COALESCE(model,''), COALESCE(decision_mode,''), gate_stopped, content_hash,
		        claims, citations, chain, signature, signing_key_id, created_at
		 FROM provenance_records
		 WHERE output_type=$1 AND output_id=$2`,
		string(outputType), outputID,
	).Scan(
		&r.ID, &r.OutputType, &r.OutputID, &r.ProjectID, &r.ExpertID, &r.ChatID, &r.WorkflowID,
		&r.Model, &r.DecisionMode, &r.GateStopped, &r.ContentHash,
		&claims, &cites, &r.Chain, &r.Signature, &r.SigningKey, &r.CreatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	r.Claims = json.RawMessage(claims)
	r.Citations = json.RawMessage(cites)
	return &r, nil
}

// verifyChain recomputes the HMAC over the stored chain text and compares
// it to the stored signature (constant-time). Pure — unit-tested.
func verifyChain(key []byte, chain, signature string) bool {
	if len(key) == 0 || chain == "" || signature == "" {
		return false
	}
	want := sign(key, []byte(chain))
	return hmac.Equal([]byte(want), []byte(signature))
}

// Verify checks the signature of a fetched record.
func (s *Service) Verify(r *Record) bool {
	if s == nil || r == nil {
		return false
	}
	return verifyChain(s.key, r.Chain, r.Signature)
}

// sign returns hex(HMAC-SHA256(key, msg)). Pure — unit-tested.
func sign(key, msg []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}

// contentHash returns hex(sha256(text)). Pure — unit-tested.
func contentHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// normalizeJSON marshals v to compact JSON, defaulting to "[]" for nil so
// JSONB columns never hold SQL NULL and the chain stays byte-stable.
func normalizeJSON(v interface{}) json.RawMessage {
	if v == nil {
		return json.RawMessage("[]")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("[]")
	}
	if len(b) == 0 || string(b) == "null" {
		return json.RawMessage("[]")
	}
	return json.RawMessage(b)
}

// artifactEventTypes mirrors the workflow ReviewerMatrix artifact types
// (reviewers.go) plus the authoring/amendment outputs. Only these get a
// provenance chain — questions/notifications/reviews do not.
var artifactEventTypes = func() map[string]struct{} {
	types := []string{
		"requirement_captured",
		"architecture_decision",
		"data_model_proposed",
		"api_contract_proposed",
		"module_design_proposed",
		"code_artifact_produced",
		"test_case_proposed",
		"design_section_written",
		"design_amended",
	}
	m := make(map[string]struct{}, len(types))
	for _, t := range types {
		m[t] = struct{}{}
	}
	return m
}()

// IsArtifactEventType reports whether an event type is a produced artifact
// (eligible for provenance). Pure — unit-tested.
func IsArtifactEventType(eventType string) bool {
	_, ok := artifactEventTypes[eventType]
	return ok
}

// RecordArtifactProvenance satisfies blackboard.ArtifactProvenanceRecorder.
// Called by blackboard.Store after a successful Post; non-artifact event
// types are ignored. Non-nil-safe: silently no-ops when disabled.
func (s *Service) RecordArtifactProvenance(
	ctx context.Context,
	workflowID, eventID uuid.UUID,
	eventType string,
	expertID *uuid.UUID,
	contentHash string,
	refs []uuid.UUID,
) {
	if !s.Enabled() || !IsArtifactEventType(eventType) {
		return
	}
	// Fresh background context: the recorder runs after Post returns and
	// must not be cancelled by the request context ending.
	_ = s.RecordArtifact(context.WithoutCancel(ctx), Artifact{
		EventID:     eventID,
		WorkflowID:  workflowID,
		EventType:   eventType,
		ExpertID:    expertID,
		ContentHash: contentHash,
		References:  refs,
	})
}

func uuidStr(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// nullStr maps "" to SQL NULL for nullable text columns so absence is
// honest (not an empty string masquerading as a value).
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
