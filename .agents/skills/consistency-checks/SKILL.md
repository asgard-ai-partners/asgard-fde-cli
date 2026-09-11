---
name: consistency-checks
description: How to check this repository's material for consistency, and what consistency checking cannot answer. Writing the pass's scope into TASK.md before running anything, running most-volatile first because upstream moves without anyone here touching it, what each check is blind to, the method for holding prose against the upstream it came from, and the two rules that end the loop - compute a number rather than copy one, and record no verdict a program could have produced. Use before saying anything here is correct, after changing any document, after pulling an upstream clone, and when a reading of this material turned out to be wrong.
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
against the command tree, a number against a measurement. Four ask something
harder, and they need a clone: `hack/check-tables.py`, `hack/validate-crs.py`,
`hack/check-coverage.py` and `hack/sources.py`.

**Nothing checks whether a sentence is true.** A page can be internally
consistent, point at real files, name real commands, and describe the platform
wrongly. That gap is where every defect found by reading has come from, and no
amount of running checks closes it.

**And consistency is not capability.** Every check above can pass while the
tool has stopped doing what `Goal.md` says it is for: make `asgard-cli init`
require a session and the corpus is still consistent, every pointer still
resolves, and Goal's first point is gone. `hack/check-goal.py` is the one
check here that runs the tool the way Goal describes - in a temporary
directory with no network, no account and no git repository - and it is the
only one that writes files.

So a claim about this repository says which of the three it rests on: a
consistency check, a capability check, or somebody's reading.

## Write the list into TASK.md first

**A consistency pass begins by writing down every item, before running any of
them.** `TASK.md` has a section for it - "The consistency pass" - and it holds
two different things:

    every check this repository has      by name, and nothing else
    every surface no check reaches       what the reading was held against

That order is the point. A pass that runs checks and reports what they said
discovers its own scope as it goes, which is how a surface nobody had listed
gets found in the middle and reported as news. **The list is complete before
the first check runs, or the pass has no scope.**

**Record no verdict for a mechanical check.** `ok` beside `--links` is a
result copied out of a script that can produce it, in prose nothing can
verify - which is the same shape as every count this material got wrong. The
pass lists those checks by name; their answer is their exit code, today.

**The only group with a state is the one no program can answer**, and what it
records is not a verdict but **what the reading was held against**. That
matters because it makes the one checkable thing about a reading checkable:
`hack/sources.py` compares each recorded reading against its clone and says
which have gone behind. **That a reading happened is nobody's to verify but
whoever claims it** - write `never` when it has not, because a row that says
so is worth more than one that reads settled.

**Then run `hack/check-pass-list.py`, because a list derived from a document
is only as complete as that document.** It reads the binary's own flags and
`hack/`'s own contents rather than another list, and fails on a check the pass
does not name - `--ask`, `--unmarked`, `--term`, `extract-crs.py` and all four
Go steps were absent from the inventory too, so deriving faithfully produced
the same holes.

**The prose surfaces carry a slug, in both documents.** They are the group
that cannot be discovered from the binary, so they are the group that drifts;
comparing their wording drifts with them, because each list words a surface
for its own context. `check-pass-list.py` compares the slugs, both ways, so
a surface in one list and not the other is reported rather than assumed.

**And know what that check cannot do.** It verifies that the list is
complete - never that anything on it passed. A table whose cells are verdicts
is a table CI validates the shape of and not the content, which is how a
complete list of unverifiable claims comes to read like assurance. That is why
the mechanical rows carry no state at all.

## Then run them, most-volatile first

**Not cheapest first.** Cheap-first optimises for the time of whoever is
running the pass. What matters is where the answer is most likely to have
changed since the last one, and that is upstream: nobody here touches
asgard-docs and it moved 286 files in nine days.

**1. The upstream, which moves without anyone touching this repository:**

    hack/sources.py                          what each clone is, how far behind,
                                             and which recorded reading is now stale
    hack/sources.py --extracts               how far each extract's source chart has moved
    hack/check-tables.py                     the pinned tables against the CRDs, and
                                             every CEL-rule count stated anywhere here
    hack/check-processors.py                 processors.md's three tables against
                                             asgard-core and asgard-docs
    hack/check-counts.py                     every count taken off a deployment
    hack/check-coverage.py                   the coverage row against the docs tree
    hack/check-coverage.py --drift           which cited pages have moved since the
                                             commit the citing document names
    hack/spec-key-gap.py                     how much of a production chart `add` never writes
    hack/extract-crs.py                      pulls the CRs out of the extracts, for the next line
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
    hack/check-pass-list.py                  this list against the binary and hack/
    hack/check-goal.py                       Goal.md's four points, against the binary

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
| `check-coverage` | whether the pages behind the numbers say anything true. It counts them. `--drift` names the pages that have moved and never says what changed in one |
| `check-processors` | what a key **means**. The definitions say whether a processor takes dynamic config; that an extra key on `http-request` is an HTTP header is in the loop that reads it, and no table upstream states it |
| `check-counts` | a count of something nobody upstream counts. It recomputes what a deployment's own documents state, and a number invented here has nothing to be held against |
| `spec-key-gap` | whether a key `add` writes is written **well**. It compares key sets, so a field emitted with the wrong value counts as covered |
| `check-pass-list` | **whether any check passed.** It holds the list against the binary and `hack/`, and a verdict is not in its reach |
| `check-goal` | whether the material is any good. It asks whether the capability is there - the corpus lands, a grep finds things, a chart gets written, the issue route is printed - never whether what landed is right |
| `sources.py` | whether a reading happened. It compares a recorded reading with its clone and reports staleness; the reading itself is nobody's to verify but whoever claims it |
| `--unchecked` | nothing - it does not fail. It prints what every document says it has **not** been held against, which is where the blocked list comes from now instead of a section in `TASK.md` that had to be maintained |
| `--orphans`, `--crossref`, `--ask`, `--unmarked`, no flag | nothing - they do not fail. They are listings for a person |

