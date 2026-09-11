# SOURCES.md

Where the material in `internal/corpus/usecase/` came from.

> **Internal only.** This file lives outside `internal/`, so it cannot be embedded
> even by accident.
> `internal/corpus/usecase/` ships to every engagement, so those files name no customer
> and no deployment. This one exists so we can trace an extract back to the chart
> it was taken from.

Keep it that way: when adding an extract, put the customer-facing shape in
`internal/corpus/usecase/` and the attribution here.

**Record the commit you read.** The wiki side has always done this; this file
had no commit for any of the eight, which made "which version of that chart
did this extract describe" unanswerable. It is answerable now, and the answer
below was reconstructed rather than recorded at the time: every clone's HEAD
is dated **before** the extracts were written (2026-09-02 onward), so what was
read is what each clone holds.

## What each clone held when its extracts were written

Read 2026-09-11. **The right-hand column is the reason this table matters**: a
chart moves, and an extract is a description of one version of one chart.

| deployment | read at |
|---|---|
| unitech-e-asgard-kube | `223a59a` (2026-09-01) |
| xxentria-asgard-kube | `57e4b4c` (2026-09-01) |
| finance-ai-asgard-kube | `3a4ce84` (2026-09-01) |
| buy123-asgard-kube | `08dac8f` (2026-08-31) |
| asgard-freyr-kube | `0594f68` (2026-08-24) |
| asgard-auto-post-kube | `d11b802` (2026-09-01) |
| asgard-industry-demo-generator | `718cc0e` (2026-08-26) |
| asgard-freyr-skills | `e0b3fe3` (2026-09-01) |

**How far each has moved since is computed, not written here:**

    hack/sources.py --extracts

There used to be a third column with those distances in it, and **both of the
two non-zero rows had rotted within nine days.** One named a commit the clone
had already moved past; the other reported six commits of drift against a clone
sitting exactly where the extract was read. A distance between two things that
both move is the one shape of claim that cannot be written down and stay true -
the left-hand column can, because a commit is fixed.

**The two furthest along are the two that matter most for one extract each.**
`asgard-freyr-skills` is the only source for
`internal/corpus/usecase/browser-operation.md`, and `asgard-freyr-kube` is
where `agents.expression` and the sandbox hooks were read. An extract resting on
a sample of one, dozens of commits back, is the shape
`internal/corpus/wiki/coverage.md` exists to make visible.

**These are not in `asgard-cli audit-material --sources`.** That check reads
what is embedded in the binary, and this file deliberately is not - it is the
only one that names a customer. Holding these rows against the clones is a
thing somebody does here, with the clones, the way `hack/check-tables.py`
holds the pinned tables against asgard-kube.

## The deployments read so far

| deployment | shape it demonstrates | CR files | referred to in extracts as |
|---|---|---|---|
| unitech-e | agent hub (5 agents / 6 semantic layers) **and** a single-agent flow agent; trigger; knowledge drive. Its AGENTS.md is the most recently maintained of the set | 37 | "a later one", "a deployment with an internal hub and a public widget" |
| freyr | supervisor + 5 subagents, `agents.expression`, sandbox hooks, shared SourceSet | 16 | "a commerce back-office with five specialists", "an earlier deployment" |
| xxentria | supervisor + 9 agents | 18 | "a manufacturing one with nine" |
| finance-ai | supervisor, 3 semantic layers | 13 | "a finance one with three" |
| buy123 | the minimal flow agent - no Agent CR at all | 8 | not yet cited |
| auto-post | 28 Plugin CRs, knowledge bases, api workflows | 60 | not yet cited |
| industry-demo-generator | 12 industries, read/write governance split, a Claude Code plugin of commands + skills | many | not yet cited |

### Which customer each one is

**This is the only file that makes this link.** Nothing under `internal/` names a
customer, and nothing under `internal/` may.

