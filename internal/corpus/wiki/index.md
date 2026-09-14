# Index

    ../wiki/<page>.md

## Products and scope

| page | covers |
|---|---|
| [`product-suite`](../wiki/product-suite.md) | the six products, what each is for and who uses it |
| [`console`](../wiki/console.md) | the two permission layers, five inconsistent pages, Workspace settings |
| [`sindri`](../wiki/sindri.md) | Project, delegation, the Sandbox and its files, the governance gate |
| [`mimir`](../wiki/mimir.md) | Thread, View, Dashboard, Knowledge |
| [`fehu`](../wiki/fehu.md) | billing and usage, how cost is broken down |

## While building

| page | covers |
|---|---|
| [`setup-path`](../wiki/setup-path.md) | the order: where a credential goes, what to build from it, why Sindri needs no import |
| [`agents`](../wiki/agents.md) | Flow Agent against Managed Agent, how to choose, UI-to-CR names |
| [`knowledge`](../wiki/knowledge.md) | Drive, Context Index, and how Knowledge Base differs |
| [`semantic-model`](../wiki/semantic-model.md) | the modelling flow, its limits, the Mimir side |
| [`tools`](../wiki/tools.md) | MCP Server, Skillset and Plugin; hook events |
| [`automation`](../wiki/automation.md) | Trigger and API, and why only cron is left |
| [`processors`](../wiki/processors.md) | what each processor type takes, and the fields that decide behaviour |
| [`workflow`](../wiki/workflow.md) | the processor types against the editor's groups; Expression is JavaScript, Template is Handlebars |
| [`settings`](../wiki/settings.md) | Completion and Embedding Model, Data Source, Connection |
| [`integration`](../wiki/integration.md) | chat platforms, the two Applications pages, the architecture |
| [`api`](../wiki/api.md) | the endpoint and its actions, the SSE sequence, four patterns, the SDK |
| [`crd-rules`](../wiki/crd-rules.md) | the validations helm lint does not run, and the one the schema cannot express |
| [`platform-unknowns`](../wiki/platform-unknowns.md) | what no source answers, and who to ask |
| [`coverage`](../wiki/coverage.md) | how many deployments each CR shape was read from - which extracts rest on a sample of one |

## In practice

| page | covers |
|---|---|
| [`operations`](../wiki/operations.md) | Asgard's outbound IPs, checking model capability, vocabulary |
| [`taiwan-channels`](../wiki/taiwan-channels.md) | the commerce channels a customer will name, what SHOPLINE cost, and what we have not built |
| [`glossary`](../wiki/glossary.md) | words that mean one thing here, and what the other senses are called. **Check it for the word you searched**: a result in the wrong sense reads exactly like an answer |
| [`what-they-read`](../wiki/what-they-read.md) | the picture a customer arrives with, and the three places it is wrong |
| [`case-studies`](../wiki/case-studies.md) | the retail stockout from three angles, plus a Flow Agent help desk |
| [`screenshots`](../wiki/screenshots.md) | which picture answers which question, and the URL to fetch it from |

## Every UI name, and the CR it is

**Goal's first point owes an agent this table and there was one row family of
it - the agent pages.** Every other mapping was stated in the prose of whichever
page discusses the feature, which a grep for the UI name does reach and which
nothing could check for completeness. **A UI name with no CR stated anywhere is
invisible**, and that is what this closes.

Read it as "the customer said X, so the chart writes Y". The page column is
where the judgement is; this table is only the name.

