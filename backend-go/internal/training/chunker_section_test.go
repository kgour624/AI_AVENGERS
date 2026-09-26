package training

import (
	"strings"
	"testing"
)

// A transcript with no headings is the common case, and it must behave exactly
// as it did before sections existed: one section, no path, same chunking rules.
func TestChunkWithoutHeadingsKeepsOldBehaviour(t *testing.T) {
	body := strings.Repeat("This is a sentence about distributed systems and caching. ", 80)
	c := NewTextChunker(DefaultChunkerConfig())

	chunks := c.Chunk(body)
	if len(chunks) == 0 {
		t.Fatal("expected chunks for a long plain transcript")
	}
	for i, ch := range chunks {
		if ch.SectionPath != "" {
			t.Fatalf("chunk %d has section path %q, want empty for a document without headings", i, ch.SectionPath)
		}
		if ch.ChunkHash == "" {
			t.Fatalf("chunk %d has no hash", i)
		}
		if ch.Index != i {
			t.Fatalf("chunk %d has index %d, want %d", i, ch.Index, i)
		}
	}
}

// Every chunk must carry the heading trail it came from, and a chunk must never
// be filed under a section it does not belong to.
func TestChunkAssignsSectionPaths(t *testing.T) {
	doc := "# RAG\n\n" +
		strings.Repeat("Retrieval augmented generation basics here. ", 40) + "\n\n" +
		"## Chunking\n\n" +
		strings.Repeat("Chunk sizes change retrieval quality a lot. ", 40) + "\n\n" +
		"### Overlap\n\n" +
		strings.Repeat("Overlap keeps a concept retrievable from both sides. ", 40) + "\n\n" +
		"# Safety\n\n" +
		strings.Repeat("Prompt injection is an input problem first. ", 40)

	c := NewTextChunker(DefaultChunkerConfig())
	chunks := c.Chunk(doc)
	if len(chunks) < 4 {
		t.Fatalf("expected several chunks, got %d", len(chunks))
	}

	wantPaths := map[string]bool{
		"RAG":              false,
		"RAG > Chunking":   false,
		"RAG > Chunking > Overlap": false,
		"Safety":           false,
	}
	for i, ch := range chunks {
		if ch.SectionPath == "" {
			t.Fatalf("chunk %d has no section path", i)
		}
		if _, ok := wantPaths[ch.SectionPath]; !ok {
			t.Fatalf("chunk %d has unexpected section path %q", i, ch.SectionPath)
		}
		wantPaths[ch.SectionPath] = true
	}
	for path, seen := range wantPaths {
		if !seen {
			t.Fatalf("no chunk came from section %q", path)
		}
	}

	// The chunk that opens a section carries its heading, so a retrieval hit is
	// recognisable and the embedding sees the words the section is about.
	first := chunks[0]
	if !strings.Contains(first.Text, "# RAG") {
		t.Fatalf("first chunk should contain its heading, got %q", first.Text[:min(80, len(first.Text))])
	}
}

// Chunks from one section all share that section's path, even when the section
// is long enough to split several times.
func TestLongSectionKeepsOnePath(t *testing.T) {
	doc := "## Only Section\n\n" + strings.Repeat("A long explanation of one topic. ", 200)
	c := NewTextChunker(DefaultChunkerConfig())
	chunks := c.Chunk(doc)
	if len(chunks) < 2 {
		t.Fatalf("expected the section to split, got %d chunk(s)", len(chunks))
	}
	for i, ch := range chunks {
		if ch.SectionPath != "Only Section" {
			t.Fatalf("chunk %d has section path %q, want %q", i, ch.SectionPath, "Only Section")
		}
	}
}

// A heading with no content under it is a table-of-contents line, not content.
func TestBareHeadingsProduceNoChunks(t *testing.T) {
	sections := splitSections("# Contents\n\n## Intro\n\n## Setup\n")
	if len(sections) != 1 {
		t.Fatalf("expected the fallback single section, got %d", len(sections))
	}
	if strings.TrimSpace(sections[0].body) == "" {
		t.Fatal("fallback section must carry the text, never nothing")
	}
}

// Bare headings must not cost the document its real content: text that sits
// outside them is still chunked, and the fallback keeps a document whose headings
// are all bare from being dropped by the sectioning step.
//
// Note what is NOT claimed here: a document of only short heading lines produces
// no chunks, because the size gate (MinSize) drops content too small to retrieve.
// That rule predates sections, applies to every document, and is not this
// change's business — asserting otherwise would pin a guarantee the chunker has
// never had.
func TestBareHeadingsDoNotLoseBodyText(t *testing.T) {
	doc := "# Contents\n\n## Intro\n\n" +
		strings.Repeat("Real course content that must survive. ", 60) +
		"\n\n## Setup\n\n"

	c := NewTextChunker(DefaultChunkerConfig())
	chunks := c.Chunk(doc)
	if len(chunks) == 0 {
		t.Fatal("content outside the bare headings must still be chunked")
	}
	for i, ch := range chunks {
		if ch.SectionPath != "Contents > Intro" {
			t.Fatalf("chunk %d has section path %q, want %q", i, ch.SectionPath, "Contents > Intro")
		}
	}
}

// Levels that were never named are skipped rather than filled with placeholders.
func TestJoinSectionPathSkipsUnnamedLevels(t *testing.T) {
	if got := joinSectionPath([]string{"A", "", "C"}); got != "A > C" {
		t.Fatalf("joinSectionPath = %q, want %q", got, "A > C")
	}
	if got := joinSectionPath([]string{"", "", "C"}); got != "C" {
		t.Fatalf("joinSectionPath = %q, want %q", got, "C")
	}
	if got := joinSectionPath(nil); got != "" {
		t.Fatalf("joinSectionPath(nil) = %q, want empty", got)
	}
}

// A heading level that jumps (h1 straight to h3) must not resurrect a stale
// title from the previous branch of the document.
func TestHeadingLevelJumpDoesNotKeepStaleTitles(t *testing.T) {
	sections := splitSections("# One\n\nbody one\n\n### Deep\n\nbody deep\n")
	var paths []string
	for _, s := range sections {
		paths = append(paths, s.path)
	}
	found := false
	for _, p := range paths {
		if p == "One > Deep" {
			found = true
		}
		if strings.Contains(p, " >  > ") {
			t.Fatalf("path keeps an unnamed level: %q", p)
		}
	}
	if !found {
		t.Fatalf("expected %q among %v", "One > Deep", paths)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
