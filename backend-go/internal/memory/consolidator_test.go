package memory

import (
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDecayWeight_HalfLife(t *testing.T) {
	got := decayWeight(1.0, decayHalfLifeDays, decayHalfLifeDays)
	if math.Abs(got-0.5) > 1e-9 {
		t.Errorf("half-life: got %v want 0.5", got)
	}
}

func TestDecayWeight_ZeroIdle(t *testing.T) {
	if got := decayWeight(0.8, 0, decayHalfLifeDays); got != 0.8 {
		t.Errorf("zero idle should keep weight: %v", got)
	}
}

func TestDecayWeight_Clamp(t *testing.T) {
	if got := decayWeight(0, 10, decayHalfLifeDays); got != 0 {
		t.Errorf("zero current: %v", got)
	}
}

func TestParseConsolidationVerdict_OK(t *testing.T) {
	raw := `{"decisions":["use postgres"],"numbers":["p95 < 200ms"],"constraints":["no mongodb"],"open_items":["shard key?"],"snr":0.8}`
	v, err := parseConsolidationVerdict(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(v.Decisions) != 1 || v.Decisions[0] != "use postgres" {
		t.Errorf("decisions: %+v", v.Decisions)
	}
	if v.SNR != 0.8 {
		t.Errorf("snr=%v", v.SNR)
	}
}

func TestParseConsolidationVerdict_Fence(t *testing.T) {
	raw := "```json\n{\"decisions\":[],\"numbers\":[],\"constraints\":[],\"open_items\":[],\"snr\":0.5}\n```"
	v, err := parseConsolidationVerdict(raw)
	if err != nil {
		t.Fatalf("fence parse: %v", err)
	}
	if v.SNR != 0.5 {
		t.Errorf("snr=%v", v.SNR)
	}
}

func TestParseConsolidationVerdict_BadJSON(t *testing.T) {
	if _, err := parseConsolidationVerdict("not json"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatConsolidationSummary_Prescriptive(t *testing.T) {
	v := &consolidationVerdict{
		Decisions:   []string{"use postgres"},
		Numbers:     []string{"3 replicas"},
		Constraints: []string{"no mongo"},
		OpenItems:   []string{},
		SNR:         0.9,
	}
	out := formatConsolidationSummary(v)
	for _, need := range []string{"Decisions:", "use postgres", "Numbers:", "Constraints:", "SNR: 0.90"} {
		if !strings.Contains(out, need) {
			t.Errorf("missing %q in:\n%s", need, out)
		}
	}
	if strings.Contains(out, "Open items:") {
		t.Errorf("empty open_items must not render a section")
	}
}

func TestMajorityExpert(t *testing.T) {
	a := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	b := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	entries := []L2Entry{
		{ExpertID: a}, {ExpertID: b}, {ExpertID: a},
	}
	if got := majorityExpert(entries); got != a {
		t.Errorf("majority=%v want %v", got, a)
	}
}
