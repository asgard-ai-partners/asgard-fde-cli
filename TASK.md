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

**Two are not checks and have no pass or fail.** `--orphans`, `--crossref`,
`--ask`, `--unmarked` and `audit-material` with no flag are listings for a
person to read, and `--term <field>` is a query: it answers a question somebody
asks it after renaming a platform field, and it is in
`.agents/skills/consistency-checks/SKILL.md` for that reason rather than here.

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


## Goal

**`asgard-cli` is what an FDE's agent asks about integrating with Asgard.** That
is the positioning, and four things follow from it - `Goal.md` is the short
version and this is the same list with what each one costs:

  1. **Give an agent the knowledge of the Asgard ecosystem.** What the platform
     has, which CR a UI name maps to, how one shape of deployment is assembled
     field by field, and where each of those has been got wrong before.

  2. **Be useful in a meeting**, which means naming the dependencies an
     integration scenario needs before somebody promises it. A credential, an
     endpoint, a network path, a test environment, an approver's queue - the
     interview in `.agents/skills/asgard-platform/guide/requirements.md` and the briefings in
     `.agents/skills/asgard-platform/brief/` are this, and the riskiest activity in an engagement is
     the one that leaves no trace in a repository.

  3. **Implement the IaC - the charts.** Skeletons for the CR kinds, the
     invariants a rendered chart has to hold, and the judgement that goes with
     both. **This is the least finished of the four**: the widest reference
     chart uses 185 spec keys and `add` never mentions 88 of them, so what it
     writes is a correct starting point and not a chart.

     **That pair read 168 and 52 for a week and no method reproduced it**, which
     is worse than being wrong by a little: it is the number that decides whether
     an FDE treats what `add` emits as a chart. `hack/spec-key-gap.py` renders
     both sides, states the method - a dotted path under `spec` with list
     indices collapsed - and fails when this line and the measurement disagree.
     Across all 19 reference charts together it is 303 keys and 171 never
     written.

  4. **When the knowledge is not here, say where to file it.**
     `asgard-cli issue-report` prints the repository URL and what a report has
     to say; `--new` writes the body with what the tool already knows filled
     in. The corpus is compiled into the binary, so an engagement cannot write
     what it learns where it will be read - the issue is the only path back,
     and it has to be two sentences rather than an essay.

**The asking is the spine and the chart work hangs off it.** An FDE does not
reach for this to be told what step they are on; they reach for it mid-sentence,
in a meeting or halfway through a chart, because their agent needs a fact about
the platform that only this tool has. So the first one has to work **with no
repository at all** - the question asked in a first meeting is the same question
asked halfway through a chart, and an FDE who must be inside an engagement to
ask it will not ask it.

### Which is why the knowledge base is the product

The value here is not the command surface. It is **70 documents and 99,000
words** that exist nowhere else - 27 wiki pages, 22 extracts, 10 guides, 7 needs
lists and 4 briefings, as `asgard-cli init` lands them - and **the whole of the
engineering problem is making them searchable by an agent.**

It said 68 and 87,000 until 2026-09-11, when both were recounted off a landed
tree rather than carried forward. `hack/check-goal.py` counts them there now,
whitespace-separated and rounded to the nearest thousand, because the tree it
already builds is the only place this number is true of anything.

That is a different requirement from making them readable. An agent finds a
document by following a pointer or by matching a term, reads what it is given,
and acts - it does not browse an index, does not notice that the page it needed
was one directory away, and cannot tell that the answer it got is the wrong
sense of the word it asked about.

**The four rules below are that requirement, and each is a way an agent silently
gets the wrong answer rather than a matter of hygiene.** Each has a command that
says whether it still holds, so a rule that stops holding shows up in a run
rather than in a list somebody maintains.

### The form: llm-wiki

The target shape is the [llm-wiki
pattern](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f), and
it is chosen for a reason rather than adopted: **synthesis happens once, into a
document, instead of on every query.** An answer that needed three documents
assembled on the spot is new knowledge and goes into a document, or the next
reader assembles it again. No embeddings and no vector index - a retrieval layer
over un-synthesised material is the thing the pattern replaces.

