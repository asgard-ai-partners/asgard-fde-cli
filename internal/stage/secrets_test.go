package stage

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// longNumber is the shape of a platform identifier: a workspace id is a decimal
// snowflake of around 19 digits. Nothing a prompt has to say needs a run of
// digits that long, so anything matching is a real identifier that got pasted in
// as an example.
var longNumber = regexp.MustCompile(`[0-9]{12,}`)

// TestPromptsCarryNoRealIdentifiers exists because one did.
//
// A workspace id looks like a harmless example - it is just a number - and it
// went into 00-init.md as one. It is a live production identifier for a real
// customer, and the prompts are embedded in a binary that ships to every
// engagement, so an example there is published to all of them. Worse, the next
// FDE's most natural move with an example id is to paste it, which binds their
// repository to somebody else's workspace.
//
// Describe the shape instead. "Around 19 digits, not a UUID" is everything the
// reader needs.
func TestPromptsCarryNoRealIdentifiers(t *testing.T) {
	err := fs.WalkDir(prompts, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		body, err := prompts.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(body), "\n") {
			if m := longNumber.FindString(line); m != "" {
				t.Errorf("%s has what looks like a real identifier (%s):\n  %s\n"+
					"Describe the shape of the value rather than giving one.", path, m, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
}
