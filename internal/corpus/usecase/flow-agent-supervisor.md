# Supervisor with subagents

A public entry point that delegates to several specialist agents.

**Seen in:** three deployments - a commerce back-office with five
specialists, a manufacturing one with nine, and a finance one with three.

**Checked:** 2026-09-02 against three supervisor deployments and the CRD. The agents field was documented as a YAML list and is a stringified JSON array; corrected. Extended 2026-09-04: the conversation loop's graph read off every reference deployment at its prod values - four processors and five relationships, edge for edge identical in three of them and one edge different in a fourth, with the counts and the variant recorded below.

**Unchecked:** how to split responsibilities between subagents, and the routing prose. Judgement, taken from deployments that have not been re-examined.

**Read the platform side first:** `../wiki/agents.md` -
what a Managed Agent and a Flow Agent each are, and which the audience decides. This page assumes you have.

## When this shape, and when not

Use it when **one audience needs several specialists** and the caller should not
have to know which one to ask. The supervisor is an orchestrator: it reads the
question, picks a subagent, and delegates.

Do **not** use it for a single-purpose agent. A chat with one specialist has no
delegation decision to make, so the extra hop only adds a paraphrase step - the
orchestrator restating the question to the subagent, and restating the answer
back. One deployment removed exactly that hop from its public widget; the prompt
moved onto the Workflow and the capabilities onto the blueprint.

The other alternative is the **agent hub**: no BotProvider at all, one `Agent` CR
per system, the caller passing `agent_hub.agent_names[]` per turn. That is the
right shape when every caller can authenticate to the platform. It is not
available to an anonymous caller, and `BotProvider.entrypoint` takes a
`Workflow`, never an `Agent`.

## The shape

    BotProvider  bp-<name>          public entry, authMode none or api-key
      -> Workflow  wf-<name>        the conversation loop
        -> SandboxBlueprint sbp-<name>
             skillSetNames          capabilities of the supervisor itself
             agents[]               -> Agent CRs, the subagents
             hooks                  optional, see below

### What "the conversation loop" is

    asgard-cli add flowagent <name> --project <p> --supervisor

writes it, with the prompt left TODO and the subagents left to be added to the
blueprint. What follows is what it writes and why each edge is where it is.


It is four processors and five relationships, and **three deployments have it
edge for edge identical** - a finance supervisor, a manufacturing one and a
commerce back-office one, read at their prod values on 2026-09-04:

    entry  ->  update-context

    update-context                  --success-->  stream-llm-completion-message
    stream-llm-completion-message   --success-->  listen-message
    stream-llm-completion-message   --failure-->  push-message
    listen-message                  --success-->  stream-llm-completion-message
    push-message                    --success-->  listen-message

**It is a loop and it has no exit.** `exits: []`, and that is not an omission -
14 of the 17 Workflows across every reference deployment declare none. A run ends
when its terminal processor finishes; only a Trigger-driven Workflow, which has
somewhere to report to, tends to declare one.

Read the loop as: prime the context once, answer, then wait for the next turn.
`listen-message` is what makes it a conversation rather than a request - it
returns to the completion processor, and the completion processor returns to it.

**The failure branch says something and stays in the loop.** `failure` goes to
`push-message`, which goes back to `listen-message`, so a turn that failed does
not end the conversation. A branch that failed and one that answered must not
look the same to the caller, which is why it is a separate processor rather than
the same one.

**One deployment differs by a single edge**: a shopping guide sends
`update-context --success--> listen-message`, waiting before it answers rather
than answering first. Both are deployed. Which one is right depends on whether
the agent opens the conversation.

**The two-processor query tool is a different shape and worth not confusing with
this one.** Two deployments have `update-context --success--> http-request`, with
the request's `success` **and** `failure` both going to `push-message`: one turn,
no waiting, and the failure path says so rather than being silent.
`../wiki/processors.md` says which relations each type emits, and a
`relationName` a type never emits is a branch never taken.

Files group as one directory per supervisor:

    supervisor/<name>/bot_provider.yaml
    supervisor/<name>/workflow.yaml
    supervisor/<name>/sandbox_blueprint.yaml
    subagent/<name>.yaml            one per specialist

## Generate it

    asgard-cli add flowagent <name> --toolset ts-<name>
    asgard-cli add agent <specialist> --layer sl-<name>   # one per specialist

That writes the structure below with the fields that fail silently already in
place. **Copying a skeleton by hand is where those get lost**, because nothing
tells you they are missing: not helm lint, not CRD validation, not a server
dry-run. The generated file marks the judgement calls TODO - those are what the
rest of this page is about.

