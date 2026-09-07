# Which shapes each reference deployment actually uses

The extracts describe CR shapes. This says how many deployments each shape was
read from, because an extract written from one chart is a description of that
chart, and nothing in it says so.

Rendered 2026-09-03 with `helm template` against each chart's own production
values, then counted by kind. Eight charts - the seven customer and demo
deployments plus auto-post, the platform's own.

| kind | in how many | where |
|---|---|---|
| Workflow | 8 of 8 | everywhere |
| SkillSet / SourceSet | 8 of 8 | everywhere |
| Syncer | 7 of 8 | all but one shopping-guide deployment |
| DataConnector | 6 of 8 | all but two, both of which reach their data through tools rather than a layer |
| BotProvider | 6 of 8 | all but one internal-audience project and two demo ones |
| SandboxBlueprint | 6 of 8 | |
| Toolset | 6 of 8 | |
| Agent | 5 of 8 | absent from three, and those three are the Flow Agent shapes |
| SemanticLayer | 5 of 8 | |
| CompletionModel | 3 of 8 | the platform's own deployment, one customer, the demo generator |
| **Trigger** | **1 of 8** | one internal-audience project, and one instance of it |
| **Plugin** | **1 of 8** | auto-post, which has 28 |
| **KnowledgeBase / Loader / Source** | **1 of 8** | auto-post only |

## What that means for the extracts

**`asgard-cli usecase trigger` is written from one Trigger in one chart.** Every
rule in it about the cursor, the cold start and what a scheduled run may not do
is generalised from a single instance. It is the most confidently written
extract with the thinnest sample, and the two facts are worth holding together.

**The knowledge-base shapes come from one deployment too.** `KnowledgeBase`,
`Loader` and `Source` appear only in auto-post, so `asgard-cli usecase knowledge-drive`
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
see `asgard-cli usecase agent-hub` against `flow-agent`.

## What this does not say

It counts kinds, not uses. Two charts declaring a `Toolset` may be using it in
ways that share nothing, and this table would show 2 either way. It is a floor
on the sample size, which is the number that was missing - not a measure of
whether the shape was understood.

**Checked:** the counts are the whole of this page and they were measured, not
estimated - eight charts rendered with `helm template` against their own
production values on 2026-09-03, then `kind:` counted. Re-running it is one
command and the method is in the Sources below, which is the point: a coverage
number nobody can re-derive is a number that will be wrong within a month, and
three in this repository already have been.

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
