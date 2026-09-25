package training

import (
	"strings"
	"testing"
)

// The gate decides whether an expert is published. Every condition and every way of
// failing one is pinned here, because this is the check that replaces "the pipeline
// says it went fine" with something a human can read.

func TestIngestGatePassesOnlyWhenEveryConditionHolds(t *testing.T) {
	healthy := GateInputs{
		Chunks:              1634,
		CharterRules:        7,
		ClarificationTopics: 5,
		MeasuredTopics:      6,
		CapabilityMeasured:  true,
		SmokeProbes:         5,
		SmokePassed:         true,
	}

	if conditions := EvaluateIngestGate(healthy); !GatePassed(conditions) {
		t.Fatalf("healthy inputs failed the gate: %v", UnmetGateReasons(conditions))
	}
	if reasons := UnmetGateReasons(EvaluateIngestGate(healthy)); len(reasons) != 0 {
		t.Fatalf("healthy inputs produced reasons: %v", reasons)
	}
}

func TestIngestGateFailsOneConditionAtATime(t *testing.T) {
	healthy := func() GateInputs {
		return GateInputs{
			Chunks:              1634,
			CharterRules:        7,
			ClarificationTopics: 5,
			MeasuredTopics:      6,
			CapabilityMeasured:  true,
			SmokeProbes:         5,
			SmokePassed:         true,
		}
	}

	tests := []struct {
		name     string
		mutate   func(*GateInputs)
		wantName string
	}{
		{
			name:     "too few chunks",
			mutate:   func(in *GateInputs) { in.Chunks = 99 },
			wantName: GateChunks,
		},
		{
			name:     "charter has too few rules",
			mutate:   func(in *GateInputs) { in.CharterRules = 4 },
			wantName: GateCharterRules,
		},
		{
			name:     "clarification covers too few topics",
			mutate:   func(in *GateInputs) { in.ClarificationTopics = 2 },
			wantName: GateClarification,
		},
		{
			name:     "capability measured but too shallow",
			mutate:   func(in *GateInputs) { in.MeasuredTopics = 2 },
			wantName: GateCapability,
		},
		{
			name:     "capability never measured",
			mutate:   func(in *GateInputs) { in.CapabilityMeasured = false; in.MeasuredTopics = 0 },
			wantName: GateCapability,
		},
		{
			name:     "smoke test failed",
			mutate:   func(in *GateInputs) { in.SmokePassed = false },
			wantName: GateSmoke,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := healthy()
			tc.mutate(&in)

			conditions := EvaluateIngestGate(in)
			if GatePassed(conditions) {
				t.Fatal("gate passed, want a failure")
			}
			reasons := UnmetGateReasons(conditions)
			if len(reasons) != 1 {
				t.Fatalf("got %d failures, want exactly 1: %v", len(reasons), reasons)
			}
			if !strings.Contains(reasons[0], tc.wantName) {
				t.Fatalf("failure %q does not name %q", reasons[0], tc.wantName)
			}
			// The reason must carry the numbers, not just the verdict: "failed" alone
			// leaves the admin with nothing to act on.
			if len(reasons[0]) < 30 {
				t.Fatalf("failure %q is too terse to act on", reasons[0])
			}
		})
	}
}

// An expert that was never measured and one that was measured and found shallow need
// different remedies, so they must not produce the same message.
func TestIngestGateDistinguishesNeverMeasuredFromShallow(t *testing.T) {
	base := GateInputs{
		Chunks: 1634, CharterRules: 7, ClarificationTopics: 5,
		SmokeProbes: 5, SmokePassed: true,
	}

	neverMeasured := base
	neverMeasured.CapabilityMeasured = false
	neverMeasured.MeasuredTopics = 0

	shallow := base
	shallow.CapabilityMeasured = true
	shallow.MeasuredTopics = 1

	neverReason := UnmetGateReasons(EvaluateIngestGate(neverMeasured))[0]
	shallowReason := UnmetGateReasons(EvaluateIngestGate(shallow))[0]

	if !strings.Contains(neverReason, "never been measured") {
		t.Fatalf("never-measured reason %q does not say it was never measured", neverReason)
	}
	if strings.Contains(shallowReason, "never been measured") {
		t.Fatalf("shallow reason %q wrongly says it was never measured", shallowReason)
	}
}

func TestCountCharterRules(t *testing.T) {
	tests := []struct {
		name    string
		charter string
		want    int
	}{
		{name: "empty", charter: "", want: 0},
		{name: "prose only", charter: "Always explain the trade-off before recommending.", want: 0},
		{
			name:    "bullets",
			charter: "- Why X beats Y\n- Why caching helps\n* Why sharding is last\n",
			want:    3,
		},
		{
			// '1.' and '2)' are list markers; a plain sentence is not a rule.
			name:    "numbered",
			charter: "1. Why latency matters\n2) Why consistency costs\nMost teams get this wrong.\n",
			want:    2,
		},
		{
			name:    "mixed with prose and blank lines",
			charter: "Reasoning charter:\n\n- Why one\n\nSome prose here.\n- Why two\n",
			want:    2,
		},
		{
			// A version number or a decimal must not be counted as a rule.
			name:    "numbers that are not list markers",
			charter: "Version 2.1 of the charter\n- Why one\n",
			want:    1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CountCharterRules(tc.charter); got != tc.want {
				t.Fatalf("CountCharterRules = %d, want %d", got, tc.want)
			}
		})
	}
}

// The smoke test's deterministic part feeds the gate's smoke condition, so it is
// pinned here: the three ways an answer can be unusable must be distinguishable.
func TestRefusalOrCitationFailure(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		want   string
	}{
		{name: "well-formed answer", answer: "Sharding splits data [1].", want: ""},
		{name: "explicit refusal", answer: insufficientContextMarker, want: FailureRefused},
		{name: "refusal with surrounding text", answer: "I cannot say. " + insufficientContextMarker, want: FailureRefused},
		{name: "empty", answer: "   ", want: FailureEmpty},
		{name: "answer without any citation", answer: "Sharding splits data across nodes.", want: FailureNotCited},
		{name: "bracketed word is not a citation", answer: "See [figure] for details.", want: FailureNotCited},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := refusalOrCitationFailure(tc.answer); got != tc.want {
				t.Fatalf("refusalOrCitationFailure = %q, want %q", got, tc.want)
			}
		})
	}
}

// buildNumberedContext is what the answer cites and what the checker validates, so
// the numbering must be shared rather than re-derived at each site.
func TestBuildNumberedContextNumbersInOrder(t *testing.T) {
	got := buildNumberedContext([]string{"alpha", "beta"}, 100)
	if !strings.Contains(got, "[1] alpha") || !strings.Contains(got, "[2] beta") {
		t.Fatalf("context numbering wrong: %q", got)
	}

	// The per-text limit must actually bound what enters a prompt.
	long := strings.Repeat("x", 50)
	clipped := buildNumberedContext([]string{long}, 10)
	if strings.Contains(clipped, strings.Repeat("x", 11)) {
		t.Fatal("context was not clipped to the per-text limit")
	}
}
