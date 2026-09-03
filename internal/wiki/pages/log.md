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
- `add` 2026-09-02 `platform-unknowns`, moved out of the customer repo's
  `docs/open-questions.md`. It was cross-engagement knowledge shipped as a copy
  into every repo, which is the arrangement the rest of this material exists to
  avoid: once an engagement filed its first question the file stopped matching
  its template, and after `scaffold --force` learned to preserve it the table
  could never be refreshed again. Two repos onboarded a quarter apart would have
  disagreed about it forever, with the older one still showing an answered
  question as open. Nothing read it programmatically, so the move changed no
  behaviour and made it searchable through `find` for the first time
- `add` 2026-09-02 P6 to that list, on whether the 30-step / 3-minute
  per-request ceiling is adjustable. Raised by a customer service requirement
  spanning product knowledge, a CRM and a ticket system, which is exactly the
  shape that reaches 30 steps
- `add` 2026-09-02 to `integration`: what the platform does not own - handoff to
  a human, pausing, resuming, per-user question counts. No CRD carries any of
  them. An earlier version of this section asserted that a LINE Official Account
  cannot run chat and a webhook at once, and that was retracted: neither source
  of truth says it, LINE's own building-a-bot page does not mention response
  modes, and it was being used to tell a customer something could not be built
- `add` 2026-09-02 to `operations`: Asgard is a hosted cloud service, so a system
  inside a customer's network is unreachable until they allowlist the four
  outbound addresses or bring the platform on over a VPN. The addresses were
  already here; what was missing was whose job the change is. From the FDE team,
  not from a document - raised by an engagement whose three scenarios all read
  internal systems
- `add` 2026-09-02 `setup-path`: the order from a handed-over credential to an
  agent a user can talk to, and the screenshot paths in asgard-docs for
  illustrating it. Every step was already documented on its own page; the
  sequence was not documented anywhere, which is the condition the README says
  obliges a page. Carries two corrections it is worth reading for on their own -
  an HTTP API's credential has no home under Settings, and Sindri has no import
  step because publishing is automatic
- `add` 2026-09-02 `screenshots`: an index of the product documentation's images
  - path, what each shows, and which situation it is for - fetched by URL rather
  than carried. Thirteen PNGs were briefly embedded instead and that was wrong:
  an image is not judgement, the copy would have gone stale against asgard-docs
  while looking current, and the site was already a source. The captions are
  harvested from asgard-docs' own alt text, so they describe the file rather
  than guess at it. Opening three of them corrected two things `setup-path` had
  asserted: the console is in English (only the captions are zh-TW), and the
  Data Source Provider list does show nine databases and no HTTP option

- `source` 2026-09-03 read asgard-core `internal/constants.go` at HEAD
  (2026-09-01) and re-read asgard-kube at `15ded0f`. Also cloned asgard-docs
  rather than fetching pages one at a time, which is how the counting below
  became possible at all
- `fix` 2026-09-03 **this entry was wrong and is corrected below.** It said five
  of thirteen processors declare a Failure output, from
  `ProcessorDefinitions`' relationship list. See the last entry on this page
- `ingest` 2026-09-03 `processors` gained the full per-processor table, from a
  `go/ast` walk of `ProcessorDefinitions` rather than a regex: outputs, required
  keys, which required keys carry a default, and which processors take arbitrary
  extra keys. The line worth the exercise is that **`semanticLayer.allowQuery`
  defaults to false and `semanticLayer.allowWrite` defaults to true** on both
  LLM processors and on `query-database` - the safe field off, the dangerous one
  on
- `fix` 2026-09-03 the definitions are **not** the config contract. `await` is
  documented on the streaming processor, set in five production deployments, and
  declared in neither asgard-core nor the CRD. A gate rule built on treating the
  list as complete called five of five correct charts wrong and was deleted.
  `processors` and `platform-unknowns` P10 both say so now
- `fix` 2026-09-03 `processors` and `workflow` said opposite things about ECMA5.
  The limit is `execute-script`'s Engine field, not every Expression: 520
  `expression:` values across the reference charts include an arrow function in a
  shipped tenant chart. Both pages carry the resolution and name each other at
  the point of the claim
- `ingest` 2026-09-03 `processors` gained `prevToolCalls`, which appears in **no
  documentation page** and was found by reading auto-post's two agent workflows.
  It carries each tool call's name, arguments and result, and is how a chart
  reacts to what an agent did without asking the model to say what it did. Also
  router-at-scale: a router's config keys **are** its branch names, and the
  production shape is a chain of one-boolean routers rather than a switch.
  `platform-unknowns` P11 asks what else is missing from the variable list
- `ingest` 2026-09-03 `coverage`, a new page: how many of eight rendered
  deployments declare each CR kind. **Trigger appears in one chart as one
  instance**; `KnowledgeBase`, `Loader` and `Source` only in auto-post; `Plugin`
  is 28 in one chart and none elsewhere. The trigger and knowledge-drive
  extracts now open by saying what they were written from
