package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

const issueRepo = "asgard-ai-partners/asgard-fde-cli"

func newIssueCmd() *cobra.Command {
	var draft bool

	cmd := &cobra.Command{
		Use:   "issue-report",
		Short: "How to report a gap in this tool, from wherever you found it",
		Long: `How to file what this tool got wrong, or did not know.

You are probably in a customer repository. **The gap does not belong there** - a
note in one engagement's docs is a note one engagement has, and the next one
starts over. It belongs upstream, where a fix reaches every engagement in one
release. Same reason this material is compiled into the binary rather than
copied into your repo.

FILE ONE WHEN

  - you searched for something, found nothing, and it turned out to exist
  - two pages told you opposite things
  - you did what a page said and it was wrong in front of a customer
  - a number or a claim here did not match what you saw
  - you needed something in a meeting that this tool does not have

Do not wait to be sure it is a defect. "I could not find X and I do not know
whether it exists" is useful: it is either a missing page or a wrong signpost,
and those need different repairs.

WHAT TO WRITE

The reader has none of your context and is very likely a future you, with no
memory of today.

  1. What I was trying to do
     The real task in a sentence - "building a discovery deck for a customer
     whose three scenarios all read internal systems", not "using the skill".

  2. The state I was in            REQUIRED, and it is what makes it a bug
                                   report rather than a complaint
     Somebody has to be able to stand where you stood. What the repo held, and
     what the customer situation was in shape. Do not assemble this by hand -
     --new collects it, and collects it safely. What a reader needs is the
     version, what the charts declare, what "asgard-cli check" says and how much
     is open; what they must never receive is the content of any of it.

     **"asgard-cli question" and "asgard-cli request" are not pasteable.** Their
     rows are the customer's own table names, column names and system names,
     which is exactly what the rule below forbids. --new reports them as counts
     for that reason, and a count carries everything a fix needs.

  3. What I ran, and what came back
     In order, with the real output pasted. Then what you expected instead. The
     gap between those two is usually the whole report.

  4. Where the answer actually was
     The one that gets left out, and the most useful. Say what you searched for
     first: "I searched for the marketplace names and got nothing; it was in
     asgard-freyr-skills the whole time." A missing page and an unfindable page
     need different fixes, and only this sentence tells them apart. If you never
     found it, say that - it is also an answer.

  5. What it cost
     Twenty minutes, or a wrong sentence to a customer, or nothing yet because
     you caught it. This decides what gets fixed first, and "nothing yet, but it
     nearly reached a slide" is a real answer.

NEVER PASTE THE CUSTOMER'S CONTENT

Their document, their system names, their hostnames, their people. Describe the
shape instead: "three capabilities, one public channel and two internal" carries
everything a fix needs and identifies nobody.

DO NOT FILE

  - a fix you already made. That is a pull request and it is better
  - "the documentation should be better". Name the sentence that misled you
  - a report with no state in it. Nobody can act on "find did not work"

--new WRITES THE REPORT

Three of the five sections above are things this tool already knows, and asking
for them by hand is why they arrive missing. --new emits the body with those
filled in and the rest marked TODO:

    asgard-cli issue-report --new > report.md
    asgard-cli issue-report --new | gh issue create --repo asgard-ai-partners/asgard-fde-cli --body-file -

**Section 2 and the search evidence are collected, not narrated.** A report is
otherwise entirely somebody's account of what happened, and the account is the
part that can be wrong - a search someone remembers running, phrased differently
from the one they ran. What --new puts in is the tool's own record: the version,
what the charts declare, what "asgard-cli check" says, how many questions,
requests and task specs are open, and every query that came back empty.

**The search evidence appears only when there is some.** It is read from
docs/.find-misses, which "asgard-cli find" writes when a query returns nothing,
and a repository where every search found something has no such file - so
section 3 arrives as a bare TODO and that is the correct output, not a bug. The
line the report closes with names what was actually collected.

**Read what it produced before filing it.** The misses are queries as they were
typed, so they can carry the customer's words; the rule above about never
pasting their content applies to what this generated exactly as much as to what
you write.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if draft {
				return writeReport(out)
			}
			fmt.Fprintf(out, "File it at:\n\n  https://github.com/%s/issues/new\n\n", issueRepo)
			fmt.Fprintf(out, "Or, with `gh` authenticated:\n\n"+
				"  gh issue create --repo %s\n\n", issueRepo)
			fmt.Fprintf(out, "Paste this line so nobody has to ask:\n\n  asgard-cli %s\n\n",
				version.Get().String())
			fmt.Fprintf(out, "Or have the body written for you, with the state and the search\nevidence already in it:\n\n  asgard-cli issue-report --new\n\n")
			fmt.Fprintf(out, "What to put in it: `asgard-cli issue-report --help`.\n"+
				"A worked example: https://github.com/%s/issues/9\n", issueRepo)
			return nil
		},
	}

	cmd.Flags().BoolVar(&draft, "new", false, "write the report body, with the state and the search evidence filled in")

	return cmd
}

// writeReport emits the issue body.
//
// The sections are the five in this command's help, in that order, because a
// reader of the repository's issues should not have to learn two shapes. What
// differs is who fills each one: the tool fills what it can observe and marks
// the rest TODO, rather than describing all five and hoping.
//
// It works outside a repository. Half this tool's job is answering a question
// asked before there is a directory, and a gap found there is worth the same
// report - it just has no repository state to carry.
func writeReport(out io.Writer) error {
	fmt.Fprintf(out, "## 1) What I was trying to do\n\n"+
		"TODO - the real task in a sentence, not \"using the tool\". Describe the\n"+
		"shape of the customer\u0027s situation, never their systems or their people.\n\n")

	fmt.Fprintf(out, "## 2) The state I was in\n\n```\nasgard-cli %s\n```\n\n", version.Get().String())

	state, err := loadState()
	if err != nil {
		fmt.Fprintf(out, "Run outside a customer repository, so there is no repository state.\n"+
			"That is a normal place to hit a gap: the question gets asked in a meeting.\n\n")
	} else {
		fmt.Fprintf(out, "```\n")
		printProjects(out, state)
		fmt.Fprintf(out, "```\n\n")
		fmt.Fprintf(out, "%d open question(s), %d open request(s), %d open task spec(s).\n\n",
			len(state.Questions), len(work.ActiveRequests(state.Requests)), len(work.ActiveTasks(state.Tasks)))
	}
	checked := writeCheck(out)

	fmt.Fprintf(out, "## 3) What I ran, and what came back\n\n")
	misses := writeMisses(out)
	fmt.Fprintf(out, "TODO - the rest, in order, with the real output pasted, then what you\nexpected instead. The gap between those two is usually the whole report.\n\n")

	fmt.Fprintf(out, "## 4) Where the answer actually was\n\n"+
		"TODO - and this is the one that gets left out. A missing page and an\n"+
		"unfindable page need different fixes, and only this sentence tells them\n"+
		"apart. If you never found it, say that; it is also an answer.\n\n")

	fmt.Fprintf(out, "## 5) What it cost\n\n"+
		"TODO - twenty minutes, a wrong sentence to a customer, or nothing yet\nbecause you caught it. This decides what gets fixed first.\n\n")

	// **What it says it collected has to be what it collected.** The old line
	// claimed the search evidence unconditionally, and a repository where every
	// `find` returned something has no misses file at all - so a reader saw a
	// bare TODO under a sentence promising evidence, and went looking for a bug
	// in the generator. Naming the parts is one line and removes that hunt.
	collected := "the version"
	if err == nil {
		collected += ", what the charts declare, what is open"
	}
	if checked {
		collected += ", the `check` report"
	}
	if misses {
		collected += ", and the searches that came back empty"
	}
	fmt.Fprintf(out, "---\n\nWritten by `asgard-cli issue-report --new`. Collected: %s.\n"+
		"The TODOs are not.\n", collected)
	return nil
}

