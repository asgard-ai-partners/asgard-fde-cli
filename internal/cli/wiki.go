package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

func newWikiCmd() *cobra.Command {
	var (
		conventions bool
		unverified  bool
		aliases     bool
		sources     bool
	)

	cmd := &cobra.Command{
		Use:   "wiki [page]",
		Short: "What the platform is made of, and who each piece is for",
		Long: `What the Asgard platform is made of, and who each piece is for.

An agent picking up a customer repository has that repository and nothing else:
it describes one customer's systems, never the platform those systems run on.
These pages are that missing half - which product a request lands in, what the
pieces are called in the UI, and where those names stop matching the resources a
chart declares.

It answers a different question from "asgard-cli usecase". An extract there says
how one shape of deployment is assembled, field by field, and assumes the reader
already knows the platform has that shape. These pages are where that assumption
comes from, so they deliberately do not repeat what an extract already covers.

With no arguments it lists the pages. Naming one prints it in full.

    asgard-cli wiki agents

**To look something up, use "asgard-cli find" instead.** It searches this and the
extracts together and names the counterpart of whatever it finds.

There was a --search here that searched this half alone. It went because it was
superseded twice: "find" already did it better, and once "asgard-cli init"
started writing these pages into a repository, grep did it without a
subprocess.

--conventions prints how the wiki is maintained: where its sources are, what a
page has to carry, and how it is kept from going stale as the platform moves.

--sources prints the documentation links a page cites - one page, or every page
with no argument. An engagement building a customer deck copied nine of them out
of the Sources blocks by hand, one page at a time; this is that step.

--aliases prints the index: what a customer says, and what to search for. It is
what "asgard-cli find" applies to a query before searching, and it is not a page
- an index inside the corpus competes with what it points at, so it lives beside
the pages the way index.md and log.md do.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if sources {
				pages, err := wiki.List()
				if err != nil {
					return err
				}
				if len(args) == 1 {
					page, err := wiki.Read(args[0])
					if err != nil {
						return err
					}
					pages = []wiki.Page{{Name: args[0], Sources: kb.SourceURLs(page)}}
				}
				printSources(out, pages, func(n string) string { return "asgard-cli wiki " + n },
					"No documentation link cited. A page with no product documentation behind\nit says so in its Sources block rather than leaving a blank - `console`,\n`fehu`, `operations` and `product-suite` are the ones where that is a\ndecision rather than an omission.\n")
				return nil
			}

			if aliases {
				body, err := wiki.Index()
				if err != nil {
					return err
				}
				fmt.Fprint(out, body)
				return nil
			}

			if unverified {
				pages, err := wiki.List()
				if err != nil {
					return err
				}
				fmt.Fprintf(out, `These pages are distilled from the product documentation, which describes a
UI. Much of what it describes is not in any chart, so they are checked less
deeply than "asgard-cli usecase" and each says how far it got.

`)
				for _, p := range pages {
					if !p.Verified() {
						fmt.Fprintf(out, "  %-16s NOTHING RECORDED - unknown, not fine\n\n", p.Name)
						continue
					}
					fmt.Fprintf(out, "  %s\n    %s\n\n", p.Name, truncate(p.Unchecked, 74))
				}
				fmt.Fprintf(out, "For the same on deployment shapes, `asgard-cli usecase --unverified`.\n")
				return nil
			}

			if conventions {
				text, err := wiki.Conventions()
				if err != nil {
					return err
				}
				fmt.Fprint(out, text)
				return nil
			}

			if len(args) == 1 {
				text, err := wiki.Read(args[0])
				recallHere("wiki", args[0])
				if err != nil {
					return err
				}
				fmt.Fprint(out, text)
				return nil
			}

			pages, err := wiki.List()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "What the platform is made of:\n\n")
			for _, p := range pages {
				fmt.Fprintf(out, "  %-20s %s\n", p.Name, p.Title)
				if p.Summary != "" {
					fmt.Fprintf(out, "  %-20s %s\n", "", truncate(p.Summary, 88))
				}
			}
			fmt.Fprintf(out, `
Read one with "asgard-cli wiki <page>". To look something up, "asgard-cli find
<terms>" searches this and the extracts together.

For how a deployment shape is actually assembled, "asgard-cli usecase".
For how the wiki is maintained, "asgard-cli wiki --conventions".
`)
			return nil
		},
	}

	cmd.Flags().BoolVar(&conventions, "conventions", false, "print how the wiki is maintained and where its sources are")
	cmd.Flags().BoolVar(&aliases, "aliases", false, "print the index: what a customer says, and what to search for")
	cmd.Flags().BoolVar(&sources, "sources", false, "print the documentation links a page cites; every page with no argument")
	cmd.Flags().BoolVar(&unverified, "unverified", false,
		"list what each page has NOT been held against a real deployment")

	return cmd
}
