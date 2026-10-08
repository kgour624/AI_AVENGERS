package vacuum

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestClassifyOpenErrorSeparatesGoneFromFlaky pins the decision that stops a dead
// job from being re-scheduled forever: a definitively missing object must land on
// INPUT_MISSING with NO retry_after, while a transient failure keeps the
// retryable STORAGE_OPEN code. The two codes are what the admin screen branches
// on (Retry vs "source missing — re-upload"), so they must never collapse.
func TestClassifyOpenErrorSeparatesGoneFromFlaky(t *testing.T) {
	now := time.Date(2026, 10, 8, 3, 0, 0, 0, time.UTC)

	// Build the missing-file error by actually opening a missing path instead of
	// fabricating one: os.Open returns the *fs.PathError wrapping ENOENT that the
	// pipeline really hands to classifyOpenError, so this test cannot drift from
	// the shape production produces.
	_, openErr := os.Open(filepath.Join(t.TempDir(), "no-such-input.extracted.md"))
	if openErr == nil {
		t.Fatal("opening a path that does not exist must fail")
	}

	goneCode, goneMsg, goneRetry := classifyOpenError(openErr, now)
	if goneCode != "INPUT_MISSING" {
		t.Errorf("missing object: code = %q, want INPUT_MISSING", goneCode)
	}
	if goneRetry != nil {
		t.Errorf("missing object: retry_after = %v, want nil — nothing may re-schedule a job whose input no longer exists", goneRetry)
	}
	if goneMsg == "" {
		t.Error("missing object: message must not be empty")
	}
	if !strings.Contains(goneMsg, "upload") {
		t.Errorf("missing object: message must point at the real fix (re-upload), got %q", goneMsg)
	}
	if !strings.Contains(goneMsg, "no-such-input.extracted.md") {
		t.Errorf("missing object: message must carry the underlying error for triage, got %q", goneMsg)
	}

	flakyErr := errors.New("connection reset while reading object")
	flakyCode, flakyMsg, flakyRetry := classifyOpenError(flakyErr, now)
	if flakyCode != "STORAGE_OPEN" {
		t.Errorf("transient failure: code = %q, want STORAGE_OPEN", flakyCode)
	}
	if flakyRetry == nil {
		t.Fatal("transient failure: retry_after must be set so the normal retry path picks the row back up")
	}
	if want := now.Add(60 * time.Second); !flakyRetry.Equal(want) {
		t.Errorf("transient failure: retry_after = %v, want %v", flakyRetry, want)
	}
	if !strings.Contains(flakyMsg, "connection reset") {
		t.Errorf("transient failure: message must carry the underlying error, got %q", flakyMsg)
	}
	if goneCode == flakyCode {
		t.Error("the two failure kinds must stay distinguishable — the UI renders Retry for one and re-upload for the other")
	}
}

// TestMissingOnlyReportsDefinitiveAbsence guards the guard itself: the retry
// endpoints may only claim "this can never succeed" when the object is really
// gone. Reporting a merely unreadable object as missing would send the admin to
// re-upload a file that is sitting right there.
func TestMissingOnlyReportsDefinitiveAbsence(t *testing.T) {
	root := t.TempDir()
	s := NewFSStorage(root)

	storedKey := filepath.Join("vacuum", "job-1", "file.extracted.md")
	full := filepath.Join(root, storedKey)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("preparing stored object: %v", err)
	}
	if err := os.WriteFile(full, []byte("cleaned text"), 0o644); err != nil {
		t.Fatalf("writing stored object: %v", err)
	}

	if s.Missing(storedKey) {
		t.Error("a stored object must not be reported missing")
	}
	if !s.Missing(filepath.Join("vacuum", "job-1", "gone.extracted.md")) {
		t.Error("an absent object must be reported missing")
	}
	if s.Missing("inline:hello world") {
		t.Error("inline: keys carry their payload in the key, so they can never be missing")
	}
}
