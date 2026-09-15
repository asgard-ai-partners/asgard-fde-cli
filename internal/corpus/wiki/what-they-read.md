---
group: In practice
description: the picture a customer arrives with, and where it is wrong
---
# What the customer read before meeting us

The product site and the overview pages are what a customer has seen by the time
of the first meeting. **They do not describe the platform we deliver**, in three
specific ways, and each one produces a conversation an FDE can prepare for or
walk into.

This page is not a complaint about marketing. It is the vocabulary and the
expectation somebody arrives with.

## "Non-technical people can drive it themselves"

The overview frames Odin as democratisation - natural language, modular tools,
non-technical staff orchestrating AI without a barrier.

**Odin is where we build the charts.** `../wiki/product-suite.md` says it
plainly: the build platform, and whoever builds it is usually us. Their staff
live in Sindri and in Mimir.

So a customer may arrive expecting their own people to assemble workflows after
handover. That is not the delivery, and it is far cheaper to say so in the first
meeting than at training. **What their staff genuinely own** is worth naming in
the same breath, because it is real: dashboards and Views in Mimir, and the
prompts and descriptions on a Managed Agent.

## A vocabulary that maps to nothing here

The overview describes three kinds of node tool:

    Basic Function       "the common cases, fill in the parameters"
    Advanced Processor   "finer-grained, for developers"
    Template             "a ready-made workflow to adjust"

**No other source has these.** The CRD and `../wiki/processors.md` have
their own processor types and no Basic/Advanced split and no Template concept.
The overview page is also one the index already flags as possibly stale.

So when a customer says "we will just start from a Template", they are using a
word from a page rather than from the product as it is described anywhere else.
**Ask them, in the meeting, which page they saw it on.** Do not adopt the word, and do not contradict it either
until you know whether that surface still exists.

## Mimir "simulates the future"

The ecosystem page calls Mimir the AI brain, using LLMs and predictive
algorithms, analysing the past and simulating the future.

What `../wiki/mimir.md` describes is conversational exploration over a
Semantic Model, producing Views and Dashboards. **Forecasting is not among the
concepts it lists.**

If a customer's evaluation list has a forecast on it, that is a question rather
than a feature - and it is one to settle before the shape is decided, because
the answer changes whether Mimir is the right product for it at all.

## The plan decides the limits

The overview says a Workspace has a price plan, and that **how many Projects can
be created depends on it**.

The quota page gives 40 Projects and 300 GB, and this material carried those as
flat numbers. They are the numbers for a plan. `../wiki/integration.md` has
the full list and that they are raised through sales; treat every one of them as
plan-dependent rather than as a property of the platform, and check before
quoting one.

## What to do with this

**Ask what they have already read**, in the first meeting, before describing
anything. It costs one sentence and it tells you which of these you are
correcting. A customer who read the site and a customer who was introduced by a
colleague arrive with different pictures, and the second is usually closer.

## Corresponding extracts

None. This is about expectation rather than assembly.

## Sources

- [Core concepts / the product triangle](https://docs.asgard-ai.com/docs/core-concepts-ecosystem)
  and asgard-docs `docs/overview/why-asgard.md` - the second is cited as a file
  because it is `draft: true`, and **that matters more here than anywhere else
  on this page: a customer cannot have read it.** A page about what they
  arrived having read must not count an unpublished one
  - asgard-docs `f00e0ee`, read 2026-09-02. Neither had been read into this
  material before; both are what a customer sees first

**Unchecked:** whether the Basic Function / Advanced Processor / Template
surface still exists in Odin. Nothing else describes it and the index already
marks the neighbouring overview page as possibly stale, but nobody has opened
the product to look. Until somebody does, it is a vocabulary to ask about rather
than one to correct.
