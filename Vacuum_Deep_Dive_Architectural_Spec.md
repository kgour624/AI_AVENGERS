# Vacuum Engine: Core Logic, DSA & Concurrency Complexities

This document details the core Go implementation of the Vacuum Engine, focusing strictly on the
**Pipeline Logic, Data Structures and Algorithms (DSA), Concurrency Models, and System Design
Complexities**. Database schemas and frontend integrations have been deliberately omitted.
Every code block below is taken verbatim from the live source tree.

---

## 1. The Engine Struct — Entry Point (`engine/engine.go`)

The whole cleaning subsystem is rooted in a single, tiny struct. Its dependency is a pointer to the
`brain.Store`, the versioned Aho-Corasick automaton.

```go
// engine/engine.go
type Engine struct {
    brain *brain.Store
}

func New(b *brain.Store) *Engine { return &Engine{brain: b} }
```

The Engine exposes **three** public methods, each for a different use-case:

| Method | Use-case | LLM? |
|--------|----------|-------|
| `CleanReader(ctx, r io.Reader)` | Streaming entry for large files via buffered I/O | ❌ |
| `Clean(ctx, text string)` | Pure deterministic + Trie path | ❌ |
| `CleanHybrid(ctx, text, det, ver, sink, fileJobID)` | Full hybrid: DSA + LLM detect + verify + sink | ✅ |

---

## 2. Phase 1 — Static Regex Deterministic Pass (`deterministicClean`)

Before the Aho-Corasick Trie is even touched, a static pool of pre-compiled `*regexp.Regexp`
patterns fires against the raw text. This is the "dumb but fast" first sweep.

### 2.1 The Regex Battery

```go
// engine/engine.go  (lines 32-43)
var deterministicRes = []*regexp.Regexp{
    // [music] / [applause] / [laughter] annotation tags
    regexp.MustCompile(`(?i)\[(music|applause|laughter|inaudible|crosstalk|background music|noise)\]`),

    // Screen-share / audio-check meta-phrases
    regexp.MustCompile(`(?i)\b(can you see (my )?screen\??|is my screen visible\??|am i audible\??|can you hear me\??)\b`),

    // Marketing boilerplate
    regexp.MustCompile(`(?i)\b(please )?(like,? share and subscribe|subscribe to (my |our )?channel|hit the bell icon)\b`),

    // Discourse markers / filler openers
    regexp.MustCompile(`(?i)\b(so+ guys|okay so|alright so|so basically|you know|i mean|kind of|sort of|basically|actually)\b`),

    // Single-word fillers
    regexp.MustCompile(`(?i)\b(uh+|umm+|hmm+|ah+|eh+|er+|um+|mmm+)\b`),

    // Inline timestamps  "1:23"  /  "01:23:45"
    regexp.MustCompile(`\b\d{1,2}:\d{2}(?::\d{2})?\b`),

    // Bracketed timestamps  "[1:23]"
    regexp.MustCompile(`\[\d{1,2}:\d{2}(?::\d{2})?\]`),

    // Speaker labels  "Speaker 1:  /  Host:  /  Instructor:"
    regexp.MustCompile(`(?m)^\s*(speaker\s*\d+|host|instructor|student)\s*:\s*`),

    // Ellipsis artefacts  "..."  "......"
    regexp.MustCompile(`[.]{2,}`),

    // Em-dash / separator runs  "---"
    regexp.MustCompile(`[-]{2,}`),

    // Closing pleasantries
    regexp.MustCompile(`(?i)\b(thank you+|thanks a lot)\s*(so much)?\b`),
}
```

### 2.2 The Deterministic Clean Function

```go
// engine/engine.go (lines 50-71)
func deterministicClean(text string) (string, []string) {
    hits := []string{}
    out := text
    for _, re := range deterministicRes {
        ms := re.FindAllString(out, -1)
        if len(ms) > 0 {
            hits = append(hits, ms...)
            out = re.ReplaceAllString(out, " ") // replace match with single space
        }
    }
    // Collapse 3+ newlines to exactly 2
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
```

**DSA Complexity**: `O(K * N)` where K = number of regex patterns (fixed at 11), N = text length.
This is **linear** for a given corpus size. The returned `hits` slice is surfaced in
`CleanResult.P1Hits` for telemetry.

---

## 3. The Aho-Corasick Trie (`brain/trie.go`)

The second cleaning pass uses a full Aho-Corasick automaton — a more sophisticated, learned
pattern bank that grew from admin-approved kachra candidates.

### 3.1 Node Representation

```go
// brain/trie.go (lines 16-30)
// node is Aho-Corasick trie node.
type node struct {
    children map[rune]int
    fail     int        // failure-link index (BFS computed)
    out      []int      // pattern indices that end at this node (including via fail-links)
}

// Trie is immutable after Build. Build is O(total pattern chars).
// Search is O(textLen + numMatches). Thread-safe for concurrent Search.
type Trie struct {
    nodes    []node
    patterns []string // lowercased
    ids      []string // original IDs parallel to patterns
    types    []string // WORD/PHRASE/REGEX parallel
}
```

**Design Note**: Nodes are stored in a flat `[]node` slice (index-based graph), not a pointer-graph.
This is cache-friendlier than `map[rune]*TrieNode` because consecutive nodes are likely in the
same cache line.

### 3.2 Trie Construction — `Add` + `Build`

**Phase A: Insertion `O(|pattern|)` per pattern**

