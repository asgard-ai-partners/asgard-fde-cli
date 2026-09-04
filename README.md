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
internal/config/    reads and writes .asgard-config.json, including each project's shape
internal/work/      reads and writes the customer repo's own records of its work
                    (requests, task specs, open questions, decision records)
internal/kb/        one implementation of listing, reading, scoring and provenance,
                    shared by every part of the material
internal/wiki/      the platform wiki, and the glossary's customer-vocabulary table
internal/usecase/   the deployment-shape extracts
internal/stage/     which stage a repo is at, derived and never stored, plus which
                    guidance the repo's own state makes relevant
internal/scaffold/  writes the non-customer-specific tree, and serves the skills in it
internal/generate/  CR skeletons, wired to what the chart already declares
internal/size/      the deployment shapes, counted off production
internal/brief/     what one activity gets wrong, addressed by intent not by stage
internal/check/     repository structure: indexes, dated names, links, orphan pages
internal/gate/      the invariant checks on a rendered chart (xref, agent split, enums)
internal/render/    renders a release's chart via helm, with placeholder asgard values
internal/binding/   reads and writes .asgard-cli.yaml, the checkout's platform binding
internal/pipelineconfig/ reads .asgard-pipeline.yaml, the deployment declaration
internal/chart/     reads a project's unrendered templates for (kind, name)
internal/tool/      resolves helm/kubectl/python3, and how to install one
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
| `scripts/db/` | the query and introspection tool-chain, with an empty target registry |
| `.agents/skills/` | the six design-time skills |
| `common/` | the runtime-skill directory |
| `.asgard-pipeline.yaml` | the deployment declaration, with one release per project to fill in |
| `projects/<slug>/` | one chart skeleton per project |

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

### `guide`

**`guide` reads one decision.** There is no command that says where the
engagement is, and that is deliberate - the two that did are gone. `next`
reported a position on a walk. `status` replaced it, reported the repository,
and then named the guidance the shape of it raised: the same rungs, in the same
order, with the numbers taken off. What the second one printed from files is now
read by `project`, `question`, `request` and `task`, each from its own file.

```bash
asgard-cli guide                   # all the guidance
asgard-cli guide requirements      # one piece of it, any time
asgard-cli find "<terms>"          # reach any of it by subject
```

```
  init           Start the onboarding
  scaffold       Write the repository skeleton
  requirements   Turn what the customer said into a request
  projects       Decide how the work splits into projects
  data-sources   Wire up the customer's databases
  read-path      Decide each project's read path
  entry-point    Decide each project's entry point
  knowledge      Decide where unstructured knowledge lives
  verify         Run the acceptance gate
  deploy         Deploy
  enhance        Add a capability to a repo that is already live
  idle           Nothing in flight
```

**None of these is a step you arrive at.** They were numbered once, and `next`
derived "stage 4 of 9" from the earliest missing CR kind. That was wrong in both
directions: it could name only one thing, so three of them were unreachable
unless you already knew their names, and a position cannot be argued with, so an
engagement working in a different order was told it was behind. An engagement
that has already gathered every requirement has no stage at all, and the tool
used to insist otherwise.

Removing the numbers was not enough. `status` raised the same rungs from
conditions instead - `!p.Has("DataConnector")` in a `switch`, so a chart missing
two things was told about the first - which is a position with the arithmetic
hidden. Nothing raises guidance now. It is reached by name with `guide` and by
subject with `find`.

**`read-path`, `entry-point` and `knowledge` print the wrong answer next to the
right one.** Those are the three decisions this engagement got wrong once and
reversed, and in each case the wrong answer is the one that looks obvious.

**A chart does not always end with an entry point.** Which CR kinds finish one
depends on the project's shape, and the shape is the thing the files cannot say
- a SemanticLayer with nothing mounted on it is either a finished Mimir
deliverable or an agent nobody has written yet. Declare it and `project` stops
reporting what that shape does not have as missing:

```bash
asgard-cli project shape insight mimir-dashboard
asgard-cli project shape insight          # the current one, and the choices
```

### `project`, `request`, `task`, `question`

**Four commands read the repository back to you**, one file each, each with
`--format json`. None of them infers anything from the others.

