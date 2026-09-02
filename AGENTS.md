# AGENTS.md

Rules for AI agents working in this repo. Human contributors follow the same ones.

## What this repo is

A single Go binary, `asgard-cli`, that drives the onboarding of one customer onto
the Asgard platform. It writes files into a *customer's* repository and never
holds state of its own.

Most of the value is not code. It is four bodies of embedded material, and the
first question when adding anything is which one it belongs to:

[STRUCTURE.md](STRUCTURE.md) walks every directory. The short version:

| material | answers | lives in |
|---|---|---|
| stage prompts | what to do at this point in an onboarding | `internal/stage/prompts/` |
| wiki pages | what the platform is, and who each piece is for | `internal/wiki/pages/` |
| usecase extracts | how one shape of deployment is assembled, field by field | `internal/usecase/extracts/` |
| scaffold templates | the part of a customer repo that is the same every time | `internal/scaffold/templates/` |

**One fact, one home; everywhere else links.** A trap that belongs to a CR
template does not also get explained in a wiki page. When you are about to repeat
a paragraph, link instead - two copies drift, and the reader cannot tell which is
current.

The reading order is wiki, then extract: an extract assumes you already know the
platform has that shape. `asgard-cli add <kind>` prints both, in that order.

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

`internal/wiki/` used to be the second. It is English now: the sources are zh-TW,
but an agent asked in Chinese queries in English, so the corpus does not need to
carry both - and one language removes the split where a Chinese question reached
only the wiki and an English one only the extracts. Product labels keep their own
names (Managed Agent, Drive, Context Index are what the UI says).

Do not "fix" the scaffold templates into English. Do not start a second exception
without saying why it earns one.

## Plain ASCII, no emoji

No emoji, and no decorative Unicode either - no check marks, arrows, box-drawing
characters. They render inconsistently across terminals, get mangled in logs and
CI output, and are awkward to grep for. This applies to the Chinese material too:
Chinese text is fine, `->` beats an arrow glyph.

Mark status with words or ASCII punctuation:

```
ok  .asgard-config.json
- workspace.id must not be empty
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

## The gate

There is no test suite - it was removed on 2026-09-02. What is left:

```bash
go build ./...
go vet ./...
gofmt -l internal/ cmd/
```

None of those sees a wrong string in an embedded template, a stage prompt that
sends a reader to a page that does not exist, or a generated CR the apiserver
would reject. **So exercise the change by hand**, against a scratch repository
outside this one:

```bash
go build -o .out/asgard-cli ./cmd/asgard-cli
cd $(mktemp -d)
/path/to/.out/asgard-cli init --workspace-id 0 && /path/to/.out/asgard-cli scaffold
/path/to/.out/asgard-cli project add app --env dev
/path/to/.out/asgard-cli add <kind> <name> --project app
/path/to/.out/asgard-cli check && /path/to/.out/asgard-cli render app dev
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
asgard-cli next --stage requirements | grep -in 'chat\|channel\|LINE\|other end\|API\|console'
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

Ten repositories, read-only, never vendored in. **The URL is the source of
truth; where you happen to clone it is not** - a local path is true on one
machine and wrong on every other:

**The contract:**

| what | source of truth |
|---|---|
| CRD definitions, the platform contract | https://github.com/asgard-ai-platform/asgard-kube |
| product documentation | https://github.com/asgard-ai-platform/asgard-docs |

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
