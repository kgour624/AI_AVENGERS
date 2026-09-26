package training

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// These tests pin the batch RECOVERY behaviour that the depth-layers incident
// exposed: a batch the model cannot answer must be split, not discarded. If a
// future change removes the split, these fail instead of production.

// jsonFor returns a valid layers payload for n passages, all layer 1.
func jsonFor(n int) string {
	out := `{"layers":[`
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf(`{"n":%d,"layer":1}`, i+1)
	}
	return out + "]}"
}

// TestClassifyBatchRecoversBySplitting: a caller that only answers batches of
// <=5 must still yield every passage, with a non-nil error for the failed top
// batch but a complete merged map.
func TestClassifyBatchRecoversBySplitting(t *testing.T) {
	var calls int
	call := func(_ context.Context, texts []string) (string, error) {
		calls++
		if len(texts) > 5 {
			return "", errors.New("empty content for big batch")
		}
		return jsonFor(len(texts)), nil
	}

	texts := make([]string, 20)
	for i := range texts {
		texts[i] = fmt.Sprintf("passage %d", i)
	}
	layers, _, _, err := classifyWithCaller(context.Background(), call, texts)

	if len(layers) != 20 {
		t.Fatalf("recovered %d of 20 passages", len(layers))
	}
	for i := 0; i < 20; i++ {
		if layers[i] != 1 {
			t.Fatalf("passage %d layer = %d, want 1", i, layers[i])
		}
	}
	// Every passage was recovered, so the outcome is a SUCCESS — a fully
	// recovered batch must not be reported as a failure to the admin.
	if err != nil {
		t.Fatalf("fully recovered batch must not error, got %v", err)
	}
	// And recovery must actually have used the split (more than one call).
	if calls <= 1 {
		t.Fatalf("calls = %d, want the batch to have been split and retried", calls)
	}
}

// TestClassifyBatchSinglePassageStopsRecursing: with a caller that always fails,
// recursion must bottom out at one passage (bounded, no infinite split) and
// return nothing rather than panicking.
func TestClassifyBatchSinglePassageStopsRecursing(t *testing.T) {
	var calls int
	call := func(_ context.Context, _ []string) (string, error) {
		calls++
		return "", errors.New("always fails")
	}
	texts := []string{"a", "b", "c", "d"}
	layers, _, _, err := classifyWithCaller(context.Background(), call, texts)
	if err == nil {
		t.Fatal("expected an error when every call fails")
	}
	if len(layers) != 0 {
		t.Fatalf("expected no layers, got %d", len(layers))
	}
	// Bounded binary split of 4 with every call failing: 1 (top) + 2 (halves)
	// + 4 (single passages) = 7 calls. Any larger number would mean the split
	// is unbounded or re-calling the same size.
	if calls != 7 {
		t.Fatalf("calls = %d, want 7 (bounded split)", calls)
	}
}

// TestClassifyBatchUnparseableSplits: a parse failure (not a call error) must
// also trigger the split, so a truncated/prose response cannot lose the batch.
func TestClassifyBatchUnparseableSplits(t *testing.T) {
	call := func(_ context.Context, texts []string) (string, error) {
		if len(texts) > 2 {
			return "sorry, I cannot do that", nil // unparseable
		}
		return jsonFor(len(texts)), nil
	}
	texts := make([]string, 10)
	layers, _, _, err := classifyWithCaller(context.Background(), call, texts)
	if len(layers) != 10 {
		t.Fatalf("recovered %d of 10 passages", len(layers))
	}
	// An unparseable response on a big batch is fully recovered by splitting, so
	// the outcome is success (same rule as the call-error case).
	if err != nil {
		t.Fatalf("fully recovered parse failure must not error, got %v", err)
	}
}

// TestClassifyBatchReportsPartialLoss: when one passage truly cannot be
// classified, the batch must be reported as failed (so the admin sees it) while
// the passages that DID work are still returned.
func TestClassifyBatchReportsPartialLoss(t *testing.T) {
	call := func(_ context.Context, texts []string) (string, error) {
		// Refuse exactly one specific passage, answer everything else.
		for _, t := range texts {
			if t == "bad" {
				return "", errors.New("cannot classify this one")
			}
		}
		return jsonFor(len(texts)), nil
	}
	texts := []string{"a", "b", "bad", "c"}
	layers, _, _, err := classifyWithCaller(context.Background(), call, texts)
	if err == nil {
		t.Fatal("a lost passage must be reported as a failure")
	}
	if len(layers) != 3 {
		t.Fatalf("recovered %d of 4 passages, want 3", len(layers))
	}
	if _, ok := layers[2]; ok {
		t.Fatal("the lost passage must not be present in the result")
	}
}
