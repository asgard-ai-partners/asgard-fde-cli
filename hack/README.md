# hack/

## Where the upstream clones are

Every check here needs one, and **the paths used to be written into the scripts
and into this file** - true on one machine, wrong on every other, and the
reason `check-tables.py` went eight upstream commits without being run. One
environment variable per source, and a default that is one person's layout:

    hack/sources.py            what each one resolves to, and how far behind it is
    hack/check-tables.py       the pinned gate tables against the CRDs
    hack/check-coverage.py     the wiki's coverage row against the docs tree
    hack/check-processors.py   wiki/processors.md's two tables against their owners
    hack/check-counts.py       counts this material asserts about a deployment

    ASGARD_KUBE          the CRDs, the platform contract
    ASGARD_DOCS          the product documentation
    ASGARD_CORE          the processor definitions the CRDs come from
    ASGARD_DEPLOYMENTS   the directory holding the reference deployment clones

**Nothing here clones or pulls.** A script that fetched would turn "read at this
commit" into "read at whatever was there when the script ran", which is the one
thing the provenance rule exists to prevent. `git -C <path> pull` is the
reader's act, and `sources.py` tells you when it is due.

This repository's own tooling. Not shipped, not embedded, and not the same thing
as `.agents/skills/db-query/scripts/`, which `asgard-cli init` writes into a
customer repo - that one reads the customer's own source systems, and is
described in `README.md`.

Everything here answers one question: **is what this repo emits still accepted by
the platform contract?** `helm lint`, `asgard-cli check` and a server-side
dry-run do not answer it. The dry-run is worse than silent, because it drops a
field it does not recognise and reports success while helm's own server-side
apply refuses.

## Reading every instruction at once

    asgard-cli audit-material [--ask] [--unmarked] [--crossref]

**In the binary, not here.** It began as a script in this directory and was
moved, for a reason worth keeping: an audit that only runs on the maintainer's
machine only finds what the maintainer can see. What a maintainer finds by
reading is inconsistency; what costs money is somebody following an instruction
into a wall, and that person has the binary and not this repository. Now they can
run it at the moment they hit one.

It also reads the **embedded** material, which is what an engagement gets. A
script over `internal/corpus/wiki/*.md` audits the input instead, and the thing
being audited is what somebody actually read.

Hidden from `--help`, because its reader edits this material and the help output
belongs to whoever is onboarding a customer.

## Two implementations of the .env format, and whether they agree

    go run ./hack/dotenv-agreement

There have to be two: `asgard-cli local-env` writes the file in Go, the
db-query scripts read it in python. A format with two implementations and
nothing comparing them drifts silently, and the way it surfaces is the worst
kind - the form shows one value and the query connects with another.

It also checks that a save changes the one value it was asked to change and
nothing else. **That file is edited by hand as well**, and a note somebody left
for the next reader is worth as much as the value beside it. Two ways to lose
one have already been caught here: a trailing comment dropped when its line was
rewritten, and a quoted value re-spelled bare on a line nobody had touched -
which is also a meaning change to any shell that sources it.

Needs python3. Without it the python half reports as **skipped**, not passed.

## Running the gate over the deployments its rules came from

    hack/verify-references.sh [parent-dir]        default: ..
    ASGARD_CLI=.out/asgard-cli hack/verify-references.sh ~/projects/asgard

**The gate had never been run over the charts its rules were written from.** It
was run over charts this tool generates, which pass by construction, and over a
scratch repository. The first time somebody rendered the six reference
deployments through it, `gate` R1b was wrong about nine Agents in a running
deployment: it counted a semantic layer and a Toolset as capability sources and
not a `SkillSet`, so every subagent of a flow-agent supervisor was told it had
"no source of capability at all" while it had one.

**A count is not a pass.** Read what each finding says, because three kinds turn
up and they need opposite responses:

    a rule that is wrong             fix the rule - R1b was this
    a chart that is wrong            tell whoever owns it
    a rule right for one shape,      the expensive kind. See R1b, and the
    applied to another               a rule right for one shape, wrong for the next

And `--rendered` sees a chart with nothing around it. R7 and R11 ask
questions whose answer can live in the owning repository rather than in the CR,
and there is nowhere to record one now: `.asgard-config.json` held
`olapOnlyLayers` and the sampleQuestions exemption, and none of the four
reference repos still carries it. So a finding may be answered somewhere this
run cannot see. **That is a reason to read a finding, not to discount one.** It
was used to discount R1b once, and R1b was a bug.

## Checking a change against the CRDs

Pull asgard-kube first. Validating against a clone from three weeks ago proves
nothing, and it moves without announcing it.

```bash
KUBE=../asgard-kube
git -C $KUBE fetch && git -C $KUBE status -sb        # say so in the PR if behind
mkdir -p .out/crdjson
# check-tables.py converts these itself; this is only for validate-crs.py below
for f in $KUBE/crd/*.yaml; do yq -o=json "$f" > .out/crdjson/$(basename $f .yaml).json; done
```

**What the templates emit.** Build a throwaway repository, add every kind, render
both environments, validate each:

```bash
go build -o .out/asgard-cli ./cmd/asgard-cli
# init, scaffold, project add, then one `add <kind>` per kind, then:
asgard-cli render <release> --quiet | yq -o=json -I=0 '.' > .out/dev.ndjson
python3 hack/validate-crs.py .out/crdjson .out/dev.ndjson
```

