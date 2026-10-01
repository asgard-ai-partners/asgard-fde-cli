---
description: the agent hub or your own BotProvider, why a public widget cannot have the hub, the project that needs no entry point at all
---
# Decide each project's entry point

This stage applies to a project with a read path and no way to reach it. It is
the second decision that is often answered wrong.

Needing an entry point:
<<range .Projects>><<if .NeedsEntryPoint>>  - <<.Slug>>
<<end>><<end>>
## First: does this project need one at all?

A chart can be finished at the read path. If what the customer wants is to
watch the same numbers each morning, the consumer is Data Insight (Mimir): they
explore the model by conversation and save Views and Dashboards in the product,
and no Agent, Toolset, entry point or BotProvider is written at all.

Nothing on disk can tell that apart from an agent nobody has written yet, so
write it in a comment in the chart, next to the layer. There is no field for it,
because no tool can check the claim.

Write that comment only when the answer to "who asks it a question" is nobody;
it is not a way to skip this stage. Without the comment, a later reader who finds a
SemanticLayer that nothing references will read it as a missed connection and
bind an Agent to it, which hands agents deliberately restricted to an API a
second path into the database. No gate catches that: cross-reference
checking validates references that exist, never one that should not.

    ../guide/read-path.md
    ../usecase/mimir-dashboard.md

## The decision

| audience | answer |
|---|---|
| a caller that can authenticate to the platform | the per-namespace preset-agent-hub. You author no BotProvider - one Agent CR per system is the whole entry surface |
| an anonymous visitor | your own BotProvider -> Workflow -> SandboxBlueprint |

## Why the agent hub cannot serve a public widget

The constraint is a field on the CRD and it is not this page's to state:
`../wiki/agents.md` has what `BotProvider.entrypoint` takes, and
`../usecase/agent-hub.md` has what that means for the shape. It cannot be
configured around: a public audience forces the second row of the table above.

> Answered wrong once: a public widget was specced onto the agent hub, then
> re-decided as a self-hosted chain after the constraint surfaced. The
> constraint was documented; ask the audience question before committing to an
> entry point.

Read the shape before writing it:

    ../usecase/agent-hub.md
    ../usecase/flow-agent-single.md
    ../usecase/flow-agent-supervisor.md

If the audience already lives on LINE, Telegram, Discord or Slack, that decides
it: those platforms reach a BotProvider and nothing else, so the agent hub is
out. One field on the CR, and it is immutable after creation:

    ../usecase/chat-channel.md

## If it is the agent hub

The Agent CR is pure subagent config and spawns nothing. Its `description` is
routing text, not a self-introduction. The orchestrator reads it, with the
sample questions appended, as the "when to spawn" line above the agent's full
prompt, so put the business nouns in it, and
where two agents look similar, say in each which side of the line it is on.

Publishing is the on/off switch: callers build agent_names[] from the *published*
agents, so the agent-published label decides whether an agent gets work.

A published agent needs at least two sampleQuestions. That is our rule, not the
platform's: the CRD's `sampleQuestions` has no minimum, so
nothing on the cluster refuses one without them and `asgard-cli verify` does
(R7). Knowing which of the two refuses it tells you whether the fix is a chart
edit or a conversation. `../usecase/agent-hub.md` has why two.

prompt.task and prompt.format are the shared block: Agent CRs have no include
mechanism, so shared text can only be duplicated, and keeping the copies verbatim
is what makes a change a single global replace. Nothing checks it, because a
chart that puts per-role substance inside those two fields is a real shape. `../usecase/agent-hub.md` says what
that leaves to you.

## If it is a self-hosted chain

    visitor -> BotProvider (public) -> Workflow -> SandboxBlueprint -> capabilities

The capabilities live on the SandboxBlueprint, one CR further out than you
would expect. Searching the Workflow CRD for the link finds nothing; do not
conclude from that that the link does not exist.

A single public agent needs no subagent. Attaching one only adds a hop where
the orchestrator restates the question and restates the answer back. Put the
prompt on the Workflow's processor and the capabilities on the blueprint.

The endpoint takes the BotProvider's own key (`generic.authMode: api-key`) or
nothing at all (`none`, the anonymous shape), and a key in front-end JavaScript
is not a secret either way. There is no per-visitor identity. The protection is on the capability side: the whole chain read-only,
zero-parameter tools so no user input reaches SQL, mounts readOnly. That
protection is lost as soon as someone adds a parameterised or write-capable tool.

## Either way: the UI metadata

Every Workflow needs its full workflow-set label set, or it belongs to no set and
the UI has nothing to list, even though it runs correctly. A hand-written
Trigger needs two labels of its own or its editor opens as a blank canvas. See
AGENTS.md, "Workflow sets".

Done when: asgard-cli verify resolves the whole chain including the entry names.

**Checked:** against asgard-kube `cbd8d70`. `BotProvider.spec.entrypoint`
is `{entry, workflow}`, both required, with no agent field; the Agent CRD's own
description reads "The Agent CR is a pure 'subagent config' resource ... it no
longer produces its own deployment"; `agentClass` carries `self == oldSelf`; the
capability fields (`agents`, `skillSetNames`, `pluginNames`, `sourceSetMounts`,
`credentialMounts`, `hooks`) are on `SandboxBlueprint.spec` and not on
`Workflow`. Two claims that read as platform rules are ours and now say so:
the two-sampleQuestions minimum (the CRD sets none - `gate` R7 does). Shared
prompt text across per-role agents is nobody's rule: a chart set may interleave
per-role content into `task` and `format`, and `gate` does not check it.

Against asgard-core `478cf5d6`,
asgard-core `internal/edgeserver/middleware/bot_provider.go`: a generic BotProvider's routes
check its own key in `X-API-KEY` or `X-Asgard-Webhook-Token`, or nothing under
`authMode: none`, so the protection of a public endpoint can only be on the
capability side.

**Unchecked:** which audience forces which shape, and that a single public agent
needs no subagent, are one engagement's judgement rather than a platform
contract.
