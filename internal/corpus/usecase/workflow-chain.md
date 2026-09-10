# Chaining processors, and the graph that wires them

A Workflow with more than one step: what the thirteen processor types are, what
crosses between them, and how `relationships` decides the order. This is the
mechanism every other shape is built out of.

**Seen in:** a two-call mail sender, a search tool that reshapes its API's
response, a conversation loop, and a nine-branch content pipeline.

**Checked:** 2026-09-02 against workflows in six deployments: prevPayload 134 uses, httpResponse 20, and JavaScript constructs including exactly the 4 occurrences of ?? this page cites.

**Unchecked:** nothing outstanding. The replacement of prevPayload by an http-request is stated in a deployment's own comment in the same words.

**Read the platform side first:** `../wiki/workflow.md` -
the processor types and the three ways a config takes a value. This page assumes you have.

## When this shape, and when not

You are already in it. A single-processor Workflow is the exception, not the rule
- a tool that does one HTTP call still wants three processors, because the third
is what tells the agent whether the first worked.

Reach for **more** processors when:

- **a step's output is another step's input** - a token, a search result to
  reshape, an id to look up.
- **the flow branches on a value** - use a `router`.
- **failure has to be reported rather than swallowed.** This is the common one and
  the one that gets left out.

Do **not** add processors to express what a prompt should decide. A workflow node
per conversation topic puts the routing in two places: one deployment removed six
topic workflows and let the orchestrator route from the prompt instead. Keep the
conversation graph minimal - greet, listen, answer, back to listen, plus a
failure branch.

## Generate it

    asgard-cli add flowagent <name> --project <project>
    asgard-cli add querytool <name> --project <project> --connector dc-<name> --toolset ts-<name>
    asgard-cli add httptool <name> --project <project>

Each writes a working chain with its relationships already wired, including the
failure branches. Add processors to that rather than starting from an empty
`spec` - the parts that fail silently (the display annotation, the workflow-set
labels, the environment id) are already right.

## The skeleton

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: wf-search
  annotations:
    asgard-ai.com/workflow-name: "商品搜尋"
spec:
  variables:
    # The only way a secret reaches a workflow: a config has value, expression
    # or template, and no valueFrom.
    - name: apiKey
      valueFrom:
        secretKeyRef:
          name: {{ include "<chart>.appSecretName" . }}
          key: search_api_key

  entries:
    - name: search
      handlingProcessor: proc-input          # where the run starts
      tooling:
        name: search_products
        description: |-
          What the model reads to decide whether to call this. Name the tool it
          could be confused with.
        allowUploadFile: false   # required, and it has no default
      inputSchema: |
        { "type": "object",
          "properties": { "q": { "type": "string" } },
          "required": ["q"] }

  # A tool workflow ends by pushing a message, so it needs no exit. A
  # conversation loop does.
  exits: []

  processors:
    # (1) Copy the arguments into context BEFORE any http-request runs, because
    #     an http-request replaces prevPayload with its own result.
    - name: proc-input
      type: update-context
      configs:
        - name: searchQuery
          expression: 'String(prevPayload.q || "").trim()'
        # Validate here rather than trusting the model: an enum it half-remembers
        # arrives as a plausible wrong string.
        - name: sortType
          expression: |-
            (() => {
              const allowed = ["DEFAULT", "SOLD_DESC", "RATING_DESC"];
              return allowed.indexOf(prevPayload.sortType) >= 0
                ? prevPayload.sortType : "DEFAULT";
            })()

    - name: proc-search
      type: http-request
      configs:
        - name: url
          value: "https://api.example.com/search"
        - name: method
          value: POST
        - name: parseJson
          value: "true"
        # Any key that is not url/method/body/parseJson becomes a header.
        - name: Content-Type
          value: application/json
        - name: Authorization
          expression: '"Bearer " + vars.apiKey'
        - name: body
          expression: 'JSON.stringify({ q: searchQuery, sortType: sortType })'

    # (3) Reshape before it reaches the model. The raw response is not the answer.
    - name: proc-response
      type: push-message
      configs:
        - name: payload
          expression: |-
            (() => {
              const items = ((httpResponse.json && httpResponse.json.results) || [])
                .map((r) => ({ id: String(r.id), name: r.name }));
              return { ok: true, itemCount: items.length, items: items };
            })()

    - name: proc-error
      type: push-message
      configs:
        - name: payload
          expression: '({ ok: false, error: prevError })'

  relationships:
    - from: {processor: proc-input,    relationName: success}
      to:   {processor: proc-search}
    - from: {processor: proc-search,   relationName: success}
      to:   {processor: proc-response}
    # Every processor that can fail needs this, or the run ends in silence and
    # the agent sees a tool that returned nothing.
    - from: {processor: proc-search,   relationName: failure}
      to:   {processor: proc-error}
