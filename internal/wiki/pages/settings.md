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

**Unchecked:** the provider list was held against the CRD; the UI form fields come
from the product documentation only.
