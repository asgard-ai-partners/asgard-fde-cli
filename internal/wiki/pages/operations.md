# Connectivity, model capability, vocabulary

Things filed under help-community that come up in practice.

## Reaching a system inside the customer's network

**Asgard is a hosted cloud service.** It runs in Asgard's own cloud, not in the
customer's data centre and not on their network, and there is no deployment that
puts it inside. Everything follows from that.

**And the agent runs in a sandbox that the platform starts**, in that cloud. So
there is no fixed machine of ours to put on their network, nothing to install
behind their firewall, and no endpoint of theirs we can reach out from. The
traffic leaves Asgard's cloud, and it leaves from the addresses below.

So the ask has exactly one shape, and it is not a menu:

    they add Asgard's outbound addresses to their allowlist

**Do not offer a VPN, a bastion or a jump host as alternatives.** They are the
shapes for connecting two networks, and this is not that: it is a hosted service
calling in from fixed addresses. Presenting three options invites their network
team to pick the one that suits their habits, and then a week is spent
discovering it does not apply. If their policy requires a VPN, that is a
conversation for their side about how the allowlist is implemented - it does not
change what we need from them.

The work is theirs, not ours. We supply the addresses; they own the change, the
approval and the schedule.

Traffic from the platform to a customer's internal database or service leaves
from these four fixed addresses:

```
35.79.216.190
52.196.233.171
54.178.218.69
57.181.108.84
```

All four have to be allowlisted, not one - which of them a given request leaves
from is not something to rely on.

### How these are handed over, which is not "in the deck"

**Do not copy them into a customer repository, a proposal, or an email that will
be forwarded.** Give them directly to the person making the firewall change,
once, and read them from here when you do.

The reason is this material's own core argument, applied to itself. The
screenshots page refuses to carry images because *a copy in one engagement goes
stale where nobody is looking* - and an address is the same kind of thing with a
worse failure. **A stale screenshot is embarrassing; a stale allowlist is the
customer's connection dropping, and they will come back to us about it.** Every
copy in a repo, a slide or a mail thread is a copy nobody will update.

It also resolves a contradiction that was sitting in plain sight: the
`proposal-deck` skill forbids **coordinates** on a customer's screen - hostnames,
connection strings, account names, not even in a screenshot's corner - and that
rule is about the customer's infrastructure. Ours is the same class of thing, and
this page was encouraging the opposite. Two documents, opposite instincts, and
nothing said they were about the same subject.

**In the meeting, ask whether the change can be made.** Then send the addresses
to whoever will make it, afterwards and directly.

**Find out who will make it, for the follow-up list** - not for a slide. That
phrasing is deliberate: an earlier version of this page said "get the name of
whoever approves it" with the boundary in the paragraph above, and an FDE read
both and copied only the imperative onto a customer deck. **A caveat beside an
imperative reads as elaboration, not as a limit.** The destination has to be
inside the instruction.

**Nothing here records how these change.** They read as constants and there is no
documented channel for a revision, which is itself worth knowing before treating
a copy of them as durable.

Two things this changes in an interview.

**It is a question with a known answer, so ask it in the first meeting.** "Can
four addresses be added to that system's firewall allowlist?" - the question,
not the addresses and not the org chart. It avoids discovering in week three
that the credential works, the query is right, and nothing can connect.

**Track the outcome, not the person.** This is worth an open question rather
than a note, because it blocks the first delivery and we cannot do it
ourselves - but what is tracked is *whether the allowlist can be changed, and
then whether it has been*. Not who signs it, not how many approvals, not how
long their process takes.

    ours     can this be changed, and is it done yet
    theirs   who signs, which queue, how long

**Asking who approves it fails this material's own filter**, and it has been
asked on a slide: `asgard-cli next --stage requirements` filter 0 tests whether
an answer changes what we build, and an approver's name does not. Worse, filter
0 names this exact case - turning an operational precondition into a design
question - and asking it in front of a customer reads as managing their internal
process.

It is fine to ask when they expect it, because a date changes our plan. It is
not fine to ask who, because a name does not.

**This page has now contradicted the interview stage twice** - once by offering a
VPN or a jump host as alternatives, once here. Its reader is usually somebody
preparing for a meeting, and it was written as though for somebody doing an
integration. **Any line here that says "ask this in the meeting" should be run
through filter 0 before it is followed.**

