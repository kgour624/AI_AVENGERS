package reliability

import (
	"context"
	"testing"
)

func TestAvailability(t *testing.T) {
	if got := Availability(0, 0); got != 1.0 {
		t.Errorf("no traffic → 1.0, got %v", got)
	}
	if got := Availability(100, 0); got != 1.0 {
		t.Errorf("100/0 errors → 1.0, got %v", got)
	}
	if got := Availability(100, 5); got < 0.949 || got > 0.951 {
		t.Errorf("100 calls/5 errs → ~0.95, got %v", got)
	}
	// Errors clamp to calls (never negative availability).
	if got := Availability(10, 50); got != 0 {
		t.Errorf("errors>calls → 0, got %v", got)
	}
	if got := Availability(10, -3); got != 1.0 {
		t.Errorf("negative errors clamp to 0 → 1.0, got %v", got)
	}
}

func TestErrorBudgetRemaining(t *testing.T) {
	// target 0.995 → allowed error rate 0.005.
	if got := ErrorBudgetRemaining(0, 0.995); got != 1.0 {
		t.Errorf("no errors → full budget, got %v", got)
	}
	if got := ErrorBudgetRemaining(0.005, 0.995); got != 0 {
		t.Errorf("exactly at budget → 0, got %v", got)
	}
	if got := ErrorBudgetRemaining(0.01, 0.995); got > 0 {
		t.Errorf("over budget → negative, got %v", got)
	}
	if got := ErrorBudgetRemaining(0.5, 1.0); got != 1.0 {
		t.Errorf("degenerate target → 1.0, got %v", got)
	}
}

func TestVerdict(t *testing.T) {
	if got := Verdict(0, 1.0, 0.25); got != VerdictUnknown {
		t.Errorf("no traffic → unknown, got %q", got)
	}
	if got := Verdict(100, 0.5, 0.25); got != VerdictOK {
		t.Errorf("healthy budget → ok, got %q", got)
	}
	if got := Verdict(100, 0.10, 0.25); got != VerdictAtRisk {
		t.Errorf("low budget → at_risk, got %q", got)
	}
	if got := Verdict(100, 0, 0.25); got != VerdictBreached {
		t.Errorf("zero budget → breached, got %q", got)
	}
	if got := Verdict(100, -1.0, 0.25); got != VerdictBreached {
		t.Errorf("negative budget → breached, got %q", got)
	}
}

func TestTransitionEvent(t *testing.T) {
	cases := []struct {
		prev, cur string
		kind      string
		ok        bool
	}{
		{"", "unhealthy", KindDegraded, true},   // first observation, down
		{"", "healthy", "", false},              // first observation, fine → no event
		{"healthy", "unhealthy", KindDegraded, true},
		{"unhealthy", "healthy", KindRecovered, true},
		{"healthy", "healthy", "", false},       // steady state → no event
		{"unhealthy", "unhealthy", "", false},   // steady state → no event
	}
	for _, c := range cases {
		kind, ok := TransitionEvent(c.prev, c.cur)
		if kind != c.kind || ok != c.ok {
			t.Errorf("TransitionEvent(%q,%q)=(%q,%v) want (%q,%v)", c.prev, c.cur, kind, ok, c.kind, c.ok)
		}
	}
}

func TestNewService_Defaults(t *testing.T) {
	s := NewService(nil, Policy{Enabled: true}, nil)
	p := s.Policy()
	if p.AvailabilityTarget != 0.995 {
		t.Errorf("default target = %v want 0.995", p.AvailabilityTarget)
	}
	if p.ErrorBudgetWindowDays != 30 {
		t.Errorf("default window = %d want 30", p.ErrorBudgetWindowDays)
	}
	if p.AtRiskThreshold != 0.25 {
		t.Errorf("default at-risk = %v want 0.25", p.AtRiskThreshold)
	}
	// Explicit values are preserved.
	s2 := NewService(nil, Policy{Enabled: true, AvailabilityTarget: 0.99, ErrorBudgetWindowDays: 7, AtRiskThreshold: 0.5}, nil)
	p2 := s2.Policy()
	if p2.AvailabilityTarget != 0.99 || p2.ErrorBudgetWindowDays != 7 || p2.AtRiskThreshold != 0.5 {
		t.Errorf("explicit policy not preserved: %+v", p2)
	}
}

func TestService_EnabledNilSafe(t *testing.T) {
	var nilSvc *Service
	if nilSvc.Enabled() {
		t.Error("nil service must be disabled")
	}
	nilSvc.RegisterProbe("x", func(_ context.Context) error { return nil }) // must not panic
	if got := len(nilSvc.Components(context.Background())); got != 0 {
		t.Errorf("nil Components → empty, got %d", got)
	}
}
