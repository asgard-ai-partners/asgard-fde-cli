# asgard-fde-cli

Command line tool for Asgard FDE (`asgard-cli`).

## Development

```bash
go build -o asgard-cli ./cmd/asgard-cli   # build
./asgard-cli version                      # run
```

Layout:

```
cmd/asgard-cli/     main; signal handling and exit codes only
internal/cli/       cobra command tree, one file per subcommand
internal/config/    reads and writes .asgard-config.json
internal/work/      reads and writes the customer repo's own records of its work
                    (requests, task specs, open questions, decision records)
internal/deploy/    reads projects/<project>/deploy.yaml, the deployment SoT
internal/render/    renders a chart via helm, the way CD does
internal/gate/      the invariant checks on a rendered chart (xref, agent split)
internal/tool/      resolves helm/kubectl/python3, and how to install one
internal/wiki/      the platform wiki: what Asgard is made of and who each piece is for
internal/version/   build information (injected by GoReleaser via ldflags)
```

To add a subcommand, write a `newXxxCmd()` in `internal/cli/` and register it in
the `cmd.AddCommand(...)` call in `root.go`.

- [STRUCTURE.md](STRUCTURE.md) - what every directory is for, including the four
  bodies of embedded material and which one a change belongs to.
- [AGENTS.md](AGENTS.md) - the conventions this repo follows, and what the gate is
  now that there is no test suite.
- [TASK.md](TASK.md) - what this repo is for, and what is not finished.

## Commands

### `init`

Create `.asgard-config.json` in the current directory, binding it to a workspace.
The workspace is the customer:

```bash
asgard-cli init
```

```
Created /path/to/acme-asgard-kube/.asgard-config.json
  workspace.id    (not set yet)
  workspace.slug  acme
  workspace.name  acme
  repository      acme-asgard-kube

The workspace id is what the platform knows this customer by. Nothing here
needs it yet - namespaces come from the slug - so it can wait until the
platform has issued one:

    asgard-cli init --workspace-id ws_xxxxxxxx

No projects yet. Add one with `asgard-cli project add <slug>`.
```

`--workspace-id` is issued by the Asgard platform and is **optional**. Nothing
this CLI generates reads it, so waiting for one does not block the work that
comes before it. Its format is not validated, pending an API to verify it.

Setting it later on an already-initialised repository changes nothing else, so it
needs no `--force`:

```bash
asgard-cli init --workspace-id 1234567890123456789
```

Replacing an id that is already recorded does need `--force`, because that binds
the repository to a different workspace. `project add` and `check` both say when
the id is still unset - a project is what the platform deploys, so that is the
point at which it is worth chasing.

`--workspace-slug` defaults to the directory name with a trailing `-asgard-kube`
removed, so running inside `acme-asgard-kube` yields `acme`.
`--workspace-name` defaults to the slug. `--project` is a repeatable shortcut
that creates projects up front, each with the `dev` environment only.

Rerunning is safe: when the config exists nothing is changed and the current
settings are printed, except that a missing workspace id is filled in. `--force`
rebinds, and keeps the projects already recorded - their charts are on disk
either way - unless `--project` gives a new list.

### `project add`

A project is the unit of deployment: one Helm chart, one namespace per
environment. Projects are added as the engagement discovers them.

```bash
asgard-cli project add internal --env dev --env prod
```

```
Added project "internal" to /path/to/.asgard-config.json
  dev   asgard-acme-internal-dev
  prod  asgard-acme-internal-prod

Before the first deploy of each environment:
  1. tf-asgard must create the namespace and its app-secret first. Declaring an
     environment before they exist makes the next tag fail at helm upgrade.
  2. the project needs at least one Syncer. CD waits for a CronJob labelled
     asgard-ai.com/syncer-name and exits 1 after 180s if it finds none, even
     when helm upgrade succeeded.
```

`--env` may be repeated and defaults to `dev`; `dev` and `prod` are the only
valid symbols and a project may declare either or both. `--name` defaults to the
slug.

