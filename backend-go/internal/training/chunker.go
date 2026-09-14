package training

// Chunker upgraded to recursive splitting strategy.
// Inspired by LangChain RecursiveCharacterTextSplitter (Byte by Byte AI course).
// Priority: \n\n (paragraph) -> \n (line) -> sentence -> word

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

// TextChunk represents a single chunk of text with metadata.
//
// ChunkHash: SHA-256 (hex) of the normalized chunk text. Used by the
// ingestion pipeline to deduplicate chunks across re-ingestion runs
// (DOMAIN_EXPERT_COLLABORATION_DESIGN.md §6.3). Normalization is:
// lowercase + collapse all whitespace runs to a single space + trim.
// This means trivial whitespace or case differences do NOT produce a
// different hash, but any real content change does.
type TextChunk struct {
	Text       string
	Index      int
	StartChar  int
	EndChar    int
	TokenCount int
	ChunkHash  string
}

// ChunkerConfig holds chunking parameters.
type ChunkerConfig struct {
	TargetSize int // Target tokens per chunk
	MinSize    int // Minimum tokens per chunk
	MaxSize    int // Maximum tokens per chunk
	Overlap    int // Overlap tokens between chunks
}

// DefaultChunkerConfig returns production-tuned defaults.
// WHY these values (updated 2026-09-14 for V4 retraining):
//
// Previous values: Target=600, Min=500, Max=800, Overlap=100
// New values:      Target=450, Min=375, Max=600, Overlap=75
//
// WHY 25% reduction:
//   V3 (Claude Opus, 2654 chunks, 98 topics) refused problems that
//   V2 (DeepSeek, 2022 chunks, 156 topics) solved correctly.
//   Root cause: larger chunks consolidate related concepts into fewer
//   chunks, reducing topic diversity. Gate 2 (coverage check) needs
//   a relevant chunk to exist — fewer, larger chunks = lower hit rate.
//   Smaller chunks = more chunks = more topics = higher Gate 2 pass rate.
//
// WHY 25% not more:
//   Below ~350 tokens chunks lose enough context that the reranker
//   (0.35 threshold) starts failing — chunk too short to carry
//   meaningful semantic signal for the embedding model.
//
// Overlap reduced proportionally (100 -> 75) to maintain the same
// overlap-to-chunk ratio (~17%) as before.
func DefaultChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		TargetSize: 450,
		MinSize:    375,
		MaxSize:    600,
		Overlap:    75,
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

// Chunk splits text into overlapping chunks and computes ChunkHash for each.
func (c *TextChunker) Chunk(text string) []TextChunk {
	text = cleanText(text)
	if text == "" {
		return nil
	}
	pieces := c.recursiveSplit(text, c.separators)
	if len(pieces) == 0 {
		return nil
	}
	chunks := c.mergeIntoChunks(pieces)
	// Populate ChunkHash on every chunk. Keeping this here (rather than
	// making the caller compute it) guarantees no chunk ever leaves the
	// chunker without a hash, which is what the ingestion pipeline relies on.
	for i := range chunks {
		chunks[i].ChunkHash = HashChunkText(chunks[i].Text)
	}
	return chunks
}

// HashChunkText returns the SHA-256 (hex) of normalized chunk text.
// Normalization: lowercase, collapse whitespace runs to single space, trim.
//
// WHY normalize before hashing:
// - Chunking is not perfectly reproducible across chunker config changes
//   or across different runs of the recursive splitter for minor input
//   variations. Whitespace and case differences would silently produce
//   "new" chunks that are semantically identical.
// - Normalization catches those cases while still detecting any real
//   content change (character additions/removals/substitutions).
//
// This is called by the chunker for every produced chunk, and can also
// be called by callers (e.g., ingestion pipeline) to hash arbitrary text
// and check against the DB before inserting.
func HashChunkText(text string) string {
	normalized := normalizeForHash(text)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

var hashWhitespaceRe = regexp.MustCompile(`\s+`)

func normalizeForHash(text string) string {
	text = strings.ToLower(text)
	text = hashWhitespaceRe.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
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

// splitSentences splits text into sentences.
func splitSentences(text string) []string {
	abbreviations := map[string]bool{
		"dr": true, "mr": true, "mrs": true, "ms": true, "prof": true,
		"sr": true, "jr": true, "vs": true, "etc": true, "e.g": true,
		"i.e": true, "fig": true, "no": true, "vol": true, "pg": true,
	}

	sentenceEndPattern := regexp.MustCompile(`([.!?]+)\s+([A-Z"'\(])`)

	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	i := 0

	for i < len(runes) {
		current.WriteRune(runes[i])

		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			j := i + 1
			for j < len(runes) && runes[j] == ' ' {
				j++
			}

			if j < len(runes) && unicode.IsUpper(runes[j]) {
				word := getLastWord(current.String())
				wordLower := strings.ToLower(strings.TrimRight(word, "."))

				if !abbreviations[wordLower] && !isDecimalNumber(current.String()) {
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

	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		sentences = append(sentences, remaining)
	}

	if len(sentences) == 0 {
		_ = sentenceEndPattern
		for _, line := range strings.Split(text, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				sentences = append(sentences, line)
			}
		}
	}

	return sentences
}

func getLastWord(text string) string {
	text = strings.TrimRight(text, " ")
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func isDecimalNumber(text string) bool {
	text = strings.TrimSpace(text)
	if len(text) < 2 {
		return false
	}
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

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	excessiveNewlines := regexp.MustCompile(`\n{3,}`)
	text = excessiveNewlines.ReplaceAllString(text, "\n\n")

	excessiveSpaces := regexp.MustCompile(`[ \t]+`)
	text = excessiveSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// estimateTokens estimates token count for a string.
// Rule of thumb: ~4 characters per token (GPT tokenizer average).
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(text) / 4
}
