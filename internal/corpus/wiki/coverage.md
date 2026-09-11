# Which shapes each reference deployment actually uses

The extracts describe CR shapes. This says how many deployments each shape was
read from, because an extract written from one chart is a description of that
chart, and nothing in it says so.

Rendered with `helm template` against each chart's own values, then counted by
kind. **Per deployment, not per chart**, because the question is how many
independent sources a shape was read from: one deployment ships twelve charts
that differ only by industry, and counting those as twelve would say the
opposite of what this page is for.

**Seven deployments.** The eighth reference repository, the Freyr skills, holds
runtime skills and no CRs at all - it is the only source for
`../usecase/browser-operation.md` and it contributes nothing here.

| kind | in how many | where |
|---|---|---|
| Workflow | 7 of 7 | everywhere |
| BotProvider | 7 of 7 | everywhere |
| SandboxBlueprint | 7 of 7 | everywhere |
| SkillSet / SourceSet / Syncer | 6 of 7 | all but the minimal flow agent |
| DataConnector | 5 of 7 | all but two, which reach their data through tools rather than a layer |
| Toolset | 5 of 7 | |
| Agent | 5 of 7 | absent from two, and both are Flow Agent shapes |
| SemanticLayer | 5 of 7 | |
| CompletionModel | 3 of 7 | the platform's own deployment, one customer, the demo generator |
| **Trigger** | **1 of 7** | one internal-audience project, and one instance of it |
| **Plugin** | **1 of 7** | auto-post, which has 28 |
| **KnowledgeBase / Loader / Source** | **1 of 7** | auto-post only |
| **Indexer** | **0 of 7** | a live CRD in no reference deployment. `../usecase/knowledge-drive.md` names it because the contract has it; nothing here has seen one configured |

**The three at 7 of 7 are the entry point**, which is the one thing every
deployment has. **The three at 1 of 7 are the thin samples**, and the sections
below are about them. **`Indexer` at 0 of 7 is thinner still**: the material
names it because the CRD does, and nobody here has seen one in a chart - read
anything this material says about it as read off the schema.

Every other kind in the rendered charts is named somewhere in this material,
which is the check in the other direction.

## What that means for the extracts

**`../usecase/trigger.md` is written from one Trigger in one chart.** Every
rule in it about the cursor, the cold start and what a scheduled run may not do
is generalised from a single instance. It is the most confidently written
extract with the thinnest sample, and the two facts are worth holding together.

**The knowledge-base shapes come from one deployment too.** `KnowledgeBase`,
`Loader` and `Source` appear only in auto-post, so `../usecase/knowledge-drive.md`
describes auto-post's arrangement of them. That it is the platform's own
deployment rather than a customer's cuts both ways - it is written by the people
who built the CRs, and it is not a customer's constraints.

**`Plugin` at 28 in one chart is the opposite problem.** One deployment uses the
shape heavily and no other uses it at all, so there is no second arrangement to
compare against and no way to tell which of auto-post's choices are the shape
and which are auto-post.

**Agent is absent from three of eight**, and those three are the Flow Agent
projects. An engagement that reaches for an Agent CR because the material talks
about Agents is choosing one of two shapes without being told there are two -
see `../usecase/agent-hub.md` against `flow-agent`.

## What this does not say

It counts kinds, not uses. Two charts declaring a `Toolset` may be using it in
ways that share nothing, and this table would show 2 either way. It is a floor
on the sample size, which is the number that was missing - not a measure of
whether the shape was understood.

**Checked:** the counts are the whole of this page and they were measured, not
estimated - re-measured 2026-09-11 by rendering every chart in the eight
reference repositories with `helm template` against its own values, 19 charts
in all, then grouping by deployment and counting `kind:`. **A coverage number
nobody can re-derive is a number that will be wrong**, so the denominator is
stated with what it excludes: seven deployments, the eighth repository holding
no CRs.

**Unchecked:** what the counts *mean*. A kind appearing in two charts is two
declarations, not two arrangements, and this page cannot tell whether they use
the shape the same way. It is a floor on the sample size behind each extract,
which is the number that was missing - not a measure of whether the shape was
understood.

## Sources

Eight charts rendered 2026-09-03 with `helm template` against each one's own
production values, then counted by `kind:`. **They are not named here** - this
page ships to every engagement, and shipped material names no customer, no
tenant and no deployment, the same rule the extracts follow. They are the
platform's own deployment, six customer or demo charts, and one industry demo
standing for the eleven more built from the same template - counting those would
inflate every row without adding a second arrangement.

Whoever maintains this tool can re-derive the list from the parent directory in
one command; nobody reading it in an engagement needs it, and the point of the
page is the sample size rather than whose sample it is.
