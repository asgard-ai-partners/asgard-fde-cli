package platform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EnvSideEffectFile names a file this CLI stamps with the current time after
// every call that changed something on the platform.
//
// It is how a screen learns that an agent changed what it shows. In the
// Workbench assistant's sandbox the agent drives this CLI, the member watches
// the Workbench page, and nothing tells the page that the agent just opened an
// issue - so the sandbox image sets this variable, and the page watches the file
// through the sandbox's file watch (asgard-odin-pm TASK-074, UC-070). Unset, the
// CLI writes nothing.
const EnvSideEffectFile = "ASGARD_CLI_SIDE_EFFECT_TIMESTAMP_FILE"

// sideEffectStamp is the file's content: one JSON object. Only `at` today, but
// an object rather than a bare timestamp so a field can be added without
// breaking a reader - a reader relies on `at` and ignores what it does not know.
type sideEffectStamp struct {
	At string `json:"at"`
}

// now is replaced by tests.
var now = time.Now

// NoteSideEffect stamps EnvSideEffectFile with the current time, if it is set.
//
// Called by the request paths for every call marked sideEffect that the
// platform answered 2xx - the change has happened even when the body that
// follows cannot be read - and directly by a command whose change the platform
// makes on its own, such as the connection "pipeline connect --continue" sees
// appear.
//
// **It never fails the command.** The change it reports has already been made;
// a stamp that could not be written is a screen that refreshes late, so it is a
// warning on stderr and nothing more. A missing directory or file is created:
// the watcher is not this CLI's to fix, but the file it watches is.
func NoteSideEffect() {
	path := os.Getenv(EnvSideEffectFile)
	if path == "" {
		return
	}
	if err := writeSideEffectStamp(path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not record the change in %s (%s): %v\n", path, EnvSideEffectFile, err)
	}
}

func writeSideEffectStamp(path string) error {
	body, err := json.Marshal(sideEffectStamp{At: now().UTC().Format("2006-01-02T15:04:05.000Z07:00")})
	if err != nil {
		return err
	}
	body = append(body, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// A temporary file renamed over the target, so a reader never sees half a
	// write. The sandbox's watch follows a file by name within its directory,
	// so a replace is seen like any other write.
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
