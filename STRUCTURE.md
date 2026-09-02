# STRUCTURE.md

What every directory in this repo is for. `README.md` is what the commands do,
`AGENTS.md` is the rules for changing them, `TASK.md` is what the repo is for and
what is unfinished. This file is the map.

## The shape in one line

A single Go binary that writes files into **somebody else's** repository. It holds
no state: everything an engagement knows ends up in the customer's repo, because
that repo is what the next agent opens.

```
cmd/asgard-cli/       main; signal handling and exit codes only
internal/             every package, none exported
source/               internal notes that must never ship
.github/             CI, the tag-driven release, and the PR template
.goreleaser.yaml      how the binary is built and published
```

## `internal/` - the code

Two halves. The first five packages are the command surface and the records it
keeps; the rest are the checks and the plumbing.

| package | go | what it holds |
|---|---|---|
| `cli` | 2313 | the cobra command tree, one file per subcommand, plus `root.go`, `repo.go` and `find.go` (which spans two packages rather than serving one) |
| `work` | 1173 | the customer repo's own records: requests, task specs, open questions, decisions |
| `gate` | 1051 | the invariant checks on a rendered chart - xref, agent split, deployability |
| `generate` | 642 | CR skeletons for ten kinds, wired to what the chart already declares |
| `check` | 616 | repository structure: indexes, dated names, links, orphan pages |
| `stage` | 381 | which stage an onboarding is at, derived from the repo and never stored |
| `scaffold` | 360 | writes the non-customer-specific tree |
| `tool` | 324 | resolves helm/kubectl/python3 and says how to install one |
| `config` | 278 | `.asgard-config.json`: the workspace and its projects |
| `wiki` | 171 | serves the platform wiki |
| `usecase` | 162 | serves the deployment-shape extracts |
| `chart` | 117 | reads a project's **unrendered** templates for (kind, name) |
| `render` | 104 | renders via `helm template`, the way CD does |
| `deploy` | 81 | `projects/<p>/deploy.yaml`, the source of truth for deploy targets |
| `version` | 73 | build information, injected by GoReleaser via ldflags |

**To add a subcommand**: write `newXxxCmd()` in `internal/cli/`, register it in the
`cmd.AddCommand(...)` call in `root.go`.

### Why `chart` reads unrendered templates

`generate` has to know what a chart already declares before it writes into it -
whether a SemanticLayer exists to mount, whether a Toolset was already emitted.
That has to work without helm, before values are filled in, so `chart` parses the
templates as text rather than rendering them.

## `internal/` - the embedded material

Most of this repo's value is not code. Four bodies of material are compiled into
the binary, and the first question when adding anything is which one it belongs
to.

| where | files | answers | language |
|---|---|---|---|
| `stage/prompts/` | 12 | what to do at this point in an onboarding | English |
| `wiki/pages/` | 18 | what the platform is, and who each piece is for | English |
| `usecase/extracts/` | 18 | how one shape of deployment is assembled, field by field | English |
| `generate/templates/` | 12 | the CR skeletons `asgard-cli add` writes | English |
| `scaffold/templates/` | 45 | the part of a customer repo that is the same every time | mixed |

They are embedded rather than written into a customer repo because a copy in one
engagement goes stale where nobody is looking, while a stale one here is fixed for
every engagement in a single release.

**One fact, one home; everywhere else links.** A trap belonging to a CR template
is not also explained in a wiki page.

### `stage/prompts/`

One file per stage, numbered in the order they happen. Ten are the numbered walk;
`11-requirements.md` (the customer interview) and `10-idle.md` sit outside it and
are reached with `next --stage`.

The prompts are Go templates with `<< >>` delimiters, rendered against the
repository's state, so a prompt can name the actual projects and requests rather
than placeholders.

### `wiki/pages/` and `usecase/extracts/`

Both are reference material and they answer different questions. The reading order
is wiki first: an extract assumes you already know the platform has that shape.

`wiki/README.md` is the wiki's own schema - its three layers, what a page must
carry, and how it is kept from going stale. `wiki/pages/index.md` and `log.md` are
its bookkeeping rather than pages about the platform, so they are readable by name
but not listed.

Every wiki page ends with two things: a source block linking the rendered page on
docs.asgard-ai.com plus the commit it was read at, and an `**Unchecked:**` line
saying which parts were never held against a deployment. The first is the only
thing that makes "upstream moved and this page did not" detectable; the second
stops a reader assuming the wiki is checked as deeply as the extracts, which by
its nature it is not.

### `generate/templates/` and `scaffold/templates/`

Both use `<< >>` delimiters so that Helm's own `{{ }}` survives into the output.

`scaffold/templates/` mirrors the generated repo one-for-one. Path segments in
capitals are placeholders expanded at write time:

```
projects/__PROJECT__/chart/values-__ENV__.yaml.tmpl
docs/spec/__SPEC_SLUG__/README.md.tmpl
```

A `.tmpl` suffix means the file is rendered; anything else is copied verbatim.
`.agents/skills/` under it holds the five design-time skills the coding agent in
the customer repo loads.

## `source/` - never ships

`source/SOURCES.md` traces each extract back to the deployment it came from, and
records the **generational conflicts**: two charts that disagree in a way that is
dated rather than a matter of taste.

It lives outside `internal/` deliberately, so `go:embed` cannot reach it even by
accident. Everything under `internal/` ships to every engagement and names no
customer; this file is the one place that does.

## Reference material that is not in this repo

Read-only, never vendored in. **The URL is the source of truth; where you clone it
is not.**

| what | source of truth |
|---|---|
| CRD definitions, the platform contract | https://github.com/asgard-ai-platform/asgard-kube |
| product documentation | https://github.com/asgard-ai-platform/asgard-docs |
| the eight reference deployments | listed with their shapes in `AGENTS.md` |

A copy taken into this repo stops tracking upstream and then reads exactly like a
current one. Record the commit you read instead.

## What is not here

- **No test suite.** Removed on 2026-09-02. The gate is `go build`, `go vet`,
  `gofmt -l`, and exercising the CLI by hand against a scratch repository. See
  "The gate" in `AGENTS.md`.
- **No `.out/` in version control.** It is gitignored and holds anything a command
  produces: hand-built binaries, command output, scratch programs.
- **No customer data anywhere.** Everything under `internal/` is generic; the
  customer's own knowledge lives in the repo the CLI writes, not in the CLI.
