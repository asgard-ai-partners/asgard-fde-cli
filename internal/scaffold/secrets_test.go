package scaffold

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
//
// allowed lists the exceptions, each of which has to be a value that is
// *functional* rather than illustrative - something the generated command will
// not work without. An example never qualifies. Adding an entry here is a
// decision to publish that value to every customer repository, so it needs a
// reason written next to it.
var allowed = map[string]string{
	// Asgard's own AWS account, in the kubectl context ARN the verification
	// step needs. Every customer runs against Asgard's clusters, separated by
	// namespace rather than by cluster, so this is not one customer's value
	// leaking into another's repo. Whether it should be a parameter at all is
	// the open question recorded in TASK.md under "Are the EKS cluster names
	// customer-specific?".
	"698306514474": "Asgard's own AWS account, in the EKS context ARN",
}

func TestEmbeddedFilesCarryNoRealIdentifiers(t *testing.T) {
	err := fs.WalkDir(templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		body, err := templates.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(body), "\n") {
			for _, m := range longNumber.FindAllString(line, -1) {
				if _, ok := allowed[m]; ok {
					continue
				}
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
