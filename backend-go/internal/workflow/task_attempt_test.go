package workflow

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// The whole skip/retry policy is one pure function, so this table is the
// specification: what may be skipped, what must be re-run, and what must never
// be silently skipped.
func TestDecideAttempt(t *testing.T) {
	tests := []struct {
		name string
		prev *TaskAttempt
		want AttemptDecision
	}{
		{
			name: "never attempted",
			prev: nil,
			want: DecisionRun,
		},
		{
			name: "already succeeded for this exact phase",
			prev: &TaskAttempt{Status: AttemptSucceeded, Attempt: 1},
			want: DecisionSkipDone,
		},
		{
			name: "a failed attempt may be retried",
			prev: &TaskAttempt{Status: AttemptFailed, Attempt: 1},
			want: DecisionRetry,
		},
		{
			// A 'running' row means the process died mid-attempt. Leaving it
			// alone would strand the unit forever; single-flight guarantees no
			// other runner is driving this workflow.
			name: "an interrupted attempt may be retried",
			prev: &TaskAttempt{Status: AttemptRunning, Attempt: 1},
			want: DecisionRetry,
		},
		{
			name: "the last allowed attempt may still be retried",
			prev: &TaskAttempt{Status: AttemptFailed, Attempt: maxTaskAttempts - 1},
			want: DecisionRetry,
		},
		{
			// The budget is spent: the unit must FAIL, not be skipped. A silent
			// skip is indistinguishable from work that never happened — the bug
			// this ledger was introduced to stop.
			name: "an exhausted failed unit is not skipped, it is abandoned",
			prev: &TaskAttempt{Status: AttemptFailed, Attempt: maxTaskAttempts},
			want: DecisionAbandoned,
		},
		{
			name: "an exhausted interrupted unit is also abandoned",
			prev: &TaskAttempt{Status: AttemptRunning, Attempt: maxTaskAttempts + 5},
			want: DecisionAbandoned,
		},
		{
			// Nobody writes this value. Treating an unknown status as "done"
			// would move the old bug to a new place.
			name: "an unknown status is not evidence of success",
			prev: &TaskAttempt{Status: "who-knows", Attempt: 1},
			want: DecisionRun,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := decideAttempt(tc.prev); got != tc.want {
				t.Fatalf("decideAttempt(%+v) = %q, want %q", tc.prev, got, tc.want)
			}
		})
	}
}

func TestMaxTaskAttemptsIsBounded(t *testing.T) {
	// A guard on the constant itself: a cap of 1 would retry nothing (one
	// transient provider error would permanently fail a unit), and a very large
	// cap would let a broken task burn credits on every resume.
	if maxTaskAttempts < 2 {
		t.Fatalf("maxTaskAttempts = %d, want at least 2 so one transient failure is survivable", maxTaskAttempts)
	}
	if maxTaskAttempts > 5 {
		t.Fatalf("maxTaskAttempts = %d, want at most 5 so a broken unit cannot burn credits forever", maxTaskAttempts)
	}
}

func TestContainsID(t *testing.T) {
	ids := []string{"a", "b"}
	if !containsID(ids, "b") {
		t.Fatal("expected b to be found")
	}
	if containsID(ids, "c") {
		t.Fatal("expected c not to be found")
	}
	if containsID(nil, "a") {
		t.Fatal("expected an empty list to contain nothing")
	}
}

func TestAttemptNumberNilSafe(t *testing.T) {
	if got := attemptNumber(nil); got != 0 {
		t.Fatalf("attemptNumber(nil) = %d, want 0", got)
	}
	if got := attemptNumber(&TaskAttempt{Attempt: 2}); got != 2 {
		t.Fatalf("attemptNumber = %d, want 2", got)
	}
}

// An unwired ledger must run the work rather than skip it. Skipping on "no data"
// is the failure mode this whole change removes.
func TestClaimWithoutLedgerRuns(t *testing.T) {
	var s *TaskAttemptStore
	if _, decision, err := s.Claim(context.Background(), uuid.Nil, PhaseHighLevelDesign, uuid.Nil); err != nil || decision != DecisionRun {
		t.Fatalf("unwired ledger: decision = %q err = %v, want %q and nil", decision, err, DecisionRun)
	}
}
