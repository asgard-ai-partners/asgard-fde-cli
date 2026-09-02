package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
)

const issueRepo = "asgard-ai-partners/asgard-fde-cli"

func newIssueCmd() *cobra.Command {
	return &cobra.Command{
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
     Somebody has to be able to stand where you stood. Which stage, what the
     repo held, what the customer situation was in shape. The fastest way to
     give most of it is to paste "asgard-cli next" and "asgard-cli check" from
     that moment - they describe the state better than a sentence, and they are
     what a reader runs to reproduce it.

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
  - a report with no state in it. Nobody can act on "find did not work"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "File it at:\n\n  https://github.com/%s/issues/new\n\n", issueRepo)
			fmt.Fprintf(out, "Or, with `gh` authenticated:\n\n"+
				"  gh issue create --repo %s\n\n", issueRepo)
			fmt.Fprintf(out, "Paste this line so nobody has to ask:\n\n  asgard-cli %s\n\n",
				version.Get().String())
			fmt.Fprintf(out, "What to put in it: `asgard-cli issue-report --help`.\n"+
				"A worked example: https://github.com/%s/issues/9\n", issueRepo)
			return nil
		},
	}
}
