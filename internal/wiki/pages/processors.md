# The processors, and the fields that decide behaviour

`asgard-cli wiki workflow` says which processors exist and how they wire
together. This page is what each one takes, and it exists because the
per-processor documentation - sixteen pages of it - had never been read into
this material. A chart author writing a Workflow was working from a type list.

Every processor has **Success** and **Failure** outputs. The error branch is
universal, not a property of the risky ones.

## The two LLM processors are not variants of each other

They share almost every field and differ in the two that matter:

| | `llm-completion` | `stream-llm-completion-message` |
|---|---|---|
| **Output Schema** | **required** | not present |
| **Await** | not present | present |
| Payload / Template | not present | present |
| what it is for | a decision or an extraction the flow then uses | text going to a person, as it is produced |

**So the choice is not "streaming or not" - it is whether the output has a
shape.** A processor whose result another processor reads needs a schema and
therefore `llm-completion`. A processor whose result a human reads needs the
streaming one, and gets no schema.

**`Await`, on the streaming one, is the field people set backwards.** On: the
flow waits for the whole response before continuing. Off: the flow continues the
moment the first token is emitted. Off is what makes a chat feel immediate, and
it means **anything after this processor runs while the model is still talking**.

## Where `allowedCubes` and `allowWrite` actually live

Both LLM processors carry the semantic layer configuration, and this is where the
warnings scattered through the extracts land:

    Semantic Layer                a layer this processor may use
    Semantic Layer Allow Query
    Semantic Layer Allow Write
    Semantic Layer Allowed Cubes

**`Allowed Cubes` is the field whose absence is the whole argument against
mounting a layer for an anonymous audience** - without it, the model composes SQL
over every cube, and the surface grows each time one is added. See
`asgard-cli next --stage read-path`.

**`Allow Write` omitted resolves to true.** The trigger extract says this and it
is a processor field, so it is written out on every entry that has a layer, on
purpose, even when false. A scheduled run with a missing `allowWrite` has a write
path into the customer's systems that nothing in the rendered chart shows.

**`Toolsets Fault Tolerant`** is off by default. On, a tool error is returned to
the model to decide whether to retry, rather than failing the step. Off is right
when a failed tool call must not be retried - which is most writes.

## Fields that decide something and look like detail

| processor | field | what it decides |
|---|---|---|
| `execute-script` | **Engine** | only `ECMA5` is supported. Not modern JavaScript - no `let`, no arrow functions, no template literals. This is the single most common surprise in this list |
| `http-request` | **Parse JSON** | off by default. On, `httpResponse` gains a `json` field. Off, the body is a string and every downstream expression has to parse it |
| `validate-payload` | **Schema** (required), **File Requirements** | this is the entry contract of an automation tool - what the caller must supply, and which file types are accepted |
| `query-sql` | **SQL Type Arguments** | parameterised queries. The alternative is string-building a query, which is the injection surface `fixed-query-tools` exists to remove |
| `retrieve-knowledge` | **Similarity Threshold** (required) | there is no safe default to fall back on. Too high returns nothing and looks like an empty knowledge base |
| `retrieve-knowledge` | **Filter Tags**, **Path Exists**, **Path Predicate** | retrieval can be scoped without splitting the knowledge base |
| `generate-embedding` | **Result Field** (required) | where the vector lands. Every model processor that produces data names its own output field |
| `push-message` | **Flush** | whether the reply buffer is sent immediately |
| `router` | **Else** | the unmatched branch. A router without one silently drops what does not match |

## Entry, Exit and Router

`flow-entry` and `flow-exit` are connection points rather than work. Two things
worth knowing from their pages:

  - **a workflow can have several entries**, which is how one Workflow serves
    more than one caller shape
  - **an exit connects workflows to each other**, which is the mechanism behind
    `asgard-cli usecase workflow-chain`

## Corresponding extracts

`asgard-cli usecase external-api` uses `http-request` field by field;
`workflow-chain` uses `router` and the entry/exit connection;
`fixed-query-tools` uses `query-sql`;
`knowledge-drive` uses `retrieve-knowledge`.

## Sources

- The sixteen pages under
  [developer-reference/processor](https://docs.asgard-ai.com/docs/developer-reference/processor)
  - asgard-docs `f00e0ee`. **None of them had been read into this material
  before 2026-09-02**, which is why `workflow.md` carried a type table and
  nothing below it
- `ProcessorDefinitions` in asgard-core `internal/constants.go`, read 2026-09-02
  through the API: the same set, with the static config definitions the
  documentation describes in prose

**Unchecked:** every field here comes from the product documentation, not from a
chart that sets it. Where an extract uses one - `http-request`, `query-sql`,
`router` - that extract is the checked version and wins. The `ECMA5` limit and
the `Await` semantics are the two most worth confirming against a running
deployment, because both change what can be written rather than how it is
configured.
