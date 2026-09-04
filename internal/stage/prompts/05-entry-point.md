# Decide each project's entry point

There is a read path but no way to reach it. **This is the second decision that
gets answered wrong.**

Needing an entry point:
<<range .Projects>><<if .NeedsEntryPoint>>  - <<.Slug>>
<<end>><<end>>
## First: does this project need one at all?

**A chart can be finished at the read path.** If what the customer wants is to
watch the same numbers each morning, the consumer is Data Insight (Mimir): they
explore the model by conversation and save Views and Dashboards in the product,
and no Agent, Toolset, entry point or BotProvider is written at all.

Nothing on disk can tell that apart from an agent nobody has written yet, so it
has to be said:

    asgard-cli project shape <project> mimir-dashboard

**That is a claim about what the customer gets, not a way to stop this stage
asking.** Declare it only when the answer to "who asks it a question" is nobody,
and then say so in the chart next to the layer - a later reader who finds a
SemanticLayer that nothing references will read it as a missed connection and
bind an Agent to it, which hands agents deliberately restricted to an API a
second path into the database. **No gate catches that**: cross-reference
checking validates references that exist, never one that should not.

    asgard-cli guide read-path
    asgard-cli usecase mimir-dashboard

## The decision

| audience | answer |
|---|---|
| a caller that can authenticate to the platform | the per-namespace preset-agent-hub. You author no BotProvider - one Agent CR per system is the whole entry surface |
| an anonymous visitor | your own BotProvider -> Workflow -> SandboxBlueprint |

## Why the agent hub cannot serve a public widget

The constraint is a field on the CRD and it is not this page's to state:
`asgard-cli wiki agents` has what `BotProvider.entrypoint` takes, and
`asgard-cli usecase agent-hub` has what that means for the shape. **The reason
it is a decision at all is that it cannot be configured around** - a public
audience forces the second row of the table above, and no amount of
configuration moves it.

> Answered wrong once: a public widget was specced onto the agent hub, then
> re-decided as a self-hosted chain after the constraint surfaced. **That is
> what this page is for** - the constraint was always readable, and reading it
> is not the same as being asked the question before committing to an answer.

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
agents, so the agent-published label decides whether an agent gets work.

**A published agent needs at least two sampleQuestions, and that is our rule
rather than the platform's** - the CRD's `sampleQuestions` has no minimum, so
nothing on the cluster refuses one without them and `asgard-cli verify` does
(R7). Knowing which of the two will stop you decides whether the fix is a chart
edit or a conversation. `asgard-cli usecase agent-hub` has why two.

prompt.task and prompt.format must be **byte-identical** across the agents in one
chart. Agent CRs have no include mechanism, so shared text can only be
duplicated; verbatim equality is what makes a change a single global replace, and
the gate diff-checks it. **Also ours, not the platform's.**

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

**Checked:** 2026-09-04 against asgard-kube `15ded0f`. `BotProvider.spec.entrypoint`
is `{entry, workflow}`, both required, with no agent field; the Agent CRD's own
description reads "The Agent CR is a pure 'subagent config' resource ... it no
longer produces its own deployment"; `agentClass` carries `self == oldSelf`; the
capability fields (`agents`, `skillSetNames`, `pluginNames`, `sourceSetMounts`,
`credentialMounts`, `hooks`) are on `SandboxBlueprint.spec` and not on
`Workflow`. Two claims that read as platform rules are **ours** and now say so:
the two-sampleQuestions minimum (the CRD sets none - `gate` R7 does) and the
byte-identical prompt text.

**Unchecked:** everything that makes this a decision rather than a lookup - which
audience forces which shape, that a single public agent needs no subagent, and
that the protection of an unauthenticated endpoint is on the capability side.
Those come from the engagement this was written in, where the first was answered
wrong once and reversed. **They have no source to be held against**, and a
reader should weigh them as one engagement's experience rather than as a
platform contract.
