# TASK.md

Where this repo stands, and what needs something a checkout does not have.
Nothing else - what it is for is [Goal.md](Goal.md), how the main capabilities
are implemented is [APPROACH.md](APPROACH.md), what lives in which directory is
[STRUCTURE.md](STRUCTURE.md), how to change it is [AGENTS.md](AGENTS.md), and
what the commands do is [README.md](README.md).

**There is one worklist here, and it is written rather than remembered.** "The
consistency pass" below is filled in before a pass runs, by
`.agents/skills/consistency-checks/SKILL.md`, and it is the scope of that pass.
Everything else follows the older rule: what could be done from a checkout has
been, git log is the record of it, and what is left is under "What is not done",
where every line names the thing it is waiting for. A finding a reader needs lives on
the document it concerns rather than here - `**Unchecked:**` on the page, a row
on `.agents/skills/asgard-platform/wiki/platform-unknowns.md`, a rule in
AGENTS.md.

## The consistency pass

**Every check this repository has is named here, and no state is recorded for
one.** `hack/check-pass-list.py` holds this list against the binary's own flags
and `hack/`'s own contents, so it cannot be missing a check - and a verdict
written into a table is a result copied from a script that can produce it, in
prose nothing can verify.

    hack/check-pass-list.py      this list is complete
    <run the check>              whether it passes

So: **the mechanical checks appear here only as names.** Their answer is their
exit code, today, not a word somebody typed.

| the checks | how to run them |
|---|---|
| `--links`, `--bare`, `--commands`, `--paths`, `--unverified`, `--sources`, `--urls` | `asgard-cli audit-material <flag>`, after building from the working tree |
| `hack/check-doc-paths.py`, `hack/check-coverage.py`, `hack/check-tables.py`, `hack/check-pass-list.py` | the ones that read this repository, and the coverage row's clone |
| `hack/sources.py`, `hack/check-processors.py`, `hack/check-counts.py`, `hack/spec-key-gap.py`, `hack/extract-crs.py`, `hack/validate-crs.py`, `hack/verify-references.sh` | the ones that need a clone of somebody else's repository |
| `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` | the compiler's half |
| `hack/check-goal.py` | **the capability, not the material.** Goal.md's four points held against the binary, in a temporary directory with no network, no account and no git repository |

**Some of them are not checks and have no pass or fail.** `--orphans`,
`--crossref`, `--ask`, `--unmarked`, `--unchecked` and `audit-material` with no
flag are listings for a person to read - `--unchecked` is the one that replaced
this file's hand-written list of what is blocked - and `--term <field>` is a
query: it answers a question somebody asks it after renaming a platform field,
and it is in `.agents/skills/consistency-checks/SKILL.md` for that reason rather
than here.

### What no check reaches

**This is the only part with a state, because it is the only part a program
cannot answer.** It records what each reading was held against, so that
`hack/sources.py` can say when a reading has gone behind - which is the one
thing about a reading that *is* checkable. **That a reading happened is nobody's
to verify but the person who claims it.**

| surface | read against | when |
|---|---|---|
| `extracts-vs-charts` the 22 extracts against the charts they came from | the eight commits in `source/SOURCES.md`'s **held against** column, which is what makes this reading checkable rather than a date | 2026-09-11 |
| `wiki-vs-docs` the 27 wiki pages against asgard-docs | **the pages whose citations have moved, which `hack/check-coverage.py --drift` names** - 5 of them, all read; the prose citing a page that has not moved stands at its own earlier reading | 2026-09-11 |
| `packages-help` the 14 packages' help against their behaviour | the working tree | 2026-09-11 |
| `flag-usage` every flag's usage text against what the flag does | the working tree, all 81 | 2026-09-11 |
| `processors-vs-palette` `wiki/processors.md`'s prose, as opposed to its two tables, which `hack/check-processors.py` now holds against asgard-core and asgard-docs | the two clones as pulled | 2026-09-11 |
| `stage-prompts` the 10 stage prompts' guidance, as opposed to their command claims | - | **never** |
| `design-time-skills` the 7 design-time skills' prose, as opposed to their command and path claims | - | **never** - 2,255 lines, of which `proposal-deck` is 1,232 |
| `deployment-diffs` the eight deployment clones' diffs since the extracts were written from them | the four that had moved, which `hack/sources.py --extracts` names; the other four were unmoved | 2026-09-11 |

## Where it stands

Two numbers, and each is here because a check reads this file for it.

**The corpus is 70 documents and over 100,000 words** - 27 wiki pages, 22 extracts,
10 guides, 7 needs lists and 4 briefings, as `asgard-cli init` lands them.
`hack/check-goal.py` counts them in the tree it builds, which is the only place
the figure is true of anything.

**The chart half is the least finished of Goal's four points.** The widest
reference chart uses 185 spec keys and `add` never mentions 88 of them; across
all nineteen it is 303 and 171. So what `add` writes is a correct starting point
and not a chart. `hack/spec-key-gap.py` renders both sides and `--missing` is
the worklist.

