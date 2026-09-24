// Package reliability owns C10: reliability-as-a-product.
//
// WHY a dedicated package:
//   SLO targets, the public /status surface and the audit-grade event log are
//   one cohesive concern ("what do we promise and did we keep it"). Keeping
//   the SLI math here (pure + tested) and the audit writes here means the
//   public and admin surfaces are views over the same computation, not two
//   divergent implementations.
//
// DESIGN (§3.1, C10): non-determinism is the job — the harness shifts the
//   probability mass. This package makes the promise explicit (published
//   targets + error budget), observable (live SLI snapshot + dependency
//   probes) and auditable (append-only slo_events). It does not invent
//   reliability: it reports what the existing observability counters and
//   dependency health checks already say.
//
// FAIL-SAFE: dependency probes are time-bounded; an audit-write failure is
//   logged, never surfaced (reliability reporting must not break serving).
package reliability

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"ai_avengers/backend/internal/observability"
)

// Verdicts for the availability SLI.
const (
	VerdictOK       = "ok"
	VerdictAtRisk   = "at_risk"
	VerdictBreached = "breached"
	VerdictUnknown  = "unknown" // no traffic yet — honest, not "ok"
)

// Event kinds (slo_events.kind).
const (
	KindDegraded     = "degraded"
	KindRecovered    = "recovered"
	KindBudgetBreach = "budget_breach"
)

// Severities (slo_events.severity).
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Policy is the configured reliability promise.
type Policy struct {
	// Enabled: expose the SLO surface + write audit events. Absent env → true.
	// Explicit false is the kill switch (status degrades to dependency-only,
	// like the pre-C10 /health).
	Enabled bool
	// AvailabilityTarget: e.g. 0.995 → allowed error rate 0.5%.
	AvailabilityTarget float64
	// ErrorBudgetWindowDays: the window the target is published for (default 30).
	// Informational for the surface (counters are process-lifetime).
	ErrorBudgetWindowDays int
	// AtRiskThreshold: error-budget fraction remaining below which the
	// verdict is "at_risk" (default 0.25).
	AtRiskThreshold float64
}

// DependencyStatus is one probed dependency's health.
type DependencyStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"` // healthy | unhealthy
	Error  string `json:"error,omitempty"`
}

// SLISnapshot is the availability service-level indicator + verdict.
type SLISnapshot struct {
	LLMCallsTotal        int64   `json:"llm_calls_total"`
	LLMErrorsTotal       int64   `json:"llm_errors_total"`
	Availability         float64 `json:"availability"`
	ErrorRate            float64 `json:"error_rate"`
	AvailabilityTarget   float64 `json:"availability_target"`
	ErrorBudgetRemaining float64 `json:"error_budget_remaining"`
	Status               string  `json:"status"`
	WindowDays           int     `json:"error_budget_window_days"`
	UptimeSeconds        int64   `json:"uptime_seconds"`
}

// StatusReport is the composed reliability surface.
type StatusReport struct {
	Status      string             `json:"status"` // ok | degraded | unknown
	Version     string             `json:"version"`
	Components  []DependencyStatus `json:"components"`
	SLO         *SLISnapshot       `json:"slo,omitempty"`
	GeneratedAt string             `json:"generated_at"`
}

// Event is one append-only reliability audit row.
type Event struct {
	ID        uuid.UUID       `json:"id"`
	Kind      string          `json:"kind"`
	Severity  string          `json:"severity"`
	Component string          `json:"component,omitempty"`
	Detail    json.RawMessage `json:"detail"`
	CreatedAt time.Time       `json:"created_at"`
}

// probe is one registered dependency health check.
type probe struct {
	name string
	fn   func(ctx context.Context) error
}

// Service computes the reliability surface + records audit events.
type Service struct {
	db     *pgxpool.Pool
	policy Policy
	probes []probe
	// mu guards last: the previous probe status per dependency, so a
	// healthy→unhealthy (or unhealthy→healthy) TRANSITION writes exactly one
	// audit event instead of one per probe call.
	mu   sync.Mutex
	last map[string]string
	log  *zap.Logger
}

// NewService builds the reliability service. db may be nil (surface only).
func NewService(db *pgxpool.Pool, policy Policy, logger *zap.Logger) *Service {
	if policy.AvailabilityTarget <= 0 || policy.AvailabilityTarget >= 1 {
		policy.AvailabilityTarget = 0.995
	}
	if policy.ErrorBudgetWindowDays <= 0 {
		policy.ErrorBudgetWindowDays = 30
	}
	if policy.AtRiskThreshold <= 0 || policy.AtRiskThreshold >= 1 {
		policy.AtRiskThreshold = 0.25
	}
	return &Service{
		db: db, policy: policy, probes: []probe{},
		last: map[string]string{}, log: logger,
	}
}