```bash
asgard-cli question    # what nobody has answered yet, and who each is with
asgard-cli request     # what the customer asked for and is not done
asgard-cli task        # the task specs that are open
asgard-cli project     # what each chart declares, and what its shape lacks
```

**Read `question` first.** The fastest way to do damage in a repository somebody
else started is to design past a question they already knew was open.

```
Projects:

  insight              mimir-dashboard
                       DataConnector, SemanticLayer
                       complete for its shape - which is not the same as deployed

  helpdesk             shape not declared
                       chart is empty
                       no shape declared, so nothing is claimed about what it lacks
                       `asgard-cli project shape helpdesk <shape>`, or `asgard-cli size` for the list
```

**A chart with no declared shape gets no verdict.** Against a declared shape,
"this shape asks for X and X is absent" is subtraction. With no shape there is
nothing to subtract from, and answering anyway means assuming a set of kinds
every chart wants - which is the ladder, rebuilt out of a default.

### `request`, `task`, `question`, `decision` - writing the records

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

`asgard-cli question`, `request`, `task` and `project` read them back, each with
`--format json`.

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

**The way in.** It searches all four parts of the material at once - the wiki,
the extracts, the stage guidance and the design-time skills - because which of
them holds an answer is usually not obvious before searching. It needs no
repository.

```bash
asgard-cli find schedule
asgard-cli find anonymous visitor
asgard-cli find 儀表板                    # translated before the search runs
asgard-cli find schedule --format json   # each hit with the command that reads it
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

It follows the link between the halves, so a hit in either hands over the other -
in the order they should be read. Every term has to appear, so an extra word
narrows rather than widens; when nothing carries them all the search widens and
says which terms it could not place.

**Ask in the language the question was asked in.** The material is English and a
customer conversation is not, so the index is applied before the search runs and
the rewrite is printed:

```
$ asgard-cli find 電商
This material is in English. "電商" was read as:

    commerce marketplace channel

PLATFORM - what the thing is (asgard-cli wiki <page>)

  taiwan-channels    The commerce channels a customer will name, and what we have
```

The index is `asgard-cli wiki --aliases`, and it is **not a page**. It sits
beside the pages, with `index.md` and `log.md`, because an index inside a
searched corpus competes with what it points at: the table lists every alias, so
it reliably carried every term of a translated query and the reader got the word
list rather than the page.

It has three tables. An **alias** replaces the word - 電商 appears nowhere in an
English corpus, so keeping it only adds a term that lands nowhere. An **entity**
is *added* to the query, because the name may be written verbatim in a page and
replacing it would throw away the best answer there is.

The entities are split, and the split is the point:

| table | means |
|---|---|
| names the material **covers** | somebody searched the deployments and recorded the answer. SHOPLINE, Shopee, momo, PChome, 蝦皮, Coupang |
| names it only **routes** | nothing here names it. The row reaches the *shape* it belongs to, which is what the material has |

**A row that routes reads exactly like a row that answers**, so `find` says
which it was:

```
$ asgard-cli find 綠界
**Nothing here names 綠界.** What follows is the shape it belongs to, which
is what this material has - not material about the product. Nobody has
searched the reference deployments for it, and until somebody does, the
answer to "do we already integrate it" is not in this tool.

  asgard-cli wiki taiwan-channels   the four, and what each one costs
  asgard-cli question add "which of the four shapes does 綠界 give us" \
      --ask "<who at the customer>"
```

and then returns `usecase external-api`, `usecase write-path`, `browser-operation`
and the unknown that blocks a payment gateway on a public site - `wiki
platform-unknowns` P8, who presses approve on an anonymous channel. **Moving a
row from the second table to the first means somebody did the search**, and
nothing else.

**A word this material has taken is flagged before the results, not after.** A
search that finds nothing is recorded and the reader is told so; a search that
finds the *wrong sense* of a word looks exactly like an answer, and nothing is
red anywhere:

```
$ asgard-cli find payment
These results use a word that means one thing here, and it may not be
the one that was asked about:

  payment      is     billing between Asgard and this customer - see
                      `asgard-cli wiki fehu`
               is not **the customer's own payment gateway**, which is an
                      external system with side effects: `asgard-cli usecase
                      write-path` ...
