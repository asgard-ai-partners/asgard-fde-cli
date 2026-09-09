# Agent hub

Several specialist agents reachable through the platform's own entry point. **You
author no BotProvider.**

**Seen in:** a deployment with five agents over six semantic layers, one agent
per source system.

**Checked:** 2026-09-02 against a hub deployment's 5 Agent CRs (no BotProvider in that project, prompt.task byte-identical across all five) and the CRD.

**Unchecked:** the delegation-design guidance - how many agents, where the line between two of them goes. No deployment contradicts it; none confirms it either.

**Read the platform side first:** `../wiki/agents.md` -
what a Managed Agent and a Flow Agent each are, and which the audience decides. This page assumes you have.

## When this shape, and when not

Use it when **every caller can authenticate to the platform** - an internal
console, another service, anything that holds a credential. It costs nothing to
run and saves a BotProvider per agent.

It is **not available to an anonymous caller**. An `Agent` CR is reachable only
through `preset-agent-hub`, which requires an authenticated caller, and
`BotProvider.entrypoint` accepts a `Workflow`, never an `Agent`. A public widget
therefore has to use a self-hosted entry point instead.

## The shape

    preset-agent-hub          provisioned by the platform in every namespace
      -> Agent  ag-<system>   pure subagent config; it spawns nothing
           semanticLayers[]   the read surface
           toolsetNames[]     any write path
           skillSetNames[]    domain skills

The caller selects agents per turn with `payload.agent_hub.agent_names[]`,
filled with the Agent CRs' `metadata.name`.

## Generate it

    asgard-cli add agent <name> --layer sl-<name>

That writes the structure below with the fields that fail silently already in
place - the display annotation, the labels the UI needs, the current field names.
**Copying the skeleton by hand is where those get lost**, because nothing tells
you they are missing: not helm lint, not CRD validation, not a server dry-run.

The generated file marks the judgement calls TODO. Those are what the rest of
this page is about.

## The skeleton

`projects/<project>/chart/app/templates/agent/ag-<system>.yaml`, one file per CR.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: Agent
metadata:
  name: ag-<system>
  annotations:
    # Required. Without it the CR applies cleanly and then shows up nameless in
    # the UI, which no lint or dry-run reveals.
    asgard-ai.com/agent-name: "<display name>"
  labels:
    # The on/off switch. Callers build agent_names[] from the published agents.
    asgard-ai.com/agent-published: "true"
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  agentClass: managed
  managed:
    aliasName: <system>
    description: |-
      <when to delegate here, in the customer's business nouns>
    prompt:
      persona: |-
        <who this agent is - per agent>
      task: |-
        <byte-identical across every agent in this chart>
      context: |-
        <what it can reach - per agent>
      format: |-
        <byte-identical across every agent in this chart>
    sampleQuestions:
      - <at least two, required while published>
      - <...>
    skillSetNames:
      - sk-base
    semanticLayers:
      - name: sl-<system>
        allowQuery: true
        allowWrite: false
