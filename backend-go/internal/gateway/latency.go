package gateway

import (
	"sort"
	"sync"
)

// latencyWindowSize is how many recent call durations are kept per provider.
//
// WHY a window and not every sample: a percentile needs enough points to be
// meaningful and few enough that memory cannot grow with uptime. 200 recent
// calls is far more than one user generates in a minute and enough to see a
// provider slowing down.
const latencyWindowSize = 200

// LatencyStats is one provider's recent call timing, for the admin surface.
//
// WHY percentiles and not an average: one slow call in twenty moves an average
// and hides, while p95 is exactly the number a reader wants — "how long do the
// bad ones take". The average is deliberately absent for that reason.
type LatencyStats struct {
	Provider string  `json:"provider"`
	Calls    int64   `json:"calls_total"`
	Samples  int     `json:"samples_in_window"`
	P50Ms    float64 `json:"p50_ms"`
	P95Ms    float64 `json:"p95_ms"`
	MaxMs    float64 `json:"max_ms"`
	LastMs   float64 `json:"last_ms"`
}

// latencyWindow keeps a rolling set of durations for one provider.
type latencyWindow struct {
	mu      sync.Mutex
	samples []float64
	calls   int64
	last    float64
	max     float64
}

// record adds one duration in milliseconds.
func (w *latencyWindow) record(ms float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.calls++
	w.last = ms
	if ms > w.max {
		w.max = ms
	}
	if len(w.samples) < latencyWindowSize {
		w.samples = append(w.samples, ms)
		return
	}
	// Ring behaviour by eviction: the window is recent calls, not the first ones.
	copy(w.samples, w.samples[1:])
	w.samples[len(w.samples)-1] = ms
}

// stats reports the window. Samples are copied under the lock and sorted outside
// it so a slow sort never blocks a call from recording its duration.
func (w *latencyWindow) stats() (calls int64, samples int, p50, p95, maxMs, last float64) {
	w.mu.Lock()
	snapshot := make([]float64, len(w.samples))
	copy(snapshot, w.samples)
	calls = w.calls
	maxMs = w.max
	last = w.last
	w.mu.Unlock()

	samples = len(snapshot)
	if samples == 0 {
		return calls, 0, 0, 0, maxMs, last
	}
	sort.Float64s(snapshot)
	return calls, samples, percentileMillis(snapshot, 50), percentileMillis(snapshot, 95), maxMs, last
}

// percentileMillis returns the p-th percentile (p in 0..100) of a SORTED slice.
//
// Nearest-rank, which needs no interpolation: with a handful of samples the
// smallest observation that covers p% of them is the honest answer, whereas
// interpolating invents a duration that no call took.
//
// Pure — unit-tested.
func percentileMillis(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	rank := int((p / 100) * float64(len(sorted)))
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}

// recordLatency stores one call's duration for a provider.
func (g *ModelGateway) recordLatency(provider string, ms float64) {
	if g == nil || provider == "" {
		return
	}
	g.latencyMu.Lock()
	if g.latency == nil {
		g.latency = make(map[string]*latencyWindow)
	}
	w, ok := g.latency[provider]
	if !ok {
		w = &latencyWindow{}
		g.latency[provider] = w
	}
	g.latencyMu.Unlock()
	w.record(ms)
}

// LatencySnapshot reports a stable, sorted view of every provider that has been
// called in this process.
func (g *ModelGateway) LatencySnapshot() []LatencyStats {
	if g == nil {
		return nil
	}
	g.latencyMu.Lock()
	windows := make(map[string]*latencyWindow, len(g.latency))
	for name, w := range g.latency {
		windows[name] = w
	}
	g.latencyMu.Unlock()

	out := make([]LatencyStats, 0, len(windows))
	for name, w := range windows {
		calls, samples, p50, p95, maxMs, last := w.stats()
		out = append(out, LatencyStats{
			Provider: name,
			Calls:    calls,
			Samples:  samples,
			P50Ms:    p50,
			P95Ms:    p95,
			MaxMs:    maxMs,
			LastMs:   last,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Provider < out[j].Provider })
	return out
}
