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
	llmCallsTotal        atomic.Int64
	llmErrorsTotal       atomic.Int64
	llmCostUSDTotal      atomic.Value // float64
	workflowsStarted     atomic.Int64
	workflowsFailed      atomic.Int64
	agentLoopIter        atomic.Int64
	outboxPublished      atomic.Int64
	entitlementDenied    atomic.Int64
	tenantDenied         atomic.Int64
	vacuumPickCalls      atomic.Int64
	vacuumJobsPicked     atomic.Int64
	vacuumJobsDone       atomic.Int64
	vacuumJobsFailed     atomic.Int64
	vacuumChunksVerified atomic.Int64
	llmCallsVacuum       atomic.Int64
	preservationFail     atomic.Int64
	dagStageMs           atomic.Int64
	kachraSuggested      atomic.Int64
	kachraVerified       atomic.Int64
	kachraMapped         atomic.Int64
	vacuumAutoPromoted   atomic.Int64
	vacuumFillerFiltered atomic.Int64
	vacuumCacheHit       atomic.Int64
	vacuumCacheMiss      atomic.Int64
	vacuumDriftFail      atomic.Int64
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
func (m *Metrics) IncTenantDenied()       { m.tenantDenied.Add(1) }
func (m *Metrics) IncVacuumPick()         { m.vacuumPickCalls.Add(1) }
func (m *Metrics) IncVacuumPicked(n int64) { m.vacuumJobsPicked.Add(n) }
func (m *Metrics) IncVacuumDone()         { m.vacuumJobsDone.Add(1) }
func (m *Metrics) IncVacuumFailed()       { m.vacuumJobsFailed.Add(1) }
func (m *Metrics) AddVacuumChunksReal(n int64) { m.vacuumChunksVerified.Add(n) }
func (m *Metrics) IncLLMVacuum()        { m.llmCallsVacuum.Add(1) }
func (m *Metrics) IncPreservationFail() { m.preservationFail.Add(1) }
func (m *Metrics) AddDAGStageMs(ms int64) { m.dagStageMs.Add(ms) }
func (m *Metrics) IncKachraSuggested(n int64) { m.kachraSuggested.Add(n) }
func (m *Metrics) IncKachraVerified(n int64)  { m.kachraVerified.Add(n) }
func (m *Metrics) IncKachraMapped(n int64)    { m.kachraMapped.Add(n) }
func (m *Metrics) KachraSuggested() int64 { return m.kachraSuggested.Load() }
func (m *Metrics) KachraVerified() int64  { return m.kachraVerified.Load() }
func (m *Metrics) KachraMapped() int64    { return m.kachraMapped.Load() }
func (m *Metrics) IncAutoPromoted(n int64)    { m.vacuumAutoPromoted.Add(n) }
func (m *Metrics) IncFillerFiltered(n int64)  { m.vacuumFillerFiltered.Add(n) }
func (m *Metrics) IncCacheHit()               { m.vacuumCacheHit.Add(1) }
func (m *Metrics) IncCacheMiss()              { m.vacuumCacheMiss.Add(1) }
func (m *Metrics) IncDriftFail()              { m.vacuumDriftFail.Add(1) }
func (m *Metrics) AutoPromoted() int64    { return m.vacuumAutoPromoted.Load() }
func (m *Metrics) FillerFiltered() int64  { return m.vacuumFillerFiltered.Load() }
func (m *Metrics) CacheHit() int64        { return m.vacuumCacheHit.Load() }
func (m *Metrics) CacheMiss() int64       { return m.vacuumCacheMiss.Load() }

// Typed getters (C10): the reliability surface reads these directly instead
// of type-asserting out of Snapshot()'s map[string]interface{}.

// LLMCallsTotal returns the process-wide LLM call counter.
func (m *Metrics) LLMCallsTotal() int64 { return m.llmCallsTotal.Load() }

// LLMErrorsTotal returns the process-wide LLM error counter.
func (m *Metrics) LLMErrorsTotal() int64 { return m.llmErrorsTotal.Load() }

// UptimeSeconds returns seconds since process start.
func (m *Metrics) UptimeSeconds() int64 { return int64(time.Since(startTime).Seconds()) }
func (m *Metrics) VacuumJobsDone() int64       { return m.vacuumJobsDone.Load() }
func (m *Metrics) VacuumChunksVerified() int64 { return m.vacuumChunksVerified.Load() }
func (m *Metrics) VacuumPickCalls() int64      { return m.vacuumPickCalls.Load() }
func (m *Metrics) LLMVacuumCalls() int64       { return m.llmCallsVacuum.Load() }
func (m *Metrics) PreservationFails() int64    { return m.preservationFail.Load() }

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

func (m *Metrics) AddVacuumChunks(n int64) { m.vacuumChunksVerified.Add(n) }

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
		"entitlement_denied":     m.entitlementDenied.Load(),
		"tenant_denied":          m.tenantDenied.Load(),
		"vacuum_pick_calls":      m.vacuumPickCalls.Load(),
		"vacuum_jobs_picked":     m.vacuumJobsPicked.Load(),
		"vacuum_jobs_done":       m.vacuumJobsDone.Load(),
		"vacuum_jobs_failed":     m.vacuumJobsFailed.Load(),
		"vacuum_chunks_verified":   m.vacuumChunksVerified.Load(),
		"vacuum_llm_calls":         m.llmCallsVacuum.Load(),
		"vacuum_preservation_fail": m.preservationFail.Load(),
		"vacuum_dag_stage_ms":      m.dagStageMs.Load(),
		"vacuum_kachra_suggested": m.kachraSuggested.Load(),
		"vacuum_kachra_verified":  m.kachraVerified.Load(),
		"vacuum_kachra_mapped":    m.kachraMapped.Load(),
		"vacuum_auto_promoted":    m.vacuumAutoPromoted.Load(),
		"vacuum_filler_filtered":  m.vacuumFillerFiltered.Load(),
		"vacuum_cache_hit":        m.vacuumCacheHit.Load(),
		"vacuum_cache_miss":       m.vacuumCacheMiss.Load(),
		"vacuum_drift_fail":       m.vacuumDriftFail.Load(),
		"uptime_seconds":          int64(time.Since(startTime).Seconds()),
	}
}
