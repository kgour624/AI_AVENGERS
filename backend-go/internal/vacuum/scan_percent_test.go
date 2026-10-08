package vacuum

import "testing"

// TestScanPercentStaysInsideThePhase2Band guards the mapping the live phase-2
// counter uses to move the progress bar. The band must stay inside the values
// phasePercent already reserves for phases 2 and 3, otherwise the counter would
// either stall the bar or make it jump past the phase it belongs to.
func TestScanPercentStaysInsideThePhase2Band(t *testing.T) {
	lo, hi := phasePercent(2), phasePercent(3)
	if hi <= lo {
		t.Fatalf("phase map is degenerate: phasePercent(2)=%d phasePercent(3)=%d", lo, hi)
	}

	cases := []struct {
		name        string
		done, total int
		want        int
	}{
		{"unknown total falls back to the band floor", 0, 0, lo},
		{"negative total falls back to the band floor", 3, -1, lo},
		{"negative done is clamped to the floor", -5, 10, lo},
		{"start of the scan is the band floor", 0, 10, lo},
		{"midpoint", 5, 10, lo + (hi-lo)/2},
		{"end of the scan is the band ceiling", 10, 10, hi},
		{"an overshooting done is clamped to the ceiling", 99, 10, hi},
	}

	for _, c := range cases {
		if got := scanPercent(c.done, c.total); got != c.want {
			t.Errorf("%s: scanPercent(%d, %d) = %d, want %d", c.name, c.done, c.total, got, c.want)
		}
	}
}

// TestScanPercentIsMonotonic verifies the bar can only ever move forwards as
// chunks complete — a non-monotonic percent is what makes a progress bar look
// like it is going backwards.
func TestScanPercentIsMonotonic(t *testing.T) {
	lo, hi := phasePercent(2), phasePercent(3)
	const total = 137
	prev := lo
	for done := 0; done <= total; done++ {
		got := scanPercent(done, total)
		if got < prev {
			t.Fatalf("scanPercent(%d, %d) = %d went backwards from %d", done, total, got, prev)
		}
		if got < lo || got > hi {
			t.Fatalf("scanPercent(%d, %d) = %d left the %d..%d band", done, total, got, lo, hi)
		}
		prev = got
	}
	if prev != hi {
		t.Fatalf("a completed scan must reach the band ceiling %d, got %d", hi, prev)
	}
}
