package gateway

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BreakerConfig controls when a provider is taken out of rotation.
//
// WHY this exists at all (P8: trust no single model or provider): the gateway
// already fell back to a second provider, but only AFTER three failed attempts
// with 1s and 2s backoff. A provider that is down for an hour therefore cost
// every single request ~3 seconds of waiting and three more hits on a service
// that was already failing — and nothing anywhere said "this provider is
// broken". A breaker turns a repeated, known failure into an instant skip for a
// bounded time, and it is the only part of the fallback chain that remembers.
type BreakerConfig struct {
	// Enabled: the kill switch. false restores the old behaviour exactly
	// (always attempt the primary, never skip).
	Enabled bool
	// FailureThreshold is how many consecutive failures open the breaker.
	// Consecutive, not a ratio: a single successful call clears the count, so an
	// occasionally flaky provider is never taken out of rotation.
	FailureThreshold int
	// Cooldown is how long the provider is skipped before one probe is allowed
	// through. Long enough that a provider coming back has time to come back;
	// short enough that a transient outage does not hide a healthy provider.
	Cooldown time.Duration
}

// DefaultBreakerConfig is the behaviour when nothing is configured.
//
// Enabled by default is deliberate: the failure it protects against (a dead
// provider draining every request of three wasted attempts) is worse than the
// failure it could cause (a provider skipped for 60s after three consecutive
// failures, with the reason logged and the fallback — when configured — used
// instead).
func DefaultBreakerConfig() BreakerConfig {
	return BreakerConfig{
		Enabled:          true,
		FailureThreshold: 3,
		Cooldown:         60 * time.Second,
	}
}

// Breaker state names, as reported to the admin surface.
const (
	BreakerClosed   = "closed"
	BreakerOpen     = "open"
	BreakerHalfOpen = "half_open"
)

