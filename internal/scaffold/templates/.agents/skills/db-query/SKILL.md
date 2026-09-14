---
name: db-query
description: Use when you need to look at a customer source system at design time - list tables, read a column's real values, check a row count, confirm a join actually holds, or verify a query before it goes into a SemanticLayer or a fixed-query Toolset. Covers eight of the nine DataConnector classes the platform can read - postgres, mysql, mssql, oracle, salesforce, netsuite, trino, athena - and not hana, whose driver SAP does not distribute openly; how a connection is configured; and how to get a credential without asking anybody to type a password into a chat.
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

Define the shorthand as a FUNCTION, not a variable. `Q="... query.py"` then
`$Q --classes` is a bash idiom that fails mute under zsh, which is the login
shell on every macOS this is delivered to: zsh does not word-split an unquoted
parameter expansion, so `$Q` is one word and the whole string is treated as a
command name. You get `command not found` and exit 127, with nothing to suggest
the shell rather than the venv or the credentials. A function word-splits
correctly in both shells.

```bash
Q() { .venv/bin/python .agents/skills/db-query/scripts/query.py "$@"; }

Q --classes                                        # the classes this tool drives
Q --class postgres --prefix UOF_DB_ --keys         # the .env keys this connection needs
Q --class postgres --prefix UOF_DB_ "select 1"
Q --class postgres --prefix UOF_DB_ -f some.sql
echo "select 1" | Q --class postgres --prefix UOF_DB_
Q --class postgres --prefix UOF_DB_ --columns sales.orders    # what columns does it have
Q --class netsuite --prefix NS_ "SELECT 1 AS ok FROM DUAL"
```

The connection summary - **never a password** - goes to stderr and the result
table to stdout, so it pipes.

`--limit` caps how many rows are pulled back (200 by default, `0` for all). It is
a client-side fetch limit, not something appended to your SQL, so it works the
same on every engine.

**A failure is one sentence, not a stack trace** - which credential or which
column, in the driver's own words. `--traceback` on the same command gives the
full thing when the sentence is not enough.

**`--prefix` is not optional and there is no default.** One repository can
connect to two PostgreSQL databases, which are two groups of keys in one `.env`
and cannot share names. The prefix is how you say which group. If you omit it,
the tool lists the prefixes it can see.

`references/connectors.md` has, for each class: the CR fields, the matching
`.env` keys, the driver, and the introspection recipes that `--columns` does not
cover (tables, primary keys, foreign keys).

**Pass `--limit 0` for a schema sweep.** The default 200 is right for looking at
data and wrong for enumerating one: an `information_schema.tables` count on a
database with several hundred base tables is silently truncated at 200 rows, and
a truncated enumeration looks exactly like a complete one. One real source in
this shape held roughly 580.

**Eight is this tool's number, not the platform's.** `DataConnectorClass` has
**nine** values and `hana` is the ninth: `spec.hana` is in the CRD and a HANA
source is a perfectly valid `DataConnector`, but SAP distributes `hdbcli` under
its own licence rather than from PyPI, so it cannot be a line in
`requirements.txt` that anyone can install. **A customer on SAP HANA is not out
of reach of the platform** - only out of reach of this skill. Introspect it with
whatever their own DBAs use and record what you found in the spec.

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

3. Open the form for whoever holds the credential:

   ```bash
   asgard-cli local-env --focus UOF_DB_HOST,UOF_DB_PASSWORD
   ```

   It serves one page on 127.0.0.1, and when they save it tells you **which
   keys now have a value and nothing else**. `--focus` highlights the ones you
   are waiting for without hiding the rest.

4. **Re-read `.env`.** They can add keys from the form - a second database you
   had not heard of - and that is deliberate, so do not assume you got back
   exactly the list you asked for.
5. Run `select 1` to confirm, then get on with the introspection.

**You write the keys because only you know what you are about to connect to.**
They know the values; naming what has to be filled in is your half of it.

### The kinds of credential, and they are not interchangeable

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

**Checked:** 2026-09-07, re-read 2026-09-11 against asgard-kube `cbd8d70` and against this skill's
own scripts. `DataConnectorClass` has nine values; `SPECS` in
`scripts/connectors.py` implements eight and its comment says which is left out
and why. **The description claimed "all eight DataConnector classes the platform
can read", and the platform reads nine** - a completeness claim that would tell
a reader asked about SAP HANA that the platform cannot reach it, when what
cannot reach it is this tool. Corrected here and in `--classes`.

**Checked against a real customer system:** postgres, netsuite, mssql. The
commands, the `--prefix` rule and the failure shapes come from one engagement's
PostgreSQL and NetSuite work. **mssql** was driven end to end by a later one:
`--keys` emitted every key, the optional named-instance one included, and
`asgard-cli local-env` filled them, `select 1` connected, the stderr summary
printed `user@host:port/database` with no password as documented, and a full
introspection ran - `information_schema.tables` counts, `sys.tables` joined to
`sys.partitions` for row counts, `--columns` on individual tables, and
ad-hoc join-verification queries across the whole database. The read-only
guard refused nothing, because every statement was a `SELECT`. The `TOP n` /
`LIMIT n` dialect note in `references/connectors.md` was correct and needed.

**Unchecked:** oracle, salesforce, trino, athena and mysql are configured from
`references/connectors.md` and the CR fields and **have not been run against a
customer's system from here**. The credential path is one engagement's plus that
one: `asgard-cli local-env` exists so nobody types a password at an agent, and
whether that survives a customer whose credentials come through their own vault
is still the first real test of it.
