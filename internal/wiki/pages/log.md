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
  says why at the point of the copy: that text is printed by `asgard-cli status`
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

- `fix` 2026-09-03 recounted asgard-docs against a clone at `f00e0ee` and against
  the live site, because `index` and `TASK.md` disagreed and neither closed
  arithmetically. **162 files, 81 cited, 81 uncited, 12 of those drafts, so 69
  published and unread**; 4 cited files are drafts, which is exactly the "4 dead
  and disclosed" that `audit-material --urls` reports against docs.asgard-ai.com.
  Two counts by different routes agreeing is what the earlier figures lacked
- `fix` 2026-09-03 `index` claimed 130 of 162 cited. 130 was the **in-scope
  denominator** from the 2026-09-02 count - "130/130, complete for what is in
  scope" - restated as a count of citations. The 2026-09-03 recount is two lines
  above this one and the index was never updated against it, so for a day the
  section that exists to warn about a wrong coverage number carried one
- `fix` 2026-09-03 `index` said 32 files were deliberately excluded while its own
  table listed 29. Four `asgard-builtin` pages were pulled back into `processors`
  and the subtraction was made in the table and not in the prose
- `fix` 2026-09-03 `asgard-core` was cited by name in six places - this index,
  `processors`, `platform-unknowns` P10 and `internal/gate/processors.go`, whose
  contract is extracted from its `ProcessorDefinitions` - and by URL in none.
  `AGENTS.md` said "Ten repositories" and listed ten without it. Verified as
  `asgard-ai-platform/asgard-core` (private) and added; the count is eleven
- `fix` 2026-09-04 `asgard-cli status` was removed, and the `next` alias with it.
  The pointer on the `11-requirements` copy above named it; what prints the open
  questions now is `asgard-cli question`. The four record commands - `question`,
  `request`, `task`, `project` - each read one file back, and none of them says
  where an engagement is
- `fix` 2026-09-04 the alias table moved off `glossary` into `aliases.md`, beside
  the pages rather than among them. It is an index, and while it was inside the
  searched corpus it competed with what it points at: it lists every alias, so it
  was reliably the one document carrying every term of a translated query, and a
  search for a commerce subject returned the word list instead of
  `asgard-cli wiki taiwan-channels`. Verified before and after against that query
- `ingest` 2026-09-04 `aliases.md` gained a second table, for names a customer
  will say. These are **added** to a query rather than replacing it, because the
  name may be written verbatim in a page - SHOPLINE is - and replacing it would
  throw away the best answer there is. Only the six the material has actually
  been searched against are in it: the four on `taiwan-channels` that seven
  reference deployments were checked for, plus SHOPLINE and Shopee
- `lint` 2026-09-04 `audit-material --orphans` ran for the first time: **11 of 61
  documents are reached by no pointer**, `glossary`, `crd-rules` and
  `case-studies` among them. Nothing could report this before, because the
  pointers were recovered by regular expression at print time rather than held as
  a field; they are `kb.Doc.Links` now. `index.md` had also never been link
  checked at all - it is Unlisted, so it was never a source - and the first run
  including it found a pointer at `wiki log` that the checker did not know was a
  real invocation
- `fix` 2026-09-04 the one-meaning-here table is applied to a query now, not only
  read by a person. It had been prose on a page nothing pointed at while the
  failure it describes went on happening: `find payment` returns `fehu`, which is
  billing between Asgard and the customer, to somebody asking about the
  customer's own payment gateway. **Nothing was wrong with the result and nothing
  was recorded** - a search that lands is not a miss - so the miss log is blind
  to exactly the dangerous case. `payment` is now a row on `glossary`, and `find`
  prints the two senses above the results
- `ingest` 2026-09-04 a search naming a system this material never had now ends
  in a question rather than a phrasing hint. What decides the work is not which
  product it is but which of the four shapes on `taiwan-channels` it presents, and
  that is the customer's answer - so the dead end points there and offers
  `asgard-cli question add`. The catalogue still prints, last: an agent that
  searches twice and gets nothing twice falls back on what it already believed
- `ingest` 2026-09-04 the entity index split into two tables: names the material
  **covers**, and names it only **routes**. The first are the six somebody
  searched the reference deployments for. The second reach the shape the thing
  belongs to and nothing more, and `find` says so - **a row that routes reads
  exactly like a row that answers**, and a reader who cannot tell them apart
  takes results about a shape as results about a product. Payment gateways are
  the first five routed rows, added because a search for one was recorded and
  that is the test this index states
