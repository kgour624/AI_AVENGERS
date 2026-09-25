package workflow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

// These tests cover the safety-critical half of 3E: an expert must not be able
// to write a client file that was never approved for reading. The rule is
// enforced in removeProtectedChanges, so each case pins one branch of it.

func gitAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for workspace merger tests")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v (%s)", args, err, string(output))
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// newExpertWorkspace creates a committed git workspace with two tracked files,
// then edits both without committing — the exact state an expert leaves behind.
func newExpertWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")

	writeFile(t, dir, "approved.go", "package main\n")
	writeFile(t, dir, "protected.go", "package main\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "seed")

	writeFile(t, dir, "approved.go", "package main // edited with permission\n")
	writeFile(t, dir, "protected.go", "package main // edited WITHOUT permission\n")
	return dir
}

func TestRemoveProtectedChangesDropsOnlyUnapprovedEdits(t *testing.T) {
	gitAvailable(t)
	dir := newExpertWorkspace(t)

	merger := NewWorkspaceMerger(zap.NewNop())
	merger.SetProtectedPathChecker(func(_ context.Context, path string) bool {
		return path == "protected.go"
	})

	merger.removeProtectedChanges(context.Background(), dir)

	if _, err := os.Stat(filepath.Join(dir, "approved.go")); err != nil {
		t.Fatalf("approved file was removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "protected.go")); !os.IsNotExist(err) {
		t.Fatalf("protected file must be removed before merge; stat err = %v", err)
	}
}

func TestRemoveProtectedChangesIsNoOpWithoutChecker(t *testing.T) {
	gitAvailable(t)
	dir := newExpertWorkspace(t)

	// No checker installed = the scratch behaviour, which must not touch files.
	NewWorkspaceMerger(zap.NewNop()).removeProtectedChanges(context.Background(), dir)

	for _, name := range []string{"approved.go", "protected.go"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s must survive when nothing is protected: %v", name, err)
		}
	}
}

func TestDirtyFilesFindsUncommittedEdits(t *testing.T) {
	gitAvailable(t)
	dir := newExpertWorkspace(t)

	files := NewWorkspaceMerger(zap.NewNop()).dirtyFiles(context.Background(), dir)

	seen := map[string]bool{}
	for _, f := range files {
		seen[f] = true
	}
	// The guard is worthless if it only saw committed work: an expert that never
	// commits is the common case, so uncommitted edits MUST be listed.
	for _, want := range []string{"approved.go", "protected.go"} {
		if !seen[want] {
			t.Fatalf("dirtyFiles() = %v, missing uncommitted edit %s", files, want)
		}
	}
}

func TestDefaultMaxWaveTasksIsBounded(t *testing.T) {
	// A10: the wave loop must never be unbounded. A zero or negative limit here
	// would make the semaphore channel zero-capacity and deadlock the wave, so
	// the default is asserted rather than assumed.
	if maxWaveTasks < 1 {
		t.Fatalf("maxWaveTasks = %d, want >= 1", maxWaveTasks)
	}
	if defaultMaxWaveTasks != maxWaveTasks && os.Getenv("MAX_WAVE_CONCURRENCY") == "" {
		t.Fatalf("maxWaveTasks = %d without an env override, want the default %d",
			maxWaveTasks, defaultMaxWaveTasks)
	}
}
