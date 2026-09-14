# Conventions

**Read the platform side first:** the pages in `../wiki/` - what the platform is made
of, and who each piece is for. These conventions are how this repo writes it
down; the wiki is what is being written down.

**Checked:** 2026-09-02: every naming prefix in the table below and the
`templates/<kind>/` layout, against four deployments.

**Unchecked:** the rest of the conventions. They are house style, and house style has no external source to check against.

Where every CR goes, what it is called, and what all of them need.
Read this once; the shape extracts assume it.

## How to reach a system, in order

Decide this per system before anything else. It determines the CRs, the effort,
and what the integration costs to maintain.

| route | what you get | shape |
|---|---|---|
| **a database you can read** | the agent composes its own queries and joins across tables. Questions nobody thought to expose an endpoint for still work | `../usecase/semantic-layer.md` |
| **an HTTP API** | a fixed set of calls, but a real contract - reviewable, and gateable for writes | `../usecase/external-api.md` |
| **any other protocol** | the agent runs the client itself in its sandbox: SNMP, SSH, a vendor CLI, a database client. **Not limited to HTTP** | `../usecase/external-api.md`, the sandbox half |
| **only a web UI** | last resort. Brittle, slow, and verifying the agent did the right thing is hard | `../usecase/browser-operation.md` |

**"No API" does not mean "browser".** The sandbox is a real environment: if a
client exists for the protocol, the agent can run it. A network device with SNMP
or SSH is reached that way, and it is far more reliable than driving its web
console. Check for a command-line route before concluding there is only a UI.

The gate rule does not move: **reads may happen in the sandbox, writes go
through a Toolset with `requestConsent`** whatever the protocol.

**A system offering both a database and an API: read from the database, write
through the API.**

Two questions that change the answer, and are worth asking before designing
anything:

- **Does something already consolidate these systems?** Middleware, an OMS, a
  warehouse that already pulls the channels in. If so, several integrations
  collapse into one database.
- **Does the vendor have an API nobody has asked for?** A web console is usually
  a client of one.

If nobody knows, that is a row in `docs/open-questions.md`, not an assumption.

## Where things go

The extracts each say where their own CR goes. This is the whole tree, because
"which directory does a Workflow live in" has no answer from one extract alone -
it depends on what the Workflow is for.

```
projects/<project>/chart/app/templates/
  agent/<name>.yaml            Agent CRs
  data_connector/<name>.yaml   DataConnector
  semantic_layer/<name>.yaml   SemanticLayer
  toolset/<name>.yaml          Toolset
  tool/<name>.yaml             a Workflow that backs one tool
  workflow/<name>.yaml         a Workflow that is not a tool - a conversation
                               loop, a Trigger's entrypoint, an action behind a
                               Toolset
  trigger/<name>.yaml          Trigger
  skill_set/<name>.yaml        SkillSet + its SourceSet + its Syncer, one file
  source_set/<name>.yaml       a SourceSet that is not a skill set's store
  plugin/<name>.yaml           Plugin + the SkillSet it bundles
  bot_provider/<name>.yaml     BotProvider
  sandbox_blueprint/<name>.yaml
  _helpers.tpl
  NOTES.txt
```

A self-hosted entry point is often grouped instead, one directory per entry
point, holding its BotProvider, Workflow and SandboxBlueprint together:

```
  supervisor/<name>/{bot_provider,workflow,sandbox_blueprint}.yaml
```

**One CR per file**, except where CRs have no independent life: a SkillSet with
its SourceSet and Syncer, or a Plugin with the SkillSet it bundles, go in one
file separated by `---`.

### Names

| kind | prefix | the rest |
|---|---|---|
| `Agent` | `ag-` | the system or role it covers |
| `SemanticLayer` | `sl-` | the source system |
| `DataConnector` | `dc-` | the source system, matching its layer |
| `Toolset` | `ts-` | what the set is for |
| `Workflow` | `wf-` | **`<verb>-<noun>`** for an action or a tool; the entry point's name for a conversation loop |
| `Trigger` | `tr-` | `<verb>-<noun>`, matching its entrypoint workflow |
| `SkillSet` | `sk-` | the domain |
| `SourceSet` | `ss-` | what it stores |
| `Syncer` | `syn-` | what it fills |
| `BotProvider` | `bp-` | the entry point |
| `SandboxBlueprint` | `sbp-` | the entry point, matching its BotProvider |
| `Plugin` | `pg-` | the bundle |

