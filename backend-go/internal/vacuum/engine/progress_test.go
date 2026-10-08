package engine

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"ai_avengers/backend/internal/vacuum/brain"
	"ai_avengers/backend/internal/vacuum/llm"
)

// progressTestDetector is a deterministic stand-in for the Gemini classifier.
// It sleeps so the goroutines genuinely overlap, and it only reports a span for
// chunks containing "unwanted" — which lets a single run exercise both the
// mapped path and the early-return path that used to freeze the counter.
type progressTestDetector struct{ delay time.Duration }

func (d progressTestDetector) Detect(ctx context.Context, chunkText string) ([]llm.KachraSpan, error) {
	if d.delay > 0 {
		select {
		case <-time.After(d.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if !strings.Contains(chunkText, "unwanted") {
		return nil, nil
	}
	return []llm.KachraSpan{{
		Text:       "unwanted",
		Reason:     "marker span used by the progress test",
		Type:       "filler",
		Confidence: 0.95,
	}}, nil
}

// progressTestVerifier accepts every candidate, so the only thing under test is
// the progress plumbing — not the LLM confidence gate.
type progressTestVerifier struct{}

func (progressTestVerifier) Verify(_ context.Context, _ string, spans []llm.KachraSpan) ([]llm.KachraSpan, error) {
	return spans, nil
}

// progressTestText builds roughly 240 KB of low-entropy prose — many sentences,
// so ChunkText can actually cut at its 4000-byte budget.
//
// The periods are load-bearing: chunker.splitSentences only breaks on . ! ? or a
// newline, so a wall of prose with no sentence punctuation comes back as ONE
// sentence and therefore ONE chunk. That is a real property of the chunker, not
// a bug, but it would leave the counter with nothing to count and make this test
// prove nothing. The words are deliberately none of the deterministic P1
// patterns, so afterP1 stays close to the input.
func progressTestText(marked int) string {
	const (
		sentence = "Alpha bravo charlie delta echo foxtrot golf hotel india juliet. "
		withSpan = "Alpha bravo charlie delta echo foxtrot golf hotel unwanted juliet. "
		count    = 4000
	)
	// Spread the marked sentences evenly so most chunks take the no-kachra
	// early-return path and only some take the mapped path — that mix is what
	// makes the "counter still reaches the end" assertion meaningful.
	stride := 0
	if marked > 0 {
		stride = count / marked
	}
	var sb strings.Builder
	sb.Grow(count * len(sentence))
	for i := 0; i < count; i++ {
		if stride > 0 && i%stride == 0 {
			sb.WriteString(withSpan)
			continue
		}
		sb.WriteString(sentence)
	}
	return sb.String()
}

type progressEvent struct {
	stage string
	done  int
	total int
}

// TestCleanHybridEmitsLiveChunkCounter is the regression test for the phase-2
// visibility upgrade: an admin must be able to watch the LLM classifier advance
// chunk by chunk instead of staring at a frozen percentage.
func TestCleanHybridEmitsLiveChunkCounter(t *testing.T) {
	var mu sync.Mutex
	var events []progressEvent
	ctx := WithProgress(context.Background(), func(stage string, done, total int) {
		mu.Lock()
		events = append(events, progressEvent{stage: stage, done: done, total: total})
		mu.Unlock()
	})

	eng := New(brain.NewStore(nil, nil))
	res, err := eng.CleanHybrid(ctx, progressTestText(40), progressTestDetector{delay: 3 * time.Millisecond}, progressTestVerifier{}, nil, "")
	if err != nil {
		t.Fatalf("CleanHybrid returned error: %v", err)
	}
	if res == nil || !res.Verified {
		t.Fatalf("expected a verified CleanResult, got %+v", res)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 {
		t.Fatal("no progress events emitted: the injected callback was never called")
	}

	// 1. The first event opens the counter at zero with the real denominator.
	first := events[0]
	if first.stage != "scanning" || first.done != 0 {
		t.Fatalf("first event = %+v, want stage=scanning done=0", first)
	}
	total := first.total
	if total < 2 {
		t.Fatalf("expected a multi-chunk file so the counter can move, got total=%d", total)
	}

	// 2. Every scanning event keeps the same denominator, and the only other
	//    stage is the single closing "assembling" event.
	var scanEvents, assembleEvents []progressEvent
	for _, e := range events {
		switch e.stage {
		case "scanning":
			if e.total != total {
				t.Fatalf("scanning event %+v changed the total, want %d", e, total)
			}
			scanEvents = append(scanEvents, e)
		case "assembling":
			assembleEvents = append(assembleEvents, e)
		default:
			t.Fatalf("unexpected stage %q", e.stage)
		}
	}

	// 3. Progress is monotonic and never overshoots.
	prev := -1
	for _, e := range scanEvents {
		if e.done < prev {
			t.Fatalf("counter went backwards: %d after %d", e.done, prev)
		}
		if e.done > total {
			t.Fatalf("counter overshot: %d > %d", e.done, total)
		}
		prev = e.done
	}

	// 4. It always reaches the end. This is the assertion that would fail if the
	//    increment lived inline instead of in a defer, because the common
	//    no-kachra path returns before any inline increment could run.
	last := scanEvents[len(scanEvents)-1]
	if last.done != total {
		t.Fatalf("last scanning event = %+v, want done=%d", last, total)
	}

	// 5. The throttle still emits movement rather than a single jump.
	if len(scanEvents) < 3 {
		t.Fatalf("expected several counter updates, got %d (%+v)", len(scanEvents), scanEvents)
	}

	// 6. Assembly is announced exactly once, closed out on the full scan.
	if len(assembleEvents) != 1 {
		t.Fatalf("want exactly 1 assembling event, got %d", len(assembleEvents))
	}
	if a := assembleEvents[0]; a.done != total || a.total != total {
		t.Fatalf("assembling event = %+v, want done=%d total=%d", a, total, total)
	}

	t.Logf("counter emitted %d scanning updates over %d input chunks (output chunks=%d)", len(scanEvents), total, len(res.Chunks))
}

// TestCleanHybridProgressIsObserveOnly pins the additive contract: injecting a
// callback must not alter a single byte of the cleaning output. Without this the
// change would be a silent behaviour change rather than an upgrade.
func TestCleanHybridProgressIsObserveOnly(t *testing.T) {
	text := progressTestText(40)
	eng := New(brain.NewStore(nil, nil))
	det := progressTestDetector{}
	ver := progressTestVerifier{}

	withCallback, err := eng.CleanHybrid(WithProgress(context.Background(), func(string, int, int) {}), text, det, ver, nil, "")
	if err != nil {
		t.Fatalf("with callback: %v", err)
	}
	withoutCallback, err := eng.CleanHybrid(context.Background(), text, det, ver, nil, "")
	if err != nil {
		t.Fatalf("without callback: %v", err)
	}

	if withCallback.CleanedText != withoutCallback.CleanedText {
		t.Fatal("injecting a progress callback changed the cleaned text")
	}
	if withCallback.SHA256In != withoutCallback.SHA256In || withCallback.SHA256Out != withoutCallback.SHA256Out {
		t.Fatalf("injecting a progress callback changed a hash: in %s/%s out %s/%s",
			withCallback.SHA256In, withoutCallback.SHA256In, withCallback.SHA256Out, withoutCallback.SHA256Out)
	}
	if len(withCallback.Chunks) != len(withoutCallback.Chunks) {
		t.Fatalf("chunk count changed: %d vs %d", len(withCallback.Chunks), len(withoutCallback.Chunks))
	}
	if len(withCallback.Hits) != len(withoutCallback.Hits) {
		t.Fatalf("hit count changed: %d vs %d", len(withCallback.Hits), len(withoutCallback.Hits))
	}
	if len(withCallback.P1Hits) != len(withoutCallback.P1Hits) {
		t.Fatalf("P1 hit count changed: %d vs %d", len(withCallback.P1Hits), len(withoutCallback.P1Hits))
	}
}

// TestWithProgressNilCallbackIsSafe covers the nil guards: a caller that has
// nothing to report must be able to pass nil and get the old behaviour back.
func TestWithProgressNilCallbackIsSafe(t *testing.T) {
	eng := New(brain.NewStore(nil, nil))
	ctx := WithProgress(context.Background(), nil)
	if ctx == nil {
		t.Fatal("WithProgress(ctx, nil) returned a nil context")
	}
	if _, err := eng.CleanHybrid(ctx, "alpha bravo charlie delta echo.", progressTestDetector{}, progressTestVerifier{}, nil, ""); err != nil {
		t.Fatalf("a nil callback changed behaviour: %v", err)
	}

	// reportProgress must also survive a nil context and a context that never
	// went through WithProgress.
	reportProgress(nil, "scanning", 1, 1)
	reportProgress(context.Background(), "scanning", 1, 1)
}

// TestCleanHybridBlankTextEmitsNoCounter documents that the empty-input fast path
// returns before the counter is ever opened, so a caller never sees a 0/0 event.
func TestCleanHybridBlankTextEmitsNoCounter(t *testing.T) {
	var mu sync.Mutex
	var events []progressEvent
	ctx := WithProgress(context.Background(), func(stage string, done, total int) {
		mu.Lock()
		events = append(events, progressEvent{stage: stage, done: done, total: total})
		mu.Unlock()
	})

	if _, err := New(brain.NewStore(nil, nil)).CleanHybrid(ctx, "   \n\t  ", progressTestDetector{}, progressTestVerifier{}, nil, ""); err != nil {
		t.Fatalf("blank text: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 0 {
		t.Fatalf("blank input should emit no counter events, got %+v", events)
	}
}


