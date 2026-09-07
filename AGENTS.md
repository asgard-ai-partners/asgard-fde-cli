# AGENTS.md

Rules for AI agents working in this repo. Human contributors follow the same ones.

## What this repo is

A single Go binary, `asgard-cli`, that does two things for an FDE:

  1. **Answers an agent's questions about integrating with Asgard** - what the
     platform has, which CR a UI name maps to, how one shape is assembled field
     by field, and where each has been got wrong before. This works with no
     repository present, and has to keep working that way.
  2. **Helps assemble a project's chart** - CR skeletons, the invariants a
     rendered chart has to hold, and the judgement that goes with both.

It writes files into a *customer's* repository and never holds state of its own.
[TASK.md](TASK.md) states the goal in full and why it changed; this file is how
to change the code and the material without breaking it.

**Most of the value is not code.** It is one corpus in four parts, and the first
question when adding anything is which part it belongs to:

| part | answers | lives in |
|---|---|---|
| wiki pages | what the platform is, and who each piece is for | `internal/wiki/pages/` |
| usecase extracts | how one shape of deployment is assembled, field by field | `internal/usecase/extracts/` |
| stage prompts | what to weigh at one point in the work | `internal/stage/prompts/` |
| scaffold templates | the part of a customer repo that is the same every time, including the skills the customer's agent loads | `internal/scaffold/templates/` |

[STRUCTURE.md](STRUCTURE.md) walks every directory.

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
and a schema a person changes deliberately. `internal/wiki/README.md` states it
for the wiki and is the longer version; these four rules apply to all four
parts. All four hold as of 2026-09-04, and each has a command that says so -
`find --unverified`, `audit-material --links`, `audit-material --orphans`. A
rule that stops holding shows up there rather than in a list somebody has to
maintain.

**One schema.** A document opens with a `# ` title and a summary paragraph, and
carries `**Checked:**` / `**Unchecked:**` - what it has been held against, and
what it has not. `internal/kb` parses that and nothing else; a part of the corpus
that needs its own reader, its own search and its own match type is a fifth thing
that will drift. Do not add one. If new material does not fit `kb.Corpus`, that
is a reason to change `kb`, not to write a second one.

**A pointer is data, not prose.** Write a cross-reference in the canonical form -
`asgard-cli wiki <page>`, `asgard-cli usecase <name>`, `asgard-cli guide
<name>`, `asgard-cli brief <activity>` - because `audit-material --links`
resolves exactly those and fails on one that goes nowhere. A pointer written any
other way is invisible to the gate, and a dead pointer reads correctly right up
to the moment somebody follows it.

**Guidance is retrieved by subject or by condition, never by position.** "You are
at step 4" is a claim about a walk that no engagement actually performs. What a
reader can act on is a claim about the repository in front of them, or about the
thing they are about to do. `asgard-cli brief` was added for exactly this reason
and its package comment is the argument; `asgard-cli find` searches all four
parts by subject for the same one. **Anything reachable only by having arrived
somewhere is unreachable**, and the measured version of that is in TASK.md: an
engagement read the stage it was told it was in and what that stage pointed at,
and never opened the page it spent a day needing.

**What an agent reads has to be parseable.** Column-aligned output is for the
FDE, and it is not an interface. A command an agent acts on owes it a form that
does not have to be recovered from `%-9s`.

## Language

**Go code, command output, error messages, command help, and this repo's own
English documents are English.** That is `README.md`, `TASK.md`, this file, and
`source/SOURCES.md`.

```go
// Save writes cfg to path.
return fmt.Errorf("write %s: %w", path, err)
```

**One body of material is deliberately Chinese**: parts of
`internal/scaffold/templates/`, because the generated repository is read by the
customer's own engagement, in their language.

`internal/wiki/` used to be the second. It is English now, and one language
removes the split where a Chinese question reached only the wiki and an English
one only the extracts.

**The corpus carries one language because the mapping lives somewhere else, not
because the reader translates.** `asgard-cli wiki --aliases` is the mapping,
`asgard-cli find` applies it to a query before searching, and the rewrite is
printed so a reader can see what was searched. So the instruction to an agent is
the opposite of what this paragraph used to imply: **give `find` the customer's
own words.** Translating only after a search came back empty was tried and was
worse: the table matched the row about the word itself. Product labels keep
their own names (Managed Agent, Drive, Context Index are what the UI says).

**An index is not a page and does not live among them.** That table was a
section of `pages/glossary.md` until it was measured: because it lists every
alias, it was reliably the one document carrying every term of a translated
query, so a search for a subject returned the word list rather than the page.
It sits beside `pages/` now, with `index.md` and `log.md`. Anything that
catalogues the corpus goes there, and `audit-material --links` still reads it -
`--orphans` deliberately does not, because a list that names every page makes
every page look reached.

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

