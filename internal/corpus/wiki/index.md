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
| [`processors`](../wiki/processors.md) | what each of the 13 takes, and the fields that decide behaviour |
| [`workflow`](../wiki/workflow.md) | the 13 processors; Expression is JavaScript, Template is Handlebars |
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

**A coverage number here names its denominator in the same sentence or it does
not go here.** A percentage measured against one source out of nine reads as a
statement about the material, and an in-scope denominator restated as a count
of citations reads the same way. Both are the failure this section exists to
warn about.

Counted 2026-09-03 against a clone at `f00e0ee`, the commit this wiki records,
by deriving each file's published URL (`slug:` where one is declared) and
matching it against every page's source block:

| | count | how |
|---|---|---|
| files under `docs/` | 162 | `find docs -name '*.md*'` |
| `draft: true`, so not published | 16 | frontmatter |
| cited by some page's source block | 81 | URL match, per file |
| uncited | 81 | the remainder |
| - of those, drafts | 12 | |
| - **published and uncited** | **69** | **the number that means anything** |
| cited but draft | 4 | the four `audit-material --urls` reports as dead and disclosed |

The last row is the check on the rest: `audit-material --urls` fetches all 81
cited links against the live site and gets 4 404s, all of them drafts the site
does not publish, all disclosed in their own citations. Two independent counts
agreeing is what the earlier figures never had.

That is a statement about the product documentation, and it was being read as a
statement about the material - which is how a wiki with nothing about SHOPLINE,
nothing about Mimir as a deliverable, and nothing about the largest chart
repository in existence could report itself complete.

The sources this material is actually built from:

| source | what it holds | state |
|---|---|---|
| asgard-docs | the product documentation | 77 / 162 cited at `f00e0ee`; 85 published and uncited; 28 deliberately excluded below. **Computed, not counted** - `hack/check-coverage.py` in asgard-fde-cli recomputes it and fails when this row drifts |
| asgard-kube `crd/` | the contract | read per page, per field, and dated on the page |
| **asgard-kube `pkg/apis/`** | **the Go types the CRDs are generated from, with the reasoning as comments** | read once, 2026-09-02, for the validation rules - `../wiki/crd-rules.md`. 134KB of declarations; what has been taken is the behavioural comments, not the field list |
| **[asgard-core](https://github.com/asgard-ai-platform/asgard-core)** `internal/constants.go` | **the processor definitions the CRD is generated from** | walked 2026-09-03 at `5da86c6` and re-walked 2026-09-11 at `623ceb5`. Every processor's required keys, defaults and declared outputs are in `../wiki/processors.md`, and asgard-fde-cli's `hack/check-processors.py` is what holds that table against the literal |
| **asgard-freyr-skills** | **nine runtime skills, incl. the SHOPLINE pair** | one page - `../usecase/skill-layers.md`. The eighth reference repository, and the only one with no CRs |
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

**And it is not counted by hand.** `hack/check-coverage.py` in asgard-fde-cli
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

**28 files at `f00e0ee`, and the count is computed** - it was 29 by hand, and
before that 32, because four `asgard-builtin` pages came back into
`../wiki/processors.md` and the subtraction was done in one place and not the
other. asgard-fde-cli's `hack/check-coverage.py` counts it now.

**`superpowers/` no longer exists upstream.** Those five pages are gone from
asgard-docs as of `23409b3`, so at the clone's HEAD the excluded set is 23 and
not 28 - a row that excludes a directory can stop being an exclusion by the
directory being deleted, which is not something this material can notice.

**The denominator moved by four rather than five**, because
`developer-reference/processor/query-llm-database` arrived in the same span:
158 pages at `23409b3` against 162 at `f00e0ee`, with the same 77 cited.
asgard-fde-cli's `hack/check-coverage.py --head` prints both, and the difference
is the size of what re-reading would cover rather than a defect.

**"Uncited" is not "unread", and the gap is two families.** 15 of the 85 are
the per-processor reference pages and 11 more are the SSE event pages under
`developer-reference/api-doc/send-message/sse-response/`, both of which this
material points at **by URL pattern rather than by link** - `../wiki/processors.md`
gives the pattern and one example, `api.md` says "one page per event" and links
one. That is deliberate: 26 links to pages whose content is a field table would
be 26 things to keep resolving. A further 14 are the excluded message-template
shapes. **So the number is a measure of how much is linked, not of how much has
been read**, and it is worth knowing which before treating it as a backlog.

**A live URL is not the file path under `docs/`, and matching them literally
undercounts this row by six.** Four channel pages are served from capitalised
files - `integration/line` from `integration/LINE.mdx` - and two more are a
directory's `index.mdx` reached without the `index`; 14 of the 158 pages declare
a `slug:` that differs from where they sit. The numerator is resolved through
that frontmatter for this reason. **Computing a number does not make it right;
it makes it re-derivable**, which is the only reason this was catchable.

| excluded | count at `f00e0ee` | why |
|---|---|---|
| `developer-reference/asgard-builtin/message-template-*` | 14 | Message template shapes - button, carousel, image, video, location. Genuinely lookup material, and per-channel. Read the source when writing one |
| `help-community/release-notes/` | 10 | historical, and does not describe the present |
| `superpowers/` | 5 | the documentation site's own redesign plans, not an Asgard feature |

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
