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

## 4) Scope

In scope:

- TODO

Out of scope:

- TODO

## 5) Open questions

Anything this request cannot proceed without goes in ` + "`docs/open-questions.md`" + `
as well, one row each, with what it blocks and who can answer it:

    asgard-cli question add "<question>" --blocks {{.ID}} --ask "<who>"

It goes there rather than only here because a question buried in a spec
disappears when that spec reaches ` + "`done`" + `, and ` + "`asgard-cli next`" + ` reads that
file on every run.

## 6) Task specs

Written with ` + "`asgard-cli task add \"<title>\" --request {{.ID}}`" + `.

| Task ID | Title | Status |
|---|---|---|

## 7) Log

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

## 4) Execution Log / Change Log

- {{.Created}} spec created, status ` + "`{{.Status}}`" + `
`