To add a subcommand: write `newXxxCmd()` in `internal/cli/`, register it in the
`cmd.AddCommand(...)` call in `root.go`.

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
`.agents/skills/` and `plugins/asgard-fde/` are the agent-skill and Claude Code
plugin layouts, `.claude-plugin/` holds a manifest with a schema, and
`db-query/scripts/` holds the scripts its own `SKILL.md` documents. Naming what
those are for is the external convention's job, and a README in each would be a
second answer to a question already answered.

## The gate

There is no test suite - it was removed on 2026-09-02. What is left:

```bash
go build ./...
go vet ./...
gofmt -l internal/ cmd/
asgard-cli audit-material --links
asgard-cli audit-material --commands
asgard-cli audit-material --urls   # needs the network
```

<<<<<<< HEAD
That is this repository's gate. **A customer repository's gate is one command,
`asgard-cli gate`**, and the difference is deliberate: the thing an agent runs
after every edit has to be one command whose definition lives in the binary,
not a list in a markdown file that goes stale. This list is for the maintainer,
who is editing the binary - and when a step is added to `gate`, nothing here
needs changing, which is the point.
=======
`--links` and `--commands` run in CI (the `material` job). `--urls` does not: a
third party's outage is not this repository's build failure.
>>>>>>> origin/feat/skills-from-platform

`--urls` fetches every `docs.asgard-ai.com` link the material cites and fails on
a 404. It is separate because it needs the network, and a gate that only works
online is one that fails on a plane. A citation that already says the link 404s
- a page marked `draft: true`, which asgard-docs does not publish - is reported
and does not fail, so disclosing one is how you keep it.

`--links` resolves every `asgard-cli wiki <page>`, `usecase <extract>`,
`brief <activity>` and `guide <name>` the material writes - in prose and
in the generator's own `Wiki:`, `Extract:` and `AlsoRead:` fields - and exits 1
on one that goes nowhere. **Run it after renaming or removing a page**, which is
the only way to leave a dead pointer behind; it reads correctly and resolves to
nothing, and the reader who follows it cannot tell that from a page they failed
to find.

`--commands` is `--links` for the tool itself: it resolves every
`asgard-cli <command>` this material writes - prose, help screens and the
scaffold templates alike - against the command tree the binary answers to, and
exits 1 on one that does not exist. **It reads what is embedded**, which is what
an engagement gets; this file, `README.md` and `STRUCTURE.md` are not in it,
because they are read from a checkout rather than shipped.

It exists because that failure shipped. `asgard-cli pipeline deliveries` was
named in six documents - the verification skill, a scaffolded `AGENTS.md` and
`README`, two stage prompts - as the one place a push that produced no run
explains itself, and no such command had ever been built. A person re-reading a
provenance line found it, weeks later.

None of those sees a wrong string in an embedded template, a pointer that
resolves to the wrong page rather than to none, or a generated CR the apiserver
would reject. **So exercise the change by hand**, against a scratch repository
outside this one:

```bash
go build -o .out/asgard-cli ./cmd/asgard-cli
cd $(mktemp -d)
/path/to/.out/asgard-cli init && /path/to/.out/asgard-cli scaffold
/path/to/.out/asgard-cli project add app
/path/to/.out/asgard-cli add <kind> <name> --project app
/path/to/.out/asgard-cli check && /path/to/.out/asgard-cli render <release>
```

If the change touched a CR template, validate the rendered output against the
CRD schemas in [`asgard-kube/crd/`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) before believing it.
`helm lint` and `asgard-cli check` do not compare against the platform contract,
which is how four required fields were missing from two templates until somebody
looked.

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
`init`, `scaffold`, `project add`, `add` and `render` in sequence. Do that, in a
throwaway directory, for anything that touches a command.

**Does it still work with no repository?**
Half the job is answering a question, and that question gets asked in a meeting,
before the engagement has a directory. `wiki`, `usecase`, `find`, `brief`, `size`
and `guide` all answer outside one - reference material that requires an
engagement is unavailable exactly when somebody is deciding whether to have one,
and `asgard-cli init` is what somebody with no directory runs first - it needs
no repository, no session and no network. A new command that calls `config.Find`
before it can say anything has quietly left that half. Run it in an empty
directory.

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
`asgard-cli find <the thing>` finds the other copies.

**Did you break the thing that was enforcing it?**
Renumbering the request template's sections silently broke `work.go`, which
appended status transitions under a heading by its literal number. If a rule was
being kept by something, changing the shape it keeps is the same change.

**Does anything point at what you added?**
Material nothing links to is not read, and the writer never finds out, because
the file is there. The `find` command shipped with every routing document still
sending readers to the two narrower commands it replaced - it worked, and nothing
mentioned it outside its own source file. Writing something and pointing at it
are separate acts; `source/FINDINGS.md` records this repo doing only the first,
repeatedly. Grep for the name of what you added.

