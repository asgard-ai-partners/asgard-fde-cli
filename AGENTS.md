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
[APPROACH.md](APPROACH.md) is how the main capabilities are implemented, and
this file is how to change the code and the material without breaking it.

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

**Five of those six are written into a customer repository**, under
`.agents/skills/asgard-platform/`, by `asgard-cli init` - the scaffold templates
are the repository. So a change to any of them ships twice: into the binary,
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

The reading order is wiki, then extract: an extract assumes you already know the
platform has that shape. `asgard-cli add <kind>` prints both, in that order.

## The corpus is a wiki, and these are its rules

The shape is the [llm-wiki
pattern](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f) -
raw sources that are never vendored in, a corpus that is rewritten continuously,
and a schema a person changes deliberately. `internal/corpus/wiki/README.md`
states it for the wiki and is the longer version; the five rules below apply to
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
  breaks them in silence. 141 were outside the graph at once, half of them in
  the two indexes.

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

**Prose and material carry almost all of the risk here, so the checks are
aimed there.** A handful of Go tests cover parsing and matching rules where a
wrong answer is silent; everything else is checked by reading what ships.

```bash
go build ./...
go vet ./...
gofmt -l internal/ cmd/
go test ./...
asgard-cli audit-material --links
asgard-cli audit-material --commands
asgard-cli audit-material --bare
asgard-cli audit-material --paths
asgard-cli audit-material --unverified
asgard-cli audit-material --sources
hack/check-doc-paths.py
hack/check-tables.py                     # needs $ASGARD_KUBE
hack/check-coverage.py                   # needs $ASGARD_DOCS
asgard-cli audit-material --urls   # needs the network
```

`--paths` and `check-doc-paths.py` are the same rule from the two sides. The
audit reads what **lands** in a customer repository and fails on a path only we
have; the script reads the documents that never land - Goal, README, AGENTS,
STRUCTURE, APPROACH, TASK - where naming our own paths is the point, and fails
when one of them is gone. It checks a package-qualified Go symbol the same way, and
resolves it **inside the package that owns it** - a search of the whole tree
cannot tell one package's Index from strings.Index, and it passed a reference
to a symbol whose entire package had been deleted. **Do not write a deleted
symbol in backticks**, even to explain that it is deleted: this check reads
this file, and it will resolve it. It needs the checkout, which is why it is in
`hack/`.

Everything above except `--urls` runs in CI. `--urls` does not: a third party's
outage is not this repository's build failure, and a gate that only works
online is one that fails on a plane.

That is this repository's gate. **A customer repository's gate is one command,
`asgard-cli gate`**, and the difference is deliberate: the thing an agent runs
after every edit has to be one command whose definition lives in the binary,
not a list in a markdown file that goes stale. This list is for the maintainer,
who is editing the binary - and when a step is added to `gate`, nothing here
needs changing, which is the point.

`--urls` fetches every `docs.asgard-ai.com` link the material cites and fails on
a 404. A citation that already says the link 404s - a page marked `draft: true`,
which asgard-docs does not publish - is reported and does not fail, so
disclosing one is how you keep it.

`--links` resolves every pointer the material writes - the paths
`../wiki/<page>.md`, `../usecase/<name>.md` and the same for the other three
kinds - in prose, in the templates, and in the generator's own `Wiki:`,
`Extract:` and `AlsoRead:` fields, and exits 1 on one that goes nowhere. It
also fails a path whose target `init` does not write into a repository,
because that one resolves here and not there.
**Run it after renaming or removing a page**, which is the only way to leave a
dead pointer behind; it reads correctly and resolves to nothing, and the reader
who follows it cannot tell that from a page they failed to find.

`--commands` is `--links` for the tool itself: it resolves every
`asgard-cli <command>` this material writes - prose, help screens and the
scaffold templates alike - against the command tree the binary answers to, and
exits 1 on one that does not exist. **It reads what is embedded**, which is what
an engagement gets; this file, `README.md` and `STRUCTURE.md` are not in it,
because they are read from a checkout rather than shipped.

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

**"Every check passes" is not "the repository is correct", and the difference
is this table.** Every defect found by reading rather than by a check came from
a surface in the third group. So a claim that something is done says which
group it was in.

**Mechanical, and fails the build.** Run them and the answer is not a
judgement:

