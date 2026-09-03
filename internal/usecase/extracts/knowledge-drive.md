# Knowledge drive

**The `KnowledgeBase`, `Loader` and `Source` shapes here come from one
deployment** - auto-post, the platform's own - which is the only one of eight
that declares any of them (`asgard-cli wiki coverage`). That cuts both ways: it
was written by the people who built the CRs, and it is not a customer's
constraints.


Unstructured knowledge - documents, FAQs, web pages - as a mounted `SourceSet`
with a knowledge graph over it.

**Seen in:** a public widget whose knowledge comes from a product catalogue
database, a crawl of its own marketing site, and manually uploaded documents.

**Checked:** 2026-09-02 against a Drive with two Syncers and a contextIndex, and the CRD. Two required Syncer fields and a misplaced column flag were corrected.

**Unchecked:** the contextIndex.prompt guidance. It is advice about the customer's own data and has no source outside the engagement that wrote it.

**Read the platform side first:** `asgard-cli wiki knowledge` -
Drive, Context Index, and how Knowledge Base differs. This page assumes you have.

## When this shape, and when not

Use it when the customer's knowledge is **not rows in a database**: product
documents, FAQ spreadsheets, pages on a website. The agent queries a knowledge
graph and then reads the few files it points at.

Do **not** use it for structured data that a query answers exactly. The two are
complementary and neither does the other's job: a Drive answers "what is this
machine roughly, how do I choose, how do I fix it"; a query answers "how many,
which ones, what is the phone number".

**Prefer this over `KnowledgeBase` for new work, but know what that claim
rests on.** `KnowledgeBase`, `Loader`, `Indexer` and `Source` are all live CRDs,
none carries a deprecation marker, and the console ships the feature with its
own documented UI. What happened is narrower: one engagement built knowledge on
`KnowledgeBase` + `Loader` + a retrieval workflow and moved it to a Drive with a
Context Index (TASK-013). So an older chart containing one is not automatically
wrong - it is a shape somebody chose before that experience existed. Ask the
platform team before telling a customer the mechanism is going away.

## The shape

    SourceSet  ss-<name>-knowledge     declares no members - the paths its
      catalog/                         Syncers write to are what is in it
      website/
      docs/                            uploaded by hand, no Syncer
      .context-index/                  the platform's, from spec.contextIndex
      <- SandboxBlueprint.sourceSetMounts, readOnly at /knowledge

`spec.contextIndex` is **the whole switch**: setting it makes the reconciler
derive three same-named CRs (a Workflow, a SandboxBlueprint and a Trigger) that
mount the Drive **writable** and run the indexer on a cron.

## Generate it

    asgard-cli add knowledgedrive <name> --connector dc-<name>

That writes the structure below with the fields that fail silently already in
place. **Copying a skeleton by hand is where those get lost**, because nothing
tells you they are missing: not helm lint, not CRD validation, not a server
dry-run. The generated file marks the judgement calls TODO - those are what the
rest of this page is about.

## The skeleton

`templates/source_set/ss-<name>-knowledge.yaml`, plus one Syncer per member.

```yaml
apiVersion: asgard-ai.com/v1alpha1
kind: SourceSet
metadata:
  name: ss-<name>-knowledge
  annotations:
    asgard-ai.com/source-set-name: "<display name>"
    asgard-ai.com/source-set-description: "<what knowledge this holds>"
  labels:
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  apiKey:
    valueFrom:
      secretKeyRef:
        key: asgard_resource_api_key
        name: {{ include "<chart>.appSecretName" . }}
  # No member registry: what is in the Drive is whatever its Syncers write,
  # plus anything uploaded by hand. The paths are the truth.
  #
  # Setting contextIndex derives three same-named CRs that build the graph on a
  # cron. To pause, label the SourceSet context-index-suspend - do not remove it.
  contextIndex:
    cron:
      schedule: {{ .Values.<name>Knowledge.contextIndex.schedule | quote }}
      timeZone: {{ .Values.<name>Knowledge.timeZone | quote }}
    prompt: |-
      <domain knowledge only - this is appended to the platform's own indexing
      instructions. e.g. newest partition wins per product_id>
---
apiVersion: asgard-ai.com/v1alpha1
kind: Syncer
metadata:
  name: syn-<name>-catalog-db
  annotations:
    asgard-ai.com/syncer-name: "<display name>"
  labels:
    asgard-ai.com/syncer-suspend: {{ .Values.<name>Knowledge.catalogSync.suspend | quote }}
    {{- include "<chart>.labels" . | nindent 4 }}
spec:
  sourceSetName: ss-<name>-knowledge
  # Relative path inside the volume: no leading /, no . or .. segment, no //.
  # Must end with / for database and web syncers.
  destinationPath: "catalog/"
  # Where the incremental cursor is kept. Must NOT end with /.
  statePath: ".syncer-state/<name>-catalog"
  syncerClass: database
  schedule: {{ .Values.<name>Knowledge.catalogSync.schedule | quote }}
  timeZone: Asia/Taipei
  database:
    dataConnectorName: dc-<system>
    batchSize: {{ .Values.<name>Knowledge.catalogSync.batchSize }}
    # Immutable. Changing the projection means a new Syncer.
    columns:
      - name: product_id
        isIdentifier: true
      - name: <...the rest of the projection...>
      # Strictly-greater-than cursor, kept in the Syncer status. At most one.
      - name: row_updated_at
        isMaxValueColumn: true
    query: |
      select ... from ...
```

**`isMaxValueColumn` and `isIdentifier` sit on a column, not on the `database`
block.** Written one level up, beside `dataConnectorName`, they are unknown
fields: the apiserver drops them without a word, and what is left is a Syncer
that re-reads the whole table every run. Nothing in the rendered chart, in
`helm lint` or in the apply output says so.

