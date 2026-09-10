// Package corpus holds the material itself, in the shape a repository receives
// it.
//
// **The layout is the whole reason this package exists.** `asgard-cli init`
// writes the wiki and the extracts into a customer repository as `wiki/` and
// `usecase/` side by side, so a pointer written as a path resolves the same
// way here and there:
//
//	from a wiki page to an extract   ../usecase/write-path.md
//
// Held in two separate package directories, the same pointer would need a
// different number of `../` in each tree, and the only form that survived both
// was an invocation - which cost a subprocess on every hop and stopped being
// possible when the reader commands went.
//
// go:embed cannot reach across a package, so the files sit under whichever
// package embeds them: one tree, matching the landed one.
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