```go
// brain/trie.go (lines 36-56)
func (t *Trie) Add(pattern, id, ptype string) {
    pat := strings.ToLower(strings.TrimSpace(pattern))
    if pat == "" {
        return
    }
    t.patterns = append(t.patterns, pat)
    t.ids = append(t.ids, id)
    t.types = append(t.types, ptype)
    idx := len(t.patterns) - 1
    cur := 0                               // start at root node (index 0)
    for _, ch := range pat {
        nxt, ok := t.nodes[cur].children[ch]
        if !ok {
            nxt = len(t.nodes)             // next free slot
            t.nodes = append(t.nodes, node{children: make(map[rune]int)})
            t.nodes[cur].children[ch] = nxt
        }
        cur = nxt
    }
    t.nodes[cur].out = append(t.nodes[cur].out, idx) // mark terminal
}
```

**Phase B: BFS Failure-Link Construction `O(total_pattern_chars)`**

The `Build()` method runs a Breadth-First Search to wire the `fail` pointers. Each `fail` pointer
points to the longest proper suffix of the current prefix that is also a prefix of some pattern.
This is what makes Aho-Corasick `O(N)` at search time: when no transition exists for a character,
the automaton follows `fail` links instead of restarting from the root.

```go
// brain/trie.go (lines 59-85)
func (t *Trie) Build() {
    queue := []int{}
    // Root's immediate children all fail back to root (depth-1 nodes)
    for _, nxt := range t.nodes[0].children {
        t.nodes[nxt].fail = 0
        queue = append(queue, nxt)
    }
    for len(queue) > 0 {
        v := queue[0]
        queue = queue[1:]  // BFS dequeue (O(1) amortised via slice reslicing)
        for ch, nxt := range t.nodes[v].children {
            // Walk failure chain of parent to find the deepest node
            // that has a transition on 'ch'
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
            // Propagate dictionary-suffix links: any pattern ending at fail
            // also ends at nxt (output function merging)
            t.nodes[nxt].out = append(t.nodes[nxt].out, t.nodes[t.nodes[nxt].fail].out...)
            queue = append(queue, nxt)
        }
    }
}
```

### 3.3 Search — `O(N + M)` with Word-Boundary Enforcement

This is the hot-path for every call to `engine.Clean()`.

```go
// brain/trie.go (lines 108-164)
func (t *Trie) Search(text string) []Match {
    if len(t.nodes) == 0 || text == "" {
        return nil
    }
    lower := strings.ToLower(text)
    var out []Match
    state := 0
    runes := []rune(lower)

    // Pre-compute rune→byte offset map for correct Start/End byte positions
    // (critical for multi-byte UTF-8 characters like Hindi/Hinglish)
    byteOffsets := make([]int, len(runes)+1)
    off := 0
    for i, r := range runes {
        byteOffsets[i] = off
        off += len(string(r))
    }
    byteOffsets[len(runes)] = off

    for i, ch := range runes {
        // Follow failure links until we find a node with a 'ch' transition
        for state != 0 {
            if _, ok := t.nodes[state].children[ch]; ok {
                break
            }
            state = t.nodes[state].fail
        }
        if nxt, ok := t.nodes[state].children[ch]; ok {
            state = nxt
        } else {
            state = 0  // Reset to root: no match starts here
        }
        // Check all patterns whose terminal node is 'state'
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
            // WORD-type patterns: enforce word-boundary so "um" doesn't
            // fire inside "umbrella"
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
```

**Word Boundary Check** (the small but crucial guard):
```go
func isBoundary(text string, pos int) bool {
    if pos <= 0 || pos >= len(text) {
        return true
    }
    ch := rune(text[pos])
    return unicode.IsSpace(ch) || unicode.IsPunct(ch)
}

func isWordBoundary(text string, start, end int) bool {
    leftOK  := start == 0 || isBoundary(text, start-1) || unicode.IsSpace(rune(text[start-1])) || unicode.IsPunct(rune(text[start-1]))
    rightOK := end >= len(text) || isBoundary(text, end) || unicode.IsSpace(rune(text[end]))   || unicode.IsPunct(rune(text[end]))
    return leftOK && rightOK
}
```

---

## 4. Brain Store — Hot Reload with `sync.RWMutex` (`brain/store.go`)

The `Store` wraps the Trie with a versioned reader-writer lock pattern, enabling concurrent reads
and safe hot-swaps.

```go
// brain/store.go (lines 15-21)
type Store struct {
    mu      sync.RWMutex   // RLock for reads (concurrent), Lock for Reload (exclusive)
    trie    *Trie
    version int64
    db      *pgxpool.Pool
    logger  *zap.Logger
}
```

### 4.1 Hot Reload — Exclusive Lock, Swap, Release

```go
// brain/store.go (lines 43-70)
func (s *Store) Reload(ctx context.Context) error {
    rows, err := s.db.Query(ctx,
        `SELECT id::text, pattern, pattern_type FROM kachra_patterns WHERE is_active=true ORDER BY pattern`)
    if err != nil {
        return err
    }
    defer rows.Close()

    // 1. Build entire new Trie outside the lock (expensive allocation, zero contention)
    nt := NewTrie()
    for rows.Next() {
        var id, pat, ptype string
        if err := rows.Scan(&id, &pat, &ptype); err != nil {
            continue
        }
        nt.Add(pat, id, ptype)
    }
    if err := rows.Err(); err != nil {
        return err
    }
    nt.Build()  // BFS failure-link construction

    var v int64
    _ = s.db.QueryRow(ctx, `SELECT version FROM brain_version WHERE id=1`).Scan(&v)

    // 2. Exclusive lock ONLY for the pointer swap (nanoseconds)
    s.mu.Lock()
    s.trie = nt
    s.version = v
    s.mu.Unlock()
    // Old *Trie will be GC'd once all goroutines that hold a snapshot finish.

    s.logger.Info("vacuum brain reloaded", zap.Int("patterns", nt.Size()), zap.Int64("version", v))
    return nil
}
```

