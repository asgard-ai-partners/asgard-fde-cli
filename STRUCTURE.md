# STRUCTURE.md

What every directory in this repo is for. What the tool is for is
[Goal.md](Goal.md), how the main capabilities are implemented is
[APPROACH.md](APPROACH.md), what the commands do is [README.md](README.md), the
rules for changing them are [AGENTS.md](AGENTS.md), and where it stands is
[TASK.md](TASK.md). This file is the map.

## The shape in one line

A single Go binary that answers questions about integrating with Asgard, and
writes chart skeletons into **somebody else's** repository. It holds no state:
everything an engagement knows ends up in the customer's repo, because that repo
is what the next agent opens.

The first half works with no repository at all, and has to keep doing so - the
question gets asked in a meeting, before there is a directory. `Goal.md` states
the goal; this file is the map.

```
CLAUDE.md             @AGENTS.md, so the rules load without being asked for
.agents/skills/       this repo's own maintenance skills, not the ones that ship
cmd/asgard-cli/       main; signal handling and exit codes only
internal/             every package, none exported
source/               internal notes that must never ship
hack/                 this repo's own tooling: the CRD contract check
.github/             CI, the tag-driven release, and the PR template
.goreleaser.yaml      how the binary is built and published
```

## `internal/` - the code

Largest first. The top of the table is the command surface and the records it
keeps; below it are the checks, the material servers and the plumbing. **No
line counts**: they were here and they were wrong within a week of every
change, and a stale number reads as a fact.

| package | what it holds |
|---|---|
| `cli` | the cobra command tree, one file per subcommand, plus `root.go`, `repo.go`, `format.go` and `audit.go` (which spans every corpus rather than serving one) |
| `gate` | the invariant checks on a rendered chart - xref, agent split, deployability, enums, constraints, conditional CEL shapes |
| `scaffold` | writes the non-customer-specific tree, serves the design-time skills inside it, writes the platform corpus under `.agents/skills/asgard-platform/`, and keeps `.asgard-scaffold.json` - the record of which CLI wrote the files this binary ships |
| `auth` | the OAuth 2.0 + PKCE sign-in and the credential store, which is the only file this CLI keeps outside a repository |
| `work` | the customer repo's own records: requests, task specs, open questions and decisions |
| `check` | repository structure: indexes, dated names, links, orphan pages |
| `platform` | the platform API client: workspaces, the whole `/v1/iac` surface, and `/v1/docs` |
| `generate` | CR skeletons for ten kinds, wired to what the chart already declares |
| `localenv` | the local environment file a chart's placeholders are filled from |
| `kb` | one implementation of listing, reading, provenance and the link graph, shared by every corpus |
| `stage` | the onboarding guidance: a static half that lands as files, and the half rendered against this repository |
| `size` | the deployment shapes, counted off production, and what one costs before anything is added |
| `brief` | what one activity gets wrong, addressed by intent rather than by position |
| `tool` | resolves helm/kubectl/python3 and says how to install one |
| `skills` | the platform's fetched reference material, and the record of which version is here. **Not the design-time skills** - those are `scaffold/templates/.agents/skills/`, and the two are different halves that happen to land in the same directory |
| `render` | renders via `helm template`, with the reserved `asgard` block supplied as placeholders |
| `binding` | reads and writes `.asgard-cli.yaml`, the checkout's platform binding |
| `gitrepo` | the checkout's root and its remotes, read and never compared to anything |
| `needs` | what a shape has to be given by the customer, as seven documents written into a repository beside the extracts |
| `repo` | what a customer repository is made of, by looking at it |
| `pipelineconfig` | reads `.asgard-pipeline.yaml`, the deployment declaration |
| `chart` | reads a project's **unrendered** templates for (kind, name) |
| `wiki` | serves the platform wiki, and the two tables in `aliases.md` beside it |
| `version` | build information, injected by GoReleaser via ldflags |
| `browser` | opens a URL, or says it could not |
| `usecase` | serves the deployment-shape extracts |
| `corpus` | the material itself, in the layout a repository receives it: `wiki/` and `usecase/` side by side, so a pointer can become a path that resolves in both trees |
| `selfsrc` (module root) | this repository's own Go source, embedded so the binary can audit the commands it prints. At the root because `go:embed` only reaches downward |

