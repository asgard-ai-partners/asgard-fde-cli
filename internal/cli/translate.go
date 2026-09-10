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
