// Package usecase serves the extracts: how each shape of Asgard deployment is
// built, taken from ones already in production.
//
// An agent about to author a chart reaches for these. They are embedded rather
// than written into a customer repo, because an example nobody is using becomes
// stale boilerplate there, while a stale one here is fixed for every engagement
// in a single release.
//
// The reading and searching are internal/kb's; what lives here is the corpus.
package usecase

import (
	"embed"

	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
)

//go:embed extracts
var extracts embed.FS

// Extract is one shape. It is kb.Doc under a name that reads at the call site.
type Extract = kb.Doc

// Match is one extract that matched a search, with the lines that matched.
type Match = kb.Match

var corpus = kb.Corpus{
	FS:  extracts,
	Dir: "extracts",
	// README explains how to read an extract; it is not one.
	Unlisted: map[string]bool{"README": true},
	Noun:     "extract",
	Command:  "asgard-cli usecase",
}

// List returns every extract, sorted by name.
func List() ([]Extract, error) { return corpus.List() }

// All returns every extract including the README that explains how to read
// one, for writing the extracts out rather than listing them.
func All() ([]Extract, error) { return corpus.All() }

// Read returns one extract in full.
func Read(name string) (string, error) { return corpus.Read(name) }

// Index returns the README, which explains how the extracts are organised.
func Index() (string, error) { return corpus.File("extracts/README.md") }

// Search finds extracts mentioning all of the given terms.
func Search(query string) ([]Match, error) { return corpus.Search(query) }
