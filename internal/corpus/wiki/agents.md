# Flow Agent and Managed Agent

They are not two of the same thing. A Managed Agent is the configuration of one
agent; a Flow Agent is an entry point from outside, and it can mount Managed
Agents.

| | Managed Agent | Flow Agent |
|---|---|---|
| what it is | one LLM conversational agent's configuration | the entry point that receives requests and coordinates Managed Agents |
| published to | the Agent Hub (Sindri) | it is the outward endpoint itself |
| who can reach it | callers that can authenticate | can be opened to anonymous visitors |
| can mount | MCP Server, Skillset, Drive, Semantic Model, Browser | Plugin, MCP Server, Skillset, **Managed Agent** |

That last cell is the point: a Flow Agent can mount one or more Managed Agents,
so the two compose rather than compete.

## Managed Agent

To create one:

- **Agent Name** (required)
- **Agent Alias** (required) - starts with a lowercase letter, and takes only
  letters, digits and hyphens
- **Description** (required) - what this agent can do. It is shown on the agent's
  profile, and it is what the orchestrator reads to decide whether to delegate
- **Prompt** (required), in four fields
  - Persona - who this agent is
  - Task - the list of things it does
  - Context - the background it needs while doing them
  - Format - the shape its replies follow
- **Profile Picture**, **On-boarding Settings** (optional) - sample questions and
  a custom menu

Mountable resources: MCP Servers, Skillsets, Drives, Semantic Model, Browser
Configuration.

Five built-in templates can be applied directly: a simple online help desk,
Customer Support, Knowledge Base Q&A, Data Analyst, General Assistant.

Every Managed Agent is published to Sindri and can also be managed from the
Management Console. An enabled agent serves immediately; a disabled one stops
serving but keeps its configuration. **There is no publish step and nothing to
import on the Sindri side** - what makes an agent findable there is the
Description, because that is what the orchestrator routes on.
[`setup-path.md`](../wiki/setup-path.md) is the order this sits in.

## Flow Agent

Creating one needs only a Name and a Description. Everything else is optional,
under Advanced Sandbox Settings: Plugins, MCP Servers, Skillsets, Managed Agents.

## Choosing between them

The deciding fact is **whether the caller can authenticate**.

| situation | what to use |
|---|---|
| internal users, all signed in | Managed Agents alone - users pick one in the Agent Hub |
| anonymous visitors, a single job | a Flow Agent alone, mounting no Managed Agent |
| anonymous visitors, several specialisms | a Flow Agent mounting several Managed Agents |

The middle row is the one people get wrong. Mounting one Managed Agent for a
single-job widget makes the Flow Agent restate the question to that one agent and
restate the answer back - a round trip that adds no judgement. One deployment
removed exactly that layer after shipping it.

## Two hard limits

Not preferences - things that cannot be built:

1. **An anonymous caller cannot use the Agent Hub.** The Hub requires a caller
   that can authenticate.
2. **An outward entry point can only point at a Workflow, never at an Agent.**
   `BotProvider.entrypoint` takes a `Workflow`.

So "point a public website straight at a Managed Agent" is not implementable,
however reasonable it looks in the UI.

One project first designed a public website around the Agent Hub and later moved
it to a Flow Agent (TASK-007 to TASK-011). The deciding question is the caller's
ability to authenticate, not what other projects did.

## UI names against CR names

The product documentation describes objects in an interface; a chart declares
resources. They are not one to one.

| UI | chart |
|---|---|
| Managed Agent | one `Agent` CR (`agentClass: managed`) |
| Flow Agent | **three** CRs: `BotProvider` + `Workflow` + `SandboxBlueprint` |
| Agent Hub | `preset-agent-hub`, provisioned by the platform in every namespace |
| the Managed Agents a Flow Agent mounts | `SandboxBlueprint.spec.agents[]` |
| the Prompt's four fields | `Agent.spec.managed.prompt.{persona,task,context,format}` |

`agentClass` has only one value, `managed`, so the CRD's `Agent` **is** the UI's
Managed Agent - there is no other kind. The name suggests a pair of symmetric
options and there is not one.

For how each shape is assembled: `../usecase/agent-hub.md` (an internal
hub), `../usecase/flow-agent-single.md` (anonymous, one job), `../usecase/flow-agent-supervisor.md`
(anonymous, several specialists), `../usecase/browser-operation.md` (giving one a browser).

## Sources

- [Managed Agent](https://docs.asgard-ai.com/docs/product-suite/odin/features/agent-hub-managed-agent)
  - asgard-docs `f00e0ee`
- [Flow Agent](https://docs.asgard-ai.com/docs/product-suite/odin/features/agent-hub-flow-agent)
  - asgard-docs `f00e0ee`
- The two limits and the name mapping: checked 2026-09-02 against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` -
  `BotProviderSpec.Entrypoint`, `AgentClass`, `SandboxBlueprintSpec.Agents`

**Unchecked:** the name mapping and the two limits were held against the CRD and
one deployment; the UI flows the product documentation describes were not held
against any deployment.
