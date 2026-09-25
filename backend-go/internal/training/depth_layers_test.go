package training

import (
	"strings"
	"testing"
)

// The layer label is what turns "does this expert only know what things are?" from a
// feeling into a fact, so both the parsing and the sentences derived from it are pinned.

func TestParseDepthLayersAcceptsWellFormedAnswers(t *testing.T) {
	response := "```json\n" + `{"layers":[{"n":1,"layer":1},{"n":2,"layer":3}]}` + "\n```"

	layers, rejected, err := parseDepthLayers(response, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected != 0 {
		t.Fatalf("rejected = %d, want 0", rejected)
	}
	if len(layers) != 2 {
		t.Fatalf("got %d layers, want 2", len(layers))
	}
	// Indexes are 0-based after parsing, matching how the caller indexes its batch.
	if layers[0] != LayerDefinition {
		t.Fatalf("passage 1 = %d, want %d", layers[0], LayerDefinition)
	}
	if layers[1] != LayerFailure {
		t.Fatalf("passage 2 = %d, want %d", layers[1], LayerFailure)
	}
}

func TestParseDepthLayersDropsWhatItCannotTrust(t *testing.T) {
	tests := []struct {
		name         string
		entry        string
		count        int
		wantRejected int
		wantKept     int
	}{
		{
			// A layer applied to the WRONG chunk is worse than no layer: an
			// unclassified chunk is visibly pending, a mislabelled one is silently
			// wrong for as long as the corpus lives.
			name:         "passage number beyond the batch",
			entry:        `{"n":9,"layer":2}`,
			count:        3,
			wantRejected: 1,
		},
		{
			name:         "zero is not a passage",
			entry:        `{"n":0,"layer":2}`,
			count:        3,
			wantRejected: 1,
		},
		{
			name:         "layer out of range",
			entry:        `{"n":1,"layer":7}`,
			count:        3,
			wantRejected: 1,
		},
		{
			name:         "layer zero",
			entry:        `{"n":1,"layer":0}`,
			count:        3,
			wantRejected: 1,
		},
		{
			name:     "valid entry",
			entry:    `{"n":2,"layer":3}`,
			count:    3,
			wantKept: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			layers, rejected, err := parseDepthLayers(`{"layers":[`+tc.entry+`]}`, tc.count)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rejected != tc.wantRejected {
				t.Fatalf("rejected = %d, want %d", rejected, tc.wantRejected)
			}
			if len(layers) != tc.wantKept {
				t.Fatalf("kept %d, want %d", len(layers), tc.wantKept)
			}
		})
	}
}

func TestParseDepthLayersKeepsFirstAnswerForADuplicate(t *testing.T) {
	layers, rejected, err := parseDepthLayers(
		`{"layers":[{"n":1,"layer":1},{"n":1,"layer":3}]}`, 1)
	if err != nil || rejected != 0 {
		t.Fatalf("rejected=%d err=%v", rejected, err)
	}
	if len(layers) != 1 || layers[0] != LayerDefinition {
		t.Fatalf("layers = %v, want the first answer kept", layers)
	}
}

func TestParseDepthLayersHandlesUnusableResponses(t *testing.T) {
	if _, _, err := parseDepthLayers("I cannot help with that.", 3); err == nil {
		t.Fatal("prose without JSON must be an error")
	}
	if _, _, err := parseDepthLayers("", 3); err == nil {
		t.Fatal("empty response must be an error")
	}
	// A model that answered with no entries is not an error: every chunk stays
	// unclassified, which is visible and resumable.
	layers, rejected, err := parseDepthLayers(`{"layers":[]}`, 3)
	if err != nil || rejected != 0 || len(layers) != 0 {
		t.Fatalf("empty list should parse cleanly: layers=%v rejected=%d err=%v", layers, rejected, err)
	}
}

// The finding that matters most: a topic with no failure-mode content cannot answer
// "what breaks", and that must be said in words rather than left as a number.
func TestLayerFindingsNameTheGaps(t *testing.T) {
	report := &DepthLayerReport{
		TotalChunks:            100,
		Classified:             100,
		Definition:             80,
		Mechanics:              15,
		Failure:                5,
		TopicsWithoutFailure:   7,
		TopicsWithoutMechanics: 2,
	}

	findings := layerFindings(report)
	joined := strings.Join(findings, " | ")

	if !strings.Contains(joined, "7 topic(s) have NO failure-mode") {
		t.Fatalf("missing the failure-mode gap: %v", findings)
	}
	if !strings.Contains(joined, "2 topic(s) have no mechanics") {
		t.Fatalf("missing the mechanics gap: %v", findings)
	}
	// 80% definitions is the "broad but shallow" case and must be stated as such.
	if !strings.Contains(joined, "80%") || !strings.Contains(joined, "broad but shallow") {
		t.Fatalf("missing the definition-share finding: %v", findings)
	}
}

func TestLayerFindingsHandleUnclassifiedAndEmpty(t *testing.T) {
	// Nothing stored at all: one clear statement, not a wall of zeroes.
	empty := layerFindings(&DepthLayerReport{})
	if len(empty) != 1 || !strings.Contains(empty[0], "No chunks") {
		t.Fatalf("empty corpus findings = %v", empty)
	}

	// Partly classified: the admin must be told the picture is incomplete BEFORE any
	// conclusion is drawn from it.
	partial := layerFindings(&DepthLayerReport{TotalChunks: 100, Classified: 0, Unclassified: 100})
	if len(partial) == 0 {
		t.Fatal("partial classification produced no findings")
	}
	if !strings.Contains(partial[0], "not classified yet") {
		t.Fatalf("first finding should say the picture is incomplete: %v", partial)
	}
}

func TestLayerFindingsReportBalance(t *testing.T) {
	// A corpus with real spread across the three kinds gets no scary sentence.
	balanced := layerFindings(&DepthLayerReport{
		TotalChunks: 90, Classified: 90,
		Definition: 30, Mechanics: 30, Failure: 30,
	})
	if len(balanced) != 1 || !strings.Contains(balanced[0], "balanced") {
		t.Fatalf("balanced corpus findings = %v", balanced)
	}
}

func TestDepthLayerNames(t *testing.T) {
	if DepthLayerName(LayerDefinition) != "what / why" {
		t.Fatalf("layer 1 name = %q", DepthLayerName(LayerDefinition))
	}
	if DepthLayerName(LayerMechanics) != "how / trade-offs" {
		t.Fatalf("layer 2 name = %q", DepthLayerName(LayerMechanics))
	}
	if DepthLayerName(LayerFailure) != "failure / edge cases" {
		t.Fatalf("layer 3 name = %q", DepthLayerName(LayerFailure))
	}
	if DepthLayerName(0) != "not classified" {
		t.Fatalf("unset name = %q", DepthLayerName(0))
	}
}
