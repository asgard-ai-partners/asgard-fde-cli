---
name: semantic-layer-modeling
description: Use when modeling or extending a SemanticLayer CR — connect to the customer databases defined in .env, introspect schema (tables, columns, keys, relationships), and turn that into cubes / dimensions / measures / joins / sampleQueries. Also the skill for any ad-hoc query against those databases.
version: 1.1.0
alwaysApply: false
---

# SemanticLayer Modeling (design time)

This repo's job is to define the Asgard CRs that let the customer build Agentic applications.
The single highest-value, highest-effort artifact in that set is the **`SemanticLayer` CR** —
the read surface the agent composes SQL against. Modeling one well requires looking at the
**real database**, not guessing.

This skill covers: connecting to those databases, introspecting them, and turning what you find
into a `SemanticLayer` CR that matches this repo's conventions.

> **Design time, not runtime.** This skill is for the *coding agent* working in this repo.
> It is never synced into the platform. Skills the *deployed* agent uses at runtime live in
> `common/skills/<skill>/SKILL.md` and are bound via a `SkillSet` CR — a completely separate
> mechanism. Do not confuse the two.

## When To Use

- Adding or extending a `SemanticLayer` CR (new cube, dimension, measure, join, `sampleQueries`).
- Adding a `DataConnector` for a database not yet wired up.
- Verifying that a cube's `sqlTable` / dimension `sql` actually resolves against the live schema.
- Any ad-hoc query against the five databases (spot-checking data, confirming a status code's real
  values, checking row counts before writing a `sampleQuery`).
- Diagnosing an agent answer that looks wrong and might be a modeling problem rather than a prompt
  problem.

## When To Skip

- Pure chart plumbing that does not touch a `SemanticLayer` or `DataConnector`.
- Prompt / skill / `deploy.yaml` / CI changes.

## Prerequisites (one-time)

```bash
python3 -m venv .venv
.venv/bin/pip install -r scripts/db/requirements.txt   # psycopg[binary], pymssql, pyyaml, pyjwt[crypto]
cp .env.example .env                                   # then fill in the values
```

`.env` is gitignored and holds one connection group per database, mirroring the `DataConnector`
values in that project's `chart/values-dev.yaml` (one group per connector). Copy the non-password
fields from there; passwords come from the cluster secret:

```bash
kubectl get secret app-secret -n <namespace> \
  -o jsonpath='{.data.<key>_db_password}' | base64 -d
```

Never write a password into `.env.example`, a values file, a spec, or a commit. Never echo one
into the transcript.

## The Databases

**Fill this table in as you wire each database up.** It is the map between the
customer's systems, the local tooling and the CRs, and it is the first thing the
next agent reads.

Named targets are registered in `scripts/db/pgenv.py` -> `DB_TARGETS`:

| target | driver | `.env` prefix | System | Wired to |
|---|---|---|---|---|
| | | | | |

Adding a database = register it in `DB_TARGETS`, add its `.env` group to
`.env.example`, and add the matching `DataConnector` CR. The non-password fields
belong in that project's own `chart/values-dev.yaml`.

> **A schema name containing a hyphen must be double-quoted in SQL.**
> `select ... from "db-something_site".products`. Without the quotes PostgreSQL
> parses it as `db` minus `something_site`, and the error message says nothing
> about schemas, which makes it hard to connect to the cause.

**NetSuite, if the customer runs it, is a source but not a `DB_TARGETS` entry** —
it has no DBAPI driver. It is a REST service (SuiteQL), so it gets its own pair
of modules and its own `.env` group:

| target | transport | `.env` prefix | System | Wired to |
|---|---|---|---|---|
| (NetSuite) | SuiteQL over REST | `NS_` | NetSuite ERP | `dc-netsuite` -> `sl-netsuite` |

`scripts/db/nsenv.py` (config + OAuth) and `scripts/db/nsquery.py` (CLI) mirror
`pgenv.py` / `query.py`. Five `.env` keys: `NS_HOST`, `NS_CONSUMER_KEY`,
`NS_CERTIFICATE_ID`, `NS_PRIVATE_KEY_PEM`, `NS_SIGNATURE_ALGORITHM`.
Host/cert/algorithm can be copied from the project's `chart/values-dev.yaml` ->
`netsuite.*`; the consumer key and private key are secrets (`app-secret` ->
`netsuite_consumer_key` / `netsuite_private_key_pem`). The private key is a
PKCS#8 PEM and may be written on one line with literal `\n`.

## Querying