```

That table is `asgard-cli wiki glossary`, and it is applied to a query rather
than only read by a person. It existed as prose for a long time while the
failure it describes went on happening.

**A dead end asks a question instead of guessing.** When a query names a system
this material has never had, the useful answer is not a phrasing hint - it is
that the work is decided by which of four shapes the system presents, and that
nobody here can answer it:

```
**That is a question for the customer, and not one this tool can answer.**

  asgard-cli question add "which of the four shapes does <it> give us" \
      --ask "<who at the customer>"
```

**A search that finds nothing is recorded**, in an engagement, to a file that is
not committed - a query carries whatever words the customer used.

```bash
asgard-cli reading --misses      # what this engagement searched for and did not find
asgard-cli issue-report --new    # the report, with that evidence already in it
```

That is the one part of a defect report nobody has to be believed about: the
tool witnessed it. Each line is either a missing index row or a missing page,
and the two need different fixes. A row is added when a search came back empty
and the subject turned out to exist under another name; a row nobody has needed
is a guess.

### `brief`, `size`, `reading`, `issue-report`

Four ways in, none of them a position. None needs a
repository except `reading`.

**`brief`** answers "the thing I am about to do - where will I get it wrong",
which no repository report can: the riskiest activity leaves no trace in one,
because talking to a customer changes no file, and meetings happen at every
stage.

```bash
asgard-cli brief                     # the activities
asgard-cli brief customer-meeting    # before any customer conversation
asgard-cli brief write-chart         # before authoring CRs
```

**`size`** is what one capability is made of before it is written - the first
question a proposal is asked, and the basis of a quote. The counts come from
deployments in production rather than from reasoning, which matters most where
the intuitive answer is wrong: **the flow-agent shapes contain no `Agent` CR at
all.** The same shapes are what a project declares with `project shape`.

```bash
asgard-cli size                      # the shapes, and what each costs empty
asgard-cli size flow-agent-single --databases 2 --queries 4
```

**`reading`** reports which pages this engagement opened and which it never did.
Every read of `wiki`, `usecase` and `guide` inside an engagement appends
a line to `docs/.reading-log`. It records page names and nothing about the
customer, so it is safe to commit - and worth committing, because six months
later it says what the person before you knew. **The pages that cost the most
are the right ones nobody opened.**

**`issue-report`** is how a gap in this tool gets filed. The gap does not belong
in the customer repository: a note in one engagement is a note one engagement
has, and the next one starts over.

### `check`

The first step of the acceptance gate, and the only one that needs no external
tool. It verifies the invariants a chart render cannot see - the ones that
otherwise surface at deploy time, or when the next person picks the repo up:

```bash
asgard-cli check                    # whole repo
asgard-cli check erp                # project-scoped checks limited to erp
asgard-cli check --format json      # errors and warnings as separate arrays
```

```
warn   common/skills/ has no skill directories yet
ok  structure is consistent (1 project(s): [erp])
```

- the root README's project table matches the directories under `projects/`
- `.asgard-pipeline.yaml` parses, names no release twice, and every release it
  declares points at a chart directory that has a `Chart.yaml`
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

`asgard-cli verify` adds the invariants a render carries, including the CRDs'
conditional CEL rules: a credential that sets neither a literal nor a reference
or both, a class block missing or doubled, a `toolsetClass` without the block it
requires. **Every one of those renders, lints and passes a server-side dry-run**,
and is refused at apply. Forty of the 79 rules are `self == oldSelf` and cannot
be seen offline at all.

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
asgard-cli verify --format json        # one record per render, each check named
asgard-cli doctor                      # which external tools are here, and how to get them
```

**`check` and `verify` are the pair an agent works hardest**, because they are
the gate it is trying to turn green - so both take `--format json`. In text a
warning and a failure differ by one word at the left margin and only one of them
is fatal; in JSON they are separate arrays. A failing JSON run exits 1 and
prints nothing to stderr: the report already says it failed, and a second
account of it on another stream is a second source for one fact.

`render` replaces the generated repo's `common/render.sh`, and `verify` replaces
its `check_chart_xref.py` and `check_agent_split.py`. The old chain was:

