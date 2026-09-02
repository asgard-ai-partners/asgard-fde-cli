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

Two things this changes in an interview.

**It is a question with a known answer, so ask it in the first meeting.** "Is
this system reachable from outside your network, and who can add four addresses
to the allowlist?" hands their network admin exactly what they need. The failure
this avoids is discovering in week three that the credential works, the query is
right, and nothing can connect.

**It has an owner and a lead time on their side.** An allowlist change is a
firewall change, and in most companies that is a ticket, an approval and a
window - not something the person in the meeting can do that afternoon. That is
why it is one of the few things worth tracking as an open question rather than
handing back as a note: it blocks the first delivery and we cannot do it
ourselves. Get the name of whoever approves it.

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
