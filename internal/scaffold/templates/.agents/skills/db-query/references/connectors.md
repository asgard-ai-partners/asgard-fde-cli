# What each connector needs

One section per `DataConnectorClass`. Each lists the fields the CR declares, the
`.env` key that carries the same field at design time, and how to introspect that
engine once you are connected.

## The naming rule

**A `.env` key is `<PREFIX>` + the CR field name in upper snake case.**

    spec.postgres.sslMode      ->  <PREFIX>SSL_MODE
    spec.athena.accessKeyId    ->  <PREFIX>ACCESS_KEY_ID
    spec.netsuite.privateKeyPem -> <PREFIX>PRIVATE_KEY_PEM

The rule exists so that what you fill in locally maps one-to-one onto what the CR
declares. Nobody has to translate between two vocabularies, and a field that has
no `.env` key is a field somebody forgot.

**`<PREFIX>` is yours to choose, and this skill will not pick one for you.** An
engagement with two PostgreSQL databases has two groups of keys in one `.env`,
and they cannot share names. Pick a prefix per system - `UOF_DB_`, `ERP_DB_`,
`NS_` - and pass it to every command with `--prefix`.

`query.py --class <class> --prefix <PREFIX> --keys` prints the block to append.

## Where the values come from, and where they do not

The connection coordinates come from **the customer**: the person who runs that
system. The password comes from them too.

**Not from the cluster.** A running `DataConnector` reads its password from a
Kubernetes Secret the platform provisions, and that is a different mechanism with
a different lifecycle - see the boundary table in `SKILL.md`. The value may well
be the same string; how you obtain it is not.

## postgres

`spec.postgres` - driver `psycopg[binary]>=3`

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `host` | `<P>HOST` | yes | bare host, no scheme |
| `port` | `<P>PORT` | yes in the CR | design time defaults to 5432 |
| `user` | `<P>USER` | yes | |
| `password` | `<P>PASSWORD` | yes | secret |
| `database` | `<P>DATABASE` | yes | |
| `sslMode` | `<P>SSL_MODE` | no | the CRD enum is `disable`, `require`, `verify-ca`, `verify-full` - narrower than libpq's, so `prefer` is not declarable even though libpq accepts it. Leave the key empty to let libpq decide locally, and set the CR to what the customer's server actually requires |

> **A schema name containing a hyphen must be double-quoted in SQL.**
> `select ... from "sales-eu".products`. Without the quotes PostgreSQL parses it
> as `sales` minus `eu`, and the error message says nothing about schemas.

## mysql

`spec.mysql` - driver `PyMySQL>=1.1`

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `host` | `<P>HOST` | yes | |
| `port` | `<P>PORT` | yes in the CR | design time defaults to 3306 |
| `user` | `<P>USER` | yes | |
| `password` | `<P>PASSWORD` | yes | secret |
| `database` | `<P>DATABASE` | yes | |

MySQL has no schema layer: a "schema" is a database. `information_schema` is
still there, and `table_schema` holds the database name.

## mssql

`spec.mssql` - driver `pymssql>=2.3`

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `host` | `<P>HOST` | yes | |
| `port` | `<P>PORT` | yes in the CR | design time defaults to 1433 |
| `user` | `<P>USER` | yes | |
| `password` | `<P>PASSWORD` | yes | secret |
| `database` | `<P>DATABASE` | yes | |
| `instance` | `<P>INSTANCE` | no | a named instance; most deployments leave it empty. `query.py` joins it as `host\instance` |

Dialect: `TOP n`, not `LIMIT n`.

## oracle

`spec.oracle` - driver `oracledb>=2` (thin mode - no Oracle Instant Client)

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `host` | `<P>HOST` | yes | |
| `port` | `<P>PORT` | yes in the CR | design time defaults to 1521 |
| `user` | `<P>USER` | yes | |
| `password` | `<P>PASSWORD` | yes | secret |
| `serviceName` | `<P>SERVICE_NAME` | one of | |
| `sid` | `<P>SID` | one of | |

