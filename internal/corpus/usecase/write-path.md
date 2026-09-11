# Write paths and the approval gate

Anything with a side effect. **This is the shape the platform is built around**,
and the one a customer is usually asking for when they say "and then it does it
for me".

**Seen in:** a deployment whose entire architecture is stated as one rule, and a
notification chain whose single outward action is deliberately mocked.

**Checked:** 2026-09-02 against 72 gated and 14 ungated tool entries across the deployments, with consent level consistent within every Toolset.

**Unchecked:** the reversibility table. It matches what the deployments do, but the deployments did not derive it from this table.

**Read the platform side first:** `../wiki/sindri.md` -
what the governance gate looks like to the person approving it. This page assumes you have.

**Nothing to link.** This shape has no product documentation page - see
`../wiki/sindri.md`. When a customer asks for documentation on the approval
gate, and they will, the answer is a cropped screenshot and this page.

## The rule the platform is built on

> **Reading is autonomous. Writing stops at a gate and waits for a human.**

Everything a `SELECT` can answer goes through a read path with no gate. Anything
with a side effect goes through **`Toolset` -> `Workflow`, with
`requestConsent: true`**, and the harness parks the run until a person approves
it.

That is not a safety add-on, it is the product. A deployment that lets an agent
write without a gate has given up the thing that makes it deployable against real
business systems.

**What this shape costs to get is `../needs/write-path.md`** - what has to come
from the customer before any of it can be built, and the answer that
changes the plan is never the one nobody asked for.

## When this shape, and when not

Use it whenever the action **changes something outside the agent**: placing an
order, updating a record, sending a message, publishing a listing, calling a
partner API that does any of those.

Do **not** use it for reads, however expensive or slow. A gate on a read trains
people to click approve without looking, which is worse than no gate.

**If you are reading this while writing the chart and nobody asked these during
the interview, stop and go back.** They are interview questions - the list above
is what to ask - and a write designed from assumptions is the most expensive
rework in this repository's history A customer
describing "AI prepares it, a person checks it, then it goes out" is describing
this shape, and it changes the project's architecture: a write path needs its own
spec, its own credentials, and a decision about what happens when the human says
no.

## Unknown: what this looks like on an anonymous channel

Every image of the gate is Sindri's dialog, and Sindri is where authenticated
staff work. **On a public channel the person approving is usually the visitor in
the conversation**, not an operator - so whether `requestConsent` renders there,
and as what, is not established. `../wiki/platform-unknowns.md` P8.

Do not promise a customer an approval step on a public channel by showing them a
screenshot of the internal one. A prompt asking "shall I go ahead?" is not the
gate; it is the model being polite, and it can be talked past.

## What to ask the customer, before any of the below

This page was implementation-only, and an FDE asked to design a write had to
invent the questions - which is how a list arrived that designed the customer's
permission model for them. Four things, and only four:

    what may the token we get actually do?
    what does creating one of these require - fields, validation rules?
      and ask for their document rather than for the answer
    is there a test environment we can write into?
    which field on the record identifies the end customer it is for?

**The last one is the only part of their record that is ours.** The agent fills
it. Everything else about how that record behaves inside their system - who it
is attributed to, whether it routes by creator, whether it counts against
somebody's SLA - is theirs, and asking reads as designing their organisation.

**The third one comes before deciding to mock.** See below.

## Two questions to settle before writing any of it

**1. Does the write point outward, or at a source system?**

Pointing **outward** - a notification endpoint, a partner platform, a channel the
customer publishes on - is the common case and the safer one. The systems of
record stay read-only, and a mistake is visible and usually reversible by a
person.

Pointing **at a source system** - updating the ERP, changing stock - is a
different risk class. It needs an explicit decision recorded, and usually a
narrower tool than the one first proposed. Do not let it arrive by accident
because a tool "just needed to update one field".

**2. Who is the human, and what are they looking at?**

