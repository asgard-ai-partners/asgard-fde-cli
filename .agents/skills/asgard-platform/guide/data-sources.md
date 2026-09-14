# Wire up the customer's databases

**This stage is a project that has no DataConnector yet**, so nothing in it
can be read.

    what to ASK THEM FOR      docs > source code > API spec > the DB > the UI
    how to READ it, below     a database  >  an API  >  a screen

They have the same shape and opposite purposes, and an FDE uses both in one
meeting. **The asking ladder runs first and makes this page easier**: with the
system's documentation - or better, its source - the fields, validation rules and
status codes are read rather than extracted from somebody by questioning, and
what arrives here is a design decision rather than a guess.

If you are at this stage with no documentation, that is worth going back for
before writing a connector. `../guide/requirements.md` has the full
ladder and why source code sits above a spec.

## Take the most capable route each system offers

    a database we can read   -> DataConnector + SemanticLayer
                                ../usecase/semantic-layer.md
    an API                   -> a Workflow with an http-request processor,
                                or a skill teaching the agent to call it
                                ../usecase/external-api.md
                                ../usecase/api-oauth.md  (if it needs a token)
    a screen only            -> last resort. Brittle and slow, and it breaks
                                when the vendor changes their UI. Raise it as
                                a question before designing around it.
                                ../usecase/browser-operation.md

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

Load the `db-query` skill under .agents/skills/ and connect to the real system.
Reasoning about a schema is not verifying it - `semantic-layer-modeling` records
what that cost last time: a plausible-looking query written from a spec document
turned out to reference a column that does not exist, to sum to zero because of a
missing filter, and to be wrong about where half the rows came from. None of that
is visible in column names.

    ../usecase/semantic-layer.md

## The steps

  1. Pick a prefix for the system - one per source system - and write the keys
     it needs into .env, which is gitignored:

         cp .env.example .env
         .venv/bin/python .agents/skills/db-query/scripts/query.py \
           --class postgres --prefix UOF_DB_ --keys >> .env

  2. Hand the filling-in to whoever owns that system:

         asgard-cli local-env --focus UOF_DB_HOST,UOF_DB_PASSWORD

     It opens a form on 127.0.0.1 and tells you which keys got a value, never
     what the value was. **Never ask anybody to type a password into the
     conversation.** The coordinates and the password both come from them - not
     from the cluster. The Secret in the cluster is the deployed CR's copy, on
     its own lifecycle; reading it to get a design-time credential conflates two
     mechanisms that must stay apart.

     **Re-read .env after they save.** They can add keys - a second database
     nobody had mentioned - and that is deliberate.

  3. Verify the connection before writing any CR:

         .venv/bin/python .agents/skills/db-query/scripts/query.py \
           --class postgres --prefix UOF_DB_ "select 1"

  4. Introspect: tables, columns, primary keys, foreign keys. db-query's
     references/connectors.md has the recipes per engine. Foreign keys are often
     absent in older business systems - infer relationships from naming, then
     **verify with a join query** before writing anything down.

  5. Write one DataConnector CR per system, under
     projects/<project>/chart/app/templates/data_connector/dc-<system>.yaml.
     The .env key suffixes are the CR field names, so this step is a transcription
     rather than a translation. Connection coordinates are declared as chartValues
     in .asgard-pipeline.yaml and their values set on the platform per release;
     the password is declared there too, under appSecret, and set with
     `variables set --kind secret`. It is only ever a secretKeyRef, and the
     Secret's name comes from the platform - `asgard-cli add dataconnector`
     prints the exact key names it wrote.

Read-only throughout. SELECT and introspection only, and db-query enforces it.

Done when: every project that reads something has its DataConnector, and
asgard-cli verify resolves it.

**Checked:** 2026-09-04, re-read 2026-09-11 against asgard-kube `cbd8d70`. `dataConnectorClass` is
one of postgres, mysql, mssql, oracle, salesforce, hana, netsuite, trino, athena,
and is immutable after creation - so "a database we can read" covers nine engines
and picking the wrong one is a replacement rather than an edit. db-query has a
design-time path for eight of them; hana has none, because SAP's driver is not
installable from PyPI. Every credential
block takes a `secretKeyRef`, with the CRD enforcing exactly one of
[secretKeyRef configMapKeyRef] and exactly one of [value valueFrom], so the
instruction to keep the password out of values and git is supported by the
contract rather than only by convention.

**Unchecked:** the ladder. That source code beats a spec, a database beats an
API and an API beats a screen is this engagement's ordering of what to reach
for, and **no source states it** - the pages it points at describe each shape
without ranking them. Read it as the order that has paid off here, and the last
rung as the one to raise as a question rather than design around.
