package runner

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Stage is a single DAG step. Inputs are produced by prior stages; stage reads from ctx-bound state.
type Stage struct {
	Name string
	Run  func(ctx context.Context, jobID uuid.UUID) error
}

// DAG is a 6-stage sequential executor with stage timing and errgroup semantics.
// Phase 6: up to 20-100 pods via caller's semaphore; DAG itself is sequential per job.
type DAG struct {
	stages []Stage
}

func New(stages ...Stage) *DAG { return &DAG{stages: stages} }

type StageTiming struct {
	Name       string `json:"name"`
	DurationMs int64  `json:"duration_ms"`
	Err        string `json:"err,omitempty"`
}

// Run executes stages sequentially. First failure short-circuits and returns error.
// Each stage is context-aware; cancellation aborts remaining stages.
func (d *DAG) Run(ctx context.Context, jobID uuid.UUID) ([]StageTiming, error) {
	timings := make([]StageTiming, 0, len(d.stages))
	for _, st := range d.stages {
		if ctx.Err() != nil {
			return timings, ctx.Err()
		}
		t0 := time.Now()
		err := st.Run(ctx, jobID)
		dur := time.Since(t0).Milliseconds()
		timing := StageTiming{Name: st.Name, DurationMs: dur}
		if err != nil {
			timing.Err = err.Error()
			timings = append(timings, timing)
			return timings, err
		}
		timings = append(timings, timing)
	}
	return timings, nil
}

// RunParallel runs Run for many jobIDs concurrently with concurrency cap.
// Mirrors Pipeline.PickAndExecute's sem(5) but up to 20-100 pods.
func RunParallel(ctx context.Context, dag *DAG, jobIDs []uuid.UUID, concurrency int) (map[uuid.UUID][]StageTiming, map[uuid.UUID]error) {
	if concurrency <= 0 {
		concurrency = 5
	}
	if concurrency > 100 {
		concurrency = 100
	}
	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	timings := map[uuid.UUID][]StageTiming{}
	errs := map[uuid.UUID]error{}
	var wg sync.WaitGroup
	for _, id := range jobIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func(jid uuid.UUID) {
			defer wg.Done()
			defer func() { <-sem }()
			t, err := dag.Run(ctx, jid)
			mu.Lock()
			timings[jid] = t
			if err != nil {
				errs[jid] = err
			}
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return timings, errs
}