```bash
.venv/bin/python scripts/db/query.py -d <target> "<SQL>"
.venv/bin/python scripts/db/query.py -d <target> -f some.sql
echo "select 1" | .venv/bin/python scripts/db/query.py -d <target>

# NetSuite (SuiteQL) — no -d, it is the only NetSuite account
.venv/bin/python scripts/db/nsquery.py "SELECT 1 AS ok FROM DUAL"
.venv/bin/python scripts/db/nsquery.py --columns transactionline   # 反推欄位名
.venv/bin/python scripts/db/nsquery.py --all "SELECT ..."          # 翻頁撈完
```

The tool prints the connection summary (no password) to stderr and a formatted table to stdout.

> **Read only.** These are live business databases and every read path in this repo is deliberately
> read-only — `SemanticLayer` in `internal`, fixed query tools in `website`, no write path anywhere.
> Issue `SELECT` /
> introspection queries only. Never `INSERT` / `UPDATE` / `DELETE` / `CREATE` / `DROP` /
> `TRUNCATE`, and never run `scripts/db/apply.py` against these targets unless the user explicitly
> asks for a seed operation and names the target.

Dialect matters. Check the driver column of the table above before writing SQL:
PostgreSQL takes `LIMIT n`, MSSQL takes `TOP n`.

## Introspection Recipes

**PostgreSQL**:

```sql
-- tables in a schema
select table_schema, table_name
from information_schema.tables
where table_schema = '<schema>' and table_type = 'BASE TABLE'
order by table_name;

-- columns + types + nullability
select column_name, data_type, is_nullable, character_maximum_length
from information_schema.columns
where table_schema = '<schema>' and table_name = '<table>'
order by ordinal_position;

-- primary keys
select kcu.column_name
from information_schema.table_constraints tc
join information_schema.key_column_usage kcu
  on tc.constraint_name = kcu.constraint_name and tc.table_schema = kcu.table_schema
where tc.constraint_type = 'PRIMARY KEY'
  and tc.table_schema = '<schema>' and tc.table_name = '<table>';

-- foreign keys (the raw material for `joins`)
select tc.table_name as from_table, kcu.column_name as from_col,
       ccu.table_name as to_table,  ccu.column_name as to_col
from information_schema.table_constraints tc
join information_schema.key_column_usage kcu
  on tc.constraint_name = kcu.constraint_name
join information_schema.constraint_column_usage ccu
  on ccu.constraint_name = tc.constraint_name
where tc.constraint_type = 'FOREIGN KEY' and tc.table_schema = '<schema>';
```

**MSSQL**:

```sql
-- tables
select TABLE_SCHEMA, TABLE_NAME from INFORMATION_SCHEMA.TABLES
where TABLE_TYPE = 'BASE TABLE' order by TABLE_NAME;

-- columns
select COLUMN_NAME, DATA_TYPE, IS_NULLABLE, CHARACTER_MAXIMUM_LENGTH
from INFORMATION_SCHEMA.COLUMNS
where TABLE_NAME = '<table>' order by ORDINAL_POSITION;

-- primary keys
select kcu.COLUMN_NAME
from INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
join INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu on tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
where tc.CONSTRAINT_TYPE = 'PRIMARY KEY' and tc.TABLE_NAME = '<table>';
```

**NetSuite (SuiteQL)** has no `information_schema`. Reverse-engineer a table by taking one row:

```bash
.venv/bin/python scripts/db/nsquery.py --columns transaction
```

Four things about SuiteQL that will waste your time if you don't know them — all found the hard way
while modelling Item Receipt (2026-08-18):

- **A NULL column is omitted from the row entirely**, so `--columns` on one sample under-reports.
  `transaction`'s sample was a Customer Deposit and hid `createdfrom` — which turned out not to
  exist on that table at all. Sample a row that actually has the fields you care about
  (e.g. `WHERE mainline = 'F'`).
- **A non-existent column raises a bare `500 UNEXPECTED_ERROR`**, not a "no such column" message.
  So does a `GROUP BY` it dislikes. When a query 500s, bisect it column by column.
- **Dates lose their time on a plain SELECT.** `SELECT t.lastmodifieddate` returns `2026/08/18`;
  `TO_CHAR(t.lastmodifieddate, 'YYYY-MM-DD HH24:MI:SS')` returns `2026-08-18 17:36:40`. Compare with
  `TO_DATE(...)`. Getting this wrong silently degrades an incremental cursor to day precision.
- **Column names come back lowercase** whatever case you write. SuiteQL identifiers are
  case-insensitive, so both work; new cubes here use lowercase to match what NetSuite returns.

