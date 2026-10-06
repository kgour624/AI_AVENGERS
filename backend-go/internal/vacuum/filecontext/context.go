package filecontext

import (
	"sync"
)

// Interval represents a cleaned byte range [Start,End).
type Interval struct{ Start, End int }

// IntervalTree is simple sorted interval set with merge + overlap check.
// Prevents infinite loop on 4x kachra/line by tracking visited ranges.
type IntervalTree struct {
	mu        sync.Mutex
	intervals []Interval
}

func (t *IntervalTree) Insert(s, e int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.intervals = append(t.intervals, Interval{s, e})
}

func (t *IntervalTree) Overlaps(s, e int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, iv := range t.intervals {
		if s < iv.End && e > iv.Start {
			return true
		}
	}
	return false
}

func (t *IntervalTree) All() []Interval {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make([]Interval, len(t.intervals))
	copy(cp, t.intervals)
	return cp
}

// FileContext holds per-file concurrent state.
type FileContext struct {
	Visited       sync.Map // key string "start:end" -> struct{}
	CheckpointSet sync.Map // chunkIndex int -> struct{} for contiguous checkpoint
	Intervals     IntervalTree
	mu            sync.Mutex
	phase         int
}

func New() *FileContext { return &FileContext{} }

func (f *FileContext) SetPhase(p int) { f.mu.Lock(); f.phase = p; f.mu.Unlock() }
func (f *FileContext) Phase() int     { f.mu.Lock(); defer f.mu.Unlock(); return f.phase }

// TryVisit returns true if this range was not visited before (and marks it).
func (f *FileContext) TryVisit(start, end int) bool {
	k := makeKey(start, end)
	_, loaded := f.Visited.LoadOrStore(k, struct{}{})
	return !loaded
}

func makeKey(a, b int) string { return itoa(a) + ":" + itoa(b) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
