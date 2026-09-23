//go:build windows

package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
)

// oldSuffix marks the displaced binary. Windows will not let a running .exe be
// deleted, so it is renamed aside and swept up by a later run.
const oldSuffix = ".old"

// replace moves the staged binary over the running one.
//
// **Windows locks a running .exe against being written or deleted, but not
// against being renamed.** So the running one is moved aside first and the new
// one takes its name - the same outcome as the Unix rename, in two steps
// instead of one, with a rollback in between because two steps can fail
// halfway.
//
// The displaced file cannot be removed while this process is executing from
// it. It is left, and `SweepOld` removes it on a later run.
func replace(staged, target string) error {
	old := target + oldSuffix
	_ = os.Remove(old)

	if err := os.Rename(target, old); err != nil {
		return fmt.Errorf("nothing was changed: %w", err)
	}
	if err := os.Rename(staged, target); err != nil {
		// Put it back. Failing here without this leaves no binary at the name
		// at all, which is worse than not updating.
		if back := os.Rename(old, target); back != nil {
			return fmt.Errorf("%s is now at %s and could not be moved back (%v): %w", target, old, back, err)
		}
		return fmt.Errorf("nothing was changed: %w", err)
	}
	return nil
}

// SweepOld removes a binary an earlier update displaced, if it is no longer in
// use. It reports nothing: a failure means the file is still running
// somewhere, which is not this run's problem.
func SweepOld(target string) {
	_ = os.Remove(filepath.Clean(target) + oldSuffix)
}
