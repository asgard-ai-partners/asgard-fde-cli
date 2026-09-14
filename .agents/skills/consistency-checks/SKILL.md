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
harder, and they need a clone: `go run ./hack tables`, `go run ./hack validate-crs`,
`go run ./hack coverage` and `go run ./hack sources`.

**Nothing checks whether a sentence is true.** A page can be internally
consistent, point at real files, name real commands, and describe the platform
wrongly. That gap is where every defect found by reading has come from, and no
amount of running checks closes it.

**And consistency is not capability.** Every check above can pass while the
tool has stopped doing what `Goal.md` says it is for: make `asgard-cli init`
require a session and the corpus is still consistent, every pointer still
resolves, and Goal's first point is gone. `go run ./hack goal` is the one
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
`go run ./hack sources` compares each recorded reading against its clone and says
which have gone behind. **That a reading happened is nobody's to verify but
whoever claims it** - write `never` when it has not, because a row that says
so is worth more than one that reads settled.

**Then run `go run ./hack pass-list`.** It compares no names - the pass is
printed rather than copied - and what it holds is the part no program can
derive: that every check says what it needs, that every prose surface in
`TASK.md` records the date it was read, and that this repository's own
maintenance skill never appears in a scaffolded tree.

**The prose surfaces live in one document, and that is `TASK.md`.** They were
in two, worded for each document's own context, and the check's whole job was
holding one copy against the other - so a rename in one was a failure in the
other and the repair was to retype it. **A row records when it was read**,
because that is the one thing a program can act on: `go run ./hack sources`
cannot say whether a source has moved since a reading that gives no date.
`pass-list` fails a row without one, and fails a second table of the same
surfaces appearing anywhere else.

**And know what that check cannot do.** It verifies that the list is
complete - never that anything on it passed. A table whose cells are verdicts
is a table CI validates the shape of and not the content, which is how a
complete list of unverifiable claims comes to read like assurance. That is why
the mechanical rows carry no state at all.

## Then run them, most-volatile first

    go run ./hack pass

**That prints the pass, in the order to run it, derived from the binary's own
flags and the gate's own subcommands - so it is not written down anywhere,
including here.** A list of checks kept in a document drifts from the checks;
four documents were carrying one and each had drifted in its own direction.

**Not cheapest first**, which is the order that optimises for the time of
whoever is running the pass. What matters is where the answer is most likely to
have changed:

  - **Upstream first.** Nobody here touches asgard-docs or asgard-kube, and they
    move without anyone noticing. **Pull before running them**, or they check a
    clone rather than the platform.
  - **Then the material** - and **build the binary from the working tree first**,
    because every audit reads what is embedded, so an older binary checks an
    older corpus.
  - **Then the compiler**, which already holds most of what it can.
  - **The network last**, because it is the only one that needs it, and a third
    party's outage is not this repository's failure.

**Then the prose**, which is the part no check does, and the section below is
how. Leave it last because a stale clone makes it worthless, and step 1 is what
tells you whether the clones are stale.

## What each one is blind to

**The names are the gate's own** - `go run ./hack list` prints them with what
each is for; this table is the other half, which the tool cannot print.

| check | what it will not catch |
|---|---|
| `--links` | a pointer that resolves to the **wrong** page. It checks that the target exists, never that it is the right one |
| `--bare` | a reference with no hyphen in it. `agents` is a CR field, `verify` is a command, and telling those from a page name is not possible in prose |
| `--commands` | a command written in prose without backticks, and a wrong **argument** - that was tried and reported twenty correct lines |
| `--paths` | a path relative to the skill that writes it, which is correct and looks wrong from the root |
| `--unverified` | whether the marker is true. It reports the marker's presence, and for `needs` and `brief` the marker is one shared constant |
| `--sources` | whether a recorded commit is current. Nothing inside this repository can know that; `go run ./hack sources` reads the clones |
| `tables` | a constraint the CRD expresses in CEL rather than in the schema |
| `validate-crs` | whether the CR does what the page says it does |
| `coverage` | whether the pages behind the numbers say anything true. It counts them. `--drift` names the pages that have moved and never says what changed in one |
| `processors` | what a key **means**. The definitions say whether a processor takes dynamic config; that an extra key on `http-request` is an HTTP header is in the loop that reads it, and no table upstream states it |
| `counts` | a count of something nobody upstream counts. It recomputes what a deployment's own documents state, and a number invented here has nothing to be held against |
| `spec-key-gap` | whether a key `add` writes is written **well**. It compares key sets, so a field emitted with the wrong value counts as covered |
| `pass-list` | **whether any check passed**, and **whether a listed surface was actually read.** It holds the shape of the list - that every check says what it needs, that every prose surface carries a date, that there is no second table of them - and a verdict is not in its reach |
| `goal` | whether the material is any good. It asks whether the capability is there - the corpus lands, a grep finds things, a chart gets written, the issue route is printed - never whether what landed is right |
| `sources` | whether a reading happened. It compares a recorded reading with its clone and reports staleness; the reading itself is nobody's to verify but whoever claims it |
| `doc-paths` | whether a document's prose is right. It resolves the paths, the Go symbols and whether every command in the tree is named in both READMEs, and says nothing about what the sentence around one claims |
| `--unchecked` | nothing - it does not fail. It prints what every document says it has **not** been held against, which is where the blocked list comes from now instead of a section in `TASK.md` that had to be maintained |
| `--orphans`, `--crossref`, `--ask`, `--unmarked`, no flag | nothing - they do not fail. They are listings for a person |

