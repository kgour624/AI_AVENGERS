package chunker

import (
	"regexp"
	"strings"
	"unicode"
)

// Chunk is a semantic unit with heading placeholder.
type Chunk struct {
	Index    int
	Text     string
	Start    int // byte offset in original
	End      int
	Headings []string
}

// Very small pronoun set for anaphora check (Improvement #1).
var pronounRe = regexp.MustCompile(`(?i)^\s*(he|she|it|they|this|that|these|those|here|there|him|her|them)\b`)

// semanticCut decides boundary using cheap heuristics:
// - sentence boundary (.,?,!, para break) + length budget
// - pronoun check: if next sentence starts with pronoun, don't cut.
func isPronounStart(s string) bool { return pronounRe.MatchString(strings.TrimSpace(s)) }

func splitSentences(text string) []string {
	// Keep delimiters; split on sentence boundaries
	re := regexp.MustCompile(`[^.!?\n]+[.!?\n]*`)
	parts := re.FindAllString(text, -1)
	if len(parts) == 0 {
		return []string{text}
	}
	return parts
}

// ChunkText produces chunks with maxChars budget and pronoun-aware merging.
// Improvement #1: no cut if next sentence starts with pronoun/anaphora.
func ChunkText(text string, maxChars int) []Chunk {
	if maxChars <= 0 {
		maxChars = 4000
	}
	sents := splitSentences(text)
	var chunks []Chunk
	var cur strings.Builder
	curStart := 0
	curLen := 0
	chunkStart := 0
	byteOff := 0

	flush := func() {
		if cur.Len() == 0 {
			return
		}
		chunks = append(chunks, Chunk{Index: len(chunks), Text: cur.String(), Start: chunkStart, End: byteOff})
		cur.Reset()
		curLen = 0
	}

	for i, s := range sents {
		sLen := len(s)
		nextStartsPronoun := false
		if i+1 < len(sents) {
			nextStartsPronoun = isPronounStart(sents[i+1])
		}
		// Would overflow?
		if curLen+sLen > maxChars && curLen > 0 {
			// If next starts with pronoun, merge instead of cut -> extend budget slightly
			if nextStartsPronoun && curLen+sLen < maxChars+800 {
				// merge, don't flush
			} else {
				flush()
				chunkStart = byteOff
			}
		}
		if cur.Len() == 0 {
			chunkStart = byteOff
		}
		cur.WriteString(s)
		curLen += sLen
		byteOff += sLen

		// Paragraph break forces flush (double newline heuristic)
		if strings.Contains(s, "\n\n") {
			flush()
			chunkStart = byteOff
		}
		_ = curStart
		_ = unicode.IsSpace
	}
	flush()
	if len(chunks) == 0 && text != "" {
		chunks = append(chunks, Chunk{Index: 0, Text: text, Start: 0, End: len(text)})
	}
	return chunks
}