- `lint` 2026-09-04 the glossary page was being returned twice for one query -
  once as the sense block above the results and once as a search hit - now that
  its one-meaning table is applied at query time. Dropped from the results when
  the sense block has already printed the row: one finding shown twice reads as
  two
- `fix` 2026-09-04 command help is a link source now. Sixty-odd pointers into the
  corpus live in Long and Short strings and had never been resolved by anything -
  a page renamed out from under one would have gone dead silently. It also counts
  as a pointer for `--orphans`, because it is read at the moment somebody is
  deciding what to run, which is what discovery means here; the index still does
  not. Four of the eleven orphans were reachable from help all along
- `fix` 2026-09-04 **a pointer that wrapped across a line was invisible to
  everything.** This material is hard wrapped at about 78 columns, so one near the
  right margin is split in two, and both `find`'s counterpart and
  `audit-material --links` were matching a single space. Six real pointers in the
  corpus had never resolved, reading perfectly to a person the whole time. The
  pattern takes one space **or one wrap** - not any run of whitespace, which made
  a help screen's column padding resolve `asgard-cli guide` + "all of it" to a
  document called `all`
- `fix` 2026-09-04 the seven real orphans were closed with a sentence in the
  document whose reader needs the target, not with a link added to clear a list:
  `wiki knowledge`'s Knowledge Base half now points at its extract the way the
  Drive half already did; `guide verify` step 4 points at `crd-rules`, which is
  what that step catches; `guide deploy` names the state between two pieces of
  work; `guide init` names the guidance for the step it tells you to run;
  `proposal-deck` points at `case-studies` for a worked scenario and at
  `demo-generation` for what to promise when we have none of their systems;
  `add` points at `conventions`, which is what it generates applied. **0 of 61**
- `fix` 2026-09-04 `guide init` still said `asgard-cli project` "works out where
  the onboarding is" - a position claim left behind when `status` was removed
- `lint` 2026-09-04 the provenance pass ran over the nineteen documents that
  carried no `**Checked:**` or `**Unchecked:**` line - the twelve pieces of
  guidance and the seven skills - against asgard-kube `15ded0f`, the six
  reference repositories and the gate itself. **0 of 25, 0 of 21, 0 of 12,
  0 of 7.** It was done as a checking task and not a writing one, and four
  statements were wrong:
- `fix` 2026-09-04 `06-knowledge` said the Syncer member keys "was retired".
  `destinationMemberKey` and `stateMemberKey` are **deprecated and still
  accepted**, kept so pre-rename objects stay readable, so a chart that sets them
  passes lint, dry-run and the gate while being wrong
- `fix` 2026-09-04 `08-deploy` described the CD Syncer guard from a sample of
  two. Six reference repositories run that step and **three guard on the count
  while three do not** - an even split, which is why the instruction is "read
  your workflow" and not a rule
- `fix` 2026-09-04 `asgard-cr-verification` listed six reference kinds the gate
  resolves; it resolves seven. `SandboxBlueprint.pluginNames[]` to a `Plugin` was
  missing
- `fix` 2026-09-04 three documents stated **our** rules as the platform's: the
  two-`sampleQuestions` minimum, the byte-identical prompt text, and the refusal
  of `allowedCubes`. The CRD permits all three; the gate refuses them. A reader
  who cannot tell which will stop them cannot tell whether the fix is a chart
  edit or an argument with us
- `ingest` 2026-09-04 the CRDs' conditional CEL rules are checked. Of the 79,
  forty are `self == oldSelf` and cannot be seen in a render; the rest are two
  families - exactly one of a set of sibling fields, and a discriminator that
  implies its block - and nothing checked either. A credential setting both
  `value` and `valueFrom`, a `DataConnector` declaring two engine blocks, a
  `Syncer` declaring none, a `toolsetClass: mcp-server` with no
  `mcpServerConfig`: **every one of those renders, lints and passes a
  server-side dry-run, and is refused at apply.** `gate.Shapes` is X1/X2/X3
