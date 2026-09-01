package usecase

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// longNumber is the shape of an Asgard platform identifier: a workspace id is a
// decimal snowflake of around 19 digits. Nothing shipped here needs a run of
// digits that long, so a match is a real identifier pasted in as an example.
var longNumber = regexp.MustCompile(`[0-9]{12,}`)

// TestEmbeddedFilesCarryNoRealIdentifiers guards the same mistake that reached
// a stage prompt once: a live workspace id used as an example. Everything under
// this package's embed is compiled into a binary that ships to every
// engagement, so an example there is published to all of them - and the next
// reader's most natural move with an example id is to paste it, which binds
// their repository to another customer's workspace.
//
// Describe the shape of the value instead of giving one.
func TestEmbeddedFilesCarryNoRealIdentifiers(t *testing.T) {
	err := fs.WalkDir(extracts, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		body, err := extracts.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(body), "\n") {
			if m := longNumber.FindString(line); m != "" {
				t.Errorf("%s has what looks like a real identifier (%s):\n  %s",
					path, m, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
}
