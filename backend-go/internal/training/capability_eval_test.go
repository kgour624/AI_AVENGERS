package training

import (
	"strings"
	"testing"
)

// The numbers and shapes here come from the real failure the feature exists to
// detect: an expert whose corpus holds the answer but whose retrieval, citations
// or grounding did not deliver it. Each check is tested separately because they
// fail separately and need different fixes.

func TestParseGeneratedQuestionsDropsUnboundEntries(t *testing.T) {
	response := "```json\n" + `{"questions":[
		{"passage":1,"level":1,"question":"What is consistent hashing?"},
		{"passage":2,"level":2,"question":"How does a ring reduce reshuffling?"},
		{"passage":9,"level":3,"question":"this one names a passage that does not exist"},
		{"passage":2,"level":3,"question":"duplicate for passage 2"},
		{"passage":3,"level":99,"question":"What breaks when a node dies?"}
	]}` + "\n```"

	got, err := parseGeneratedQuestions(response, 3)
	if err != nil {
		t.Fatalf("parseGeneratedQuestions returned error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d questions, want 3 (out-of-range and duplicate passages dropped)", len(got))
	}

	// Passage 9 must be DROPPED, not clamped: a question bound to the wrong chunk
	// would score retrieval against content that cannot answer it, which would
	// look like a retrieval failure.
	for _, q := range got {
		if q.PassageIdx < 0 || q.PassageIdx > 2 {
			t.Fatalf("passage index %d is outside the supplied range", q.PassageIdx)
		}
	}

	// An out-of-range level is clamped to the shallowest, never to the deepest:
	// claiming more depth than the model asked for would inflate the measurement.
	last := got[2]
	if last.PassageIdx != 2 {
		t.Fatalf("passage 3 mapped to index %d, want 2", last.PassageIdx)
	}
	if last.Level != CapabilityLevelDefinition {
		t.Fatalf("level 99 clamped to %d, want %d", last.Level, CapabilityLevelDefinition)
	}
}

func TestParseGeneratedQuestionsRejectsUselessResponses(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{name: "prose without json", response: "I cannot help with that request."},
		{name: "empty questions array", response: `{"questions":[]}`},
		{name: "all passages out of range", response: `{"questions":[{"passage":7,"level":1,"question":"x"}]}`},
		{name: "blank question text", response: `{"questions":[{"passage":1,"level":1,"question":"   "}]}`},
		{name: "empty response", response: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseGeneratedQuestions(tc.response, 3); err == nil {
				t.Fatal("want an error, got nil")
			}
		})
	}
}