// Enabled reports whether the SLO surface is on. Nil-safe.
func (s *Service) Enabled() bool { return s != nil && s.policy.Enabled }

// Policy exposes the resolved policy (for the surface to publish targets).
func (s *Service) Policy() Policy { return s.policy }

// RegisterProbe adds a dependency health check, preserving registration order.
// Nil receiver / nil fn are ignored (safe wiring).
func (s *Service) RegisterProbe(name string, fn func(ctx context.Context) error) {
	if s == nil || fn == nil || name == "" {
		return
	}
	s.probes = append(s.probes, probe{name: name, fn: fn})
}

// Snapshot computes the availability SLI from the process-wide counters.
// Pure-ish (reads observability.Global); the math is in the pure helpers.
func (s *Service) Snapshot() SLISnapshot {
	calls := observability.Global.LLMCallsTotal()
	errs := observability.Global.LLMErrorsTotal()
	target := s.policy.AvailabilityTarget
	remaining := ErrorBudgetRemaining(ErrorRate(calls, errs), target)
	return SLISnapshot{
		LLMCallsTotal:        calls,
		LLMErrorsTotal:       errs,
		Availability:         Availability(calls, errs),
		ErrorRate:            ErrorRate(calls, errs),
		AvailabilityTarget:   target,
		ErrorBudgetRemaining: remaining,
		Status:               Verdict(calls, remaining, s.policy.AtRiskThreshold),
		WindowDays:           s.policy.ErrorBudgetWindowDays,
		UptimeSeconds:        observability.Global.UptimeSeconds(),
	}
}

// Components runs every registered probe (3s bound each) and audits status
// transitions. Probe order is registration order (deterministic output).
func (s *Service) Components(ctx context.Context) []DependencyStatus {
	if s == nil {
		return []DependencyStatus{}
	}
	out := make([]DependencyStatus, 0, len(s.probes))
	for _, p := range s.probes {
		pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := p.fn(pctx)
		cancel()
		st := DependencyStatus{Name: p.name, Status: "healthy"}
		if err != nil {
			st.Status = "unhealthy"
			st.Error = err.Error()
		}
		out = append(out, st)
		s.auditTransition(ctx, p.name, st.Status, err)
	}
	return out
}

// Status composes the overall report: dependency components + SLO snapshot.
// Overall = degraded if any dependency is unhealthy or the budget is breached;
// at_risk if the SLI is at risk; ok otherwise.
func (s *Service) Status(ctx context.Context, version string) StatusReport {
	components := s.Components(ctx)
	report := StatusReport{
		Version:     version,
		Components:  components,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Status:      "ok",
	}

	unhealthy := false
	for _, c := range components {
		if c.Status != "healthy" {
			unhealthy = true
			break
		}
	}

	if s.Enabled() {
		snap := s.Snapshot()
		report.SLO = &snap
		switch {
		case snap.Status == VerdictBreached:
			report.Status = "degraded"
			s.auditBudgetBreach(ctx, snap)
		case unhealthy:
			report.Status = "degraded"
		case snap.Status == VerdictAtRisk:
			report.Status = VerdictAtRisk
		}
		return report
	}

	if unhealthy {
		report.Status = "degraded"
	}
	return report
}

// Record appends one audit event (best-effort; a failure is logged only).
func (s *Service) Record(ctx context.Context, kind, severity, component string, detail map[string]interface{}) {
	if s == nil || s.db == nil {
		return
	}
	if severity != SeverityInfo && severity != SeverityWarning && severity != SeverityCritical {
		severity = SeverityInfo
	}
	payload, err := json.Marshal(detail)
	if err != nil {
		payload = []byte("{}")
	}
	var comp interface{}
	if component != "" {
		comp = component
	}
	if _, err := s.db.Exec(ctx,
		`INSERT INTO slo_events (kind, severity, component, detail)
		 VALUES ($1,$2,$3,$4::jsonb)`,
		kind, severity, comp, string(payload),
	); err != nil && s.log != nil {
		s.log.Warn("slo audit write failed", zap.String("kind", kind), zap.Error(err))
	}
}

