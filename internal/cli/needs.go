package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/needs"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

// newNeedsCmd answers what a scenario has to be got from the customer.
//
// **This is the half of a meeting the rest of the tool could not answer.**
// Everything else here says what the platform has and how a shape is
// assembled; this says what the shape cannot start without, which is the
// question a customer's answer is needed for and the one that costs weeks when
// it is asked late.
//
// It resolves a scenario the way `find` does - the same search, the same alias
// index - so a scenario can be described in the customer's own words. What it
// prints is theirs to obtain, one line each, with the document that owns the
// claim beside it.
func newNeedsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "needs [terms]",
		Short: "What a scenario has to be got from the customer before it can be built",
		Long: `What an integration scenario has to be got from the customer before anybody
can build it.

Describe the scenario in whatever words it was described to you, including the
customer's:

    asgard-cli needs 庫存查詢
    asgard-cli needs "answer questions on LINE"
    asgard-cli needs semantic-layer            one shape, by name
    asgard-cli needs                           every shape

**Not what we will build - what they have to give us.** A credential, an
endpoint, a network path, a test environment, a login. The expensive mistake in
an engagement is not choosing the wrong shape; it is week three, when the
allowlist nobody asked for turns out to need a ticket, an approval and a window.

Every line says which document owns it, because none of this is new: it was
written across the interview, a wiki page and three extracts, and reaching it
meant reading all five and assembling the answer. **Read the owning document
before promising anything** - the line here is a reminder, not the reasoning.

What it cannot tell you is which shape a system actually presents. That is the
customer's answer, not this tool's: an open API with a test environment, an open
API production-only, a data export, or only a web back office. The four are on
` + "`asgard-cli wiki taiwan-channels`" + ` and they differ in cost by more than an order
of magnitude.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if err := checkFormat(format); err != nil {
				return err
			}

			var found []needs.Shape
			switch {
			case len(args) == 0:
				found = needs.Shapes()
			default:
				query := strings.Join(args, " ")
				if s, ok := needs.Find(query); ok {
					found = []needs.Shape{s}
					break
				}
				var err error
				var landed []string
				found, landed, err = resolveNeeds(query)
				if err != nil {
					return err
				}
				// Which words carried the result, before showing it. A
				// scenario described in a sentence has common words in it, and
				// a partial match reads exactly like a whole one - the same
				// property `find` reports rather than hides.
				if format != formatJSON {
					reportTerms(out, query, landed)
				}
				if len(found) == 0 {
					return noShapeMatched(out, query)
				}
			}

			if format == formatJSON {
				return writeJSON(out, found)
			}
			// One row per shape actually printed, so a scenario that landed on
			// three of them says so. Not recorded for the bare listing, which
			// is a table of contents rather than a page.
			if len(args) > 0 {
				for _, s := range found {
					recallHere("needs", s.Name)
				}
			}
			printNeeds(out, found)
			return nil
		},
	}

	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	return cmd
}

// resolveNeeds maps a scenario to the shapes it touches.
//
// **It searches term by term and unions the result, which is the opposite of
// what `find` does and right here.** `find` drops the partial matches once
// something carries every term, because an extra word should narrow a lookup.
// A scenario is not a lookup: it is a sentence with common words in it, and it
// touches several shapes on purpose. "answer questions on LINE" carries three
// words that reach almost everything and one that reaches the answer, and the
// narrowing rule threw the answer away.
//
// Ranked by how many of the query's terms reached each shape, so the sentence's
// distinctive word still decides the order.
func resolveNeeds(query string) ([]needs.Shape, []string, error) {
	search := query
	if rewritten := translate(query); rewritten != "" {
		search = rewritten
	}
	var out []needs.Shape
	landedSet := map[string]bool{}
	seen := map[string]bool{}

	// **What the index already says, before any searching.** A row in
	// `asgard-cli wiki --aliases` may carry a pointer - 客服 says "customer
	// service, help desk - and `asgard-cli usecase chat-channel`" - and that
	// pointer is somebody's recorded answer to this exact question. Searching
	// the translated words instead ranked external-api above chat-channel for
	// 客服, which is the index knowing the answer and nothing using it.
	for _, name := range indexedShapes(query) {
		s, ok := needs.Find(name)
		if !ok || seen[s.Name] {
			continue
		}
		seen[s.Name] = true
		out = append(out, s)
	}

	// Then the ranked search: it orders by how much of the query a document
	// carries, which is what puts chat-channel at the top of "LINE".
	matches, err := usecase.Search(search)
	if err == nil {
		for _, m := range matches {
			s, ok := needs.Find(m.Name)
			if !ok || seen[s.Name] {
				continue
			}
			seen[s.Name] = true
			out = append(out, s)
			for _, t := range m.Terms {
				landedSet[t] = true
			}
		}
	}

	// Then whatever the narrowing rule cut. `usecase.Search` drops the partial
	// matches once some document carries every term, which is right for a
	// lookup and wrong for a sentence: "answer questions on LINE" has three
	// words that reach almost everything and one that reaches the answer, and
	// the rule threw the answer away. Recovered after the ranked ones rather
	// than mixed in, so a distinctive term still decides the order.
	for _, term := range kb.Terms(search) {
		byTerm, err := usecase.Search(term)
		if err != nil {
			continue // one word the corpus rejects is not the query
		}
		for _, m := range byTerm {
			s, ok := needs.Find(m.Name)
			if !ok || seen[s.Name] {
				continue
			}
			seen[s.Name] = true
			out = append(out, s)
			for _, t := range m.Terms {
				landedSet[t] = true
			}
		}
	}

	landed := make([]string, 0, len(landedSet))
	for t := range landedSet {
		landed = append(landed, t)
	}
	return out, landed, nil
}

// indexedShapes returns the shapes the alias index points at for this query.
//
// A row's pointer is a fact recorded by whoever added the row, and
// `audit-material --links` resolves it like every other pointer here - so this
// reads the index rather than re-deriving what it already holds.
func indexedShapes(query string) []string {
	var out []string
	seen := map[string]bool{}
	tables := []map[string]string{wiki.Aliases()}
	for _, e := range wiki.EntityRows() {
		tables = append(tables, map[string]string{e.Word: e.Search})
	}
	for _, term := range kb.Terms(query) {
		for _, table := range tables {
			row := lookup(table, term)
			if row == "" {
				continue
			}
			links, _ := kb.Links(row)
			for _, l := range links {
				if l.Kind == "usecase" && !seen[l.Name] {
					seen[l.Name] = true
					out = append(out, l.Name)
				}
			}
		}
	}
	return out
}

func printNeeds(out io.Writer, found []needs.Shape) {
	fmt.Fprintf(out, "What the customer has to give us, by the shape it is for.\n\n"+
		"**This is theirs to provide, not ours to design.** Ask for exactly the thing\n"+
		"named - offering options invites the other side to pick one that does not\n"+
		"apply, and the week it takes to find that out is the week you were saving.\n\n")

	for _, s := range found {
		fmt.Fprintf(out, "%s - %s\n\n", strings.ToUpper(s.Name), s.What)
		for _, it := range s.Items {
			fmt.Fprintf(out, "  - %s\n", wrapAt(it.Ask, 72, 4))
			fmt.Fprintf(out, "    %s\n", wrapAt(it.Why, 70, 4))
			fmt.Fprintf(out, "    %s\n\n", it.From)
		}
	}

	fmt.Fprintf(out, "**Which shape a system actually presents is the customer's answer, not this\n"+
		"tool's.** An open API with a test environment, an open API production-only, a\n"+
		"data export, or only a web back office - the four are on\n"+
		"`asgard-cli wiki taiwan-channels` and they differ in cost by more than an\n"+
		"order of magnitude.\n\n"+
		"Write each answer down where it will be answered rather than remembered:\n\n"+
		"    asgard-cli question add \"<what is still unknown>\" --ask \"<who at the customer>\"\n")
}

// noShapeMatched is the dead end, and it hands over the list rather than
// advice: a scenario nothing matched is usually named differently here, and the
// shapes are few enough to read.
func noShapeMatched(out io.Writer, query string) error {
	fmt.Fprintf(out, "No shape here matched %q.\n\n"+
		"That is usually the subject being named differently rather than absent -\n"+
		"`asgard-cli find %s` searches every part of the material and translates\n"+
		"from the customer's words first. The shapes that carry a dependency list:\n\n",
		query, query)

	names := needs.Names()
	sort.Strings(names)
	for _, n := range names {
		s, _ := needs.Find(n)
		fmt.Fprintf(out, "  %-20s %s\n", n, s.What)
	}
	fmt.Fprintf(out, "\nA shape with no list here is not one with no dependencies - it is one\n"+
		"nobody has written them down for. `asgard-cli issue-report --new` is how\n"+
		"that gets fixed for the next engagement.\n")
	return nil
}
