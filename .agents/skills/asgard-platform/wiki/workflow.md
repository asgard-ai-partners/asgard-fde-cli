# Workflow and Processor

A Workflow is the program that actually runs. In the UI it is a diagram; in a
chart it is the `Workflow` CR's `entries` / `exits` / `processors` /
`relationships`.

## Processor types

Thirteen values in the CRD's `ProcessorType` enum, against the group each sits
under in the editor's node menu:

| group | node, as the menu labels it | `type` |
|---|---|---|
| 流程控制 | Router | `router` |
| Message | Push Message | `push-message` |
| | Listen Message | `listen-message` |
| Model | LLM Completion | `llm-completion` |
| | Stream LLM Completion Message | `stream-llm-completion-message` |
| | Generate Embedding | `generate-embedding` |
| Action | Update Context | `update-context` |
| | Execute Script | `execute-script` |
| Query | SQL | `query-database` |
| | Retrieve Knowledge | `retrieve-knowledge` |
| API | HTTP 請求 | `http-request` |
| Automation Tool | Validate Payload | `validate-payload` |
| CRD only | LLM Query Database | `llm-query-database` |

**Entry and Exit are nodes in that menu and are not processor types.** The
editor draws them under 流程控制 and the documentation gives each its own page,
but a Workflow declares them as its own `entries` and `exits` arrays - so
looking for an `entry` value in the enum finds nothing, and a processor written
with one is rejected. Two more rows of the menu are not general either:
Validate Payload and Response appear only in an Automation Tool workflow, and
**Response has no type at all** - the output is a `push-message` scoped to
`automation_tool`, which `../wiki/processors.md` sets out. `llm-query-database`
is the opposite case: a legal type the node menu will not add.

At most 100 processors and 1000 relationships.

`stream-llm-completion-message` is the workhorse of agent conversation:
semanticLayers, toolsets and sandboxBlueprint all hang off its config.

## Three ways a value is set, exactly one at a time

Every config value is a Literal, an Expression or a Template, and the CRD
enforces exactly one.

| form | syntax | for |
|---|---|---|
| Literal (`value`) | plain text | something fixed |
| Expression (`expression`) | **JavaScript** | computation and logic |
| Template (`template`) | **Handlebars** | rendering |

Expression is JavaScript, not CEL, and it is **not** restricted to ECMA5 -
that limit is `execute-script`'s Engine field and applies to a script body, not
to these. One shipped tenant chart evaluates
`prevBlobs.map(b => b.blobId).join(',')`, which is the evidence;
`../wiki/processors.md` carries the count behind it. `||`, `??`, `String()`
and `encodeURIComponent` work, as
do built-in helpers such as `history(0, -1)` and `urlEncode(...)`.

**`const` and `let` are a different question and the answer is don't.** No
chart uses either, because an Expression field holds one expression rather than
statements - a declaration has nowhere to go. If you need statements, that is
`execute-script`, and there you are back inside ECMA5.

This was once recorded as CEL, and anything written as CEL neither works nor
explains why. The current documentation and the charts agree.

## Reading conversation context

Two built-ins with different reach:

| | returns | for |
|---|---|---|
| `prevMessage` | the previous message | echo, single-turn lookups |
| `history(start, end)` | a slice of the conversation | multi-turn, anything needing context |

```javascript
history(0, -1)    // everything
history(-3, -1)   // the last three
history(-1, -1)   // just the previous one
```

An LLM chatbot answering in context wants `history`; `prevMessage` only covers a
single turn.

## What crosses between processors

How `prevPayload` and `httpResponse` actually behave - in particular that an
`http-request` **replaces** `prevPayload` - is in
`../usecase/workflow-chain.md`.

## Entry and tooling

A Workflow entry can carry a `tooling` block (name, description,
allowUploadFile). With it, a `Toolset` can expose that entry as a tool.

`tooling.description` is the only text the model reads when choosing between
tools.