A Workflow that backs a tool and the tool's own name are **different**: the CR is
`wf-search-products`, the `tooling.name` the model sees is `search_products`.
Snake case there, kebab case in the CR name.

Where several CRs form one chain, **repeat the same suffix across them** so the
chain is visible in a file listing: `bp-website` / `wf-website` / `sbp-website`.

## What every CR needs

Three things, on every CR, regardless of kind.

```yaml
metadata:
  name: <prefix>-<name>
  annotations:
    # The kind's own display annotation. Without it the CR applies cleanly and
    # then appears NAMELESS in the platform UI - and nothing in lint, CRD
    # validation or a server dry-run reveals that.
    asgard-ai.com/<kind>-name: "<display name in the customer's language>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
```

The annotation key is the kind in kebab case, and **fifteen kinds are
enforced**: `agent-name`, `semantic-layer-name`, `data-connector-name`,
`skill-set-name`, `source-set-name`, `syncer-name`, `toolset-name`,
`workflow-name`, `trigger-name`, `bot-provider-name`,
`sandbox-blueprint-name`, `knowledge-base-name`, `loader-name`,
`completion-model-name` and `plugin-name`. Several kinds also take a
`-description`.

**The one exemption is a SkillSet a Plugin bundles**, which is not presented in
the UI at all and so has no name to be missing - `../usecase/plugin.md` is the
shape.

### `asgard-ai.com/product`, which every generated chart stamps

Every CR `asgard-cli project add` writes carries
`asgard-ai.com/product: platform`, from the chart's own `<chart>.labels`
helper. **It is not a display annotation and not one of the fifteen above**, so
nothing in the gate asks for it and a chart without it deploys.

What it does is name the product that owns the resource, as the IAM product
code, and **asgard-core echoes it into every audit-log event so the Console can
filter by product** - it is the dimension an audit event is filed under. The three IAM product codes are `product: agent-hub`, `product: data-insight`
and `product: platform`, and **unlabelled means no product** rather than a default - a resource with no
label is one the Console's product filter will not find.

`platform` is right for a chart's own CRs: they are applied by the pipeline
rather than created inside a product. The other two are stamped by the platform
itself - workflow-service on a Flow Agent, the preset reconcilers on the
per-namespace preset BotProviders - so a chart never writes them.

**Do not drop it from a migrated chart because a grep of this material came
back empty.** It did, and the label read as retired; it is current, and what it
costs to remove is a gap in the audit trail that nothing reports.

`asgard-cli verify` enforces the display annotations. **The authority for every
platform key is the workflow-service source, `internal/shared.go`** - read it
rather than inferring a convention from another chart. Anything not on the
checker's list is invisible to every check, so a new UI label is a check to add,
not just a line to write.

## Values, and what never goes in them

Connection coordinates, endpoints, schedules and suspend switches go in
the platform's variables, one group per thing, declared as `chartValues`. `values.yaml` declares a default
for **every** `.Values.*` the chart itself owns - the lint `asgard-cli gate` runs,
which supplies only the reserved `asgard` block and no environment file,
is the only step that proves it, and without it a missing default is masked in
the gate and nil-pointers for anyone running plain `helm template`.

**Credentials never go in values, and neither does the name.** A chart writes a
`secretKeyRef` and reads the object's name through its `appSecretName` helper:
each Release has its own Secret, created and maintained by the Platform, whose
name arrives as `.Values.asgard.appSecretName`. A literal name in a template
points at an object that is not in that namespace, and nothing catches it - the
CR is valid, the dry run passes, apply succeeds, and it fails at runtime.

`asgard_resource_api_key` is the **conventional** name for a platform resource
credential. `SourceSet.apiKey`, `Toolset.apiKey` and `BotProvider.adminApiKey`
are all of that kind, and `asgard-cli add` points them at that one key. Whether
they share one credential or need separate ones is a requirement about rotation
scope, not a rule this page can state: the Platform reads whatever
`secretKeyRef.key` says and never the name itself.