// writeCheck puts `asgard-cli check` into section 2, and reports whether it did.
//
// The help has always named `check` as one of the things worth pasting, and
// --new has never included it. It is the most reproducible half of "the state I
// was in": every other line of section 2 says what the repository holds, and
// this is the only one that says whether what it holds is coherent.
//
// A failing check is the interesting case and must not stop the report - a
// repository broken enough to fail it is a repository somebody is more likely
// to be filing about, not less.
func writeCheck(out io.Writer) bool {
	root := repo.Root(".")
	if root == "" {
		return false
	}
	report, err := check.Run(root)
	if err != nil {
		return false
	}
	fmt.Fprintf(out, "`asgard-cli check` at that moment:\n\n```\n")
	for _, f := range report.Warnings() {
		fmt.Fprintf(out, "warn   %s\n", f.Message)
	}
	for _, f := range report.Errors() {
		fmt.Fprintf(out, "error  %s\n", f.Message)
	}
	if report.OK() && len(report.Warnings()) == 0 {
		fmt.Fprintf(out, "ok  structure is consistent (whole repo)\n")
	}
	fmt.Fprintf(out, "```\n\n")
	return true
}

// writeMisses puts the recorded empty searches into the report.
//
// This is the only part of a defect report that is not somebody's account of
// what happened. A search that came back empty was witnessed by the tool, so it
// is evidence rather than recollection - and it is the half of section 4 that
// separates a missing page from an unfindable one.
func writeMisses(out io.Writer) bool {
	root := repo.Root(".")
	if root == "" {
		return false
	}
	misses, err := work.Misses(root)
	if err != nil || len(misses) == 0 {
		return false
	}
	fmt.Fprintf(out, "Searches this engagement ran that the material did not answer, recorded by\n"+
		"`asgard-cli find` at the time:\n\n```\n")
	for _, m := range misses {
		fmt.Fprintf(out, "%s  %-9s %s\n", m.Date, m.Kind, m.Query)
	}
	fmt.Fprintf(out, "```\n\n**Check these before filing** - a query carries whatever words were typed.\n\n")
	return true
}
