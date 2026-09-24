package workflow

import "testing"

func TestAttackIsClean(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"NONE", true},
		{"none", true},
		{"NONE — no issues found", true},
		{"", false}, // empty → fail closed (not clean)
		{"FINDINGS:\n- [SEVERITY=high] SQL injection", false},
		{"NONE\nFINDINGS:\n- [SEVERITY=low] typo", false}, // contradictory
		{"Looks fine to me", false},
	}
	for _, c := range cases {
		if got := AttackIsClean(c.in); got != c.want {
			t.Errorf("AttackIsClean(%q)=%v want %v", c.in, got, c.want)
		}
	}
}

func TestClassifyAttackSeverity(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"NONE", "none"},
		{"FINDINGS:\n- [SEVERITY=high] auth bypass", "high"},
		{"- SEVERITY=medium path traversal", "medium"},
		{"- SEVERITY: LOW style nit", "low"},
		{"- [HIGH] injection", "high"},
		{"vague complaint", "unknown"},
	}
	for _, c := range cases {
		if got := ClassifyAttackSeverity(c.in); got != c.want {
			t.Errorf("ClassifyAttackSeverity(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestParseDebateVerdict_FailClosed(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"PASS", DebatePass},
		{"pass — residual low only", DebatePass},
		{"FAIL\nhigh severity remains", DebateFail},
		{"ESCALATE\nproduct call needed", DebateEscalate},
		{"", DebateFail},            // empty → fail closed
		{"maybe ok?", DebateFail},   // unparseable → fail closed
		{"I think this is fine", DebateFail},
		{"Verdict: PASS\nreason", DebatePass},
	}
	for _, c := range cases {
		if got := ParseDebateVerdict(c.in); got != c.want {
			t.Errorf("ParseDebateVerdict(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestIsHighStakes(t *testing.T) {
	if !IsHighStakes("architecture_decision") {
		t.Fatal("architecture_decision must be high-stakes")
	}
	if !IsHighStakes("code_artifact_produced") {
		t.Fatal("code_artifact_produced must be high-stakes")
	}
	if IsHighStakes("requirement_captured") {
		t.Fatal("requirement_captured must stay on cheap single-pass path")
	}
	if IsHighStakes("review_comment") {
		t.Fatal("review_comment is not an artifact under debate")
	}
}

func TestSetDebatePolicy_NilSafeAndDefaultHops(t *testing.T) {
	var nilCV *CrossVerifier
	nilCV.SetDebatePolicy(DebatePolicy{Enabled: true}) // must not panic

	cv := &CrossVerifier{}
	if cv.debateEnabled() {
		t.Fatal("zero-value CrossVerifier must have debate disabled")
	}
	cv.SetDebatePolicy(DebatePolicy{Enabled: true, MaxHops: 0})
	if !cv.debateEnabled() {
		t.Fatal("enabled policy must turn debate on")
	}
	if cv.debate.MaxHops != defaultDebateMaxHops {
		t.Fatalf("MaxHops 0 must snap to default %d, got %d", defaultDebateMaxHops, cv.debate.MaxHops)
	}
}
