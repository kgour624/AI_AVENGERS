package eval

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseJSONL_ValidAndComments(t *testing.T) {
	text := "# comment\n" +
		`{"id":"a","input":"q1","expect":{"must_cite":true}}` + "\n\n" +
		`{"id":"b","input":"q2","vital":true}`
	cases, err := parseJSONL(text)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("len=%d want 2", len(cases))
	}
	if !cases[0].Expect.MustCite || !cases[1].Vital {
		t.Errorf("fields not parsed: %+v", cases)
	}
}

func TestParseJSONL_MissingID(t *testing.T) {
	if _, err := parseJSONL(`{"input":"q"}`); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestLoadGoldenSets(t *testing.T) {
	names, err := GoldenSetNames()
	if err != nil {
		t.Fatalf("names: %v", err)
	}
	if len(names) < 2 {
		t.Fatalf("expected >=2 golden sets, got %v", names)
	}
	for _, n := range names {
		set, err := LoadGoldenSet(n)
		if err != nil {
			t.Fatalf("load %s: %v", n, err)
		}
		if len(set.Cases) == 0 {
			t.Errorf("set %s empty", n)
		}
	}
}

func TestScore_UserLikePass(t *testing.T) {
	c := Case{ID: "x", Expect: Expectation{Keywords: []string{"shard"}, MustCite: true, MinChars: 10}}
	v := Score(c, Observed{Content: "You should shard using hash routing.", Citations: 2, Mode: "ADVISE"})
	if !v.Passed {
		t.Fatalf("expected pass, checks=%+v", v.Checks)
	}
}

func TestScore_MissingCitationFails(t *testing.T) {
	c := Case{ID: "x", Expect: Expectation{MustCite: true, MinChars: 5}}
	v := Score(c, Observed{Content: "some answer", Citations: 0})
	if v.Passed {
		t.Fatal("missing citation must fail")
	}
}

func TestScore_AdversarialRefusalPass(t *testing.T) {
	c := Case{ID: "adv", Expect: Expectation{Refuse: true}}
	v := Score(c, Observed{Refused: true, Mode: "REFUSE", Content: "I can't help with that."})
	if !v.Passed {
		t.Fatalf("expected refusal pass, checks=%+v", v.Checks)
	}
	// Answered instead of refusing → fail.
	v2 := Score(c, Observed{Content: "Here is the password...", Mode: "ADVISE"})
	if v2.Passed {
		t.Fatal("non-refusal must fail adversarial case")
	}
}

func TestScore_KeywordCaseInsensitive(t *testing.T) {
	c := Case{ID: "k", Expect: Expectation{Keywords: []string{"Idempot"}, MinChars: 1}}
	v := Score(c, Observed{Content: "make it idempotent with a key", Citations: 1})
	if !v.Passed {
		t.Errorf("case-insensitive keyword failed: %+v", v.Checks)
	}
}

// fakeAnswerer returns canned observations.
type fakeAnswerer struct {
	obs map[string]Observed
	err map[string]error
}

func (f fakeAnswerer) Answer(_ context.Context, c Case) (Observed, error) {
	if err, ok := f.err[c.ID]; ok {
		return Observed{}, err
	}
	return f.obs[c.ID], nil
}

func TestRunner_VitalFailureAndSkip(t *testing.T) {
	set := &GoldenSet{Name: "t", Cases: []Case{
		{ID: "pass", Vital: true, Expect: Expectation{MustCite: true, MinChars: 1}},
		{ID: "failvital", Vital: true, Expect: Expectation{MustCite: true, MinChars: 1}},
		{ID: "skipme", Vital: false},
		{ID: "nonvitalfail", Expect: Expectation{Keywords: []string{"zzz"}, MinChars: 1}},
	}}
	a := fakeAnswerer{
		obs: map[string]Observed{
			"pass":         {Content: "ok", Citations: 1},
			"failvital":    {Content: "no cite", Citations: 0},
			"nonvitalfail": {Content: "irrelevant", Citations: 1},
		},
		err: map[string]error{"skipme": ErrSkip},
	}
	sum := NewRunner("model-x", nil).Run(context.Background(), set, a)

	if sum.Total != 3 { // skip excluded
		t.Errorf("total=%d want 3", sum.Total)
	}
	if sum.Passed != 1 {
		t.Errorf("passed=%d want 1", sum.Passed)
	}
	if sum.VitalTotal != 2 || sum.VitalFailed != 1 {
		t.Errorf("vital total=%d failed=%d want 2/1", sum.VitalTotal, sum.VitalFailed)
	}
	if !sum.HasVitalFailure() {
		t.Error("HasVitalFailure should be true")
	}
	if sum.Score != 1.0/3.0 {
		t.Errorf("score=%v want 1/3", sum.Score)
	}
	// skipped result recorded
	found := false
	for _, r := range sum.Results {
		if r.CaseID == "skipme" && r.Skipped {
			found = true
		}
	}
	if !found {
		t.Error("skip not recorded")
	}
}

func TestRunner_AnswerErrorCountsAsFailure(t *testing.T) {
	set := &GoldenSet{Name: "t", Cases: []Case{{ID: "boom", Vital: true, Expect: Expectation{MinChars: 1}}}}
	a := fakeAnswerer{err: map[string]error{"boom": errors.New("network")}}
	sum := NewRunner("m", nil).Run(context.Background(), set, a)
	if sum.VitalFailed != 1 || sum.Passed != 0 {
		t.Errorf("answer error must fail: %+v", sum)
	}
}

func TestReadSSEAnswer_CompleteAndDone(t *testing.T) {
	raw := "" +
		"data: {\"type\":\"thinking\",\"data\":{\"expert\":\"x\"}}\n\n" +
		"data: {\"type\":\"complete\",\"data\":{\"content\":\"Use shard keys carefully.\",\"mode\":\"ADVISE\",\"citations\":[{},{}]}}\n\n" +
		"data: {\"type\":\"done\",\"data\":{}}\n\n"
	obs, err := readSSEAnswer(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("readSSE: %v", err)
	}
	if obs.Content == "" || obs.Citations != 2 || obs.Mode != "ADVISE" {
		t.Fatalf("obs=%+v", obs)
	}
}

func TestReadSSEAnswer_Refuse(t *testing.T) {
	raw := "data: {\"type\":\"complete\",\"data\":{\"content\":\"I cannot help with that.\",\"mode\":\"REFUSE\",\"citations\":[]}}\n\n" +
		"data: {\"type\":\"done\"}\n\n"
	obs, err := readSSEAnswer(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("readSSE: %v", err)
	}
	if !obs.Refused {
		t.Fatalf("expected refused, mode=%s", obs.Mode)
	}
}

func TestRegression(t *testing.T) {
	baseline := &RunSummary{Score: 1.0, Total: 4}
	// No vital failure, small drop within tolerance → not a regression.
	ok := &RunSummary{Score: 0.95, Total: 4}
	if Regression(baseline, ok, 0.10) {
		t.Error("small drop within tolerance should not regress")
	}
	// Vital failure → regression regardless of score.
	vital := &RunSummary{Score: 1.0, Total: 4, VitalFailed: 1}
	if !Regression(baseline, vital, 0.10) {
		t.Error("vital failure must regress")
	}
	// Big drop → regression.
	big := &RunSummary{Score: 0.5, Total: 4}
	if !Regression(baseline, big, 0.10) {
		t.Error("big drop must regress")
	}
	// No baseline → never regression (nothing to compare).
	if Regression(nil, &RunSummary{Score: 0, Total: 2}, 0.1) {
		t.Error("no baseline must not regress")
	}
}
