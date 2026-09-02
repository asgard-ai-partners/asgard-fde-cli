# Models, data sources and connections

Four things under Odin's Settings. All of them are prerequisites for something
else.

| setting | used by | CR |
|---|---|---|
| Completion Model | agents, workflows, Semantic Models | `CompletionModel` |
| Embedding Model | semantic search and retrieval | `EmbeddingModel` |
| Data Source | Semantic Models and agent queries | `DataConnector` |
| Connection | Syncers and Loaders needing third-party authorisation | `OAuthProvider` + `OAuthCredential` |

## Completion Model

Four built-in tiers, marked Built-in and neither deletable nor editable:

| built-in | for |
|---|---|
| Builtin (Vision) | tasks with image input |
| Builtin (Balanced) | general use |
| Builtin (Complex) | harder reasoning |
| Builtin (Fast) | when latency matters most |

A custom model takes Name, Model Provider, Model Name and API Key.

The built-in tiers are semantic aliases rather than specific model names, which
is usually the right default: customers rarely have a view, and a hardcoded model
name becomes something to come back and fix when the model is retired.

**A builtin tier is a logical model, and a router resolves it.** This is what
the alias buys, and none of it is visible from the platform side:

  - one logical name is backed by **several provider-model pairs**, and the
    selection policy is weighted-random, weighted round-robin or ordered
    fallback
  - **automatic failover**: a 5xx or a timeout from one provider retries the
    next candidate rather than failing the run
  - the managed key lives in the router's own environment

So a builtin tier is not "a model we picked for you" - it is a pool with
failover. **A custom `CompletionModel` gives that up**: one provider, one key,
one point of failure, and an outage at that provider is an outage for the
customer.

**And it is only available on Odin.** Sindri and Mimir use the platform's
designated models and the LLM cannot be swapped there - see
[`fehu.md`](fehu.md). So a custom `CompletionModel` does not make a hub agent or
a dashboard use the customer's key, and "we will use our own model" has a
different answer per product.

That is the trade to state when a customer asks for a specific model. They may
still want it - a compliance requirement, an existing contract, a model they have
tested against - and those are good reasons. "We prefer this one" usually is not.

**In a chart, a custom model is a `CompletionModel` CR** - the built-in tiers are
that CR's `builtin` class rather than the absence of one, which is the reading
that gets this wrong. Three reference deployments declare their own, with the
provider's key as a secretKeyRef into app-secret:

| | |
|---|---|
| `completionModelClass` | `aoai-chat`, `openai-chat`, `gemini`, `anthropic`, `mistral`, `builtin` |
| provider block | exactly one of `aoaiChat`, `openaiChat`, `gemini`, `anthropic`, `mistral`, `builtin` |
| the key | `spec.<provider>.apiKey.valueFrom.secretKeyRef` |

**`completionModelClass` is immutable**, so moving a customer from one provider
to another is a new CR rather than an edit - the same trap as
`BotProvider.botProviderClass`. The exactly-one rule is a CRD validation, so a CR
carrying two provider blocks is refused by the apiserver and passes `helm lint`.

## Embedding Model

Only one built-in, Builtin (Balanced).

The custom form's fields change with the provider. The default, Azure OpenAI
Embedding Model, takes six required fields: Name, Model Provider, Resource Name,
Deployment ID, API Version and API Key.

## Data Source

Nine providers, matching `DataConnectorClass` exactly:

MySQL, SAP Hana, SalesForce, Oracle, Microsoft SQL Server, PostgreSQL, NetSuite,
Trino, Athena.

The form takes Name, Provider, Host, Port, Database, User and Password. Test
Connection can be run before Save. The form marks nothing as required.

In a chart, the non-secret coordinates go in `chart/values-<env>.yaml` and the
password is always a secretKeyRef into `app-secret`.

**An HTTP API does not go here.** Data Source is these nine database providers
and nothing else, and Connection below is OAuth to five named services. A REST
API with a key or a bearer token is configured on the tool that calls it - an
`http-request` step in a Workflow, or an MCP Server's environment variables.
[`setup-path.md`](setup-path.md) has the fork; `asgard-cli usecase external-api`
has the shape.

## Connection

Manages authorisation to third-party apps and services over OAuth, as distinct
from a Data Source's host-and-password. The form has only Name and Type, and the
button says Authorize.

The Type list is grouped by purpose:

| group | options |
|---|---|
| syncer | Dropbox, Google Drive, OneDrive |
| For Loader | Google Drive, OneDrive |
| For Trigger | Google Drive, Google Sheets, OneDrive, OneDrive Workbook (Excel) |

The same service appears in more than one group, so the services that can
actually be authorised are Dropbox, Google Drive, OneDrive, Google Sheets and
OneDrive Workbook - five, not eleven.

The "For Trigger" group is stale: the trigger classes it corresponds to were
removed and only cron is left. See [`automation.md`](automation.md).

## Before writing the chart

`asgard-cli usecase semantic-layer` has the DataConnector fields - coordinates in
`values-<env>.yaml`, password always a secretKeyRef. Connection has no extract of
its own, because OAuth authorisation happens in the UI rather than being declared
in a chart.

## Sources

- [Completion Model](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/completion-model),
  [Embedding Model](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/embedding-model),
  [Data Source](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/data-source),
  [Connection](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/connection)
  - asgard-docs `f00e0ee`
- The CR mapping and the provider list: checked 2026-09-02 against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` -
  `DataConnectorClass`, `CompletionModelClass`, `EmbeddingModelClass`

- The router's behaviour behind a builtin alias - logical models, the three
  selection policies, failover on 5xx or timeout, and the managed key in its own
  environment: `asgard-router`'s README, read 2026-09-02

**Checked:** 2026-09-02 against asgard-kube `15ded0f`
(`completionModelClass` enum, the immutability rule and the ExactlyOneOf
validation) and against three deployments that declare their own model.

**Unchecked:** the provider list was held against the CRD; the UI form fields come
from the product documentation only.
