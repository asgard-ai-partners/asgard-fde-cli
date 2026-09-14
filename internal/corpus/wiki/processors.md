# The processors, and the fields that decide behaviour

`../wiki/workflow.md` says which processors exist and how they wire
together. This page is what each one takes, and it exists because the
per-processor documentation - fifteen pages and an introduction - had never
been read into this material. A chart author writing a Workflow was working from a type list.

**Almost every processor has a Failure output, and one does not.** Of the
thirteen, only `update-context` documents Success alone; `http-request` fails on
a non-200 or a network error and produces `prevError`, and `query-database`,
`retrieve-knowledge`, `execute-script` and `push-message` all document one too.
Draw the error branch.

**`ProcessorDefinitions`' `StaticRelationships` is not the contract.** It lists
five processors with a Failure relation and the rest with Success or nothing,
and it is incomplete in both directions: it gives `listen-message` no
relationships at all while every production chart continues from one, and
`http-request` Success only while four production charts across two
repositories route `failure` off it and the documentation describes that branch
in full.

That list is the third thing in `ProcessorDefinitions` to be incomplete -
`await` and the config keys are the others. **Treat it as what the definitions
declare, never as what the platform accepts**, and when the two disagree the
deployed charts win. Do not build a check on it.

## Every field is one of three kinds of value

This is the vocabulary the per-processor pages all assume, and `workflow.md` says
only that one is JavaScript and one is Handlebars:

    Literal      a fixed value
    Expression   JavaScript, evaluated per run. `prevMessage || '訪客'`
    Template     Handlebars, for producing text. `{{#if prevMessage}}...{{/if}}`

**The ECMA5 limit is `execute-script`'s Engine field, and does not reach the
Expression fields.** The deployed charts settle it. Across 520
`expression:` values in the reference charts exactly one uses an arrow function
- `prevBlobs.map(b => b.blobId).join(',')`, in a shipped tenant chart - so
arrow functions evaluate. What no chart uses anywhere is `const`, `let` or a
template literal, and the reason is structural rather than a limit: **an
Expression field holds one expression, not statements**, so there is nothing for
a declaration to do in it.

`execute-script` is the other thing. Its **Engine** takes `ECMA5` and the
documentation says only `ECMA5` is supported, and that body *is* statements -
which is where the restriction bites, and where it belongs in your head.

### The variables in scope

| variable | type | what it is |
|---|---|---|
| `prevMessage` | string | the previous user message |
| `prevBlobs` | array of Blob | files attached to it |
| `prevPayload` | object | the payload from whatever called in - **this is what a BotProvider passes through**, and what `pluginNames` and `sourceSetMounts` expressions read |
| `prevError` | | the previous step's error, on a Failure branch |
| `customChannelId` | string | the conversation key, chosen by the caller |
| `customMessageId` | string | the message id, optional |
| `prevToolCalls` | array | **what the agent just called, and what came back.** Not in the documentation, not in the expression pages, and used in eight places in a production chart |

    interface Blob {
      blobId: number; fileType: FileType; fileName?: string;
      size: number; mime: string;
    }
    type FileType = 'BINARY' | 'IMAGE' | 'VIDEO' | 'AUDIO' | 'DOCUMENT'

**`prevToolCalls` is undocumented and is the one that unlocks post-processing.**
The documentation's variable page does not list it; it is in the charts. Each
entry carries the tool's name, the arguments it was called with, and its
result:

    prevToolCalls[i].toolName          the tool that was called
    prevToolCalls[i].parameter.<arg>   the arguments, by their own names
    prevToolCalls[i].output.isSuccess  whether it worked
    prevToolCalls[i].output.data       what it returned

So a step **after** the model can ask "did the agent call this tool, and did it
work" - which is how a chart reacts to what an agent did without asking the
model to say what it did. The idiom in the chart it came from is
`prevToolCalls.filter(c => c.toolName == "x" && c.output.isSuccess &&
c.output.data).length > 0`, and reading `.parameter` back off the same entry is
how the arguments reach the reply.

