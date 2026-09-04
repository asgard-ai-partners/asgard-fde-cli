# Wire up the customer's databases

Projects exist but no DataConnector does, so nothing can be read yet.

Missing a connector:
<<range .Projects>><<if not (.Has "DataConnector")>>  - <<.Slug>>
<<end>><<end>>
## Two ladders, and this page is only one of them

    what to ASK THEM FOR      docs > source code > API spec > the DB > the UI
    how to READ it, below     a database  >  an API  >  a screen

They have the same shape and opposite purposes, and an FDE uses both in one
meeting. **The asking ladder runs first and makes this page easier**: with the
system's documentation - or better, its source - the fields, validation rules and
status codes are read rather than extracted from somebody by questioning, and
what arrives here is a design decision rather than a guess.

If you are at this stage with no documentation, that is worth going back for
before writing a connector. `asgard-cli guide requirements` has the full
ladder and why source code sits above a spec.

## Take the most capable route each system offers

    a database we can read   -> DataConnector + SemanticLayer
                                asgard-cli usecase semantic-layer
    an API                   -> a Workflow with an http-request processor,
                                or a skill teaching the agent to call it
                                asgard-cli usecase external-api
                                asgard-cli usecase api-oauth  (if it needs a token)
    a screen only            -> last resort. Brittle and slow, and it breaks
                                when the vendor changes their UI. Raise it as
                                a question before designing around it.
                                asgard-cli usecase browser-operation

**A system with both a database and an API: take the database** for reading. The
agent composes its own queries, joins across tables, and is not limited to the
calls someone thought to expose. Keep the API for writes.

**Before designing five integrations, ask whether one already exists.** A
customer selling on several channels usually has something that consolidates
them - middleware, an OMS, a warehouse that already pulls orders in. If they do,
those channels are rows in one database rather than five external APIs, and it
is a much better design. If nobody knows, that is an open question, not an
assumption: put it in `docs/open-questions.md` and say which branch you are
proceeding on.

## Do not guess a schema

Load the `semantic-layer-modeling` skill under .agents/skills/ and connect to the
real database. Reasoning about a schema is not verifying it - the same skill
records what that cost last time: a plausible-looking query written from a spec
document turned out to reference a column that does not exist, to sum to zero
because of a missing filter, and to be wrong about where half the rows came from.
None of that is visible in column names.

    asgard-cli usecase semantic-layer

## The steps

  1. cp .env.example .env, then fill in the connection values. The non-password
     fields come from the customer; the passwords come from the cluster secret:

         kubectl get secret app-secret -n <namespace> \
           -o jsonpath='{.data.<key>}' | base64 -d

  2. Register each database in DB_TARGETS in scripts/db/pgenv.py.

  3. Verify the connection before writing any CR:

         .venv/bin/python scripts/db/query.py -d <target> "select 1"

  4. Introspect: tables, columns, primary keys, foreign keys. The skill has the
     recipes for PostgreSQL and MSSQL. Foreign keys are often absent in older
     business systems - infer relationships from naming, then **verify with a
     join query** before writing anything down.

  5. Write one DataConnector CR per database, under
     projects/<project>/chart/app/templates/data_connector/dc-<system>.yaml.
     Connection coordinates are declared as chartValues in .asgard-pipeline.yaml
     and their values set on the platform per release; the
     password is a secretKeyRef into app-secret and never enters values or git.

Read-only throughout. SELECT and introspection only.

Done when: every project that reads something has its DataConnector, and
asgard-cli verify resolves it.

**Checked:** 2026-09-04 against asgard-kube `15ded0f`. `dataConnectorClass` is
one of postgres, mysql, mssql, oracle, salesforce, hana, netsuite, trino, athena,
and is immutable after creation - so "a database we can read" covers ten engines
and picking the wrong one is a replacement rather than an edit. Every credential
block takes a `secretKeyRef`, with the CRD enforcing exactly one of
[secretKeyRef configMapKeyRef] and exactly one of [value valueFrom], so the
instruction to keep the password out of values and git is supported by the
contract rather than only by convention.

**Unchecked:** the ladder. That source code beats a spec, a database beats an
API and an API beats a screen is this engagement's ordering of what to reach
for, and **no source states it** - the pages it points at describe each shape
without ranking them. Read it as the order that has paid off here, and the last
rung as the one to raise as a question rather than design around.