**Two packages this table used to list are gone.** `config` held
`.asgard-config.json` - the workspace slug, the project list and each project's
"shape" - and every field of it failed the same test: *a value belongs in a
config file only when nothing on disk implies it and the platform cannot be
asked.* `deploy` held `projects/<p>/deploy.yaml`, which the Pipeline cut-over
replaced with the one declaration at the repository root.

**Why nothing decides when a chart is finished.** A shape is the one thing the
files cannot say: a SemanticLayer with nothing mounted on it is either a
finished Mimir deliverable or an agent nobody has written yet, and those are
identical on disk. There is nowhere to record which one it is, deliberately:
that record is somebody's note of intent, and subtracting a chart's contents
from a note of intent is sound arithmetic on an input this tool cannot check.
`asgard-cli size` counts the shapes off deployments in production, for a person
to compare against.

**To add a subcommand**: write `newXxxCmd()` in `internal/cli/` and register it
through `addTo(cmd, group..., ...)` in `root.go`. The group is required - cobra
panics on a `GroupID` the parent does not have - so a command cannot be added
without deciding where in the help it belongs.

### Why `chart` reads unrendered templates

`generate` has to know what a chart already declares before it writes into it -
whether a SemanticLayer exists to mount, whether a Toolset was already emitted.
That has to work without helm, before values are filled in, so `chart` parses the
templates as text rather than rendering them.

## `internal/` - the embedded material

Most of this repo's value is not code. It is compiled into the binary, and the
first question when adding anything is which part it belongs to.

**Four of these are one corpus** - grep reaches them together and
they share one schema: a `# ` title, a summary, and `**Checked:**` /
`**Unchecked:**`. `asgard-cli audit-material --unverified` is the check, and it is 0 of 25,
0 of 21, 0 of 12, 0 of 7. The fifth, `generate/templates/`, is not searched: it
is what `add` writes, not something anybody reads to decide.

| where | files | answers | language | searched |
|---|---|---|---|---|
| `corpus/wiki/` | 27 | what the platform is, and who each piece is for | English | yes |
| `corpus/usecase/` | 22 | how one shape of deployment is assembled, field by field | English | yes |
| `stage/prompts/` | 12 | what to weigh at one point in the work | English | yes |
| `scaffold/templates/.agents/skills/` | 7 | what the agent in a customer repo loads to do one kind of work | mixed | yes |
| `.agents/skills/asgard-platform/` | 73 | the wiki and the extracts written out so an agent can grep them, with a generated `index.md` mapping both halves as paths; from `scaffold/corpus.go`, not a template | md | yes |
| `scaffold/templates/` | 45 | the part of a customer repo that is the same every time | mixed | the skills only |
| `generate/templates/` | 12 | the CR skeletons `asgard-cli add` writes | English | no |

They are compiled into the binary and **also written into a customer
repository** by `asgard-cli init`, under `.agents/skills/asgard-platform/`.
That copy is generated and replaced when the binary's version moves; edit
the originals here.

**One fact, one home; everywhere else links.** A trap belonging to a CR template
is not also explained in a wiki page.

**No images.** The platform's screenshots are the product documentation's and are
fetched by URL; `corpus/wiki/screenshots.md` indexes which one answers which
question. Carrying the files themselves would put a copy here that goes stale
against asgard-docs while looking current, which is the failure the whole
embedding argument above exists to avoid - it holds for text because text is
where the judgement is, and inverts for assets.

### `stage/prompts/`

One file per piece of guidance. **The filenames are numbered and nothing else
is** - the numbers are the order they are usually reached in, kept because they
sort, and they are not a position anybody is at.

**Nothing raises guidance.** An onboarding is not linear, so a command that
derived one stage from the earliest missing CR kind told an engagement working
in a different order that it was behind, and could name only one thing at a
time. Guidance is reached by name with `asgard-cli guide`, or by grepping
`guide/` for the subject.