- `lint` 2026-09-04 the new check ran over the six reference deployments - 107
  CRs across unitech-e, buy123, finance-ai, xxentria, netbridge and freyr - and
  reported nothing, which is what a rule the platform already enforces should do
  against charts that are deployed. It was also run against a chart broken on
  purpose in all four ways and caught each
- `ingest` 2026-09-04 P12: `ImageGenerationModel`, `TranscriptionModel` and
  `SourceSetEditorServer` are CRDs in the contract with full schemas and appear
  in no documentation page and no material here. Recorded as an unknown because
  which of the two it is - the platform ahead of its documentation, or three
  kinds not meant to be reached for - changes the answer, and nobody has asked
- `fix` 2026-09-04 two findings that were living in TASK.md moved to the
  documents whose readers need them, and TASK.md's worklist is gone with them.
  **The interview is agent-shaped and so is whoever runs it** - every question
  after 2b assumes the deliverable is something you talk to, which cost one
  proposal - now sits at question 2b in `asgard-cli guide requirements`. **A rule
  that produces a good artefact of one kind silently produces a bad one of
  another** - six `proposal-deck` defects passed the skill's own checks - is a
  question in AGENTS.md's "Before you say it is done"
- `ingest` 2026-09-04 the conversation loop is written down. `flow-agent-supervisor`
  said "Workflow wf-<name> — the conversation loop" and never said what the loop
  is; it is four processors and five relationships, **edge for edge identical in
  three deployments** and one edge different in a fourth, read off every
  reference deployment at its prod values. The two-processor query tool is a
  separate shape and is named as one so the two stop being confused
- `lint` 2026-09-04 **a rule was invented and shipped, then reverted.** Reading
  one production Workflow that declared an exit, the generator's template was
  given `exits` plus a relationship graph, with a comment asserting that a
  processor reaching no exit "leaves the run with nowhere to finish", and a gate
  check was added for it. **14 of the 17 Workflows across every reference
  deployment declare `exits: []`**, and two single-processor ones declare no
  relationships either - the template was already right. Both changes are
  reverted. The failure is the one this material records twice: asserting from a
  sample without checking the rest, and this time the assertion was on its way
  into every customer repository
- `ingest` 2026-09-04 `wiki --sources` and `usecase --sources`, from an
  engagement that did it by hand: building a customer deck it opened each page it
  had used, read the Sources block at the foot, copied nine URLs out and checked
  each one itself. Every part of that except the checking was already known. A
  document's outbound documentation links are `kb.Doc.Sources` now, read at parse
  time beside its pointers, and `audit-material --urls` uses the same extractor
  instead of a second copy of the pattern
- `lint` 2026-09-04 **no extract cites a documentation link, and none of the 22
  does.** That is the convention - an extract assumes the reader knows the
  platform has the shape, and the counterpart page is where the links live - so
  `usecase --sources` says that and names the three commands that get there,
  rather than printing "none" and dead-ending
- `fix` 2026-09-04 two deck decisions moved to where the reader actually passes.
  `guide requirements` now says, before it points at the skill: **which of the
  three decks is it**, and **does it take images at all**. Both were in the skill
  and both were read too late. The engagement that measured it walked
  requirements first and the skill only after being corrected - so a rule inside
  the skill cannot be read before the work starts. It cost that engagement four
  downloaded screenshots, two cropped, none used: filling every page with the
  customer's own words leaves no room, and one kind of evidence per page then
  excludes the picture
- `ingest` 2026-09-04 `check` reports a command name in a customer repository
  that this build does not have. **Renaming a command silently breaks every
  repository already scaffolded**, and nothing detected it: `scaffold` writes
  AGENTS.md, the design-time skills and four READMEs and then never overwrites
  them - which is right, because an FDE edits them - so the old name stays until
  somebody types it and gets `unknown command`, then guesses. Found by an
  engagement doing exactly that. The command list comes from the cobra tree
  rather than a constant, because a constant is a second copy that goes stale at
  the moment of a rename
- `lint` 2026-09-04 the new check ran against a live engagement with a known
  answer: **12 of 12, exact match** with a list assembled by hand, including one
  buried in prose as the citation for a correction (`docs/open-questions.md:637`)
  rather than sitting in a command table. Zero on a freshly scaffolded
  repository, and flags neither `asgard-cli --help`, nor the `<command>`
  placeholder, nor a subcommand after a real command