The gate shows the call being made. If the person approving cannot tell from that
whether it is right, the gate is theatre. Shape the tool's arguments so the
approval is reviewable: one listing at a time rather than a batch of forty, the
resolved values rather than an id the reviewer would have to look up.

## The shape

    Toolset  ts-<name>
      tools[]
        entrypoint: (workflow, entry)
        requestConsent: true        <- the gate
      -> Workflow  wf-<verb-noun>
           the actual call: http-request outward, or query-database for a write

Bind it with `Agent.managed.toolsetNames` in the hub shape, or
`SandboxBlueprint.toolsetNames` in a flow agent.

## Generate it

    asgard-cli add httptool <name> --toolset ts-<name> --write

That writes the structure below with the fields that fail silently already in
place - the display annotation, the labels the UI needs, the current field names.
**Copying the skeleton by hand is where those get lost**, because nothing tells
you they are missing: not helm lint, not CRD validation, not a server dry-run.

The generated file marks the judgement calls TODO. Those are what the rest of
this page is about.

## The skeleton

`templates/toolset/ts-<name>.yaml` for the gate, and
`templates/tool/wf-<verb-noun>.yaml` for the call behind it. Keep a write
Toolset in its own file, separate from any read-only one - they have opposite
`requestConsent` values and merging them later is how a gate gets dropped.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: Toolset
metadata:
  name: ts-<name>
  annotations:
    asgard-ai.com/toolset-name: "<display name>"
    asgard-ai.com/toolset-description: "<what this can change, in one line>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  toolsetClass: workflow-tooling
  apiKey:
    valueFrom:
      secretKeyRef:
        key: asgard_resource_api_key
        name: {{ include "<chart>.appSecretName" . }}
  tools:
    - entrypoint:
        entry: entry-main
        workflow: wf-<verb-noun>
      # The gate. The harness intercepts the call and waits for a person.
      requestConsent: true
