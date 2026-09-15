---
group: While building
description: a tool's description and a runtime skill reach the model in one context - which of the two owns a fact when both could hold it
---
# Which of the tool description and the skill owns a fact

A tool's `tooling.description` and a runtime skill's `SKILL.md` are **not two
audiences**. They are two texts handed to the same reader at the same moment:
the tools of a Toolset arrive in the agent's sandbox as `mcp__<toolset>__<tool>`
and the skills its SkillSet carries are loaded into the same context. So a
subject covered by both is covered twice, to one reader, with nothing saying
which half is current.

**Checked:** 2026-09-15 against `../wiki/tools.md` - how a Toolset's tools reach the agent, and what a Skillset holds - and against `../wiki/workflow.md`, where `tooling.description` is the only text the model reads when choosing between tools.

**Unchecked:** the division below is a rule rather than a platform behaviour, and nothing enforces it. No check has the two halves in front of it: by the time a skill is bound it is a file in another repository, and `asgard-cli verify` reads CRs.

## Prefer the skill

**Anything true across more than one tool belongs in the skill. The tool
description carries only what decides a choice between tools:**

    the description   what this one returns, and which sibling it is confused
                      with
    the skill         what a failure means, whose data this is, what an empty
                      result means, how two systems' entities correspond, what
                      to do when nothing is found

The two have different grain. A description is **per tool** and a skill is **per
subject**, so a fact true of five tools written into descriptions is written
five times and corrected in five places - and the copy that is missed is the one
a model reads on the turn that matters.

They are also not equally cheap to correct. A description is a field inside a
`Workflow` CR, so changing it is a chart edit and a release; a skill is a file a
Syncer brings in. What most often needs correcting after a customer conversation
is the interpretation, and that is the half that should be cheapest to change.

**This does not thin the description out to a label.** `../usecase/external-api.md`
is where it is doing the most work - the model can answer a question about an
external system from memory and be wrong, so the description is what tells it to
call the tool at all. That instruction is about *this* tool, so it stays.

## The stale half tells the model to do something impossible

Not the usual duplication problem, where two copies drift and a person cannot
tell which is current.

An engagement built a set of read-only HTTP tools over one API, plus a skill for
the same subject, each following its own page. The descriptions came out right.
The skill went on describing the HTTP layer the tools now hide: the base URL,
the header, the query parameter, that one field arrives as stringified JSON
needing a second parse, that one status code means two things. The model is
handed a function and a result object. It never sets a header, never picks a
parameter - and the skill instructs it to parse a field that does not reach it.

The cost is wasted context at best. At worst the model tries to satisfy the
instruction and reports failure on a call that succeeded, which is the shape an
agent fills with a confident invented answer.

**So read the pair, because nothing else will.** When a tool is added to a
subject a skill already covers, the skill is part of that change: whatever the
tool now hides comes out of it.

## Where each one is written

`../usecase/fixed-query-tools.md` and `../usecase/external-api.md` are how to
write a `tooling.description`; `../usecase/skill-set.md` is what belongs in a
runtime skill and how its own description decides whether the agent loads it at
all. **Those two descriptions are not the same text and do not follow the same
rule** - a skill's description is a trigger condition for loading it, and it is
read before the skill's body is.