Foreign keys are often **absent** in older business systems. When they are, infer relationships
from naming (`sno`, `code`, `departmentcode` and the like) and **verify with a join query**
before writing a `joins` entry:

```sql
-- does this really join 1:N and not blow up?
select count(*) from a join b on a.sno = b.sno;
select count(distinct sno) from b;
```

## Writing the `SemanticLayer` CR

Live under `projects/<project>/chart/app/templates/semantic_layer/sl-<system>.yaml`. Shape (see
existing files for full examples):

```yaml
spec:
  completionModelName: preset-balanced   # REQUIRED on SemanticLayer (unlike Agent)
  effort: medium                         # optional; omitting it is NOT "medium" — it
                                         # means the LLM processors send no effort at
                                         # all and the model's own default applies.
                                         # In this chart: {{ .Values.defaultSemanticLayerEffort | quote }}
                                         # low|medium|high|xhigh|max|auto|disabled
  dataConnectorName: dc-<system>
  locale: zh-TW
  timezone: Asia/Taipei
  cubes:
    - name: <schema>.<table>             # convention: fully-qualified, matches sqlTable
      sqlTable: <schema>.<table>
      title: 派工紀錄                     # 繁中人類可讀
      description: 記錄餐飲設備維修案件的派工資訊。
      primaryKeyDimensions: [id]         # set it when the table has a PK
      dimensions:
        - name: sno
          sql: '{CUBE}.sno'
          title: 叫修編號
          description: 維修申請的流水號。
          type: string                   # string | number | time | boolean
      measures:
        - name: total_returned_qty
          sql: '{CUBE}.ReturnQty'
          title: 退貨數量合計
          description: 退貨數量加總。
          type: sum                      # sum | count | avg | min | max …
  joins:
    - name: <from_cube>_<from_dim>_<to_cube>_<to_dim>   # flattened, dots -> underscores
      description: 維修案件明細與派工紀錄透過叫修編號關聯
      relationship: one_to_many          # one_to_one | one_to_many | many_to_one
      from: {cube: <schema>.<table_a>, dimensions: [sno]}
      to:   {cube: <schema>.<table_b>, dimensions: [sno]}
  sampleQueries:
    - comment: 未完成出貨的訂單視圖。<what it answers, and any 口徑 caveat>
      sql: |
        select ...
```

### Conventions that are easy to get wrong

- **`description` on every cube / dimension / measure, in 繁體中文.** This is not decoration — it
  is what the agent reads to decide which column answers a question. A dimension with no
  description is effectively invisible to the model.
- **Common analysis views go in top-level `sampleQueries[]`** (`comment` + `sql`), **never** as a
  cube-level `sql:` virtual cube.
- **Run every `sampleQuery` against the live DB before committing it.** A `sampleQuery` that errors
  or returns nonsense actively misleads the agent. Record the row count you observed in the
  `comment` if it helps set expectations.
- **`measures` are optional** — several layers here have none. Add one only when there is a real
  aggregate the business asks for; do not mechanically add `count` to every cube.
- The `Agent` binds layers with **no `allowedCubes`**, so every cube you add is immediately
  queryable. Adding a cube widens the agent's reach — that is why this needs a spec (below).
- `Agent.managed.semanticLayers[]` needs `allowQuery: true` and (today) `allowWrite: false`.

## Spec Gate

Per `docs/spec-driven-development.md`, **a new `SemanticLayer`/`DataConnector`, or widening which
cubes the agent may query, requires SDD** — the databases are real. Write the task spec under
`requirements/tasks/`, register it in `requirements/tasks/_index.md`, get it to `ready`, and wait
for explicit instruction before implementing.

Introspection and ad-hoc read queries are *research* and do not need a spec — they are exactly how
you gather the "known context" a spec needs.

## Verify

After editing a `SemanticLayer`, run the repo's gate (see the `asgard-cr-verification` skill):

```bash
asgard-cli check
helm lint projects/<project>/chart/app -f projects/<project>/chart/values-dev.yaml
asgard-cli verify <project>
```

`asgard-cli verify` confirms `SemanticLayer.dataConnectorName` resolves to a real
`DataConnector` and that the `Agent` references only layers that exist. It does **not** validate
SQL against the live schema — that is what your introspection queries are for.

> **Reasoning about a schema is not verifying it.** `sl-netsuite`'s first Item Receipt draft was
> written from a spec document's SuiteQL and looked entirely plausible; running it found that
> `transaction.createdfrom` does not exist, that a missing `isinventoryaffecting = 'T'` filter makes
> quantities sum to **zero**, and that 54% of Item Receipts are transfer orders rather than
> purchases. None of that is visible in column names. Run the query.
