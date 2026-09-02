# log

Append-only. One prefix per line, so it stays scannable and parseable:

    ingest   read a new source, wrote or rewrote pages
    lint     ran an audit, recorded the result
    fix      fixed something an audit found
    source   the raw sources moved to a new commit

**Only ever appended to.** A line written wrongly is corrected by another line,
never by editing the old one - what this layer is for is what happened at the
time, and editing removes exactly that.

---

- `source` 2026-09-02 read asgard-docs at `f00e0ee` (2026-08-31) and asgard-kube
  at `15ded0f`
- `ingest` 2026-09-02 created the wiki; the three layers and the conventions went
  into `README.md`
- `ingest` 2026-09-02 `product-suite`, from the product suite index and the
  features overview
- `lint` 2026-09-02 `overview/asgard-features` is marked `draft: true` and links
  to `/docs/user-guide/...` paths that stopped existing in the 2026-08-31 rebuild.
  **Its detail cannot be relied on**; only the two lists corroborated elsewhere
  (model providers, integration outlets) were taken, and the source block says so
- `ingest` 2026-09-02 `agents`, from the Managed Agent and Flow Agent pages
- `lint` 2026-09-02 the product vocabulary and the CRD's do not line up, and the
  gap misleads: the UI's Flow Agent is three CRs (`BotProvider` + `Workflow` +
  `SandboxBlueprint`), while `agentClass` has only `managed` - so "Managed Agent"
  and "Flow Agent" look like a symmetric pair and are not. The mapping table went
  into `pages/agents.md`
- `ingest` 2026-09-02 `index.md` listed ten uncovered subjects with their source
  paths
- `ingest` 2026-09-02 `knowledge`, `semantic-model`, `tools`, `automation`,
  `settings`, `mimir`, `sindri` - from every file under `product-suite/odin/features/`,
  `mimir/features/` and `sindri/features/`
- `lint` 2026-09-02 the Drive Syncer UI offers five sources (Google Drive,
  OneDrive, Git, Web Crawler, Data Source) while the CRD's `SyncerClass` has ten -
  bot, dropbox, ftp, sftp and smb can only be declared in a chart. Recorded in
  `knowledge`
- `lint` 2026-09-02 Settings -> Connection's "For Trigger" group lists Google
  Sheets, OneDrive Workbook and others whose trigger classes were removed from
  `TriggerClass`; only cron remains. Recorded in `automation` and `settings`
- `lint` 2026-09-02 the Semantic Model page lists six data sources, the Data
  Source page nine. The documentation gives no reason; marked unconfirmed
- `fix` 2026-09-02 removed the chart-level cautions repeated on several pages
  (allowWrite, Toolset tool fields, the SkillSet trio, allowedCubes). Those belong
  to `asgard-cli usecase` and to the templates; the wiki points at them
- `fix` 2026-09-02 the README gained the "the reader is an agent" framing: these
  pages exist because an agent picking up a customer repository has only that
  repository, which never describes the platform
- `ingest` 2026-09-02 the remaining coverage: `console`, `integration`,
  `workflow`, `case-studies`, `fehu`, `operations`. Sources were
  `management-console/`, `integration/`, `integration-with-asgard/`,
  `developer-reference/`, the case studies, `fehu/` and `help-community/`
- `lint` 2026-09-02 the four `integration-with-asgard/` pages (api, line, slack,
  discord) are all `draft: true` and describe a "Published -> add integrated"
  interface inside a Project, which does not match Odin's current Applications ->
  Customized Integration. Marked possibly stale
- `lint` 2026-09-02 the glossary's Processor entry lists an older set of nodes and
  does not match the current `ProcessorType`; `operations` says to take `workflow`
  as current
- `lint` 2026-09-02 `integration/Discord.mdx` is an empty file
- `source` 2026-09-02 the product documentation states plainly that Expression is
  JavaScript and Template is Handlebars, and asgard-kube `15ded0f` makes no claim
  either way (its only CEL references are the CRD's own validation rules). The
  conflict recorded in this repo's `source/FINDINGS.md` - documentation saying CEL while
  charts wrote JavaScript - no longer exists in the current sources
- `fix` 2026-09-02 decided not to write pages for the variable and message
  template lists under `developer-reference/asgard-builtin/`: they are lookup
  material, and copying them here only produces a copy that goes stale. `index`
  says to read the source directly
- `lint` 2026-09-02 measured source coverage by listing all 162 asgard-docs files
  against every page's source block. The first measurement found only 65 cited.
  The earlier claim of complete coverage was wrong
- `ingest` 2026-09-02 filled the 48 genuinely missing sources, added the `api`
  page (endpoint, SSE sequence, four integration patterns, SDK and migration), and
  extended existing pages: `prevMessage` / `history` and the inputSchema editor in
  `workflow`, LLM Completion troubleshooting and OpenAI-compatible models in
  `operations`, Odin's Workspace layer and the Project quota in `console`, user
  settings in `sindri`
- `lint` 2026-09-02 measured again: 130/130, complete for what is in scope. The
  denominator excludes the 32 files deliberately left out - 5 superpowers, 10
  release-note entries, 18 asgard-builtin references
- `fix` 2026-09-02 every page gained an `**Unchecked:**` line. The wiki is checked
  less deeply than `usecase` by nature, and not saying so lets a reader assume the
  two are equally reliable
- `fix` 2026-09-02 `tools` claimed the SkillSet 1:1:1 rule without exception,
  contradicting the 28-Plugin deployment that shares one store. The page had been
  written from the product documentation and never held against a chart; the
  exception was added
- `fix` 2026-09-02 translated every page into English and turned the source
  blocks into docs.asgard-ai.com links. The reasoning that the wiki should match
  its zh-TW sources did not hold: an agent asked in Chinese queries in English, so
  the corpus does not have to carry both languages, and one language removes the
  split where a Chinese question could reach only the wiki and an English one only
  the extracts
