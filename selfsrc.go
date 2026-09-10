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
