package vacuum

import (
	"context"
	"sync"
	"time"
)

// ProgressEvent is one stage transition or incremental counter for a file_jobs
// job. It is the single shape streamed to the admin UI over SSE.
//
// WHY a numeric phase + a string stage: the DB already stores phase (0..6) and
// the UI renders "phase/6"; keeping phase lets a refreshing client map straight
// onto the same numbers, while the string stage is what the SSE reducer keys
// terminal states off (done/failed/canceled/quarantined) without magic ints.
type ProgressEvent struct {
	JobID   string `json:"job_id"`
	Phase   int    `json:"phase"`
	Stage   string `json:"stage"`
	Status  string `json:"status"`
	Percent int    `json:"percent"`
	Message string `json:"message,omitempty"`
	Done    int    `json:"done"`
	Total   int    `json:"total"`
	At      int64  `json:"at"`
}

// progressHub is a tiny in-process pub/sub keyed by job id.
//
// WHY in-process and not Redis: a vacuum job runs inside exactly one server
// process (the goroutine started by ExecuteJob), and its SSE subscribers are
// admin browsers talking to that same process. Cross-process fan-out would add
// a broker for no benefit here; if the backend is ever scaled horizontally the
// hub is the single place that would need to move.
type progressHub struct {
	mu   sync.RWMutex
	subs map[string]map[chan ProgressEvent]struct{}
	last map[string]ProgressEvent
}

func newProgressHub() *progressHub {
	return &progressHub{
		subs: make(map[string]map[chan ProgressEvent]struct{}),
		last: make(map[string]ProgressEvent),
	}
}

var progressBus = newProgressHub()

// Publish records the latest event for a job and fans it out to every live
// subscriber. It is non-blocking: a subscriber whose buffer is full is skipped
// rather than stalling the pipeline goroutine that produced the event.
func (h *progressHub) Publish(ev ProgressEvent) {
	if ev.At == 0 {
		ev.At = time.Now().UnixMilli()
	}
	h.mu.Lock()
	h.last[ev.JobID] = ev
	for ch := range h.subs[ev.JobID] {
		select {
		case ch <- ev:
		default:
		}
	}
	h.mu.Unlock()
}

// Subscribe registers a buffered channel for one job and returns an unsubscribe
// func. The latest known snapshot (if any) is delivered first so a client that
// refreshes mid-run re-attaches to the current phase instead of a blank screen.
func (h *progressHub) Subscribe(jobID string) (<-chan ProgressEvent, func()) {
	ch := make(chan ProgressEvent, 64)
	h.mu.Lock()
	if h.subs[jobID] == nil {
		h.subs[jobID] = make(map[chan ProgressEvent]struct{})
	}
	h.subs[jobID][ch] = struct{}{}
	last, ok := h.last[jobID]
	h.mu.Unlock()
	if ok {
		select {
		case ch <- last:
		default:
		}
	}
	var once sync.Once
	unsub := func() {
		once.Do(func() {
			h.mu.Lock()
			if m := h.subs[jobID]; m != nil {
				delete(m, ch)
				if len(m) == 0 {
					delete(h.subs, jobID)
				}
			}
			// close under the lock: Publish also sends under the lock, so a
			// send can never race a close and panic.
			close(ch)
			h.mu.Unlock()
		})
	}
	return ch, unsub
}

// Forget drops the retained snapshot for a finished job so a completed job's
// last event is not replayed forever and the map cannot grow unbounded.
func (h *progressHub) Forget(jobID string) {
	h.mu.Lock()
	delete(h.last, jobID)
	h.mu.Unlock()
}

// Last returns the most recent event published for a job, when one is still
// retained. Cancel uses it to keep the progress bar where the run actually
// stopped instead of rewinding it to 0%.
func (h *progressHub) Last(jobID string) (ProgressEvent, bool) {
	h.mu.RLock()
	ev, ok := h.last[jobID]
	h.mu.RUnlock()
	return ev, ok
}

// publishProgress is the thin helper the pipeline calls at each stage.
func publishProgress(jobID, stage, status, msg string, phase, percent, done, total int) {
	progressBus.Publish(ProgressEvent{
		JobID:   jobID,
		Phase:   phase,
		Stage:   stage,
		Status:  status,
		Percent: percent,
		Message: msg,
		Done:    done,
		Total:   total,
	})
}

// terminalStage reports whether an event ends a job's lifecycle (SSE closes).
func terminalStage(stage string) bool {
	switch stage {
	case "done", "failed", "canceled", "quarantined":
		return true
	}
	return false
}

// phasePercent maps a DB phase to a coarse progress percentage, so a refreshing
// (or freshly attached) client shows a sane bar before any counter event lands.
func phasePercent(phase int) int {
	switch {
	case phase <= 0:
		return 0
	case phase == 1:
		return 5
	case phase == 2:
		return 20
	case phase == 3:
		return 45
	case phase == 4:
		return 70
	case phase == 5:
		return 85
	default:
		return 100
	}
}

// runMu guards runCancels, the registry of in-flight jobs' cancel funcs.
var (
	runMu      sync.Mutex
	runCancels = map[string]context.CancelFunc{}
)

// registerRun stores the cancel func for a job so a cancel request can reach
// the goroutine executing it, even though that goroutine runs detached from the
// HTTP request that started it.
func registerRun(jobID string, cancel context.CancelFunc) {
	runMu.Lock()
	runCancels[jobID] = cancel
	runMu.Unlock()
}

// unregisterRun removes a finished job's cancel func.
func unregisterRun(jobID string) {
	runMu.Lock()
	delete(runCancels, jobID)
	runMu.Unlock()
}

// cancelRun cancels a running job. Returns true when a live run was found.
func cancelRun(jobID string) bool {
	runMu.Lock()
	cancel, ok := runCancels[jobID]
	if ok {
		delete(runCancels, jobID)
	}
	runMu.Unlock()
	if ok && cancel != nil {
		cancel()
	}
	return ok
}

// isRunning reports whether this process still owns a live run for a job.
//
// WHY the orphan reaper needs it: a long phase-2 scan can legitimately leave a
// row without a single DB write for a long time, so "nothing written recently"
// alone is not proof of death. A registered cancel func is that proof of life.
func isRunning(jobID string) bool {
	runMu.Lock()
	_, ok := runCancels[jobID]
	runMu.Unlock()
	return ok
}

// cancelSnapshot picks the numbers published with the terminal cancel event.
//
// WHY not simply 0: cancel is the last event a client ever sees, and the old
// hardcoded 0 rewound the bar to 0% the instant the admin pressed Cancel, which
// reads as "restarting" rather than "stopped". Precedence is: the last live
// event, then the coarse percentage of the DB phase, and only a job that never
// left phase 0 reports a completed 100%.
func cancelSnapshot(jobID string, dbPhase int) (phase, percent, done, total int) {
	phase = dbPhase
	percent = phasePercent(dbPhase)
	if last, ok := progressBus.Last(jobID); ok {
		if last.Phase > 0 {
			phase = last.Phase
		}
		if last.Percent > 0 {
			percent = last.Percent
		}
		done, total = last.Done, last.Total
	}
	if percent <= 0 {
		percent = 100
	}
	return phase, percent, done, total
}
