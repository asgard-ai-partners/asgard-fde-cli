package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

func newWikiCmd() *cobra.Command {
	var (
		search      string
		conventions bool
		unverified  bool
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

With no arguments it lists the pages. Naming one prints it in full. --search
finds pages by what the customer said, which is useful when you know the
requirement but not what the feature is called:

    asgard-cli wiki agents
    asgard-cli wiki --search "匿名 訪客"
    asgard-cli wiki --search dashboard

--conventions prints how the wiki is maintained: where its sources are, what a
page has to carry, and how it is kept from going stale as the platform moves.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

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

			if search != "" {
				matches, err := wiki.Search(search)
				if err != nil {
					return err
				}
				if len(matches) == 0 {
					fmt.Fprintf(out, "Nothing matched %q. List every page with `asgard-cli wiki`.\n", search)
					return nil
				}
				fmt.Fprintf(out, "%d page(s) mention %q:\n\n", len(matches), search)
				for _, m := range matches {
					fmt.Fprintf(out, "  %s\n    %s\n", m.Name, m.Title)
					for _, line := range m.Lines {
						fmt.Fprintf(out, "      %s\n", truncate(line, 96))
					}
					fmt.Fprintln(out)
				}
				fmt.Fprintf(out, "Read one with `asgard-cli wiki <page>`.\n")
				return nil
			}

			if len(args) == 1 {
				text, err := wiki.Read(args[0])
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
Read one with "asgard-cli wiki <page>", or find one by what the customer said
with "asgard-cli wiki --search <terms>".

For how a deployment shape is actually assembled, "asgard-cli usecase".
For how the wiki is maintained, "asgard-cli wiki --conventions".
`)
			return nil
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "find pages mentioning all of these terms")
	cmd.Flags().BoolVar(&conventions, "conventions", false, "print how the wiki is maintained and where its sources are")
	cmd.Flags().BoolVar(&unverified, "unverified", false,
		"list what each page has NOT been held against a real deployment")

	return cmd
}