| surface | check |
|---|---|
| every document pointer, in material, templates and help | `--links` |
| a document named with no path | `--bare` |
| every command named, in material, templates, help, this repository's Go string literals and its own documents | `--commands` |
| a landed document naming a path only this repository has | `--paths` |
| every document carrying a provenance marker | `--unverified` |
| every upstream cited being declared in the raw-sources table | `--sources` |
| every documentation URL being live | `--urls` (needs the network) |
| the pinned enum and constraint tables against the CRDs | `hack/check-tables.py` (needs `$ASGARD_KUBE`) |
| the four numbers in the coverage row | `hack/check-coverage.py` (needs `$ASGARD_DOCS`) |
| every path and package-qualified Go symbol this repository's own documents name | `hack/check-doc-paths.py` |
| generated CRs and the extracts' skeletons against the CRD schemas | `hack/validate-crs.py` (needs `$ASGARD_KUBE`) |
| the gate over the reference charts | `hack/verify-references.sh` (needs the clones) |
| build, vet, gofmt, tests | CI |

**Reported, and deliberately not enforced.** Each needs a person to read it,
and a green build says nothing about them:

| surface | why it cannot fail |
|---|---|
| `--orphans` | a document reached only by search is still reached |
| `--crossref` | a sentence describing another command reads correctly alone; only opening that command settles it |
| `audit-material` with no flag | every bold imperative on one screen, because the failure is two opposing ones never being in front of the same reader |
| `hack/sources.py` | how far each clone is behind, which is information rather than a verdict |

**Checked by nothing, and verified by reading.** This is the group that has
produced every finding, so it carries a date:

| surface | last read, and how |
|---|---|
| the 22 extracts' prose against the charts they came from | 2026-09-11, every field name against the pulled clones and the CRDs, every count by rendering all 19 charts |
| the 27 wiki pages' prose against asgard-docs | the prose stands at `f00e0ee`; the 17 processor pages were read at `23409b3` on 2026-09-11, and **the rest have not been re-read** |
| the 10 stage prompts | 2026-09-11, for what they claim another command does; not for their guidance |
| the 7 design-time skills' prose | 2026-09-11, for command and path claims only |
| every flag's usage text against what the flag does | 2026-09-11, all 81 |
| **`internal/localenv`, `platform`, `auth`, `work`, `skills`, `gitrepo`, `render`, `binding`, `chart`, `tool`, `browser`, `version`, `pipelineconfig`, `repo`** | **never read end to end - about 7,300 lines.** Their string literals are in `--commands` and any symbol a document names is in `check-doc-paths.py`; nothing checks that their help describes what they do |

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

**Does anything point at what you added?**
Material nothing links to is not read, and the writer never finds out, because
the file is there. Writing something and pointing at it are separate acts, and
only the first one feels like finishing. `audit-material --orphans` is the
mechanical half; grep for the name of what you added is the other.

**Is the claim verified, or asserted?**
The wiki was described as covering its sources completely, in the repository, in
prose. Counting the citations against the source files gave 65 of 162. If a claim
is countable, count it before writing it, and put the number somewhere the next
person can recount it - `internal/corpus/wiki/index.md` carries that one.

**Would this be recognisable to the customer it came from?**
`--help` shipped a real customer's repository name as its example, and a stage
prompt's illustration is still modelled on one engagement's actual systems.
Nothing under `internal/` may name or portray a customer.

**Is "done" as wide as what you checked?**
Say what was verified and what was not. "The extracts are correct" and "the
extracts' YAML skeletons validate against the CRD" are different claims, and
reporting the first when you did the second is how a review passes something
broken.

**Did you say how the number was measured?**
Four coverage figures in this repo have been wrong, and each was believed rather
than checked: an index claiming 100%, a comparison that was case-sensitive, 39
divided by every image in asgard-docs rather than by the ones any page uses, and
the guess that preceded it. A percentage with no method beside it is a claim
nobody can check and everybody repeats. Write the denominator and how you got
it, or write the raw counts and no percentage.

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
processor config key the definitions do not declare - and fired on five of five
production charts, always for `await`, which is real and documented. One flagged
camelCase terms absent from the CRD - 29 findings, 0 defects. One flagged an
extract for listing two of ten syncer classes, which is exactly what an extract
about a git-backed SkillSet should do. **A checker that cries wolf teaches people
to change what it can see rather than what is wrong**, and this material has
already caused that once, in a customer deck. Delete it; do not soften it.