The prompts are Go templates with `<< >>` delimiters, rendered against the
repository's state, so a prompt can name the actual projects and requests rather
than placeholders. The half that needs no repository lands as a file; a
paragraph still carrying a template action after substitution is dropped from
what lands, because a file would freeze one moment of this repository's state.

### The index, and why it is not a page

`corpus/wiki/index.md` and `corpus/aliases.md` are the corpus's own
bookkeeping, readable by name and absent from any list of pages.

**An index inside a searched corpus competes with what it points at.** The
alias table lists every alias, so it carries every term of any translated query
and is reliably the document matching all of them - a query for a Chinese term
returned the word list rather than the page about it. It is beside the pages
rather than inside them.

`aliases.md` is applied to a query before searching, so the question can be
asked in the customer's own words. Two tables, and they behave differently on
purpose: an **alias** replaces the word, because a Chinese term appears nowhere
in an English corpus and keeping it only adds a term that lands nowhere; an
**entity** - a marketplace, a product - is added to the query, because the name
may be written verbatim in a page and replacing it would throw away the best
answer there is.

**A row nobody has needed is a guess.** Rows come from searches that came back
empty in a real engagement, which is also what `asgard-cli issue-report --new`
is for: the material's gaps are reported rather than guessed at.

### The link graph

Every document carries `kb.Doc.Links`, read when the document is parsed rather
than when a reader is printed one. A pointer resolved at the point of use is a
second regular expression that can disagree with the first about what a
document points at.

Holding it as data is what makes `--orphans` possible: **what does nothing point
at.** A dead pointer is loud, and a document nothing points at is silent and
costs more. The index does not count as a pointer there, because a page can sit
in it under a title nobody recognises while somebody spends a day on its
subject.

### `corpus/wiki/` and `corpus/usecase/`

Both are reference material and they answer different questions. The reading
order is wiki first: an extract assumes you already know the platform has that
shape.

`corpus/wiki/README.md` is the schema - the three layers, what a page must
carry, and how it is kept from going stale. `corpus/wiki/index.md` is
bookkeeping rather than a page about the platform, so it is readable by name
and not listed.

`corpus/wiki/glossary.md` carries the table of words with two senses here, and
it is on the page rather than in Go so that somebody reading the glossary can
see it and extend it, and so `audit-material --links` resolves the pointers its
rows carry.

Every wiki page ends with two things: a source block linking the rendered page
on docs.asgard-ai.com plus the commit it was read at, and an `**Unchecked:**`
line saying which parts were never held against a deployment. The first is the
only thing that makes "upstream moved and this page did not" detectable; the
second stops a reader assuming the wiki is checked as deeply as the extracts,
which by its nature it is not.


### `generate/templates/` and `scaffold/templates/`

Both use `<< >>` delimiters so that Helm's own `{{ }}` survives into the output.

`scaffold/templates/` mirrors the generated repo one-for-one. Path segments in
capitals are placeholders expanded at write time:

```
projects/__PROJECT__/chart/values-__ENV__.yaml.tmpl
docs/spec/__SPEC_SLUG__/README.md.tmpl
```

A `.tmpl` suffix means the file is rendered; anything else is copied verbatim.
`.agents/skills/` under it holds the seven design-time skills the coding agent
in the customer repo loads.

**The line is authority, not subject.** What ships here is what is true of any
Asgard, whoever is running it: the repository skeleton, the docs layers, the
declaration template, how to write plain Chinese, how to model a semantic layer
from a customer's own database. A skill stating what a particular server
accepts, rejects or calls things - the CRD shapes, the processor catalogue - is
served from the platform by `asgard-cli skill update` instead, because a
customer's server can be several versions from whichever release they installed
this from. See the embed comment in `scaffold/scaffold.go`.

`scaffold/skills.go` reads them as a `kb.Corpus`, so they carry links and
provenance like every other body of material and the same audits reach them.
They land in the same directory as the platform corpus and are a different
half of it: these are how to work, that is what the platform is.

## `source/` - never ships

`source/SOURCES.md` traces each extract back to the deployment it came from, and
records the **generational conflicts**: two charts that disagree in a way that is
dated rather than a matter of taste.

It lives outside `internal/` deliberately, so `go:embed` cannot reach it even by
accident. Everything under `internal/` ships to every engagement and names no
customer; this file is the one place that does.