**Every one of them can be absent, and ECMA5 has no optional chaining.** So the
documentation's own examples are all defensive, and a chart that is not will
throw at run time on the turn where a user sends no file:

    prevBlobs && prevBlobs[0] && prevBlobs[0].fileName
    prevPayload && 'property' in prevPayload ? prevPayload.property : '預設值'

### The seven built-in functions

| function | what it does |
|---|---|
| `history(start, end)` | conversation history as plain text, one line per turn. Indices are inclusive and **negative counts from the end** - `history(0, -1)` is everything |
| `historySize()` | how many turns there are |
| | **`prevMessage` against `history`** is the choice: one turn, or context. `history(-3, -1)` is the last three, and an echo or a single lookup wants neither |
| `urlEncode(s)` | for building a URL in an `http-request` |
| `xpathExtract(...)` | pull a value out of XML or HTML |
| `vecToStr(...)` | a vector as a string |
| `isoNow()` / `isoToday()` | the timestamp and the date |

**`history` is the one that decides a design.** Feeding a whole conversation into
a prompt is one call, and it is also how a run reaches the 30-step and
three-minute ceiling early. `historySize()` first, then a bounded window.

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
`../guide/read-path.md`.

**The safe field defaults off and the dangerous one defaults on.** In the
definitions, `semanticLayer.allowQuery` defaults to **false** and
`semanticLayer.allowWrite` defaults to **true**, on both LLM processors. So a
processor configured by adding only what you want has a write path, and the
rendered chart shows nothing. That is why every deployment writes
`allowWrite: false` out on every entry that has a layer, even where it is
obviously false.

**`query-database` carries the same pair and is easier to miss**, because the
name reads as a read tool: `allowWrite` defaults to true there too, and its
`allowedTables` is the counterpart of `allowedCubes` - both default to empty,
meaning unrestricted. A scheduled run with a missing `allowWrite` has a write
path into the customer's systems.

The chart keys are `semanticLayer.allowQuery`, `semanticLayer.allowWrite` and
`semanticLayer.allowedCubes`, not the spaced labels the builder shows.

**`Toolsets Fault Tolerant`** is off by default. On, a tool error is returned to
the model to decide whether to retry, rather than failing the step. Off is right
when a failed tool call must not be retried - which is most writes.

## `effort` fails the turn when the model does not take it

The reasoning-effort field takes `low`, `medium`, `high`, `xhigh`, `max`, or
`auto` to leave it to the model. It sets thinking depth and token spend
together, so a higher level is slower and costs more.

**Sending a level a model does not support fails that turn outright** - not a
degraded answer, a failure. Some models do not accept the parameter at all: a
non-reasoning model like `gpt-4.1-mini` rejects it, and the speed-oriented
Haiku class is the other to know.

**Omitting the field is not the same as disabling it, and that is the trap.**
There are three states, not two:

    a level          `--effort <level>` is sent
    the field absent no `--effort` is sent - **and the driver treats a model id
                     it does not recognise as supporting effort, so it supplies
                     a level of its own, on the high side**
    `disabled`       no `--effort`, and the spawn declares the capability as
                     unsupported, so nothing is supplied

So a chart on a non-reasoning model must write `disabled` explicitly. Leaving
the field out is how one deployment's token spend went up without anyone
changing a prompt. `effort` is per model rather than per chart: switching
`completionModelName` can turn a working chart into one that fails on every
turn, or one that quietly buys reasoning nobody asked for.

**`input` on `stream-llm-completion-message` is normally left empty.** Empty
means the processor uses what the pipe handed it - `prevMessage` in context,
which is what the user actually said. Set it only when the input has been
worked on first: a summary prepended, or the question rewritten into something
clearer.

## Fields that decide something and look like detail

