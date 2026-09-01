---
name: asgard-cr-verification
description: Use before committing any change under projects/*/chart or common/ — runs this repo's acceptance gate (repo consistency, helm lint, render + CR cross-reference + Agent-split invariants) plus, against a live cluster, a server-side dry-run AND a CRD-fidelity check that finds fields the CRDs would silently prune. Also the reference for first-time environment setup.
version: 1.2.0
alwaysApply: false
---

# Asgard CR Verification (design time)

There is no application to run in this repo — every artifact is a declarative Kubernetes custom
resource under `asgard-ai.com/v1alpha1`. Verification therefore means **rendering the charts and
checking the resources reference each other correctly**.

Run the gate before committing any change under `projects/*/chart/`, `common/`, or a `deploy.yaml`.
**Any red step stops the work** — report the specific error, do not push past it.

> **Design time.** This is for the coding agent working in this repo. Runtime skills for the
> deployed agent live in `common/skills/` and are bound via `SkillSet` CRs.

## Prerequisites (one-time)

```bash
python3 -m venv .venv
.venv/bin/pip install -r scripts/db/requirements.txt   # only step 4 (check_crd_fidelity.py) needs this
```

Also required on PATH: **`helm`** (steps 2 and 3) and **`kubectl`** (step 4).
Nothing else — no `yq`, no bash, and steps 1 to 3 need no Python at all, so the
gate runs unchanged on Windows. Run **`asgard-cli doctor`** to see what is
installed and how to install what is not; it works out the command for the
machine you are on.
`kubectl` only for the optional dry-run step.

If a step fails on a missing tool, say which tool and the install command; do not silently skip it.

## The Gate

### 1. Repo consistency

```bash
asgard-cli check            # whole repo
asgard-cli check <project>  # one project
```

Expect `✓ repo consistency OK`. Checks the structural invariants a chart render cannot see:
the root README project table matches `projects/`, every project has a well-formed `deploy.yaml`
whose declared values files exist, `common/skills/*/SKILL.md` carry `name` + `description`
frontmatter, the `requirements/` indexes are present, and the **`docs/` spec layer** is intact —
required files, the living spec's module index matching the actual module files,
`YYYY-MM-DD-<topic>.md` filenames under `decisions/` and `meeting-notes/`, and every relative
markdown link inside `docs/` resolving.

### 2. Chart static validation

```bash
helm lint projects/<project>/chart/app                                            # bare values.yaml
helm lint projects/<project>/chart/app -f projects/<project>/chart/values-dev.yaml
```

Expect exit 0 from both. **Run the bare one too** — it is the only step that proves
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
- `SemanticLayer.dataConnectorName` → a real `DataConnector`
- `SkillSet`/`Syncer`.`sourceSetName` → a real `SourceSet`
- `Loader.knowledgeBaseName` → a real `KnowledgeBase`; `Loader.database.dataConnectorName` → a real `DataConnector`
- **every `(workflow, entry)` entrypoint** — `Toolset.tools[]`, `Workflow.exits[].handlingWorkflow`,
  `BotProvider.entrypoint`. A wrong **entry** name is as fatal as a wrong workflow name and apply
  catches neither, so both halves are resolved.
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

Always use `asgard-cli render`, not a bare `helm template`. It takes the namespace and values
files from `deploy.yaml` and applies the same overlay order, `--namespace` and release name as CI,
so what you check is exactly what CD deploys.

### Run step 3 for every project x env combination

`asgard-cli render` refuses an env a project has not declared in its `deploy.yaml`. That
refusal is by design and is never "fixed" by inventing a namespace.

Projects legitimately have **different shapes**, and neither is incomplete:

| | reads through a semantic layer | reads through fixed tools |
|---|---|---|
| Read path | one `SemanticLayer` per Agent | no `SemanticLayer`, a `Toolset` of zero-parameter queries |
| Entry | the platform's `preset-agent-hub`, every caller can authenticate | own `BotProvider`, public and unauthenticated |
| `Agent` CRs | one per source system | often none: a Flow Agent keeps the prompt on the Workflow and the capabilities on a `SandboxBlueprint` |

Which shape a project takes follows from its audience, not from preference. An
anonymous visitor cannot reach the agent hub, and a semantic layer mounted
without `allowedCubes` is arbitrary SQL over every cube in it.

## Optional: validate against the real CRDs

`asgard-ai-platform/asgard-kube` holds the actual `openAPIV3Schema` for every kind. Validating the
rendered output against it catches field typos and enum violations that `asgard-cli verify`
cannot see, and needs no cluster:

```bash
git clone --depth 1 https://github.com/asgard-ai-platform/asgard-kube.git /tmp/asgard-kube
# then validate each rendered doc against /tmp/asgard-kube/crd/*.yaml (openAPIV3Schema per kind)
```

That repo is also the authority when you need to know whether a field exists — prefer it over
inferring shape from a CR dump. Note what it *cannot* tell you: `Workflow.processors[].configs[].name`
is a free-form string, so the set of valid config names for a processor type is not discoverable
there. Do not guess one — a wrong config name lints clean, passes CRD validation, and then does
nothing at runtime.

## 4. Against a live cluster — TWO checks, and they catch different things

Both need a cluster with the Asgard CRDs installed and are read-only. Run both whenever a CRD-shape
change is in play (a field added or **removed** upstream in `asgard-kube`).

### 4a. Server-side dry-run — does the apiserver accept it?

```bash
asgard-cli render <project> dev | kubectl apply --dry-run=server -n <any-ns> -f -
```

Catches pattern violations (the `Trigger` cron regex), CEL rules (`exactly one of value/expression/
template`), and `Required value`. Nothing is persisted.

### 4b. CRD fidelity — is anything being silently dropped?

```bash
asgard-cli render <project> dev | python3 scripts/check_crd_fidelity.py - --context <ctx>
```

> **Step 4a does NOT catch a field the CRD does not declare.** CRDs prune unknown fields by
> default, so dry-run reports `created` while the field is quietly discarded. A deprecated
> `Toolset.spec.instruction` passed 25/25 dry-runs and then failed the real deploy, because
> **helm's server-side apply builds a typed patch and refuses**:
>
> ```
> .spec.instruction: field not declared in schema
> ```
>
> 4a answers "will it be accepted", 4b answers "will it be kept". A CD failure is the only other
> place this surfaces — which is far too late.

**Point it at the cluster you deploy to, not a local kind.** The two can disagree, and the one that
matters is the deploy target:

```bash
asgard-cli render internal dev | python3 scripts/check_crd_fidelity.py - \
  --context arn:aws:eks:ap-northeast-1:698306514474:cluster/asgard-ai-eks
```

If no cluster is reachable, this step is **not run** — that is not the same as passing. Say so.

## Reporting

Summarize PASS/FAIL per step. Declare success only when every step is green. On failure, name the
project and the step, and quote the actual error output rather than paraphrasing it.
