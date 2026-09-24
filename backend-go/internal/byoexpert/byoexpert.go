// Package byoexpert owns C8: customer-registered ("bring your own") experts.
//
// WHY a dedicated package:
//   Self-service expert registration + corpus ingest is a distinct capability
//   with its own authorization rule (tenant entitlement + tenant ownership)
//   and its own audit trail. Keeping it in one place mirrors the C4 tenant
//   package: the rule is identical at every call site and fails closed (P3).
//
// DESIGN (§3.1 + C4):
//   Deterministic metadata pre-filter — a byo expert is an `experts` row with
//   origin='byo' and tenant_id = the caller's tenant (migration 033). Reads and
//   asserts are explicit predicates; we never "hope" a list endpoint filters.
//   Entitlement is read from tenants.settings (allow_byo_expert / byo_max_experts)
//   — least-privilege: default deny, an admin grants it per tenant.
//
// FAIL CLOSED (P3):
//   - BYO disabled (kill switch)           → ErrDisabled
//   - caller scope not a tenant (global)   → ErrScopeRequired (admins use the admin API)
//   - scope unresolved (tenant_id NULL)    → ErrScopeRequired
//   - tenant missing/suspended/not entitled→ ErrNotEntitled
//   - quota reached                        → ErrQuotaExceeded
//   - expert not owned by the caller       → ErrNotOwned (never reveals existence)
//
// SOLID: SRP (registration/ingest/audit only); DIP (ingest is an Ingestor
// interface, so this package does not depend on the training pipeline's guts).
package byoexpert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/tenant"
	"ai_avengers/backend/internal/training"
)

// Sentinel errors — the HTTP layer maps these to status codes.
var (
	ErrDisabled      = errors.New("byo experts are disabled")
	ErrNotEntitled   = errors.New("tenant is not entitled to register experts")
	ErrScopeRequired = errors.New("a tenant scope is required (admins use the admin API)")
	ErrQuotaExceeded = errors.New("tenant byo-expert quota exceeded")
	ErrSlugTaken     = errors.New("expert slug is already in use")
	ErrNotFound      = errors.New("expert not found")
	ErrNotOwned      = errors.New("expert is not a tenant-owned byo expert")
	ErrInvalidInput  = errors.New("invalid input")
)

