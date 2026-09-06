---
name: db-query
description: Use when you need to look at a customer source system at design time - list tables, read a column's real values, check a row count, confirm a join actually holds, or verify a query before it goes into a SemanticLayer or a fixed-query Toolset. Covers all eight DataConnector classes the platform can read (postgres, mysql, mssql, oracle, salesforce, netsuite, trino, athena), how a connection is configured, and how to get a credential without asking anybody to type a password into a chat.
version: 1.0.0
alwaysApply: false
---

# Querying a customer's source system (design time)

Modelling anything against a customer's data - a `SemanticLayer`, a fixed-query
`Toolset`, a Syncer's incremental cursor - requires looking at the **real**
system. Reasoning about a schema is not verifying it.

This skill is the one way this repository reaches a source system, and it is
**read-only in both directions**: it reads their data, and it never writes.

> **Design time, not runtime.** This is for the coding agent working in this
> repository, on this laptop. The deployed agent reaches the same databases
> through a `DataConnector` in the cluster, with credentials the platform
> provisions. The two are separate mechanisms; see the boundary below.

## When to use

- Before writing any `SemanticLayer` cube, dimension, measure or join.
- Before committing a `sampleQuery` - **run it**; a query that errors or returns
  nonsense is worse than an absent one, because the agent treats it as an example.
- When a `DataConnector` is about to be written and nobody has confirmed the
  coordinates connect.
- When an agent answer looks wrong and it might be the model of the data rather
  than the prompt.
- Any ad-hoc read against a system this repository already connects to.

## When to skip

- Chart plumbing, prompts, `.asgard-pipeline.yaml` - nothing there touches a
  source system.
- Anything that would **write** to the customer's system. There is no path for it
  here and adding one is not the answer; see below.

## Setup, once

```bash
python3 -m venv .venv
.venv/bin/pip install -r .agents/skills/db-query/scripts/requirements.txt
```

Installing all of it is fine, but unnecessary - the drivers are imported lazily,
so you only need the ones for the classes this engagement actually connects to.
A missing one names itself when you first use that class.

## Commands

```bash
Q=".venv/bin/python .agents/skills/db-query/scripts/query.py"

$Q --classes                                        # the eight supported classes
$Q --class postgres --prefix UOF_DB_ --keys         # the .env keys this connection needs
$Q --class postgres --prefix UOF_DB_ "select 1"
$Q --class postgres --prefix UOF_DB_ -f some.sql
echo "select 1" | $Q --class postgres --prefix UOF_DB_
$Q --class postgres --prefix UOF_DB_ --columns sales.orders    # what columns does it have
$Q --class netsuite --prefix NS_ "SELECT 1 AS ok FROM DUAL"
```

The connection summary - **never a password** - goes to stderr and the result
table to stdout, so it pipes.

`--limit` caps how many rows are pulled back (200 by default, `0` for all). It is
a client-side fetch limit, not something appended to your SQL, so it works the
same on every engine.

**`--prefix` is not optional and there is no default.** One repository can
connect to two PostgreSQL databases, which are two groups of keys in one `.env`
and cannot share names. The prefix is how you say which group. If you omit it,
the tool lists the prefixes it can see.

`references/connectors.md` has, for each class: the CR fields, the matching
`.env` keys, the driver, and the introspection recipes that `--columns` does not
cover (tables, primary keys, foreign keys).

## Getting a credential, without asking for a password in chat

**Never ask the user to tell you a password.** Not in the conversation, not in a
comment, not "just paste it and I will remove it after". A credential that has
been through a transcript has to be treated as disclosed.

The flow is:

1. Decide the prefix for this system - one per source system, e.g. `UOF_DB_`.
2. Generate the keys and append them to `.env`, which is gitignored:

   ```bash
   .venv/bin/python .agents/skills/db-query/scripts/query.py \
     --class postgres --prefix UOF_DB_ --keys >> .env
   ```

3. Ask the user to fill in the values - **name the keys, never ask for a value**.
4. Run `select 1` to confirm, then get on with the introspection.

**You write the keys because only you know what you are about to connect to.**
The user knows the values; naming what has to be filled in is your half of it.

### Three kinds of credential, and they are not interchangeable

| | who uses it | where it lives | who fills it |
|---|---|---|---|
| **design time** | the coding agent, on this laptop | this repo's `.env`, never committed | the customer, or the FDE |
| pipeline variables | one run, on the platform | the Platform | the FDE, with `asgard-cli pipeline variables set` |
| runtime secret | the CR in the cluster | a Kubernetes Secret | the platform provisions it |

The **value** may be the same string in all three - it is the same database. How
it is set is not, and mixing them up is how a production password ends up
somewhere it cannot be withdrawn from. In particular: **do not read the runtime
Secret to obtain a design-time credential.** Ask the person who owns the system.

## Read only

These are the customer's live business systems.

`query.py` refuses anything that is not `SELECT`, `WITH`, `SHOW` or `DESCRIBE`,
and refuses those too if a write keyword appears anywhere in the statement -
because `WITH x AS (INSERT ... RETURNING *) SELECT ...` starts with `WITH` and
writes. The check is deliberately blunt and will occasionally refuse a harmless
query (MySQL's `TRUNCATE()` numeric function, for one). Phrase it another way.

**If the customer genuinely needs something written, they run it.** Not this
tool, not this agent, and the decision goes in `docs/decisions/` first.

## What to do with what you find

Introspection and ad-hoc reads are **research**: they need no spec, and they are
exactly how the "known context" a spec needs gets gathered.

Writing what you found into a `SemanticLayer` or a `DataConnector` is not
research. `docs/spec-driven-development.md` has the rule; the short form is that
a new `DataConnector` or `SemanticLayer`, or widening which cubes an agent may
query, is spec-driven work.

Record what surprised you. A status column whose real values are not what the
documentation says, a join that turns out to be many-to-many, a table that is
empty in practice - none of that is visible again later unless somebody writes it
down, and the next reader will otherwise re-derive it from the same queries.
