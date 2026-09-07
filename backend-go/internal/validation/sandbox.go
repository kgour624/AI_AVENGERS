package validation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// SandboxConfig controls resource limits for subprocess execution.
type SandboxConfig struct {
	// WallTimeLimit is the maximum wall-clock time for the subprocess.
	// Default: 5 seconds.
	WallTimeLimit time.Duration

	// MemoryLimitMB is the memory limit in megabytes.
	// Applied via ulimit on Linux. Ignored on other platforms.
	// Default: 512 MB.
	MemoryLimitMB int
}

// DefaultSandboxConfig returns production-safe defaults.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		WallTimeLimit: 5 * time.Second,
		MemoryLimitMB: 512,
	}
}

// RunResult is the output of a sandboxed subprocess run.
type RunResult struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	DurationMs int64
	TimedOut   bool
}

// Run executes a command in a sandboxed subprocess.
//
// workDir: directory where the command runs (artifact temp dir).
// command: executable + args, e.g. ["go", "vet", "./..."].
// env: additional environment variables (merged with minimal safe env).
//
// Mental execution:
//   Run(["gofmt", "-l", "main.go"], "/tmp/artifact_abc", nil)
//   → exec.CommandContext(5s, "gofmt", "-l", "main.go")
//   → cmd.Dir = /tmp/artifact_abc
//   → cmd.Run() → stdout="main.go\n" (file needs formatting)
//   → RunResult{Stdout: "main.go\n", ExitCode: 0}
func Run(ctx context.Context, cfg SandboxConfig, workDir string, command []string, env []string) (*RunResult, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf("sandbox: command is empty")
	}

	// Apply wall-time limit
	limit := cfg.WallTimeLimit
	if limit <= 0 {
		limit = 5 * time.Second
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, command[0], command[1:]...)
	cmd.Dir = workDir

	// Minimal safe environment: only PATH + any caller-provided vars.
	// WHY not inherit full env: prevents leaking secrets (API keys, tokens)
	// from the server process into the validator subprocess.
	cmd.Env = append([]string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"), // needed by go tools
		"GOPATH=" + os.Getenv("GOPATH"),
		"GOROOT=" + os.Getenv("GOROOT"),
	}, env...)

	// Apply memory limit on Linux via ulimit
	// WHY only Linux: ulimit SysProcAttr is Linux-specific.
	// On macOS/Windows (dev machines), we skip it gracefully.
	if runtime.GOOS == "linux" && cfg.MemoryLimitMB > 0 {
		applyMemoryLimit(cmd, cfg.MemoryLimitMB)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	duration := time.Since(start).Milliseconds()

	result := &RunResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: duration,
	}

	// Determine exit code
	if runErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			result.TimedOut = true
			result.ExitCode = -1
			return result, nil // timeout is not a system error
		}
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil // non-zero exit is a validation failure, not a system error
		}
		// System error (binary not found, permission denied, etc.)
		return nil, fmt.Errorf("sandbox: exec failed: %w", runErr)
	}

	result.ExitCode = 0
	return result, nil
}

// WriteTempFile writes content to a temp file in dir and returns the path.
// Caller is responsible for cleanup (defer os.Remove(path)).
func WriteTempFile(dir, filename, content string) (string, error) {
	path := dir + "/" + filename
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("sandbox: write temp file: %w", err)
	}
	return path, nil
}

// MakeTempDir creates a temporary directory for artifact validation.
// Caller must defer os.RemoveAll(dir).
func MakeTempDir() (string, error) {
	dir, err := os.MkdirTemp("", "ai_avengers_validation_*")
	if err != nil {
		return "", fmt.Errorf("sandbox: make temp dir: %w", err)
	}
	return dir, nil
}
