# TASK.md

Where this repo stands, and what needs something a checkout does not have.
**Nothing else** - what it is for is [Goal.md](Goal.md), how the capabilities
are implemented is [APPROACH.md](APPROACH.md), what lives where is
[STRUCTURE.md](STRUCTURE.md), how to change it is [AGENTS.md](AGENTS.md), and
what the commands do is [README.md](README.md).

**What a pass cannot derive is written here rather than remembered**, which is
"What no check reaches" below and nothing else;
`.agents/skills/consistency-checks/SKILL.md` is the method. **A finding a reader
needs lives on the document it concerns, not here** - an `**Unchecked:**` marker
on the page, a row on
`.agents/skills/asgard-platform/wiki/platform-unknowns.md`, a rule in AGENTS.md.

## The consistency pass

**The list is not here.** It is derived from the binary's own flags, `hack/`'s
contents and the Go gate's subcommands, so it cannot be missing one:

    go run ./hack pass

**No check carries a verdict anywhere.** A word typed beside one is a result
copied out of something that can produce it; the answer is its exit code, today.
What follows is the only part of a pass a program cannot produce.

### What no check reaches

**The only part carrying a state, because it is the only part a program cannot
answer.** Each row records what the reading was held against rather than a
verdict: that a reading happened is nobody's to verify but whoever claims it,
and whether its source has moved since is what `go run ./hack sources` reports.

| surface | read against | when |
|---|---|---|
| `glossary-collisions` every word the glossary says has one meaning here, against the help screens and the pages | read 2026-09-14; no check is possible, because which sense a bare word carries needs a reader | 2026-09-14 |
| `provenance-markers` every `**Checked:**` and `**Seen in:**` line in the corpus | read for what it actually claims: a count on one is not evidence anybody read the source, so the counts are gone and the scope is named instead. The commit each extract was read at lives in `source/SOURCES.md`, which `go run ./hack sources` holds against the clones | 2026-09-14 |
| `extracts-vs-charts` the 22 extracts against the charts they came from | the eight commits in `source/SOURCES.md`'s **held against** column, which is what makes this reading checkable rather than a date | 2026-09-11 |
| `wiki-vs-docs` the 27 wiki pages against asgard-docs | **the pages whose citations have moved, which `go run ./hack coverage --drift` names** - now 0, each citation carrying the commit it was held against; the prose citing a page that has not moved stands at its own earlier reading | 2026-09-14 |
| `packages-help` the 14 packages' help against their behaviour | the working tree | 2026-09-11 |
| `root-documents` Goal, AGENTS, APPROACH, STRUCTURE, README and its Chinese half - the files that state the rules, as opposed to the paths and commands in them | all six read end to end against the binary, the tree and the CRDs, **including AGENTS' twelve questions and four contradictions**: each one's citable claim was re-checked, the four contradictions are all still described accurately, and the `await` count moved from five of five to seven of seven | 2026-09-14 |
| `needs-and-briefs` the 7 needs lists and 4 briefings - Go rather than markdown, and in no group until now | every platform claim against asgard-kube `cbd8d70`, every command and flag against the binary, and every row's pointer against the document it names | 2026-09-14 |
| `command-help` the 2,866 lines of `--help` across 77 screens - the largest reader-facing surface, and in no group until now | every screen read end to end; every command, flag and count in them run or recomputed against the binary and the CRDs | 2026-09-14 |
| `gate-messages` the 392 error and warning strings the checks print, and `hack/verify-references.sh` | the 33 carrying a claim, against asgard-kube `cbd8d70` and the deployment clones; every reference chart rendered, including the third layout the script had never looked for | 2026-09-14 |
| `generator-templates` the 1,100 lines of CR skeleton `add` writes into a customer's chart | every field and claim against asgard-kube `cbd8d70` and asgard-core `623ceb5`; each kind generated, rendered and verified in a scratch repository | 2026-09-14 |
| `indexes-and-counts` the two indexes, and every count in the material weighed against what a reader does with it | every alias term grepped against the landed tree; the counts that are a claim recomputed by `go run ./hack counts`, the ones standing in for a yes deleted in favour of a named example | 2026-09-14 |
| `flag-usage` every flag's usage text against what the flag does | the working tree, all 81 | 2026-09-11 |
| `processors-vs-palette` `wiki/processors.md`'s prose, as opposed to its two tables, which `go run ./hack processors` now holds against asgard-core and asgard-docs | the two clones as pulled | 2026-09-11 |
| `stage-prompts` the 10 stage prompts' guidance, as opposed to their command claims | all 2,355 lines read end to end; the platform claims they carry against asgard-kube `cbd8d70` and asgard-docs `23409b3`, and every command and flag they write run against the binary | 2026-09-14 |
| `design-time-skills` the 7 design-time skills' prose, as opposed to their command and path claims | read end to end, 2,255 lines; every platform claim in them against the CRDs and asgard-core `623ceb5` | 2026-09-14 |
| `scaffold-templates` the 2,783 lines a customer repository receives that are not skills - of which `AGENTS.md.tmpl` is 912 | read; its platform field claims against asgard-kube `cbd8d70`, its gate table against the binary's eight steps, its shape-C field forms against the CRD | 2026-09-14 |
| `deployment-diffs` the eight deployment clones' diffs since the extracts were written from them | the four that had moved, which `go run ./hack sources --extracts` names; the other four were unmoved | 2026-09-11 |

