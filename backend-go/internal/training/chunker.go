package training

// Chunker upgraded to recursive splitting strategy.
// Inspired by LangChain RecursiveCharacterTextSplitter (Byte by Byte AI course).
// Priority: \n\n (paragraph) -> \n (line) -> sentence -> word

import (
	"strings"
	"unicode"
)

// TextChunk represents a single chunk of text with metadata.
type TextChunk struct {
	Text       string
	Index      int
	StartChar  int
	EndChar    int
	TokenCount int
}

// ChunkerConfig holds chunking parameters.
type ChunkerConfig struct {
	TargetSize int // Target tokens per chunk
	MinSize    int // Minimum tokens per chunk
	MaxSize    int // Maximum tokens per chunk
	Overlap    int // Overlap tokens between chunks
}

// DefaultChunkerConfig returns production-tuned defaults.
// WHY these values (Byte by Byte AI + Apna Kiro):
// 500-800 tokens: Large enough for context, small enough for precise retrieval.
// 100 token overlap: Prevents losing context at chunk boundaries.
// Course taught: overlap ensures boundary concepts are not cut off.
func DefaultChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		TargetSize: 600,
		MinSize:    500,
		MaxSize:    800,
		Overlap:    100,
	}
}

// TextChunker splits text into overlapping chunks.
// Uses recursive splitting: paragraph -> line -> sentence -> word.
//
// WHY recursive splitting (Byte by Byte AI course):
// Simple sentence splitting cuts mid-concept.
// Recursive approach tries larger boundaries first so semantic
// units (paragraphs, sections) stay together when possible.
// This is exactly how LangChain RecursiveCharacterTextSplitter works.
type TextChunker struct {
	cfg        ChunkerConfig
	separators []string
}

// NewTextChunker creates a new chunker.
func NewTextChunker(cfg ChunkerConfig) *TextChunker {
	return &TextChunker{
		cfg: cfg,
		// Priority: largest semantic unit first
		// WHY this order: paragraph > line > sentence > word
		// Matches LangChain RecursiveCharacterTextSplitter default separators
		separators: []string{"\n\n", "\n", ". ", "! ", "? ", " "},
	}
}

// Chunk splits text into overlapping chunks.
func (c *TextChunker) Chunk(text string) []TextChunk {
	text = cleanText(text)
	if text == "" {
		return nil
	}
	pieces := c.recursiveSplit(text, c.separators)
	if len(pieces) == 0 {
		return nil
	}
	return c.mergeIntoChunks(pieces)
}

// recursiveSplit splits text using the first separator that produces
// pieces within target size. Falls back to next separator if needed.
//
// Mental execution:
// text="A\n\nB\n\nC", separators=["\n\n","\n",". "]
// Try \n\n -> ["A","B","C"] each fits -> done
//
// text="Very long paragraph..."
// Try \n\n -> still too large -> try \n -> still too large
// Try ". " -> ["sentence1","sentence2"] fits -> done
func (c *TextChunker) recursiveSplit(text string, separators []string) []string {
	if len(separators) == 0 {
		return c.splitBySize(text)
	}
	sep := separators[0]
	parts := strings.Split(text, sep)
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if estimateTokens(part) <= c.cfg.MaxSize {
			result = append(result, part)
		} else {
			result = append(result, c.recursiveSplit(part, separators[1:])...)
		}
	}
	return result
}

// mergeIntoChunks combines pieces into chunks with overlap.
//
// WHY overlap (Byte by Byte AI - "form of" example):
// If a concept spans chunk boundary, overlap ensures both chunks
// contain enough context to be retrieved by semantic search.
func (c *TextChunker) mergeIntoChunks(pieces []string) []TextChunk {
	var chunks []TextChunk
	chunkIndex, charPos := 0, 0
	i := 0
	for i < len(pieces) {
		var current []string
		tokenCount, j := 0, i
		for j < len(pieces) {
			pt := estimateTokens(pieces[j])
			if tokenCount+pt > c.cfg.MaxSize && len(current) > 0 {
				break
			}
			current = append(current, pieces[j])
			tokenCount += pt
			j++
			if tokenCount >= c.cfg.TargetSize {
				break
			}
		}
		chunkText := strings.TrimSpace(strings.Join(current, " "))
		if chunkText != "" && tokenCount >= c.cfg.MinSize/2 {
			chunks = append(chunks, TextChunk{
				Text: chunkText, Index: chunkIndex,
				StartChar: charPos, EndChar: charPos + len(chunkText),
				TokenCount: tokenCount,
			})
			chunkIndex++
			charPos += len(chunkText)
		}
		overlapTokens, overlapStart := 0, j-1
		for overlapStart > i && overlapTokens < c.cfg.Overlap {
			overlapTokens += estimateTokens(pieces[overlapStart])
			overlapStart--
		}
		nextStart := overlapStart + 1
		if nextStart <= i {
			nextStart = i + 1
		}
		i = nextStart
	}
	return chunks
}

