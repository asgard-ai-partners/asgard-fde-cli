---
group: While building
description: what no source answers, and who to ask
---
# What the platform's documentation does not answer

Questions no source settles - not the CRDs, not the product documentation. Each
one is here because an engagement hit it and had to proceed without an answer.

**Ask the platform team before the meeting, not the customer and not during one.** Do not assume. Then
write the answer into the customer repo's `docs/decisions/`, cite it wherever the
design depends on it, and tell whoever maintains `asgard-cli` so the next
engagement starts with the answer rather than rediscovering it.

**Checked:** 2026-09-02, re-read 2026-09-11 against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `cbd8d70` and asgard-docs `f00e0ee` - each row was re-checked to confirm neither source answers it. P15 was added 2026-09-15 and held against asgard-kube `cbd8d70` `pkg/apis/` for the label declarations, asgard-core `623ceb50` for what the platform's own generated workflow writes, and asgard-docs `23409b30` - none of the three names a label key.

**Unchecked:** whether the platform team has since answered any of these outside those two repositories.

## The list

| # | Question | When it matters |
|---|---|---|
| P1 | **Can one user be allowed to see or operate only some of the resources, while another sees others?** The entry point authenticates the caller, but nothing found so far scopes what a given user may then reach. Asked because a customer's own test plan required limiting which equipment different users could query or operate, and there was no answer. | Any customer with more than one team. A security review asks it early |
| P2 | **Is there a record of what an agent did - who asked, what was called, what came back - and can it be shown to an auditor?** The label is `asgard-ai.com/product`, which asgard-core echoes into every audit-log event so the Console can filter by product - `../usecase/conventions.md` has it - so something is recorded and the dimension it is filed under is known. What is still unknown is **what each event contains and how long it is kept**, which is the half an auditor asks about. Asked because a customer's test plan required keeping a record of equipment operations. | Anything with a side effect, and every regulated customer |
| P3 | **Between a human approving a call and that call going out, can the content change?** Asked because an engagement had to write "the approved listing is the one that gets published" as an assumption with nothing to cite. | Any case where the approved content **is** the deliverable - a listing, a message, a document |
| ~~P4~~ | ~~Can the platform reach a system over something other than HTTP?~~ **Answered: yes.** The agent gets a sandbox, so any protocol with a client - SNMP, SSH, a vendor CLI - is reachable by running that client. Reads may happen there; **writes still go through a Toolset with `requestConsent`**. See `../usecase/external-api.md`. | - |
| P5 | **Before an agent can operate a system that only has a web console, someone has to write down what that console contains** - every page, what each page does, and what is behind the buttons the menu does not show. One deployment has such a document covering 88 pages. **Is producing it tooling-assisted, or does a person work through the system by hand?** And when the vendor restyles their UI, is it redone? | Any web-only system. It is the difference between a day and a week, and it decides the maintenance cost |
| ~~P6~~ | ~~**Are the per-request limits adjustable per workspace?**~~ **Partly answered 2026-09-02.** They can be raised: the quota page says to contact sales or write to **service@asgard-ai.com**. **What is still unanswered is by how much, and whether it is tiered.** So the sentence to say in a meeting is "that is the default, and it can be raised" - not "that is the limit". The difference matters to a proposal | any multi-system agent |
| P8 | **What does the approval gate look like on an anonymous channel, and who presses it?** Every screenshot of it is Sindri's dialog, which is where authenticated staff work. `../usecase/write-path.md` says a prompt-level "shall I go ahead?" is not a real gate, and does not say what a real one is on LINE. It matters because on an anonymous channel the person approving a write is usually **the visitor in the conversation**, not staff - so the question is whether `requestConsent` renders there at all, and as what | any public-channel capability with a write in it, which is most customer-service scenarios |
| P7 | **What is a "step"?** The ceiling is 30 per request and the platform enforces it - the error is `Max execution steps 30 reached` - and **no source defines the unit**. Not the quota page, not the CRDs, not the processor documentation. A tool call? A model turn? A processor execution? Without it you cannot tell whether a design will exceed it, and you cannot answer the customer who asks. **Until it is answered, do not put "30 steps" in front of a customer** - say the three-minute ceiling, which they can check | any design that consults several systems in one turn, and any proposal that quotes the limit |
| P9 | **Does the message endpoint carry a `/generic/` path segment or not?** The API reference documents `{base_url}/generic/ns/{namespace}/bot-provider/{name}/message/sse`; the SDK overview and a production tenant's chart README both give the same URL without `/generic`, and the SDK appends `/message/sse` to the endpoint you pass it without adding anything. One of the two 404s. Asked because the material carried the reference form as though it were the one to hand out | the moment a front-end team is given an integration URL - which is early, and by someone who will not find out for a week |
| P10 | **ANSWERED.** The editor palette is the source, and it is now read into `../wiki/processors.md` - per processor, which keys are the author's and which the platform sets, from asgard-docs `23409b3` - the Sources entry below has when. `await` and `temperature` are author keys on the streaming processor; they were never missing entries. Some accept **dynamic config** - `../wiki/processors.md` says which - so for those the palette is a starting set and a key an author adds beyond it is legitimate - which is why a chart cannot be validated against a closed key list, and is now a reason rather than a gap |
| P11 | **What else is in scope in an expression that nobody has written down?** `prevToolCalls` carries the agent's tool calls with their arguments and results, is what makes post-processing possible, and appears in no documentation page - it was found by reading a chart. If one variable is undocumented, the list of six on the expression page is a floor rather than the set. Asked because the material taught that list as complete | any design that needs to react to what an agent did, which is most write paths and every confirmation message |
| P14 | **Where does an engagement get the `X-API-KEY` that a front end sends?** The documentation routes it to the project's Integration -> App settings page, and an engagement holding the account could not find that page, or anything else issuing a key, on 2026-09-14. Either it moved, it is named something else, or it is gated on something nobody has identified. **This is not P13.** That one is the platform-minted resource key a CR reads and it is narrowed; this is the key a caller puts in a header, and nothing here answers it. **Until it is answered, do not promise a customer their front end can call the API on day one** - the chart deploys and the first request is refused | any project with its own front end, which is every `generic` BotProvider with `authMode: api-key` |
| P13 | **NARROWED: a chart does not need to read it.** Referencing `preset-agent-hub` / `api_key` directly works - verified against a deployed namespace 2026-09-14, and `../usecase/conventions.md` has the shape - so nothing is blocked any more. What is still unanswered is the original question, for anyone who needs the value itself rather than a CR that reads it: **how does an engagement read the platform-minted resource api key?** Tracked on the platform side as `asgard-core#299`, which is where the answer belongs. The platform creates it per namespace - a Secret `preset-agent-hub`, key `api_key`, generated once and never rotated - and the same value backs `BotProvider.apiKey` and `SourceSet.apiKey` (`../usecase/conventions.md`). What nothing here can do is read it: it is a Kubernetes Secret, no cluster credential is ever issued to a client, and `asgard-cli pipeline manifest` reads back only what the helm release deployed, which that Secret is not. So either a Console page shows it, or somebody with cluster access hands it over, and **no source says which**. Asked because it blocked a first deploy: the chart rendered, every local check was green, and the one value nobody could obtain was the one the CRD requires | every chart with a SourceSet, a Toolset or a BotProvider, which is nearly all of them |
| P15 | **Which keys in a Workflow's `labels` maps does anything read?** Entry, exit, processor, relationship and variable each take one, and asgard-kube declares every one of them as a bare `map[string]string` in `pkg/apis/asgard/v1alpha1/types.go` with no key enumerated anywhere - so **the contract cannot answer this and never will**. The material teaches `display_name` and `description` because that is what the deployments write; the demo generator's chart generator also writes `default` on an entry and `relationship_id` on a relationship, and neither appears in the CRD, in asgard-core or in asgard-docs. The platform's own generated workflow writes no entry, processor or relationship labels at all, which says the engine does not read them and says nothing about what does. The canvas editor is the candidate and it lives in `asgard-ai-platform-web`, which nothing here clones - **the same second-hand source as P10's palette, and a different question**; P10 is answered and closed, so this one is its own row rather than an extension of it. **Until it is answered, do not teach either key and do not generate them** - `default` reads as "the entry to start at", and a workflow that starts in the wrong place is not something any check reports. Asked while deciding, key by key, what the extracts should carry out of a reference chart and what `asgard-cli add` should write: these three had no answer and the extracts proceeded without them | any Workflow meant to be opened in the editor after it is deployed, and every judgement about what to copy out of a chart that already runs **One lead, and it is second-hand too**: a reference deployment's own chart comments record that since 2026-08-31 a single entry is automatically the default, so `entries.labels.default` may be the older shape rather than a live key - which would make this the same generational split `source/SOURCES.md` tracks, and is worth checking before anybody writes the label into a new chart. |
| ~~P12~~ | ~~Are `ImageGenerationModel`, `TranscriptionModel` and `SourceSetEditorServer` meant to be reached for?~~ **Answered 2026-09-14: no. All three are internal.** They are CRDs in the contract with full schemas, and an engagement finding one should read the absence of documentation as the answer rather than as a gap. **So image generation and transcription are not a chart to write** - a customer asking for either needs the platform team, not a CR. `../wiki/settings.md` carries the credential shape for the day that changes, and says it is not a route today | - |

## Why this is not in the customer's repository

These questions are the platform's, not one engagement's, so they do not belong
in a customer's `docs/open-questions.md`: a copy in one engagement goes stale
where nobody is looking, while a stale one here is fixed
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
- P10's answer: the editor palette, read into `../wiki/processors.md` from
  asgard-docs `23409b3` on 2026-09-11. That page carries the table and the
  reasoning; this row records only that the question is closed
- P6 and P7's numbers, and that quotas are raised by contacting sales:
  [Quota and limits](https://docs.asgard-ai.com/docs/help-community/quota-limits)
  - asgard-docs `f00e0ee`, read in full 2026-09-02. An earlier reading took four
  of its eight numbers and its closing section, which is where the answer to P6
  was
