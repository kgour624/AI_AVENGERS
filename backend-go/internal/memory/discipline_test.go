package memory

import "testing"

func TestShouldRememberKeepsFacts(t *testing.T) {
	keep := []string{
		"Decided to use PostgreSQL for the order ledger",
		"We agreed the API is versioned with a header",
		"Constraint: the deploy must finish before 18:00 UTC",
		"Auth: JWT",                  // short but declarative
		"Should we use Kafka? yes",   // a question that also records the answer
		"Added a retry with backoff to the ingest worker",
	}
	for _, c := range keep {
		if ok, reason := ShouldRemember(c); !ok {
			t.Errorf("expected to keep %q (dropped as %s)", c, reason)
		}
	}
}

func TestShouldRememberDropsNonFacts(t *testing.T) {
	drop := map[string]string{
		"":                   "empty",
		"   ":                "empty",
		"ok":                 "acknowledgement",
		"Thanks!":            "acknowledgement",
		"hmm":                "acknowledgement",
		"what about caching?": "bare_question",
		"...":                "no_content",
		"---":                "no_content",
		"yes":                "acknowledgement",
	}
	for c, wantReason := range drop {
		if ok, reason := ShouldRemember(c); ok {
			t.Errorf("expected to drop %q", c)
		} else if reason != wantReason {
			t.Errorf("drop reason for %q = %q, want %q", c, reason, wantReason)
		}
	}
}
