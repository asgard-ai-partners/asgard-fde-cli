package cli

import (
	"fmt"
	"io"
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
	var misses bool

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

--misses is the other half, and it is the more actionable one: every search
this engagement ran that the material did not answer. **A query that found
nothing is the one part of a defect report nobody has to be believed about** -
everything else is somebody's account of what happened, and this is the tool's
own record that a search was run and the corpus had no answer. Each line is a
candidate row for ` + "`asgard-cli wiki --aliases`" + `, or a page that does not exist.

That file is not committed, because a query carries whatever words the customer
used. The reading log is, because it carries only page names of this tool.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root := repo.Root(".")
			if root == "" {
				return errNotInRepo()
			}

			if misses {
				return printMisses(out, root)
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

	cmd.Flags().BoolVar(&misses, "misses", false, "searches this engagement ran that the material did not answer")

	return cmd
}

// printMisses reports the searches that came back empty.
//
// In the order they happened, not collapsed by count: a run of misses on one
// afternoon is one subject somebody could not reach, and a frequency table
// would hide exactly that.
func printMisses(out io.Writer, root string) error {
	misses, err := work.Misses(root)
	if err != nil {
		return err
	}
	if len(misses) == 0 {
		fmt.Fprintf(out, "Nothing recorded in %s.\n\nEither every search here landed, or none has been run inside this\nrepository - `asgard-cli find` records a miss only when it is run in one.\n", work.MissLog)
		return nil
	}

	fmt.Fprintf(out, "%d search(es) the material did not answer, from %s:\n\n", len(misses), work.MissLog)
	for _, m := range misses {
		fmt.Fprintf(out, "  %s  %-9s %s\n", m.Date, m.Kind, m.Query)
	}
	fmt.Fprintf(out, "\n  miss      nothing in the four bodies carried any term\n"+
		"  unplaced  results came back, but these terms appeared in none of them\n\n"+
		"**Each line is one of two different defects, and they need different\nrepairs.** If the subject exists under another name, the index is missing a\n"+
		"row: `asgard-cli wiki --aliases` carries the rule for adding one. If it\ndoes not exist, the material is missing a page, and that is worth filing:\n\n"+
		"    asgard-cli issue-report --new\n\n"+
		"This file is not committed - a query carries whatever words the customer\nused - so it lives only as long as this checkout. File what it shows.\n")
	return nil
}
