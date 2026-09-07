package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/repo"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/scaffold"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/stage"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/usecase"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/work"
)

func newFindCmd() *cobra.Command {
	var (
		format     string
		unverified bool
	)

	cmd := &cobra.Command{
		Use:   "find <terms>",
		Short: "Search all four parts of the corpus at once",
		Long: `Search the wiki, the extracts, the stage guidance and the skills together.

Four parts answer different questions - what the platform has, how one shape is
assembled, what to weigh when deciding, and what the agent in a customer repo
loads to do one kind of work - and which of them holds an answer is often not
obvious before searching. This searches all four.

The stage prompts are searchable **by subject** rather than only by position in
the work. An onboarding is not linear - three of this engagement's most expensive
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
    asgard-cli find --unverified             what has never been held against anything
    asgard-cli find knowledge graph

An extra term narrows: what carries every one of them is listed first and alone.
When nothing carries all of them the search widens rather than returning nothing,
and says which terms it could not place - a query taken from a customer's own
document usually has a word or two this material has never heard of, and one of
them should not suppress what the rest would have found.

**Ask in Chinese.** The corpus is English and a customer conversation is not, so
a term the glossary has a name for is translated before the search runs and the
rewrite is printed. Terms with no entry are searched as they were, and any that
land nowhere are named at the top of the results along with the word this
material uses, if it has one. The table is on the glossary page:

    asgard-cli wiki glossary

To read one in full: "asgard-cli wiki <page>" or "asgard-cli usecase <name>".

--format json emits each hit with the command that reads it in full, the
counterpart it names, and the terms it actually carried - plus the query as
asked and the query as searched, which differ whenever the glossary had a name
for one of the words.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if err := checkFormat(format); err != nil {
				return err
			}
			if unverified {
				return printUnverified(out)
			}
			query := strings.Join(args, " ")

			// Translate before searching rather than after failing. The
			// corpus is English and the question was asked in a meeting that
			// was not, so a word with a name here should reach the material
			// every time - not only on the queries that happen to match
			// nothing at all.
			//
			// Doing it the other way round was worse than doing nothing: the
			// glossary carries the table, so a Chinese term matched the
			// glossary row about itself, the search was no longer empty, and
			// the reader was handed the word list instead of the answer.
			search := query
			rewritten := translate(query)
			if rewritten != "" {
				search = rewritten
			}

			found, err := searchAll(search)
			if err != nil {
				return err
			}

			if format == formatJSON {
				return writeJSON(out, findReport(query, search, found))
			}

			if rewritten != "" {
				fmt.Fprintf(out, "This material is in English. %q was read as:\n\n    %s\n\n", query, search)
			}
			if found.empty() {
				recordMiss("miss", query)
				return nothingMatched(out, query)
			}
			// A term nothing carried is a candidate index row just as much as
			// a query that missed outright, and it is the commoner case: five
			// words of a customer's sentence land and the sixth - their word
			// for the subject - lands nowhere.
			if unplaced := unmatched(search, found.terms()); len(unplaced) > 0 {
				recordMiss("unplaced", strings.Join(unplaced, " "))
			}
			reportRouted(out, query)
			// The glossary row is printed above in full when a sense fires, so
			// the page is dropped from the results: the same row twice, once
			// as an explanation and once as a search hit, reads as two
			// findings and is one.
			if reportSenses(out, query, search) {
				found.drop("asgard-cli wiki glossary")
			}
			printResults(out, search, found)
			return nil
		},
	}

	cmd.Flags().StringVar(&format, formatFlag, formatText, formatUsage)
	cmd.Flags().BoolVar(&unverified, "unverified", false,
		"list what carries no record of having been held against anything, across all four parts")

	return cmd
}

// printUnverified lists what says nothing about having been checked.
//
// A document carrying neither a Checked nor an Unchecked line is UNKNOWN, not
// fine - see kb.Doc.Verified. The wiki and the extracts carry one each; the
// guidance and the skills carry none, and that is a true report rather than a
// formatting gap. **Do not close it by writing the lines.** An `Unchecked:`
// line written to satisfy a listing converts UNKNOWN into a claim, which is
// worse than the silence it replaces.
func printUnverified(out io.Writer) error {
	fmt.Fprintf(out, "What carries no record of having been held against anything.\n\n"+
		"Neither line present means UNKNOWN - not that the document is wrong, and\nnot that it is right.\n\n")

	for _, p := range parts() {
		docs, err := p.list()
		if err != nil {
			return err
		}
		var bare []kb.Doc
		for _, d := range docs {
			if !d.Verified() {
				bare = append(bare, d)
			}
		}
		fmt.Fprintf(out, "%s\n  %d of %d\n", strings.Fields(p.heading)[0], len(bare), len(docs))
		for _, d := range bare {
			fmt.Fprintf(out, "    %s\n", d.Name)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "For what a checked document says it has NOT been held against:\n"+
		"`asgard-cli wiki --unverified`, `asgard-cli usecase --unverified`.\n")
	return nil
}

// findJSON is what `find --format json` emits. It names the query as asked and
// the query as searched, because the two differ whenever the glossary had a
// name for one of the words and a caller acting on the results needs to know
// which question was actually answered.
type findJSON struct {
	Query    string    `json:"query"`
	Searched string    `json:"searched"`
	Missed   []string  `json:"termsFoundNowhere"`
	Aliases  []string  `json:"aliases"`
	Results  []hitJSON `json:"results"`
}

type hitJSON struct {
	Part        string   `json:"part"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Read        string   `json:"read"`
	Counterpart string   `json:"counterpart,omitempty"`
	Lines       []string `json:"lines"`
	Terms       []string `json:"terms"`
}

func findReport(query, searched string, r results) findJSON {
	out := findJSON{
		Query:    query,
		Searched: searched,
		Missed:   []string{},
		Aliases:  []string{},
		Results:  []hitJSON{},
	}

	found := map[string]bool{}
	for _, t := range r.terms() {
		found[t] = true
	}
	for _, t := range kb.Terms(searched) {
		if !found[t] {
			out.Missed = append(out.Missed, t)
		}
	}
	out.Aliases = append(out.Aliases, aliasesFor(kb.Terms(query))...)

	for i, p := range r.parts {
		for _, m := range r.hits[i] {
			h := hitJSON{
				Part:  strings.ToLower(strings.Fields(p.heading)[0]),
				Name:  m.Name,
				Title: m.Title,
				Read:  p.read(m.Name),
				Lines: m.Lines,
				Terms: m.Terms,
			}
			if p.counterpart != "" {
				if to := m.Counterpart(p.counterpart); to != "" {
					h.Counterpart = p.opens(to)
				}
			}
			out.Results = append(out.Results, h)
		}
	}
	return out
}

// part is one part of the corpus, how a reader opens a document in it, and
// which counterpart a hit there should hand over.
//
// There were four of these written out four times - four search calls, four
// result loops, three helpers to collect what matched - because two of the four
// were not kb corpora and so could not share a type. They are now, so this is a
// table and one loop.
type part struct {
	heading string
	search  func(string) ([]kb.Match, error)
	list    func() ([]kb.Doc, error)
	read    func(name string) string

	// counterpart is the kind of document a hit here should hand over, and
	// opens says how to read it. The pointer itself comes off kb.Doc, which
	// carries every document's outbound links from parse time.
	counterpart string
	opens       func(name string) string

	// showsPath prints where a document lives, for the one part whose
	// location a reader cannot derive from the heading.
	showsPath bool
}

func parts() []part {
	return []part{{
		heading:     "PLATFORM - what the thing is (asgard-cli wiki <page>)",
		search:      wiki.Search,
		list:        wiki.List,
		read:        func(n string) string { return "asgard-cli wiki " + n },
		counterpart: "usecase",
		opens:       func(n string) string { return "-> field level: asgard-cli usecase " + n },
	}, {
		heading:     "SHAPES - how it is assembled (asgard-cli usecase <name>)",
		search:      usecase.Search,
		list:        usecase.List,
		read:        func(n string) string { return "asgard-cli usecase " + n },
		counterpart: "wiki",
		opens:       func(n string) string { return "-> what it is:  asgard-cli wiki " + n },
	}, {
		heading: "DECISIONS - what to weigh (asgard-cli guide <name>)",
		search:  stage.Search,
		list:    stage.Docs,
		read:    func(n string) string { return "asgard-cli guide " + n },
	}, {
		heading:   "SKILLS - what the agent in the customer repo loads (.agents/skills/)",
		search:    scaffold.Search,
		list:      scaffold.List,
		read:      scaffold.Path,
		showsPath: true,
	}}
}

// results are what one query found across every part of the corpus.
type results struct {
	parts []part
	hits  [][]kb.Match
}

func (r results) empty() bool {
	for _, h := range r.hits {
		if len(h) > 0 {
			return false
		}
	}
	return true
}

// drop removes one document from the results, for a hit already shown in a
// better form above them. It is matched on the command that reads the document,
// because that is the one identifier a part and a name share.
func (r *results) drop(readCmd string) {
	for i, p := range r.parts {
		kept := r.hits[i][:0]
		for _, m := range r.hits[i] {
			if p.read(m.Name) != readCmd {
				kept = append(kept, m)
			}
		}
		r.hits[i] = kept
	}
}

// broadResult is where a result stops narrowing anything, and broadQuery is
// how long a query has to be for that to be the query's fault rather than the
// subject being everywhere.
const (
	broadResult = 12
	broadQuery  = 4
)

// count is how many documents matched across every part.
func (r results) count() int {
	n := 0
	for _, h := range r.hits {
		n += len(h)
	}
	return n
}

func (r results) terms() []string {
	var out []string
	for _, h := range r.hits {
		for _, m := range h {
			out = append(out, m.Terms...)
		}
	}
	return out
}

// searchAll runs one query against every part of the corpus.
func searchAll(query string) (results, error) {
	r := results{parts: parts()}
	for _, p := range r.parts {
		hits, err := p.search(query)
		if err != nil {
			return r, err
		}
		r.hits = append(r.hits, hits)
	}
	return r, nil
}

func printResults(out io.Writer, query string, r results) {
	// Say which words carried the result before showing it. A partial match
	// reads exactly like a whole one, and a reader who does not know that
	// "shopee" found nothing will take what came back as the material on the
	// subject they asked about.
	reportTerms(out, query, r.terms())

	for i, p := range r.parts {
		hits := r.hits[i]
		if len(hits) == 0 {
			continue
		}
		fmt.Fprintf(out, "%s\n\n", p.heading)
		for _, m := range hits {
			fmt.Fprintf(out, "  %-18s %s\n", m.Name, truncate(m.Title, 84))
			for _, line := range m.Lines {
				fmt.Fprintf(out, "  %-18s %s\n", "", truncate(line, 84))
			}
			if p.counterpart != "" {
				if to := m.Counterpart(p.counterpart); to != "" {
					fmt.Fprintf(out, "  %-18s %s\n", "", p.opens(to))
				}
			}
			// Only the skills need this line: the other three headings name
			// the command in full, and a skill's path is not derivable from
			// its name by a reader who has not seen the layout.
			if p.showsPath {
				fmt.Fprintf(out, "  %-18s -> read it:     %s\n", "", p.read(m.Name))
			}
			fmt.Fprintln(out)
		}
	}

	// **A query that matched almost everything narrowed nothing**, and the
	// ranking cannot fix that: a question asked in a sentence - "which
	// questions to ask the customer" - is six common words, every one of them
	// in most documents, so the section that answers it can be fifteen rows
	// down while every term reports as landed. Saying so beats leaving a
	// scanner to stop before it.
	// Gated on the query being a sentence, not on the count alone: "semantic
	// layer" reaches 34 documents and is a perfectly good two-word lookup.
	// What does not narrow is four or more words, most of them common.
	if n := r.count(); n > broadResult && len(kb.Terms(query)) >= broadQuery {
		fmt.Fprintf(out, "**%d documents matched, which means this query narrowed almost nothing.**\n"+
			"Every term landed, so nothing above says otherwise - a question asked in a\n"+
			"sentence is mostly common words. One or two distinctive nouns work better\n"+
			"here, and the section that answers a question about **what to ask** is\n"+
			"DECISIONS rather than PLATFORM.\n\n", n)
	}

	// The reading order is the same whichever half you landed in.
	fmt.Fprintf(out, "Read the platform side first; an extract assumes you have.\n")
}

// reportRouted says when a name in the query reached the shape it belongs to
// rather than material about the name itself.
//
// **Without this the two are indistinguishable.** A query naming a payment
// gateway is rewritten into the terms for a write path and comes back with
// `usecase write-path` and `usecase external-api` - which is the right material
// and is not material about that gateway. Nothing on the screen says so, and an
// agent reads a confident set of results as an answer about the product it
// asked about. That is the same failure as taking the wrong sense of a word,
// arriving by a different door.
//
// A name in the covered table says nothing here: somebody searched the
// deployments for it, the answer is on the page the row points at, and the
// results are about the thing that was asked.
func reportRouted(out io.Writer, query string) {
	entities := wiki.EntityRows()
	seen := map[string]bool{}
	var routed []string
	for _, term := range kb.Terms(query) {
		if e, ok := entityFor(entities, term); ok && !e.Covered && !seen[e.Word] {
			seen[e.Word] = true
			routed = append(routed, e.Word)
		}
	}
	if len(routed) == 0 {
		return
	}

	fmt.Fprintf(out, "**Nothing here names %s.** What follows is the shape it belongs to, which\n"+
		"is what this material has - not material about the product. Nobody has\n"+
		"searched the reference deployments for it, and until somebody does, the\n"+
		"answer to \"do we already integrate it\" is not in this tool.\n\n"+
		"What is still the customer's to answer is which of four shapes it gives us,\n"+
		"and that changes what gets built:\n\n"+
		"  asgard-cli wiki taiwan-channels   the four, and what each one costs\n"+
		"  asgard-cli question add \"which of the four shapes does %s give us\" \\\n"+
		"      --ask \"<who at the customer>\"\n\n"+
		"**Before writing a question down, put it through the interview's own test:\n"+
		"imagine the most specific answer possible, then ask what you would do\n"+
		"differently.** A perfect answer that changes nothing is not a question -\n"+
		"`asgard-cli guide requirements` has the test and the shape of one that\n"+
		"works.\n\n"+
		"**And this is a gap worth filing**, because the next engagement asks the\n"+
		"same thing and gets the same answer. What to write is\n"+
		"`asgard-cli issue-report --help`; `asgard-cli issue-report --new` writes it\n"+
		"with this search already in it:\n\n"+
		"  https://github.com/%s/issues/new\n"+
		"  gh issue create --repo %s\n\n",
		strings.Join(routed, ", "), routed[0], issueRepo, issueRepo)
}

// reportSenses warns when a query used a word this material has taken.
//
// **It fires on a search that succeeded**, which is the point. A search that
// finds nothing is recorded, and the reader is told so; a search that finds the
// wrong sense of a word looks exactly like an answer. `payment` here is billing
// between Asgard and the customer - `find payment` returns Fehu - and an agent
// asked how to integrate a customer's payment gateway gets that, reads it as
// responsive, and nothing anywhere is red.
//
// Capped, because nine of these words are common query terms and a wall of
// definitions above every result is how a warning gets skipped.
func reportSenses(out io.Writer, query, searched string) bool {
	terms := append(kb.Terms(query), kb.Terms(searched)...)
	seen := map[string]bool{}
	var hits []wiki.Sense
	for _, s := range wiki.Senses() {
		if seen[s.Word] {
			continue
		}
		for _, t := range terms {
			if t == s.Word {
				seen[s.Word] = true
				hits = append(hits, s)
				break
			}
		}
	}
	if len(hits) == 0 {
		return false
	}
	if len(hits) > 3 {
		hits = hits[:3]
	}

	fmt.Fprintf(out, "These results use a word that means one thing here, and it may not be\nthe one that was asked about:\n\n")
	// Wrapped rather than truncated: the "is not" column is where the pointer
	// to the right material lives, and it is the half that was being cut off.
	for _, h := range hits {
		fmt.Fprintf(out, "  %-12s is     %s\n", h.Word, wrapAt(h.Means, 56, 22))
		fmt.Fprintf(out, "  %-12s is not %s\n\n", "", wrapAt(h.Not, 56, 22))
	}
	fmt.Fprintf(out, "\n`asgard-cli wiki glossary` has the rest. **A result in the wrong sense reads\nexactly like an answer**, and nothing here can tell them apart.\n\n")
	return true
}

// unmatched returns the query terms that appear in none of the results.
func unmatched(query string, matched []string) []string {
	found := map[string]bool{}
	for _, t := range matched {
		found[t] = true
	}
	var out []string
	for _, t := range kb.Terms(query) {
		if !found[t] {
			out = append(out, t)
		}
	}
	return out
}

// recordMiss writes a search the material did not answer, when the search was
// run inside an engagement.
//
// `find` works with no repository at all - that is the point of it, the question
// gets asked before there is a directory - so this finds one if there is one and
// does nothing if there is not. What it writes goes to work.MissLog, which is
// not committed: see the contract there.
func recordMiss(kind, query string) {
	root := repo.Root(".")
	if root == "" {
		return
	}
	work.Miss(root, kind, query)
}

// reportTerms names the query terms that appear nowhere in what was returned.
func reportTerms(out io.Writer, query string, matched []string) {
	terms := kb.Terms(query)
	if len(terms) < 2 {
		return
	}

	found := map[string]bool{}
	for _, t := range matched {
		found[t] = true
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

	// A word this material knows under another name is the common case, and it
	// used to be reported only when the whole query missed. A query with one
	// good term and one untranslated one therefore got results, a "matched 1 of
	// 2" line, and no way to find out that the missing half had a name here.
	printAliases(out, missing)
}

// printAliases names the words this material uses for the terms that missed.
func printAliases(out io.Writer, missing []string) {
	hints := aliasesFor(missing)
	if len(hints) == 0 {
		return
	}
	fmt.Fprintf(out, "Of those, this material has a name for:\n\n")
	for _, h := range hints {
		fmt.Fprintf(out, "  %s\n", h)
	}
	fmt.Fprintln(out)
}

// aliasesFor returns "<word>  <the words this material uses>" for each term
// that has a row in the index.
//
// The tables used to be a map in this file, and then a section of the glossary
// page. They are neither: an index is bookkeeping, and it belongs beside the
// corpus rather than inside it. See wiki.Aliases.
func aliasesFor(terms []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, term := range terms {
		for word, means := range wiki.Aliases() {
			if strings.Contains(term, word) && !seen[word] {
				seen[word] = true
				out = append(out, fmt.Sprintf("%-8s %s", word, means))
			}
		}
		for _, e := range wiki.EntityRows() {
			if strings.Contains(term, e.Word) && !seen[e.Word] {
				seen[e.Word] = true
				out = append(out, fmt.Sprintf("%-8s %s", e.Word, e.Search))
			}
		}
	}
	sort.Strings(out)
	return out
}

// entityFor returns the entity row covering this term.
func entityFor(entities []wiki.Entity, term string) (wiki.Entity, bool) {
	for _, e := range entities {
		if strings.Contains(term, e.Word) {
			return e, true
		}
	}
	return wiki.Entity{}, false
}

// lookup returns the index row covering this term, or "".
func lookup(table map[string]string, term string) string {
	for word, means := range table {
		if strings.Contains(term, word) {
			return means
		}
	}
	return ""
}

// suggest returns the alias hints for every term in a query.
func suggest(query string) []string { return aliasesFor(kb.Terms(query)) }

// translate rewrites a query into the words this material uses, before the
// search runs rather than after it has failed.
//
// **The two tables behave differently and that is the point.** An alias
// replaces the word: 電商 appears nowhere in an English corpus, so keeping it
// would only add a term that lands nowhere and report "matched 3 of 4". An
// entity is added to the query instead, because the name may be written
// verbatim in a page - SHOPLINE is, in two skills - and replacing it with
// "commerce, marketplace, channel" would throw away the best answer there is.
//
// Only the bare terms of a row are searched for: the part before any " - "
// aside is a pointer for a reader, not something to match on.
func translate(query string) string {
	aliases, entities := wiki.Aliases(), wiki.EntityRows()
	if len(aliases) == 0 && len(entities) == 0 {
		return ""
	}

	var out []string
	var changed bool
	seen := map[string]bool{}
	add := func(w string) {
		if w = strings.TrimSpace(w); w != "" && !seen[strings.ToLower(w)] {
			seen[strings.ToLower(w)] = true
			out = append(out, w)
		}
	}
	addRow := func(row string) {
		bare, _, _ := strings.Cut(row, " - ")
		for _, w := range strings.Split(bare, ",") {
			add(w)
		}
	}

	for _, term := range kb.Terms(query) {
		if means := lookup(aliases, term); means != "" {
			changed = true
			addRow(means)
			continue
		}
		// The term itself first, so a page naming it outranks the material
		// its category reaches.
		add(term)
		if e, ok := entityFor(entities, term); ok {
			changed = true
			addRow(e.Search)
		}
	}
	if !changed {
		return ""
	}
	return strings.Join(out, " ")
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
		"shape reaches it, a device protocol is a question about the sandbox.\n\n", query)

	if hints := suggest(query); len(hints) > 0 {
		fmt.Fprintf(out, "Terms in that query, in the words this material uses:\n\n")
		for _, h := range hints {
			fmt.Fprintf(out, "  %s\n", h)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "**If that was a system the customer named, this material never had it.**\n"+
		"It describes platform parts and deployment shapes, not a catalogue of other\n"+
		"people's products - and what decides the work is not which product it is.\n"+
		"It is which of four shapes the thing presents to us, and those differ in\n"+
		"cost by more than an order of magnitude:\n\n"+
		"  asgard-cli wiki taiwan-channels   the four, and what each one costs\n\n"+
		"**That is a question for the customer, and not one this tool can answer.**\n"+
		"The answer changes what gets built, so it belongs where somebody will\n"+
		"answer it rather than in a decision made here:\n\n"+
		"  asgard-cli question add \"which of the four shapes does <it> give us\" \\\n"+
		"      --ask \"<who at the customer>\"\n\n"+
		"**Before writing that question down, put it through the test the interview\n"+
		"uses**, because this is the one place in this tool that hands you a\n"+
		"`question add` unprompted and the reader of it is, by definition, not in\n"+
		"the interview: **imagine the most specific answer possible, then ask what\n"+
		"you would do differently.** \"Ming issues it\" and \"Ming spends five hours a\n"+
		"day on it\" are both perfect answers and neither changes anything, so\n"+
		"neither is a question. What we need FROM them is ours to chase - a\n"+
		"credential, an endpoint, a network path, a document, an account. How they\n"+
		"staff a channel is theirs.\n\n"+
		"  asgard-cli guide requirements   the test in full, and the shape of a\n"+
		"                                  question that works\n\n")

	fmt.Fprintf(out, "**This search was recorded.** If the subject does exist here under another\n"+
		"name, that is a missing row in the index rather than a missing page, and\n"+
		"the two need different repairs:\n\n"+
		"  asgard-cli wiki --aliases      the index, and the rule for adding a row\n"+
		"  asgard-cli reading --misses    every search here that came back empty\n"+
		"  asgard-cli issue-report --new  a report with this evidence already in it\n\n")

	// The whole catalogue, last. It is here rather than replaced by advice
	// about phrasing because an agent that searches twice, gets nothing twice
	// and falls back on what it already believed has stopped using this
	// material - so the dead end hands over the contents. What goes above it is
	// the part that is actionable without reading 68 titles.
	fmt.Fprintf(out, "Everything there is, in full:\n\n")
	for _, p := range parts() {
		docs, err := p.list()
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n\n", p.heading)
		for _, d := range docs {
			fmt.Fprintf(out, "  %-24s %s\n", d.Name, truncate(d.Title, 78))
		}
		fmt.Fprintln(out)
	}
	return nil
}
