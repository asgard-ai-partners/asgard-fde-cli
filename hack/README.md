# hack/

## Where the upstream clones are

Every check here needs one, and **the paths used to be written into the scripts
and into this file** - true on one machine, wrong on every other, and the
reason `check-tables.py` went eight upstream commits without being run. One
environment variable per source, and a default that is one person's layout:

    hack/sources.py          what each one resolves to, and how far behind it is

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

And `--rendered` sees a chart with nothing around it. R7, R10 and R11 ask
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
asgard-cli render <project> dev  --quiet | yq -o=json -I=0 '.' > .out/dev.ndjson
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

## What this catches that nothing else does

Required fields with no default (a `Workflow` entry's `tooling.allowUploadFile`
was missing from two extracts, so a reader copying one got a rejected CR), fields
absent from the schema, enums, patterns, `maxItems`, and the `ExactlyOneOf` CEL
rules.
