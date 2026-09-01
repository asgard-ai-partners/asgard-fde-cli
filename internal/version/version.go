// Package version exposes build information for `asgard-cli version` and --version.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// GoReleaser injects these at build time with -ldflags -X. A plain `go build`
// or `go install` leaves them at dev, and debug.ReadBuildInfo fills the gap.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

// Info identifies a single build.
type Info struct {
	Version  string `json:"version"`
	Commit   string `json:"commit,omitempty"`
	Date     string `json:"date,omitempty"`
	Go       string `json:"go"`
	Platform string `json:"platform"`
}

// Get returns the build information of the running binary.
func Get() Info {
	info := Info{
		Version:  version,
		Commit:   commit,
		Date:     date,
		Go:       runtime.Version(),
		Platform: runtime.GOOS + "/" + runtime.GOARCH,
	}

	// A binary built without ldflags (`go install ...@latest`) can still take its
	// version and commit from the module and VCS metadata.
	if info.Version == "dev" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if v := bi.Main.Version; v != "" && v != "(devel)" {
				info.Version = v
			}
			for _, s := range bi.Settings {
				switch s.Key {
				case "vcs.revision":
					info.Commit = s.Value
				case "vcs.time":
					info.Date = s.Value
				case "vcs.modified":
					if s.Value == "true" {
						info.Version += "-dirty"
					}
				}
			}
		}
	}

	return info
}

// String returns the one-line summary used by --version.
func (i Info) String() string {
	s := i.Version
	if i.Commit != "" {
		c := i.Commit
		if len(c) > 7 {
			c = c[:7]
		}
		s += fmt.Sprintf(" (%s)", c)
	}
	return fmt.Sprintf("%s %s %s", s, i.Platform, i.Go)
}
