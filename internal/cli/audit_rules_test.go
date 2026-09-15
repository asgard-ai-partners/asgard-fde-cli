package cli

import (
	"strings"
	"testing"
)

// The `audit-material` rules had no test at all, and the newest of them -
// `halfPath`, which catches a pointer whose `.md` is missing - is the one most
// likely to be wrong: it is a pattern over prose, and the corpus is full of the
// string it has to stay silent on. AGENTS.md records three checks deleted
// rather than tuned because they fired on correct material, so what these hold
// is both halves - what each rule catches, and what it must never touch.
//
// **The sources are written here, not read from the corpus.** A test that
// asserts against 70 documents fails for reasons that are not its subject. The
// one thing taken from the corpus is the set of names a pointer may resolve to:
// `checkBare` resolves through `targets()`, so a fixture has to use names that
// are really there, and `anchors` is the guard on that.

// bareSource is one in-memory document for the audits to read.
func bareSource(body string) []source {
	return []source{{label: "wiki", name: "fixture", body: body}}
}

// findings runs a check over one body and returns the lines it reported.
//
// The count is taken from the reported lines rather than from the trailing
// tally, so a check that stops printing what it found cannot pass by printing
// a number.
func findings(t *testing.T, prefix string, check func(*strings.Builder, []source) error, body string) ([]string, error) {
	t.Helper()
	var out strings.Builder
	err := check(&out, bareSource(body))
	var lines []string
	for _, l := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(l, prefix) {
			lines = append(lines, l)
		}
	}
	return lines, err
}

func bareFindings(t *testing.T, body string) ([]string, error) {
	t.Helper()
	return findings(t, "bare ", func(w *strings.Builder, s []source) error { return checkBare(w, s) }, body)
}

func pathFindings(t *testing.T, body string) ([]string, error) {
	t.Helper()
	return findings(t, "unrooted ", func(w *strings.Builder, s []source) error { return checkPaths(w, s) }, body)
}

// TestTheAnchorDocumentsExist is the one assertion here that reads the real
// corpus, and it is deliberate: every fixture below names `usecase
// write-path`, `wiki glossary`, `usecase flow-agent-single` or `guide
// projects`, and a rule that fires only on a name `targets()` knows would go
// silent - passing every "must not fire" test for the wrong reason - the day
// one of them was renamed.
func TestTheAnchorDocumentsExist(t *testing.T) {
	known, err := targets()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []struct{ kind, name string }{
		{"usecase", "write-path"},
		{"usecase", "flow-agent-single"},
		{"wiki", "glossary"},
		{"guide", "projects"},
	} {
		if !known[a.kind][a.name] {
			t.Fatalf("the fixtures below point at %s/%s and it is gone; rename it in them, or every "+
				"silence they assert is a rule that stopped resolving rather than a rule that held", a.kind, a.name)
		}
	}
}

// TestHalfPathCatchesAPointerWithoutItsExtension is the shape the rule was
// written for: the path is there, the `.md` is not, so `kb.Link` never reads it
// as a pointer and `--links` never resolves it. It reads perfectly and goes
// nowhere.
func TestHalfPathCatchesAPointerWithoutItsExtension(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"inline", "The write path is assembled in `../usecase/write-path` field by field."},
		{"split across a line break", "The write path is assembled in `../usecase/\nwrite-path` field by field."},
		{"wrapped with indentation", "See\n    ../usecase/\n    write-path\nfor the shape."},
		{"another kind", "The terms are on ../wiki/glossary and nowhere else."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := bareFindings(t, tc.body)
			if err == nil {
				t.Error("a pointer with no `.md` was reported as clean; `--links` cannot see this shape either, so nothing does")
			}
			if len(lines) != 1 {
				t.Fatalf("want 1 finding, got %d: %q", len(lines), lines)
			}
		})
	}
}

// TestHalfPathNamesBothPaths holds the message itself. A reader who is told
// only that a line is wrong has to work out what was meant; the repair is one
// character, and the report is where it should be visible.
func TestHalfPathNamesBothPaths(t *testing.T) {
	lines, err := bareFindings(t, "assembled in `../usecase/write-path`, field by field")
	if err == nil {
		t.Fatal("no finding at all")
	}
	if len(lines) != 1 {
		t.Fatalf("want 1 finding, got %d: %q", len(lines), lines)
	}
	if !strings.Contains(lines[0], "`../usecase/write-path`") {
		t.Errorf("the report does not name the path that is there: %q", lines[0])
	}
	if !strings.Contains(lines[0], "`../usecase/write-path.md`") {
		t.Errorf("the report does not name the path that was meant: %q", lines[0])
	}
}