| repo | who |
|---|---|
| [unitech-e-asgard-kube](https://github.com/asgard-ai-platform/unitech-e-asgard-kube) | 台新 |
| [xxentria-asgard-kube](https://github.com/asgard-ai-platform/xxentria-asgard-kube) | 森鉅 |
| [finance-ai-asgard-kube](https://github.com/asgard-ai-platform/finance-ai-asgard-kube) | FinanceAI |
| [buy123-asgard-kube](https://github.com/asgard-ai-platform/buy123-asgard-kube) | Buy123 |
| [asgard-freyr-kube](https://github.com/asgard-ai-platform/asgard-freyr-kube) | Freyr |
| [asgard-auto-post-kube](https://github.com/asgard-ai-platform/asgard-auto-post-kube) | Heimdall |
| [asgard-industry-demo-generator](https://github.com/asgard-ai-platform/asgard-industry-demo-generator) | Demo Generator |
| [asgard-freyr-skills](https://github.com/asgard-ai-platform/asgard-freyr-skills) | Freyr (runtime skills, a separate repo from the chart) |

The demo generator and auto-post are Asgard's own rather than a customer engagement. **`Heimdall` is
also the name of a product in the suite** (Media & PR AI, see `internal/corpus/wiki/
product-suite`) - the repo and the product are not the same thing, and an extract
saying "Heimdall" without saying which is ambiguous.

The platform contract itself is
[asgard-kube](https://github.com/asgard-ai-platform/asgard-kube), and the product
documentation is
[asgard-docs](https://github.com/asgard-ai-platform/asgard-docs). Neither is a
deployment; both are listed in `AGENTS.md` alongside these.

## Generational conflicts found so far

These matter more than any single extract: two charts disagree, and the
disagreement is dated rather than a matter of taste.

| topic | earlier | later | which wins |
|---|---|---|---|
| SkillSet to SourceSet | one shared store, several skill sets slicing it with searchPaths (freyr) | 1:1:1, one file per skill set (unitech-e, changed 2026-08-28) | later. The earlier shape leaves the UI unable to find a skill set's git config |
| `bot-provider-type` label | stamped, "front end breaks without it" (freyr) | dropped, platform derives it from `botProviderClass` (unitech-e, workflow-service #336) | later, but stamping it anyway is harmless |
| canvas metadata | hand-written `node_positions` ConfigMaps (demo generator) | dropped, platform auto-lays-out (unitech-e, #336-#340, 2026-08-31) | later |
| `Agent.managed.completionModelName` | required (demo generator's GOAL.md) | Agent takes no model; the caller picks per turn (unitech-e) | later |
| `KnowledgeBase` | in use (auto-post) | one deployment replaced it with a SourceSet Drive + contextIndex (TASK-013). Live and unmarked in the CRDs; do not restate as a platform deprecation | later |

## Where the platform's own documentation disagrees with every chart

Not a generational conflict - a documentation error, and one an agent would act
on.

| topic | the CRD documentation says | every chart does | evidence |
|---|---|---|---|
| `config.expression` | "CEL 表達式" | **JavaScript**: arrow functions (26 occurrences), `const` (11), `String()` (7), `encodeURIComponent` (5), `??` (4), `JSON.stringify` (3) | CEL has none of those constructs. Counted across every chart in the deployments listed above, on 2026-09-01 |

`internal/corpus/usecase/workflow-chain.md` states the corrected version and
says the docs are wrong, because an agent handed "it is CEL" writes something
that cannot work and has no way to find out why.

**unitech-e is the most recently maintained**, so it wins a conflict unless
there is a reason to think otherwise. Record the reason when there is.

## Which file each kind's material came from

Attribution for `internal/corpus/usecase/`, moved here from `TASK.md` because
this is the file that is allowed to name a deployment. Sizes are from when they
were read, and are a rough guide to how much of the knowledge is in the header
comments rather than the YAML.

| kind | source |
|---|---|
| `DataConnector` | `unitech-e/.../data_connector/dc-bpm.yaml` (585B) |
| `Toolset` | `unitech-e/.../toolset/ts-catalog.yaml` (3.2KB, exemplary comments) |
| `SandboxBlueprint` | `unitech-e/.../agent/sbp-website.yaml` (2.3KB) |
| `Agent` | `unitech-e/.../agent/ag-bpm.yaml` (5.5KB) |
| `Workflow` | `unitech-e/.../workflow/wf-website.yaml` (12KB) |
| `SkillSet` trio | `unitech-e/.../skill_set/sk-base.yaml` (3.9KB) |
| `SourceSet` + `Syncer` | `unitech-e/.../source_set/ss-website-knowledge.yaml` (4.6KB) |
| `Trigger` | `unitech-e/.../trigger/tr-pr-arrival-notify.yaml` (2.6KB) |
| `BotProvider` | `unitech-e/.../agent/bp-website.yaml` (4.4KB) |
| `SemanticLayer` | too large to ship whole (23KB-253KB); take one cube plus the `sampleQueries` shape |
| `CompletionModel`, write-path `Workflow` with `requestConsent` | `industry-demo-generator/retail/chart/...` |

`SemanticLayer` has no single source small enough to ship: take one cube plus the
`sampleQueries` shape. `CompletionModel` and the write-path `Workflow` with
`requestConsent` come from the demo generator rather than a kube repo.

**Where the two disagree, prefer `unitech-e`** - it is the more recently
maintained. The disagreements are dated rather than contradictory, and the table
above this one records them.
