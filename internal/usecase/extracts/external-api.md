# External HTTP APIs

Reading from, and writing to, a system that is not a database - a SaaS platform,
a partner's REST API, an internal service.

**Seen in:** a shopping assistant whose entire read path is two HTTP APIs, and a
notification chain that posts outward to a mail endpoint.

**Checked:** 2026-09-02 against a search tool's http-request workflow - update-context first, parseJson, header configs, httpResponse - and the CRD.

**Unchecked:** the advice on measuring the request against the real API. That is the step this page exists to make you take, and it cannot be done here.

**Read the platform side first:** `asgard-cli wiki api` -
the endpoint, the SSE event sequence, and the four integration patterns. This page assumes you have.

## First: is an API the right route at all?

Per system, take the most capable route it offers:

| route | what you get | shape |
|---|---|---|
| **a database you can read** | the agent composes its own queries and joins across tables | `asgard-cli usecase semantic-layer` |
| **an API** | a fixed set of calls, but a real contract - reviewable, and gateable for writes | this extract |
| **a screen a person clicks** | last resort. Brittle, slow, and it breaks whenever the vendor changes their UI | no extract yet - raise it as a question first |

**A system offering both: read from the database, write through the API.** The
database gives the agent questions nobody thought to expose an endpoint for; the
API gives writes a gate.

**Before building one integration per external system, ask whether the customer
already consolidates them.** Middleware, an OMS, a warehouse that already pulls
the channels in - if one exists, several external APIs collapse into one database
and the design gets simpler in every dimension. If nobody knows, that is an open
question in `docs/open-questions.md`, not an assumption to design on.

## When this shape, and when not

There are **two ways to reach an external API**, and the choice is not stylistic:

| | **A. `http-request` in a Workflow** | **B. the agent calls it from its sandbox** |
|---|---|---|
| Where the call happens | the platform, inside a tool | inside the agent's sandbox, taught by a skill |
| Protocol | HTTP only | **anything a client exists for** - SNMP, SSH, a vendor CLI, a database client |
| The contract | an `inputSchema` you wrote | whatever the skill describes |
| Reviewable | yes - the URL, the body and the parsing are in version control | no - the agent composes the call |
| Can gate on human approval | **yes**, `requestConsent: true` | no |
| Use for | **anything with a side effect**, and reads you want pinned | reads only, where the surface is wide and exploratory |

**Every write goes through A, with `requestConsent: true`.** Never let a sandbox
make a side-effecting call to an external system: there is no gate, no record of
what was sent, and no way to review the request shape before it goes out.

Reads may use B in two cases:

- **the API is large and the useful calls are not knowable in advance** - a back
  office with fifty endpoints, where a tool per endpoint is the wrong trade.
  Give the agent a skill describing the contract and let it compose calls.
- **the system does not speak HTTP at all.** The sandbox is a real environment,
  so a protocol with a client - SNMP, SSH, a vendor CLI - is reached by running
  that client. `http-request` is one processor type, not the limit of what the
  platform can integrate.

Either way the reads stay read-only, and **a write still goes through A**.

For the second case the skill carries what a person would need: how to
authenticate, which commands are safe, how to read the output, and which
commands are refused outright. Credentials arrive the same way as for HTTP -
written into the sandbox by a hook, never baked into the skill.

## The shape (A)

    Workflow  wf-<verb-noun>
      entries[].inputSchema      the tool's parameters - real ones, unlike a SQL tool
      processors:
        update-context           pull the parameters out FIRST
        http-request             the call
        push-message             shape the response for the agent
    Toolset  ts-<name>
      tools[] -> (workflow, entry), requestConsent per tool

## Generate it

    asgard-cli add httptool <name> --toolset ts-<name>

That writes the structure below with the fields that fail silently already in
place - the display annotation, the labels the UI needs, the current field names.
**Copying the skeleton by hand is where those get lost**, because nothing tells
you they are missing: not helm lint, not CRD validation, not a server dry-run.

The generated file marks the judgement calls TODO. Those are what the rest of
this page is about.

## The skeleton