```

The Workflow behind it is the call itself - see `../usecase/external-api.md`
for the `http-request` shape, including where the credential goes
and why parameters have to be pulled into context before the call.

## Designing the gate - the part the generator leaves TODO

### What the human is looking at

The gate shows the call. **If the person approving cannot tell from those
arguments whether it is right, the gate is theatre** - and worse than none,
because it manufactures a record of approval.

Three things follow:

- **Resolve values before they reach the gate.** An id the reviewer would have
  to look up is not reviewable; the name is.
- **One action per call.** Forty listings in one argument gets approved as a
  batch, which means unreviewed.
- **Include what makes it checkable.** For a published price, the price. For a
  message, the whole text. The approval screen is not the place to be terse.

### Split preparing from doing

`requestConsent` is per tool, so a job that prepares something and then acts on
it is **two tools**: preparing is read-only and ungated, acting is gated.

That also makes the conversation better - the person sees the draft, asks for a
change, sees it again, and only then approves the one call that matters.

### What happens when they say no

Decide it, and say it in the prompt: does the agent revise and re-ask, or stop
and report? A run that silently loops on refusal is worse than one that stops.

### Deciding what needs a gate at all

Not everything with a side effect deserves the same treatment, but the line is
not about risk in the abstract - it is about **reversibility and audience**:

| | gate |
|---|---|
| changes a system of record | yes, and it needs its own spec |
| sends something to a person outside the company | yes |
| writes to a scratch area only this agent reads | no |
| anything a `SELECT` could have answered | it is not a write; do not gate a read |

## Fields that are not obvious

### The prompt must not describe the approval flow

**Human approval is the harness's mechanism, not the agent's.** Writing "ask the
user for confirmation before proceeding" into a prompt does not add a gate - it
adds a second, fake one that the model can talk itself past, and it confuses the
real one.

Describe **what the tool does**. The gate happens whether or not the prompt
mentions it.

### The gate is per tool, so split preparing from doing

`requestConsent` sits on a tool, not on a workflow or an agent. So a task that
prepares something and then acts on it is **two tools**: preparing is read-only
and ungated, acting is gated.

That is also the better shape for review. The human sees the call that changes
something, not the research that led to it.

### What the approval actually pins has not been verified

The harness intercepts the call and waits. **Whether the arguments the human
approved are guaranteed to be the arguments that get sent - with no further model
turn in between - is not documented anywhere we have found**, and it matters for
anything where the approved content is the deliverable.

Do not assume either answer. If a customer's requirement depends on it - "the
person checks the listing, and that exact listing is what goes up" - raise it
with the platform team, write the answer into a decision record, and cite it.
This paragraph should be replaced by that citation.

### A scheduled run cannot use a consenting tool

Nobody is there at 03:00. `requestConsent: true` parks the run until it times
out, so a Trigger's toolset needs `requestConsent: false` - which means the
scheduled path and the write path **cannot share a Toolset**.

Keep them as separate CRs even when they call the same API, and say why in the
header. Someone will eventually try to merge them.

### `Toolset.spec.instruction` does not exist

It was removed from the CRD. Usage guidance lives in the Workflow's
`entries[].tooling.description`.

Adding it back is a trap worth knowing precisely: the CRD **silently prunes**
undeclared fields, so `kubectl apply --dry-run=server` reports success while the
field is discarded, and then helm's server-side apply fails **in CD** with
`field not declared in schema`. That broke a release once, after passing 25 of 25
dry-runs.

### One credential or several is a requirement, not a convention

`asgard_resource_api_key` is the conventional name the skeleton uses for a
platform resource credential, and `asgard-cli add` points every CR of that kind
at it. Whether they in fact share one is a question about rotation scope and
blast radius - a requirement, not something a template settles. A token for the
**external** service is its own key either way, and only ever a `secretKeyRef`.

**Do not declare a `secretKeyRef` for a key that is not both declared and set.**
Config evaluation fails at call time, not at apply time, so the chart deploys
and the tool breaks the first time somebody uses it. The key has to be declared
under `appSecret:` in `.asgard-pipeline.yaml` **and** given a value with
`asgard-cli pipeline variables set --kind secret` - a value with no declaration
is stored and never injected, which `variables list` reports as `ORPHAN` and the
run reports as `vars/orphan`. Leave the variables list empty until both are done.

## A mock is a fallback, not the default

**Ask the customer whether there is a test environment before you design a mock, and record the answer in `docs/open-questions.md` if it does not come back in the meeting.** If there is, write into it -
the whole path is then genuinely proved, including the fields, the validation
rules and the status codes, and none of that is proved by a mock. Reaching for
a mock before asking loses the strongest version of the first delivery, and
"shall we really create the ticket, or just draft it" is a question that assumes
they have only production.

Two situations where it is right, and they are different:

    no test environment, and they will not
    let us write to production                a mock. Their call, not ours
    the endpoint does not exist yet -
    not provided, or later in the plan        a mock

A mock that echoes the request back is the right thing to ship in either case. The whole chain gets exercised, and the drafted call lands
in the invocation record for review.

**Returning success is deliberate**: a failure would stop a cursor and the path
would never run end to end.

That makes **disclosure the safety property**. Both the tool's description and
the agent's prompt must require the summary to say plainly that nothing was
actually sent. A log that reads as though customers were emailed, or listings
were published, is the real damage a mock can do - and it is discovered late, by
someone who trusted it.

## Verify

```bash
asgard-cli check
asgard-cli verify <project>
```

The xref check resolves the `(workflow, entry)` pair. **Nothing checks that
`requestConsent` is set correctly** - a write tool with it missing or `false`
lints clean, deploys clean, and then acts without asking.

So check it by reading, every time, and make it part of review:

```bash
asgard-cli render <release> | grep -A2 'requestConsent'
```

Every `true` should be a tool that changes something; every `false` should be a
read or a scheduled path. Anything else is the gate being wrong in one of the two
directions that matter.