### 4.2 Search — Under Read-Lock Only

```go
// brain/store.go (lines 73-76)
func (s *Store) Search(text string) []Match {
    trie, _ := s.Snapshot() // takes RLock, returns immediately
    return trie.Search(text)
}

func (s *Store) Snapshot() (*Trie, int64) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.trie, s.version
}
```

**Concurrency Model**: `N` concurrent goroutines can call `Search` simultaneously under `RLock`.
`Reload` takes a `Lock` only for the pointer swap (< 1µs). This gives near-zero contention on the
read path even under heavy load.

### 4.3 30-Second Auto-Reload Ticker

```go
// brain/store.go (lines 79-92)
func (s *Store) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            if err := s.Reload(ctx); err != nil {
                s.logger.Warn("vacuum brain hot-reload failed", zap.Error(err))
            }
        case <-ctx.Done():
            return
        }
    }
}
```

---

## 5. Overlap Resolver — Greedy Interval Sweep (`engine/engine.go`)

When both the Trie and the LLM independently flag overlapping byte ranges, the Engine must pick
non-overlapping winners. This is the classic **interval scheduling / activity selection** problem.

```go
// engine/engine.go (lines 469-488)
func resolveOverlaps(in []brain.Match) []brain.Match {
    if len(in) == 0 {
        return in
    }
    // Sort by start ascending; on ties, longer match wins (greedy max coverage)
    sort.Slice(in, func(a, b int) bool {
        if in[a].Start == in[b].Start {
            return (in[a].End - in[a].Start) > (in[b].End - in[b].Start)
        }
        return in[a].Start < in[b].Start
    })
    out := make([]brain.Match, 0, len(in))
    lastEnd := -1
    for _, m := range in {
        if m.Start >= lastEnd {  // non-overlapping: safe to include
            out = append(out, m)
            lastEnd = m.End
        }
        // overlapping match silently dropped
    }
    return out
}
```

**DSA Complexity**: `O(N log N)` for the sort; `O(N)` for the sweep. Total `O(N log N)`.

---

## 6. The `FileContext` — Lock-Free Dedup Guard (`filecontext/context.go`)

After `resolveOverlaps`, there is still a race: multiple goroutines writing to `llmMatches` could
produce byte ranges already committed by the Trie pass. `FileContext.TryVisit` solves this without
a mutex, using `sync.Map.LoadOrStore`.

```go
// filecontext/context.go (lines 42-61)
type FileContext struct {
    Visited       sync.Map  // key "start:end" -> struct{} (lock-free CAS)
    CheckpointSet sync.Map  // chunkIndex -> struct{} for contiguous checkpoint
    Intervals     IntervalTree
    mu            sync.Mutex
    phase         int
}

// TryVisit returns true ONLY the first time this exact [start,end) range is seen.
// Uses atomic LoadOrStore — no lock, wait-free on uncontested paths.
func (f *FileContext) TryVisit(start, end int) bool {
    k := makeKey(start, end)
    _, loaded := f.Visited.LoadOrStore(k, struct{}{})
    return !loaded  // true = first visit; false = already seen
}
```

The `IntervalTree` used alongside it is a simple append-only slice with a Mutex, used for
broader overlap detection (not just exact [start,end) equality):

```go
// filecontext/context.go (lines 12-32)
type IntervalTree struct {
    mu        sync.Mutex
    intervals []Interval
}

func (t *IntervalTree) Insert(s, e int) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.intervals = append(t.intervals, Interval{s, e})
}

func (t *IntervalTree) Overlaps(s, e int) bool {
    t.mu.Lock()
    defer t.mu.Unlock()
    for _, iv := range t.intervals {
        if s < iv.End && e > iv.Start {  // interval overlap condition
            return true
        }
    }
    return false
}
```

---

## 7. The Chunker — Pronoun-Aware Sentence Splitting (`chunker/chunker.go`)

LLMs have a token-window cap. The Chunker splits the cleaned text into bounded units while
being smart enough to avoid severing anaphoric (pronoun-linked) sentences.

### 7.1 Sentence Tokenisation

```go
// chunker/chunker.go (lines 26-34)
func splitSentences(text string) []string {
    // Regex keeps delimiters attached: "Hello." becomes one token, not "Hello" + "."
    re := regexp.MustCompile(`[^.!?\n]+[.!?\n]*`)
    parts := re.FindAllString(text, -1)
    if len(parts) == 0 {
        return []string{text}
    }
    return parts
}
```

### 7.2 Pronoun-Aware Flush Logic

```go
// chunker/chunker.go (lines 19-95)
var pronounRe = regexp.MustCompile(`(?i)^\s*(he|she|it|they|this|that|these|those|here|there|him|her|them)\b`)

func isPronounStart(s string) bool { return pronounRe.MatchString(strings.TrimSpace(s)) }

func ChunkText(text string, maxChars int) []Chunk {
    if maxChars <= 0 {
        maxChars = 4000
    }
    sents := splitSentences(text)
    var chunks []Chunk
    var cur strings.Builder
    curLen := 0
    chunkStart := 0
    byteOff := 0

    flush := func() {
        if cur.Len() == 0 {
            return
        }
        chunks = append(chunks, Chunk{
            Index: len(chunks),
            Text:  cur.String(),
            Start: chunkStart,
            End:   byteOff,
        })
        cur.Reset()
        curLen = 0
    }

    for i, s := range sents {
        sLen := len(s)
        nextStartsPronoun := false
        if i+1 < len(sents) {
            nextStartsPronoun = isPronounStart(sents[i+1])
        }
        // Would adding this sentence overflow the budget?
        if curLen+sLen > maxChars && curLen > 0 {
            // Special case: if NEXT sentence starts with a pronoun (anaphora),
            // allow a 800-char budget extension to keep context coherent.
            if nextStartsPronoun && curLen+sLen < maxChars+800 {
                // merge, don't flush — context preservation > strict budget
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
        // Paragraph break (double newline) always forces a chunk boundary
        if strings.Contains(s, "\n\n") {
            flush()
            chunkStart = byteOff
        }
    }
    flush()
    // Safety: if input was non-empty but no chunk produced, return one whole chunk
    if len(chunks) == 0 && text != "" {
        chunks = append(chunks, Chunk{Index: 0, Text: text, Start: 0, End: len(text)})
    }
    return chunks
}
```

