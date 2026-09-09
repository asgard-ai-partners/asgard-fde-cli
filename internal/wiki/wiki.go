// Package wiki serves the platform wiki: what Asgard is made of and who each
// piece is for, distilled from the product documentation and the CRD contract.
//
// It answers a different question from internal/usecase. An extract there says
// how one shape of deployment is assembled, field by field, and assumes you
// already know the platform has that shape. These pages are what an agent needs
// before that: it has the customer's repository and nothing else, and that
// repository describes one customer's systems, never the platform they run on.
//
// **The pages themselves are in internal/corpus**, beside the extracts and in
// the layout a repository receives them - that package's doc comment says why.
// `asgard-cli init` writes both halves into a customer repository, and
// `scaffold.replaceCorpus` replaces them when the binary's version moves.
//
// The reading and searching are internal/kb's; this package is the way in.
package wiki

import (
	"sort"
	"strings"

	corpusfs "github.com/asgard-ai-partners/asgard-fde-cli/internal/corpus"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
)

// Page is one wiki page. It is kb.Doc under a name that reads at the call site.
type Page = kb.Doc

// Match is one page that matched a search, with the lines that matched.
type Match = kb.Match

var corpus = kb.Corpus{
	FS:  corpusfs.FS,
	Dir: "wiki",
	// The index, the log and the conventions are the wiki's own bookkeeping
	// rather than pages about the platform. Readable by name, absent from a
	// listing. README joined them when the pages moved under `wiki/`: it used
	// to sit outside the corpus directory, so nothing had to exclude it.
	Unlisted: map[string]bool{"index": true, "log": true, "README": true},
	Noun:     "wiki page",
	Command:  "asgard-cli wiki",
}

// List returns every page, sorted by name.
func List() ([]Page, error) { return corpus.List() }

// All returns every page including the index and the log, for writing the wiki
// out rather than listing it.
func All() ([]Page, error) { return corpus.All() }

// Read returns one page in full.
func Read(name string) (string, error) { return corpus.Read(name) }

// Conventions returns the wiki's own README: the three layers, the three
// operations, and the rules a page has to follow.
func Conventions() (string, error) { return corpus.File("wiki/README.md") }

// Search finds pages mentioning all of the given terms.
func Search(query string) ([]Match, error) { return corpus.Search(query) }

// The two tables in aliases.md. Matched on the heading rather than on position,
// so the file can be reordered.
const (
	aliasFile      = "aliases.md"
	aliasHeading   = "## What a customer says, in the words this material uses"
	coveredHeading = "## Names the material covers"
	routedHeading  = "## Names it only routes"
)

// Entity is a name a customer will say - a product, a marketplace, a payment
// gateway - and the terms that reach the material about it.
type Entity struct {
	Word   string
	Search string

	// Covered is true when somebody searched the reference deployments for
	// this name and recorded what came back, so the row routes to a page that
	// answers it. False means the row routes to the **shape** the thing
	// belongs to and nothing here names the thing itself.
	//
	// **The distinction is the whole reason there are two tables.** A row that
	// routes reads exactly like a row that answers, and a reader who cannot
	// tell them apart takes results about a shape as results about a product -
	// which is the same failure as taking the wrong sense of a word, arriving
	// by a different door.
	Covered bool
}

// EntityRows returns every name a customer will say, from both tables.
//
// These are **added** to a query rather than replacing it: the name may be
// written verbatim in a page - SHOPLINE is, in two skills - and replacing it
// with its category would throw away the best answer there is.
func EntityRows() []Entity {
	var out []Entity
	for word, search := range table(coveredHeading) {
		out = append(out, Entity{Word: word, Search: search, Covered: true})
	}
	for word, search := range table(routedHeading) {
		out = append(out, Entity{Word: word, Search: search})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Word < out[j].Word })
	return out
}

// Sense is a word that means one thing in this material, and what the other
// sense is called instead.
type Sense struct {
	Word  string
	Means string
	Not   string
}

