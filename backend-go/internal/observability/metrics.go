package observability

import (
	"sync/atomic"
	"time"
)

// Metrics is a simple in-memory counter store.
//
// WHY not Prometheus client library: avoids binary size for small deployments.
// /metrics returns JSON; switch Render later without caller changes.
//
// SOLID: OCP — new metrics = new field + method. SRP — count only.
// Pattern: Singleton (Global), Null Object (zero value is valid).
type Metrics struct {
	llmCallsTotal      atomic.Int64
	llmErrorsTotal     atomic.Int64
	llmCostUSDTotal    atomic.Value // float64
	workflowsStarted   atomic.Int64
	workflowsFailed    atomic.Int64
	agentLoopIter      atomic.Int64
	outboxPublished    atomic.Int64
	entitlementDenied  atomic.Int64
	tenantDenied       atomic.Int64
}

// Global is the process-wide metrics instance.
var Global = &Metrics{}

var startTime = time.Now()

func init() {
	Global.llmCostUSDTotal.Store(float64(0))
}

func (m *Metrics) IncLLMCall()          { m.llmCallsTotal.Add(1) }
func (m *Metrics) IncLLMError()         { m.llmErrorsTotal.Add(1) }
func (m *Metrics) IncWorkflowStarted()  { m.workflowsStarted.Add(1) }
func (m *Metrics) IncWorkflowFailed()   { m.workflowsFailed.Add(1) }
func (m *Metrics) IncAgentLoopIter()    { m.agentLoopIter.Add(1) }
func (m *Metrics) IncOutboxPublished()  { m.outboxPublished.Add(1) }
func (m *Metrics) IncEntitlementDenied() { m.entitlementDenied.Add(1) }
func (m *Metrics) IncTenantDenied()      { m.tenantDenied.Add(1) }

// Typed getters (C10): the reliability surface reads these directly instead
// of type-asserting out of Snapshot()'s map[string]interface{}.

// LLMCallsTotal returns the process-wide LLM call counter.
func (m *Metrics) LLMCallsTotal() int64 { return m.llmCallsTotal.Load() }

// LLMErrorsTotal returns the process-wide LLM error counter.
func (m *Metrics) LLMErrorsTotal() int64 { return m.llmErrorsTotal.Load() }

// UptimeSeconds returns seconds since process start.
func (m *Metrics) UptimeSeconds() int64 { return int64(time.Since(startTime).Seconds()) }

func (m *Metrics) AddLLMCost(usd float64) {
	for {
		old := m.llmCostUSDTotal.Load()
		var oldF float64
		if old != nil {
			oldF = old.(float64)
		}
		if m.llmCostUSDTotal.CompareAndSwap(old, oldF+usd) {
			return
		}
	}
}

// Snapshot returns current metric values for the /metrics handler.
func (m *Metrics) Snapshot() map[string]interface{} {
	cost := float64(0)
	if v := m.llmCostUSDTotal.Load(); v != nil {
		cost = v.(float64)
	}
	return map[string]interface{}{
		"llm_calls_total":       m.llmCallsTotal.Load(),
		"llm_errors_total":      m.llmErrorsTotal.Load(),
		"llm_cost_usd_total":    cost,
		"workflows_started":     m.workflowsStarted.Load(),
		"workflows_failed":      m.workflowsFailed.Load(),
		"agent_loop_iter":       m.agentLoopIter.Load(),
		"outbox_published":      m.outboxPublished.Load(),
		"entitlement_denied":    m.entitlementDenied.Load(),
		"tenant_denied":         m.tenantDenied.Load(),
		"uptime_seconds":        int64(time.Since(startTime).Seconds()),
	}
}
