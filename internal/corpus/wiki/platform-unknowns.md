# What the platform's documentation does not answer

Questions no source settles - not the CRDs, not the product documentation. Each
one is here because an engagement hit it and had to proceed without an answer.

**Ask the platform team before the meeting, not the customer and not during one.** Do not assume. Then
write the answer into the customer repo's `docs/decisions/`, cite it wherever the
design depends on it, and tell whoever maintains `asgard-cli` so the next
engagement starts with the answer rather than rediscovering it.

**Checked:** 2026-09-02 against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` and asgard-docs `f00e0ee` - each row was re-checked to confirm neither source answers it.

**Unchecked:** whether the platform team has since answered any of these outside those two repositories.

## The list

| # | Question | When it matters |
|---|---|---|
| P1 | **Can one user be allowed to see or operate only some of the resources, while another sees others?** The entry point authenticates the caller, but nothing found so far scopes what a given user may then reach. Asked because a customer's own test plan required limiting which equipment different users could query or operate, and there was no answer. | Any customer with more than one team. A security review asks it early |
| P2 | **Is there a record of what an agent did - who asked, what was called, what came back - and can it be shown to an auditor?** CRs carry a label the platform echoes into audit events, so something is recorded, but not what or for how long. Asked because a customer's test plan required keeping a record of equipment operations. | Anything with a side effect, and every regulated customer |
| P3 | **Between a human approving a call and that call going out, can the content change?** Asked because an engagement had to write "the approved listing is the one that gets published" as an assumption with nothing to cite. | Any case where the approved content **is** the deliverable - a listing, a message, a document |
| ~~P4~~ | ~~Can the platform reach a system over something other than HTTP?~~ **Answered: yes.** The agent gets a sandbox, so any protocol with a client - SNMP, SSH, a vendor CLI - is reachable by running that client. Reads may happen there; **writes still go through a Toolset with `requestConsent`**. See `../usecase/external-api.md`. | - |
| P5 | **Before an agent can operate a system that only has a web console, someone has to write down what that console contains** - every page, what each page does, and what is behind the buttons the menu does not show. One deployment has such a document covering 88 pages. **Is producing it tooling-assisted, or does a person work through the system by hand?** And when the vendor restyles their UI, is it redone? | Any web-only system. It is the difference between a day and a week, and it decides the maintenance cost |
| ~~P6~~ | ~~**Are the per-request limits adjustable per workspace?**~~ **Partly answered 2026-09-02.** They can be raised: the quota page says to contact sales or write to **service@asgard-ai.com**. **What is still unanswered is by how much, and whether it is tiered.** So the sentence to say in a meeting is "that is the default, and it can be raised" - not "that is the limit". The difference matters to a proposal | any multi-system agent |
| P8 | **What does the approval gate look like on an anonymous channel, and who presses it?** Every screenshot of it is Sindri's dialog, which is where authenticated staff work. `../usecase/write-path.md` says a prompt-level "shall I go ahead?" is not a real gate, and does not say what a real one is on LINE. It matters because on an anonymous channel the person approving a write is usually **the visitor in the conversation**, not staff - so the question is whether `requestConsent` renders there at all, and as what | any public-channel capability with a write in it, which is most customer-service scenarios |
| P7 | **What is a "step"?** The ceiling is 30 per request and the platform enforces it - the error is `Max execution steps 30 reached` - and **no source defines the unit**. Not the quota page, not the CRDs, not the processor documentation. A tool call? A model turn? A processor execution? Without it you cannot tell whether a design will exceed it, and you cannot answer the customer who asks. **Until it is answered, do not put "30 steps" in front of a customer** - say the three-minute ceiling, which they can check | any design that consults several systems in one turn, and any proposal that quotes the limit |
| P9 | **Does the message endpoint carry a `/generic/` path segment or not?** The API reference documents `{base_url}/generic/ns/{namespace}/bot-provider/{name}/message/sse`; the SDK overview and a production tenant's chart README both give the same URL without `/generic`, and the SDK appends `/message/sse` to the endpoint you pass it without adding anything. One of the two 404s. Asked because the material carried the reference form as though it were the one to hand out | the moment a front-end team is given an integration URL - which is early, and by someone who will not find out for a week |
| P10 | **Which config keys does a processor actually accept?** `ProcessorDefinitions` in asgard-core is the only machine-readable list and it is demonstrably incomplete: `await` is documented on `stream-llm-completion-message` with real semantics, is set in five separate production deployments, and appears in neither that list nor the CRD. `temperature` is the same. So there is no source that says what a processor takes, and a chart cannot be checked against one. Asked because a gate built on treating the definitions as complete called five of five correct charts wrong | writing a Workflow by hand, and any attempt to validate one before it reaches a cluster |
| P11 | **What else is in scope in an expression that nobody has written down?** `prevToolCalls` carries the agent's tool calls with their arguments and results, is what makes post-processing possible, and appears in no documentation page - it was found by reading a chart. If one variable is undocumented, the list of six on the expression page is a floor rather than the set. Asked because the material taught that list as complete | any design that needs to react to what an agent did, which is most write paths and every confirmation message |
| P12 | **Are `ImageGenerationModel`, `TranscriptionModel` and `SourceSetEditorServer` meant to be reached for?** All three are CRDs in the contract at `15ded0f`, with full schemas and provider blocks, and they appear in **no product documentation page and in no material here**. An engagement whose customer needs image generation or transcription would find the CRD and nothing else - no shape, no traps, no statement either way. Which of the two it is matters: the platform may be ahead of its documentation, or these may not be for an engagement to use. Cheap to answer and nobody has asked | any request for image generation, transcription, or an editable source set |

## Why this is not in the customer's repository

It used to be, in `docs/open-questions.md`, and that was wrong for the reason
this whole body of material is embedded rather than copied: a copy in one
engagement goes stale where nobody is looking, while a stale one here is fixed
for every engagement in a single release. A repo that filed its first question in
March and one that filed its first in September would have disagreed about this
list forever, and the older one would have been the one still showing an answered
question as open.

The customer repo's `docs/open-questions.md` holds that engagement's own
questions and nothing else.

## Adding a row

Every row here came from an engagement actually hitting it - a customer's test
plan asking for something with no answer, or a design that had to proceed on an
assumption with nothing to cite. Add one the same way.

A question that merely seems like a good thing to know does not belong here. It
dilutes the list, and the list only works because every row is one somebody has
already paid for.

When one gets answered, do not delete it: strike it through and write the answer
in, the way P4 reads now. The fact that it was once unknown explains the shape of
designs that were made without it.

## Sources

- Each row names the engagement or the customer requirement that produced it
- P4's answer: `../usecase/external-api.md`
- P6 and P7's numbers, and that quotas are raised by contacting sales:
  [Quota and limits](https://docs.asgard-ai.com/docs/help-community/quota-limits)
  - asgard-docs `f00e0ee`, read in full 2026-09-02. An earlier reading took four
  of its eight numbers and its closing section, which is where the answer to P6
  was
