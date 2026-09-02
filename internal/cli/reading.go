package cli

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/config"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newReadingCmd() *cobra.Command {
	return &cobra.Command{
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
of ` + "`wiki`" + `, ` + "`usecase`" + ` and ` + "`next --stage`" + ` inside an engagement appends a line to
` + "`docs/.reading-log`" + `; this reads it back.

**It records page names of this tool and nothing about the customer**, so it is
safe to commit, and worth committing - six months later it says what the person
before you knew.

What it cannot see is a page opened and misread. That still needs a person.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			path, err := config.Find(".")
			if err != nil {
				return err
			}
			root := filepath.Dir(path)

			reads, err := work.Readings(root)
			if err != nil {
				return err
			}
			if len(reads) == 0 {
				fmt.Fprintf(out, "Nothing recorded yet. The log fills as `wiki`, `usecase` and\n"+
					"`next --stage` are used inside this repository.\n")
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
				if !opened["stage/"+string(s.Name)] {
					never = append(never, "next --stage "+string(s.Name))
				}
			}
			sort.Strings(never)

			total := len(pages) + len(extracts) + len(stage.List())
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
}