**What the extracts teach.** These are what somebody copies by hand, so they are
checked the same way. They are chart fragments rather than parseable YAML, so
they are defused first - Helm actions and `<placeholder>` text become sentinels
the validator knows not to report on:

```bash
python3 hack/extract-crs.py .out/extracts.ndjson
python3 hack/validate-crs.py .out/crdjson .out/extracts.ndjson
```

Both should print `0 schema violation(s)`. Put the counts and the asgard-kube
commit in the PR body - `.github/pull_request_template.md` asks for them.

## Checking the pinned tables against the CRDs

    hack/check-tables.py $KUBE/crd

`internal/gate` holds three copies of the platform contract, extracted from
asgard-kube's **Go types**. The Go types are not the contract; the generated
CRDs are, and the two are not the same document. `status` carries three values
in the Asgard types and six in the CRD, because Kubernetes' own condition
schema uses that field name - the wrong three sat in the enum table for a day.

Run it after regenerating a table and whenever asgard-kube moves. A field it
reports as absent from the CRD is not automatically a bug - `baseAgentName`
lives inside a JSON string rather than in the schema - but it is always
something to explain rather than leave.

**It also holds every CEL-rule count this repository states**, for the same
reason and against the same trap: 79 is the `XValidation` markers in the Go
types, 231 is what the generator emits from them, and this material had the
marker count written down as the CRDs' own for a week.

**And every immutable field.** 41 of the enforced rules are `self == oldSelf`,
carried on 40 kind-and-property pairs across twelve kinds, and nothing offline
can tell you a chart will be refused at apply - but **which fields they are** is
computable, and an immutable field nobody has written down is one an FDE meets
after the tag is pushed. So this checks that `wiki/crd-rules.md` names all
eleven class fields, states the Syncer's 21 and the total, and that no immutable
Syncer field is missing from the corpus. Pairs rather than distinct paths:
`bot.botProviderName` is immutable on the Loader and on the Syncer, and those
are two fields somebody can be refused on.

## How far the generated chart is from a real one

    hack/spec-key-gap.py             recompute, and check TASK.md's claim
    hack/spec-key-gap.py --missing   the keys production uses and `add` never writes

**The number behind "the chart half is the least finished of the four."** It
decides whether an FDE treats what `add` emits as a chart or as a starting
point, and it stood at "168 spec keys, 52 never mentioned" for a week with no
method that reproduced either figure. It is 185 and 88 for the widest reference
chart, 303 and 171 across all nineteen, and both sides are rendered here rather
than quoted.

`--missing` is the useful half: it is the worklist for closing the gap.

Two traps it had to be taught. `$ASGARD_DEPLOYMENTS` is somebody's projects
directory and also holds scratch repositories this tool scaffolded - those pass
by construction, one of them at 0 keys not written - so only the deployments
`source/SOURCES.md` declares are counted. And list indices are collapsed, or
`processors.0.configs` and `processors.7.configs` count apart and the gap
appears to close as a chart grows.

## Recomputing a count that came out of somebody else's document

    hack/check-counts.py           against the clones as they stand
    hack/check-counts.py --dump    print what upstream counts, and stop

**A number copied out of a document that states its own count is the cheapest
thing in this material to get wrong, and the most expensive to notice**: nothing
about "88" reads differently from "93". A pass that set out to recount SHOPLINE's
back-office map took a figure off a different tally and wrote it into seven
places, where it sat for a week looking exactly as authoritative as the truth.

So each of those counts is recomputed from the clone, and every place this
material states one has to agree. **A claim whose wording has drifted out of
every pattern is a failure rather than a pass** - that is how a count stops
being checked without anybody deciding to stop checking it.

## Re-walking the processor definitions

    hack/check-processors.py            against the clones as they stand
    hack/check-processors.py --dump     print what upstream says, and stop

`wiki/processors.md` is the most claim-dense page in the corpus - thirteen
processors, their required keys, their defaults, their outputs, and which keys
an author may set - and every one of those claims belongs to a file in somebody
else's repository. **The two tables on it have two different owners, and they
disagree on purpose:**

    the definitions table   asgard-core `internal/constants.go` -
                            what the runtime validates a Workflow against
    the palette table       asgard-docs' per-page `metadata.json` -
                            what the builder lets an author type

So each table is checked against its own owner and never against the other. A
processor appearing or vanishing fails: the page says thirteen in four places.

**The literal is walked by brace depth rather than matched by pattern.** An
earlier pattern-based extraction of that same literal attributed one
processor's fields to the next, and a table confidently wrong about `allowWrite`
is worse than no table at all. Every identifier must resolve to a string or the
script exits - an unresolved one means the literal grew a shape the walk does
not understand, which is exactly when its output must not be trusted.

Writing it found six things reading had missed, including `validate-payload`'s
`schema` marked as having a default it does not have - which told a reader that
omitting it was a silent choice when it is a rejected CR.

## What this catches that nothing else does

Required fields with no default (a `Workflow` entry's `tooling.allowUploadFile`
was missing from two extracts, so a reader copying one got a rejected CR), fields
absent from the schema, enums, patterns, `maxItems`, and the `ExactlyOneOf` CEL
rules.