An entry can also carry an `inputSchema` (a JSON Schema string) validating the
incoming payload, and `files` (per-slot type limits and context aliases).

### Editing an inputSchema

On a processor with a Schema field, the expand icon at the top right of the
editor opens a window that switches between Schema and JSON modes with live
two-way preview.

The editor does not save. Return to the processor's settings page and press save,
or the change never reaches the system.

## Platform built-ins

Two further reference sets live under `developer-reference/asgard-builtin/` and
are worth reading when needed rather than summarising here: the Expression
variable and ECMAScript function lists, and the Handlebars helpers plus the
button, carousel, chart, location, video and quick-reply message templates.

## Before writing the chart

`../usecase/workflow-chain.md` covers what actually passes between
processors; `../usecase/fixed-query-tools.md` is the shape of a zero-parameter query tool.

## What goes in a field

Every processor field takes one of three kinds of value - Literal, Expression
(JavaScript) or Template (Handlebars) - and the six variables in scope, the
seven built-in functions and the `Blob` shape are in
[`processors`](../wiki/processors.md). **The ECMA5 limit is `execute-script`'s
Engine field and does not reach an Expression**, which is the same thing this
page says above and the deployed charts settle: one of them evaluates an arrow
function. What no chart uses in an Expression is `const` or `let`, and that is
structural rather than a limit - the field holds one expression, not
statements.

## The editor's canvas is a ConfigMap

A Workflow renders as a diagram in the platform's editor, and **where each node
sits is not on the Workflow**. It is a plain Kubernetes `ConfigMap` holding one
key:

```yaml
kind: ConfigMap
metadata:
  name: cfgmap-<workflow>
data:
  node_positions: |-
    { "workflow": {"x":40,"y":40},
      "entry":     {"main": {"x":40,"y":140}},
      "processor": {"submit-claim": {"x":420,"y":140}, ... },
      "exit":      {"finish": {"x":1180,"y":140}} }
```

The Workflow binds it by annotation - `asgard-ai.com/workflow-config-name` - and
without one the graph opens as a pile at the origin and somebody drags it apart
by hand, once per environment.

**This is the second thing that decides whether a chart-authored Workflow is
usable in the UI**, and the two fail the same way and are never mentioned
together:

    the ConfigMap missing            the nodes open on top of each other
    project-environment-id missing   the editor opens as a blank canvas

One deployment carries 80 of these, one per Workflow. `ConfigMap` is not an
Asgard CR and appears in no CRD, which is why nothing else here mentions it -
and why it is easy to conclude it is somebody else's concern.

## Sources

- All 16 files under
  [processor](https://docs.asgard-ai.com/docs/developer-reference/processor/introduction)
  - asgard-docs `23409b3`, read 2026-09-14. The type table above is the CRD's
  `ProcessorType` enum at asgard-kube `cbd8d70` against that page's groups; the
  two are not the same list, which is why Entry and Exit now say what they are
- [Expression forms](https://docs.asgard-ai.com/docs/developer-reference/asgard-builtin/expression-introduction)
  - asgard-docs `23409b3`, read 2026-09-14
- [Architecture](https://docs.asgard-ai.com/docs/developer-reference/architecture)
  - asgard-docs `f00e0ee`
- [Conversation context](https://docs.asgard-ai.com/docs/help-community/other/retrieve-conversation-context)
  and its near-duplicate `compare-conversation-context-retrieve-method`
  - asgard-docs `f00e0ee`
- [JSON Schema editor](https://docs.asgard-ai.com/docs/help-community/other/json-schema)
  - asgard-docs `f00e0ee`
- Processor list and the limits: checked 2026-09-02 against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `cbd8d70` -
  `ProcessorType`, `WorkflowSpec`
- Expression being JavaScript: confirmed 2026-09-02 from both sides - the product
  documentation states it, the CRD makes no claim, and six deployments use
  JavaScript constructs throughout

**Unchecked:** the processor list and value forms were held against the CRD, and
`prevPayload` behaviour against six deployments; the message templates were not
examined.
