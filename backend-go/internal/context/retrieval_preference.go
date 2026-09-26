package context

import (
	"sort"
	"strings"
)

// RetrievalPreference biases ranking toward part of an expert's material.
//
// WHY a preference and not a filter: the design's pipeline puts a metadata filter
// BEFORE retrieval for the things that must never be seen (permissions, dates).
// Depth and section are not that. One question can genuinely need a beginner
// paragraph and a research-level one, and a client asking about "chunking" is not
// asking to have every other section hidden. Filtering first would remove the
// chunk that answers the question, and the expert would report "not in my
// material" about material it has. So this only ever reorders what the search
// already found.
type RetrievalPreference struct {
	// Section is matched case-insensitively as a substring of the chunk's heading
	// trail, so "chunking" matches a chunk filed under "RAG > Chunking".
	Section string
	// Layer is the chunk's depth layer (1 surface / 2 working / 3 deep).
	// 0 means no layer preference.
	Layer int
	// Boost multiplies the score of a matching chunk. A value <= 1 is replaced by
	// the default: a preference that demotes what it prefers is a bug, not a
	// choice.
	Boost float64
}

// DefaultPreferenceBoost is a nudge, not a takeover: strong enough to lift a
// matching chunk over a near-equal neighbour, small enough that a clearly better
// non-matching chunk still wins.
const DefaultPreferenceBoost = 1.25

// retrievalPoolFactor is how many times the caller's limit is fetched from the
// reranker, so a preference has room to promote something the reranker had placed
// just outside the returned set.
const retrievalPoolFactor = 3

// Active reports whether anything was actually asked for.
func (p RetrievalPreference) Active() bool {
	return strings.TrimSpace(p.Section) != "" || p.Layer > 0
}

// effectiveBoost returns the multiplier to apply, falling back to the default.
func (p RetrievalPreference) effectiveBoost() float64 {
	if p.Boost > 1 {
		return p.Boost
	}
	return DefaultPreferenceBoost
}

// Matches reports whether a chunk satisfies the preference.
//
// When both a section and a layer are given, satisfying EITHER one is enough. A
// preference is a nudge, and requiring both would quietly turn it into the filter
// this type exists to avoid.
func (p RetrievalPreference) Matches(sectionPath string, layer int) bool {
	if !p.Active() {
		return false
	}
	if s := strings.TrimSpace(p.Section); s != "" {
		if strings.Contains(strings.ToLower(sectionPath), strings.ToLower(s)) {
			return true
		}
	}
	if p.Layer > 0 && layer == p.Layer {
		return true
	}
	return false
}

// applyPreferenceBoost returns the order of indices by descending boosted score.
//
// Pure, and unit-tested. It only reorders: every index appears exactly once, so a
// preference can never remove a chunk that a plain search would have returned —
// the property that makes this safe to leave on.
func applyPreferenceBoost(scores []float64, matches []bool, boost float64) []int {
	idx := make([]int, len(scores))
	boosted := make([]float64, len(scores))
	for i := range idx {
		idx[i] = i
		boosted[i] = scores[i]
		if i < len(matches) && matches[i] {
			boosted[i] = scores[i] * boost
		}
	}
	sort.SliceStable(idx, func(a, b int) bool { return boosted[idx[a]] > boosted[idx[b]] })
	return idx
}
