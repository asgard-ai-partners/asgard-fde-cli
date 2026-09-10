package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/wiki"
)

// Query translation: the customer's words into the words this material uses.
//
// The corpus is English and a customer conversation usually is not, so a term
// taken from what somebody actually said matches nothing - which reads
// identically to a subject the material lacks. `internal/corpus/aliases.md`
// holds two tables: words that REPLACE a query term, and names that are ADDED
// to it. Whatever applies one has to print what it actually searched for,
// because the answer is to a different question from the one asked.

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

// lookup returns the index row covering this term, or "".
func lookup(table map[string]string, term string) string {
	for word, means := range table {
		if strings.Contains(term, word) {
			return means
		}
	}
	return ""
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
		"  .agents/skills/asgard-platform/wiki/taiwan-channels.md   the four, and what each one costs\n"+
		"  asgard-cli question add \"which of the four shapes does %s give us\" \\\n"+
		"      --ask \"<who at the customer>\"\n\n"+
		"**Before writing a question down, put it through the interview's own test:\n"+
		"imagine the most specific answer possible, then ask what you would do\n"+
		"differently.** A perfect answer that changes nothing is not a question -\n"+
		"`.agents/skills/asgard-platform/guide/requirements.md` has the test and the shape of one that\n"+
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
	fmt.Fprintf(out, "\n`.agents/skills/asgard-platform/wiki/glossary.md` has the rest. **A result in the wrong sense reads\nexactly like an answer**, and nothing here can tell them apart.\n\n")
	return true
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
