There is a read path but no way to reach it. **This is the second decision that
gets answered wrong.**

Needing an entry point:
<<range .Projects>><<if not (.Has "Agent" "BotProvider")>>  - <<.Slug>>
<<end>><<end>>
## The decision

| audience | answer |
|---|---|
| a caller that can authenticate to the platform | the per-namespace preset-agent-hub. You author no BotProvider - one Agent CR per system is the whole entry surface |
| an anonymous visitor | your own BotProvider -> Workflow -> SandboxBlueprint |

## Why the agent hub cannot serve a public widget

An Agent CR is reachable **only** through preset-agent-hub, and that endpoint
requires an authenticated caller. BotProvider.entrypoint accepts a Workflow,
**never an Agent**. So a public audience forces the second shape, no matter how
it is configured.

> Answered wrong once: a public widget was specced onto the agent hub, then
> re-decided as a self-hosted chain after the constraint surfaced.

Read the shape before writing it:

    asgard-cli usecase agent-hub
    asgard-cli usecase flow-agent-single
    asgard-cli usecase flow-agent-supervisor

If the audience already lives on LINE, Telegram, Discord or Slack, that decides
it: those platforms reach a BotProvider and nothing else, so the agent hub is
out. One field on the CR, and it is immutable after creation:

    asgard-cli usecase chat-channel

## If it is the agent hub

The Agent CR is pure subagent config and spawns nothing. Its `description` is
**routing text**, not a self-introduction - it is the only thing the orchestrator
sees when deciding whether to delegate, so put the business nouns in it, and
where two agents look similar, say in each which side of the line it is on.

Publishing is the on/off switch: callers build agent_names[] from the *published*
agents, so the agent-published label decides whether an agent gets work. A
published agent needs at least two sampleQuestions.

prompt.task and prompt.format must be **byte-identical** across the agents in one
chart. Agent CRs have no include mechanism, so shared text can only be
duplicated; verbatim equality is what makes a change a single global replace, and
the gate diff-checks it.

## If it is a self-hosted chain

    visitor -> BotProvider (public) -> Workflow -> SandboxBlueprint -> capabilities

The capabilities live on the **SandboxBlueprint**, one CR further out than you
would expect. Searching the Workflow CRD for the link finds nothing, and that has
produced a wrong conclusion before.

A single public agent needs **no subagent**. Attaching one only adds a hop where
the orchestrator restates the question and restates the answer back. Put the
prompt on the Workflow's processor and the capabilities on the blueprint.

The endpoint is public and unauthenticated - a key in front-end JavaScript is not
a key. **The protection is on the capability side**: the whole chain read-only,
zero-parameter tools so no user input reaches SQL, mounts readOnly. That argument
stops holding the moment someone adds a parameterised or write-capable tool.

## Either way: the UI metadata

Every Workflow needs its full workflow-set label set, or it belongs to no set and
the UI has nothing to list - while runtime is perfectly fine. A hand-written
Trigger needs two labels of its own or its editor opens as a blank canvas. See
AGENTS.md, "Workflow sets".

Done when: asgard-cli verify resolves the whole chain including the entry names.