## The one that is a query rather than a check

    asgard-cli audit-material --term <field>

**Run it whenever you change anything a second place might restate** - a
renamed platform field, a rule reworded, a check whose behaviour moved. It has
no pass or fail - it answers a question somebody asks it - so it is here and
not in `TASK.md`'s pass, which lists checks that can be run and answered
without being told what to look for.

**Let its scope be the scope.** It reads every surface this repository is
responsible for: the material, the scaffold templates, every `--help` screen,
the CLI's own string literals, these documents, the maintenance skills and the
gate under `hack/` - whose `What:` strings are what `go run ./hack list` prints. A
claim is taught in several of those at once - a template that writes a field,
an extract that explains it, a prompt that mentions it, a help screen that
names it - and fixing one leaves the rest teaching what is no longer true.

**Do not decide that set by hand.** Choosing which files to sweep is the same
mistake as writing down a list that can be generated, and it fails the same
way: the set you remember is the set you have been editing, which is not where
the other copy is. A sweep that names its own files comes back clean over the
files it never opened, and reads exactly like a sweep that found nothing.

**The same rule one level up: the patterns are not yours either.**
`go run ./hack introduced` reads the lines THIS change adds and lists every
count-shaped one, with no phrase list to remember. Over the whole corpus that
detector reports more than a thousand lines and is useless; over a diff it
reports a few dozen and is the item above that gets skipped. **Existing counts
are a backlog no check closes** - they are found by reading, and saying the
sweep is finished is the claim that keeps being wrong. The ones a change adds
are bounded, and they are the regression.

## Holding prose against its source

The part no check does. It is linear in the prose, so do it by claim rather
than by document.

0. **`asgard-cli audit-material --unchecked`** first, because it is the scope.
   Every document names the surface it has not been held against, and reading
   that output is how you find out what a pass is for before spending it.
   **This replaced a hand-written list**, four of whose entries outlived the
   thing they described.
1. **`go run ./hack sources`** next. A reading held against a stale clone proves
   nothing, and most of the deployment clones have been behind by tens
   of commits at once. `git -C <path> pull` before reading.
2. **Let `--drift` set the scope rather than the diff.**
   `go run ./hack coverage --drift` lists every cited page that has moved since
   the commit the citing document names. That turns "the wiki against
   asgard-docs" into the handful that moved, which is a reading somebody can
   actually do - where reading the whole diff is not, and claiming it is means moving every
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

**A count on a provenance line is not evidence of a reading.** `**Checked:**`
records what was read and the commit; a tally beside it - "against 72 gated and
14 ungated tool entries" - says only that something was counted, while reading
exactly like a statement that somebody looked. **The right number is the
dangerous one**: it makes the check green and the page unread, and everyone who
passes it inherits whatever the sentence next to it says. Name the scope
instead - every CR of that kind, in these deployments, at this commit - because
a scope cannot be satisfied by counting.

**Before asking whether a count can be computed, ask what the reader does with
it.** That question comes first and it is the one that gets skipped. Three
kinds, and only one of them earns a number:

    the count IS the claim        how much of a chart `add` never writes, how
                                  many pages a back office has, how far a clone
                                  is behind. Keep it, and compute it.
    the count is evidence         "does an arrow function evaluate" - the
                                  answer is yes or no, and a tally is a weak
                                  way to say yes. Name the deployed example
                                  instead: it cannot rot, and it is checkable
                                  by opening one file.
    the count is decoration       "the 13 processors" in an index row. The row
                                  reads the same without it and cannot go
                                  stale. Delete it.

**A tally standing in for a yes is the one that keeps going wrong.** "One arrow
function in 520 values" was recounted three times, was wrong every time, and
each recount used a different denominator - while `prevBlobs.map(b => b.blobId)`
in a shipped chart answered the same question and could not have been wrong.
**Writing a checker for a number like that is treating the symptom**: the number
never needed to be there.

**An index carries no arithmetic.** `internal/corpus/wiki/index.md` had a
seven-row ledger deriving its coverage figure by hand and, beside it, a record
of what the figure used to be - which is a changelog, and a changelog is what a
count grows when nothing can recompute it. An index says which page answers a
question. One number survives there, the row `go run ./hack coverage`
recomputes and fails on.

**So before writing a count, decide whether it can be computed instead.**
`go run ./hack coverage` exists because a coverage row was hand-counted and
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
  `go run ./hack coverage` all began as a defect somebody found by reading.
- If no, say so where the claim lives. A `**Unchecked:**` line that names what
  nobody has held against anything is worth more than a claim that reads as
  settled.
- If a check would fire on correct material, do not write it. `AGENTS.md` asks
  that under "Would this check fire on material that is correct?", and an
  argument-count check failed it.

**Checked:** every command and script named here is in this repository and
does what is said - `--commands` and `go run ./hack doc-paths` resolve them,
and this file is in the set both read. The ordering claim is checkable too:
`go run ./hack sources` reports how far each clone is behind, which is the measure
of what moves without anyone here touching it.

**Unchecked:** whether most-volatile-first is the best order. That upstream
moves most is measured - `go run ./hack sources` reports it - but that running it
first finds more, sooner, is a design rather than a measurement. What is
settled is that cheapest-first was the wrong principle: it optimised for the
time of whoever runs the pass.
