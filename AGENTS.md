# AGENTS.md

Rules for AI agents working in this repo. Human contributors follow the same ones.

## What this repo is

**[Goal.md](Goal.md) is what this tool is for, and it is the one to read first.**
Four points, and every one of them is a claim you can break by accident:

  1. **An agent can get the Asgard platform's knowledge** - what the platform
     has, which CR a UI name maps to, how one shape is assembled field by
     field, and where each has been got wrong before. **Offline, with no
     repository**, because the question is asked in a meeting and a tool that
     needs a directory first will not get asked.
  2. **It is useful in a meeting**, naming what has to be obtained from the
     customer before a shape can be built - a credential, an endpoint, a
     network path, a test environment, an approval queue.
  3. **It helps write the Helm charts**, and deliberately not the namespace,
     the environment id, or whether the thing deploys.
  4. **When the knowledge is missing, its own output says where to file that** -
     out of the tool, not worked out by the agent.

The four are not independent. Point 1 is the spine; 2 and 3 are the same
knowledge coming out in a meeting and in a chart; 4 is the way back in when it
is not there. **So the knowledge base is the product and the commands are not**,
which is why most of this file is about the material.

Read the four before changing anything that prints, because three of them are
about what an agent can reach and the fourth is about what the tool says when
it cannot.

It writes files into a *customer's* repository and never holds state of its
own. The division: `Goal.md` is the goal, [TASK.md](TASK.md) is where it stands
and what is missing, [README.md](README.md) is what the commands do,
[APPROACH.md](APPROACH.md) is how the main capabilities are implemented,
[STRUCTURE.md](STRUCTURE.md) is what lives in which directory, and this file is
how to change the code and the material without breaking it.

**Before saying anything here is correct, load
`.agents/skills/consistency-checks/SKILL.md`.** It is the method: what to run,
in what order, what each check is blind to, and how to do the part no check
does. `go run ./hack pass` is the list itself, derived rather than written down;
what a pass cannot derive is in `TASK.md` under "What no check reaches", and it
is written there **before** a pass runs, because a pass that discovers its own
scope as it goes finds a surface nobody listed and reports it as news.

**Most of the value is not code.** It is one corpus, and the first question when
adding anything is which part it belongs to:

| part | answers | lives in |
|---|---|---|
| wiki pages | what the platform is, and who each piece is for | `internal/corpus/wiki/` |
| usecase extracts | how one shape of deployment is assembled, field by field | `internal/corpus/usecase/` |
| needs | what to get from the customer before a shape can be built | `internal/needs/` |
| briefs | what has actually been got wrong before one activity | `internal/brief/` |
| stage prompts | what to weigh at one point in the work | `internal/stage/prompts/` |
| scaffold templates | the part of a customer repo that is the same every time, including the skills the customer's agent loads | `internal/scaffold/templates/` |

**All of those but the scaffold templates are written into a customer
repository**, under `.agents/skills/asgard-platform/`, by `asgard-cli init` -
the scaffold templates are the repository. So a change to any of them ships twice: into the binary,
and into every repository that runs `init` after it. `internal/scaffold/corpus.go`
is what writes them and `scaffold.replaceCorpus` is what replaces them when the
version moves.

`needs` and `brief` are Go rather than markdown because each row carries the
document that owns its claim, and that pointer is checked; the markdown is
rendered from them.

[STRUCTURE.md](STRUCTURE.md) walks every directory.

**Present tense, and no log.** A rule is stated as the rule. **What went wrong
before is not the justification** - it reads as evidence and functions as a
changelog, it grows without bound, and every reader pays for it in context.
What belongs on a page is the constraint and what it costs to break; what
happened to this repository belongs in `git log` and nowhere else. A platform
or customer failure is different: that is a fact about the platform, it is what
the material is for, and it carries its provenance.

**Nor in a commit message.** The subject says what changed, the body says why
the rule is what it is. Not a post-mortem.

**One fact, one home; everywhere else links.** A trap that belongs to a CR
template does not also get explained in a wiki page. When you are about to repeat
a paragraph, link instead - two copies drift, and the reader cannot tell which is
current.

**The general form: reference the source of truth, never restate it.** A
paragraph is the obvious copy; the expensive ones do not look like copies at
all, and every kind of drift this repository has had is one of them:

    a count in prose            the tree is the source. Adding one page forced
                                edits to "70 documents", "27 wiki pages" and
                                "22 extracts" in three files, none of which any
                                goal asks for
    a constraint pinned here    the CRD is the source, and it deletes rules as
                                readily as it adds them
    a list of what exists       the binary is the source - the checks, the
                                commands, the kinds
    a value copied into a       the platform is the source, and copying it
    chart                       needs a route to obtain it that may not exist
    the same constant in        one declaration, and the other package reads it
    two packages

**A checker that holds a copy is not a fix, it is the coupling made
compulsory.** `hack goal` required TASK.md to state a document count that
`Goal.md` never asks for; every page added then turned the check red and the
repair was to retype a number the check had just computed. What it holds now is
the claim - that every document the binary carries lands in a repository -
compared set against set, with no number written anywhere.

So before writing a number, a list, or a constant: **ask what owns it, and
whether this can point at that instead.** If nothing owns it, it is a judgement
and belongs in prose. If something does, prose gets the judgement and the
program gets the value.

The reading order is wiki, then extract: an extract assumes you already know the
platform has that shape. `asgard-cli add <kind>` prints both, in that order.

## The corpus is a wiki, and these are its rules