## The skeleton

    templates/supervisor/<name>/bot_provider.yaml
    templates/supervisor/<name>/workflow.yaml
    templates/supervisor/<name>/sandbox_blueprint.yaml
    templates/subagent/<name>.yaml        one per specialist

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: BotProvider
metadata:
  name: bp-<name>
  annotations:
    asgard-ai.com/bot-provider-name: "<display name>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  botProviderClass: generic
  entrypoint:
    workflow: wf-<name>
    entry: entry-main
  maxUnsupervisedSteps: 30
  disabled: false
  debugMode: on-demand
  adminApiKey:
    valueFrom:
      secretKeyRef:
        name: {{ include "<chart>.appSecretName" . }}
        key: asgard_resource_api_key
  generic:
    authMode: api-key       # or none, for an anonymous audience
    apiKey:
      valueFrom:
        secretKeyRef:
          name: {{ include "<chart>.appSecretName" . }}
          key: asgard_resource_api_key
---
apiVersion: asgard-ai.com/v1alpha1
kind: SandboxBlueprint
metadata:
  name: sbp-<name>
  annotations:
    asgard-ai.com/sandbox-blueprint-name: "<display name>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  skillSetNames:
    value: "sk-base,sk-<domain>"
  # A STRINGIFIED JSON ARRAY, not a YAML list. Every field on a
  # SandboxBlueprint is a ValueExprTemplate - an object taking value,
  # expression or template - so a bare YAML list here is rejected by the
  # apiserver. skillSetNames two lines up has the same shape for the same
  # reason; agents is the one people get wrong, because its contents look
  # like a list.
  agents:
    value: |-
      [{"baseAgentName":"ag-<specialist>"},
       {"baseAgentName":"ag-<other>"}]
```

Each subagent is an ordinary `Agent` CR - see the agent-hub extract for its
skeleton. The difference is only how it is reached.

Each entry carries **exactly one** of `baseAgentName` or `aliasName`; giving
both, or neither, is an error raised while the blueprint is evaluated rather
than at apply time.

- **`baseAgentName`** references an existing Agent CR as the base, and the other
  fields on the entry are an override delta on top of it: `description` and the
  four `prompt` sections are **appended** to the base, `skillSetNames` and
  `toolsetNames` are **added and deduped**, `sourceSetMounts` and
  `semanticLayers` **replace** the base entry with the same mountPath or name,
  and `browser` overrides outright.
- **`aliasName`** defines an ephemeral subagent with no Agent CR at all, and the
  same fields are then its whole definition rather than a delta.

Either way the resolved description and the fused prompt must be non-empty.

**A resolved agent can turn the browser on for the whole sandbox.** `browser` is
OR-aggregated: a blueprint saying `enabled: "false"` does not hold if any agent
it resolves asks for one. The same aggregation applies to the subagents'
skillSets, toolsets and semanticLayers, which land on the main orchestrator too -
so listing a capability on the blueprint as well is redundancy, not a
requirement. Both supervisor deployments list them anyway, so the supervisor's
own capability set reads without having to compute the aggregation.

## Designing the split - the part the generator leaves TODO

### How many subagents

**One per area of responsibility a person would recognise**, not one per system
and not one per tool. The user asks in their own terms; the roster should match
those terms.

Two tests before adding one:

- **Can you state, in one sentence, what goes to it and what does not?** If two
  subagents need each other's names in their descriptions to be told apart, they
  are one subagent.
- **Would a person in this business recognise it as a job?** "Inventory" and
  "exceptions" are jobs. "The API-calling one" is not.

One specialist means **no subagent at all** - put the prompt on the workflow and
the capabilities on the blueprint.

### The supervisor's own prompt

It is a router, not an expert. It needs to know what each specialist is for and
what to do when none of them fits - and **not** the domain knowledge, which
belongs to the specialists.

Keep the supervisor's own capabilities minimal. Anything it can do itself is
something it will do instead of delegating.

### Each subagent's `description`

Same rule as the agent-hub shape: it is **routing text, the only thing read when
deciding whether to delegate**. Business nouns, and an explicit boundary where
two look similar.

### When to compute the roster

A static list is right until the caller genuinely needs a different roster per
conversation - per tenant, per user's permissions, per brand. Then the
`expression` form earns its complexity. Do not start there.


### Publish the supervisor to the Agent Hub

`asgard-ai.com/agent-hub-published: "true"` on the BotProvider is what makes this
supervisor appear in the Hub's agent list, which is where an internal console
finds it. Both supervisor deployments carry it.

The opposite case is a public widget, which must **not** carry it - see
`../usecase/flow-agent-single.md`. Copying a supervisor's BotProvider into a
public one is how that mistake actually happened.

## Fields that are not obvious

### `agents` can be an expression, not a list

Several blueprint fields are `ValueExprTemplate`: they take either a static
`value:` or an `expression:` of JavaScript that the platform evaluates **per
turn**, with the BotProvider's payload available as `prevPayload`.

A static list of subagents is the simple case. Computing it lets the caller shape
the roster per conversation:

```yaml
  agents:
    expression: |-
      (() => {
        const addons = prevPayload.subagent_addons || {};
        const bases = [
          { baseAgentName: "ag-brand-manager", alias: "brand-manager" },
          ...
        ];
        return bases.map(b => {
          const addon = addons[b.alias] || {};
          const entry = { baseAgentName: b.baseAgentName };
          if (addon.prompt) entry.prompt = { persona: addon.prompt };
          if (Array.isArray(addon.skill_set_names) && addon.skill_set_names.length > 0) {
            entry.skillSetNames = addon.skill_set_names;
          }
          return entry;
        });
      })()
