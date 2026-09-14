# A read surface for dashboards, with no agent on it

A SemanticLayer whose consumer is **Data Insight (Mimir)** rather than an Agent.
The customer explores it by conversation and saves the useful answers as Views
and Dashboards; nothing in the chart mounts the layer.

**Seen in:** a finance deployment whose entire architecture is stated as one
rule - two read paths in one chart that never touch, domain Agents on a
GraphQL API and SemanticLayers that exist only for Mimir.

**Checked:** 2026-09-02 against that deployment's chart and its AGENTS.md -
every layer, every cube and join in them, and
`Agent.spec.managed.semanticLayers[]` empty on every Agent - and against
the CRD.

**Unchecked:** what the customer's Mimir side looks like. Views, Dashboards and
Knowledge are created in the product by the people who use it, not declared in a
chart, so nothing here can be held against a deployment.

**Read the platform side first:** `../wiki/mimir.md` -
Thread, View, Dashboard and Knowledge, and that Mimir reads the models Odin
builds and never edits them. This page assumes you have.

## When this shape, and when not

**The question is what they do with the answer**, and it is asked in the
interview at 2b, next to the one about who is on the other end:

    look one thing up, in the moment      an agent
    watch the same numbers every day      this shape
    both, for different people            both - two deliveries, one model

`../wiki/product-suite.md` says it in one line: **Mimir is often what the
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

**One layer per section of what they will look at.** The deployment splits its
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
- **Expect the gate to report it, and expect that to be the wrong reading.**
  `asgard-cli verify` has R11 for exactly this shape: *a SemanticLayer that no
  Agent binds*. It is a warning and its own rule text says why - the render
  cannot tell "deliberately unbound" from "somebody has not finished the read
  path", and neither can this tool. **So on this shape R11 always fires and is
  always wrong**, which is the thing to write in the chart before the next
  reader treats it as work to do. A green gate is not evidence this survived;
  neither is a red one evidence it did not

## Fields that are not obvious

**There is no `allowedCubes` here, and no second setting that narrows the
surface.** `allowedCubes` is not a field on `SemanticLayer` at all - it lives on
`Agent.spec.managed.semanticLayers[]` and on the two LLM processors' config, and
this shape has neither. Even where it exists, `gate` R4 refuses an Agent that
sets one, on the standing decision that a bound layer is queryable in full.

So **the only exposure control in this shape is which cubes and dimensions exist
in the layer.** Adding one widens what a person can reach, and nothing can narrow
it again. That is a stronger statement than "restrict it later", and it decides
things: a table that keeps credential columns - a password, a hardware-key serial
- beside the display name is not defensible behind a whitelist here, because
there is no whitelist. Either the column is not modelled, or it is exposed.

**Write that reason next to the cube**, because "not modelled" is a discipline
and the next person adding a dimension will not know about it. The `instruction`
field and a comment on the cube are where it survives a diff.

**`sampleQuestions` is what the product's front page shows, and this is the
only field in the chart that reaches it.** Data Insight renders each string as a
button under the layer's prompt box; a reader clicks one and it is sent as their
question. With the array empty they get a blank box and have to guess what the
layer knows, which for an internal audience exploring by conversation is the
difference between a page somebody uses on day one and one they open once.

    spec:
      sampleQuestions:
        - 哪些客戶的 KYC 狀態為待覆審(review_due)?
        - 各客戶類型的委託資產規模分布如何?

Plain strings, not objects. Write them the way the audience refers to things,
anchor each to a subject that is actually in the layer, and keep each one
self-contained - a button is read with no context around it.

**`Agent.spec.managed.sampleQuestions` is a different field on a different CRD,
and its rules do not transfer.** `../usecase/agent-hub.md` has a whole
section for that one, and `gate` R7 enforces a minimum of two on a published
Agent - our rule, not the platform's. Searching for the field name reaches the
Agent guidance and nothing else, which is worse than reaching nothing: it is
confident, well-written and about another CR. There is no minimum here, and no
gate rule; there is just a front page that is blank until somebody fills it.

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

`../wiki/mimir.md` for what a Thread, View and Dashboard are;
`../wiki/semantic-model.md` for the modelling flow and its limits;
`../wiki/product-suite.md` for which product a request belongs to.
