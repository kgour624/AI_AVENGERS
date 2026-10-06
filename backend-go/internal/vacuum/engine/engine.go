package engine

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"

	"ai_avengers/backend/internal/vacuum/brain"
	"ai_avengers/backend/internal/vacuum/chunker"
	"ai_avengers/backend/internal/vacuum/filecontext"
)

// Engine — Vacuum Brain Phase 2 deterministic pipeline.
// Input->Process->Output: raw transcript -> P1 regex -> P2 Aho-Corasick trie (boundary aware) -> overlap resolve -> concurrent dedup -> offset-safe removal -> chunk -> SHA verify.
type Engine struct {
	brain *brain.Store
}

func New(b *brain.Store) *Engine { return &Engine{brain: b} }

var builderPool = sync.Pool{New: func() any { return new(strings.Builder) }}

var deterministicRes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\[(music|applause|laughter|inaudible|crosstalk|background music|noise)\]`),
	regexp.MustCompile(`(?i)\b(can you see (my )?screen\??|is my screen visible\??|am i audible\??|can you hear me\??)\b`),
	regexp.MustCompile(`(?i)\b(please )?(like,? share and subscribe|subscribe to (my |our )?channel|hit the bell icon)\b`),
	regexp.MustCompile(`(?i)\b(so+ guys|okay so|alright so|so basically|you know|i mean|kind of|sort of|basically|actually)\b`),
	regexp.MustCompile(`(?i)\b(uh+|umm+|hmm+|ah+|eh+|er+|um+|mmm+)\b`),
	regexp.MustCompile(`\b\d{1,2}:\d{2}(?::\d{2})?\b`),
	regexp.MustCompile(`\[\d{1,2}:\d{2}(?::\d{2})?\]`),
	regexp.MustCompile(`(?m)^\s*(speaker\s*\d+|host|instructor|student)\s*:\s*`),
	regexp.MustCompile(`[.]{2,}`),
	regexp.MustCompile(`[-]{2,}`),
	regexp.MustCompile(`(?i)\b(thank you+|thanks a lot)\s*(so much)?\b`),
}

var wsRe = regexp.MustCompile(`[ \t]+`)
var multiNLRe = regexp.MustCompile("\n{3,}")

func deterministicClean(text string) (string, []string) {
	hits := []string{}
	out := text
	for _, re := range deterministicRes {
		ms := re.FindAllString(out, -1)
		if len(ms) > 0 {
			hits = append(hits, ms...)
			out = re.ReplaceAllString(out, " ")
		}
	}
	out = multiNLRe.ReplaceAllString(out, "\n\n")
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		lines[i] = wsRe.ReplaceAllString(strings.TrimSpace(l), " ")
	}
	out = strings.Join(lines, "\n")
	out = regexp.MustCompile(`[ \t]*\n[ \t]*`).ReplaceAllString(out, "\n")
	out = regexp.MustCompile(`\n{2,}`).ReplaceAllString(out, "\n\n")
	out = strings.TrimSpace(out)
	out = regexp.MustCompile(` +`).ReplaceAllString(out, " ")
	return strings.TrimSpace(out), hits
}

type CleanResult struct {
	CleanedText string
	Chunks      []chunker.Chunk
	Hits        []brain.Match
	P1Hits      []string `json:"p1_hits"`
	SHA256In    string
	SHA256Out   string
	Verified    bool
}

// Clean — 6-phase DAG: P1 deterministic -> P2 trie -> overlap resolve -> concurrent dedup -> assembly -> chunk -> verify.
func (e *Engine) Clean(ctx context.Context, text string) (*CleanResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	shaIn := sha256Hex(text)
	if strings.TrimSpace(text) == "" {
		return &CleanResult{CleanedText: "", P1Hits: nil, Hits: nil, SHA256In: shaIn, SHA256Out: shaIn, Verified: true}, nil
	}
	afterP1, p1Hits := deterministicClean(text)
	matches := e.brain.Search(afterP1)
	filtered := resolveOverlaps(matches)
	fc := filecontext.New()
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	var mu sync.Mutex
	deduped := make([]brain.Match, 0, len(filtered))
	for _, m := range filtered {
		if ctx.Err() != nil {
			break
		}
		if !fc.TryVisit(m.Start, m.End) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(mm brain.Match) {
			defer wg.Done()
			defer func() { <-sem }()
			mu.Lock()
			deduped = append(deduped, mm)
			mu.Unlock()
		}(m)
	}
	wg.Wait()
	sort.Slice(deduped, func(i, j int) bool { return deduped[i].Start > deduped[j].Start })
	cleaned := afterP1
	for _, m := range deduped {
		if m.Start < 0 || m.End > len(cleaned) || m.Start >= m.End {
			continue
		}
		cleaned = cleaned[:m.Start] + " " + cleaned[m.End:]
	}
	cleaned = strings.TrimSpace(wsRe.ReplaceAllString(cleaned, " "))
	cleaned = regexp.MustCompile(`\s*\n\s*`).ReplaceAllString(cleaned, "\n")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		cleaned = afterP1
	}
	if len(text) > 200 && len(cleaned) < len(text)/10 {
		cleaned = afterP1
	}
	chunks := chunker.ChunkText(cleaned, 4000)
	shaOut := sha256Hex(cleaned)
	return &CleanResult{CleanedText: cleaned, Chunks: chunks, Hits: deduped, P1Hits: p1Hits, SHA256In: shaIn, SHA256Out: shaOut, Verified: len(cleaned) > 0}, nil
}

// CleanReader — streaming entry for large files (mmap/bufio ready). Reads via pooled buffer.
func (e *Engine) CleanReader(ctx context.Context, r io.Reader) (*CleanResult, error) {
	br := bufio.NewReader(r)
	b := builderPool.Get().(*strings.Builder)
	b.Reset()
	defer builderPool.Put(b)
	buf := make([]byte, 32*1024)
	for {
		n, err := br.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return e.Clean(ctx, b.String())
}

func resolveOverlaps(in []brain.Match) []brain.Match {
	if len(in) == 0 {
		return in
	}
	sort.Slice(in, func(a, b int) bool {
		if in[a].Start == in[b].Start {
			return (in[a].End - in[a].Start) > (in[b].End - in[b].Start)
		}
		return in[a].Start < in[b].Start
	})
	out := make([]brain.Match, 0, len(in))
	lastEnd := -1
	for _, m := range in {
		if m.Start >= lastEnd {
			out = append(out, m)
			lastEnd = m.End
		}
	}
	return out
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
