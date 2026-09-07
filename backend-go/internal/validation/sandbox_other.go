//go:build !linux

package validation

import "os/exec"

// applyMemoryLimit is a no-op on non-Linux platforms.
// Memory limiting via rlimit is Linux-specific.
// On macOS/Windows (dev machines), validation still runs but without
// the memory cap. Production runs on Linux (Docker), so the cap applies there.
func applyMemoryLimit(_ *exec.Cmd, _ int) {}