```

That is how a caller extends a subagent's persona or skills per conversation
without a CR change. `skillSetNames` on the blueprint itself is a plain
comma-separated string in `value`.

### The subagent is an ordinary `Agent` CR

`agentClass: managed`, `aliasName` (the delegation name, `^[a-z0-9][a-z0-9-]*$`,
conventionally the CR name without its `ag-` prefix), and `description` - which
is **routing text for the orchestrator, not a self-introduction**. It is the only
thing the orchestrator sees when deciding whether to delegate, so it carries the
business nouns.

`prompt` requires all four of `persona` / `task` / `context` / `format`, but they
may be empty strings. Subagents in one deployment use `persona` only and leave
the rest `""` with a comment saying why - the CRD requires the keys, not the
content.

### BotProvider auth is a choice, and it is the security boundary

    authMode: none      an anonymous public widget. There is no key a browser
                        could keep secret, so protection has to be on the
                        capability side: read-only chain, zero-parameter tools,
                        read-only mounts.
    authMode: api-key   a caller that can hold a credential, with the key read
                        from the release's own Secret.

`adminApiKey` is separate from visitor auth: it guards the admin API, and the
skeleton points it at `asgard_resource_api_key` - the conventional name, shared
with the other platform resource credentials by convention rather than by rule.

Also on the BotProvider: `maxUnsupervisedSteps` (30 in one deployment) caps how
far the orchestrator runs without a human, and `debugMode: on-demand`.

## What it cost someone

### Sandbox hooks: `user-prompt-submit`, never `session-start`

A deployment that writes runtime config into the sandbox with a hook records
why the obvious event is wrong (2026-08-21):

> session-start hook 進 Sandbox CR spec,內容一變 generation +1 -> pod 對話中被
> 重建。user-prompt-submit 由 driver 每 turn 用當輪 payload 重新評估、走 task
> 交付、完全不進 CR spec.

A JWT carries `jti`/`iat`, so the string changes on every issue. Putting it in a
`session-start` hook meant the Sandbox spec changed every time, the generation
bumped, and **the pod was rebuilt mid-conversation**. `user-prompt-submit` is
evaluated per turn by the driver and never enters the CR spec, which also fixed
tokens going stale after 8 hours.

Two more details in that hook worth copying: `umask 077` so a file holding a
token is 600, and writing to a temp file then `mv` for an atomic replace - a
mid-run message otherwise lets a running tool read half a config file.

### One label you will see in older charts

`asgard-ai.com/bot-provider-type` is derived by the platform from
`spec.botProviderClass` (workflow-service #336), so new charts do not stamp it.
Older ones do, with a comment saying the front end breaks without it - that was
true before #336. Stamping it anyway is harmless; **omitting it against a cluster
older than #336 is not**, so check the cluster you deploy to before removing it
from a chart that has it.

## Verify

```bash
asgard-cli gate               # every local check, the lint step included
asgard-cli verify <project>   # or one step alone, while iterating
```

**Never run `helm lint` by hand**: without the reserved `asgard` values file
that `gate` supplies, every chart that labels anything fails. `asgard-cli gate
--help` says why.

The xref check follows the whole chain including `agents[].baseAgentName` parsed
out of the JSON the CRD stores it in - a typo there silently drops a subagent,
and the symptom is an agent that "can't call any tools".

**Two things no check catches**, both needing a real conversation:

- an `agents` or `hooks` **expression** that evaluates to the wrong thing for a
  given payload. The syntax is checked, the logic is not
- a wrong `configs[].name` on a processor, since that field is a free-form string

For a hook, confirm what it produced rather than that it ran:

```bash
kubectl get -n <namespace> sandboxes.asgard-ai.com
# then, in a conversation, have the agent read the file the hook writes
```
