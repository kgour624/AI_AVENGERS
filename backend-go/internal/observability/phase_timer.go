// Package observability provides phase-level timing instrumentation
// for the AI Avengers request pipeline.
//
// DESIGN PATTERN: Observer (GoF)
//   PhaseTimer is a passive observer that records timing data as
//   pipeline phases start and stop. Callers do not need to know how
//   timing is stored or reported — they just call Start/Stop.
//
// SOLID:
//   SRP: PhaseTimer has exactly one responsibility — measure phase durations.
//   OCP: New phases are added by calling Start/Stop with a new name.
//        No modification to PhaseTimer itself required.
//   ISP: Callers only see the interface they need (Recorder).
package observability

import (
	"sync"
	"time"
)

// Recorder is the interface callers use to record phase timings.
// Keeping it minimal (ISP) so callers don't depend on reporting methods.
type Recorder interface {
	Start(phase string)
	Stop(phase string)
}

// PhaseResult holds the measured duration of one pipeline phase.
type PhaseResult struct {
	Phase    string
	Duration time.Duration
}

// PhaseTimer implements Recorder.
// Thread-safe: multiple goroutines can record phases concurrently.
//
// WHY sync.Mutex not sync.RWMutex:
//   Both Start and Stop write — no read-heavy path here.
//   Mutex is simpler and correct.
type PhaseTimer struct {
	mu      sync.Mutex
	starts  map[string]time.Time
	results []PhaseResult
}

// NewPhaseTimer creates a ready-to-use PhaseTimer.
func NewPhaseTimer() *PhaseTimer {
	return &PhaseTimer{
		starts:  make(map[string]time.Time),
		results: make([]PhaseResult, 0, 8),
	}
}

// Start records the start time of a named phase.
// Calling Start twice for the same phase overwrites the previous start.
func (t *PhaseTimer) Start(phase string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.starts[phase] = time.Now()
}

// Stop records the duration of a named phase.
// If Start was never called for this phase, Stop is a no-op.
func (t *PhaseTimer) Stop(phase string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	start, ok := t.starts[phase]
	if !ok {
		return
	}
	t.results = append(t.results, PhaseResult{
		Phase:    phase,
		Duration: time.Since(start),
	})
	delete(t.starts, phase)
}

// Results returns a snapshot of all completed phase timings.
// Order matches the order Stop was called.
func (t *PhaseTimer) Results() []PhaseResult {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]PhaseResult, len(t.results))
	copy(out, t.results)
	return out
}

// TotalMs returns the sum of all recorded phase durations in milliseconds.
func (t *PhaseTimer) TotalMs() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	var total time.Duration
	for _, r := range t.results {
		total += r.Duration
	}
	return total.Milliseconds()
}

// NoopRecorder is a Recorder that does nothing.
// Used when observability is disabled (nil-safe pattern).
//
// WHY: Callers can always call recorder.Start/Stop without nil checks.
// Follows the Null Object pattern (GoF).
type NoopRecorder struct{}

func (n *NoopRecorder) Start(_ string) {}
func (n *NoopRecorder) Stop(_ string)  {}
