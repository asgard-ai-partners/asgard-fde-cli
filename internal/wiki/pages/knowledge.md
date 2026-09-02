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

## Two switches, and one does not do what its name suggests

`asgard-ai.com/syncer-suspend: "true"` stops the **scheduler** and **does not
stop CD**: the deploy step runs `kubectl create job --from`, which works on a
suspended CronJob. That is deliberate - the skills Syncer relies on it - so CD
cannot be changed to skip suspended ones.

To stop a Syncer running on deploy as well, it needs this repo's own opt-out
label `asgard-ai.com/syncer-cd-trigger: "false"`, which only CD reads. **Silence
takes both.** See `asgard-cli usecase skill-set` for the CD side.

## Knowledge Base

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

`asgard-cli usecase knowledge-drive` has the full Drive-plus-Syncer shape.

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

They are defaults rather than ceilings - `asgard-cli wiki integration` has all
eight numbers and how they are raised.

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

**Unchecked:** the syncer classes and contextIndex were held against the CRD and
one deployment; the UI steps come from the product documentation only.
