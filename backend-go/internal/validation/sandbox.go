package validation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
//
//	Run(["gofmt", "-l", "main.go"], "/tmp/artifact_abc", nil)
//	→ exec.CommandContext(5s, "gofmt", "-l", "main.go")
//	→ cmd.Dir = /tmp/artifact_abc
//	→ cmd.Run() → stdout="main.go\n" (file needs formatting)
//	→ RunResult{Stdout: "main.go\n", ExitCode: 0}
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

	// Minimal safe environment — see MinimalEnv.
	//
	// This used to build the list inline with os.Getenv for PATH, HOME, GOPATH
	// and GOROOT. Two problems with that, both fixed by moving to MinimalEnv:
	// GOCACHE was missing (the Dockerfile sets it precisely because the default
	// location is not writable for the non-root user), and os.Getenv cannot tell
	// "unset" from "empty", so an unset GOROOT was passed to the child as
	// GOROOT= rather than left absent.
	cmd.Env = MinimalEnv(env...)

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

// subprocessEnvAllowlist is every environment variable a build or test tool is
// allowed to see.
//
// Why an allowlist, and why this matters more than it looks:
//
// The api process environment holds, from docker-compose.yml: JWT_SECRET,
// ENCRYPTION_KEY (the key that decrypts clients' git tokens), OPENROUTER_API_KEY,
// CAVOTI_API_KEY, AIDER_PROXY_TOKEN, GITHUB_CLIENT_SECRET, GITLAB_CLIENT_SECRET,
// and DATABASE_URL with its password.
//
// Every one of those was being handed to subprocesses that run code we did not
// write. `go test ./...` runs test code a model generated;
// `npm run build` runs whatever scripts a package.json declares — and since §17
// the package.json in question can belong to a CLIENT's cloned repository, which
// is third-party code with third-party dependencies. A single postinstall or test
// file reading os.Environ() would have every credential the platform holds.
//
// An allowlist and not a blocklist: a blocklist has to be updated every time
// someone adds a secret to the compose file, and the failure mode of forgetting
// is silent.
//
// Absent on purpose and worth stating: there is no network restriction here.
// Limiting that needs namespaces, not an environment change.
var subprocessEnvAllowlist = []string{
	// Basics every tool needs.
	"PATH", "HOME", "TMPDIR", "LANG", "LC_ALL", "TZ",
	// Go. GOCACHE and GOFLAGS are load-bearing: the Dockerfile sets GOCACHE
	// because the default is not writable for the non-root user, and GOFLAGS
	// carries -mod=mod.
	"GOCACHE", "GOPATH", "GOMODCACHE", "GOROOT", "GOTMPDIR", "GOFLAGS",
	"GOPROXY", "GOSUMDB", "GONOSUMDB", "GOPRIVATE", "GOOS", "GOARCH", "CGO_ENABLED",
	// Node. npm_config_cache is set by the Dockerfile for the same reason as
	// GOCACHE.
	"npm_config_cache", "NODE_ENV", "NODE_PATH",
	// Python.
	"PYTHONPATH", "PYTHONDONTWRITEBYTECODE",
}

// MinimalEnv returns the environment a build or test subprocess may see: the
// allowlisted variables that are actually set, plus any extras the caller adds.
//
// os.LookupEnv rather than os.Getenv, so a variable that is not set is left
// ABSENT instead of being passed through as NAME=. The two are not equivalent to
// every tool — an empty GOROOT is not the same as no GOROOT — and "unset" is the
// state the parent process is actually in.
func MinimalEnv(extra ...string) []string {
	out := make([]string, 0, len(subprocessEnvAllowlist)+len(extra))
	for _, name := range subprocessEnvAllowlist {
		if value, ok := os.LookupEnv(name); ok {
			out = append(out, name+"="+value)
		}
	}
	return append(out, extra...)
}

// WriteTempFile writes content to a temp file in dir and returns the path.
// Caller is responsible for cleanup (defer os.Remove(path)).
//
// filename may contain directory separators. Artifacts carry their real
// path inside the project ("foundational/logger/logger.go"), not a bare
// base name. os.WriteFile does not create intermediate directories, so
// every artifact in a subdirectory failed with ENOENT and the pipeline
// reported it as "validation failed" — the generated code was fine, the
// sandbox was. MkdirAll on the parent removes that whole false-failure class.
func WriteTempFile(dir, filename, content string) (string, error) {
	path := filepath.Join(dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", fmt.Errorf("sandbox: make temp subdir: %w", err)
	}
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
