---
name: consistency-checks
description: How to check this repository's material for consistency, and what consistency checking cannot answer. Every mechanical check in one order with what each one is blind to, the method for holding prose against the upstream it came from, the claim shapes that rot, and the rule that ends the loop - compute a number rather than copy one. Use before saying anything here is correct, after changing any document, after pulling an upstream clone, and when a reading of this material turned out to be wrong.
---

# Checking this material

**This is for asgard-fde-cli itself**, not for a customer repository. What a
customer's agent runs is `asgard-cli gate`.

The inventory of what each check covers is `AGENTS.md` under "What is checked,
and what is not". **This is the method**: the order to run them in, what each
one is blind to, and how to do the part no check does.

## The distinction that matters

**Almost every check here is a consistency check.** It asks whether two things
inside this repository agree - a pointer against the file it names, a command
against the command tree, a number against a measurement. Three ask something
harder, and they need a clone: `hack/check-tables.py`, `hack/validate-crs.py`,
`hack/check-coverage.py`.

**Nothing checks whether a sentence is true.** A page can be internally
consistent, point at real files, name real commands, and describe the platform
wrongly. That gap is where every defect found by reading has come from, and no
amount of running checks closes it.

So a claim about this repository says which of the two it rests on.

## Write the list into TASK.md first

**A consistency pass begins by writing down every item, before running any of
them.** `TASK.md` has a section for it - "The consistency pass" - and the skill
fills it in: every check, every surface, its group, and its state.

That order is the point. A pass that runs checks and reports what they said
discovers its own scope as it goes, which is how a surface nobody had listed
gets found in the middle and reported as news. **The list is complete before
the first check runs, or the pass has no scope.**

Derive it from `AGENTS.md`'s inventory - all three groups, not just the
mechanical one - and mark each item `ok`, `fixed`, `needs a clone` or
`not read`. An item nobody looked at says so; that is more useful than its
absence.

## Then run them, most-volatile first

**Not cheapest first.** Cheap-first optimises for the time of whoever is
running the pass. What matters is where the answer is most likely to have
changed since the last one, and that is upstream: nobody here touches
asgard-docs and it moved 286 files in nine days.

**1. The upstream, which moves without anyone touching this repository:**

    hack/sources.py                          what each clone is, and how far behind
    hack/check-tables.py                     the pinned tables against the CRDs
    hack/check-coverage.py                   the coverage row against the docs tree
    hack/validate-crs.py                     generated CRs and extracts against the schemas
    hack/verify-references.sh                the gate over the reference charts

Pull first, or these check a clone rather than the platform.

**2. The material, which moves when somebody edits it:**

    asgard-cli audit-material --links        every pointer resolves, and a path's target lands
    asgard-cli audit-material --bare         a document named with no path
    asgard-cli audit-material --commands     every command named exists
    asgard-cli audit-material --paths        a landed document naming a path only we have
    asgard-cli audit-material --unverified   a document with no provenance marker
    asgard-cli audit-material --sources      every upstream cited is declared
    hack/check-doc-paths.py                  paths and symbols in our own documents

**Build the binary from the working tree first.** Every audit reads what is
embedded, so a check run against an older binary is checking an older corpus.

**3. The code, which the compiler already mostly holds:**

    go build ./... && go vet ./... && gofmt -l . && go test ./...

**4. The network, last, because it is the only one that needs it:**

    asgard-cli audit-material --urls         every documentation link is live

**Then the prose**, which is the part no check does, and the section below is
how. Leave it last because a stale clone makes it worthless, and step 1 is what
tells you whether the clones are stale.

## What each one is blind to

| check | what it will not catch |
|---|---|
| `--links` | a pointer that resolves to the **wrong** page. It checks that the target exists, never that it is the right one |
| `--bare` | a reference with no hyphen in it. `agents` is a CR field, `verify` is a command, and telling those from a page name is not possible in prose |
| `--commands` | a command written in prose without backticks, and a wrong **argument** - that was tried and reported twenty correct lines |
| `--paths` | a path relative to the skill that writes it, which is correct and looks wrong from the root |
| `--unverified` | whether the marker is true. It reports the marker's presence, and for `needs` and `brief` the marker is one shared constant |
| `--sources` | whether a recorded commit is current. Nothing inside this repository can know that; `hack/sources.py` reads the clones |
| `check-tables` | a constraint the CRD expresses in CEL rather than in the schema |
| `validate-crs` | whether the CR does what the page says it does |
| `--orphans`, `--crossref` | nothing - they do not fail. They are listings for a person |

## Holding prose against its source

The part no check does. It is linear in the prose, so do it by claim rather
than by document.

1. **`hack/sources.py`** first. A reading held against a stale clone proves
   nothing, and five of the eight deployment clones have been behind by tens
   of commits at once. `git -C <path> pull` before reading.
2. **Read the diff when the diff is small.** Four commits with one change that
   reaches a page is a real reading. 286 files is not, and claiming it is
   means moving every citation on a reading nobody did.
3. **Move the commit only for what you read.** A page's provenance is per
   claim, so one page legitimately cites two commits of one upstream. Mark a
   commit you have **not** read with ` (unread)` so `--sources` lets it stand.
4. **Take the strongest source available.** For a CR field the CRD beats the
   documentation; for how a field behaves in practice a chart beats both. The
   `effort` field is the example: the docs give its levels, and a chart's own
   values file carries the part that costs money - omitting it is not
   disabling it.

## The claim shapes that rot

Every number this material got wrong was copied from somewhere that moves.
Every judgement it recorded has held.

    rots      a count of anything upstream: CRs, Plugins, pages, nodes
    rots      a field name, an enum value, a file path somebody else owns
    holds     a rule, a trade-off, a failure mode, what a word means here

**So before writing a count, decide whether it can be computed instead.**
`hack/check-coverage.py` exists because a coverage row was hand-counted and
three of its four numbers were wrong; the row cannot be wrong now, because
nobody copies it. A number a script can derive does not belong in prose.

**And before restating a field rule, point at it.** `internal/corpus/wiki/README.md`
says chart-writing cautions belong to the extracts and field rules to the CRD.
Most of what has been wrong here was this repository restating something
another repository owns.

## When a reading turns out to be wrong

Fix the claim, then ask the second question: **would a check have caught it?**

- If yes, and there is no check - write it. `--paths`, `--sources` and
  `check-coverage.py` all began as a defect somebody found by reading.
- If no, say so where the claim lives. A `**Unchecked:**` line that names what
  nobody has held against anything is worth more than a claim that reads as
  settled.
- If a check would fire on correct material, do not write it. That is the
  ninth question in `AGENTS.md`, and an argument-count check failed it.

**Checked:** every command and script named here is in this repository and
does what is said - `--commands` and `hack/check-doc-paths.py` resolve them,
and this file is in the set both read. The ordering claim is checkable too:
`hack/sources.py` reports how far each clone is behind, which is the measure
of what moves without anyone here touching it.

**Unchecked:** whether this order is the best one. It is cheapest-first and
each step rules out a class the next cannot see, which is a design rather than
a measurement.