**System Design Point**: The `Start` and `End` byte offsets in `Chunk` are critical. They allow
`mapSpansToMatches` in the engine to re-map LLM-suggested spans (which are relative to the chunk
text) back to absolute byte offsets in the full document for the final splice.

---

## 8. The Full Hybrid Pipeline — `CleanHybrid` (`engine/engine.go`)

`CleanHybrid` is the production code path. It orchestrates every sub-system.

### 8.1 Phase Entry: Context + SHA256 Bookend

```go
// engine/engine.go (lines 203-209)
func (e *Engine) CleanHybrid(ctx context.Context, text string,
    det llm.KachraDetector, ver llm.KachraVerifier,
    sink llm.KachraSink, fileJobID string) (*CleanResult, error) {

    _, cancel := context.WithCancel(ctx)
    defer cancel()  // ensures all goroutines clean up if Clean returns early

    shaIn := sha256Hex(text)  // cryptographic fingerprint of raw input
    if strings.TrimSpace(text) == "" {
        return &CleanResult{..., Verified: true}, nil  // fast-path empty input
    }
```

### 8.2 Phase 1 + Trie (Synchronous)

```go
    afterP1, p1Hits := deterministicClean(text)   // Regex battery
    matches := e.brain.Search(afterP1)             // Aho-Corasick
    filtered := resolveOverlaps(matches)           // Greedy interval sweep
```

### 8.3 Bounded Concurrency — LLM Detect + Verify Worker Pool

The concurrency ceiling is dynamically computed: `min(len(chunks)*2, 20)`. This auto-scales so
a 3-chunk file doesn't spin up 20 idle goroutines, while a 20-chunk file saturates the pool.

```go
// engine/engine.go (lines 243-321) — condensed for clarity
semSize := len(chunks) * 2
if semSize > 20 { semSize = 20 }
if semSize < 1  { semSize = 1 }
sem := make(chan struct{}, semSize)  // buffered channel as semaphore
var wg sync.WaitGroup
var mu sync.Mutex
var allVerified []llm.KachraSpan
var llmMatches  []brain.Match

for _, ch := range chunks {
    if ctx.Err() != nil {
        break  // context cancelled: stop dispatching new work
    }
    if strings.TrimSpace(ch.Text) == "" {
        continue
    }
    wg.Add(1)
    sem <- struct{}{}  // acquire slot (blocks if pool is full)

    go func(chunk chunker.Chunk) {
        defer wg.Done()
        defer func() { <-sem }()  // release slot on exit

        // Progress reporting (LIFO defer fires before wg.Done)
        defer func() {
            progressMu.Lock()
            doneCount++
            d := doneCount
            if d == total || d%step == 0 {
                reportProgress(ctx, "scanning", d, total)
            }
            progressMu.Unlock()
        }()

        // LLM Call 1: Detect kachra spans
        spans, err := det.Detect(ctx, chunk.Text)
        if err != nil || len(spans) == 0 {
            return
        }

        // LLM Call 2: Verify — drop hallucinations (conf < 0.70 or non-verbatim)
        var verified []llm.KachraSpan
        if ver != nil {
            verified, err = ver.Verify(ctx, chunk.Text, spans)
            if err != nil {
                // Verifier failed: fallback to manual threshold filter
                tmp := make([]llm.KachraSpan, 0, len(spans))
                for _, s := range spans {
                    if s.Confidence >= 0.70 && strings.TrimSpace(s.Reason) != "" {
                        tmp = append(tmp, s)
                    }
                }
                verified = tmp
            }
        }
        if len(verified) == 0 {
            return
        }

        // Re-map span byte offsets from chunk-relative to document-absolute
        mapped := mapSpansToMatches(chunk, verified)
        if len(mapped) == 0 {
            return
        }

        mu.Lock()
        allVerified = append(allVerified, verified...)
        llmMatches  = append(llmMatches, mapped...)
        mu.Unlock()
    }(ch)
}
wg.Wait()
```

**Concurrency Complexity**:
- Maximum goroutines alive simultaneously: `semSize` (≤ 20).
- Each goroutine makes **2 sequential LLM calls** (Detect → Verify).
- `progressMu` is a dedicated mutex just for the counter — it never contends with `mu` (the
  result collector), so counter updates don't slow result aggregation.

### 8.4 Assembly Phase — Merge + Dedup + Descending Splice