// TestHalfPathIsSilentOnTheCorrectForm is the test that matters most.
//
// `../usecase/write-path.md` is the commonest string in this corpus and it is
// the right one. A rule that fires on it would report every correct pointer in
// the material, which is the failure AGENTS.md has three deleted checks for -
// and the answer to a check that cries wolf is that somebody turns it off.
func TestHalfPathIsSilentOnTheCorrectForm(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"plain", "assembled in `../usecase/write-path.md`, field by field"},
		{"two on one line", "see `../usecase/write-path.md` and `../wiki/glossary.md`"},
		{"markdown link", "see [the write path](../usecase/write-path.md)"},
		{"wrapped after the slash", "see `../usecase/\nwrite-path.md` for the shape"},
		{"in a sentence with no backticks", "The shape is ../usecase/write-path.md and the terms are ../wiki/glossary.md."},
		{"a trailing sentence stop", "The shape is in ../usecase/write-path.md."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := bareFindings(t, tc.body)
			if err != nil || len(lines) != 0 {
				t.Errorf("a correct pointer was reported: %q (%v)", lines, err)
			}
		})
	}
}

// TestHalfPathIsSilentOnAPathThatNamesNoDocument keeps the rule to what it can
// actually judge. `../usecase/not-a-real-page` is either a typo or prose about
// a document that does not exist, and this check has no standing to say which:
// resolving names is `--links`'s job on the form it can see.
func TestHalfPathIsSilentOnAPathThatNamesNoDocument(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"no such extract", "see `../usecase/not-a-real-page` for the shape"},
		{"no such page", "see `../wiki/not-a-real-page` for the shape"},
		{"a directory, named as one", "the extracts live under `../usecase/`"},
		{"a glob in a command", "grep -n allowWrite ../usecase/*.md"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := bareFindings(t, tc.body)
			if err != nil || len(lines) != 0 {
				t.Errorf("reported a path naming no document: %q (%v)", lines, err)
			}
		})
	}
}

// TestHalfPathAcrossAWhitespaceRun is where the newline in the character class
// costs something, and it is the boundary of the rule rather than a bug in it.
//
// `[\s]*` was widened to cross a line break so that a pointer wrapped mid-path
// is still caught - the shape a reader is least likely to spot. The price is
// that a directory named at the end of a line and a document name opening the
// next one read to this rule as one pointer. The corpus writes a bare
// directory in backticks, which ends the match at the backtick, so the shape
// below does not occur in it - but it is available, and narrowing the class
// would give up the case the rule was written for.
func TestHalfPathAcrossAWhitespaceRun(t *testing.T) {
	lines, _ := bareFindings(t, "The extracts live under ../usecase/\nwrite-path is one of them.")
	if len(lines) != 1 {
		t.Fatalf("want 1 finding, got %d: %q", len(lines), lines)
	}
	// The same words with the directory in backticks, which is how this
	// material writes one, and the match stops at the backtick.
	lines, err := bareFindings(t, "The extracts live under `../usecase/`\n`write-path` is one of them.")
	for _, l := range lines {
		if strings.Contains(l, "the path is there") {
			t.Errorf("a backticked directory was read as half a pointer: %q (%v)", l, err)
		}
	}
}

// TestHalfPathReadsFencesAndURLs records that the rule has no notion of either.
//
// **Neither is an exception this adds.** A fenced block in this material is
// mostly a skeleton somebody copies by hand, so a half pointer in one is copied
// too, and a `../` path inside a URL is not a shape the corpus writes. What the
// test pins is that the behaviour is known rather than assumed - if this ever
// has to change, the narrower pattern is the fix, not a fence-stripping pass
// over every body.
func TestHalfPathReadsFencesAndURLs(t *testing.T) {
	fenced := "```\n# see ../usecase/write-path for the shape\n```"
	lines, _ := bareFindings(t, fenced)
	if len(lines) != 1 {
		t.Errorf("fenced half pointer: want 1 finding, got %d: %q", len(lines), lines)
	}
	// A URL cannot carry `../` without being a broken URL, so there is nothing
	// to be silent about; this is the assertion that it stays that way.
	lines, err := bareFindings(t, "https://github.com/asgard-ai-platform/asgard-kube/tree/main/usecase/write-path")
	if err != nil || len(lines) != 0 {
		t.Errorf("a URL path was read as a half pointer: %q (%v)", lines, err)
	}
}

// TestBareCatchesANameWithNoPathAtAll is the older rule: a document named with
// its kind and no path - the form `kb.Link` deliberately stopped reading, so a
// writer who writes it writes a pointer nothing can follow and nothing reports.
func TestBareCatchesANameWithNoPathAtAll(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"a guide", "Read `guide projects` before the interview."},
		{"a page", "The terms are on `wiki glossary`."},
		{"an extract", "The shape is in `usecase write-path`."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := bareFindings(t, tc.body)
			if err == nil {
				t.Error("a document named with no path was reported as clean")
			}
			if len(lines) != 1 {
				t.Fatalf("want 1 finding, got %d: %q", len(lines), lines)
			}
		})
	}
}

