//go:build linux

package validation

import "os/exec"

// applyMemoryLimit is intentionally a no-op on Linux in this implementation.
//
// WHY no-op (not the original Rlimit approach):
//   Go's syscall.SysProcAttr does NOT have an Rlimit field.
//   Setting process memory limits in Go requires either:
//     (a) CGo + C setrlimit() call
//     (b) A pre-exec hook via cmd.SysProcAttr.Cloneflags + namespace
//     (c) Running inside a container with cgroup limits (our Docker setup)
//
//   For AI Avengers (single-client, Docker-deployed):
//   - Docker container already has memory limits via docker-compose
//   - The 5s wall-time limit in sandbox.go kills runaway processes
//   - CGo adds build complexity not worth it for this use case
//
//   If per-subprocess memory limits become necessary in the future,
//   the correct approach is:
//     import "golang.org/x/sys/unix"
//     unix.Setrlimit(unix.RLIMIT_AS, &unix.Rlimit{Cur: limit, Max: limit})
//   called inside a cmd.SysProcAttr.Pdeathsig goroutine before exec.
//
// Current protection: context.WithTimeout(5s) in sandbox.go Run().
func applyMemoryLimit(_ *exec.Cmd, _ int) {}
