package work

// The two spec skeletons. They are Go templates rather than files under
// templates/ because they are the shape of a record this package also parses:
// the Meta status line and the log section are written here and rewritten by
// SetRequestStatus / SetTaskStatus, so keeping them in one file is what stops
// the writer and the reader from drifting apart.
//
// Everything a person has to answer is marked TODO. Nothing is guessed on their
// behalf: a filled-in field that was never actually decided is worse than an
// empty one, because the next reader cannot tell the difference.

// The log headings, named once. Both templates below write them and work.go
// appends transitions under them, so a section renumbered in one place and not
// the other is a status that silently stops being recorded.
const (
	requestLogHeading = "## 8) Log"
	taskLogHeading    = "## 4) Execution Log / Change Log"
)

const requestTemplate = `# {{.ID}} - {{.Title}}

## Meta

- Request ID: {{.ID}}
- Status: ` + "`{{.Status}}`" + `
- Raised: {{.Raised}}
- Priority: {{if .Priority}}{{.Priority}}{{else}}TODO{{end}}
- Audience: {{if .Audience}}{{.Audience}}{{else}}TODO - internal authenticated users, or public anonymous visitors{{end}}
- Target project: {{if .Project}}{{.Project}}{{else}}TODO - decided by the audience, see section 2{{end}}

## 1) What was asked for

{{.Title}}

TODO in the customer's own words, before it is translated into platform terms.
The translation is section 2's job, and keeping the original is what lets the
next reader check that the translation was right.

Everything you write here is 繁體中文 somebody reads months from now to work out
what was agreed - see ` + "`.agents/skills/plain-chinese/`" + `. Their words stay exactly as
they said them; yours are the ones the rules apply to.

## 2) Who is on the other end

TODO. This is the question that decides everything downstream, because the
entry point and the read path follow the audience and cannot be shared:

    internal, authenticated callers   -> the platform's agent hub, semantic layers
    public, anonymous visitors        -> your own BotProvider, fixed query tools

  - Same audience as an existing project -> this request goes into that project.
  - A new audience -> a new project: ` + "`asgard-cli project add <slug>`" + `.

Record the answer in Meta above as the target project.

## 3) Systems it has to read

One row per system, including the ones that sound obvious. "How to reach it" has
a preference order and it is not a matter of taste: a database we can read beats
an API, and an API beats a screen a person clicks. Take the highest available
per system.

| system | what it holds | how to reach it | who asks | what they want | read or write |
|---|---|---|---|---|---|

TODO. Ask as well whether the customer already has something that consolidated
these for them - a middleware layer, an OMS, a warehouse that pulls the channels
in. If they do, several rows collapse into one database, and that changes the
design more than any other single answer.

A row wanting **write** is not a project decision but a warning: the standing
architecture is read-only, and every write path needs its own spec and human
approval.

## 4) How each system is actually reached

One block per system in section 3 that we will connect to. This is the record of
what was agreed in the room: without it, the coordinates live in somebody's chat
history and the next person asks the customer the same questions again.

> **This file is committed to git. No password, token, key or connection string
> containing one goes in it - not "temporarily", not redacted-but-guessable.**
>
> A secret that reaches a commit is not fixed by deleting the line, because the
> history keeps it. It is fixed by rotating the credential, which means going
> back to the customer to ask for a new one. Write the *name* of the key and
> where it lives; never its value.

    coordinates  -> here, and in projects/<project>/chart/values-<env>.yaml
    passwords    -> .env locally (gitignored), app-secret in the cluster

### <system name>

- Kind: TODO - postgres, mysql, mssql, oracle, hana, netsuite, trino, athena,
  an HTTP API, or a web console with no API at all
- Host / base URL: TODO
- Port: TODO
- Database / schema / tenant: TODO
- Account: TODO - the username, and **whether it is read-only**. Ask for a
  read-only account explicitly; the one they offer first usually is not.
- Reachable from the cluster? TODO - yes, or it needs an IP allowlist, a VPN,
  or a bastion. This is the answer that most often turns a one-day integration
  into a three-week one, and it is free to ask on day one.
- Credential owner: TODO - the person or team who issues it, by name or role.
  A credential with no owner is not a dependency, it is a delay.
- Secret key name: TODO - the ` + "`.env`" + ` key (` + "`<TARGET>_DB_PASSWORD`" + `) and the
  ` + "`app-secret`" + ` key (` + "`<target>_db_password`" + `). The names, not the values.
- Confirmed working: TODO - the date somebody actually connected with it, and
  how. Credentials that were only ever pasted into a chat have not been tested,
  and an untested one fails at the least convenient moment.

Once a database is agreed, register it in three places or it will not work end
to end - ` + "`.env.example`" + ` at the repo root lists them:

    .env                             the real values, locally, never committed
    scripts/db/pgenv.py DB_TARGETS   so the introspection tooling can reach it
    chart/values-<env>.yaml          the non-secret coordinates the CR reads

and create the CR with ` + "`asgard-cli add dataconnector <name> --db-class <class>`" + `.

**An HTTP API instead of a database** needs the same block plus the auth scheme
(bearer, OAuth client credentials, mTLS), who holds the client id and secret,
whether there is a sandbox environment, and the rate limit. Ask for the rate
limit even when it sounds generous - it decides whether a Syncer can backfill.

## 5) Scope

The smallest version they said they would accept as proof (question 6b - their
answer, not ours):

- TODO

In scope, to deliver that:

- TODO

Out of scope for now, with what would have to be answered before it comes back:

| deferred | what has to be answered first |
|---|---|
| TODO | TODO |

Out of scope, full stop:

- TODO

## 6) Open questions

Anything this request cannot proceed without goes in ` + "`docs/open-questions.md`" + `
as well, one row each, with what it blocks and who can answer it:

    asgard-cli question add "<question>" --blocks {{.ID}} --ask "<who>"

It goes there rather than only here because a question buried in a spec
disappears when that spec reaches ` + "`done`" + `, and ` + "`asgard-cli next`" + ` reads that
file on every run.

## 7) Task specs

Written with ` + "`asgard-cli task add \"<title>\" --request {{.ID}}`" + `.

| Task ID | Title | Status |
|---|---|---|

` + requestLogHeading + `

- {{.Raised}} raised, status ` + "`{{.Status}}`" + `
`

