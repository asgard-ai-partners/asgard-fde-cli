//go:build !windows

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

// startDetached starts cmd in a session of its own, so that it outlives the
// command that started it - and the shell that ran that command, which is
// what the Workbench sandbox's agent does with every command it runs.
func startDetached(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	// Not waited on: it is not this process's child to reap once this exits,
	// and waiting would tie the two together again.
	return cmd.Process.Release()
}

// processAlive reports whether pid is a running process.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// stopProcess asks pid to stop.
func stopProcess(pid int) {
	if p, err := os.FindProcess(pid); err == nil {
		_ = p.Signal(syscall.SIGTERM)
	}
}
