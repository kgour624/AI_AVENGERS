package memory

import (
	"strings"
	"unicode"
)

// T6: memory discipline — a memory is only worth storing if it can change what
// the system does next time. The chat path can hand L2Store.Append almost any
// sentence a conversation produced; storing acknowledgements, greetings and
// one-off questions fills the project memory with rows that never influence a
// future answer, and they then compete with real decisions in retrieval.
//
// WHY the rules are conservative: a false positive here DELETES a fact the
// expert would have used. So the gate only drops things that are unambiguously
// not facts — it never tries to judge "importance", which is what importance and
// the consolidator's SNR pass are for. Anything declarative is kept.
//
// WHY deterministic (no LLM): this runs on every append, including the ones the
// ingest and workflow paths make. A model call to decide whether to remember
// would cost more than the memory it saves, and would make retention
// non-reproducible. The discipline is a filter, not a judge.

// minFactRunes is the shortest content that can still carry a fact. Below this
// the text is a fragment ("ok", "yes", "hmm") rather than a statement.
const minFactRunes = 5

// nonFactPhrases are conversational turns that never change future behaviour.
// Matched against the whole (trimmed, lowercased) content, so "ok" is dropped
// but "ok, use Postgres" is kept (it carries a decision).
var nonFactPhrases = map[string]bool{
	"ok": true, "okay": true, "k": true, "thanks": true, "thank you": true,
	"got it": true, "cool": true, "great": true, "sure": true, "yes": true,
	"no": true, "hmm": true, "hello": true, "hi": true, "hey": true,
	"bye": true, "good": true, "nice": true, "perfect": true, "done": true,
	"understood": true, "alright": true, "yep": true, "nope": true,
}

// decisionMarkers are verbs/markers that make even a short line a fact worth
// keeping, and that make a question ("should we use X? yes") a real decision.
var decisionMarkers = []string{
	"decid", "chose", "choose", "will use", "we use", "agreed", "agreement",
	"requirement", "constraint", "must ", "should ", "prefer", "plan to",
	"need to", "use ", "using ", "migrat", "implement", "config", "deploy",
	"version", "schema", "policy", "budget", "deadline", "owner",
}

// ShouldRemember reports whether an L2 entry's content is worth storing, and a
// stable reason when it is not (for logs/metrics).
func ShouldRemember(content string) (bool, string) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false, "empty"
	}
	lower := strings.ToLower(trimmed)

	// Match the phrase against the content with trailing punctuation removed, so
	// "Thanks!" and "ok." are the same acknowledgement as "thanks" and "ok".
	if nonFactPhrases[strings.ToLower(strings.TrimRight(trimmed, "!?.,;:…"))] {
		return false, "acknowledgement"
	}

	// A line with no letters or digits at all ("...", "---") carries nothing.
	hasAlnum := false
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlnum = true
			break
		}
	}
	if !hasAlnum {
		return false, "no_content"
	}

	if !containsDecisionMarker(lower) {
		// A single question on its own ("what about caching?") is a request for
		// information, not a stored fact. If it also carries a decision marker
		// it is kept — that phrasing usually records the answer too.
		if strings.HasSuffix(trimmed, "?") && !strings.Contains(trimmed, "\n") {
			return false, "bare_question"
		}
		if len([]rune(trimmed)) < minFactRunes {
			return false, "too_short"
		}
	}
	return true, ""
}

func containsDecisionMarker(lower string) bool {
	for _, m := range decisionMarkers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}
