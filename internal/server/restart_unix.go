//go:build !windows

package server

import (
	"os"
	"syscall"
)

// restart replaces the current process image with the updated binary.
// On Unix, syscall.Exec keeps the same PID (important for Docker/systemd),
// so the process manager sees no interruption.
func restart(execPath string) {
	_ = syscall.Exec(execPath, os.Args, os.Environ())
	// If Exec somehow fails, fall back to exit so a process manager can restart.
	os.Exit(0)
}
