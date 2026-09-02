package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
)

func newUsecaseCmd() *cobra.Command {
	var (
		search     string
		unverified bool
	)

	cmd := &cobra.Command{
		Use:   "usecase [name]",
		Short: "How each shape of deployment is built, from ones already in production",
		Long: `How each shape of deployment is built, taken from deployments already in
production.

Read one before authoring a chart. Each extract answers four questions: when to
use the shape and when not, which CRs it needs and how they reference each
other, the fields that are not obvious and what a wrong value does, and what the
mistake cost someone the last time.

An extract assumes you already know the platform has that shape. Where that
assumption comes from is "asgard-cli wiki", and every extract names the page
that covers it. Read that first if the shape itself is new to you.

Not all of them carry the same weight. Each opens with what it was held against
and what it was not - a field name is checkable against the CRD, a shape against
a deployment, and a design rationale against nothing at all. --unverified lists
that for every shape at once; it is worth reading before a first engagement.

With no arguments it lists the shapes. Naming one prints it in full.

    asgard-cli usecase flow-agent-supervisor
    asgard-cli usecase --search schedule      # this half only

**To look something up, use "asgard-cli find" instead.** It searches this and the
wiki together and names the counterpart of whatever it finds, which is what you
want when you know the requirement but not which half holds the answer. --search
here is the narrow form, for when you already know it is a deployment shape.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if unverified {
				list, err := usecase.List()
				if err != nil {
					return err
				}
				fmt.Fprintf(out, `Every shape carries what it was checked against and what it was not.
Weight a claim by that, and re-check before betting a deployment on one.

`)
				for _, e := range list {
					if !e.Verified() {
						fmt.Fprintf(out, "  %-24s NOTHING RECORDED - nobody has written down what this was\n", e.Name)
						fmt.Fprintf(out, "  %-24s checked against. That is unknown, not fine.\n\n", "")
						continue
					}
					fmt.Fprintf(out, "  %s\n", e.Name)
					fmt.Fprintf(out, "    checked    %s\n", wrapAt(e.Checked, 68, 15))
					if e.Unchecked != "" {
						fmt.Fprintf(out, "    unchecked  %s\n", wrapAt(e.Unchecked, 68, 15))
					}
					fmt.Fprintln(out)
				}
				fmt.Fprintf(out, "Read one in full with `asgard-cli usecase <name>`.\n")
				return nil
			}

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
				recallHere("usecase", args[0])
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
Read one with "asgard-cli usecase <name>". To look something up, "asgard-cli
find <terms>" searches this and the wiki together.

Every shape says what it was checked against and what it was not. Before betting
a deployment on one, see "asgard-cli usecase --unverified" - some of these are
read off the platform contract and have never run anywhere.

Which entry point and which read path are not preferences: both follow from
whether the caller can authenticate. See "asgard-cli next --stage read-path".

For what the platform is and who each piece is for, "asgard-cli wiki".
`)
			return nil
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search this half only; `asgard-cli find` searches both")
	cmd.Flags().BoolVar(&unverified, "unverified", false,
		"list only what has NOT been held against a real deployment, and what about each is unchecked")

	return cmd
}

// truncate keeps a listing readable without hiding that it was cut.
// wrapAt breaks a provenance line so it stays readable in a terminal, indenting
// continuations to line up under the first.
func wrapAt(s string, width, indent int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	line := 0
	for i, w := range words {
		if i > 0 && line+1+len([]rune(w)) > width {
			b.WriteString("\n" + strings.Repeat(" ", indent))
			line = 0
		} else if i > 0 {
			b.WriteString(" ")
			line++
		}
		b.WriteString(w)
		line += len([]rune(w))
	}
	return b.String()
}

// truncate cuts to n runes, not n bytes. Slicing a string by byte splits a
// multi-byte character in half and prints a replacement glyph, which the wiki
// pages hit on every line because they are written in Chinese.
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "..."
}
