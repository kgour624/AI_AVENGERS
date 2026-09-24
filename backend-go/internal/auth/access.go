package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExpertAccess decides which AI experts a user may list and invoke.
//
// Roles:
//   - admin: all trained experts (admin panel uses separate routes)
//   - client: all trained/active experts (legacy self-register path)
//   - domain_expert: only rows in user_expert_grants
func FilterAllowedExpertIDs(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, role string, requested []uuid.UUID) (allowed []uuid.UUID, denied []uuid.UUID, err error) {
	if len(requested) == 0 {
		return nil, nil, nil
	}
	if role == "admin" || role == "client" {
		return requested, nil, nil
	}

	grantSet, err := GrantedExpertSet(ctx, db, userID)
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

// GrantedExpertSet returns the set of expert IDs granted to userID.
func GrantedExpertSet(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) (map[uuid.UUID]struct{}, error) {
	rows, err := db.Query(ctx,
		`SELECT expert_id FROM user_expert_grants WHERE user_id = $1`,
		userID,
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

// MustHaveExpertAccess returns an error message when domain_expert lacks a grant.
// admin and client always pass.
func MustHaveExpertAccess(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID, role string, expertIDs []uuid.UUID) error {
	_, denied, err := FilterAllowedExpertIDs(ctx, db, userID, role, expertIDs)
	if err != nil {
		return err
	}
	if len(denied) > 0 {
		return fmt.Errorf("one or more experts are not assigned to this account")
	}
	return nil
}
