# hack/

This repository's own tooling. Not shipped, not embedded, and not the same thing
as the `scripts/` directory `asgard-cli scaffold` writes into a customer repo -
that one holds the four acceptance gates and the database tool-chain, and is
described in `README.md`.

Everything here answers one question: **is what this repo emits still accepted by
the platform contract?** `helm lint`, `asgard-cli check` and a server-side
dry-run do not answer it. The dry-run is worse than silent, because it drops a
field it does not recognise and reports success while helm's own server-side
apply refuses.

## Reading every instruction at once

    python3 hack/imperatives.py [--ask] [--unmarked]

Three self-contradictions have reached a customer, all found by somebody walking
into one, and the cause is structural: the instructions are spread across ninety
files, so **there is no moment at which two opposing ones are in front of the
same reader.** This makes that moment - it lists every bolded imperative with
its file, which is a few hundred lines and therefore one sitting.

It detects nothing, deliberately. Verb-pair matching over prose produces noise,
and a checker that cries wolf teaches people to change what it can see rather
than what is wrong - a failure this repository has already had, in a deck.

`--ask` narrows to instructions about asking a customer, which is where all
three incidents were. `--unmarked` narrows to those that do not say who they are
for or where the answer goes; most instructions have one obvious reader, so that
list is where to look rather than a list of defects.

**Run it after adding any instruction to the material**, and read the `--ask`
list whole - it is 24 lines and it is the set where every incident so far has
happened. That is the point of the tool: not that it judges, but that the whole
set fits in one reading.

**Its first run found two.** An extract telling a chart author to have asked
about writes during the interview - a meeting they were not at - and a build
stage asking a customer question after the interview had closed. Both were
rewritten to say who they are for.

## Checking a change against the CRDs

Pull asgard-kube first. Validating against a clone from three weeks ago proves
nothing, and it moves without announcing it.

```bash
KUBE=../asgard-kube
git -C $KUBE fetch && git -C $KUBE status -sb        # say so in the PR if behind
mkdir -p .out/crdjson
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

## What this catches that nothing else does

Required fields with no default (a `Workflow` entry's `tooling.allowUploadFile`
was missing from two extracts, so a reader copying one got a rejected CR), fields
absent from the schema, enums, patterns, `maxItems`, and the `ExactlyOneOf` CEL
rules.
