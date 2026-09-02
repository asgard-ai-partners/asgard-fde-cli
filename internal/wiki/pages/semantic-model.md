# Semantic Model

Connects a database, has AI understand its structure, and then answers in natural
language, producing charts and reports. Built in Odin; the finished model appears
in Data Insight (Mimir).

CRs: `SemanticLayer`, plus the `DataConnector` it points at.

## Building one

Three steps: Basic Information, Table Setting, Modeling.

### Basic Information

| field | required | |
|---|---|---|
| Name | yes | |
| Completion Model | yes | **must support at least 60,000 Max Output Tokens** |
| Effort | | reasoning depth, applied to every query; Default unless set |
| Language | yes | the language the AI answers in |
| Timezone | yes | |
| Data Source | yes | PostgreSQL, MySQL, MS SQL, Oracle, Trino, Athena |
| MCP Servers | | |
| Plugins | | |

The 60,000-token floor is a real limit: the wrong model fails during modelling.

Six data sources are listed here, while the Data Source settings page supports
nine - it also has SAP Hana, SalesForce and NetSuite. Whether a Semantic Model
can use those three is not documented.

### Table Setting

The left pane lists every table under that Data Source **across all schemas**,
searchable and selectable, **up to 100**. The right pane previews the columns and
rows of whichever table is selected.

### Modeling

The AI assistant connects to the database and works through each table: reading
columns and types, sampling rows, checking which column combinations are unique
enough to be a primary key, then writing the Name, Description, Primary Key and
Columns.

On entry the panel offers three quick actions: Build all the tables, Start with
one table, Find the relationships. The left side holds Table semantics,
Relationships, Summary, Sample Questions, Rule and SQL Scripts, each with its own
Discuss button.

**Ten tables took about three minutes** in practice.

Before adding relationships the assistant asks about the things the data cannot
settle - whether a column that currently holds one value will hold others later,
what two similar columns each mean - and continues once answered.

> **Save is manual, bottom right.** Switching or closing the tab without saving
> discards both the table selection and the built schema, and reopening starts
> from an empty Table Setting.

## The Mimir side

Mimir reads and presents a semantic model and **does not edit it**. Building and
publishing happen in Odin.

Mimir's Data Model button expands the current model's structure:

- **Summary** - the business area it covers
- **Model Source** - which semantic model this structure came from
- **Table / Columns** - names, types and descriptions
- **Relationships** - the columns joining this table to others
- **Measures** - predefined measures with their aggregation and source column. A
  question can name a measure directly instead of describing the arithmetic
- **Preview Data** - ten real rows

## Semantic Model or fixed query tools

For an anonymous public audience the answer is usually fixed zero-parameter query
tools rather than a Semantic Model. The reasoning, and what `allowedCubes` does,
is in `asgard-cli usecase fixed-query-tools`; the modelling itself is in
`semantic-layer`, and opening a write path is in `write-path`.

## Sources

- [Semantic Model](https://docs.asgard-ai.com/docs/product-suite/odin/features/data-insight-semantic-model)
  - asgard-docs `f00e0ee`
- [Data Model](https://docs.asgard-ai.com/docs/product-suite/mimir/features/data-model)
  - asgard-docs `f00e0ee`
- `completionModelName` being required: checked 2026-09-02 against 11 real
  `SemanticLayer` CRs across three deployments
- The six-versus-nine data source difference: compared against the
  [Data Source](https://docs.asgard-ai.com/docs/product-suite/odin/features/settings/data-source)
  page. The documentation gives no reason, so this is **unconfirmed**

**Unchecked:** the modelling guidance and the UI steps come from the product
documentation only. `completionModelName` was held against 11 real CRs.
