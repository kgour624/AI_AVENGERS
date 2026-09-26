package training

import (
	"testing"

	"github.com/google/uuid"
)

// Pins the fix for the NOT NULL incident: a nil RetrievedIDs (retrieval miss /
// empty context) must become an empty array, never NULL, before it reaches the
// database. If this regresses, expert_capability_results.retrieved_ids fails the
// INSERT again and the eval run aborts.
func TestNonNilIDs(t *testing.T) {
	if got := nonNilIDs(nil); got == nil || len(got) != 0 {
		t.Fatalf("nonNilIDs(nil) = %#v, want non-nil empty slice", got)
	}
	id := uuid.New()
	got := nonNilIDs([]uuid.UUID{id})
	if len(got) != 1 || got[0] != id {
		t.Fatalf("nonNilIDs must preserve values, got %#v", got)
	}
}
