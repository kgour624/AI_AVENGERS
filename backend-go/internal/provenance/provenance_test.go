package provenance

import (
	"encoding/json"
	"testing"
)

func TestSign_DeterministicAndKeySensitive(t *testing.T) {
	msg := []byte(`{"a":1}`)
	s1 := sign([]byte("key-one"), msg)
	s2 := sign([]byte("key-one"), msg)
	if s1 != s2 {
		t.Fatalf("sign not deterministic: %s vs %s", s1, s2)
	}
	if sign([]byte("key-two"), msg) == s1 {
		t.Fatalf("different keys produced the same signature")
	}
	if len(s1) != 64 { // hex of sha256
		t.Fatalf("signature length=%d want 64", len(s1))
	}
}

func TestVerifyChain_TamperDetected(t *testing.T) {
	key := []byte("secret")
	chain := `{"output_type":"chat_message","output_id":"x"}`
	sig := sign(key, []byte(chain))

	if !verifyChain(key, chain, sig) {
		t.Fatal("valid chain failed verification")
	}
	if verifyChain(key, chain+` `, sig) {
		t.Fatal("tampered chain passed verification")
	}
	if verifyChain([]byte("other"), chain, sig) {
		t.Fatal("wrong key passed verification")
	}
	if verifyChain(key, chain, "") {
		t.Fatal("empty signature passed verification")
	}
}

func TestContentHash_Stable(t *testing.T) {
	if contentHash("hello") != contentHash("hello") {
		t.Fatal("content hash not stable")
	}
	if contentHash("hello") == contentHash("hello!") {
		t.Fatal("different content produced the same hash")
	}
	if len(contentHash("hello")) != 64 {
		t.Fatalf("hash length=%d want 64", len(contentHash("hello")))
	}
}

func TestNormalizeJSON_DefaultsToEmptyArray(t *testing.T) {
	if got := string(normalizeJSON(nil)); got != "[]" {
		t.Errorf("nil → %q want []", got)
	}
	if got := string(normalizeJSON([]int{})); got != "[]" {
		t.Errorf("empty slice → %q want []", got)
	}
	// A concrete value marshals through.
	got := normalizeJSON([]map[string]string{{"chunk_id": "abc"}})
	var back []map[string]string
	if err := json.Unmarshal(got, &back); err != nil || len(back) != 1 || back[0]["chunk_id"] != "abc" {
		t.Errorf("round-trip failed: %s (err=%v)", got, err)
	}
}

func TestIsArtifactEventType(t *testing.T) {
	for _, yes := range []string{"architecture_decision", "code_artifact_produced", "design_section_written"} {
		if !IsArtifactEventType(yes) {
			t.Errorf("%q should be an artifact type", yes)
		}
	}
	for _, no := range []string{"artifact_approved", "artifact_blocked", "question_to_expert", ""} {
		if IsArtifactEventType(no) {
			t.Errorf("%q should NOT be an artifact type", no)
		}
	}
}
