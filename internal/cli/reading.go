package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/brief"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/needs"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newReadingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reading",
		Short: "Which pages this engagement opened, and which it never did",
		Long: `What this engagement has read of the material, and what it has not.

**The pages that cost the most are not the wrong ones. They are the right ones
nobody opened.** In one engagement ` + "`wiki operations`" + ` sat in the index under the
title Connectivity while an FDE spent a day on connectivity and never followed
it - because nothing pointed there at the moment it was needed. Discovery is by
pointer, not by browsing, and a page nobody points at when it is needed does not
exist.

That gap is the only part of this measurable without asking anybody. Every read
of ` + "`wiki`" + `, ` + "`usecase`" + `, ` + "`guide`" + `, ` + "`brief`" + ` and ` + "`needs`" + ` inside an
engagement appends a line to ` + "`docs/.reading-log`" + `, under the name of the
command that reads it; this reads it back.

**The briefings are the ones worth looking for.** Each exists because somebody
actually got that activity wrong, so whether a predecessor was briefed before
talking to a customer changes what you do - and for a long time those reads
were the ones this log did not record.

**It records page names of this tool and nothing about the customer**, so it is
safe to commit, and worth committing - six months later it says what the person
before you knew.

What it cannot see is a page opened and misread. That still needs a person.

**When the material has no answer, that is worth filing rather than working
around.** ` + "`asgard-cli issue-report --new`" + ` writes the report with what the tool
knows already filled in.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root := repo.Root(".")
			if root == "" {
				return errNotInRepo()
			}

			reads, err := work.Readings(root)
			if err != nil {
				return err
			}
			if len(reads) == 0 {
				fmt.Fprintf(out, "Nothing recorded yet. The log fills as `wiki`, `usecase`, `guide`,\n"+
					"`brief` and `needs` are used inside this repository.\n")
				return nil
			}

			opened := map[string]bool{}
			fmt.Fprintf(out, "Opened, most-read first:\n\n")
			for _, r := range reads {
				opened[r.Kind+"/"+r.Name] = true
				when := r.First
				if r.Last != r.First {
					when = r.First + " to " + r.Last
				}
				fmt.Fprintf(out, "  %2d  %-9s %-22s %s\n", r.Count, r.Kind, r.Name, when)
			}

			var never []string
			pages, err := wiki.List()
			if err != nil {
				return err
			}
			for _, p := range pages {
				if !opened["wiki/"+p.Name] {
					never = append(never, "wiki "+p.Name)
				}
			}
			extracts, err := usecase.List()
			if err != nil {
				return err
			}
			for _, e := range extracts {
				if !opened["usecase/"+e.Name] {
					never = append(never, "usecase "+e.Name)
				}
			}
			for _, s := range stage.List() {
				if !opened["guide/"+string(s.Name)] {
					never = append(never, "guide "+string(s.Name))
				}
			}
			// **The briefings are the column worth reading.** Each exists
			// because somebody got that activity wrong, so an engagement with
			// `customer-meeting` in this list has a person who walked into a
			// room without it - which is the whole reason the log is
			// committed. They were never recorded until now.
			for _, a := range brief.Activities {
				if !opened["brief/"+a.Name] {
					never = append(never, "brief "+a.Name)
				}
			}
			shapes := needs.Shapes()
			for _, s := range shapes {
				if !opened["needs/"+s.Name] {
					never = append(never, "needs "+s.Name)
				}
			}
			sort.Strings(never)

			total := len(pages) + len(extracts) + len(stage.List()) + len(brief.Activities) + len(shapes)
			fmt.Fprintf(out, "\n%d of %d pages opened.\n", total-len(never), total)

			if len(never) > 0 {
				fmt.Fprintf(out, "\nNever opened:\n\n")
				for _, n := range never {
					fmt.Fprintf(out, "  %s\n", n)
				}
				fmt.Fprintf(out, "\n**Most of these are irrelevant to this engagement and that is fine.**\n"+
					"The ones worth a minute are those you would expect to be relevant - a page\n"+
					"about the thing you spent a day on. Every expensive mistake found in this\n"+
					"material so far was a page in that column.\n")
			}
			return nil
		},
	}

	return cmd
}

// printMisses reports the searches that came back empty.
//
// In the order they happened, not collapsed by count: a run of misses on one
// afternoon is one subject somebody could not reach, and a frequency table
