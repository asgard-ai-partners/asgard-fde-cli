// Package selfsrc carries this repository's own Go source, so the binary can
// audit the commands it prints.
//
// **A string this tool prints makes the same claim a document does**: that
// this build answers to what it names. `audit-material --commands` resolves
// the material and the scaffold templates against the command tree, and until
// this existed it could not resolve the tool's own output - so a renamed
// command left dead names in the help, in `check`'s findings and in the
// generator's warnings, and the only thing that ever caught one was the check
// written into a customer's repository, an engagement later.
//
// It lives at the module root because `go:embed` only reaches downward, and
// the packages that print are spread across `internal/`. The pattern is
// `*.go` at one level deep, which is every package here and no material - the
// corpus must not be embedded twice.
package selfsrc

import "embed"

//go:embed internal/*/*.go
var Go embed.FS

// Docs is this repository's own documentation - Goal, README, AGENTS,
// STRUCTURE, APPROACH, TASK.
//
// **They are read by an agent working in this repository**, which is an
// audience, and until this existed nothing held them to the standard the
// shipped material is held to. STRUCTURE.md was sending a reader to
// `asgard-cli scaffold` and README.md documented a command and a log file
// that had both been deleted, while `--commands` reported 0 dead - because
// these are read from a checkout rather than shipped, and the audit reads
// what ships.
//
// They are still not material: they do not land anywhere, so `--paths` and
// the provenance rules do not apply to them. What applies is that a command
// they name has to exist.
//
//go:embed *.md
var Docs embed.FS
