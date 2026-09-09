# Drive and Knowledge Base

Both hold unstructured knowledge and both are live. Drive is what this repo
recommends for new work, and the reason is one engagement's experience rather
than anything the platform has announced.

| | Drive | Knowledge Base |
|---|---|---|
| status | live | live - no deprecation marker on any of its four CRDs, and the console ships the feature |
| CRs | `SourceSet` (+ `Syncer`) | `KnowledgeBase` + `Loader` + `Indexer` + `Source` |
| retrieval | a Context Index knowledge graph | RAG over chunks |

Start new knowledge on a Drive with a Context Index. That is a recommendation
with a traceable origin - one deployment moved off `KnowledgeBase` + `Loader` +
a retrieval workflow and found the Drive simpler to keep correct - and not a
statement that the platform is retiring anything. Before repeating it to a
customer as platform direction, ask the platform team, because nothing in the
CRDs or the product documentation says so.

## Drive

A mountable file store an agent can read and write. Creating one needs a Name and
a Description, plus optional Reference Paths (a path may not start with `/`, and
may not contain `../` or `./`).

The detail page has four tabs:

| tab | |
|---|---|
| Files | a file browser, read-only until Edit is pressed. Also Open in Advance Editor |
| Syncers | scheduled pulls from external sources |
| Context Index | builds a searchable index over the Drive |
| Settings | the same fields as creation |

### Syncers

The table shows Folder, Source, Active, Status, Schedule. In Edit mode, New
Syncer opens a three-step wizard: Type, Basic, Settings.

**The UI offers five source types; the CRD supports ten.**

| source | in the UI | `syncerClass` |
|---|---|---|
| Google Drive | yes | `google-drive` |
| OneDrive | yes | `onedrive` |
| Git | yes | `git` |
| Web Crawler | yes | `web` |
| Data Source | yes | `database` |
| Bot | no | `bot` |
| Dropbox | no | `dropbox` |
| FTP / SFTP / SMB | no | `ftp` / `sftp` / `smb` |

The bottom five can only be declared in a chart. If a customer's data lives on an
FTP server or an SMB share, that route works - it just cannot be built in the UI.

### Context Index

Builds a searchable index over the Drive so an agent queries the index instead of
reading every file each time. It refreshes on a schedule. Off by default; Enable
Context Index turns it on.

It maps to `SourceSet.spec.contextIndex`. Setting the field makes the platform
derive three CRs - a Workflow, a SandboxBlueprint and a Trigger - that run the
indexer.

To pause indexing use the label `asgard-ai.com/context-index-suspend`. Clearing
the `contextIndex` field instead tears down those three derived CRs and renames
the built index aside.

## The index runs after the Syncers, not with them

`contextIndex.cron` and each Syncer's `schedule` are independent and nothing
orders them. Put the index **after** the Syncers on the same day - the deployment
runs its two Syncers at 09:00 and the index at 10:00, both `Asia/Taipei`.

Reversed or simultaneous, the index walks the volume before the day's content
lands, so the graph describes yesterday, every day, without ever failing. An
incremental update over an unchanged Drive finishes in seconds, so the gap costs
nothing.

The derived CRs take the SourceSet's name with a `-ci` suffix, so a Drive named
`ss-<name>-knowledge` produces `ss-<name>-knowledge-ci`. That is what to look for
on a cluster when the index is not running.

## Two labels, and neither reads the other

`asgard-ai.com/syncer-suspend: "true"` stops the **scheduler**, and nothing
else. It does not stop a deploy from firing the Syncer, and that is deliberate -
the skills Syncer relies on exactly that.

What fires it on a deploy is a **second, opt-in label**,
`asgard-ai.com/auto-fire-on-rollout: "true"`: the platform's apply step fires
the Syncers of the release that carry it and waits for them. Only that runner
reads the label; the Syncer module ignores it, and neither label reads the
other. **A suspended Syncer with no auto-fire label never runs at all**, and the
symptom is an empty drive or an agent with zero skills - with a green gate, a
succeeded run and no error anywhere.

**The polarity flipped.** Firing on deploy used to be the default, opted out of
with `asgard-ai.com/syncer-cd-trigger: "false"` - a label **the platform does
not read at all**, left from the CD workflows that predate the pipeline. Silence
is the default now, so noise is what has to be asked for. See
`../usecase/skill-set.md`.

## Knowledge Base

