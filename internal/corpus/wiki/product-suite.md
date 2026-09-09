# Six products

Asgard is not one product but six. Deciding which one a request lands in shapes
everything after it.

| product | also called | in one line | who uses it |
|---|---|---|---|
| **Odin** | Asgard Studio | the build platform: design workflows, produce Bots, Automation Tools and knowledge bases | whoever builds it, usually us |
| **Mimir** | Data Insight | explore data by conversation, build charts and dashboards | the people who need the numbers |
| **Sindri** | Agents Hub | where agents are published and run | staff, every day |
| **Management Console** | - | Workspace settings, and each product's accounts and roles | the customer's IT or admins |
| **Heimdall** | Media & PR AI | PR monitoring and AI-assisted review | PR and marketing |
| **Fehu** | - | AI usage billing | whoever pays |

## Which ones an FDE actually touches

Odin and Sindri are the main ground. Odin is where agents, data connections,
knowledge bases and automations are built; what is built there is published to
Sindri for users. Sindri offers no creating or editing.

Management Console is a prerequisite. Workspaces, accounts and permissions are
set there, and so is the per-resource grant a semantic model or agent needs after
it is built. Skip that step and the customer reports seeing nothing - which is
not a chart problem.

Mimir is often what the customer actually wants. When someone says "I want an AI
that answers stock questions", ask what they do with the answer: glancing at it
each morning is a Dashboard, looking one thing up is an agent.

Heimdall and Fehu rarely appear during onboarding unless the customer came for
them.

## What Odin holds

This layer maps to the CRs in a chart; see the name mapping in
[`agents.md`](agents.md).

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
Meta LLaMA, Mistral, OpenAI. Models are split by purpose into Completion Model
and Embedding Model.

The platform also offers built-in options under semantic aliases - complex,
balanced, fast, vision - rather than specific model names. **That is usually the
right default**: customers rarely have a view, and a hardcoded model name becomes
something to come back and fix when the model is retired.

## Outward integration

An Asgard application can be reached through an SDK, an API, Discord, LINE, Slack
or Telegram. For Taiwanese customers LINE is usually the first one asked about.

## Corresponding extracts

This page is a scoping judgement rather than one shape, so it points at no single
extract. Once the product is settled, follow the extract named on that product's
page.

## Sources

- [Product suite](https://docs.asgard-ai.com/docs/product-suite)
  - asgard-docs `f00e0ee`
- [Core concepts](https://docs.asgard-ai.com/docs/core-concepts-ecosystem) and
  [deployment architecture](https://docs.asgard-ai.com/docs/deployment-architecture)
  - asgard-docs `f00e0ee`. The first is brand framing - Mimir decides, Sindri
  executes, Odin enables; the second says Asgard is delivered as cloud SaaS only,
  with no on-premises option
- Odin's feature list: the filenames under
  `docs/product-suite/odin/features/` at asgard-docs `f00e0ee`
- Model providers and integration outlets also appear in
  [Asgard features](https://docs.asgard-ai.com/docs/overview/asgard-features) - **`draft: true`, 404s on the published site**,
  which is marked `draft` and links to paths removed in the 2026-08-31 rebuild -
  only the two lists corroborated elsewhere were taken from it

**Unchecked:** everything here comes from the product documentation; none of it
was held against a deployment. Which product suits which need is a judgement.
