package cli

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
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

Four bodies answer different questions - what the platform has, how one shape is
assembled, what to weigh when deciding, and what the agent in a customer repo
loads to do one kind of work - and which of them holds an answer is often not
obvious before searching. This searches all four.

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

An extra term narrows: what carries every one of them is listed first and alone.
When nothing carries all of them the search widens rather than returning nothing,
and says which terms it could not place - a query taken from a customer's own
document usually has a word or two this material has never heard of, and one of
them should not suppress what the rest would have found.

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
			skills, err := scaffold.SearchSkills(query)
			if err != nil {
				return err
			}

			if len(pages) == 0 && len(extracts) == 0 && len(stages) == 0 && len(skills) == 0 {
				return nothingMatched(out, query)
			}

			// Say which words carried the result before showing it. A partial
			// match reads exactly like a whole one, and a reader who does not
			// know that "shopee" found nothing will take what came back as the
			// material on the subject they asked about.
			reportTerms(out, query, matchedTerms(pages), matchedTerms(extracts), stageTerms(stages), skillTerms(skills))

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

			if len(skills) > 0 {
				fmt.Fprintf(out, "SKILLS - what the agent in the customer repo loads (.agents/skills/)\n\n")
				for _, m := range skills {
					fmt.Fprintf(out, "  %-18s %s\n", m.Name, truncate(m.Description, 84))
					for _, line := range m.Lines {
						fmt.Fprintf(out, "  %-18s %s\n", "", truncate(line, 84))
					}
					fmt.Fprintf(out, "  %-18s -> read it:     %s\n\n", "", m.Path)
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

// reportTerms names the query terms that appear nowhere in what was returned.
func reportTerms(out io.Writer, query string, sets ...[]string) {
	terms := kb.Terms(query)
	if len(terms) < 2 {
		return
	}

	found := map[string]bool{}
	for _, set := range sets {
		for _, t := range set {
			found[t] = true
		}
	}

	var missing []string
	for _, t := range terms {
		if !found[t] {
			missing = append(missing, t)
		}
	}
	if len(missing) == 0 {
		return
	}

	fmt.Fprintf(out, "Matched %d of %d terms. These results answer that much of the query, not\n"+
		"all of it. Nothing here mentions:\n\n    %s\n\n",
		len(terms)-len(missing), len(terms), strings.Join(missing, "  "))
}

// nothingMatched is the dead end, and it is the one place the tool can lose a
// reader entirely: an agent that searches twice, gets nothing twice, and falls
// back on what it already believed has stopped using the material this command
// exists to serve. So it does not offer advice about phrasing - it hands over
// the contents, which is what the reader would have to ask for next.
func nothingMatched(out io.Writer, query string) error {
	fmt.Fprintf(out, "No term in %q appears anywhere in the four bodies.\n\n"+
		"That usually means the subject is named differently here, not that it is\n"+
		"absent. This material is in English, and it describes platform parts rather\n"+
		"than a customer's systems: a marketplace integration is under whichever\n"+
		"shape reaches it, a device protocol is a question about the sandbox.\n\n"+
		"Everything there is, in full:\n\n", query)

	pages, err := wiki.List()
	if err != nil {
		return err
	}
	extracts, err := usecase.List()
	if err != nil {
		return err
	}

	listDocs(out, "PLATFORM - what the thing is", "asgard-cli wiki <page>", pages)
	listDocs(out, "SHAPES - how it is assembled", "asgard-cli usecase <name>", extracts)

	fmt.Fprintf(out, "DECISIONS - what to weigh (asgard-cli next --stage <name>)\n\n")
	for _, s := range stage.List() {
		fmt.Fprintf(out, "  %-18s %s\n", s.Name, truncate(s.Title, 84))
	}
	fmt.Fprintln(out)

	skills, err := scaffold.Skills()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "SKILLS - loaded by the agent in the customer repo (.agents/skills/<name>/SKILL.md)\n\n")
	for _, s := range skills {
		fmt.Fprintf(out, "  %-24s %s\n", s.Name, truncate(s.Description, 78))
	}
	fmt.Fprintln(out)
	return nil
}

func listDocs(out io.Writer, heading, command string, docs []kb.Doc) {
	fmt.Fprintf(out, "%s (%s)\n\n", heading, command)
	for _, d := range docs {
		fmt.Fprintf(out, "  %-18s %s\n", d.Name, truncate(d.Title, 84))
	}
	fmt.Fprintln(out)
}

// matchedTerms and stageTerms collect what landed. Two functions because the
// stage corpus is not a kb.Corpus - it is the prompts, which are files with a
// position in the walk - so its Match carries a Stage rather than a Doc.
func matchedTerms(matches []kb.Match) []string {
	var out []string
	for _, m := range matches {
		out = append(out, m.Terms...)
	}
	return out
}

func stageTerms(matches []stage.Match) []string {
	var out []string
	for _, m := range matches {
		out = append(out, m.Terms...)
	}
	return out
}

func skillTerms(matches []scaffold.SkillMatch) []string {
	var out []string
	for _, m := range matches {
		out = append(out, m.Terms...)
	}
	return out
}
