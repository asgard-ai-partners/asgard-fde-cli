# STRUCTURE.md

What every directory in this repo is for. `README.md` is what the commands do,
`AGENTS.md` is the rules for changing them, `TASK.md` is what the repo is for and
what is unfinished. This file is the map.


**How the mechanisms work, and the failure each was built from, is
[APPROACH.md](APPROACH.md).** This file is where things live.

## The shape in one line

A single Go binary that answers questions about integrating with Asgard, and
writes chart skeletons into **somebody else's** repository. It holds no state:
everything an engagement knows ends up in the customer's repo, because that repo
is what the next agent opens.

The first half works with no repository at all, and has to keep doing so - the
question gets asked in a meeting, before there is a directory. `TASK.md` states
the goal; this file is the map.

```
cmd/asgard-cli/       main; signal handling and exit codes only
internal/             every package, none exported
source/               internal notes that must never ship
hack/                 this repo's own tooling: the CRD contract check
.github/             CI, the tag-driven release, and the PR template
.goreleaser.yaml      how the binary is built and published
```

## `internal/` - the code

Ordered by size. The top of the table is the command surface and the records it
keeps; below it are the checks, the material servers and the plumbing.

| package | go | what it holds |
|---|---|---|
| `cli` | 12467 | the cobra command tree, one file per subcommand, plus `root.go`, `repo.go`, `format.go` and `find.go` (which spans every corpus rather than serving one) |
| `gate` | 1976 | the invariant checks on a rendered chart - xref, agent split, deployability, enums, constraints, conditional CEL shapes |
| `work` | 1678 | the customer repo's own records: requests, task specs, open questions, decisions, and the two reading logs |
| `platform` | 1218 | the platform API client: workspaces, the whole `/v1/iac` surface, and `/v1/docs` |
| `check` | 1179 | repository structure: indexes, dated names, links, orphan pages |
| `auth` | 1118 | the OAuth 2.0 + PKCE sign-in and the credential store, which is the only file this CLI keeps outside a repository |
| `generate` | 714 | CR skeletons for ten kinds, wired to what the chart already declares |
| `kb` | 798 | one implementation of listing, reading, scoring, provenance and the link graph, shared by every corpus |
| `scaffold` | 1603 | writes the non-customer-specific tree, serves the design-time skills inside it, exports the wiki and the extracts as files under `.agents/skills/asgard-platform/`, and keeps `.asgard-scaffold.json` - the record of which CLI wrote the files this binary ships |
| `stage` | 565 | the onboarding prompts, rendered against the repository's state |
| `size` | 380 | the deployment shapes, counted off production, and what one costs before anything is added |
| `tool` | 329 | resolves helm/kubectl/python3 and says how to install one |
| `brief` | 395 | what one activity gets wrong, addressed by intent rather than by position |
| `skills` | 268 | the platform's fetched reference material, and the record of which version is here |
| `binding` | 224 | reads and writes `.asgard-cli.yaml`, the checkout's platform binding |
| `corpus` | 39 | the material itself, in the layout a repository receives it: `wiki/` and `usecase/` side by side, so a pointer can become a path that resolves in both trees |
| `wiki` | 242 | serves the platform wiki, and the two index tables beside it |
| `render` | 206 | renders via `helm template`, with the reserved `asgard` block supplied as placeholders |
| `repo` | 147 | what a customer repository is made of, by looking at it |
| `pipelineconfig` | 147 | reads `.asgard-pipeline.yaml`, the deployment declaration |
| `gitrepo` | 120 | the checkout's root and its remotes, read and never compared to anything |
| `chart` | 117 | reads a project's **unrendered** templates for (kind, name) |
| `version` | 73 | build information, injected by GoReleaser via ldflags |
| `usecase` | 46 | serves the deployment-shape extracts |
| `needs` | 188 | what a shape has to be given by the customer, as seven documents written into a repository beside the extracts |
| `browser` | 39 | opens a URL, or says it could not |

**Two packages this table used to list are gone.** `config` held
`.asgard-config.json` - the workspace slug, the project list and each project's
"shape" - and every field of it failed the same test: *a value belongs in a
config file only when nothing on disk implies it and the platform cannot be
asked.* `deploy` held `projects/<p>/deploy.yaml`, which the Pipeline cut-over
replaced with the one declaration at the repository root.

**Why nothing decides when a chart is finished.** A shape is the one thing the
files cannot say: a SemanticLayer with nothing mounted on it is either a
finished Mimir deliverable or an agent nobody has written yet, and those are
identical on disk. It used to be recorded per project so that `stage.Gaps` could
subtract - and recording it meant treating somebody's note of intent as a
specification this tool could check. Both are gone. `size` still counts the
shapes off deployments in production, for a person to compare against.

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

