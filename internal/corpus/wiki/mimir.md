---
group: Products and scope
description: Thread, View, Dashboard, Knowledge
---
# Mimir - Data Insight

Explores data by conversation and produces charts and dashboards. It reads the
Semantic Models built in Odin and **does not edit them**.

Four concepts, from the most transient to the most fixed:

```
Thread     one question and answer - the act of exploring
View       a chart saved with the SQL behind it and the chosen presentation
Dashboard  several Views on one page, for ongoing tracking
Knowledge  the organisation's own query conventions
```

## Thread

Sending a question in the prompt box creates a Thread. A reply has several parts:

- the execution steps, expandable, showing what it actually did
- an automatically generated conversation title
- a result card: Data result when only a table came back, Chart result when a
  chart did
- the conclusion and observations
- suggested follow-up questions
- a prompt for feedback

Follow-ups within the same Thread keep the context.

Check the semantic model selected at the top before asking. Asked about something
that model does not cover, Mimir says the data at hand cannot answer it and lists
what the model actually can answer, rather than improvising.

## View

A Thread is the exploring; a View is the conclusion kept.

Created from a reply's Data result, then View data, then Create View. The dialog
has Name (required, prefilled with the original question and usually worth
rewriting into a chart title), Description, SQL, and Visualization.

The Visualization list always offers Table, plus one entry per chart Mimir drew
for that answer. So a reply that only produced a table can only be saved as a
table - to have a chart to choose, ask for one when asking the question.

## Dashboard

Several Views on one page. Suited to metrics already asked about and confirmed
worth watching.

Creating one needs only a name. The page offers + View (a three-step wizard:
Select Views, Preview, done - several at once), Refresh, Share and Edit Mode for
sizing and placing charts.

## Knowledge

Stores the organisation's own query conventions so answers match expectations
without restating them each time. Split into Question-SQL pairs and Instructions.

**Question-SQL pairs** are a question plus the SQL your team accepts for it.
Mimir refers to that shape for similar questions instead of deriving one afresh.
The typical use is settling what the data itself cannot: when one table carries
both `store_id` and `store_name`, and both `on_hand` and `safety_stock`, which
two columns a shortfall subtracts and whether a store shows as a code or a name
are team conventions.

**Instructions** set the definitions and formats an answer has to follow.

Knowledge belongs to **the currently selected semantic model**, not to the
Project. Knowledge created under model A is invisible under model B.

## Mimir or an agent

When a customer says "I want an AI that answers stock questions", ask what they
do with the answer: glancing at it each morning is a Dashboard, looking one thing
up is an agent. Getting that wrong builds something used once and left.

## Corresponding extracts

Mimir reads Semantic Models and produces no CRs of its own. Building the model is
`../usecase/semantic-layer.md`.

## Sources

- [Thread](https://docs.asgard-ai.com/docs/product-suite/mimir/features/thread),
  [View](https://docs.asgard-ai.com/docs/product-suite/mimir/features/view),
  [Dashboard](https://docs.asgard-ai.com/docs/product-suite/mimir/features/dashboard),
  [Knowledge](https://docs.asgard-ai.com/docs/product-suite/mimir/features/knowledge),
  [Data Model](https://docs.asgard-ai.com/docs/product-suite/mimir/features/data-model)
  - asgard-docs `f00e0ee`

**Unchecked:** everything here comes from the product documentation. Mimir
produces no CRs, so there is no chart to hold it against.
