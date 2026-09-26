package eval

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ErrSkip lets an Answerer signal that a case cannot run in this
// environment (e.g. its expert_slug is not present). Skipped cases are
// excluded from the score, not counted as failures.
var ErrSkip = errors.New("eval: case skipped")

// Observed is what an Answerer produced for one case.
type Observed struct {
	Content   string
	Citations int
	Mode      string
	Refused   bool
	CostUSD   float64
	LatencyMs int64
}

// Answerer produces an observed answer for a case. The real implementation
// (HTTPAnswerer) calls the running system; tests inject a fake.
type Answerer interface {
	Answer(ctx context.Context, c Case) (Observed, error)
}

// CaseResult is the scored outcome for one case.
type CaseResult struct {
	CaseID    string  `json:"case_id"`
	Vital     bool    `json:"vital"`
	Passed    bool    `json:"passed"`
	Skipped   bool    `json:"skipped,omitempty"`
	Checks    []Check `json:"checks,omitempty"`
	Error     string  `json:"error,omitempty"`
	CostUSD   float64 `json:"cost_usd"`
	LatencyMs int64   `json:"latency_ms"`
}

// RunSummary is the aggregate of one golden-set run.
type RunSummary struct {
	ID          uuid.UUID    `json:"id"`
	Suite       string       `json:"suite"`
	Model       string       `json:"model"`
	Total       int          `json:"total"`
	Passed      int          `json:"passed"`
	VitalTotal  int          `json:"vital_total"`
	VitalFailed int          `json:"vital_failed"`
	Score       float64      `json:"score"`
	CostUSD     float64      `json:"cost_usd"`
	Results     []CaseResult `json:"results"`
	CreatedAt   time.Time    `json:"created_at"`
}

// HasVitalFailure reports whether any must-have case failed (merge blocker).
func (s *RunSummary) HasVitalFailure() bool {
	return s != nil && s.VitalFailed > 0
}

// Runner executes a golden set against an Answerer.
type Runner struct {
	model  string
	logger *zap.Logger
	// judge (T3) is OPTIONAL. Nil → deterministic scoring only (CI default).
	judge Judge
}

// NewRunner builds a runner. model is recorded on the run for delta context.
func NewRunner(model string, logger *zap.Logger) *Runner {
	return &Runner{model: model, logger: logger}
}

// WithJudge layers an LLM judge on top of the deterministic scorer and returns
// the runner for chaining. When set, each non-refusal case gains an extra
// "judge" component check; the deterministic checks and the vital-failure gate
// are unchanged. A judge error is logged and the component is omitted, so a
// model outage can never flip a case.
func (r *Runner) WithJudge(j Judge) *Runner {
	r.judge = j
	return r
}

// Run evaluates every case in the set. Cases the Answerer skips are
// excluded from the totals (a missing expert is an environment gap, not
// a regression).
func (r *Runner) Run(ctx context.Context, set *GoldenSet, a Answerer) *RunSummary {
	sum := &RunSummary{
		Suite:     set.Name,
		Model:     r.model,
		Results:   []CaseResult{},
		CreatedAt: time.Now().UTC(),
	}
	for _, c := range set.Cases {
		if ctx.Err() != nil {
			break
		}
		obs, err := a.Answer(ctx, c)
		if errors.Is(err, ErrSkip) {
			sum.Results = append(sum.Results, CaseResult{CaseID: c.ID, Vital: c.Vital, Skipped: true})
			continue
		}
		if err != nil {
			res := CaseResult{CaseID: c.ID, Vital: c.Vital, Passed: false, Error: err.Error()}
			sum.Results = append(sum.Results, res)
			sum.Total++
			if c.Vital {
				sum.VitalTotal++
				sum.VitalFailed++
			}
			continue
		}
		v := Score(c, obs)
		passed := v.Passed
		checks := v.Checks

		// T3: optional LLM-judge component. Refusals are judged by the
		// deterministic refusal check alone — grading a deliberate refusal with a
		// helpfulness rubric would penalise correct behaviour.
		if r.judge != nil && !c.Expect.Refuse {
			jr, judgeErr := r.judge.Judge(ctx, c, obs)
			if judgeErr != nil {
				if r.logger != nil {
					r.logger.Warn("eval judge failed — component omitted (deterministic result stands)",
						zap.String("case_id", c.ID), zap.Error(judgeErr))
				}
			} else {
				checks = append(checks, Check{
					Name:   "judge",
					Passed: jr.Passed,
					Detail: detailf("score=%.2f feedback=%s", jr.Score, jr.Feedback),
				})
				if !jr.Passed {
					passed = false
				}
			}
		}

		res := CaseResult{
			CaseID:    c.ID,
			Vital:     c.Vital,
			Passed:    passed,
			Checks:    checks,
			CostUSD:   obs.CostUSD,
			LatencyMs: obs.LatencyMs,
		}
		sum.Results = append(sum.Results, res)
		sum.Total++
		sum.CostUSD += obs.CostUSD
		if passed {
			sum.Passed++
		}
		if c.Vital {
			sum.VitalTotal++
			if !passed {
				sum.VitalFailed++
			}
		}
	}
	if sum.Total > 0 {
		sum.Score = float64(sum.Passed) / float64(sum.Total)
	}
	if r.logger != nil {
		r.logger.Info("eval run complete",
			zap.String("suite", sum.Suite),
			zap.Int("total", sum.Total),
			zap.Int("passed", sum.Passed),
			zap.Int("vital_failed", sum.VitalFailed),
			zap.Float64("score", sum.Score),
		)
	}
	return sum
}

// ScoreDelta returns current.Score - baseline.Score (0 when no baseline).
// Pure — unit-tested.
func ScoreDelta(baseline, current *RunSummary) float64 {
	if baseline == nil || current == nil || baseline.Total == 0 {
		return 0
	}
	return current.Score - baseline.Score
}

// Regression reports whether the current run regressed against the
// baseline: any vital failure, OR a score drop beyond tolerance. A run
// with no baseline is never a regression (nothing to compare). Pure.
func Regression(baseline, current *RunSummary, tolerance float64) bool {
	if current == nil {
		return false
	}
	if current.HasVitalFailure() {
		return true
	}
	if baseline == nil {
		return false
	}
	return ScoreDelta(baseline, current) < -tolerance
}
