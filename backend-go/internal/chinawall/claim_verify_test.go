package chinawall

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestExtractMaterialClaims_SkipsCodeAndShort(t *testing.T) {
	ans := "Short.\n\n" +
		"This is a long enough material claim about PostgreSQL sharding strategy for multi-tenant systems.\n\n" +
		"```go\nfmt.Println(\"hi\")\n```\n\n" +
		"## Heading\n\n" +
		"Another sufficiently long claim that discusses connection pooling under load."
	got := extractMaterialClaims(ans)
	if len(got) < 2 {
		t.Fatalf("expected >=2 material claims, got %d: %v", len(got), got)
	}
	for _, c := range got {
		if strings.Contains(c, "fmt.Println") {
			t.Errorf("code leaked into claims: %q", c)
		}
		if strings.HasPrefix(c, "##") {
			t.Errorf("heading leaked: %q", c)
		}
	}
}

func TestParseClaimVerifyResponse_OK(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cands := []string{
		"PostgreSQL is the primary store for tenant data in this design.",
		"MongoDB is required for every write path.",
	}
	raw := `[
	  {"i":0,"verdict":"supported","confidence":0.9,"justification":"chunk says postgres","span_start":0,"span_end":10,"chunk_ids":["11111111-1111-1111-1111-111111111111"]},
	  {"i":1,"verdict":"refuted","confidence":0.8,"justification":"chunk forbids mongo","span_start":0,"span_end":7,"chunk_ids":["11111111-1111-1111-1111-111111111111"]}
	]`
	got, err := parseClaimVerifyResponse(raw, cands)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Verdict != ClaimSupported || len(got[0].ChunkIDs) != 1 || got[0].ChunkIDs[0] != id {
		t.Errorf("claim0: %+v", got[0])
	}
	if got[1].Verdict != ClaimRefuted {
		t.Errorf("claim1 verdict=%s", got[1].Verdict)
	}
}

func TestParseClaimVerifyResponse_MissingIndexUnverifiable(t *testing.T) {
	cands := []string{"Only claim that is long enough to matter here right now."}
	raw := `[]`
	got, err := parseClaimVerifyResponse(raw, cands)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got[0].Verdict != ClaimUnverifiable {
		t.Errorf("missing idx must be unverifiable, got %s", got[0].Verdict)
	}
}

func TestNormalizeClaimReports_P9NoEvidence(t *testing.T) {
	ans := "Use PostgreSQL with read replicas for the catalogue workload under peak traffic loads."
	reports := []ClaimReport{{
		Claim:      ans,
		Verdict:    ClaimSupported,
		Confidence: 0.9,
		// no ChunkIDs
	}}
	got := normalizeClaimReports(ans, reports)
	if got[0].Verdict != ClaimUnverifiable {
		t.Errorf("supported without chunks must demote, got %s", got[0].Verdict)
	}
	if got[0].SpanStart != 0 || got[0].SpanEnd != len(ans) {
		t.Errorf("span should locate claim in answer: %d-%d", got[0].SpanStart, got[0].SpanEnd)
	}
}

func TestNormalizeClaimReports_Untraceable(t *testing.T) {
	ans := "Something else entirely is written here as the final answer body text."
	reports := []ClaimReport{{
		Claim:      "This claim text is not in the answer at all and should be demoted now.",
		Verdict:    ClaimSupported,
		ChunkIDs:   []uuid.UUID{uuid.MustParse("11111111-1111-1111-1111-111111111111")},
		Confidence: 1,
	}}
	got := normalizeClaimReports(ans, reports)
	if got[0].Verdict != ClaimUnverifiable {
		t.Errorf("untraceable supported must demote, got %s", got[0].Verdict)
	}
}

func TestAnnotateUnverified(t *testing.T) {
	claim := "Redis is mandatory for every session store in multi-region deployments today."
	ans := "Intro.\n\n" + claim + "\n\nOutro."
	reports := []ClaimReport{{
		Claim: claim, Verdict: ClaimUnverifiable, SpanStart: 0, SpanEnd: len(claim),
	}}
	// Labels are REMOVED by product decision: a response must never contain
	// [UNVERIFIED]/[REFUTED] or a verification footer.
	out := annotateUnverified(ans, reports)
	if strings.Contains(out, "[UNVERIFIED]") || strings.Contains(out, "[REFUTED]") {
		t.Errorf("answer must not carry verification labels: %s", out)
	}
	if strings.Contains(out, "Verification:") {
		t.Errorf("answer must not carry a verification footer: %s", out)
	}
	if out != ans {
		t.Errorf("answer must be returned unchanged, got: %s", out)
	}
}

func TestGenericClaimIsNotReLabeledUnverified(t *testing.T) {
	answer := "Trained fact. [GENERIC] React can use a controlled input for the form."
	claims := extractMaterialClaims(answer)
	for _, claim := range claims {
		if strings.Contains(claim, "[GENERIC]") {
			t.Fatalf("generic claim must not be sent to corpus-only verifier: %q", claim)
		}
	}
	reports := []ClaimReport{{
		Claim:     "React can use a controlled input for the form.",
		Verdict:   ClaimSupported, // even if a verifier/model returns this optimistically
		SpanStart: 0, SpanEnd: len("React can use a controlled input for the form."),
	}}
	normalized := normalizeClaimReports(answer, reports)
	if normalized[0].Verdict != ClaimSupported {
		t.Fatalf("explicitly allowed generic content must not be demoted, got %q", normalized[0].Verdict)
	}
	if got := annotateUnverified(answer, normalized); got != answer {
		t.Fatalf("generic-labelled answer should keep its [GENERIC] label, got %q", got)
	}
}

func TestNormalizeVerdict(t *testing.T) {
	if normalizeVerdict("SUPPORTED") != ClaimSupported {
		t.Error("supported")
	}
	if normalizeVerdict("maybe") != ClaimUnverifiable {
		t.Error("default unverifiable")
	}
}

func TestChunksForVerification_PrefersCitations(t *testing.T) {
	a := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	b := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	chunks := []CourseChunk{{ID: a, Text: "a"}, {ID: b, Text: "b"}}
	cites := []Citation{{ChunkID: b}}
	got := chunksForVerification(chunks, cites)
	if len(got) != 1 || got[0].ID != b {
		t.Errorf("got %+v", got)
	}
}
