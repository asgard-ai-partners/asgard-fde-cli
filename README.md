# asgard-fde-cli

**English** | [繁體中文](README.zh-TW.md)

Command line tool for Asgard FDE (`asgard-cli`).

## Install

**This repository is private, and so is every release of it.** A GitHub release
takes the visibility of its repository: the page, the notes and every asset are
reachable only by an account with read access, and an unauthenticated request
for an asset URL gets a 404 rather than a 403. Nothing is published anywhere
else - no Homebrew tap, no Scoop bucket, no package repository (see
[Releasing](#releasing) for why).

So every install path needs credentials, and the shortest one uses the ones
`gh` already holds:

```bash
gh release download --repo asgard-ai-partners/asgard-fde-cli \
  --pattern '*_darwin_arm64.tar.gz' --output - | tar xz asgard-cli
sudo mv asgard-cli /usr/local/bin/
asgard-cli doctor          # says whether helm is on PATH
```

Swap the pattern for your platform - assets cover darwin / linux / windows on
amd64 / arm64, and Debian or RPM hosts can take the `.deb` / `.rpm` instead.
A file downloaded by a CLI is not quarantined by Gatekeeper the way a browser
download is, so no `xattr` step is needed.

If you already build Go, the module works directly once git can reach the
private repo:

```bash
GOPRIVATE=github.com/asgard-ai-partners/* \
  go install github.com/asgard-ai-partners/asgard-fde-cli/cmd/asgard-cli@latest
```

`asgard-cli version` reports what a release built; a `go build` with no ldflags
falls back to the module and VCS metadata rather than claiming a version it
does not have.

## Development

```bash
go build -o asgard-cli ./cmd/asgard-cli   # build
./asgard-cli version                      # run
```

Layout:

```
cmd/asgard-cli/     main; signal handling and exit codes only
internal/cli/       cobra command tree, one file per subcommand
internal/repo/      what a customer repository is made of, by looking at it
internal/auth/      the OAuth flow and the credential store, which is the only file
                    this CLI keeps outside a repository
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

To add a subcommand, write a `newXxxCmd()` in `internal/cli/` and register it
through `addTo(cmd, group..., ...)` in `root.go`. The group is required - cobra
panics on a `GroupID` the parent does not have - so a command cannot be added
without deciding where in the help it belongs.

- [STRUCTURE.md](STRUCTURE.md) - what every directory is for, including the four
  bodies of embedded material and which one a change belongs to.
- [AGENTS.md](AGENTS.md) - the conventions this repo follows, and what the gate is
  now that there is no test suite.
- [TASK.md](TASK.md) - what this repo is for, and what is not finished.

## Commands

### `init`

Write the repository skeleton here, so that a coding agent can take over.

```bash
mkdir acme-asgard-kube && cd acme-asgard-kube
asgard-cli init
```

```
This writes the Asgard repository skeleton into

    /path/to/acme-asgard-kube

and the repository will be called acme-asgard-kube, after that directory.

Write it here? [Y/n]

This is not a git repository yet. The skeleton expects one: the customer's
design-time credentials live in a .env that a .gitignore line keeps out of git.

Run `git init` here? [Y/n]

  created      .agents/skills/asgard-fde-onboarding/SKILL.md
  ...
45 created in /path/to/acme-asgard-kube

Now open this directory in your coding agent and say:

    Connect this repo to the Asgard platform
```

**This is the one command written for a person, and the only one that asks
questions.** Everything else here is written for a coding agent working in a
repository that already exists - and until this has run, that repository does
not: no `AGENTS.md`, no `CLAUDE.md`, no `.agents/skills/`. An agent opened in an
empty directory knows nothing about Asgard at all, which is why asking it to run
a command that needs a workspace id was circular.

**It touches no network and needs no account.** The skeleton is a fact about
this tool, not about any platform, so it can be written on a plane, before a
workspace exists, or before anybody has signed in. That is what makes it
possible to run first.

**Connecting the checkout to a platform is deliberately not part of it.**
Signing in, choosing a workspace, creating a pipeline and fetching the material
describing the server all come afterwards, guided by the agent this command just
equipped - which is a better guide than a list of six commands somebody follows
by hand. `asgard-cli gate` says what is still missing at any point.

Run it again whenever this CLI has moved on or a project was added: existing
files are left alone and reported as skipped. `--force` takes the newer shipped
material, discarding local edits to the skeleton; files this tool writes into -
the indexes, the open-questions table, the living spec - are preserved either
way and reported. `--yes` asks nothing, which is also what happens when stdin is
not a terminal, so a re-run from an agent or from CI needs no interaction.

It refuses to write into a home directory or a filesystem root. Forty-five files
one directory up from where they were meant is the mistake worth a guard.

**There is no separate `scaffold` command.** There used to be - it was this
without the platform steps, back when `init` had platform steps. Once `init`
stopped needing a session the two did the same thing, and two commands doing the
same thing is a question that gets asked.

It writes the part of a customer repo that is the same for every engagement:

| | |
|---|---|
| `AGENTS.md` | the platform contract, with the customer-specific sections marked TODO |
| `docs/` | the four-layer model (meeting-notes / decisions / living spec) and the SDD rules |
| `requirements/` | the task and request indexes |
| `.agents/skills/` | the seven design-time skills that hold for any Asgard, `db-query` among them; the ones describing a particular server come from `asgard-cli skill update` |
| `assets/` | the runtime-skill directory |
| `.asgard-pipeline.yaml` | the deployment declaration, with one release per project to fill in |
| `projects/<slug>/` | one chart skeleton per project |

What it does **not** write is the customer's own knowledge: which systems exist,
how the projects split, what the CRs look like. That is what the onboarding
produces, and no template can generate it.

The generated skeleton passes its own gate on the first run:

```bash
asgard-cli gate                               # everything this machine can check
```

**Do not run `helm lint` by hand.** The platform injects a reserved
`.Values.asgard` block into every render, and a chart must not declare it in its
own `values.yaml` - so a bare lint fails on every chart that reads
`.Values.asgard.projectEnvironmentId`, which is every chart that labels
anything. `gate` supplies that one file and nothing else.

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

**A chart does not always end with an entry point**, and no command here says
whether one is finished. A SemanticLayer with nothing mounted on it is either a
finished Mimir deliverable or an agent nobody has written yet, and the files
cannot tell the two apart. `asgard-cli size <shape>` lists what a shape is made
of, for a person to compare against; nothing records a chart's intended shape,
because a note of what somebody meant to build is not something this tool can
check.

### `project`, `request`, `task`, `question`

**Four commands read the repository back to you**, one file each, each with
`--format json`. None of them infers anything from the others.

```bash
asgard-cli question    # what nobody has answered yet, and who each is with
asgard-cli request     # what the customer asked for and is not done
asgard-cli task        # the task specs that are open
asgard-cli project     # what each chart declares
```

**Read `question` first.** The fastest way to do damage in a repository somebody
else started is to design past a question they already knew was open.

```
Projects:

  insight              DataConnector, SemanticLayer
  helpdesk             chart is empty
```

**It says what each chart HAS and nothing about what it lacks.** That used to be
measured against a "shape" recorded per project, and reporting "this shape asks
for X and X is absent" meant treating somebody's note of intent as a
specification. The list itself is the repository - the chart paths the
declaration names, and the directories under `projects/` - so there is no second
copy of it to drift.

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
                      asgard-cli verify
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
all.**

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
warn   assets/skills/ has no skill directories yet
ok  structure is consistent (1 project(s): [erp])
```

- the root README's project table matches the directories under `projects/`
- `.asgard-pipeline.yaml` parses, names no release twice, and every release it
  declares points at a chart directory that has a `Chart.yaml`
- runtime skills under `assets/skills/` carry `name` and `description`
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
asgard-cli render internal-dev         # manifests to stdout, summary to stderr
asgard-cli verify                      # render every release, check the invariants
asgard-cli verify --rendered file.yaml # check a stream that is already rendered
asgard-cli verify --format json        # one record per render, each check named
asgard-cli doctor                      # which external tools are here, and how to get them
```

`render` takes a **release**, not a project and an environment. Where a chart
deploys is a release in `.asgard-pipeline.yaml`, and one chart can have several;
`asgard-cli pipeline releases` lists the ones the platform has.

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
implementations over the same rendered chart: same findings, same counts.

**What used to be step 4 is not a local step any more.** It ran `kubectl` and a
`check_crd_fidelity.py` against a cluster; both moved to the platform's plan at
the Pipeline cut-over, because the checks worth the most - the apiserver's own
CEL, pattern and required validation, and the unknown-field pruning a dry run
hides - need a cluster, and **no client is ever issued credentials for one**.
The local half is `asgard-cli gate`; the authority is the plan report.

### helm and kubectl are prerequisites, not dependencies

Nothing about how asgard-cli is distributed can install them. A tar.gz, a zip and
`go install` carry no dependency metadata and never can, and a dependency
declared on a Homebrew tap or a Scoop bucket would only cover people who install
that way - which is nobody, because those channels are deliberately off for a
private repository (see [Releasing](#private-and-the-channels-that-off-follows-from)).
Declaring one anyway would read as a guarantee that does not hold.

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

`init`, `project`, `request`, `task`, `question`, `decision`
and `check` need none of these tools. `render`, `verify` and the three chart
steps of `gate` need helm. **kubectl is optional**: nothing in this binary talks
to a cluster, and `doctor` lists it because a person debugging a deployment
still wants to know whether it is there.

### The files

Four, and each is somebody else's answer to a different question.

| file | who writes it | who reads it | committed |
|---|---|---|---|
| `.asgard-pipeline.yaml` | a person | **the platform**, on every run | yes |
| `.asgard-cli.yaml` | `asgard-cli` | `asgard-cli` only | yes |
| `os.UserConfigDir()/asgard-cli/credentials.json` | `asgard-cli login` | `asgard-cli` | **never** |
| `os.UserConfigDir()/asgard-cli/profiles.json` | `asgard-cli profile set` | `asgard-cli` | **never** (but it can be handed to a colleague) |

**`.asgard-pipeline.yaml` is the declaration**, and the only file a deployment
depends on: which releases exist, which chart each deploys, what triggers it,
which keys it takes. See `internal/pipelineconfig`.

**`.asgard-cli.yaml` is the binding**: which workspace and which pipeline this
checkout acts on, and nothing else. Both fields are required and neither is
derived. See `internal/binding`, whose package comment explains why it lives
beside the declaration rather than at the repository root.

**`profiles.json` names a platform this binary does not have compiled in** -
an on-prem deployment, or a stack running locally. It holds no secret and it is
optional: with no file, every profile is the hosted platform. See
[`profile`](#profile).

**`credentials.json` is the secret this CLI keeps outside a repository.**
0600, one file for every profile, and nothing beside it. There was a
`config.json` there too, holding a default profile, a default workspace per
profile and a map of custom profiles; it is gone. Every field was a preference
some flag or environment variable already expressed, and every one of them was a
thing an upgrade had to keep understanding. A credential is the one thing that
genuinely has to live there: it is a secret, it is per-person rather than
per-repository, and it cannot be re-derived.

A leftover `config.json` is **an error, not a warning**. The retired
`defaultProfile` was usually `dev`, so ignoring the file silently would move
every command to `prod` - a customer's platform. The first command that resolves
a profile refuses instead, names each retired key and what replaces it, and says
to delete the file.

**What is deliberately not stored anywhere:** which projects the repository has
(read off the declaration's chart paths and `projects/*/`), the customer's
display name (asked of the platform when a template needs it), and what a chart
is "meant to be" (a claim about intent no tool can check). The rule they each
failed: *a value belongs in a config file only when nothing on disk implies it
and the platform cannot be asked.* See
`docs/decisions/2026-09-05-asgard-cli-config-surface.md` in `asgard-odin-pm`.

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

A **profile** is one Asgard installation. `--profile` picks per command and
`ASGARD_PROFILE` sets it for a shell; with neither it is `default`, which is the
hosted platform. **Nothing records a current profile**, and `login --set-default`
used to: a preference on one machine is a preference two people running the same
command do not share.

Profiles other than the hosted one are written with
[`asgard-cli profile`](#profile). `ASGARD_PLATFORM_API`, `ASGARD_ISSUER` and
`ASGARD_CLIENT_ID` still override one field at a time on top of whichever
profile applies, for a one-off.

**`ASGARD_PLATFORM_API` is named for the service, not for "the API".** This tool
talks to one Asgard service today and is expected to grow into others - the
Control Center API is the next one - so the general word is not spent on
whichever arrived first. `ASGARD_ISSUER`, `ASGARD_CLIENT_ID` and `ASGARD_TOKEN`
stay general on purpose: every Asgard service authenticates against the same
Casdoor and accepts the same token, so those three genuinely are about all of
them. It was called `ASGARD_API` in v0.1.0; a shell that still exports that name
is refused with the rename rather than quietly ignored, because ignoring it
would send every command to the built-in prod URL.

With no browser - CI, a container, an agent sandbox - set `ASGARD_TOKEN` to an
access token instead. It bypasses the store completely, reading nothing from
disk and writing nothing to it.

### `profile`

**If you use the hosted Asgard platform, you need none of this.** With no file
at all, every command reaches it - that is what `default` means, and why it is
the default. These commands exist for the two installations this binary cannot
know about: an on-prem deployment, and a stack running locally.

```bash
asgard-cli profile list              # what is configured, and what applies now
asgard-cli profile show [name]       # the three values, and where each came from
asgard-cli profile set onprem --platform-api https://asgard.acme.internal \
    --issuer https://iam.acme.internal --client-id abc123
asgard-cli profile remove onprem
```

A profile holds three values and **each falls back on its own** to the hosted
platform's:

| | |
|---|---|
| `--platform-api` | where the Asgard Platform API is |
| `--issuer` | the Casdoor that issues tokens for it |
| `--client-id` | the application this CLI presents itself as (not a secret - it is disclosed to the browser on every sign-in) |

**An on-prem installation sets all three.** Its API and the Casdoor that issues
tokens for it are the same deployment, and a token from one is not accepted by
the other - so setting the API alone leaves you signing in against the hosted
Casdoor and presenting that token to somebody else's server. `profile show`
prints where every value came from and warns when the two disagree:

```
profile        onprem
platform api   https://asgard.acme.internal    profiles.json
issuer         https://iam.asgard-ai.com       the hosted platform (this profile does not set it)
client id      r21ntx0eb5igyokl3px4            the hosted platform (this profile does not set it)

WARNING: profile "onprem" takes its Platform API from profiles.json and its
identity provider from the hosted platform ...
```

It warns rather than refuses, because a local Platform API against a real
Casdoor is a legitimate way to develop.

**There is no `profile use`.** Recording which profile is current is the one
field of the retired `config.json` that is not coming back - it was invisible on
the machine that had it and absent on every other. Which profile is `--profile`,
`ASGARD_PROFILE`, or `default`.

`profile set` is the only command that creates `profiles.json`, and only when
run. Nothing writes it as a side effect. It holds no secret, so it can be handed
to a colleague setting up the same installation; credentials are a separate file
and cannot.

#### Working against our own development platform

`dev` is not a built-in name. Our development platform is one installation among
the ones this tool meets, not a second kind of thing, and compiling it in would
put an internal endpoint in every customer's binary.

Write it like any other, with the three values from the internal setup notes -
**they are not in this repository**:

```bash
asgard-cli profile set dev \
    --issuer       <internal>  \
    --client-id    <internal>  \
    --platform-api <internal>
asgard-cli login --profile dev
export ASGARD_PROFILE=dev        # for a shell
```

### `workspace`, `pipeline use`

Which workspace and which pipeline this checkout acts on - the two facts nothing
in the repository implies and the platform cannot be asked on your behalf.

```bash
asgard-cli workspace list            # what this account can reach
asgard-cli workspace use <id>        # record the workspace for this checkout
asgard-cli pipeline list             # what that workspace has
asgard-cli pipeline use <id>         # record the pipeline
asgard-cli workspace show            # which apply here, and why
```

Both land in `.asgard-cli.yaml`, beside the declaration they belong to, and
**that file is committed**: whoever clones the repository, and whatever agent
works in it, then needs no flags. The platform never reads it.

**Nothing is ever inferred, including from a list of one.** A command with
nothing recorded lists the candidates and refuses. A rule that resolves while a
list holds one entry starts resolving to something nobody chose on the day it
holds two, and nobody is watching that day.

**No git remote is read as identity.** The pipeline used to be found by matching
the checkout's `origin` against the workspace's pipelines; a repository may have
any number of remotes, and which one carries that name is its owner's business.
The gap that leaves - a repository copied wholesale into another repository of
the *same* workspace keeps a pipeline id that still resolves - is stated in
`.asgard-cli.yaml`'s own header rather than guarded by a rule that fires on the
wrong input. Run `pipeline use` after copying a repository.

**`workspace use` clears the pipeline line** when the workspace changes, because
a pipeline belongs to one workspace. Every pipeline command then refuses and
names the remedy, and `asgard-cli gate`'s `binding` step goes red - which is the
failure landing on the next command instead of on whichever later one happened
to be destructive. Commit both lines together.

Resolution order, highest first: `--workspace`, `ASGARD_WORKSPACE`, the
checkout's `.asgard-cli.yaml`. That is the whole list. `workspace show` reports
*why* that workspace, which is the useful half: acting in the wrong one is the
failure the order exists to prevent.

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
asgard-cli pipeline runs watch --release <name> --ref <tag>
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

### Private, and the channels that off follows from

Every release is private because the repository is. That is not a limitation to
work around while the audience is internal - it is the point - but it does
decide the install path, which is why the release notes carry a
`gh release download` line rather than a `brew install` one.

The bottom of `.goreleaser.yaml` has ready-made **Homebrew tap** and **Scoop
bucket** blocks, and they stay commented out. A tap or a bucket is a second
repository that whoever installs has to be able to read; making that one private
too means every user runs `brew tap` against a repo needing credentials, which
is more setup than the one-line download it would replace, for a smaller
audience than a tap exists to serve. Audience decided internal-only, 2026-09-06.
**If this ever goes public, enabling them is the first thing to revisit** - the
blocks and their prerequisites (`HOMEBREW_TAP_TOKEN`, `SCOOP_BUCKET_TOKEN`) are
left in place for that.

Other things worth knowing:

- **CGO**: builds run with `CGO_ENABLED=0` so cross-compilation fits on a single
  runner. Pulling in a cgo dependency (sqlite and friends) means switching to
  zig cc or per-platform runners.
- **macOS signing**: the binaries are unsigned. That is survivable *because* the
  install path is a CLI download - `gh` and `curl` do not set the
  `com.apple.quarantine` attribute that makes Gatekeeper refuse an unsigned
  binary; a browser does. Handing somebody a release URL to click is the case
  that breaks, and `anchore/quill` is the answer if that ever becomes the normal
  way in.
- **Every PR builds a release.** `ci.yml`'s `build` job runs
  `goreleaser release --snapshot --clean --skip=publish`, so a config or
  cross-compilation break is caught before it is a failed tag.