**Is the claim verified, or asserted?**
The wiki was described as covering its sources completely, in the repository, in
prose. Counting the citations against the source files gave 65 of 162. If a claim
is countable, count it before writing it, and put the number somewhere the next
person can recount it - `internal/wiki/pages/index.md` carries that one.

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
Three checks were written this way and deleted rather than tuned. One flagged a
processor config key the definitions do not declare - and fired on five of five
production charts, always for `await`, which is real and documented. One flagged
camelCase terms absent from the CRD - 29 findings, 0 defects. One flagged an
extract for listing two of ten syncer classes, which is exactly what an extract
about a git-backed SkillSet should do. **A checker that cries wolf teaches people
to change what it can see rather than what is wrong**, and this material has
already caused that once, in a customer deck. Delete it; do not soften it.

**If it is a pinned copy of the platform's contract, which way can it go stale?**
`gate` holds three: the processor definitions, the CRD enums, the CRD patterns.
A copy can only be wrong by being behind, so a rule built on one is a **warning**
- the platform adds a value and a correct chart looks wrong. The exception is a
condition broken against every version, like a required key with no default, and
that one may fail. Say which you are writing before you write it.

## Two rules the material contradicted itself on

Three self-contradictions have been found, all by somebody walking into one, and
they come in two shapes.

**Opposite instructions in two pages.** Twice, both `wiki operations` against the
interview stage - offering a VPN or a jump host as alternatives when there is one
shape, and asking who approves a firewall change when filter 0 rejects exactly
that. Both times the FDE followed the page in front of them.

    anything that tells a reader to ask a customer something
    has to pass filter 0 first: does the answer change what we build?

**One word, two meanings.** `sandbox` meant the platform's agent runtime and the
customer's test environment, twenty lines apart in one section. Nothing
contradicted anything; the reader took the wrong sense.

    `wiki/pages/glossary.md` lists the terms that already mean something
    specific. Check it before introducing a word, and before using one of
    those for something else.

**A third shape, and it is the one that scales worst.** An instruction that is
right for one reader is copied by another. `wiki operations` says to get the name
of whoever approves a firewall change - correct for tracking, wrong on a slide -
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

**Adjacency was tried and it failed.** `wiki operations` had the boundary two
lines above the instruction; an FDE read both sentences and copied only the
imperative onto a customer slide. An imperative is the shape a reader scanning
for "what do I do" hooks on, and prose beside one reads as elaboration rather
than as a limit - especially when the imperative is short and bold and the
caveat is a paragraph.

Embedded, the destination travels with the words. Removing it becomes a
deliberate act, and that act is the judgement that was missing.

**A fourth shape: handing work to another skill without its constraints.**
`proposal-deck` said to load the typesetting skill and let it produce the file,
and did not carry over that skill's own ordering requirement. An FDE arriving
from this side did not know the requirement existed and worked in the opposite
order for a whole deck. **That is an omission rather than a contradiction, and
it fails the same way** - follow the page and get hurt, find out by being hurt.

    when a page delegates, it carries the constraints of what it delegates to,
    or it says explicitly which of them do not apply here.

That one was resolved by a decision rather than a repair - only the layout
language is borrowed now, not the process - but the shape stands.

**None of the four is enforced by anything.** They are conventions, and the
record is that a convention catches this only after somebody has been caught by
it. Issue #9 tracks whether any can be made mechanical. `hack/imperatives.py` is the
first step and is deliberately not a detector: it lists every instruction in the
material on one screen, because the cause is that no two opposing ones are ever
in front of the same reader. Its first run found two instructions whose reader
was wrong.

## Three questions the material has to keep answering

The seven above are asked of a change. These three are asked of the tool, because
they are what it is for, and each has been *nearly* true while missing something
specific. None of them is settled by reading - run the check.

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
asgard-cli render <project> <env> | \
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
asgard-cli guide requirements | grep -in 'chat\|channel\|LINE\|other end\|API\|console'
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
cat internal/usecase/extracts/*.md internal/generate/templates/*.tmpl \
    internal/wiki/pages/*.md | grep -o 'kind: [A-Z][A-Za-z]*' | sort -u
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
reasoning survives as comments. `wiki crd-rules` was written from the second.

**`asgard-core` was cited by name six times and by URL nowhere**, including in
`internal/gate/processors.go`, whose processor contract is extracted from its
`ProcessorDefinitions`. That makes it a pinned copy of the platform's contract,
so the rule below about which way a pinned copy can go stale applies to it - and
`wiki platform-unknowns` P10 records that the list is **demonstrably
incomplete**: `await` and `temperature` are set in five production deployments
and appear in neither it nor the CRD. A gate rule built on treating it as
complete called five of five correct charts wrong, and was deleted.

**The reference deployments** - every extract under `internal/usecase/extracts/`
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

Clone them wherever you like. Pull before relying on any of them, and **record
the commit you read** in whatever you write - a copy taken into this repo stops
tracking upstream and then reads exactly like a current one.

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
