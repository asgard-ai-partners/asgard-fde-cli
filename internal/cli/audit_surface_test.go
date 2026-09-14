package cli

import (
	"strings"
	"testing"
)

// **A sweep is only as wide as the set it is given, and the set drifted.**
//
// `--commands`, `--sources` and `--term` each assembled their own source set by
// hand and the three were not the same. The narrowest was `--term`'s - the one
// AGENTS.md sends a reader to for "what else claimed the thing you just
// changed" - and it read neither this repository's own documents nor its
// maintenance skills. A sweep run exactly as instructed came back clean over
// surfaces it had never opened, which reads identically to one that found
// nothing.
//
// `everySurface` is the one assembly now, and this holds it to the parts that
// were missing.
//
// **Asserted by file name, not by content.** `goStrings` reads this package's
// own Go source, test files included, so a probe written as a string literal
// here is satisfied by this file - the first version of this test passed
// against a set with two collectors deleted, for exactly that reason.
func TestEverySurfaceReachesWhatASweepMustSee(t *testing.T) {
	all, err := everySurface(NewRootCmd())
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]bool{}
	labels := map[string]int{}
	for _, s := range all {
		names[s.name] = true
		labels[s.label]++
	}

	for _, want := range []string{"wiki", "usecase", "template", "scaffold", "source", "repo", "hack"} {
		if labels[want] == 0 {
			t.Errorf("everySurface carries nothing labelled %q, so a sweep is blind to it", want)
		}
	}

	// The ones that were missing, each named by the collector that supplies
	// it: repoDocs, repoSkills, goStrings over this package, and repoHack -
	// the maintainer's gate, whose `What:` strings `go run ./hack list` prints
	// and whose descriptions outlive the behaviour they describe otherwise.
	for _, want := range []string{
		"AGENTS.md",
		".agents/skills/consistency-checks/SKILL.md",
		"internal/cli/gate.go",
		"hack/passlist.go",
	} {
		if !names[want] {
			t.Errorf("everySurface does not carry %s; a term sweep reports it clean without reading it", want)
		}
	}

	// A help screen's own prose reaches the sweep through `goStrings`, because
	// that is where it is written. This is the assertion that the path works
	// end to end rather than that the file is merely present.
	helm := false
	for _, s := range all {
		if s.name == "internal/cli/gate.go" && strings.Contains(s.body, "Do not run helm by hand here") {
			helm = true
		}
	}
	if !helm {
		t.Error("a help screen's prose is not in the swept body of the file that writes it")
	}
}