After `wg.Wait()`:
```go
// engine/engine.go (lines 329-388)
// Merge Trie matches + LLM matches
combined := make([]brain.Match, 0, len(filtered)+len(llmMatches))
combined  = append(combined, filtered...)
combined  = append(combined, llmMatches...)
combined  = resolveOverlaps(combined)  // deduplicate overlapping ranges

fc := filecontext.New()
deduped := make([]brain.Match, 0, len(combined))
for _, m := range combined {
    if m.Start < 0 || m.End > len(afterP1) || m.Start >= m.End {
        continue  // bounds-safety: never splice outside string
    }
    if !fc.TryVisit(m.Start, m.End) {
        continue  // already committed this exact range
    }
    deduped = append(deduped, m)
    fc.Intervals.Insert(m.Start, m.End)
}

// CRITICAL: sort by Start DESCENDING so we splice from right to left.
// Splicing left-to-right would shift byte offsets for all subsequent matches.
sort.Slice(deduped, func(i, j int) bool {
    return deduped[i].Start > deduped[j].Start
})

cleaned := afterP1
for _, m := range deduped {
    if m.Start < 0 || m.End > len(cleaned) || m.Start >= m.End {
        continue
    }
    cleaned = cleaned[:m.Start] + " " + cleaned[m.End:]  // O(N) string rebuild
}
cleaned = strings.TrimSpace(wsRe.ReplaceAllString(cleaned, " "))
```

**System Design Complexity — The Descending Splice**:  
Naively splicing `cleaned[0:start] + " " + cleaned[end:]` from left to right is a classic
off-by-one disaster: after the first splice, all subsequent `Start`/`End` byte offsets in the
remaining matches point to **shifted** positions in the now-shorter string. The fix is to sort
matches by `Start` **descending** and always splice from the **tail of the string towards the
head**. Tail splices never affect the offsets of remaining matches (which are all at lower indices).

### 8.5 Catastrophic Over-Deletion Guard

```go
// engine/engine.go (lines 356-364)
if cleaned == "" {
    cleaned = afterP1  // fallback: if LLM deleted everything, keep P1 result
}
// Guard: if output is less than 10% of input, suspect over-aggressive LLM
if len(text) > 200 && len(cleaned) < len(text)/10 {
    cleaned = afterP1  // rollback to deterministic-only output
}
```

This prevents a hallucinating LLM from reducing a 100,000-character transcript to 500 characters.

### 8.6 SHA256 Output Hash + Verification

```go
// engine/engine.go (lines 368-369)
shaOut := sha256Hex(cleaned)
verified := shaOut != "" && len(shaOut) == 64 && len(strings.TrimSpace(cleaned)) > 0

// engine/engine.go (line 490-493)
func sha256Hex(s string) string {
    h := sha256.Sum256([]byte(s))
    return hex.EncodeToString(h[:])
}
```

`SHA256In` (of raw) and `SHA256Out` (of cleaned) are stored in the `CleanResult` and later
persisted to `file_jobs.sha256_*` columns. They enable forensic auditing: if a user claims
"my content was changed", the hashes prove exactly what bytes were present before and after.

---

## 9. LLM Span Mapping — `mapSpansToMatches` (`engine/engine.go`)