- `fix` 2026-09-03 the deploy guidance said CD refuses a project with no Syncer.
  That is one repository's workflow: another counts what the chart declares and
  skips at zero, and a production chart runs with none. Corrected in the deploy
  stage, the deploy gate's warning and two `add` notices
- `ingest` 2026-09-03 the deploy stage gained what a tag actually selects - the
  prefix is the whole decision and **prod is the fallback**, a tag deploys the
  commit it points at rather than the branch, `helm upgrade` runs without
  `--wait` so a green deploy means the CRs applied and nothing more, and the
  shared values file layers before the project's
- `fix` 2026-09-03 `api` cited a directory URL that 404s, and described
  `RESET_CHANNEL` as the only way into a channel - so the obvious front end
  destroys the conversation on every reload. The SDK's rejoin path is in the page
  now, along with the fact that the endpoint has two shapes in circulation and
  the one to use comes from the deployment. `platform-unknowns` P9
- `lint` 2026-09-03 fetched all 82 `docs.asgard-ai.com` links in the material.
  **Six were 404s.** Two directory URLs with no landing page, and four pages
  marked `draft: true`, which the site does not publish - the files are readable
  in a checkout, so the content is sound and only the links were broken. All six
  fixed or disclosed, and `asgard-cli audit-material --urls` now checks this
- `lint` 2026-09-03 counted the documentation properly against a clone. 162
  pages, 79 cited, **16 of them drafts**; 679 png files of which only 168 are
  referenced by any page. The screenshot index said "roughly a hundred" and then
  briefly said 6%, both wrong - it is 39 of 168, and was expanded the same day to
  113 of 168
- `ingest` 2026-09-03 `screenshots` grew by 74 images, every remaining one a page
  actually references and that carries alt text. Also: `integration/LINE` carries
  **three** images and the page said two, and the missed one is LINE's own
  Developers Console - not Asgard's, so Odin's redesign cannot date it, and
  usable in front of a customer today
- `lint` 2026-09-03 checked every enumerated set in the material against the CRD
  enums, and every completeness claim in it. **Nothing disagrees with the
  platform.** A third check was written and thrown away for flagging an extract
  that mentioned two of ten syncer classes, which is exactly right for an extract
  about a git-backed SkillSet
- `lint` 2026-09-03 checked all 98 backticked camelCase terms in the material
  against the CRD's 308 json field names plus the processor config keys. The 29
  that are not CRD fields are all correct - expression variables, API response
  fields, helm values, JavaScript builtins. **No retired field is being taught.**
  `audit-material --term` makes the next sweep one command
- `fix` 2026-09-03 `chat-channel` re-checked field by field against the CRD: all
  five credential blocks match, `generic`'s two fields are optional while every
  chat class's are required, and the CRD enforces `ExactlyOneOf` across the five
  class blocks. Also records the trap - grepping the parent directory for
  `botProviderClass: line` returns this tool's own scaffold output
- `fix` 2026-09-03 the systems-inventory example in the projects stage was
  modelled on one real customer. It is a composite now, and says so, because a
  worked example is the easiest place for the no-named-customer rule to slip
- `lint` 2026-09-03 the extracts index was a third out of date - eight of
  twenty-two written and never grouped. `asgard-cli usecase` listed them all
  along, so nothing broke; what was missing is the grouping, which is the only
  thing an index does. Both indexes now carry a rule saying adding a page means
  adding a row
- `lint` 2026-09-03 this log itself had no entry after 2026-09-02, while the
  material changed all day. The README's ingest procedure has appending here as
  step 5 and it was skipped every time. The entries above were written from the
  commit log afterwards, which is worse than writing them at the time: a line
  written a day later records what was done and not what was surprising about it
- `fix` 2026-09-03 swept the corpus for claims corrected in one page and left
  standing in another - the first failure shape in AGENTS.md, and the one this
  day's work was most likely to create. Four found: `integration` still handed
  out the `/generic` endpoint URL that `api` had stopped trusting; `skill-set`
  still said CD refuses a project with no Syncer; `index` still described the
  ECMA5 limit as applying to every expression and counted six variables in
  scope; `api` attributed the missing optional chaining to ECMA5 rather than to
  the variables being absent. **All four were pages nobody edited today**, which
  is exactly why the shape survives: the reader who corrects a claim is not
  reading the page that repeats it
- `ingest` 2026-09-03 `plain-chinese`, a seventh design-time skill in the
  scaffold, from the reading below. It was first folded into `proposal-deck` as
  prose and that was the wrong shape: the rules apply to every 繁體中文 thing an
  engagement sends a customer - a reply in `open-questions.md`, a cube
  `description` a model reads to choose what to query, a deployed agent's
  prompt - and a deck skill is not where any of those authors look. It is a
  skill now, `proposal-deck` points at it and keeps only the three shapes a
  slide produces, and the routing skill lists it