**Do not declare a `secretKeyRef` for a key that does not exist yet.** Config
evaluation fails at call time, not at apply time, so the chart deploys and the
CR breaks the first time it is used.

## Two field shapes that are easy to get wrong

**`ValueExprTemplate`** takes either a static `value:` or an `expression:` of
JavaScript evaluated per turn, with the caller's payload as `prevPayload`.

On a `SandboxBlueprint`, the `*Names` fields are **comma-separated strings**, not
lists, and `sourceSetMounts` is a **JSON string** the controller unmarshals.

```yaml
  toolsetNames:
    value: "ts-a,ts-b"
  sourceSetMounts:
    value: '[{"sourceSetName": "ss-x", "mountPath": "/knowledge", "readOnly": true}]'
```

**Every config value is a string**, including booleans: `parseJson` is `"true"`,
a suspend label is `"true"`, not `true`.

## Where a sandbox mount may go, and where it kills the agent

Four rules, and the first one has taken a deployment's every conversation down.

**Never mount anything under the agent's `~/.claude`.** kubelet creates a
mount's missing **parent** directory as `root:root` before the container starts,
the sandbox runs as uid 1000, and `~/.claude` is the one directory the CLI driver
**must** write - its appended-system-prompt seed file lands there. So a
read-only mount at `~/.claude/anything` makes the driver die at start-up, and
the symptom looks nothing like a mount: the conversation fails with the driver
never coming ready. The platform solves this for its own nested mounts by giving
the parent (`<home>/local-plugin`) its own fsGroup-owned emptyDir, and **there
is no such volume for `~/.claude`**.

**The sandbox working directory is where it does work.** `/work` is the
sandbox's home for the channel, `/work/.claude` does not exist on a fresh
sandbox, and a project-level skills directory under it is discovered the same way
- so `/work/.claude/skills/<name>` is a mount that both lands and loads.

**`sourceSetMounts` is the only mechanism that takes a `subPath`.** A SkillSet
always mounts its **whole** SourceSet - the controller joins the volume path with
an empty member sub - and its `searchPaths` are static. So a SkillSet cannot
expose part of a store, and using one to carry per-customer or per-brand files
puts **every** tenant's files in **every** sandbox. Per-tenant isolation is
`sourceSetMounts` with a `subPath`, or it does not exist.

**Skill discovery has two chains and they do not meet.** A SkillSet reaches the
agent through the environment's skill-set JSON and the runtime plugin allowlist;
a `sourceSetMounts` mount is on neither, and is found only by being a directory
inside a skills directory. Mounted anywhere else it is a readable file nothing
loads, which reads as "the skill did not work" rather than as a mount in the
wrong place.

### What the CRD checks here, and what it does not

| field | validated as | what that leaves to you |
|---|---|---|
| `sourceSetMounts[].mountPath` | `^/.+` | **nothing else.** No traversal check, no character rule. A mountPath built from a name a customer chose is a path-traversal surface, and two entries resolving to the same mountPath are deduped **first-wins**, silently dropping one |
| `sourceSetMounts[].subPath` | rejected if absolute, or containing `.`, `..` or a double slash | a rejected one takes the **whole Sandbox** not-ready with `InvalidSourceSetMountSubPath`, which stops every conversation - far worse than skipping one entry. So validate in the expression and **skip** a bad entry rather than repairing it |
| `credentialMounts[].mountPath` | `^/.+`, and it is a **directory** | the token lands in it as `access_token`. A single-file mount would need a `subPath`, and kubelet never refreshes those - the token would be a one-time copy taken at pod start, which is the whole thing this field exists to avoid |

**Checked:** 2026-09-11 against asgard-core `623ceb5` - its
asgard-core `internal/bpoperator/reconciler/sb_reconciler.go` for the mount
construction, the subPath validation and the local-plugin emptyDir, and
asgard-core `internal/processor/clidriver/options.go` for the seed file the
driver writes - and against asgard-kube `cbd8d70`, its asgard-kube
`pkg/apis/asgard/v1alpha1/types.go` for the three patterns and the directory
rule. The `~/.claude` failure was reported by a
deployment that had hit it in production, and the mechanism was then re-read in
the platform source rather than taken from the report.
