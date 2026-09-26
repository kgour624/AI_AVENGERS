package gateway

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

// fixedClock lets the open → cooldown → half-open cycle be tested without
// sleeping through a real minute.
type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time            { return c.t }
func (c *fixedClock) Advance(d time.Duration)   { c.t = c.t.Add(d) }

func newTestBreaker(cfg BreakerConfig) (*Breaker, *fixedClock) {
	clock := &fixedClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	b := NewBreaker(cfg, zap.NewNop())
	b.SetClock(clock.Now)
	return b, clock
}

// A healthy provider is never skipped, however many calls it serves.
func TestBreakerStaysClosedWhileCallsSucceed(t *testing.T) {
	b, _ := newTestBreaker(BreakerConfig{Enabled: true, FailureThreshold: 3, Cooldown: time.Minute})
	for i := 0; i < 50; i++ {
		if allowed, reason := b.Allow("anthropic"); !allowed {
			t.Fatalf("call %d was skipped: %s", i, reason)
		}
		b.RecordSuccess("anthropic")
	}
	if st := b.Snapshot(); len(st) != 1 || st[0].State != BreakerClosed {
		t.Fatalf("state = %+v, want a single closed entry", st)
	}
}

// The breaker opens only at the threshold, and one success resets the count —
// an occasionally flaky provider must never be taken out of rotation.
func TestBreakerOpensOnlyAfterConsecutiveFailures(t *testing.T) {
	b, _ := newTestBreaker(BreakerConfig{Enabled: true, FailureThreshold: 3, Cooldown: time.Minute})
	fail := errors.New("boom")

	b.RecordFailure("openrouter", fail)
	b.RecordFailure("openrouter", fail)
	if allowed, _ := b.Allow("openrouter"); !allowed {
		t.Fatal("two failures must not open a threshold of three")
	}

	b.RecordSuccess("openrouter") // the flaky call that worked
	b.RecordFailure("openrouter", fail)
	b.RecordFailure("openrouter", fail)
	if allowed, _ := b.Allow("openrouter"); !allowed {
		t.Fatal("a success must reset the consecutive count")
	}

	b.RecordFailure("openrouter", fail)
	allowed, reason := b.Allow("openrouter")
	if allowed {
		t.Fatal("three consecutive failures must open the breaker")
	}
	if reason == "" {
		t.Fatal("a skipped call must come with a reason the log can print")
	}
}

// After the cooldown exactly one probe goes through; a success closes the
// breaker, a failure restarts the cooldown.
func TestBreakerHalfOpenThenRecoversOrReopens(t *testing.T) {
	b, clock := newTestBreaker(BreakerConfig{Enabled: true, FailureThreshold: 2, Cooldown: 30 * time.Second})
	fail := errors.New("still down")

	b.RecordFailure("deepseek", fail)
	b.RecordFailure("deepseek", fail)
	if allowed, _ := b.Allow("deepseek"); allowed {
		t.Fatal("breaker should be open inside the cooldown")
	}

	clock.Advance(31 * time.Second)
	allowed, _ := b.Allow("deepseek")
	if !allowed {
		t.Fatal("after the cooldown one probe must be allowed")
	}
	// A second concurrent caller must not join the probe.
	if again, _ := b.Allow("deepseek"); again {
		t.Fatal("only one probe may be in flight")
	}

	// The probe succeeded.
	b.RecordSuccess("deepseek")
	if allowed, _ := b.Allow("deepseek"); !allowed {
		t.Fatal("a successful probe must close the breaker")
	}

	// Now the probe fails instead: the cooldown restarts.
	b.RecordFailure("deepseek", fail)
	b.RecordFailure("deepseek", fail)
	clock.Advance(31 * time.Second)
	b.Allow("deepseek") // probe
	b.RecordFailure("deepseek", fail)
	if allowed, reason := b.Allow("deepseek"); allowed {
		t.Fatalf("a failed probe must reopen the breaker, got allowed (%s)", reason)
	}
}

// The kill switch must restore the old behaviour exactly.
func TestBreakerDisabledNeverSkips(t *testing.T) {
	b, _ := newTestBreaker(BreakerConfig{Enabled: false, FailureThreshold: 1, Cooldown: time.Minute})
	for i := 0; i < 10; i++ {
		b.RecordFailure("gemini", errors.New("down"))
	}
	if allowed, reason := b.Allow("gemini"); !allowed {
		t.Fatalf("a disabled breaker must always allow, got skipped: %s", reason)
	}
	if st := b.Snapshot(); len(st) != 0 {
		t.Fatalf("a disabled breaker should report nothing, got %+v", st)
	}
}

// A zero config must not silently disable the protection.
func TestBreakerZeroConfigUsesDefaults(t *testing.T) {
	b := NewBreaker(BreakerConfig{}, zap.NewNop())
	cfg := b.Config()
	if !cfg.Enabled || cfg.FailureThreshold <= 0 || cfg.Cooldown <= 0 {
		t.Fatalf("zero config produced %+v, want the defaults", cfg)
	}
}

// Providers are tracked independently: one being down must not skip another.
func TestBreakerTracksProvidersSeparately(t *testing.T) {
	b, _ := newTestBreaker(BreakerConfig{Enabled: true, FailureThreshold: 1, Cooldown: time.Minute})
	b.RecordFailure("cavoti", errors.New("down"))

	if allowed, _ := b.Allow("cavoti"); allowed {
		t.Fatal("the failing provider must be skipped")
	}
	if allowed, _ := b.Allow("anthropic"); !allowed {
		t.Fatal("a healthy provider must not be affected")
	}
	if st := b.Snapshot(); len(st) != 1 || st[0].Provider != "cavoti" {
		t.Fatalf("snapshot = %+v, want only the provider with an opinion", st)
	}
}

// The status surface must be readable: retry_after while open, half_open after
// the cooldown, and a stable order.
func TestBreakerSnapshotReport(t *testing.T) {
	b, clock := newTestBreaker(BreakerConfig{Enabled: true, FailureThreshold: 1, Cooldown: 60 * time.Second})
	b.RecordFailure("zeta", errors.New("down"))
	b.RecordFailure("alpha", errors.New("down"))

	st := b.Snapshot()
	if len(st) != 2 || st[0].Provider != "alpha" || st[1].Provider != "zeta" {
		t.Fatalf("snapshot order = %+v, want alpha before zeta", st)
	}
	if st[0].State != BreakerOpen || st[0].RetryAfterSeconds <= 0 {
		t.Fatalf("open entry = %+v, want an open state with a retry hint", st[0])
	}
	if st[0].OpenedAt == nil {
		t.Fatal("an open breaker must say when it opened")
	}

	clock.Advance(61 * time.Second)
	st = b.Snapshot()
	if st[0].State != BreakerHalfOpen {
		t.Fatalf("after the cooldown state = %q, want %q", st[0].State, BreakerHalfOpen)
	}
}