Everything else about what this repo is belongs to another file and is not
restated here: [Goal.md](Goal.md) is what it is for, [APPROACH.md](APPROACH.md)
is how the four rules are implemented, `internal/corpus/wiki/README.md` is the
llm-wiki pattern and the three operations, [AGENTS.md](AGENTS.md) is how to
change it.

**The write-back path is the one part of the pattern that is not solved.** The
corpus is compiled into the binary - correctly, so a stale page is fixed once
for every engagement rather than rotting inside one - so an engagement that
learns something cannot write it where it will be read. What exists instead is
`issue-report`, and the slow step in it is the right one: a claim entering
material that ships to everybody passes a person.


## Non-goals

Refused, not pending. Each is here so that "missing" and "decided against" are
not the same answer.

  - **Sequencing an engagement.** No command says which step you are on, and a
    command reachable only by having reached the one before it is a defect.
    Onboardings are not linear: three decisions in the engagement this was built
    from were made, built and reversed.
  - **Provisioning git** - not `git init`, not a remote, not authenticating to
    one. Asgard is growing its own mechanism. **This needed saying in the stage
    prompts rather than only here**, because an agent that finds the directory
    is not a repository, and reads that a push is what deploys, offers to set
    one up every time. Reading a checkout is different and is done: the origin
    remote is how a pipeline is matched without writing a platform id into the
    repository.
  - **Packaging or bundling helm and kubectl.** A tar.gz, a zip, a dmg and
    `go install` carry no dependency metadata and never can. Bundling is worse
    than a prerequisite: four platform/arch combinations at ~50MB, and
    **kubectl has to stay within one minor of the cluster's API server**, so a
    pinned copy goes stale. `asgard-cli doctor` reports them with the install
    line for the current machine.
  - **Anything that talks to a cluster.** No cluster credential is ever issued
    to a client - which is also why the apiserver's own CEL, pattern and
    required validation of a rendered CR cannot happen here.
  - **Reimplementing what the platform checks.** `pipeline` is a wrapper over
    its API and holds no rules of its own; a second copy of a lint rule
    disagrees with the server the first time either changes.
  - **Validating `workspace.id`**, waiting on the API, and optional since
    nothing rendered reads it; and **reading or writing
    `platformMainEnvironmentId`**, which exists only after tf-asgard has created
    the namespace and so belongs to the generated repo's values files.
  - **A Homebrew tap and a Scoop bucket**, decided against 2026-09-06 while the
    audience is internal. A tap is a **second repository that whoever installs
    has to be able to read**, and this one is private: making the tap private
    too means every user runs `brew tap` against a repo needing credentials,
    which is more setup than the `gh release download` line the release notes
    already give them, for a smaller audience than a tap exists to serve.
    `.goreleaser.yaml` carries the configuration commented out, in the
    `homebrew_casks` shape rather than the deprecated `brews` one, and the two
    steps it needs. **Going public is what makes this worth revisiting**, and it
    is then the first thing to.
  - **Pinning a helm major version.** Answered by warning instead: Homebrew and
    scoop both ship Helm 4 while the scaffolded gate and skills were written for
    3, and `template` and `lint` both still exist - so `doctor` reports the major
    as a note and says to check that CI uses the same one. A pin would fail an
    install that works.

**The offline rule has a boundary rather than being absolute.** `init` writes
the whole corpus with no network, no repository and no login, and `size` and
`guide` answer without one either - the question they answer is asked in a
meeting, before there is an engagement to log in to. `login` and `pipeline` are
the exception, and the first half must never acquire it.


## Open questions

Not carried here. Each lives where whoever can answer it will be standing: the
platform's on `.agents/skills/asgard-platform/wiki/platform-unknowns.md` with
who to ask and what each blocks, an engagement's in its own
`docs/open-questions.md` which `asgard-cli question` reads back, and the FDE's
under "What is not done" below.


## What is not done

**Not listed here. It is read out of the material:**

    asgard-cli audit-material --unchecked

Every document carries an `**Unchecked:**` marker naming the surface it has not
been held against, and that listing prints all 74 of them. **This section used
to be a hand-maintained copy of those markers**, and four of its entries
outlived the thing they described - two claimed they needed clones this machine
has, one named a decision that had been made and recorded, and one named
something that appears nowhere in this repository. A list that can be generated
is not written down, the same rule this file applies to a count.

What is left, when you read it, needs one of four things nobody here can
produce: **a console login** for the five pages whose source is a UI,
**a cluster** to watch one of the 40 immutable fields actually be refused,
**a first customer on a chat platform** - `usecase/chat-channel.md` says why it
is that page's first test - and **an answer from the platform team** on the
three CRDs in `wiki/platform-unknowns.md` P12 that have no documentation
anywhere.

**None of them blocks an engagement**, and two are close to closed: the field
list and the credential asks are written down now, so what a cluster and a
customer add is confirmation rather than knowledge. The one that can still cost
a day is LINE, because `botProviderClass` is immutable - choosing wrong is a new
BotProvider rather than an edit.