`templates/tool/wf-<verb-noun>.yaml`.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: Workflow
metadata:
  name: wf-<verb-noun>
  annotations:
    asgard-ai.com/workflow-name: "<display name>"
    {{- include "<chart>.workflowSetAnnotations" (dict "name" "wf-<verb-noun>" "displayName" "<display name>") | nindent 4 }}
  labels:
    {{- include "<chart>.workflowSetLabels" (dict "name" "wf-<verb-noun>" "type" "automation_tool") | nindent 4 }}
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  variables: []
  entries:
    - name: entry-main
      handlingProcessor: proc-input
      labels:
        display_name: Entry
      inputSchema: |-
        {
          "type": "object",
          "properties": {
            "q": {
              "type": "string",
              "description": "<what to pass, and what NOT to - e.g. the thing being searched for, not the whole conversation>"
            }
          },
          "required": ["q"]
        }
      tooling:
        name: <snake_case_name>
        description: |-
          <what it does, what it returns, and when the agent must call it>
        allowUploadFile: false
  exits: []
  processors:
    # Take the parameters out of the payload BEFORE the call. See below.
    - name: proc-input
      type: update-context
      labels:
        display_name: <read parameters>
      configs:
        - name: searchQuery
          expression: |-
            (() => {
              return String(prevPayload.q || "").trim();
            })()

    - name: proc-call
      type: http-request
      labels:
        display_name: HTTP Request
      configs:
        - name: url
          value: {{ .Values.<service>.endpoint | quote }}
        - name: method
          value: POST
        - name: parseJson
          value: "true"
        - name: body
          expression: |-
            (() => {
              return JSON.stringify({ q: searchQuery });
            })()
        # Headers are configs, named after the header itself.
        - name: Content-Type
          value: application/json

    - name: proc-response
      type: push-message
      labels:
        display_name: Response
      configs:
        - name: payload
          expression: |-
            (() => {
              const items = (httpResponse && httpResponse.json && httpResponse.json.items) || [];
              return { items };
            })()