Both notes in the output are ordering traps: each one fails during CD rather than
here, which is far too late to find out.

### `scaffold`

Write the repository skeleton next to `.asgard-config.json`:

```bash
asgard-cli scaffold
```

It writes the part of a customer repo that is the same for every engagement:

| | |
|---|---|
| `AGENTS.md` | the platform contract, with the customer-specific sections marked TODO |
| `docs/` | the four-layer model (meeting-notes / decisions / living spec) and the SDD rules |
| `requirements/` | the task and request indexes |
| `scripts/check_*.py` | the four acceptance gates |
| `scripts/db/` | the query and introspection tool-chain, with an empty target registry |
| `.agents/skills/` | the six design-time skills |
| `common/` | `asgard-cli render`, the per-env overlay points, the runtime-skill directory |
| `.github/workflows/main.yaml` | tag-driven CD |
| `projects/<slug>/` | one chart skeleton per project, with `values-<env>.yaml` only for the environments that project declares |

What it does **not** write is the customer's own knowledge: which systems exist,
how the projects split, what the CRs look like. That is what the onboarding
produces, and no template can generate it.

Re-running is safe. Existing files are left alone and counted as already
present, so it can be run again after adding a project, or when a file was
deleted by hand. `--force` overwrites, which discards local edits.

The generated skeleton passes its own gate on the first run:

```bash
asgard-cli check                              # structure is consistent
helm lint projects/<slug>/chart/app           # 0 charts failed
```

### `next`

**The command to run after every step.** It reports which stage the onboarding is
at and what that stage needs, worked out from the repository itself - which files
exist, which CR kinds each chart declares - not from a counter in the config. So
it stays right when somebody does a step by hand, and it answers in an empty
directory too, where the answer is how to begin.

```bash
asgard-cli next                          # where am I, what now
asgard-cli next --list                   # every stage
asgard-cli next --stage requirements     # read one out of order
```

```
  0  init           Start the onboarding
  1  scaffold       Write the repository skeleton
  -  requirements   Turn what the customer said into a request
  2  projects       Decide how the work splits into projects
  3  data-sources   Wire up the customer's databases
  4  read-path      Decide each project's read path
  5  entry-point    Decide each project's entry point
  6  knowledge      Decide where unstructured knowledge lives
  7  verify         Run the acceptance gate
  8  deploy         Deploy
  9  enhance        Add a capability to a repo that is already live
  -  idle           Nothing in flight
```

Three of these sit outside the numbered walk. `requirements` is the customer
interview - it produces the request that stage 2 consumes, and it is read
deliberately because a conversation leaves no trace on disk until somebody writes
it down. `enhance` is the loop for a repo already live. `idle` is what an FDE sees
most often once a repo is running: nothing open, so the only question left is what
the customer wants next.

**Stages 4, 5 and 6 print the wrong answer next to the right one.** Those are the
three decisions this engagement got wrong once and reversed, and in each case the
wrong answer is the one that looks obvious.

Open questions print first, on every run, before anything else the command has to
say.

### `request`, `task`, `question`, `decision`

Work arrives as a **request**: one thing the customer wants that the agent cannot
do today. Everything else hangs off it.

```bash
asgard-cli request add "warehouse staff want to ask about stock levels in chat"
asgard-cli request target REQ-001 erp
asgard-cli request ready REQ-001

asgard-cli task add "expose stock levels" --request REQ-001 --project erp --complexity M
asgard-cli task ready TASK-001
asgard-cli task start TASK-001
asgard-cli task done TASK-001

asgard-cli question add "which stock figure is authoritative" --blocks REQ-001 --ask "the warehouse lead"
asgard-cli question answered 1 "location 608 only" --decision 2026-09-04-safety-stock.md

asgard-cli decision add "the website reads through fixed query tools" --module architecture.md
```

**None of this state lives in the CLI.** Each command writes a file in the
customer's repository, because that repository is what the next agent opens:

