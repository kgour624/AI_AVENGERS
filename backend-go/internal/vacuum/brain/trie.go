package brain

import (
	"strings"
	"unicode"
)

// Match is a single kachra hit with word-boundary guaranteed.
type Match struct {
	PatternID string
	Pattern   string
	Start     int // byte offset inclusive
	End       int // byte offset exclusive
}

// node is Aho-Corasick trie node.
type node struct {
	children map[rune]int
	fail     int
	out      []int // pattern indices ending here
}

// Trie is immutable after Build. Build is O(total pattern chars).
// Search is O(textLen + numMatches). Thread-safe for concurrent Search.
type Trie struct {
	nodes    []node
	patterns []string // lowercased
	ids      []string // original IDs parallel to patterns
	types    []string // WORD/PHRASE/REGEX parallel
}

// NewTrie creates empty trie.
func NewTrie() *Trie { return &Trie{nodes: []node{{children: make(map[rune]int)}}} }

// Add inserts pattern (already lowercased trimmed). Returns index.
func (t *Trie) Add(pattern, id, ptype string) {
	pat := strings.ToLower(strings.TrimSpace(pattern))
	if pat == "" {
		return
	}
	t.patterns = append(t.patterns, pat)
	t.ids = append(t.ids, id)
	t.types = append(t.types, ptype)
	idx := len(t.patterns) - 1
	cur := 0
	for _, ch := range pat {
		nxt, ok := t.nodes[cur].children[ch]
		if !ok {
			nxt = len(t.nodes)
			t.nodes = append(t.nodes, node{children: make(map[rune]int)})
			t.nodes[cur].children[ch] = nxt
		}
		cur = nxt
	}
	t.nodes[cur].out = append(t.nodes[cur].out, idx)
}

// Build computes failure links BFS. Must call after all Add before Search.
func (t *Trie) Build() {
	queue := []int{}
	for _, nxt := range t.nodes[0].children {
		t.nodes[nxt].fail = 0
		queue = append(queue, nxt)
	}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for ch, nxt := range t.nodes[v].children {
			f := t.nodes[v].fail
			for f != 0 {
				if _, ok := t.nodes[f].children[ch]; ok {
					break
				}
				f = t.nodes[f].fail
			}
			if failNxt, ok := t.nodes[f].children[ch]; ok {
				t.nodes[nxt].fail = failNxt
			} else {
				t.nodes[nxt].fail = 0
			}
			t.nodes[nxt].out = append(t.nodes[nxt].out, t.nodes[t.nodes[nxt].fail].out...)
			queue = append(queue, nxt)
		}
	}
}

// isBoundary returns true if position is word boundary (start/end/space/punct).
func isBoundary(text string, pos int) bool {
	if pos <= 0 || pos >= len(text) {
		return true
	}
	ch := rune(text[pos])
	if ch == 0 {
		return true
	}
	return unicode.IsSpace(ch) || unicode.IsPunct(ch)
}

// isWordBoundary checks both sides of match for WORD patterns.
func isWordBoundary(text string, start, end int) bool {
	leftOK := start == 0 || isBoundary(text, start-1) || unicode.IsSpace(rune(text[start-1])) || unicode.IsPunct(rune(text[start-1]))
	rightOK := end >= len(text) || isBoundary(text, end) || unicode.IsSpace(rune(text[end])) || unicode.IsPunct(rune(text[end]))
	return leftOK && rightOK
}

// Search returns all matches with word-boundary enforcement for WORD type.
// PHRASE matches are substring (intentional multi-word), REGEX not handled here.
func (t *Trie) Search(text string) []Match {
	if len(t.nodes) == 0 || text == "" {
		return nil
	}
	lower := strings.ToLower(text)
	var out []Match
	state := 0
	runes := []rune(lower)
	// Map rune index to byte offset for correct Start/End
	byteOffsets := make([]int, len(runes)+1)
	off := 0
	for i, r := range runes {
		byteOffsets[i] = off
		off += len(string(r))
	}
	byteOffsets[len(runes)] = off

	for i, ch := range runes {
		for state != 0 {
			if _, ok := t.nodes[state].children[ch]; ok {
				break
			}
			state = t.nodes[state].fail
		}
		if nxt, ok := t.nodes[state].children[ch]; ok {
			state = nxt
		} else {
			state = 0
		}
		if len(t.nodes[state].out) == 0 {
			continue
		}
		for _, pi := range t.nodes[state].out {
			pat := t.patterns[pi]
			patRunes := []rune(pat)
			startRune := i - len(patRunes) + 1
			if startRune < 0 {
				continue
			}
			startByte := byteOffsets[startRune]
			endByte := byteOffsets[i+1]
			// Enforce word boundary for WORD type
			if t.types[pi] == "WORD" {
				if !isWordBoundary(text, startByte, endByte) {
					continue
				}
			}
			out = append(out, Match{
				PatternID: t.ids[pi],
				Pattern:   pat,
				Start:     startByte,
				End:       endByte,
			})
		}
	}
	return out
}

// Size returns number of patterns.
func (t *Trie) Size() int { return len(t.patterns) }
