# Connectivity, model capability, vocabulary

Things filed under help-community that come up in practice.

## Asgard's outbound IPs

Traffic from the platform to a customer's internal database or service leaves
from these four fixed addresses:

```
35.79.216.190
52.196.233.171
54.178.218.69
57.181.108.84
```

A customer whose firewall restricts source IPs needs all four allowlisted.

This is the concrete answer to "can it reach us from the cluster" during an
interview. It is exactly what the customer's network admin wants, and it can be
handed over in the first meeting rather than after a connection fails.

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
only. How a platform Environment corresponds to a chart's dev/prod is
undocumented and unconfirmed.