`../usecase/knowledge-base.md` has the `KnowledgeBase` + `Loader` +
`Indexer` + `Source` shape field by field, the way `knowledge-drive` has the
Drive one. Reading a chart that uses it is the case it exists for.

For recognising older charts only. Creating one needs a Name and an Alias Name
(lowercase letter first, then letters, digits and underscores).

Content is split across All, Manual Upload and Auto Load tabs.

**Manual Upload** takes CSV, XLSX, PDF, PPTX, DOCX and JSON Lines. A CSV goes
through three steps: Import File, Processing (preview the columns, choose which
to include, toggle Skip Header, pick an Identifier column - required), Finish.

**Auto Load** has three external sources: Crawler, Data Source, Asgard App. The
schedule is Daily or Weekly plus a time.

## What belongs in a Drive

Things a query cannot answer exactly: documents, FAQs, web pages. Counts, prices
and stock levels belong to a Semantic Model or a fixed query tool - putting a
number in a Drive makes the agent paraphrase a figure it should have read.

## Before writing the chart

`../usecase/knowledge-drive.md` has the full Drive-plus-Syncer shape.

## Ten Loaders, workspace-wide

A Loader is one recurring pull, and the platform allows **10 per Workspace** -
shared across every project in it, not per knowledge base. Indexers are capped at
150 and Processors at 500 on the same basis.

**A customer with a dozen document sources exceeds this before anything else in
the quota list**, and the failure arrives when the eleventh is created rather
than at design time. Two consequences worth carrying into an interview:

  - **count the sources, not the documents.** Fifty files behind one crawl is
    one Loader; five files from five places is five
  - a Drive with a Syncer is a different mechanism and is not counted here -
    which is one more reason it is the recommendation for new work

They are defaults rather than ceilings - `../wiki/integration.md` has all
eight numbers and how they are raised.

## Citations are possible, and they are not automatic

"Where did that answer come from" is on most customers' lists, and the answer is
**yes, if the Workflow is built to return them**.

Retrieval happens entirely server-side - there is no separate knowledge endpoint,
the same message call runs the retrieve processor and the model - and the sources
come back on `asgard.message.complete`, inside the message's `template`:

    fact.messageComplete.message.template.sources[]
      title, and url or fileName

**`template`'s shape is whatever the Workflow puts there.** So citations are a
design decision made when the chart is written, not a switch, and a front end
that does not read `template` shows an answer with no provenance no matter what
the retrieval did. Decide it before the chart, because retrofitting it means
touching the workflow and the front end together.

## Retrieval quality depends on how the question is asked

Worth handing over rather than discovering. The documentation's own guidance:

    specific keywords beat vague ones     "how many working days for a refund"
    one topic per question                not "tell me everything you have"
    context helps                         "as a business customer, what is the
                                          renewal process"

**A customer whose staff ask broad questions will judge the knowledge base as
bad**, and they will be describing their questions rather than the corpus. This
belongs in the handover and in the sample questions on the agent - `On-boarding
Settings` exists for exactly this, and a good set of starter questions teaches
the shape without anyone reading a guide.

## Sources

- [Drive](https://docs.asgard-ai.com/docs/product-suite/odin/features/drive)
  - asgard-docs `f00e0ee`
- [Knowledge](https://docs.asgard-ai.com/docs/product-suite/odin/features/knowledge-base-knowledge)
  - asgard-docs `f00e0ee`
- The syncer-class table and the contextIndex behaviour: checked 2026-09-02
  against [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube)
  `15ded0f` - `SyncerClass`, `SourceSetContextIndex`
- The schedule ordering and the two switches: read off a deployment's own Drive
- Preferring a Drive over `KnowledgeBase`: one deployment's migration, recorded
  in this repo's `source/SOURCES.md`. Held against
  [asgard-kube](https://github.com/asgard-ai-platform/asgard-kube) `15ded0f` on
  2026-09-02 - `knowledgebases`, `loaders`, `indexers` and `sources` all exist
  and none is marked deprecated - and against asgard-docs `f00e0ee`, which
  documents the feature as current

- Citations on `message.template.sources`, that retrieval is server-side on the
  same endpoint, and the question-shape guidance:
  [Knowledge base query](https://docs.asgard-ai.com/docs/developer-reference/examples/knowledge-base-query)
  - asgard-docs `f00e0ee`, read 2026-09-02

**Unchecked:** the syncer classes and contextIndex were held against the CRD and
one deployment; the UI steps come from the product documentation only.
