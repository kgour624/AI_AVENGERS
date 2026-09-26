package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scope is declared in tenant.go; this file holds the database-side half of the
// isolation added in A11.

// SettingStatements returns the transaction-local settings that scope a
// transaction to this Scope, and the arguments for them.
//
// Pure — unit-tested. WHY it returns SQL and arguments rather than a finished
// string: the tenant id goes through a bind parameter (set_config), so a value
// from outside can never be spliced into SQL. Building the statement by string
// concatenation is the one way this helper could become the leak it exists to
// prevent.
//
// Transactions are transaction-LOCAL on purpose. A pooled connection is reused, so
// a session-wide setting would carry one request's tenant into the next request
// that borrows the same connection — a leak that would look like an intermittent
// ghost rather than a bug.
func SettingStatements(scope Scope) (string, []interface{}) {
	if scope.Global {
		// Global == system scope: admin traffic and background work. It is an
		// explicit statement about the transaction, and the policies only honour
		// this exact value, so "privileged" is always a decision someone made.
		return "SELECT set_config('app.system', 'on', true)", nil
	}
	if scope.TenantID == nil {
		// A scope that is neither global nor owned by a tenant is UNKNOWN, and the
		// only safe reading of "I do not know which client this is" is to see
		// nothing. Setting the tenant to the empty string makes
		// app_current_tenant() NULL, which every policy denies.
		return "SELECT set_config('app.tenant_id', '', true)", nil
	}
	return "SELECT set_config('app.tenant_id', $1, true)", []interface{}{scope.TenantID.String()}
}

// ApplyScope sets the isolation settings on an open transaction.
func ApplyScope(ctx context.Context, tx pgx.Tx, scope Scope) error {
	sql, args := SettingStatements(scope)
	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("apply tenant scope: %w", err)
	}
	return nil
}

// WithTenantTx runs fn inside a transaction scoped to `scope`.
//
// This is the shape every request-scoped read should take once the deployment
// moves request traffic to the non-owner role (see migration 053 for the steps).
// It is provided now, before enforcement is switched on, for the same reason the
// policies are: the switch is a configuration change, and the code it needs should
// already exist and be tested when that change is made — not be written in a hurry
// during it.
func WithTenantTx(ctx context.Context, pool *pgxpool.Pool, scope Scope, fn func(tx pgx.Tx) error) error {
	if pool == nil {
		return fmt.Errorf("with tenant tx: no pool")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("with tenant tx: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ApplyScope(ctx, tx, scope); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("with tenant tx: commit: %w", err)
	}
	return nil
}

// ScopeForUserID is the scope a plain tenant user's requests run under.
//
// A nil tenant id is NOT turned into a global scope: an unknown tenant must fail
// closed, which is what Resolve already does for a NULL users.tenant_id. Note that
// even if this check were removed, SettingStatements would still deny — the empty
// tenant setting is denied by every policy — so the two cannot drift into a
// fail-open pair.
func ScopeForUserID(tenantID *uuid.UUID) (Scope, error) {
	if tenantID == nil {
		return Scope{}, ErrScopeUnknown
	}
	return Scope{TenantID: tenantID}, nil
}
