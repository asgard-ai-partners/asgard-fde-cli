# TASK.md

Working notes for `asgard-fde-cli`. Rewrite when a new task starts - except
"What the engagement knows" and "What using it has caught", which are worth
carrying forward regardless: each line there cost somebody a real investigation.

## Status

The whole flow works and has been walked end to end against a real filesystem
twice: once as an FDE using it, once as an agent with no context and nothing to go
on but what the commands print. Both audience shapes reach a green four-step gate,
and a chart that also satisfies CD.

```
asgard-cli init / scaffold / project add          the workspace and the skeleton
            request / task / question / decision  the records, dated and status-tracked
            next                                  which stage, and what it needs
            add <kind> <name>                     CR skeletons, correct where failure is silent
            usecase [name]                        how each deployment shape is built
            check                                 repository structure
            render <project> <env>                what CD will apply
            verify [project]                      the invariants a render cannot see
            doctor                                the external tools, and how to install one
```

| package | holds |
|---|---|
| `internal/config` | `.asgard-config.json`: the workspace, its projects, `olapOnlyLayers` |
| `internal/scaffold` | the non-customer-specific tree, `go:embed`, `<< >>` delimiters so Helm's `{{ }}` survives |
| `internal/stage` | ten stages plus idle; the stage is derived from the repo, never stored |
| `internal/work` | the repo's own records: requests, task specs, open questions, decisions |
| `internal/generate` | CR skeletons for eleven kinds, wired to what the chart already declares |
| `internal/chart` | reads a project's **unrendered** templates for (kind, name) |
| `internal/deploy` | `projects/<p>/deploy.yaml`, the source of truth for deploy targets |
| `internal/render` | renders via `helm template`, the way CD does |
| `internal/gate` | xref, agent split, deployability |
| `internal/check` | repository structure (was `check_repo_consistency.py`) |
| `internal/tool` | resolves helm/kubectl/python3 and says how to install one |
| `internal/usecase` | sixteen extracts from deployments already in production, plus `conventions`. Every one is reachable from a generator or a prompt, and a test fails the build otherwise |

## Goal