## The one that is a query rather than a check

    asgard-cli audit-material --term <field>

**Run it when a platform field is renamed or retired**, and only then. It has
no pass or fail - it answers a question somebody asks it - so it is here and
not in `TASK.md`'s pass, which lists checks that can be run and answered
without being told what to look for.

It reads the templates as well as the prose. A renamed field is taught in
three places - a template that writes it, an extract that explains it, a
prompt that mentions it - and fixing one leaves the other two teaching a field
that no longer exists.

## Holding prose against its source

The part no check does. It is linear in the prose, so do it by claim rather
than by document.

0. **`asgard-cli audit-material --unchecked`** first, because it is the scope.
   Every document names the surface it has not been held against, and reading
   those 74 lines is how you find out what a pass is for before spending it.
   **This replaced a hand-written list**, four of whose entries outlived the
   thing they described.
1. **`hack/sources.py`** next. A reading held against a stale clone proves
   nothing, and five of the eight deployment clones have been behind by tens
   of commits at once. `git -C <path> pull` before reading.
2. **Let `--drift` set the scope rather than the diff.**
   `hack/check-coverage.py --drift` lists every cited page that has moved since
   the commit the citing document names. That turned "the 27 wiki pages against
   asgard-docs" into five pages, which is a reading somebody can actually do -
   where reading a 286-file diff is not, and claiming it is means moving every
   citation on a reading nobody did.
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
    rots      a distance between two things that both move
    holds     a rule, a trade-off, a failure mode, what a word means here
    holds     a count with a commit beside it

**A count with no commit beside it cannot be checked twice.** That is the whole
difference between the two count rows: 29 Plugins is meaningless and "29
Plugins at `edb0ad0`" is permanent, because the commit does not move. A
distance cannot be rescued the same way - "26 commits behind" compares two
moving things, and `source/SOURCES.md` carried a column of those in which both
non-zero rows had rotted within nine days. Compute a distance, never write one.

**The worst of them is a count copied out of a document that states its own.**
Nothing about 88 reads differently from 93. A pass set out to recount SHOPLINE's
page ledger, took a figure off a different tally, and wrote it into seven
places, where it sat looking exactly as authoritative as the truth - while the
file it came from said 88 in two of its own headings and asserted it with a
script. If the source counts itself, read its count; if it does not, compute
yours and leave the method beside it.

**And two numbers that describe the same thing are the ones to be most careful
with.** 79 is the `XValidation` markers in asgard-kube's Go types; 231 is what
the generator emits from them. This material had the marker count written down
as the CRDs' own, in the one page whose subject is what the CRDs enforce. A live
URL against a file path is the same trap: eight processor pages answer at a name
that is not their file's, which made the coverage row understate itself by six
pages and put a 404 into a naming table twice, in opposite directions.

**So before writing a count, decide whether it can be computed instead.**
`hack/check-coverage.py` exists because a coverage row was hand-counted and
three of its four numbers were wrong; the row cannot be wrong now, because
nobody copies it. A number a script can derive does not belong in prose.

**And before restating a field rule, point at it.** `internal/corpus/wiki/README.md`
says chart-writing cautions belong to the extracts and field rules to the CRD.
Most of what has been wrong here was this repository restating something
another repository owns.

**The same rule applies to a check's own result.** `ok` written beside
`--links` is a copied number wearing a different hat: a script produces it, so
prose must not. What a document may record about a check is its name; what it
may record about a reading is what the reading was held against. Anything
else is a claim nothing can hold against anything.

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

**Unchecked:** whether most-volatile-first is the best order. That upstream
moves most is measured - `hack/sources.py` reports it - but that running it
first finds more, sooner, is a design rather than a measurement. What is
settled is that cheapest-first was the wrong principle: it optimised for the
time of whoever runs the pass.
