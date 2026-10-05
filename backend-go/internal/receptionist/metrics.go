package receptionist

import "sync"

// In-process, dependency-free metrics.
// WHY: prometheus/client_golang is NOT a dependency of this module (see go.mod),
// so the previous prometheus-based metrics.go could not compile. This file keeps
// every existing call site (sessionsTotal / expertLatency / finalCost /
// pressureGauge) compiling and still records values, without a new dependency.
type Metric struct {
mu     sync.Mutex
name   string
labels []string
count  float64
sum    float64
last   float64
series map[string]float64
}

func newMetric(name string, labels ...string) *Metric {
return &Metric{name: name, labels: labels, series: make(map[string]float64)}
}

// seriesKey - label values ko stable key me badalta hai
func seriesKey(vals []string) string {
out := ""
for i, v := range vals {
if i > 0 {
out += "|"
}
out += v
}
return out
}

// WithLabelValues - label series register karta hai (no-op routing)
func (m *Metric) WithLabelValues(vals ...string) *Metric {
m.mu.Lock()
defer m.mu.Unlock()
k := seriesKey(vals)
if _, ok := m.series[k]; !ok {
m.series[k] = 0
}
return m
}

func (m *Metric) Add(v float64) {
m.mu.Lock()
defer m.mu.Unlock()
m.count++
m.sum += v
m.last = v
}

func (m *Metric) Inc() { m.Add(1) }

func (m *Metric) Observe(v float64) { m.Add(v) }

func (m *Metric) Set(v float64) {
m.mu.Lock()
defer m.mu.Unlock()
m.last = v
}

func (m *Metric) Value() float64 {
m.mu.Lock()
defer m.mu.Unlock()
return m.last
}

func (m *Metric) Name() string { return m.name }

func (m *Metric) Count() float64 {
m.mu.Lock()
defer m.mu.Unlock()
return m.count
}

func (m *Metric) Sum() float64 {
m.mu.Lock()
defer m.mu.Unlock()
return m.sum
}

// Snapshot - label series ka copy (debug/health endpoint ke liye)
func (m *Metric) Snapshot() map[string]float64 {
m.mu.Lock()
defer m.mu.Unlock()
out := make(map[string]float64, len(m.series))
for k, v := range m.series {
out[k] = v
}
return out
}

var (
sessionsTotal = newMetric("receptionist_sessions_total", "phase")
expertLatency = newMetric("receptionist_expert_latency_ms", "expert")
finalCost     = newMetric("receptionist_final_cost_usd")
pressureGauge = newMetric("receptionist_pressure_score", "session_id")
)