| record | file | carries |
|---|---|---|
| request | `requirements/requests/REQ-xxx-<name>.md` + its registry row | date raised, status, target project, the customer's own wording |
| task spec | `requirements/tasks/TASK-xxx-<name>.md` + its queue row | date created, status, the SDD sections, a dated log of every transition |
| open question | a row in `docs/open-questions.md` | date raised, what it blocks, who can answer |
| decision | `docs/decisions/YYYY-MM-DD-<topic>.md` + a traceability row | the date in the file name, the module it changed |

The reason these are commands rather than instructions to write a file is that
each record lives in more than one place. A task's status is in the queue table,
in the spec's own `Meta`, and in the spec's execution log; a request's target
project is in the registry's Spec column and in its `Meta`. Moving one by hand
means three or four edits, and a repository where two of them disagree gives the
next reader no way to tell which is current. Every command here moves all of
them, and stamps the date rather than asking for it.

`asgard-cli next` reads all four back. Open questions print first, on every run,
before anything else it has to say.

### `add`

Write a CR skeleton into a project's chart, correct in the parts that fail
silently.

```bash
asgard-cli add                                   # list the kinds
asgard-cli add dataconnector erp --db-class postgres --project erp
asgard-cli add flowagent support --bot-class line --project site
```

```
created projects/erp/chart/app/templates/data_connector/dc-erp.yaml
updated projects/erp/chart/app/values.yaml (added the values it reads)

Next:
  1. what it is:      asgard-cli wiki settings
  2. how to build it: asgard-cli usecase semantic-layer
  3. fill in the TODOs
  4. verify:          asgard-cli check
                      asgard-cli verify erp
```

Ten kinds: `dataconnector`, `semanticlayer`, `agent`, `httptool`, `querytool`,
`skillset`, `trigger`, `knowledgedrive`, `plugin`, `flowagent`.

**What it generates is a skeleton**: the structure and the traps are right, the
content is marked TODO. The parts worth generating are the ones nothing catches -
a missing display annotation shows a nameless resource in the UI, a Workflow
without its set labels is invisible there, a Trigger without its own two labels
opens as a blank canvas, and a field renamed upstream still lints clean under its
old name. None of those is caught by `helm lint`, by CRD validation, or by a
server-side dry-run.

It reads the chart before writing into it, so a reference it emits points at
something that exists: one SemanticLayer in the chart is mounted, several are
refused by name, a SkillSet is referenced only if one is there. A second query
tool does not re-emit a Toolset the first one already wrote.

The two pointers it prints are in reading order and answer different questions -
`wiki` says what the thing is, `usecase` says how it is assembled and assumes you
already know the first.

### `wiki`, `usecase`

Two bodies of reference material, embedded in the binary rather than written into
a customer repo: a copy in one engagement goes stale where nobody is looking,
while a stale page here is fixed for every engagement in one release.

```bash
asgard-cli wiki                       # what the platform is made of
asgard-cli wiki agents
asgard-cli wiki --search "匿名 訪客"
asgard-cli wiki --conventions         # how the wiki is maintained

asgard-cli usecase                    # how each deployment shape is built
asgard-cli usecase flow-agent-single
asgard-cli usecase --search schedule
```

| | answers | written from |
|---|---|---|
| `wiki` | what the platform is, who each piece is for, and where the UI's names stop matching the resources a chart declares | the product documentation, [asgard-docs](https://github.com/asgard-ai-platform/asgard-docs), checked against the CRDs |
| `usecase` | how one shape of deployment is assembled, field by field, and what a wrong value costs | deployments already in production |

An extract assumes you already know the platform has that shape; a wiki page is
where that assumption comes from. `asgard-cli add` prints one of each.

