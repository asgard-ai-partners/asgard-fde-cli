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
| asgard-docs | the product documentation | 71 / 162 cited at `f00e0ee`; 91 published and unread; 28 deliberately excluded below. **Computed, not counted** - `hack/check-coverage.py` in asgard-fde-cli recomputes it and fails when this row drifts |
| asgard-kube `crd/` | the contract | read per page, per field, and dated on the page |
| **asgard-kube `pkg/apis/`** | **the Go types the CRDs are generated from, with the reasoning as comments** | read once, 2026-09-02, for the validation rules - `../wiki/crd-rules.md`. 134KB of declarations; what has been taken is the behavioural comments, not the field list |
| **[asgard-core](https://github.com/asgard-ai-platform/asgard-core)** `internal/constants.go` | **the processor definitions the CRD is generated from** | read once, 2026-09-02, for the type list. Its per-processor config definitions are not carried anywhere |
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
158 pages at `23409b3` against 162 at `f00e0ee`, with the same 71 cited.
asgard-fde-cli's `hack/check-coverage.py --head` prints both, and the difference
is the size of what re-reading would cover rather than a defect.

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
