# TASK.md

What this repo is for, and what needs something a checkout does not have.
Nothing else - how it is built is [STRUCTURE.md](STRUCTURE.md), how to change it
is [AGENTS.md](AGENTS.md), what the commands do is [README.md](README.md), and
the defects it has produced and why nothing caught them are in
`source/FINDINGS.md`.

**There is one worklist here**, under "Landing the rest of the material", and it
is there because somebody asked for that work. Everything else follows the older
rule: what could be done from a checkout has been, git log is the record of it,
and what is left is under "What is not done" where every line names the thing it
is waiting for. A finding that a reader
needs lives on the document it concerns rather than here - `**Unchecked:**` on
the page, a row on `asgard-cli wiki platform-unknowns`, a rule in AGENTS.md.

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
     interview in `asgard-cli guide requirements` and the briefings in
     `asgard-cli brief` are this, and the riskiest activity in an engagement is
     the one that leaves no trace in a repository.

  3. **Implement the IaC - the charts.** Skeletons for the CR kinds, the
     invariants a rendered chart has to hold, and the judgement that goes with
     both. **This is the least finished of the four**: a production chart uses
     168 spec keys and `add` never mentions 52 of them, so what it writes is a
     correct starting point and not a chart.

  4. **When the knowledge is not here, say where to file it.** A search that
     came back empty is recorded, `asgard-cli reading --misses` reads it back,
     and `asgard-cli issue-report --new` writes the report with that evidence
     already in it. The corpus is compiled into the binary, so an engagement
     cannot write what it learns where it will be read - the issue is the only
     path back, and it has to be two sentences rather than an essay.

**The asking is the spine and the chart work hangs off it.** An FDE does not
reach for this to be told what step they are on; they reach for it mid-sentence,
in a meeting or halfway through a chart, because their agent needs a fact about
the platform that only this tool has. So the first one has to work **with no
repository at all** - the question asked in a first meeting is the same question
asked halfway through a chart, and an FDE who must be inside an engagement to
ask it will not ask it.

### Which is why the knowledge base is the product

The value here is not the command surface. It is 68 documents and 87,000 words
that exist nowhere else, and **the whole of the engineering problem is making
them searchable by an agent.**

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

What that costs and what it does not buy is the next section.

### What this tool does not do

