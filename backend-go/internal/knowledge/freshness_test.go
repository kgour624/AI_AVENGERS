package knowledge

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCorpusAgeDays(t *testing.T) {
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	if got := CorpusAgeDays(nil, now); got != -1 {
		t.Errorf("nil newest must be -1, got %d", got)
	}
	old := now.AddDate(0, 0, -200)
	if got := CorpusAgeDays(&old, now); got != 200 {
		t.Errorf("age=%d want 200", got)
	}
	future := now.Add(48 * time.Hour)
	if got := CorpusAgeDays(&future, now); got != 0 {
		t.Errorf("future must clamp to 0, got %d", got)
	}
}

func TestIsStale(t *testing.T) {
	if IsStale(-1, 180) {
		t.Error("unknown age (-1) must never be stale")
	}
	if IsStale(200, 0) {
		t.Error("non-positive window disables staleness")
	}
	if IsStale(180, 180) {
		t.Error("exactly at window is not stale (must exceed)")
	}
	if !IsStale(181, 180) {
		t.Error("181 > 180 must be stale")
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		chunks, mismatch int
		stale            bool
		want             string
	}{
		{0, 0, false, FreshnessEmpty},
		{10, 3, false, FreshnessNeedsReembed},
		{10, 0, true, FreshnessStale},
		{10, 0, false, FreshnessFresh},
		// mismatch outranks stale
		{10, 1, true, FreshnessNeedsReembed},
	}
	for _, c := range cases {
		if got := Classify(c.chunks, c.mismatch, c.stale); got != c.want {
			t.Errorf("Classify(%d,%d,%v)=%s want %s", c.chunks, c.mismatch, c.stale, got, c.want)
		}
	}
}

func TestPolarityConflict(t *testing.T) {
	// Opposite polarity + shared subject term -> flagged with a subject term.
	if term, ok := polarityConflict("The cache stores messages for 30 seconds", "The cache does not store messages for 30 seconds"); !ok || term == "" {
		t.Fatalf("opposite polarity overlap = (%q, %v), want a non-empty term and true", term, ok)
	}
	// Same polarity -> never flagged.
	if _, ok := polarityConflict("The cache stores messages for 30 seconds", "The cache stores messages for 30 seconds"); ok {
		t.Fatal("same-polarity statements must not be flagged")
	}
	// Negation but no shared subject -> not flagged.
	if _, ok := polarityConflict("Database uses indexes for lookups", "The cache does not store sessions"); ok {
		t.Fatal("unrelated negated statements must not be flagged")
	}
}

func TestPolicyDefault(t *testing.T) {
	f := NewFreshness(nil, Policy{}, nil)
	if f.MaxAgeDays() != 180 {
		t.Errorf("default max age must be 180, got %d", f.MaxAgeDays())
	}
	var nilf *Freshness
	if nilf.Enabled() {
		t.Error("nil service must be disabled")
	}
	if nilf.MaxAgeDays() != 180 {
		t.Error("nil service MaxAgeDays must be safe")
	}
}

func TestNewFreshness_EnabledRequiresDB(t *testing.T) {
	// A service constructed with no db is disabled but safe to call.
	var f *Freshness = NewFreshness(nil, Policy{MaxCorpusAgeDays: 90}, nil)
	if f.Enabled() {
		t.Error("no-db service must report disabled")
	}
	if _, err := f.ScanExpert(context.Background(), uuid.Nil); err == nil {
		t.Error("disabled service ScanExpert must error")
	}
}