// measuredLevelFrom is the judgement the whole feature rests on, so every branch
// gets pinned.
func TestMeasuredLevelIsCumulative(t *testing.T) {
	tests := []struct {
		name   string
		passed map[int]int
		cases  map[int]int
		want   int
	}{
		{
			name:   "all three levels passed",
			passed: map[int]int{1: 1, 2: 1, 3: 1},
			cases:  map[int]int{1: 1, 2: 1, 3: 1},
			want:   3,
		},
		{
			name:   "definitions and mechanics pass, failure modes fail",
			passed: map[int]int{1: 2, 2: 2, 3: 0},
			cases:  map[int]int{1: 2, 2: 2, 3: 2},
			want:   2,
		},
		{
			name:   "only definitions pass",
			passed: map[int]int{1: 1, 2: 0, 3: 0},
			cases:  map[int]int{1: 1, 2: 1, 3: 1},
			want:   1,
		},
		{
			name:   "nothing passed",
			passed: map[int]int{1: 0, 2: 0, 3: 0},
			cases:  map[int]int{1: 1, 2: 1, 3: 1},
			want:   0,
		},
		{
			// A failure-mode question that happens to pass, while definitions fail,
			// must NOT count as depth. That is luck, not capability.
			name:   "deep question passes but definitions fail",
			passed: map[int]int{1: 0, 2: 0, 3: 1},
			cases:  map[int]int{1: 1, 2: 1, 3: 1},
			want:   0,
		},
		{
			// Half is the threshold (majority), so one of two passing counts.
			name:   "half of a level passes",
			passed: map[int]int{1: 1, 2: 1},
			cases:  map[int]int{1: 2, 2: 2},
			want:   2,
		},
		{
			name:   "just below half does not count",
			passed: map[int]int{1: 1, 2: 1},
			cases:  map[int]int{1: 2, 2: 3},
			want:   1,
		},
		{
			// Nothing was asked at level 2, so nothing is claimed there — and the
			// levels above it cannot be reached either.
			name:   "missing level breaks the chain",
			passed: map[int]int{1: 1, 3: 1},
			cases:  map[int]int{1: 1, 3: 1},
			want:   1,
		},
		{
			name:   "no cases at all",
			passed: map[int]int{},
			cases:  map[int]int{},
			want:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := measuredLevelFrom(tc.passed, tc.cases); got != tc.want {
				t.Fatalf("measuredLevelFrom = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCountCitationMarkers(t *testing.T) {
	tests := []struct {
		name   string
		answer string
		want   int
	}{
		{name: "no markers", answer: "Sharding splits data across nodes.", want: 0},
		{name: "single marker", answer: "Sharding splits data [1].", want: 1},
		{name: "multiple markers", answer: "Sharding [1] splits data [2][3].", want: 3},
		{name: "multi digit marker", answer: "See [12].", want: 1},
		{
			// A word in brackets is not a citation; the check must not be
			// satisfiable by accident.
			name:   "bracketed word is not a citation",
			answer: "See [figure] and [note].",
			want:   0,
		},
		{name: "markdown link is not a citation", answer: "See [docs](http://x).", want: 0},
		{name: "unclosed bracket", answer: "See [1 and more.", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := countCitationMarkers(tc.answer); got != tc.want {
				t.Fatalf("countCitationMarkers = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParseJudgeVerdictDefaultsToUnverifiable(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{name: "supported", response: `{"verdict":"supported","reason":"ok"}`, want: "supported"},
		{name: "refuted with prose around it", response: "Here you go:\n```json\n{\"verdict\":\"REFUTED\"}\n```", want: "refuted"},
		{
			// An unreadable judge must never be read as a pass.
			name:     "unparseable",
			response: "I think it is probably fine.",
			want:     "unverifiable",
		},
		{name: "unknown verdict value", response: `{"verdict":"maybe"}`, want: "unverifiable"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseJudgeVerdict(tc.response); got != tc.want {
				t.Fatalf("parseJudgeVerdict = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildTopicReportKeepsTheFailuresAsClaims(t *testing.T) {
	results := []CapabilityCaseResult{
		{Level: 1, Question: "What is sharding?", Passed: true, HitRank: 1, Grounded: "supported"},
		{Level: 1, Question: "When is sharding used?", Passed: true, HitRank: 2, Grounded: "supported"},
		{Level: 2, Question: "How does rebalancing work?", Passed: false, FailureReason: FailureRetrievalMissed},
		{Level: 3, Question: "What breaks when a shard is hot?", Passed: false, FailureReason: FailureUngrounded},
	}

	report := buildTopicReport("sharding", 14, results)

	if report.Cases != 4 || report.Passed != 2 {
		t.Fatalf("cases=%d passed=%d, want 4 and 2", report.Cases, report.Passed)
	}
	// Level 2 was asked and did not pass, so the level stops at 1.
	if report.MeasuredLevel != 1 {
		t.Fatalf("measured level = %d, want 1", report.MeasuredLevel)
	}
	if len(report.CanHandle) != 2 {
		t.Fatalf("can_handle has %d entries, want the 2 questions that passed", len(report.CanHandle))
	}
	if len(report.CannotHandle) != 2 {
		t.Fatalf("cannot_handle has %d entries, want the 2 that failed", len(report.CannotHandle))
	}
	// The limit must carry the reason, so the list says WHY, not just what.
	if !strings.Contains(report.CannotHandle[0], FailureRetrievalMissed) {
		t.Fatalf("cannot_handle entry %q does not name the failure reason", report.CannotHandle[0])
	}
	// Reasons are de-duplicated and sorted, so the report does not repeat itself.
	if len(report.FailureReasons) != 2 {
		t.Fatalf("failure reasons = %v, want 2 distinct", report.FailureReasons)
	}
}

func TestBuildTopicReportCapsTheClaimLists(t *testing.T) {
	var results []CapabilityCaseResult
	for i := 0; i < 12; i++ {
		results = append(results, CapabilityCaseResult{
			Level: 1, Question: "q", Passed: true, HitRank: 1, Grounded: "supported",
		})
	}
	for i := 0; i < 9; i++ {
		results = append(results, CapabilityCaseResult{
			Level: 2, Question: "f", Passed: false, FailureReason: FailureNotCited,
		})
	}

	report := buildTopicReport("caching", 0, results)
	if len(report.CanHandle) != canHandleLimit {
		t.Fatalf("can_handle = %d entries, want the cap of %d", len(report.CanHandle), canHandleLimit)
	}
	if len(report.CannotHandle) != cannotHandleLimit {
		t.Fatalf("cannot_handle = %d entries, want the cap of %d", len(report.CannotHandle), cannotHandleLimit)
	}
}

// The link comparison decides whether an expensive feature earns its place, so the
// verdict rule and its refusal to judge a tiny sample are pinned here.
func TestLinkVerdict(t *testing.T) {
	tests := []struct {
		name        string
		commonCodes int
		withoutHits int
		withHits    int
		want        string
	}{
		{name: "links found more sources", commonCodes: 9, withoutHits: 6, withHits: 9, want: "improved"},
		{name: "links found fewer sources", commonCodes: 9, withoutHits: 9, withHits: 7, want: "regressed"},
		{name: "identical result", commonCodes: 15, withoutHits: 15, withHits: 15, want: "unchanged"},
		{
			// The case that produced this work: one mode answered 9 questions and the
			// other 15, and the totals looked equal. A verdict needs shared questions.
			name:        "too few shared questions",
			commonCodes: 2,
			withoutHits: 2,
			withHits:    1,
			want:        "inconclusive",
		},
		{name: "no shared questions", commonCodes: 0, withoutHits: 0, withHits: 0, want: "inconclusive"},
		{name: "exactly at the minimum", commonCodes: 3, withoutHits: 2, withHits: 3, want: "improved"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := linkVerdict(tc.commonCodes, tc.withoutHits, tc.withHits); got != tc.want {
				t.Fatalf("linkVerdict(%d, %d, %d) = %q, want %q",
					tc.commonCodes, tc.withoutHits, tc.withHits, got, tc.want)
			}
		})
	}
}

func TestMinimumComparableCasesIsMeaningful(t *testing.T) {
	// A guard on the constant itself: a comparison judged from fewer than three
	// questions is a coin toss, and lowering this silently would let it be presented
	// as evidence.
	if minimumComparableCases < 3 {
		t.Fatalf("minimumComparableCases = %d, want at least 3", minimumComparableCases)
	}
}
