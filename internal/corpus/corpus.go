// Package corpus holds the material itself, in the shape a repository receives
// it.
//
// **The layout is the whole reason this package exists.** The wiki and the
// extracts used to live under their own packages, at `internal/wiki/pages/` and
// `internal/usecase/extracts/`, and `asgard-cli init` writes them into a
// customer repository as `wiki/` and `usecase/` side by side. Those two trees
// disagree about what a relative path means:
//
//	from a wiki page to that extract
//	  the old layout   ../../usecase/extracts/write-path.md
//	  as landed        ../usecase/write-path.md
//
// So a pointer written as a path could be right in one tree and wrong in the
// other, and that is why every pointer in this material is an invocation -
// `asgard-cli usecase write-path` - which means the same thing from anywhere.
// Layout independence was bought with a subprocess on every hop.
//
// One tree, matching the landed one, is what lets a pointer become a path that
// resolves in both. That is the point of the move and nothing else: no document
// changed, and go:embed cannot reach across a package, so the files have to sit
// under whichever package embeds them.
//
// `internal/wiki` and `internal/usecase` are still the way in. They hold the
// two `kb.Corpus` values and everything that reads them; this package holds
// only the bytes.
package corpus

import "embed"

// FS is the material: `wiki/` and `usecase/` side by side, with the alias index
// at the root because it applies to both.
//
// `aliases.md` is named rather than globbed, and the two directories are not
// `all:`, so nothing arrives here by being dropped in the folder - a file joins
// the corpus by being a document in one of the two halves.
//
//go:embed wiki usecase aliases.md
var FS embed.FS
