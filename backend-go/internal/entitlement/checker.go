// Package entitlement owns account→expert authorization (sellable-experts seam).
//
// WHY a dedicated package (not auth):
//   Auth is identity (who you are). Entitlement is product access (what you may
//   use). Selling experts externally needs plans/API keys/quotas later without
//   touching chat/workflow call sites — those depend only on ports.EntitlementCheck.
//
// Seed: migration 022 user_expert_grants + role rules (admin/client unrestricted,
// domain_expert grant-scoped). Same behaviour as the previous auth/access.go
// helpers so core chat/workflow/project paths do not change.
//
// SOLID: SRP (only access decisions). DIP (implements ports.EntitlementCheck).
// Pattern: Strategy — Checker is swappable (e.g. future plan-based adapter).
package entitlement

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/ports"
)

// Checker implements ports.EntitlementCheck against user_expert_grants.
type Checker struct {
	db *pgxpool.Pool
}

// NewChecker returns a Postgres-backed entitlement checker.
// Compile-time check that Checker satisfies the port.
var _ ports.EntitlementCheck = (*Checker)(nil)

func NewChecker(db *pgxpool.Pool) *Checker {
	return &Checker{db: db}
}

// Can implements ports.EntitlementCheck.
// capability is reserved for future plan/API-key scopes; Stage 0 treats all
// capabilities the same (grant presence).
func (c *Checker) Can(ctx context.Context, accountID uuid.UUID, role string, expertID uuid.UUID, _ ports.Capability) (ports.AccessDecision, error) {
	if role == "admin" || role == "client" {
		return ports.AccessDecision{Allowed: true}, nil
	}
	set, err := c.GrantedSet(ctx, accountID)
	if err != nil {
		return ports.AccessDecision{}, err
	}
	if _, ok := set[expertID]; ok {
		return ports.AccessDecision{Allowed: true}, nil
	}
	return ports.AccessDecision{
		Allowed: false,
		Reason:  "expert is not assigned to this account",
	}, nil
}

// FilterAllowed implements ports.EntitlementCheck.
func (c *Checker) FilterAllowed(ctx context.Context, accountID uuid.UUID, role string, requested []uuid.UUID) (allowed, denied []uuid.UUID, err error) {
	if len(requested) == 0 {
		return nil, nil, nil
	}
	if role == "admin" || role == "client" {
		return requested, nil, nil
	}

	grantSet, err := c.GrantedSet(ctx, accountID)
	if err != nil {
		return nil, nil, err
	}
	for _, id := range requested {
		if _, ok := grantSet[id]; ok {
			allowed = append(allowed, id)
		} else {
			denied = append(denied, id)
		}
	}
	return allowed, denied, nil
}

// MustAllow implements ports.EntitlementCheck.
func (c *Checker) MustAllow(ctx context.Context, accountID uuid.UUID, role string, expertIDs []uuid.UUID) error {
	_, denied, err := c.FilterAllowed(ctx, accountID, role, expertIDs)
	if err != nil {
		return err
	}
	if len(denied) > 0 {
		observability.Global.IncEntitlementDenied()
		return fmt.Errorf("one or more experts are not assigned to this account")
	}
	return nil
}

// GrantedSet implements ports.EntitlementCheck.
func (c *Checker) GrantedSet(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	rows, err := c.db.Query(ctx,
		`SELECT expert_id FROM user_expert_grants WHERE user_id = $1`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("list expert grants: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		out[id] = struct{}{}
	}
	return out, nil
}