```

It goes in `projects/<project>/chart/app/templates/workflow/wf-<name>.yaml`, or
`templates/tool/` where the chart groups tool workflows.

## The thirteen processor types

| type | does |
|---|---|
| `listen-message` | wait for user input, including files |
| `push-message` | send a message, or return a tool's result |
| `llm-completion` | call a model synchronously; supports structured output |
| `stream-llm-completion-message` | stream a model's reply; the agentic one |
| `router` | branch on conditions |
| `http-request` | one HTTP request |
| `update-context` | put values into context |
| `query-database` | query through a DataConnector |
| `generate-embedding` | vectors, for RAG preprocessing |
| `retrieve-knowledge` | search a knowledge store |
| `execute-script` | custom logic |
| `validate-payload` | validate input and files |

In practice a handful carry almost everything: `query-database` and
`push-message` dominate the real charts, with `update-context` and
`http-request` appearing wherever a system is reached over HTTP.

## What crosses between processors

| name | holds | when it changes |
|---|---|---|
| `prevPayload` | the tool call's arguments, or the previous message's data | replaced by an `http-request`'s result |
| `prevMessage` | the previous user message | on each `listen-message` |
| `prevBlobs` | files the user uploaded | on upload |
| `httpResponse` | **the most recent** HTTP response | after every `http-request` |
| `prevError` | why the previous processor failed | on a `failure` branch |
| `vars.<name>` | a `variables` entry, including secrets | never |
| anything an `update-context` set | by the name you gave it | when you set it again |

**Configs are evaluated immediately before their processor runs.** That is what
makes a two-call chain work - at the second call's config time, `httpResponse` is
still the first call's - and it is also why inserting a processor between them
loses the value. It also means **a broken expression is a runtime failure on
first use, not a deploy failure**: nothing evaluates configs at apply time.

### `expression` is JavaScript

Arrow functions, `const`, `String()`, `encodeURIComponent`, `JSON.stringify`,
`??`, `.map()` - every real chart uses them, and none of them is CEL.

This was worth stating because the two sources once disagreed: the platform
documentation described `expression` as a CEL expression while every chart wrote
JavaScript, so anyone who trusted the docs wrote something that could not work and
had no way to see why. **Both now say JavaScript** - the product documentation
states it directly, and the CRD makes no claim either way. The rule is unchanged;
only the reason to distrust the docs has gone.

Write JavaScript, and the idiom the charts use is an immediately-invoked arrow
function when there is more than one statement:

```yaml
expression: |-
  (() => {
    const x = prevPayload.thing;
    return { ok: true, value: x };
  })()
```

Note the parentheses around a bare object literal - `({ ok: true })` - without
them it parses as a block.

The three config value types are `value` (static string), `expression`, and
`template` (Handlebars, `{{name}}`). One per config.

## The graph

```yaml
relationships:
  - from: {processor: a, relationName: success}
    to:   {processor: b}
  - from: {processor: b, relationName: success}
    to:   {exit: finish}
```

Relation names:

| processor | relations |
|---|---|
| most types | `success`, `failure` |
| `router` | one per condition you define in its configs, plus a built-in `else` |

A `router` is the only type with dynamic relations: each config's **name** is a
relation, and its expression decides whether that branch is taken.

```yaml
- name: route
  type: router
  configs:
    - name: isVIP
      expression: "userLevel === 'VIP'"
    - name: hasQuestion
      expression: "question != null"

relationships:
  - from: {processor: route, relationName: isVIP}       ; to: {processor: vip}
  - from: {processor: route, relationName: hasQuestion} ; to: {processor: ask}
  - from: {processor: route, relationName: else}        ; to: {processor: default}
```

### Two disciplines that are not optional

1. **Every processor that can fail needs a `failure` relationship.** Without one
   the run stops there. A tool call that stops has returned nothing, which the
   model reads as an empty answer rather than an error - and then it tells the
   user something that did not happen. This is the single most common omission.
2. **A tool's last processor reports success or failure explicitly**, as a field
   the prompt is told to read (`ok: true|false`). The status code is not visible
   to the model; only what you push is.

## Designing the chain

- **Reshape the response before the model sees it.** An API's raw JSON has
  fields the model will quote, ids in the wrong type, and nesting it will get
  wrong. One deployment measured that its search API returned numeric ids while
  the downstream tool needed strings, and cast in the workflow rather than hoping
  - **write down what you measured**, with the date, because the next person
  cannot tell a deliberate cast from a leftover.
- **Validate an enum in `update-context`, not in the prompt.** A model that
  half-remembers an allowed value produces a plausible wrong one, and the API
  answers 400 to something that looks fine in the log.
- **Name processors for what they do, not their type.** `proc-search` beats
  `http-request-1` in a nine-node graph, and the `labels.display_name` is what
  the Platform UI shows.
- **Keep one Workflow to one job.** Two jobs in one graph share a failure path,
  and then one job's error message answers the other job's caller.

## Verify

    asgard-cli verify <project>

It resolves every reference into and out of this Workflow - including both halves
of an entrypoint, because a wrong entry name is as dead as a wrong workflow name
- and requires the workflow-set labels and the display annotation, without which
the Platform UI has nothing to list.

What it cannot check, and what to check by hand:

- **that every `failure` has a relationship.** Nothing enforces it.
- **that expressions parse.** They are strings until called. Render the chart and
  read them, and exercise the tool once against the real system.
- **that a `router` covers its cases.** A value matching no condition and no
  `else` stops the run.
- **that `relationships` reaches an exit** in a conversation workflow. A tool
  workflow ends on `push-message` and needs no exit; a loop that never reaches
  one leaks a channel until `channelMaxIdleMs`.
