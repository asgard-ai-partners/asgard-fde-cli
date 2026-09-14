// Package selfsrc carries this repository's own Go source, so the binary can
// audit the commands it prints.
//
// **A string this tool prints makes the same claim a document does**: that
// this build answers to what it names. `audit-material --commands` resolves
// the material and the scaffold templates against the command tree; without
// this it cannot resolve the tool's own output, and a renamed command leaves
// dead names in the help, in `check`'s findings and in the generator's
// warnings - none of which any check here reads.
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
// audience. They are read from a checkout rather than shipped, so without
// this the audit - which reads what ships - cannot see them, and a command
// they name can be gone with nothing to say so.
//
// They are still not material: they do not land anywhere, so `--paths` and
// the provenance rules do not apply to them. What applies is that a command
// they name has to exist.
//
//go:embed *.md
var Docs embed.FS

// Skills is this repository's own `.agents/skills/` - the ones an agent
// maintaining this material loads, as distinct from the design-time skills
// under `internal/scaffold/templates/` that land in a customer repository.
//
// Embedded for the same reason as Docs: they name commands and paths, and the
// audits read what is embedded.
//
//go:embed .agents/skills/*/SKILL.md
var Skills embed.FS

// Hack is the maintainer's gate - `hack/`, which is Go for the reason
// AGENTS.md gives: a check that is not compiled is one nobody runs until it is
// wrong.
//
// **Embedded because it is prose an audience reads.** `go run ./hack list`
// prints each check's own `What:` string, and those describe what a check
// covers to the only person who can act on it. Without this they are outside
// every sweep: a check whose behaviour moves leaves its description standing,
// and the description is what the next maintainer believes.
//
//go:embed hack/*.go hack/internal/*/*.go
var Hack embed.FS
