// Package usecase serves the extracts: how each shape of Asgard deployment is
// built, taken from ones already in production.
//
// An agent about to author a chart reaches for these. **The extracts themselves
// are in internal/corpus**, beside the wiki and in the layout a repository
// receives them - that package's doc comment says why. `asgard-cli init` writes
// both halves into a customer repository.
//
// The reading and searching are internal/kb's; this package is the way in.
package usecase

import (
	corpusfs "github.com/asgard-ai-partners/asgard-fde-cli/internal/corpus"
	"github.com/asgard-ai-partners/asgard-fde-cli/internal/kb"
)

// Extract is one shape. It is kb.Doc under a name that reads at the call site.
type Extract = kb.Doc

// Match is one extract that matched a search, with the lines that matched.
type Match = kb.Match

var corpus = kb.Corpus{
	FS:  corpusfs.FS,
	Dir: "usecase",
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
func Index() (string, error) { return corpus.File("usecase/README.md") }

// Search finds extracts mentioning all of the given terms.
func Search(query string) ([]Match, error) { return corpus.Search(query) }