// ListEvents returns the audit trail, newest first. limit clamps to [1,500].
func (s *Service) ListEvents(ctx context.Context, kind string, limit int) ([]Event, error) {
	if s == nil || s.db == nil {
		return []Event{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := `SELECT id, kind, severity, COALESCE(component,''), detail, created_at FROM slo_events`
	args := []interface{}{}
	if kind != "" {
		q += ` WHERE kind = $1`
		args = append(args, kind)
	}
	q += fmt.Sprintf(` ORDER BY created_at DESC LIMIT %d`, limit)

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list slo events: %w", err)
	}
	defer rows.Close()

	out := []Event{}
	for rows.Next() {
		var ev Event
		var detail []byte
		if err := rows.Scan(&ev.ID, &ev.Kind, &ev.Severity, &ev.Component, &detail, &ev.CreatedAt); err != nil {
			return nil, err
		}
		ev.Detail = json.RawMessage(detail)
		out = append(out, ev)
	}
	return out, rows.Err()
}

// ============================================================
// internals
// ============================================================

// auditTransition records exactly one event per status change. First-time
// unhealthy is a "degraded" event; unhealthy→healthy is "recovered".
func (s *Service) auditTransition(ctx context.Context, name, status string, probeErr error) {
	s.mu.Lock()
	prev := s.last[name]
	s.last[name] = status
	s.mu.Unlock()

	kind, ok := TransitionEvent(prev, status)
	if !ok {
		return
	}
	severity := SeverityWarning
	detail := map[string]interface{}{"previous": prev, "current": status}
	if status == "unhealthy" {
		severity = SeverityCritical
		if probeErr != nil {
			detail["error"] = probeErr.Error()
		}
	}
	s.Record(ctx, kind, severity, name, detail)
}

// auditBudgetBreach writes at most one budget_breach event per process hour to
// avoid hammering the log while the budget stays exhausted.
var lastBreachMu sync.Mutex
var lastBreachAt time.Time

func (s *Service) auditBudgetBreach(ctx context.Context, snap SLISnapshot) {
	lastBreachMu.Lock()
	if time.Since(lastBreachAt) < time.Hour {
		lastBreachMu.Unlock()
		return
	}
	lastBreachAt = time.Now()
	lastBreachMu.Unlock()

	s.Record(ctx, KindBudgetBreach, SeverityCritical, "", map[string]interface{}{
		"availability":           snap.Availability,
		"availability_target":    snap.AvailabilityTarget,
		"error_rate":             snap.ErrorRate,
		"error_budget_remaining": snap.ErrorBudgetRemaining,
		"llm_calls_total":        snap.LLMCallsTotal,
		"llm_errors_total":       snap.LLMErrorsTotal,
	})
}

// ============================================================
// pure SLI math (unit-tested)
// ============================================================

// Availability returns success/total in [0,1]; 1.0 when there is no traffic
// (an empty denominator is "nothing failed", not a division by zero).
func Availability(calls, errors int64) float64 {
	if calls <= 0 {
		return 1.0
	}
	if errors < 0 {
		errors = 0
	}
	if errors > calls {
		errors = calls
	}
	return float64(calls-errors) / float64(calls)
}

// ErrorRate returns errors/total in [0,1]; 0 when there is no traffic.
func ErrorRate(calls, errors int64) float64 {
	if calls <= 0 {
		return 0
	}
	if errors < 0 {
		errors = 0
	}
	if errors > calls {
		errors = calls
	}
	return float64(errors) / float64(calls)
}

// ErrorBudgetRemaining returns the fraction of the allowed error budget still
// unspent (1.0 = none spent, 0 = exactly exhausted, <0 = overspent). The
// allowed error rate is (1 - target). A degenerate target (>=1) means the
// budget is infinite → 1.0.
func ErrorBudgetRemaining(errorRate, target float64) float64 {
	allowed := 1.0 - target
	if allowed <= 0 {
		return 1.0
	}
	return 1.0 - (errorRate / allowed)
}

// Verdict maps traffic + remaining budget to a status. Pure: no traffic →
// unknown (never a false "ok").
func Verdict(calls int64, budgetRemaining, atRiskThreshold float64) string {
	if calls <= 0 {
		return VerdictUnknown
	}
	if budgetRemaining <= 0 {
		return VerdictBreached
	}
	if budgetRemaining < atRiskThreshold {
		return VerdictAtRisk
	}
	return VerdictOK
}

// TransitionEvent is pure: it decides whether a probe status change deserves
// an audit event, and which one. "" previous means first observation.
func TransitionEvent(prev, current string) (kind string, ok bool) {
	if prev == current {
		return "", false
	}
	if current == "unhealthy" {
		return KindDegraded, true
	}
	if current == "healthy" && prev == "unhealthy" {
		return KindRecovered, true
	}
	return "", false
}
