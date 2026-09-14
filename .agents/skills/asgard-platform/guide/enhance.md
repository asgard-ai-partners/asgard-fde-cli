# Add a capability to a repo that is already live

The onboarding is done: every project reads through a shape, is reachable, and
has been deployed. **From here the repo is not "finished", it is live** - and
the work changes character. New capability arrives as a request, not as a stage.

**The commonest later request is not an agent.** A repo that is live has a
SemanticLayer somebody trusts, and the next thing asked for is usually to *see*
it - the same numbers every morning, without asking. That is Mimir, it needs no
new CR beyond the model that already exists, and it is the cheapest thing this
walk ever delivers. Check before designing an agent for it:
`../usecase/mimir-dashboard.md`, and question 2b of
`../guide/requirements.md`.

This is the loop for adding one.

## 0. Open it as a request

    asgard-cli request add "<what they asked for, in their words>"

Everything below is that request's own sections. Opening it first is what gives
the work an ID, a date and a status, so `asgard-cli request` reports it and the next
agent to open the repo can see it without being told.

## 1. Which project does it belong to?

Ask the same question that decided the original split: **who is on the other
end?**

  - Same audience as an existing project -> it goes in that project.
  - A new audience -> it is a **new project**: `asgard-cli project add <slug>`,
    which writes its chart skeleton too, and it finds its own way from there.

Do not put a public capability into an internal project because the data happens
to be nearby. The entry point and the read path follow the audience, and mixing
them is how a semantic layer ends up reachable from a public endpoint.

## 2. Does it need a spec first?

**Yes, if it does any of these:**

  - adds a `SemanticLayer` or `DataConnector`
  - widens which cubes an agent may query
  - introduces a write path (`../usecase/write-path.md`)

The databases are real, so those changes need to be reviewed before they are
made, not after. Load the `spec-workflow` skill, then:

    asgard-cli task add "<title>" --request <request-id> --project <project>
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

  - `docs/spec/<spec-slug>/<module>.md` for the area you are touching. Your
    change is a **delta against it**, so you have to know what it currently says.
  - `docs/decisions/` when something looks odd. It is usually deliberate, and the
    record says why.
  - `AGENTS.md` for how to write the CR itself.

## 4. Implement, reusing what is there

Prefer an existing `DataConnector`, `SkillSet`, `SourceSet` over a new one. A
second CR that does the same job as an existing one is the thing reviewers catch
late and it is expensive to unpick.

If it is a new system rather than a new question about an old one, wiring it up
is the same work as the first time: introspect the real database, do not guess
the schema.

    ../guide/data-sources.md

## 5. Gate, then deploy

Run the full acceptance gate (`../guide/verify.md`). Then tag.

## 6. Close the loop - this is the step that gets skipped

A task that changed behaviour **is not done** until:

  - the delta is applied to `docs/spec/<spec-slug>/<module>.md`, so the living
    spec describes the system as it is now;
  - each decision that got settled has its own
    `docs/decisions/YYYY-MM-DD-<topic>.md`, written the day it was settled;
  - the module index and the task index both reflect it.

A task spec stops being read once it reaches `done`. The living spec is what the
next person reads, and if the delta never reaches it, the next engagement starts
from a description of a system that no longer exists.

**Checked:** 2026-09-04 - the CR kinds it names are in the contract at
asgard-kube `cbd8d70`, and the loop it describes is the one the commands
implement: a request, then a task spec, then the chart change, then the living
spec. `asgard-cli request` and `asgard-cli task` are what move each status, and
each writes the three places by hand editing would miss.

**Unchecked:** which change needs a spec first and which does not. The rule -
anything touching a `SemanticLayer` or `DataConnector`, widening which cubes an
agent may query, or introducing a write path - is this engagement's line, drawn
where a change points at the customer's live systems. **No source states it**,
and it is deliberately conservative: the cost of a spec nobody needed is an hour,
and the cost of the other mistake was paid once.