What that costs and what it does not buy is "The design of record" below.

### What this tool does not do

It does not write the customer's Kubernetes resources. The end state of an
engagement is still a repository shaped like
[`unitech-e-asgard-kube`](https://github.com/asgard-ai-platform/unitech-e-asgard-kube),
and roughly 90% of that repo's 916-line `AGENTS.md` is knowledge only the
engagement can earn.

**And it reports no position.** No command says where an engagement is.
Guidance is reached by name with `guide` or by grepping `guide/` for the
subject; the records are read by `project`, `question`, `request` and `task`,
one file each.

## The design of record: one corpus

Every body of material is under one schema, in the form named above. How that
is implemented is [APPROACH.md](APPROACH.md); the rules a document has to follow
are [AGENTS.md](AGENTS.md). What is below is what the pattern buys and what it
does not.

### Three layers

What separates them is which one may be rewritten.

| layer | contents | may be edited |
|---|---|---|
| raw sources | asgard-docs, asgard-kube, asgard-core, the reference deployments | read-only, never vendored in; only the commit is recorded |
| the corpus | `corpus/wiki/`, `corpus/usecase/`, `stage/prompts/`, the scaffolded skills | rewritten continuously, and only ever describes the present |
| the schema | `AGENTS.md`, and each corpus's own `README.md` | changed deliberately, by a person |

### Three operations

**ingest** a source: read it, confirm the reading, write or rewrite the pages it
touches, update the index, and put the commit you read it at in each page's own
source block. Step "the pages it touches", plural, is the one that gets skipped;
`internal/corpus/wiki/README.md` says why there is no separate ledger of
readings.

**query**: grep the corpus first, through `aliases.md` if the question did not
arrive in English. An answer that needed three documents assembled on the spot
is new knowledge - it goes into a document, or the next reader assembles it
again.

**lint**: contradictions, stale claims, orphans, dead pointers, and the fourth
kind that only happens here - upstream moved and the corpus did not. Nothing
inside the corpus can detect that one; only going back to the source can, which
is why every document carries its sources.

### Four rules, and where each one stands

Each is stated in AGENTS.md as a rule; what follows is how far the code holds
it.

  - **One schema.** Held. Every document opens with a title and a summary and
    carries `**Checked:**` and `**Unchecked:**`, and `audit-material
    --unverified` reports **0 of 63** across the four bodies where the marker
    is written per document. It says so differently for `needs` and `brief`,
    whose eleven documents share one provenance constant, because a check that
    cannot fail should not report a pass. `kb.Ref` and `kb.ParseDoc` are what
    let a body join without being one `dir/name.md` per document - a numbered
    prompt file, a skill directory with YAML frontmatter.

    **It was closed by checking, not by writing the lines**, which is the whole
    of the distinction: the guidance and the skills were held against
    asgard-kube `cbd8d70`, against the eight reference repositories, and against
    the gate that enforces the rules they describe. A marker is a claim about a
    reading, so writing one without doing the reading is the only way to make
    this check lie.

  - **Links are data.** Held. Every document carries `kb.Doc.Links`, read when
    it is parsed, with the ones inside its counterpart section marked
    deliberate. Resolving a pointer at the point of use instead would be a
    second regular expression that can disagree with the first about what a
    document points at. The lint that was impossible is now
    `audit-material --orphans`, and **0 of 67 documents are reached by no
    pointer**. The index deliberately does not count as one.

  - **Retrieval is by subject, never by position.** Held. Nothing raises a
    document at a reader: there is no command that derives where an engagement
    stands, because an onboarding is not linear and a position cannot be
    argued with. `guide` names a piece of guidance; grep reaches any document
    by subject.

  - **Everything an agent reads has a parseable form.** Held. `project`,
    `question`, `request`, `task`, `check` and `verify` take `--format json`,
    and the last two are the pair that mattered most: they are the gate an
    agent is trying to turn green, and in text a warning and a failure differ
    by one word at the left margin while only one is fatal. A JSON run that
    fails exits 1 and prints nothing to stderr, because the report already says
    it failed and a second account of it is a second source for one fact. The
    material has no format flag - it is whole documents on disk.

### What the pattern does not buy, and what is unsolved

**No embeddings, no vector index, and now no search command either.** The point
of the pattern is that synthesis happens once, into a document, instead of on
every query. `asgard-cli init` writes the documents into the repository and
grep is the way in, which is one fewer thing to keep working than a scoring
function was.

**The write-back path runs through a person, and the person is at the far end.**
In the original, a good answer becomes a new page. Here the corpus is compiled
into the binary - correctly, so that a stale document is fixed once for every
engagement rather than rotting inside one - so an engagement that learns
something cannot write it where it will be read. The loop that exists instead:

    a grep comes back empty     ->  the agent has the gap in front of it
    the agent files it          ->  `issue-report --new`, evidence filled in
    somebody ingests it here    ->  a page, or a row in the alias index
    the next release            ->  every engagement has it

**The slow step is the right one.** ingest's first instruction is "read it, and
confirm the reading with the person before writing", and the only human in this
loop stands exactly where a claim enters material that ships to everybody.

What makes it work is that **the report is not prose filled in from memory.**
The reader and the writer here are both agents, so every narrated field is
somebody's account of what happened and can be wrong - a search remembered as
run, phrased differently from the one that was run. So `--new` fills in the two
things nobody has to be believed about: the exact build, and the repository
state or the fact that there is no repository. The four narrated fields are
marked TODO, and **"run outside a customer repository" is printed as a normal
answer** rather than an absence, because that is where the question gets asked.

**Nothing records the query.** A search that came back empty was logged once and
the mechanism is gone: a query is whatever words the customer used, so the file
was a customer's vocabulary sitting in a repository, and the agent that ran the
search already has the gap in front of it without being told.

## Non-goals

- **Not a workflow engine.** The tool reports what the repository contains and
  what is relevant to that; it does not sequence an engagement, and a command
  that can only be reached by having reached the one before it is a defect.
  Onboardings are not linear - three of the decisions in the engagement this was
  built from were made, built, and reversed.
- **Provisioning git.** Not `git init`, not adding a remote, not authenticating
  to one: Asgard is growing its own mechanism for provisioning a customer
  repository. This needed saying **in the prompts**, not just here - an agent
  that finds the directory is not a repository, and reads that a push is what
  deploys, offers to set one up on its own every time. Stages 1 and 8 tell it
  not to, and why.

  Reading a checkout is a different thing and is done: the origin remote is how
  a pipeline is matched without writing a platform id into the repository.
- **helm and kubectl as packaged dependencies.** Nothing about how this is
  distributed can install them: a tar.gz, a zip, a dmg and `go install` carry no
  dependency metadata and never can, and a Homebrew or Scoop dependency would only
  cover people installing that way - nobody, until those repositories exist. They
  are prerequisites, reported by `asgard-cli doctor` with the install line for the
  current machine.
- **Bundling helm or kubectl into the release.** Four platform/arch combinations
  at ~50MB each, and **kubectl has to stay within one minor of the cluster's API
  server**, so a pinned copy goes stale and is worse than none.
- **Validating `workspace.id` against the platform.** Waiting on the API. It is optional as of 2026-09-02, since nothing rendered reads it; `project add` and `check` say when it is still unset.
- **Reading or writing `platformMainEnvironmentId`.** Per project per env, and it
  only exists after tf-asgard has created the namespace, so it belongs to the
  generated repo's values files.
- **Any command that talks to a cluster.** No cluster credential is ever issued to
  a client, which is also why the checking a pipeline does cannot be done here:
  the apiserver's own CEL, pattern and required validation of a rendered CR needs
  the apiserver.
- **Reimplementing what the platform checks.** `pipeline` is a wrapper over the
  platform's API and holds no rules of its own. A second copy of a lint rule is a
  copy that disagrees with the server the first time either changes, and the
  local half of the loop is the native tools - `helm lint`, `helm template` - plus
  `verify`, which checks what a dry run passes and runtime still fails.

**The offline rule now has a boundary rather than being absolute.**
`asgard-cli init` writes the whole corpus with no network, no repository and no
login, and `size` and `guide` answer without one either. That has to stay true:
the question they answer is asked in a meeting, before there is an engagement
to log in to. `login` and
`pipeline` are the exception, and they are an exception the first half must never
acquire.

## Open questions

They are not carried here. Each lives where the person who can answer it will be
standing:

  - **The platform's**, with who to ask and what each blocks -
    `.agents/skills/asgard-platform/wiki/platform-unknowns.md`. Five of them are also written into
    every scaffolded `docs/open-questions.md`, because every engagement hits
    them.
  - **This engagement's** - `asgard-cli question`, which reads
    `docs/open-questions.md` back and reports who each is waiting on.
  - **The FDE's**, listed under "What is not done" below: they need a decision
    rather than work.

## What is not done

Everything here needs something this repository does not have. A line is here
because it cannot be done from a checkout, not because nobody got to it - what
could be done from one has been, and git log is the record.

**A console login and an afternoon.** `console`, `sindri`, `mimir`, `fehu` and
`settings` describe a UI, so their source is product documentation rather than a
chart, and they are checked less deeply than the extracts by nature. Each says
how far it got on its own `**Unchecked:**` line and `asgard-cli audit-material --unverified`
lists them. Two mechanical passes found nothing and a third was written and
thrown away for calling correct material wrong. One person with access could
settle all five.

**A cluster.** 41 of the CRDs' 231 enforced CEL rules are `self == oldSelf`, comparing a
proposal against the object already on it. A render is one object with no
history, so nothing offline can see them - `botProviderClass` is the one that
bites, and `.agents/skills/asgard-platform/usecase/chat-channel.md` documents it instead. The same
applies to `.agents/skills/asgard-platform/guide/verify.md` step 4: a step that cannot be run is not a
step that passed.

**A customer on a chat platform.** Every `BotProvider` across every reference
deployment is `generic`. `chat-channel`'s credential blocks were checked field by
field against the CRD and match, but nothing there has run, and its
`**Unchecked:**` line says so. The first customer on LINE is that page's first
test.

**A clone of asgard-docs, and one of asgard-kube, together.** Goal's first
point names "which CR a UI name maps to" as one of four things this material
owes an agent, and **there is one such table: `wiki/agents.md`, five rows,
the agent family only.** Every other mapping is stated in the prose of
whichever page discusses it - Data Source on `settings`, Drive on `knowledge`,
MCP Server on `tools` - which a grep for the UI name does reach, and which
nothing can check for completeness. A UI name with no CR stated anywhere is
invisible.

The table that would fix it cannot be written from a checkout: it needs the
UI's own vocabulary from asgard-docs held against the kinds in asgard-kube,
and inventing a row is worse than not having one. **Do not put it in
`aliases.md`** - that file's rule is that every row is a term somebody
actually searched for, and a bulk import of UI names is exactly the guess it
forbids.

**A clone of asgard-docs, and a reader for one slice of it.** 85 pages are
cited by no wiki page at `f00e0ee`, 81 at the clone's HEAD, and
`hack/check-coverage.py` computes both. Almost all of it is release notes and
site plans. **The one slice worth reading is `help-community/faq`**, because
nothing has checked whether the answers a customer gets there agree with what
this material tells an FDE to say - and a customer quoting their own
documentation back is the one disagreement that cannot be argued with.

**An answer from the platform team.** The unknowns are on
`.agents/skills/asgard-platform/wiki/platform-unknowns.md`, with who to ask and what each blocks. P12 is
the cheapest: three CRDs - `ImageGenerationModel`, `TranscriptionModel`,
`SourceSetEditorServer` - exist in the contract and appear in no documentation
and no material here, and nobody has asked whether they are meant to be reached
for.

**A decision from the FDE**, deferred deliberately: the Homebrew tap and Scoop
bucket, which are a repository and a secret away; Linux packaging beyond nfpm's
defaults; whether the helm major version should be pinned; and whether the EKS
cluster names are customer-specific.
