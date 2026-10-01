---
description: the ladder for what to ask them for, filling .env without anybody typing a password, introspecting the real schema rather than guessing it
---
# Wire up the customer's databases

This stage applies to a project that has no DataConnector yet, so nothing in it
can be read.

Missing a connector:
<<range .Projects>><<if not (.Has "DataConnector")>>  - <<.Slug>>
<<end>><<end>>
## Two ladders, and this page is only one of them

    what to ASK THEM FOR      docs > source code > API spec > the DB > the UI
    how to READ it, below     a database  >  an API or another protocol  >
                              a screen

They have the same shape and different purposes, and an FDE uses both in one
meeting. Use the asking ladder first. With the system's documentation - or
better, its source - the fields, validation rules and status codes can be read
directly instead of obtained by questioning, and the choices on this page are
made from facts instead of guesses.

If you are at this stage with no documentation, go back and get it before
writing a connector. `../guide/requirements.md` has the full
ladder and why source code sits above a spec.

## Take the most capable route each system offers

    a database we can read   -> DataConnector + SemanticLayer
                                ../usecase/semantic-layer.md
    an API                   -> a Workflow with an http-request processor,
                                or a skill teaching the agent to call it
                                ../usecase/external-api.md
                                ../usecase/api-oauth.md  (if it needs a token)
    another protocol         -> a runtime skill, with the agent running the
                                client in its own sandbox: SSH, SNMP, a vendor
                                CLI. As good as an API, and open only where the
                                credential arrives per turn from the caller -
                                nothing puts a static service key in a sandbox
                                ../usecase/skill-set.md
    a screen only            -> last resort. Brittle and slow, and it breaks
                                when the vendor changes their UI. Raise it as
                                a question before designing around it.
                                ../usecase/browser-operation.md

For a system with both a database and an API, use the database for reading. The
agent composes its own queries, joins across tables, and is not limited to the
calls someone thought to expose. Keep the API for writes.

Before designing five integrations, ask whether one already exists. A
customer selling on several channels usually has something that consolidates
them - middleware, an OMS, a warehouse that already pulls orders in. If they do,
those channels are rows in one database rather than five external APIs, and it
is a much better design. If nobody knows, record it as an open question
rather than assuming an answer: put it in `docs/open-questions.md` and say which branch you are
proceeding on.

## Do not guess a schema

Load the `db-query` skill under .agents/skills/ and connect to the real system.
A schema inferred from documents has to be verified against the database.
`semantic-layer-modeling` records what happened when it was not: a plausible-looking query written from a spec document
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
     what the value was. Never ask anybody to type a password into the
     conversation. The coordinates and the password both come from them, not
     from the cluster. The Secret in the cluster is the deployed CR's copy, on
     its own lifecycle; reading it to get a design-time credential conflates two
     mechanisms that must stay apart.

     Re-read .env after they save. They can add keys - a second database
     nobody had mentioned - and the form allows that on purpose.

  3. Verify the connection before writing any CR:

         .venv/bin/python .agents/skills/db-query/scripts/query.py \
           --class postgres --prefix UOF_DB_ "select 1"

  4. Introspect: tables, columns, primary keys, foreign keys. db-query's
     references/connectors.md has the recipes per engine. Foreign keys are often
     absent in older business systems - infer relationships from naming, then
     verify with a join query before writing anything down.

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

**Checked:** against asgard-kube `cbd8d70`. `dataConnectorClass` is
one of postgres, mysql, mssql, oracle, salesforce, hana, netsuite, trino, athena,
and is immutable after creation - so "a database we can read" covers nine engines
and picking the wrong one is a replacement rather than an edit. db-query has a
design-time path for eight of them; hana has none, because SAP's driver is not
installable from PyPI. Every credential
block takes a `secretKeyRef`, with the CRD enforcing exactly one of
[secretKeyRef configMapKeyRef] and exactly one of [value valueFrom], so the
instruction to keep the password out of values and git is supported by the
contract rather than only by convention.

**Unchecked:** the ladder - source code over a spec, a database over an API, an
API over a screen - is this engagement's ordering and no source states it;
`../usecase/browser-operation.md` ranks only its own rung, a screen as the last
resort.
