// Package tenant owns C4: tenant isolation boundaries and enterprise controls.
//
// WHY a dedicated package:
//   Isolation is a cross-cutting authorization concern (data, memory, cost,
//   experts). Living in one place keeps the rule identical at every call site
//   and keeps it fail-closed (P3): if a caller's tenant scope cannot be
//   resolved, every assertion denies rather than lets the request through.
//
// DESIGN (§3.1): a deterministic metadata pre-filter. Every tenant-owned row
// carries an explicit tenant_id column (migration 030) and every read/assert
// is an explicit predicate — never "hope the vector search separates tenants".
// experts.tenant_id NULL = platform/global expert, visible to all tenants
// (this is what keeps today's single-tenant catalog working unchanged).
//
// STAGED: when IsolationEnabled is false the Service resolves a global scope
// and every assertion is a no-op — a pure kill switch for rollback. When true,
// admin is global and non-admin callers resolve their tenant from
// users.tenant_id (backfilled to the default tenant by migration 030).
//
// SOLID: SRP (access decisions only); DIP (message/expert/admin handlers depend
// on the Service, not on raw SQL).
package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/observability"
)

var (
	// ErrScopeUnknown: the caller's tenant could not be determined. Fail closed.
	ErrScopeUnknown = errors.New("tenant scope could not be resolved")
	// ErrTenantDenied: the resource belongs to another tenant (or is unknown).
	ErrTenantDenied = errors.New("resource is outside this tenant")
)

// Tenant is the isolation boundary — an enterprise customer org.
type Tenant struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	Status       string          `json:"status"`
	Settings     json.RawMessage `json:"settings"`
	CreatedAt    time.Time       `json:"created_at"`
	UserCount    int             `json:"user_count"`
	ProjectCount int             `json:"project_count"`
	ExpertCount  int             `json:"expert_count"`
}

// Scope is the resolved isolation scope of a request.
type Scope struct {
	// TenantID is the caller's tenant. nil means "unresolved" (fail closed)
	// unless Global is set.
	TenantID *uuid.UUID
	// Global bypasses tenant filtering (admin, or isolation disabled).
	Global bool
}

// Allows reports whether a row owned by rowTenant is visible in this scope.
// A nil rowTenant is a platform/global row and is visible to every tenant.
// An unresolved non-global scope denies everything (fail closed).
func (s Scope) Allows(rowTenant *uuid.UUID) bool {
	if s.Global {
		return true
	}
	if rowTenant == nil {
		return true
	}
	if s.TenantID == nil {
		return false
	}
	return *s.TenantID == *rowTenant
}

// Service resolves and enforces tenant scope.
type Service struct {
	db      *pgxpool.Pool
	enabled bool
	logger  *zap.Logger
}

// NewService builds the tenant service. enabled=false is the kill switch:
// every scope resolves global and every assertion is a no-op.
func NewService(db *pgxpool.Pool, enabled bool, logger *zap.Logger) *Service {
	return &Service{db: db, enabled: enabled, logger: logger}
}

// Enabled reports whether isolation enforcement is on.
// Safe on a nil receiver (unwired service → disabled → fail-open no-op).
func (s *Service) Enabled() bool {
	return s != nil && s.enabled && s.db != nil
}