```
bash render.sh  ->  yq  ->  helm template  ->  python3 + PyYAML
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

`init`, `scaffold`, `project`, `request`, `task`, `question`, `decision`
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

### `login`, `logout`, `whoami`

Sign in to the Asgard platform, so that the `pipeline` commands can act as you.

```bash
asgard-cli login                     # sign in to prod
asgard-cli login --profile dev       # sign in to dev
asgard-cli login --no-browser        # print the URL instead of opening one
asgard-cli whoami                    # ask the platform who the session is
asgard-cli logout --all              # forget every stored session
```

OAuth 2.0 authorization code with PKCE over a loopback redirect, which is what
RFC 8252 asks for on a machine that has a browser. The binary ships no client
secret. The session is stored under this user account - never inside a customer
repository - at `os.UserConfigDir()/asgard-cli/`, 0600.

Two profiles exist: `prod` (the default) and `dev`. `--profile` picks per
command, `ASGARD_PROFILE` sets it for a shell, and `login --set-default` records
one. `ASGARD_API`, `ASGARD_ISSUER` and `ASGARD_CLIENT_ID` override a profile's
fields one at a time, for a platform running somewhere else.

With no browser - CI, a container, an agent sandbox - set `ASGARD_TOKEN` to an
access token instead. It bypasses the store completely, reading nothing from
disk and writing nothing to it.

### `workspace`

Choose which workspace the platform commands act in.

```bash
asgard-cli workspace list            # what this account can reach
asgard-cli workspace use <id>        # bind this repository to one
asgard-cli workspace show            # which one applies here, and why
```

The binding is kept beside the credentials, keyed by profile and by the
repository's origin remote - not in the repository, because the declaration
contract for a customer repository is a chart and one `.asgard-pipeline.yaml`,
and not on the platform, which does not know which directory holds its
repository. A repository scaffolded with a `workspace.id` in
`.asgard-config.json` is read from there first.

`workspace show` reports *why* that workspace, which is the useful half: acting
in the wrong one is the failure the resolution order exists to prevent.

### `pipeline`

Deploy the repository through the platform's IaC pipeline.

```bash
asgard-cli pipeline connect                     # connect GitHub to this workspace
asgard-cli pipeline connections                 # the installations connected
asgard-cli pipeline repos                       # what one can reach
asgard-cli pipeline create --name <name>        # bind this repository
asgard-cli pipeline show                        # the pipeline bound to this checkout
asgard-cli pipeline projects                    # projects a release can deploy into
asgard-cli pipeline release create <name> --project <id>
asgard-cli pipeline releases                    # created releases, and the ghost rows
asgard-cli pipeline variables list --release <name>
asgard-cli pipeline variables set --release <name> --kind secret <key> --from-file <path>
asgard-cli pipeline runs watch --release <name> --commit $(git rev-parse HEAD)
asgard-cli pipeline runs approve <run-id>
```

**These hold no rules of their own.** Whether a change is deployable is the
platform's answer: it renders the chart, checks every rendered CR against the
cluster's own CRDs with a server-side dry run, and reports back. That is not
reproducible here - no cluster credential is ever issued to a client - so the
loop is: change the chart, check what can be checked locally with `helm lint`
and `asgard-cli verify`, push, and read the plan back with `runs watch`.

Which release a command acts on comes from the name the declaration uses. Which
workspace, and which pipeline when a repository carries more than one, come from
`.asgard-cli.yaml` beside the declaration - written by `workspace use` and
`pipeline create`, and committed, so a clone and an agent both inherit it.

**The platform never reads that file.** A run reads the declaration at the
pipeline's config path and the chart it names, and nothing else, so nothing in
`.asgard-cli.yaml` can make a deployment succeed or fail. It exists so the
commands need no `--workspace`, and so an agent landing in a fresh clone can see
what the checkout is pointed at without a call. `--workspace` and
`ASGARD_WORKSPACE` outrank it, which is the safe direction: acting on a test
workspace when the customer's was meant costs a confusing error, and the reverse
deploys to a customer.

A secret's value can only be given with `--from-file` (or `--from-file -` for
standard input): a value typed as an argument is in the shell history and in the
process list. Files are read verbatim, so a PEM keeps its newlines.

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
