package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// What a managed region means for the file it is in, and the four shapes a
// difference can take.
//
// **The region is the shipped material and the rest of the file is not.** Both
// Write and InspectShipped decide through here, and they used to decide
// differently: Write carried the rule and the inspection did not, so
// `asgard-cli init` reported nothing to do about AGENTS.md while
// `asgard-cli gate` reported that it would be updated - one file, one
// repository, two answers, and no way to tell which was right by reading them.
func TestClassifyTreatsTheRegionAsTheShippedMaterial(t *testing.T) {
	region := func(body string) string {
		return "top\n<!-- asgard-cli:managed:start -->\n" + body + "\n<!-- asgard-cli:managed:end -->\ntail\n"
	}
	rec := Entry{CLIVersion: "0.9.0"}

	for _, c := range []struct {
		name           string
		disk, rendered string
		want           Status
	}{
		{
			// The engagement answered a TODO above the marker. Nothing this
			// CLI owns has moved, so there is nothing to report and nothing to
			// write - and reporting `edited` here would fire on every
			// repository whose AGENTS.md has been filled in, which is all of
			// the real ones.
			name:     "answered above the marker",
			disk:     "the engagement's answer\n" + region("ours"),
			rendered: "top\n" + region("ours"),
			want:     Skipped,
		},
		{
			// The CLI's half moved. mergeManaged replaces the region and
			// leaves everything around it, so this is a refresh rather than
			// anything needing --force.
			name:     "the region moved",
			disk:     "the engagement's answer\n" + region("ours, v1"),
			rendered: "top\n" + region("ours, v2"),
			want:     Updated,
		},
		{
			// A template that drops its marker is not a licence to delete what
			// accumulated behind it.
			name:     "the render has no region any more",
			disk:     "the engagement's answer\n" + region("ours"),
			rendered: "a whole new file with no marker\n",
			want:     Skipped,
		},
		{
			// No region anywhere: the ordinary whole-file comparison, and this
			// record says this CLI wrote what is on disk.
			name:     "no region, and the record covers it",
			disk:     "ours, v1\n",
			rendered: "ours, v2\n",
			want:     Updated,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := rec
			if c.want == Updated && !strings.Contains(c.disk, "managed:start") {
				// The whole-file path needs the record to match the disk;
				// the region path never reads the digest at all.
				e.Digest = digest([]byte(c.disk))
			}
			if got := classify([]byte(c.disk), []byte(c.rendered), e, "0.9.0"); got != c.want {
				t.Errorf("classify() = %s, want %s", got, c.want)
			}
		})
	}
}

var interpolation = regexp.MustCompile(`<<\s*(range|if)\b`)

// Nothing above a managed region marker may be GENERATED.
//
// **Everything above the marker is never overwritten**, which is what makes it
// safe for the engagement to write there - and it makes it the wrong place for
// content derived from the repository, because nothing will ever bring it back
// into step and nothing reports that it is out. AGENTS.md carried a second copy
// of the project list there for a release: `asgard-cli project add` wrote the
// chart, README's table refreshed inside its own region, and AGENTS.md went on
// saying "none yet" with `init` reporting the file as already present.
//
// A `range` or an `if` is the shape that goes wrong, because it is derived from
// a set that grows. A single interpolated name is not: README's `# <RepoName>`
// title is a seed the engagement may rename deliberately, and a stale one is
// visible to anybody reading the line.
func TestNoGeneratedListAboveAManagedRegion(t *testing.T) {
	err := fs.WalkDir(templates, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := fs.ReadFile(templates, path)
		if err != nil {
			return err
		}
		loc := managedRegion.FindIndex(body)
		if loc == nil {
			return nil
		}
		if m := interpolation.Find(body[:loc[0]]); m != nil {
			t.Errorf("%s generates content above its managed region (%q); it will never be "+
				"refreshed and nothing will report it stale", path, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A second scaffold run leaves everything above the marker exactly as the
// engagement left it.
//
// This is the incident the rule exists for, over the real templates rather
// than a fixture: AGENTS.md gained a region, the whole-file branch took the
// file because the merge had already brought the region into step, and the
// answers somebody had written above the marker went with it.
func TestWriteKeepsWhatIsAboveTheMarker(t *testing.T) {
	root := t.TempDir()
	if _, err := Write(root, nil, false); err != nil {
		t.Fatal(err)
	}

	agents := filepath.Join(root, "AGENTS.md")
	before, err := os.ReadFile(agents)
	if err != nil {
		t.Fatal(err)
	}
	loc := managedRegion.FindIndex(before)
	if loc == nil {
		t.Fatal("the scaffolded AGENTS.md has no managed region, so this test proves nothing")
	}

	const answer = "- **Projects**: billing reads the ERP's SO table. THE ENGAGEMENT WROTE THIS.\n"
	edited := append(append([]byte(answer), before[:loc[0]]...), before[loc[0]:]...)
	if err := os.WriteFile(agents, edited, 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := Write(root, []string{"billing"}, false)
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(agents)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), answer) {
		t.Error("the second run wrote away what was above the marker")
	}
	for _, r := range results {
		if r.Path == "AGENTS.md" && r.Status != Skipped && r.Status != Updated {
			t.Errorf("AGENTS.md reported %s; an answered file is neither edited nor behind", r.Status)
		}
	}

	// **And --force does not take it either.** The flag means "take the newer
	// shipped material", and in a file with a region the shipped material is
	// the region - which is what the scaffolded AGENTS.md promises its reader
	// about the half above the marker.
	if _, err := Write(root, []string{"billing"}, true); err != nil {
		t.Fatal(err)
	}
	forced, err := os.ReadFile(agents)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(forced), answer) {
		t.Error("--force wrote away what was above the marker")
	}

	// And the inspection has to agree with what the write just did, which is
	// the divergence this change removes.
	inspected, err := InspectShipped(root, []string{"billing"})
	if err != nil {
		t.Fatal(err)
	}
	var written, seen Status
	for _, r := range results {
		if r.Path == "AGENTS.md" {
			written = r.Status
		}
	}
	for _, r := range inspected {
		if r.Path == "AGENTS.md" {
			seen = r.Status
		}
	}
	if written != seen {
		t.Errorf("Write reported %s and InspectShipped reported %s for the same file", written, seen)
	}
}
