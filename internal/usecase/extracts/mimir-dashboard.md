# A read surface for dashboards, with no agent on it

A SemanticLayer whose consumer is **Data Insight (Mimir)** rather than an Agent.
The customer explores it by conversation and saves the useful answers as Views
and Dashboards; nothing in the chart mounts the layer.

**Seen in:** a finance deployment whose entire architecture is stated as one
rule - two read paths in one chart that never touch, three domain Agents on a
GraphQL API and three SemanticLayers that exist only for Mimir.

**Checked:** 2026-09-02 against that deployment's chart and its AGENTS.md - three
layers, 31 cubes, 510 dimensions, 21 joins, `Agent.spec.managed.semanticLayers[]`
empty on all three Agents - and against the CRD.

**Unchecked:** what the customer's Mimir side looks like. Views, Dashboards and
Knowledge are created in the product by the people who use it, not declared in a
chart, so nothing here can be held against a deployment.

**Read the platform side first:** `asgard-cli wiki mimir` -
Thread, View, Dashboard and Knowledge, and that Mimir reads the models Odin
builds and never edits them. This page assumes you have.

## When this shape, and when not

**The question is what they do with the answer**, and it is asked in the
interview at 2b, next to the one about who is on the other end:

    look one thing up, in the moment      an agent
    watch the same numbers every day      this shape
    both, for different people            both - two deliveries, one model

`asgard-cli wiki product-suite` says it in one line: **Mimir is often what the
customer actually wants.** "I want an AI that answers stock questions" is a
statement about stock questions, not about an agent. Someone glancing at a
figure every morning wants a page that is already open; an agent makes them ask
for it again each time, and they will stop.

**The two are not exclusive and the model is shared.** The expensive mistake is
not choosing wrong - it is not asking, building the agent, and discovering at
the demo that what they wanted was a dashboard. The modelling survives either
way; the delivery does not.

When **not** this shape: a question whose answer changes what someone does in
the next minute, anything a person needs while a customer is on the phone,
anything with a write in it. Mimir reads. A dashboard cannot approve a transfer.

## The shape

    DataConnector  dc-<name>     read-only credentials to the database
      <- SemanticLayer  sl-<name>    cubes, dimensions, joins
           consumer: Mimir, in the product. Nothing in the chart references it.

That is the whole chart side. There is **no Agent, no Toolset and no
SandboxBlueprint** in this path, and that is the point rather than an omission.

**One layer per section of what they will look at.** The deployment splits three
layers along the dashboard areas of the customer's own site - fundamentals,
chips, market - rather than along the database's tables. The layer is a reading
surface for a person exploring, so it is organised the way they think about the
subject, not the way the schema grew.

## The trap, and the gate does not catch it

**A SemanticLayer that no Agent references looks exactly like a mistake.** The
next person to read the chart - or the next agent - sees a layer nobody mounts,
compares it against the other engagements where every Agent binds one, and
helpfully binds it.

Doing that gives Agents that were deliberately restricted to an API a second,
unreviewed path straight into the database. **It is a behaviour change presented
as a fix**, and it is the single most likely way this shape gets destroyed.

Two things follow:

- **Say so in the chart.** The deployment carries the reason as a comment on the
  layer itself and as a rule in its AGENTS.md, in the file the person doing the
  "fixing" is already looking at. A rule that lives only in a review comment
  loses to a plausible-looking diff
- **Do not expect the gate to hold it.** Cross-reference checking validates the
  references that exist; it has nothing to say about one that should not. A
  green gate is not evidence this survived, and `asgard-cli verify` is the same
  - it reports an Agent with no capability, never a layer with no consumer

## Fields that are not obvious

**`allowedCubes` still matters, for the same reason as anywhere else.** A layer
without it is arbitrary SQL over every cube, and the surface grows by itself
each time a table is added. That a person rather than an agent is on the other
end does not change it - it changes who is surprised.

**The credential is read-only and this is the one shape where that is easy to
get.** There is no write path to argue for, so ask for a read replica.

**Test every cube against the live database before handing it over.** The
deployment carries a script that does exactly this and reports per cube, because
a cube that does not resolve fails inside Mimir as an unhelpful error in front of
the customer, not at deploy time. 31 cubes is more than anyone checks by hand.

## What it cost someone

The deployment's own AGENTS.md leads with this rule rather than filing it under
conventions, and names the two other engagements whose charts look different, so
that a reader who has seen those does not conclude this one is broken. That is
the shape of the incident: nothing failed, and the risk was a later maintainer
being reasonable.

## Read the platform side first

`asgard-cli wiki mimir` for what a Thread, View and Dashboard are;
`asgard-cli wiki semantic-model` for the modelling flow and its limits;
`asgard-cli wiki product-suite` for which product a request belongs to.