```

`toolsetNames` replaces or joins `semanticLayers` when the capability is a tool
rather than a read surface.

## One system, one agent

Each Agent mounts **exactly one** semantic layer, so its search space is that one
system's cubes rather than all of them combined. An agent mounting two has the
search space the split was meant to shrink.

**No `allowedCubes`** on the binding: an agent may query any table in its own
layer, and the restriction is which layer it mounts rather than which cubes
within it. Setting it is a deliberate departure that needs a reason - and the
reason a *public* audience needs one is why that audience gets fixed query tools
instead of a layer at all. See `../usecase/semantic-layer.md`.

Zero layers is legal when the agent's capability comes from toolsets instead.
Zero of both is not - an agent with no capability source at all is almost always
a layer deleted by accident, and the gate rejects it.

## Fields that are not obvious

**`description` is routing text, not a self-introduction.** It is the only thing
the orchestrator sees when deciding whether to delegate, so it carries the
business nouns. Where two agents look similar, each one's description says which
side of the line it is on.

The boundary that gets misrouted is the one where two systems hold overlapping
data. "How much of this item is left" and "which bin is it in, and when does it
expire" sound like the same question and belong to different systems. Say so in
both descriptions.

**Publishing is the on/off switch.** Callers build `agent_names[]` from the
*published* agents, so `asgard-ai.com/agent-published` decides whether an agent
gets delegated work - nothing else does. Leaving it `"true"` and omitting
`sampleQuestions` does **not** hold an agent back; the gate rejects that
combination, and a published agent needs at least two sample questions.

**`aliasName`** is the delegation name, `^[a-z0-9][a-z0-9-]*$`, conventionally the
CR name without its `ag-` prefix.

**No `completionModelName` on an Agent.** The orchestrator's model is chosen per
turn by the caller (`agent_hub.completion_model_name`), not baked into the CR.
`SemanticLayer.spec.completionModelName` is a different, still-required field.

**`prompt` requires all four of `persona` / `task` / `context` / `format`**, and
`aliasName` + `description` are required too.

## Designing the parts the generator leaves TODO

### How many agents

**One per source system.** Not one per question, not one per department. The
split exists so an agent's search space is one system's tables rather than all of
them combined, and the orchestrator picks between them per turn.

A system nobody asks questions about does not need an agent. A question that
spans two systems does not need a third agent - the orchestrator selects both and
reconciles the answers.

### `description` - the only thing the orchestrator reads

Write it as an instruction to a router, not as a biography:

    當使用者的請求涉及<業務名詞>時,委派給<這個 agent>。
    它可以<能做什麼>。舉凡「<像這樣的問題>」的請求都應委派給它。

Put the customer's own nouns in it - the words they use for their orders, their
parts, their tickets. The orchestrator matches on those.

**Where two agents look similar, each description says which side of the line it
is on.** "How much stock is left" and "which shelf is it on, and when does it
expire" sound like one question and belong to different systems. If you cannot
state the boundary in one sentence per side, the split is wrong.

### The prompt, in four sections that do not overlap

| section | holds | per agent? |
|---|---|---|
| `persona` | who it is, and that it works **through tools** rather than from memory | yes |
| `task` | the operating rules: when to reach for a tool, what to do when it cannot | **byte-identical across the chart** |
| `context` | what this agent can reach, described by capability | yes |
| `format` | how to answer, and what never to say | **byte-identical across the chart** |

**Keep it high level. Do not name CRs, tools, skills or columns in a prompt.**
Say "query through the semantic model" and "act through the system's tools", so
that adding a cube or renaming a tool does not mean editing prose. A capability
the agent should have is bound with a SkillSet, not described in a prompt.

The two shared sections are shared because there is no include mechanism: they
are duplicated on purpose, so a change is one global replace and the gate can
diff them.

### `sampleQuestions` - two per agent, and they are demo openers

Each one has to trigger the multi-step reasoning on its own, without the agent
having to ask a follow-up for an id nobody knows:

- **Anchor a real subject.** Use something that exists in the data - a customer
  name, a part number - rather than "this order". "這張單" with no antecedent
  forces the agent to ask "which one?" and the demo stalls.
- **Match how people refer to things.** Business subjects in plain language
  (customer plus item); technical identifiers as the codes people actually say
  on the floor (a part number, a spec).
- **Point at a decision, not a single number.** "Which of these is at risk"
  pulls several steps; "what is the stock of X" pulls one.
- **Keep each one self-contained**, and inside what this agent can actually
  reach.

## The constraint that surprises people

**`prompt.task` and `prompt.format` must be byte-identical across every Agent in
one chart.** Agent CRs have no include mechanism, so shared prompt text can only
be duplicated; keeping it verbatim is what makes a change a single global replace
and lets the gate diff-verify it. `persona` and `context` are per-agent.

Editing one agent's `task` and not the others fails the gate. That is the point.

## Verify

```bash
asgard-cli gate               # every local check, the lint step included
asgard-cli verify <project>   # or one step alone, while iterating
```

**Never run `helm lint` by hand**: without the reserved `asgard` values file
that `gate` supplies, every chart that labels anything fails. `asgard-cli gate
--help` says why.

The last one is what catches this shape's specific mistakes: two layers on one
agent, a layer mounted twice, a published agent with fewer than two sample
questions, and `task`/`format` that have drifted apart.

## Adding a system

Adding a semantic layer means adding an Agent for it, with its routing
`description` and prompt `context`. **The CR binding alone tells neither the
orchestrator nor the agent that the capability exists** - nothing infers it.
