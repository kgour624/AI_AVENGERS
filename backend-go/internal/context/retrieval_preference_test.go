package context

import "testing"

// The preference must be a nudge: it reorders, it never removes. This is the
// property that makes it safe to leave switched on for a caller that asks for it —
// a chunk the plain search would have returned is still in the result set.
func TestApplyPreferenceBoostReordersOnly(t *testing.T) {
	scores := []float64{0.90, 0.80, 0.70}
	matches := []bool{false, true, false}

	order := applyPreferenceBoost(scores, matches, DefaultPreferenceBoost)
	if len(order) != 3 {
		t.Fatalf("boosting changed the number of candidates: %d", len(order))
	}
	seen := map[int]bool{}
	for _, i := range order {
		if seen[i] {
			t.Fatalf("index %d appears twice — a preference must never duplicate", i)
		}
		seen[i] = true
	}
	// 0.80 * 1.25 = 1.00 beats 0.90, so the matching chunk leads.
	if order[0] != 1 {
		t.Fatalf("order = %v, want the matching chunk first", order)
	}
}

// A clearly better non-matching chunk must still win: a preference that overrules
// quality is a filter wearing a different name.
func TestPreferenceDoesNotOverruleAClearlyBetterChunk(t *testing.T) {
	scores := []float64{0.95, 0.40}
	matches := []bool{false, true}

	order := applyPreferenceBoost(scores, matches, DefaultPreferenceBoost)
	if order[0] != 0 {
		t.Fatalf("order = %v, want the much better chunk to stay first", order)
	}
}

func TestPreferenceMatchesSectionAndLayer(t *testing.T) {
	tests := []struct {
		name        string
		pref        RetrievalPreference
		sectionPath string
		layer       int
		want        bool
	}{
		{name: "inactive preference matches nothing", pref: RetrievalPreference{}, sectionPath: "RAG > Chunking", layer: 3, want: false},
		{name: "section substring, case-insensitive", pref: RetrievalPreference{Section: "chunking"}, sectionPath: "RAG > Chunking", layer: 1, want: true},
		{name: "section not present", pref: RetrievalPreference{Section: "safety"}, sectionPath: "RAG > Chunking", layer: 1, want: false},
		{name: "layer match", pref: RetrievalPreference{Layer: 3}, sectionPath: "", layer: 3, want: true},
		{name: "layer mismatch", pref: RetrievalPreference{Layer: 3}, sectionPath: "", layer: 1, want: false},
		// Either criterion is enough: requiring both would quietly turn the nudge
		// into the filter this type exists to avoid.
		{name: "either criterion is enough", pref: RetrievalPreference{Section: "safety", Layer: 3}, sectionPath: "RAG", layer: 3, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pref.Matches(tc.sectionPath, tc.layer); got != tc.want {
				t.Fatalf("Matches(%q, %d) = %v, want %v", tc.sectionPath, tc.layer, got, tc.want)
			}
		})
	}
}

// A boost of 1 or less would demote what the caller asked to prefer, which is a
// bug rather than a choice, so it falls back to the default.
func TestEffectiveBoostFallsBackForNonBoostValues(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "unset", in: 0, want: DefaultPreferenceBoost},
		{name: "one is not a boost", in: 1, want: DefaultPreferenceBoost},
		{name: "below one is not a boost", in: 0.5, want: DefaultPreferenceBoost},
		{name: "explicit boost is honoured", in: 2, want: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := (RetrievalPreference{Boost: tc.in}).effectiveBoost(); got != tc.want {
				t.Fatalf("effectiveBoost(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestPreferenceActive(t *testing.T) {
	if (RetrievalPreference{}).Active() {
		t.Fatal("an empty preference must be inactive")
	}
	if (RetrievalPreference{Section: "  "}).Active() {
		t.Fatal("a whitespace section must be inactive")
	}
	if !(RetrievalPreference{Layer: 2}).Active() {
		t.Fatal("a layer preference must be active")
	}
}

// The pool the reranker is asked for must be larger than the caller's limit, or a
// preference could never promote a chunk the reranker had left out.
func TestRetrievalPoolFactorIsLargerThanOne(t *testing.T) {
	if retrievalPoolFactor <= 1 {
		t.Fatalf("retrievalPoolFactor = %d, want > 1 so the boost has room to promote", retrievalPoolFactor)
	}
}