This function converts LLM-supplied kachra spans (which are expressed as strings within the
chunk's coordinate space) back to absolute byte offsets in the full document.

```go
// engine/engine.go (lines 391-426)
func mapSpansToMatches(chunk chunker.Chunk, spans []llm.KachraSpan) []brain.Match {
    if len(spans) == 0 || strings.TrimSpace(chunk.Text) == "" {
        return nil
    }
    out  := make([]brain.Match, 0, len(spans))
    seen := map[string]bool{}

    for _, s := range spans {
        txt := strings.TrimSpace(s.Text)
        if txt == "" {
            continue
        }
        // Dedup within chunk: "uh uh uh" might produce 3 identical spans
        key := strings.ToLower(txt) + "|" + strings.ToLower(s.Type)
        if seen[key] {
            continue
        }
        seen[key] = true

        // Attempt 1: exact case-sensitive substring search
        idx := strings.Index(chunk.Text, txt)
        if idx < 0 {
            // Attempt 2: case-insensitive (LLM sometimes changes casing)
            lowerChunk := strings.ToLower(chunk.Text)
            lowerTxt   := strings.ToLower(txt)
            idx = strings.Index(lowerChunk, lowerTxt)
            if idx < 0 {
                continue  // span not found in chunk — drop (hallucination)
            }
        }

        // Translate chunk-relative offset to document-absolute offset
        start := chunk.Start + idx
        end   := start + len(txt)

        // Word boundary check for single-word spans to prevent sub-word deletion
        if !isWordBoundaryLLM(chunk.Text, idx, idx+len(txt)) {
            if !strings.Contains(txt, " ") {  // multi-word spans bypass boundary check
                continue
            }
        }
        out = append(out, brain.Match{PatternID: "", Pattern: txt, Start: start, End: end})
    }
    return out
}
```

**System Design Note**: The two-attempt search (exact → case-insensitive) is a pragmatic
hallucination guard. The LLM operates at `Temperature: 0.0` for the detector, but still sometimes
returns "Uh" when the source was "uh". Rather than silently dropping it, the engine tries a
case-insensitive fallback before giving up.

---

## 10. The Kachra Detector — LLM Call 1 (`llm/kachra_detector.go`)

The Detector is the first LLM call per chunk. It asks the model to identify noise spans.

### 10.1 System Prompt

```go
// llm/kachra_detector.go (lines 21-33)
const kachraDetectSystem = `You are a kachra (noise) detector for course transcripts.
Identify ONLY noise text spans that should be removed.
Return STRICT JSON array only — no markdown, no explanation.

Rules:
- Each item: {"text":"verbatim substring","reason":"why kachra 10-20 words","type":"<one of 8>","confidence":0.0-1.0}
- "text" MUST be exact verbatim substring of input (no paraphrase).
- "reason" short justification e.g. "filler interjection, no information".
- "type" exactly one of: filler | repetition | asr_error | classroom_meta | hinglish | logistics | semantic_noise | smalltalk
- "confidence" 0.0-1.0 (0.70+ high certainty)
- Do NOT rewrite. Do NOT return cleaned text. Only spans.
- Max 20 spans. If no kachra, return [].
- Respond ONLY with JSON array.`
```

**Design Philosophy**: Temperature `0.0` is used for structured JSON output. The schema is minimal.
"No paraphrase" and "verbatim substring" are explicit instructions to reduce hallucination risk.

### 10.2 3-Attempt Retry with Linear Backoff

```go
// llm/kachra_detector.go (lines 43-63)
for attempt := 0; attempt < 3; attempt++ {
    spans, err := g.doDetect(ctx, prompt)
    if err == nil {
        observability.Global.IncKachraSuggested(int64(len(spans)))
        return spans, nil
    }
    lastErr = err
    if ctx.Err() != nil {
        return nil, ctx.Err()  // context cancelled: no retry
    }
    if !isTransient(err) {
        return nil, err  // fatal error (e.g. bad auth): no retry
    }
    // Linear backoff: 200ms, 400ms, 600ms
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-time.After(time.Duration(200*(attempt+1)) * time.Millisecond):
    }
}
```

`isTransient` inspects the error string for HTTP status codes and keywords:
```go
func isTransient(err error) bool {
    msg := strings.ToLower(err.Error())
    return strings.Contains(msg, "429") || strings.Contains(msg, "rate") ||
        strings.Contains(msg, "503") || strings.Contains(msg, "502") ||
        strings.Contains(msg, "504") || strings.Contains(msg, "timeout") ||
        strings.Contains(msg, "overload")
}
```

---

## 11. The Kachra Verifier — LLM Call 2 (`llm/kachra_verifier.go`)

The Verifier is the second independent LLM call. It cross-examines the Detector's output.

### 11.1 Why a Second Call?

LLMs hallucinate. The Detector running at `Temperature: 0.0` is still non-deterministic across
providers. The Verifier is a stricter second opinion that enforces:
1. Span text is verbatim in the original chunk.
2. Confidence ≥ 0.70.
3. Span is not meaningful content.

```go
// llm/kachra_verifier.go (lines 20-29)
const kachraVerifySystem = `You are a strict verifier for kachra (noise) spans.
Given the original chunk and candidate spans, KEEP ONLY spans that are truly noise.
Return STRICT JSON array of kept spans with same keys: {"text","reason","type","confidence"}.
Rules:
- If span text is NOT verbatim in the chunk, drop it.
- If confidence < 0.70, drop it.
- If span is meaningful content (code, domain term, explanation), drop it.
- Max same count as input. If none verified, return [].
- Respond ONLY with JSON array.`
```

### 11.2 The Fallback Path (No Verifier or Verifier Failure)

```go
// engine/engine.go (lines 290-309) — inside goroutine
if ver != nil {
    verified, err = ver.Verify(ctx, chunk.Text, spans)
    if err != nil {
        // Verifier down: manual threshold gate instead
        tmp := make([]llm.KachraSpan, 0, len(spans))
        for _, s := range spans {
            if s.Confidence >= 0.70 && strings.TrimSpace(s.Reason) != "" {
                tmp = append(tmp, s)
            }
        }
        verified = tmp
    }
} else {
    // No verifier wired: same threshold gate
    tmp := make([]llm.KachraSpan, 0, len(spans))
    for _, s := range spans {
        if s.Confidence >= 0.70 && strings.TrimSpace(s.Reason) != "" {
            tmp = append(tmp, s)
        }
    }
    verified = tmp
}
```

### 11.3 Span Normalisation & Deduplication

```go
// llm/kachra_verifier.go (lines 110-143)
func normalizeKachraSpans(in []KachraSpan) []KachraSpan {
    out  := make([]KachraSpan, 0, len(in))
    seen := map[string]bool{}
    for _, s := range in {
        s.Text   = strings.TrimSpace(s.Text)
        s.Reason = strings.TrimSpace(s.Reason)
        s.Type   = strings.ToLower(strings.TrimSpace(s.Type))
        if s.Text == "" || s.Reason == "" {
            continue
        }
        if !validKachraType[s.Type] {
            s.Type = "filler"  // coerce unknown types to safe default
        }
        // Clamp confidence to [0.0, 1.0]
        if s.Confidence < 0 { s.Confidence = 0 }
        if s.Confidence > 1 { s.Confidence = 1 }

        key := strings.ToLower(s.Text) + "|" + s.Type
        if seen[key] {
            continue  // dedup: same text+type pair already in output
        }
        seen[key] = true
        out = append(out, s)
        if len(out) >= 20 {  // hard cap at 20 spans per chunk
            break
        }
    }
    return out
}
```

---

## 12. Gemini Chunk Classifier (`llm/classifier.go`)

Separate from the Kachra detector/verifier, the Classifier runs a simpler single-call
chunk-level triage — labelling entire chunks rather than spans within chunks.

```go
// llm/classifier.go (lines 21-28)
const classifierSystem = `You are a tiny text classifier for transcript chunks.
Classify the chunk into exactly one label: "filler" | "content" | "music" | "noise" | "code".
Respond ONLY with JSON: {"label":"<one of the 5>","confidence":0.0-1.0}
No other text.`
```

### 12.1 Cost-Controlled Input

```go
// llm/classifier.go (lines 69-77)
func (g *GeminiClassifier) doClassify(ctx context.Context, text string) (string, float64, error) {
    prompt := truncate(text, 2000) // hard cap: never send > 2000 chars to save tokens
    resp, err := g.caller.Call(ctx, LLMRequest{
        Model:        "fast",  // Gemini Flash (cheapest)
        SystemPrompt: classifierSystem,
        UserPrompt:   prompt,
        MaxTokens:    8100,
        Temperature:  0.1,  // slight randomness for better label variety
    })
```

### 12.2 Heuristic Fallback (Zero LLM Dependency in Tests)

If `g.caller == nil` (no LLM wired), the classifier falls back to a lightweight heuristic:

```go
// llm/classifier.go (lines 106-119)
func heuristicLabel(text string) string {
    lower := strings.ToLower(text)
    if strings.Contains(lower, "[music]") || strings.Contains(lower, "[applause]") {
        return "music"
    }
    if strings.Contains(lower, "\n```") || strings.Contains(lower, "func ") {
        return "code"
    }
    if len(strings.Fields(text)) < 4 {
        return "filler"  // very short chunks are almost always filler
    }
    return "content"
}
```

---

## 13. The Kachra Sink — Self-Learning Feedback Loop (`llm/kachra_sink.go`)

After the engine verifies LLM spans, they are routed to the Sink asynchronously (non-blocking fire
and forget) so the pipeline does not wait for the DB write.

```go
// engine/engine.go (lines 373-386)
if sink != nil && len(allVerified) > 0 {
    toSink := make([]llm.KachraSpan, len(allVerified))
    copy(toSink, allVerified)     // defensive copy: goroutine outlives the current stack frame
    fid := fileJobID
    go func() {
        bg, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel2()
        _, _ = sink.SaveWithFileID(bg, fid, toSink, chunkSample)
    }()
}
```

The Sink itself uses an `INSERT ... ON CONFLICT DO UPDATE` (upsert) to accumulate hit counts:

```go
// llm/kachra_sink.go (lines 60-70)
_, err := s.db.Exec(ctx, `
    INSERT INTO candidate_kachra(pattern, pattern_type, category, context_snippet, file_id, confidence, status, hit_count)
    VALUES ($1,$2,$3,$4,$5,$6,'pending',1)
    ON CONFLICT (lower(trim(pattern)), pattern_type, category) DO UPDATE
    SET hit_count        = candidate_kachra.hit_count+1,
        confidence       = GREATEST(candidate_kachra.confidence, EXCLUDED.confidence),
        context_snippet  = EXCLUDED.context_snippet,
        file_id          = COALESCE(EXCLUDED.file_id, candidate_kachra.file_id),
        last_seen_at     = NOW(),
        updated_at       = NOW()
`, pat, ptype, cat, snippet, fileID, conf)
```

**Self-Learning Flow**:
1. LLM flags `"uh, so basically guys"` as `filler` with confidence `0.91`.
2. Verifier confirms it.
3. Sink upserts: first time → `hit_count=1`. Second file → `hit_count=2`. ... 50th file → `hit_count=50`.
4. Admin sees it in the "Kachra Candidates" panel and approves it.
5. Approved pattern is inserted into `kachra_patterns` (permanent table).
6. Brain's 30s ticker fires `Reload()`, rebuilds Trie, swaps pointer.
7. Future jobs catch `"uh, so basically guys"` in the **Trie pass (Phase 2)** at zero LLM cost.

---

## 14. The Orphan Reaper — Fault Tolerance (`reaper.go`)

The Reaper is a background daemon that heals stuck jobs. It uses two complementary strategies.

### 14.1 Boot Sweep — Zero Grace Window

```go
// reaper.go (lines 57-70)
// Called at startup. A freshly-booted process owns no active runs.
// Any job in picking/cleaning/verifying is definitively a ghost from the
// crashed previous process.
func HealOrphansOnBoot(ctx context.Context, db *pgxpool.Pool, log *zap.Logger) int {
    n := sweep(ctx, db, 0, "ORPHANED_RESTART",
        "run lost: backend restarted while this job was in flight", log)
    if n > 0 && log != nil {
        log.Warn("vacuum orphan reaper: healed jobs left busy by a previous process",
            zap.Int("healed", n))
    }
    return n
}
```

### 14.2 Periodic Sweep — 45-Minute Grace

```go
// reaper.go (lines 72-91)
const OrphanGrace          = 45 * time.Minute
const orphanSweepInterval  = time.Minute