// splitBySize splits text into pieces of max size (last resort).
func (c *TextChunker) splitBySize(text string) []string {
	words := strings.Fields(text)
	var pieces []string
	var current []string
	tokens := 0
	for _, word := range words {
		wt := estimateTokens(word)
		if tokens+wt > c.cfg.MaxSize && len(current) > 0 {
			pieces = append(pieces, strings.Join(current, " "))
			current, tokens = nil, 0
		}
		current = append(current, word)
		tokens += wt
	}
	if len(current) > 0 {
		pieces = append(pieces, strings.Join(current, " "))
	}
	return pieces
}

// GetStats returns statistics about the chunks.
func (c *TextChunker) GetStats(chunks []TextChunk) map[string]interface{} {
	if len(chunks) == 0 {
		return map[string]interface{}{"total": 0}
	}
	total, minT, maxT := 0, chunks[0].TokenCount, chunks[0].TokenCount
	for _, ch := range chunks {
		total += ch.TokenCount
		if ch.TokenCount < minT {
			minT = ch.TokenCount
		}
		if ch.TokenCount > maxT {
			maxT = ch.TokenCount
		}
	}
	return map[string]interface{}{
		"total": len(chunks), "avg_tokens": total / len(chunks),
		"min_tokens": minT, "max_tokens": maxT,
	}
}

// NewTextChunker creates a new chunker with given config.
func NewTextChunker(cfg ChunkerConfig) *TextChunker {
	return &TextChunker{cfg: cfg}
}

// Chunk splits text into overlapping chunks with sentence boundary preservation.
//
// Mental execution:
// Input: "Sharding distributes data. Consistent hashing ensures even distribution.
//
//	Virtual nodes prevent hotspots."
//
// Step 1: Split into sentences
// Step 2: Accumulate sentences until target size reached
// Step 3: When target reached, save chunk
// Step 4: Start next chunk with overlap (last N tokens from previous chunk)
// Step 5: Repeat until all sentences processed
func (c *TextChunker) Chunk(text string) []TextChunk {
	// Clean text first
	text = cleanText(text)
	if text == "" {
		return nil
	}

	// Split into sentences
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil
	}

	var chunks []TextChunk
	chunkIndex := 0
	charPos := 0

	// Sliding window over sentences
	sentenceStart := 0

	for sentenceStart < len(sentences) {
		// Build current chunk
		var chunkSentences []string
		tokenCount := 0
		sentenceEnd := sentenceStart

		for sentenceEnd < len(sentences) {
			sentenceTokens := estimateTokens(sentences[sentenceEnd])

			// If adding this sentence exceeds max, stop
			// But always include at least one sentence (prevents infinite loop)
			if tokenCount+sentenceTokens > c.cfg.MaxSize && len(chunkSentences) > 0 {
				break
			}

			chunkSentences = append(chunkSentences, sentences[sentenceEnd])
			tokenCount += sentenceTokens
			sentenceEnd++

			// Stop if we've reached target size
			if tokenCount >= c.cfg.TargetSize {
				break
			}
		}

		// Build chunk text
		chunkText := strings.Join(chunkSentences, " ")
		chunkText = strings.TrimSpace(chunkText)

		if chunkText != "" && tokenCount >= c.cfg.MinSize/2 {
			// Accept chunks that are at least half the minimum size
			// WHY half: Last chunk of a transcript may be shorter
			chunks = append(chunks, TextChunk{
				Text:       chunkText,
				Index:      chunkIndex,
				StartChar:  charPos,
				EndChar:    charPos + len(chunkText),
				TokenCount: tokenCount,
			})
			chunkIndex++
			charPos += len(chunkText)
		}

		// Calculate overlap: how many sentences to go back
		// WHY overlap in sentences not tokens:
		// Sentence-level overlap preserves semantic units
		overlapTokens := 0
		overlapStart := sentenceEnd - 1
		for overlapStart > sentenceStart && overlapTokens < c.cfg.Overlap {
			overlapTokens += estimateTokens(sentences[overlapStart])
			overlapStart--
		}

		// Move window forward (but not past overlap)
		nextStart := overlapStart + 1
		if nextStart <= sentenceStart {
			// Safety: always advance at least one sentence
			nextStart = sentenceStart + 1
		}
		sentenceStart = nextStart
	}

	return chunks
}

