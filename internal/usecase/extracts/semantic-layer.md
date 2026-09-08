# SemanticLayer

The read surface an agent composes SQL against. **The highest-effort artifact in
a chart, and the one that most rewards looking at the real database.**

**Seen in:** deployments with one layer per source system; the largest single
layer runs to hundreds of cubes.

**Checked:** 2026-09-02 against 11 SemanticLayer CRs across three deployments (completionModelName present on every one) and the CRD.

**Unchecked:** the modelling guidance - cube granularity, what belongs in instruction. Checkable only against a database, and not checked.

**Read the platform side first:** `asgard-cli wiki semantic-model` -
what a Semantic Model is, how it is built, and its limits. This page assumes you have.

## When this shape, and when not

Use it when the audience is **internal and authenticated**, and the questions are
open-ended. The agent writes its own SQL over the cubes you expose, which is
what makes it able to answer a question nobody wrote a tool for.

Do **not** use it for a public audience. Mounted without `allowedCubes` it lets
the agent compose arbitrary SQL over every cube in it, and **the exposed surface
grows by itself every time a cube is added** - nobody goes back and narrows it.
Excluding the sensitive tables is not the fix: the shape itself is the risk. Use
fixed query tools there instead.

## Do not guess a schema

Connect to the database and introspect it. Reasoning about a schema is not
verifying it, and the failures are invisible in column names:

- a receipt line can have a **twin line with the quantity negated** (the
  offsetting accounting entry), so a sum without the right filter returns **zero**
- a log table can be **~90% duplicate rows**, so a raw join fans out ~30x and
  silently inflates every count
- a view with a window function can take **1m51s where the same expressions
  against the base table take 2.1s**, because no filter pushes down
- a "safety stock" figure can be defined as coming from one specific location
  only, never the row's own, and must never be summed across locations

Every one of those was found by running a query, and each one is now a line in a
layer's `instruction`.

## The shape

    DataConnector  dc-<system>     connection coordinates + secretKeyRef
      <- SemanticLayer  sl-<system>
           cubes[]                 one per table
             dimensions[]          one per column you expose
             measures[]            aggregates, optional
           joins[]
           sampleQueries[]         the analysis views
      <- Agent.managed.semanticLayers[]

## Generate it

    asgard-cli add dataconnector <name> --db-class postgres
    asgard-cli add semanticlayer <name> --connector dc-<name>

That writes the structure below with the fields that fail silently already in
place - the display annotation, the labels the UI needs, the current field names.
**Copying the skeleton by hand is where those get lost**, because nothing tells
you they are missing: not helm lint, not CRD validation, not a server dry-run.

The generated file marks the judgement calls TODO. Those are what the rest of
this page is about.

## The skeleton

`templates/data_connector/dc-<system>.yaml` and
`templates/semantic_layer/sl-<system>.yaml`.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: DataConnector
metadata:
  name: dc-<system>
  annotations:
    asgard-ai.com/data-connector-name: "<display name>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  dataConnectorClass: mssql          # or postgres
  mssql:
    host: {{ .Values.<system>DB.host | quote }}
    port: {{ .Values.<system>DB.port }}
    user: {{ .Values.<system>DB.user | quote }}
    database: {{ .Values.<system>DB.database | quote }}
    password:
      valueFrom:
        secretKeyRef:
          key: <system>_db_password
          name: {{ include "<chart>.appSecretName" . }}
---
apiVersion: asgard-ai.com/v1alpha1
kind: SemanticLayer
metadata:
  name: sl-<system>
  annotations:
    asgard-ai.com/semantic-layer-name: "<display name>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  completionModelName: {{ .Values.defaultCompletionModelName | quote }}
  effort: {{ .Values.defaultSemanticLayerEffort | quote }}
  dataConnectorName: dc-<system>
  locale: zh-TW
  timezone: Asia/Taipei
  instruction: |-
    <business rules the column names do not convey: 口徑, tables not to trust,
    joins that fan out, views that do not push filters down>
  cubes:
    - name: <schema>.<table>
      sqlTable: <schema>.<table>
      title: <human readable>
      description: <what this table is, in the customer's language>
      primaryKeyDimensions: [<pk column>]
      dimensions:
        - name: <ColumnName>          # the column its sql selects, not an alias
          sql: '{CUBE}.<ColumnName>'
          title: <human readable>
          description: <what it means - a dimension without this is invisible>
          type: string                 # string | number | time | boolean
      measures:
        - name: total_<thing>
          sql: '{CUBE}.<Column>'
          title: <human readable>
          description: <what it aggregates>
          type: sum
  joins:
    - name: <from_cube>_<from_dim>_<to_cube>_<to_dim>
      description: <what the relationship is>
      relationship: one_to_many
      from: {cube: <schema>.<table_a>, dimensions: [<col>]}
      to:   {cube: <schema>.<table_b>, dimensions: [<col>]}
  sampleQueries:
    - comment: <what it answers, and any caveat about the numbers>
      sql: |
        select ...
