# Workflow and Processor

A Workflow is the program that actually runs. In the UI it is a diagram; in a
chart it is the `Workflow` CR's `entries` / `exits` / `processors` /
`relationships`.

## Processor types

Thirteen of them. The CRD's `ProcessorType` against the documentation's grouping:

| group | processor | `type` |
|---|---|---|
| Flow | Entry / Exit / Router | `router` |
| Message | Push / Listen | `push-message` / `listen-message` |
| Model | LLM Completion | `llm-completion` |
| | Stream LLM Completion | `stream-llm-completion-message` |
| | Generate Embedding | `generate-embedding` |
| Query | SQL | `query-database` |
| | LLM Query Database | `llm-query-database` |
| | Retrieve Knowledge | `retrieve-knowledge` |
| Action | Update Context | `update-context` |
| | Execute Script | `execute-script` |
| API | HTTP Request | `http-request` |
| AutoTool | Validate Payload | `validate-payload` |

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

Expression is JavaScript, not CEL. Arrow functions, `const`, `||`, `??`,
`String()` and `encodeURIComponent` all work, as do built-in helpers such as
`history(0, -1)` and `urlEncode(...)`.

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
`asgard-cli usecase workflow-chain`.

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

`asgard-cli usecase workflow-chain` covers what actually passes between
processors; `fixed-query-tools` is the shape of a zero-parameter query tool.

## Sources

- All 16 files under
  [processor](https://docs.asgard-ai.com/docs/developer-reference/processor/introduction)
  - asgard-docs `f00e0ee`
- [Expression forms](https://docs.asgard-ai.com/docs/developer-reference/asgard-builtin/expression-introduction)
  - asgard-docs `f00e0ee`
- [Architecture](https://docs.asgard-ai.com/docs/developer-reference/architecture)
  - asgard-docs `f00e0ee`
- [Conversation context](https://docs.asgard-ai.com/docs/help-community/other/retrieve-conversation-context)
  and its near-duplicate `compare-conversation-context-retrieve-method`
  - asgard-docs `f00e0ee`
- [JSON Schema editor](https://docs.asgard-ai.com/docs/help-community/other/json-schema)
  - asgard-docs `f00e0ee`
- Processor list and the limits: checked 2026-09-02 against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` -
  `ProcessorType`, `WorkflowSpec`
- Expression being JavaScript: confirmed 2026-09-02 from both sides - the product
  documentation states it, the CRD makes no claim, and six deployments use
  JavaScript constructs throughout

**Unchecked:** the processor list and value forms were held against the CRD, and
`prevPayload` behaviour against six deployments; the message templates were not
examined.
