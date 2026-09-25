package training

import (
	"strings"
	"testing"
)

// The numbers in these tests are the real ones from production: a 395-chunk
// document whose corpus holds 277 rows, because 118 chunks repeated text that
// was already there. That case used to be reported as a database mismatch.

func TestCountDistinctHashesSeparatesRepeatsFromChunks(t *testing.T) {
	tests := []struct {
		name           string
		hashes         []string
		wantUnique     int
		wantDuplicates int
	}{
		{
			name:           "all distinct",
			hashes:         []string{"a", "b", "c"},
			wantUnique:     3,
			wantDuplicates: 0,
		},
		{
			name:           "one repeat",
			hashes:         []string{"a", "b", "a"},
			wantUnique:     2,
			wantDuplicates: 1,
		},
		{
			name:           "every chunk the same",
			hashes:         []string{"a", "a", "a", "a"},
			wantUnique:     1,
			wantDuplicates: 3,
		},
		{
			name:           "empty run",
			hashes:         nil,
			wantUnique:     0,
			wantDuplicates: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks := make([]TextChunk, 0, len(tc.hashes))
			for _, h := range tc.hashes {
				chunks = append(chunks, TextChunk{ChunkHash: h})
			}

			unique, duplicates := countDistinctHashes(chunks)
			if unique != tc.wantUnique {
				t.Fatalf("unique = %d, want %d", unique, tc.wantUnique)
			}
			if duplicates != tc.wantDuplicates {
				t.Fatalf("duplicates = %d, want %d", duplicates, tc.wantDuplicates)
			}
		})
	}
}

// TestExpectedStoredIsTheIdentity guards the rule the verification depends on:
// the rows a file should hold after a store is what it held before, plus what the
// INSERT actually created. If this drifts, every run reports a false mismatch.
func TestExpectedStoredIsTheIdentity(t *testing.T) {
	tests := []struct {
		name        string
		preexisting int
		inserted    int
		want        int
	}{
		{name: "fresh file, everything new", preexisting: 0, inserted: 395, want: 395},
		{name: "resumed run inserts nothing", preexisting: 277, inserted: 0, want: 277},
		{name: "append onto an existing file", preexisting: 277, inserted: 12, want: 289},
		{name: "nothing anywhere", preexisting: 0, inserted: 0, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := RunLedger{Stats: StoreStats{PreexistingForFile: tc.preexisting, Inserted: tc.inserted}}
			if got := l.ExpectedStored(); got != tc.want {
				t.Fatalf("ExpectedStored() = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestDescribeLedgerShowsWhyTheCountsDiffer pins the sentence the admin reads:
// it must state the parsed count, the duplicates merged and the stored count, so
// "277 stored, 395 parsed" is explained rather than looking like data loss.
func TestDescribeLedgerShowsWhyTheCountsDiffer(t *testing.T) {
	l := RunLedger{
		Stats: StoreStats{
			Parsed:             395,
			UniqueHashes:       277,
			Duplicates:         118,
			Inserted:           277,
			Reused:             0,
			PreexistingForFile: 0,
		},
		StoredForFile: 277,
		CorpusTotal:   277,
	}

	got := l.describeLedger()
	for _, want := range []string{"parsed 395", "duplicates merged 118", "inserted 277", "stored 277"} {
		if !strings.Contains(got, want) {
			t.Fatalf("describeLedger() = %q, missing %q", got, want)
		}
	}
}
