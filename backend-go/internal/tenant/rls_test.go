package tenant

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// The tenant id must travel as a bind parameter. A statement built by
// concatenation is the one way this helper could become the leak it exists to
// prevent, so the property is pinned rather than trusted.
func TestSettingStatementsNeverSpliceTheTenantID(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sql, args := SettingStatements(Scope{TenantID: &id})

	if strings.Contains(sql, id.String()) {
		t.Fatalf("the tenant id was spliced into the SQL: %q", sql)
	}
	if !strings.Contains(sql, "$1") {
		t.Fatalf("expected a bind parameter, got %q", sql)
	}
	if !strings.Contains(sql, "set_config") || !strings.Contains(sql, "true") {
		t.Fatalf("expected a transaction-local set_config, got %q", sql)
	}
	if len(args) != 1 || args[0] != id.String() {
		t.Fatalf("args = %v, want the tenant id", args)
	}
}

// Global scope is the privileged path, and it must say so explicitly.
func TestSettingStatementsForGlobalScope(t *testing.T) {
	sql, args := SettingStatements(Scope{Global: true})
	if !strings.Contains(sql, "app.system") || !strings.Contains(sql, "'on'") {
		t.Fatalf("global scope should set app.system=on, got %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("global scope needs no arguments, got %v", args)
	}
}

// An unknown tenant must NOT become a global scope — that inversion is how a
// fail-closed check turns into a fail-open one.
func TestScopeForUserIDRefusesUnknownTenant(t *testing.T) {
	if _, err := ScopeForUserID(nil); err == nil {
		t.Fatal("a nil tenant id must fail closed, not run as global")
	}
	id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	scope, err := ScopeForUserID(&id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scope.Global || scope.TenantID == nil || *scope.TenantID != id {
		t.Fatalf("scope = %+v, want the caller's own tenant", scope)
	}
}