// Expert is the owner-facing view of a byo expert (includes drafts).
type Expert struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Domain          string     `json:"domain"`
	Description     string     `json:"description"`
	TenantID        *uuid.UUID `json:"tenant_id"`
	Origin          string     `json:"origin"`
	CreatedByUserID *uuid.UUID `json:"created_by_user_id"`
	TrainingStatus  string     `json:"training_status"`
	TotalChunks     int        `json:"total_chunks"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Entitlement is the resolved BYO entitlement for a scope.
type Entitlement struct {
	Allowed    bool   `json:"allowed"`
	MaxExperts int    `json:"max_experts"`
	Used       int    `json:"used"`
	Reason     string `json:"reason,omitempty"`
}

// Event is one append-only audit row.
type Event struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  *uuid.UUID      `json:"tenant_id"`
	UserID    *uuid.UUID      `json:"user_id"`
	ExpertID  *uuid.UUID      `json:"expert_id"`
	Action    string          `json:"action"`
	Detail    json.RawMessage `json:"detail"`
	CreatedAt time.Time       `json:"created_at"`
}

// Policy is the configured BYO policy (from config).
type Policy struct {
	// Enabled: master kill switch (BYO_EXPERT_ENABLED). Absent env → true.
	Enabled bool
	// DefaultMaxExperts: per-tenant quota when the tenant has no explicit
	// byo_max_experts setting. Default 5.
	DefaultMaxExperts int
}

// RegisterInput is the validated payload for Register.
type RegisterInput struct {
	Name        string
	Slug        string
	Domain      string
	Description string
}

// Ingestor runs the transcript ingestion pipeline. Implemented by
// *training.IngestionPipeline (wired in cmd/server). Interface here keeps the
// dependency one-way and mockable.
type Ingestor interface {
	IngestTranscript(ctx context.Context, jobID, expertID uuid.UUID, expertName, transcript, sourceFile string, replaceExisting bool) (*training.IngestionResult, error)
	// PrepareTranscript converts an uploaded document into ingestion-ready text
	// (stage 0) and persists it for resume. WHY part of this interface: BYO
	// uploads accept the same formats as the admin path, so the conversion must
	// not be duplicated (or forgotten) in this second entry point.
	PrepareTranscript(ctx context.Context, jobID, expertID uuid.UUID, filename string, data []byte) (string, error)
}

// Service resolves entitlement + owns byo-expert lifecycle.
type Service struct {
	db       *pgxpool.Pool
	tenants  *tenant.Service
	ingestor Ingestor
	policy   Policy
	logger   *zap.Logger
}

// NewService builds the BYO service. ingestor may be nil (ingest disabled).
func NewService(db *pgxpool.Pool, tenants *tenant.Service, ingestor Ingestor, policy Policy, logger *zap.Logger) *Service {
	if policy.DefaultMaxExperts <= 0 {
		policy.DefaultMaxExperts = 5
	}
	return &Service{db: db, tenants: tenants, ingestor: ingestor, policy: policy, logger: logger}
}

// Enabled reports whether BYO is on and the service is wired. Nil-safe.
func (s *Service) Enabled() bool {
	return s != nil && s.policy.Enabled && s.db != nil
}

// Resolve returns the caller's tenant scope (nil-safe; delegates to C4).
func (s *Service) Resolve(ctx context.Context, userID uuid.UUID, role string) (tenant.Scope, error) {
	if s == nil || s.tenants == nil {
		// Unwired tenant layer → global. Register still fails closed
		// (ErrScopeRequired) because a tenant scope is mandatory for BYO.
		return tenant.Scope{Global: true}, nil
	}
	return s.tenants.Resolve(ctx, userID, role)
}

// tenantSettings is the subset of tenants.settings this package reads.
type tenantSettings struct {
	AllowByoExpert bool `json:"allow_byo_expert"`
	ByoMaxExperts  int  `json:"byo_max_experts"`
}

// parseSettings is pure (unit-tested): invalid JSON → zero value (deny).
func parseSettings(raw json.RawMessage) tenantSettings {
	var ts tenantSettings
	if len(raw) == 0 {
		return ts
	}
	_ = json.Unmarshal(raw, &ts)
	return ts
}

// Entitlement resolves whether the scope may register byo experts and how many.
// Reason is set (and Allowed=false) on every denial so callers can explain.
func (s *Service) Entitlement(ctx context.Context, scope tenant.Scope) (Entitlement, error) {
	if !s.Enabled() {
		return Entitlement{Allowed: false, Reason: "byo_disabled"}, nil
	}
	// Global scope (admin / isolation disabled): BYO needs a tenant owner, so
	// it is not available here — admins provision via the admin API.
	if scope.Global {
		return Entitlement{Allowed: false, MaxExperts: s.policy.DefaultMaxExperts, Reason: "global_scope"}, nil
	}
	if scope.TenantID == nil {
		return Entitlement{Allowed: false, Reason: "tenant_scope_unknown"}, nil
	}

	var raw json.RawMessage
	var status string
	err := s.db.QueryRow(ctx,
		`SELECT settings, status FROM tenants WHERE id=$1`, *scope.TenantID,
	).Scan(&raw, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entitlement{Allowed: false, Reason: "tenant_not_found"}, nil
		}
		return Entitlement{}, fmt.Errorf("read tenant entitlement: %w", err)
	}

	ts := parseSettings(raw)
	max := ts.ByoMaxExperts
	if max <= 0 {
		max = s.policy.DefaultMaxExperts
	}
	used, _ := s.countByo(ctx, *scope.TenantID)

	switch {
	case status != "active":
		return Entitlement{Allowed: false, MaxExperts: max, Used: used, Reason: "tenant_suspended"}, nil
	case !ts.AllowByoExpert:
		return Entitlement{Allowed: false, MaxExperts: max, Used: used, Reason: "not_entitled"}, nil
	default:
		return Entitlement{Allowed: true, MaxExperts: max, Used: used}, nil
	}
}

// Register creates a tenant-owned byo expert (status draft). It enforces
// entitlement + quota and records an audit event. Fail closed on any denial.
func (s *Service) Register(ctx context.Context, scope tenant.Scope, userID uuid.UUID, in RegisterInput) (*Expert, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	if scope.Global || scope.TenantID == nil {
		return nil, ErrScopeRequired
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	in.Domain = strings.TrimSpace(in.Domain)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" || in.Slug == "" || in.Domain == "" {
		return nil, fmt.Errorf("%w: name, slug and domain are required", ErrInvalidInput)
	}
	if !ValidSlug(in.Slug) {
		return nil, fmt.Errorf("%w: slug must be lowercase letters, digits and single dashes", ErrInvalidInput)
	}

	ent, err := s.Entitlement(ctx, scope)
	if err != nil {
		return nil, err
	}
	if !ent.Allowed {
		s.recordEvent(ctx, scope, &userID, nil, "denied_entitlement", map[string]interface{}{"reason": ent.Reason})
		return nil, fmt.Errorf("%w: %s", ErrNotEntitled, ent.Reason)
	}
	if ent.Used >= ent.MaxExperts {
		s.recordEvent(ctx, scope, &userID, nil, "denied_quota", map[string]interface{}{"used": ent.Used, "max": ent.MaxExperts})
		return nil, fmt.Errorf("%w (%d/%d)", ErrQuotaExceeded, ent.Used, ent.MaxExperts)
	}

	var e Expert
	err = s.db.QueryRow(ctx,
		`INSERT INTO experts
		   (name, slug, domain, description, tenant_id, origin, created_by_user_id, training_status)
		 VALUES ($1,$2,$3,$4,$5,'byo',$6,'draft')
		 RETURNING id, name, slug, domain, COALESCE(description,''),
		           tenant_id, origin, created_by_user_id, training_status, total_chunks, created_at`,
		in.Name, in.Slug, in.Domain, in.Description, *scope.TenantID, userID,
	).Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
		&e.TenantID, &e.Origin, &e.CreatedByUserID, &e.TrainingStatus, &e.TotalChunks, &e.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrSlugTaken
		}
		return nil, fmt.Errorf("register byo expert: %w", err)
	}

	s.recordEvent(ctx, scope, &userID, &e.ID, "registered", map[string]interface{}{"slug": e.Slug, "domain": e.Domain})
	return &e, nil
}

// ListMine returns the tenant's byo experts (incl. drafts) so the owner can see
// and ingest them. Global scope (admin) sees every byo expert.
func (s *Service) ListMine(ctx context.Context, scope tenant.Scope) ([]Expert, error) {
	if !s.Enabled() {
		return []Expert{}, nil
	}
	q := `SELECT id, name, slug, domain, COALESCE(description,''),
	             tenant_id, origin, created_by_user_id, training_status, total_chunks, created_at
	      FROM experts
	      WHERE origin='byo' AND deleted_at IS NULL`
	args := []interface{}{}
	if !scope.Global {
		if scope.TenantID == nil {
			return []Expert{}, nil // fail closed: no scope → no rows
		}
		q += ` AND tenant_id = $1`
		args = append(args, *scope.TenantID)
	}
	q += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list byo experts: %w", err)
	}
	defer rows.Close()

	out := []Expert{}
	for rows.Next() {
		var e Expert
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
			&e.TenantID, &e.Origin, &e.CreatedByUserID, &e.TrainingStatus, &e.TotalChunks, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetOwned returns a single byo expert the caller is allowed to manage.
func (s *Service) GetOwned(ctx context.Context, scope tenant.Scope, expertID uuid.UUID) (*Expert, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	e, err := s.getOwned(ctx, scope, expertID)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// StartIngest creates an ingestion job for an owned byo expert and runs the
// pipeline in the background (2h timeout, same posture as the admin path).
// Returns the job id immediately.
func (s *Service) StartIngest(ctx context.Context, scope tenant.Scope, userID uuid.UUID, expertID uuid.UUID, filename string, content []byte) (uuid.UUID, error) {
	if !s.Enabled() {
		return uuid.Nil, ErrDisabled
	}
	if s.ingestor == nil {
		return uuid.Nil, fmt.Errorf("%w: ingestion pipeline not wired", ErrDisabled)
	}
	e, err := s.getOwned(ctx, scope, expertID)
	if err != nil {
		return uuid.Nil, err
	}
	if strings.TrimSpace(filename) == "" {
		filename = "transcript.txt"
	}

	var jobID uuid.UUID
	err = s.db.QueryRow(ctx,
		`INSERT INTO ingestion_jobs (expert_id, job_type, status, source_path)
		 VALUES ($1, 'transcript', 'pending', $2)
		 RETURNING id`,
		expertID, filename,
	).Scan(&jobID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create ingestion job: %w", err)
	}

	// NOTE (D3): transcript_content is no longer written here. For non-text
	// uploads (PDF/DOCX/XLSX/...) the stored transcript is the EXTRACTED text,
	// which PrepareTranscript writes in the goroutine below — the bytes at this
	// point may not be text at all. The write is no longer best-effort either:
	// resume depends on it, so a failure fails the job instead of being ignored.
	_, _ = s.db.Exec(ctx,
		`UPDATE experts SET is_training=TRUE, updated_at=NOW() WHERE id=$1`, expertID)

	s.recordEvent(ctx, scope, &userID, &expertID, "ingest_started",
		map[string]interface{}{"job_id": jobID.String(), "file": filename, "bytes": len(content)})

	// Background run — same 2h guard as the admin path so a hung step cannot
	// leak a goroutine forever. Failures are audited, not returned.
	go func() {
		bg, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		// Stage 0 (D3): convert the upload to text. PrepareTranscript marks the
		// job failed with an admin-readable reason if the document cannot be read.
		transcript, prepErr := s.ingestor.PrepareTranscript(bg, jobID, expertID, filename, content)
		if prepErr != nil {
			s.logger.Warn("byo ingest: transcript preparation failed",
				zap.String("job_id", jobID.String()),
				zap.String("expert_id", expertID.String()),
				zap.String("filename", filename),
				zap.Error(prepErr),
			)
			s.recordEvent(bg, scope, &userID, &expertID, "ingest_failed",
				map[string]interface{}{"job_id": jobID.String(), "error": prepErr.Error()})
			return
		}

		_, runErr := s.ingestor.IngestTranscript(bg, jobID, expertID, e.Name, transcript, filename, false)
		if runErr != nil {
			s.logger.Error("byo ingest failed",
				zap.String("job_id", jobID.String()),
				zap.String("expert_id", expertID.String()),
				zap.Error(runErr),
			)
			s.recordEvent(bg, scope, &userID, &expertID, "ingest_failed",
				map[string]interface{}{"job_id": jobID.String(), "error": runErr.Error()})
			return
		}
		s.recordEvent(bg, scope, &userID, &expertID, "ingest_completed",
			map[string]interface{}{"job_id": jobID.String()})
	}()

	return jobID, nil
}

// SetEntitlement grants/revokes BYO for a tenant by merging tenants.settings.
// Used by the admin API. maxExperts <= 0 removes the explicit quota (falls
// back to the policy default).
func (s *Service) SetEntitlement(ctx context.Context, tenantID uuid.UUID, allow bool, maxExperts int) error {
	if s == nil || s.db == nil {
		return ErrDisabled
	}
	var max interface{}
	if maxExperts > 0 {
		max = maxExperts
	}
	tag, err := s.db.Exec(ctx,
		`UPDATE tenants
		 SET settings = settings || jsonb_build_object('allow_byo_expert', $2::boolean, 'byo_max_experts', $3::int),
		     updated_at = NOW()
		 WHERE id=$1`,
		tenantID, allow, max,
	)
	if err != nil {
		return fmt.Errorf("set byo entitlement: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("tenant not found")
	}
	return nil
}

// ListEvents returns the append-only audit trail (admin view). tenantID nil →
// all tenants. limit is clamped to [1, 500].
func (s *Service) ListEvents(ctx context.Context, tenantID *uuid.UUID, limit int) ([]Event, error) {
	if s == nil || s.db == nil {
		return []Event{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := `SELECT id, tenant_id, user_id, expert_id, action, detail, created_at
	      FROM byo_expert_events`
	args := []interface{}{}
	if tenantID != nil {
		q += ` WHERE tenant_id = $1`
		args = append(args, *tenantID)
	}
	q += fmt.Sprintf(` ORDER BY created_at DESC LIMIT %d`, limit)

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list byo events: %w", err)
	}
	defer rows.Close()

	out := []Event{}
	for rows.Next() {
		var ev Event
		if err := rows.Scan(&ev.ID, &ev.TenantID, &ev.UserID, &ev.ExpertID, &ev.Action, &ev.Detail, &ev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

// ============================================================
// internals
// ============================================================

// getOwned loads an expert and enforces: origin='byo' AND (global OR tenant match).
// A missing/foreign expert both surface ErrNotFound/ErrNotOwned without leaking
// existence to the caller.
func (s *Service) getOwned(ctx context.Context, scope tenant.Scope, expertID uuid.UUID) (Expert, error) {
	var e Expert
	err := s.db.QueryRow(ctx,
		`SELECT id, name, slug, domain, COALESCE(description,''),
		        tenant_id, origin, created_by_user_id, training_status, total_chunks, created_at
		 FROM experts WHERE id=$1 AND deleted_at IS NULL`,
		expertID,
	).Scan(&e.ID, &e.Name, &e.Slug, &e.Domain, &e.Description,
		&e.TenantID, &e.Origin, &e.CreatedByUserID, &e.TrainingStatus, &e.TotalChunks, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Expert{}, ErrNotFound
		}
		return Expert{}, fmt.Errorf("load byo expert: %w", err)
	}
	if e.Origin != "byo" {
		// Platform/admin experts are managed through the admin API only.
		return Expert{}, ErrNotOwned
	}
	if !scope.Global {
		if scope.TenantID == nil || e.TenantID == nil || *e.TenantID != *scope.TenantID {
			return Expert{}, ErrNotOwned
		}
	}
	return e, nil
}

// countByo returns how many byo experts a tenant already owns (quota usage).
func (s *Service) countByo(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var n int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM experts
		 WHERE tenant_id=$1 AND origin='byo' AND deleted_at IS NULL`, tenantID,
	).Scan(&n)
	return n, err
}

