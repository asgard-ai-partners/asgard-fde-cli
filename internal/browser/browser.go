// Package browser opens a URL in the desktop's browser.
package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open asks the desktop to open a URL.
//
// There is no dependency for this because there is nothing to depend on: each
// platform has one command, they have not changed in a decade, and a library
// would be a third-party package in a binary that currently has two.
//
// A failure here is never fatal to the flow that called it. Every caller
// prints the URL as well, so a machine with no desktop - a container, a CI
// runner, an SSH session - falls back to the person opening it themselves,
// which is exactly what --no-browser asks for explicitly.
func Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		// rundll32 rather than `cmd /c start`, which treats the first quoted
		// argument as a window title and mangles a URL containing `&`.
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: %w", cmd.Path, err)
	}
	// The browser outlives this process, so the child is deliberately not
	// waited on; releasing it stops the CLI holding a zombie until it exits.
	go func() { _ = cmd.Wait() }()
	return nil
}
