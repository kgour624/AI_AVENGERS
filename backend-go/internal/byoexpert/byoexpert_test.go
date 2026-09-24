package byoexpert

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"ai_avengers/backend/internal/tenant"
)

func TestValidSlug(t *testing.T) {
	ok := []string{"my-expert", "expert1", "a", "sales-tax-2026", "x9-y8"}
	bad := []string{"", "My-Expert", "expert_1", "expert 1", "-lead", "trail-", "double--dash",
		"UPPER", "with.dot", "with/slash"}
	for _, s := range ok {
		if !ValidSlug(s) {
			t.Errorf("ValidSlug(%q)=false want true", s)
		}
	}
	for _, s := range bad {
		if ValidSlug(s) {
			t.Errorf("ValidSlug(%q)=true want false", s)
		}
	}
	// Length cap.
	long := make([]byte, 130)
	for i := range long {
		long[i] = 'a'
	}
	if ValidSlug(string(long)) {
		t.Error("over-long slug must be rejected")
	}
}

func TestParseSettings_FailClosed(t *testing.T) {
	// Missing/empty/garbage settings → zero value (deny, quota default).
	for _, raw := range []string{"", "null", "not-json", "[]"} {
		got := parseSettings(json.RawMessage(raw))
		if got.AllowByoExpert {
			t.Errorf("parseSettings(%q) allowed=true want false", raw)
		}
		if got.ByoMaxExperts != 0 {
			t.Errorf("parseSettings(%q) max=%d want 0", raw, got.ByoMaxExperts)
		}
	}

	got := parseSettings(json.RawMessage(`{"allow_byo_expert":true,"byo_max_experts":7,"other":"x"}`))
	if !got.AllowByoExpert {
		t.Error("expected allow_byo_expert=true")
	}
	if got.ByoMaxExperts != 7 {
		t.Errorf("ByoMaxExperts=%d want 7", got.ByoMaxExperts)
	}
}

func TestService_EnabledNilSafe(t *testing.T) {
	var nilSvc *Service
	if nilSvc.Enabled() {
		t.Error("nil service must report disabled")
	}
	// Disabled by policy even though wired.
	s := NewService(nil, nil, nil, Policy{Enabled: false}, nil)
	if s.Enabled() {
		t.Error("policy-disabled service must report disabled")
	}
	// Enabled requires a db handle (fail closed when unwired).
	s2 := NewService(nil, nil, nil, Policy{Enabled: true}, nil)
	if s2.Enabled() {
		t.Error("service without db must report disabled")
	}
}

func TestNewService_DefaultQuota(t *testing.T) {
	s := NewService(nil, nil, nil, Policy{Enabled: true, DefaultMaxExperts: 0}, nil)
	if s.policy.DefaultMaxExperts != 5 {
		t.Errorf("DefaultMaxExperts=%d want 5", s.policy.DefaultMaxExperts)
	}
	s2 := NewService(nil, nil, nil, Policy{Enabled: true, DefaultMaxExperts: 9}, nil)
	if s2.policy.DefaultMaxExperts != 9 {
		t.Errorf("DefaultMaxExperts=%d want 9", s2.policy.DefaultMaxExperts)
	}
}

func TestEntitlement_DisabledKillSwitch(t *testing.T) {
	// Disabled (kill switch) → denied with reason, no error, no DB touch.
	// (The tenant-specific branches need a live pool and are covered by the
	// manual/DB checks, consistent with the other C-branch pure tests.)
	dis := NewService(nil, nil, nil, Policy{Enabled: false}, nil)
	ent, err := dis.Entitlement(context.Background(), tenant.Scope{TenantID: ptrUUID(uuid.New())})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ent.Allowed || ent.Reason != "byo_disabled" {
		t.Fatalf("disabled service must deny with byo_disabled, got %+v", ent)
	}
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
