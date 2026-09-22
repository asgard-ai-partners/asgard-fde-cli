//go:build !windows

package selfupdate

import (
	"fmt"
	"os"
)

// replace moves the staged binary over the running one.
//
// **Replacing a running executable is a rename here, and that is safe.** The
// process already running holds the old inode open and keeps running from it;
// the name now points at the new file, so the next invocation is the new
// version. Writing into the existing file instead would corrupt the binary
// under the process executing it.
//
// The staging directory is beside the target, so this is a rename within one
// filesystem and either happens or does not. There is no window in which the
// name points at a partial file.
func replace(staged, target string) error {
	// Carry the old file's mode across: a binary installed as 0755 root-owned
	// stays that, and one somebody made group-writable on purpose stays that
	// too. The staged file is 0755 as a floor rather than as a decision.
	if info, err := os.Stat(target); err == nil {
		if err := os.Chmod(staged, info.Mode().Perm()); err != nil {
			return err
		}
	}
	if err := os.Rename(staged, target); err != nil {
		return fmt.Errorf("nothing was changed: %w", err)
	}
	return nil
}

// SweepOld does nothing here, and exists so the caller needs no build tag.
//
// Windows has to displace the running binary to a second name before the new
// one can take the first, and that file cannot be removed until the process
// executing it exits. A rename leaves nothing behind.
func SweepOld(string) {}
