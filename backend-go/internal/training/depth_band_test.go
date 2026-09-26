package training

import "testing"

// The declared depth band is displayed next to the chunk count it is derived
// from, so the boundaries are pinned: a topic sitting on a boundary must not
// change band depending on which code path computed it.
func TestDepthLevelForChunkCount(t *testing.T) {
	tests := []struct {
		name       string
		chunkCount int
		want       int
	}{
		{name: "no chunks is still a band, not an error", chunkCount: 0, want: 1},
		{name: "one chunk", chunkCount: 1, want: 1},
		{name: "just below the 2 band", chunkCount: 4, want: 1},
		{name: "bottom of the 2 band", chunkCount: 5, want: 2},
		{name: "top of the 2 band", chunkCount: 14, want: 2},
		{name: "bottom of the 3 band", chunkCount: 15, want: 3},
		{name: "top of the 3 band", chunkCount: 29, want: 3},
		{name: "bottom of the 4 band", chunkCount: 30, want: 4},
		{name: "top of the 4 band", chunkCount: 49, want: 4},
		{name: "bottom of the 5 band", chunkCount: 50, want: 5},
		{name: "well above the top band", chunkCount: 5000, want: 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := depthLevelForChunkCount(tc.chunkCount); got != tc.want {
				t.Fatalf("depthLevelForChunkCount(%d) = %d, want %d", tc.chunkCount, got, tc.want)
			}
		})
	}
}

// Every band must map to a label, because the two are written together by the
// corpus recompute. A gap here would store an empty complexity_ceiling.
func TestEveryDepthBandHasALabel(t *testing.T) {
	for level := 1; level <= 5; level++ {
		if label := depthLevelToComplexity(level); label == "" {
			t.Fatalf("depth level %d has no complexity label", level)
		}
	}
}