const taskTemplate = `# {{.ID}} - {{.Title}}

## Meta

- Task ID: {{.ID}}
- Status: ` + "`{{.Status}}`" + `
- Created: {{.Created}}
- Project: {{if .Project}}{{.Project}}{{else}}TODO{{end}}
- Request: {{if .Request}}{{.Request}}{{else}}-{{end}}
- Complexity: {{if .Complexity}}{{.Complexity}}{{else}}TODO - S, M or L{{end}}
- Owner: {{if .Owner}}{{.Owner}}{{else}}TODO{{end}}
- Spec mode: spec
- Living spec module: docs/spec/{{.SpecSlug}}/TODO.md

## Writing this

繁體中文, to ` + "`.agents/skills/plain-chinese/`" + `. A task spec is read by whoever picks
this up, which is often not you, and 不僅……更是…… costs them the same second it
costs a customer.

## 1) Requirements

### Background

TODO. Read ` + "`docs/spec/{{.SpecSlug}}/`" + ` for the module this touches first: this
spec is a delta against what that module says the system does today.

### Goal

TODO.

### In scope

- TODO

### Out of scope

- TODO

### Known context

TODO - the schema facts, the credentials, the platform behaviour this depends on.
Introspect the real database rather than guessing it; the ` + "`semantic-layer-modeling`" + `
skill under ` + "`.agents/skills/`" + ` says how.

### Open questions

TODO, or none. Anything that blocks the work belongs in
` + "`docs/open-questions.md`" + ` too, because this section vanishes when the task
reaches ` + "`done`" + `.

### Acceptance criteria

EARS style, one ` + "`R#`" + ` each. Every ` + "`R#`" + ` has to map to an implementation task in
section 3 and to a verification entry in section 2, and the readiness gate is
exactly that mapping being complete.

- R1 WHEN TODO THE SYSTEM SHALL TODO

## 2) Design

### Resources and templates

TODO - which CRs, which files under ` + "`projects/{{if .Project}}{{.Project}}{{else}}<project>{{end}}/chart/app/templates/`" + `.
Generate each one with ` + "`asgard-cli add <kind> <name>`" + ` rather than by hand: the
parts that fail silently are the parts it gets right.

Say explicitly whether this adds a ` + "`SemanticLayer`" + ` or ` + "`DataConnector`" + `, widens
which cubes an agent may query, or introduces a write path. Those three point at
the customer's live systems, which is why they are reviewed before they are made.

### Data and secret dependencies

TODO. One ` + "`app-secret`" + ` per namespace, and ` + "`asgard_resource_api_key`" + ` is shared by
every CR that needs a platform credential.

### External API contracts

TODO, or none.

### Acceptance test matrix

| R# | how it is verified |
|---|---|

### Verification plan

    asgard-cli check {{if .Project}}{{.Project}}{{else}}<project>{{end}}
    helm lint projects/{{if .Project}}{{.Project}}{{else}}<project>{{end}}/chart/app
    helm lint projects/{{if .Project}}{{.Project}}{{else}}<project>{{end}}/chart/app -f projects/{{if .Project}}{{.Project}}{{else}}<project>{{end}}/chart/values-dev.yaml
    asgard-cli verify {{if .Project}}{{.Project}}{{else}}<project>{{end}}

## 3) Implementation Tasks

| # | task | covers |
|---|---|---|

` + taskLogHeading + `

- {{.Created}} spec created, status ` + "`{{.Status}}`" + `
`