// TestBareIsSilentOnAPhraseThatMerelyStartsWithAKind. `wiki` and `usecase` are
// ordinary words in this material, and a backticked phrase opening with one is
// prose. The rule earns its silence from the second half: the name has to
// resolve.
func TestBareIsSilentOnAPhraseThatMerelyStartsWithAKind(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"a phrase", "Every `wiki page` opens with a title."},
		{"a sentence", "`usecase extracts are read after the page` is the reading order."},
		{"an unresolvable name", "Read `guide nowhere-at-all` first."},
		{"the kind alone", "The `wiki` is the corpus."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := bareFindings(t, tc.body)
			if err != nil || len(lines) != 0 {
				t.Errorf("reported prose as a bare name: %q (%v)", lines, err)
			}
		})
	}
}

// TestNameOnlyCatchesADocumentNameOnItsOwn - no kind, no path, just the name
// in backticks. Not a path, so `--links` skips it; no kind, so `bare` skips it.
func TestNameOnlyCatchesADocumentNameOnItsOwn(t *testing.T) {
	lines, err := bareFindings(t, "The single-agent shape is `flow-agent-single`.")
	if err == nil {
		t.Error("a bare document name was reported as clean")
	}
	if len(lines) != 1 {
		t.Fatalf("want 1 finding, got %d: %q", len(lines), lines)
	}
	if !strings.Contains(lines[0], "`../usecase/flow-agent-single.md`") {
		t.Errorf("the report does not say what to write instead: %q", lines[0])
	}
}

// TestNameOnlyIsSilentWhereAHyphenIsNotAPointer holds the four exemptions the
// rule depends on to be usable at all. Each of them is correct prose that a
// looser pattern would rewrite.
func TestNameOnlyIsSilentWhereAHyphenIsNotAPointer(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"a single word that is also a field", "`agents` is a stringified JSON array."},
		{"a hyphenated token that is no document", "`bot-provider-class` is immutable once applied."},
		{"a markdown link label", "See [`write-path`](../usecase/write-path.md) for the shape."},
		{"a design-time skill", "Load the `knowledge-base` skill before writing one."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := bareFindings(t, tc.body)
			if err != nil || len(lines) != 0 {
				t.Errorf("reported correct prose as a bare name: %q (%v)", lines, err)
			}
		})
	}
}

// TestNameOnlyIsSilentOnADocumentNamingItself. A page writing its own name in
// its own prose is not pointing anywhere.
func TestNameOnlyIsSilentOnADocumentNamingItself(t *testing.T) {
	var out strings.Builder
	err := checkBare(&out, []source{{
		label: "usecase", name: "write-path",
		body: "`write-path` is about the shape where the agent writes back.",
	}})
	if err != nil {
		t.Errorf("a document naming itself was reported: %v", err)
	}
}

// TestPathsReportsAFileOnlyThisRepositoryHas. These documents land in a
// customer's checkout, where `internal/` is not there and the path resolves to
// nothing - the same failure as a dead pointer, one directory up.
func TestPathsReportsAFileOnlyThisRepositoryHas(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"a package file", "The pinned copy is in `internal/gate/processors.go`."},
		{"the maintainer's gate", "Run the check in `hack/passlist.go`."},
		{"a root document", "The goal is stated in `Goal.md`."},
		{"an earlier layout", "The extracts used to live in `extracts/`."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := pathFindings(t, tc.body)
			if err == nil {
				t.Error("a path only this repository has was reported as clean")
			}
			if len(lines) != 1 {
				t.Fatalf("want 1 finding, got %d: %q", len(lines), lines)
			}
		})
	}
}

// TestPathsIsSilentOnACustomerRepositoryPath. `projects/`, `docs/` and
// `assets/` are what a scaffolded repository is made of, so a landed document
// naming one is pointing at a file the reader has.
func TestPathsIsSilentOnACustomerRepositoryPath(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"a chart", "The chart is at `projects/app/Chart.yaml`."},
		{"the docs directory", "Write the handover into `docs/handover.md`."},
		{"the skills directory", "A SkillSet syncing from git leaves `assets/skills/` empty."},
		{"the scaffolded corpus", "The pages land under `.agents/skills/asgard-platform/wiki/`."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := pathFindings(t, tc.body)
			if err != nil || len(lines) != 0 {
				t.Errorf("reported a path the reader has: %q (%v)", lines, err)
			}
		})
	}
}

// TestPathsIsSilentWhenTheRepositoryIsNamedOnTheLine. Citing a file in another
// repository is provenance and is the point; what makes it resolvable is the
// repository name beside it.
func TestPathsIsSilentWhenTheRepositoryIsNamedOnTheLine(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"an upstream", "The definitions are in asgard-core `internal/processor/definitions.go`."},
		{"this tool, named", "asgard-fde-cli carries the pinned copy in `internal/gate/processors.go`."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := pathFindings(t, tc.body)
			if err != nil || len(lines) != 0 {
				t.Errorf("reported a cited path that names its repository: %q (%v)", lines, err)
			}
		})
	}
	// And the same line with the repository taken off it is reported, so the
	// silence above is the naming and not the shape of the path.
	lines, err := pathFindings(t, "The definitions are in `internal/processor/definitions.go`.")
	if err == nil || len(lines) != 1 {
		t.Errorf("the same path with no repository named was not reported: %q (%v)", lines, err)
	}
}
