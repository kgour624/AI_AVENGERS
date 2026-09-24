package tenant

import (
	"testing"

	"github.com/google/uuid"
)

func ptr(id uuid.UUID) *uuid.UUID { return &id }

func TestScope_Allows(t *testing.T) {
	t1 := uuid.New()
	t2 := uuid.New()

	global := Scope{Global: true}
	if !global.Allows(ptr(t1)) || !global.Allows(nil) {
		t.Error("global must allow everything")
	}

	scoped := Scope{TenantID: ptr(t1)}
	if !scoped.Allows(ptr(t1)) {
		t.Error("same tenant must be allowed")
	}
	if scoped.Allows(ptr(t2)) {
		t.Error("other tenant must be denied")
	}
	if !scoped.Allows(nil) {
		t.Error("platform/global rows (nil tenant) must be visible to all")
	}

	unresolved := Scope{} // no tenant, not global → fail closed on tenant rows
	if unresolved.Allows(ptr(t1)) {
		t.Error("unresolved scope must deny tenant-owned rows")
	}
}

func TestScope_UnresolvedDeniesTenantRowsAllowsGlobal(t *testing.T) {
	// A row with no tenant is a platform row → visible even to an unresolved
	// scope (it carries no tenant secret). A tenant-owned row is denied.
	unresolved := Scope{}
	if !unresolved.Allows(nil) {
		t.Error("platform row should be visible")
	}
	if unresolved.Allows(ptr(uuid.New())) {
		t.Error("tenant row must be denied for unresolved scope")
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Acme Corp":        "acme-corp",
		"  Foo__Bar  ":     "foo-bar",
		"Ünïcode & Co.":    "n-code-co",
		"already-slugged":  "already-slugged",
		"---":              "",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q)=%q want %q", in, got, want)
		}
	}
}

func TestService_NilSafe(t *testing.T) {
	var s *Service
	if s.Enabled() {
		t.Error("nil service must be disabled")
	}
}
