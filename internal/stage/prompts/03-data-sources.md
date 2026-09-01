Projects exist but no DataConnector does, so nothing can be read yet.

Missing a connector:
<<range .Projects>><<if not (.Has "DataConnector")>>  - <<.Slug>>
<<end>><<end>>
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
     Connection coordinates go in that project's chart/values-<env>.yaml; the
     password is a secretKeyRef into app-secret and never enters values or git.

Read-only throughout. SELECT and introspection only.

Done when: every project that reads something has its DataConnector, and
asgard-cli verify resolves it.
