---
name: asgard-cr-verification
description: Use before pushing any change under projects/*/chart or .asgard-pipeline.yaml — runs this repo's local gate (repo consistency, helm lint, render + CR cross-reference + Agent-split invariants), then reads back the platform's plan, which is where every CR is checked against the real cluster's CRDs. Also the reference for first-time environment setup.
version: 2.0.0
alwaysApply: false
---

# Asgard CR Verification (design time)

There is no application to run in this repo — every artifact is a declarative Kubernetes custom
resource under `asgard-ai.com/v1alpha1`. Verification therefore means **rendering the charts and
checking the resources reference each other correctly**.

Run the gate before pushing any change under `projects/*/chart/` or `.asgard-pipeline.yaml`.
**Any red step stops the work** — report the specific error, do not push past it.

> **Design time.** This is for the coding agent working in this repo. Runtime skills for the
> deployed agent live in `common/skills/` and are bound via `SkillSet` CRs.

## Prerequisites (one-time)

```bash
python3 -m venv .venv
.venv/bin/pip install -r scripts/db/requirements.txt   # only the scripts/db tooling needs this
```

Also required on PATH: **`helm`** (steps 2 and 3). Nothing else — no `yq`, no
bash, no `kubectl`, and no Python for the gate itself, so it runs unchanged on
Windows. Run **`asgard-cli doctor`** to see what is installed and how to install
what is not; it works out the command for the machine you are on.

**No step here needs a cluster, because no cluster credential is issued to a
client.** The checks that need one run on the platform, in step 4.

If a step fails on a missing tool, say which tool and the install command; do not silently skip it.

## The Gate

### 1. Repo consistency

```bash
asgard-cli check            # whole repo
asgard-cli check <project>  # one project
```

Expect `✓ repo consistency OK`. Checks the structural invariants a chart render cannot see:
the root README project table matches `projects/`, `.asgard-pipeline.yaml` names a chart that
exists for every release it declares, `common/skills/*/SKILL.md` carry `name` + `description`
frontmatter, the `requirements/` indexes are present, and the **`docs/` spec layer** is intact —
required files, the living spec's module index matching the actual module files,
`YYYY-MM-DD-<topic>.md` filenames under `decisions/` and `meeting-notes/`, and every relative
markdown link inside `docs/` resolving.

### 2. Chart static validation

```bash
helm lint projects/<project>/chart/app     # bare values.yaml
```

Expect exit 0. **The bare form is the point** — it is the only step that proves
`values.yaml` declares a default for every `.Values.*` a template reads. With an env file
overlaid a missing default is masked, so a template referencing an undeclared value renders fine
in the gate and then nil-pointers for anyone who runs plain `helm template`.

### 3. Render + CR cross-reference

```bash
asgard-cli verify <project>
```

Expect `✓ xref OK`. This is the step that catches dangling references, which `helm lint` cannot
see because to Helm these are opaque CRs:

- `Agent.managed.toolsetNames[]` → a real `Toolset`
- `Agent.managed.skillSetNames[]` → a real `SkillSet`
- `Agent.managed.semanticLayers[].name` → a real `SemanticLayer`
- `SemanticLayer.dataConnectorName` → a real `DataConnector`; `SemanticLayer.toolsetNames[]` → a real `Toolset`
- `SkillSet`/`Syncer`.`sourceSetName` → a real `SourceSet`; `Syncer.database.dataConnectorName` → a real `DataConnector`
- `Loader.knowledgeBaseName` → a real `KnowledgeBase`; `Loader.database.dataConnectorName` → a real `DataConnector`
- **every `(workflow, entry)` entrypoint** — `Toolset.tools[]`, `Workflow.exits[].handlingWorkflow`,
  `BotProvider.entrypoint`. A wrong **entry** name is as fatal as a wrong workflow name and apply
  catches neither, so both halves are resolved.
- `SandboxBlueprint.skillSetNames[]` → a `SkillSet`, `.toolsetNames[]` → a `Toolset`,
  `.pluginNames[]` → a `Plugin`, `.sourceSetMounts[].sourceSetName` → a `SourceSet`
- **the Flow Agent chain** — a processor's `sandboxBlueprint` config → a real `SandboxBlueprint`,
  and `SandboxBlueprint.spec.agents[].baseAgentName` → a real `Agent` (parsed out of the JSON
  string the CRD stores it in). This is how a public widget reaches its subagent; a typo here
  silently drops it, and the symptom is an agent that "can't call any tools".
  ⚠️ It cannot catch an **invented config key** — `configs[].name` is a free-form string, so a
  misspelt key passes lint, passes CRD validation, and does nothing at runtime.

It also checks that every CR the Platform UI renders carries its `<kind>-name` annotation
(`asgard-ai.com/agent-name`, `semantic-layer-name`, `data-connector-name`, …). Without it the CR
applies fine and then shows up nameless in the UI, so nothing but this check catches it. `Agent`
needs only `agent-name` — its description comes from `spec.managed.description`.

`asgard-cli verify` covers what a *reference* check cannot: the invariants of the
**one-system-one-Agent** split. Expect `✓ agent split OK`.

- each `Agent` mounts **at most one** semantic layer, and no layer is mounted twice — otherwise the
  search space the split was meant to shrink grows back. **Zero is legal**: a public-facing project
  may deliberately mount none and use fixed-query Toolsets instead (a layer without `allowedCubes`
  is arbitrary SQL over every cube, and the surface grows with each table added)
- an `Agent` with neither `semanticLayers` nor `toolsetNames` fails — zero layers is a choice, but
  an agent with no capability source at all is a mistake (usually a layer deleted by accident)
- no `allowedCubes` anywhere (the repo's standing decision: any table in the layer is queryable)
- every layer listed in `OLAP_ONLY_LAYERS` is on no `Agent` — those are the Data Insight system's OLAP sources
- every `Agent` has ≥2 `sampleQuestions`, except those listed in `EXEMPT_SAMPLE_QUESTIONS`
- `prompt.task` and `prompt.format` are **byte-identical** across all Agents. Agent CRs have no
  include mechanism, so shared prompt text is duplicated; verbatim equality is what makes a change
  a single global replace. **Editing one Agent's `task`/`format` and not the others fails here.**

Use `asgard-cli render <release>`, not a bare `helm template`: it takes the chart from the
release's entry in `.asgard-pipeline.yaml` and supplies the reserved `.Values.asgard.*` block, so a
chart that reads those renders instead of failing on a missing key.

**What it renders is not what deploys.** The values are placeholders and nothing is checked against
a cluster. That is step 4's job, and step 4 is not optional.

### Run step 3 for every release

`asgard-cli verify` with no argument does every release the declaration names.

Projects legitimately have **different shapes**, and neither is incomplete:

| | reads through a semantic layer | reads through fixed tools |
|---|---|---|
| Read path | one `SemanticLayer` per Agent | no `SemanticLayer`, a `Toolset` of zero-parameter queries |
| Entry | the platform's `preset-agent-hub`, every caller can authenticate | own `BotProvider`, public and unauthenticated |
| `Agent` CRs | one per source system | often none: a Flow Agent keeps the prompt on the Workflow and the capabilities on a `SandboxBlueprint` |

Which shape a project takes follows from its audience, not from preference. An
anonymous visitor cannot reach the agent hub, and a semantic layer mounted
without `allowedCubes` is arbitrary SQL over every cube in it.

## Reading the CRDs directly

`asgard-ai-platform/asgard-kube` holds the actual `openAPIV3Schema` for every kind, and is the
authority when you need to know whether a field exists — prefer it over inferring shape from a CR
dump:

```bash
git clone --depth 1 https://github.com/asgard-ai-platform/asgard-kube.git /tmp/asgard-kube
```

Note what it *cannot* tell you: `Workflow.processors[].configs[].name` is a free-form string, so
the set of valid config names for a processor type is not discoverable there. Do not guess one — a
wrong config name lints clean, passes CRD validation, and then does nothing at runtime.

## 4. The platform's plan — the step that cannot be run here, and cannot be skipped

Push, then read the plan back:

```bash
git push origin <tag>
asgard-cli pipeline runs watch --release <name> --commit $(git rev-parse HEAD)
```

The plan renders the chart with the release's **real** values and real ids, then sends **every CR
to the apiserver with a server-side dry run**. That is where these are caught, and nowhere else:

- **pattern violations** (the `Trigger` cron regex), **CEL rules** (`exactly one of
  value/expression/template`), and `Required value` — reported as `crd/dry-run-rejected`.
- **a field the CRD does not declare** — reported as `crd/unknown-field`. This one is worth
  understanding: CRDs prune unknown fields by default, so a plain `kubectl apply
  --dry-run=server` reports `created` while the field is quietly discarded. A deprecated
  `Toolset.spec.instruction` passed 25 of 25 dry-runs and then failed the real deploy, because
  helm's server-side apply builds a typed patch and refuses:

      .spec.instruction: field not declared in schema

  "Will it be accepted" and "will it be kept" are different questions. The plan asks both.

It also reports `chart/kind-not-allowed` for anything rendered that is not an `asgard-ai.com`
resource, `chart/hook-not-allowed`, remote chart dependencies, and `vars/required-missing` for a
declared key with no value on the platform.

A run stops at review unless the release has auto-apply on:

```bash
asgard-cli pipeline runs approve <run-id>
```

**If the run was never created, the push matched nothing** — either no release's pattern matched
the ref, or the one it matched was never created on the platform. `asgard-cli pipeline deliveries`
says which; nothing else will.

## Reporting

Summarize PASS/FAIL per step. Declare success only when every step is green, **including step 4**.
Steps 1 to 3 passing is not the gate passing: they check what a dry run passes and runtime still
fails, which is the half a client can check. On failure, name the release and the step, and quote
the actual error output rather than paraphrasing it — a plan report names the rule that fired.

**Checked:** 2026-09-04, twice, and the second pass found more than the first.

Every `asgard-cli` command this page names was run against the built binary:
`check`, `verify`, `render` and `doctor` resolve; `pipeline runs` and its
subcommands resolve; **`pipeline deliveries` did not exist** and has been added.
That one mattered more than the others - this page names it as the only place a
push that produced no run explains itself, which was true of the platform and
not of the tool.

Every lint rule code quoted in step 4 resolves in `asgard-iac`'s `pkg/lint`:
`config/schema`, `config/release-missing`, `config/chart-missing`,
`vars/required-missing`, `chart/kind-not-allowed`, `chart/hook-not-allowed`,
`crd/dry-run-rejected`, `crd/unknown-field`.

The reference list was held against the references the gate actually resolves,
read out of its own source rather than from this page, and **was four short**:
`SemanticLayer.toolsetNames[]`, `Syncer.database.dataConnectorName`,
`SandboxBlueprint.toolsetNames[]` and `SandboxBlueprint.sourceSetMounts[]` are
all resolved and none was mentioned. They are now.

The `≥2 sampleQuestions` and byte-identical prompt rules are the **gate's**, not
the platform's - the CRD sets no minimum and no equality constraint - and a
reader needs that difference to know whether a failure is fixed in the chart or
argued with us.

**Unchecked:** the field paths themselves, against the CRDs. The previous pass
held them against asgard-kube `15ded0f`; this one did not re-fetch, so what is
current is that the **gate** resolves them, not that each still exists in the
contract. The four added above are named the way the gate names them, which is
the same thing the gate would fail on.

Also unchecked: step 4 end to end from a repository this skill was scaffolded
into. It has been run - a real push produced a plan of 29 CRs, was approved, and
applied - but from the fixture repository rather than from a scaffolded one, so
what is proven is the platform's behaviour rather than this page's instructions
for reaching it.
