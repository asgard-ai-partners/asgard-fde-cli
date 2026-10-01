package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/check"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/feedback"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/version"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newIssueCmd() *cobra.Command {
	var (
		draft       bool
		send        string
		email, name string
	)

	cmd := &cobra.Command{
		Use:   "issue-report",
		Short: "How to report a gap in this tool, from wherever you found it",
		Long: `How to file what this tool got wrong, or did not know.

You are probably in a customer repository. **The gap does not belong there** - a
note in one engagement's docs is a note one engagement has, and the next one
starts over. It belongs upstream, where a fix reaches every engagement in one
release. Same reason this material is compiled into the binary rather than
copied into your repo.

**It goes to the maintainers alone.** --send delivers it as a User Feedback in
the maintainers' Sentry project, which nobody outside them reads - not as an
issue on this tool's repository, which is public.

**This is for the tool, never for the customer.** A report leaves the
engagement for a third party's service. What is wrong in the customer's
systems or in what they run goes on the Workbench (asgard-cli workbench
create), and what the engagement records for itself stays in this repository -
the table below is the whole rule.

` + trackersHelp + `

FILE ONE WHEN

  - you searched for something, found nothing, and it turned out to exist
  - two pages told you opposite things
  - you did what a page said and it was wrong in front of a customer
  - a number or a claim here did not match what you saw
  - you needed something in a meeting that this tool does not have
  - a command failed, or its message blamed the wrong thing

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

--new WRITES THE REPORT, --send FILES IT

**One section is something this tool already knows, and it is the required
one.** Section 2 is the state you were in, asking for it by hand is why it
arrives missing, and it is what separates a bug report from a complaint. --new
emits the body with that section collected and the rest marked TODO:

    asgard-cli issue-report --new > report.md
    (fill in every TODO)
    asgard-cli issue-report --send report.md --email you@example.com

**Section 2 is collected, not narrated.** A report is otherwise entirely
somebody's account of what happened, and the account is the part that can be
wrong. What --new puts in is the tool's own record: the version, what the
charts declare, what "asgard-cli check" says, and how many questions, requests
and task specs are open. The line the report closes with names what was
actually collected.

**--send refuses a report that still has a TODO in it**, and names the
sections. A section you cannot answer is answered by saying so - "I never
found it" is an answer to section 4 - not by leaving the marker. It reads the
file, or stdin when the file is "-", and prints the id Sentry filed it under;
quote that id when you follow it up. Without --email nobody can answer you,
and with no network it fails and leaves the file where it was, to send later.

**Read what it produced before sending it.** The rule above about never pasting
a customer's content applies to what this generated exactly as much as to what
you write.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if draft && send != "" {
				return fmt.Errorf("--new writes a report and --send files one: write it, fill it in, then send it")
			}
			if draft {
				return writeReport(out)
			}
			if send != "" {
				return sendReport(cmd, send, email, name)
			}
			fmt.Fprintf(out, "Write the report, with this repository's own state already in it:\n\n"+
				"  asgard-cli issue-report --new > report.md\n\n")
			fmt.Fprintf(out, "Fill in every TODO, then send it to the maintainers:\n\n"+
				"  asgard-cli issue-report --send report.md --email you@example.com\n\n")
			fmt.Fprintf(out, "It arrives as a Sentry User Feedback that only the maintainers read.\n"+
				"What to put in it: `asgard-cli issue-report --help`.\n")
			return nil
		},
	}

	cmd.Flags().BoolVar(&draft, "new", false, "write the report body, with this repository's own state filled in")
	cmd.Flags().StringVar(&send, "send", "", `send this report file to the maintainers ("-" reads stdin)`)
	cmd.Flags().StringVar(&email, "email", "", "with --send: where the maintainers can answer you (default: nowhere)")
	cmd.Flags().StringVar(&name, "name", "", "with --send: your name (default: none)")

	return cmd
}

// todoSection matches the header of a numbered report section.
var todoSection = regexp.MustCompile(`(?m)^## (\d)\)`)

// sendReport files the report at path as a Sentry User Feedback.
//
// It refuses one still carrying --new's TODO markers: a report sent with them
// is the complaint the help says not to file, and the sender is the only
// person who can fill them in.
func sendReport(cmd *cobra.Command, path, email, name string) error {
	var (
		body []byte
		err  error
	)
	if path == "-" {
		body, err = io.ReadAll(cmd.InOrStdin())
	} else {
		body, err = os.ReadFile(path)
	}
	if err != nil {
		return fmt.Errorf("read report: %w", err)
	}
	report := strings.TrimSpace(string(body))
	if report == "" {
		return fmt.Errorf("%s is empty: write it with asgard-cli issue-report --new", path)
	}
	if open := unfilled(report); len(open) > 0 {
		return fmt.Errorf("%s still has a TODO in section %s: answer each one, or say why you cannot, before sending",
			path, strings.Join(open, ", "))
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "sending to the asgard-cli maintainers' Sentry User Feedback\n")
	client := &http.Client{Timeout: 30 * time.Second}
	id, err := feedback.Send(cmd.Context(), client, feedback.Report{
		Body:    report + "\n",
		Email:   email,
		Name:    name,
		Release: "asgard-cli " + version.Get().String(),
	})
	if err != nil {
		return fmt.Errorf("%w (the report is still at %s; send it again later)", err, path)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Sent. Sentry filed it as %s - quote that when you follow it up.\n", id)
	if email == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "No --email was given, so the maintainers have no way to answer you.\n")
	}
	return nil
}

// unfilled names the sections that still hold a "TODO - " marker.
func unfilled(report string) []string {
	var open []string
	idx := todoSection.FindAllStringSubmatchIndex(report, -1)
	for i, m := range idx {
		end := len(report)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		if strings.Contains(report[m[1]:end], "TODO - ") {
			open = append(open, report[m[2]:m[3]])
		}
	}
	if len(idx) == 0 && strings.Contains(report, "TODO - ") {
		open = append(open, "(no section headers)")
	}
	return open
}

// writeReport emits the report body.
//
// The sections are the ones in this command's help, in that order, because a
// reader of the repository's issues should not have to learn two shapes. What
// differs is who fills each one: the tool fills what it can observe and marks
// the rest TODO, rather than describing them all and hoping.
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
	fmt.Fprintf(out, "TODO - the rest, in order, with the real output pasted, then what you\nexpected instead. The gap between those two is usually the whole report.\n\n")

	fmt.Fprintf(out, "## 4) Where the answer actually was\n\n"+
		"TODO - and this is the one that gets left out. A missing page and an\n"+
		"unfindable page need different fixes, and only this sentence tells them\n"+
		"apart. If you never found it, say that; it is also an answer.\n\n")

	fmt.Fprintf(out, "## 5) What it cost\n\n"+
		"TODO - twenty minutes, a wrong sentence to a customer, or nothing yet\nbecause you caught it. This decides what gets fixed first.\n\n")

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