**The CRD enforces exactly one of `serviceName` and `sid`**, and so does
`query.py`. Filling both is not "belt and braces", it is a rejected CR.

Oracle stores identifiers upper case in its dictionary, and has no
`information_schema` - use `all_tab_columns` / `all_tables`.

## salesforce

`spec.salesforce` - no driver; OAuth 2.0 + REST over the standard library

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `host` | `<P>HOST` | yes | `<org>.my.salesforce.com`; a sandbox is `<org>.sandbox.my.salesforce.com` |
| `consumerKey` | `<P>CONSUMER_KEY` | yes | secret |
| `consumerSecret` | `<P>CONSUMER_SECRET` | yes | secret |

**The Connected App must have the client credentials flow enabled and a run-as
user assigned.** Without it the token endpoint answers `400
unsupported_grant_type`, which does not say which of the two is missing.

Salesforce speaks SOQL, not SQL: no `JOIN`, no `information_schema`. Relationships
are traversed with dots (`Account.Name` from `Contact`), and the field list of an
object comes from the describe endpoint - `query.py --columns <SObject>` calls it.

## netsuite

`spec.netsuite` - driver `pyjwt[crypto]>=2.8` (SuiteQL over REST)

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `host` | `<P>HOST` | yes | account host, no scheme |
| `consumerKey` | `<P>CONSUMER_KEY` | yes | secret |
| `signatureAlgorithm` | `<P>SIGNATURE_ALGORITHM` | yes in the CR | `ES256`, `ES512` or `PS256`; design time defaults to `ES256` |
| `certificateId` | `<P>CERTIFICATE_ID` | yes | |
| `privateKeyPem` | `<P>PRIVATE_KEY_PEM` | yes | secret. PKCS#8 PEM. May be written on one line with literal `\n` - `netsuite.py` restores the newlines |

Three things that cost time, all found the hard way:

- **The OAuth scope must be `rest_webservices`.** SuiteQL and REST Web Services
  use it; RESTlets use `restlets`. The wrong one gives `401
  INVALID_LOGIN_ATTEMPT`, which does not mention scope.
- **A non-existent column raises a bare `500 UNEXPECTED_ERROR`**, not "no such
  column". So does a `GROUP BY` it dislikes. When a query 500s, bisect it column
  by column.
- **Dates lose their time on a plain SELECT.** `SELECT t.lastmodifieddate`
  returns `2026/08/18`; `TO_CHAR(t.lastmodifieddate, 'YYYY-MM-DD HH24:MI:SS')`
  returns `2026-08-18 17:36:40`. Getting this wrong silently degrades an
  incremental cursor to day precision.

## trino

`spec.trino` - driver `trino>=0.330`

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `scheme` | `<P>SCHEME` | yes in the CR | `http` or `https`; design time defaults to `https` |
| `host` | `<P>HOST` | yes | |
| `port` | `<P>PORT` | yes in the CR | design time defaults to 443 |
| `user` | `<P>USER` | yes | |
| `password` | `<P>PASSWORD` | no | secret |
| `jwtAccessToken` | `<P>JWT_ACCESS_TOKEN` | no | secret |
| `sslCertificatePem` | `<P>SSL_CERTIFICATE_PEM` | no | secret. A PEM, which `query.py` writes to a temporary file because the HTTP client wants a CA bundle path |

**All three credentials are optional**, in the CRD and here. A Trino behind a
gateway that authenticates for it needs none of them, and that is a normal
deployment rather than a half-filled configuration.

**Over `http`, no credential is sent at all** - the Trino client refuses basic
auth without TLS, and a JWT over cleartext is the token given away. Setting
`SCHEME=http` together with a password or a JWT is refused rather than quietly
downgraded: pick `https`, or leave the credential empty.

**Trino federates, so a table is `catalog.schema.table`** - three parts, not two.
`information_schema` exists once per catalog, so an unqualified query answers
`MISSING_CATALOG_NAME`, which does not mention what is missing.
`--columns tpch.sf1.orders` works; `--columns sf1.orders` is refused with the
reason. `show catalogs` lists what is mounted.

