# Decide each project's read path

Data connectors exist but nothing reads through them yet. **This is the first of
the three decisions that get answered wrong.**

Needing a read path:
<<range .Projects>><<if not (.Has "SemanticLayer" "Toolset")>>  - <<.Slug>>
<<end>><<end>>
## The decision

| audience | what they do with it | answer |
|---|---|---|
| internal, authenticated | ask, in the moment | SemanticLayer, one per source system, one Agent each |
| internal, authenticated | **watch, every day** | **SemanticLayer and nothing else** - see below |
| public, anonymous | ask, in the moment | a Toolset of zero-parameter fixed queries, no semantic layer |

The second column is the one this page used not to have, and the row it adds is
not a variant of the first - it is a different product.

## When nobody is asking

**A layer with no Agent on it is a finished deliverable, not an unfinished one.**
If what the customer wants is to see the same numbers each morning, the consumer
is **Data Insight (Mimir)**: they explore the model by conversation and save what
is useful as Views and Dashboards, in the product, and no Agent, Toolset, entry
point or BotProvider is written at all.

    ask one thing, now        an agent. Everything below applies
    watch the same numbers    a layer, and the chart stops there
    both, different people    both - two deliveries over one model

When the chart stops there, **say so in the chart, next to the layer.** There
used to be a field for it - a "shape" recorded per project - and it is gone:
this tool cannot check such a claim and has no business judging it, so the only
place the answer belongs is where the next reader will be looking anyway.

`asgard-cli usecase mimir-dashboard` is the shape, including the trap: a later
reader finds a SemanticLayer nothing references, assumes it is a missed
connection, and binds it to an Agent - which hands agents deliberately restricted
to an API a second path into the database. **Nothing in the gate catches that**,
because cross-reference checking validates references that exist, never one that
should not. Say so in the chart, next to the layer.

`asgard-cli guide requirements` asks this as question 2b, so the answer
should already be in the request. If it is not, it was not asked - and the
expensive version of this mistake is not choosing wrong, it is building the agent
and finding out at the demo that they wanted a page that was already open.

## Why a semantic layer is wrong for a public audience

A bound layer lets the agent compose arbitrary SQL over every cube in it, and
**the exposed surface grows by itself every time a cube is added**. Nobody goes
back to narrow it.

**Narrowing it is not an option that gets overlooked - it is refused.**
`allowedCubes` exists on the Agent's mount, and `asgard-cli verify` rejects any
Agent that sets it (R4), on the standing decision that a bound layer is
queryable in full. That is ours rather than the platform's: the CRD allows the
field. The reason to refuse it is that a per-agent allowlist makes the exposed
surface look bounded while the layer underneath keeps growing, so excluding the
sensitive tables is not the fix - the shape itself is the risk.

With fixed tools, what can be asked is decided by a few statements in version
control, no user input reaches SQL so the injection surface is zero, and widening
it takes a CR change and a review.

> This was answered wrong once already: a public site got a SemanticLayer first,
> and it was later removed and replaced with five zero-parameter query tools.

## "Zero-parameter" means the model supplies nothing, not that everyone sees the same rows

The two are constantly confused, and the confusion turns into telling a customer
that something is impossible when it is not.

An anonymous channel can absolutely answer "where is MY order". What it
cannot do is let the **model** choose whose case to look up. The customer's
identity is injected server-side on every turn - it arrives in the request the
front end sends, never as a tool argument the model fills in - and the query
filters on it. The tool still takes no parameters from the model, so the
injection surface is still zero, which is the whole point of the rule.

    the model         picks WHICH question         zero parameters
    the caller        supplies WHO is asking       every turn, server-side

Whatever sits in front - a website, a chat channel, a support desk - is what
knows who is talking and passes it through. If nothing in front knows, then the
per-customer half genuinely cannot be built, and that is a question for the
customer rather than a design to work around.

    asgard-cli usecase per-turn-credentials

Read that before promising anything of the form "查我的...". It is also where the
failure mode lives: a query that forgets to filter on the injected identity fails
on the anonymous path only, which is the path nobody tests.

Read the shape before writing it:

    asgard-cli usecase semantic-layer
    asgard-cli usecase fixed-query-tools

A fixed query tool is a Workflow, so read how a chain passes values between its
processors before writing one - that is where the silent failures are:

    asgard-cli usecase workflow-chain

## If the answer is a semantic layer

One system, one layer, one Agent. An Agent mounting two layers has the search
space the split was meant to shrink.

  - Every cube, dimension and measure needs a description in 繁體中文. It is what
    the agent reads to decide which column answers a question; one without a
    description is invisible to the model. **This is the place plain Chinese
    matters twice** - `.agents/skills/plain-chinese/` - because 至關重要 in a
    description is not just noise, it is noise the agent has to guess past every
    time it chooses a column.
  - A dimension's name is the column its sql selects, never a re-cased alias.
  - Common analysis views go in top-level sampleQueries, never as a cube-level
    sql: virtual cube. **Run every one against the live database before
    committing it** - a sampleQuery that errors actively misleads the agent.

## If the answer is fixed tools

Every tool takes zero parameters, and requestConsent is false because they are
read-only.

**Do not ship two tools that sit on the same FROM/JOIN and differ only in
projection.** Each one's description then has to name the other as the
alternative, and that is exactly where a model picks wrong. Merge them and make
the difference a column value instead of a tool choice.

Tool usage guidance goes in each tool's Workflow entries[].tooling.description.
Toolset.spec.instruction does not exist any more - adding it back passes
dry-run and then fails the real deploy.

Done when: every project reads through one shape or the other, and
asgard-cli check plus asgard-cli verify are green.

**Checked:** 2026-09-04 against asgard-kube `15ded0f`. `Toolset` declares no
`instruction` field, so the note about adding it back is current;
`SemanticLayer.spec` carries `cubes` and top-level `sampleQueries`; `allowedCubes`
is on the **Agent's** semanticLayers mount rather than on the layer, and the CRD
permits it - refusing it is `gate` R4 and now says so.

**Unchecked:** the decision itself. Which audience gets a layer and which gets
fixed tools, that a description in 繁體中文 is what the model matches on, and that
a sampleQuery which errors actively misleads - all of that comes from the
engagement this was written in, where the public-site fork was answered wrong
once and reversed. **Nothing here has a source to be held against**, and the
instruction to run every sampleQuery against the live database before committing
it is the one line that has to survive a reader who trusts the rest.