## Where it stands

Two numbers, and each is here because a check reads this file for it.

**The corpus is 70 documents and over 100,000 words** - 27 wiki pages, 22 extracts,
10 guides, 7 needs lists and 4 briefings, as `asgard-cli init` lands them.
`go run ./hack goal` counts them in the tree it builds, which is the only place
the figure is true of anything.

**The chart half is the least finished of Goal's four points.** The widest
reference chart uses 185 spec keys and `add` never mentions 88 of them; across
all nineteen it is 303 and 171. So what `add` writes is a correct starting point
and not a chart. `go run ./hack spec-key-gap` renders both sides and `--missing` is
the worklist.

**The write-back path is the one part of the design that is not solved.** The
corpus is compiled into the binary - correctly, so a stale page is fixed once
for every engagement rather than rotting inside one - so an engagement that
learns something cannot write it where it will be read. `issue-report` is the
way back, and its slow step is the right one: a claim entering material that
ships to everybody passes a person.


## Non-goals

Refused, not pending - so that "missing" and "decided against" are not the same
answer. **The reason lives where the refusal is implemented**; what is here is
the decision.

  - **Sequencing an engagement.** A command reachable only by having reached the
    one before it is a defect - onboardings are not linear.
  - **Provisioning git.** Not `git init`, not a remote, not authenticating to
    one; Asgard is growing its own mechanism. Stages 1 and 8 say so too, because
    an agent that finds no repository offers to make one otherwise. *Reading* a
    checkout is different and is done.
  - **Packaging or bundling helm and kubectl.** They are prerequisites, reported
    by `asgard-cli doctor`; a tar.gz, a zip and `go install` carry no dependency
    metadata and never can, and a bundled `kubectl` has to stay within one minor
    of the cluster's API server.
  - **Pinning a helm major.** `doctor` warns instead - `versionNote` in
    `internal/cli/doctor.go` says why a pin would fail an install that works.
  - **A Homebrew tap and a Scoop bucket**, while the audience is internal. A tap
    is a second repository whoever installs has to be able to read, and this one
    is private. `.goreleaser.yaml` carries the whole argument and the
    configuration, commented out; going public is what makes it worth revisiting.
  - **Anything that talks to a cluster.** No cluster credential is ever issued to
    a client, which is also why a rendered CR's CEL and pattern validation cannot
    happen here.
  - **Reimplementing what the platform checks.** `pipeline` wraps its API and
    holds no rules of its own; a second copy disagrees with the server the first
    time either changes.
  - **Validating `workspace.id`** (waiting on the API, and optional since nothing
    rendered reads it) and **reading or writing `platformMainEnvironmentId`**
    (it exists only after tf-asgard creates the namespace).

**The offline rule has a boundary rather than being absolute.** `init`, `size`
and `guide` answer with no network, no repository and no login, because the
question they answer is asked in a meeting. `login` and `pipeline` are the
exception, and the first half must never acquire it.

## Open questions

Not carried here. Each lives where whoever can answer it will be standing: the
platform's on `.agents/skills/asgard-platform/wiki/platform-unknowns.md`, with
who to ask and what each blocks, and an engagement's in its own
`docs/open-questions.md`, which `asgard-cli question` reads back.


## What is not done

    asgard-cli audit-material --unchecked

Every document names the surface it has not been held against, and that prints
all 74. **Nothing is listed here** - a list that can be generated is not written
down, the same rule this file applies to a count, and the marker is on the page
where a reader meets the claim rather than in a file they have to think to open.

**Most of what it prints is not waiting on anybody.** A page says what it was
held against and what it was not, and the honest answer for a shape nobody has
deployed is that the first engagement to do it is its first test - which is a
thing to read before building one, not a task. The five pages whose source is a
UI screen are the one group where somebody with an account could close the gap
in an afternoon.
