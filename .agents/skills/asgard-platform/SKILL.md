---
name: asgard-platform
description: The Asgard platform as greppable files - what the platform has, which CR a UI name maps to, how each deployment shape is assembled field by field, what to get from the customer before one can be built, and where each has been got wrong before. Use when writing or reading an Asgard CR or Helm chart, when a customer names something and you need to know what it maps to, before a customer meeting, or before answering any question about what the platform can do. Read aliases.md first when the question came in a language other than English.
---

# The Asgard platform, as files

The platform knowledge an agent in a customer repository does not otherwise
have. **That repository describes one customer's systems and never the platform
those systems run on**; this is the missing half.

    wiki/       what the platform has, and which CR a UI name maps to
    usecase/    how ONE deployment shape is assembled, field by field
    needs/      what to get from the customer before a shape can be built
    brief/      what this activity gets wrong, before you do it
    guide/      which decision to make now, and what reversing it costs
    aliases.md  what a customer said -> what to search for

**[`index.md`](index.md) is the map** - all five in one place as paths, the
rule that turns a pointer into a path, and what is deliberately not here. Start
there. `wiki/index.md` and `usecase/README.md` group their own documents by
the question each answers, and `wiki/README.md` says what a page has to
carry.

**It is generated. Editing it is meaningless** - `asgard-cli init` writes it
from the corpus inside that binary and the next run replaces it, so an edit is
a claim about the platform that no other engagement sees.

## Grep it

    grep -ril "<term>" .
    grep -n "<term>" wiki/processors.md

## Two things grep will not do for you

**Read `aliases.md` before searching a question that arrived in another
language.** The material is English; a term taken from what somebody actually
said matches nothing, and grep reports that identically to a subject the
material genuinely lacks. The file has two tables - words that replace a term,
and names that are added to it.

**Check `wiki/glossary.md` for the word you searched.** A handful of words
mean one thing here and something else to a customer. `payment` is billing
between Asgard and the customer, and also the customer's own payment gateway:
both sets of results are correct, nothing contradicts anything, and the wrong
one reads exactly like an answer. **This is the failure a search cannot
report**, because it found something.

## When the answer is not here

A page that is wrong, and a question these files do not answer, are both worth
filing rather than working around. **Nothing an engagement learns reaches the
next one any other way** - this material is compiled into the binary, so a note
in one repository is a note one repository has:

    asgard-cli issue-report --new

## Staleness

These files came from one binary and a newer one may carry different pages.
Nothing in this directory can tell you which:

    asgard-cli init

That compares what is here against the running binary and reports five states.
**`ahead` is the one worth knowing**: these files were written by a newer
build of this CLI than the one you are running, so your binary is the stale
half and `--force` would be a downgrade.

The platform's own reference material is a separate half with its own record -
`asgard-cli skill status` - because a customer's server can be several
versions from this CLI in either direction, and only the server can say what it
accepts.
