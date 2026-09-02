The onboarding is done: every project reads through a shape, is reachable, and
has been deployed. **From here the repo is not "finished", it is live** - and
the work changes character. New capability arrives as a request, not as a stage.

**The commonest later request is not an agent.** A repo that is live has a
SemanticLayer somebody trusts, and the next thing asked for is usually to *see*
it - the same numbers every morning, without asking. That is Mimir, it needs no
new CR beyond the model that already exists, and it is the cheapest thing this
walk ever delivers. Check before designing an agent for it:
`asgard-cli usecase mimir-dashboard`, and question 2b of
`asgard-cli next --stage requirements`.

This is the loop for adding one.

## 0. Open it as a request

    asgard-cli request add "<what they asked for, in their words>"

Everything below is that request's own sections. Opening it first is what gives
the work an ID, a date and a status, so `asgard-cli next` walks it and the next
agent to open the repo can see it without being told.

## 1. Which project does it belong to?

Ask the same question that decided the original split: **who is on the other
end?**

  - Same audience as an existing project -> it goes in that project.
  - A new audience -> it is a **new project**: `asgard-cli project add <slug>`,
    then `asgard-cli scaffold`, and it walks its own way through stages 3-8.

Do not put a public capability into an internal project because the data happens
to be nearby. The entry point and the read path follow the audience, and mixing
them is how a semantic layer ends up reachable from a public endpoint.

## 2. Does it need a spec first?

**Yes, if it does any of these:**

  - adds a `SemanticLayer` or `DataConnector`
  - widens which cubes an agent may query
  - introduces a write path (`asgard-cli usecase write-path`)

The databases are real, so those changes need to be reviewed before they are
made, not after. Load the `spec-workflow` skill, then:

    asgard-cli task add "<title>" --request <<.RequestID>> --project <project>
    asgard-cli task ready <id>       once every R# maps to a task and a check
    asgard-cli task start <id>       when the user asks for implementation

Those write the spec, register it, and move the status in the index, the spec's
Meta and the spec's log together. Doing it by hand is three edits, and a repo
where two of them disagree tells the next reader nothing.

**No, if it is a small, clearly scoped change** - a new sampleQuery, a prompt
correction, an annotation that was missing.

Task IDs are global across projects, and `task add` takes the next free one from
the files on disk. Two branches numbering independently is how a collision
happens.

## 3. Read before writing

  - `docs/spec/<<.SpecSlug>>/<module>.md` for the area you are touching. Your
    change is a **delta against it**, so you have to know what it currently says.
  - `docs/decisions/` when something looks odd. It is usually deliberate, and the
    record says why.
  - `AGENTS.md` for how to write the CR itself.

## 4. Implement, reusing what is there

Prefer an existing `DataConnector`, `SkillSet`, `SourceSet` over a new one. A
second CR that does the same job as an existing one is the thing reviewers catch
late and it is expensive to unpick.

If it is a new system rather than a new question about an old one, that is
stage 3 again: introspect the real database, do not guess the schema.

## 5. Gate, then deploy

Run the full acceptance gate (`asgard-cli next --stage verify`). Then tag.

## 6. Close the loop - this is the step that gets skipped

A task that changed behaviour **is not done** until:

  - the delta is applied to `docs/spec/<<.SpecSlug>>/<module>.md`, so the living
    spec describes the system as it is now;
  - each decision that got settled has its own
    `docs/decisions/YYYY-MM-DD-<topic>.md`, written the day it was settled;
  - the module index and the task index both reflect it.

A task spec stops being read once it reaches `done`. The living spec is what the
next person reads, and if the delta never reaches it, the next engagement starts
from a description of a system that no longer exists.
