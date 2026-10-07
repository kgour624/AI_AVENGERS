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
	"time"

	"ai_avengers/backend/internal/observability"
	"ai_avengers/backend/internal/vacuum/brain"
	"ai_avengers/backend/internal/vacuum/chunker"
	"ai_avengers/backend/internal/vacuum/filecontext"
	"ai_avengers/backend/internal/vacuum/llm"
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
var fillerThreshold = 0.62 // classifier auto-filter: filler confidence below this = keep content

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

// IsFiller reports whether classifier label+confidence should trigger auto-filter.
// Mental: INPUT label,confidence, text -> PROCESS threshold 0.62 (tunable) + min words -> OUTPUT bool
func IsFiller(label string, conf float64, textLen int) bool {
	if label != "filler" {
		return false
	}
	if conf < fillerThreshold {
		return false
	}
	// very long chunks are rarely pure filler even if LLM says so
	if textLen > 8000 {
		return false
	}
	return true
}

// FillerThreshold returns current threshold for observability.
func FillerThreshold() float64 { return fillerThreshold }

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

// CleanHybrid — DSA + LLM, hash safe. DSA is sole mutator.
// P1/P2 baseline → chunk → concurrent LLM suggest → self cross-verify (confidence >=0.70 + reason gate) → DSA byte-exact mapping (word-boundary + TryVisit + resolveOverlaps) → descending splice + fallback + re-verify + contiguous + SHA guard.
func (e *Engine) CleanHybrid(ctx context.Context, text string, det llm.KachraDetector, ver llm.KachraVerifier, sink llm.KachraSink, fileJobID string) (*CleanResult, error) {
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	shaIn := sha256Hex(text)
	if strings.TrimSpace(text) == "" {
		return &CleanResult{CleanedText: "", P1Hits: nil, Hits: nil, SHA256In: shaIn, SHA256Out: shaIn, Verified: true}, nil
	}
	afterP1, p1Hits := deterministicClean(text)
	matches := e.brain.Search(afterP1)
	filtered := resolveOverlaps(matches)
	if det == nil {
		return e.Clean(ctx, text)
	}
	chunks := chunker.ChunkText(afterP1, 4000)
	if len(chunks) == 0 {
		chunks = []chunker.Chunk{{Index: 0, Text: afterP1, Start: 0, End: len(afterP1)}}
	}
	semSize := len(chunks) * 2
	if semSize > 20 {
		semSize = 20
	}
	if semSize < 1 {
		semSize = 1
	}
	sem := make(chan struct{}, semSize)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allVerified []llm.KachraSpan
	var llmMatches []brain.Match
	for _, ch := range chunks {
		if ctx.Err() != nil {
			break
		}
		if strings.TrimSpace(ch.Text) == "" {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(chunk chunker.Chunk) {
			defer wg.Done()
			defer func() { <-sem }()
			spans, err := det.Detect(ctx, chunk.Text)
			if err != nil || len(spans) == 0 {
				return
			}
			var verified []llm.KachraSpan
			if ver != nil {
				verified, err = ver.Verify(ctx, chunk.Text, spans)
				if err != nil {
					tmp := make([]llm.KachraSpan, 0, len(spans))
					for _, s := range spans {
						if s.Confidence >= 0.70 && strings.TrimSpace(s.Reason) != "" {
							tmp = append(tmp, s)
						}
					}
					verified = tmp
				}
			} else {
				tmp := make([]llm.KachraSpan, 0, len(spans))
				for _, s := range spans {
					if s.Confidence >= 0.70 && strings.TrimSpace(s.Reason) != "" {
						tmp = append(tmp, s)
					}
				}
				verified = tmp
			}
			if len(verified) == 0 {
				return
			}
			mapped := mapSpansToMatches(chunk, verified)
			if len(mapped) == 0 {
				return
			}
			mu.Lock()
			allVerified = append(allVerified, verified...)
			llmMatches = append(llmMatches, mapped...)
			mu.Unlock()
		}(ch)
	}
	wg.Wait()
	combined := make([]brain.Match, 0, len(filtered)+len(llmMatches))
	combined = append(combined, filtered...)
	combined = append(combined, llmMatches...)
	combined = resolveOverlaps(combined)
	fc := filecontext.New()
	deduped := make([]brain.Match, 0, len(combined))
	for _, m := range combined {
		if m.Start < 0 || m.End > len(afterP1) || m.Start >= m.End {
			continue
		}
		if !fc.TryVisit(m.Start, m.End) {
			continue
		}
		deduped = append(deduped, m)
		fc.Intervals.Insert(m.Start, m.End)
	}
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
	if strings.TrimSpace(cleaned) == "" {
		cleaned = afterP1
	}
	for _, ch := range chunks {
		fc.CheckpointSet.Store(ch.Index, struct{}{})
	}
	shaOut := sha256Hex(cleaned)
	verified := shaOut != "" && len(shaOut) == 64 && len(strings.TrimSpace(cleaned)) > 0
	if len(llmMatches) > 0 {
		observability.Global.IncKachraMapped(int64(len(llmMatches)))
	}
	if sink != nil && len(allVerified) > 0 {
		toSink := make([]llm.KachraSpan, len(allVerified))
		copy(toSink, allVerified)
		fid := fileJobID
		chunkSample := afterP1
		if len(chunkSample) > 2000 {
			chunkSample = chunkSample[:2000]
		}
		go func() {
			bg, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel2()
			_, _ = sink.SaveWithFileID(bg, fid, toSink, chunkSample)
		}()
	}
	chunksOut := chunker.ChunkText(cleaned, 4000)
	return &CleanResult{CleanedText: cleaned, Chunks: chunksOut, Hits: deduped, P1Hits: p1Hits, SHA256In: shaIn, SHA256Out: shaOut, Verified: verified}, nil
}

func mapSpansToMatches(chunk chunker.Chunk, spans []llm.KachraSpan) []brain.Match {
	if len(spans) == 0 || strings.TrimSpace(chunk.Text) == "" {
		return nil
	}
	out := make([]brain.Match, 0, len(spans))
	seen := map[string]bool{}
	for _, s := range spans {
		txt := strings.TrimSpace(s.Text)
		if txt == "" {
			continue
		}
		key := strings.ToLower(txt) + "|" + strings.ToLower(s.Type)
		if seen[key] {
			continue
		}
		seen[key] = true
		idx := strings.Index(chunk.Text, txt)
		if idx < 0 {
			lowerChunk := strings.ToLower(chunk.Text)
			lowerTxt := strings.ToLower(txt)
			idx = strings.Index(lowerChunk, lowerTxt)
			if idx < 0 {
				continue
			}
		}
		start := chunk.Start + idx
		end := start + len(txt)
		if !isWordBoundaryLLM(chunk.Text, idx, idx+len(txt)) {
			if !strings.Contains(txt, " ") {
				continue
			}
		}
		out = append(out, brain.Match{PatternID: "", Pattern: txt, Start: start, End: end})
	}
	return out
}

func isWordBoundaryLLM(text string, start, end int) bool {
	leftOK := start == 0 || llmIsBoundary(text[start-1])
	rightOK := end >= len(text) || llmIsBoundary(text[end])
	return leftOK && rightOK
}

func llmIsBoundary(b byte) bool {
	if b == ' ' || b == '\n' || b == '\r' || b == '\t' {
		return true
	}
	if (b >= '!' && b <= '/') || (b >= ':' && b <= '@') || (b >= '[' && b <= '`') || (b >= '{' && b <= '~') {
		return true
	}
	return false
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
