package training

import (
	"regexp"
	"strings"
	"unicode"
)

// TranscriptCleaner removes noise from raw course transcripts before ingestion.
//
// WHY cleaning matters:
//   Raw transcripts contain 45-55% noise (greetings, timestamps, quiz mechanics,
//   filler words). This noise gets embedded into chunks and lowers reranker scores
//   below the China Wall threshold (0.35), causing Gate 2 to refuse all questions
//   even when the expert has relevant knowledge.
//
// Noise patterns documented in: TRANSCRIPT_NOISE_TABLE.md
//
// Design:
//   Line-by-line processing with three rule types:
//   1. KEEP   — line contains technical content, always preserve
//   2. REMOVE — line is pure noise, delete entirely
//   3. CLEAN  — line has mixed content, remove noise symbols, keep text
//
//   KEEP overrides REMOVE. If a line has a technical term, it is kept
//   even if it also matches a noise pattern.
type TranscriptCleaner struct {
	// speakerLine matches "Name (HH:MM:SS)" attribution lines
	// e.g. "Sneha Mehra (00:04:33)" or "Sneha Mehra (00:04:33)  "
	speakerLine *regexp.Regexp

	// percentResult matches quiz result lines like "88 % people got it right"
	percentResult *regexp.Regexp

	// trailingBackslash matches OCR artifacts like "5\\.", "n minus 1\\."
	trailingBackslash *regexp.Regexp

	// hesitationMark is the ⁓ symbol used in transcripts for speech pauses
	hesitationMark string
}

// NewTranscriptCleaner creates a new cleaner.
func NewTranscriptCleaner() *TranscriptCleaner {
	return &TranscriptCleaner{
		speakerLine:       regexp.MustCompile(`^[A-Za-z\s]+\(\d{2}:\d{2}:\d{2}\)\s*$`),
		percentResult:     regexp.MustCompile(`\d+\s*%\s*people`),
		trailingBackslash: regexp.MustCompile(`(\d+)\\+\.?`),
		hesitationMark:    "\u2053", // ⁓ character
	}
}

// Clean processes a raw transcript and returns cleaned text.
//
// Mental execution:
//   Input:  "Sneha Mehra (00:04:33)\nHello everyone\nArray indexing starts from 0"
//   Line 1: speakerLine match → REMOVE
//   Line 2: greeting match → REMOVE
//   Line 3: contains "array", "index" → KEEP
//   Output: "Array indexing starts from 0"
func (c *TranscriptCleaner) Clean(transcript string) string {
	lines := strings.Split(transcript, "\n")
	cleaned := make([]string, 0, len(lines))

	for _, line := range lines {
		result := c.processLine(line)
		if result != "" {
			cleaned = append(cleaned, result)
		}
	}

	// Join and collapse excessive blank lines (max 2 consecutive)
	text := strings.Join(cleaned, "\n")
	text = collapseBlankLines(text)

	return strings.TrimSpace(text)
}

// processLine applies cleaning rules to a single line.
// Returns empty string if line should be removed entirely.
func (c *TranscriptCleaner) processLine(line string) string {
	// Step 1: Basic cleanup
	line = strings.TrimRight(line, " \t")

	// Step 2: Remove separator artifacts (+| or |)
	if isSeparatorLine(line) {
		return ""
	}

	// Step 3: Remove speaker attribution lines ("Sneha Mehra (00:04:33)")
	if c.speakerLine.MatchString(line) {
		return ""
	}

	// Step 4: Clean hesitation marks (⁓) — remove symbol, keep text
	line = strings.ReplaceAll(line, c.hesitationMark+" ", "")
	line = strings.ReplaceAll(line, c.hesitationMark, "")
	line = strings.TrimSpace(line)

	// Step 5: Fix OCR artifacts (trailing backslashes on numbers)
	line = c.trailingBackslash.ReplaceAllString(line, "$1")

	// Step 6: If line is now empty or only whitespace, remove
	if strings.TrimSpace(line) == "" {
		return ""
	}

	// Step 7: KEEP check — if line has technical content, always keep
	// This overrides all REMOVE rules below.
	if containsTechnicalContent(line) {
		return line
	}

	// Step 8: REMOVE checks — pure noise lines
	if isNoiseLine(line) {
		return ""
	}

	// Step 9: Short line check — lines < 25 chars with no technical terms
	// are almost always noise ("Okay.", "Yes.", "Sure.", "Nice.", "you")
	if len(strings.TrimSpace(line)) < 25 && !containsTechnicalContent(line) {
		return ""
	}

	return line
}