The shape is the [llm-wiki
pattern](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f) -
raw sources that are never vendored in, a corpus that is rewritten continuously,
and a schema a person changes deliberately. The [Open Knowledge
Format](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
arrived at the same shape from the other side and is worth reading for the
vocabulary. **It says of itself that it solves the format and nothing else** -
going stale, what else a change is owed, and whether what a document points at
is still there are outside it, and those are what the rules below and their
checks are; `Goal.md` says why that is the half worth having. Take its
frontmatter; do not take it as a reason to relax a rule below.
`internal/corpus/wiki/README.md`
states it for the wiki and is the longer version; the rules below apply to
every part, and each has a check that says so, so a rule that stops holding
shows up in `asgard-cli gate` rather than in a list somebody has to maintain.

**One schema.** A document opens with a `# ` title and a summary paragraph, and
carries `**Checked:**` / `**Unchecked:**` - what it has been held against, and
what it has not. `internal/kb` parses that and nothing else. A body of material
that needs its own reader and its own parse is one that drifts from the rest
with nothing mechanical noticing: if new material does not fit `kb.Corpus`,
change `kb`. `Docs` and `ParseDoc` are the two hooks for a body that is not one
`<name>.md` per document - a numbered prompt file, a skill directory with YAML
frontmatter - and both still come out as `kb.Doc`.
**Check: `audit-material --unverified`.**

**A document declares two things about itself, and the index is rendered from
them.** YAML frontmatter, at the very top, and nothing else goes in it:

    ---
    group: While building
    description: what each processor type takes, and the fields that decide behaviour
    ---
    # The processors, and the fields that decide behaviour

**Both are there because neither can be derived.** A grouping is the question a
section asks. A description is what somebody would come to the document FOR,
which is not what it opens by saying - `console.md`'s thesis is "The Console
does no business work", and what an index owes a reader is "the permission
layers, the pages that disagree with each other, Workspace settings". Rendering
the second from the first was tried and gives rows like "They are not two of the
same thing". Everything else in a row - the title, the path, the order - is
computed.

**So never edit an index by hand.** `go run ./hack index --write` renders
`wiki/index.md` and `usecase/README.md` from the documents; the check fails when
either has drifted. Adding a page is the page and two lines on it. What stays
written is which sections exist and in what order, and the order of rows inside
one - that is a reading order rather than an alphabet, it is read back out of
the file, and a new document is appended so that adding one moves nothing
somebody arranged.

**A description does not follow its page.** It is authored, so when the page
changes the row can quietly stop describing it - two did in one sitting. A
document points at itself in `source/reconciled.json` for exactly that, and
`go run ./hack reconcile` reports it.
**Check: `go run ./hack index`.**

**Every pointer is a path**, because every document lands:

    ../wiki/<page>.md      from a document inside one of the directories
    wiki/<page>.md         from `aliases.md` or `index.md`, at the landed root

`internal/corpus/` holds the material in the layout a repository receives it, so
one form resolves in both trees. It is written `../` even between two documents
in the same directory, because a pointer that carries its own kind beats two
forms that need to know where the reader is standing.

Two ways of writing one that the gate refuses, and both read perfectly to a
person:

- **an invocation.** The tool's name, a kind and a page name used to be how a
  document pointed at a document, and a document pointing at a document points
  at a file. It is not written here as an example, because `--commands` reads
  this file and would resolve it - which is the check working. A help screen is
  different: `guide` is a command, and naming it there is telling somebody to
  run it.
- **a name with no path.** `write-path` and `[`x`](x.md)` both resolve for
  whoever is reading right now and are checked by nothing, so a renamed page
  breaks them in silence, and the two indexes are where they accumulate,
  because an index is nothing but references.

**Check: `audit-material --links` and `--bare`.**

**Nothing points outward that a repository does not have.** These files land in
somebody else's checkout, where "this repo" is theirs and `source/SOURCES.md`
is not there. Citing a file in another repository is right - that is provenance
- so the rule is to **name the repository on the same line**.
**Check: `audit-material --paths`.**

**Guidance is retrieved by subject, never by position.** "You are at step 4" is
a claim about a walk no engagement performs, and a position cannot be argued
with. What a reader can act on is a claim about the repository in front of them
or about the thing they are about to do - which is what `brief/` is and why
`guide` renders against this repository. **Anything reachable only by having
arrived somewhere is unreachable**: an engagement read the stage it was told it
was in, and what that stage pointed at, and never opened the page it spent a day
needing. **Check: `audit-material --orphans`**, which is the same rule from the
other end.

**What an agent reads has to be parseable.** Column-aligned output is for the
FDE and is not an interface; a command an agent acts on owes it a form that does
not have to be recovered from `%-9s`. The material itself has no format flag,
because it is whole documents on disk.

## Language

**Go code, command output, error messages, command help, and every document
this repository ships are English.** That is the whole corpus, plus
`README.md`, `TASK.md`, `STRUCTURE.md`, `APPROACH.md`, this file and
`source/SOURCES.md`.

**Two documents here are Chinese on purpose.** `Goal.md` is the goal as its
author states it, and `README.zh-TW.md` is the Chinese half of the README. Both
are read by people rather than shipped to an engagement.

```go
// Save writes cfg to path.
return fmt.Errorf("write %s: %w", path, err)
```

**One body of material is deliberately Chinese**: parts of
`internal/scaffold/templates/`, because the generated repository is read by the
customer's own engagement, in their language.

**The corpus carries one language because the mapping lives somewhere else, not
because the reader translates.** `internal/corpus/aliases.md` is the mapping,
and it lands beside the material so it can be read before a grep rather than
after one comes back empty. **Translate the customer's words, then search.**
Translating only after a search failed was tried and was worse: the query
matched the alias table's own row about the word. Product labels keep their own
names (Managed Agent, Drive, Context Index are what the UI says).

**An index is not a page and does not live among them.** The alias table was a
section of the glossary until it was measured: because it lists every alias, it
was reliably the one document carrying every term of a translated query, so a
search for a subject returned the word list rather than the page. `aliases.md`
sits at `internal/corpus/`, above both halves, because it applies to both;
`internal/corpus/wiki/index.md` is the wiki's own catalogue and stays with the
pages it lists. Both are `Unlisted`, so `List()` hides them and `All()` does
not. `audit-material --links` reads them - their rows carry pointers - and
`--orphans` deliberately does not count them, because a list that names every
page makes every page look reached.

Do not "fix" the scaffold templates into English. Do not start a second exception
without saying why it earns one.

## `hack/` is Go

**A check that is not compiled is a check nobody runs until it is wrong.** These
scripts are the maintainer's gate and most of them run only when somebody runs
them by hand, so a typo in a branch that is rarely taken sits there for weeks -
and one pass added every one of these: a missing import, a key that did not
exist in the map it indexed, a `git grep` output parsed by the wrong field, a helper
called with the wrong signature, and a substring test that passed on a longer
word. **Every one of those is a compile error in Go**, and `go build ./...` and
`go vet ./...` already run on every push.

So a check goes in `hack/` as Go, under the one `main` package with a
subcommand each, and shares what it needs through `hack/internal/`. Three things
follow:

  - **No `yq`, and no subprocess for YAML.** `gopkg.in/yaml.v3` is already a
    dependency, so a CRD is unmarshalled rather than shelled out to and parsed
    back out of JSON.
  - **Reading the corpus goes through `internal/kb`**, not a second regular
    expression. The whole reason that package exists is that two readers of one
    format drift.
  - **`go test ./...` can reach them**, which nothing could before.

**A shell script is still a shell script** when what it does is drive other
programs - `verify-references.sh` renders charts with helm and pipes them into
the binary, and rewriting that in Go buys nothing.

## Plain ASCII, no emoji

No emoji, and no decorative Unicode either - no check marks, arrows, box-drawing
characters. They render inconsistently across terminals, get mangled in logs and
CI output, and are awkward to grep for. This applies to the Chinese material too:
Chinese text is fine, `->` beats an arrow glyph.

Mark status with words or ASCII punctuation:

```
ok  .asgard-pipeline.yaml
- workspace must not be empty
```

Some scaffold templates and extracts still carry decorative characters from
before this rule. Fix one when you are editing the file for another reason; do
not sweep them all as a change of its own.

## Every command must document itself

A command that cannot explain itself is not finished:

- `Short` is a one-line summary, shown in the parent's command list.
- `Long` explains what the command does, its defaults, and how it fails. It must
  say more than `Short` repeats. Include an example when the usage is not obvious.
- Every flag needs usage text, and says there if it is required or what it
  defaults to.
- **Never register a flag the code does not read.** An option that silently does
  nothing is worse than no option at all - and one kind reading `--layer` while
  its own help said `--layers` shipped exactly that way.

cobra gives every command a `--help` for free but does not stop it being empty,
and nothing here checks. Read the `Long` of a neighbouring command and match it.

To add a subcommand: write `newXxxCmd()` in `internal/cli/`, and register it
through `addTo(cmd, group..., ...)` in `root.go`. **Not `AddCommand`** - the
group is required, cobra panics on a `GroupID` the parent does not have, and a
command added the other way lands in "Additional Commands" where the next
reader will see it and ask why.

## Every directory the scaffold writes must document itself

Same rule as the commands above, applied to `internal/scaffold/templates/`. A
directory that arrives in a customer's repository with nothing in it but its own
name is a question the FDE has to ask somebody.

**Say both halves: what goes in it, and what does not.**

The second half is the load-bearing one, and it is not politeness. `assets/skills/`
and `.agents/skills/` differ by one word and are opposite mechanisms - one is
synced into a running system, the other is read by the agent editing this
repository - and what stops somebody putting a file in the wrong one is a
sentence in a README saying which is which. Every "not" in those files is there
because the confusion is available, not to fill space.

So, for **a directory whose name is ours to choose**:

- It contains a `README.md` (or an `_index.md` where a command writes into the
  same file) that names its purpose in the first line.
- It says what it is **not**, naming the directory somebody would otherwise
  confuse it with.
- One that starts empty says **why it is there at all** - `assets/skills/` is
  empty in a fresh repository on purpose, because a path invented per engagement
  is a path that differs per engagement, and `SkillSet.searchPaths` carries it.

**git cannot track an empty directory**, so this is not only documentation: the
README is what makes the directory exist in a clone at all.

The rule does not reach a directory whose shape somebody else decided.
`.agents/skills/` and the Claude Code plugin under
`internal/scaffold/templates/plugins/` are somebody else's layouts,
`.claude-plugin/` holds a manifest with a schema, and `db-query/scripts/` holds
the scripts its own `SKILL.md` documents. Naming what
those are for is the external convention's job, and a README in each would be a
second answer to a question already answered.

## A command that changes something says what it is changing

Before it acts, on stderr, in one line: the Platform API it is about to change,
the profile in effect, and **how that profile came to be the one in effect**.
`actingOn` in `internal/cli/context.go` is the helper; `actingLocally` is the
variant for the two commands that write the checkout's binding rather than
reaching the platform.

**The URL is the identity and the name is a local label.** Two people's `dev`
can point at different installations and one of them can be production, so a
line naming only the profile names the one part that is nobody else's fact.

**It has no conditions.** Printing it only when the profile was implicit - not
typed on the command line - was considered and refused: once somebody types
`--profile` out of habit the line disappears, and its absence then carries no
information, exactly the way a skipped gate step is not a pass.

It goes to stderr so that `--format json` keeps a parseable stdout. Adding the
same facts *into* the JSON was not done for the pipeline group: those commands
return a bare object or a bare array, and wrapping them to add three fields
would change the shape every existing reader depends on. `gate` was already
emitting a map, so it carries a `platform` key.

## The gate

**Prose and material carry almost all of the risk here, so the checks are aimed
there.** A handful of Go tests cover parsing and matching rules where a wrong
answer is silent; everything else is checked by reading what ships.

    go run ./hack verified run the gate, skipping what has not changed
    go run ./hack pass     every check, in the order to run them
    go run ./hack list     what each one is for, and what it needs

**`verified` is the one to run.** It keys each check on the digests of the
classes of input it reads and skips the ones whose inputs have not moved, so a
pass does not start from zero - which is most of why the same parts get read
again and again in one sitting. **A skip is printed differently from a pass**,
and the record lives in `.out/` rather than in the repository, because it is a
claim about this machine's tree and a shared one would be a claim about
somebody else's.

**The list is not written down here.** It is derived from the binary's own
flags, this directory's contents and the gate's own subcommands, so it cannot go
stale - which a list in a markdown file does, and this one had: it named a couple of
the Go checks and none of the rest.

`make gate` runs all of it except `--urls`, which is `make audit-urls`. The
Makefile **mirrors CI rather than defining it**: `.github/workflows/ci.yml` is
what decides whether a change merges, so a check added there and not to the
Makefile is still enforced, and one added to the Makefile alone is a
convenience.

`--paths` and `go run ./hack doc-paths` are the same rule from the two sides. The
audit reads what **lands** in a customer repository and fails on a path only we
have; the script reads the documents that never land - the root documents, both
READMEs and `hack/`'s own - where naming our own paths is the point, and fails
when one of them is gone. It checks a package-qualified Go symbol the same way, and
resolves it **inside the package that owns it** - a search of the whole tree
cannot tell one package's Index from strings.Index, and it passed a reference
to a symbol whose entire package had been deleted. **Do not write a deleted
symbol in backticks**, even to explain that it is deleted: this check reads
this file, and it will resolve it. It needs the checkout, which is why it is in
`hack/`.

**CI runs the checks that need nothing but this repository**: the audits, the
Go steps, and the `hack` checks that read only this tree. **Which ones those
are is not written here** - `.github/workflows/ci.yml` is the list, and it
carries a comment naming what it leaves out and why. The rest need a clone of
somebody else's repository - and asgard-core is private - so they are the
maintainer's to run, and `go run ./hack sources` says when one is due. `--urls` is excluded for
its own reason: a third party's outage is not this repository's build failure,
and a gate that only works online is one that fails on a plane.

**A customer repository's gate is one command, `asgard-cli gate`**, and the
difference is deliberate: the thing an agent runs after every edit has to be one
command whose definition lives in a binary. This one is for the maintainer, who
is editing that binary.

Three of the audits are worth a sentence each beyond what `--help` says:

  - **`--links`** resolves every pointer, in prose, in the templates and in the
    generator's own `Wiki:`, `Extract:` and `AlsoRead:` fields - and fails a
    path whose target `init` does not write into a repository, because that one
    resolves here and not there. **Run it after renaming or removing a page**,
    which is the only way to leave a dead pointer behind: it reads correctly,
    resolves to nothing, and the reader who follows it cannot tell that from a
    page they failed to find.
  - **`--commands`** is the same thing for the tool itself, and **it reads what
    is embedded**, which is what an engagement gets - plus these root documents,
    which `selfsrc.Docs` embeds for exactly this. They land nowhere, so the
    provenance rules do not reach them; what does reach them is that a command
    they name has to exist. **A sample of the tool's own output is read as
    prose**, so a fenced block reproducing a line where `asgard-cli` is followed
    by an ordinary English word is a reference to a command of that name. Elide
    it rather than rewriting what the tool prints.
  - **`--urls`** reports rather than fails on a citation that already says the
    link 404s - a page marked `draft: true`, which asgard-docs does not publish -
    so disclosing one is how you keep it.

None of those sees a wrong string in an embedded template, a pointer that
resolves to the wrong page rather than to none, or a generated CR the apiserver
would reject. **So exercise the change by hand**, against a scratch repository
outside this one:

```bash
go build -o .out/asgard-cli ./cmd/asgard-cli
cd $(mktemp -d)
/path/to/.out/asgard-cli init -y
/path/to/.out/asgard-cli project add app
/path/to/.out/asgard-cli add <kind> <name> --project app
/path/to/.out/asgard-cli check && /path/to/.out/asgard-cli render <release>
```

If the change touched a CR template, validate the rendered output against the
CRD schemas in [`asgard-kube/crd/`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) before believing it.
`helm lint` and `asgard-cli check` do not compare against the platform contract,
which is how four required fields were missing from two templates until somebody
looked.

## What is checked, and what is not

**This section says which of three kinds of thing covers a surface; the method
is `.agents/skills/consistency-checks/SKILL.md` and the list is
`go run ./hack pass`.** It runs **most-volatile first** - upstream before the
material, because nobody here touches asgard-docs and it moves without anyone
here noticing.

**"Every check passes" is not "the repository is correct", and the difference
is this table.** Every defect found by reading rather than by a check came from
a surface in the third group. So a claim that something is done says which
group it was in.

### Never report a pass as green while `TASK.md` has work in it

**Do not answer "all green", or anything that reads as it, while `TASK.md`
carries an unfinished item.** Report the checks as what they are - a list of
exit codes - and then say what is still open, by name.

The reason is not modesty. Green means every check this repository has passes,
and the surfaces that produce the defects are the ones with no check: **17 were
found by reading in a single pass during which everything was green**, among
them a table saying a required field had a default it does not have, a count of
88 written as 93 in seven places, and a pair of numbers no method could
reproduce. So a green run is evidence about the checks and almost none about
the material, and reporting it as a verdict on the repository tells the reader
the opposite of what it means.

**`TASK.md` is where that is read from.** "What is not done" is the list, every
line names what it is waiting for, and "The consistency pass" is the scope of
the current pass. If either has an entry, the answer is "the checks pass and
these are open", never "green".

**Mechanical, and fails the build.** Run them and the answer is not a
judgement. **What each one covers is the check's own description**, printed by
`go run ./hack list` and by `asgard-cli audit-material --help`, because a
description kept here is a second copy that drifts from the code implementing
it - this one did, and said `pass-list` checked something it had stopped
checking.

**Reported, and deliberately not enforced.** Each needs a person to read it,
and a green build says nothing about them:

| surface | why it cannot fail |
|---|---|
| `--orphans` | a document reached only by search is still reached |
| `--crossref` | a sentence describing another command reads correctly alone; only opening that command settles it |
| `audit-material` with no flag, `--ask`, `--unmarked` | every bold imperative on one screen, because the failure is two opposing ones never being in front of the same reader. `--ask` narrows to the ones telling a reader to ask a customer, which is where filter 0 applies |
| `--unchecked` | **what every document says it has NOT been held against**, which is the opposite question to `--unverified` and the only one that had no answer: a page whose marker names a whole surface passed the check and nothing put that surface in front of a reader. Every document is expected to have one, so it cannot fail - and it is what `TASK.md` reads its blocked list out of instead of keeping one |
| `--term <field>` | the sweep for a renamed platform field, across prose and templates. It cannot fail on its own: it only answers a question somebody asks it |
| `go run ./hack sources` | how far each clone is behind, which is information rather than a verdict |
| `go run ./hack reconcile` | **which pointers have not been read against their target since it moved.** `--links` says a pointer resolves and `related` says who points at a document; neither says whether the claim behind the pointer was ever checked against what it points at. `source/SOURCES.md` records that for an upstream commit; `source/reconciled.json` records it for a pointer inside the corpus, and it is committed for the same reason - a reading is a claim about the material. Record one with `reconcile <document>` **after** reading it, and note there is deliberately no way to record them all at once - that would be a claim nobody made |
| `go run ./hack related` | **which documents point at the ones a change touched.** `--orphans` asks whether anything points at a document; this asks what does, off the same graph. It is the re-read a change actually owes, as against the corpus - and it cannot fail, because a pointer is a question rather than a defect |
| `go run ./hack introduced` | **the count-shaped lines THIS change adds.** Over the corpus the same detector reports more than a thousand lines and would be a rule to delete rather than soften; over a diff it reports a few dozen. Existing counts are a backlog found by reading, and it closes - a document reviewed at a recorded digest does not come back until it changes. This is what stops it refilling behind the reading |

**Checked by nothing, and verified by reading.** This is the group that has
produced every finding.

**The rows are in [TASK.md](TASK.md), not here.** Two tables of the same
surfaces is two tables that drift, and these two did - the same slugs,
different prose, and a check whose whole job was to hold one copy against the
other. `TASK.md` is where the state of this repository is written down, a
reading is state, and this file is the rule for what a row has to say.

**A verdict typed into a table is not checkable and does not belong in one.**
`ok` beside a script is a result copied out of something that can produce it;
`read` beside a document is an honour-system claim no program can confirm.
What *is* checkable is whether the thing a reading was held against has moved
since, and `go run ./hack sources` reports that.

So a row records three things and no fourth:

    surface        what it is, named so that a reader can tell whether their
                   change lands in it
    held against   the clone, the commit, the binary, the rendered chart - the
                   thing that can be gone back to. **Not a count**: a tally is
                   evidence something was counted and reads exactly like
                   evidence something was read
    when           the date

**What a reading found goes where it was fixed, not into the row.** A
correction belongs on the document it corrects, a rule belongs in this file,
and the history belongs in `git log` - a findings column grows without bound,
is read by everybody, and is the changelog this file's first section forbids.

**Add a row when you add a surface, and move one up when you write its
check.** A surface that is in none of the three groups is one nobody has
decided about, which is the state every finding came out of.

## Before you say it is done

The gate above is mechanical and catches almost nothing that has actually gone
wrong here. These are the questions that would have. Each one is on the list
because skipping it shipped something.

`.github/pull_request_template.md` asks for the answers to three of them, which
is where they are hardest to skip - a PR body is written at the moment somebody
believes the work is finished. The rest are here because they are cheaper to ask
while the work is still being done.

**`gh pr create --body` replaces that template rather than filling it in.** The
template only prefills the web form, so a PR opened from the command line - which
is how they are opened here - silently skips it. Structure the body you pass
around the template's sections, or read it first with
`cat .github/pull_request_template.md`.

**Did you run it, or only read it?**
Auditing material is not using the tool. `project add` left the repository with a
project the config declared and the disk did not - `check` passed, `render`
failed - and that survived a long review of the material because nobody had run
`init`, `project add`, `add`, `check` and `render` in sequence. Do that, in a
throwaway directory, for anything that touches a command.

**Does it still work with no repository?**
Half the job is answering a question, and that question gets asked in a meeting,
before the engagement has a directory. `asgard-cli init` in an empty directory
writes the whole corpus, and `asgard-cli size` and `asgard-cli guide` answer
outside a repository too - reference material that requires an engagement is
unavailable exactly when somebody is deciding whether to have one. None of
those needs a repository, a session or a network. A new command that resolves a
profile, or reads the checkout's binding, before it can say anything has
quietly left that half - `actingLocally` in `internal/cli/context.go` is the
line where that starts. Run it in an empty directory.

**Which kind of artefact is this recipe for, and what inverts for the others?**
A rule that produces a good artefact of one kind silently produces a bad one of
another, and a checker that validates shape cannot tell them apart. Six defects
in the `proposal-deck` skill were found by an engagement building a real deck
and **every one passed the skill's own checks** - density, rhythm and content
all green - because the skill carried the proposal's rules and applied them to a
discovery deck, where several of them invert. Titles as assertions became
conclusions stated before the questions that would support them; the
three-to-five item bound compressed a customer's document into something only
its author could read. Say which kind a recipe is for.

**What else claimed the thing you just changed?**
A trap lives in a template, an extract, a stage prompt and a wiki page, and only
one of them is in your diff. Changing `allowWrite` in a template while an extract
still teaches the old value leaves the repository disagreeing with itself.
A term sweep - `asgard-cli audit-material <term>` - finds the other copies.

**Did you break the thing that was enforcing it?**
Renumbering the request template's sections silently broke `work.go`, which
appended status transitions under a heading by its literal number. If a rule was
being kept by something, changing the shape it keeps is the same change.

**If it is a list, does it say which of its items are load-bearing?**
**A reference that lists things uniformly invites use of all of them.** Prose
says what a thing is for and a schema says what shape it takes; **only the
worked example says how often a thing is right**, and nobody reads an example
for frequency. A design reference handed an agent thirteen CSS classes as one
inventory - the deck held up as correct uses six, one of them zero times - and
the agent filled the rest, because a list of thirteen reads as thirteen things
worth using.

Two ways out, and both are cheaper than the defect:

    split the list           a working set and a specialised one, so the
                             specialised ones are opted into rather than
                             defaulted to
    put the rule where it    the rule lived in full in one file, in part on the
    is used                  schema, and not at all on the thing somebody is
                             looking at while writing the line

**Do not close it by adding a copy at the point of use** - that is a fourth
home for a fact with three already. Move it, or point at it.

**Does anything point at what you added?**
Material nothing links to is not read, and the writer never finds out, because
the file is there. A document with no `group:` is in no index, which
`go run ./hack index` reports as its own failure rather than leaving to a
reader. Writing something and pointing at it are separate acts, and
only the first one feels like finishing. `audit-material --orphans` is the
mechanical half; grep for the name of what you added is the other.

**Is the claim verified, or asserted?**
The wiki was described as covering its sources completely, in prose. Counting
the citations against the source files gave a two-thirds job. If a claim is
countable, count it before writing it, and **put the number where the next
person can recount it** - `internal/corpus/wiki/index.md` carries that one and
`go run ./hack coverage` recomputes it.

**Would this be recognisable to the customer it came from?**
`--help` shipped a real customer's repository name as its example, and a stage
prompt's illustration is still modelled on one engagement's actual systems.
Nothing under `internal/` may name or portray a customer.

**Is "done" as wide as what you checked?**
Say what was verified and what was not. "The extracts are correct" and "the
extracts' YAML skeletons validate against the CRD" are different claims, and
reporting the first when you did the second is how a review passes something
broken.

**Do not ask whether to commit until you have finished checking.**
Asking moves the checking onto the person answering, and they answer on the
assumption that you already did it. A pass that ends with a question is a pass
that ended early.

What that means in practice, after the last edit and before the question:

    every check, again, from the top   not the ones you were watching
    what else claimed the thing        `go run ./hack related` for the
                                       documents, `reconcile` for the ones not
                                       read since, `audit-material --term` for
                                       a word - let their scope be the scope
    what YOU just introduced           `go run ./hack introduced`, then the id
                                       you added, the pointer you moved, the
                                       wording a check matches on

**The scope of a sweep is the tool's, not yours.** Deciding by hand which files
a sweep covers is the same mistake as writing a list that can be generated, and
it fails the same way: the set you remember is the set you have been editing,
which is not where the other copy is. `--term` reads every surface this
repository is responsible for - the material, the scaffold templates, every
`--help` screen, the CLI's own string literals, these documents, the maintenance
skills and the gate under `hack/` - because `everySurface` assembles it once. It was three
hand-written assemblies and they were not the same set; the narrowest was
`--term`'s, so a sweep run exactly as this section instructs came back clean
over surfaces it had never read.

The third is the one that gets skipped, because it is not a surface anybody
listed - it is a surface you created in the last ten minutes. A row added to
`platform-unknowns.md` with an id that already existed, and four pointers at
the old number, went out under "要 commit 嗎?" twice. **Nothing was red.** The
identifiers were new, so no check knew to hold them against anything, and the
question was asked as though the work were finished.

**A question about the work is not a substitute for finishing it**, and "the
gate is green" is not the same sentence as "I checked what I changed".

**Never annotate a source with a count.**
A provenance marker records **what was read and the commit it was read at**, and
nothing else. "Checked against 72 gated and 14 ungated tool entries" and
"checked against 11 SemanticLayer CRs across three deployments" are not
statements that anybody read those CRs - they are statements that something was
counted, and they read exactly like the first. That is the damage, and it is
worse when the number is right: **a correct count is what stops the next person
opening the page.** The check is green, the figure recomputes, and the sentence
beside it has been believed by everyone who has passed it.

So a marker names the scope - which deployment, which shape, which commit - and
the claim carries no tally:

    no    checked against 11 SemanticLayer CRs across three deployments
    yes   checked against every SemanticLayer CR in three deployments,
          and completionModelName is present on all of them

The second cannot be satisfied by counting, which is the point. **Deleting a
count deletes its checker too** - `go run ./hack counts` reports a row whose
claim has gone rather than passing silently, so the two stay in step.

**Does the number need to be there at all?**
**Ask that before asking whether it can be computed.** A count earns its place
when the count is the claim - how much of a chart `add` never writes, how many
pages a back office has. It does not when it is standing in for a yes: "one
arrow function in 520 values" was recounted three times, was wrong every time,
each recount used a different denominator, and one deployed expression answered
the same question permanently. **Writing a checker for a number like that is
treating the symptom.** And an index carries none: `internal/corpus/wiki/index.md`
had a seven-row ledger deriving its coverage figure by hand, with a record of
what the figure used to be beside it. `.agents/skills/consistency-checks/SKILL.md`
has the three cases.

**Did you say how the number was measured?**
Every coverage figure this repository has stated has been wrong at least once,
and each was believed rather than checked: an index claiming 100%, a comparison
that matched a URL against a file path, an image count divided by every file in
asgard-docs rather than by the ones any page uses. **A percentage with no method
beside it is a claim nobody can check and everybody repeats.** Write the
denominator and how you got it, or write the raw counts and no percentage - and
prefer a check that recomputes it to either.

**Would this check fire on material that is correct?**
**A static check over prose cannot tell an instruction from a mention**, and
the boundary is worth knowing before writing one. Validating an invocation's
argument count against cobra's own `Args` looks exact and is not: a quoted
argument is four words, a line continuation is a token, `(asgard-cli render)`
in a parenthetical takes the next three words, and an example block aligns a
trailing description with a single space - and one space is significant here
by rule. A check that fires on correct material is worse than none, because
the answer is to turn it off. **Prefer the check that runs where the mistake
is made**: `render` rejects a project-and-environment with its own message at
the moment somebody types it.
Three checks were written this way and deleted rather than tuned. One flagged a
processor config key the definitions do not declare - and fired on every
production chart with a streaming processor in it, always for `await`, which is
real and documented. It was five of five when the check was written and is seven
of seven now, which is the shape of evidence that should have stopped it being
written at all. One flagged
camelCase terms absent from the CRD - 29 findings, 0 defects. One flagged an
extract for listing two of ten syncer classes, which is exactly what an extract
about a git-backed SkillSet should do. **A checker that cries wolf teaches people
to change what it can see rather than what is wrong**, and this material has
already caused that once, in a customer deck. Delete it; do not soften it.

**Does the report name what it counts, in every category, always?**
A number with no list is not actionable, and a category that prints only when
another one is empty is worse than one that never prints: the count goes on
saying something is owed and the report stops saying what. `reconcile` printed
its never-recorded pointers only when nothing had moved, so a single moved
pointer hid every unread one behind "1 never recorded" - and what it hid was a
pointer added the same afternoon, which is when a pointer is least likely to
have been read. **Print the strongest category first and cap it rather than
suppressing it**: a capped list with "and N more" under it is something somebody
can finish, where a number alone is a thing to believe. The same rule as a skip that lies, one
layer out - an answer that is right and unusable reads exactly like one that is
both.

**If it is a pinned copy of the platform's contract, which way can it go stale?**
**Both ways, and the expensive direction is the one nobody expects.** A
constraint that loosens upstream leaves a pin that reports correct charts as
wrong, which costs more than a pin that has stopped catching something: the
platform deletes constraints as readily as it adds them, and a regex it
decides was wrong becomes a warning against the thing it now accepts.
`go run ./hack tables` is what catches it, and it takes the asgard-kube
checkout as its only argument so that running it is one command.
`gate` holds three: the processor definitions, the CRD enums, the CRD patterns.
A copy can only be wrong by being behind, so a rule built on one is a **warning**
- the platform adds a value and a correct chart looks wrong. The exception is a
condition broken against every version, like a required key with no default, and
that one may fail. Say which you are writing before you write it.

## Four ways this material has contradicted itself

Every one was found by somebody walking into it. None is enforced by anything,
which is the point of writing them down.

**1. Opposite instructions in two pages.** Twice, both
`internal/corpus/wiki/operations.md` against the interview stage - offering a
VPN or a jump host as alternatives when there is one shape, and asking who
approves a firewall change when filter 0 rejects exactly that. Both times the
FDE followed the page in front of them.

    anything that tells a reader to ask a customer something
    has to pass filter 0 first: does the answer change what we build?

**2. One word, two meanings.** `sandbox` meant the platform's agent runtime and
the customer's test environment, twenty lines apart in one section. Nothing
contradicted anything; the reader took the wrong sense.

    internal/corpus/wiki/glossary.md lists the terms that already mean
    something specific. Check it before introducing a word, and before using
    one of those for something else.

**3. An instruction right for one reader, copied by another.** This is the one
that scales worst. `internal/corpus/wiki/operations.md` says to get the name of
whoever approves a firewall change - correct for tracking, wrong on a slide -
and an FDE put it on one. Filter 0 says to ask who issues an account; same
outcome. **Three of one deck's six worst questions were copied out of this
material rather than reasoned into existence.**

    a correction only in the canonical place does not work.
    somebody copying reads one page, and it is not that one.

So when a line tells a reader to find out who somebody is, or to ask a customer
anything, **the destination goes inside the imperative** - not beside it:

    no    who issues the credential          ...and a paragraph explaining
                                             that the name is for tracking
    yes   who issues the credential, for the tracking row

**Adjacency was tried and it failed.**
`internal/corpus/wiki/operations.md` had the boundary two lines above the
instruction; an FDE read both sentences and copied only the
imperative onto a customer slide. An imperative is the shape a reader scanning
for "what do I do" hooks on, and prose beside one reads as elaboration rather
than as a limit - especially when the imperative is short and bold and the
caveat is a paragraph.

Embedded, the destination travels with the words. Removing it becomes a
deliberate act, and that act is the judgement that was missing.

**4. Handing work to another skill without its constraints.**
`proposal-deck` said to load the typesetting skill and let it produce the file,
and did not carry over that skill's own ordering requirement. An FDE arriving
from this side did not know the requirement existed and worked in the opposite
order for a whole deck. **That is an omission rather than a contradiction, and
it fails the same way** - follow the page and get hurt, find out by being hurt.

    when a page delegates, it carries the constraints of what it delegates to,
    or it says explicitly which of them do not apply here.

That one was resolved by a decision rather than a repair - only the layout
language is borrowed now, not the process - but the shape stands.

**None of the four is enforced by anything**, and that is what separates them
from the rules above, each of which has a check. A convention catches this
kind of thing only after somebody has been caught by it.

What would help is not a detector - all four read correctly line by line - but
putting every instruction in the material on one screen, because the cause is
that no two opposing ones are ever in front of the same reader.
`audit-material` with no flag is the closest thing: it lists every bold
imperative across every part, and `--ask` narrows to the ones that tell a
reader to ask a customer something, which is where filter 0 applies.

## Three questions the material has to keep answering

The questions above are asked of a change. These three are asked of the tool,
because they are what it is for, and each has been *nearly* true while missing
something specific. None of them is settled by reading - run the check.

**Does what it generates still satisfy the contract?**

This repo does not define CRDs; it consumes them. So the claim is not that a CRD
is correct, it is that **what the templates emit, and what the extracts show a
reader, are both accepted by the CRDs as they stand today** - which move without
telling anyone.

**Pull first.** Validating against a clone from three weeks ago proves nothing,
and asgard-kube moves without announcing it. Record the commit, and whether it
was head when you looked - "checked against the CRD" without one is not an
answer.

Render every kind and validate the result against the schemas: required fields,
fields that are not in the schema, enums, patterns, and the `ExactlyOneOf` rules.
Do the same for the YAML skeletons in the extracts, because those are what
somebody copies by hand. `go run ./hack extract-crs` pulls them out and
`go run ./hack validate-crs` checks them; this is not a check to do by eye, and
`hack/README.md` is the procedure.

If the contract has moved since this repo last looked, say what changed and what
it means here. A retired field, a flipped default and a new required field each
land differently - the first breaks a template silently, the second changes
behaviour with no diff at all, and only the third fails loudly.

```bash
asgard-cli render <release> | \
  <validate each document against asgard-kube/crd/*.yaml openAPIV3Schema>
```

**Neither `helm lint` nor `asgard-cli check` nor a server-side dry-run does this.**
That is how four required fields went missing from two templates, and how an
extract taught `agents` as a YAML list when the field is a stringified JSON
array. A dry-run is worse than useless here: it silently drops a field it does
not recognise and reports success, while helm's own server-side apply refuses.

**Can it get an FDE to the right integration by asking?**

Every entry point the platform offers has to be reachable through a question the
interview actually asks. `botProviderClass` is immutable once applied, so a
channel decided by assumption is a new BotProvider rather than an edit.

```bash
# the classes the platform has
grep -A3 'botProviderClass' ~/asgard-kube/crd/asgard-ai.com_botproviders.yaml | grep enum

# whether the interview asks about each route
grep -in 'chat\|channel\|LINE\|other end\|API\|console' \
    internal/stage/prompts/11-requirements.md
```

The interview asked who was on the other end - the question deciding hub against
flow agent - without ever asking **which channel**, which is a separate decision
on an immutable field.

**Does an agent have enough to assemble a chart?**

Every CR kind a chart may need has to have a generator, an extract, or a written
statement that it is not for an engagement to reach for. A kind with a CRD and
nothing else leaves the reader with a schema and no judgement.

```bash
# every kind the platform defines
ls ~/asgard-kube/crd/*.yaml | sed 's/.*com_//;s/s.yaml//'

# what any of the material mentions
cat internal/corpus/usecase/*.md internal/generate/templates/*.tmpl \
    internal/corpus/wiki/*.md | grep -o 'kind: [A-Z][A-Za-z]*' | sort -u
```

That comparison leaves three, and **all three are a deliberate absence rather
than a gap**: `ImageGenerationModel`, `TranscriptionModel` and
`SourceSetEditorServer` are internal, confirmed 2026-09-14. A kind with a CRD
and no material is the state this comparison exists to surface - it just turned
out, this time, that the answer was "not for an engagement". **Ask before
writing the page**, because the cost of the two mistakes is not symmetric:
material for a kind nobody may use invites somebody to use it.

**The same question one level down is `go run ./hack spec-key-gap`**, which asks
it per spec key rather than per kind, and mechanically. A key a production chart
uses is written by `add`, named in a commented skeleton where somebody meets it,
absent on purpose with the document that says why, or nowhere - and only the
last is a gap. **The third state is a pointer, not an opinion**: the row names
the document carrying the decision, and the check fails when that document is
gone or has stopped naming the key. So a key you decide against gets its reason
written where a reader meets it, the same as a kind does.

## Reference material lives outside this repo

The reference repositories, read-only, never vendored in. **The URL is the source of
truth; where you happen to clone it is not** - a local path is true on one
machine and wrong on every other:

**The contract:**

| what | source of truth |
|---|---|
| CRD definitions, the platform contract | https://github.com/asgard-ai-platform/asgard-kube |
| product documentation | https://github.com/asgard-ai-platform/asgard-docs |
| the processor definitions the CRD is generated from | https://github.com/asgard-ai-platform/asgard-core (private) |

`asgard-kube` is read at two depths and they answer differently: `crd/` is the
contract, and `pkg/apis/` is the Go types it is generated from, where the
reasoning survives as comments. `internal/corpus/wiki/crd-rules.md` was written
from the second.

**`internal/gate/processors.go` holds a pinned copy of asgard-core's
`ProcessorDefinitions`**, so the rule above about which way a pinned copy can go
stale applies to it - **and that list is a subset of what a chart may set rather
than the config contract**, because a processor declaring dynamic config takes
keys no definition names. `internal/corpus/wiki/processors.md` is where that is
recorded, and "would this check fire on material that is correct" above has the
cost of treating it as closed.

**The reference deployments** - every extract under `internal/corpus/usecase/`
was taken from one of these, and a claim about how a shape is built should be
checkable against at least one:

    unitech-e-asgard-kube            xxentria-asgard-kube
    finance-ai-asgard-kube           buy123-asgard-kube
    asgard-freyr-kube                asgard-auto-post-kube
    asgard-industry-demo-generator   asgard-freyr-skills

all under https://github.com/asgard-ai-platform/. **What shape each one
demonstrates is in `source/SOURCES.md`** and not restated here: that file owns
the attribution, states the commit each extract was written from and held
against, and its counts are recomputed by `go run ./hack counts`. A second copy
of them here would be a second copy nothing checks.

Which customer each belongs to, and how the extracts refer to one without naming
it, is in `source/SOURCES.md` - the one file here allowed to make that link, and
the reason it sits outside `internal/`.

Clone them wherever you like and point an environment variable at each:
`ASGARD_KUBE`, `ASGARD_DOCS`, `ASGARD_CORE`, and `ASGARD_DEPLOYMENTS` for the
directory holding the deployment clones. `go run ./hack sources` prints what they
resolve to and how far behind each one is.

Pull before relying on any of them, and **record the commit you read** in
whatever you write - a copy taken into this repo stops tracking upstream and
then reads exactly like a current one. `asgard-cli audit-material --sources`
holds every recorded commit against every other; **nothing can tell you a
recorded commit has gone stale**, which is what `go run ./hack sources` is for and why it
reads the clone rather than fetching it.

## Put generated files in `.out/`

Anything a command produces that does not belong in version control goes to
`.out/` at the repo root - not scattered, and not in `/tmp`:

- throwaway debug scripts and scratch programs
- command output, logs, sample data downloaded for inspection
- binaries built by hand
- profiling reports (`*.pprof`)

```bash
mkdir -p .out
go build -o .out/asgard-cli ./cmd/asgard-cli   # or: make build
```

`.out/` is gitignored and can be deleted at any time: `rm -rf .out`.
