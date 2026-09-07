//go:build linux

package validation

import (
	"os/exec"
	"syscall"
)

// applyMemoryLimit sets a virtual memory limit on the subprocess via rlimit.
// Only compiled on Linux. Called from sandbox.go's Run() function.
//
// WHY RLIMIT_AS (address space) not RLIMIT_DATA:
//   RLIMIT_AS limits total virtual memory including stack + heap + mmap.
//   This is the most effective limit for preventing runaway processes.
//   RLIMIT_DATA only limits the data segment, not mmap allocations
//   (which is how most modern allocators work).
func applyMemoryLimit(cmd *exec.Cmd, memoryMB int) {
	limitBytes := uint64(memoryMB) * 1024 * 1024
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Rlimit = []syscall.Rlimit{
		{
			Type: syscall.RLIMIT_AS,
			Cur:  limitBytes,
			Max:  limitBytes,
		},
	}
}
