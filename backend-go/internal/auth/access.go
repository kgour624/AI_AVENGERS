package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ai_avengers/backend/internal/entitlement"
	"ai_avengers/backend/internal/ports"
)

// Default entitlement checker used by thin wrappers below.
// Callers that already hold a ports.EntitlementCheck should use it directly;
// these helpers keep pre-Stage-0 call sites working without signature churn.
func checker(db *pgxpool.Pool) ports.EntitlementCheck {
	return entitlement.NewChecker(db)
}

// FilterAllowedExpertIDs decides which AI experts a user may list and invoke.
//
// Roles:
//   - admin: all trained experts (admin panel uses separate routes)
//   - client: all trained/active experts (legacy self-register path)
//   - domain_expert: only rows in user_expert_grants
//
// Delegates to entitlement.Checker (ports.EntitlementCheck) so sellable-expert
// plans can land later without rewriting chat/workflow/project.
func FilterAllowedExpertIDs(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, role string, requested []uuid.UUID) (allowed []uuid.UUID, denied []uuid.UUID, err error) {
	return checker(db).FilterAllowed(ctx, userID, role, requested)
}

// GrantedExpertSet returns the set of expert IDs granted to userID.
func GrantedExpertSet(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	return checker(db).GrantedSet(ctx, userID)
}

// MustHaveExpertAccess returns an error message when domain_expert lacks a grant.
// admin and client always pass.
func MustHaveExpertAccess(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, role string, expertIDs []uuid.UUID) error {
	return checker(db).MustAllow(ctx, userID, role, expertIDs)
}