```

Connection coordinates are `chartValues` declared in `.asgard-pipeline.yaml` and set on the
platform, one group per connector.
The password is only ever a `secretKeyRef`.

## Fields that are not obvious

**`completionModelName` is required** on a SemanticLayer, unlike an Agent. Take
it from a chart value rather than writing a model name into the template.

**`effort` omitted is not the same as `medium`.** Omitting it means the LLM
processors send no effort parameter at all and the model's own default applies.
Set it explicitly, from a value.

**A dimension's `name` is the column its `sql` selects** - `name: ProductId` with
`sql: '{CUBE}.ProductId'`, never a re-cased alias. Measure names are aggregates
rather than columns, so they stay descriptive (`total_shipped_qty`).

**`description` on every cube, dimension and measure, in the customer's
language.** This is not decoration: it is what the agent reads to decide which
column answers a question. **A dimension with no description is effectively
invisible to the model.**

**Analysis views go in top-level `sampleQueries[]`** (`comment` + `sql`), never as
a cube-level `sql:` virtual cube. **Run every one against the live database
before committing it** - a sampleQuery that errors or returns nonsense actively
misleads the agent that reads it.

**`primaryKeyDimensions`** where the table has a key. **Foreign keys are often
absent** in older business systems; infer relationships from naming, then verify
with a join query before writing a `joins` entry:

    select count(*) from a join b on a.sno = b.sno;
    select count(distinct sno) from b;

**Business rules that are not in the column names go in `instruction`.** That is
where the traps above live, in the layer that owns them.

## Designing the parts the generator leaves TODO

### Which tables become cubes

**Work backwards from the questions, not forwards from the schema.** Take the
queries the customer already runs - the reports someone maintains, the SQL in a
spreadsheet, the two or three things they ask every week - and reverse-engineer
which tables those need. Those tables are the layer.

A layer covering part of a system is a **normal, finished state**, not a
half-built one. Say so at the top of the file:

    # partial 語意層,目前涵蓋兩塊:
    #   1. 料件 × 倉別庫存 —— 由兩支常用的查詢逆推
    #   2. 到貨 —— TASK-002 為到貨通知新增
    # 其餘模組尚未建模。

The cube count follows the **kind of question**, not the size of the database:

| the agent's job | cubes, roughly |
|---|---|
| point lookups - "what is the status of this one" | a handful |
| history and elapsed-time analysis | ten or so |
| operational statistics across a whole process | dozens |

A system with 400 tables where people ask three questions gets a small layer.
Adding cubes "because they are there" widens what the agent can be asked without
widening what it can answer well, and every added cube is more search space
between the question and the right table.

### `instruction` - the rules the column names do not carry

This is the highest-value field in the CR, and the test for what belongs in it is
exact:

> **rules that the column names cannot tell you, and that produce a wrong answer
> if broken.**

Not a description of the data - a list of the ways a plausible query is wrong.
The shapes that keep recurring:

- **a value that must come from one specific row, never the row's own.** A
  "target level" configured in one place, joined a second time with that filter
  pinned - and therefore something that **must never be summed** across rows.
- **a column whose type is not what it looks like.** A flag stored as the string
  `'T'`/`'F'`, where `= false` silently matches nothing.
- **rows that must always be excluded.** Codes with a suffix meaning
  discontinued, soft-deleted rows, test records.
- **the convention for what to show.** "Only list locations where at least one
  quantity is above zero" - obvious to a person, invisible to a model.
- **a table that is already aggregated**, so summing it double-counts.

Each one is a sentence and a reason. Write them as you find them during
introspection, because you will not remember them afterwards - and nobody else
can recover them from the schema.

### `description` on every cube, dimension and measure

This is what the agent reads to decide which column answers a question, so write
it as **what it means to the business**, not what it is technically:

    ✗ 供應商 ID 欄位
    ✓ 每個供應商獨一無二的識別碼,用於區分不同供應商

A dimension with no description is invisible to the model. A dimension whose
description only restates its name is nearly as bad.

### `sampleQueries` - the analysis views

These are the reports the customer already lives on, expressed once so nobody
rebuilds them per conversation. Take them from what people actually run.

**Every one must be executed against the live database before it is committed**,
and it is worth putting the row count you saw in the `comment` - a query that
returns 89 rows sets a different expectation from one that returns 4.

## Two rules about mounting

**`allowWrite: false`** on every binding; the standing architecture is read-only.

**No `allowedCubes`, and this one is not a convention you can depart from.**
An agent may query any table in its own layer - the restriction is which layer it
mounts, not which cubes within it - and `gate` R4 refuses an Agent that sets the
field. Note where it is: `Agent.spec.managed.semanticLayers[]`, never the
`SemanticLayer` itself, which has no such field. So for a layer with no Agent on
it at all there is nothing to set and nothing to narrow, and the only exposure
control is which cubes and dimensions the layer declares - see
`asgard-cli usecase mimir-dashboard`.

**`sampleQuestions` on the layer is not this field's counterpart either.** It is
what Data Insight renders as the buttons under a layer's prompt box, it takes
plain strings, and it is the one field in the chart that changes what a person
sees before they type anything. `usecase mimir-dashboard` is where it is written
up, because that is the shape whose consumer is a person.

## The OLAP exception

A company-wide warehouse layer is a **superset** of the per-system layers, with
unrestricted cross-schema joins. That freedom is the point of a warehouse and
exactly why it is the wrong tool for a chat agent. Such a layer is mounted on no
agent at all.

**Nothing enforces that, and nothing should.** The gate used to, from a list of
layer names recorded per customer - which meant a tool that cannot know what a
customer is building was carrying that customer's use case in a config field.
What the gate does now is report a layer no Agent binds, as an observation; a
layer bound to an Agent that should not be is a review's job, not a rule's.

Where a system has no live database - a third-party SaaS reached only through an
ETL - a layer over the warehouse is the only option. Say so in the layer's
`instruction` and make the agent disclose the lag whenever recency matters.

## Verify

```bash
# every sampleQuery, against the real database, before committing
.venv/bin/python .agents/skills/db-query/scripts/query.py \
  --class <class> --prefix <PREFIX> -f query.sql

asgard-cli check
helm lint projects/<project>/chart/app
asgard-cli verify <project>
```

`asgard-cli verify` confirms `dataConnectorName` resolves and that the agents
reference only layers that exist. **It does not validate SQL against the live
schema** - that is what the introspection queries are for.