**If it is a pinned copy of the platform's contract, which way can it go stale?**
**Both ways, and the expensive direction is the one nobody expects.** A
constraint that loosens upstream leaves a pin that reports correct charts as
wrong, which costs more than a pin that has stopped catching something: the
platform deletes constraints as readily as it adds them, and a regex it
decides was wrong becomes a warning against the thing it now accepts.
`hack/check-tables.py` is what catches it, and it takes the asgard-kube
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
from the five rules above, each of which has a check. A convention catches this
kind of thing only after somebody has been caught by it.

What would help is not a detector - all four read correctly line by line - but
putting every instruction in the material on one screen, because the cause is
that no two opposing ones are ever in front of the same reader.
`audit-material` with no flag is the closest thing: it lists every bold
imperative across every part, and `--ask` narrows to the ones that tell a
reader to ask a customer something, which is where filter 0 applies.

## Three questions the material has to keep answering

The twelve above are asked of a change. These three are asked of the tool,
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
somebody copies by hand. `hack/README.md` is the procedure and `hack/` holds the
two scripts; this is not a check to do by eye.

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

That comparison currently leaves three: `ImageGenerationModel`,
`TranscriptionModel` and `SourceSetEditorServer` are in the contract and in no
page, template or extract - and in no product documentation either, which is why
nobody noticed. `TASK.md` carries it.

## Reference material lives outside this repo

Eleven repositories, read-only, never vendored in. **The URL is the source of
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

**`asgard-core` was cited by name six times and by URL nowhere**, including in
`internal/gate/processors.go`, whose processor contract is extracted from its
`ProcessorDefinitions`. That makes it a pinned copy of the platform's contract,
so the rule above about which way a pinned copy can go stale applies to it - and
`internal/corpus/wiki/platform-unknowns.md` P10 records that the list is **demonstrably
incomplete**: `await` and `temperature` are set in five production deployments
and appear in neither it nor the CRD. A gate rule built on treating it as
complete called five of five correct charts wrong, and was deleted.

**The reference deployments** - every extract under `internal/corpus/usecase/`
was taken from one of these, and a claim about how a shape is built should be
checkable against at least one:

| repo | shape it demonstrates |
|---|---|
| [unitech-e-asgard-kube](https://github.com/asgard-ai-platform/unitech-e-asgard-kube) | agent hub (5 agents / 6 semantic layers) and a single-agent flow agent; trigger; knowledge drive. The most recently maintained of the set, so it wins a generational conflict |
| [xxentria-asgard-kube](https://github.com/asgard-ai-platform/xxentria-asgard-kube) | supervisor + 9 agents |
| [finance-ai-asgard-kube](https://github.com/asgard-ai-platform/finance-ai-asgard-kube) | supervisor + 3 agents over 3 semantic layers |
| [buy123-asgard-kube](https://github.com/asgard-ai-platform/buy123-asgard-kube) | the minimal flow agent - **no Agent CR at all**, 8 CRs in the whole chart |
| [asgard-freyr-kube](https://github.com/asgard-ai-platform/asgard-freyr-kube) | supervisor + 5 subagents, `agents.expression`, sandbox hooks, a `tenants/` layout |
| [asgard-auto-post-kube](https://github.com/asgard-ai-platform/asgard-auto-post-kube) | **28 Plugin CRs**, knowledge bases, api workflows; 3 BotProviders |
| [asgard-industry-demo-generator](https://github.com/asgard-ai-platform/asgard-industry-demo-generator) | 12 industries, the read/write governance split, a Claude Code plugin of commands + skills |
| [asgard-freyr-skills](https://github.com/asgard-ai-platform/asgard-freyr-skills) | runtime skills as a repository of their own - the only source for browser operation, and what `browser-operation` was written from |

Which customer each belongs to, and how the extracts refer to one without naming
it, is in `source/SOURCES.md` - the one file here allowed to make that link, and
the reason it sits outside `internal/`.

Clone them wherever you like and point an environment variable at each:
`ASGARD_KUBE`, `ASGARD_DOCS`, `ASGARD_CORE`, and `ASGARD_DEPLOYMENTS` for the
directory holding the deployment clones. `hack/sources.py` prints what they
resolve to and how far behind each one is.

Pull before relying on any of them, and **record the commit you read** in
whatever you write - a copy taken into this repo stops tracking upstream and
then reads exactly like a current one. `asgard-cli audit-material --sources`
holds every recorded commit against every other; **nothing can tell you a
recorded commit has gone stale**, which is what `sources.py` is for and why it
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
go build -o .out/asgard-cli ./cmd/asgard-cli
```

`.out/` is gitignored and can be deleted at any time: `rm -rf .out`.