## athena

`spec.athena` - driver `PyAthena>=3`

| CR field | `.env` key | required | notes |
|---|---|---|---|
| `region` | `<P>REGION` | yes | e.g. `ap-northeast-1` |
| `accessKeyId` | `<P>ACCESS_KEY_ID` | yes | secret |
| `secretAccessKey` | `<P>SECRET_ACCESS_KEY` | yes | secret |
| `outputLocation` | `<P>OUTPUT_LOCATION` | yes | **must start with `s3://`** - the CRD has a rule for it. Athena writes every result set there, so the key needs write access to that bucket even for a read-only query |
| `workGroup` | `<P>WORK_GROUP` | no | |

Athena reads a Glue catalogue: `information_schema` works, and the "database" is
a Glue database.

## hana

`spec.hana` exists in the CRD - `host`, `port`, `user`, `password` - and
**`db-query` has no path for it.** SAP's Python driver (`hdbcli`) is distributed
under SAP's own licence rather than from PyPI under an open one, so it cannot be
a line in `requirements.txt` that anyone can install.

A HANA source is still a valid `DataConnector`. Introspect it with whatever the
customer's own DBAs use, and record what you found in the spec.

## Introspection recipes

`query.py --columns <table>` covers the common case. These are the rest.

### `information_schema` engines: postgres, mysql, mssql, trino, athena

```sql
-- tables in a schema
select table_schema, table_name
from information_schema.tables
where table_schema = '<schema>' and table_type = 'BASE TABLE'
order by table_name;

-- primary keys (postgres, mysql, mssql)
select kcu.column_name
from information_schema.table_constraints tc
join information_schema.key_column_usage kcu
  on tc.constraint_name = kcu.constraint_name and tc.table_schema = kcu.table_schema
where tc.constraint_type = 'PRIMARY KEY'
  and tc.table_schema = '<schema>' and tc.table_name = '<table>';

-- foreign keys, the raw material for a SemanticLayer's joins (postgres)
select tc.table_name as from_table, kcu.column_name as from_col,
       ccu.table_name as to_table,  ccu.column_name as to_col
from information_schema.table_constraints tc
join information_schema.key_column_usage kcu
  on tc.constraint_name = kcu.constraint_name
join information_schema.constraint_column_usage ccu
  on ccu.constraint_name = tc.constraint_name
where tc.constraint_type = 'FOREIGN KEY' and tc.table_schema = '<schema>';
```

Trino and Athena have `information_schema.tables` and `.columns` but no
constraint tables - a lake has no foreign keys to read. Infer, then verify.

### oracle

```sql
select table_name from all_tables where owner = '<SCHEMA>' order by table_name;

select cc.column_name
from all_constraints c
join all_cons_columns cc on c.constraint_name = cc.constraint_name
where c.constraint_type = 'P' and c.owner = '<SCHEMA>' and c.table_name = '<TABLE>';
```

Identifiers are upper case in the dictionary whatever case you wrote them in.

### netsuite

No `information_schema`. Take one row and read its keys - that is what
`--columns` does. **A NULL column is omitted from the row entirely**, so one
sample under-reports: take a row that actually has the fields you care about
(`WHERE mainline = 'F'` on a transaction, for instance) and ask again.

### salesforce

`--columns <SObject>` calls the describe endpoint, which returns every field with
its type and label. To list the objects themselves, the endpoint is
`/services/data/v60.0/sobjects`.

## Foreign keys are often absent

Older business systems frequently declare none. When they do not, infer the
relationship from naming and **verify it with a join query before writing it into
a `SemanticLayer`**:

```sql
select count(*) from a join b on a.sno = b.sno;
select count(distinct sno) from b;
```

A relationship that is 1:N in the customer's head and 1:1 in their data - or the
reverse - produces a layer that returns plausible wrong numbers.
