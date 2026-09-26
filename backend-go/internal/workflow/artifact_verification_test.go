package workflow

import (
	"encoding/json"
	"strings"
	"testing"
)

// Artifacts are free-form: each event type posts its own shape and the shapes are
// written by a model. These cases pin what the verifier will read, because a
// silent miss here would report "nothing to verify" for a real artifact.
func TestArtifactTextAndQuestion(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantText     string
		wantQuestion string
	}{
		{
			name:     "plain text key",
			content:  `{"text":"The system shards by tenant id."}`,
			wantText: "The system shards by tenant id.",
		},
		{
			name:         "title becomes the question, body the text",
			content:      `{"title":"How is data partitioned?","body":"By tenant id, then by time."}`,
			wantText:     "By tenant id, then by time.",
			wantQuestion: "How is data partitioned?",
		},
		{
			name:         "architecture decision shape",
			content:      `{"title":"Storage","decision":"Use object storage with lifecycle rules."}`,
			wantText:     "Use object storage with lifecycle rules.",
			wantQuestion: "Storage",
		},
		{
			name:     "bare string event",
			content:  `"A single sentence artifact."`,
			wantText: "A single sentence artifact.",
		},
		{
			name:     "unknown keys fall back to a sorted concatenation",
			content:  `{"zeta":"last","alpha":"first"}`,
			wantText: "first\n\nlast",
		},
		{
			name:     "non-string values are ignored",
			content:  `{"count":3,"flag":true,"body":"real text"}`,
			wantText: "real text",
		},
		{
			name:     "empty content",
			content:  `{}`,
			wantText: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			text, question := artifactTextAndQuestion(json.RawMessage(tc.content))
			if text != tc.wantText {
				t.Fatalf("text = %q, want %q", text, tc.wantText)
			}
			if question != tc.wantQuestion {
				t.Fatalf("question = %q, want %q", question, tc.wantQuestion)
			}
		})
	}
}

// The fallback must be stable: if the same artifact produced different reference
// text on two runs, its own verdicts could not be compared with each other.
func TestArtifactTextFallbackIsDeterministic(t *testing.T) {
	content := json.RawMessage(`{"delta":"d","beta":"b","alpha":"a"}`)
	first, _ := artifactTextAndQuestion(content)
	for i := 0; i < 20; i++ {
		again, _ := artifactTextAndQuestion(content)
		if again != first {
			t.Fatalf("fallback text changed between calls: %q then %q", first, again)
		}
	}
}

func TestVerificationQueryIsRuneSafe(t *testing.T) {
	short := "a short artifact"
	if got := verificationQuery(short); got != short {
		t.Fatalf("short text should pass through unchanged, got %q", got)
	}

	// Multi-byte characters: cutting by bytes would split a character and put a
	// broken rune into the retrieval query.
	long := strings.Repeat("ज", 500)
	got := verificationQuery(long)
	if len([]rune(got)) != 400 {
		t.Fatalf("query runes = %d, want 400", len([]rune(got)))
	}
	if !strings.HasPrefix(long, got) {
		t.Fatal("query must be a prefix of the artifact, never a mid-character cut")
	}
}

// Code artifacts are validated by compiling them, which is a stronger check than
// claim matching. Pinning the exclusion keeps a future edit from quietly running
// the claim verifier over source text and reporting noise as evidence.
func TestCodeArtifactsAreNotClaimVerified(t *testing.T) {
	if verifiedArtifactTypes["code_artifact_produced"] {
		t.Fatal("code artifacts must not go through claim verification")
	}
	for _, want := range []string{"architecture_decision", "design_section_written"} {
		if !verifiedArtifactTypes[want] {
			t.Fatalf("%s should be claim-verified", want)
		}
	}
}
