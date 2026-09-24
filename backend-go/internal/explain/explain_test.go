package explain

import (
	"encoding/json"
	"testing"
)

func TestGateTimeline_AllPassed(t *testing.T) {
	steps := GateTimeline(0)
	if len(steps) != 5 {
		t.Fatalf("want 5 gates, got %d", len(steps))
	}
	for _, s := range steps {
		if s.Status != "passed" {
			t.Errorf("gate %d status=%q want passed", s.N, s.Status)
		}
	}
}

func TestGateTimeline_StoppedAtGate(t *testing.T) {
	steps := GateTimeline(2)
	if len(steps) != 5 {
		t.Fatalf("want 5 gates, got %d", len(steps))
	}
	if steps[0].Status != "passed" {
		t.Errorf("gate 1 = %q want passed", steps[0].Status)
	}
	if steps[1].Status != "stopped" {
		t.Errorf("gate 2 = %q want stopped", steps[1].Status)
	}
	if steps[2].Status != "not_reached" {
		t.Errorf("gate 3 = %q want not_reached", steps[2].Status)
	}
}

func TestGateTimeline_Sentinels(t *testing.T) {
	// Structure-permission ASK sentinel.
	ask := GateTimeline(-1)
	if len(ask) != 1 || ask[0].N != 0 || ask[0].Status != "asked" {
		t.Fatalf("sentinel -1 = %+v want single gate0 asked", ask)
	}
	// Unknown → everything not_reached (never a false "passed").
	un := GateTimeline(99)
	for _, s := range un {
		if s.Status != "not_reached" {
			t.Errorf("unknown gateStopped: gate %d = %q want not_reached", s.N, s.Status)
		}
	}
	// Gate 5 (China Wall) stop.
	g5 := GateTimeline(5)
	if g5[4].Status != "stopped" {
		t.Errorf("gate 5 = %q want stopped", g5[4].Status)
	}
}

func TestModeLabel_ModeExplanation(t *testing.T) {
	for _, m := range []string{"ADVISE", "ASK", "WARN", "PUSH_BACK", "REFUSE"} {
		if ModeLabel(m) == "Unknown" {
			t.Errorf("ModeLabel(%q) returned Unknown", m)
		}
		if ModeExplanation(m, 0) == "" {
			t.Errorf("ModeExplanation(%q) empty", m)
		}
	}
	if ModeLabel("") != "Unknown" {
		t.Error("empty mode must map to Unknown")
	}
	if ModeExplanation("", 0) != "No decision explanation is available for this message." {
		t.Error("unknown mode must give the fallback explanation")
	}
}

func TestParseSources_FailClosed(t *testing.T) {
	// Empty / garbage → empty slice (never panic, never null).
	for _, raw := range []string{"", "not-json", "null", "{}"} {
		got := parseSources(json.RawMessage(raw))
		if got == nil {
			t.Errorf("parseSources(%q) returned nil; want non-nil empty slice", raw)
		}
	}
	got := parseSources(json.RawMessage(`[{"chunk_id":"11111111-1111-1111-1111-111111111111","text":"x","score":0.9,"source_name":"f.txt","chunk_index":3}]`))
	if len(got) != 1 || got[0].Text != "x" || got[0].SourceName != "f.txt" || got[0].ChunkIndex != 3 {
		t.Fatalf("parseSources did not decode citation: %+v", got)
	}
}

func TestDerefStr(t *testing.T) {
	if derefStr(nil) != "" {
		t.Error("nil → empty")
	}
	s := "hi"
	if derefStr(&s) != "hi" {
		t.Error("ptr → value")
	}
}