When an FDE onboards a customer, the coding agent working in that customer's repo
needs enough context to do the work. The end state is a repo shaped like
[`unitech-e-asgard-kube`](https://github.com/asgard-ai-platform/unitech-e-asgard-kube).

**That repo is the output. The CLI's job is to emit the prompts that drive an
agent to produce it, plus the parts identical for every customer.** It does not
write the customer's Kubernetes resources, because roughly 90% of that repo's
916-line `AGENTS.md` is knowledge only the engagement can earn.

## The finding that shapes everything: onboarding is not linear

Of the target repo's 13 task specs, **three were superseded and one reverted** -
and they are the three decisions that matter most:

| decision | first answer | final answer | why it flipped |
|---|---|---|---|
| how `website` reads catalog data | a SemanticLayer (TASK-003) | five zero-parameter query tools (TASK-010) | a layer without `allowedCubes` is arbitrary SQL over every cube, and the exposed surface grows by itself each time a table is added |
| how `website` is reached | the platform's agent hub (TASK-007) | Flow Agent: own BotProvider -> Workflow -> SandboxBlueprint (TASK-011) | an anonymous visitor cannot authenticate to the agent hub, and `BotProvider.entrypoint` takes a Workflow, never an Agent |
| how `website` stores knowledge | `KnowledgeBase` + `Loader` + Toolset (TASK-012) | `SourceSet` Drive + contextIndex + read-only mount (TASK-013) | the KnowledgeBase mechanism is deprecated platform-side |

**So a linear wizard is the wrong shape.** Each was answered once, implemented,
and reversed after contact with reality. What the next customer needs is not a
script that walks them into the same first answer - it is the question with the
trap already visible.

That is the CLI's highest-value output: **stages 4, 5 and 6 print the wrong
answer next to the right one**, because the wrong one is what looks obvious.

## The unit of work is a request, not a stage

State lives in the customer's repository, never in the CLI, because the
repository is what the next agent opens.

| record | file | carries |
|---|---|---|
| request | `requirements/requests/REQ-xxx-*.md` + registry row | raised date, status, target project, the customer's own wording |
| task spec | `requirements/tasks/TASK-xxx-*.md` + queue row | created date, status, the SDD sections, a dated log of every transition |
| open question | a row in `docs/open-questions.md` | raised date, what it blocks, who can answer |
| decision | `docs/decisions/YYYY-MM-DD-*.md` + traceability row | the date in the file name, the module it changed |

Three things follow, each removing a hand edit:

1. **A stage is a property of a request.** `Current()` returns `Projects` when an
   open request has no target project - not when `projects/` is empty. A repo
   with no projects and no request is not waiting to be interviewed; it is
   waiting for a requirement, and the answer is to ask for one.
2. **A status transition moves three or four places at once.** A task's status is
   in the queue table, the spec's `Meta`, and the spec's log; a request's target
   project is in the registry's Spec column and its `Meta`. The commands write all
   of them and stamp the date.
3. **`next` has an idle state.** Nothing open and every chart complete -> it does
   not print a stage. It reports what the repo holds and asks the one question
   that can move the work on. `Verify` is reached only when something *is* open,
   because a deployed repo and one waiting for its first tag are identical on
   disk.

`decision add` reads `docs/decisions/_decision-template.md` **from the customer
repo**, not an embedded copy, so an engagement that edited the template keeps its
version. It links the record from the living spec's traceability table, which the
scaffold marks with an ASCII `<!-- traceability -->` anchor: the generated
documents are in the customer's language, and an anchor the CLI must match should
not depend on which language that is.

## The stages

| # | stage | CLI writes | prompt asks the agent to |
|---|---|---|---|
| 0 | init | - | get the workspace id; run `init` in the directory that becomes the workspace |
| 1 | scaffold | the whole non-customer-specific tree | - |
| 2 | projects | `projects/<slug>/` + chart skeleton, README row | interview: inventory the systems and audiences, decide the split, record the decision |
| 3 | data sources | - | fill `.env`, introspect the real databases, write `DataConnector` CRs |
| 4 | read path | - | SemanticLayer, or fixed query tools when the audience is public. **Carries the TASK-010 reversal** |
| 5 | entry point | - | pick the shape from the audience. **Carries the TASK-011 reversal** |
| 6 | knowledge | - | SourceSet Drive + contextIndex. **Carries the TASK-013 reversal** |
| 7 | verify | - | the four-step gate |
| 8 | deploy | - | tf-asgard namespace + app-secret first, collect `platformMainEnvironmentId`, then tag |
| 9 | enhance | - | the loop for a repo already live; not part of the linear walk |
| - | idle | - | nothing open: what does the customer want next |

**Stage is derived from the repo, not stored.** Whether `projects/` has entries,
whether a chart has a `DataConnector`, whether a request names a project - all
facts on disk. A counter in the config would be a second source of truth that goes
stale the moment someone does the work by hand.

## What the target repo looks like

```
AGENTS.md                     single source of truth for coding agents (916 lines)
CLAUDE.md                     one line: @AGENTS.md
README.md                     project registry table, deployment status, gates
projects/<project>/
  deploy.yaml                 which envs deploy, namespace + values per env
  chart/app/templates/<cr>/   one directory per CR kind, one file per CR
  chart/values-{dev,prod}.yaml
common/
  values-{dev,prod}.yaml      cross-project per-env values
  skills/<skill>/SKILL.md     RUNTIME skills, synced into the platform
docs/
  README.md                   index + the four-layer model
  spec/<slug>/                living spec: what the system does now
  decisions/                  dated, immutable
  meeting-notes/              raw input before it converges
  open-questions.md           what nobody has answered yet
requirements/{requests,tasks}/ executable task specs, draft -> ready -> in-progress -> done
references/                   background material, not implementable
scripts/                      check_crd_fidelity.py, db/ introspection tools
.agents/skills/<skill>/       DESIGN-TIME skills, for the coding agent in the repo
.github/workflows/main.yaml   tag-driven CD
```

Structural facts a generator has to respect:

1. **A project is the unit.** One project = one Helm chart = one K8s namespace.
   Adding one means `projects/<p>/chart` + `projects/<p>/deploy.yaml` + a row in
   the root README table, and `asgard-cli check` enforces that the table matches
   the folders on disk.
2. **Envs are exactly `dev` and `prod`**, and the two may be asymmetric. A project
   deploys to an env only if its `deploy.yaml` declares it.
3. **Namespace convention**: `asgard-<customer>-<project-slug>-<env>`. Slugs are
   short deliberately: names derived from a namespace inherit its length, and
   Kubernetes caps it at 63 characters.
4. **Two kinds of skills, never mixed.** `.agents/skills/` is design time (read by
   the coding agent from the working directory); `common/skills/` is runtime
   (synced into the cluster and bound by `SkillSet`).
5. **Documentation is four layers separated by tense**: meeting-notes (raw) ->
   decisions (dated, immutable) -> spec (living, present tense) -> tasks (delta).
6. **Terraform comes first.** A namespace and its `app-secret` are created by
   tf-asgard before a project can declare that env.
7. **The audience decides the split**, not the system. Two audiences need different
   entry points and different read paths, and those cannot be shared.

## What the engagement knows, and the CLI cannot

The reason the CLI emits prompts rather than resources. No template produces
these:

- every NetSuite receipt line has a twin line with the quantity negated, so
  `SUM(quantity)` without `isinventoryaffecting = 'T'` returns 0
- `calltimeview` must never be queried: a windowed view where no filter pushes
  down, 1m51s versus 2.1s for the same answer
- `call_status_logs` is ~90% duplicate rows and fans out a join ~30x
- the safety stock level comes only from Location 608, never the row's own

## What is the same for every customer, and is written as files

The platform contract - the things that have to be right character by character,
so regenerating them from a prompt would drift:

- the four-layer docs model and the SDD status flow
- the gate: `asgard-cli check` / `verify`, and `scripts/check_crd_fidelity.py`
- the design-time skills under `.agents/skills/`
- tag-driven CD (`dev-x.y.z` -> dev, `x.y.z` -> prod), matrix built from
  `projects/*/deploy.yaml`
- **display annotations**: every listed CR needs `asgard-ai.com/<kind>-name` or it
  shows up nameless in the UI, and nothing in lint or dry-run reveals it
- the workflow-set label set, and the two extra labels a hand-written `Trigger`
  needs or its editor opens as a blank canvas
- `platformMainEnvironmentId` is per project per env
- one `app-secret` per namespace, one `asgard_resource_api_key` shared by every CR
  that needs a platform credential
- **a field the CRD does not declare is silently pruned**, and server dry-run
  still reports success; only helm's server-side apply fails

`AGENTS.md` is both: the platform-contract half is written by the CLI, and the
hard-won half is left as marked sections the agent fills as stages 3-6 discover
things.

## The two ordering traps

Both fail during CD rather than locally, which is far too late. `project add`
prints them, and `asgard-cli verify` warns on the second:

1. **Terraform first.** A namespace and its `app-secret` must exist before a
   project declares that env, or the next tag fails at `helm upgrade`.
2. **CD requires at least one `Syncer` per deployed project.** The trigger step
   polls for CronJobs labelled `asgard-ai.com/syncer-name` and **exits 1 after
   180s if it finds none**, even when `helm upgrade` succeeded. The CD workflow's
   own comment calls zero Syncers a configuration error. A Syncer comes from a
   `SkillSet` or a knowledge drive.

## Config schema

```json
{
  "workspace": {
    "id": "<a ~19-digit number, issued by the platform>",
    "slug": "acme",
    "name": "acme"
  },
  "projects": [
    { "slug": "internal", "name": "internal", "environments": ["dev", "prod"] },
    { "slug": "website",  "name": "website",  "environments": ["dev"] }
  ],
  "olapOnlyLayers": []
}
```

`slug` is load bearing. Two names are **derived** from it rather than stored, so
the config cannot drift from the layout on disk:

- repo name: `<workspace.slug>-asgard-kube`
- namespace: `asgard-<workspace.slug>-<project.slug>-<env>`

`workspace.id` is a **numeric snowflake, roughly 19 digits, not a UUID**. It is
issued by the platform and the FDE holds it when `init` runs, so it stays
required; the format is not validated, because an API to verify it is coming and
one example is not a pattern.

It sits at a different level from `platformMainEnvironmentId`, which **is** a UUID
and is per project per env. The platform hierarchy is workspace -> project ->
environment, and the repo layout mirrors it.

`olapOnlyLayers` are semantic layers deliberately bound to no Agent, because they
feed Data Insight rather than a chat agent. It lives in the config rather than as
a flag, because a policy that applies only when somebody remembers an argument is
not a policy.

## Nothing shipped names another customer

A rule, not an accident, and three sets of tests enforce it:

- `internal/usecase/extracts/` and the stage prompts **name no customer and no
  deployment**. `source/SOURCES.md` holds the attribution and lives outside
  `internal/` so it cannot be embedded even by accident.
- `TestExtractsNameNoCustomer` and `TestWriteCarriesNoCustomerContent` fail the
  build on any reference deployment's name in shipped content.
- Four `secrets_test.go` files fail on any run of 12+ digits in an embedded file.

Two leaks of this class have been found and fixed. A live workspace id had shipped
in a stage prompt as an example - and prompts go to every engagement, so it was
published to all of them, and the natural thing for the next reader to do with an
example id is paste it, binding their repository to another customer's workspace.
The second was `common/README.md` recommending two other engagements' repository
layouts, in a paragraph that also hardcoded "exactly two projects" from the repo
it was copied out of; the guard had listed only the primary source repo until
then.

**So the CLI recommends no reference repo to the agent.** The extracts are the
generalised substitute, and that is deliberate - see the next section.

## The design rule the prior art settles

[`asgard-industry-demo-generator`](https://github.com/asgard-ai-platform/asgard-industry-demo-generator)
already solves the prompt/context problem, as a Claude Code plugin of commands and
skills. Its `minting-asgard-chart/SKILL.md` states the rule outright:

> 欄位規格的**完整定義在 GOAL.md §5**;以下為行動流程提示,**不複述欄位表**。

| where | holds |
|---|---|
| the repo's own SoT (`AGENTS.md`) | the specification: field tables, naming, annotation contracts |
| a skill | **how to act**, and a pointer back to the spec. It never restates the field table |
| a command | a staged procedure with hard gates |

**One fact, one home.** A skill that copied the field table would be a second copy
that goes stale, and the reader cannot tell which is current.

Where our situation differs: their skills lean on "比照既有產業 - 7 個產業結構相同",
and **their reference examples live in the repo** - eight complete industries. A
new customer's repo has none of that, and pointing an agent at another customer's
repository is not an option. Filling that gap is what `asgard-cli usecase` is for,
and why the extracts are context on demand rather than files in the repo: an
unused example in a repo becomes stale boilerplate, while a stale example inside
the CLI is one release away from being fixed for every customer at once.

What a CR example has to carry: in the real charts **the header comments are three
times longer than the YAML, and they carry the knowledge** - why fixed query tools
instead of a SemanticLayer, why all five take zero parameters, why seven tools
became five, why `spec.instruction` must not come back. An example stripped of
those degrades into a restatement of the CRD schema, which an agent can already
read from `asgard-kube`. **The comments are the product.**

## Reference deployments

**Internal only.** Cloned under `~/projects/asgard/`. `source/SOURCES.md` is the
one file allowed to name a deployment, and holds three things this one does not:
how each extract refers to a repo without naming it, which file each kind's
material came from, and the **generational conflicts** - two charts disagreeing in
a way that is dated rather than a matter of taste.

| repo | shape it demonstrates |
|---|---|
| unitech-e | agent hub (5 agents / 6 semantic layers) **and** a single-agent flow agent; trigger; knowledge drive. **Most recently maintained - it wins a generational conflict** |
| xxentria | supervisor + 9 agents |
| finance-ai | supervisor, 3 semantic layers |
| asgard-industry-demo-generator | 12 industries; the read/write governance split; a Claude Code plugin of commands + skills |
| buy123 | the minimal flow agent - no Agent CR at all; the whole read path is two HTTP APIs |
| asgard-freyr-kube | supervisor + 5 subagents; `agents.expression`; sandbox hooks; `ts-ui-event`; `tenants/` layout |
| asgard-auto-post | **28 Plugin CRs**; knowledge bases; api workflows |
| asgard-freyr-skills | **runtime skills as their own repo**, and the only source for browser operation |

**Three were listed as "not yet cited" for too long** - buy123, auto-post and the
demo generator - and being uncited is not the same as being read. Two rounds of
reading them produced `chat-channel`, `api-oauth` and `workflow-chain`, and found
that three mechanisms every shape is built on had no extract at all: the token
chain, the `relationships` graph, and what crosses between processors. Still
uncovered: auto-post's **28 Plugin CRs**, and the `router` shape as used in its
nine-branch content pipeline.

## Browser operation: the material exists, the extract does not

Nothing in the kube repos teaches how a web-only system becomes an agent
capability. **`asgard-freyr-skills` does**, and it is the only source for a system
with neither a database nor an API - the gap an IoT scenario walked into, where a
network device with only a web console had no shape to reach for.

It is a three-repo contract: skills versioned in `asgard-freyr-skills`, API
contracts vendored from `asgard-freyr-api` by a sync script (generated pages never
hand-edited), and the kube repo injecting `api_base_url` and the user's JWT
through a config file before loading the skills through a SkillSet.

Its worked example is **not a recording of clicks** - it is reference documents an
agent produced by exploring the system: a page map (88 L1 pages), an operation map
(everything the menu cannot see), an inventory of 14 cross-domain iframe origins,
and a four-level self-exploration discipline.

The post-mortem (TASK-006) rejected the obvious single cause and found three
independent mechanisms, each a real trap:

- **A1 - extraction is shaped like controls.** It collected tabs, fields, buttons
  and column names, and **discarded prose that names deeper objects**.
- **A2 - crawling follows links, and an editor entry is not a link.** The button
  had no `href` and no accessible name, so a link-following crawl **structurally
  cannot** reach the page builder behind it.
- **A3 - "cannot read the iframe" was treated as "cannot cover it".** The 14 inner
  origins had been recorded all along; nobody tried navigating to them directly,
  so 14 pages went uncovered.

The discipline that matters most is **never guess a route** - forbidden at every
one of the four levels. A route may come only from a link visible in the current
UI, or one already recorded and measured:

> 猜錯而被導回首頁只是運氣好;猜中一個**語意相近但不同**的頁面,然後在上面操作 ——
> 那是本 skill 最難察覺的失敗模式。

The same reasoning appears in `ts-ui-event` from the kube side, pinning the model
with schema rather than prose: every legal path is an `enum` in the tool's
`inputSchema`.

## Open questions

### Are the EKS cluster names customer-specific?

`.github/workflows/main.yaml` hardcodes `asgard-ai-eks` (dev) and
`asgard-ai-eks-prod` (prod), and `asgard-cr-verification/SKILL.md` carries the
full ARN including Asgard's AWS account id, which ships into every customer repo.

These read as **Asgard's own clusters, shared by every customer**, with customers
separated by namespace rather than cluster, which would make the CD workflow
portable verbatim. That is inferred from the tag-to-cluster table in `AGENTS.md`,
not verified. If it is wrong the cluster name becomes a fourth thing the scaffold
must parameterise - and the account id becomes something to parameterise either
way. It is allowlisted in `internal/scaffold/secrets_test.go` with that reasoning
written next to it.

### Should the helm major version be pinned?

Homebrew and Scoop both ship **Helm 4** now (4.2.4 as of 2026-09-01), while the
generated repo was written for 3. `template` and `lint` are what the CLI uses and
both still exist, so `asgard-cli doctor` prints a note rather than failing. CI
picks its own helm version independently, so the two can differ silently.

### Browser operation, before it can be an extract

Whether the page and operation maps were produced with tooling or by hand, and how
long one takes. Both decide whether an extract is advice or a commitment.

### Platform unknowns

Five are carried in the scaffolded `docs/open-questions.md` (P1-P5) rather than
here, because every engagement hits them: per-user resource scoping, an audit
record of what an agent did, whether approved content can change before it goes
out, non-HTTP protocol reach (answered: yes, via the sandbox), and the cost of
producing a web console's page map.

## Non-goals

- **Anything to do with git.** Not `git init`, not adding a remote, not
  authenticating to one: Asgard is growing its own mechanism for provisioning a
  customer repository. This needed saying **in the prompts**, not just here - the
  CLI never printed a word about git, but an agent that finds the directory is not
  a repository and reads a tag-driven CD workflow offers to set one up on its own
  every time. Stages 1 and 8 tell it not to, and why.
- **helm and kubectl as packaged dependencies.** Nothing about how this is
  distributed can install them: a tar.gz, a zip, a dmg and `go install` carry no
  dependency metadata and never can, and a Homebrew or Scoop dependency would only
  cover people installing that way - nobody, until those repositories exist. They
  are prerequisites, reported by `asgard-cli doctor` with the install line for the
  current machine.
- **Bundling helm or kubectl into the release.** Four platform/arch combinations
  at ~50MB each, and **kubectl has to stay within one minor of the cluster's API
  server**, so a pinned copy goes stale and is worse than none.
- **Validating `workspace.id` against the platform.** Waiting on the API.
- **Reading or writing `platformMainEnvironmentId`.** Per project per env, and it
  only exists after tf-asgard has created the namespace, so it belongs to the
  generated repo's values files.
- **Any command that talks to a cluster or the platform.** The CLI stays offline.

## What using it has caught

Each of these was invisible to every check that existed, and **why it was
invisible is the reusable part**.

| defect | why nothing caught it |
|---|---|
| **The generator wrote skeletons that could not pass the gate.** `add agent` hardcoded `skillSetNames: [sk-base]`, which nothing creates and which `add skillset` cannot create without a git URL the agent does not have, and it ignored the `SemanticLayer` in the same chart. The flowagent blueprint had the same hardcoded name. | The generator never read the chart it was writing into. Fixed by `generate.Resolve`: one layer is mounted (the choice is forced by "one system, one layer, one Agent"), several are refused with their names, a SkillSet is referenced only if one exists. |
| **Two CRs with one name rendered clean.** Several fixed query tools belong to one Toolset - the normal shape - and each `add querytool` emitted its own copy of that Toolset. Kubernetes keeps whichever applies last, so a tool silently vanished. | helm renders both, apply reports success twice. Now `verify` rejects a duplicate `kind/name`, and `Resolve` stops the second tool re-emitting the set. |
| **A project with no Syncer passed the gate and would fail CD.** | CD's requirements are not the apiserver's. Now `gate.Deployability` warns - a warning, not a failure, because zero Syncers is correct until skills are added and failing would leave the gate red through the middle of every engagement. |
| **The gate said what was wrong and not what to do.** `no SkillSet/sk-base` has an obvious cheap resolution - delete the reference - and it is the wrong one. | Written for a reader who already knew the design. Every dangling reference now names the `asgard-cli add` command, with the real project in it so it runs as printed. |
| **The generated `common/README.md` taught a file deleted two rounds earlier**, and recommended two other customers' repository layouts. | It was the one file excluded from that rename and never caught up, and prose compiles under no test. Now `TestGeneratedDocsOnlyNameFilesThatExist` fails on any `common/` or `scripts/` path the scaffold does not write, and the no-customer guard lists every reference deployment. |
| **`gate.Read` failed on a non-mapping YAML document** where the Python original skipped it. | A captured stderr line would have failed the gate for the wrong reason - the exact case the Python tolerance existed for. |
| **`next --stage` was silently ignored with no config.** | A flag the code does not read, which `AGENTS.md` forbids. |
| **A live workspace id shipped in a stage prompt as an example.** | It is just a number, and prompts are prose. |
| **`add <kind> <name>` doubled the kind prefix**, writing `metadata.name: dc-dc-erp`. | helm lint and a server dry-run both accept it; only the xref gate catches it, and only once something references the name that was meant. |
| **`check_agent_split.py` crashed with a raw `ModuleNotFoundError`** when PyYAML was missing, while its siblings printed a friendly line. | Since ported to Go, so the dependency is gone entirely. |
| **Nothing supported a chat platform.** Every extract and the BotProvider template hardcoded `botProviderClass: generic`, so a customer wanting LINE - the most likely channel for a Taiwanese customer - had nothing. | Three reference deployments were listed as "not yet cited" and had not been read. The class vocabulary (`generic \| line \| telegram \| discord \| slack`) was only in the platform's CRD contract, which no extract drew from. Fixed by `usecase chat-channel` and `add flowagent --bot-class`. |
| **`expression` is documented as CEL and is JavaScript.** The platform's CRD documentation says CEL; every chart uses arrow functions, `const`, `String()`, `encodeURIComponent` and `??`, none of which CEL has. | Nobody had read the docs and the charts against each other. An agent handed "it is CEL" writes something that cannot work and has no way to find out why. Recorded in `source/SOURCES.md` as a docs-versus-reality conflict, and corrected in `workflow-chain`. |
| **Three mechanisms every shape is built on had no extract**: the two-call token chain, the `relationships` graph, and what crosses between processors (`prevPayload` being replaced by an `http-request`, `httpResponse` meaning "most recent"). | The extracts were organised by deployment shape, so a mechanism used by all of them belonged to none. |
| **Five extracts were reachable from nothing** - `browser-operation`, `write-path` and the three new ones - so the only way to find one was to browse a list of sixteen for a mechanism you did not know you needed. | Writing material and pointing at it are separate acts, and only the first had a test. `TestEveryExtractIsReachable` now fails the build when a generator kind or a stage prompt does not send the reader to an extract, verified against a deliberately orphaned one. |
| **A second CR sharing a values block silently got no entry.** `botProviders:` exists after the first flow agent, so the second one's whole snippet was skipped as "already there" and its template read `.Values.botProviders.<name>.disabled` off a map with no such entry. | `appendValues` compared only the top-level key. **Bare `helm lint` catches it** - which is exactly what the bare pass is for - but only after the file is on disk, and an env file overlaid masks it entirely. |

Two lessons that are mine rather than the code's:

- **Say a mechanism cannot deliver the guarantee before building it.** Two rounds
  went into packaging metadata that was inert on every install path this project
  actually has.
- **"Not deployable" and "not functional" are different claims.** The generated
  chart applies cleanly; it is empty. Using the wrong word hid the real
  deployability gap (the Syncer) for a round.

Verified against `unitech-e-asgard-kube`: its `internal` project declares the same
CR kinds the CLI generates and carries **zero TODOs**. Same skeleton, content
filled in during the engagement. So the structure is right, and what remains is
content - which is what the staged prompts are for.

## What is not done

Ordered by what an engagement would hit first. Everything here is known, not
discovered - a line being here means somebody decided it could wait, and the
reason is next to it.

### Reference material still missing

- **auto-post's 28 `Plugin` CRs have no extract.** `plugin` covers one; nothing
  covers the shape at that scale, or what governs which plugin an agent may
  reach.
- **`router` at scale.** `workflow-chain` gives the mechanism and a two-branch
  example. auto-post's content pipeline has nine branches across several
  workflows, and the question that shape answers - when a branch belongs in the
  graph rather than in the prompt - is exactly the one an FDE gets wrong.
- **The `execute-script`, `validate-payload`, `generate-embedding` and
  `retrieve-knowledge` processors** appear in the platform's contract and in no
  chart that has been read. Either they are unused in practice, which is worth
  knowing, or the sample of deployments read so far is too small.

### Gates that could be stronger

- **`project add`'s terraform prerequisite is a notice, not a gate.** The demo
  generator's equivalent command *refuses* to continue without the namespace and
  both `platformMainEnvironmentId` values. Theirs is stronger. Doing the same here
  needs the CLI to be able to check the condition, which today it cannot -
  it is offline by design.
- **Nothing checks that a `tooling.description` names the tool it could be
  confused with.** It is the single field where a wrong value makes a model pick
  the wrong tool, three extracts say so, and it is unenforceable by anything but
  a person reading it.
- **`internal/cli` sits at 70% coverage.** The uncovered part is the output paths
  of the older commands - `add`, `usecase`, `next --list`. Not wrong, just
  untested.

### Deliberately deferred

- **A Claude Code plugin in the customer repo.**
  `.claude-plugin/marketplace.json` + `plugins/asgard-fde/{commands,skills}`, as
  the demo generator does it. The `asgard-fde-onboarding` design-time skill does
  the routing part today, which was most of the value; the rest waits until the
  plugin format is worth committing to.
- **`--template-dir` to override embedded prompt text.** Prompts are embedded and
  versioned with the binary, which is the right default. This is for when
  iterating on prompt text against a live engagement turns out to be painful, and
  it has not yet.
- **Homebrew tap and Scoop bucket.** The config is written and guarded by its
  token; enabling either is creating a repository and adding a secret. The FDE
  deferred this deliberately.
- **Linux packaging beyond what nfpm does by default.** Deferred by the FDE.

### Questions that block nothing yet

Listed in full under "Open questions": whether the EKS cluster names are
customer-specific, whether the helm major version should be pinned, and what
producing a web console's page map actually costs. The five platform unknowns
live in the scaffolded `docs/open-questions.md`, because every engagement hits
them and the answers belong where the engagement is.

## Cross-platform

The gate used to be bash calling `yq`, piped into Python needing PyYAML in a
virtualenv. **Three of those four do not work on Windows without WSL** - while
helm and kubectl both ship native Windows builds, so the two tools everyone
assumed were the problem were the only part that was fine.

```
before:  bash render.sh -> yq -> helm template -> python3 + PyYAML
now:     asgard-cli verify -> helm template -> internal/gate
```

`common/render.sh`, `check_chart_xref.py` and `check_agent_split.py` are deleted
from the scaffold. The xref rules were ported one for one and **the port was
checked against the Python original on the same rendered chart**: identical
findings, identical counts. `scripts/` keeps only `check_crd_fidelity.py` (needs a
cluster) and `db/` (needs psycopg/pymssql).

Windows specifics: `exec.LookPath` rather than a hardcoded `.exe` (it reads
`PATHEXT` and honours Git Bash and scoop shims); never through a shell; the
scaffold writes `.gitattributes` with `eol=lf`, because a CRLF shebang makes the
kernel look for an interpreter whose name ends in a carriage return. `python3` on
Windows resolves to an **App Execution Alias** that opens the Microsoft Store, so
`tool.Path()` rejects the zero-byte stub and tries `python`.

## Working here

- **`AGENTS.md` rules apply**: English, plain ASCII with no emoji, every command
  documents itself (`internal/cli/help_test.go` fails the build otherwise),
  generated files go in `.out/`.
- **Scratch repo**: `/Users/jjchen/projects/asgard/deleteme` is a throwaway
  directory for exercising the CLI against a real filesystem - scaffolding a
  customer repo end to end and running the real gate against it. `helm`, real
  paths and real relative links cannot be faked in a `t.TempDir()`. Nothing there
  is precious.
- **The gate on this repo**: `go test ./...`, `go vet ./...`, `gofmt -l`,
  `goreleaser check`.