// GetStats returns statistics about the chunks.
func (c *TextChunker) GetStats(chunks []TextChunk) map[string]interface{} {
	if len(chunks) == 0 {
		return map[string]interface{}{"total": 0}
	}

	totalTokens := 0
	minTokens := chunks[0].TokenCount
	maxTokens := chunks[0].TokenCount

	for _, chunk := range chunks {
		totalTokens += chunk.TokenCount
		if chunk.TokenCount < minTokens {
			minTokens = chunk.TokenCount
		}
		if chunk.TokenCount > maxTokens {
			maxTokens = chunk.TokenCount
		}
	}

	return map[string]interface{}{
		"total":      len(chunks),
		"avg_tokens": totalTokens / len(chunks),
		"min_tokens": minTokens,
		"max_tokens": maxTokens,
	}
}

// splitSentences splits text into sentences.
// Handles edge cases: abbreviations, decimal numbers, ellipsis.
//
// Mental execution:
// Input: "Dr. Smith said 3.14 is pi. Never use SELECT *. Always index."
// Output: ["Dr. Smith said 3.14 is pi.", "Never use SELECT *.", "Always index."]
//
// Edge cases handled:
// - "Dr." "Mr." "Mrs." "Prof." — not sentence ends
// - "3.14" — decimal, not sentence end
// - "..." — ellipsis, not sentence end
// - "e.g." "i.e." — not sentence ends
func splitSentences(text string) []string {
	// Common abbreviations that end with period but are NOT sentence ends
	abbreviations := map[string]bool{
		"dr": true, "mr": true, "mrs": true, "ms": true, "prof": true,
		"sr": true, "jr": true, "vs": true, "etc": true, "e.g": true,
		"i.e": true, "fig": true, "no": true, "vol": true, "pg": true,
	}

	// Split on sentence-ending punctuation followed by whitespace and capital letter
	// or end of string
	sentenceEndPattern := regexp.MustCompile(`([.!?]+)\s+([A-Z"'\(])`)

	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	i := 0

	for i < len(runes) {
		current.WriteRune(runes[i])

		// Check if this could be a sentence end
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			// Look ahead for whitespace + capital
			j := i + 1
			for j < len(runes) && runes[j] == ' ' {
				j++
			}

			if j < len(runes) && unicode.IsUpper(runes[j]) {
				// Check if this is an abbreviation
				word := getLastWord(current.String())
				wordLower := strings.ToLower(strings.TrimRight(word, "."))

				if !abbreviations[wordLower] && !isDecimalNumber(current.String()) {
					// This is a sentence end
					sentence := strings.TrimSpace(current.String())
					if sentence != "" {
						sentences = append(sentences, sentence)
					}
					current.Reset()
				}
			}
		}

		i++
	}

	// Add remaining text as last sentence
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		sentences = append(sentences, remaining)
	}

	// If no sentences found (e.g., no capital letters), split by newlines
	if len(sentences) == 0 {
		_ = sentenceEndPattern // suppress unused warning
		for _, line := range strings.Split(text, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sentences = append(sentences, line)
			}
		}
	}

	return sentences
}

// getLastWord extracts the last word from text.
func getLastWord(text string) string {
	text = strings.TrimRight(text, " ")
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// isDecimalNumber checks if the period is part of a decimal number.
// e.g., "3.14" or "version 2.0"
func isDecimalNumber(text string) bool {
	text = strings.TrimSpace(text)
	if len(text) < 2 {
		return false
	}
	// Check if character before period is a digit
	runes := []rune(text)
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == '.' {
			if i > 0 && unicode.IsDigit(runes[i-1]) {
				return true
			}
			return false
		}
	}
	return false
}

// cleanText normalizes text for chunking.
// Removes excessive whitespace, normalizes line endings.
func cleanText(text string) string {
	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Remove excessive blank lines (more than 2 consecutive)
	excessiveNewlines := regexp.MustCompile(`\n{3,}`)
	text = excessiveNewlines.ReplaceAllString(text, "\n\n")

	// Normalize whitespace within lines
	excessiveSpaces := regexp.MustCompile(`[ \t]+`)
	text = excessiveSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// estimateTokens estimates token count for a string.
// Rule of thumb: ~4 characters per token (GPT tokenizer average).
// WHY estimate not exact: Exact tokenization requires loading a tokenizer.
// 4 chars/token is accurate enough for chunking decisions.
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}