```

Endpoints and non-secret settings are `chartValues` set on the platform; a token is
declared under `appSecret` and read with a `secretKeyRef`, never a value. The Secret
belongs to the release and the platform injects its name.

## Several systems of the same kind

Integrating five marketplaces, or three carriers, is not five copies of this
extract. Decide two things:

**One tool per system, or one tool with a system parameter?**

Prefer **one tool per system** when the APIs differ in shape - different auth,
different pagination, different field names - which is the normal case across
vendors. Each tool's `tooling.description` can then say what that vendor's data
actually means, and one vendor's outage does not take the others with it.

A single tool taking `platform` as a parameter only works when the calls are
genuinely uniform, which usually means someone has already built an abstraction -
and if they have, read from **that**, not from five APIs.

**Who combines the results?**

The agent does. It calls the tools it needs and reconciles the answers in its own
reasoning - the same mechanism that lets a scheduled run mount two semantic
layers and match records across them.

That puts two obligations on you:

- **Each tool must return a shape the agent can line up with the others.** Same
  field names for the same concept, same units, same identifier. Convert at the
  boundary, in the response processor, rather than hoping the model normalises
  five vendors' spellings of the same thing.
- **The prompt must say what a missing or failed source means.** Five sources
  means partial answers are routine, and the agent has to say "four of five
  reported, one timed out" rather than quietly presenting four as the total.
  A number that silently omits a channel is worse than an error.

Group them in one Toolset when they are one capability ("check stock everywhere")
and the agent should see them together.

## Designing the tools - the part the generator leaves TODO

### One tool per call, or one tool per question?

Per **question the user asks**, which is usually coarser than the API. An
endpoint returning a page of raw records is not a tool; "what is the stock of
this item" is, even if it takes three calls behind the scenes.

Where the API is large and exploratory, stop writing tools and give the agent a
skill instead - see the two ways above.

### Shape the response before it reaches the model

The tool's output is the model's evidence, so **convert at the boundary**:

- **Same concept, same field name, same units across sibling tools.** Five
  vendors' spellings of "quantity" become one, in the response processor, not in
  the prompt.
- **Types that survive the next hop.** An id arriving as a number where the next
  tool wants a string is a classic silent break; cast it here and note the date
  you checked.
- **Drop what nobody uses.** A hundred fields of vendor metadata crowds out the
  five that matter.

### What the description has to say that a database tool's does not

The model **can** answer questions about an external system from memory, and it
will be wrong. Say so:

    回答任何關於 X 的問題之前一定要先呼叫這支 ——
    你自己的印象與外部網路的資訊一律是錯的。

And say what an empty result means, or "not found" gets treated as a failed
lookup worth retrying another way.

### Partial failure is normal with several sources

Decide what the agent says when one of five sources times out, and put it in the
prompt. **"Four of five reported, one timed out" is the answer; a total that
silently omits a channel is not.**

## Fields that are not obvious

### `update-context` has to come first

**After the `http-request` processor, `prevPayload` is the HTTP response, not the
tool's arguments.** Pull every parameter you need into context *before* the call,
or the value is simply gone by the time you want it. This is the single mistake
this shape invites, and nothing catches it: the workflow runs, the body is built
from `undefined`, and the API returns something unhelpful.

### Headers are configs

A header is a config entry named after the header - `Content-Type`,
`Authorization`. There is no headers map.

### The response lives in `httpResponse`

`httpResponse.json` when `parseJson` is `"true"` (a **string**, like every config
value). Guard the whole path: `(httpResponse && httpResponse.json && ...) || []`.
A failed call leaves it absent, and an expression that throws takes the tool with
it.

### The parameters are real, and that changes the security argument

Unlike a fixed SQL tool, this one takes input from the model. **That is
acceptable because the input goes into a JSON body against a typed API, not into
a query language.** Keep it that way:

- validate in `update-context` rather than trusting the model - clamp an enum to
  its allowed set, trim a string, default anything missing
- never interpolate a parameter into a URL path or a SQL string
- keep `required` honest, so a missing argument fails loudly instead of
  silently searching for `""`

### Write down what you measured about the response

Field types across a boundary are a classic source of silent breakage - an id
that arrives as a number where the next tool wants a string. Convert at the
boundary and say so in a comment with the date you checked. The API can change;
your note is what makes that discoverable.

## The `tooling.description` is doing more work here than anywhere else

With a database tool, the agent cannot answer without it. With an external API it
can - **from memory, and it will be wrong**. Say so explicitly:

> call this before answering any question about X. Your own impression, and
> anything from the open web, is wrong here.

And say what an empty result means, or the agent will treat it as a failed
lookup and try something else:

> an empty list means the platform does not carry it. Do not go looking another
> way.

## Writes: the approval gate

A tool with a side effect sets `requestConsent: true` on its entry in the
Toolset. The harness intercepts the call and waits for a human. **The prompt
should not describe an approval flow** - the gate is the harness's, and prompting
around it only invites the model to work past it.

Two consequences worth planning for:

- **A scheduled run cannot use a consenting tool.** Nobody is there at 03:00, so
  `requestConsent: true` parks the run until it times out. A Trigger's toolset
  needs `requestConsent: false`, which means a write path and a scheduled path
  **cannot share a Toolset**. Keep them separate CRs and say why in the header.
- **A mock must announce itself.** Where the endpoint is not wired up yet,
  returning success is deliberate - a failure would stop a cursor and the chain
  would never be exercised. That makes disclosure the safety property: the tool
  description and the prompt must both require the summary to say nothing was
  actually sent.

## Verify

```bash
asgard-cli check
helm lint projects/<project>/chart/app
asgard-cli verify <project>
```

The xref check resolves the `(workflow, entry)` pairs. **It cannot check the URL,
the body, or the parsing** - and a wrong `configs[].name` lints clean, passes CRD
validation and then does nothing at runtime, because that field is a free-form
string.

So exercise it for real before trusting it:

```bash
# call the API directly with the same body the expression builds
curl -X POST "<endpoint>" -H 'Content-Type: application/json' -d '<body>'
```

Compare that against what the tool returns in a conversation. Anything the two
disagree about is in your expressions.