The order the rest of the setup follows once the path is open is
[`setup-path.md`](setup-path.md).

## Checking what a model supports

Confirm a model's input and output types before choosing it, or a workflow breaks
at run time. GPT-4o, for instance, takes and produces text, takes images but does
not produce them, and does not handle audio at all.

| provider | list |
|---|---|
| OpenAI | platform.openai.com/docs/models |
| Azure AI | ai.azure.com/catalog |
| Anthropic | docs.anthropic.com/en/docs/about-claude/models/overview |
| Mistral | docs.mistral.ai/getting-started/models/models_overview |
| Gemini | ai.google.dev/gemini-api/docs/models |
| Voyage | docs.voyageai.com/docs/pricing |

Semantic Model adds a hard floor of its own: the Completion Model must support at
least 60,000 Max Output Tokens.

## Common LLM Completion failures

| symptom | cause | what to do |
|---|---|---|
| the step fails | the prompt or history exceeds the model's context window | check the input length, chat history and embedded data especially; cap output with `MaxToken` |
| no response | the provider account has no billing enabled, or the quota is spent | check the payment method and remaining quota |
| the request is refused | the API key is wrong, expired, or has stray whitespace | check the key |

To reproduce a context-window overflow deliberately, set `MaxToken` to `0`.

## An empty iFrame

Usually the iFrame has not been made public. Check Public iFrame under App-Share
Settings, save, and try again.

## Connecting an OpenAI-compatible model

Services whose API is OpenAI-compatible, such as DeepSeek, need no new provider:

1. Model Provider: **OpenAI Chat Completion**
2. Model Name: **Other**, then the real model name (`deepseek-chat`)
3. Endpoint: the service's address (`https://api.deepseek.com`)

## Environment

A Project can hold several Environments. The system creates Main by default and a
user can add others - Main as production and another for development, say - and
merge a finished one back into the main environment.

This is not the same thing as a chart's `dev` / `prod`: those map to namespaces
and values files, while a platform Environment is a division inside a Project.
**How the two correspond is not documented, and is unconfirmed.**

## Vocabulary

| term | meaning |
|---|---|
| Workspace | the **smallest unit a subscription is billed against** |
| Project | a project under a Workspace; its resources - knowledge bases, settings, apps - are shared within it |
| Collection | a set of workflows, a workflow set |
| Workflow | processors connected into a flow, with a start and an end |
| Processor | the smallest processing node |

That a Workspace is the billing unit is worth keeping in mind: when a customer
asks about cost structure, this is where the effect of how things are divided
begins.

## Corresponding extracts

Connectivity and vocabulary produce no CRs, so there is no extract for them.

## Sources

- [Outbound IPs](https://docs.asgard-ai.com/docs/help-community/other/vpn-white-list-ip),
  [checking model support](https://docs.asgard-ai.com/docs/help-community/other/check-model-support),
  [an empty iFrame](https://docs.asgard-ai.com/docs/help-community/faq/iframe-display-blank),
  [LLM Completion troubleshooting](https://docs.asgard-ai.com/docs/help-community/other/troubleshooting-llm-completion),
  [DeepSeek](https://docs.asgard-ai.com/docs/help-community/faq/deepseek),
  [channel log](https://docs.asgard-ai.com/docs/help-community/faq/how-to-show-channel-log)
  - asgard-docs `f00e0ee`
- [Environment](https://docs.asgard-ai.com/docs/overview/asgard-environment)
  - asgard-docs `f00e0ee`, marked `draft`
- [Glossary](https://docs.asgard-ai.com/docs/help-community/glossary)
  - asgard-docs `f00e0ee`. Its Processor entry lists an older set of nodes and
  does not match the current `ProcessorType`; take [`workflow.md`](workflow.md)
  as the current one

**Unchecked:** the IPs and the model lists come from the product documentation
only. That Asgard is hosted and therefore always needs the customer to open the
path inward is the FDE team's account of how every engagement so far has gone,
recorded 2026-09-02; no document states it, and it has not been held against a
deployment that failed for the opposite reason. How a platform Environment corresponds to a chart's dev/prod is
undocumented and unconfirmed.