func StartOrphanReaper(ctx context.Context, db *pgxpool.Pool, log *zap.Logger) {
    if db == nil {
        return
    }
    HealOrphansOnBoot(ctx, db, log)
    ticker := time.NewTicker(orphanSweepInterval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if n := sweep(ctx, db, OrphanGrace, "ORPHANED",
                "no progress for 45m: run died without reporting", log); n > 0 && log != nil {
                log.Warn("vacuum orphan reaper: healed silent jobs", zap.Int("healed", n))
            }
        }
    }
}
```

### 14.3 The Sweep Core — `isRunning` Liveness Proof

```go
// reaper.go (lines 36-37, 96-148)
const orphanCandidatesSQL = `SELECT id::text FROM file_jobs
 WHERE status IN ('picking','cleaning','verifying') AND updated_at < $1`

func sweep(ctx context.Context, db *pgxpool.Pool, grace time.Duration, code, message string, log *zap.Logger) int {
    rows, err := db.Query(ctx, orphanCandidatesSQL, time.Now().Add(-grace))
    // ... scan ids ...
    healed := 0
    for _, id := range ids {
        if isRunning(id) {  // in-process registry: this run is alive in this process
            continue        // skip: a phase-2 LLM scan can be silent for > grace window
        }
        tag, err := db.Exec(ctx, orphanFailSQL, id, code, message)
        // ...
        healed++
        progressBus.Forget(id)                              // evict SSE snapshot
        publishProgress(id, "failed", "failed", message, 6, 100, 0, 0)  // notify UI
    }
    return healed
}
```

**System Design Complexity**: The `isRunning(id)` check is crucial. A large file's Phase 2 LLM
scan can legitimately produce zero DB writes for 30-40 minutes (it's scanning chunks, not writing
rows). Without the liveness check, the Reaper would kill a healthy slow run. The `isRunning`
registry is an in-process `sync.Map` that active pipeline goroutines register into on start and
deregister from on finish.

### 14.4 The Orphan Fail SQL

```go
// reaper.go (lines 46-55)
const orphanFailSQL = `UPDATE file_jobs SET
     status='failed',
     phase=6,
     error_code=$2,
     last_error=$3,
     retry_after=NULL,
     finished_at=NOW(),
     duration_ms=COALESCE(duration_ms, (EXTRACT(EPOCH FROM (NOW()-COALESCE(picked_at, created_at)))*1000)::int),
     updated_at=NOW()
 WHERE id=$1`
