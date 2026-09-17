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
| `generated-repo-end-to-end` the tool run the way an engagement runs it | `init` through `verify` in a scratch repository, every generated kind rendered and the whole render validated against the CRD schemas | 2026-09-15 |
| `provenance-markers` every `**Checked:**` and `**Seen in:**` line in the corpus | read for what it actually claims: a count on one is not evidence anybody read the source, so the counts are gone and the scope is named instead. The commit each extract was read at lives in `source/SOURCES.md`, which `go run ./hack sources` holds against the clones | 2026-09-14 |
| `extracts-vs-charts` the extracts against the charts they came from | the commits in `source/SOURCES.md`'s **held against** column, which is what makes this reading checkable rather than a date | 2026-09-11 |
| `wiki-vs-docs` the wiki pages against asgard-docs | **the pages whose citations have moved, which `go run ./hack coverage --drift` names** - now 0, each citation carrying the commit it was held against; the prose citing a page that has not moved stands at its own earlier reading | 2026-09-14 |
| `packages-help` every package's help against its behaviour | the working tree | 2026-09-11 |
| `root-documents` Goal, AGENTS, APPROACH, STRUCTURE, TASK, README and its Chinese half - the files that state the rules, as opposed to the paths and commands in them, which `go run ./hack doc-paths` has | every pointer into the corpus opened and the sentence around it held against what that page says now; every Go symbol, command and CI step against the tree; the CEL counts against asgard-kube `cbd8d70` | 2026-09-15 |
| `needs-and-briefs` the needs lists and briefings - Go rather than markdown, and in no group until now | every platform claim against asgard-kube `cbd8d70`, every command and flag against the binary, and every row's pointer against the document it names | 2026-09-14 |
| `command-help` every `--help` screen - the largest reader-facing surface, and in no group until now | every screen read end to end; every command, flag and count in them run or recomputed against the binary and the CRDs | 2026-09-14 |
| `gate-messages` every error and warning string the checks print, and `hack/verify-references.sh` | the ones carrying a claim, against asgard-kube `cbd8d70` and the deployment clones; every reference chart rendered, including the third layout the script had never looked for | 2026-09-14 |
| `generator-templates` the CR skeletons `add` writes into a customer's chart, and the commented shapes beside them - which ship the same way, because a shape somebody copies by hand is a shape that runs | asgard-kube `cbd8d70` and asgard-core `623ceb5`; every kind generated in a scratch repository, rendered, and the render validated against the CRD schemas | 2026-09-15 |
| `indexes-and-counts` the two indexes, and every count in the material weighed against what a reader does with it | every alias term grepped against the landed tree; the counts that are a claim left to `go run ./hack counts` to recompute, the ones standing in for a yes replaced by a named example, and the rest deleted. The sweep reached every part of the corpus, the stage prompts, the scaffold templates, the design-time skills, every `--help` screen and the root documents. **What is left is check-held or load-bearing**: a figure `counts`, `tables`, `processors`, `coverage` or `shapes` recomputes, a platform quota a customer is sized against, or a number that is itself the argument. **The backlog closes and the regression is what needs a mechanism** - a document reviewed at a recorded digest does not come back until it changes, while a count added tomorrow is indistinguishable from a correct one, which is what `go run ./hack introduced` reads off the diff | 2026-09-14 |
| `frontmatter-descriptions` the `description:` each wiki page and extract declares about itself, which `go run ./hack index` renders into a row and holds against the frontmatter - never against the document | each description read against its own page, and against the heading of the group its row sits under: in `usecase/README.md` the heading poses the question and the rows answer it in the same terms, so a row that reads thin on its own can be the right one and rewriting it breaks the comparison | 2026-09-15 |
| `flag-usage` every flag's usage text against what the flag does | the working tree, every one | 2026-09-11 |
| `processors-vs-palette` `wiki/processors.md`'s prose, as opposed to its two tables, which `go run ./hack processors` now holds against asgard-core and asgard-docs | the two clones as pulled. **The palette is at second hand and stays there** - asgard-docs records it from a repository nothing here clones, which that page's `**Unchecked:**` marker names | 2026-09-11 |
| `stage-prompts` the stage prompts' guidance, as opposed to their command claims | read end to end; the platform claims they carry against asgard-kube `cbd8d70` and asgard-docs `23409b3`, and every command and flag they write run against the binary | 2026-09-14 |
| `design-time-skills` the design-time skills' prose, as opposed to their command and path claims | read end to end; every platform claim in them against the CRDs and asgard-core `623ceb5` | 2026-09-14 |
| `scaffold-templates` what a customer repository receives that is not a skill, `AGENTS.md.tmpl` above all | read; its platform field claims against asgard-kube `cbd8d70`, its gate table against the binary's own steps, its shape-C field forms against the CRD | 2026-09-14 |
| `deployment-diffs` the deployment clones' diffs since the extracts were written from them | the ones that had moved, which `go run ./hack sources --extracts` names; the rest were unmoved | 2026-09-11 |

## Where it stands

**The corpus is over 100,000 words** - wiki pages, extracts, guides, needs
lists and briefings, as `asgard-cli init` lands them.

