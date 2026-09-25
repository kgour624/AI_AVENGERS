package training

import (
	"strings"
	"testing"
)

// The numbers here are the production ones: a 395-chunk document whose corpus
// holds 277 rows, because 118 chunks repeated text that was already there. The
// whole point of the diagnostic is that this case and a genuinely truncated store
// produce the same totals and need opposite responses.

func TestClassifyFileIntegritySeparatesDuplicatesFromALostTail(t *testing.T) {
	tests := []struct {
		name               string
		stored             int
		minIndex           int
		maxIndex           int
		distinct           int
		expected           int
		wantVerdict        string
		wantShortfall      int
		wantMissingIndices int
		wantTailMissing    bool
	}{
		{
			// Every index from 0..394 was offered; 118 were repeats. The last
			// chunk landed, so nothing is missing.
			name:               "duplicate text inside the document",
			stored:             277,
			minIndex:           0,
			maxIndex:           394,
			distinct:           277,
			expected:           395,
			wantVerdict:        integrityDuplicatesMerged,
			wantShortfall:      118,
			wantMissingIndices: 118,
			wantTailMissing:    false,
		},
		{
			// The store stopped at index 275, so the tail of the document is
			// absent. Identical shortfall shape, opposite meaning.
			name:               "store stopped early",
			stored:             276,
			minIndex:           0,
			maxIndex:           275,
			distinct:           276,
			expected:           395,
			wantVerdict:        integrityTailMissing,
			wantShortfall:      119,
			wantMissingIndices: 0,
			wantTailMissing:    true,
		},
		{
			name:               "everything present",
			stored:             100,
			minIndex:           0,
			maxIndex:           99,
			distinct:           100,
			expected:           100,
			wantVerdict:        integrityComplete,
			wantShortfall:      0,
			wantMissingIndices: 0,
			wantTailMissing:    false,
		},
		{
			// Append mode: an earlier run already stored rows for this file, so
			// there are MORE rows than this run parsed. Not a loss.
			name:               "more rows than this run parsed",
			stored:             25,
			minIndex:           0,
			maxIndex:           24,
			distinct:           25,
			expected:           10,
			wantVerdict:        integrityComplete,
			wantShortfall:      0,
			wantMissingIndices: 0,
			wantTailMissing:    false,
		},
		{
			name:               "nothing stored",
			stored:             0,
			minIndex:           0,
			maxIndex:           0,
			distinct:           0,
			expected:           395,
			wantVerdict:        integrityUnchecked,
			wantShortfall:      0,
			wantMissingIndices: 0,
			wantTailMissing:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyFileIntegrity(
				"transcript.txt", tc.stored, tc.minIndex, tc.maxIndex, tc.distinct, tc.expected)

			if got.Verdict != tc.wantVerdict {
				t.Fatalf("verdict = %q, want %q", got.Verdict, tc.wantVerdict)
			}
			if got.Shortfall != tc.wantShortfall {
				t.Fatalf("shortfall = %d, want %d", got.Shortfall, tc.wantShortfall)
			}
			if got.MissingIndices != tc.wantMissingIndices {
				t.Fatalf("missing indices = %d, want %d", got.MissingIndices, tc.wantMissingIndices)
			}
			if got.TailMissing != tc.wantTailMissing {
				t.Fatalf("tail missing = %v, want %v", got.TailMissing, tc.wantTailMissing)
			}
			if got.SourceFile != "transcript.txt" {
				t.Fatalf("source file = %q, want it carried through", got.SourceFile)
			}
		})
	}
}

// A single row is the edge case where "the last chunk landed" is trivially true
// and the span arithmetic must not invent a gap.
func TestClassifyFileIntegrityHandlesSingleRow(t *testing.T) {
	got := classifyFileIntegrity("one.txt", 1, 0, 0, 1, 1)
	if got.Verdict != integrityComplete {
		t.Fatalf("verdict = %q, want %q", got.Verdict, integrityComplete)
	}
	if got.MissingIndices != 0 {
		t.Fatalf("missing indices = %d, want 0", got.MissingIndices)
	}
}

// The verdict must reach the admin as a sentence that names the cause, otherwise
// the audit screen is just numbers again.
func TestFileIntegritySentenceNamesTheCause(t *testing.T) {
	truncated := classifyFileIntegrity("talk.md", 276, 0, 275, 276, 395)
	if sentence := truncated.classifySentence(); !strings.Contains(sentence, "stopped early") {
		t.Fatalf("tail-missing sentence = %q, want it to say the store stopped early", sentence)
	}

	dupes := classifyFileIntegrity("talk.md", 277, 0, 394, 277, 395)
	if sentence := dupes.classifySentence(); !strings.Contains(sentence, "repeat text") {
		t.Fatalf("duplicates sentence = %q, want it to explain the repeats", sentence)
	}
}

func TestValidateActions(t *testing.T) {
	t.Run("accepts every documented action", func(t *testing.T) {
		if err := ValidateActions(ReconcileActions()); err != nil {
			t.Fatalf("ValidateActions(ReconcileActions()) = %v, want nil", err)
		}
	})

	t.Run("rejects an empty list", func(t *testing.T) {
		err := ValidateActions(nil)
		if err == nil {
			t.Fatal("ValidateActions(nil) = nil, want an error")
		}
		// The message must tell the caller what to do instead.
		for _, want := range ReconcileActions() {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q does not list valid action %q", err.Error(), want)
			}
		}
	})

	t.Run("rejects an unknown action by name", func(t *testing.T) {
		err := ValidateActions([]string{ReconcileActionExpertStats, "delete_everything"})
		if err == nil {
			t.Fatal("ValidateActions with an unknown action = nil, want an error")
		}
		if !strings.Contains(err.Error(), "delete_everything") {
			t.Fatalf("error %q does not name the offending action", err.Error())
		}
	})
}
