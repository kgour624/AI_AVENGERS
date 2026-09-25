package training

import (
	"strings"
	"testing"
)

// The graph is read by retrieval, which pulls a neighbour's chunks into the candidate
// set. So an edge naming a topic that does not exist would create a neighbour matching
// no corpus rows, and an unaccepted relation would make the expansion behaviour depend
// on the model's wording. These tests pin the validation that prevents both.

func TestParseConceptEdgesAcceptsWellFormedEdges(t *testing.T) {
	response := "```json\n" + `{"edges":[
		{"from":"consistent_hashing","to":"sharding","relation":"part_of","rationale":"the ring assigns keys to shards"},
		{"from":"caching","to":"cache_stampede","relation":"fails_with","rationale":"thundering herd when a hot key expires"}
	]}` + "\n```"

	allowed := map[string]bool{"consistent_hashing": true, "sharding": true, "caching": true, "cache_stampede": true}

	edges, rejected, err := parseConceptEdges(response, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected != 0 {
		t.Fatalf("rejected = %d, want 0", rejected)
	}
	if len(edges) != 2 {
		t.Fatalf("got %d edges, want 2", len(edges))
	}
	if edges[0].From != "consistent_hashing" || edges[0].Relation != RelationPartOf {
		t.Fatalf("first edge parsed wrong: %+v", edges[0])
	}
}

func TestParseConceptEdgesRejectsWhatItCannotTrust(t *testing.T) {
	allowed := map[string]bool{"caching": true, "sharding": true}

	tests := []struct {
		name         string
		edge         string
		wantRejected int
		wantKept     int
	}{
		{
			// A topic the corpus does not have. Accepting it would add a neighbour that
			// matches nothing.
			name:         "invented topic",
			edge:         `{"from":"caching","to":"quantum_cache","relation":"used_with"}`,
			wantRejected: 1,
		},
		{
			name:         "invented source topic",
			edge:         `{"from":"invented","to":"caching","relation":"used_with"}`,
			wantRejected: 1,
		},
		{
			// A relation outside the closed set: expansion behaviour must not depend on
			// model wording.
			name:         "unknown relation",
			edge:         `{"from":"caching","to":"sharding","relation":"sometimes_related_to"}`,
			wantRejected: 1,
		},
		{
			// The table has a CHECK for this; catching it here keeps a malformed row out
			// of the database instead of failing the whole insert.
			name:         "self edge",
			edge:         `{"from":"caching","to":"caching","relation":"part_of"}`,
			wantRejected: 1,
		},
		{
			name:     "valid",
			edge:     `{"from":"caching","to":"sharding","relation":"used_with"}`,
			wantKept: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			edges, rejected, err := parseConceptEdges(`{"edges":[`+tc.edge+`]}`, allowed)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rejected != tc.wantRejected {
				t.Fatalf("rejected = %d, want %d", rejected, tc.wantRejected)
			}
			if len(edges) != tc.wantKept {
				t.Fatalf("kept %d edges, want %d", len(edges), tc.wantKept)
			}
		})
	}
}

func TestParseConceptEdgesNormalisesRelationCase(t *testing.T) {
	allowed := map[string]bool{"a": true, "b": true}
	edges, rejected, err := parseConceptEdges(
		`{"edges":[{"from":"a","to":"b","relation":"PART_OF"}]}`, allowed)
	if err != nil || rejected != 0 || len(edges) != 1 {
		t.Fatalf("edges=%v rejected=%d err=%v", edges, rejected, err)
	}
	if edges[0].Relation != RelationPartOf {
		t.Fatalf("relation = %q, want %q", edges[0].Relation, RelationPartOf)
	}
}

func TestParseConceptEdgesDeduplicates(t *testing.T) {
	allowed := map[string]bool{"a": true, "b": true}
	// The same relation twice is one edge, and a duplicate is not a rejection: it is
	// the model repeating itself, not inventing something.
	edges, rejected, err := parseConceptEdges(
		`{"edges":[{"from":"a","to":"b","relation":"part_of"},{"from":"a","to":"b","relation":"part_of"}]}`,
		allowed)
	if err != nil || rejected != 0 {
		t.Fatalf("rejected=%d err=%v", rejected, err)
	}
	if len(edges) != 1 {
		t.Fatalf("got %d edges, want 1 after dedupe", len(edges))
	}
}

func TestParseConceptEdgesHandlesUnusableResponses(t *testing.T) {
	allowed := map[string]bool{"a": true, "b": true}

	if _, _, err := parseConceptEdges("I cannot help with that.", allowed); err == nil {
		t.Fatal("prose without JSON must be an error, not an empty graph")
	}
	if _, _, err := parseConceptEdges("", allowed); err == nil {
		t.Fatal("empty response must be an error")
	}

	// A well-formed response with no edges is NOT an error: the model answered, it just
	// found no relations, and the caller decides what that means.
	edges, rejected, err := parseConceptEdges(`{"edges":[]}`, allowed)
	if err != nil || rejected != 0 || len(edges) != 0 {
		t.Fatalf("empty edge list should parse cleanly: edges=%v rejected=%d err=%v", edges, rejected, err)
	}
}

// The closed relation set is what keeps expansion behaviour predictable, so it is
// asserted directly rather than trusted to the CHECK constraint alone.
func TestConceptRelationsAreTheClosedSet(t *testing.T) {
	want := []string{"prerequisite_of", "part_of", "contrasts_with", "used_with", "fails_with"}
	got := ConceptRelations()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	// The CHECK constraint lists the same values; a mismatch would let an edge through
	// validation and then fail at the database.
	joined := strings.Join(got, ",")
	for _, relation := range want {
		if !strings.Contains(joined, relation) {
			t.Fatalf("%q missing from the relation set", relation)
		}
	}
}
