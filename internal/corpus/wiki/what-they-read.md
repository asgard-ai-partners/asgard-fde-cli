---
group: In practice
description: the picture a customer arrives with, and where it is wrong
---
# What the customer read before meeting us

The product site and the core-concepts page are what a customer has seen by the
time of the first meeting. They do not describe the platform we deliver, in three
specific ways, and each one produces a conversation an FDE should prepare for.

This page records the vocabulary and the expectations a customer arrives with.

## "Non-technical people can drive it themselves"

The core-concepts page frames Odin as democratisation - natural language, modular tools,
non-technical staff orchestrating AI without a barrier.

Odin is where we build the charts. `../wiki/product-suite.md` describes it as
the build platform, and whoever builds it is usually us. Their staff
live in Sindri and in Mimir.

So a customer may arrive expecting their own people to assemble workflows after
handover. That is not the delivery; say so in the first meeting rather than at
training. In the same conversation, name what their staff do own: dashboards and Views in Mimir, and the
prompts and descriptions on a Managed Agent.

## A vocabulary that maps to nothing here

An unpublished overview page describes three kinds of node tool:

    Basic Function       "the common cases, fill in the parameters"
    Advanced Processor   "finer-grained, for developers"
    Template             "a ready-made workflow to adjust"

No other source has these. The CRD and `../wiki/processors.md` have
their own processor types and no Basic/Advanced split and no Template concept,
and the runtime has no such concept either. The page is marked `draft: true`,
so the documentation site does not show it and a customer did not get the words
from there.

So when a customer says "we will just start from a Template", the word came from
somewhere other than the published documentation - a deck, an older copy, a
conversation. Ask them, in the meeting, where they saw it. Do not adopt the
word, and do not contradict it either until you know whether the editor shows
anything by that name.

## Mimir "simulates the future"

The ecosystem page calls Mimir the AI brain, using LLMs and predictive
algorithms, analysing the past and simulating the future.

What `../wiki/mimir.md` describes is conversational exploration over a
Semantic Model, producing Views and Dashboards. Forecasting is not among the
concepts it lists.

If a customer's evaluation list has a forecast on it, treat it as an open
question rather than a feature, and settle it before the shape is decided, because
the answer changes whether Mimir is the right product for it at all.

## The plan decides the limits

The Fehu plan page says plans, prices and names differ by Workspace and
contract, and the unpublished overview page says how many Projects a Workspace
can create depends on its plan.

The quota page gives 40 Projects and 300 GB, and this material carried those as
flat numbers. They are the numbers for a plan. `../wiki/integration.md` has
the full list and that they are raised through sales; treat every one of them as
plan-dependent rather than as a property of the platform, and check before
quoting one. The step ceiling is the exception: it is a chart field, not a plan
number.

## What to do with this

Ask what they have already read, in the first meeting, before describing
anything. The answer tells you which of these you are correcting. A customer who read the site and a customer who was introduced by a
colleague arrive with different pictures, and the second is usually closer.

## Corresponding extracts

None. This is about expectation rather than assembly.

## Sources

- [Core concepts / the product triangle](https://docs.asgard-ai.com/docs/core-concepts-ecosystem)
  - asgard-docs `21c920f6`, `docs/core-concepts-ecosystem.mdx`: the
  democratisation framing and Mimir simulating the future. Published, and what
  a customer sees first
- [Plan](https://docs.asgard-ai.com/docs/product-suite/fehu/plan)
  - asgard-docs `6261fdff`, `docs/product-suite/fehu/plan.mdx`
- asgard-docs `21c920f6`, `docs/overview/why-asgard.md` - cited as a file
  because it is `draft: true`, like every page under `docs/overview/`, so a
  customer cannot have read it there. It is the only page carrying Basic
  Function, Advanced Processor and Template, and the Workspace-plan-Project
  wording

**Checked:** against asgard-docs `21c920f6` (`docs/overview/`,
`docs/core-concepts-ecosystem.mdx`, `docs/product-suite/fehu/plan.mdx`), asgard-kube `3da0365`
(`pkg/apis/asgard/v1alpha1/types.go`, `ProcessorType`) and the whole of
asgard-core `478cf5d6`, none of which has a Basic Function, Advanced Processor
or Template concept outside that draft page.

**Unchecked:** whether the Odin editor shows anything under those three names,
which only a Console account can settle.