**To look something up, use [`find`](#find)**, which searches both and names the
counterpart of whatever it hits. `--search` on either command is the narrow form,
for when you already know which half holds the answer.

The wiki's own conventions - its three layers, what a page must carry, and how it
is kept from going stale as the platform moves - are in `asgard-cli wiki
--conventions`.

### `find`

Search both bodies of reference material at once, when it is not obvious which
holds the answer.

```bash
asgard-cli find schedule
asgard-cli find anonymous visitor
```

```
PLATFORM - what the thing is (asgard-cli wiki <page>)

  automation         Trigger and API
                     Starts a conversation with an agent on a schedule.
                     -> field level: asgard-cli usecase trigger

SHAPES - how it is assembled (asgard-cli usecase <name>)

  trigger            Trigger
                     -> what it is:  asgard-cli wiki automation

Read the platform side first; an extract assumes you have.
```

It follows the link between the two, so a hit in either half hands over the
other - in the order they should be read. Every term has to appear, so an extra
word narrows rather than widens.

### `check`

The first step of the acceptance gate, and the only one that needs no external
tool. It verifies the invariants a chart render cannot see - the ones that
otherwise surface at deploy time, or when the next person picks the repo up:

```bash
asgard-cli check              # whole repo
asgard-cli check erp          # project-scoped checks limited to erp
```

```
warn   common/skills/ has no skill directories yet
ok  structure is consistent (1 project(s): [erp])
```

- the root README's project table matches the directories under `projects/`
- every project has a `deploy.yaml`, its envs are `dev` or `prod`, and the values
  files it names exist, along with the shared `common/values-<env>.yaml`
- runtime skills under `common/skills/` carry `name` and `description`
  frontmatter, with the name matching the directory
- the SDD entry points under `requirements/` are present
- the `docs/` spec layer is intact: required files, the living spec's module index
  matching the files on disk, dated filenames, and every relative link inside
  `docs/` resolving
- **no page is an orphan** - a document under `docs/` or `requirements/` that
  nothing links to is not read, and the person who wrote it never finds out,
  because the file is still there. A warning rather than an error: a decision
  recorded today and not yet applied is an orphan for as long as that takes.

Naming projects limits the project-scoped checks to those; the repo-wide checks
always run. It exits non-zero when anything fails, and warnings do not fail it.

The rot this cannot see - two pages that contradict each other, a claim a newer
source superseded, a concept every document explains in passing and none owns -
needs a reader. The `knowledge-base` skill under `.agents/skills/` in the
generated repo is the pass for that.

### `render`, `verify`, `doctor`

These three are why the acceptance gate now runs on Windows.

```bash
asgard-cli render erp dev              # manifests to stdout, summary to stderr
asgard-cli verify                      # render every declared env, check invariants
asgard-cli verify --rendered file.yaml # check a stream that is already rendered
asgard-cli doctor                      # which external tools are here, and how to get them
```

`render` replaces the generated repo's `common/render.sh`, and `verify` replaces
its `check_chart_xref.py` and `check_agent_split.py`. The old chain was:

```
bash render.sh  ->  yq (read deploy.yaml)  ->  helm template  ->  python3 + PyYAML
```

Four external dependencies, of which **three do not work on Windows without WSL
or Git Bash** - while helm and kubectl both have native Windows builds. The
prerequisite is now `helm` alone, and one binary:

```
asgard-cli verify  ->  helm template  ->  internal/gate
```

`verify` renders in process, so there is no pipeline and no temporary file, and
`render` keeps the manifests on stdout with everything else on stderr so the pipe
forms still work identically in cmd, PowerShell and bash.

The gate rules were ported one for one, and the port was checked by running both
implementations over the same rendered chart: same findings, same counts. Step 4
of the gate still needs `kubectl` and `check_crd_fidelity.py`, because it talks
to a cluster.

### helm and kubectl are prerequisites, not dependencies

Nothing about how asgard-cli is distributed can install them. A tar.gz, a zip and
`go install` carry no dependency metadata and never can, and a dependency
declared on a Homebrew tap or a Scoop bucket would only cover people who install
that way - which is nobody today, since neither repository exists. Declaring one
anyway would read as a guarantee that does not hold.

So the binary is the mechanism. Every command that needs helm resolves it through
`internal/tool` first and refuses with the install line for the machine it is on,
and `asgard-cli doctor` reports all of them at once:

```
$ asgard-cli doctor
platform  darwin/arm64

MISSING helm               render a chart (asgard-cli render) and lint it
ok    kubectl              v1.35.1
ok    python3 (optional)   Python 3.14.7

helm is not on PATH, and asgard-cli needs it to render a chart and lint it

Install it with:

    brew install helm
```

It works out the command for the machine it runs on, including which Linux
distribution - neither kubectl nor helm is in the Debian or Ubuntu default
repositories, so the honest answer there is not an `apt install`. It exits
non-zero when a required tool is missing, so it works as a CI preflight.

`init`, `scaffold`, `next`, `project`, `request`, `task`, `question`, `decision`
and `check` need none of these tools. `render` and `verify` need helm; step 4 of
the gate needs kubectl.

### The config file

```json
{
  "workspace": {
    "id": "7ab7f523-3cd9-7e87-a873-6f1fa6028104",
    "slug": "acme",
    "name": "acme"
  },
  "projects": [
    { "slug": "internal", "name": "internal", "environments": ["dev", "prod"] },
    { "slug": "website",  "name": "official site", "environments": ["dev"] }
  ]
}
```

The slug is load bearing. Two names are **derived** from it rather than stored,
so the config cannot drift from the layout on disk:

| derived | rule | example |
|---|---|---|
| repository | `<workspace.slug>-asgard-kube` | `acme-asgard-kube` |
| namespace | `asgard-<workspace.slug>-<project.slug>-<env>` | `asgard-acme-internal-dev` |

Both commands validate before writing, so an invalid config is never created:

| check | why |
|---|---|
| slugs are DNS-1123 labels | a slug becomes part of a namespace, so anything else is rejected later by the apiserver |
| environments are `dev` or `prod`, no duplicates | those are the only symbols the platform accepts |
| project slugs are unique | two projects with one slug would fight over a namespace |
| every derived namespace fits in 63 characters | names derived from a namespace inherit its length; without this the failure surfaces in `helm upgrade` during CD |

Every problem is reported at once rather than one per run.

`.asgard-config.json` is project configuration and belongs in version control.
Read and write it through `internal/config` (`config.Load` / `config.Save` /
`config.Find`) rather than assembling JSON inside a command. `config.Find` walks
up from the current directory, so a command still resolves the project root when
run from a subdirectory.

## Releasing

Releases are driven by [GoReleaser](https://goreleaser.com). Pushing a tag triggers
`.github/workflows/release.yml`:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

One run produces binaries for linux / darwin / windows x amd64 / arm64, `.deb` /
`.rpm` / `.apk` packages and checksums, all attached to the GitHub Release. The
changelog is grouped automatically from conventional commit messages (`feat:`,
`fix:`).

To verify locally without publishing anything (output lands in `dist/`):

```bash
goreleaser check
goreleaser release --snapshot --clean --skip=publish
```

### Channels not yet enabled

The bottom of `.goreleaser.yaml` has ready-made **Homebrew tap** and **Scoop
bucket** blocks. Create the corresponding repo and a PAT with write access to it,
then uncomment:

| Channel | Prerequisite |
| --- | --- |
| Homebrew | Create `asgard-ai-partners/homebrew-tap`, secret `HOMEBREW_TAP_TOKEN` |
| Scoop | Create `asgard-ai-partners/scoop-bucket`, secret `SCOOP_BUCKET_TOKEN` |

Other things worth knowing:

- **CGO**: builds run with `CGO_ENABLED=0` so cross-compilation fits on a single
  runner. Pulling in a cgo dependency (sqlite and friends) means switching to
  zig cc or per-platform runners.
- **macOS signing**: an unsigned binary downloaded through a browser is blocked by
  Gatekeeper (installing via Homebrew is not affected). Add `anchore/quill` when
  notarization becomes necessary.
- **LICENSE**: the archive config picks up `LICENSE*`. No such file exists yet, so
  GoReleaser prints one warning, which does not affect the release.
