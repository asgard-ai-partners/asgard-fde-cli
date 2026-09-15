package main

import "testing"

// reported is the decision runIntroduced makes about one added line: the
// pattern matches it, and nothing in notAClaim takes it back. It is here
// rather than in introduced.go because the two regexps are the surface worth
// pinning and neither is worth wrapping in a function to get at.
func reported(line string) bool {
	claim := aDate.ReplaceAllString(line, " ")
	if !countShaped.MatchString(claim) {
		return false
	}
	for _, re := range notAClaim {
		if re.MatchString(claim) {
			return false
		}
	}
	return true
}

// The shape it exists for: a number, or a number written as a word, against a
// set that can move. Every one of these is a sentence somebody has to recount
// when the set changes, which is the whole argument for printing it.
func TestCountShapedMatchesAClaimAboutASet(t *testing.T) {
	for _, line := range []string{
		"the five rules below apply to every part",
		"the wiki is 26 wiki pages and an index",
		"Seven of the twenty-four kinds have a generator",
		"2026-09-11, all 81 flags read against what the flag does",
		"checked against all seven reference deployments",
	} {
		if !countShaped.MatchString(line) {
			t.Errorf("countShaped missed a count-shaped line: %q", line)
			continue
		}
		if !reported(line) {
			t.Errorf("notAClaim swallowed a real claim: %q", line)
		}
	}
}

// What it must not print. A count-shaped line is only worth a reader's time
// when the number stands for a set that can move; a date, a link, a commit and
// a quantity quoted with its unit are none of those. Each row says which of
// the two mechanisms lets the line through, because they are different repairs
// if one stops working.
func TestNotAClaimKeepsOutWhatCannotGoStale(t *testing.T) {
	for _, c := range []struct{ line, why string }{
		{"**Checked:** 2026-09-14 against the CRDs, and every field is present", "the pattern itself: a date is not a count"},
		{"see https://github.com/asgard-ai-platform/asgard-kube for all nine classes", "the URL pattern"},
		{"held against asgard-kube `cbd8d70`, and its six fields are present", "the commit-hash pattern"},
		{"the request gets 30 seconds before the poller gives up", "the unit pattern"},
		{"a rendered page and its two assets must stay under 16MB", "the unit pattern"},
		{"the retry window is 5m and the deadline is 3 minutes", "the unit pattern"},
	} {
		if reported(c.line) {
			t.Errorf("%s did not keep this out: %q", c.why, c.line)
		}
	}
}

// **Pinned as noise, deliberately.** A structural "two things" or "three
// questions" enumerated in the same sentence is not a claim about a set, and
// the pattern reports it anyway. That is the design: the scope is one change's
// added lines, a few dozen at most, and a person reads them. A filter for this
// would have to tell an enumeration from a count, which is a reader's job - and
// a test asserting these do NOT match would be pinning a filter nobody wrote.
func TestCountShapedAlsoMatchesStructuralPhrasingAndThatIsAccepted(t *testing.T) {
	for _, line := range []string{
		"there are two things that will bite you here",
		"three questions the material has to keep answering",
		"one of the four ways this material has contradicted itself",
	} {
		if !reported(line) {
			t.Errorf("a filter now takes this out, which no rule here asked for: %q", line)
		}
	}
}

// **The mask, and why it is not a veto.** A date's own digits are read as a
// count when a plural noun follows them, and that is the whole of what the
// date pattern is for; nothing else on a line stops being a claim because the
// line is dated. Vetoing the line took out the one place the forbidden shape
// lives - a count annotating a source is what AGENTS.md says never to write,
// and every line that annotates a source carries the date it was read at.
func TestADateIsMaskedRatherThanVetoingTheLine(t *testing.T) {
	digits := "the wiki was re-read on 2026-09-14 pages first, then the extracts"
	if !countShaped.MatchString(digits) {
		t.Fatalf("the pattern no longer reads a date's digits as a count, so the mask is now unmotivated: %q", digits)
	}
	if reported(digits) {
		t.Errorf("the date's own digits are being reported as a count: %q", digits)
	}
	dated := "**Checked:** 2026-09-14, against 11 SemanticLayer CRs in three deployments"
	if !reported(dated) {
		t.Errorf("a count annotating a source is invisible because the line carries a date: %q", dated)
	}
}

// This exercises the git invocation and the wiring, and nothing else: the
// check is a listing and cannot fail, so a nil return says only that it did
// not panic and that `git diff -U0` was accepted. It is not coverage of what
// the listing decides.
func TestRunIntroducedAgainstThisRepository(t *testing.T) {
	if out := captureStdout(t, func() {
		if err := runIntroduced(nil); err != nil {
			t.Errorf("runIntroduced: %v", err)
		}
	}); out == "" {
		t.Error("runIntroduced printed nothing")
	}
}
