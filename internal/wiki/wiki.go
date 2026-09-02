// Package wiki serves the platform wiki: what Asgard is made of and who each
// piece is for, distilled from the product documentation and the CRD contract.
//
// It answers a different question from internal/usecase. An extract there says
// how one shape of deployment is assembled, field by field, and assumes you
// already know the platform has that shape. These pages are what an agent needs
// before that: it has the customer's repository and nothing else, and that
// repository describes one customer's systems, never the platform they run on.
//
// Embedded rather than written into a customer repo, for the same reason the
// extracts are: a copy in one engagement goes stale where nobody is looking,
// while a stale page here is fixed for every engagement in one release.
//
// The reading and searching are internal/kb's; what lives here is the corpus.
package wiki

import (
	"embed"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
)

//go:embed pages README.md
var content embed.FS

// Page is one wiki page. It is kb.Doc under a name that reads at the call site.
type Page = kb.Doc

// Match is one page that matched a search, with the lines that matched.
type Match = kb.Match

var corpus = kb.Corpus{
	FS:  content,
	Dir: "pages",
	// The index and the log are the wiki's own bookkeeping rather than pages
	// about the platform. Readable by name, absent from a listing.
	Unlisted: map[string]bool{"index": true, "log": true},
	Noun:     "wiki page",
	Command:  "asgard-cli wiki",
}

// List returns every page, sorted by name.
func List() ([]Page, error) { return corpus.List() }

// Read returns one page in full.
func Read(name string) (string, error) { return corpus.Read(name) }

// Conventions returns the wiki's own README: the three layers, the three
// operations, and the rules a page has to follow.
func Conventions() (string, error) { return corpus.File("README.md") }

// Search finds pages mentioning all of the given terms.
func Search(query string) ([]Match, error) { return corpus.Search(query) }