**Four of these are one corpus** - `asgard-cli find` searches them together and
they share one schema: a `# ` title, a summary, and `**Checked:**` /
`**Unchecked:**`. `asgard-cli find --unverified` is the check, and it is 0 of 25,
0 of 21, 0 of 12, 0 of 7. The fifth, `generate/templates/`, is not searched: it
is what `add` writes, not something anybody reads to decide.

| where | files | answers | language | searched |
|---|---|---|---|---|
| `corpus/wiki/` | 28 | what the platform is, and who each piece is for | English | yes |
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

There was a walk: `stage.Current` derived one stage from the earliest missing CR
kind and reported it as where the onboarding stood. It went, and `stage.Relevant`
replaced it - guidance raised from conditions the repository meets, printed with
the condition beside each. That was the same ladder: its three per-project cases
were the old rungs in a `switch`, so a chart missing two things heard about the
first. It is gone too, and so is `stage.Gaps`, which subtracted what a chart
declares from what its declared shape asked for - the shape was a note of
intent, and there is no longer anywhere to record one. **Nothing raises guidance
now**; `find` reaches any document by subject. A document reachable only by
arriving at it is unreachable, and the measured version of that is in TASK.md.

The prompts are Go templates with `<< >>` delimiters, rendered against the
repository's state, so a prompt can name the actual projects and requests rather
than placeholders.

### The index, and why it is not a page

`corpus/wiki/index.md`, `corpus/wiki/log.md` and `corpus/aliases.md` are the corpus's own
bookkeeping. The first two were always unlisted; the third used to be a section
of `corpus/wiki/glossary.md` and was moved for a measured reason.

**An index inside a searched corpus competes with what it points at.** The alias
table lists every alias, so it carried every term of any translated query and
was reliably the one document matching all of them: `find 電商` returned the word
list rather than `taiwan-channels`. It is beside the pages now.

`asgard-cli find` applies it to a query before searching, so the question can be
asked in the customer's own words. Two tables, and they behave differently on
purpose: an **alias** replaces the word, because a Chinese term appears nowhere
in an English corpus and keeping it only adds a term that lands nowhere; an
**entity** - a marketplace, a product - is added to the query, because the name
may be written verbatim in a page and replacing it would throw away the best
answer there is.

Rows come from searches that came back empty. `find` records those in an
engagement, `asgard-cli reading --misses` reads them back, and
`issue-report --new` puts them in a report. A row nobody has needed is a guess.

### The link graph

Every document carries `kb.Doc.Links`, read when it is parsed. `find` names a
hit's counterpart off that field, and `audit-material --links` resolves the same
field - it was two regular expressions at two points of use, which could
disagree about what a document pointed at.

Holding it as data is what makes `--orphans` possible: **what does nothing point
at.** A dead pointer is loud, and a document nothing points at is silent and
costs more. The index does not count as a pointer there, because `wiki
operations` sat in it under the title Connectivity while an FDE spent a day on
connectivity and never opened it.

### `corpus/wiki/` and `corpus/usecase/`

Both are reference material and they answer different questions. The reading order
is wiki first: an extract assumes you already know the platform has that shape.

`corpus/wiki/README.md` is the schema - the three layers, what a page must carry, and
how it is kept from going stale. `corpus/wiki/index.md` and `log.md` are its
bookkeeping rather than pages about the platform, so they are readable by name
but not listed.

`corpus/wiki/glossary.md` carries one table the code reads: the words a customer
uses against the words this material uses. `find` translates a query through it
before searching, because the corpus is English and the conversation it came
from was not. It lives on the page rather than in Go so that somebody reading
the glossary can see it and extend it, and so `audit-material --links` resolves
the pointers its rows carry.

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
`.agents/skills/` under it holds the seven design-time skills the coding agent in
the customer repo loads. **Only five, and the line is authority.** A skill that
states what a particular Asgard server accepts or calls things - the CRD shapes,
the processor catalogue, `asgard-cr-verification` - is served from the platform
and written by `asgard-cli skill update`, because a customer's server can be
several versions from whichever release they installed this from. See the embed
comment in `scaffold/scaffold.go`.

Those seven are **searchable**, and `scaffold/skills.go` is what makes them so.
They are material like the wiki is material, and leaving them out of `find` meant
an agent asked to build a deck searched for one and was told nothing matched
anywhere - while the skill that owns the subject sat in the repository it was
standing in.

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

Not to be confused with `.agents/skills/db-query/scripts/`, which `asgard-cli
scaffold` writes into a **customer** repo - that is the tool-chain for reading
the customer's own source systems at design time.

## Reference material that is not in this repo

Read-only, never vendored in. **The URL is the source of truth; where you clone it
is not.**

| what | source of truth |
|---|---|
| CRD definitions, the platform contract | https://github.com/asgard-ai-platform/asgard-kube |
| product documentation | https://github.com/asgard-ai-platform/asgard-docs |
| the processor definitions the CRD is generated from | https://github.com/asgard-ai-platform/asgard-core (private) |
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