// recordEvent appends one audit row. Best-effort: an audit failure must never
// break the user-facing operation (the operation itself already succeeded).
func (s *Service) recordEvent(ctx context.Context, scope tenant.Scope, userID, expertID *uuid.UUID, action string, detail map[string]interface{}) {
	if s == nil || s.db == nil {
		return
	}
	var tid *uuid.UUID
	if !scope.Global {
		tid = scope.TenantID
	}
	payload, err := json.Marshal(detail)
	if err != nil {
		payload = []byte("{}")
	}
	if _, err := s.db.Exec(ctx,
		`INSERT INTO byo_expert_events (tenant_id, user_id, expert_id, action, detail)
		 VALUES ($1,$2,$3,$4,$5::jsonb)`,
		tid, userID, expertID, action, string(payload),
	); err != nil && s.logger != nil {
		s.logger.Warn("byo audit write failed",
			zap.String("action", action),
			zap.Error(err),
		)
	}
}

// slugPattern: lowercase alphanumerics separated by single dashes.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidSlug is pure (unit-tested): the accepted expert slug shape.
func ValidSlug(slug string) bool {
	if slug == "" || len(slug) > 120 {
		return false
	}
	return slugPattern.MatchString(slug)
}

// isUniqueViolation reports whether err is a Postgres unique-constraint error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