// BreakerStatus is one provider's breaker state for the status endpoint.
type BreakerStatus struct {
	Provider string `json:"provider"`
	// State: closed (normal), open (skipped), half_open (one probe allowed).
	State               string     `json:"state"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	OpenedAt            *time.Time `json:"opened_at,omitempty"`
	// RetryAfterSeconds is how long until the next probe is allowed. 0 unless open.
	RetryAfterSeconds int    `json:"retry_after_seconds,omitempty"`
	LastError         string `json:"last_error,omitempty"`
}

// breakerEntry is the mutable state of one provider's breaker.
type breakerEntry struct {
	failures      int
	openedAt      time.Time
	probeInFlight bool
	lastError     string
}

// Breaker decides whether a provider may be attempted right now.
//
// It is a pure state machine apart from its clock, which is injectable so the
// open → cooldown → half-open → closed cycle can be unit-tested without sleeping
// through a real minute.
type Breaker struct {
	cfg    BreakerConfig
	logger *zap.Logger
	now    func() time.Time

	mu      sync.Mutex
	entries map[string]*breakerEntry
}

// NewBreaker builds a breaker. A zero config is replaced by the defaults, so a
// caller that has no opinion still gets the protection rather than a disabled
// feature nobody notices is disabled.
func NewBreaker(cfg BreakerConfig, logger *zap.Logger) *Breaker {
	// An entirely zero config means "no opinion" and gets the defaults. A config
	// that sets a threshold or a cooldown is an opinion — including the opinion
	// that the breaker should be off — and is honoured as given. Without this
	// distinction a caller that passed nothing would silently get a DISABLED
	// breaker, which is the one outcome nobody would notice.
	if cfg.FailureThreshold <= 0 && cfg.Cooldown <= 0 && !cfg.Enabled {
		cfg = DefaultBreakerConfig()
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = DefaultBreakerConfig().FailureThreshold
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = DefaultBreakerConfig().Cooldown
	}
	return &Breaker{cfg: cfg, logger: logger, now: time.Now, entries: make(map[string]*breakerEntry)}
}

// Config returns the active configuration (read by the admin surface).
func (b *Breaker) Config() BreakerConfig {
	if b == nil {
		return DefaultBreakerConfig()
	}
	return b.cfg
}

// SetClock replaces the clock. Test-only, and named plainly so it cannot be
// mistaken for production wiring.
func (b *Breaker) SetClock(now func() time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.now = now
}

func (b *Breaker) entry(provider string) *breakerEntry {
	e, ok := b.entries[provider]
	if !ok {
		e = &breakerEntry{}
		b.entries[provider] = e
	}
	return e
}

// Allow reports whether a call to this provider may be attempted, with a reason
// when it may not.
//
// A disabled breaker always allows. When the cooldown has elapsed exactly one
// caller is let through (half-open) so recovery is discovered by a real call
// rather than assumed; concurrent callers are refused while that probe is in
// flight, which keeps a recovering provider from being hit by a burst.
func (b *Breaker) Allow(provider string) (bool, string) {
	if b == nil || !b.cfg.Enabled {
		return true, ""
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	// Asking must not create an entry. The status screen lists providers the
	// breaker has an opinion about, and a read that registered every provider
	// ever asked about would fill it with "closed, no failures" rows that say
	// nothing.
	e, ok := b.entries[provider]
	if !ok || e.openedAt.IsZero() {
		return true, ""
	}
	elapsed := b.now().Sub(e.openedAt)
	if elapsed < b.cfg.Cooldown {
		remaining := (b.cfg.Cooldown - elapsed).Round(time.Second)
		return false, fmt.Sprintf(
			"provider %s is open after %d consecutive failures; %s until the next probe",
			provider, e.failures, remaining,
		)
	}
	if e.probeInFlight {
		return false, fmt.Sprintf("provider %s: a recovery probe is already in flight", provider)
	}
	e.probeInFlight = true
	return true, ""
}

// RecordSuccess clears the breaker: one good call is proof enough.
func (b *Breaker) RecordSuccess(provider string) {
	if b == nil || !b.cfg.Enabled {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	e := b.entry(provider)
	if !e.openedAt.IsZero() {
		b.logger.Info("provider recovered",
			zap.String("provider", provider),
			zap.Int("consecutive_failures_before", e.failures),
		)
	}
	e.failures = 0
	e.openedAt = time.Time{}
	e.probeInFlight = false
	e.lastError = ""
}

// RecordFailure counts a failure and opens the breaker at the threshold.
//
// A failure while half-open restarts the cooldown: the probe proved the provider
// is still unwell, and treating that as a new outage is the honest reading.
func (b *Breaker) RecordFailure(provider string, cause error) {
	if b == nil || !b.cfg.Enabled {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	e := b.entry(provider)
	e.failures++
	e.probeInFlight = false
	if cause != nil {
		e.lastError = cause.Error()
	}

	wasOpen := !e.openedAt.IsZero()
	if !wasOpen && e.failures < b.cfg.FailureThreshold {
		return
	}
	if !wasOpen {
		b.logger.Warn("provider breaker opened",
			zap.String("provider", provider),
			zap.Int("consecutive_failures", e.failures),
			zap.Duration("cooldown", b.cfg.Cooldown),
			zap.String("last_error", e.lastError),
		)
	}
	e.openedAt = b.now()
}

// Snapshot reports every provider the breaker has an opinion about, sorted by
// provider name for a stable screen.
func (b *Breaker) Snapshot() []BreakerStatus {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]BreakerStatus, 0, len(b.entries))
	for name, e := range b.entries {
		st := BreakerStatus{
			Provider:            name,
			State:               BreakerClosed,
			ConsecutiveFailures: e.failures,
			LastError:           e.lastError,
		}
		if !e.openedAt.IsZero() {
			opened := e.openedAt
			st.OpenedAt = &opened
			elapsed := b.now().Sub(e.openedAt)
			if elapsed >= b.cfg.Cooldown {
				st.State = BreakerHalfOpen
			} else {
				st.State = BreakerOpen
				st.RetryAfterSeconds = int((b.cfg.Cooldown - elapsed).Round(time.Second).Seconds())
			}
		}
		out = append(out, st)
	}
	// Insertion into a map is random; a screen that reshuffles is unreadable.
	sortBreakerStatuses(out)
	return out
}

func sortBreakerStatuses(items []BreakerStatus) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j-1].Provider > items[j].Provider; j-- {
			items[j-1], items[j] = items[j], items[j-1]
		}
	}
}
