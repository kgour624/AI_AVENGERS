package gateway

import "testing"

// Nearest-rank percentiles: with a handful of samples the smallest observation
// covering p% is the honest answer — interpolating would invent a duration that
// no call actually took.
func TestPercentileMillis(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	tests := []struct {
		p    float64
		want float64
	}{
		{p: 0, want: 10},
		{p: 50, want: 60},
		{p: 95, want: 100},
		{p: 100, want: 100},
	}
	for _, tc := range tests {
		if got := percentileMillis(sorted, tc.p); got != tc.want {
			t.Fatalf("percentileMillis(p=%v) = %v, want %v", tc.p, got, tc.want)
		}
	}
	if got := percentileMillis(nil, 95); got != 0 {
		t.Fatalf("an empty window must report 0, got %v", got)
	}
}

// One slow call must show up in the p95 rather than being averaged away — that
// is the whole reason percentiles are reported instead of a mean.
func TestLatencyWindowKeepsTheSlowTail(t *testing.T) {
	var w latencyWindow
	for i := 0; i < 19; i++ {
		w.record(100)
	}
	w.record(9000)

	calls, samples, p50, p95, maxMs, last := w.stats()
	if calls != 20 || samples != 20 {
		t.Fatalf("calls=%d samples=%d, want 20/20", calls, samples)
	}
	if p50 != 100 {
		t.Fatalf("p50 = %v, want 100 — the tail must not move the median", p50)
	}
	if p95 != 9000 {
		t.Fatalf("p95 = %v, want 9000 — the slow call must be visible", p95)
	}
	if maxMs != 9000 || last != 9000 {
		t.Fatalf("max=%v last=%v, want both 9000", maxMs, last)
	}
}

// The window is bounded: memory cannot grow with uptime, and the old samples it
// evicts are the ones that stop counting.
func TestLatencyWindowIsBounded(t *testing.T) {
	var w latencyWindow
	for i := 0; i < latencyWindowSize+50; i++ {
		w.record(10)
	}
	calls, samples, _, _, _, _ := w.stats()
	if calls != int64(latencyWindowSize+50) {
		t.Fatalf("calls = %d, want every call counted", calls)
	}
	if samples != latencyWindowSize {
		t.Fatalf("samples = %d, want the window capped at %d", samples, latencyWindowSize)
	}
}
