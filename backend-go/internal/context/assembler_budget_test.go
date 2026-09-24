package context

import (
	"testing"

	"ai_avengers/backend/internal/chinawall"
	"ai_avengers/backend/internal/memory"
)

func TestTrimChunksToBudget_KeepsBestFirst(t *testing.T) {
	// Each chunk ~ text len/4 tokens. Make 10 chunks of 40 chars = 10 tokens each.
	mk := func(i int) chinawall.CourseChunk {
		return chinawall.CourseChunk{Text: string(make([]byte, 40)), RerankScore: float32(1 - i*0.01)}
	}
	var chunks []chinawall.CourseChunk
	for i := 0; i < 10; i++ {
		chunks = append(chunks, mk(i))
	}
	kept, used := trimChunksToBudget(chunks, 30) // room for 3 chunks
	if len(kept) != 3 {
		t.Errorf("kept=%d want 3", len(kept))
	}
	if used != 30 {
		t.Errorf("used=%d want 30", used)
	}
	// Order preserved (best first)
	if kept[0].RerankScore < kept[2].RerankScore {
		t.Errorf("order not best-first")
	}
}

func TestTrimChunksToBudget_ZeroBudget(t *testing.T) {
	chunks := []chinawall.CourseChunk{{Text: "hello world"}}
	kept, used := trimChunksToBudget(chunks, 0)
	if len(kept) != 0 || used != 0 {
		t.Errorf("zero budget must keep nothing: kept=%d used=%d", len(kept), used)
	}
}

func TestEnforceHardCeiling_EvictsLowestValueFirst(t *testing.T) {
	budget := 100
	asm := &AssembledContext{
		RollingSummary: "summary text here", // ~4 tokens
		RecentMessages: []Message{
			{Role: "user", Content: longText(20)},
			{Role: "assistant", Content: longText(20)},
		},
		RelevantHistory: []HistoryEntry{
			{OneLineSummary: longText(30)},
			{OneLineSummary: longText(30)},
		},
		ProjectContext: &memory.ProjectContext{
			L2Entries: []memory.L2Entry{{Content: longText(20)}},
		},
		CourseChunks: []chinawall.CourseChunk{{Text: longText(40)}},
	}
	tokensUsed := 4 + 20 + 20 + 30 + 30 + 20 + 40 // = 164
	got := enforceHardCeiling(asm, budget, tokensUsed, nil)
	if got > budget {
		t.Errorf("still over budget: %d > %d", got, budget)
	}
	// Semantic history (lowest value) evicted first.
	if len(asm.RelevantHistory) != 0 {
		t.Errorf("history should be evicted first, left=%d", len(asm.RelevantHistory))
	}
	// Rolling summary never evicted unless everything else gone.
	if asm.RollingSummary == "" {
		t.Errorf("rolling summary evicted too eagerly")
	}
}

func TestEnforceHardCeiling_NoOpWhenUnder(t *testing.T) {
	asm := &AssembledContext{RecentMessages: []Message{{Content: "hi"}}}
	if got := enforceHardCeiling(asm, 1000, 50, nil); got != 50 {
		t.Errorf("under budget must be unchanged, got %d", got)
	}
}

func TestEnforceHardCeiling_KeepsReplyThreadDropsSummaryLastResort(t *testing.T) {
	// Only a huge reply thread and a summary; the reply thread is never
	// evicted, so the summary is dropped as the last resort.
	budget := 10
	asm := &AssembledContext{
		RollingSummary: longText(40), // 10 tokens
		ReplyThread:    []ReplyThreadEntry{{Content: longText(80)}}, // 20 tokens
	}
	tokensUsed := 10 + 20
	got := enforceHardCeiling(asm, budget, tokensUsed, nil)
	if len(asm.ReplyThread) != 1 {
		t.Errorf("reply thread must never be evicted")
	}
	if asm.RollingSummary != "" {
		t.Errorf("summary should drop as last resort when reply pinned")
	}
	if got != 20 {
		t.Errorf("got=%d want 20 (reply thread bound)", got)
	}
}

func longText(tokens int) string {
	b := make([]byte, tokens*4)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