- `fix` 2026-09-04 `brief` is named where an FDE passes. It was in no first
  screen at all: not in root help, not in the scaffolded `AGENTS.md`, not in the
  onboarding skill - the three things read before any work starts. An engagement
  cold-running the tool was never guided to it and said so. It is the command
  that exists for the riskiest activity in an engagement, and the reason it
  needs pointing at is the reason it exists: **talking to a customer changes no
  file**, so nothing derived from what the repository contains can raise it
- `fix` 2026-09-04 the deck-versus-questions warning compared modification
  times, so **any** edit to `docs/open-questions.md` silenced it. The edit that
  did was a command rename in the prose, made for an unrelated reason by an agent
  with no idea a warning was being switched off, and a meeting's answers were
  never written back while the gate said ok. It now compares the dates **inside**
  the rows against the meeting's own date, so prose cannot silence it - verified
  both ways in a scratch repository
- `lint` 2026-09-04 **and it still does not catch the case that produced it.** A
  meeting dated the same day as the newest row ties, and nothing orders a tie: a
  deck dated 2026-09-03 beside a question raised 2026-09-03 passes. A timestamp
  of any granularity has that hole. The shape that would catch it is a warning
  that does not clear itself - standing until somebody records that the questions
  were worked - and that needs `check` to write state, which it does not do. The
  limitation is in the code beside the check rather than left to be rediscovered
- `lint` 2026-09-04 the gate ran over the deployments its rules came from, for
  the first time. It had only ever run over charts this tool generates - which
  pass by construction - and over a scratch repository. **R1b was wrong about
  nine Agents in a running deployment**: it counted a semantic layer and a
  Toolset as capability sources and not a `SkillSet`, so every subagent of a
  flow-agent supervisor was told it had "no source of capability at all" while
  it had one. Fixed
- `lint` 2026-09-04 what the same run leaves standing, unadjudicated: 12
  `SourceSet` and 10 `Syncer` findings about the `managed-by=skill-set` label and
  a SourceSet shared by two SkillSets, 3 missing display annotations, 7 × R7
  (a published Agent with one sampleQuestion, where the label does say published),
  2 × R12 and one each on a Toolset and a BotProvider. **These are not noise and
  they are not mine to close** - each is either a chart to tell somebody about or
  a rule right for one shape applied to another, and telling those apart needs
  whoever built the chart. `hack/verify-references.sh` is how to see them again
- `fix` 2026-09-04 that script's first glob was `*-asgard-kube`, which silently
  missed `asgard-freyr-kube` - one of ours is named `asgard-<name>-kube` - and
  with it 15 findings. A glob makes that kind of miss without saying anything
- `fix` 2026-09-04 R12 was an agent-hub rule firing on flow-agent subagents, and
  it is scoped now. Measured across every reference deployment: the five
  agent-hub Agents share **one** `prompt.task`; the seventeen blueprint
  subagents have **thirteen** distinct ones, six of them empty because their
  prompt lives on the Workflow's processor. Three supervisor deployments out of
  three, so it is the convention rather than a mistake three engagements made -
  and a subagent's task is what makes it a specialist. Verified as scoping and
  not disabling: an agent-hub Agent with a deliberately altered task still fails
- `lint` 2026-09-04 **R7 was suspected of the same error and is not wrong.** The
  evidence went the other way: freyr's six subagents carry no `agent-published`
  label at all, finance-ai's three carry it and have exactly two sample
  questions each, and xxentria's nine carry it with one question in seven of
  them. So the label is meaningful on a subagent rather than boilerplate, and
  seven Agents in a running chart are published with one question. Either the
  questions or the label is wrong there, and both are that chart's to decide
- `fix` 2026-09-04 three more rules were right for one shape and firing on
  another, all found by running the gate over the reference deployments:
  **the SkillSet pairing rules** demanded `managed-by=skill-set` and one
  SourceSet per SkillSet, exempting only a SkillSet a Plugin bundles - and none
  of the four deployments using the shared shape has a Plugin at all. What the
  shape costs is UI presentation, not acceptance, so it warns; the 1:1:1 case
  with a genuinely absent label still fails, verified by removing one on purpose.
  **The Syncer path rule** reported "missing destinationPath" on three Syncers
  that set `destinationMemberKey`, the pre-rename spelling the CRD still accepts
  - the field was not missing and the object was not refused, so it is a
  migration to name. Same for `stateMemberKey`