It does not write the customer's Kubernetes resources. The end state of an
engagement is still a repository shaped like
[`unitech-e-asgard-kube`](https://github.com/asgard-ai-platform/unitech-e-asgard-kube),
and roughly 90% of that repo's 916-line `AGENTS.md` is knowledge only the
engagement can earn.

**And it reports no position.** No command says where an engagement is. Guidance
is reached by subject through `find` or by name through `guide`; the records are
read by `project`, `question`, `request` and `task`, one file each.

## The design of record: one corpus, four parts

The target shape is the [llm-wiki
pattern](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f).
`internal/wiki/` already implements it, deliberately and with the pattern named
in its own `README.md`. **The work is to bring the other three bodies of
material under the same schema, not to invent one.**

### Three layers

What separates them is which one may be rewritten.

| layer | contents | may be edited |
|---|---|---|
| raw sources | asgard-docs, asgard-kube, the eight reference deployments | read-only, never vendored in; only the commit is recorded |
| the corpus | `corpus/wiki/`, `corpus/usecase/`, `stage/prompts/`, the scaffolded skills | rewritten continuously, and only ever describes the present |
| the schema | `AGENTS.md`, and each corpus's own `README.md` | changed deliberately, by a person |

### Three operations

**ingest** a source: read it, confirm the reading, write or rewrite the pages it
touches, update the index, append to the log. Step "the pages it touches",
plural, is the one that gets skipped.

**query**: search the corpus first. An answer that needed three documents
assembled on the spot is new knowledge - it goes into a document, or the next
reader assembles it again.

**lint**: contradictions, stale claims, orphans, dead pointers, and the fourth
kind that only happens here - upstream moved and the corpus did not. Nothing
inside the corpus can detect that one; only going back to the source can, which
is why every document carries its sources.

### Four rules, and where each one stands

These are the acceptance criteria for the work in the worklist below. Each is
stated in AGENTS.md as a rule; what follows is how far the code holds it.

  - **One schema.** Held. Every document in all four bodies opens with a title
    and a summary and carries `**Checked:**` and `**Unchecked:**` -
    **0 of 25, 0 of 21, 0 of 12, 0 of 7** unverified. `kb.Scan` and `kb.Rank`
    are the one scoring implementation and `kb.Ref` lets a corpus address files
    that are not `dir/name.md`, which is how a prompt is reachable as
    `read-path` and a skill as `<name>/SKILL.md`.

    **It was closed by checking, not by writing the lines**, which is the whole
    of the distinction: the guidance and the skills were held against
    asgard-kube `15ded0f`, against the six reference repositories, and against
    the gate that enforces the rules they describe. What that pass produced is
    in "The provenance pass" below - four corrections, not four provenance
    lines.

  - **Links are data.** Held. Every document carries `kb.Doc.Links`, read when
    it is parsed, with the ones inside its counterpart section marked
    deliberate. `find` names a hit's counterpart off that field and
    `audit-material --links` resolves the same field, so the two can no longer
    disagree about what a document points at - which they could, being two
    regular expressions at two points of use. The lint that was impossible is
    now `audit-material --orphans`, and its first run said **11 of 61 documents
    are reached by no pointer**. The index deliberately does not count as one.

  - **Retrieval is by subject, never by position.** Held, and the rule is
    narrower than it was: it used to say "by subject or by condition", and the
    condition half was the ladder with the numbers off. Gone with it are
    `stage.Current`, `stage.Relevant`, the numbering and "stage 4 of 9" - and,
    since 2026-09-05, `stage.Gaps` too. It was arithmetic, subtracting what a
    chart declares from what its **declared shape** asked for, and it was sound
    arithmetic on an unsound input: the shape was a note of what somebody meant
    to build, which is not something this tool can check. There is nowhere to
    record one now. Nothing raises a document for a reader; `guide` names them
    and `find` reaches them by subject.

  - **Everything an agent reads has a parseable form.** Held. `find`,
    `project`, `question`, `request`, `task`, `check` and `verify` take
    `--format json`, and the last two are the pair that mattered most: they are
    the gate an agent is trying to turn green, and in text a warning and a
    failure differ by one word at the left margin while only one is fatal. A
    JSON run that fails exits 1 and prints nothing to stderr, because the report
    already says it failed and a second account of it is a second source for one
    fact. `wiki` and `usecase` stay prose - they are whole documents.

### What the pattern does not buy, and what is unsolved

**No embeddings, no vector index.** The point of the pattern is that synthesis
happens once, into a document, instead of on every query. A retrieval layer over
un-synthesised material is the thing it replaces. Substring and term matching in
`internal/kb` is sufficient and stays.

**The write-back path runs through a person, and the person is at the far end.**
In the original, a good answer becomes a new page. Here the corpus is compiled
into the binary - correctly, so that a stale document is fixed once for every
engagement rather than rotting inside one - so an engagement that learns
something cannot write it where it will be read. The loop that exists instead:

    a search comes back empty   ->  `find` records it, in docs/.find-misses
    the agent files it          ->  `issue-report --new`, with that evidence in
    somebody ingests it here    ->  a page, or a row in the alias index
    the next release            ->  every engagement has it

**The slow step is the right one.** ingest's first instruction is "read it, and
confirm the reading with the person before writing", and the only human in this
loop stands exactly where a claim enters material that ships to everybody.

What changed to make it work is that **the report stopped being prose to be
filled in from memory.** The reader and the writer here are both agents, so
every narrated field is somebody's account of what happened and can be wrong -
a search remembered as run, phrased differently from the one that was run. The
recorded miss is the one part nobody has to be believed about: the tool
witnessed it. `--new` puts that, the version and the repository state in; the
four narrated fields are marked TODO.

**The miss file is not committed.** A query is whatever words the customer used,
so unlike `docs/.reading-log` - which carries page names of this tool and
nothing else - it stays out of the repository, and the scaffold's `.gitignore`
says so. It only has to live long enough for the issue to be filed; what reaches
the next engagement is the fix in the next release.

## Landing the rest of the material

**The test is grep, and it is the only test.** `asgard-cli init` writes the wiki
and the extracts into `.agents/skills/asgard-platform/` so that an agent in a
customer repository can find a document by a word rather than by a subprocess
and a ranking pass. What else belongs there is decided by asking whether that
retrieval gets better - not by whether the material is valuable.

Three things follow from the test, and they rule as much out as in:

  - **Static, or it cannot be a file.** Anything rendered from the repository's
    own state freezes one moment into a committed file, and nothing detects that
    kind of staleness: it is not behind the binary, it is behind the directory
    next door.
  - **Found by a word, not by a name.** A document somebody is told to read is
    already reachable; one they arrive at carrying a term is what grep is for.
  - **Read while working in this repository.** Material for a meeting held before
    the repository exists is served by the binary, which is where it has to be.

`asgard-cli find` searches four parts of the material. **Two have landed** - the
wiki and the extracts - and the two below have not.

### 1. `needs` - do this first

Seven shapes, about thirty rows, each with what to ask, why, and the document
that owns the claim. **Fully static**, no repository dependency, and `Item`
already carries json tags, so serialising it is the small part.

It passes the test outright: `allowlist`, `read-only`, `test environment`,
`Channel Access Token` are words an FDE arrives with, and the answer is one row
plus its source. Today it is 125 lines of Go that no grep can reach, and it is
the command Goal's second point names.

### 2. `brief` - same shape, nearly as cheap

Four briefs, 28 recorded ways to get something wrong. **Static** - `Render`
takes a writer and no data.

It passes the test less cleanly, and the reason is worth keeping: a brief is read
**by name before an activity**, not found by a word, so landing it wins less than
`needs` does. What earns it a place is that two of the four - `write-chart` and
`handover` - are read while inside the repository, and a phrase like "what must
never appear on a customer's screen" is one somebody would grep for.

### 3. `guide` - the most valuable, and the only one that needs designing

Ten documents, 2313 lines: the interview and its order, how the work splits into
projects, each project's read path and entry point, where knowledge lives, deploy,
and adding a capability to something already live. It is the third kind of
material - `wiki` is what the platform has, `usecase` is what to put in a field,
and this is **which decision to make now and what it costs to change later**.

**It is already being pointed at from material that has landed**: eight pointers
across seven exported documents, six of them to `asgard-cli guide requirements`,
which is where filter 0 lives and which every `needs` row cites. Those resolve
through the binary and not on disk - a dead pointer in a greppable corpus, the
same class of defect as the missing index.

**It cannot be exported as it stands, and this is the real work.** 56 template
sites across the ten files, of two kinds:

    repository state   <<range .Projects>>, <<.RequestID>>, <<with .Requests>>,
                       <<if not (.Has "DataConnector")>>, <<.Status>>
    a placeholder      <<.SpecSlug>> in an example path

Dumped verbatim it ships files full of `<<range .Projects>>`; rendered, it commits
one moment's repository state. So each document has to be **split into its static
half and its live half** - the decision knowledge becomes a file, and "here is
what your repository currently has" stays a command.

Two of the ten need no split at all: `06-knowledge.md` and `07-verify.md` have
zero template sites and could land today.

**The live half is not `gate`'s to serve.** `gate` answers whether the repository
is in a state to go on; the inventory the prompts interpolate is three other
commands, and the tool already says so in its own words - the repo check's
message for a document naming the removed `status` reads: *where it meant "what
is still open", `asgard-cli question`, `asgard-cli request` and `asgard-cli
task`; where it meant "what does each chart declare and still lack",
`asgard-cli project`.* That mapping is the one the split should use, and a
landed document has to name the command rather than imply it.

### Ruled out, with the reason

  - **`gate` and `check` rule explanations.** Exported, the next rule change
    makes the file a lie, and nobody greps for a rule - it finds you, and the
    message it prints carries its own reasoning.
  - **`size`.** A calculator, not a document: `--databases 2 --queries 4`. Grep
    does no arithmetic, and its shape table is worthless without the sum.
  - **`generate`'s twelve CR templates.** 16 of the 21 extracts already carry a
    `## The skeleton`, and the extracts have landed. A second copy of the same
    fields, without the cautions attached to them, breaks one fact one home.
  - **Everything under Build, Check and Deploy.** Repository views and actions.

### The target state: static knowledge lives in the repository, and nothing reads it for you

**Stated by the FDE, and it is a removal rather than a reduction.** Everything
that is not dynamic is static knowledge, it **has to** be on disk in the
customer's repository, and it is read there - by grep, by an agent, by a person.
`wiki`, `usecase` and eventually `find` stop being commands. The binary's job
narrows to writing the material out and keeping it current.

Two kinds of thing stay dynamic, and only these two:

    the repository's own state    which projects exist, what each chart
                                  declares and lacks, what is still open
                                  -> `project`, `question`, `request`, `task`
    the platform's own contract   what THIS customer's server accepts, which
                                  can be several versions from this binary in
                                  either direction
                                  -> `skill status` / `skill update`

**Updating the landed material is the software's own update path**, not a
separate fetch: a new binary carries new pages, and `init` replaces the
directory because the stamp says the version moved. That mechanism exists -
`replaceCorpus` - so the update story for landed knowledge is "upgrade the CLI
and re-run init", with nothing to remember.

#### `find` is to be removed, and printing how to search is an acceptable end

The four things it does that a grep cannot do not survive as one category once
the material has landed. **Three of the four are driven by data that lands with
it**: `aliases.md` for translation, `glossary.md` for the senses.

    naming the counterpart   dissolves. Once pointers are paths, the
                             counterpart IS a path in the document and grep
                             has it for free
    translating the query    becomes an instruction - read aliases.md, then
                             grep. SKILL.md already says so
    recording a dead query   becomes an instruction - run issue-report when
                             the search finds nothing
    warning on a word with   has no equivalent. It fires on a SUCCESSFUL
    two senses here          search, which is the failure nothing else can
                             see, and a landed glossary only helps a reader
                             who thought to open it

So the one thing lost is a mechanism becoming an instruction, and **the
difference is whether it still works when nobody follows it.** That argument
proves too much if taken alone - by it, every command stays - and the counter is
that an instruction in a loaded skill is cheaper than a command and is followed
reasonably well.

**It is testable rather than arguable.** The payment mistake is on record:
`find payment` returned Fehu's billing to somebody asking about a customer's
payment gateway. Land the material, write the instruction, and see whether it
recurs.

**A `find` that only prints how to search is accepted as an end state** - and
worth being precise about, because such a thing is a document rather than a
command. If its output is guidance, that guidance is `SKILL.md`, which already
lands. The only thing the command form would add is reachability with no
repository, and `init` in an empty directory is already that.

#### What has to be true first

**1. A command-form pointer was chosen because it does not depend on the
layout.** `asgard-cli usecase write-path` means the same thing from anywhere. A
path does not, and the two trees disagree:

    from a wiki page to that extract
      before the move      ../../usecase/extracts/write-path.md
      as landed            ../usecase/write-path.md

Three ways out, and only the second reaches the target:

  - Rewrite them to paths at export. The landed pages then differ from the
    binary's copy, so nobody can diff the two, and the rewrite has its own
    failure modes.
  - **Make this repository's tree match the landed one** - one `corpus/` holding
    `wiki/` and `usecase/` - so a relative path is correct in both.
  - Teach the mapping in a generated root index and leave the pointers as
    commands. Cheapest, and it keeps a subprocess at the one place the landing
    was meant to remove it: the map.

**2. `--links`, `--orphans` and the counterpart are built on that pointer
form.** `linkRe` is the basis of two of the four acceptance rules, so changing
the form rebuilds that machinery - mechanical, and a mistake in it lapses
silently rather than failing. The order is protected: change the form first and
`--commands` reports every pointer left behind.

**3. Reading a whole document is not a blocker.** It was written here as one and
it is not: the full text lands, so `cat` reads it, and with no repository `init`
in an empty directory produces the same files. The cost is 51 files and a
skeleton to read one page - ergonomics, not a missing capability. `find` prints
excerpts only, and that stops mattering once the documents are on disk.

**The corpus stays embedded in the binary either way.** `init` needs it to write
anything out, so removing `find` removes a search implementation, not the
material.

The sequence, each step verifiable alone and none of them leaving the tool worse
if it stops there:

  1. ~~Move this repository's corpus to `corpus/{wiki,usecase}/`.~~ **Done** -
     `internal/corpus` holds the material now, in the layout a repository
     receives it, and `internal/wiki` and `internal/usecase` are the way in
     rather than the place. No document changed.
  2. ~~Convert the pointers to paths and rebuild `kb.Link` on them.~~ **Done** -
     121 pointers between the wiki and the extracts are now
     `../wiki/<page>.md` and `../usecase/<name>.md`, which resolve in this tree
     and in a repository alike. The 8 pointing at `guide` stay invocations
     until step 4 lands it, and `wiki log` stays one permanently. `kb.Link`
     records which form a pointer took, and `--links` fails a path whose
     target `init` does not write - the check exists because converting the
     `log` pointer produced a link that resolved here and went nowhere in a
     repository.
  3. ~~Generate the root `index.md`.~~ **Done** - built from the jobs that were
     actually written rather than from the corpus, so it cannot name a document
     the export skipped. It carries what the per-half indexes cannot: both
     halves in one place as paths, the shape a pointer takes, and what is
     deliberately not there. `needs` and `brief` get their rows when step 4
     lands them; until then they are in its "what is not here" table.
  4. Land `needs`, `brief`, and `guide`'s static half. **`needs` is done** -
     seven documents under `needs/`, one per shape so that a grep hit carries
     which shape it belongs to, and each extract now points at its own. Making
     them documents also fixed a claim the package had been making and not
     keeping: its comment said `--links` resolved every `From`, and `needs` was
     in no source at all, so a `From` naming an extract that does not exist
     passed with 0 dead. **`brief` is done too** - four documents under `brief/`, with the same gap
     found and closed: a `Where` naming a page that does not exist passed with
     0 dead before they were documents. **`guide` is done, and step 4 with it** - the ten
     stages land as `guide/<name>.md`, minus the paragraphs that render this
     repository's own state. The split is 10 paragraphs of 542, and three
     sentences had to be reworded in the source rather than dropped, because a
     state claim is not always a template action: "Projects exist but no
     DataConnector does" is prose, true only of the repository the command was
     run in. They now say which repository the stage is for, which reads
     correctly in both places.

     With all five kinds landing, **every document pointer in the material is a
     path** - 73 more converted, across the extracts, the pages and the stage
     prompts. `asgard-cli wiki log` is the only invocation left.
  5. Delete `wiki` and `usecase`. `--commands` confirms nothing still names
     them. **Split into three, because the deletion is not the hard part:**

     - ~~The two `--search` flags.~~ **Done** - superseded twice, and the help
       said so itself.
     - Convert the ~180 remaining references. **This is the work**, and it is
       not a rename: the right replacement depends on where the text ends up.
       Classified, and the landing half is **done** - `aliases.md`, the only
       one of the 21 that needed it. What is left is the three groups that do
       not land as they are: 22 in the scaffold templates, 61 in Go strings
       and help, and 27 in this repository's own documentation, where a path
       into a customer repository would be wrong.
     - Delete the commands.

     **Step 5 is coupled to step 6 and the sequence did not say so.** `find`
     prints `-> field level: asgard-cli usecase external-api`, so deleting the
     commands while `find` survives leaves the tool naming something it does
     not have. `find`'s counterpart output has to move to the path form in the
     same change.

     And `asgard-cli wiki log` goes with it. Log is the one page `init` does
     not write, so deleting the command makes it reachable only from this
     repository - which is where its reader already is, but the index sentence
     that points at it has to say so.
  6. Delete `find`, after the sense instruction has been given a release to be
     wrong in.

**Steps 1 and 2 were the ones to do early**, because between them they touch
every document in the corpus and so collide with any other edit to material that
changes most weeks. What is left does not: steps 3 and 4 add files, and 5 and 6
remove commands.

**Two forms coexist until step 4**, which is a state to get out of rather than a
design. `brief` and `guide` are invocations because a path to them would resolve
nowhere; landing them is what makes the form uniform.

### Three defects the first landing introduced, all fixed

**All three were found by running `asgard-cli gate` inside a scaffolded
repository** - the check that should have been run before that commit and was
not. Landing material into a customer repository makes it subject to that
repository's own checks, and nothing in this repository's audits sees that.

  - The exported `SKILL.md` named `asgard-cli scaffold`, which this build has no
    such command for - it is `init`. One word, and the exact failure `2bce648`
    was written about.
  - It also read "a newer **asgard-cli than** the one you are running", and the
    repo check takes `asgard-cli` plus the next word in prose as a command name.
    Prose naming the tool rather than invoking it now says "this CLI", which is
    the corpus's own convention.
  - **`wiki/log.md` is no longer landed.** It names the removed `asgard-cli
    status` twice in historical entries that are correct, so every customer
    repository carried two warnings for them. The wiki's own index already said
    why it does not belong: log is the provenance layer and *"an FDE looking for
    an answer should never land there"*. `Corpus.All` treats every unlisted
    document alike and these two are not alike - `index` is the map and has to
    travel, `log` is for whoever maintains this repository. `wiki.Landing` is
    that distinction; 51 files land rather than 52.

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

**The offline rule now has a boundary rather than being absolute.** `wiki`,
`usecase`, `find`, `brief`, `size` and `guide` answer with no network, no
repository and no login, and that has to stay true: the question they answer is
asked in a meeting, before there is an engagement to log in to. `login` and
`pipeline` are the exception, and they are an exception the first half must never
acquire.

## Open questions

They are not carried here. Each lives where the person who can answer it will be
standing:

  - **The platform's**, with who to ask and what each blocks -
    `asgard-cli wiki platform-unknowns`. Five of them are also written into
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
how far it got on its own `**Unchecked:**` line and `asgard-cli wiki --unverified`
lists them. Two mechanical passes found nothing and a third was written and
thrown away for calling correct material wrong. One person with access could
settle all five.

**A cluster.** Forty of the CRDs' 79 CEL rules are `self == oldSelf`, comparing a
proposal against the object already on it. A render is one object with no
history, so nothing offline can see them - `botProviderClass` is the one that
bites, and `asgard-cli usecase chat-channel` documents it instead. The same
applies to `asgard-cli guide verify` step 4: a step that cannot be run is not a
step that passed.

**A customer on a chat platform.** Every `BotProvider` across every reference
deployment is `generic`. `chat-channel`'s credential blocks were checked field by
field against the CRD and match, but nothing there has run, and its
`**Unchecked:**` line says so. The first customer on LINE is that page's first
test.

**A clone of asgard-docs.** 69 published pages are cited by no wiki page. Two
slices are worth reading and the rest is release notes and site plans:
`developer-reference/processor`, because `wiki processors` was written from
asgard-core's definitions and P10 says that list is demonstrably incomplete; and
`help-community/faq`, because nothing has checked whether the answers a customer
gets there agree with what this material tells an FDE to say.

**An answer from the platform team.** The unknowns are on
`asgard-cli wiki platform-unknowns`, with who to ask and what each blocks. P12 is
the cheapest: three CRDs - `ImageGenerationModel`, `TranscriptionModel`,
`SourceSetEditorServer` - exist in the contract and appear in no documentation
and no material here, and nobody has asked whether they are meant to be reached
for.

**A decision from the FDE**, deferred deliberately: the Homebrew tap and Scoop
bucket, which are a repository and a secret away; Linux packaging beyond nfpm's
defaults; whether the helm major version should be pinned; and whether the EKS
cluster names are customer-specific.
