package selflearning

import (
	"reflect"
	"testing"
)

func TestCountTokens(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"   ", 0},
		{"one two three", 3},
		{"  padded   words  ", 2},
	}
	for _, c := range cases {
		if got := countTokens(c.in); got != c.want {
			t.Errorf("countTokens(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestShouldProcess(t *testing.T) {
	long := "Alice is on a chessboard and wants to move a knight from one corner to another using valid moves"
	if !shouldProcess(long) {
		t.Errorf("long english question should be processed")
	}
	if shouldProcess("What is BFS?") {
		t.Errorf("short question should be skipped")
	}
	// Space-less CJK question long enough by rune count must NOT be skipped.
	cjk := "这是一个非常长的问题关于图的最短路径算法和广度优先搜索的实现细节以及复杂度分析"
	if !shouldProcess(cjk) {
		t.Errorf("long spaceless CJK question should be processed")
	}
	// Short spaceless string stays skipped.
	if shouldProcess("图论") {
		t.Errorf("short spaceless question should be skipped")
	}
}

func TestConstraintTokens(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"numbers", "find min moves for 8 pieces on 8x8 board", []string{"8"}},
		{"caps", "use BFS and DFS with API calls", []string{"BFS", "DFS", "API"}},
		{"quoted", `the "two-sum" problem and 'edge-case'`, []string{"two-sum", "edge-case"}},
		{"dedup", "100 items, 100 again", []string{"100"}},
		{"none", "just some plain words here", nil},
	}
	for _, c := range cases {
		got := constraintTokens(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestMissingConstraintTokens(t *testing.T) {
	original := "Find minimum moves for 3 knights on a 20x20 grid using BFS"
	// Extraction keeps BFS but drops 3 and 20.
	extracted := "BFS shortest path on grid, minimum distance between cells"
	missing := missingConstraintTokens(original, extracted)
	wantMissing := map[string]bool{"3": true, "20": true}
	if len(missing) != 2 {
		t.Fatalf("got %v, want missing {3,20}", missing)
	}
	for _, m := range missing {
		if !wantMissing[m] {
			t.Errorf("unexpected missing token %q", m)
		}
	}

	// Extraction that preserves everything → no missing.
	good := "BFS on a 20x20 grid with 3 knights, minimum distance"
	if got := missingConstraintTokens(original, good); len(got) != 0 {
		t.Errorf("expected no missing constraints, got %v", got)
	}

	// No constraints in original → nothing to miss.
	if got := missingConstraintTokens("explain recursion", "recursion and base case"); len(got) != 0 {
		t.Errorf("expected no missing constraints, got %v", got)
	}
}

func TestEvaluateExtraction(t *testing.T) {
	original := "Find minimum moves for 3 knights on a 20x20 grid using BFS"
	// Perfect extraction: preserves 3, 20, BFS.
	perfect := "BFS on a 20x20 grid with 3 knights, find minimum distance between cells"
	score, reasons := evaluateExtraction(original, perfect)
	if score <= 0 || score > 1 {
		t.Fatalf("score out of range: %f", score)
	}
	if len(reasons) == 0 {
		t.Fatalf("expected reasons to be populated")
	}
	// Preservation should yield a high score.
	if score < 0.9 {
		t.Errorf("expected high score for perfect extraction, got %f (%v)", score, reasons)
	}

	// Lossy extraction: dropped 3 and 20 → lower score.
	lossy := "BFS shortest path on grid, minimum distance between cells"
	lossyScore, _ := evaluateExtraction(original, lossy)
	if lossyScore >= score {
		t.Errorf("lossy score %f should be lower than perfect %f", lossyScore, score)
	}

	// No-constraint original with a same-length rewrite → near 1.0.
	s1, _ := evaluateExtraction("explain recursion in detail please", "recursion explanation with base case and stack")
	if s1 <= 0 || s1 > 1 {
		t.Errorf("score out of range: %f", s1)
	}
}
