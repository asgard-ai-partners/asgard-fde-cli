package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
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

You are probably in a customer repository. File the gap upstream, not there: a
note in one engagement's docs reaches only that engagement, and a fix upstream
reaches every engagement in one release. This material is compiled into the
binary rather than copied into your repo for the same reason.

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

  2. The state I was in            REQUIRED: without it nobody can reproduce
                                   what you saw
     Somebody has to be able to stand where you stood. What the repo held, and
     what the customer situation was in shape. Do not assemble this by hand -
     --new collects it, and collects it safely. What a reader needs is the
     version, what the charts declare, what "asgard-cli check" says and how much
     is open; what they must never receive is the content of any of it.

     Do not paste "asgard-cli question" or "asgard-cli request" output. Their
     rows are the customer's own table names, column names and system names,
     which the rule below forbids. --new reports them as counts
     for that reason, and a count carries everything a fix needs.

  3. What I ran, and what came back
     In order, with the real output pasted. Then what you expected instead.

  4. Where the answer actually was
     This section is often left out and is the most useful. Say what you searched for
     first: "I searched for the marketplace names and got nothing; it was in
     asgard-freyr-skills the whole time." A missing page and an unfindable page
     need different fixes, and only this sentence tells them apart. If you never
     found it, say that.

  5. What it cost
     Twenty minutes, or a wrong sentence to a customer, or nothing yet because
     you caught it. This decides what gets fixed first, and "nothing yet, but it
     nearly reached a slide" is a real answer.

  6. What I now know
     For a discovery rather than a defect: something the platform does that
     the material does not say, which you found out on a deployment. Write the
     claim and the shape it was seen on, never the customer. For a discovery,
     sections 3 to 5 are optional.

NEVER PASTE THE CUSTOMER'S CONTENT

Their document, their system names, their hostnames, their people. Describe the
shape instead: "three capabilities, one public channel and two internal" carries
everything a fix needs and identifies nobody.

DO NOT FILE

  - a fix you already made. Open a pull request instead
  - "the documentation should be better". Name the sentence that misled you
  - a report with no state in it. Nobody can act on "find did not work"

--new WRITES THE REPORT

--new emits the report body with section 2, the state you were in, filled in
from what this tool can observe, and the other sections marked TODO:

    asgard-cli issue-report --new > report.md
    asgard-cli issue-report --new | gh issue create --repo asgard-ai-partners/asgard-fde-cli --body-file -

Section 2 comes from the tool's own record rather than your account: the version, what the
charts declare, what "asgard-cli check" says, how many questions, requests
and task specs are open, and the paths of shipped files this repository has
edited in place - paths only, never their contents. The line the report closes with names what was
actually collected.

Read what it produced before filing it. The rule above about never pasting a
customer's content applies to generated text as well as to what you write.`,
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
			fmt.Fprintf(out, "Or have the body written for you, with this repository's own\nstate already in it:\n\n  asgard-cli issue-report --new\n\n")
			fmt.Fprintf(out, "What to put in it: `asgard-cli issue-report --help`.\n"+
				"A worked example: https://github.com/%s/issues/9\n", issueRepo)
			return nil
		},
	}

	cmd.Flags().BoolVar(&draft, "new", false, "write the report body, with this repository's own state filled in")

	return cmd
}

// writeReport emits the issue body.
//
// The sections are the ones in this command's help, in that order. The tool
// fills what it can observe and marks the rest TODO.
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
	edited := writeEdited(out)

	todo := "TODO"
	fmt.Fprintf(out, "## 3) What I ran, and what came back\n\n")
	fmt.Fprintf(out, "%s - the rest, in order, with the real output pasted, then what you\nexpected instead.\n\n", todo)

	fmt.Fprintf(out, "## 4) Where the answer actually was\n\n"+
		"%s - what you searched for first, and where the answer was. A missing\n"+
		"page and an unfindable page need different fixes. If you never found it,\n"+
		"say that.\n\n", todo)

	fmt.Fprintf(out, "## 5) What it cost\n\n"+
		"%s - twenty minutes, a wrong sentence to a customer, or nothing yet\nbecause you caught it.\n\n", todo)

	writeLearned(out)

	// **What it says it collected has to be what it collected.** A line
	// claiming evidence the run did not gather leaves a reader looking at a
	// bare TODO under a promise, hunting for a bug in the generator. Naming
	// the parts is one line and removes that hunt.
	collected := "the version"
	if err == nil {
		collected += ", what the charts declare, what is open"
	}
	if checked {
		collected += ", the `check` report"
	}
	if edited {
		collected += ", the edited shipped files"
	}
	fmt.Fprintf(out, "---\n\nWritten by `asgard-cli issue-report --new`. Collected: %s.\n"+
		"The TODOs are not.\n", collected)
	return nil
}

// writeLearned emits section 6, for a discovery rather than a defect. It is
// always present, because a defect report sometimes carries one too.
func writeLearned(out io.Writer) {
	fmt.Fprintf(out, "## 6) What I now know\n\n"+
		"TODO, or delete this section if there is nothing. The claim, and the\n"+
		"shape it was seen on (never the customer).\n\n")
}

// writeEdited lists the shipped files this repository changed in place, by
// path, and reports whether it listed any. An edit to shipped material is an
// engagement disagreeing with it in writing, which is worth a maintainer's
// look. Paths only: the contents may carry the customer's own names.
func writeEdited(out io.Writer) bool {
	root := repo.Root(".")
	if root == "" {
		return false
	}
	projects, err := repo.Projects(root)
	if err != nil {
		return false
	}
	results, err := scaffold.InspectShipped(root, projects)
	if err != nil {
		return false
	}
	var paths []string
	for _, r := range results {
		if r.Status == scaffold.Edited {
			paths = append(paths, r.Path)
		}
	}
	if len(paths) == 0 {
		return false
	}
	fmt.Fprintf(out, "Shipped files edited in place in this repository:\n\n```\n")
	for _, p := range paths {
		fmt.Fprintf(out, "%s\n", p)
	}
	fmt.Fprintf(out, "```\n\n")
	return true
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
