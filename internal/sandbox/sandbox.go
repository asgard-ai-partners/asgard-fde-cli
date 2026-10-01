// Package sandbox says whether this process runs in the Workbench assistant's
// sandbox on the platform, for the packages that behave differently there and
// cannot reach the one that holds the identity (internal/auth imports the
// browser, so the browser cannot import it).
package sandbox

import "os"

// EnvMode is set to "true" by the Workbench assistant's sandbox image. Any
// other value is as if it were absent.
const EnvMode = "ASGARD_SANDBOX_MODE"

// Enabled reports whether this process runs in the Workbench sandbox.
func Enabled() bool { return os.Getenv(EnvMode) == "true" }
