package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
)

func newUsecaseCmd() *cobra.Command {
	var search string

	cmd := &cobra.Command{
		Use:   "usecase [name]",
		Short: "How each shape of deployment is built, from ones already in production",
		Long: `How each shape of deployment is built, taken from deployments already in
production.

Read one before authoring a chart. Each extract answers four questions: when to
use the shape and when not, which CRs it needs and how they reference each
other, the fields that are not obvious and what a wrong value does, and what the
mistake cost someone the last time.

With no arguments it lists the shapes. Naming one prints it in full. --search
finds shapes by what the customer asked for, which is useful when you know the
requirement but not what the shape is called:

    asgard-cli usecase --search "public anonymous"
    asgard-cli usecase --search schedule
    asgard-cli usecase flow-agent-supervisor`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if search != "" {
				matches, err := usecase.Search(search)
				if err != nil {
					return err
				}
				if len(matches) == 0 {
					fmt.Fprintf(out, "Nothing matched %q. List every shape with `asgard-cli usecase`.\n", search)
					return nil
				}
				fmt.Fprintf(out, "%d shape(s) mention %q:\n\n", len(matches), search)
				for _, m := range matches {
					fmt.Fprintf(out, "  %s\n    %s\n", m.Name, m.Title)
					for _, line := range m.Lines {
						fmt.Fprintf(out, "      %s\n", truncate(line, 96))
					}
					fmt.Fprintln(out)
				}
				fmt.Fprintf(out, "Read one with `asgard-cli usecase <name>`.\n")
				return nil
			}

			if len(args) == 1 {
				content, err := usecase.Read(args[0])
				if err != nil {
					return err
				}
				fmt.Fprint(out, content)
				return nil
			}

			list, err := usecase.List()
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Shapes, each taken from a deployment already in production:\n\n")
			for _, e := range list {
				fmt.Fprintf(out, "  %-24s %s\n", e.Name, e.Title)
				if e.Summary != "" {
					fmt.Fprintf(out, "  %-24s %s\n", "", truncate(e.Summary, 88))
				}
			}
			fmt.Fprintf(out, `
Read one with "asgard-cli usecase <name>", or find one by what the customer
asked for with "asgard-cli usecase --search <terms>".

Which entry point and which read path are not preferences: both follow from
whether the caller can authenticate. See "asgard-cli next --stage read-path".
`)
			return nil
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "find shapes mentioning all of these terms")

	return cmd
}

// truncate keeps a listing readable without hiding that it was cut.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n]) + "..."
}
