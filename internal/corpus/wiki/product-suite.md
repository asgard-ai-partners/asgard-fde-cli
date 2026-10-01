---
group: Products and scope
description: what each product is for, and who uses it
---
# Six products

Asgard is six products. Decide which one a request belongs to first, because
that decides what gets built.

| product | also called | in one line | who uses it |
|---|---|---|---|
| Odin | Asgard Studio | the build platform: design workflows, produce Bots, Automation Tools and knowledge bases | whoever builds it, usually us |
| Mimir | Data Insight | explore data by conversation, build charts and dashboards | the people who need the numbers |
| Sindri | Agents Hub | where agents are published and run | staff, every day |
| Management Console | - | Workspace Management, and each product's accounts and roles | the customer's IT or admins |
| Heimdall | Media & PR AI | PR monitoring and AI-assisted review | PR and marketing |
| Fehu | - | AI usage billing | whoever pays |

## Which ones an FDE actually touches

Odin and Sindri are the main ground. Odin is where agents, data connections,
knowledge bases and automations are built; what is built there is published to
Sindri for users. Sindri offers no creating or editing.

Management Console is a prerequisite. Workspaces, accounts and permissions are
set there, and so is the per-resource grant a semantic model or agent needs after
it is built. If that step is skipped the customer reports seeing nothing, and
the chart is not the cause.

Mimir is often what the customer actually wants. When someone says "I want an AI
that answers stock questions", ask what they do with the answer: glancing at it
each morning is a Dashboard, looking one thing up is an agent.

Heimdall and Fehu rarely appear during onboarding unless the customer came for
them.

## What Odin holds

This layer maps to the CRs in a chart; see the name mapping in
[`agents.md`](../wiki/agents.md).

| Odin feature | what it does |
|---|---|
| Agent Hub | creating and configuring Managed Agents and Flow Agents |
| Applications | outward integration: customised integrations, SDK and API |
| Automation | Trigger (schedules) and API |
| Data Insight | Semantic Model - the read surface over a database |
| Drive | where document knowledge lives |
| Knowledge Base | the older knowledge mechanism - live, but a Drive is preferred for new work |
| MCP Servers | external tools |
| Plugins | capability bundles |
| Skillsets | skill sets |
| Settings | Completion Model, Embedding Model, Connection, Data Source |

## Model providers

Customers can bring their own: Azure OpenAI, Anthropic Claude, Google Gemini,
Mistral, OpenAI. Models are split by purpose into Completion Model and Embedding
Model, and the two lists differ: an Embedding Model can also be Voyage AI. Meta
LLaMA appears in an older product page but is not a class the platform accepts -
a customer asking for it has no custom-model route. The classes are in
[`settings.md`](../wiki/settings.md).

The platform also offers built-in options under semantic aliases - complex,
balanced, fast, vision - rather than specific model names. That is usually the
right default: customers rarely have a preference, and a hardcoded model name becomes
something to come back and fix when the model is retired.

## Outward integration

An Asgard application can be reached through an SDK, an API, Discord, LINE, Slack
or Telegram. The SDK and the API are the `generic` BotProvider class; the other
four are one class each. For Taiwanese customers LINE is usually the first one
asked about.

## Corresponding extracts

This page is a scoping judgement rather than one shape, so it points at no single
extract. Once the product is settled, follow the extract named on that product's
page.

## Sources

- [Product suite](https://docs.asgard-ai.com/docs/product-suite)
  - asgard-docs `6261fdff`
- [Core concepts](https://docs.asgard-ai.com/docs/core-concepts-ecosystem) and
  [deployment architecture](https://docs.asgard-ai.com/docs/deployment-architecture)
  - asgard-docs `21c920f6`. The first is brand framing - Mimir decides, Sindri
  executes, Odin enables; the second says Asgard is delivered as cloud SaaS only,
  with no on-premises option
- Odin's feature list: the filenames under
  `docs/product-suite/odin/features/` at asgard-docs `21c920f6`
- Model providers and integration outlets also appear in asgard-docs
  `docs/overview/asgard-features.md` - cited as a file: it is `draft: true`,
  so there is no page to link to. It also links to paths that no longer
  exist, and its provider list names Meta LLaMA, which the CRD does not accept

**Checked:** against asgard-kube `3da0365` - `CompletionModelClass`,
`EmbeddingModelClass` and `BotProviderClass` in
asgard-kube `pkg/apis/asgard/v1alpha1/types.go` - for the provider list and the
integration outlets, and against asgard-docs `21c920f6` for the rest.

**Unchecked:** what each product's screens offer is taken from the product
documentation, and nobody here has opened them in a Console account.