| processor | field | what it decides |
|---|---|---|
| `execute-script` | **Engine** | only `ECMA5` is supported, so the script body is not modern JavaScript - no `let`, no arrow functions, no template literals. **This applies to the script body only**; Expression fields elsewhere are ordinary JavaScript, and `../wiki/workflow.md` gives the evidence |
| `http-request` | **Parse JSON** | off by default. On, `httpResponse` gains a `json` field. Off, the body is a string and every downstream expression has to parse it |
| `validate-payload` | **Schema** (required), **File Requirements** | this is the entry contract of an automation tool - what the caller must supply, and which file types are accepted |
| `query-database` | **SQL Type Arguments** | parameterised queries, supplied as extra keys rather than a static field. The alternative is string-building a query, which is the injection surface `../usecase/fixed-query-tools.md` exists to remove. **The docs page for this processor is called `query-sql`** - see the naming table below |
| `retrieve-knowledge` | **Similarity Threshold** (required) | there is no safe default to fall back on. Too high returns nothing and looks like an empty knowledge base |
| `retrieve-knowledge` | **Filter Tags**, **Path Exists**, **Path Predicate** | retrieval can be scoped without splitting the knowledge base |
| `generate-embedding` | **Result Field** (required) | where the vector lands. Every model processor that produces data names its own output field |
| `push-message` | **Flush** | whether the reply buffer is sent immediately |
| `router` | **Else** | the unmatched branch. A router without one silently drops what does not match |
| `router` | **its own config keys** | a router has no fixed fields: **each config key you add is a branch name**, and its value is a boolean expression. A relationship out of the router carries that key as its `relationName`, and `else` is the static one. See below |

## Defaults the documentation does not give

The type definitions carry a default per config key and the documentation pages
do not, so a field marked 必填 there can still have a value it falls back to -
which changes whether leaving it out is an error or a silent choice.

The three worth knowing before the table:

| processor | field | |
|---|---|---|
| `retrieve-knowledge` | `sampleK` | **required, and defaults to 20.** How many chunks come back. Nothing in the documentation gives the number, so a retrieval returning "too much" or "not enough" is being tuned against an invisible 20 |
| `validate-payload` | `path` | defaults to `"$"` - the whole payload. Set it to validate a subtree instead |
| both LLM processors, and `query-database` | `allowWrite` | **defaults to true**, while `allowQuery` beside it defaults to false. The section above is about this one |

`retrieve-knowledge` requires **five** fields - the knowledge bases, the query
text, the similarity threshold, the result field and `sampleK` - and four of
them have no default at all, so a partially configured one fails rather than
guessing. **`stream-llm-completion-message` requires only two**, and notably not
`prompt`, where `llm-completion` requires it: the streaming one is a channel,
and what it says can come from its `input` instead.

### Every processor, its outputs, and what it requires

Extracted from the definitions, so these are the keys a chart writes rather than
the labels the builder shows. `=` marks a **required key that also has a
default** - omitting one of those is a silent choice rather than an error.

**This table is a subset of what a chart may set, not the contract.** `await`,
described above and set in five separate production deployments, is in neither
`ProcessorDefinitions` nor the CRD; `temperature` is the same. So a key missing
from the row below is not a key you may not use - read the row as "these are
declared". **The wider set is the editor palette**, and it is a table of its
own: "What the editor lets an author set" below, which has both of those keys
and says which keys are the platform's rather than yours. `asgard-cli verify` reflects this: it fails
on a **required** key with no default, because that is broken against any
version, and does not complain about a key it has never heard of.