## `hack/` - the contract check

`hack/validate-crs.py` holds rendered CRs and the extracts' skeletons against
asgard-kube's schemas: required fields, unknown fields, enums, patterns,
`maxItems`, `ExactlyOneOf`. `hack/extract-crs.py` gets the skeletons out of the
extracts, which are chart fragments rather than parseable YAML.

It exists because nothing else looks. `helm lint`, `asgard-cli check` and a
server-side dry-run all pass a document the apiserver would reject or silently
prune. `hack/README.md` is the procedure, and the PR template asks for its
output.

`.agents/skills/consistency-checks/` is this repository's own maintenance
skill - how to run a consistency pass, what each check is blind to, and how to
hold prose against the upstream it came from. **Not the design-time skills**:
those are `scaffold/templates/.agents/skills/` and land in a customer
repository. This one never leaves here, and `selfsrc` embeds it so the audits
read it.

`hack/sources.py` is where the upstream clones are: one environment variable
per source, resolved from the shell, then `.env`, then a default that is one
person's layout. `.env.example` is the template - **`.env.template` would be
gitignored**, because the rule is `.env.*` with `!.env.example` carved out.
Running it prints what each resolves to and how far behind it is, and nothing
here clones or pulls: a script that fetched would turn "read at this commit"
into "read at whatever was there when the script ran".

`hack/check-doc-paths.py` holds every path and every package-qualified Go
symbol named by this repository's own documents - the seven at the root, plus
this directory's README and scripts - against what is on disk - the
mirror of `audit-material --paths`, which does the same for what lands in a
customer's repository. A symbol resolves inside the package that owns it,
because a search of the whole tree cannot tell one package's Index from
another's.

`hack/check-coverage.py` recomputes the asgard-docs coverage row in
`internal/corpus/wiki/index.md` and fails when the page drifts from it. That
row is the material's own claim about how complete it is, and the two things
easiest to confuse in it are the number of links the material writes and the
number of pages there are.

`hack/check-tables.py` is the other half of the contract check: it holds the gate's pinned enum and
constraint tables against the generated CRDs, so a table that has fallen behind
the platform is reported rather than quietly warning about the wrong thing.
`hack/verify-references.sh` runs the whole gate over the reference deployments.
Neither ships in the binary - both need repositories that are not vendored in.

Not to be confused with `.agents/skills/db-query/scripts/`, which
`asgard-cli init` writes into a **customer** repo - that is the tool-chain for
reading the customer's own source systems at design time.

## Reference material that is not in this repo

Read-only, never vendored in. **The URL is the source of truth; where you clone
it is not** - so where you cloned it is an environment variable, not a path in
a script: `ASGARD_KUBE`, `ASGARD_DOCS`, `ASGARD_CORE`, `ASGARD_DEPLOYMENTS`.
`hack/sources.py` prints what each resolves to and how far behind it is.

| what | source of truth |
|---|---|
| CRD definitions, the platform contract | https://github.com/asgard-ai-platform/asgard-kube |
| product documentation | https://github.com/asgard-ai-platform/asgard-docs |
| the processor definitions the CRD is generated from | https://github.com/asgard-ai-platform/asgard-core (private) |
| the eight reference deployments | listed with their shapes in `AGENTS.md` |

A copy taken into this repo stops tracking upstream and then reads exactly like a
current one. Record the commit you read instead.

## What is not here

- **Almost no tests, and the ones here are deliberate.** A handful cover
  parsing and matching rules where a wrong answer is silent - `gate/credref`,
  `generate/refkeys`, `check/environments`, the pipeline manifest. Prose and
  material are covered by `audit-material` instead, which reads what ships
  rather than a copy of it. `go test ./...` runs in CI alongside `go vet` and
  `gofmt -l`; see "The gate" in `AGENTS.md`.
- **No `.out/` in version control.** It is gitignored and holds anything a command
  produces: hand-built binaries, command output, scratch programs.
- **No customer data anywhere.** Everything under `internal/` is generic; the
  customer's own knowledge lives in the repo the CLI writes, not in the CLI.