- `ingest` 2026-09-03 the `writing-humanizer` skill
  (github.com/shyuan/writing-humanizer), a zh-TW AI-writing-smell remover, read
  and its transferable half folded into `proposal-deck`. Not installed - it is
  written for essays and this is a deck. What transferred: the two absolute bans
  (`不僅……更是……` and 意義蓋章 endings), the four shapes that show up on slides
  (四字標籤清單, 句內粗體排比, 元論述, 升華結尾), the rule-of-three reflex, a
  vocabulary table, and the mandatory second pass with its rationalisation traps.
  **What did not transfer is the point**: that skill's strongest rule is that
  headings-and-bullets prose is AI, and a slide *is* a list - applying it whole
  would destroy the deck. The skill says so where an agent would reach for it,
  which is failure shape 3 in AGENTS.md handled at the source rather than
  discovered
- `fix` 2026-09-03 `plain-chinese` existed and almost nothing pointed at it.
  Ten places instruct an agent to write 繁體中文 and five had no route to the
  rules: the read-path stage and two other pages telling it to write cube
  descriptions, the CR display annotation section, the living spec's own
  language line, and both prompt templates. **The prompts are the ones that
  mattered** - an agent whose prompt says 賦能 answers in that register all day,
  to customers, and nobody reviews it after the first week. Each pointer went
  inside the imperative rather than near it, which is the third failure shape in
  AGENTS.md
- `fix` 2026-09-03 the first routing pass found five gaps by grepping 繁體中文,
  and that pattern was the wrong instrument: it finds places that *say* Chinese,
  not places that *are* Chinese. A second pass over the fields an author fills
  in found **nine more** - the cube and dimension `description`s in the semantic
  layer template itself, `sampleQuestions` (which the end user reads before
  anything else the agent says), the knowledge drive's indexing prompt, both
  `tooling.description`s and the four `*-description` UI annotations. Every
  generated CR that asks for Chinese now routes. Internal documents - decision
  records, meeting notes, task specs - deliberately do not, and the skill's own
  table says so rather than leaving it to be inferred
- `fix` 2026-09-03 the internal documents are routed too, on the FDE's call. The
  skill had carved them out - decision records, task specs, meeting notes as
  "your call, nobody outside reads them" - and that exemption is gone: a
  decision record outlives the engagement, and a task spec is read by whoever
  picks the work up. The request and task record templates, the decision and
  meeting-note templates, the living spec's module rules and each project's
  chart README all point at the skill now. **The one thing never rewritten is
  what the customer themselves said** - a quoted question, section 1 of a
  request, their words in a meeting note. Those are the record, and editing them
  destroys what they are for. Every file a scaffolded repo produces that asks
  for prose now routes; verified by generating one of each and checking the
  count
- `fix` 2026-09-03 `proposal-deck` was carrying its own copy of six rules that
  `plain-chinese` already had - including the inversion. Asked why it did not
  just point at the skill, and there was no answer: it was duplication written
  one commit after the rule against duplication. It is a pointer now, keeping
  only what is genuinely deck-specific, which turned out to be **when** to run
  the pass rather than any rule. The copy in `11-requirements` stays and now
  says why at the point of the copy: that text is printed by `asgard-cli next`
  and a skill file is only read if somebody opens it
- `lint` 2026-09-03 verified the routing both directions rather than asserting
  it again. Every reference to the skill resolves as a relative path from the
  file that makes it - 24 of them in a repo with one of every kind, checked by
  walking the paths rather than by eye, because they are written from four
  different directory depths. And every file that asks an author for prose has
  one: the last gap was the repo's own README, whose 說明 column and surrounding
  paragraphs are the first Chinese anybody reads. What is deliberately not
  routed is `requirements/requests/_index.md`, whose only TODO is a file path
- `fix` 2026-09-03 **the Failure-output claim made this morning was wrong**, and
  the text it replaced was closer to right. `ProcessorDefinitions`'
  `StaticRelationships` gives `http-request` Success only; the documentation
  gives it a Failure branch producing `prevError`, and four production charts
  across two repositories route `failure` off it. Five of the six processors
  named that morning document a Failure branch - only `update-context` does not.

  Found by writing a gate rule on that list, which flagged the four correct
  charts. The rule is deleted and the table it read is cut back to the one field
  a rule still uses, because a table that is partly wrong invites the next rule
  to be built on the wrong part - which is what happened.

  **Third time `ProcessorDefinitions` has proved incomplete**: `await`, the
  config keys, now the relationships. The lesson is not new and it caught the
  person who had written it down twice already
- `lint` 2026-09-03 held `internal/gate`'s pinned tables against the **generated
  CRDs** rather than the Go types they were extracted from, and the two are not
  the same document. `status` was in the enum table with three values from a
  Loader's job status; the CRD has six, because **Kubernetes' own condition
  schema uses that field name**. It is out of the table for the same reason as
  `type`, `format` and `alias`. The code's comment had also claimed status
  fields were skipped, which was never implemented - the comment described an
  intention.

  `baseAgentName` came out of the constraint table for a better reason:
  `SandboxBlueprint.spec.agents` is a **string** holding JSON, so the subagent
  shape is payload rather than schema and no walker over the decoded document
  ever reaches it. Nearly reported as drift between the types and the CRDs
  before this repo's own `gate/xref.go` explained it.

  `hack/check-tables.py` makes the comparison repeatable. It found both in its
  first run, which is the argument for it existing