| processor | outputs | extra keys | required keys, with any default |
|---|---|---|---|
| `execute-script` | Success | no | `engine` `script` |
| `generate-embedding` | Success + Failure | no | `embeddingModel` `input` `resultField` |
| `http-request` | Success | **yes** | `url` `method` `parseJson` =false |
| `listen-message` | none | no | *none* |
| `llm-completion` | Success + Failure | **yes** | `completionModel` `prompt` `outputSchema` |
| `llm-query-database` | Success + Failure | no | `semanticLayer` `query` `resultField` `completionModel` `maxTokens` |
| `push-message` | Success | no | `message` ="" `flush` =false `isDebug` =false **(the platform's)** |
| `query-database` | Success | **yes** | `dataConnector` `sql` `resultField` |
| `retrieve-knowledge` | Success | **yes** | `knowledgeBases` `textQuery` `similarityThreshold` `resultField` `sampleK` =20 |
| `router` | Else | **yes**, + branches | *none* |
| `stream-llm-completion-message` | Success + Failure | no | `completionModel` `isDebug` =false **(the platform's)** |
| `update-context` | Success | **yes** | *none* |
| `validate-payload` | Success + Failure | **yes** | `schema` |

**`update-context` and `router` require nothing and take arbitrary keys**, which
is the whole of how they work: on `update-context` the extra key names *are* the
context variables you are setting, and on `router` they are the cases, with a
dynamic output per case and the static `Else` for what matches none. This is why
neither has a static field list to look up.

### What an extra key means, per processor

"Takes extra keys" is not one mechanism. Each processor reads them its own way,
and **the key name is the parameter** - so a typo is not an error, it is a
different parameter or a silently ignored one.

| processor | an extra key is | the shape |
|---|---|---|
| `update-context` | a context variable you are setting | the key **is** the variable name |
| `router` | a branch | the key is the branch name, the value a boolean or an expression returning one. An unknown value type fails the step |
| `http-request` | an **HTTP header** | the key is the header name. **The value must be a string** or the step fails - a number written bare is a failure at run time, not at render |
| `query-database` | one SQL placeholder | `sql.args.<n>.type` **and** `sql.args.<n>.value`, both, numbered **from 1** |
| `retrieve-knowledge` | one JSON-path filter | `path.<n>.exists` or `path.<n>.predicate`, each series numbered **from 0** and counted independently |
| `validate-payload` | one file requirement | `file.<n>.type`, `file.<n>.alias`, numbered **from 0** |
| `llm-completion` | **nothing.** It declares dynamic config and no code reads it | an extra key here is accepted and ignored |

**The numbered ones stop at the first gap.** Each is a loop that breaks the
moment an index is missing, so `sql.args.1.*` and `sql.args.3.*` with no `2`
sends **one** argument, not two - and nothing says so. The same holds for a
`file.0.type` with no `file.0.alias` where the alias series is read separately,
and for `path.<n>.*`. Renumber after deleting one.

**Two of them start at 0 and one starts at 1.** There is no rule behind it to
remember; it is per processor, and getting it wrong on `query-database` means an
argument list that is silently empty.

**Extra keys are also where `sqlTypeArguments` lives.** The documentation gives
`query-sql` a **SQL Type Arguments** field for the `$1`, `$2` placeholders, and
there is no such static key - the parameters are dynamic config on the
processor. Looking for the key in the type definitions and not finding it does
not mean the documentation is wrong.

**The documentation and the definitions disagree about what is required, and
the direction has flipped.** At `f00e0ee` the `query-sql` page gave
`ResultField` a 預設值為 `result` that the definition does not have; asgard-docs
resynced the processor pages against asgard-core `internal/constants.go` on
2026-09-04 and that default is gone. What is left is the other direction: the streaming page
marks **MaxTokens（必填）** where the definition has `maxTokens` optional with no
default, and `Temperature` 可選 where the definition has no such key on that
processor at all.

**So neither side is the one to trust for requiredness.** The definitions are
what the runtime validates against - that is what `asgard-cli verify` reflects -
and a field the documentation calls 必填 may be one the platform accepts without.
Write it out anyway: a page that says 必填 is a page somebody will be held to.

## What the editor lets an author set, which is a third source

**Three sources describe a processor's configuration and they do not agree.**
The CRD enum says which types exist; `ProcessorDefinitions` in asgard-core says
which keys are declared; and the **editor palette** says which keys an author
can actually set in the builder. asgard-docs added a page per processor on
2026-09-09, each one verified against the palette rather than only against the
code, and states why the code alone is not enough: the definitions "omit the
section a `has_dynamic_config` processor generates, and list keys the editor
hides".

The table below is the palette's view. **`author` is what somebody filling in a
node sees; `platform` is set for them and is not theirs to write.**

| type | author sets | platform sets | dynamic | scope |
|---|---|---|---|---|
| `execute-script` | `engine` `script` | - | no | general |
| `update-context` | *none* | - | **yes** | general |
| `http-request` | `url` `method` `parseJson` `body` | - | **yes** | general |
| `router` | one per branch | - | **yes** | general |
| `listen-message` | *none* | - | no | general |
| `push-message` | `message` `template` `flush` `payload` | `isDebug` | no | **bot, automation_tool** |
| `validate-payload` | `schema` `path` | - | no | **automation_tool** |
| `generate-embedding` | `embeddingModel` `input` `resultField` | - | no | general |
| `llm-completion` | `completionModel` `prompt` `outputSchema` `maxTokens` `temperature` `effort` `toolsets` `semanticLayers` `openai.webSearch.enabled` `openai.webSearch.searchContextSize` `sandboxBlueprint` | `blobs` `semanticLayer` `semanticLayer.allowQuery` `semanticLayer.allowWrite` `semanticLayer.allowedCubes` | no | general |
| `stream-llm-completion-message` | the same, plus `input` `await` `semanticLayers.dataVisualization`, minus `outputSchema` | the same, plus `isDebug` | no | general |
| `query-database` | `dataConnector` `sql` `resultField` | `allowWrite` `allowedTables` | **yes** | general |
| `retrieve-knowledge` | `knowledgeBases` `textQuery` `similarityThreshold` `resultField` `filterTags` `sampleK` | - | **yes** | general |
| `llm-query-database` | `semanticLayer` `query` `resultField` `completionModel` `maxTokens` `temperature` | - | no | **not in the palette** |

Four things in it that are not anywhere else:

**`await` and `temperature` are author keys on the streaming processor.** They
are set in five production deployments and appear in neither the definitions nor
the CRD. The palette was the only source that described them until 2026-09-09,
when `model-stream-llm-completion` gained an **Await** section, a **Temperature**
section and a worked example that writes `await` as a config key - so the
documentation now carries what the definitions do not. **A key absent from the
definitions is not a key you may not use** - check the palette or the page
rather than guess.

**`llm-query-database` is in the CRD and in the definitions and not in the
builder.** An author cannot add it from the editor; a chart can still declare
it. Nothing says whether that is deliberate.

**A processor is scoped to a kind of workflow set.** `validate-payload` is
`automation_tool` only, `push-message` is `bot` and `automation_tool`, and
everything else is `general`. So `push-message` is the same CRD type reached
from two different places, which is why the documentation has two pages for it -
`message-push` and `automation-tool-response`.

**`isDebug`, `allowWrite`, `allowedTables` and the `semanticLayer.*` keys are
the platform's, not the author's.** The definitions list `isDebug` as a
required key with a default of false, which reads as something to write out; the
palette says the platform sets it. The same holds for the per-layer permissions
- see "Where `allowedCubes` and `allowWrite` actually live" above, which the
palette confirms from the other side.

**The definitions under-report Failure branches**, and the palette agrees with
the documentation against them: `execute-script`, `http-request`,
`push-message`, `query-database` and `retrieve-knowledge` all have one. The row
in the table above this section says Success alone for several of those, because
it was extracted from the definitions. Read the definitions as declared, not as
complete.

**For a field's meaning, read the page rather than this one.** Each carries the
property panel as a screenshot, every field's default, and a worked example -
`https://docs.asgard-ai.com/docs/developer-reference/processor/<name>`, where
`<name>` is the documentation's name and not the chart's. The mismatch is the
section below.

## The documentation's names are not the chart's names

Sixteen pages sit under `developer-reference/processor` as of asgard-docs
`23409b3`, plus an introduction - fifteen at `f00e0ee`. **The new one is
`query-llm-database`**, which is the one type that is documented and is not in
the editor palette: a chart can declare it, an author cannot add it from the
builder. The CRD enum has thirteen types. They do not line up, and the mismatches are
each a place where searching for what you read finds nothing:

| the page is called | the chart writes |
|---|---|
| SQL, at `processor/query-sql` | `query-database` |
| Entry, at `processor/entry` | *not a processor* - `spec.entries` |
| Exit, at `processor/exit` | *not a processor* - `spec.exits` |
| Response, at `processor/automation-tool-response` | `push-message`. **There is no `response` type** - an Automation Tool's final output is the same processor a bot replies with, scoped to `automation_tool` |
| LLM Database, at `processor/query-llm-database` | `llm-query-database` - **the two words are swapped**, so the page and the type do not find each other |
| MCP Servers, the field label on both LLM processors | `toolsets`, a comma-separated list of Toolset names |

### And a page has a third name: the file it is in

**Half the processor pages are served at a URL that is not their file name.**
Eight of the sixteen declare a `slug:` in their frontmatter, so
`flow-entry.mdx` answers at `processor/entry`, `message-push.mdx` at
`processor/push-message`, `model-stream-llm-completion.mdx` at
`processor/stream-llm-completion`. The file is named for the processor's family
and the URL for the builder's label.

Which matters twice. **A link built from a file name 404s**, and that is how
the Entry row in the table above got a URL with `flow-` on the front of it,
written while correcting something else on the same line. And **a grep of the
docs repository finds the family name**, so searching it for `entry` finds a
file called something else.

The documentation's own index page links by URL and gets them right. Its
category headings are the node menu's, which is a fourth naming of the same
thirteen things and the one an author actually sees:

    流程控制    Entry, Exit, Router
    Message     Push Message, Listen Message
    Model       LLM Completion, Stream LLM Completion Message, Generate Embedding
    Action      Update Context, Execute Script
    Query       SQL, Retrieve Knowledge
    API         HTTP 請求
    Automation Tool   Validate Payload, Response - **not in the Flow Agent
                menu at all**, only in an Automation Tool workflow
    CRD only    LLM Query Database

**`llm-query-database` got its page on 2026-09-09 and is still the one type
the editor will not add.** It
takes `semanticLayer`, `query`, `resultField`, `completionModel` and `maxTokens`
- all five required - plus an optional `temperature`. It is the processor for
"let a model answer this question against the layer" as a single step, where
the alternative is an `llm-completion` with a layer mounted and a prompt. Note
what it does **not** have: no `allowQuery`, no `allowWrite`, no `allowedCubes`.
So the field the read-path argument turns on does not exist on it, and its
access is whatever the SemanticLayer itself permits. Check that before reaching
for it as the safer-looking option, because it is not obviously safer - it is
differently scoped, and the scoping is somewhere else.

## Router at scale is a chain, not a switch

The mental model that costs a day is "a router is a switch with N branches".
The largest use of routers read here - a content-generation deployment with five
across two workflows - is not that shape at all. Each router asks **one boolean
question and has one named branch**, and its `else` goes to the next router.

    proc-route-if-vscode-open-file  is-true -> push the "open in VSCode" CTA
                                    else    -> proc-route-if-download-link
    proc-route-if-download-link     is-true -> push the download CTA
                                    else    -> the next one

Written out, the mechanism is:

  - the router's **config key is the branch name**, and its value is an
    expression returning a boolean. `is-true` is a name somebody chose, not a
    keyword
  - a relationship out of the router names that same key as its `relationName`
  - `else` is the one relationship the type declares statically

**So the cost of a branch is one processor, and branches compose by chaining.**
That is why the shape is a cascade: each question is independent of the others,
several can fire in one turn, and adding a sixth means adding one router and
re-pointing one `else` rather than editing a nine-way condition.

### When a branch belongs in the graph rather than in the prompt

This is the decision an FDE gets wrong, and the deployment answers it clearly.
Every one of those routers runs **after** the model has finished, and asks about
`prevToolCalls` - did the agent call this tool, did it succeed. None of them
asks the model anything.

**Put it in the graph when the answer is a fact about what happened**, and in
the prompt when it is a judgement about what to say. "Did the agent
successfully call `vscode_open_file`" is a fact, it is checkable, and a chart
that checks it will be right every time - where a prompt asking the model to
remember to mention the file will be right most times. The reply the router
builds also reads `.parameter.absolute_path` back off the tool call, so the
content of the message comes from the call rather than from the model's
recollection of it.

The corollary is the test: **if you would have to ask the model to tell you
whether something happened, that branch belongs in the graph.**

## Entry, Exit and Router

**Neither is a processor type.** The documentation files them under
`developer-reference/processor` and the builder draws them as nodes, but the CRD
enum has thirteen types and neither is among them: a Workflow carries
`spec.entries` and `spec.exits` as their own lists, siblings of
`spec.processors`. An entry is `{name, handlingProcessor}` - a named way in that
points at the processor which handles it - and an exit is
`{name, handlingWorkflow}`, pointing at another workflow's entry. Search a chart
for a `flow-entry` processor and you will find nothing.

Two things worth knowing from their pages:

  - **a workflow can have several entries**, which is how one Workflow serves
    more than one caller shape
  - **an exit connects workflows to each other**, which is the mechanism behind
    `../usecase/workflow-chain.md`

## Corresponding extracts

`../usecase/external-api.md` uses `http-request` field by field;
`../usecase/workflow-chain.md` uses `router` and the entry/exit connection;
`../usecase/fixed-query-tools.md` uses `query-database`;
`../usecase/knowledge-drive.md` uses `retrieve-knowledge`.

## Sources

- The value types, variables and functions:
  [expression-introduction](https://docs.asgard-ai.com/docs/developer-reference/asgard-builtin/expression-introduction),
  [expression-variable](https://docs.asgard-ai.com/docs/developer-reference/asgard-builtin/expression-variable),
  [expression-ecma-script-functions](https://docs.asgard-ai.com/docs/developer-reference/asgard-builtin/expression-ecma-script-functions)
  - asgard-docs `f00e0ee`. That section was listed here as deliberately not
  covered, on the grounds that lookup material only goes stale - a judgement made
  before anyone noticed the ECMA5 limit lives in it
- The pages under `developer-reference/processor/`, whose landing page is
  [introduction](https://docs.asgard-ai.com/docs/developer-reference/processor/introduction)
  - **the bare directory URL 404s**; cite the introduction, not the directory
  - asgard-docs `f00e0ee` for the prose
- `effort`'s levels, that an unsupported one fails the turn, and that an empty
  `input` falls back to `prevMessage`: asgard-docs `23409b3`
  `docs/developer-reference/processor/model-llm-completion.mdx` and
  `model-stream-llm-completion.mdx`, read 2026-09-11
- **The three states of `effort`, and that omitting it is not disabling it**:
  read 2026-09-11 off `buy123-asgard-kube`, which carries
  `defaultEffort: "disabled"` with the reasoning in its own values file and in
  `cm-gpt-4.1-mini.yaml`, and cites asgard-core `internal/processor/clidriver/options.go`.
  Confirmed there at `623ceb5`:
  `disabled` becomes a `ModelCapabilities` declaration on the spawn rather than
  a flag, and unset makes the CLI "supply its own default level for any model it
  believes supports effort, and it believes that of every model id it cannot
  recognize (which is all of ours)". **This is the one part of it seen to
  fire** - that deployment's token spend was the symptom. The documentation says
  the same thing more softly (未設定時採用模型預設值), which is why the strong form
  is sourced to the driver
- The editor palette per processor - which keys are the author's, which the
  platform sets, which types accept dynamic config, and which workflow-set
  types each is scoped to: **asgard-docs `23409b3`**, read 2026-09-11 from the
  seventeen `metadata.json` files under
  `content-generator/services/developer-reference/docs/processor/`. Those
  record the palette as a third source beside the CRD enum and asgard-core's
  definitions, and the pages are verified against it rather than only against
  the code. **This is the only part of this page read at that commit** - the
  prose above it is still at `f00e0ee`
- `ProcessorDefinitions` in asgard-core `internal/constants.go`: the per-key
  `IsRequired` and `DefaultValue` the documentation does not carry, the
  Success/Failure declarations, and which processors take arbitrary extra keys.
  Extracted 2026-09-03 at asgard-core `5da86c6` by walking that literal and
  resolving the key constants to their string values - **an earlier
  pattern-based attempt misaligned**, attributing one processor's fields to the
  next, and was discarded rather than published
  - **re-walked 2026-09-11 at `623ceb5` and the thirteen rows are unchanged.**
  asgard-core's `internal/constants.go` itself moved four times in between - a
  card tool, an upload cap, an agent-hub prompt version - which is why "the file
  has not changed" is not the claim to make. The walk is
  asgard-fde-cli `go run ./hack processors`, and it holds both tables on this
  page against that literal
  - and it is **incomplete**: checked against five rendered production charts,
  where `await` appears on the streaming processor in all five and is declared
  nowhere in it
- **What an extra key means, per processor**: read 2026-09-11 at asgard-core
  `623ceb5` off the task implementations themselves - one file per processor
  under asgard-core `internal/processor/task/`. The definitions say only
  *whether* a processor takes dynamic config; the key shapes, the two different
  starting indices and the break-at-the-first-gap behaviour are in the loops that
  read them, and nowhere else. `llm-completion` declaring dynamic config that no
  code reads was found the same way
- The Failure outputs: **the documentation**, one page per processor, after the
  type definitions were found to disagree with four production charts. Checked
  2026-09-03 across `api-http-request`, `query-sql`, `query-retrieve-knowledge`,
  `action-execute-script`, `message-push` and `action-update-context` - five
  document a Failure branch and `update-context` does not
- `prevToolCalls`, the router branch mechanism and the cascade shape: read
  2026-09-03 off `asgard-auto-post-kube`'s two agent workflows, which are the
  only charts anywhere in the reference set that use `prevToolCalls` - eight
  uses, all of them post-processing. It appears in no documentation page and in
  no other deployment
- The thirteen-value `ProcessorType` enum in asgard-kube
  `pkg/apis/asgard/v1alpha1/types.go`, and `WorkflowSpec` beside it, which is
  what settles that entries and exits are not processors

**Unchecked:** the field *meanings* come from the product documentation, not
from a chart that sets them - the names, requiredness, defaults and outputs now
come from the type definitions, which is why they are allowed to contradict the
documentation above and win. Where an extract uses one - `http-request`, `query-sql`,
`router` - that extract is the checked version and wins. The `Await` semantics were
the one most worth confirming, because they change what can be written rather
than how it is configured - and `23409b3` states them per connection point,
which is where the description above now comes from. The `ECMA5` question was
the other, and the charts have answered it.

What is left unchecked is the **palette**: this page has it at second hand, from
asgard-docs' `metadata.json` files, and the file those cite -
`asgard-ai-platform-web` `src/components/react-flow/workflow/processors.json` -
is in a repository nothing here has a clone of.
