package workflow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func TestInstallNodeDependenciesRunsNPMWithoutLifecycleScripts(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm verification requires npm on PATH")
	}

	workspace := t.TempDir()
	writeTestFile(t, workspace, "package.json", `{"name":"verify-fixture","version":"1.0.0","scripts":{"postinstall":"node -e \"require('fs').writeFileSync('postinstall-ran','yes')\""}}`)

	runner := &AiderRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, ok := runner.installNodeDependencies(ctx, workspace)
	if !ok {
		t.Fatalf("installNodeDependencies() unavailable: %+v", result)
	}
	if !hasWorkspaceDir(workspace, "node_modules") {
		t.Fatal("npm install did not create node_modules")
	}
	if _, err := os.Stat(filepath.Join(workspace, "postinstall-ran")); !os.IsNotExist(err) {
		t.Fatalf("postinstall script ran despite --ignore-scripts; stat err=%v", err)
	}
	writeTestFile(t, workspace, "package.json", `{"name":"verify-fixture","version":"1.0.0","dependencies":{"left-pad":"1.3.0"}}`)
	result, ok = runner.installNodeDependencies(ctx, workspace)
	if !ok {
		t.Fatalf("installNodeDependencies() after manifest change unavailable: %+v", result)
	}
}

func TestInstallNodeDependenciesReportsTimeout(t *testing.T) {
	workspace := t.TempDir()
	writeTestFile(t, workspace, "package.json", `{"name":"timeout-fixture","version":"1.0.0"}`)

	runner := &AiderRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, ok := runner.installNodeDependencies(ctx, workspace)
	if ok || result.Status != VerifyUnavailable {
		t.Fatalf("installNodeDependencies() = (%+v, %v), want unavailable", result, ok)
	}
}

func TestNodeDependencyFingerprintChangesWithLockfile(t *testing.T) {
	workspace := t.TempDir()
	writeTestFile(t, workspace, "package.json", `{"name":"fingerprint-fixture","version":"1.0.0"}`)
	first, err := nodeDependencyFingerprint(workspace)
	if err != nil {
		t.Fatalf("nodeDependencyFingerprint() error = %v", err)
	}
	writeTestFile(t, workspace, "package-lock.json", `{"lockfileVersion":3,"packages":{}}`)
	second, err := nodeDependencyFingerprint(workspace)
	if err != nil {
		t.Fatalf("nodeDependencyFingerprint() after lockfile error = %v", err)
	}
	if first == second {
		t.Fatal("fingerprint did not change after lockfile was added")
	}
}
