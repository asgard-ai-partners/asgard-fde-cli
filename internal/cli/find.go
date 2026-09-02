package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

// pointer finds the cross-reference a page or extract makes to the other body
// of material. That link is what makes a search in one language reach material
// written in the other.
var (
	toWiki    = regexp.MustCompile(`asgard-cli wiki ([a-z][a-z-]*)`)
	toUsecase = regexp.MustCompile(`asgard-cli usecase ([a-z][a-z-]*)`)
)

func newFindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find <terms>",
		Short: "Search both bodies of material at once",
		Long: `Search the platform wiki and the deployment extracts together.

Three bodies answer different questions - what the platform has, how one shape is
assembled, and what to weigh when deciding - and which of them holds an answer is
often not obvious before searching. This searches all three.

The stage prompts are searchable **by subject** rather than only by position in
the walk. An onboarding is not linear - three of this engagement's most expensive
decisions were reversed after contact with reality - so an agent that reads the
repository and forms its own view of where things stand is doing the right thing.
What it then needs is the guidance for the question at hand, without having to
arrive at a stage to be handed it.

It also **follows the link between them**: a wiki page that matches names the
extract covering the same subject at field level, and an extract names its wiki
page. So a hit in either half hands over the other, in the order they should be
read.

    asgard-cli find schedule
    asgard-cli find anonymous visitor
    asgard-cli find knowledge graph

Every term has to appear, so an extra word narrows rather than widens.

To read one in full: "asgard-cli wiki <page>" or "asgard-cli usecase <name>".`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			query := strings.Join(args, " ")

			pages, err := wiki.Search(query)
			if err != nil {
				return err
			}
			extracts, err := usecase.Search(query)
			if err != nil {
				return err
			}
			stages, err := stage.Search(query)
			if err != nil {
				return err
			}

			if len(pages) == 0 && len(extracts) == 0 && len(stages) == 0 {
				fmt.Fprintf(out, "Nothing matched %q in any of the three bodies.\n\n"+
					"They are searched literally and every term has to appear, so try fewer\n"+
					"words. List them with `asgard-cli wiki`, `asgard-cli usecase` and\n"+
					"`asgard-cli next --list`.\n", query)
				return nil
			}

			if len(pages) > 0 {
				fmt.Fprintf(out, "PLATFORM - what the thing is (asgard-cli wiki <page>)\n\n")
				for _, m := range pages {
					fmt.Fprintf(out, "  %-18s %s\n", m.Name, m.Title)
					for _, line := range m.Lines {
						fmt.Fprintf(out, "  %-18s %s\n", "", truncate(line, 84))
					}
					if body, err := wiki.Read(m.Name); err == nil {
						if to := firstRef(toUsecase, body); to != "" {
							fmt.Fprintf(out, "  %-18s -> field level: asgard-cli usecase %s\n", "", to)
						}
					}
					fmt.Fprintln(out)
				}
			}

			if len(extracts) > 0 {
				fmt.Fprintf(out, "SHAPES - how it is assembled (asgard-cli usecase <name>)\n\n")
				for _, m := range extracts {
					fmt.Fprintf(out, "  %-18s %s\n", m.Name, m.Title)
					for _, line := range m.Lines {
						fmt.Fprintf(out, "  %-18s %s\n", "", truncate(line, 84))
					}
					if body, err := usecase.Read(m.Name); err == nil {
						if to := firstRef(toWiki, body); to != "" {
							fmt.Fprintf(out, "  %-18s -> what it is:  asgard-cli wiki %s\n", "", to)
						}
					}
					fmt.Fprintln(out)
				}
			}

			if len(stages) > 0 {
				fmt.Fprintf(out, "DECISIONS - what to weigh at this point in the work (asgard-cli next --stage <name>)\n\n")
				for _, m := range stages {
					fmt.Fprintf(out, "  %-18s %s\n", m.Name, m.Title)
					for _, line := range m.Lines {
						fmt.Fprintf(out, "  %-18s %s\n", "", truncate(line, 84))
					}
					fmt.Fprintln(out)
				}
			}

			// The reading order is the same whichever half you landed in.
			fmt.Fprintf(out, "Read the platform side first; an extract assumes you have.\n")
			return nil
		},
	}

	return cmd
}

// primarySection is where a page or extract names its counterpart deliberately,
// as opposed to mentioning one in passing. Taking the first match anywhere sends
// a reader to whichever reference happened to appear earliest, which on the
// knowledge page was the CD note pointing at skill-set rather than the page's
// own knowledge-drive.
var primarySection = regexp.MustCompile(
	`(?s)##+ (?:Before writing the chart|Corresponding extracts|Read the platform side first)[^
]*
(.*?)(?:
##|\z)`)

func firstRef(re *regexp.Regexp, body string) string {
	if sec := primarySection.FindStringSubmatch(body); sec != nil {
		if m := re.FindStringSubmatch(sec[1]); m != nil {
			return m[1]
		}
	}
	if m := re.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return ""
}