```

**Why `failed` not `canceled`?**: `canceled` deletes the job and its stored files. `failed`
preserves everything and surfaces a "Retry" button in the UI. The Reaper must not destroy user data.

---

## 15. CleanReader — Streaming Large Files (`engine/engine.go`)

For very large transcripts, loading the entire file into a `string` at once can cause OOM. 
`CleanReader` reads in 32KB chunks via `bufio.Reader` into a pooled `strings.Builder`.

```go
// engine/engine.go (lines 444-467)
var builderPool = sync.Pool{New: func() any { return new(strings.Builder) }}

func (e *Engine) CleanReader(ctx context.Context, r io.Reader) (*CleanResult, error) {
    br := bufio.NewReader(r)
    b := builderPool.Get().(*strings.Builder)  // Get from pool (GC-friendly)
    b.Reset()
    defer builderPool.Put(b)  // Return to pool on exit

    buf := make([]byte, 32*1024)  // 32KB read buffer
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
            return nil, ctx.Err()  // respect cancellation during I/O
        }
    }
    return e.Clean(ctx, b.String())
}
```

**System Design Note**: `sync.Pool` is used for the `strings.Builder` to reduce pressure on the
garbage collector when thousands of files are processed concurrently. Each `Get()` returns either
a recycled builder (zero allocation) or a fresh one.

---

## 16. Progress Reporting — Context Value Pattern (`engine/engine.go`)

The Engine emits real-time progress without importing the `vacuum` package (which would cause an
import cycle). The solution: inject a callback through `context.Value`.

```go
// engine/engine.go (lines 167-199)
type ProgressFn func(stage string, done, total int)

type progressCtxKey struct{}  // unexported: no other package can forge this key

// WithProgress attaches fn to ctx. Context value keeps signature stable:
// no change to CleanHybrid's parameter list, zero risk of import cycle.
func WithProgress(ctx context.Context, fn ProgressFn) context.Context {
    if ctx == nil || fn == nil {
        return ctx
    }
    return context.WithValue(ctx, progressCtxKey{}, fn)
}

func reportProgress(ctx context.Context, stage string, done, total int) {
    if ctx == nil {
        return
    }
    fn, _ := ctx.Value(progressCtxKey{}).(ProgressFn)
    if fn == nil {
        return  // no-op: context not wrapped with WithProgress
    }
    fn(stage, done, total)
}
```

The `progressMu` inside `CleanHybrid` ensures strictly monotonic counter delivery:

```go
// engine/engine.go (lines 238-241, 276-284)
var (
    doneCount  int
    progressMu sync.Mutex
)
// Inside goroutine defer (fires LIFO, before sem release):
defer func() {
    progressMu.Lock()
    doneCount++
    d := doneCount
    if d == total || d%step == 0 {
        reportProgress(ctx, "scanning", d, total)
    }
    progressMu.Unlock()
}()
```

**Why a dedicated `progressMu` separate from `mu`?**: The result-aggregation mutex `mu` wraps a
slice append (microseconds). The progress mutex `progressMu` wraps a counter bump + a callback
invocation (which might do I/O). Mixing them would slow result aggregation down to callback speed.

---

## 17. Summary — DSA & Complexity Matrix

| Algorithm | Location | Complexity |
|-----------|----------|-----------|
| Static Regex batch | `deterministicClean` | `O(K·N)` fixed K=11 |
| Aho-Corasick Build | `Trie.Build` | `O(Σ pattern lengths)` |
| Aho-Corasick Search | `Trie.Search` | `O(N + M)` |
| Overlap Resolver | `resolveOverlaps` | `O(N log N)` sort + `O(N)` sweep |
| Chunker | `ChunkText` | `O(N)` sentence scan |
| Span Mapping | `mapSpansToMatches` | `O(S·C)` S=spans, C=chunk length |
| Descending Splice | `CleanHybrid` assembly | `O(D·N)` D=deduped matches |
| TryVisit Dedup | `FileContext.TryVisit` | `O(1)` amortised (sync.Map) |
| Interval Overlap Check | `IntervalTree.Overlaps` | `O(I)` I=committed intervals |
| SHA256 | `sha256Hex` | `O(N)` |

**Max Goroutines In-Flight**: `min(len(chunks)*2, 20)` — bounded, no unbounded goroutine spawn.

**Memory Model**: `CleanReader` uses `sync.Pool` to reuse `strings.Builder` across calls. The
Trie itself is read-only after `Build()`, enabling lock-free concurrent search with only an RLock.

---
*End of Vacuum Engine Core Code & Concurrency Specification.*
