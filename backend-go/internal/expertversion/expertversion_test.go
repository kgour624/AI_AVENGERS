package expertversion

import "testing"

func TestSHA256Hex_Stable(t *testing.T) {
	if sha256Hex("abc") != sha256Hex("abc") {
		t.Fatal("not stable")
	}
	if sha256Hex("abc") == sha256Hex("abd") {
		t.Fatal("different input, same hash")
	}
	if len(sha256Hex("abc")) != 64 {
		t.Fatalf("len=%d want 64", len(sha256Hex("abc")))
	}
}

func TestJaccardDistance(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want float64
	}{
		{"identical", []string{"a", "b"}, []string{"a", "b"}, 0},
		{"disjoint", []string{"a"}, []string{"b"}, 1},
		{"half", []string{"a", "b"}, []string{"a", "c"}, 2.0 / 3.0}, // inter1 union3 -> 1-1/3
		{"both empty", nil, nil, 0},
		{"one empty", []string{"a"}, nil, 1},
	}
	for _, tc := range cases {
		got := jaccardDistance(tc.a, tc.b)
		if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestClassifyDrift_CapabilityWins(t *testing.T) {
	prev := &Version{CorpusHash: "aaa", CharterHash: "ccc", Topics: []string{"sql", "indexing"}, ChunkCount: 10}
	cur := &Version{CorpusHash: "bbb", CharterHash: "ddd", Topics: []string{"kafka", "streams"}, ChunkCount: 12}
	dt, score, _, drifted := classifyDrift(prev, cur, 0.30)
	if !drifted {
		t.Fatal("expected drift")
	}
	if dt != "capability" {
		t.Errorf("type=%s want capability", dt)
	}
	if score != 1.0 { // fully disjoint topic sets
		t.Errorf("score=%v want 1.0", score)
	}
}

func TestClassifyDrift_CharterThenCorpus(t *testing.T) {
	// Same topics, charter changed → charter drift.
	prev := &Version{CorpusHash: "aaa", CharterHash: "ccc", Topics: []string{"sql"}}
	cur := &Version{CorpusHash: "aaa", CharterHash: "ddd", Topics: []string{"sql"}}
	dt, _, _, drifted := classifyDrift(prev, cur, 0.30)
	if !drifted || dt != "charter" {
		t.Errorf("type=%s drifted=%v want charter/true", dt, drifted)
	}
	// Only corpus changed.
	cur2 := &Version{CorpusHash: "bbb", CharterHash: "ccc", Topics: []string{"sql"}}
	dt2, _, _, drifted2 := classifyDrift(prev, cur2, 0.30)
	if !drifted2 || dt2 != "corpus" {
		t.Errorf("type=%s drifted=%v want corpus/true", dt2, drifted2)
	}
}

func TestClassifyDrift_NoChange(t *testing.T) {
	prev := &Version{CorpusHash: "aaa", CharterHash: "ccc", Topics: []string{"sql"}}
	cur := &Version{CorpusHash: "aaa", CharterHash: "ccc", Topics: []string{"sql"}}
	_, _, _, drifted := classifyDrift(prev, cur, 0.30)
	if drifted {
		t.Fatal("identical versions reported drift")
	}
}

func TestClassifyDrift_ThresholdBoundary(t *testing.T) {
	// topic drift exactly at threshold counts as drift (>=).
	prev := &Version{Topics: []string{"a", "b", "c", "d"}, CorpusHash: "x", CharterHash: "y"}
	cur := &Version{Topics: []string{"a", "b", "c", "e"}, CorpusHash: "x", CharterHash: "y"}
	// inter 3 union 5 → drift 0.4 ≥ 0.4
	dt, score, _, drifted := classifyDrift(prev, cur, 0.40)
	if !drifted || dt != "capability" {
		t.Errorf("type=%s drifted=%v score=%v", dt, drifted, score)
	}
}
