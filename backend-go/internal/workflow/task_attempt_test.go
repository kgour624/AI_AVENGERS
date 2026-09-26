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

func TestDecideRevisionAttempt(t *testing.T) {
	original := uuid.New()
	redesign := uuid.New()
	succeeded := &TaskAttempt{RevisionID: original, Status: AttemptSucceeded, Attempt: 1}
	if got := decideRevisionAttempt(succeeded, original); got != DecisionSkipDone {
		t.Fatalf("same revision decision = %q, want %q", got, DecisionSkipDone)
	}
	if got := decideRevisionAttempt(succeeded, redesign); got != DecisionRun {
		t.Fatalf("new redesign decision = %q, want %q", got, DecisionRun)
	}
	failed := &TaskAttempt{RevisionID: original, Status: AttemptFailed, Attempt: maxTaskAttempts}
	if got := decideRevisionAttempt(failed, original); got != DecisionAbandoned {
		t.Fatalf("exhausted same revision decision = %q, want %q", got, DecisionAbandoned)
	}
	if got := decideRevisionAttempt(failed, redesign); got != DecisionRun {
		t.Fatalf("new revision after exhausted prior = %q, want %q", got, DecisionRun)
	}
}

func TestParseDesignRevisionIDAndChangeGoal(t *testing.T) {
	revision := uuid.New()
	if got := parseDesignRevisionID(revision.String()); got != revision {
		t.Fatalf("parsed revision = %s, want %s", got, revision)
	}
	if got := parseDesignRevisionID("not-a-uuid"); got != uuid.Nil {
		t.Fatalf("invalid revision = %s, want uuid.Nil", got)
	}
	if got := withChangeGoal("base description", " \n "); got != "base description" {
		t.Fatalf("empty change goal altered task: %q", got)
	}
	if got := withChangeGoal("base description", "Add a bulk endpoint"); got != "base description\n\nCLIENT-REQUESTED REDESIGN:\nAdd a bulk endpoint" {
		t.Fatalf("change goal not attached as expected: %q", got)
	}
}

func TestTaskStatusRevisionKeySeparatesEvents(t *testing.T) {
	workflowID, expertID := uuid.New(), uuid.New()
	originalKey := taskStatusDedupKey(workflowID, expertID, PhaseHighLevelDesign, "in_progress", RevisionInitial, 1)
	redesignID := uuid.New()
	redesignKey := taskStatusDedupKey(workflowID, expertID, PhaseHighLevelDesign, "in_progress", redesignID, 1)
	if originalKey == redesignKey {
		t.Fatal("redesign task status key must differ from original task status key")
	}
	otherPhaseKey := taskStatusDedupKey(workflowID, expertID, PhaseDetailedDesign, "in_progress", redesignID, 1)
	retryKey := taskStatusDedupKey(workflowID, expertID, PhaseHighLevelDesign, "in_progress", redesignID, 2)
	if otherPhaseKey == redesignKey {
		t.Fatal("task status keys must separate design phases")
	}
	if retryKey == redesignKey {
		t.Fatal("task status keys must separate retry numbers")
	}
	if again := taskStatusDedupKey(workflowID, expertID, PhaseHighLevelDesign, "in_progress", redesignID, 1); again != redesignKey {
		t.Fatal("same revision and attempt must have a repeatable event identity")
	}
}

func TestRedesignRevisionIDUsesApprovalIDAndIsStableForRecovery(t *testing.T) {
	workflowID := uuid.New()
	approvalID := uuid.New()
	if got := redesignRevisionID(workflowID, approvalID); got != approvalID {
		t.Fatalf("approval revision id = %s, want approval id %s", got, approvalID)
	}
	first := redesignRevisionID(workflowID, uuid.Nil)
	second := redesignRevisionID(workflowID, uuid.Nil)
	if first == uuid.Nil || first != second {
		t.Fatalf("fallback revision must be stable and non-zero, got %s and %s", first, second)
	}
}

func TestWithChangeGoalAddsOnlyNonEmptyGoal(t *testing.T) {
	base := "Update the API design"
	if got := withChangeGoal(base, " \n "); got != base {
		t.Fatalf("empty goal changed the task: %q", got)
	}
	if got := withChangeGoal(base, "Add a bulk move endpoint"); got != base+"\n\nCLIENT-REQUESTED REDESIGN:\nAdd a bulk move endpoint" {
		t.Fatalf("change goal missing or malformed: %q", got)
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

func TestSnapshotRunnerStateKeepsRedesignRevision(t *testing.T) {
	revision := uuid.New().String()
	state := &runnerState{
		Phase:              PhaseDetailedDesign,
		CompletedExpertIDs: []string{"expert-1"},
		DesignRevisionID:   revision,
		RedesignGoal:       "Add a bulk endpoint",
	}
	snapshot := snapshotRunnerState(state, nil)
	if snapshot.DesignRevisionID != revision || snapshot.RedesignGoal != state.RedesignGoal {
		t.Fatalf("snapshot lost revision data: %+v", snapshot)
	}
	snapshot.CompletedExpertIDs[0] = "changed"
	if state.CompletedExpertIDs[0] != "expert-1" {
		t.Fatal("snapshot must not alias the source completed-expert slice")
	}
}

// An unwired ledger must run the work rather than skip it. Skipping on "no data"
// is the failure mode this whole change removes.
func TestClaimWithoutLedgerRuns(t *testing.T) {
	var s *TaskAttemptStore
	if _, decision, err := s.Claim(context.Background(), uuid.Nil, PhaseHighLevelDesign, uuid.Nil, uuid.Nil); err != nil || decision != DecisionRun {
		t.Fatalf("unwired ledger: decision = %q err = %v, want %q and nil", decision, err, DecisionRun)
	}
}
