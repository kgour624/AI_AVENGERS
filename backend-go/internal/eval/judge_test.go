package eval

import (
	"context"
	"errors"
	"testing"
)

func TestParseJudgeResult(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		floor     float64
		wantScore float64
		wantPass  bool
		wantErr   bool
	}{
		{name: "plain json above floor", raw: `{"score":0.8,"feedback":"ok"}`, floor: 0.6, wantScore: 0.8, wantPass: true},
		{name: "plain json below floor", raw: `{"score":0.4,"feedback":"missing x"}`, floor: 0.6, wantScore: 0.4, wantPass: false},
		{name: "fenced json is tolerated", raw: "```json\n{\"score\":0.9,\"feedback\":\"\"}\n```", floor: 0.6, wantScore: 0.9, wantPass: true},
		{name: "score clamped high", raw: `{"score":1.5}`, floor: 0.6, wantScore: 1.0, wantPass: true},
		{name: "score clamped low", raw: `{"score":-0.3}`, floor: 0.6, wantScore: 0.0, wantPass: false},
		{name: "garbage fails", raw: `not json`, floor: 0.6, wantErr: true},
		{name: "exactly at floor passes", raw: `{"score":0.6}`, floor: 0.6, wantScore: 0.6, wantPass: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseJudgeResult(tc.raw, tc.floor)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Score != tc.wantScore || got.Passed != tc.wantPass {
				t.Fatalf("score=%v passed=%v, want score=%v passed=%v", got.Score, got.Passed, tc.wantScore, tc.wantPass)
			}
		})
	}
}

type fakeJudge struct {
	result JudgeResult
	err    error
	calls  int
}

func (f *fakeJudge) Judge(context.Context, Case, Observed) (JudgeResult, error) {
	f.calls++
	return f.result, f.err
}

func TestRunnerJudgeLayerAddsComponentAndKeepsGate(t *testing.T) {
	set := &GoldenSet{Name: "t", Cases: []Case{
		{ID: "c1", Input: "q1", Vital: true, Expect: Expectation{Keywords: []string{"x"}}},
		{ID: "c2", Input: "q2", Expect: Expectation{Refuse: true}},
	}}
	answerer := mapAnswerer{byID: map[string]Observed{
		"c1": {Content: "x is here", Citations: 1},
		"c2": {Content: "I cannot help with that", Refused: true},
	}}

	// A judge that fails every case must still flip a deterministic PASS to FAIL.
	judge := &fakeJudge{result: JudgeResult{Score: 0.1, Passed: false, Feedback: "bad"}}
	run := NewRunner("test-model", nil).WithJudge(judge).Run(context.Background(), set, answerer)

	if judge.calls != 1 {
		t.Fatalf("judge calls = %d, want 1 (refusal case must not be judged)", judge.calls)
	}
	// c1 is judged and fails; c2 (refusal) passes on the deterministic check
	// alone, so exactly one case passes overall.
	if run.Passed != 1 {
		t.Fatalf("passed = %d, want 1 (only the refusal case passes)", run.Passed)
	}
	if run.VitalFailed != 1 {
		t.Fatalf("vital_failed = %d, want 1 (judge must fail the vital c1)", run.VitalFailed)
	}
	if run.Results[0].Passed {
		t.Fatalf("c1 must fail once the judge component fails: %+v", run.Results[0])
	}
	if !run.Results[1].Passed {
		t.Fatalf("refusal case should pass deterministically: %+v", run.Results[1])
	}
}

func TestRunnerJudgeErrorOmitsComponent(t *testing.T) {
	set := &GoldenSet{Name: "t", Cases: []Case{
		{ID: "c1", Input: "q1", Expect: Expectation{Keywords: []string{"x"}}},
	}}
	answerer := mapAnswerer{byID: map[string]Observed{"c1": {Content: "x", Citations: 1}}}
	judge := &fakeJudge{err: errors.New("model down")}

	run := NewRunner("test-model", nil).WithJudge(judge).Run(context.Background(), set, answerer)
	if run.Passed != 1 || run.Total != 1 {
		t.Fatalf("judge error must not flip the case: passed=%d total=%d", run.Passed, run.Total)
	}
	for _, ch := range run.Results[0].Checks {
		if ch.Name == "judge" {
			t.Fatalf("judge check must be omitted on error: %+v", ch)
		}
	}
}

type mapAnswerer struct {
	byID map[string]Observed
}

func (f mapAnswerer) Answer(_ context.Context, c Case) (Observed, error) {
	o, ok := f.byID[c.ID]
	if !ok {
		return Observed{}, ErrSkip
	}
	return o, nil
}