// Resolve returns the caller's scope.
// admin (or isolation off) → global. Otherwise the caller's tenant is read
// from users.tenant_id; a NULL tenant fails closed (P3).
func (s *Service) Resolve(ctx context.Context, userID uuid.UUID, role string) (Scope, error) {
	if !s.Enabled() || role == "admin" {
		return Scope{Global: true}, nil
	}
	var tid *uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT tenant_id FROM users WHERE id=$1 AND deleted_at IS NULL`, userID,
	).Scan(&tid)
	if err != nil {
		return Scope{}, fmt.Errorf("resolve tenant scope: %w", err)
	}
	if tid == nil {
		return Scope{}, ErrScopeUnknown
	}
	return Scope{TenantID: tid}, nil
}

// AssertProject denies when projectID is not visible to scope. A missing or
// deleted project also fails closed (the caller must not learn its existence
// and must not proceed).
func (s *Service) AssertProject(ctx context.Context, scope Scope, projectID uuid.UUID) error {
	if !s.Enabled() || scope.Global {
		return nil
	}
	var ptid *uuid.UUID
	err := s.db.QueryRow(ctx,
		`SELECT tenant_id FROM projects WHERE id=$1 AND deleted_at IS NULL`, projectID,
	).Scan(&ptid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			observability.Global.IncTenantDenied()
			return ErrTenantDenied
		}
		return err
	}
	if !scope.Allows(ptid) {
		observability.Global.IncTenantDenied()
		return ErrTenantDenied
	}
	return nil
}

// AssertExperts denies when any requested expert is owned by another tenant.
// Global (NULL-tenant) experts are always allowed. Unknown experts deny.
func (s *Service) AssertExperts(ctx context.Context, scope Scope, expertIDs []uuid.UUID) error {
	if !s.Enabled() || scope.Global || len(expertIDs) == 0 {
		return nil
	}
	if scope.TenantID == nil {
		return ErrScopeUnknown
	}
	rows, err := s.db.Query(ctx,
		`SELECT id FROM experts
		 WHERE id = ANY($1) AND deleted_at IS NULL
		   AND (tenant_id IS NULL OR tenant_id = $2)`,
		expertIDs, *scope.TenantID,
	)
	if err != nil {
		return fmt.Errorf("assert experts: %w", err)
	}
	defer rows.Close()

	visible := make(map[uuid.UUID]struct{}, len(expertIDs))
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		visible[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range expertIDs {
		if _, ok := visible[id]; !ok {
			observability.Global.IncTenantDenied()
			return ErrTenantDenied
		}
	}
	return nil
}

// ============================================================
// ENTERPRISE CONTROLS (admin)
// ============================================================

// List returns every tenant with rolled-up counts (admin view).
func (s *Service) List(ctx context.Context) ([]Tenant, error) {
	if !s.Enabled() {
		return []Tenant{}, nil
	}
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.name, t.slug, t.status, t.settings, t.created_at,
		       (SELECT COUNT(*) FROM users u WHERE u.tenant_id=t.id AND u.deleted_at IS NULL),
		       (SELECT COUNT(*) FROM projects p WHERE p.tenant_id=t.id AND p.deleted_at IS NULL),
		       (SELECT COUNT(*) FROM experts e WHERE e.tenant_id=t.id AND e.deleted_at IS NULL)
		FROM tenants t
		ORDER BY t.created_at`)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	out := []Tenant{}
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.Settings, &t.CreatedAt,
			&t.UserCount, &t.ProjectCount, &t.ExpertCount); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Create adds a tenant. An empty slug is derived from the name.
func (s *Service) Create(ctx context.Context, name, slug string) (*Tenant, error) {
	if !s.Enabled() {
		return nil, errors.New("tenant service not enabled")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("tenant name is required")
	}
	if strings.TrimSpace(slug) == "" {
		slug = slugify(name)
	}
	var t Tenant
	err := s.db.QueryRow(ctx,
		`INSERT INTO tenants (name, slug) VALUES ($1,$2)
		 RETURNING id, name, slug, status, settings, created_at`,
		name, slug,
	).Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.Settings, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}
	return &t, nil
}

// SetUserTenant moves a user into a tenant (enterprise onboarding).
func (s *Service) SetUserTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	if !s.Enabled() {
		return errors.New("tenant service not enabled")
	}
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET tenant_id=$2, updated_at=NOW()
		 WHERE id=$1 AND deleted_at IS NULL`,
		userID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("set user tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

// SetExpertTenant homes an expert into a tenant (NULL = platform/global).
// This is the enterprise control that makes an expert tenant-private.
func (s *Service) SetExpertTenant(ctx context.Context, expertID uuid.UUID, tenantID *uuid.UUID) error {
	if !s.Enabled() {
		return errors.New("tenant service not enabled")
	}
	tag, err := s.db.Exec(ctx,
		`UPDATE experts SET tenant_id=$2, updated_at=NOW()
		 WHERE id=$1 AND deleted_at IS NULL`,
		expertID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("set expert tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("expert not found")
	}
	return nil
}

// slugify lowercases and turns runs of non-alphanumerics into single dashes.
// Pure — unit-tested.
func slugify(name string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
