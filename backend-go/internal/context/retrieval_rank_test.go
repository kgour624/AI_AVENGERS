package context

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// salientTerms is the safety boundary: its output is joined into a tsquery, so the
// whitelist is tested as a property, not as a convenience.

func TestSalientTermsDropsQuestionScaffolding(t *testing.T) {
	got := salientTerms("How does rebalancing work in consistent hashing?")

	want := []string{"rebalancing", "consistent", "hashing"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v (question order must be preserved)", got, want)
		}
	}
}

func TestSalientTerms(t *testing.T) {
	tests := []struct {
		name     string
		question string
		wantLen  int
	}{
		{name: "empty", question: "", wantLen: 0},
		{name: "only stopwords", question: "how does it work", wantLen: 0},
		{name: "single chars dropped", question: "a b c kafka", wantLen: 1},
		{name: "bare numbers dropped", question: "2024 100 kafka", wantLen: 1},
		{name: "duplicates collapsed", question: "kafka kafka redis", wantLen: 2},
		{name: "short but meaningful terms kept", question: "s3 io", wantLen: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := salientTerms(tc.question); len(got) != tc.wantLen {
				t.Fatalf("salientTerms(%q) = %v, want %d terms", tc.question, got, tc.wantLen)
			}
		})
	}
}

func TestSalientTermsCapsTheTermCount(t *testing.T) {
	// A long question must not turn the keyword filter into "match anything".
	question := strings.Repeat("alpha bravo charlie delta echo foxtrot golf hotel india juliet ", 4)
	if got := salientTerms(question); len(got) > maxKeywordTerms {
		t.Fatalf("got %d terms, want at most %d", len(got), maxKeywordTerms)
	}
}

// The property that makes the tsquery construction safe: whatever the question
// contains, the terms that reach SQL contain only [a-z0-9_]. Anything else would be
// tsquery syntax injected by the caller.
func TestSalientTermsOutputIsTsquerySafe(t *testing.T) {
	questions := []string{
		"kafka | redis & !drop (secret):*",
		"'; DROP TABLE course_chunks; --",
		"' OR '1'='1",
		"what is a & b | c ! (d) :e",
		"caf\u00e9 na\u00efve \u4e2d\u6587",
	}

	for _, question := range questions {
		for _, term := range salientTerms(question) {
			for _, r := range term {
				isLower := r >= 'a' && r <= 'z'
				isDigit := r >= '0' && r <= '9'
				if !isLower && !isDigit && r != '_' {
					t.Fatalf("term %q from %q contains %q, which is not whitelisted", term, question, r)
				}
			}
		}
	}
}

func TestReciprocalRankFusionPromotesAgreement(t *testing.T) {
	both := uuid.New()        // in both lists
	vectorOnly := uuid.New()  // top of the vector list
	keywordOnly := uuid.New() // top of the keyword list

	// "both" is 2nd in one list and 2nd in the other; vectorOnly is 1st but appears
	// once. Agreement between two independent retrievers must beat a single first
	// place, which is the reason RRF exists.
	got := reciprocalRankFusion([][]uuid.UUID{
		{vectorOnly, both},
		{keywordOnly, both},
	})
	if len(got) != 3 {
		t.Fatalf("got %d ids, want 3", len(got))
	}
	if got[0] != both {
		t.Fatalf("first id = %v, want the id present in both lists (%v)", got[0], both)
	}
}

func TestReciprocalRankFusionIsDeterministic(t *testing.T) {
	a, b := uuid.New(), uuid.New()

	first := reciprocalRankFusion([][]uuid.UUID{{a}, {b}})
	// Same input, run again: a map iteration order leak would show up here as a
	// different order, which would make two eval runs incomparable.
	for i := 0; i < 20; i++ {
		again := reciprocalRankFusion([][]uuid.UUID{{a}, {b}})
		if len(again) != len(first) {
			t.Fatalf("length changed between runs")
		}
		for j := range again {
			if again[j] != first[j] {
				t.Fatalf("order changed between runs: %v vs %v", again, first)
			}
		}
	}
}

func TestReciprocalRankFusionHandlesEmptyAndNil(t *testing.T) {
	if got := reciprocalRankFusion(nil); len(got) != 0 {
		t.Fatalf("nil input produced %v", got)
	}
	if got := reciprocalRankFusion([][]uuid.UUID{{}, {}}); len(got) != 0 {
		t.Fatalf("empty lists produced %v", got)
	}

	id := uuid.New()
	got := reciprocalRankFusion([][]uuid.UUID{{uuid.Nil, id}})
	if len(got) != 1 || got[0] != id {
		t.Fatalf("nil ids must be skipped, got %v", got)
	}
}

// Rank position must matter: an id ranked first is worth more than the same id
// ranked last, otherwise fusion would be no better than a union.
func TestReciprocalRankFusionRespectsRank(t *testing.T) {
	top := uuid.New()
	bottom := uuid.New()

	got := reciprocalRankFusion([][]uuid.UUID{
		{top, uuid.New(), uuid.New(), uuid.New(), bottom},
	})
	if len(got) != 5 {
		t.Fatalf("got %d ids, want 5", len(got))
	}
	if got[0] != top {
		t.Fatalf("first = %v, want the rank-1 id %v", got[0], top)
	}
	if got[len(got)-1] != bottom {
		t.Fatalf("last = %v, want the rank-5 id %v", got[len(got)-1], bottom)
	}
}
