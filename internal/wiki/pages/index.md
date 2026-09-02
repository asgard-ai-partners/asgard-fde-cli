# Index

    asgard-cli wiki <page>

## Products and scope

| page | covers |
|---|---|
| [`product-suite`](product-suite.md) | the six products, what each is for and who uses it |
| [`console`](console.md) | the two permission layers, five inconsistent pages, Workspace settings |
| [`sindri`](sindri.md) | Project, delegation, the Sandbox and its files, the governance gate |
| [`mimir`](mimir.md) | Thread, View, Dashboard, Knowledge |
| [`fehu`](fehu.md) | billing and usage, how cost is broken down |

## While building

| page | covers |
|---|---|
| [`setup-path`](setup-path.md) | the order: where a credential goes, what to build from it, why Sindri needs no import |
| [`agents`](agents.md) | Flow Agent against Managed Agent, how to choose, UI-to-CR names |
| [`knowledge`](knowledge.md) | Drive, Context Index, and how Knowledge Base differs |
| [`semantic-model`](semantic-model.md) | the modelling flow, its limits, the Mimir side |
| [`tools`](tools.md) | MCP Server, Skillset and Plugin; hook events |
| [`automation`](automation.md) | Trigger and API, and why only cron is left |
| [`workflow`](workflow.md) | the 13 processors; Expression is JavaScript, Template is Handlebars |
| [`settings`](settings.md) | Completion and Embedding Model, Data Source, Connection |
| [`integration`](integration.md) | chat platforms, the two Applications pages, the architecture |
| [`api`](api.md) | the endpoint and its actions, the SSE sequence, four patterns, the SDK |
| [`platform-unknowns`](platform-unknowns.md) | what no source answers, and who to ask |

## In practice

| page | covers |
|---|---|
| [`operations`](operations.md) | Asgard's outbound IPs, checking model capability, vocabulary |
| [`case-studies`](case-studies.md) | the retail stockout from three angles, plus a Flow Agent help desk |
| [`screenshots`](screenshots.md) | which picture answers which question, and the URL to fetch it from |

## Known gaps between the documentation and the CRD

Each is written on the page it affects:

| gap | page |
|---|---|
| the UI's Flow Agent is three CRs; `agentClass` has one value | `agents` |
| the Drive Syncer UI offers five sources, the CRD supports ten | `knowledge` |
| Knowledge Base and Drive both exist and which one new work should use | `knowledge` |
| Connection's "For Trigger" group names removed trigger classes | `automation`, `settings` |
| Semantic Model lists six data sources, the settings page nine (unexplained) | `semantic-model` |
| the four `integration-with-asgard/` pages are `draft` and name an older UI | `integration` |
| the glossary's Processor list does not match the current `ProcessorType` | `operations` |
| `overview/asgard-features` is `draft` and links to removed paths | `product-suite` |

## Coverage

Measured 2026-09-02: asgard-docs holds 162 files, of which 130 are cited by some
page's source block - **100% of what is in scope**. The 32 excluded below are the
denominator's difference.

## Deliberately not covered

| excluded | count | why |
|---|---|---|
| `developer-reference/asgard-builtin/` | 18 | Expression variables, function lists and message templates. Lookup material: copying it here only produces a copy that goes stale. Read the source when needed |
| `help-community/release-notes/` | 10 | historical, and does not describe the present |
| `superpowers/` | 5 | the documentation site's own redesign plans, not an Asgard feature |

## Outside this wiki

| what you want | where |
|---|---|
| how a deployment shape is assembled, field by field | `asgard-cli usecase` |
| whether a CR field is legal or required | [asgard-kube's `crd/*.yaml`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) |
| what this customer's systems look like | that customer repo's `docs/spec/` |
| the original product documentation | [asgard-docs](https://github.com/asgard-ai-platform/asgard-docs), rendered at https://docs.asgard-ai.com |
