---
group: Products and scope
description: Thread, View, Dashboard, Knowledge
---
# Mimir - Data Insight

Explores data by conversation and produces charts and dashboards. It reads the
Semantic Models built in Odin and does not edit them.

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

A View saves a result from a Thread. It is created from a reply's Data result, then View data, then Create View. The dialog
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

Question-SQL pairs are a question plus the SQL your team accepts for it.
Mimir refers to that shape for similar questions instead of deriving one afresh.
The typical use is settling what the data itself cannot: when one table carries
both `store_id` and `store_name`, and both `on_hand` and `safety_stock`, which
two columns a shortfall subtracts and whether a store shows as a code or a name
are team conventions.

Instructions set the definitions and formats an answer has to follow.

Knowledge belongs to the currently selected semantic model, not to the
Project. Knowledge created under model A is invisible under model B.

It also belongs to the person asking. The platform describes it as what one end
user has accumulated against one model - their own question-SQL pairs and their
own instructions - and the Data Insight service attaches the asker's knowledge
to each question. An instruction is either global, applying to every question,
or matched, applying only when the question resembles one it lists. The list is
capped per question and cut from the tail, so the service sends the most
relevant entries first rather than everything a person has saved.

## Mimir or an agent

When a customer says "I want an AI that answers stock questions", ask what they
do with the answer: glancing at it each morning is a Dashboard, looking one thing
up is an agent. Choosing the wrong one builds something the customer does not
keep using.

## Corresponding extracts

Mimir reads Semantic Models and a chart writes no CR for it. The platform
derives Mimir's chat Workflow, BotProvider and SandboxBlueprint from each
`SemanticLayer` itself. Building the model is `../usecase/semantic-layer.md`.

## Sources

- [Thread](https://docs.asgard-ai.com/docs/product-suite/mimir/features/thread),
  [View](https://docs.asgard-ai.com/docs/product-suite/mimir/features/view),
  [Dashboard](https://docs.asgard-ai.com/docs/product-suite/mimir/features/dashboard),
  [Knowledge](https://docs.asgard-ai.com/docs/product-suite/mimir/features/knowledge),
  [Data Model](https://docs.asgard-ai.com/docs/product-suite/mimir/features/data-model)
  - asgard-docs `6261fdff`

**Checked:** Knowledge being per user and per semantic model, the two
instruction scopes, the cap, and the resources derived from a `SemanticLayer`,
against asgard-core `478cf5d6` `internal/bpoperator/reconciler/sl_reconciler.go`.

**Unchecked:** the Thread, View and Dashboard screens come from the product
documentation only, and only a Console account can show them.
