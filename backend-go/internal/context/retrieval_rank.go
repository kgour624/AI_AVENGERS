package context

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// I4: retrieval ranking.
//
// Three defects in the hybrid keyword half are fixed here, and all three were
// invisible because the cross-encoder reranker re-scores whatever it is handed:
//
//  1. The keyword query had NO ORDER BY. `WHERE chunk_text_tsv @@ ... LIMIT 10`
//     returns ten arbitrary rows, so the keyword half contributed whatever the
//     planner happened to find first rather than the best matches.
//  2. It used plainto_tsquery, which ANDs every lexeme. A real question
//     ("How does rebalancing work in consistent hashing?") requires ALL of
//     rebalanc/work/consist/hash to appear in one chunk, so for natural-language
//     input the keyword half usually matched nothing at all. The comment above it
//     claims exact terms like "PostgreSQL" are why it exists — true, but only for
//     one-word queries.
//  3. The two ranked lists were CONCATENATED (vector results, then keyword results).
//     Nothing was fused, so vector order always won: the keyword hits sat at the end
//     of the candidate list, and on the reranker-unavailable path — which returns the
//     first `limit` candidates — they were effectively discarded.
//
// Fixing (3) with Reciprocal Rank Fusion also makes the fallback path honest, since
// the fallback's output order IS this fused order.

// rrfK is the Reciprocal Rank Fusion constant. 60 is the value the course material
// gives (usable range 40-80). It is large enough that the difference between rank 1
// and rank 2 stays small — the point of RRF is that neither list's absolute scores
// dominate, only relative rank does.
const rrfK = 60.0

// maxKeywordTerms caps how many question terms reach the keyword query.
//
// WHY capped: the tsquery is built from these terms with OR semantics, so a long
// question would otherwise match on any common word and flood the candidate set with
// noise that the reranker then has to sort out. A handful of salient terms keeps the
// recall benefit (the ask that a chunk need not contain EVERY question word) without
// turning the filter into "match anything".
const maxKeywordTerms = 8

// keywordStopwords are dropped before building the keyword query.
//
// WHY a local list rather than relying on the 'english' dictionary: the dictionary
// removes stopwords from the tsvector, but these terms are what the QUESTION is turned
// into, and a query term that exists only as a stopword would match nothing while
// still counting against the cap. Small, explicit, and testable beats clever.
var keywordStopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "are": {}, "but": {}, "not": {}, "you": {},
	"all": {}, "any": {}, "can": {}, "how": {}, "why": {}, "what": {}, "when": {},
	"where": {}, "which": {}, "who": {}, "does": {}, "did": {}, "was": {}, "were": {},
	"this": {}, "that": {}, "these": {}, "those": {}, "with": {}, "without": {},
	"from": {}, "into": {}, "onto": {}, "over": {}, "under": {}, "about": {},
	"there": {}, "their": {}, "them": {}, "then": {}, "than": {}, "they": {},
	"have": {}, "has": {}, "had": {}, "will": {}, "would": {}, "should": {},
	"could": {}, "may": {}, "might": {}, "must": {}, "your": {}, "our": {},
	"its": {}, "it's": {}, "it": {}, "is": {}, "as": {}, "at": {}, "by": {}, "in": {},
	"of": {}, "on": {}, "or": {}, "to": {}, "be": {}, "do": {}, "if": {}, "so": {},
	"we": {}, "us": {}, "me": {}, "my": {}, "he": {}, "she": {}, "his": {}, "her": {},
	// Question scaffolding that is never the term worth searching for.
	"explain": {}, "describe": {}, "tell": {}, "give": {}, "used": {}, "use": {},
	"using": {}, "work": {}, "works": {}, "working": {}, "mean": {}, "means": {},
	"help": {}, "please": {}, "between": {}, "difference": {}, "versus": {}, "vs": {},
}

// salientTerms extracts the terms worth a keyword search from a question.
//
// The whitelist is the safety property, not a detail: only [a-z0-9_] survive, so the
// returned terms can be joined into a tsquery with ' | ' without any possibility of
// the question injecting tsquery syntax (&, |, !, parentheses, ':'). Building the
// query from user text with a general-purpose tokenizer is the injection risk that
// this avoids by construction.
//
// Pure — unit-tested.
func salientTerms(question string) []string {
	fields := strings.FieldsFunc(strings.ToLower(question), func(r rune) bool {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		return !isLower && !isDigit && r != '_'
	})

	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 2 {
			continue
		}
		if _, stop := keywordStopwords[field]; stop {
			continue
		}
		if isAllDigits(field) {
			// A bare number matches half the corpus and identifies nothing.
			continue
		}
		if _, dup := seen[field]; dup {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
		if len(out) >= maxKeywordTerms {
			break
		}
	}
	return out
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// reciprocalRankFusion merges ranked id lists into one order.
//
// Each id scores the sum of 1/(k + rank) over the lists it appears in, so an id near
// the top of either list is promoted, and an id present in both beats an id present
// in one — which is the whole point: agreement between two independent retrievers is
// evidence, and rank is comparable across retrievers in a way that raw scores are not.
//
// The tie-break is by id, deliberately: without it, ids with equal scores would come
// out in map iteration order and the same query could rank differently between runs,
// which would break the very comparison this exists to support.
//
// Pure — unit-tested.
func reciprocalRankFusion(lists [][]uuid.UUID) []uuid.UUID {
	scores := make(map[uuid.UUID]float64)
	for _, list := range lists {
		for rank, id := range list {
			if id == uuid.Nil {
				continue
			}
			scores[id] += 1.0 / (rrfK + float64(rank+1))
		}
	}

	ids := make([]uuid.UUID, 0, len(scores))
	for id := range scores {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		si, sj := scores[ids[i]], scores[ids[j]]
		if si != sj {
			return si > sj
		}
		return ids[i].String() < ids[j].String()
	})
	return ids
}
