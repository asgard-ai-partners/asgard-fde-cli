//go:build windows

package cli

import (
	"errors"
	"os/exec"
)

// The Workbench sandbox is Linux; none of this is reached on Windows.

func startDetached(*exec.Cmd) error {
	return errors.New("a background form server is only for the Workbench sandbox, which is Linux")
}

func processAlive(int) bool { return false }

func stopProcess(int) {}