// containsTechnicalContent returns true if the line contains DSA/programming terms.
// Lines with technical content are ALWAYS kept, regardless of other rules.
// WHY: A line like "O(n) time complexity" must never be removed even if it
// also contains a noise pattern.
func containsTechnicalContent(line string) bool {
	lower := strings.ToLower(line)

	for _, term := range technicalTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

// isNoiseLine returns true if the line is pure noise with no DSA value.
func isNoiseLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))

	for _, pattern := range noisePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// isSeparatorLine returns true for transcript separator artifacts.
func isSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "+|" || trimmed == "|" || trimmed == "+" ||
		trimmed == "---" || trimmed == "---+" ||
		(len(trimmed) <= 3 && strings.ContainsAny(trimmed, "+|"))
}

// collapseBlankLines reduces consecutive blank lines to maximum 2.
func collapseBlankLines(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	blankCount := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 {
				result = append(result, line)
			}
		} else {
			blankCount = 0
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// isAllNonLetter returns true if line has no actual letters (only symbols/numbers).
func isAllNonLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// ============================================================
// NOISE PATTERNS (from TRANSCRIPT_NOISE_TABLE.md)
// ============================================================

// noisePatterns are substrings that indicate a line is pure noise.
// Sourced from TRANSCRIPT_NOISE_TABLE.md — update both files together.
// All lowercase — matched against lowercased line.
var noisePatterns = []string{
	// Category 2: Filler greetings
	"hello everyone",
	"good evening",
	"good morning",
	"good afternoon",
	"hello, hello",
	"thank you, brahma",
	"thank you everyone",
	"all right, i'm back",
	"just give me one minute",
	"i'll just set the ipad",
	"am i audible",
	"is my voice clear",
	"it is slow",
	"yes, no, yes",

	// Category 4: Administrative/logistics
	"ending on 14th",
	"ending on 9th",
	"team's call",
	"we'll try to have a break",
	"advanced content will start",
	"old curriculum",
	"classes were created some extra",
	"last week of the essay",
	"next module that is final",
	"i am not sure when the advanced",
	"don't worry about the future",
	"don't worry about anything",
	"let's enjoy the present",
	"future will take care",
	"future we will take care",

	// Category 5: Student interaction
	"please answer the quiz",
	"take your time then click",
	"% people got it right",
	"top 10 winners",
	"let's take shiva as a volunteer",
	"everyone will get a chance",
	"please give a thumbs up",
	"answer it privately",
	"use question tab",
	"3, 2, 1, go",
	"just click yes",
	"click on the answer",
	"in case you clicked on a wrong answer",

	// Category 7: Language disclaimers
	"i cannot go specific in any programming language",
	"i'm writing pseudo code",
	"pseudo words",

	// Category 8: Motivational/off-topic
	"shah rukh khan",
	"mehmuna",
	"big journey is about to be over",
	"i am here, then why are you worrying",
	"i am there now",

	// Category 9: Standalone section headers (handled by short-line check)
}

// technicalTerms are substrings that indicate DSA/programming content.
// Lines containing these are ALWAYS kept.
// Sourced from TRANSCRIPT_NOISE_TABLE.md — update both files together.
var technicalTerms = []string{
	// Complexity
	"o(n)", "o(log", "o(1)", "o(n^2)", "o(n log", "o(n*",
	"time complexity", "space complexity", "n log n", "log n",

	// Data structures
	"array", "hash", "hashing", "hash table", "hash map", "hashmap",
	"linked list", "node", "stack", "queue", "deque", "priority queue",
	"heap", "min heap", "max heap", "tree", "binary tree", "bst",
	"binary search tree", "graph", "matrix", "2d array",
	"trie", "segment tree", "fenwick",

	// Algorithms
	"dfs", "bfs", "depth first", "breadth first",
	"merge sort", "quick sort", "bubble sort", "insertion sort", "selection sort",
	"binary search", "linear search",
	"recursion", "recursive", "base case",
	"dynamic programming", " dp ", "memoization", "tabulation",
	"backtracking",
	"union find", "dsu", "disjoint set",
	"two pointer", "sliding window",
	"subarray", "subsequence",
	"prefix sum", "suffix",
	"topological sort", "topological",
	"dijkstra", "bellman", "floyd",
	"mst", "kruskal", "prim",
	"greedy",
	"bit manipulation", "xor", "bitwise",

	// Programming concepts
	"for loop", "while loop", "for(", "while(",
	"function", "method", "return",
	"index", "indexing", "pointer",
	"null", "nullptr",
	"int ", "void ", "string ",
	"algorithm", "data structure",
	"swap", "traverse", "iteration",
	"sorted", "unsorted",
	"left child", "right child", "parent",
	"inorder", "preorder", "postorder",
	"connected component", "cycle", "path",
	"hamming", "distance",
	"frequency", "count", "occurrence",
	"minimum", "maximum", "optimal",
	"n minus", "n plus", "n squared",
	"ascii", "character",
	"string reversal", "palindrome",
	"stack overflow",
}
