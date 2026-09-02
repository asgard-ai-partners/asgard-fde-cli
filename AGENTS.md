# AGENTS.md

Rules for AI agents working in this repo. Human contributors follow the same ones.

## What this repo is

A single Go binary, `asgard-cli`, that drives the onboarding of one customer onto
the Asgard platform. It writes files into a *customer's* repository and never
holds state of its own.

Most of the value is not code. It is four bodies of embedded material, and the
first question when adding anything is which one it belongs to:

[STRUCTURE.md](STRUCTURE.md) walks every directory. The short version:

| material | answers | lives in |
|---|---|---|
| stage prompts | what to do at this point in an onboarding | `internal/stage/prompts/` |
| wiki pages | what the platform is, and who each piece is for | `internal/wiki/pages/` |
| usecase extracts | how one shape of deployment is assembled, field by field | `internal/usecase/extracts/` |
| scaffold templates | the part of a customer repo that is the same every time | `internal/scaffold/templates/` |

**One fact, one home; everywhere else links.** A trap that belongs to a CR
template does not also get explained in a wiki page. When you are about to repeat
a paragraph, link instead - two copies drift, and the reader cannot tell which is
current.

The reading order is wiki, then extract: an extract assumes you already know the
platform has that shape. `asgard-cli add <kind>` prints both, in that order.

## Language

**Go code, command output, error messages, command help, and this repo's own
English documents are English.** That is `README.md`, `TASK.md`, this file, and
`source/SOURCES.md`.

```go
// Save writes cfg to path.
return fmt.Errorf("write %s: %w", path, err)
```

**One body of material is deliberately Chinese**: parts of
`internal/scaffold/templates/`, because the generated repository is read by the
customer's own engagement, in their language.

`internal/wiki/` used to be the second. It is English now: the sources are zh-TW,
but an agent asked in Chinese queries in English, so the corpus does not need to
carry both - and one language removes the split where a Chinese question reached
only the wiki and an English one only the extracts. Product labels keep their own
names (Managed Agent, Drive, Context Index are what the UI says).

Do not "fix" the scaffold templates into English. Do not start a second exception
without saying why it earns one.

## Plain ASCII, no emoji

No emoji, and no decorative Unicode either - no check marks, arrows, box-drawing
characters. They render inconsistently across terminals, get mangled in logs and
CI output, and are awkward to grep for. This applies to the Chinese material too:
Chinese text is fine, `->` beats an arrow glyph.

Mark status with words or ASCII punctuation:

```
ok  .asgard-config.json
- workspace.id must not be empty
```

Some scaffold templates and extracts still carry decorative characters from
before this rule. Fix one when you are editing the file for another reason; do
not sweep them all as a change of its own.

## Every command must document itself

A command that cannot explain itself is not finished:

- `Short` is a one-line summary, shown in the parent's command list.
- `Long` explains what the command does, its defaults, and how it fails. It must
  say more than `Short` repeats. Include an example when the usage is not obvious.
- Every flag needs usage text, and says there if it is required or what it
  defaults to.
- **Never register a flag the code does not read.** An option that silently does
  nothing is worse than no option at all - and one kind reading `--layer` while
  its own help said `--layers` shipped exactly that way.

cobra gives every command a `--help` for free but does not stop it being empty,
and nothing here checks. Read the `Long` of a neighbouring command and match it.

To add a subcommand: write `newXxxCmd()` in `internal/cli/`, register it in the
`cmd.AddCommand(...)` call in `root.go`.

## The gate

There is no test suite - it was removed on 2026-09-02. What is left:

```bash
go build ./...
go vet ./...
gofmt -l internal/ cmd/
```

None of those sees a wrong string in an embedded template, a stage prompt that
sends a reader to a page that does not exist, or a generated CR the apiserver
would reject. **So exercise the change by hand**, against a scratch repository
outside this one:

```bash
go build -o .out/asgard-cli ./cmd/asgard-cli
cd $(mktemp -d)
/path/to/.out/asgard-cli init --workspace-id 0 && /path/to/.out/asgard-cli scaffold
/path/to/.out/asgard-cli project add app --env dev
/path/to/.out/asgard-cli add <kind> <name> --project app
/path/to/.out/asgard-cli check && /path/to/.out/asgard-cli render app dev
```

If the change touched a CR template, validate the rendered output against the
CRD schemas in [`asgard-kube/crd/`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) before believing it.
`helm lint` and `asgard-cli check` do not compare against the platform contract,
which is how four required fields were missing from two templates until somebody
looked.

## Reference material lives outside this repo

Ten repositories, read-only, never vendored in. **The URL is the source of
truth; where you happen to clone it is not** - a local path is true on one
machine and wrong on every other:

**The contract:**

| what | source of truth |
|---|---|
| CRD definitions, the platform contract | https://github.com/asgard-ai-platform/asgard-kube |
| product documentation | https://github.com/asgard-ai-platform/asgard-docs |

**The reference deployments** - every extract under `internal/usecase/extracts/`
was taken from one of these, and a claim about how a shape is built should be
checkable against at least one:

| repo | shape it demonstrates |
|---|---|
| [unitech-e-asgard-kube](https://github.com/asgard-ai-platform/unitech-e-asgard-kube) | agent hub (5 agents / 6 semantic layers) and a single-agent flow agent; trigger; knowledge drive. The most recently maintained of the set, so it wins a generational conflict |
| [xxentria-asgard-kube](https://github.com/asgard-ai-platform/xxentria-asgard-kube) | supervisor + 9 agents |
| [finance-ai-asgard-kube](https://github.com/asgard-ai-platform/finance-ai-asgard-kube) | supervisor + 3 agents over 3 semantic layers |
| [buy123-asgard-kube](https://github.com/asgard-ai-platform/buy123-asgard-kube) | the minimal flow agent - **no Agent CR at all**, 8 CRs in the whole chart |
| [asgard-freyr-kube](https://github.com/asgard-ai-platform/asgard-freyr-kube) | supervisor + 5 subagents, `agents.expression`, sandbox hooks, a `tenants/` layout |
| [asgard-auto-post-kube](https://github.com/asgard-ai-platform/asgard-auto-post-kube) | **28 Plugin CRs**, knowledge bases, api workflows; 3 BotProviders |
| [asgard-industry-demo-generator](https://github.com/asgard-ai-platform/asgard-industry-demo-generator) | 12 industries, the read/write governance split, a Claude Code plugin of commands + skills |
| [asgard-freyr-skills](https://github.com/asgard-ai-platform/asgard-freyr-skills) | runtime skills as a repository of their own - the only source for browser operation, and what `browser-operation` was written from |

Which customer each belongs to, and how the extracts refer to one without naming
it, is in `source/SOURCES.md` - the one file here allowed to make that link, and
the reason it sits outside `internal/`.

Clone them wherever you like. Pull before relying on any of them, and **record
the commit you read** in whatever you write - a copy taken into this repo stops
tracking upstream and then reads exactly like a current one.

## Put generated files in `.out/`

Anything a command produces that does not belong in version control goes to
`.out/` at the repo root - not scattered, and not in `/tmp`:

- throwaway debug scripts and scratch programs
- command output, logs, sample data downloaded for inspection
- binaries built by hand
- profiling reports (`*.pprof`)

```bash
mkdir -p .out
go build -o .out/asgard-cli ./cmd/asgard-cli
```

`.out/` is gitignored and can be deleted at any time: `rm -rf .out`.