| the UI calls it | the chart writes | and the page is |
|---|---|---|
| Agent Hub > Flow Agent | `Workflow` + `SandboxBlueprint` + `BotProvider`, and **no `Agent`** | `../usecase/flow-agent-single.md` |
| Agent Hub > Managed Agent | `Agent` | `../wiki/agents.md` |
| Agent Hub > Configuration > Models | **no CR of its own** - it selects `CompletionModel`s that already exist | `../wiki/settings.md` |
| Agent Hub > Configuration > Global Directory | a read-only `SourceSet`, mounted through `SandboxBlueprint.extraDirectories` | `../usecase/conventions.md` |
| Applications (Data Insight & Agent Hub) | **no CR** - a listing of what is already published | `../wiki/integration.md` |
| Applications > Customized Integration | `BotProvider` | `../usecase/chat-channel.md` |
| Automation > API | `Workflow`, with the `automation_tool` workflow-set type | `../wiki/automation.md` |
| Automation > Trigger | `Trigger`, plus the `Workflow` it enters | `../usecase/trigger.md` |
| Data Insight > Semantic Model | `SemanticLayer` | `../usecase/semantic-layer.md` |
| Drive | `SourceSet`, one `Syncer` per source | `../usecase/knowledge-drive.md` |
| Drive > Context Index | `Indexer`, and three CRs the reconciler derives from `spec.contextIndex` | `../wiki/knowledge.md` |
| Knowledge Base | `KnowledgeBase`; a `Loader` per Auto Load source, a `Source` per item | `../usecase/knowledge-base.md` |
| MCP Servers | `Toolset` | `../wiki/tools.md` |
| Plugins | `Plugin` | `../usecase/plugin.md` |
| Skillsets | `SkillSet` + its own `SourceSet` + the `Syncer` that fills it | `../usecase/skill-set.md` |
| Settings > Completion Model | `CompletionModel` | `../wiki/settings.md` |
| Settings > Embedding Model | `EmbeddingModel` | `../wiki/settings.md` |
| Settings > Data Source | `DataConnector` | `../wiki/settings.md` |
| Settings > Connection | `OAuthProvider` + `OAuthCredential`. **Not Data Source**: this is third-party OAuth, where Data Source is a credential you type | `../wiki/settings.md` |

**Four kinds are in the contract and are nobody's to create.** `Sandbox` is the
runtime object a blueprint produces, so a chart never writes one; and
`ImageGenerationModel`, `TranscriptionModel` and `SourceSetEditorServer` are
**internal** - confirmed 2026-09-14, `../wiki/platform-unknowns.md` P12. Their
absence from the documentation is the answer rather than a gap, so **a customer
asking for image generation or transcription needs the platform team and not a
chart.**

**Checked:** 2026-09-11, the UI's own vocabulary from asgard-docs `f00e0ee` -
the sixteen pages under `product-suite/odin/features/` - held against the 24
kinds in asgard-kube `cbd8d70` `crd/`. Each row's CR is the one that page says
gets created, or the one the extract in the third column writes.

**Unchecked:** Mimir's and Sindri's own pages are not in it. Their features
(Thread, View, Dashboard, My Chat, Directory) are reached rather than authored,
so a chart writes nothing for them - but nobody has confirmed that a Mimir View
leaves no CR behind.

## Known gaps between the documentation and the CRD

Each is written on the page it affects:

| gap | page |
|---|---|
| the UI's Flow Agent is three CRs; `agentClass` has one value | `../wiki/agents.md` |
| the Drive Syncer UI offers five sources, the CRD supports ten | `../wiki/knowledge.md` |
| Knowledge Base and Drive both exist and which one new work should use | `../wiki/knowledge.md` |
| Connection's "For Trigger" group names removed trigger classes | `../wiki/automation.md`, `../wiki/settings.md` |
| Semantic Model lists six data sources, the settings page nine (unexplained) | `../wiki/semantic-model.md` |
| the four `integration-with-asgard/` pages are `draft` and name an older UI | `../wiki/integration.md` |
| the glossary's Processor list does not match the current `ProcessorType` | `../wiki/operations.md` |
| `overview/asgard-features` is `draft` and links to removed paths | `../wiki/product-suite.md` |

## Coverage

**A coverage number names its denominator in the same sentence or it does not
go here.** A percentage measured against one source out of nine reads as a
statement about the material, and an in-scope denominator restated as a count
of citations reads the same way.

**And an index does not carry arithmetic.** There was a seven-row ledger here
deriving the figure by hand, and beside it a record of what the figure used to
be - which is a changelog, and it is what a count looks like when nothing can
recompute it. `go run ./hack coverage` in asgard-fde-cli derives every one of
these from the clone and **fails when the row below drifts**, so the row is the
only place a number belongs and the ledger is gone.

**A page's live URL is not its path under `docs/`**, which is why the count is
resolved through `slug:` frontmatter: channel pages are served from capitalised
files, a directory's `index.mdx` answers without the `index`, and fourteen
pages declare a slug that differs from where they sit. Matching literally
undercounts. **Computing a number does not make it right; it makes it
re-derivable**, which is the only reason that was ever catchable.

The sources this material is actually built from:

| source | what it holds | state |
|---|---|---|
| asgard-docs | the product documentation | 77 / 162 cited at `f00e0ee`; 85 published and uncited; 28 deliberately excluded below. **Computed, not counted** - `go run ./hack coverage` in asgard-fde-cli recomputes it and fails when this row drifts |
| asgard-kube `crd/` | the contract | read per page, per field, and dated on the page |
| **asgard-kube `pkg/apis/`** | **the Go types the CRDs are generated from, with the reasoning as comments** | read once, 2026-09-02, for the validation rules - `../wiki/crd-rules.md`. 134KB of declarations; what has been taken is the behavioural comments, not the field list |
| **[asgard-core](https://github.com/asgard-ai-platform/asgard-core)** `internal/constants.go` | **the processor definitions the CRD is generated from** | walked 2026-09-03 at `5da86c6` and re-walked 2026-09-11 at `623ceb5`. Every processor's required keys, defaults and declared outputs are in `../wiki/processors.md`, and asgard-fde-cli's `go run ./hack processors` is what holds that table against the literal |
| **asgard-freyr-skills** | **9 runtime skills, incl. the SHOPLINE pair** | one page - `../usecase/skill-layers.md`. The eighth reference repository, and the only one with no CRs |
| **seven deployments** | **every shape the extracts describe** | 19 charts between them; `../wiki/coverage.md` counts per deployment and says why |

**Deployment coverage cannot be measured from this material, by design.** An
extract names no customer and no deployment - it says "seen in a deployment
whose..." - so nothing here can be counted against the charts it came from. The
inventory has to be run separately, over the charts, and its result is
[`coverage`](../wiki/coverage.md) - how many deployments each CR shape was actually read
from - rather than a number here.

**A percentage here names its denominator in the same sentence, or it does not
go here.** A fraction whose numerator counts links and whose denominator counts
pages reads as better coverage than it is, and those are the two things easiest
to confuse.

**And it is not counted by hand.** `go run ./hack coverage` in asgard-fde-cli
computes all four numbers and fails when this row drifts from them.

## Deliberately not covered

**One row of this table was wrong.** `asgard-builtin/` was excluded whole as
lookup material; four of its pages are the expression language every processor
field is written in, including where the ECMA5 limit actually applies - to
`execute-script`'s engine, not to every Expression - and the variables in scope,
one of which the documentation never mentions at all.
They are now in [`processors`](../wiki/processors.md). The message-template pages remain
excluded, and that part of the judgement holds.

**An exclusion is a judgement someone made once.** Recheck one before relying on
it, particularly if it excludes a whole directory - that is the shape of an
exclusion nobody has looked inside.

**An exclusion can stop being one without anybody noticing**, and one has:
`superpowers/` is gone from asgard-docs, so at the clone's HEAD the excluded set
is smaller than the row above says and neither number is wrong. A row that
excludes a whole directory is the shape that does this.

**"Uncited" is not "unread", and the gap is two families.** The per-processor
reference pages and the SSE event pages are pointed at **by URL pattern rather
than by link** - `../wiki/processors.md` gives the pattern and one example,
`api.md` says "one page per event" and links one. That is deliberate: a link per
page whose content is a field table would be that many things to keep resolving.
**So the figure measures how much is linked, not how much has been read**, and
it is worth knowing which before treating it as a backlog.

| excluded | why |
|---|---|
| `developer-reference/asgard-builtin/message-template-*` | Message template shapes - button, carousel, image, video, location. Genuinely lookup material, and per-channel. Read the source when writing one |
| `help-community/release-notes/` | historical, and does not describe the present |
| `superpowers/` | the documentation site's own redesign plans, not an Asgard feature. **Gone from asgard-docs since**, which is how an exclusion stops being one without anybody noticing |

## Keeping this index complete

Every page in this directory has a row above.

**Where a page's own provenance is: on the page.** Each one ends with a source
block naming the rendered documentation page and the commit it was read at,
plus an `**Unchecked:**` line. There is no separate log - one existed, and a
running record of what was read when is a thing to maintain rather than a thing
to answer from, so what survived is the per-page half that a reader actually
follows.

**Adding a page means adding a row**, under the question it answers rather than
at the end. The extracts index went a third out of date this way -
eight of twenty-two written and never grouped - and nothing broke, because
the directory listed them regardless. What was lost is the only thing an
index is for: telling somebody which page answers their question when they do
not already know its name.

## Outside this wiki

| what you want | where |
|---|---|
| how a deployment shape is assembled, field by field | `../usecase/` |
| whether a CR field is legal or required | [asgard-kube's `crd/*.yaml`](https://github.com/asgard-ai-platform/asgard-kube/tree/main/crd) |
| what this customer's systems look like | that customer repo's `docs/spec/` |
| the original product documentation | [asgard-docs](https://github.com/asgard-ai-platform/asgard-docs), rendered at https://docs.asgard-ai.com |
