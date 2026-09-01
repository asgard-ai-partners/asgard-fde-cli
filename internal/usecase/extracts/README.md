# source/usecase

Extracts from Asgard deployments that are already in production, organised by the
**shape** an engagement needs rather than by CR kind. These become the answer to
"how do I write this" - the material an agent gets when it is about to author a
chart.

These are shipped to every engagement, so they name no customer, no tenant and
no deployment. Where two deployments disagree they are "an earlier one" and "a
later one", and what matters is which is newer and why.

## How to read an extract

Each file answers four questions in this order:

1. **When this shape, and when not.** The decision, with the alternative and why
   it loses. This is the part worth reading before writing anything.
2. **The shape.** Which CRs, how they reference each other, in dependency order.
3. **The fields that are not obvious.** Not a restatement of the CRD - only what
   a reader cannot infer, and what a wrong value does.
4. **What it cost someone.** Dated incidents, because a rule with a date behind
   it survives contact with someone who thinks they know better.

## Two warnings that apply to every extract

**Deployments disagree, and the disagreements are generational.** The platform
keeps changing what it derives for you, and a chart that has not been touched
since is still carrying the old workaround. Where an extract notes a conflict it
says which side is newer, and why. Never resolve one by picking the version you
happened to see first.

**The top-level layout is not settled either.** Different deployments group by
`projects/<name>/`, by `tenants/<name>/`, by domain, or use a single chart with
no grouping at all. The CR shapes below are independent of that choice.

## The extracts

**Read first:**

| file | what |
|---|---|
| [`conventions.md`](conventions.md) | where every CR goes, what it is called, what all of them need |


**Entry points** - pick by audience, never by preference:

| file | when |
|---|---|
| [`agent-hub.md`](agent-hub.md) | every caller can authenticate; several specialists |
| [`flow-agent-single.md`](flow-agent-single.md) | anonymous audience, one job |
| [`flow-agent-supervisor.md`](flow-agent-supervisor.md) | anonymous or credentialed audience, several specialists |

**Read paths** - also decided by audience:

| file | when |
|---|---|
| [`semantic-layer.md`](semantic-layer.md) | internal audience, open-ended questions |
| [`fixed-query-tools.md`](fixed-query-tools.md) | public audience, a known set of questions |
| [`knowledge-drive.md`](knowledge-drive.md) | knowledge that is documents, not rows |
| [`external-api.md`](external-api.md) | a source that is an HTTP API, not a database |
| [`browser-operation.md`](browser-operation.md) | a system with no database and no API - only a web UI. The last resort |

**Write paths** - anything with a side effect:

| file | when |
|---|---|
| [`write-path.md`](write-path.md) | the approval gate. Read this before designing any action that changes something |

**Capabilities and scheduling:**

| file | when |
|---|---|
| [`skill-set.md`](skill-set.md) | getting skills to a deployed agent |
| [`plugin.md`](plugin.md) | capability bundles chosen per request |
| [`trigger.md`](trigger.md) | work on a schedule with nobody watching |
