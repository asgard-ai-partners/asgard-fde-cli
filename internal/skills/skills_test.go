package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// Which of the two candidate directories the material is read from and written
// to.
//
// **The wrong answer here is silent and lasting.** Nothing fails: the material
// is fetched into a directory, an agent's runtime reads whichever directory it
// looks in, and the copy that is no longer in use sits there describing
// whatever server it described when it was written. Root picked the first
// directory that EXISTED for one release, which meant `asgard-cli init` -
// which creates `.agents/skills` unconditionally - moved the answer out from
// under a repository that had already fetched into `.claude/skills`.
func TestRootPrefersWhereTheMaterialAlreadyIs(t *testing.T) {
	agents := filepath.Join(".agents", "skills")
	claude := filepath.Join(".claude", "skills")

	for _, c := range []struct {
		name string
		// dirs exist; stamped also hold the record.
		dirs     []string
		stamped  []string
		override string
		want     string
	}{
		// Nothing here at all: the first candidate is what gets created.
		{name: "empty repository", want: agents},

		// The ordinary scaffolded repository. `init` wrote .agents/skills and
		// nothing has been fetched yet.
		{name: "scaffolded, nothing fetched", dirs: []string{agents}, want: agents},

		// A checkout already using Claude Code, before `init` has run.
		{name: "only claude exists", dirs: []string{claude}, want: claude},

		// The case the existence test got wrong. The material was fetched into
		// .claude/skills, then somebody ran `init` and .agents/skills appeared
		// beside it. The record is in .claude/skills, so that is still where
		// the material is.
		{
			name:    "fetched into claude, then scaffolded",
			dirs:    []string{agents, claude},
			stamped: []string{claude},
			want:    claude,
		},

		// The same repository once an update has been pointed at .agents.
		{
			name:    "fetched into agents, claude merely exists",
			dirs:    []string{agents, claude},
			stamped: []string{agents},
			want:    agents,
		},

		// Both hold a record - a repository that ran the two versions of this
		// function. Roots order decides, and it is the one an agent-neutral
		// layout puts first.
		{
			name:    "both hold a record",
			dirs:    []string{agents, claude},
			stamped: []string{agents, claude},
			want:    agents,
		},

		// An override is somebody saying it outright, and beats every rule
		// above including a record somewhere else.
		{
			name:     "override wins over a record elsewhere",
			dirs:     []string{agents, claude},
			stamped:  []string{claude},
			override: filepath.Join("docs", "skills"),
			want:     filepath.Join("docs", "skills"),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			for _, d := range c.dirs {
				if err := os.MkdirAll(filepath.Join(repoRoot, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for _, d := range c.stamped {
				if err := os.WriteFile(filepath.Join(repoRoot, d, StampName), []byte(`{"version":"9"}`), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got, want := Root(repoRoot, c.override), filepath.Join(repoRoot, c.want); got != want {
				t.Errorf("Root() = %s, want %s", got, want)
			}
		})
	}
}

// Elsewhere is what reports the copy that is not in use, and the thing it must
// not do is report a directory holding no material: an empty `.claude/skills`
// is a repository that uses Claude Code, not a second copy of anything.
func TestElsewhereReportsOnlyRealCopies(t *testing.T) {
	agents := filepath.Join(".agents", "skills")
	claude := filepath.Join(".claude", "skills")

	repoRoot := t.TempDir()
	for _, d := range []string{agents, claude} {
		if err := os.MkdirAll(filepath.Join(repoRoot, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	inUse := Root(repoRoot, "")
	if got := Elsewhere(repoRoot, inUse); len(got) != 0 {
		t.Fatalf("an empty second directory is not a copy: got %v", got)
	}

	body := []byte(`{"version":"11","files":{"a.md":"x","b.md":"y"}}`)
	if err := os.WriteFile(filepath.Join(repoRoot, claude, StampName), body, 0o644); err != nil {
		t.Fatal(err)
	}
	// Root now resolves to .claude/skills, because that is where the material
	// is - so nothing is elsewhere.
	if got := Elsewhere(repoRoot, Root(repoRoot, "")); len(got) != 0 {
		t.Fatalf("the directory in use is never elsewhere: got %v", got)
	}
	// Pointed at the other one by hand, the fetched copy is the one left
	// behind, and it is reported with what it holds.
	got := Elsewhere(repoRoot, filepath.Join(repoRoot, agents))
	if len(got) != 1 {
		t.Fatalf("Elsewhere() = %v, want one entry", got)
	}
	if got[0].Dir != claude || got[0].Version != "11" || got[0].Files != 2 {
		t.Errorf("Elsewhere() = %+v, want {%s 11 2}", got[0], claude)
	}
}