- `lint` 2026-09-04 the run over six deployments went 36 findings to 12, and the
  12 are read rather than counted: **7 x R7** (a published subagent with one
  sample question, and the label is meaningful because one deployment omits it),
  **4 missing display annotations** on a BotProvider, three SandboxBlueprints and
  a Toolset - which is the exact defect `add` exists to prevent, "applies cleanly
  and appears with no name in the UI" - and one Toolset. Every one of those is a
  chart's to fix, not a rule's to stop asking
- `lint` 2026-09-05 `audit-material --commands`: every `asgard-cli <command>` this
  material writes now resolves against the command tree, and the tree and the
  material are both in CI (the `material` job, which also runs `--links` - they
  existed as flags and had never been wired to anything). Written because
  `asgard-cli pipeline deliveries` was named in six documents as the one place a
  push that produced no run explains itself, and no such command existed: it was
  found by a person re-reading a provenance line, weeks later
- `fix` 2026-09-05 the first run resolved 629 references and found **4 dead
  flags**, none of them a command name: `project add --env` in the scaffolded
  `AGENTS.md` and the projects prompt (the environment model went with the
  Pipeline cut-over), `size --plain` in the proposal-deck skill (the plain
  reading is `size <shape>`'s second output and was never optional) and
  `add --connector-less` in `api-oauth` (an `httptool` takes `--toolset`, and has
  never taken a connector). `project add`'s own help still described `--env` and
  `--force` in prose, which is the same defect one level down and is not
  mechanically caught - a flag is only resolved where it is written after an
  invocation
- `fix` 2026-09-05 `asgard-cr-verification` left this binary. It described
  `deploy.yaml`, per-environment values files, a python CRD-fidelity script and
  cloning asgard-kube to find out whether a field exists - **a month after the
  Pipeline cut-over removed all four** - and nothing was ever going to correct
  it: `scaffold` does not overwrite a file that exists, and no version number
  covered it. It is now served from the platform's `/v1/docs/skills` alongside
  the extracted CR shapes and processor catalogue, and rewritten on every
  `asgard-cli skill update`. The line for what stays here is **authority, not
  subject**: how to run a local gate is true of any Asgard, what a server
  accepts is not
- `lint` 2026-09-05 `asgard-cli gate`: one command for everything a machine with
  no cluster can check - tools, repo, the reference material's freshness, lint,
  render, xref. Every part existed; the only thing assembling them was prose,
  and prose is what went stale. **A skip is not a pass**, and the two print
  differently, because "kubectl was not installed so the cluster step did not
  run" was reported as green in the gate this replaces
- `fix` 2026-09-05 **the bare `helm lint` instruction had been wrong since the
  Pipeline cut-over.** A chart must not declare the reserved `asgard` block -
  the platform injects it and declaring it is a warning on every plan - so
  linting with no `-f` at all fails on every chart that reads
  `.Values.asgard.projectEnvironmentId`, which is every chart that labels
  anything. Four of four charts in a real repository, all for that reason and no
  other. The gate supplies that one file and nothing else, which keeps the
  property the bare form was for: every OTHER `.Values.*` still needs a default
  in the chart's own `values.yaml`. Verified both ways - a chart with a genuinely
  undeclared value still fails
- `fix` 2026-09-05 `kubectl` was still described as "server-side dry run and CRD
  fidelity, against the cluster" and counted as a tool the gate needs. Both
  moved to the platform's plan at the cut-over and no client is issued cluster
  credentials, so it is optional now and `doctor` says why it is there
- `fix` 2026-09-05 the top-level help was a flat list of thirty commands, which is
  what made four commands that **compose** look like four commands that
  **compete**. It is four groups now - Ask, Build, Check, Deploy - and the
  grouping is the source rather than a rendering of it: a command is added by
  naming its group, so adding one without deciding where it goes does not
  compile. Within a group the order is insertion, not alphabetical, because
  alphabetical put `check` and `doctor` in front of `gate`
- `fix` 2026-09-05 `check`, `render`, `verify` and `doctor` each say which step of
  `asgard-cli gate` they are. Three of them still described the **four-step**
  gate of the pre-Pipeline world - `check` called itself "the first step", and
  `verify` called itself "steps 2 and 3" and sent the reader to a step 4 that
  needs a cluster no client is given