**The chart half is the least finished of Goal's four points.** What `add`
writes is a correct starting point and not a chart: a large share of the spec
keys the widest reference chart uses are keys it does not write.

**The size is not written here.** `go run ./hack spec-key-gap` renders both
sides and prints it, and the check fails if the gap ever closes - because then
this paragraph is what is wrong.

**A key `add` does not write is in one of four states, and only one of them is
a gap.** Written; named in a commented skeleton, where somebody meets it at the
moment they would write one; absent on purpose, with the document that carries
that decision; or nowhere, which is the worklist. `--missing` prints the fourth
and lists the other two beside it, and **it is empty**: every spec key a
production chart uses has a home. What is left is the second state, which is
where most of the shape a chart needs now lives - a key that is a choice is
better shown than guessed, and a generator that guesses is worse than one that
shows the shape and leaves the decision.

**A decision is only as good as the document carrying it.** The third state is
a list in `hack/speckeys.go`, because no program recovers a judgement from a
CRD - but what is checked is the pointer, not the judgement: each row names the
document that says why the key is absent, and the check fails when that document
is gone or has stopped naming it, the same contract `needs` and `brief` hold.
Proved by deleting the row a document carries and watching every decision
resting on it go red.

**Most of a gap can be the measurement, and this one was three separate
measurement defects.** The check used to run one flag combination per kind, so
it counted as never written every key living behind a flag it did not pass - the
`--db-class` class blocks, `--private`'s git auth, `--supervisor`'s loop; close
to a third of the figure was artefact, and the figure is what somebody would
have worked from. It probes the combinations now, derived rather than listed:
the flags `add` registers, read off its cobra calls; per kind, the fields its
own templates **branch** on, read off the template parse tree rather than a
regex, because a field that is only interpolated changes a value and never a key
path; and for a flag with a closed vocabulary, that vocabulary from the
generator. Second, it could not see a comment, so every key taught where
somebody meets it counted as owed - the number said work was due exactly where
the work was done. What a comment can be read for is the field it names, not
the path it sits under: a skeleton here is written above the key it belongs to
rather than inside it, so reconstructing the path attributes it to the previous
sibling. That leaves one limit, and the decided set is what answers it - a
decision covers the fields under it, so a CR kind `add` never generates does
not pick up a generic leaf like `key` from another kind's skeleton. Third, it read a customer's own map key as a field: an
indexer called `idx-00001` was a key `add` "never writes", which is true and
unfixable, because a generator cannot know what a customer will call theirs.
Those segments collapse to `<key>` now, which keeps what hangs off them -
`docx.indexers.<key>.chunkSize` is schema and `docx.indexers.idx-00001.chunkSize`
is one chart's. **A number nobody can reproduce is worse than no number**, and
this one was being reproduced wrongly by the thing that printed it.

**The reading backlog is closed, and how it closed is the useful part.**
`go run ./hack reconcile` lists every pointer in the corpus nobody has recorded
a reading for - including each document's own index row, which is a pointer to
itself. **How much is left is not written here**: the command prints it, in the
two forms that are not one question - never recorded, and recorded against a
target that has moved since.

**It closes one document at a time and there is deliberately no way to close it
at once**, because that would record a claim nobody made. What that cost in
practice: reading passes run in parallel each move the targets the others were
reading, so the owed set does not fall monotonically - it went from everything
to a tail of several dozen and then converged only when one reader worked it
serially. **Parallel is faster at finding defects and slower at settling
them**, and a pass that fixes nothing settles immediately.

**The write-back path works, and what is missing is narrower than "a way
back".** The corpus is compiled into the binary - correctly, so a stale page is
fixed once for every engagement rather than rotting inside one - and
`issue-report` is the route out. Counting what has come through it settles the
question this paragraph used to ask: engagements file, most of them from
somebody who is not the maintainer, and they close.

**What is actually missing is a shape for a discovery.** The template is a bug
report, so something an engagement *learned* arrives as a complaint about the
tool, and the corpus has no slot for a claim only the engagement can verify -
`../usecase/conventions.md` had to invent "settled / very likely / reasoned" by
hand to say how far a thing was proved. `.out/write-back.md` carries the
argument and a recommendation: extend `issue-report` rather than build a second
route, and give `kb` a `**Proved:**` marker beside `**Checked:**` so an
engagement's claim lands where every reader meets it. **What would show that
wrong** is the next two engagements filing discoveries with those sections left
TODO - which would mean the problem was when a person is asked, not what they
are asked for.


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

Every document names the surface it has not been held against, and that command
prints every one of them. **Nothing is listed here** - a list that can be generated is not written
down, the same rule this file applies to a count, and the marker is on the page
where a reader meets the claim rather than in a file they have to think to open.

**Most of what it prints is not waiting on anybody.** A page says what it was
held against and what it was not, and the honest answer for a shape nobody has
deployed is that the first engagement to do it is its first test - which is a
thing to read before building one, not a task. The pages whose source is a UI screen
are the one group where somebody with an account could close the gap in an
afternoon.