`batchSize` and `query` are both **required** - a `database` Syncer missing
either is rejected at apply time, which in practice means during CD.

The Syncer wraps the query as

```sql
select <columns> from (query) where <cursor> > $cursor order by <cursor> asc
```

so every name in `columns` has to appear in the projection spelled exactly the
same way, and the cursor column has to be comparable.

Mount it read-only from the blueprint:

```yaml
  sourceSetMounts:
    value: '[{"sourceSetName": "ss-<name>-knowledge", "mountPath": "/knowledge", "readOnly": true}]'
```

## The index runs after the Syncers, not with them

`contextIndex.cron` and each Syncer's `schedule` are independent fields and
nothing orders them. Put the index **after** the Syncers on the same day - the
deployment runs the two Syncers at 09:00 and the index at 10:00, both
`Asia/Taipei`.

Reversed or simultaneous, the index walks the volume before the day's content
lands and the graph describes yesterday, every day, without ever failing. An
incremental `--update` over an unchanged Drive finishes in seconds, so the gap
costs nothing.

The derived CRs are named after the SourceSet with a `-ci` suffix, so a Drive
called `ss-<name>-knowledge` produces `ss-<name>-knowledge-ci`. That is what to
look for on a cluster when the index is not running.

### Two switches, and one of them does not do what its name suggests

`asgard-ai.com/syncer-suspend: "true"` stops the **scheduler** and **does not
stop CD**: the deploy step runs `kubectl create job --from`, which works on a
suspended CronJob. That is deliberate - the skills Syncer relies on it - so CD
cannot be changed to skip suspended ones.

To stop a Syncer running on deploy as well, it needs this repo's own opt-out
label `asgard-ai.com/syncer-cd-trigger: "false"`, which only CD reads. **Silence
takes both.** See `asgard-cli usecase skill-set` for the CD side.

## The pause that is not a delete

**To pause indexing, label the SourceSet `context-index-suspend: "true"`.**

Clearing the `contextIndex` field instead tears down the three derived CRs and
**renames the index aside** - so the next run rebuilds the graph from scratch.
The field is a switch for existence, not for scheduling.

## Designing it - the part the generator leaves TODO

### What goes in the Drive, and what does not

Documents a person would read to answer the question: product literature, FAQs,
policies, the pages of a public site. **Not** anything a query answers exactly -
counts, prices, stock, contact details. Those belong to a query tool, and putting
them in a Drive makes the agent paraphrase a number it should have read.

The two are complementary: the Drive answers "what is this, how do I choose, how
do I fix it"; a query answers "how many, which ones, what is the number".

### `contextIndex.prompt` - what the indexer needs to know

It is appended to the platform's own instructions, so only domain knowledge
belongs there. In practice, three things:

- **what each folder holds**, one line each
- **the key, and which copy wins.** With incremental sync, the same record
  appears in several dated partitions - say which one is current, or the graph
  treats stale versions as facts
- **what a field means** where the name does not carry it

### Telling the agent how to read it

Put this in the consuming agent's prompt, not in the Drive:

    先用 graphify 查 /knowledge 的知識圖,拿到相關檔案與段落後再去讀那幾個檔案 ——
    不要自己遍歷整個 Drive 逐檔閱讀。

Without it the agent reads everything, slowly, and still misses things.

### The gap that embarrasses a demo

Manually uploaded documents are not there until someone uploads them, and the
graph is not useful until it has run once. **Say so in the chart README as a
post-deploy step with an owner**, because the failure mode is a demo where the
agent answers the structured questions perfectly and the knowledge ones badly.

## Fields that are not obvious

**`destinationPath` must end with `/`** for database and web syncers;
`statePath` must not. Both are relative paths inside the volume, and the CRD
rejects a leading `/`, a `.` or `..` segment, or `//`.

**The folder does not need declaring anywhere.** Writing to it is what creates
it - which is also why a typo produces a second, silently empty folder rather
than an error.

**The database Syncer is incremental.** `isMaxValueColumn` is a
strictly-greater-than cursor kept in the Syncer's status; a day with no changes
writes nothing. Two consequences worth writing into the CR header:

- an updated row leaves its old copy in an older partition, so the index prompt
  has to say newest-partition-wins per key
- a soft-deleted row never disappears from old partitions. A full rebuild means
  deleting the partitions and clearing the sync state

**`spec.database.columns` is immutable.** Changing the projection means a new
Syncer, not an edit.

**Keep the web Syncer's page list in version control** rather than enabling a
deep crawl. What the agent can see should be reviewable.

**`contextIndex.prompt` is appended** to the platform's own indexing
instructions, so only domain knowledge belongs there.

## Querying it

Mount read-only, and tell the agent in its prompt to **query the graph first and
then read only the files the graph points at**. Without that instruction it will
crawl the whole Drive.

`readOnly: true` is part of the security argument, not a detail: an agent that
answers questions about knowledge has no reason to be able to change it.

## The manual step that has to be written down

Documents nobody can sync automatically - the PDFs, the spreadsheet of FAQs -
have to be uploaded after deploy, and the Syncers and index have to run once.
**Until then the knowledge answers are poor**, and a demo will find that out
before you do. Put it in the chart README as a post-deploy step rather than
leaving it to be discovered.

## Verify

```bash
asgard-cli check
asgard-cli verify <project>
```

The xref check resolves `sourceSetName` and `database.dataConnectorName`, and
validates the path rules on `destinationPath` / `statePath`.

After deploying, confirm the first sync actually wrote something before judging
the answers:

```bash
kubectl get -n <namespace> syncers.asgard-ai.com
kubectl describe -n <namespace> syncer syn-<name>-catalog-db   # check syncState
```