// Senses returns the words with one meaning here, from the glossary's first
// table.
//
// **This is applied to a query, not only read by a person.** The table existed
// for two years as prose on a page that nothing pointed at, and the failure it
// describes went on happening: `find payment` returns Fehu, which is billing
// between Asgard and the customer, to somebody asking about the customer's own
// payment gateway. Nothing was wrong with the result and nothing was recorded -
// a search that lands is not a miss - so the one mechanism that could have
// caught it, the miss log, is blind to exactly this case.
//
// A missing or renamed table returns nothing rather than an error, for the same
// reason Aliases does: this decorates a search, it does not gate one.
func Senses() []Sense {
	body, err := Read("glossary")
	if err != nil {
		return nil
	}
	// The first table only: what follows the next heading is prose about two
	// words, and the customer-vocabulary index has moved out entirely.
	if next := strings.Index(body, "\n## "); next >= 0 {
		body = body[:next]
	}

	var out []Sense
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cols := strings.Split(strings.Trim(line, "|"), "|")
		if len(cols) != 3 {
			continue
		}
		word := strings.Trim(strings.TrimSpace(cols[0]), "*")
		means := strings.TrimSpace(cols[1])
		not := strings.TrimSpace(cols[2])
		if word == "" || word == "word" || strings.HasPrefix(word, "-") {
			continue
		}
		out = append(out, Sense{Word: word, Means: means, Not: not})
	}
	return out
}

// Index returns the alias index in full, for a reader.
func Index() (string, error) { return corpus.File(aliasFile) }

// Aliases maps a word a customer used to the words this material uses. These
// **replace** the word in a query: a Chinese term appears nowhere in an English
// corpus, so keeping it would only add a term that lands nowhere.
//
// The tables live in a file rather than in Go, because a term's other name is
// knowledge rather than configuration: it has to be readable by somebody who
// never runs the search, and `audit-material --links` resolves the pointers the
// rows carry, which it could not do inside a string constant.
//
// They live **beside** `pages/` rather than in it - the same place `index.md`
// and `log.md` sit - because an index inside the searched corpus competes with
// what it indexes. It lists every alias, so it was unusually likely to be the
// one document carrying every term of a translated query, and `find 電商`
// returned the word list rather than `taiwan-channels`.
func Aliases() map[string]string { return table(aliasHeading) }

// table reads one two-column table out of the index file.
//
// A missing or renamed section returns nothing rather than an error. This
// decorates a search; it does not gate one, and a search that fails because a
// heading moved would be worse than one that gives no hint.
func table(heading string) map[string]string {
	body, err := corpus.File(aliasFile)
	if err != nil {
		return nil
	}
	_, rest, found := strings.Cut(body, heading)
	if !found {
		return nil
	}
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}

	out := map[string]string{}
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cols := strings.Split(strings.Trim(line, "|"), "|")
		if len(cols) != 2 {
			continue
		}
		word := strings.TrimSpace(cols[0])
		means := strings.TrimSpace(cols[1])
		// Skips the header row and the |---|---| separator without having to
		// count lines: neither has a word in the first column that is a word.
		if word == "" || means == "" || word == "they said" || strings.HasPrefix(word, "-") {
			continue
		}
		out[word] = means
	}
	return out
}

// Landing returns the pages that belong in a copy written into a customer
// repository: every page, plus the index, and not the log.
//
// **The two unlisted documents are not alike**, which List cannot express and
// All does not either. The index is the map and a copy without it has none.
// The log is provenance - which commit of each source this corpus was read at -
// and this page's own index says an FDE looking for an answer should never land
// there. Writing it into a customer repository also carried its historical
// entries, which name commands the tool has since removed and are correct to,
// into a check that reads them as a repository pointing at a command that does
// not exist.
func Landing() ([]Page, error) {
	all, err := corpus.All()
	if err != nil {
		return nil, err
	}
	out := make([]Page, 0, len(all))
	for _, p := range all {
		if p.Name == "log" {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}
