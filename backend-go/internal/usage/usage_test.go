package usage

import (
	"context"
	"testing"
	"time"
)

func TestMonthStart(t *testing.T) {
	now := time.Date(2026, 9, 24, 15, 30, 0, 0, time.UTC)
	got := MonthStart(now)
	want := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("MonthStart=%v want %v", got, want)
	}
	// Non-UTC input is normalised to UTC month start.
	loc := time.FixedZone("IST", 5*3600+1800)
	got2 := MonthStart(time.Date(2026, 9, 1, 2, 0, 0, 0, loc))
	if got2.Location() != time.UTC {
		t.Errorf("MonthStart should return UTC, got %v", got2.Location())
	}
}

func TestUsagePercent(t *testing.T) {
	if got := UsagePercent(50, 100); got != 50 {
		t.Errorf("got %v want 50", got)
	}
	if got := UsagePercent(10, 0); got != 0 {
		t.Errorf("zero limit must return 0, got %v", got)
	}
}

func TestBudgetBreached(t *testing.T) {
	cases := []struct {
		spend, limit, threshold float64
		wantBreached, wantAlert bool
	}{
		{0, 100, 0.8, false, false},
		{79, 100, 0.8, false, false},
		{80, 100, 0.8, false, true},   // at threshold → alert
		{100, 100, 0.8, true, true},   // at limit → breach
		{150, 100, 0.8, true, true},
		{5, 0, 0.8, false, false},     // no limit → disabled
		{90, 100, 0, false, true},     // bad threshold → default 0.8
		{90, 100, 1.5, false, true},   // out-of-range threshold → default 0.8
	}
	for _, c := range cases {
		b, a := BudgetBreached(c.spend, c.limit, c.threshold)
		if b != c.wantBreached || a != c.wantAlert {
			t.Errorf("BudgetBreached(%v,%v,%v)=(%v,%v) want (%v,%v)",
				c.spend, c.limit, c.threshold, b, a, c.wantBreached, c.wantAlert)
		}
	}
}

func TestAttributionContext(t *testing.T) {
	if _, ok := AttributionFrom(context.Background()); ok {
		t.Error("bare context must carry no attribution")
	}
	a := Attribution{UseCase: UseCaseChat}
	ctx := WithAttribution(context.Background(), a)
	got, ok := AttributionFrom(ctx)
	if !ok || got.UseCase != UseCaseChat {
		t.Errorf("round-trip failed: %+v ok=%v", got, ok)
	}
}

func TestAttributionIsZero(t *testing.T) {
	if !(Attribution{}).IsZero() {
		t.Error("zero value must be IsZero")
	}
	if (Attribution{UseCase: UseCaseChat}).IsZero() {
		t.Error("non-empty attribution must not be IsZero")
	}
}

func TestGroupExpr_WhitelistAndFallback(t *testing.T) {
	if _, lbl := groupExpr("project"); lbl == "" {
		t.Error("project must map")
	}
	// Unknown / injection attempt falls back to tenant, never raw.
	k, _ := groupExpr("tenant_id; DROP TABLE users")
	if k != "COALESCE(u.tenant_id::text,'(global)')" {
		t.Errorf("unknown group-by must fall back to tenant, got %q", k)
	}
}

func TestService_NilSafe(t *testing.T) {
	var s *Service
	if s.Enabled() {
		t.Error("nil service must be disabled")
	}
	rows, err := s.Summary(context.Background(), Filter{})
	if err != nil || rows == nil {
		t.Errorf("nil service Summary must return empty, non-nil slice; got %v %v", rows, err)
	}
	if err := s.Record(context.Background(), Event{CostUSD: 1}); err != nil {
		t.Errorf("nil service Record must be a no-op, got %v", err)
	}
}
